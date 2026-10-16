package cli

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// view is how the items of a read show in a table. T is the typed subset of
// an item shown there, decoded from the response; `-o json` and `-o yaml`
// print the whole response instead.
type view[T any] struct {
	singular  string                     // item name for help text, e.g. "price list"
	header    []string                   // table columns; the first is the identifier that get takes
	long      []string                   // columns of long values, which tables cut (csv and -o json: whole)
	localized bool                       // the table shows translations, so its reads take --language
	row       func(T, language) []string // the cells of an item, in the language --language picks
	legend    string                     // what each read of the items adds to its text, e.g. what columns mean
}

// reader reads with a token limited to scopes and shows the items it reads
// through its view (command), or saves the file it reads (fileCommand).
// Every read command comes from a reader, built by newRead, so no command
// can pick its own scopes.
type reader[T any] struct {
	view[T]
	scopes []string // OAuth scopes besides the tenant; a write-capable one makes its reads opt-in
}

// typeNames returns the reader of a list of type names, a JSON array of
// strings such as the unit types, shown as one TYPE column.
func typeNames(singular string) reader[string] {
	return reader[string]{view: stringsView(singular, columnType)}
}

// stringsView shows items that are strings, such as names, in one column.
func stringsView(singular, column string) view[string] {
	return view[string]{
		singular: singular,
		header:   []string{column},
		row:      func(s string, _ language) []string { return []string{s} },
	}
}

// command declares one read command of a reader; newRead builds it, its
// fetch aside.
type command struct {
	use     string              // name and arguments, e.g. "get TYPE ID": {parent}, {inner} and {id} in req
	short   string              // summary
	long    string              // what it reads and leaves out; the view's legend and the scope sentence follow
	req     request             // what it requests
	fetch   fetcher             // a listing or a single response
	choices map[string][]string // the only values that an argument takes, by its name in use
	// predefined holds values that an argument takes besides others, by its
	// name in use, e.g. TYPE: PRODUCT and ORDER beside the custom entity
	// types; they complete as its choices do.
	predefined map[string][]string
	ids        argumentIDs // the reads that list the IDs of its arguments, for completion
}

// fetcher is how a read command reads its items: a listing, a single response,
// or a search, sent as POST or PUT, read as one of those.
type fetcher interface {
	// bind adds the flags of the read to cmd, and may add to its help, and
	// returns how it reads and where it finds its items.
	bind(cmd *cobra.Command) (readFunc, shape)
}

// listing lists every item of a collection, with --limit and the flags of
// filters, which it passes to its read as list options. A collection whose
// items are a plain array (items nil) also takes --stream.
type listing struct {
	filters listFilters    // what the service supports besides --limit
	read    collectionRead // how it reads the collection; GetAll when nil
	items   shape          // where the items are in the response (table and -o json); listItems when nil
}

func (l listing) bind(cmd *cobra.Command) (readFunc, shape) {
	var opts emporix.ListOptions
	l.filters.register(&opts, cmd)
	read, stream := l.collection(), false
	if l.items == nil {
		addStreamFlag(cmd, &stream)
	}
	listed := func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		if !stream {
			return read(c, ctx, path, query, opts)
		}
		run, out := opts, cmd.OutOrStdout() // a stream writes to this run's output only
		run.Stream = func(items []json.RawMessage) error { return output.WriteJSONLines(out, items) }
		if _, err := read(c, ctx, path, query, run); err != nil {
			return nil, err
		}
		return nil, errStreamed
	}
	return listed, l.found()
}

// glance implements glancer: it reads the first page of the collection, or
// the whole collection where the service sends it in one response
// (emporix.ListOptions.FirstPage).
func (l listing) glance() (readFunc, shape) {
	read := l.collection()
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		return read(c, ctx, path, query, emporix.ListOptions{FirstPage: true})
	}, l.found()
}

// collection returns how l reads its collection: read, or else GetAll.
func (l listing) collection() collectionRead {
	if l.read == nil {
		return (*emporix.Client).GetAll
	}
	return l.read
}

// found returns where l finds its items in the response: items, or else
// listItems (a plain array).
func (l listing) found() shape {
	if l.items == nil {
		return listItems
	}
	return l.items
}

// errStreamed is what a listing that streams returns once it printed its items.
var errStreamed = errors.New("the items were streamed")

// single reads one response and finds its items in it.
type single struct {
	items shape    // where the items are in the response (table and -o json)
	read  readFunc // how it reads the response; Client.Get (an empty body is not found) when nil
}

func (s single) bind(*cobra.Command) (readFunc, shape) { return s.glance() }

// glance implements glancer: one response, as bind reads it.
func (s single) glance() (readFunc, shape) {
	if s.read == nil {
		return (*emporix.Client).Get, s.items
	}
	return s.read, s.items
}

// emptyAs returns the read of one response that Emporix may send with an empty
// body, as Client.Check reads it, with the empty body read as answer, not as
// the not found that Client.Get reports for it: "{}" where the item holds
// nothing, such as the mixins of a site without any, or "true" where a check
// succeeds without a body (statusCheck).
func emptyAs(answer string) readFunc {
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		body, err := c.Check(ctx, path, query)
		if err != nil || body != nil {
			return body, err
		}
		return json.RawMessage(answer), nil
	}
}

// readFunc reads path with query: (*emporix.Client).Get, or a collectionRead
// with the list options of its command. A listing that streams (--stream)
// prints its items itself and returns errStreamed.
type readFunc func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error)

// collectionRead reads a collection, e.g. (*emporix.Client).GetAll (every
// page) or GetUnpaged (a collection the service returns whole).
type collectionRead func(c *emporix.Client, ctx context.Context, path string, query url.Values,
	opts emporix.ListOptions) (json.RawMessage, error)

// inOnePage reads a collection as one page of up to size items
// (GetAllInOnePage), for a service whose pages cannot be combined.
func inOnePage(size int) collectionRead { return pageOf((*emporix.Client).GetAllInOnePage, size) }

// onFirstPage reads a collection as one page of up to size items checked
// against its total (GetAllOnFirstPage), for a service that lists it on its
// first page only.
func onFirstPage(size int) collectionRead { return pageOf((*emporix.Client).GetAllOnFirstPage, size) }

// pageOf binds read, a read of a collection as one page, to its page size.
func pageOf(read func(*emporix.Client, context.Context, string, url.Values, int, emporix.ListOptions) (
	json.RawMessage, error), size int) collectionRead {
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values,
		opts emporix.ListOptions) (json.RawMessage, error) {
		return read(c, ctx, path, query, size, opts)
	}
}

// onePage names the page that inOnePage(size) or onFirstPage(size) reads in
// help text: "one page of up to size".
func onePage(size int) string { return "one page of up to " + strconv.Itoa(size) }

// resource declares an Emporix collection, read with `list` and `get ID`
// (newResourceCommand), or, one that Emporix lists only by search, with
// `search` and `get ID` (newSearchedResourceCommand), which filters, read and
// items then configure as they configure list. A collection that belongs to
// one item is a child (childCommand), a read that fits neither is a command
// of a reader, and the view a storefront gets without a scope is a `public`
// resource in the group of its resource. Its writes are declared as writes
// (write) to its path or itemPath, among the extra commands of its group;
// createWrite, upsertWrite and deleteWrite declare the usual ones.
type resource[T any] struct {
	view[T]
	use       string         // command name, e.g. "price-lists"
	noun      string         // the items in help text; use without hyphens when empty
	short     string         // summary of the group; "Read " + noun when empty
	long      string         // what the reads return and leave out
	about     string         // the group's help where it says more than long, e.g. what its writes change
	under     string         // the commands above the group, if any, e.g. "shipping"
	listedBy  string         // a group without list: the command line, without epx, that lists the items
	atPath    bool           // listedBy lists the items at path, as for a nested group (nested)
	path      string         // collection path; may name {site}
	query     url.Values     // query of list and get, e.g. showUnpublished=true
	listQuery url.Values     // query of list only, e.g. siteCode={site}
	scopes    []string       // OAuth scopes besides the tenant; every read requests them
	filters   listFilters    // what list takes besides --limit: what the service supports
	read      collectionRead // how list reads the collection; GetAll when nil
	items     shape          // where list finds the items in the response (table and -o json); listItems when nil
	item      shape          // how get finds the item in its response; oneItem when nil
	roots     shape          // the items of list's response that get takes, for completion; items when nil
	search    *searchListing // the search that lists a collection that Emporix lists only by search; list when nil
}

// newResourceCommand returns the group of r: list, get ID and the extra
// commands, reads or writes.
func newResourceCommand[T any](a *app, r resource[T], extra ...*cobra.Command) *cobra.Command {
	commands := append([]*cobra.Command{r.listCommand(a), r.getCommand(a)}, extra...)
	return newGroup(r.use, r.summary(), r.groupHelp(), commands...)
}

// summary returns the summary of r's group: short, else "Read " and the nouns.
func (r resource[T]) summary() string { return cmp.Or(r.short, "Read "+r.nouns()) }

// groupHelp returns the help of r's group: about, else long.
func (r resource[T]) groupHelp() string { return cmp.Or(r.about, r.long) }

// reader returns the reader of r's reads.
func (r resource[T]) reader() reader[T] { return reader[T]{view: r.view, scopes: r.scopes} }

// nouns returns the items of r as help text names them, e.g. "price lists".
func (r resource[T]) nouns() string { return cmp.Or(r.noun, strings.ReplaceAll(r.use, "-", " ")) }

// listRequest returns the request of list: the collection with query and listQuery.
func (r resource[T]) listRequest() request { return withQueries(r.path, r.query, r.listQuery) }

// ids returns the source of the IDs of r's items for completion, those of
// its first column (idsIn).
func (r resource[T]) ids() *idSource { return r.idsIn(r.header[0]) }

// idsIn returns the source of the values that column of r's table shows, for
// completion: the first page of its list, or none for a collection that
// Emporix lists only by search, that another command lists at another path
// (listedBy without atPath), or whose list needs a site where a read of one
// item takes none, as for prices: no command that names an item of r would
// give it.
func (r resource[T]) idsIn(column string) *idSource {
	list := r.listRequest()
	elsewhere := r.listedBy != "" && !r.atPath
	if r.search != nil || elsewhere || list.names(placeholderSite) && !r.itemRequest().names(placeholderSite) {
		return nil
	}
	listed := r.listing()
	if r.roots != nil {
		listed.items = r.roots
	}
	return listedIDs(r.reader(), list, listed, column)
}

// itemIDs returns the source of the ID of a command of r's item, whose use
// is itemUse: r's IDs (ids).
func (r resource[T]) itemIDs() argumentIDs { return argumentIDs{r.header[0]: r.ids()} }

// listUse is the name of the command that lists a collection.
const listUse = "list"

// getUse is the name of the command that reads one item of a collection.
const getUse = "get"

// listCommand returns `list`, which lists the collection.
func (r resource[T]) listCommand(a *app) *cobra.Command {
	req := r.listRequest()
	short := "List " + r.nouns()
	if req.names(placeholderSite) {
		short = "List the " + r.nouns() + " of a site"
	}
	return r.reader().command(a, &command{use: listUse, short: short, long: r.long, req: req, fetch: r.listing()})
}

// listing returns how list reads the collection, which a search of it reads
// alike (listing.search).
func (r resource[T]) listing() listing {
	return listing{filters: r.filters, read: r.read, items: r.items}
}

// getCommand returns `get ID`, which reads the item ID and finds it in the
// response with r.item.
func (r resource[T]) getCommand(a *app) *cobra.Command {
	return r.reader().command(a, &command{
		use: r.itemUse(getUse), short: "Show one " + r.singular, long: r.long,
		req: r.itemRequest(), fetch: single{items: itemShape(r.item)}, ids: r.itemIDs(),
	})
}

// itemUse returns the use of a command of r that takes the ID of one item:
// verb, then the name of the identifier, its first column, e.g. "get CODE".
func (r resource[T]) itemUse(verb string) string { return verb + " " + r.header[0] }

// itemPath returns the path of one item of r, the ID of a command's itemUse.
func (r resource[T]) itemPath() string { return r.path + "/" + placeholderID }

// itemRequest returns the request that reads one item of r, as get reads it.
func (r resource[T]) itemRequest() request { return request{path: r.itemPath(), query: r.query} }

// help returns long followed by the legend of rd's view and the scope sentence of rd.
func (rd reader[T]) help(long string) string {
	if rd.legend != "" {
		long += "\n" + rd.legend
	}
	return long + "\n\n" + scopeSentence(rd.scopes)
}

// readTarget is what a run of a read command requests, as newRead binds it,
// and with which scopes.
type readTarget struct {
	a         *app
	req       request
	site      *siteCode                    // --site, where req names {site}
	scopesFor func(args []string) []string // the scopes of a run with args (typeScopes)
	optedIn   func() error                 // fails a run that requests a write-capable scope without opting in
}

// newRead returns the read command c of scopes, without how it reads and
// what it does with the answer, and how a run opens the read (open): the one
// place that builds read commands, for reader.command, which prints what it
// reads, and reader.fileCommand, which saves a file. The command notes its
// scopes (scopesAnnotation), takes an ID argument where c.req names {id},
// preceded by the arguments of {parent} and {inner} where it names those
// (each one of its choices, if any), which complete as c.ids lists them,
// requires --site where it names {site}, and --allow-write-scope where one
// of scopes is write-capable (requireOptIn).
func newRead(a *app, c *command, scopes []string) (*cobra.Command, *readTarget) {
	cmd := &cobra.Command{
		Use: c.use, Short: c.short, Long: c.long,
		Annotations: map[string]string{scopesAnnotation: strings.Join(scopes, " ")},
	}
	t := &readTarget{a: a, req: c.req, site: c.req.bind(cmd, c.choices)}
	fillType := typeScopes(cmd, c.req, scopes) // after bind, whose argument rule it wraps
	t.scopesFor = func(args []string) []string { return fillType(args, scopes) }
	completeArguments(a, cmd, t.site, c.ids, c.choices, c.predefined)
	t.optedIn = requireOptIn(cmd, scopes)
	return cmd, t
}

// open fails a run with args that requests a write-capable scope without
// opting in, before it requests a token, and otherwise returns a client
// limited to the read's scopes, those of the custom entity type that its
// first argument names where they name one (typeScopes), and the path and
// query that it requests.
func (t *readTarget) open(ctx context.Context, args []string) (*emporix.Client, string, url.Values, error) {
	if err := t.optedIn(); err != nil {
		return nil, "", nil, err
	}
	client, err := t.a.client(ctx, t.scopesFor(args)...)
	if err != nil {
		return nil, "", nil, err
	}
	path, query := t.req.fill(args, *t.site)
	return client, path, query, nil
}

// command returns the read command c of rd's scopes (newRead), which reads
// as c.fetch says and prints the items it finds in the response as the rows
// of a table or CSV, or the response whole as JSON or YAML (printer).
func (rd reader[T]) command(a *app, c *command) *cobra.Command {
	// The request's arguments before fetch's bind, whose body may repeat the site
	// (siteField), and the output flags too, as it may mark flags that exclude them.
	cmd, target := newRead(a, c, rd.scopes)
	printed := addTableOutput(cmd, rd.header, rd.long...)
	read, items := c.fetch.bind(cmd)
	cmd.Long = rd.help(cmd.Long) // after bind, which may add to the help that the scope sentence ends
	lang := defaultLanguage
	cmd.RunE = func(cc *cobra.Command, args []string) error {
		client, path, query, err := target.open(cc.Context(), args)
		if err != nil {
			return err
		}
		body, err := read(client, cc.Context(), path, query)
		if errors.Is(err, errStreamed) {
			return nil
		}
		if err != nil {
			return err
		}
		found, err := items(body)
		if err != nil {
			return rd.decodeError(err)
		}
		return printed.print(cc.OutOrStdout(), body, func() ([][]string, error) { return rd.rows(found, lang) })
	}
	if rd.localized {
		addLanguageFlag(cmd, &lang)
	}
	return cmd
}

// rows returns the table rows of the items; an item without an identifier
// (first column) is a malformed response.
func (v view[T]) rows(items []json.RawMessage, lang language) ([][]string, error) {
	rows := make([][]string, len(items))
	for i, raw := range items {
		var item T
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, v.decodeError(err)
		}
		if rows[i] = v.row(item, lang); rows[i][0] == "" {
			return nil, v.decodeError(fmt.Errorf("item %d has no identifier", i+1))
		}
	}
	return rows, nil
}

func (v view[T]) decodeError(err error) error {
	return fmt.Errorf("decoding %s response: %w", v.singular, err)
}

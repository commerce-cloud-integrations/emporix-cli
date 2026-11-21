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

type view[T any] struct {
	singular  string
	header    []string
	long      []string
	localized bool
	siteAware bool
	row       func(T, language) []string
	legend    string
}

type reader[T any] struct {
	view[T]
	scopes []string
}

func typeNames(singular string) reader[string] {
	return reader[string]{view: stringsView(singular, columnType)}
}

func stringsView(singular, column string) view[string] {
	return view[string]{
		singular: singular,
		header:   []string{column},
		row:      func(s string, _ language) []string { return []string{s} },
	}
}

type command struct {
	use     string
	short   string
	long    string
	req     request
	fetch   fetcher
	choices map[string][]string
	ids     argumentIDs
	builtIn builtInTypes
}

type fetcher interface {
	bind(cmd *cobra.Command) (readFunc, shape)
}

type listing struct {
	filters listFilters
	read    collectionRead
	items   shape
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
		run, out := opts, cmd.OutOrStdout()
		run.Stream = func(items []json.RawMessage) error { return output.WriteJSONLines(out, items) }
		if _, err := read(c, ctx, path, query, run); err != nil {
			return nil, err
		}
		return nil, errStreamed
	}
	return listed, l.found()
}

func (l listing) glance() (readFunc, shape) {
	read := l.collection()
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		return read(c, ctx, path, query, emporix.ListOptions{FirstPage: true})
	}, l.found()
}

func (l listing) collection() collectionRead {
	if l.read == nil {
		return (*emporix.Client).GetAll
	}
	return l.read
}

func (l listing) found() shape {
	if l.items == nil {
		return listItems
	}
	return l.items
}

var errStreamed = errors.New("the items were streamed")

type single struct {
	items shape
	read  readFunc
}

func (s single) bind(*cobra.Command) (readFunc, shape) { return s.glance() }

func (s single) glance() (readFunc, shape) {
	if s.read == nil {
		return (*emporix.Client).Get, s.items
	}
	return s.read, s.items
}

func emptyAs(answer string) readFunc {
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		body, err := c.Check(ctx, path, query)
		if err != nil || body != nil {
			return body, err
		}
		return json.RawMessage(answer), nil
	}
}

type readFunc func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error)

type collectionRead func(c *emporix.Client, ctx context.Context, path string, query url.Values,
	opts emporix.ListOptions) (json.RawMessage, error)

func inOnePage(size int) collectionRead { return pageOf((*emporix.Client).GetAllInOnePage, size) }

func onFirstPage(size int) collectionRead { return pageOf((*emporix.Client).GetAllOnFirstPage, size) }

func pageOf(read func(*emporix.Client, context.Context, string, url.Values, int, emporix.ListOptions) (
	json.RawMessage, error), size int) collectionRead {
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values,
		opts emporix.ListOptions) (json.RawMessage, error) {
		return read(c, ctx, path, query, size, opts)
	}
}

func onePage(size int) string { return "one page of up to " + strconv.Itoa(size) }

type resource[T any] struct {
	view[T]
	use        string
	noun       string
	short      string
	long       string
	listNote   string
	about      string
	under      string
	listedBy   string
	atPath     bool
	path       string
	query      url.Values
	listQuery  url.Values
	listParams []paramFlag
	scopes     []string
	filters    listFilters
	read       collectionRead
	items      shape
	item       shape
	roots      shape
	search     *searchListing
}

func newResourceCommand[T any](a *app, r resource[T], extra ...*cobra.Command) *cobra.Command {
	commands := append([]*cobra.Command{r.listCommand(a), r.getCommand(a)}, extra...)
	return newGroup(r.use, r.summary(), r.groupHelp(), commands...)
}

func (r resource[T]) summary() string { return cmp.Or(r.short, "Read "+r.nouns()) }

func (r resource[T]) groupHelp() string { return cmp.Or(r.about, r.long) }

func (r resource[T]) reader() reader[T] { return reader[T]{view: r.view, scopes: r.scopes} }

func (r resource[T]) nouns() string { return cmp.Or(r.noun, strings.ReplaceAll(r.use, "-", " ")) }

func (r resource[T]) listRequest() request { return withQueries(r.path, r.query, r.listQuery) }

func (r resource[T]) ids() *idSource { return r.idsIn(r.header[0]) }

func (r resource[T]) idsIn(column string) *idSource {
	list := r.listRequest()
	if r.search != nil || !r.listsAtPath() || list.names(placeholderSite) && !r.itemRequest().names(placeholderSite) {
		return nil
	}
	listed := r.listing()
	if r.roots != nil {
		listed.items = r.roots
	}
	return listedIDs(r.reader(), list, listed, column)
}

func (r resource[T]) itemIDs() argumentIDs { return argumentIDs{r.header[0]: r.ids()} }

const listUse = "list"

const getUse = "get"

func (r resource[T]) listCommand(a *app) *cobra.Command {
	req := r.listRequest()
	short := "List " + r.nouns()
	if req.names(placeholderSite) {
		short = "List the " + r.nouns() + " of a site"
	}
	return r.reader().command(a, &command{use: listUse, short: short, long: joinSentences(r.long, r.listNote), req: req,
		fetch: &paramRead{params: queryFlags{optional: r.listParams}, fetch: r.listing()}})
}

func (r resource[T]) listing() listing {
	return listing{filters: r.filters, read: r.read, items: r.items}
}

func (r resource[T]) getCommand(a *app) *cobra.Command {
	return r.reader().command(a, &command{
		use: r.itemUse(getUse), short: "Show one " + r.singular, long: r.long,
		req: r.itemRequest(), fetch: single{items: itemShape(r.item)}, ids: r.itemIDs(),
	})
}

func (r resource[T]) itemUse(verb string) string { return verb + " " + r.header[0] }

func (r resource[T]) itemPath() string { return r.path + "/" + placeholderID }

func (r resource[T]) itemRequest() request { return request{path: r.itemPath(), query: r.query} }

func (r resource[T]) getLine() string { return r.commandLine(r.itemUse(getUse), r.itemRequest()) }

func (rd reader[T]) help(long string, rule typeRule) string {
	long = joinSentences(long, rd.legend)
	if rd.siteAware {
		long += "\n\n" + mixinSiteHelp
	}
	return rule.helpEnding(long, scopeSentence(rd.scopes))
}

type readTarget struct {
	a       *app
	req     request
	site    *siteCode
	rule    typeRule
	scopes  []string
	optedIn func() error
}

func newRead(a *app, c *command, scopes []string) (*cobra.Command, *readTarget) {
	cmd := &cobra.Command{Use: c.use, Short: c.short, Long: c.long,
		Annotations: map[string]string{scopesAnnotation: strings.Join(scopes, " ")}}
	t := &readTarget{a: a, req: c.req, site: c.req.bind(cmd, c.choices), scopes: scopes}
	t.rule = typeScopes(cmd, c.req, scopes, c.builtIn)
	completeArguments(a, cmd, t.site, c.ids, c.choices, t.rule)
	t.optedIn = requireOptIn(cmd, scopes)
	return cmd, t
}

func (t *readTarget) open(ctx context.Context, args []string) (*emporix.Client, string, url.Values, error) {
	if err := t.optedIn(); err != nil {
		return nil, "", nil, err
	}
	client, err := t.a.client(ctx, t.rule.fill(args, t.scopes)...)
	if err != nil {
		return nil, "", nil, err
	}
	path, query := t.req.fill(args, *t.site)
	return client, path, query, nil
}

func (rd reader[T]) command(a *app, c *command) *cobra.Command {
	cmd, target := newRead(a, c, rd.scopes)
	printed := addTableOutput(cmd, rd.header, rd.long...)
	read, items := rd.withMixinSite(c.fetch).bind(cmd)
	cmd.Long = rd.help(cmd.Long, target.rule)
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

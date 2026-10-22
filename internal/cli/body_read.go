package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// bodyRead reads with a request that Emporix serves as POST, or PUT, with a
// body but that changes no data, such as a search, a calculation or the
// parsing of a file (emporix.Client.WithBody). Its flags make the body
// (readBinder): a JSON one, which typed flags make (bodyBinder) or --body FILE
// gives whole, or a form; the read then goes on as fetch reads: a listing,
// paged as Emporix pages most searches, or a single response.
type bodyRead struct {
	method  string     // http.MethodPost or http.MethodPut
	created bool       // Emporix answers 201 Created although nothing is created (emporix.Client.AcceptCreated)
	body    readBinder // the body that its flags make; an empty JSON object when nil
	fetch   fetcher    // how the response is read: a listing or single
	// forbiddenIsRefusal says that Emporix refuses an input that it will not
	// act on with 403 Forbidden although the credential is fine: such a
	// refusal then exits 1, not 3 (emporix.Client.ForbiddenIsRefusal).
	forbiddenIsRefusal bool
}

// changesNothing says in the help of a read that Emporix serves as method,
// POST or PUT, that it changes nothing, and what it sends, as sends says.
func changesNothing(method, sends string) string {
	return "Changes nothing: Emporix serves this read as " + method + " with " + sends
}

// bodyFlag is the flag of every read with a JSON body that gives the body
// whole.
const bodyFlag = "body"

// maxBodyFileBytes bounds the body --body reads, far above any search's.
const maxBodyFileBytes = 1 << 20

func (b bodyRead) bind(cmd *cobra.Command) (readFunc, shape) {
	var body readBinder = objectBody{}
	if b.body != nil {
		body = b.body
	}
	if b.forbiddenIsRefusal {
		checkRefusalHelp(cmd)
	}
	made, sends := body.bindRead(cmd)
	cmd.Long += "\n\n" + changesNothing(b.method, sends)
	read, items := b.fetch.bind(cmd)
	sent := func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		body, err := made(ctx)
		if err != nil {
			return nil, err
		}
		sending, err := c.WithBody(b.method, body)
		if err != nil {
			return nil, err
		}
		if b.created {
			sending = sending.AcceptCreated()
		}
		if b.forbiddenIsRefusal {
			sending = sending.ForbiddenIsRefusal()
		}
		return read(sending, ctx, path, query)
	}
	return sent, items
}

// readBinder is the body of a read with a body, which its flags make: a JSON
// one (jsonRead), which typed flags make (objectBody, valueBody) or --body
// gives whole, or a form (form).
type readBinder interface {
	// bindRead adds the flags of the body to cmd and returns how a run makes
	// it, and what the read sends, as its help names it after "with".
	bindRead(cmd *cobra.Command) (made makeBody, sends string)
}

// makeBody makes the body of a request in a run, once flags are parsed.
type makeBody func(ctx context.Context) (emporix.Body, error)

// jsonRead adds to cmd --body, which gives the JSON body of a read whole,
// then the flags of typed, each of which excludes --body, and returns how a
// run makes the body and what the read sends (readBinder).
func jsonRead(cmd *cobra.Command, typed bodyBinder) (made makeBody, sends string) {
	body := bodyFile{flag: bodyFlag, limit: maxBodyFileBytes}.bindWith(cmd, "send the JSON request body in `FILE` as "+
		"given (- reads standard input) instead of the one the other flags make", typed)
	return body.make, "a JSON body, which the\nflags make or --body gives whole."
}

// bodyBinder is the body that the typed flags of a read or a write make: a
// JSON object of members (objectBody) or one value, such as an array of IDs
// (valueBody).
type bodyBinder interface {
	// bind adds the flags of the body to cmd, which defines whole, the flag
	// that gives the body whole instead (--body of a read, --file of a
	// write), and returns the body they make once flags are parsed.
	bind(cmd *cobra.Command, whole string) func() any
}

// objectBody is a body that is a JSON object, each member of which a flag
// sets, leaving out a member whose value is nil (an optional flag not given);
// without members, an empty object. A member named with dots, such as
// input.price, is a member of nested objects. Members must not overlap
// (checkMembers).
type objectBody []bodyField

func (o objectBody) bindRead(cmd *cobra.Command) (made makeBody, sends string) {
	return jsonRead(cmd, o)
}

func (o objectBody) bind(cmd *cobra.Command, whole string) func() any {
	checkMembers(o)
	members := make(map[string]func() any, len(o))
	for _, f := range o {
		members[f.member] = f.value.add(cmd, whole)
	}
	return func() any {
		object := make(map[string]any, len(members))
		for member, value := range members {
			if v := value(); v != nil {
				setMember(object, strings.Split(member, "."), v)
			}
		}
		return object
	}
}

// setMember sets the member that path names in object: the members of nested
// objects, made where missing, then the member itself.
func setMember(object map[string]any, path []string, value any) {
	for _, name := range path[:len(path)-1] {
		nested, ok := object[name].(map[string]any)
		if !ok {
			nested = map[string]any{}
			object[name] = nested
		}
		object = nested
	}
	object[path[len(path)-1]] = value
}

// checkMembers panics, as Cobra does for a flag group that names an unknown
// flag, when two fields set one member or one member holds another's (input
// and input.price): the body would then depend on the order of a map. Every
// test that builds the command tree finds it.
func checkMembers(fields []bodyField) {
	for i, f := range fields {
		for _, other := range fields[:i] {
			if f.member == other.member || strings.HasPrefix(f.member, other.member+".") ||
				strings.HasPrefix(other.member, f.member+".") {
				panic(fmt.Sprintf("body members %q and %q overlap", other.member, f.member))
			}
		}
	}
}

// valueBody is a body that is the one value a flag sets, such as a JSON array
// of IDs.
type valueBody bodyValue

func (v valueBody) bind(cmd *cobra.Command, whole string) func() any { return v.add(cmd, whole) }

func (v valueBody) bindRead(cmd *cobra.Command) (made makeBody, sends string) {
	return jsonRead(cmd, v)
}

// bodyField is a member of an objectBody that a flag sets.
type bodyField struct {
	member string    // the JSON member, e.g. "q"
	value  bodyValue // its value and the flag that sets it
}

// bodyValue is a value of a body that a flag sets: the value's own
// (flagValue), which the flag that gives the body whole excludes, or the
// --site of the read (siteField). Its value once flags are parsed is nil for
// a member left out of the body.
type bodyValue struct {
	flag string // the flag that sets it, e.g. "query"
	// add adds its own flag, if any, to cmd, whose flag whole gives the body
	// whole (bodyBinder.bind); it returns the value once flags are parsed.
	add func(cmd *cobra.Command, whole string) func() any
}

// flagValue is a value of a body set by its own flag, which add adds to the
// flags and the flag that gives the body whole excludes.
func flagValue(flag string, add func(fs *pflag.FlagSet) func() any) bodyValue {
	return commandValue(flag, func(cmd *cobra.Command) func() any { return add(cmd.Flags()) })
}

// commandValue is a flagValue whose flag add adds to cmd itself, such as one
// whose values shells complete.
func commandValue(flag string, add func(cmd *cobra.Command) func() any) bodyValue {
	return bodyValue{flag: flag, add: func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd)
		cmd.MarkFlagsMutuallyExclusive(whole, flag)
		return value
	}}
}

// flagField is a member of a body set by its own flag (flagValue).
func flagField(member, flag string, add func(fs *pflag.FlagSet) func() any) bodyField {
	return bodyField{member: member, value: flagValue(flag, add)}
}

// require returns v with its flag required unless the flag that gives the
// body whole gives it, for a value that Emporix requires: without either, the
// command fails as its flags are parsed (exit 2), and nothing is sent. The
// flag's usage says so, in its closing parenthesis if it ends with one.
func (v bodyValue) require() bodyValue {
	add := v.add
	v.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		requireFlagUnless(cmd, v.flag, whole)
		return value
	}
	return v
}

// optional returns v, left out of the body unless its flag is given, for a
// value that is never null, such as an array (objectsValue).
func (v bodyValue) optional() bodyValue {
	add := v.add
	v.add = func(cmd *cobra.Command, whole string) func() any {
		return unlessGiven(cmd.Flags(), v.flag, add(cmd, whole))
	}
	return v
}

// require returns f with its flag required unless the flag that gives the
// body whole gives it (bodyValue.require).
func (f bodyField) require() bodyField {
	f.value = f.value.require()
	return f
}

// stringField is a member of a body that is a string, set by --flag as given;
// left out unless the flag is given.
func stringField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		value := fs.String(flag, "", usage)
		return unlessGiven(fs, flag, func() any { return *value })
	})
}

// unlessGiven returns value, or, where --flag of fs is not given, a value of
// nil, which leaves its member out of the body.
func unlessGiven(fs *pflag.FlagSet, flag string, value func() any) func() any {
	return func() any {
		if !fs.Changed(flag) {
			return nil
		}
		return value()
	}
}

// queryField is the member q of a search, the Emporix q filter, set by
// --query; an empty string, which filters nothing, unless the flag is given.
func queryField() bodyField {
	return textField("q", "query", queryUsage("sent as given in the request body, not in the URL (empty: no filter)"))
}

// textField is a member of a search that is a string set by --flag, sent as
// given; an empty string unless the flag is given.
func textField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		text := fs.String(flag, "", usage)
		return func() any { return *text }
	})
}

// listField is a member of a search that is a listValue.
func listField(member, flag, usage string) bodyField {
	return bodyField{member: member, value: listValue(flag, usage)}
}

// listValue is an array of strings, each set by one --flag, in their order;
// an empty array unless the flag is given.
func listValue(flag, usage string) bodyValue {
	return flagValue(flag, func(fs *pflag.FlagSet) func() any {
		values := fs.StringArray(flag, []string{}, usage) // [] in the body, never null; the first value replaces it
		return func() any { return *values }
	})
}

// siteCodeField is a member of a search that is a site, set by --site and
// checked as the --site of a request is (siteCode), which takes the default
// site of the profile in effect (siteDefault); an empty string unless the
// flag is given.
func siteCodeField(member, usage string) bodyField {
	return flagField(member, siteFlag, func(fs *pflag.FlagSet) func() any {
		var site siteCode
		fs.Var(&site, siteFlag, usage)
		return func() any { return site }
	}).siteDefault()
}

// siteCodesField is a member of a search that is an array of sites, each set
// by one --site and checked as the --site of a request is (siteCodes), which
// takes the default site of the profile in effect (siteDefault), in their
// order; an empty array unless the flag is given.
func siteCodesField(member, usage string) bodyField {
	return flagField(member, siteFlag, func(fs *pflag.FlagSet) func() any {
		sites := siteCodes{} // [] in the body, never null
		fs.Var(&sites, siteFlag, usage)
		return func() any { return sites }
	}).siteDefault()
}

// siteDefault returns f, whose flag names a site, with that flag taking the
// default site of the profile in effect (takeProfileDefault), unless the
// flag that gives the body whole is given: a body given whole keeps its own.
func (f bodyField) siteDefault() bodyField {
	add := f.value.add
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		takeProfileDefault(cmd, f.value.flag, siteFlag, whole)
		return value
	}
	return f
}

// siteField is the member of a search that repeats the site of its request,
// whose query names {site} too: the value of the --site that the request
// requires, which reader.command adds before the fields. A body given with
// --body holds the member as given. The command must name {site}: siteField
// panics as the command is built otherwise, as Cobra does for a flag group
// of an undefined flag.
func siteField(member string) bodyField {
	return bodyField{member: member, value: bodyValue{flag: siteFlag, add: func(cmd *cobra.Command, _ string) func() any {
		site := cmd.Flags().Lookup(siteFlag)
		if site == nil {
			panic(fmt.Sprintf("%s: the body member %q repeats a site that its request does not name", cmd.Use, member))
		}
		return func() any { return site.Value.String() }
	}}}
}

// objectsValue is an array of JSON objects, one per --flag value, in their
// order, each the T that parse makes of it; an empty array unless the flag is
// given. A value that parse rejects is rejected while flags are parsed.
func objectsValue[T any](flag, usage string, parse func(value string) (T, error)) bodyValue {
	return flagValue(flag, func(fs *pflag.FlagSet) func() any {
		list := &objectList[T]{objects: []T{}, parse: parse} // [] in the body, never null
		fs.Var(list, flag, usage)
		return func() any { return list.objects }
	})
}

// objectList is the value of the repeatable flag of an objectsValue; it
// implements pflag.Value.
type objectList[T any] struct {
	given   []string
	objects []T
	parse   func(value string) (T, error)
}

// Set implements pflag.Value, adding the object of one value.
func (l *objectList[T]) Set(value string) error {
	object, err := l.parse(value)
	if err != nil {
		return err // the flag parser names the flag and the value
	}
	l.given, l.objects = append(l.given, value), append(l.objects, object)
	return nil
}

// String implements pflag.Value: the values given.
func (l *objectList[T]) String() string { return strings.Join(l.given, ",") }

// GetSlice returns the values given; it lets completion offer the repeatable
// flag again (cobra.SliceValue).
func (l *objectList[T]) GetSlice() []string { return l.given }

// Type implements pflag.Value.
func (l *objectList[T]) Type() string { return "value" }

// collectionSearch names the command that searches a collection, and the
// search's path below the collection.
const collectionSearch = "search"

// search returns how a search reads the collection that l lists: a POST
// with the body that fields make (or --body), read as l reads it, with l's
// flags but --query: a search filters with its body (queryField).
func (l listing) search(fields ...bodyField) bodyRead {
	l.filters = l.filters.without(queryable)
	return bodyRead{method: http.MethodPost, body: objectBody(fields), fetch: l}
}

// searchSentence says that a search reads and shows what it finds as the
// command line list (without epx) reads and shows its items.
func searchSentence(list string) string {
	return "It reads and shows them as `epx " + list + "` does."
}

// searchCommand returns `search`, which lists the items of r that a search
// finds: a POST to the collection's search path with the query of list and
// the body that fields make (or --body), read and shown as list reads and
// shows the collection (listing.search), as its help says (searchSentence).
func (r resource[T]) searchCommand(a *app, long string, fields ...bodyField) *cobra.Command {
	return r.searchRead(a, long+"\n"+searchSentence(r.listLine()), fields)
}

// searchRead returns `search` with the help long: a POST to the collection's
// search path with the query of list and the body that fields make (or
// --body), read as list reads the collection (listing.search).
func (r resource[T]) searchRead(a *app, long string, fields []bodyField) *cobra.Command {
	req := r.listRequest()
	req.path += "/" + collectionSearch
	return r.reader().command(a, &command{
		use: collectionSearch, short: "Search " + r.nouns(), long: long, req: req, fetch: r.listing().search(fields...),
	})
}

// querySearchCommand returns the search of r that sends the Emporix q filter
// of --query in its body (queryField), with a help generated from r's nouns.
func (r resource[T]) querySearchCommand(a *app) *cobra.Command {
	return r.searchCommand(a, "Reads the "+r.nouns()+" that an Emporix q filter finds (--query; without it,\n"+
		"every one).", queryField())
}

// searchListing declares the search of a collection that Emporix lists only
// by search, such as carts: its group has search where others have list
// (newSearchedResourceCommand), and the reads below one of its items name
// search as the source of their IDs (listLine).
type searchListing struct {
	long   string      // what search reads; the body and scope sentences follow
	fields []bodyField // the members of the body that typed flags make
}

// newSearchedResourceCommand returns the group of r, a collection that
// Emporix lists only by search (r.search): search, get ID, whose help says
// that search shows the IDs as no list is beside it, and the extra reads.
func newSearchedResourceCommand[T any](a *app, r resource[T], extra ...*cobra.Command) *cobra.Command {
	get := r
	get.long += "\n" + itemSentence(r.header[:1], r.header[:1], r.listLine())
	reads := append([]*cobra.Command{r.searchRead(a, r.search.long, r.search.fields), get.getCommand(a)}, extra...)
	return newGroup(r.use, r.summary(), r.groupHelp(), reads...)
}

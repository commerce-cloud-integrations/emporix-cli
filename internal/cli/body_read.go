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

type bodyRead struct {
	method             string
	created            bool
	body               readBinder
	fetch              fetcher
	forbiddenIsRefusal bool
}

func changesNothing(method, sends string) string {
	return "Changes nothing: Emporix serves this read as " + method + " with " + sends
}

const bodyFlag = "body"

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

type readBinder interface {
	bindRead(cmd *cobra.Command) (made makeBody, sends string)
}

type makeBody func(ctx context.Context) (emporix.Body, error)

func jsonRead(cmd *cobra.Command, typed bodyBinder) (made makeBody, sends string) {
	body := bodyFile{flag: bodyFlag, limit: maxBodyFileBytes}.bindWith(cmd, "send the JSON request body in `FILE` as "+
		"given (- reads standard input) instead of the one the other flags make", typed)
	return body.make, "a JSON body, which the flags make or --body gives whole."
}

type bodyBinder interface {
	bind(cmd *cobra.Command, whole string) func() any
}

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

type valueBody bodyValue

func (v valueBody) bind(cmd *cobra.Command, whole string) func() any { return v.add(cmd, whole) }

func (v valueBody) bindRead(cmd *cobra.Command) (made makeBody, sends string) {
	return jsonRead(cmd, v)
}

type bodyField struct {
	member string
	value  bodyValue
}

type bodyValue struct {
	flag string
	add  func(cmd *cobra.Command, whole string) func() any
}

func flagValue(flag string, add func(fs *pflag.FlagSet) func() any) bodyValue {
	return commandValue(flag, func(cmd *cobra.Command) func() any { return add(cmd.Flags()) })
}

func commandValue(flag string, add func(cmd *cobra.Command) func() any) bodyValue {
	return bodyValue{flag: flag, add: func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd)
		cmd.MarkFlagsMutuallyExclusive(whole, flag)
		return value
	}}
}

func flagField(member, flag string, add func(fs *pflag.FlagSet) func() any) bodyField {
	return bodyField{member: member, value: flagValue(flag, add)}
}

func (v bodyValue) require() bodyValue {
	add := v.add
	v.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		requireFlagUnless(cmd, v.flag, whole)
		return value
	}
	return v
}

func (v bodyValue) optional() bodyValue {
	add := v.add
	v.add = func(cmd *cobra.Command, whole string) func() any {
		return unlessGiven(cmd.Flags(), v.flag, add(cmd, whole))
	}
	return v
}

func (f bodyField) require() bodyField {
	f.value = f.value.require()
	return f
}

func stringField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		value := fs.String(flag, "", usage)
		return unlessGiven(fs, flag, func() any { return *value })
	})
}

func unlessGiven(fs *pflag.FlagSet, flag string, value func() any) func() any {
	return func() any {
		if !fs.Changed(flag) {
			return nil
		}
		return value()
	}
}

func queryField() bodyField {
	return textField("q", "query", queryUsage("sent as given in the request body, not in the URL (empty: no filter)"))
}

func textField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		text := fs.String(flag, "", usage)
		return func() any { return *text }
	})
}

func listField(member, flag, usage string) bodyField {
	return bodyField{member: member, value: listValue(flag, usage)}
}

func listValue(flag, usage string) bodyValue {
	return flagValue(flag, func(fs *pflag.FlagSet) func() any {
		values := fs.StringArray(flag, []string{}, usage)
		return func() any { return *values }
	})
}

func siteCodeField(member, usage string) bodyField {
	return flagField(member, siteFlag, func(fs *pflag.FlagSet) func() any {
		var site siteCode
		fs.Var(&site, siteFlag, usage)
		return func() any { return site }
	}).siteDefault()
}

func siteCodesField(member, usage string) bodyField {
	return flagField(member, siteFlag, func(fs *pflag.FlagSet) func() any {
		sites := siteCodes{}
		fs.Var(&sites, siteFlag, usage)
		return func() any { return sites }
	}).siteDefault()
}

func (f bodyField) siteDefault() bodyField {
	add := f.value.add
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		takeProfileDefault(cmd, f.value.flag, siteFlag, whole)
		return value
	}
	return f
}

func siteField(member string) bodyField {
	return bodyField{member: member, value: bodyValue{flag: siteFlag, add: func(cmd *cobra.Command, _ string) func() any {
		site := cmd.Flags().Lookup(siteFlag)
		if site == nil {
			panic(fmt.Sprintf("%s: the body member %q repeats a site that its request does not name", cmd.Use, member))
		}
		return func() any { return site.Value.String() }
	}}}
}

func objectsValue[T any](flag, usage string, parse func(value string) (T, error)) bodyValue {
	return flagValue(flag, func(fs *pflag.FlagSet) func() any {
		list := &objectList[T]{objects: []T{}, parse: parse}
		fs.Var(list, flag, usage)
		return func() any { return list.objects }
	})
}

type objectList[T any] struct {
	given   []string
	objects []T
	parse   func(value string) (T, error)
}

// Set implements pflag.Value, adding the object of one value.
func (l *objectList[T]) Set(value string) error {
	object, err := l.parse(value)
	if err != nil {
		return err
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

const collectionSearch = "search"

func (l listing) search(fields ...bodyField) bodyRead {
	l.filters = l.filters.without(queryable)
	return bodyRead{method: http.MethodPost, body: objectBody(fields), fetch: l}
}

func searchSentence(list string) string {
	return "It reads and shows them as `epx " + list + "` does."
}

func (r resource[T]) searchCommand(a *app, long string, fields ...bodyField) *cobra.Command {
	return r.searchRead(a, long+" "+searchSentence(r.listLine()), fields)
}

func (r resource[T]) searchRead(a *app, long string, fields []bodyField) *cobra.Command {
	req := r.listRequest()
	req.path += "/" + collectionSearch
	return r.reader().command(a, &command{
		use: collectionSearch, short: "Search " + r.nouns(), long: long, req: req, fetch: r.listing().search(fields...),
	})
}

func (r resource[T]) querySearchCommand(a *app) *cobra.Command {
	return r.searchCommand(a, "Reads the "+r.nouns()+" that an Emporix q filter finds (--query; without it, "+
		"every one).", queryField())
}

type searchListing struct {
	long   string
	fields []bodyField
}

func newSearchedResourceCommand[T any](a *app, r resource[T], extra ...*cobra.Command) *cobra.Command {
	get := r
	get.long = joinSentences(get.long, itemSentence(r.header[:1], r.header[:1], r.listLine()))
	reads := append([]*cobra.Command{r.searchRead(a, r.search.long, r.search.fields), get.getCommand(a)}, extra...)
	return newGroup(r.use, r.summary(), r.groupHelp(), reads...)
}

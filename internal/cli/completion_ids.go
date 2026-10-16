package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// envCompleteIDs is the environment variable that lets shell completion read
// the API: set to 1 (or true), the arguments that name items and --site
// complete with the IDs and site codes of the tenant (listedItems).
const envCompleteIDs = "EPX_COMPLETE_IDS"

// defaultCompletionTimeout bounds what one completion reads, its token
// included, as epx runs (app.completionTimeout): a shell waits for it.
const defaultCompletionTimeout = 3 * time.Second

// completionSize is the most IDs one completion offers.
const completionSize = 100

// completionHelp follows the help of `epx completion`: how to opt in to
// completion that reads the API.
func completionHelp() string {
	return "With " + envCompleteIDs + "=1 in the environment, the arguments that name items (such as ID of\n" +
		"`epx products get ID`) and --site complete too, with up to " + strconv.Itoa(completionSize) +
		" of the IDs and site codes that\n" +
		"the first page of their list shows (the whole list where Emporix sends it in one response),\n" +
		"read with the API key in effect and the scopes of that list within " +
		defaultCompletionTimeout.String() + ".\n" +
		"Without it, completion never calls the API; when the API cannot be read, it offers no IDs\n" +
		"and prints nothing."
}

// glancer is a fetch that reads without flags of its own, a listing or a
// single response, so that shell completion can read it too: glance returns
// how it reads one page at most and where its items are in the response.
type glancer interface {
	glance() (readFunc, shape)
}

// argumentIDs are the reads that list the IDs that the arguments of a
// command take, for shell completion, by the argument's name in its use,
// e.g. PRICE_LIST_ID; a nil read lists none.
type argumentIDs map[string]*idSource

// idSource is the read that lists the IDs that one argument of a command
// takes, for shell completion (completeArguments): a list, read as its own
// command reads it but its first page only, or a single response that holds
// the items; it offers each item's ID, the value of one column of the list's
// table, described by its NAME where that column identifies the item (offer).
type idSource struct {
	req    request  // what it reads: {parent} and {inner} name the arguments before, {site} --site
	scopes []string // the list's scopes, none write-capable
	read   readFunc // the glance of the list (glancer)
	items  shape    // where the items are in the response

	// offer returns the IDs of items that start with prefix.
	offer func(items []json.RawMessage, prefix string) []cobra.Completion
	// directive is how shells complete what offer returns, besides never
	// with file names: NoSpace where another ID may follow (severalIDs).
	directive cobra.ShellCompDirective
}

// idSeparator separates the IDs of an argument that takes several (severalIDs).
const idSeparator = ","

// severalIDs returns src for an argument that takes several of its IDs in
// one, separated by commas (e.g. a1,a2), or nil (none) where src is nil: it
// completes the last of them, after those typed before it, and offers none
// of those again, nor an ID that holds a comma; shells add no space after
// it, so that a comma and another ID may follow.
func (src *idSource) severalIDs() *idSource {
	if src == nil {
		return nil
	}
	several := *src
	several.directive = cobra.ShellCompDirectiveNoSpace
	several.offer = func(items []json.RawMessage, typed string) []cobra.Completion {
		cut := strings.LastIndex(typed, idSeparator) + 1
		before, last := typed[:cut], typed[cut:]
		given := strings.Split(before, idSeparator)
		var offered []cobra.Completion
		for _, id := range src.offer(items, last) {
			if value := completionValue(id); !strings.Contains(value, idSeparator) && !slices.Contains(given, value) {
				offered = append(offered, before+id)
			}
		}
		return offered
	}
	return &several
}

// completionValue returns the value of a completion, without its description.
func completionValue(c cobra.Completion) string {
	value, _, _ := strings.Cut(c, "\t")
	return value
}

// listedIDs returns the source of the IDs that column of rd's table shows for
// the items that req lists, as f reads them, or nil (none) where rd's scopes
// are write-capable, which completion never requests, or f is no glancer,
// such as a search whose body comes from flags. Only the IDs of the first
// column, the item's identifier, are described by the item's NAME: the value
// of another column, such as a TYPE, may name many items. A column that the
// table lacks is a mistake in the declaration, so it panics while the command
// tree is built, which every test of the tree does.
func listedIDs[T any](rd reader[T], req request, f fetcher, column string) *idSource {
	at := slices.Index(rd.header, column)
	if at < 0 {
		panic(fmt.Sprintf("completion: %s is not a column of %s", column, rd.singular))
	}
	g, glances := f.(glancer)
	if !glances || writeCapable(rd.scopes) {
		return nil
	}
	read, items := g.glance()
	name := -1 // the column that describes an ID: NAME, for the identifier only
	if n := slices.Index(rd.header, columnName); at == 0 && n > 0 {
		name = n
	}
	return &idSource{
		req: req, scopes: rd.scopes, read: read, items: items,
		offer: func(found []json.RawMessage, prefix string) []cobra.Completion {
			return offeredIDs(rd.view, found, at, name, prefix)
		},
	}
}

// offeredIDs returns up to completionSize IDs of items, the values of column
// of v's table, that start with prefix, each described by the value of the
// column name (none where it is negative or empty), once each and in the
// order listed. Only a NAME describes an ID, as no other column is known to
// keep no secret. An item that v cannot show, or an ID that no argument could
// take or that holds characters a terminal acts on, is left out; a
// description shows those characters as spaces.
func offeredIDs[T any](v view[T], items []json.RawMessage, column, name int, prefix string) []cobra.Completion {
	var ids []cobra.Completion
	seen := map[string]bool{}
	for _, raw := range items {
		if len(ids) == completionSize {
			break
		}
		var item T
		if json.Unmarshal(raw, &item) != nil {
			continue
		}
		cells := v.row(item, defaultLanguage)
		id := cells[column]
		if seen[id] || !isPathSegment(id) || output.Sanitize(id) != id || !strings.HasPrefix(id, prefix) {
			continue
		}
		seen[id] = true
		if name >= 0 && cells[name] != "" {
			id = cobra.CompletionWithDesc(id, output.Sanitize(cells[name]))
		}
		ids = append(ids, id)
	}
	return ids
}

// completeArguments makes shells complete the arguments of cmd, a read or a
// write, each with the values that values give it, by its name, such as its
// choices, which need no read and complete without the opt-in too, then with
// the IDs that ids list for it, as listedItems reads them, given the site
// that --site gives, else the default site of the profile in effect. Values
// or a read for an argument that cmd does not take, or a read that needs a
// site where cmd takes no --site, could never complete: that is a mistake in
// the declaration of cmd, so it panics while the command tree is built,
// which every test of the tree does.
func completeArguments(a *app, cmd *cobra.Command, site *siteCode, ids argumentIDs, values ...map[string][]string) {
	names := declaredArguments(cmd.Use)
	fixed := argumentValues(values)
	completes := slices.Collect(maps.Keys(fixed))
	for name, src := range ids {
		if src == nil {
			continue
		}
		completes = append(completes, name)
		if src.req.names(placeholderSite) && cmd.Flags().Lookup(siteFlag) == nil {
			panic(fmt.Sprintf("%s: completes %s from a list of a site, without --site", cmd.Use, name))
		}
	}
	for _, name := range completes {
		if !slices.Contains(names, name) {
			panic(fmt.Sprintf("%s: completes %s, which is not one of its arguments", cmd.Use, name))
		}
	}
	if len(completes) == 0 {
		return
	}
	cmd.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) (
		[]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) >= len(names) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		name := names[len(args)]
		src := ids[name]
		var items []json.RawMessage
		// Cobra runs no hook while it completes, so the default site of the profile in effect applies
		// here, as it does where the command runs; only where the user opted in, as it reads the profiles.
		if src != nil && a.completesIDs() && a.applyProfileDefaults(c) == nil {
			items = a.listedItems(c.Context(), src, args, *site)
		}
		return offerArgument(fixed[name], src, items, toComplete)
	}
}

// argumentValues returns values, each the values that arguments take by
// their names, together by name.
func argumentValues(values []map[string][]string) map[string][]string {
	all := map[string][]string{}
	for _, byName := range values {
		for name, given := range byName {
			all[name] = append(all[name], given...)
		}
	}
	return all
}

// offerArgument returns what an argument completes with that starts with typed:
// those of values, which need no read, then the IDs that src offers of items
// (none where src is nil), each once, up to completionSize in all; and how
// shells complete them.
func offerArgument(values []string, src *idSource, items []json.RawMessage, typed string) (
	[]cobra.Completion, cobra.ShellCompDirective) {
	offered := slices.DeleteFunc(slices.Clone(values), func(v string) bool { return !strings.HasPrefix(v, typed) })
	if src == nil {
		return offered, cobra.ShellCompDirectiveNoFileComp
	}
	for _, id := range src.offer(items, typed) {
		if len(offered) < completionSize && !slices.Contains(values, completionValue(id)) {
			offered = append(offered, id)
		}
	}
	return offered, cobra.ShellCompDirectiveNoFileComp | src.directive
}

// explainCompletion says in the help of completion, the command that
// generates the completion scripts, how to opt in to completion that reads
// the API.
func explainCompletion(completion *cobra.Command) {
	completion.Long = strings.TrimSuffix(completion.Long, "\n") + "\n\n" + completionHelp()
}

// listedItems returns the items that src lists, its request filled in with the
// arguments args and the site. It reads only where the user opted in
// (envCompleteIDs), with src's scopes, one page at most, within the app's
// completion timeout, which covers the token and a profile's client secret
// too; it returns none at all, and says nothing, when its request needs an
// argument or a site not given (or one that addresses no single item), or
// anything fails.
func (a *app) listedItems(ctx context.Context, src *idSource, args []string, site siteCode) []json.RawMessage {
	if !a.completesIDs() || !src.fillable(args, site) {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, a.completionTimeout)
	defer cancel()
	client, err := a.client(ctx, src.scopesFor(args)...)
	if err != nil {
		return nil
	}
	path, query := src.req.fill(args, site)
	body, err := src.read(client, ctx, path, query)
	if err != nil {
		return nil
	}
	items, err := src.items(body)
	if err != nil {
		return nil
	}
	return items
}

// fillable reports whether args and site fill in every placeholder of src's
// request, each with a value that addresses a single item, and, where src's
// scopes name a custom entity type, args name one first (typeScopes).
func (src *idSource) fillable(args []string, site siteCode) bool {
	n := len(src.req.arguments())
	if n > len(args) || slices.ContainsFunc(args[:n], func(arg string) bool { return !isPathSegment(arg) }) {
		return false
	}
	if slices.ContainsFunc(src.scopes, namesType) && (n == 0 || !typeIDPattern.MatchString(args[0])) {
		return false
	}
	return site != "" || !src.req.names(placeholderSite)
}

// scopesFor returns the scopes of src's read for args, which fill it in
// (fillable): where they name a custom entity type, that of the first
// argument, as the command that lists the items requests them.
func (src *idSource) scopesFor(args []string) []string {
	if !slices.ContainsFunc(src.scopes, namesType) {
		return src.scopes
	}
	return withType(src.scopes, args[0])
}

// completesIDs reports whether the user opted in to completion that reads
// the API: envCompleteIDs is 1 or true.
func (a *app) completesIDs() bool {
	on, err := strconv.ParseBool(a.getenv(envCompleteIDs))
	return err == nil && on
}

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

const envCompleteIDs = "EPX_COMPLETE_IDS"

const defaultCompletionTimeout = 3 * time.Second

const completionSize = 100

func completionHelp() string {
	return "With " + envCompleteIDs + "=1 in the environment, the arguments that name items (such as ID of " +
		"`epx products get ID`), --site and --saved-search complete too, with up to " + strconv.Itoa(completionSize) +
		" of the IDs and site codes that " +
		"the first page of their list shows (the whole list where Emporix sends it in one response), " +
		"read with the API key in effect and the scopes of that list within " +
		defaultCompletionTimeout.String() + ". " +
		"Without it, completion never calls the API; when the API cannot be read, it offers no IDs " +
		"and prints nothing."
}

type glancer interface {
	glance() (readFunc, shape)
}

type argumentIDs map[string]*idSource

type idSource struct {
	req    request
	scopes []string
	rule   typeRule
	read   readFunc
	items  shape

	offer     func(items []json.RawMessage, prefix string) []cobra.Completion
	directive cobra.ShellCompDirective
}

const idSeparator = ","

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

func completionValue(c cobra.Completion) string {
	value, _, _ := strings.Cut(c, "\t")
	return value
}

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
	var described descriptionColumn
	if n := slices.Index(rd.header, columnName); at == 0 && n > 0 {
		described = descriptionColumn{at: n, shown: true}
	}
	return &idSource{
		req: req, scopes: rd.scopes, rule: newTypeRule(rd.scopes, builtInTypes{}), read: read, items: items,
		offer: func(found []json.RawMessage, prefix string) []cobra.Completion {
			return offeredIDs(rd.view, found, at, described, prefix)
		},
	}
}

type descriptionColumn struct {
	at    int
	shown bool
}

func (d descriptionColumn) of(cells []string) string {
	if !d.shown {
		return ""
	}
	return cells[d.at]
}

func offeredIDs[T any](v view[T], items []json.RawMessage, column int, described descriptionColumn,
	prefix string) []cobra.Completion {
	var ids []cobra.Completion
	seen := map[string]bool{}
	for _, raw := range items {
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
		if description := described.of(cells); description != "" {
			id = cobra.CompletionWithDesc(id, output.Sanitize(description))
		}
		ids = append(ids, id)
	}
	return ids
}

func completeArguments(a *app, cmd *cobra.Command, site *siteCode, ids argumentIDs, choices map[string][]string,
	rule typeRule) {
	names := declaredArguments(cmd.Use)
	fixed := rule.builtIn.values(cmd.Use, choices)
	if len(completedArguments(cmd, names, fixed, ids)) == 0 {
		return
	}
	first := ids[names[0]].accepting(rule)
	cmd.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) (
		[]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) >= len(names) || typeRefused(rule, args) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		name := names[len(args)]
		src := ids[name]
		if len(args) == 0 {
			src = first
		}
		return offerArgument(fixed[name], src, a.completionItems(c, src, args, site), toComplete)
	}
}

func typeRefused(rule typeRule, args []string) bool {
	return len(args) > 0 && rule.accepts(args[0]) != nil
}

func (a *app) completionItems(c *cobra.Command, src *idSource, args []string, site *siteCode) []json.RawMessage {
	if src == nil || !a.completesIDs() {
		return nil
	}
	// Cobra runs no hook while it completes.
	if err := a.applyProfileDefaults(c); err != nil {
		return nil
	}
	return a.listedItems(c.Context(), src, args, *site)
}

func completedArguments(cmd *cobra.Command, names []string, fixed map[string][]string, ids argumentIDs) []string {
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
	return completes
}

func (src *idSource) accepting(rule typeRule) *idSource {
	if src == nil {
		return nil
	}
	accepted := *src
	accepted.offer = func(items []json.RawMessage, typed string) []cobra.Completion {
		return slices.DeleteFunc(src.offer(items, typed), func(id cobra.Completion) bool {
			return rule.accepts(completionValue(id)) != nil
		})
	}
	return &accepted
}

func offerArgument(values []string, src *idSource, items []json.RawMessage, typed string) (
	[]cobra.Completion, cobra.ShellCompDirective) {
	offered := slices.DeleteFunc(slices.Clone(values), func(v string) bool { return !strings.HasPrefix(v, typed) })
	if src == nil {
		return offered, cobra.ShellCompDirectiveNoFileComp
	}
	for _, id := range src.offer(items, typed) {
		value := completionValue(id)
		if len(offered) < completionSize && !slices.Contains(values, value) {
			offered = append(offered, id)
		}
	}
	return offered, cobra.ShellCompDirectiveNoFileComp | src.directive
}

func explainCompletion(completion *cobra.Command) {
	completion.Long = strings.TrimSuffix(completion.Long, "\n") + "\n\n" + completionHelp()
}

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

func (src *idSource) fillable(args []string, site siteCode) bool {
	n := len(src.req.arguments())
	if n > len(args) || slices.ContainsFunc(args[:n], func(arg string) bool { return !isPathSegment(arg) }) {
		return false
	}
	if src.rule.custom() && (n == 0 || src.rule.accepts(args[0]) != nil) {
		return false
	}
	return site != "" || !src.req.names(placeholderSite)
}

func (src *idSource) scopesFor(args []string) []string { return src.rule.fill(args, src.scopes) }

func (a *app) completesIDs() bool {
	on, err := strconv.ParseBool(a.getenv(envCompleteIDs))
	return err == nil && on
}

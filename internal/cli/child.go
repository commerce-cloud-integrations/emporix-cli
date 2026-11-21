package cli

import (
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type child[T any] struct {
	view[T]
	use       string
	sub       string
	short     string
	long      string
	query     url.Values
	listQuery url.Values
	scopes    []string
	fetch     fetcher
	unknown   unknownParent
}

type unknownParent string

const (
	notFound        unknownParent = "is not found"
	listsNothing    unknownParent = "lists nothing"
	answeredDeleted unknownParent = "is answered as one that Emporix deleted (204 No Content)"
)

func childCommand[P, T any](a *app, parent resource[P], c child[T]) *cobra.Command {
	rd := c.reader(parent.scopes)
	return rd.command(a, &command{
		use:   c.use + " " + parent.argument(),
		short: c.short,
		long:  joinSentences(c.long, parent.argumentSentence(unknownFor(rd.scopes, cmp.Or(c.unknown, notFound)))),
		req:   c.listRequest(parent.path),
		fetch: c.fetch,
		ids:   argumentIDs{parent.argument(): parent.ids()},
	})
}

func (c child[T]) reader(parentScopes []string) reader[T] {
	if c.scopes != nil {
		return reader[T]{view: c.view, scopes: c.scopes}
	}
	return reader[T]{view: c.view, scopes: parentScopes}
}

func (c child[T]) collection() string {
	if c.use == listUse && c.sub == "" {
		panic(c.singular + ": a child listed by " + listUse + " names its path in sub")
	}
	return cmp.Or(c.sub, c.use)
}

func childAt[P, T any](parent resource[P], c child[T]) collectionAt {
	target := childItemTarget(parent, c, getUse, nil)
	return collectionAt{path: c.below(parent.path), parent: parent.argument(), item: strings.Join(target.args, " "),
		itemPath: target.path, get: parent.commandLine(target.use, request{path: target.path, query: c.query}),
		parentIDs: argumentIDs{parent.argument(): parent.ids()}, itemIDs: childItemIDs(parent, c, &target)}
}

func (c child[T]) listRequest(parentPath string) request {
	return withQueries(c.below(parentPath), c.query, c.listQuery)
}

func (c child[T]) below(parentPath string) string {
	return parentPath + "/" + placeholderParent + "/" + c.collection()
}

func childListLine[P, T any](parent resource[P], c child[T]) string {
	return parent.commandLine(c.use+" "+parent.argument(), c.listRequest(parent.path))
}

func (r resource[T]) nested(group string) resource[T] {
	r.under, r.use, r.listedBy, r.atPath = strings.TrimSpace(r.under+" "+r.use), group, r.listLine(), r.listsAtPath()
	return r
}

func (r resource[T]) listsAtPath() bool { return r.listedBy == "" || r.atPath }

func (c child[T]) searchUse() string {
	if c.use == listUse {
		return collectionSearch
	}
	return collectionSearch + "-" + c.use
}

func childSearchCommand[P, T any](a *app, parent resource[P], c child[T], short, long string,
	fields ...bodyField) *cobra.Command {
	members, listed := c.fetch.(listing)
	if !listed {
		panic(c.searchUse() + ": only a listed child collection has a search")
	}
	search := c
	search.use = c.searchUse()
	search.sub = c.collection() + "/" + collectionSearch
	search.short = short
	search.long = joinSentences(long, searchSentence(childListLine(parent, c)))
	search.fetch = members.search(fields...)
	return childCommand(a, parent, search)
}

type childSearch struct {
	short  string
	long   string
	fields []bodyField
}

type childGroup struct {
	short  string
	long   string
	search *childSearch
	items  []*childItem
	writes []*cobra.Command
}

func newChildGroup[P, T any](a *app, parent resource[P], c child[T], g *childGroup) *cobra.Command {
	if parent.listedBy == "" {
		panic(c.singular + ": a child in a group of its own is below a nested parent")
	}
	commands := []*cobra.Command{childCommand(a, parent, c)}
	if g.search != nil {
		commands = append(commands, childSearchCommand(a, parent, c, g.search.short, g.search.long, g.search.fields...))
	}
	for _, item := range g.items {
		commands = append(commands, childItemCommand(a, parent, c, item))
	}
	return newGroup(parent.use, g.short, g.long, append(commands, g.writes...)...)
}

type childItem struct {
	use     string
	short   string
	long    string
	columns []string
	choices map[string][]string
	items   shape
}

func childItemCommand[P, T any](a *app, parent resource[P], c child[T], item *childItem) *cobra.Command {
	target := childItemTarget(parent, c, item.use, item.columns)
	return c.reader(parent.scopes).command(a, &command{
		use:     target.use,
		short:   item.short,
		long:    joinSentences(item.long, childItemHelp(parent, c, &target)),
		req:     request{path: target.path, query: c.query},
		fetch:   single{items: itemShape(item.items)},
		choices: item.choices,
		ids:     childItemIDs(parent, c, &target),
	})
}

type childTarget struct {
	use     string
	args    []string
	columns []string
	path    string
}

func childItemIDs[P, T any](parent resource[P], c child[T], target *childTarget) argumentIDs {
	ids := argumentIDs{parent.argument(): parent.ids()}
	if len(target.columns) == 1 {
		ids[target.args[0]] = childIDs(parent, c, target.columns[0])
	}
	return ids
}

func childIDs[P, T any](parent resource[P], c child[T], column string) *idSource {
	return listedIDs(c.reader(parent.scopes), c.listRequest(parent.path), c.fetch, column)
}

func childItemHelp[P, T any](parent resource[P], c child[T], target *childTarget) string {
	return childItemAnswerHelp(parent, c, target, notFound)
}

func childItemAnswerHelp[P, T any](parent resource[P], c child[T], target *childTarget, item unknownParent) string {
	unknown := unknownFor(c.reader(parent.scopes).scopes, notFound)
	arguments := itemAnswerSentence(target.args, target.columns, childListLine(parent, c), item)
	return parent.argumentSentence(unknown) + " " + arguments
}

func childItemTarget[P, T any](parent resource[P], c child[T], verb string, columns []string) childTarget {
	if len(columns) == 0 {
		columns = c.header[:1]
	}
	args, key := c.itemArguments(verb, columns)
	return childTarget{
		use:  childItemUse(parent, verb, args),
		args: args, columns: columns, path: c.below(parent.path) + "/" + key,
	}
}

func childItemUse[P any](parent resource[P], verb string, args []string) string {
	return strings.Join(append([]string{verb, parent.argument()}, args...), " ")
}

func (c child[T]) itemArguments(use string, columns []string) (args []string, key string) {
	placeholders := []string{placeholderInner, placeholderID}
	if len(columns) > len(placeholders) {
		panic(fmt.Sprintf("%s: an item takes at most %d arguments after its parent's", use, len(placeholders)))
	}
	for _, column := range columns {
		if !slices.Contains(c.header, column) {
			panic(fmt.Sprintf("%s: %s is not a column of %s", use, column, c.singular))
		}
		args = append(args, strings.ReplaceAll(column, " ", "_"))
	}
	return args, strings.Join(placeholders[len(placeholders)-len(columns):], "/")
}

func itemSentence(args, columns []string, line string) string {
	return itemAnswerSentence(args, columns, line, notFound)
}

func itemAnswerSentence(args, columns []string, line string, unknown unknownParent) string {
	verb := "is"
	if len(args) > 1 {
		verb = "are"
	}
	return fmt.Sprintf("%s %s the %s that `epx %s` shows; an unknown one %s.", strings.Join(args, " and "), verb,
		strings.Join(columns, " and "), line, unknown)
}

func (r resource[T]) argument() string {
	return strings.ToUpper(strings.ReplaceAll(r.singular, " ", "_")) + "_" + r.header[0]
}

func (r resource[T]) argumentSentence(unknown unknownParent) string {
	return fmt.Sprintf("%s; an unknown one %s.", r.argumentSource(), unknown)
}

func (r resource[T]) argumentSource() string { return r.source(r.argument()) }

func (r resource[T]) idSource() string { return r.source(r.header[0]) }

func (r resource[T]) source(arg string) string {
	return fmt.Sprintf("%s is the %s that `epx %s` shows", arg, r.header[0], r.listLine())
}

func (r resource[T]) listLine() string {
	use := listUse
	if r.search != nil {
		use = collectionSearch
	}
	return cmp.Or(r.listedBy, r.commandLine(use, r.listRequest()))
}

func (r resource[T]) commandLine(use string, req request) string {
	line := strings.TrimSpace(r.under + " " + r.use + " " + use)
	if req.names(placeholderSite) {
		line += " --site S"
	}
	return line
}

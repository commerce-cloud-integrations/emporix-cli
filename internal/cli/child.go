package cli

import (
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// child declares a collection that belongs to one item of a parent resource,
// such as the prices of a price list: `use PARENT_ID` reads it with the
// parent's scopes unless it declares its own (childCommand), and a read of
// one of its items, where Emporix offers one, comes from it too
// (childItemCommand). Its commands are in the parent's group, or, where it
// has writes, in a group of its own below a nested parent (resource.nested),
// which newChildGroup makes, e.g. `segments customers list SEGMENT_ID`.
type child[T any] struct {
	view[T]
	use       string        // command name, e.g. "prices"
	sub       string        // the collection's path below the parent item; use when empty
	short     string        // summary of the command
	long      string        // what it reads; the argument and scope sentences follow
	query     url.Values    // query of the list and of one item (childItem), e.g. showUnpublished=true
	listQuery url.Values    // query of the list only, e.g. expand=fees
	scopes    []string      // OAuth scopes besides the tenant; nil: the parent's, empty: none
	fetch     fetcher       // a listing, or a single response holding the items
	unknown   unknownParent // what an unknown parent item gives; notFound when empty (unknownFor may override it)
}

// unknownParent is what a child read of an unknown parent item gives, as
// help text says it.
type unknownParent string

// What a child read of an unknown parent item gives.
const (
	notFound     unknownParent = "is not found"
	listsNothing unknownParent = "lists nothing" // Emporix answers with an empty collection
	// answeredDeleted is what a delete of an unknown item gives where Emporix
	// answers it as one that it deleted, as unknownAsDeleted says.
	answeredDeleted unknownParent = "is answered as one that Emporix deleted (204 No Content)"
)

// childCommand returns the command of c for one item of parent, read with
// c's scopes or else the parent's.
func childCommand[P, T any](a *app, parent resource[P], c child[T]) *cobra.Command {
	rd := c.reader(parent.scopes)
	return rd.command(a, &command{
		use:   c.use + " " + parent.argument(),
		short: c.short,
		long:  c.long + "\n" + parent.argumentSentence(unknownFor(rd.scopes, cmp.Or(c.unknown, notFound))),
		req:   c.listRequest(parent.path),
		fetch: c.fetch,
		ids:   argumentIDs{parent.argument(): parent.ids()},
	})
}

// reader returns the reader of c's items, with c's scopes or else parentScopes.
func (c child[T]) reader(parentScopes []string) reader[T] {
	if c.scopes != nil {
		return reader[T]{view: c.view, scopes: c.scopes}
	}
	return reader[T]{view: c.view, scopes: parentScopes}
}

// collection returns the path of c's collection below the parent item: sub,
// or else use. A child listed by `list`, below a group of its own (nested),
// names its path in sub: one without sub is a mistake in its declaration (it
// would send to .../list), so it panics while the command tree is built,
// which every test of the tree does.
func (c child[T]) collection() string {
	if c.use == listUse && c.sub == "" {
		panic(c.singular + ": a child listed by " + listUse + " names its path in sub")
	}
	return cmp.Or(c.sub, c.use)
}

// childAt returns where the items of c, a child collection of parent, are
// for its writes: below one item of parent, whose argument they take first,
// each item named as a read of one (get, below parent) names it.
func childAt[P, T any](parent resource[P], c child[T]) collectionAt {
	target := childItemTarget(parent, c, getUse, nil)
	return collectionAt{path: c.below(parent.path), parent: parent.argument(), item: strings.Join(target.args, " "),
		itemPath: target.path, get: parent.commandLine(target.use, request{path: target.path, query: c.query}),
		parentIDs: argumentIDs{parent.argument(): parent.ids()}, itemIDs: childItemIDs(parent, c, &target)}
}

// listRequest returns the request of c's list below the parent item
// {parent} of the collection at parentPath, with query and listQuery.
func (c child[T]) listRequest(parentPath string) request {
	return withQueries(c.below(parentPath), c.query, c.listQuery)
}

// below returns the path of c's collection below the parent item {parent} of
// the collection at parentPath.
func (c child[T]) below(parentPath string) string {
	return parentPath + "/" + placeholderParent + "/" + c.collection()
}

// childListLine returns the command line, without epx, that lists c below
// one item of parent, e.g. "price-lists prices PRICE_LIST_ID", or, for a
// parent nested in a group of its own, "sites mixins list SITE_CODE".
func childListLine[P, T any](parent resource[P], c child[T]) string {
	return parent.commandLine(c.use+" "+parent.argument(), c.listRequest(parent.path))
}

// nested returns r as the parent of the commands of a group of their own,
// named group, below r's group, such as `sites mixins`: their command lines
// name that group (commandLine), while r's list still lists the items that
// their first argument names (listedBy), e.g. "sites mixins list SITE_CODE",
// where SITE_CODE is the CODE that `epx sites list` shows, at r's path
// (atPath), which completion reads too.
func (r resource[T]) nested(group string) resource[T] {
	r.under, r.use, r.listedBy, r.atPath = strings.TrimSpace(r.under+" "+r.use), group, r.listLine(), true
	return r
}

// searchUse returns the name of the search of c: search beside the list of
// a group of its own (use list, its parent nested), else search-<use>, e.g.
// search-prices.
func (c child[T]) searchUse() string {
	if c.use == listUse {
		return collectionSearch
	}
	return collectionSearch + "-" + c.use
}

// childSearchCommand returns the search of c (searchUse) PARENT_ID, which
// lists the items of c, a child collection of parent, that a search finds
// below one item of parent: a POST to the search path below c's collection,
// with c's query and listQuery and the body that fields make (or --body),
// read as c lists its items (listing.search) and shown as c shows them, with
// c's scopes or else the parent's. Only a collection that c lists (fetch is a
// listing) has a search: another fetch is a mistake in the declaration of c,
// so it panics while the command tree is built, which every test of the tree
// does.
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
	search.long = long + "\n" + searchSentence(childListLine(parent, c))
	search.fetch = members.search(fields...)
	return childCommand(a, parent, search)
}

// childSearch declares the search of a child collection: its summary, what
// it reads (the search sentence follows) and the fields of its body.
type childSearch struct {
	short  string
	long   string
	fields []bodyField
}

// childGroup declares the commands of a child collection in a group of its
// own, as a child with writes has: the group's summary and help, the search
// of the collection, if any, the reads of one item and the writes.
type childGroup struct {
	short  string
	long   string
	search *childSearch
	items  []*childItem
	writes []*cobra.Command
}

// newChildGroup returns the group of c below one item of parent, a parent
// nested in that group (resource.nested), named as the help of its commands
// names it (childListLine): list (c.use) PARENT_ID, search PARENT_ID where g
// declares one, the reads of one item and the writes of g. A parent that is
// not nested is a mistake in the declaration, as the help of the commands
// would name a group that does not exist, so it panics while the command
// tree is built, which every test of the tree does.
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

// childItem declares the read of one item of a child collection, such as one
// price of a price list: `use PARENT_ID ARGUMENTS` reads it whole, with the
// child's view, scopes and query (not its listQuery).
type childItem struct {
	use     string              // command name, e.g. "price"
	short   string              // summary of the command
	long    string              // what it reads; the argument and scope sentences follow
	columns []string            // the child's columns whose values name the item, in path order; its first when empty
	choices map[string][]string // the only values that an argument takes, by its name, e.g. TYPE: PRODUCT, CATEGORY
	items   shape               // what the table shows of the response, in the child's view; oneItem (the item) when nil
}

// childItemCommand returns the command of item, which reads one item of c's
// collection below one item of parent.
func childItemCommand[P, T any](a *app, parent resource[P], c child[T], item *childItem) *cobra.Command {
	target := childItemTarget(parent, c, item.use, item.columns)
	return c.reader(parent.scopes).command(a, &command{
		use:     target.use,
		short:   item.short,
		long:    item.long + childItemHelp(parent, c, &target),
		req:     request{path: target.path, query: c.query},
		fetch:   single{items: itemShape(item.items)},
		choices: item.choices,
		ids:     childItemIDs(parent, c, &target),
	})
}

// childTarget is how a command addresses one item of a child collection
// below one item of its parent, a read (childItemCommand) or a write.
type childTarget struct {
	use     string   // the command's name and arguments, e.g. "get-b2b SEGMENT_ID CUSTOMER LEGAL_ENTITY"
	args    []string // the arguments that name the item, after the parent's
	columns []string // the columns of the child's table whose values they take
	path    string   // the path of the item, which the arguments fill in
}

// childItemIDs returns the sources of the IDs of the arguments of a command
// on one item of c below one item of parent, which target addresses: the
// parent's, then, where one column names the item, that column's below the
// parent item (two columns would offer the values of each, not the pairs that
// name an item).
func childItemIDs[P, T any](parent resource[P], c child[T], target *childTarget) argumentIDs {
	ids := argumentIDs{parent.argument(): parent.ids()}
	if len(target.columns) == 1 {
		ids[target.args[0]] = childIDs(parent, c, target.columns[0])
	}
	return ids
}

// childIDs returns the source of the values that column of c's table shows
// for the items of c below the item of parent that a command's first argument
// names, for completion.
func childIDs[P, T any](parent resource[P], c child[T], column string) *idSource {
	return listedIDs(c.reader(parent.scopes), c.listRequest(parent.path), c.fetch, column)
}

// childItemHelp returns what follows the help of a command on one item of c
// below one item of parent, which target addresses: where the parent's
// argument and the item's come from, and what an unknown one gives: not
// found, or no scope where the reads of c request those of the custom entity
// type that the parent is (unknownFor).
func childItemHelp[P, T any](parent resource[P], c child[T], target *childTarget) string {
	return childItemAnswerHelp(parent, c, target, notFound)
}

// childItemAnswerHelp is childItemHelp of a command that Emporix answers as
// item says for an unknown item of a known parent, e.g. answeredDeleted.
func childItemAnswerHelp[P, T any](parent resource[P], c child[T], target *childTarget, item unknownParent) string {
	unknown := unknownFor(c.reader(parent.scopes).scopes, notFound)
	arguments := itemAnswerSentence(target.args, target.columns, childListLine(parent, c), item)
	return "\n" + parent.argumentSentence(unknown) + "\n" + arguments
}

// childItemTarget returns how the command verb addresses one item of c below
// one item of parent: by the parent's argument, then by the values of columns
// of c's table, its first when columns is empty (itemArguments).
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

// childItemUse returns the name and arguments of the command verb on one item
// of a child collection below one item of parent: the parent's argument, then
// args, those of the item (childItemTarget), e.g. "update SITE_CODE NAME".
func childItemUse[P any](parent resource[P], verb string, args []string) string {
	return strings.Join(append([]string{verb, parent.argument()}, args...), " ")
}

// itemArguments returns the arguments of the item read use, named after the
// columns of c's table whose values they take (LEGAL_ENTITY for LEGAL ENTITY),
// and the path of the item below c's collection, which they fill in: {id}, or
// {inner}/{id} for two. A column that the table lacks, or a third one, is a
// mistake in the declaration of use, so it panics while the command tree is
// built, which every test of the tree does.
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

// itemSentence says that the arguments of an item read are the values of the
// columns that the command line shows, and that an unknown one is not found.
func itemSentence(args, columns []string, line string) string {
	return itemAnswerSentence(args, columns, line, notFound)
}

// itemAnswerSentence says that the arguments of an item command are the values
// of the columns that the command line shows, and what an unknown one gives.
func itemAnswerSentence(args, columns []string, line string, unknown unknownParent) string {
	verb := "is"
	if len(args) > 1 {
		verb = "are"
	}
	return fmt.Sprintf("%s %s the %s that `epx %s` shows; an unknown one %s.", strings.Join(args, " and "), verb,
		strings.Join(columns, " and "), line, unknown)
}

// argument returns the name of the argument that names one item of r in a
// child read, e.g. PRICE_LIST_ID or CUSTOMER_NUMBER.
func (r resource[T]) argument() string {
	return strings.ToUpper(strings.ReplaceAll(r.singular, " ", "_")) + "_" + r.header[0]
}

// argumentSentence says where the argument of a child read comes from and
// what an unknown one gives.
func (r resource[T]) argumentSentence(unknown unknownParent) string {
	return fmt.Sprintf("%s; an unknown one %s.", r.argumentSource(), unknown)
}

// argumentSource says where the argument that names one item of r in a child
// read comes from, e.g. "PRICE_LIST_ID is the ID that `epx price-lists list`
// shows".
func (r resource[T]) argumentSource() string { return r.source(r.argument()) }

// idSource says where the ID argument of a command of one item of r (itemUse)
// comes from, e.g. "ID is the ID that `epx vendors list` shows".
func (r resource[T]) idSource() string { return r.source(r.header[0]) }

// source says that the argument arg names one item of r by the identifier
// that r's list shows.
func (r resource[T]) source(arg string) string {
	return fmt.Sprintf("%s is the %s that `epx %s` shows", arg, r.header[0], r.listLine())
}

// itemArgumentsSentence says where the arguments of a read of one item below
// an item of r come from: r's argument, from r's list, and item, which the
// command line source (without epx) shows; unknown says what an unknown one
// gives, e.g. "an unknown one is not found (exit 4)".
func (r resource[T]) itemArgumentsSentence(item, source, unknown string) string {
	return fmt.Sprintf("%s and %s are the IDs that `epx %s` and `epx %s` show;\n%s.", r.argument(), item,
		r.listLine(), source, unknown)
}

// listLine returns the command line, without epx, that lists r's items, e.g.
// "iam users list", "shipping zones list --site S", for a collection that
// Emporix lists only by search (r.search), "carts search", or, for a group
// with neither, listedBy.
func (r resource[T]) listLine() string {
	use := listUse
	if r.search != nil {
		use = collectionSearch
	}
	return cmp.Or(r.listedBy, r.commandLine(use, r.listRequest()))
}

// commandLine returns the command line of r's read use, which sends req,
// e.g. "shipping zones list --site S".
func (r resource[T]) commandLine(use string, req request) string {
	line := strings.TrimSpace(r.under + " " + r.use + " " + use)
	if req.names(placeholderSite) {
		line += " --site S"
	}
	return line
}

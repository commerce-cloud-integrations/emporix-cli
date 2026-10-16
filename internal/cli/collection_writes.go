package cli

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// collectionAt is where the items of a collection are, as its writes name
// them: its path, the argument of the item that it is below, if any, the
// arguments that name one of its items and its path, and the read that
// prints one item as an update takes it.
type collectionAt struct {
	path     string // e.g. /price/{tenant}/price-lists/{parent}/prices
	parent   string // the argument of the item it is below, {parent} in path, e.g. PRICE_LIST_ID; none for the tenant's
	item     string // the arguments that name one item, e.g. ID
	itemPath string // the path of one item, which they fill in, e.g. path + "/{id}"
	get      string // the command line, without epx, that reads one item, e.g. "price-lists prices get PRICE_LIST_ID ID"
	// parentIDs completes the parent's argument, itemIDs that and the item's
	// arguments (completeArguments).
	parentIDs, itemIDs argumentIDs
}

// use returns the use of the write verb of the collection, which takes the
// argument of the item that it is below, if any, e.g. "create PRICE_LIST_ID".
func (c *collectionAt) use(verb string) string { return strings.TrimSpace(verb + " " + c.parent) }

// itemUse returns the use of the write verb of one item, e.g. "update
// PRICE_LIST_ID ID".
func (c *collectionAt) itemUse(verb string) string { return c.use(verb) + " " + c.item }

// at returns where the items of r are, for its writes.
func (r resource[T]) at() collectionAt {
	return collectionAt{path: r.path, item: r.header[0], itemPath: r.itemPath(),
		get: r.commandLine(r.itemUse(getUse), r.itemRequest()), itemIDs: r.itemIDs()}
}

// collectionWrites declares the writes of the items of a collection: create
// (a POST of --file to the collection), update ID (a PUT of --file to one
// item, which replaces it, or, where Emporix creates a missing one too, an
// upsert) and delete ID (a DELETE of one item, without a body), with the
// help that each gives; and its bulk writes, if any, whose path update and
// delete then refuse as an ID.
type collectionWrites struct {
	at           collectionAt
	one          string       // one item with its article, in summaries, e.g. "a price model"
	scopes       []string     // the scopes of each write; one write-capable at least
	language     languageMode // whether create and update send localized text, and how (languageMode)
	languageHelp string       // what their help says of localized text; that of their language mode when empty
	upsert       bool         // update creates a missing item too
	create       string       // what create sends and Emporix answers; how it is sent and the scope sentence follow
	update       string       // what update sends, alike
	remove       string       // what delete deletes, alike
	// bulk declares its bulk writes, sent to the collection with these
	// scopes; none when nil.
	bulk *bulkWrites
}

// writes returns create, update and delete, as w declares them, which a
// caller may change, such as their scopes, before it makes their commands
// (writeCommands).
func (w *collectionWrites) writes() (create, update, remove *write) {
	replace := "Replace "
	if w.upsert {
		replace = "Upsert "
	}
	create = &write{
		use: w.at.use(createUse), short: "Create " + w.one, long: w.create, method: http.MethodPost,
		req: request{path: w.at.path}, scopes: w.scopes, body: &payload{}, language: w.language,
		languageHelp: w.languageHelp, ids: w.at.parentIDs,
	}
	update = &write{
		use: w.at.itemUse("update"), short: replace + w.one, long: w.update, method: http.MethodPut,
		req: request{path: w.at.itemPath}, scopes: w.scopes, body: &payload{}, language: w.language,
		languageHelp: w.languageHelp, ids: w.at.itemIDs,
	}
	remove = deleteWrite(w.at.itemUse(deleteUse), "Delete "+w.one, w.remove, request{path: w.at.itemPath}, w.scopes)
	remove.ids = w.at.itemIDs
	if w.bulk != nil {
		update.reserved = reservedFor(w.bulk.update, bulkUpdateUse)
		remove.reserved = reservedFor(w.bulk.delete, bulkDeleteUse)
	}
	return create, update, remove
}

// reservedFor returns the ID that a write of one item reserves for the bulk
// write named use, which help declares: the bulk path segment; none where
// help is empty, as the collection then has no such bulk write.
func reservedFor(help, use string) map[string]string {
	if help == "" {
		return nil
	}
	return map[string]string{bulkSegment: use}
}

// commands returns the commands of w's writes and of its bulk writes.
func (w *collectionWrites) commands(a *app) []*cobra.Command {
	create, update, remove := w.writes()
	writes := []write{*create, *update, *remove}
	if w.bulk != nil {
		bulk := *w.bulk
		bulk.path, bulk.args, bulk.scopes, bulk.ids = w.at.path, w.at.parent, w.scopes, w.at.parentIDs
		writes = append(writes, bulk.writes()...)
	}
	return writeCommands(a, writes)
}

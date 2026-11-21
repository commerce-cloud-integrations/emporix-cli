package cli

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

type collectionAt struct {
	path               string
	parent             string
	item               string
	itemPath           string
	get                string
	parentIDs, itemIDs argumentIDs
}

func (c *collectionAt) use(verb string) string { return strings.TrimSpace(verb + " " + c.parent) }

func (c *collectionAt) itemUse(verb string) string { return c.use(verb) + " " + c.item }

func (r resource[T]) at() collectionAt {
	return collectionAt{path: r.path, item: r.header[0], itemPath: r.itemPath(),
		get: r.getLine(), itemIDs: r.itemIDs()}
}

type collectionWrites struct {
	at           collectionAt
	one          string
	scopes       []string
	language     languageMode
	languageHelp string
	upsert       bool
	create       string
	update       string
	remove       string
	bulk         *bulkWrites
}

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

func reservedFor(help, use string) map[string]string {
	if help == "" {
		return nil
	}
	return map[string]string{bulkSegment: use}
}

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

package cli

import "net/http"

// The writes below declare the usual writes of one item of a resource, each
// to r's collection or to one of its items, which a group builds with
// write.command among its extra commands: create (createWrite), update ID
// where Emporix creates a missing item (upsertWrite) and delete ID
// (deleteWrite). Each requests scopes and says in long, its help, what it
// sends and what Emporix answers, and where ID comes from (idSource) if it
// takes one; its ID completes from r's list (itemIDs). A write that differs in
// more than that is a write of its own.

// createWrite declares create of one item of r: a POST of the JSON body of
// --file to r's collection.
func (r resource[T]) createWrite(scopes []string, long string) *write {
	return &write{
		use: createUse, short: "Create one " + r.singular, long: long, method: http.MethodPost,
		req: request{path: r.path}, scopes: scopes, body: &payload{},
	}
}

// upsertWrite declares update ID of r, of a service that replaces the item
// ID or creates it where none exists (an upsert): a PUT of the JSON body of
// --file to the item.
func (r resource[T]) upsertWrite(scopes []string, long string) *write {
	return &write{
		use: r.itemUse("update"), short: "Upsert one " + r.singular, long: long, method: http.MethodPut,
		req: request{path: r.itemPath()}, scopes: scopes, body: &payload{}, ids: r.itemIDs(),
	}
}

// deleteWrite declares delete ID of r: a DELETE of the item ID, which sends
// no body (itemDelete).
func (r resource[T]) deleteWrite(scopes []string, long string) *write {
	return itemDelete(r, "Delete one "+r.singular, long, scopes)
}

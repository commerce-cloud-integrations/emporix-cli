package cli

import "net/http"

func (r resource[T]) createWrite(scopes []string, long string) *write {
	return &write{
		use: createUse, short: "Create one " + r.singular, long: long, method: http.MethodPost,
		req: request{path: r.path}, scopes: scopes, body: &payload{},
	}
}

func (r resource[T]) upsertWrite(scopes []string, long string) *write {
	return &write{
		use: r.itemUse("update"), short: "Upsert one " + r.singular, long: long, method: http.MethodPut,
		req: request{path: r.itemPath()}, scopes: scopes, body: &payload{}, ids: r.itemIDs(),
	}
}

func (r resource[T]) deleteWrite(scopes []string, long string) *write {
	return itemDelete(r, "Delete one "+r.singular, long, scopes)
}

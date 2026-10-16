package cli

import (
	"cmp"
	"net/http"

	"github.com/spf13/cobra"
)

// returnRequest is the part of a Returns Service return shown in tables.
type returnRequest struct {
	ID             string          `json:"id"`
	ApprovalStatus string          `json:"approvalStatus"`
	Received       *bool           `json:"received"`
	Orders         []returnedOrder `json:"orders"`
	Requestor      struct {
		personName
		CustomerID string `json:"customerId"`
		Email      string `json:"email"`
	} `json:"requestor"`
	Total    valueAmount `json:"total"`
	Metadata struct {
		CreatedAt string `json:"createdAt"`
	} `json:"metadata"`
}

// returnedOrder is an order whose items are returned.
type returnedOrder struct {
	ID string `json:"id"`
}

// requestor returns who requested the return: their email, else their name,
// else their customer id.
func (r *returnRequest) requestor() string {
	return cmp.Or(r.Requestor.Email, r.Requestor.fullName(), r.Requestor.CustomerID)
}

// newReturnsCommand reads, creates, changes and deletes return requests.
func newReturnsCommand(a *app) *cobra.Command {
	returns := resource[returnRequest]{
		use:   "returns",
		short: "Read, create, change and delete returns",
		long: "Reads return requests, personal data included. STATUS is the approval status (PENDING,\n" +
			"APPROVED, REJECTED or CLOSED), RECEIVED whether the goods arrived, ORDERS the orders whose\n" +
			"items are returned, REQUESTOR who asked (email, name or customer id) and TOTAL the value\n" +
			"returned: the order entries' deprecated unitPrice times the quantity returned, net or gross\n" +
			"as that unitPrice is. Items (with their reasons and prices), the return reason, approver,\n" +
			"submitter, expiry date and assisted-buying entries are only in -o json.\n" +
			"`--query submitter.userType:CUSTOMER` lists the returns that customers made themselves\n" +
			"(EMPLOYEE: those made on their behalf).\n" +
			"`epx returns create`, `update`, `patch` and `delete` change returns.",
		path:    "/return/{tenant}/returns",
		scopes:  []string{"returns.returns_read"},
		filters: listFilters{kinds: filterable},
		view: view[returnRequest]{
			singular: "return",
			header:   []string{columnID, columnStatus, "RECEIVED", "ORDERS", "REQUESTOR", columnTotal, columnCreated},
			row: func(r returnRequest, _ language) []string {
				return []string{
					r.ID, r.ApprovalStatus, formatBool(r.Received),
					joinEach(r.Orders, ", ", func(o returnedOrder) string { return o.ID }), r.requestor(), r.Total.String(),
					r.Metadata.CreatedAt,
				}
			},
		},
	}
	return newResourceCommand(a, returns, commandsOf(a, returnWrites(returns))...)
}

// returnsManageScope is the scope of the Returns Service's writes made by an
// employee on behalf of a customer (returns.returns_manage_own is a
// customer's, with the customer's own token).
const returnsManageScope = "returns.returns_manage"

// returnWrites returns the writes of returns: create, update ID, patch ID
// and delete ID.
func returnWrites(returns resource[returnRequest]) []*write {
	item := request{path: returns.itemPath()}
	manage := []string{returnsManageScope}
	source := "\n" + returns.idSource() + "."
	return []*write{
		{
			use: createUse, short: "Create a return",
			long: "Creates a return on behalf of a customer from the JSON object in --file: its orders\n" +
				"(each with its id and the items to return, each with id and quantity), reason (code or\n" +
				"details), requestor (the customer, by customerId) and submitter (who submits it), and\n" +
				"optionally approver ({\"type\": \"USER\" or \"GROUP\", \"id\": an IAM user or group id};\n" +
				"Emporix refuses an id unknown for that type with 400 Bad Request, exit 1), expiryDate and\n" +
				"mixins. Emporix answers with the return's id. It does not check quantities against earlier\n" +
				"returns: sending the body again creates another return, even beyond the entry's amount.\n" +
				"It fails with 500 Internal Server Error (exit 1) for an order whose entries lack the\n" +
				"deprecated unitPrice (nothing is created).",
			method: http.MethodPost, req: request{path: returns.path}, scopes: manage, body: &payload{},
		},
		{
			use: returns.itemUse("update"), short: "Replace a return",
			long: "Replaces the return ID with the return in --file, a JSON object as\n" +
				"`epx returns get ID -o json` prints it, metadata.version included (Emporix refuses a body\n" +
				"without metadata with 400 Bad Request). Emporix replaces only the latest version: an older\n" +
				"metadata.version, like an unknown return, is not found (exit 4)." + source,
			method: http.MethodPut, req: item, scopes: manage, body: &payload{}, ids: returns.itemIDs(),
		},
		{
			use: returns.itemUse("patch"), short: "Change a return",
			long: "Changes the return ID with the JSON Patch (RFC 6902) in --file: an array of\n" +
				"operations, each with op (add, remove or replace), path and value, e.g.\n" +
				`[{"op": "replace", "path": "/approvalStatus", "value": "APPROVED"}]. Emporix requires its` +
				"\nread scope for it too, and answers a token without it with 400 Bad Request." + source,
			method: http.MethodPatch, req: item, scopes: []string{"returns.returns_read", returnsManageScope},
			body: operationsFile(), ids: returns.itemIDs(),
		},
		returns.deleteWrite(manage, "Deletes the return ID; it sends no body. Emporix deletes a PENDING or an\n"+
			"APPROVED return (the reference documents 400 Bad Request, exit 1, where the status does not\n"+
			"allow it)."+unknownAsDeleted("ID")+source),
	}
}

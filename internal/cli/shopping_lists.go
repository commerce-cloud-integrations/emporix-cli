package cli

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// customerShoppingLists is the shopping lists of a customer: each list is a
// member named after it, beside the customer's id and metadata.
type customerShoppingLists struct {
	CustomerID string
	Lists      []shoppingList // by name
}

// shoppingList is one shopping list of a customer, named by its member.
type shoppingList struct {
	Name  string
	Items []json.RawMessage // counted only
}

// UnmarshalJSON reads the lists from the members that are lists (listOf), in
// the order of their names, and the customer from the member id, as Emporix
// sends it, or else customerId, as the reference names it, where either is
// text rather than a list.
func (c *customerShoppingLists) UnmarshalJSON(data []byte) error {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}
	*c = customerShoppingLists{}
	var id, customerID string
	for _, name := range slices.Sorted(maps.Keys(members)) {
		raw := members[name]
		list, err := listOf(name, raw)
		switch {
		case err != nil:
		case list != nil:
			c.Lists = append(c.Lists, *list)
		case name == "id":
			err = json.Unmarshal(raw, &id)
		case name == "customerId":
			err = json.Unmarshal(raw, &customerID)
		}
		if err != nil {
			return fmt.Errorf("member %q: %w", name, err)
		}
	}
	c.CustomerID = cmp.Or(id, customerID)
	return nil
}

// listOf reads the shopping list that the member name holds, or nil where it
// holds none: a list is an object with items or a name. The other members,
// such as the customer's id and metadata (an object of neither), are not
// lists; a list named metadata is one all the same.
func listOf(name string, raw json.RawMessage) (*shoppingList, error) {
	if len(raw) == 0 || raw[0] != '{' { // a member of a decoded object starts with its value
		return nil, nil
	}
	var list struct {
		Name  json.RawMessage   `json:"name"`
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil || (list.Name == nil && list.Items == nil) {
		return nil, err
	}
	return &shoppingList{Name: name, Items: list.Items}, nil
}

// String returns the name of l and how many items it holds.
func (l shoppingList) String() string {
	if len(l.Items) == 1 {
		return l.Name + " (1 item)"
	}
	return fmt.Sprintf("%s (%d items)", l.Name, len(l.Items))
}

// newShoppingListsCommand reads, creates, replaces and deletes customers'
// shopping lists. The reference documents no paging for the list of every
// customer's lists, and Emporix sends no total, so it is read with one
// request.
func newShoppingListsCommand(a *app) *cobra.Command {
	const long = "Reads customers' shopping lists, personal data included: each customer's lists by name,\n" +
		"with how many items they hold. Items are only in -o json. CUSTOMER is the customer id the\n" +
		"lists are kept under (the reference calls it the customer number; for a customer that\n" +
		"Emporix numbered, the two are the same); get takes it, and a customer without shopping\n" +
		"lists is not found."
	lists := resource[customerShoppingLists]{
		use:    "shopping-lists",
		noun:   "customers' shopping lists",
		short:  "Read, create, replace and delete customers' shopping lists",
		long:   long,
		about:  long + "\ncreate, update and delete create, replace and delete a customer's lists.",
		path:   "/shoppinglist/{tenant}/shopping-lists",
		scopes: []string{"shoppinglist.shoppinglist_read"},
		read:   (*emporix.Client).GetUnpaged,
		view: view[customerShoppingLists]{
			singular: "customer's shopping lists",
			header:   []string{columnCustomer, "LISTS"},
			row: func(c customerShoppingLists, _ language) []string {
				return []string{c.CustomerID, joinEach(c.Lists, ", ", shoppingList.String)}
			},
		},
	}
	return newResourceCommand(a, lists, shoppingListWrites(a, lists)...)
}

// listNameParam names one shopping list of a customer, as the query parameter and the flag
// of `shopping-lists delete`.
const listNameParam = "name"

// shoppingListWrites returns the writes of customers' shopping lists: create
// (POST of one list), update CUSTOMER (PUT of one list of the customer),
// delete CUSTOMER --name NAME (one list) and delete-all CUSTOMER (every list
// of the customer: the same DELETE without name). Each requests only
// shoppinglist_manage, which the reference requires of an employee, as epx
// is. The reference names shoppinglist_delete for a delete too, but live,
// shoppinglist_manage alone deletes, and a token with shoppinglist_delete
// alone is taken for a customer's and refused (400 Bad Request).
func shoppingListWrites(a *app, lists resource[customerShoppingLists]) []*cobra.Command {
	manage := []string{"shoppinglist.shoppinglist_manage"}
	const unknown = "A customer without shopping lists is not found (exit 4)."
	const unknownList = "A customer without shopping lists, or without a list of that name, is not found (exit 4)."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a shopping list of a customer", method: http.MethodPost,
			long: "Creates the shopping list in --file, a JSON object with the customerId of the customer it\n" +
				"belongs to, the list's name and its items (each with its id, productId and quantity), and\n" +
				"prints what Emporix answers with: the customer's id. A list of that name that the customer\n" +
				"has already is a conflict (exit 5). Emporix does not check the customer: it creates the list\n" +
				"of an unknown customer id too.",
			example: "  epx shopping-lists create --file list.json --dry-run",
			req:     request{path: lists.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: lists.itemUse("update"), short: "Replace a shopping list of a customer", method: http.MethodPut,
			long: "Replaces the shopping list of CUSTOMER that the JSON in --file names with it: an object\n" +
				"with the list's name and its items (each with its id, productId and quantity). Send its\n" +
				"metadata.version as read to replace it only if nobody changed it since: Emporix answers\n" +
				"409 Conflict (exit 5) otherwise.\n" + unknownList + " Update creates no list; create does.",
			req: request{path: lists.itemPath()}, scopes: manage, body: &payload{}, ids: lists.itemIDs(),
		}).command(a),
		(&write{
			use: lists.itemUse(deleteUse), short: "Delete a shopping list of a customer", method: http.MethodDelete,
			long: "Deletes the shopping list --name of CUSTOMER; its other lists stay.\n" +
				unknownList + "\n`epx shopping-lists get CUSTOMER` shows the lists that are left.",
			req: request{path: lists.itemPath()}, scopes: manage, body: noBody{}, ids: lists.itemIDs(),
			params: queryFlags{required: []paramFlag{{param: listNameParam, flag: listNameParam,
				usage: "the `NAME` of the list to delete"}}},
		}).command(a),
		(&write{
			use: lists.itemUse("delete-all"), short: "Delete every shopping list of a customer",
			method: http.MethodDelete,
			long:   "Deletes every shopping list of CUSTOMER.\n" + unknown,
			req:    request{path: lists.itemPath()}, scopes: manage, body: noBody{}, ids: lists.itemIDs(),
			deletes: deletion{many: "every shopping list of the customer"},
		}).command(a),
	}
}

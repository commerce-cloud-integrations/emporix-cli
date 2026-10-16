package cli

import "github.com/spf13/cobra"

// contactAssignment is the part of a Customer Management contact assignment
// shown in tables: a customer acting as a contact of a legal entity.
type contactAssignment struct {
	ID          string `json:"id"`
	LegalEntity struct {
		ID string `json:"id"`
	} `json:"legalEntity"`
	Customer struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Email   string `json:"email"`
	} `json:"customer"`
	Type    string `json:"type"`
	Primary *bool  `json:"primary"`
}

// newContactAssignmentsCommand reads which customers are contacts of which
// legal entities, and in what role, and creates, upserts and deletes those
// assignments.
func newContactAssignmentsCommand(a *app) *cobra.Command {
	assignments := resource[contactAssignment]{
		use:   "contact-assignments",
		short: "Read and change which customers are contacts of which legal entities",
		about: "Reads and changes contact assignments, the customers that act as contacts of a legal\n" +
			"entity, personal data included.\n" + versionedWritesNote,
		long: "Reads contact assignments, the customers that act as contacts of a legal entity,\n" +
			"personal data included. CUSTOMER is the customer's id, not the NUMBER that `epx customers\n" +
			"get` takes. TYPE is PRIMARY, BILLING, LOGISTICS or CONTACT; the PRIMARY column is the\n" +
			"assignment's primary flag, separate from the TYPE PRIMARY. Phone numbers and the legal\n" +
			"entity's details are only in -o json.",
		path:    "/customer-management/{tenant}/contact-assignments",
		scopes:  []string{"customermanagement.contactassignment_read"},
		filters: listFilters{kinds: filterable},
		view: view[contactAssignment]{
			singular: "contact assignment",
			header:   []string{columnID, columnLegalEntity, columnCustomer, columnName, columnEmail, columnType, "PRIMARY"},
			row: func(c contactAssignment, _ language) []string {
				return []string{
					c.ID, c.LegalEntity.ID, c.Customer.ID, spaced(c.Customer.Name, c.Customer.Surname), c.Customer.Email,
					c.Type, formatBool(c.Primary),
				}
			},
		},
	}
	writes := assignments.versionedWrites(a, []string{"customermanagement.contactassignment_manage"}, &versioned{
		members: "legalEntity and customer (required, each as {\"id\": ...} of one that exists: 400 Bad Request,\n" +
			"exit 1, otherwise; the customer's id, not its number), type (PRIMARY, BILLING, LOGISTICS\n" +
			"or CONTACT, the default) and primary (true or false).",
		required: "legalEntity and customer; a type left out becomes CONTACT, a primary left out false",
		createNote: "A second assignment of a customer to the same legal entity is a conflict (exit 5) too.\n" +
			"Assigning a customer to a subsidiary also assigns it to the parent company, as a CONTACT\n" +
			"in an assignment of its own, which takes a delete of its own.",
	})
	return newResourceCommand(a, assignments, writes...)
}

package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// sequenceSchema is the part of a Sequential ID Service schema shown in tables.
type sequenceSchema struct {
	ID       string      `json:"id"`
	Type     string      `json:"schemaType"`
	Name     string      `json:"name"`
	Active   *bool       `json:"active"`
	PreText  string      `json:"preText"`
	Digits   json.Number `json:"numberOfDigits"`
	PostText string      `json:"postText"`
	Counter  json.Number `json:"counter"`
	Max      json.Number `json:"maxValue"`
}

// The scopes of the Sequential ID Service: of its reads and of the writes
// that issue IDs (write.readScoped), and of the writes that create, activate
// and delete schemas.
const (
	sequenceViewScope   = "sequentialid.schema_view"
	sequenceManageScope = "sequentialid.schema_manage"
)

// newSequentialIDsCommand reads and changes the schemas that number orders,
// invoices, customers and similar records, and issues IDs from them. The
// service returns them whole (live, it ignores pageNumber and pageSize and
// sends no total), so each list is read with one request.
func newSequentialIDsCommand(a *app) *cobra.Command {
	const columns = "TYPE is what a schema numbers, PREFIX and SUFFIX the text around the number\n" +
		"(placeholders such as __year__ are filled in when an ID is issued), DIGITS the number's\n" +
		"length, COUNTER the schema's current counter (it starts at the schema's start value) and\n" +
		"MAX the highest number it can reach."
	whole := (*emporix.Client).GetUnpaged
	schemas := resource[sequenceSchema]{
		use:   "sequential-ids",
		noun:  "sequential ID schemas",
		short: "Read, create, activate and delete sequential ID schemas; issue IDs",
		about: "Reads, creates, activates and deletes the schemas that number orders, invoices, quotes,\n" +
			"customers and similar records, and issues IDs from them.",
		long: "Reads the schemas that number orders, invoices, quotes, customers and similar records:\n" +
			columns + "\nThe service returns the schemas whole, so they are read with one request. An unknown\n" +
			"ID is not found (exit 4); Emporix answers an ID that is not 24 hexadecimal digits with an\n" +
			"internal error (500 Internal Server Error, exit 1).",
		path:   "/sequential-id/{tenant}/schemas",
		scopes: []string{sequenceViewScope},
		read:   whole,
		view: view[sequenceSchema]{
			singular: "sequential ID schema",
			header: []string{
				columnID, columnType, columnName, columnActive, "PREFIX", "DIGITS", "SUFFIX", "COUNTER", "MAX",
			},
			row: func(s sequenceSchema, _ language) []string {
				return []string{
					s.ID, s.Type, s.Name, formatBool(s.Active), s.PreText, s.Digits.String(), s.PostText,
					s.Counter.String(), s.Max.String(),
				}
			},
		},
	}
	forType := schemas.reader().command(a, &command{
		use:   "for-type TYPE",
		short: "List the sequential ID schemas of a type",
		long: "Reads the schemas of one TYPE, such as orderNoSequence, invoiceNoSequence or\n" +
			"customerNoSequence (`epx sequential-ids list` shows the tenant's types); an unknown type\n" +
			"lists nothing.\n" + columns,
		req:   request{path: schemas.path + "/types/" + placeholderID},
		fetch: listing{read: whole},
		ids:   argumentIDs{columnType: schemas.idsIn(columnType)},
	})
	return newResourceCommand(a, schemas, append([]*cobra.Command{forType}, sequenceSchemaWrites(a, schemas)...)...)
}

package cli

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"
)

// productTemplate is the part of a Product Service product template shown in tables.
type productTemplate struct {
	ID         string            `json:"id"`
	Name       localized         `json:"name"`
	Attributes []json.RawMessage `json:"attributes"`
	Metadata   itemMetadata      `json:"metadata"`
}

// newProductTemplatesCommand reads product templates, which need the
// product.product_template_read scope (Emporix answers 403 without it), and
// holds their writes, deletes included.
func newProductTemplatesCommand(a *app) *cobra.Command {
	templates := resource[productTemplate]{
		use:   "product-templates",
		short: "Read, create, update and delete product templates",
		about: "Reads and changes product templates, the attribute sets shared by products based on them:\n" +
			"create makes one, update replaces one, each from --file; delete deletes one.",
		long: "Reads product templates, the attribute sets shared by products based on them.\n" +
			"ATTRIBUTES counts the template's attributes (-o json prints their definitions).",
		path:    "/product/{tenant}/product-templates",
		filters: listFilters{kinds: filterable},
		scopes:  []string{"product.product_template_read"},
		view: view[productTemplate]{
			singular:  "product template",
			header:    []string{columnID, columnName, "ATTRIBUTES", columnVersion},
			localized: true,
			row: func(t productTemplate, lang language) []string {
				return []string{t.ID, t.Name.in(lang), strconv.Itoa(len(t.Attributes)), t.Metadata.Version.String()}
			},
		},
	}
	writes := append(productTemplateWrites(a, templates), itemDelete(templates, "Delete a product template",
		"Deletes the product template ID. Emporix deletes only a template that no product and no\n"+
			"parent variant is based on: it refuses one in use with 400 Bad Request (exit 1). An unknown\n"+
			"ID is no error: Emporix answers as for a deleted template.",
		[]string{"product.product_template_manage"}).command(a))
	return newResourceCommand(a, templates, writes...)
}

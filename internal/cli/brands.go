package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// brand is the part of a Brand Service brand shown in tables.
type brand struct {
	ID string `json:"id"`
	plainAndLocalizedName
}

// newBrandsCommand reads, creates, changes and deletes the tenant's brands.
// The Brand Service has no tenant path segment: the tenant comes from the
// token.
func newBrandsCommand(a *app) *cobra.Command {
	brands := resource[brand]{
		use:     "brands",
		short:   "Read, create, change and delete brands",
		long:    "Reads brands.",
		path:    "/brand/brands",
		filters: listFilters{kinds: filterable},
		view: view[brand]{
			singular:  "brand",
			header:    []string{columnID, columnName},
			localized: true,
			row:       func(b brand, lang language) []string { return []string{b.ID, b.displayName(lang)} },
		},
	}
	manage := []string{"brand.brand_manage"}
	const answer = " It prints the brand that Emporix answers with."
	return newResourceCommand(a, brands,
		(&write{
			use: createUse, short: "Create a brand", method: http.MethodPost,
			long: "Creates a brand from the JSON in --file, such as\n" +
				`{"name": "Acme", "localizedName": {"en": "Acme", "de": "Acme"}}, and prints the brand that` + "\n" +
				"Emporix answers with, including its id. Emporix generates the id unless the JSON names one;\n" +
				"an id that exists already is a conflict, while a name that exists already makes a second\n" +
				"brand.",
			req: request{path: brands.path}, scopes: manage, body: &payload{}, language: mapsOnly,
		}).command(a),
		(&write{
			use: brands.itemUse("update"), short: "Replace a brand", method: http.MethodPut,
			long: "Replaces the brand ID with the JSON in --file: a member that it leaves out is removed.\n" +
				"An unknown ID is not found." + answer,
			req: request{path: brands.itemPath()}, scopes: manage, body: &payload{}, language: mapsOnly,
			ids: brands.itemIDs(),
		}).command(a),
		(&write{
			use: brands.itemUse("patch"), short: "Change members of a brand", method: http.MethodPatch,
			long: "Changes the members of the brand ID that the JSON in --file names, such as name or\n" +
				"localizedDescription; the others keep their values. A map of translations in it, such as\n" +
				"localizedName, adds to the brand's: the languages it names change, the others stay. Unlike\n" +
				"update, Emporix does not check a metadata.version in it. An unknown ID is not found." + answer,
			req: request{path: brands.itemPath()}, scopes: manage, body: &payload{}, language: mapsOnly,
			ids: brands.itemIDs(),
		}).command(a),
		// The operation names brand.brand_delete, which the service's scopes do not list (only
		// brand.brand_manage, which its 403 answer names too) and which matches the per-verb scopes
		// that Emporix deprecated for *_manage in its other services.
		itemDelete(brands, "Delete a brand", "Deletes the brand ID.", manage).command(a),
	)
}

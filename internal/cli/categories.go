package cli

import (
	"net/url"
	"slices"

	"github.com/spf13/cobra"
)

// category is the part of a Category Service category shown in tables.
type category struct {
	ID        string `json:"id"`
	ParentID  string `json:"parentId"`
	Code      string `json:"code"`
	Type      string `json:"type"`
	Published *bool  `json:"published"`
	plainAndLocalizedName
}

// categoryAssignment is the part of a Category Service assignment of an item,
// a product, to a category shown in tables.
type categoryAssignment struct {
	ID  string `json:"id"`
	Ref struct {
		ID            string    `json:"id"`
		Type          string    `json:"type"`
		Published     *bool     `json:"published"`
		LocalizedName localized `json:"localizedName"`
	} `json:"ref"`
}

// withUnpublished returns the query of a Category Service read that includes
// unpublished categories (with the category.category_read_unpublished scope).
func withUnpublished() url.Values { return url.Values{"showUnpublished": {queryTrue}} }

// newCategoriesCommand reads categories, published and unpublished (Emporix
// omits unpublished ones without showUnpublished and the read_unpublished
// scope), what lies above and below one and what is assigned to it, and the
// categories a product is assigned to; it creates, changes and deletes
// categories and assigns products to them.
func newCategoriesCommand(a *app) *cobra.Command {
	r := resource[category]{
		use:     "categories",
		short:   "Read and change categories and the products assigned to them",
		long:    "Reads categories, published and unpublished.",
		path:    "/category/{tenant}/categories",
		query:   withUnpublished(),
		filters: listFilters{kinds: filterable},
		scopes:  []string{"category.category_read_unpublished"},
		view:    categoryView(),
	}
	// The assignments of a product are sorted but not filtered.
	reads := []*cobra.Command{
		r.searchCommand(a, "Reads the categories, published and unpublished, that an Emporix q filter finds\n"+
			"(--query; without it, every category).", queryField()),
		r.reader().command(a, &command{
			use:   forProductUse,
			short: "List the categories a product is assigned to",
			long: r.long + productIDNote + "\nAn unknown id lists nothing, because Emporix does not distinguish it\n" +
				"from a product without assignments.",
			req:   request{path: "/category/{tenant}/assignments/references/" + placeholderID, query: r.query},
			fetch: listing{filters: listFilters{kinds: sortable}},
			ids:   argumentIDs{productID: products().ids()},
		}),
		childCommand(a, r, child[category]{
			use:   "subcategories",
			short: "List the categories below a category",
			long: "Reads every category below a category, at all levels, published and unpublished; PARENT\n" +
				"is the id of the category each one lies directly below.",
			query: r.query,
			fetch: listing{filters: listFilters{kinds: sortable}},
			view:  subcategoryView(),
		}),
		childCommand(a, r, child[category]{
			use:   "parents",
			short: "List the categories above a category",
			long: "Reads the categories above a category, published and unpublished, from its root down\n" +
				"unless --sort says otherwise.",
			query: r.query,
			fetch: listing{filters: listFilters{kinds: sortable}},
			view:  categoryView(),
		}),
		newCategoryAssignmentsCommand(a, r),
	}
	return newResourceCommand(a, r, slices.Concat(reads, categoryWrites(a, r))...)
}

// newCategoryAssignmentsCommand returns the group of the products assigned to
// a category of categories: list CATEGORY_ID, which reads them, the writes
// that assign products (assignmentWrites) and those that unassign them
// (assignmentDeletes).
func newCategoryAssignmentsCommand(a *app, categories resource[category]) *cobra.Command {
	assignments := child[categoryAssignment]{
		use:   listUse,
		sub:   "assignments",
		short: "List the products assigned to a category",
		long: "Reads the products assigned to a category, published and unpublished. PRODUCT is the\n" +
			"product id (as in `epx products list`) and TYPE the kind of item, a product.",
		query: categories.query,
		fetch: listing{filters: listFilters{kinds: sortable}},
		view: view[categoryAssignment]{
			singular:  "category assignment",
			header:    []string{columnID, columnProduct, columnType, columnPublished, columnName},
			localized: true,
			row: func(c categoryAssignment, lang language) []string {
				return []string{c.ID, c.Ref.ID, c.Ref.Type, formatBool(c.Ref.Published), c.Ref.LocalizedName.in(lang)}
			},
		},
	}
	return newGroup("assignments", "Read and change the products assigned to a category",
		"Reads the products assigned to a category, assigns products to it and unassigns them.",
		slices.Concat([]*cobra.Command{childCommand(a, categories, assignments)},
			assignmentWrites(a, categories, assignments),
			assignmentDeletes(a, categories, assignments))...)
}

// categoryView is the table of categories.
func categoryView() view[category] {
	return view[category]{
		singular:  "category",
		header:    []string{columnID, columnCode, columnType, columnPublished, columnName},
		localized: true,
		row: func(c category, lang language) []string {
			return []string{c.ID, c.Code, c.Type, formatBool(c.Published), c.displayName(lang)}
		},
	}
}

// subcategoryView is categoryView with the PARENT of each category.
func subcategoryView() view[category] {
	v := categoryView()
	v.header = slices.Insert(v.header, 1, columnParent)
	row := v.row
	v.row = func(c category, lang language) []string { return slices.Insert(row(c, lang), 1, c.ParentID) }
	return v
}

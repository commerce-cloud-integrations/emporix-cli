package cli

import (
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// catalog is the part of a Catalog Service catalog shown in tables.
type catalog struct {
	ID         string    `json:"id"`
	Name       localized `json:"name"`
	Status     string    `json:"status"`
	Visibility struct {
		Visible *bool  `json:"visible"`
		From    string `json:"from"`
		To      string `json:"to"`
	} `json:"visibility"`
	PublishedSites []string `json:"publishedSites"`
	CategoryIDs    []string `json:"categoryIds"`
}

// newCatalogsCommand reads catalogs, visible or not, and the catalogs a
// category is assigned to, and creates and changes catalogs.
func newCatalogsCommand(a *app) *cobra.Command {
	r := resource[catalog]{
		use:   "catalogs",
		short: "Read and change catalogs",
		long: "Reads catalogs, visible or not. STATUS and VISIBLE are separate Emporix fields and can\n" +
			"disagree. ROOT CATEGORIES counts the categories assigned directly to the catalog.",
		path:    "/catalog/{tenant}/catalogs",
		filters: listFilters{kinds: sortable},
		view: view[catalog]{
			singular:  "catalog",
			header:    []string{columnID, columnName, columnStatus, "VISIBLE", columnFrom, columnTo, "SITES", "ROOT CATEGORIES"},
			localized: true,
			row: func(c catalog, lang language) []string {
				return []string{
					c.ID, c.Name.in(lang), c.Status, formatBool(c.Visibility.Visible), c.Visibility.From, c.Visibility.To,
					strings.Join(c.PublishedSites, ","), strconv.Itoa(len(c.CategoryIDs)),
				}
			},
		},
	}
	forCategory := r.reader().command(a, &command{
		use:   "for-category CATEGORY_ID",
		short: "List the catalogs a category is assigned to",
		long: r.long + "\nCATEGORY_ID is the id of a category assigned to catalogs directly, a root category\n" +
			"(an unindented row of `epx category-trees list`); a category below one, or an unknown\n" +
			"id, lists nothing.",
		req:   request{path: r.path + "/categories/" + placeholderID},
		fetch: listing{filters: listFilters{kinds: sortable}},
		ids:   argumentIDs{"CATEGORY_ID": categoryTrees().ids()},
	})
	return newResourceCommand(a, r, slices.Concat([]*cobra.Command{forCategory}, catalogWrites(a, r))...)
}

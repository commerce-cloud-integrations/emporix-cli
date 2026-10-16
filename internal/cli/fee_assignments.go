package cli

import (
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// feeAssignment is the part of a Fee Service item fee shown in tables: the
// fees assigned to one item on a site.
type feeAssignment struct {
	ID       string   `json:"id"`
	ItemYrn  string   `json:"itemYrn"`
	SiteCode string   `json:"siteCode"`
	FeeIDs   []string `json:"feeIds"`
}

// newFeeAssignmentsCommand reads and changes item fees (feeAssignmentWrites),
// which say which fees apply to which item on a site. It reads every one of a
// site (list) or those that the Fee Service's searches find by item or
// product, which Emporix serves as POST.
// The searches by product page in their body, not in their query. Emporix
// requires each member of their bodies (it rejects a body without one with
// 400, seen live), so their flags are required, unless --body is given.
// Tables identify an item fee by its item, which the reference requires, as
// it does not require the id.
func newFeeAssignmentsCommand(a *app) *cobra.Command {
	rd := reader[feeAssignment]{view: view[feeAssignment]{
		singular: "item fee",
		header:   []string{columnItem, columnID, columnSite, "FEES"},
		row: func(f feeAssignment, _ language) []string {
			return []string{f.ItemYrn, f.ID, f.SiteCode, strings.Join(f.FeeIDs, ",")}
		},
		legend: "ITEM is the item's YRN, ID the item fee's id and FEES the ids of its fees, as `epx fees list`\n" +
			"shows them; -o json shows the response as sent.",
	}}
	const (
		long        = "Reads item fees: which fees apply to which item (a product or a payment type) on a site."
		itemFees    = feeService + "/itemFees"
		pagedInBody = "\nEmporix pages this search in its body, so epx sets pageNumber and pageSize there, also in\n" +
			"the body of --body."
	)
	reads := []*cobra.Command{
		rd.command(a, &command{
			use:   listUse,
			short: "List the item fees of a site",
			long: long + "\nEmporix documents main as the default site, so --site is required; an unknown site lists\n" +
				"nothing.",
			req:   request{path: itemFees, query: url.Values{siteCodeParam: {placeholderSite}}},
			fetch: listing{},
		}),
		rd.command(a, &command{
			use:   collectionSearch,
			short: "Search the item fees of items",
			long: long + "\nReads the item fees of the items that --item names. Emporix requires a site, which --site\n" +
				"names in the query and in the body the flags make, but does not filter by it: the item\n" +
				"fees of other sites come too. --item takes the YRN of an item,\n" + yrnNote + ".",
			req: request{path: itemFees + "/" + collectionSearch, query: url.Values{siteCodeParam: {placeholderSite}}},
			// Emporix returns this search whole: it ignores paging in the query and in the body (seen live).
			fetch: listing{read: (*emporix.Client).GetUnpaged}.search(
				listField("itemYrns", "item", "the `YRN` of an item whose fees to read (repeatable)").require(),
				siteField(siteCodeParam),
			),
		}),
		rd.command(a, &command{
			use:   "search-by-product",
			short: "Search the item fees of a product on sites",
			long: long + "\nReads the item fees of the product that --product names on the sites that --site names: its\n" +
				"id (as in `epx products list`), not its code." + pagedInBody,
			req: request{path: itemFees + "/searchByProductId"},
			fetch: listing{read: (*emporix.Client).GetAllBodyPaged}.search(
				textField("productId", "product", "the `ID` of the product whose fees to read").require(),
				siteCodesField("siteCodes", "a site `code` to read them on, e.g. main (repeatable)").require(),
			),
		}),
		rd.command(a, &command{
			use:   "search-by-products",
			short: "Search the item fees of products on a site",
			long: long + "\nReads the item fees of the products that --products names on the site that --site names:\n" +
				"their ids (as in `epx products list`), not their codes, separated by commas." + pagedInBody,
			req: request{path: itemFees + "/searchByProductIds"},
			fetch: listing{read: (*emporix.Client).GetAllBodyPaged}.search(
				textField("productIds", "products", "the `IDs` of the products whose fees to read, separated by "+
					"commas, e.g. 100200,100201").require(),
				siteCodeField(siteCodeParam, "the site `code` to read them on, e.g. main").require(),
			),
		}),
	}
	return newGroup("assignments", "Read and change which items have which fees",
		long+"\nIts writes assign fees to an item, by its YRN, or to a product, by its id.",
		slices.Concat(reads, feeAssignmentWrites(a))...)
}

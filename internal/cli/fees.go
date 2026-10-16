package cli

import (
	"encoding/json"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// feeService is the path of the Fee Service.
const feeService = "/fee/{tenant}"

// yrnNote shows how to give the YRN that names an item, such as a product,
// in the Fee Service: an example, quoted for the shell.
const yrnNote = "e.g. 'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'\n" +
	"(quoted, because the shell would end the command at ;)"

// fee is the part of a Fee Service fee shown in tables.
type fee struct {
	ID             string      `json:"id"`
	Code           string      `json:"code"`
	Name           localized   `json:"name"`
	FeeType        string      `json:"feeType"`
	FeePercentage  json.Number `json:"feePercentage"` // kept as sent, e.g. 2.50
	FeeAbsolute    money       `json:"feeAbsolute"`
	ItemType       string      `json:"itemType"`
	SiteCode       string      `json:"siteCode"`
	Active         *bool       `json:"active"`
	Taxable        *bool       `json:"taxable"`
	ActiveTimespan struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	} `json:"activeTimespan"`
}

// amount returns the value that the fee type uses: a percentage for PERCENT,
// else an amount in its currency. A missing value stays empty.
func (f *fee) amount() string {
	if f.FeeType == "PERCENT" {
		return percent(f.FeePercentage)
	}
	return f.FeeAbsolute.String()
}

// feeItem finds the fee that get reads. The API reference documents it
// wrapped with the items it applies to, {"fee": {...}, "itemYRNs": {...}}; a
// body without a fee member is taken as the fee itself, since Emporix has
// deviated from its reference for single reads before.
func feeItem(body json.RawMessage) ([]json.RawMessage, error) {
	var wrapper map[string]json.RawMessage
	if json.Unmarshal(body, &wrapper) == nil {
		if wrapped, ok := wrapper["fee"]; ok {
			return oneItem(wrapped)
		}
	}
	return oneItem(body)
}

// newFeesCommand reads and changes fees, the fees assigned to a product or
// another item, and which items have which fees (assignments). Fees are
// filtered by a site that Emporix documents as defaulting to main, so the
// site is required rather than left implicit.
func newFeesCommand(a *app) *cobra.Command {
	r := resource[fee]{
		use:   "fees",
		short: "Read and change fees and which items have them",
		about: "Reads and changes fees, surcharges on items such as products or payment types, and in\n" +
			"`assignments` which items have which fees. Emporix filters fees by site and documents main as\n" +
			"the default site, so list, for-product and for-item require --site. create, update and delete\n" +
			"change one fee.",
		long: "Reads fees, surcharges on items such as products or payment types. Emporix filters fees\n" +
			"by site and documents main as the default site, so list, for-product and for-item require\n" +
			"--site; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or\n" +
			"ABSOLUTE_MULTIPLY_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage\n" +
			"or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or\n" +
			"PAYMENTTYPE), and FROM and TO the active time span.",
		path:      feeService + "/fees",
		listQuery: url.Values{siteCodeParam: {placeholderSite}},
		item:      feeItem,
		view: view[fee]{
			singular: "fee",
			header: []string{
				columnID, columnCode, columnName, columnType, "AMOUNT", "ITEM TYPE", columnSite, columnActive, "TAXABLE",
				columnFrom, columnTo,
			},
			localized: true,
			row: func(f fee, lang language) []string {
				return []string{
					f.ID, f.Code, f.Name.in(lang), f.FeeType, f.amount(), f.ItemType, f.SiteCode,
					formatBool(f.Active), formatBool(f.Taxable), f.ActiveTimespan.StartDate, f.ActiveTimespan.EndDate,
				}
			},
		},
	}
	// The fees of one item are not paged; without expand Emporix returns only their ids.
	item, product := feeOwners()
	feesOf := func(o feeOwner) request {
		return request{
			path:  o.fees(placeholderID),
			query: url.Values{siteCodeParam: {placeholderSite}, "expand": {queryTrue}},
		}
	}
	return newResourceCommand(a, r, append([]*cobra.Command{
		r.reader().command(a, &command{
			use:   forProductUse,
			short: "List the fees of a product on a site",
			long:  r.long + productIDNote + "\nAn unknown id lists nothing.",
			req:   feesOf(product),
			fetch: listing{read: (*emporix.Client).GetUnpaged},
			ids:   argumentIDs{productID: products().ids()},
		}),
		r.reader().command(a, &command{
			use:   "for-item YRN",
			short: "List the fees of an item on a site",
			long: r.long + "\nYRN names the item, " + yrnNote + ". An unknown item lists nothing;\n" +
				"Emporix fails on a malformed YRN.",
			req:   feesOf(item),
			fetch: listing{read: (*emporix.Client).GetUnpaged},
		}),
		newFeeAssignmentsCommand(a),
	}, feeWrites(a, r)...)...)
}

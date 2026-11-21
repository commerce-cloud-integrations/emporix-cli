package cli

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const feeService = "/fee/{tenant}"

const yrnNote = "e.g. 'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID' " +
	"(quoted, because the shell would end the command at ;)"

type fee struct {
	ID             string      `json:"id"`
	Code           string      `json:"code"`
	Name           localized   `json:"name"`
	FeeType        string      `json:"feeType"`
	FeePercentage  json.Number `json:"feePercentage"`
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

func (f *fee) amount() string {
	if f.FeeType == "PERCENT" {
		return percent(f.FeePercentage)
	}
	return f.FeeAbsolute.String()
}

func feeItem(body json.RawMessage) ([]json.RawMessage, error) {
	var wrapper map[string]json.RawMessage
	if json.Unmarshal(body, &wrapper) != nil || wrapper["fee"] == nil {
		return nil, errors.New("expected the fee in the member fee of a JSON object")
	}
	return oneItem(wrapper["fee"])
}

func newFeesCommand(a *app) *cobra.Command {
	r := resource[fee]{
		use:   "fees",
		short: "Read and change fees and which items have them",
		about: "Reads and changes fees, surcharges on items such as products or payment types, and in " +
			"`assignments` which items have which fees. Emporix filters fees by site and documents main as " +
			"the default site, so list, for-product and for-item require --site. create, update and delete " +
			"change one fee.",
		long: "Reads fees, surcharges on items such as products or payment types. Emporix filters fees " +
			"by site and documents main as the default site, so list, for-product and for-item require " +
			"--site; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or " +
			"ABSOLUTE_MULTIPLY_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage " +
			"or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or " +
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
			long:  joinSentences(r.long, productIDNote, "An unknown id lists nothing."),
			req:   feesOf(product),
			fetch: listing{read: (*emporix.Client).GetUnpaged},
			ids:   argumentIDs{productID: products().ids()},
		}),
		r.reader().command(a, &command{
			use:   "for-item YRN",
			short: "List the fees of an item on a site",
			long: r.long + " YRN names the item, " + yrnNote + ". An unknown item lists nothing; " +
				"Emporix fails on a malformed YRN.",
			req:   feesOf(item),
			fetch: listing{read: (*emporix.Client).GetUnpaged},
		}),
		newFeeAssignmentsCommand(a),
	}, feeWrites(a, r)...)...)
}

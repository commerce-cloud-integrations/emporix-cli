package cli

import (
	"cmp"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// priceMatch is the part of a price that the Price Service matches for an
// item shown in tables.
type priceMatch struct {
	PriceID     string    `json:"priceId"`
	PriceListID string    `json:"priceListId"`
	ItemID      priceItem `json:"itemId"`  // as Emporix sends it and the API reference's example names it
	ItemRef     priceItem `json:"itemRef"` // as the API reference's schema names it
	Site        struct {
		Code string `json:"code"`
	} `json:"site"`
	Currency string `json:"currency"`
	Location struct {
		CountryCode string `json:"countryCode"`
	} `json:"location"`
	Quantity struct {
		Quantity json.Number `json:"quantity"` // kept as sent
		UnitCode string      `json:"unitCode"`
	} `json:"quantity"`
	OriginalValue  json.Number `json:"originalValue"` // kept as sent, e.g. 0.0000317
	EffectiveValue json.Number `json:"effectiveValue"`
	TotalValue     json.Number `json:"totalValue"`
	IncludesTax    *bool       `json:"includesTax"`
}

// in returns an amount of the match's currency.
func (m *priceMatch) in(amount json.Number) string {
	return money{Amount: amount, Currency: m.Currency}.String()
}

// priceMatchCommand returns `match`, which matches the best price of each
// product that --product names, as Emporix would price one unit for a
// customer of the site and country in the currency, in one response.
func priceMatchCommand(a *app) *cobra.Command {
	return reader[priceMatch]{view: priceMatchView(), scopes: []string{priceReadScope}}.command(a, &command{
		use:   "match",
		short: "Match the best price of products",
		long: "Matches the price of one unit of each product that --product names (the product id as in\n" +
			"`epx products list`) that Emporix would charge on the site --site names, for the country\n" +
			"--country names, in the currency --currency names: of the prices that apply, the lowest\n" +
			"after discounts, converted with the tenant's exchange rates and tax classes where the price\n" +
			"is in another currency or for another country. ID is the matched price (of the price list\n" +
			"PRICE LIST, if any), ORIGINAL and EFFECTIVE (discounted) the price of one unit, TOTAL that\n" +
			"of QUANTITY, gross where TAX is true. Other quantities and units, a customer or a legal\n" +
			"entity go in --body.",
		req: request{path: "/price/{tenant}/match-prices"},
		fetch: bodyRead{method: http.MethodPost, fetch: single{items: priceMatches}, body: objectBody{
			matchItemsField().require(),
			stringField("siteCode", siteFlag, "the `code` of the site whose price to match").require().siteDefault(),
			matchLocationField().require(),
			stringField("targetCurrency", "currency", "the `code` of the currency to match the price in, "+
				"e.g. EUR").require(),
		}},
	})
}

// priceMatches returns the matched prices of a response: a JSON array, as
// Emporix answers, or one JSON object, as the API reference documents it.
func priceMatches(body json.RawMessage) ([]json.RawMessage, error) {
	if items, err := listItems(body); err == nil {
		return items, nil
	}
	if item, err := oneItem(body); err == nil {
		return item, nil
	}
	return nil, errors.New("expected a JSON array, or one price as a JSON object")
}

// priceMatchView is the table of matched prices.
func priceMatchView() view[priceMatch] {
	return view[priceMatch]{
		singular: "price match",
		header: []string{
			columnID, columnItem, "PRICE LIST", columnSite, columnCountry, "QUANTITY", "ORIGINAL", "EFFECTIVE",
			columnTotal, "TAX",
		},
		row: func(m priceMatch, _ language) []string {
			return []string{
				m.PriceID, cmp.Or(m.ItemID.ID, m.ItemRef.ID), m.PriceListID, m.Site.Code, m.Location.CountryCode,
				spaced(m.Quantity.Quantity.String(), m.Quantity.UnitCode), m.in(m.OriginalValue),
				m.in(m.EffectiveValue), m.in(m.TotalValue), formatBool(m.IncludesTax),
			}
		},
	}
}

// matchItem is an item whose price to match: one unit of a product.
type matchItem struct {
	ItemID   priceItem `json:"itemId"`
	Quantity struct {
		Quantity int `json:"quantity"`
	} `json:"quantity"`
}

// matchLocation is the location whose price to match: a country.
type matchLocation struct {
	CountryCode string `json:"countryCode"`
}

// matchLocationField is the member targetLocation of a price match, the
// country that --country names.
func matchLocationField() bodyField {
	const flag = "country"
	return flagField("targetLocation", flag, func(fs *pflag.FlagSet) func() any {
		code := fs.String(flag, "", "the `code` of the country to match the price for, e.g. NL")
		return func() any { return matchLocation{CountryCode: *code} }
	})
}

// matchItemsField is the member items of a price match: one unit of each
// product that --product names, in their order.
func matchItemsField() bodyField {
	const flag = "product"
	return flagField("items", flag, func(fs *pflag.FlagSet) func() any {
		ids := fs.StringArray(flag, nil, "a product `ID` whose price to match (repeatable)")
		return func() any {
			items := make([]matchItem, len(*ids))
			for i, id := range *ids {
				items[i].ItemID = priceItem{ItemType: productItemType, ID: id}
				items[i].Quantity.Quantity = 1
			}
			return items
		}
	})
}

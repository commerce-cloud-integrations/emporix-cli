package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// price is the part of a Price Service price shown in tables.
type price struct {
	ID       string    `json:"id"`
	ItemID   priceItem `json:"itemId"`
	Currency string    `json:"currency"`
	Location struct {
		CountryCode string `json:"countryCode"`
	} `json:"location"`
	Restrictions struct {
		SiteCodes []string `json:"siteCodes"`
		Validity  validity `json:"validity"`
	} `json:"restrictions"`
	TierValues priceTiers `json:"tierValues"`
}

// priceItem is the item a price is for, such as a product.
type priceItem struct {
	ItemType string `json:"itemType"`
	ID       string `json:"id"`
}

// productItemType is the item type of a product, as a price's item and a
// segment's item name it.
const productItemType = "PRODUCT"

// priceTiers are the values of a price's tiers, in the order sent; each
// refers to a tier of the price model by an id that only -o json shows.
type priceTiers []priceTier

// priceTier is the value of one price tier.
type priceTier struct {
	PriceValue json.Number `json:"priceValue"` // kept as sent, e.g. 15.90
}

func (t priceTier) String() string { return trimAmount(t.PriceValue.String()) }

// String joins the tier values with commas.
func (t priceTiers) String() string { return joinEach(t, ",", priceTier.String) }

// newPricesCommand reads and changes prices. Lists are filtered by site: the
// Price Service, although it documents siteCode with the default main, lists
// the prices of every site without it, so --site is required rather than
// leaving the filter implicit.
func newPricesCommand(a *app) *cobra.Command {
	r := resource[price]{
		use:   "prices",
		short: "Read and change prices",
		about: "Reads and changes prices. Without a site, Emporix lists the prices of every site, so list\n" +
			"and for-product require --site and read that site's prices only; an unknown site lists\n" +
			"nothing. create and update change one price, bulk-create and bulk-update many, each from\n" +
			"--file, and delete deletes one.",
		long: "Reads prices. Without a site, Emporix lists the prices of every site, so list and\n" +
			"for-product require --site and read that site's prices only; an unknown site lists\n" +
			"nothing. TIERS lists the value of each price tier as sent; tier quantities and net or\n" +
			"gross come from the price model (`epx price-models get` with the priceModelId of -o json).\n" +
			"Customer and price-list restrictions and sale prices are only in -o json.",
		path:      "/price/{tenant}/prices",
		listQuery: url.Values{siteCodeParam: {placeholderSite}},
		filters:   listFilters{kinds: filterable},
		view: view[price]{
			singular: "price",
			header: []string{
				columnID, columnItem, columnType, columnCurrency, columnCountry, "SITES", columnTiers, columnFrom, columnTo,
			},
			row: func(p price, _ language) []string {
				return []string{
					p.ID, p.ItemID.ID, p.ItemID.ItemType, p.Currency, p.Location.CountryCode,
					strings.Join(p.Restrictions.SiteCodes, ","), p.TierValues.String(),
					p.Restrictions.Validity.From, p.Restrictions.Validity.To,
				}
			},
		},
	}
	ofProduct := url.Values{siteCodeParam: {placeholderSite}, "itemId": {placeholderID}, "itemType": {productItemType}}
	reads := []*cobra.Command{
		r.reader().command(a, &command{
			use:   forProductUse,
			short: "List the prices of a product on a site",
			long:  r.long + productIDNote + "\nAn unknown id lists nothing.",
			req:   request{path: r.path, query: ofProduct},
			fetch: listing{filters: r.filters},
			ids:   argumentIDs{productID: products().ids()},
		}),
		priceSearchCommand(a, r),
		priceMatchCommand(a),
	}
	return newResourceCommand(a, r, append(reads, priceWrites(a, r)...)...)
}

// tierValuesHelp says, in the help of a create of the Price Service, what
// the tierValues of a price hold: Emporix requires each tier's id where
// priceModelId names a price model.
const tierValuesHelp = "tierValues: the id of each tier of the price model that priceModelId names, as\n" +
	"`epx price-models get ID -o json` prints it, with a priceValue"

// staleBulkItemHelp ends the help of a bulk-update of the Price Service,
// which answers an item whose metadata.version is stale with code 500, where
// its update of one item answers 409 Conflict.
const staleBulkItemHelp = "\nAn item that failed with 500 may be one whose metadata.version is stale: Emporix\n" +
	"reports it so, not as 409 Conflict. Read it again before you retry it."

// priceManageScope is the scope of the Price Service's writes of prices (the
// *_by_vendor scopes are those of a vendor's own prices).
const priceManageScope = "price.price_manage"

// priceWrites returns the writes of prices: create, update (an upsert), both
// in bulk, and delete.
func priceWrites(a *app, prices resource[price]) []*cobra.Command {
	scopes := []string{priceManageScope}
	bulk := &bulkWrites{
		path: prices.path, nouns: prices.nouns(), scopes: scopes, maxItems: 200,
		create: "Creates the prices in --file, a JSON array of prices, each as create takes it.",
		update: "Replaces the prices in --file, a JSON array of prices, each with its id, or creates\n" +
			"those that do not exist (an upsert)." + staleBulkItemHelp,
	}
	return append([]*cobra.Command{
		(&write{
			use: createUse, short: "Create a price", method: http.MethodPost,
			long: "Creates the price in --file, a JSON object with at least itemId, currency, location and\n" +
				tierValuesHelp + ". Without an id, Emporix\n" +
				"makes one, which the response gives. Emporix does not refuse an id that a price has\n" +
				"already: it stores a second price with it, and each delete of that id removes one.\n" +
				"Leave id out, or change a price with update.",
			req: request{path: prices.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: "update PRICE_ID", short: "Upsert a price", method: http.MethodPut,
			long: "Replaces the price PRICE_ID with the one in --file, a JSON object as\n" +
				"`epx prices get PRICE_ID -o json` prints it, or creates it with that id where none exists\n" +
				"(an upsert; not with --partial: Emporix answers a missing price with 404 Not Found, exit\n" +
				"4). Send the price's metadata.version as read to change it only if nobody changed it\n" +
				"since: Emporix answers 409 Conflict otherwise.\n" + prices.argumentSource() + ".",
			req: request{path: prices.itemPath()}, scopes: scopes, body: &payload{},
			switches: []paramFlag{partialSwitch("price")},
			reserved: bulk.reserved(http.MethodPut, prices.itemPath()),
		}).command(a),
		deleteWrite("delete PRICE_ID", "Delete a price", "Deletes the price PRICE_ID."+unknownAsDeleted("PRICE_ID")+
			"\n"+prices.argumentSource()+".",
			request{path: prices.itemPath()}, scopes).command(a),
	}, bulk.commands(a)...)
}

// priceSearchCommand returns `search`, which reads the prices of the products
// --product names, in one response, shown as list shows prices.
func priceSearchCommand(a *app, prices resource[price]) *cobra.Command {
	return reader[price]{view: prices.view, scopes: []string{priceReadScope}}.command(a, &command{
		use:   collectionSearch,
		short: "Search the prices of products",
		long: "Reads the prices of up to 100 products (--product, the product id as in `epx products\n" +
			"list`), on the site and in the currency that --site and --currency name, else on every\n" +
			"site and in every currency, shown as list shows prices; -o json also lists the products\n" +
			"without a price. Other criteria (a customer, a date, net or gross values) go in --body.",
		req: request{path: prices.path + "/" + collectionSearch},
		fetch: bodyRead{method: http.MethodPost, fetch: single{items: pricesOfProducts}, body: objectBody{
			listField("itemIds", "product", "a product `ID` whose prices to read (repeatable, up to 100)").require(),
			stringField("siteCode", "site", "the `code` of the site whose prices to read"),
			stringField("currency", "currency", "the `code` of the currency whose prices to read, e.g. EUR"),
		}},
	})
}

// priceReadScope is the scope of the Price Service's price search and price
// matching (the price reads need none).
const priceReadScope = "price.price_read"

// pricesOfProducts returns the prices of a price search's response, which
// lists each product's prices ({"itemId": …, "prices": […]}), in the order sent.
func pricesOfProducts(body json.RawMessage) ([]json.RawMessage, error) {
	var products []struct {
		Prices *[]json.RawMessage `json:"prices"`
	}
	if err := json.Unmarshal(body, &products); err != nil || products == nil {
		return nil, errors.New("expected a JSON array of products with their prices")
	}
	var prices []json.RawMessage
	for i, product := range products {
		if product.Prices == nil {
			return nil, fmt.Errorf("product %d has no prices array", i+1)
		}
		prices = append(prices, *product.Prices...)
	}
	return prices, nil
}

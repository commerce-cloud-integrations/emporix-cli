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

type priceItem struct {
	ItemType string `json:"itemType"`
	ID       string `json:"id"`
}

const productItemType = "PRODUCT"

type priceTiers []priceTier

type priceTier struct {
	PriceValue json.Number `json:"priceValue"`
}

func (t priceTier) String() string { return trimAmount(t.PriceValue.String()) }

func (t priceTiers) String() string { return joinEach(t, ",", priceTier.String) }

func newPricesCommand(a *app) *cobra.Command {
	r := resource[price]{
		use:   "prices",
		short: "Read and change prices",
		about: "Reads and changes prices. Without a site, Emporix lists the prices of every site, so list " +
			"and for-product require --site and read that site's prices only; an unknown site lists " +
			"nothing. create and update change one price, bulk-create and bulk-update many, each from " +
			"--file, and delete deletes one.",
		long: "Reads prices. Without a site, Emporix lists the prices of every site, so list and " +
			"for-product require --site and read that site's prices only; an unknown site lists " +
			"nothing. TIERS lists the value of each price tier as sent; tier quantities and net or " +
			"gross come from the price model (`epx price-models get` with the priceModelId of `-o json`). " +
			"Customer and price-list restrictions and sale prices are only in `-o json`.",
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
			long:  joinSentences(r.long, productIDNote, "An unknown id lists nothing."),
			req:   request{path: r.path, query: ofProduct},
			fetch: listing{filters: r.filters},
			ids:   argumentIDs{productID: products().ids()},
		}),
		priceSearchCommand(a, r),
		priceMatchCommand(a),
	}
	return newResourceCommand(a, r, append(reads, priceWrites(a, r)...)...)
}

const tierValuesHelp = "tierValues: the id of each tier of the price model that priceModelId names, as " +
	"`epx price-models get ID -o json` prints it, with a priceValue"

const staleBulkItemHelp = "An item that failed with 500 may be one whose metadata.version is stale: Emporix " +
	"reports it so, not as 409 Conflict. Read it again before you retry it."

const priceManageScope = "price.price_manage"

func priceWrites(a *app, prices resource[price]) []*cobra.Command {
	scopes := []string{priceManageScope}
	bulk := &bulkWrites{
		path: prices.path, nouns: prices.nouns(), scopes: scopes, maxItems: 200,
		create: "Creates the prices in --file, a JSON array of prices, each as create takes it.",
		update: "Replaces the prices in --file, a JSON array of prices, each with its id, or creates " +
			"those that do not exist (an upsert). " + staleBulkItemHelp,
	}
	return append([]*cobra.Command{
		(&write{
			use: createUse, short: "Create a price", method: http.MethodPost,
			long: "Creates the price in --file, a JSON object with at least itemId, currency, location and " +
				tierValuesHelp + ". Without an id, Emporix " +
				"makes one, which the response gives. Emporix does not refuse an id that a price has " +
				"already: it stores a second price with it, and each delete of that id removes one. " +
				"Leave id out, or change a price with update.",
			req: request{path: prices.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: "update PRICE_ID", short: "Upsert a price", method: http.MethodPut,
			long: "Replaces the price PRICE_ID with the one in --file, a JSON object as " +
				"`epx prices get PRICE_ID -o json` prints it, or creates it with that id where none exists " +
				"(an upsert; not with --partial: Emporix answers a missing price with 404 Not Found, exit " +
				"4). Send the price's metadata.version as read to change it only if nobody changed it " +
				"since: Emporix answers 409 Conflict otherwise. " + prices.argumentSource() + ".",
			req: request{path: prices.itemPath()}, scopes: scopes, body: &payload{},
			switches: []paramFlag{partialSwitch("price")},
			reserved: bulk.reserved(http.MethodPut, prices.itemPath()),
		}).command(a),
		deleteWrite("delete PRICE_ID", "Delete a price", "Deletes the price PRICE_ID. "+unknownAsDeleted("PRICE_ID")+
			" "+prices.argumentSource()+".",
			request{path: prices.itemPath()}, scopes).command(a),
	}, bulk.commands(a)...)
}

func priceSearchCommand(a *app, prices resource[price]) *cobra.Command {
	return reader[price]{view: prices.view, scopes: []string{priceReadScope}}.command(a, &command{
		use:   collectionSearch,
		short: "Search the prices of products",
		long: "Reads the prices of up to 100 products (--product, the product id as in `epx products " +
			"list`), on the site and in the currency that --site and --currency name, else on every " +
			"site and in every currency, shown as list shows prices; `-o json` also lists the products " +
			"without a price. Other criteria (a customer, a date, net or gross values) go in --body.",
		req: request{path: prices.path + "/" + collectionSearch},
		fetch: bodyRead{method: http.MethodPost, fetch: single{items: pricesOfProducts}, body: objectBody{
			listField("itemIds", "product", "a product `ID` whose prices to read (repeatable, up to 100)").require(),
			stringField("siteCode", "site", "the `code` of the site whose prices to read"),
			stringField("currency", "currency", "the `code` of the currency whose prices to read, e.g. EUR"),
		}},
	})
}

const priceReadScope = "price.price_read"

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

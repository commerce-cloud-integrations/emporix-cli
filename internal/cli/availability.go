package cli

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
)

// productAvailability is the part of an Availability Service record shown in tables.
type productAvailability struct {
	ProductID           string      `json:"productId"`
	Site                string      `json:"site"`
	StockLevel          json.Number `json:"stockLevel"` // kept as sent
	Available           *bool       `json:"available"`
	DistributionChannel string      `json:"distributionChannel"`
}

// availabilityService is the path of the Availability Service, whose records
// are addressed by product and site.
const availabilityService = "/availability/{tenant}/availability"

// availabilityManageScope is the scope of the Availability Service's writes
// (the *_by_vendor scope is that of a vendor's own products).
const availabilityManageScope = "availability.availability_manage"

// newAvailabilityCommand reads and changes stock per site. A record is
// addressed by site and product, so get, create and update take a product id
// and require --site; a search reads the records of every site, and a bulk
// write names the product and site of each record.
func newAvailabilityCommand(a *app) *cobra.Command {
	const onSite = "Reads product availability (stock) on the site --site names, the tenant's own\navailability."
	stock := reader[productAvailability]{view: view[productAvailability]{
		singular: "availability",
		header:   []string{columnProduct, columnSite, "STOCK", "AVAILABLE", "CHANNEL"},
		row: func(p productAvailability, _ language) []string {
			return []string{p.ProductID, p.Site, p.StockLevel.String(), formatBool(p.Available), p.DistributionChannel}
		},
	}}
	records := listing{filters: listFilters{kinds: filterable}}
	onSiteList := request{path: availabilityService + "/site/" + placeholderSite}
	listed := argumentIDs{productID: listedIDs(stock, onSiteList, records, columnProduct)} // those with a record on --site
	commands := []*cobra.Command{
		stock.command(a, &command{
			use:   listUse,
			short: "List the availability of every product on a site",
			long:  onSite + "\nA bundle is listed as a whole and per product; an unknown site lists nothing.",
			req:   onSiteList,
			fetch: records,
		}),
		stock.command(a, &command{
			use:   "get PRODUCT_ID",
			short: "Show the availability of a product on a site",
			long: onSite + productIDNote + "\nThe availability of a bundle's products is only in -o json " +
				"(bundleAvailabilities).",
			req:   request{path: availabilityService + "/" + placeholderID + "/" + placeholderSite},
			fetch: single{items: oneItem},
			ids:   listed,
		}),
		stock.command(a, &command{
			use:   collectionSearch,
			short: "Search the availability of products",
			long: "Reads the availability records, on every site, that an Emporix q filter finds (--query;\n" +
				"without it, every record), shown as list shows them, bundles left out. --body with a JSON\n" +
				"array of product ids reads theirs instead, bundles included.",
			req:   request{path: availabilityService + "/" + collectionSearch},
			fetch: records.search(queryField()),
		}),
	}
	return newGroup("availability", "Read and change availability",
		"Reads product availability (stock), the tenant's own availability: on a site with list and\n"+
			"get, which require --site, and on every site with search. create, update and delete change\n"+
			"the availability of a product on a site, bulk-create, bulk-update and bulk-delete that of\n"+
			"many.",
		append(commands, availabilityWrites(a, listed)...)...)
}

// availabilityWrites returns the writes of availability: create, whose
// PRODUCT_ID completes with every product, update (an upsert) and delete,
// whose PRODUCT_ID completes from listed, and each in bulk.
func availabilityWrites(a *app, listed argumentIDs) []*cobra.Command {
	one := request{path: availabilityService + "/" + placeholderID + "/" + placeholderSite}
	scopes := []string{availabilityManageScope}
	const bundle = "The availability of a bundle is that of its scarcest product and cannot be "
	del := deleteWrite("delete PRODUCT_ID", "Delete the availability of a product on a site",
		"Deletes the availability of the product PRODUCT_ID on the site --site names.\n"+bundle+"deleted."+
			productIDNote, one, scopes)
	del.ids = listed
	return append([]*cobra.Command{
		(&write{
			use: "create PRODUCT_ID", short: "Create the availability of a product on a site", method: http.MethodPost,
			long: "Creates the availability of the product PRODUCT_ID on the site --site names, from --file:\n" +
				"a JSON object with stockLevel, available, popularity and distributionChannel (ASSORTMENT,\n" +
				"HOME_DELIVERY or PICKUP), as `epx availability get PRODUCT_ID --site S -o json` prints\n" +
				"them. Emporix answers 409 Conflict when the product has one on that site already. It\n" +
				"advises against a record on a site named main, which aggregates the availability of the\n" +
				"other sites." + productIDNote,
			req: one, scopes: scopes, body: &payload{}, ids: argumentIDs{productID: products().ids()},
		}).command(a),
		(&write{
			use: "update PRODUCT_ID", short: "Upsert the availability of a product on a site", method: http.MethodPut,
			long: "Replaces the availability of the product PRODUCT_ID on the site --site names with the one\n" +
				"in --file, a JSON object as create takes it, or creates it where there is none (an\n" +
				"upsert). " + bundle + "set." + productIDNote,
			req: one, scopes: scopes, body: &payload{}, ids: listed,
		}).command(a),
		del.command(a),
	}, (&bulkWrites{
		path: availabilityService, nouns: "availability", scopes: scopes, maxItems: 200,
		create: "Creates the availability records in --file, a JSON array of records, each with the\n" +
			"productId and site it is for and the fields that create takes. Emporix advises against a\n" +
			"record on a site named main, which aggregates the availability of the other sites.",
		update: "Replaces the availability records in --file, a JSON array of records as bulk-create\n" +
			"takes them, or creates those that do not exist (an upsert).",
		delete: "Deletes the availability records that --file names, a JSON array of objects with the\n" +
			`productId and site of each, such as [{"productId": "100200", "site": "AcmeStore_NL"}].` + "\n" +
			bundle + "deleted.",
	}).commands(a)...)
}

package cli

import (
	"github.com/spf13/cobra"
)

// priceListManage is the scope of every write of price lists and of their
// prices.
const priceListManage = "price.pricelist_manage"

// priceListLanguageHelp follows what a write of a price list sends, in its
// help: the Price Service takes its localized text as a map of translations
// also with a Content-Language that names one language, a map of that one
// language then (its reference's header_contentLanguage).
const priceListLanguageHelp = "Localized text in the body, such as a name, is a map of translations, e.g.\n" +
	`{"en": "Wholesale", "nl": "Groothandel"}, as -o json of a read prints it, also with --` +
	contentLanguageFlag + "\nLANGUAGE: then a map of that language only, e.g. " + `{"en": "Wholesale"}` + " with --" +
	contentLanguageFlag + " en."

// priceListWrites returns the writes of price lists: create, update ID (an
// upsert) and delete ID, which deletes the prices of the list too.
func priceListWrites(a *app, lists resource[priceList]) []*cobra.Command {
	at := lists.at()
	create, update, remove := (&collectionWrites{
		at: at, one: "a price list", scopes: []string{priceListManage}, language: stringsOrMaps,
		languageHelp: priceListLanguageHelp, upsert: true,
		create: "Creates the price list in --file, a JSON object with name, currency, siteCode and validity\n" +
			"(from and to), and optionally countries or regions, and customerGroups or legalEntityId,\n" +
			"e.g. " + `{"name": {"en": "Wholesale"}, "currency": "EUR", "siteCode": "AcmeStore_NL",` + "\n" +
			`"countries": ["NL"], "validity": {"from": "2026-01-01T00:00:00.000Z"}}` + ". Without an id, Emporix\n" +
			"makes one, which the response gives.",
		update: "Replaces the price list ID with the one in --file, a JSON object as\n" +
			"`epx " + at.get + " -o json` prints it, or creates it with that id where none exists\n" +
			"(an upsert).",
		remove: "Deletes the price list ID and the prices in it, which Emporix deletes after it, in the\n" +
			"background." + unknownAsDeleted("ID"),
	}).writes()
	remove.deletes = deletion{many: "the price list and every price in it"}
	return writeCommands(a, []write{*create, *update, *remove})
}

// priceListPriceWrites returns the writes of the prices of a price list,
// prices being their collection below pricesOf (newPriceListPricesCommand):
// create, update ID (an upsert) and delete ID, and in bulk bulk-create,
// bulk-update (an upsert) and bulk-delete, each below the price list
// PRICE_LIST_ID.
func priceListPriceWrites(a *app, pricesOf resource[priceList], prices child[priceListPrice]) []*cobra.Command {
	at := childAt(pricesOf, prices)
	of := "\n" + pricesOf.argumentSource() + "."
	id := "\nID is the ID that `epx " + childListLine(pricesOf, prices) + "` shows."
	const shape = "a JSON object\nwith itemId (itemType and id) and\n" + tierValuesHelp
	return (&collectionWrites{
		at: at, one: "a price of a price list", scopes: []string{priceListManage}, upsert: true,
		create: "Adds the price in --file to the price list PRICE_LIST_ID, in its currency: " + shape + ", e.g.\n" +
			`{"itemId": {"itemType": "PRODUCT", "id": "100200"}, "priceModelId": "pm1",` + "\n" +
			`"tierValues": [{"id": "t1", "priceValue": 15.90}]}` + ".\nWithout an id, Emporix makes one, which the " +
			"response gives." + of,
		update: "Replaces the price ID of the price list PRICE_LIST_ID with the one in --file, a JSON object\n" +
			"as `epx " + at.get + " -o json` prints it, or creates it with\nthat id where none exists " +
			"(an upsert)." + of + id,
		remove: "Deletes the price ID of the price list PRICE_LIST_ID." + unknownAsDeleted("ID") + of + id,
		bulk: &bulkWrites{
			nouns: "prices of a price list", maxItems: 200, deleteWhole: true,
			create: "Adds the prices in --file to the price list PRICE_LIST_ID: a JSON array of prices, each\n" +
				"as create takes it." + of,
			update: "Replaces the prices in --file in the price list PRICE_LIST_ID, a JSON array of prices,\n" +
				"each with its id, or creates those that do not exist (an upsert)." + staleBulkItemHelp + of,
			delete: "Deletes the prices of the price list PRICE_LIST_ID whose ids --file gives, a JSON array of\n" +
				`ids, e.g. ["p1", "p2"].` + of,
		},
	}).commands(a)
}

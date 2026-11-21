package cli

import (
	"github.com/spf13/cobra"
)

const priceListManage = "price.pricelist_manage"

func priceListLanguageHelp() string {
	return "Localized text in the body, such as a name, is a map of translations, e.g. " +
		literal(`{"en": "Wholesale", "nl": "Groothandel"}`) + ", as `-o json` of a read prints it, also with --" +
		contentLanguageFlag + " LANGUAGE: then a map of that language only, its key written as --" + contentLanguageFlag +
		" gives it, e.g. " +
		literal(`{"en": "Wholesale"}`) + " with --" +
		contentLanguageFlag + " en."
}

func priceListWrites(a *app, lists resource[priceList]) []*cobra.Command {
	at := lists.at()
	create, update, remove := (&collectionWrites{
		at: at, one: "a price list", scopes: []string{priceListManage}, language: stringsOrMaps,
		languageHelp: priceListLanguageHelp(), upsert: true,
		create: "Creates the price list in --file, a JSON object with name, currency, siteCode and validity " +
			"(from and to), and optionally countries or regions, and customerGroups or legalEntityId, " +
			"e.g. " + literal(`{"name": {"en": "Wholesale"}, "currency": "EUR", "siteCode": "AcmeStore_NL", `+
			`"countries": ["NL"], "validity": {"from": "2026-01-01T00:00:00.000Z"}}`) + ". Without an id, Emporix " +
			"makes one, which the response gives.",
		update: "Replaces the price list ID with the one in --file, a JSON object as " +
			"`epx " + at.get + " -o json` prints it, or creates it with that id where none exists " +
			"(an upsert).",
		remove: "Deletes the price list ID and the prices in it, which Emporix deletes after it, in the " +
			"background. " + unknownAsDeleted("ID"),
	}).writes()
	remove.deletes = deletion{many: "the price list and every price in it"}
	return writeCommands(a, []write{*create, *update, *remove})
}

func priceListPriceWrites(a *app, pricesOf resource[priceList], prices child[priceListPrice]) []*cobra.Command {
	at := childAt(pricesOf, prices)
	of := pricesOf.argumentSource() + "."
	id := "ID is the ID that `epx " + childListLine(pricesOf, prices) + "` shows."
	const shape = "a JSON object with itemId (itemType and id) and " + tierValuesHelp
	return (&collectionWrites{
		at: at, one: "a price of a price list", scopes: []string{priceListManage}, upsert: true,
		create: "Adds the price in --file to the price list PRICE_LIST_ID, in its currency: " + shape + ", e.g. " + literal(
			`{"itemId": {"itemType": "PRODUCT", "id": "100200"}, "priceModelId": "pm1", `+
				`"tierValues": [{"id": "t1", "priceValue": 15.90}]}`) + ". Without an id, Emporix makes one, which the " +
			"response gives. " + of,
		update: "Replaces the price ID of the price list PRICE_LIST_ID with the one in --file, a JSON object " +
			"as `epx " + at.get + " -o json` prints it, or creates it with that id where none exists " +
			"(an upsert). " + of + " " + id,
		remove: "Deletes the price ID of the price list PRICE_LIST_ID. " + unknownAsDeleted("ID") + " " + of + " " + id,
		bulk: &bulkWrites{
			nouns: "prices of a price list", maxItems: 200, deleteWhole: true,
			create: "Adds the prices in --file to the price list PRICE_LIST_ID: a JSON array of prices, each " +
				"as create takes it. " + of,
			update: "Replaces the prices in --file in the price list PRICE_LIST_ID, a JSON array of prices, " +
				"each with its id, or creates those that do not exist (an upsert). " + staleBulkItemHelp + " " + of,
			delete: "Deletes the prices of the price list PRICE_LIST_ID whose ids --file gives, a JSON array of " +
				"ids, e.g. " + literal(`["p1", "p2"]`) + "." + " " + of,
		},
	}).commands(a)
}

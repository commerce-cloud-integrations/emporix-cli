package cli

import (
	"net/http"
)

// quoteManageScope is the scope of the Quote Service's writes made by an
// employee (quote.quote_manage_own is a customer's, with the customer's own
// token).
const quoteManageScope = "quote.quote_manage"

// quoteExternalPrices is what --allow-external-prices allows on quotes
// create (externalPrices).
const quoteExternalPrices = "a body with external prices (price type EXTERNAL)"

// quoteWrites returns the writes of quotes: create, patch ID and delete ID.
// Only create takes --allow-external-prices: Emporix refuses an external item
// price in a patch with that scope too (live).
func quoteWrites(quotes resource[quote]) []*write {
	item := request{path: quotes.itemPath()}
	manage := []string{quoteManageScope}
	return []*write{
		{
			use: createUse, short: "Create a quote",
			long: "Creates a quote from the JSON object in --file: a whole quote, with customerId (or\n" +
				"customerEmail), siteCode, currency, billingAddressId, shippingAddressId and items (each\n" +
				"with its product, price and quantity), or a cart's, with its cartId and optionally the\n" +
				"addresses and shipping. Without an id, Emporix makes one; it answers with the quote's id.\n" +
				"A quote made from a cart closes the cart (status CLOSED, with the quoteId): Emporix then\n" +
				"answers its reads and its delete as not found (exit 4), while `epx carts search` lists it.\n" +
				"Without --allow-external-prices, Emporix refuses an item with price type EXTERNAL with 400\n" +
				"Bad Request (exit 1). It fails with 500 Internal Server Error (exit 1) on a cart without\n" +
				"items, and answers a product that is not published as not found (exit 4).",
			method: http.MethodPost, req: request{path: quotes.path}, scopes: manage, body: &payload{},
			choice: externalPrices(quoteExternalPrices),
		},
		{
			use: quotes.itemUse("patch"), short: "Change a quote",
			long: "Changes the quote ID with the operations in --file: a JSON array of operations, each\n" +
				"with op (ADD, REMOVE or REPLACE), path (such as /status, /validTo, /items or\n" +
				"/items/ITEM_ID) and value, e.g.\n" +
				`[{"op": "REPLACE", "path": "/status", "value": {"value": "OPEN"}}]. Emporix completes` +
				"\nthe prices and product names of items from their ids afterwards. Emporix refuses to add\n" +
				"or replace an item with price type EXTERNAL with 400 Bad Request (exit 1), also with\n" +
				"cart.cart_manage_external_prices (create accepts one).\n" + quotes.idSource() + ".",
			method: http.MethodPatch, req: item, scopes: manage, body: operationsFile(), ids: quotes.itemIDs(),
		},
		quotes.deleteWrite(manage, "Deletes the quote ID; it sends no body."+unknownAsDeleted("ID")+"\n"+
			quotes.idSource()+"."),
	}
}

// quoteReasonWrites returns the writes of quote reasons: create, update ID
// and delete ID. Their message is localized text.
func quoteReasonWrites(reasons resource[quoteReason]) []*write {
	item := request{path: reasons.itemPath()}
	manage := []string{quoteManageScope}
	return []*write{
		{
			use: createUse, short: "Create a quote reason",
			long: "Creates the quote reason in --file: its type (CHANGE or DECLINE), code and message, and\n" +
				"optionally an id, which Emporix makes otherwise; it answers with the id. A reason with the\n" +
				"id exists already: a conflict (exit 5).",
			method: http.MethodPost, req: request{path: reasons.path}, scopes: manage, body: &payload{},
			language: mapsOnly,
		},
		{
			use: reasons.itemUse("update"), short: "Replace a quote reason",
			long: "Replaces the quote reason ID with the one in --file: its type, code, message and\n" +
				"metadata.version, as `epx quotes reasons get ID -o json` prints it. Emporix changes only\n" +
				"that version of the reason and answers an older one with 409 Conflict (exit 5); an\n" +
				"unknown reason is not found (exit 4).\n" + reasons.idSource() + ".",
			method: http.MethodPut, req: item, scopes: manage, body: &payload{}, language: mapsOnly,
			ids: reasons.itemIDs(),
		},
		reasons.deleteWrite(manage, "Deletes the quote reason ID; it sends no body."+unknownAsDeleted("ID")+"\n"+
			reasons.idSource()+"."),
	}
}

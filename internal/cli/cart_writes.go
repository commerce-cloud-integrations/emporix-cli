package cli

import (
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The writes of the Cart Service: carts, their items and discounts, and the
// chains of cart commands that execute runs. Each requests cart.cart_manage,
// the only scope that the reference names for a service token, and those
// whose body may hold an external price, product, fee or discount also
// cart.cart_manage_external_prices, only with --allow-external-prices.

// externalPrices returns the choice of a write whose body may hold an
// external price, product, fee or discount, of a cart or a quote, which
// Emporix accepts only with the Cart Service's
// cart.cart_manage_external_prices: --allow-external-prices requests it, to
// allow what, e.g. "external discounts (externalDiscounts)"; epx requests it
// only when told so.
func externalPrices(what string) queryChoice {
	return queryChoice{options: []choiceOption{{
		flag: "allow-external-prices", usage: "allow " + what, scopes: []string{"cart.cart_manage_external_prices"},
	}}}
}

// cartWrites returns the writes of carts: create; update, delete, refresh,
// change-site, change-currency and merge CART_ID; and execute CART_ID,
// which runs a chain of cart commands.
func cartWrites(a *app, carts resource[cart]) []*cobra.Command {
	scopes, cartID, one := []string{cartScope}, carts.argument(), carts.itemPath()
	about := "\n" + carts.argumentSentence(notFound)
	action := func(name string) request { return request{path: one + "/" + name} }
	// posted is the command use CART_ID, a POST to the cart's action name
	// whose body is the JSON object of field, a required flag, or --file.
	posted := func(use, name, short, long string, field bodyField) *cobra.Command {
		return (&write{
			use: use + " " + cartID, short: short, method: http.MethodPost, long: long + about,
			req: action(name), scopes: scopes, body: &payload{typed: objectBody{field.require()}},
		}).command(a)
	}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a customer's cart", method: http.MethodPost,
			long: "Creates a customer's cart from the JSON object of --file: its customerId, siteCode,\n" +
				"currency and type (such as shopping or wishlist; a customer may hold one open cart of\n" +
				"each type: another is a conflict, exit 5), and optionally legalEntityId, channel,\n" +
				"deliveryWindow, addresses and mixins. Emporix answers with the cartId. A cart of an\n" +
				"anonymous customer needs, instead of a customerId, the session-id header of the\n" +
				"customer's session, which epx does not send: without either, Emporix answers 400 Bad\n" +
				"Request.",
			req: request{path: carts.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: "update " + cartID, short: "Update a cart", method: http.MethodPut,
			long: "Updates the cart " + cartID + " with the JSON object of --file: customerId, currency, type,\n" +
				"status (OPEN or CLOSED), deliveryWindow, zipCode and countryCode (both or neither: one\n" +
				"alone is answered 400 Bad Request, exit 1), addresses, mixins and externalDiscounts, as\n" +
				"`epx carts get " + cartID + " -o json` shows them. Emporix creates no missing cart, and\n" +
				"ignores metadata.version in the body: only `epx carts execute --versioning` checks the\n" +
				"cart's version." + about,
			req: request{path: one}, scopes: scopes, body: &payload{},
			choice: externalPrices("external discounts (externalDiscounts)"),
		}).command(a),
		(&write{
			use: "delete " + cartID, short: "Delete a cart", method: http.MethodDelete,
			long: "Deletes the cart " + cartID + " with its items. A CLOSED cart, such as one merged into\n" +
				"another, is not found." + about,
			req: request{path: one}, scopes: scopes, body: noBody{},
		}).command(a),
		(&write{
			use: "refresh " + cartID, short: "Refresh the prices of a cart's items", method: http.MethodPut,
			long: "Refreshes the cart " + cartID + ": Emporix reads the prices of its items again and assigns\n" +
				"those that changed. It sends no body." + about,
			req: action("refresh"), scopes: scopes, body: noBody{},
		}).command(a),
		posted("change-site", "changeSite", "Move a cart to another site",
			"Moves the cart "+cartID+" to the site that --site names, whose language, currency,\n"+
				"shipping, tax and payment the cart then takes; where the site's currency differs, Emporix\n"+
				"matches the prices of the cart's items in it.",
			stringField("siteCode", "site", "the `code` of the site to move the cart to")),
		posted("change-currency", "changeCurrency", "Change the currency of a cart",
			"Changes the currency of the cart "+cartID+" to the one that --currency names; Emporix\n"+
				"matches the prices of the cart's items in it.",
			stringField("currency", "currency", "the three-letter `code` of the currency, e.g. USD")),
		posted("merge", "merge", "Merge anonymous carts into a customer's cart",
			"Merges the anonymous carts that --cart names into the cart "+cartID+", which must belong\n"+
				"to a customer. Emporix answers with no body and closes the merged carts rather than\n"+
				"deleting them: `epx carts search` still lists them as CLOSED, but get and delete no\n"+
				"longer find them.",
			listField("carts", "cart", "the `ID` of an anonymous cart to merge (repeatable)")),
		cartExecute(cartID, action("execute"), scopes, about).command(a),
	}
}

// cartExecute declares execute CART_ID, which runs a chain of up to 10 cart
// commands on one cart in one request, sent to req.
func cartExecute(cartID string, req request, scopes []string, about string) *write {
	return &write{
		use: "execute " + cartID, short: "Run a chain of cart commands", method: http.MethodPost,
		long: "Runs a chain of commands on the cart " + cartID + " in one request, from the JSON object of\n" +
			"--file: {\"commands\": [...]}, each with its type (AddCartItem, UpdateCartItem,\n" +
			"DeleteCartItem, DeleteCartItems, GetCart, AddCartItemsBatch, UpdateCartItemsBatch,\n" +
			"UpdateCart, ApplyCartDiscount, GetCartDiscounts, DeleteCartDiscounts, DeleteCartDiscount,\n" +
			"RefreshCart or ValidateCart), the body of that operation as data, and the rest of its\n" +
			"request as options (itemId, partial, expandCalculation, resourceVersion, codes,\n" +
			"discountIndex, …). epx prints the result of each command that ran, the data of each\n" +
			"as its operation answers. Without --on-error resume, Emporix runs no command after one\n" +
			"that failed, but keeps what those before it changed. With --versioning explicit, Emporix\n" +
			"answers a stale resourceVersion as a failed command with code 500 (its message names 409\n" +
			"CONFLICT), so epx exits 1 and says it may have been carried out; it was not." + about,
		example: "  epx carts execute 6a86b20c2f5961330a8b3eb6 --file add-then-get.json --dry-run\n" +
			"  epx carts execute 6a86b20c2f5961330a8b3eb6 --file chain.json --on-error resume --yes",
		req: req, scopes: scopes, body: &chain{most: 10},
		params: queryFlags{optional: []paramFlag{
			{param: "onError", flag: "on-error", choices: []string{"fail", "resume"},
				usage: "whether Emporix stops after a command that failed (fail, the default) or runs the rest (resume)"},
			{param: "versioning", flag: "versioning", choices: []string{"skip", "explicit", "follow"},
				usage: "how the writes check the cart's version (default skip): skip checks none, explicit each " +
					"command's options.resourceVersion, follow the first's, then the version each write makes"},
		}},
		choice: externalPrices("a command with an external price, product, fee or discount"),
	}
}

// The most items that Emporix adds and updates in one request, as the
// reference documents.
const (
	maxCartItemsAdded   = 200
	maxCartItemsUpdated = 50
)

// cartItemWrites returns the writes of items, the items of a cart below
// itemsOf (carts nested in their group): create, bulk-create, bulk-update and
// delete-all CART_ID, and update and delete CART_ID ID, which address one
// item as get reads it.
func cartItemWrites(a *app, itemsOf resource[cart], items child[cartItem]) []*cobra.Command {
	scopes, cartID := []string{cartScope}, itemsOf.argument()
	update, remove := childItemTarget(itemsOf, items, "update", nil), childItemTarget(itemsOf, items, "delete", nil)
	collection, batch := items.below(itemsOf.path), itemsOf.itemPath()+"/itemsBatch"
	about := "\n" + itemsOf.argumentSentence(notFound)
	item := childItemHelp(itemsOf, items, &update)
	external := externalPrices("an item with an external price, product, fee or discount")
	const takes = "product (its id) or itemYrn, quantity and price (priceId, originalAmount,\n" +
		"effectiveAmount and currency, as `epx prices match` finds them), and optionally\n" +
		"keepAsSeparateLineItem, externalFees, externalDiscounts and mixins.\n"
	return []*cobra.Command{
		(&write{
			use: createUse + " " + cartID, short: "Add a product to a cart", method: http.MethodPost,
			long: "Adds a product to the cart " + cartID + " as a new item, from the JSON object of --file: its\n" +
				takes + "--site names the cart's site. Emporix answers 201 with the itemId (e.g. 0 for the first\n" +
				"item) and yrn. A product that the cart holds already, added without\n" +
				"keepAsSeparateLineItem, raises the quantity of its item (the same itemId), so a create\n" +
				"sent twice adds twice. A product object\n" +
				"in the body is stored as sent, in place of the one Emporix would store. The item's\n" +
				"mixins.productCustomAttributes (leadTime, nonDelivery) feed\n" +
				"`epx carts delivery-restrictions`.\n" +
				"Emporix refuses a product that is not published with 400 Bad Request (exit 1), and an\n" +
				"item with an external price without --allow-external-prices with 403 Forbidden\n" +
				"(exit 3)." + about,
			req:    request{path: collection, query: url.Values{siteCodeParam: {placeholderSite}}},
			scopes: scopes, body: &payload{}, choice: external,
		}).command(a),
		(&write{
			use: update.use, short: "Update an item of a cart", method: http.MethodPut,
			long: "Replaces an item of the cart " + cartID + " with the JSON object of --file, whose members are\n" +
				"those that create takes, or with --partial changes only the members it gives." + item,
			req: request{path: update.path}, scopes: scopes, body: &payload{}, choice: external,
			switches: []paramFlag{{param: partialParam, flag: partialParam,
				usage: "change only the members that the body gives; without it, the body replaces the whole item"}},
		}).command(a),
		(&write{
			use: remove.use, short: "Remove an item from a cart", method: http.MethodDelete,
			long: "Removes an item from the cart " + cartID + "." + item,
			req:  request{path: remove.path}, scopes: scopes, body: noBody{},
		}).command(a),
		(&write{
			use: "delete-all " + cartID, short: "Remove every item from a cart", method: http.MethodDelete,
			long: "Removes every item from the cart " + cartID + "." + about,
			req:  request{path: collection}, scopes: scopes, body: noBody{},
			deletes: deletion{many: "every item of the cart"},
		}).command(a),
		(&write{
			use: bulkCreateUse + " " + cartID, short: "Add products to a cart", method: http.MethodPost,
			long: "Adds the items in --file, a JSON array of items, each as create takes it, to the cart\n" +
				cartID + "." + about,
			req: request{path: batch}, scopes: scopes, body: &payload{maxItems: maxCartItemsAdded}, choice: external,
		}).command(a),
		(&write{
			use: bulkUpdateUse + " " + cartID, short: "Update items of a cart", method: http.MethodPut,
			long: "Updates the items in --file, a JSON array of items, each with the id of an item of the\n" +
				"cart " + cartID + " and the whole item, as update takes it without --partial. Emporix\n" +
				"refuses an entry with only quantity for that item (400 \"The Price must be\n" +
				"defined.\"). `epx carts execute` changes some members of several items\n" +
				"(UpdateCartItemsBatch with options.partial)." + about,
			req: request{path: batch}, scopes: scopes, body: &payload{maxItems: maxCartItemsUpdated}, choice: external,
		}).command(a),
	}
}

// discountCodesParam is the query parameter, and the flag, that names the
// discounts that delete-all removes.
const discountCodesParam = "codes"

// cartDiscountWrites returns the writes of discounts, the discounts of a cart
// below discountsOf (carts nested in their group): create and delete-all
// CART_ID, and delete CART_ID INDEX, which removes the discount at INDEX.
func cartDiscountWrites(a *app, discountsOf resource[cart], discounts child[cartDiscount]) []*cobra.Command {
	scopes, cartID := []string{cartScope}, discountsOf.argument()
	remove := childItemTarget(discountsOf, discounts, "delete", []string{columnIndex})
	collection := discounts.below(discountsOf.path)
	about := "\n" + discountsOf.argumentSentence(notFound)
	return []*cobra.Command{
		(&write{
			use: createUse + " " + cartID, short: "Apply a discount to a cart", method: http.MethodPost,
			long: "Applies a discount to the cart " + cartID + ": the coupon that --code names, or the JSON\n" +
				"object of --file, with the coupon's code and optionally its name, amount or discountRate,\n" +
				"currency and discountCalculationType (SUBTOTAL or TOTAL). One discount per request.\n" +
				"Emporix may limit the coupons per cart (a tenant setting; one tenant allows one: a second is\n" +
				"refused with 400 Bad Request, exit 1). It refuses a coupon restricted to customers\n" +
				"(restrictions.validFor), even to the cart's own customer, with " + forbiddenRefusal + ",\n" +
				"\"User not authorized.\", when an API key credential such as epx's applies it." +
				about,
			// The Cart Service answers a missing scope with the same message (its reference's
			// 403); epx checks cart.cart_manage before it sends, so a 403 here refuses the coupon.
			req: request{path: collection}, scopes: scopes, answers: emporix.Answers{ForbiddenIsRefusal: true},
			body: &payload{typed: objectBody{stringField("code", "code", "the `code` of the coupon to apply").require()}},
		}).command(a),
		(&write{
			use: remove.use, short: "Remove a discount from a cart", method: http.MethodDelete,
			long: "Removes the discount at INDEX from the cart " + cartID + ". The discounts after it then move\n" +
				"up by one: list them again before you remove another by INDEX, or remove discounts by\n" +
				"code with delete-all --codes." + childItemHelp(discountsOf, discounts, &remove),
			req: request{path: remove.path}, scopes: scopes, body: noBody{},
		}).command(a),
		(&write{
			use: "delete-all " + cartID, short: "Remove the discounts of a cart", method: http.MethodDelete,
			long: "Removes every discount from the cart " + cartID + ", or with --codes only those with the codes\n" +
				"it names." + about,
			req: request{path: collection}, scopes: scopes, body: noBody{},
			deletes: deletion{many: "every discount of the cart (or those that --codes names)"},
			params: queryFlags{optional: []paramFlag{{param: discountCodesParam, flag: discountCodesParam,
				usage: "remove only the discounts with these `CODES`, separated by commas"}}},
		}).command(a),
	}
}

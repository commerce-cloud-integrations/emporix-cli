package cli

import (
	"net/http"
	"slices"
)

// The write cases of the Cart Service (see TestWrites): every write requests
// cart.cart_manage, and --allow-external-prices adds
// cart.cart_manage_external_prices.

// cartItemJSON is an item as items create takes it.
const cartItemJSON = `{"product":{"id":"100200"},"quantity":2,"price":{"priceId":"p1","originalAmount":15.90,` +
	`"effectiveAmount":15.90,"currency":"EUR"}}`

// cartChainJSON is a chain of two cart commands: add an item, then read the
// calculated cart.
const cartChainJSON = `{"commands":[{"type":"AddCartItem","data":` + cartItemJSON + `},` +
	`{"type":"GetCart","options":{"expandCalculation":true}}]}`

// cartChainResults answers cartChainJSON: both commands succeeded.
const cartChainResults = `{"results":[{"index":0,"type":"AddCartItem","code":201,"status":"Created",` +
	`"data":{"itemId":"3"}},{"index":1,"type":"GetCart","code":200,"status":"OK","data":{"id":"c1"}}]}`

// externalScopes returns the scopes of a cart write with
// --allow-external-prices.
func externalScopes() []string { return []string{cartCaseScope, "cart.cart_manage_external_prices"} }

// cartWriteCases returns the write cases of carts, their items and discounts.
func cartWriteCases() []writeCase {
	return slices.Concat(cartChanges(), cartItemChanges(), cartDiscountChanges())
}

// cartChanges returns the write cases of carts.
func cartChanges() []writeCase {
	scopes := []string{cartCaseScope}
	return []writeCase{
		{
			args: []string{"carts", "create"}, file: `{"siteCode":"AcmeStore_NL","currency":"EUR","type":"shopping"}`,
			method: http.MethodPost, path: "/cart/acme1/carts", scopes: scopes, status: http.StatusCreated,
			body: `{"cartId":"c1"}`, want: "{\n  \"cartId\": \"c1\"\n}\n",
		},
		{
			args: []string{"carts", "update", "c 1"}, file: `{"status":"CLOSED"}`, method: http.MethodPut,
			path: "/cart/acme1/carts/c%201", scopes: scopes, status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "update", "c1", "--allow-external-prices"},
			file: `{"externalDiscounts":[{"id":"d1","value":5}]}`, method: http.MethodPut, path: "/cart/acme1/carts/c1",
			scopes: externalScopes(), status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "delete", "c1"}, method: http.MethodDelete, path: "/cart/acme1/carts/c1",
			scopes: scopes, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "refresh", "c1"}, method: http.MethodPut, path: "/cart/acme1/carts/c1/refresh",
			scopes: scopes, status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "change-site", "c1", "--site", "AcmeStore_BE"}, sent: `{"siteCode":"AcmeStore_BE"}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/changeSite", scopes: scopes, status: http.StatusOK,
		},
		{
			args: []string{"carts", "change-currency", "c1", "--currency", "USD"}, sent: `{"currency":"USD"}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/changeCurrency", scopes: scopes, status: http.StatusOK,
		},
		{
			args: []string{"carts", "merge", "c1", "--cart", "a1", "--cart", "a2"}, sent: `{"carts":["a1","a2"]}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/merge", scopes: scopes, status: http.StatusOK,
		},
		// --file replaces the typed flags of a small body.
		{
			args: []string{"carts", "change-site", "c1"}, file: `{"siteCode":"AcmeStore_NL"}`, method: http.MethodPost,
			path: "/cart/acme1/carts/c1/changeSite", scopes: scopes, status: http.StatusOK,
		},
		{
			args: []string{"carts", "change-currency", "c1"}, file: `{"currency":"EUR"}`, method: http.MethodPost,
			path: "/cart/acme1/carts/c1/changeCurrency", scopes: scopes, status: http.StatusOK,
		},
		{
			args: []string{"carts", "merge", "c1"}, file: `{"carts":["a1"]}`, method: http.MethodPost,
			path: "/cart/acme1/carts/c1/merge", scopes: scopes, status: http.StatusOK,
		},
		{
			args: []string{"carts", "execute", "c1"}, file: cartChainJSON, method: http.MethodPost,
			path: "/cart/acme1/carts/c1/execute", scopes: scopes, status: http.StatusMultiStatus, body: cartChainResults,
			want: "{\n  \"results\": [\n    {\n      \"code\": 201,\n      \"data\": {\n        \"itemId\": \"3\"\n      },\n" +
				"      \"index\": 0,\n      \"status\": \"Created\",\n      \"type\": \"AddCartItem\"\n    },\n" +
				"    {\n      \"code\": 200,\n      \"data\": {\n        \"id\": \"c1\"\n      },\n      \"index\": 1,\n" +
				"      \"status\": \"OK\",\n      \"type\": \"GetCart\"\n    }\n  ]\n}\n",
		},
		{
			args: []string{"carts", "execute", "c1", "--on-error", "resume", "--versioning", "follow",
				"--allow-external-prices"},
			file: cartChainJSON, method: http.MethodPost, path: "/cart/acme1/carts/c1/execute",
			query: "onError=resume&versioning=follow", scopes: externalScopes(),
			status: http.StatusMultiStatus, body: `{"results":[{"index":0,"code":201},{"index":1,"code":200}]}`,
			want: "{\n  \"results\": [\n    {\n      \"code\": 201,\n      \"index\": 0\n    },\n" +
				"    {\n      \"code\": 200,\n      \"index\": 1\n    }\n  ]\n}\n",
		},
	}
}

// cartItemChanges returns the write cases of the items of a cart.
func cartItemChanges() []writeCase {
	scopes := []string{cartCaseScope}
	return []writeCase{
		{
			args: []string{"carts", "items", "create", "c1", "--site", "AcmeStore_NL"}, file: cartItemJSON,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/items", query: "siteCode=AcmeStore_NL", scopes: scopes,
			status: http.StatusCreated, body: `{"itemId":"0"}`, want: "{\n  \"itemId\": \"0\"\n}\n",
		},
		{
			args:   []string{"carts", "items", "create", "c1", "--site", "AcmeStore_NL", "--allow-external-prices"},
			file:   `{"itemType":"EXTERNAL","product":{"id":"ext-1","name":"Cable"},"quantity":1}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/items", query: "siteCode=AcmeStore_NL",
			scopes: externalScopes(), status: http.StatusCreated,
		},
		{
			args: []string{"carts", "items", "update", "c1", "1"}, file: cartItemJSON, method: http.MethodPut,
			path: "/cart/acme1/carts/c1/items/1", query: "partial=false", scopes: scopes, status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "items", "update", "c1", "0", "--partial"}, file: `{"quantity":3}`,
			method: http.MethodPut, path: "/cart/acme1/carts/c1/items/0", query: "partial=true", scopes: scopes,
			status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "items", "delete", "c1", "0"}, method: http.MethodDelete,
			path: "/cart/acme1/carts/c1/items/0", scopes: scopes, confirmBy: "0", status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "items", "delete-all", "c1"}, method: http.MethodDelete,
			path: "/cart/acme1/carts/c1/items", scopes: scopes, confirmBy: "c1",
			many: "every item of the cart", status: http.StatusNoContent,
		},
		{
			// Emporix answers the result of each item with its code as status, a number.
			args: []string{"carts", "items", "bulk-create", "c1"}, file: "[" + cartItemJSON + "," + cartItemJSON + "]",
			method: http.MethodPost, path: "/cart/acme1/carts/c1/itemsBatch", scopes: scopes, status: http.StatusOK,
			body: `[{"index":0,"status":201,"id":"1"},{"index":1,"status":201,"id":"2"}]`,
			want: "[\n  {\n    \"id\": \"1\",\n    \"index\": 0,\n    \"status\": 201\n  },\n" +
				"  {\n    \"id\": \"2\",\n    \"index\": 1,\n    \"status\": 201\n  }\n]\n",
		},
		{
			args: []string{"carts", "items", "bulk-update", "c1"}, file: `[{"id":"1","quantity":5}]`,
			method: http.MethodPut, path: "/cart/acme1/carts/c1/itemsBatch", scopes: scopes,
			status: http.StatusMultiStatus, body: `[{"index":0,"id":"1","code":204,"status":"No Content"}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"1\",\n    \"index\": 0,\n    \"status\": \"No Content\"\n  }\n]\n",
		},
	}
}

// cartDiscountChanges returns the write cases of the discounts of a cart.
func cartDiscountChanges() []writeCase {
	const cartDiscountsDeleted = "every discount of the cart (or those that --codes names)"
	scopes := []string{cartCaseScope}
	return []writeCase{
		{
			args: []string{"carts", "discounts", "create", "c1", "--code", "SAVE10"}, sent: `{"code":"SAVE10"}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/discounts", scopes: scopes, status: http.StatusCreated,
			body: `{"code":"SAVE10","discountIndex":0}`, want: "{\n  \"code\": \"SAVE10\",\n  \"discountIndex\": 0\n}\n",
			refuses: true,
		},
		{
			args: []string{"carts", "discounts", "create", "c1"}, file: `{"code":"TEN","amount":10,"currency":"EUR"}`,
			method: http.MethodPost, path: "/cart/acme1/carts/c1/discounts", scopes: scopes, status: http.StatusCreated,
			refuses: true,
		},
		{
			args: []string{"carts", "discounts", "delete", "c1", "1"}, method: http.MethodDelete,
			path: "/cart/acme1/carts/c1/discounts/1", scopes: scopes, confirmBy: "1", status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "discounts", "delete-all", "c1"}, method: http.MethodDelete,
			path: "/cart/acme1/carts/c1/discounts", scopes: scopes, confirmBy: "c1",
			many: cartDiscountsDeleted, status: http.StatusNoContent,
		},
		{
			args: []string{"carts", "discounts", "delete-all", "c1", "--codes", "SAVE10,TEN"}, method: http.MethodDelete,
			path: "/cart/acme1/carts/c1/discounts", query: "codes=SAVE10%2CTEN", scopes: scopes,
			confirmBy: "c1", many: cartDiscountsDeleted, status: http.StatusNoContent,
		},
	}
}

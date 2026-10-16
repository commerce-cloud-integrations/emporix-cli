package cli

import (
	"net/http"
	"slices"
)

// The read cases of the Cart Service (see TestReads), which reads carts only
// with the write-capable cart.cart_manage scope: every read is opt-in.

// cartsBody is a page of carts: a customer's, with its final price, and an
// anonymous one.
const cartsBody = `[{"id":"c1","status":"OPEN","siteCode":"AcmeStore_NL","type":"shopping","customerId":"C0123",
	"legalEntityId":"le1","currency":"EUR","items":[{"id":"0"},{"id":"1"}],
	"calculatedPrice":{"finalPrice":{"netValue":285.592,"grossValue":339.485}},
	"metadata":{"createdAt":"2026-10-01T08:00:00Z","modifiedAt":"2026-10-02T09:30:00Z","version":3}},
	{"id":"c2","status":"CLOSED","siteCode":"AcmeStore_NL","sessionId":"s-1","currency":"EUR","items":[]}]`

// cartsTable is the table of cartsBody.
const cartsTable = "ID  STATUS  SITE          TYPE      CUSTOMER  LEGAL ENTITY  ITEMS  TOTAL        MODIFIED\n" +
	"c1  OPEN    AcmeStore_NL  shopping  C0123     le1           2      339.485 EUR  2026-10-02T09:30:00Z\n" +
	"c2  CLOSED  AcmeStore_NL                                    0                   \n"

// cartItemsBody holds items as the items reads send them, without unitPrice
// and calculatedPrice: one with a localized product name, a price and its tax
// line, one with a plain name only.
const cartItemsBody = `[{"id":"0","type":"INTERNAL","quantity":2,"effectiveQuantity":2,"keepAsSeparateLineItem":false,
	"product":{"id":"100200","code":"100200","sku":"100200","name":"Transistor",
	"localizedName":{"en":"Transistor","nl":"Transistor NL"}},
	"price":{"priceId":"p1","currency":"EUR","originalAmount":10,"effectiveAmount":10.00},
	"itemTaxInfo":[{"name":"STANDARD","rate":19,"netValue":20,"grossValue":23.80,
	"value":{"amount":3.8,"currency":"EUR"}}]},
	{"id":"1","type":"EXTERNAL","quantity":1.5,"effectiveQuantity":1.5,"product":{"id":"ext-1","name":"Cable"}}]`

// cartCaseScope is the scope every cart read case states, as written (not
// as carts.go names it).
const cartCaseScope = "cart.cart_manage"

// cartReads returns the Cart Service reads.
func cartReads() []readCase { return slices.Concat(cartLookups(), cartItemReads(), cartContents()) }

// cartLookups returns the reads of carts: by ID, by criteria and by search.
func cartLookups() []readCase {
	const scope = cartCaseScope
	return []readCase{
		{
			args:   []string{"carts", "get", "c 1", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c%201",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"c 1","siteCode":"AcmeStore_NL","customerId":"C0123","currency":"EUR",
				"items":[{"id":"0"}],"calculatedPrice":{"finalPrice":{"grossValue":10.50}}}`,
			want: "ID   STATUS  SITE          TYPE  CUSTOMER  LEGAL ENTITY  ITEMS  TOTAL      MODIFIED\n" +
				"c 1          AcmeStore_NL        C0123                   1      10.50 EUR  \n",
		},
		{
			// Every criterion given is sent with the site; create never is.
			args: []string{"carts", "find", "--site", "AcmeStore_NL", "--customer", "C0123", "--legal-entity", "le 1",
				"--type", "shopping", "--allow-write-scope"},
			path:   "/cart/acme1/carts",
			query:  "customerId=C0123&legalEntityId=le+1&siteCode=AcmeStore_NL&type=shopping",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"c1","status":"OPEN","siteCode":"AcmeStore_NL","type":"shopping","customerId":"C0123",
				"legalEntityId":"le 1","currency":"EUR"}`,
			want: "ID  STATUS  SITE          TYPE      CUSTOMER  LEGAL ENTITY  ITEMS  TOTAL  MODIFIED\n" +
				"c1  OPEN    AcmeStore_NL  shopping  C0123     le 1          0             \n",
		},
		{
			args:   []string{"carts", "find", "--site", "AcmeStore_NL", "--session", "s-1", "--allow-write-scope"},
			path:   "/cart/acme1/carts",
			query:  "sessionId=s-1&siteCode=AcmeStore_NL",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"c2","status":"OPEN","siteCode":"AcmeStore_NL","sessionId":"s-1"}`,
			want: "ID  STATUS  SITE          TYPE  CUSTOMER  LEGAL ENTITY  ITEMS  TOTAL  MODIFIED\n" +
				"c2  OPEN    AcmeStore_NL                                0             \n",
		},
		{
			// Without --query, the body's q filters nothing; pages are read as lists read them.
			args:    []string{"carts", "search", "--allow-write-scope"},
			path:    "/cart/acme1/carts/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    cartsBody,
			want:    cartsTable,
		},
		{
			args:    []string{"carts", "search", "--query", "status:OPEN customerId:C0123", "--allow-write-scope"},
			path:    "/cart/acme1/carts/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"status:OPEN customerId:C0123"}`,
			body:    cartsBody,
			want:    cartsTable,
		},
	}
}

// cartItemReads returns the reads of the items of one cart, all or one.
func cartItemReads() []readCase {
	const scope = cartCaseScope
	return []readCase{
		{
			// Emporix documents no paging for a cart's items: one request.
			args:   []string{"carts", "items", "list", "c1", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/items",
			scopes: []string{scope},
			mode:   getWhole,
			body:   cartItemsBody,
			want: "ID  PRODUCT  NAME        QUANTITY  PRICE      GROSS\n" +
				"0   100200   Transistor  2         10.00 EUR  23.80 EUR\n" +
				"1   ext-1    Cable       1.5                  \n",
		},
		{
			args:   []string{"carts", "items", "list", "c1", "--language", "nl", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/items",
			scopes: []string{scope},
			mode:   getWhole,
			body:   cartItemsBody,
			want: "ID  PRODUCT  NAME           QUANTITY  PRICE      GROSS\n" +
				"0   100200   Transistor NL  2         10.00 EUR  23.80 EUR\n" +
				"1   ext-1    Cable          1.5                  \n",
		},
		{
			// Live, the items reads send gross values with 40 decimals: trimmed to the cents.
			args:   []string{"carts", "items", "list", "c2", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c2/items",
			scopes: []string{scope},
			mode:   getWhole,
			body: `[{"id":"0","product":{"id":"100200","name":"Acme"},"quantity":3,
				"price":{"effectiveAmount":10.0,"currency":"EUR"},
				"itemTaxInfo":[{"grossValue":35.7000000000000000000000000000000000000000}]}]`,
			want: "ID  PRODUCT  NAME  QUANTITY  PRICE     GROSS\n" +
				"0   100200   Acme  3         10.0 EUR  35.70 EUR\n",
		},
		{
			args:   []string{"carts", "items", "get", "c1", "0/1", "--language", "de", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/items/0%2F1",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"0/1","quantity":3,"product":{"id":"100200","name":"Transistor",
				"localizedName":{"de":"Transistor DE","en":"Transistor"}},
				"price":{"currency":"EUR","effectiveAmount":1.25},"itemTaxInfo":[{"grossValue":2.14},{"grossValue":2.01}]}`,
			want: "ID   PRODUCT  NAME           QUANTITY  PRICE     GROSS\n" +
				"0/1  100200   Transistor DE  3         1.25 EUR  2.14 EUR, 2.01 EUR\n",
		},
	}
}

// cartContents returns the reads of what else one cart holds: its discounts,
// delivery restrictions and validation.
func cartContents() []readCase {
	const scope = cartCaseScope
	return []readCase{
		{
			// AMOUNT trimmed (trimAmount), RATE as sent; the discounts read sends no valid, so no column shows it.
			args:   []string{"carts", "discounts", "list", "c1", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/discounts",
			scopes: []string{scope},
			mode:   getWhole,
			body: `[{"code":"SAVE7","name":"Save 7%","discountRate":7,"discountCalculationType":"SUBTOTAL","discountIndex":0},
				{"code":"TEN","amount":10.00,"currency":"EUR","discountCalculationType":"TOTAL","discountIndex":1},
				{"code":"ANY","discountIndex":2}]`,
			want: "CODE   NAME     AMOUNT     RATE  APPLIES TO  INDEX\n" +
				"SAVE7  Save 7%             7%    SUBTOTAL    0\n" +
				"TEN             10.00 EUR        TOTAL       1\n" +
				"ANY                                          2\n",
		},
		{
			args:   []string{"carts", "delivery-restrictions", "c1", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/dtRestrictions",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"leadTime":10,"nonDelivery":["1","3","5"]}`,
			want:   "LEAD HOURS  NON-DELIVERY DAYS\n10          1, 3, 5\n",
		},
		{
			// Nothing to report: the missing lead time is named.
			args:   []string{"carts", "delivery-restrictions", "c2", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c2/dtRestrictions",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{}`,
			want:   "LEAD HOURS    NON-DELIVERY DAYS\nnot reported  \n",
		},
		{
			args:   []string{"carts", "validate", "c1", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c1/validate",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"isValid":false,"itemsValidationDetails":[
				{"id":"0","errors":[{"errorCode":"CART-ITEM-UNIT-PRICE-100001","message":"not found","field":"items.unitPrice"}]},
				{"id":"1","errors":[{"errorCode":"E1"},{"errorCode":"E2"}]}]}`,
			want: "VALID  ERRORS\nfalse  0: CART-ITEM-UNIT-PRICE-100001; 1: E1, E2\n",
		},
		{
			args:   []string{"carts", "validate", "c2", "--allow-write-scope"},
			path:   "/cart/acme1/carts/c2/validate",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"isValid":true}`,
			want:   "VALID  ERRORS\ntrue   \n",
		},
	}
}

package cli

import (
	"net/http"
	"strings"
	"testing"
)

// The read cases of the searches of orders, prices, price lists and
// availability, the order count and price matching, which Emporix serves as
// POST with a JSON body, or HEAD, but which change nothing (see TestReads).

// commerceSearchReads returns the searches of the Order, Price and
// Availability services, the order count and price matching.
func commerceSearchReads() []readCase {
	return append(orderSearchReads(), priceSearchReads()...)
}

// orderSearchReads returns the order search and count and the availability search.
func orderSearchReads() []readCase {
	return []readCase{
		{
			// Without --query, the body's q filters nothing; pages are read as list reads them.
			args:    []string{"orders", "search"},
			path:    "/order-v2/acme1/salesorders/search",
			scopes:  []string{"order.order_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body: `[{"id":"EON1002","orderType":"MASTER_ORDER","status":"CREATED","created":"2026-10-02T09:00:00.000Z",
				"siteCode":"AcmeStore_NL","customer":{"email":"bo@example.com"},"entries":[{"id":"e1"}],"currency":"EUR",
				"calculatedPrice":{"finalPrice":{"grossValue":12.10}}}]`,
			want: "ID       TYPE          STATUS   CREATED                   CUSTOMER        SITE          ITEMS  TOTAL\n" +
				"EON1002  MASTER_ORDER  CREATED  2026-10-02T09:00:00.000Z  bo@example.com  AcmeStore_NL  1      12.10 EUR\n",
		},
		{
			args:    []string{"orders", "search", "--query", "status:(CREATED,CONFIRMED) created:>2026-10-01"},
			path:    "/order-v2/acme1/salesorders/search",
			scopes:  []string{"order.order_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"status:(CREATED,CONFIRMED) created:>2026-10-01"}`,
			body:    `[{"id":"EON1003","status":"CONFIRMED","customer":{"name":"Acme Ltd"}}]`,
			want: "ID       TYPE  STATUS     CREATED  CUSTOMER  SITE  ITEMS  TOTAL\n" +
				"EON1003        CONFIRMED           Acme Ltd        0      \n",
		},
		{
			// The count is the X-Total-Count of a HEAD request, which reads no orders.
			args:    []string{"orders", "count"},
			path:    "/order-v2/acme1/salesorders",
			scopes:  []string{"order.order_read"},
			filters: listFilters{kinds: queryable},
			mode:    getWhole,
			method:  http.MethodHead,
			total:   "1250",
			want:    "COUNT\n1250\n",
		},
		{
			args:    []string{"orders", "count", "--query", "status:CREATED"},
			path:    "/order-v2/acme1/salesorders",
			query:   "q=status%3ACREATED",
			scopes:  []string{"order.order_read"},
			filters: listFilters{kinds: queryable},
			mode:    getWhole,
			method:  http.MethodHead,
			total:   "0",
			want:    "COUNT\n0\n",
		},
		{
			// Records come from every site; like list, the search needs no scope.
			args:    []string{"availability", "search", "--query", "available:false"},
			path:    "/availability/acme1/availability/search",
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"available:false"}`,
			body: `[{"id":"main:001234","productId":"001234","site":"main","stockLevel":0,"available":false,
				"distributionChannel":"HOME_DELIVERY"},{"productId":"100200","site":"AcmeStore_NL","stockLevel":3.50}]`,
			want: "PRODUCT  SITE          STOCK  AVAILABLE  CHANNEL\n" +
				"001234   main          0      false      HOME_DELIVERY\n" +
				"100200   AcmeStore_NL  3.50              \n",
		},
	}
}

// priceSearchReads returns the searches of prices, price lists and their
// prices, and price matching.
func priceSearchReads() []readCase {
	const priceScope, listScope = "price.price_read", "price.pricelist_read"
	return []readCase{
		{
			// Without --site or --currency, the body names only the products (every site and
			// currency); a product without prices adds no row.
			args:   []string{"prices", "search", "--product", "001234", "--product", "ZZ 1"},
			path:   "/price/acme1/prices/search",
			scopes: []string{priceScope},
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"itemIds":["001234","ZZ 1"]}`,
			body: `[{"itemId":"001234","prices":[{"id":"p1","itemId":{"itemType":"PRODUCT","id":"001234"},
				"currency":"EUR","location":{"countryCode":"NL"},"restrictions":{"siteCodes":["AcmeStore_NL"]},
				"tierValues":[{"priceValue":15.90}]},{"id":"p3","itemId":{"itemType":"PRODUCT","id":"001234"},
				"currency":"CHF"}]},{"itemId":"ZZ 1","prices":[]}]`,
			want: "ID  ITEM    TYPE     CURRENCY  COUNTRY  SITES         TIERS  FROM  TO\n" +
				"p1  001234  PRODUCT  EUR       NL       AcmeStore_NL  15.90        \n" +
				"p3  001234  PRODUCT  CHF                                           \n",
		},
		{
			args: []string{"prices", "search", "--product", "100200", "--site", "AcmeStore_NL", "--currency", "EUR"},
			path: "/price/acme1/prices/search", scopes: []string{priceScope}, mode: getWhole, method: http.MethodPost,
			sent: `{"currency":"EUR","itemIds":["100200"],"siteCode":"AcmeStore_NL"}`,
			body: `[{"itemId":"100200","prices":[]}]`,
			want: "ID  ITEM  TYPE  CURRENCY  COUNTRY  SITES  TIERS  FROM  TO\n",
		},
		{
			// One unit of each product; values keep their text (0.0000317); the item is itemId as
			// Emporix sends it, or itemRef as the API reference's schema names it.
			args: []string{"prices", "match", "--product", "001234", "--product", "100200", "--site", "AcmeStore_NL",
				"--country", "NL", "--currency", "EUR"},
			path:   "/price/acme1/match-prices",
			scopes: []string{priceScope},
			mode:   getWhole,
			method: http.MethodPost,
			sent: `{"items":[{"itemId":{"itemType":"PRODUCT","id":"001234"},"quantity":{"quantity":1}},` +
				`{"itemId":{"itemType":"PRODUCT","id":"100200"},"quantity":{"quantity":1}}],"siteCode":"AcmeStore_NL",` +
				`"targetCurrency":"EUR","targetLocation":{"countryCode":"NL"}}`,
			body: `[{"priceId":"p1","itemId":{"itemType":"PRODUCT","id":"001234"},"site":{"code":"AcmeStore_NL"},
				"currency":"EUR","location":{"countryCode":"NL"},"quantity":{"quantity":1,"unitCode":"pc"},
				"originalValue":15.90,"effectiveValue":14.31,"totalValue":14.31,"includesTax":true,
				"tax":{"taxClass":"STANDARD","taxRate":21}},
				{"priceId":"p9","priceListId":"pl1","itemRef":{"itemType":"PRODUCT","id":"100200"},"currency":"EUR",
				"originalValue":0.0000317,"effectiveValue":0.0000317,"totalValue":0.0000317,"includesTax":false}]`,
			want: "ID  ITEM    PRICE LIST  SITE          COUNTRY  QUANTITY  ORIGINAL       EFFECTIVE      " +
				"TOTAL          TAX\n" +
				"p1  001234              AcmeStore_NL  NL       1 pc      15.90 EUR      14.31 EUR      " +
				"14.31 EUR      true\n" +
				"p9  100200  pl1                                          0.0000317 EUR  0.0000317 EUR  " +
				"0.0000317 EUR  false\n",
		},
		{
			args:    []string{"price-lists", "search", "--query", "currency:EUR", "--language", "nl"},
			path:    "/price/acme1/price-lists/search",
			scopes:  []string{listScope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"currency:EUR"}`,
			body:    `[{"id":"pl3","name":{"en":"Retail","nl":"Detailhandel"},"currency":"EUR","countries":["NL"]}]`,
			want: "ID   NAME          CURRENCY  SITE  COUNTRIES  REGIONS  CUSTOMER GROUPS  LEGAL ENTITY  FROM  TO\n" +
				"pl3  Detailhandel  EUR             NL                                                       \n",
		},
		{
			// An unknown price list lists nothing, as prices does.
			args:    []string{"price-lists", "prices", "search", "pl 1", "--query", "itemId.id:100200"},
			path:    "/price/acme1/price-lists/pl%201/prices/search",
			scopes:  []string{listScope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"itemId.id:100200"}`,
			body:    `[{"id":"p2","itemId":{"itemType":"PRODUCT","id":"100200"},"tierValues":[{"priceValue":9.50}]}]`,
			want:    "ID  ITEM    TYPE     MODEL  TIERS\np2  100200  PRODUCT         9.50\n",
		},
	}
}

// instanceAndDocumentSearchReads returns the searches of custom instances
// and of the documents of the Search Service.
func instanceAndDocumentSearchReads() []readCase {
	return []readCase{
		{
			args: []string{"custom-entities", "instances", "search", "DOCUMENT", "--query", "owner.type:EMPLOYEE",
				"--language", "nl"},
			path:    "/schema/acme1/custom-entities/DOCUMENT/instances/search",
			scopes:  []string{"custom.document_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"owner.type:EMPLOYEE"}`,
			body: `[{"id":"125","name":{"en":"Manual","nl":"Handleiding"},"owner":{"type":"EMPLOYEE","userId":"u2"},
				"metadata":{"version":1}}]`,
			want: "ID   NAME         OWNER        VERSION\n125  Handleiding  EMPLOYEE u2  1\n",
		},
		{
			// A saved search runs with the text given; SCORE keeps its text (3.10).
			args: []string{"search", "documents", "DOCUMENT", "--saved-search", "manuals", "--text", "diesel engine",
				"--language", "nl"},
			path:    "/search/acme1/search/DOCUMENT",
			scopes:  []string{"search.search_read", "custom.document_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"query":"diesel engine","searchQueryId":"manuals"}`,
			body: `[{"id":"d1","name":{"en":"Diesel engine","nl":"Dieselmotor"},"status":"ACTIVE","_score":4.2},
				{"id":"d2","name":{"en":"Diesel generator"},"_score":3.10}]`,
			want: "ID  NAME                   SCORE\nd1  Dieselmotor            4.2\nd2  Diesel generator (en)  3.10\n",
		},
	}
}

// TestPriceSearchChecksItsShape checks that a price search fails, also with
// -o json, on a response that does not list each product with its prices.
func TestPriceSearchChecksItsShape(t *testing.T) {
	const notAList = "decoding price response: expected a JSON array of products with their prices"
	for body, want := range map[string]string{
		`{"itemId":"100200","prices":[]}`:            notAList,
		`[{"itemId":"100200"}]`:                      "decoding price response: product 1 has no prices array",
		`[{"itemId":"100200","prices":[]},"100201"]`: notAList,
	} {
		for _, format := range []string{"table", "json"} {
			fake := newFakeEmporix(t)
			fake.route("/price/acme1/prices/search", allLanguages(respond(http.StatusOK, body)))
			code, stdout, stderr := runCLI(t, fake.env(), "prices", "search", "--product", "100200", "-o", format)
			if code != ExitError || stdout != "" || !strings.Contains(stderr, want) {
				t.Errorf("%s, -o %s: exit %d, stdout %q, stderr %q; want %q", body, format, code, stdout, stderr, want)
			}
		}
	}
}

// TestPriceMatchReadsAnArrayOrOneObject checks that prices match reads its
// response as Emporix sends it, an array, or as the API reference documents
// it, one object, and fails, also with -o json, on any other.
func TestPriceMatchReadsAnArrayOrOneObject(t *testing.T) {
	args := []string{"prices", "match", "--product", "100200", "--site", "main", "--country", "NL", "--currency", "EUR"}
	const match = `{"priceId":"p1","itemId":{"id":"100200"},"currency":"EUR","originalValue":9.50,` +
		`"effectiveValue":9.50,"totalValue":9.50}`
	const table = "ID  ITEM    PRICE LIST  SITE  COUNTRY  QUANTITY  ORIGINAL  EFFECTIVE  TOTAL     TAX\n" +
		"p1  100200                                       9.50 EUR  9.50 EUR   9.50 EUR  \n"
	for _, body := range []string{match, "[" + match + "]"} {
		fake := newFakeEmporix(t)
		fake.route("/price/acme1/match-prices", allLanguages(respond(http.StatusOK, body)))
		if code, stdout, stderr := runCLI(t, fake.env(), args...); code != ExitOK || stdout != table {
			t.Errorf("%s: exit %d, stderr %q, stdout:\n%s\nwant:\n%s", body, code, stderr, stdout, table)
		}
	}
	const want = "decoding price match response: expected a JSON array, or one price as a JSON object"
	for _, body := range []string{`{}`, `"p1"`, `null`} {
		for _, format := range []string{"table", "json"} {
			fake := newFakeEmporix(t)
			fake.route("/price/acme1/match-prices", allLanguages(respond(http.StatusOK, body)))
			code, stdout, stderr := runCLI(t, fake.env(), append(args, "-o", format)...)
			if code != ExitError || stdout != "" || !strings.Contains(stderr, want) {
				t.Errorf("%s, -o %s: exit %d, stdout %q, stderr %q; want %q", body, format, code, stdout, stderr, want)
			}
		}
	}
}

// TestOrderCountOutput checks that -o json prints the count alone, and that
// a response without a count fails.
func TestOrderCountOutput(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/order-v2/acme1/salesorders", respond(http.StatusOK, "", "X-Total-Count", "1250"))
	if code, stdout, stderr := runCLI(t, fake.env(), "orders", "count", "-o", "json"); code != ExitOK ||
		stdout != "1250\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.route("/order-v2/acme1/salesorders", respond(http.StatusOK, ""))
	code, stdout, stderr := runCLI(t, fake.env(), "orders", "count")
	if code != ExitError || stdout != "" ||
		!strings.Contains(stderr, "HEAD /order-v2/{tenant}/salesorders: the service reported no count in X-Total-Count") {
		t.Fatalf("no count: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

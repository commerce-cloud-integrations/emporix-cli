package cli

import "testing"

const pricesBody = `[
	{"id":"p1","itemId":{"itemType":"PRODUCT","id":"001234"},"currency":"EUR","location":{"countryCode":"NL"},
	 "priceModelId":"m1","restrictions":{"siteCodes":["AcmeStore_NL","main"],
	 "validity":{"from":"2026-01-01T00:00:00Z","to":"2027-01-01T00:00:00Z"}},
	 "tierValues":[{"priceValue":15.9000},{"priceValue":14.5}]},
	{"id":"p2","itemId":{"itemType":"PRODUCT","id":"ZZ1"},"currency":"EUR","tierValues":[]}
]`

const pricesTable = "ID  ITEM    TYPE     CURRENCY  COUNTRY  SITES              TIERS       FROM" +
	"                  TO\n" +
	"p1  001234  PRODUCT  EUR       NL       AcmeStore_NL,main  15.90,14.5  2026-01-01T00:00:00Z  " +
	"2027-01-01T00:00:00Z\n" +
	"p2  ZZ1     PRODUCT  EUR                                                                     \n"

const feeBody = `{"id":"f1","code":"deposit","name":{"en":"Deposit","nl":"Statiegeld"},"feeType":"ABSOLUTE",
	"feeAbsolute":{"amount":3.5,"currency":"EUR"}}`

const feeTable = "ID  CODE     NAME        TYPE      AMOUNT   ITEM TYPE  SITE  ACTIVE  TAXABLE  FROM  TO\n" +
	"f1  deposit  Statiegeld  ABSOLUTE  3.5 EUR                                          \n"

func priceReads() []readCase {
	return []readCase{
		{
			args:    []string{"prices", "list", "--site", "AcmeStore_NL"},
			path:    "/price/acme1/prices",
			query:   "siteCode=AcmeStore_NL",
			filters: listFilters{kinds: filterable},
			body:    pricesBody,
			want:    pricesTable,
		},
		{
			args:    []string{"prices", "for-product", "001 234/x", "--site", "main"},
			path:    "/price/acme1/prices",
			query:   "itemId=001+234%2Fx&itemType=PRODUCT&siteCode=main",
			filters: listFilters{kinds: filterable},
			body:    pricesBody,
			want:    pricesTable,
		},
		{
			args: []string{"prices", "get", "p2"},
			path: "/price/acme1/prices/p2",
			mode: getWhole,
			body: `{"id":"p2","itemId":{"itemType":"PRODUCT","id":"ZZ1"}}`,
			want: "ID  ITEM  TYPE     CURRENCY  COUNTRY  SITES  TIERS  FROM  TO\n" +
				"p2  ZZ1   PRODUCT                                         \n",
		},
	}
}

func priceModelReads() []readCase {
	return []readCase{
		{
			args:    []string{"price-models", "list", "--language", "nl"},
			path:    "/price/acme1/priceModels",
			scopes:  []string{"price.pricemodel_read"},
			filters: listFilters{kinds: sortable},
			body: `[
		{"id":"pm1","name":{"en":"Default price model","nl":"Standaard"},"default":true,"includesTax":false,
			"includesMarkup":false,"measurementUnit":{"quantity":1.0,"unitCode":"pc"},
			"tierDefinition":{"tierType":"BASIC","tiers":[{"id":"t1","minQuantity":{"quantity":0.0,"unitCode":"pc"}}]}},
		{"id":"pm2","name":{"en":"Tiered"},"includesTax":true,"measurementUnit":{"quantity":50,"unitCode":"g"},
			"tierDefinition":{"tierType":"TIERED","tiers":[{"minQuantity":{"quantity":0,"unitCode":"g"}},
				{"minQuantity":{"quantity":100,"unitCode":"g"}},{"minQuantity":{"quantity":500,"unitCode":"g"}}]}},
		{"id":"pm3","measurementUnit":{"unitCode":"kg"},
			"tierDefinition":{"tiers":[{"minQuantity":{"quantity":5}},{"minQuantity":{"unitCode":"pc"}}]}},
		{"id":"pm4"}]`,
			want: "ID   NAME         DEFAULT  TYPE    TIERS              UNIT    TAX    MARKUP\n" +
				"pm1  Standaard    true     BASIC   0.0 pc             1.0 pc  false  false\n" +
				"pm2  Tiered (en)           TIERED  0 g, 100 g, 500 g  50 g    true   \n" +
				"pm3                                5, pc              kg             \n" +
				"pm4                                                                  \n",
		},
		{
			args:   []string{"price-models", "get", "pm1", "--language", "nl"},
			path:   "/price/acme1/priceModels/pm1",
			scopes: []string{"price.pricemodel_read"},
			mode:   getWhole,
			body:   `{"id":"pm1","name":{"en":"Default","nl":"Standaard"},"tierDefinition":{"tierType":"BASIC"}}`,
			want: "ID   NAME       DEFAULT  TYPE   TIERS  UNIT  TAX  MARKUP\n" +
				"pm1  Standaard           BASIC                    \n",
		},
	}
}

func priceListReads() []readCase {
	return []readCase{
		{
			args:    []string{"price-lists", "list", "--language", "nl"},
			path:    "/price/acme1/price-lists",
			scopes:  []string{"price.pricelist_read"},
			filters: listFilters{kinds: filterable},
			body: `[
		{"id":"pl1","name":{"en":"Wholesale","nl":"Groothandel"},"currency":"EUR","countries":["NL","BE"],
			"customerGroups":["cg1","cg2"],"siteCode":"AcmeStore_NL",
			"validity":{"from":"2026-01-01T00:00:00.000Z","to":"2027-01-01T00:00:00.000Z"}},
		{"id":"pl2","name":{"de":"Händler"},"currency":"CHF","regions":["DACH"],"legalEntityId":"le1","siteCode":"main"},
		{"id":"pl3"}]`,
			want: "ID   NAME          CURRENCY  SITE          COUNTRIES  REGIONS  CUSTOMER GROUPS  LEGAL ENTITY  " +
				"FROM                      TO\n" +
				"pl1  Groothandel   EUR       AcmeStore_NL  NL,BE               cg1,cg2                        " +
				"2026-01-01T00:00:00.000Z  2027-01-01T00:00:00.000Z\n" +
				"pl2  Händler (de)  CHF       main                     DACH                      le1           " +
				"                          \n" +
				"pl3                                                                                           " +
				"                          \n",
		},
		{
			args:    []string{"price-lists", "prices", "list", "pl/1"},
			path:    "/price/acme1/price-lists/pl%2F1/prices",
			scopes:  []string{"price.pricelist_read"},
			filters: listFilters{kinds: filterable},
			body: `[
		{"id":"p1","itemId":{"itemType":"PRODUCT","id":"001234"},"priceModelId":"pm2",
			"tierValues":[{"id":"t1","priceValue":15.90},{"id":"t2","priceValue":14}]},
		{"id":"p2","itemId":{"itemType":"PRODUCT","id":"100200"},"tierValues":[{"priceValue":9.5}]}]`,
			want: "ID  ITEM    TYPE     MODEL  TIERS\n" +
				"p1  001234  PRODUCT  pm2    15.90,14\n" +
				"p2  100200  PRODUCT         9.5\n",
		},
		{
			args:   []string{"price-lists", "get", "pl1", "--language", "nl"},
			path:   "/price/acme1/price-lists/pl1",
			scopes: []string{"price.pricelist_read"},
			mode:   getWhole,
			body:   `{"id":"pl1","name":{"en":"Wholesale","nl":"Groothandel"},"currency":"EUR","siteCode":"main"}`,
			want: "ID   NAME         CURRENCY  SITE  COUNTRIES  REGIONS  CUSTOMER GROUPS  LEGAL ENTITY  FROM  TO\n" +
				"pl1  Groothandel  EUR       main                                                           \n",
		},
	}
}

func feeReads() []readCase {
	return []readCase{
		{
			args:  []string{"fees", "list", "--site", "main", "--language", "nl"},
			path:  "/fee/acme1/fees",
			query: "siteCode=main",
			body: `[
		{"id":"f1","code":"deposit","name":{"en":"Deposit","nl":"Statiegeld"},"feeType":"ABSOLUTE_MULTIPLY_ITEMQUANTITY",
			"feeAbsolute":{"amount":0.15,"currency":"EUR"},"itemType":"PRODUCT","siteCode":"main","active":true,
			"taxable":false,"activeTimespan":{"startDate":"2026-01-01T00:00:00Z","endDate":"2027-01-01T00:00:00Z"}},
		{"id":"f2","code":"card","name":{"en":"Card surcharge"},"feeType":"PERCENT","feePercentage":2.50,
			"itemType":"PAYMENTTYPE","siteCode":"main","active":false},
		{"id":"f3","feeType":"PERCENT","feeAbsolute":{"amount":3.5,"currency":"EUR"}},
		{"id":"f4","feeType":"ABSOLUTE","feePercentage":5,"feeAbsolute":{"currency":"EUR"}},
		{"id":"f5","feeType":"PERCENT","feePercentage":0},
		{"id":"f6","feeType":"ABSOLUTE","feeAbsolute":{"amount":0,"currency":"EUR"}}]`,
			want: "ID  CODE     NAME                 TYPE                            AMOUNT    ITEM TYPE    SITE  ACTIVE  " +
				"TAXABLE  FROM                  TO\n" +
				"f1  deposit  Statiegeld           ABSOLUTE_MULTIPLY_ITEMQUANTITY  0.15 EUR  PRODUCT      main  true    " +
				"false    2026-01-01T00:00:00Z  2027-01-01T00:00:00Z\n" +
				"f2  card     Card surcharge (en)  PERCENT                         2.50%     PAYMENTTYPE  main  false   " +
				"                               \n" +
				"f3                                PERCENT                                                              " +
				"                               \n" +
				"f4                                ABSOLUTE                                                             " +
				"                               \n" +
				"f5                                PERCENT                         0%                                   " +
				"                               \n" +
				"f6                                ABSOLUTE                        0 EUR                                " +
				"                               \n",
		},
		{
			args: []string{"fees", "get", "f1", "--language", "nl"},
			path: "/fee/acme1/fees/f1",
			mode: getWhole,
			body: `{"fee":` + feeBody + `,"itemYRNs":{"siteCode":"main","itemYrns":["urn:x"]}}`,
			want: feeTable,
		},
		{
			args: []string{"fees", "get", "f2"},
			path: "/fee/acme1/fees/f2",
			mode: getWhole,
			body: `{"id":"f2","code":"deposit","name":"Deposit","feeType":"ABSOLUTE","feeAbsolute":{"amount":3.5,
				"currency":"EUR"}}`,
			want: "ID  CODE     NAME     TYPE      AMOUNT   ITEM TYPE  SITE  ACTIVE  TAXABLE  FROM  TO\n" +
				"f2  deposit  Deposit  ABSOLUTE  3.5 EUR                                          \n",
		},
		{
			args:  []string{"fees", "for-product", "100 200", "--site", "AcmeStore_NL", "--language", "nl"},
			path:  "/fee/acme1/productFees/100%20200/fees",
			query: "expand=true&siteCode=AcmeStore_NL",
			mode:  getWhole,
			body:  "[" + feeBody + "]",
			want:  feeTable,
		},
		{
			args: []string{
				"fees", "for-item", "urn:yaas:saasag:caasproduct:product:acme1;100200", "--site", "AcmeStore_NL",
				"--language", "nl",
			},
			path:  "/fee/acme1/itemFees/urn:yaas:saasag:caasproduct:product:acme1%3B100200/fees",
			query: "expand=true&siteCode=AcmeStore_NL",
			mode:  getWhole,
			body:  "[" + feeBody + "]",
			want:  feeTable,
		},
	}
}

func availabilityReads() []readCase {
	return []readCase{
		{
			args:    []string{"availability", "list", "--site", "AcmeStore NL"},
			path:    "/availability/acme1/availability/site/AcmeStore%20NL",
			filters: listFilters{kinds: filterable},
			body: `[
		{"id":"a1","productId":"001234","site":"AcmeStore NL","stockLevel":12.0,"available":true,
		 "distributionChannel":"ASSORTMENT"},
		{"productId":"ZZ1","site":"AcmeStore NL"}]`,
			want: "PRODUCT  SITE          STOCK  AVAILABLE  CHANNEL\n" +
				"001234   AcmeStore NL  12.0   true       ASSORTMENT\n" +
				"ZZ1      AcmeStore NL                    \n",
		},
		{
			args: []string{"availability", "get", "001 234/x", "--site", "main"},
			path: "/availability/acme1/availability/001%20234%2Fx/main",
			mode: getWhole,
			body: `{"productId":"001 234/x","site":"main","stockLevel":0,"available":false}`,
			want: "PRODUCT    SITE  STOCK  AVAILABLE  CHANNEL\n001 234/x  main  0      false      \n",
		},
	}
}

func TestPriceMatchHelpSaysHowPriceListsRank(t *testing.T) {
	assertHelpHolds(t, map[string][]string{"prices match": {
		"Prices of price lists come before catalog prices",
		"rank by customer (a list of one of the customer's groups, of its legal entity or of a parent of that " +
			"entity first)",
		"then by country (a list of the target country first), then by how many of the country's regions " +
			"they name (more first)",
		"Only the first rank that has a price for the product is compared",
	}})
}

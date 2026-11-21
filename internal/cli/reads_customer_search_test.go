package cli

import "net/http"

func customerSearchReads() []readCase {
	return append(customerAndVendorSearchReads(), segmentSearchReads()...)
}

func customerAndVendorSearchReads() []readCase {
	return []readCase{
		{
			args:    []string{"customers", "search"},
			path:    "/customer/acme1/customers/search",
			scopes:  []string{"customer.customer_read"},
			mode:    getAllHybrisCounted,
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body: `[{"id":"c-7f3a","customerNumber":"10001","firstName":"Ada","lastName":"Example",
				"contactEmail":"ada@example.com","company":"Acme Ltd","businessModel":"B2B","active":true,"onHold":false},
				{"id":"c-8e2b","customerNumber":"10002","contactEmail":"bo@example.com","businessModel":"B2C"}]`,
			want: "NUMBER  NAME         EMAIL            COMPANY   BUSINESS MODEL  ACTIVE  ON HOLD\n" +
				"10001   Ada Example  ada@example.com  Acme Ltd  B2B             true    false\n" +
				"10002                bo@example.com             B2C                     \n",
		},
		{
			args:    []string{"customers", "search", "--query", `lastName:"Example" metadata.createdAt:>2026-01-01`},
			path:    "/customer/acme1/customers/search",
			scopes:  []string{"customer.customer_read"},
			mode:    getAllHybrisCounted,
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"lastName:\"Example\" metadata.createdAt:>2026-01-01"}`,
			body:    `[{"id":"c-7f3a","customerNumber":"10001","firstName":"Ada","lastName":"Example"}]`,
			want: "NUMBER  NAME         EMAIL  COMPANY  BUSINESS MODEL  ACTIVE  ON HOLD\n" +
				"10001   Ada Example                                          \n",
		},
		{
			args:    []string{"legal-entities", "search", "--query", "type:SUBSIDIARY"},
			path:    "/customer-management/acme1/legal-entities/search",
			scopes:  []string{"customermanagement.legalentity_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"type:SUBSIDIARY"}`,
			body: `[{"id":"le2","name":"Acme Retail","type":"SUBSIDIARY","parentId":"le1",
				"accountLimit":{"currency":"EUR","value":250},"legalInfo":{"legalName":"Acme Retail BV",
				"countryOfRegistration":"NL"}}]`,
			want: "ID   NAME         TYPE        PARENT  LEGAL NAME      COUNTRY  ACCOUNT LIMIT\n" +
				"le2  Acme Retail  SUBSIDIARY  le1     Acme Retail BV  NL       250 EUR\n",
		},
		{
			args:    []string{"vendors", "search", "--query", "type:COMPANY"},
			path:    "/vendor/acme1/vendors/search",
			scopes:  []string{"vendor.vendor_read"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"type:COMPANY"}`,
			body: `[{"id":"v1","name":"Pumps Inc","type":"COMPANY",
				"legalInfo":{"legalName":"Pumps Inc.","countryOfRegistration":"DE"}}]`,
			want: "ID  NAME       TYPE     PARENT  LEGAL NAME  COUNTRY\n" +
				"v1  Pumps Inc  COMPANY          Pumps Inc.  DE\n",
		},
	}
}

func segmentSearchReads() []readCase {
	const scope = "customersegment.segment_read"
	return []readCase{
		{
			args:    []string{"segments", "search", "--language", "nl"},
			path:    "/customer-segment/acme1/segments/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body: `[{"id":"s1","name":{"en":"Gold","nl":"Goud"},"status":"ACTIVE","siteCode":"main",
				"validity":{"from":"2026-01-01T00:00:00Z"}},{"id":"s2","name":{"en":"Trial"},"status":"INACTIVE"}]`,
			want: "ID  NAME        STATUS    SITE  FROM                  TO\n" +
				"s1  Goud        ACTIVE    main  2026-01-01T00:00:00Z  \n" +
				"s2  Trial (en)  INACTIVE                              \n",
		},
		{
			args:    []string{"segments", "customers", "search", "s 1", "--query", "assignmentType:MANUAL"},
			path:    "/customer-segment/acme1/segments/s%201/customers/search",
			query:   "sort=id",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"assignmentType:MANUAL"}`,
			body: `[{"customer":{"id":"c-7f3a","firstName":"Ada","lastName":"Example","email":"ada@example.com"},
				"legalEntity":{"id":"le1"},"assignmentType":"MANUAL"}]`,
			want: "CUSTOMER  NAME         EMAIL            LEGAL ENTITY  ASSIGNMENT\n" +
				"c-7f3a    Ada Example  ada@example.com  le1           MANUAL\n",
		},
		{
			args:    []string{"segments", "groups", "search", "s 1", "--language", "nl"},
			path:    "/customer-segment/acme1/segments/s%201/groups/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body: `[{"segmentId":"s 1","group":{"id":"g1","name":{"en":"Wholesale","nl":"Groothandel"}}},
				{"segmentId":"s 1","group":{"id":"g2","name":{"en":"Retail"}}}]`,
			want: "GROUP  NAME\ng1     Groothandel\ng2     Retail (en)\n",
		},
		{
			args:    []string{"segments", "items", "search", "s 1", "--query", "type:PRODUCT", "--language", "nl"},
			path:    "/customer-segment/acme1/segments/s%201/items/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"type:PRODUCT"}`,
			body: `[{"segmentId":"s 1","type":"PRODUCT","item":{"id":"p1","code":"100200",
				"name":{"en":"Pump","nl":"Pomp"}}}]`,
			want: "ID  SEGMENT  TYPE     CODE    NAME\np1  s 1      PRODUCT  100200  Pomp\n",
		},
	}
}

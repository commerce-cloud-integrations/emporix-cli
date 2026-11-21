package cli

import "net/http"

func customerReads() []readCase {
	return []readCase{
		{
			args:    []string{"customers", "list"},
			path:    "/customer/acme1/customers",
			scopes:  []string{"customer.customer_read"},
			mode:    getAllHybrisCounted,
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"c-7f3a","customerNumber":"10001","firstName":"Ada","middleName":"B.","lastName":"Example",
			"contactEmail":"ada@example.com","company":"Acme Ltd","businessModel":"B2B",
			"active":true,"onHold":false},
		{"id":"c-8e2b","customerNumber":"10002","firstName":"Bo","contactEmail":"bo@example.com","businessModel":"B2C",
			"active":false,"onHold":true},
		{"id":"c-9d1c","customerNumber":"10003","email":"login@example.com"}]`,
			want: "NUMBER  NAME            EMAIL            COMPANY   BUSINESS MODEL  ACTIVE  ON HOLD\n" +
				"10001   Ada B. Example  ada@example.com  Acme Ltd  B2B             true    false\n" +
				"10002   Bo              bo@example.com             B2C             false   true\n" +
				"10003                                                                      \n",
		},
		{
			args:   []string{"customers", "addresses", "list", "100/01"},
			path:   "/customer/acme1/customers/100%2F01/addresses",
			scopes: []string{"customer.customer_read"},
			mode:   getWhole,
			response: `[
		{"id":"a1","contactName":"Ada Example","companyName":"Acme Ltd","street":"Main Street","streetNumber":"1",
			"zipCode":"1000 AA","city":"Amsterdam","country":"NL","isDefault":true,"tags":["BILLING","SHIPPING"]},
		{"id":"a2","contactName":"Ada Example","street":"Side Road","country":"BE","isDefault":false}]`,
			want: "ID  CONTACT      COMPANY   STREET         ZIP      CITY       COUNTRY  DEFAULT  TAGS\n" +
				"a1  Ada Example  Acme Ltd  Main Street 1  1000 AA  Amsterdam  NL       true     BILLING,SHIPPING\n" +
				"a2  Ada Example            Side Road                          BE       false    \n",
		},
		{
			args:     []string{"customers", "get", "100/01"},
			path:     "/customer/acme1/customers/100%2F01",
			scopes:   []string{"customer.customer_read"},
			mode:     getWhole,
			response: `{"id":"c-7f3a","customerNumber":"100/01","firstName":"Ada","lastName":"Example"}`,
			want: "NUMBER  NAME         EMAIL  COMPANY  BUSINESS MODEL  ACTIVE  ON HOLD\n" +
				"100/01  Ada Example                                          \n",
		},
		{
			args:     []string{"customers", "password-migration", "get"},
			path:     "/customer/acme1/config/password-migration-retention",
			scopes:   []string{"customer.import_read"},
			mode:     getWhole,
			response: `{"retentionEndDate":"2027-01-31","emailReminderDate":"2027-01-24","emailNotificationsEnabled":false}`,
			want:     "ENDS        REMINDER    EMAILS\n2027-01-31  2027-01-24  false\n",
		},
	}
}

func sessionContextReads() []readCase {
	return []readCase{
		{
			args:   []string{"session-contexts", "get", "s 1/2", "--allow-write-scope"},
			path:   "/session-context/acme1/context/s%201%2F2",
			scopes: []string{"sessioncontext.context_manage"},
			mode:   getWhole,
			response: `{"sessionId":"s 1/2","customerId":"C100","siteCode":"AcmeStore_NL","currency":"EUR","language":"nl",
				"targetLocation":"NL","cartId":"cart1","context":{"channel":"web"},"metadata":{"version":2}}`,
			want: "ID     CUSTOMER  SITE          CURRENCY  LANGUAGE  LOCATION  CART\n" +
				"s 1/2  C100      AcmeStore_NL  EUR       nl        NL        cart1\n",
		},
	}
}

func legalEntityReads() []readCase {
	return []readCase{
		{
			args:    []string{"legal-entities", "list"},
			path:    "/customer-management/acme1/legal-entities",
			scopes:  []string{"customermanagement.legalentity_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"le1","name":"Acme","type":"COMPANY","accountLimit":{"currency":"EUR","value":1000.99},
			"legalInfo":{"legalName":"Acme Ltd","countryOfRegistration":"NL","taxRegistrationNumber":"NL001"}},
		{"id":"le2","name":"Acme Retail","type":"SUBSIDIARY","parentId":"le1"},
		{"id":"le3"}]`,
			want: "ID   NAME         TYPE        PARENT  LEGAL NAME  COUNTRY  ACCOUNT LIMIT\n" +
				"le1  Acme         COMPANY             Acme Ltd    NL       1000.99 EUR\n" +
				"le2  Acme Retail  SUBSIDIARY  le1                          \n" +
				"le3                                                        \n",
		},
		{
			args:     []string{"legal-entities", "get", "le/2"},
			path:     "/customer-management/acme1/legal-entities/le%2F2",
			scopes:   []string{"customermanagement.legalentity_read"},
			mode:     getWhole,
			response: `{"id":"le/2","name":"Acme Retail","type":"SUBSIDIARY","parentId":"le1"}`,
			want: "ID    NAME         TYPE        PARENT  LEGAL NAME  COUNTRY  ACCOUNT LIMIT\n" +
				"le/2  Acme Retail  SUBSIDIARY  le1                          \n",
		},
		{
			args:   []string{"legal-entities", "hierarchy", "le/2"},
			path:   "/customer-management/acme1/legal-entities/le%2F2/parent-hierarchy",
			scopes: []string{"customermanagement.legalentity_read"},
			mode:   getWhole,
			response: `[{"id":"le/2","name":"Acme Retail","type":"SUBSIDIARY","parentId":"le1"},
				{"id":"le1","name":"Acme","type":"COMPANY"}]`,
			want: "ID    NAME         TYPE        PARENT  LEGAL NAME  COUNTRY  ACCOUNT LIMIT\n" +
				"le/2  Acme Retail  SUBSIDIARY  le1                          \n" +
				"le1   Acme         COMPANY                                  \n",
		},
	}
}

func customerManagementReads() []readCase {
	return []readCase{
		{
			args:    []string{"customer-locations", "list"},
			path:    "/customer-management/acme1/locations",
			scopes:  []string{"customermanagement.location_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"loc1","name":"Head office","type":"HEADQUARTER","contactDetails":{"emails":["info@example.com",
			"sales@example.com"],"phones":["+31 20 000"],"addressLine1":"Main Street 1","postcode":"1000 AA",
			"city":"Amsterdam","countryCode":"NL","tags":["billing"]}},
		{"id":"loc2","name":"Depot","type":"WAREHOUSE"}]`,
			want: "ID    NAME         TYPE         ADDRESS        POSTCODE  CITY       COUNTRY  " +
				"EMAILS                              TAGS\n" +
				"loc1  Head office  HEADQUARTER  Main Street 1  1000 AA   Amsterdam  NL       " +
				"info@example.com,sales@example.com  billing\n" +
				"loc2  Depot        WAREHOUSE                                                 " +
				"                                    \n",
		},
		{
			args:     []string{"customer-locations", "get", "loc/1"},
			path:     "/customer-management/acme1/locations/loc%2F1",
			scopes:   []string{"customermanagement.location_read"},
			mode:     getWhole,
			response: `{"id":"loc/1","name":"Head office","type":"OFFICE"}`,
			want: "ID     NAME         TYPE    ADDRESS  POSTCODE  CITY  COUNTRY  EMAILS  TAGS\n" +
				"loc/1  Head office  OFFICE                                            \n",
		},
		{
			args:    []string{"contact-assignments", "list"},
			path:    "/customer-management/acme1/contact-assignments",
			scopes:  []string{"customermanagement.contactassignment_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"ca1","legalEntity":{"id":"le1","name":"Acme"},"type":"PRIMARY","primary":true,
			"customer":{"id":"c-7f3a","name":"Ada","surname":"Example","email":"ada@example.com","phone":"+31 20 000"}},
		{"id":"ca2","legalEntity":{"id":"le1"},"customer":{"id":"c-8e2b"},"type":"BILLING","primary":false}]`,
			want: "ID   LEGAL ENTITY  CUSTOMER  NAME         EMAIL            TYPE     PRIMARY\n" +
				"ca1  le1           c-7f3a    Ada Example  ada@example.com  PRIMARY  true\n" +
				"ca2  le1           c-8e2b                                  BILLING  false\n",
		},
		{
			args:     []string{"contact-assignments", "get", "ca/1"},
			path:     "/customer-management/acme1/contact-assignments/ca%2F1",
			scopes:   []string{"customermanagement.contactassignment_read"},
			mode:     getWhole,
			response: `{"id":"ca/1","legalEntity":{"id":"le1"},"customer":{"id":"c-7f3a"},"type":"CONTACT"}`,
			want: "ID    LEGAL ENTITY  CUSTOMER  NAME  EMAIL  TYPE     PRIMARY\n" +
				"ca/1  le1           c-7f3a                 CONTACT  \n",
		},
	}
}

func segmentReads() []readCase {
	return []readCase{
		{
			args:    []string{"segments", "list", "--language", "nl"},
			path:    "/customer-segment/acme1/segments",
			scopes:  []string{"customersegment.segment_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"s1","name":{"en":"Gold","nl":"Goud"},"status":"ACTIVE","siteCode":"main",
			"validity":{"from":"2026-01-01T00:00:00Z","to":"2026-12-31T00:00:00Z"}},
		{"id":"s2","name":{"en":"Trial"},"status":"INACTIVE"}]`,
			want: "ID  NAME        STATUS    SITE  FROM                  TO\n" +
				"s1  Goud        ACTIVE    main  2026-01-01T00:00:00Z  2026-12-31T00:00:00Z\n" +
				"s2  Trial (en)  INACTIVE                              \n",
		},
		{
			args:     []string{"segments", "get", "s/1", "--language", "nl"},
			path:     "/customer-segment/acme1/segments/s%2F1",
			scopes:   []string{"customersegment.segment_read"},
			mode:     getWhole,
			response: `{"id":"s/1","name":{"en":"Gold","nl":"Goud"},"status":"ACTIVE"}`,
			want:     "ID   NAME  STATUS  SITE  FROM  TO\ns/1  Goud  ACTIVE              \n",
		},
		{
			args:    []string{"segments", "customers", "list", "s/1"},
			path:    "/customer-segment/acme1/segments/s%2F1/customers",
			query:   "sort=id",
			scopes:  []string{"customersegment.segment_read"},
			filters: listFilters{kinds: filterable},
			response: `[
			{"customer":{"id":"c-7f3a","firstName":"Ada","lastName":"Example","email":"ada@example.com"},
				"legalEntity":{"id":"le1","name":"Acme"},"assignmentType":"MANUAL"},
			{"customer":{"id":"c-8e2b"},"assignmentType":"AUTOMATIC"}]`,
			want: "CUSTOMER  NAME         EMAIL            LEGAL ENTITY  ASSIGNMENT\n" +
				"c-7f3a    Ada Example  ada@example.com  le1           MANUAL\n" +
				"c-8e2b                                                AUTOMATIC\n",
		},
		{
			args:     []string{"segments", "groups", "list", "s/1", "--language", "nl"},
			path:     "/customer-segment/acme1/segments/s%2F1/groups",
			scopes:   []string{"customersegment.segment_read"},
			filters:  listFilters{kinds: filterable},
			response: `[{"segmentId":"s/1","group":{"id":"g1","name":{"en":"Wholesale","nl":"Groothandel"}}}]`,
			want:     "GROUP  NAME\ng1     Groothandel\n",
		},
		{
			args:    []string{"segments", "items", "list", "s/1", "--language", "nl"},
			path:    "/customer-segment/acme1/segments/s%2F1/items",
			scopes:  []string{"customersegment.segment_read"},
			filters: listFilters{kinds: filterable},
			response: `[{"segmentId":"s/1","type":"PRODUCT","item":{"id":"p1","code":"100200","name":{"en":"Pump"}}},
			{"segmentId":"s/1","type":"CATEGORY","item":{"id":"cat1","name":{"en":"Tools","nl":"Tuinmeubels"}}}]`,
			want: "ID    SEGMENT  TYPE      CODE    NAME\n" +
				"p1    s/1      PRODUCT   100200  Pump (en)\n" +
				"cat1  s/1      CATEGORY          Tuinmeubels\n",
		},
		{
			args:    []string{"segments", "all-items", "--language", "nl"},
			path:    "/customer-segment/acme1/segments/items",
			scopes:  []string{"customersegment.segment_read"},
			filters: listFilters{kinds: filterable},
			response: `[{"segmentId":"s2","type":"PRODUCT","item":{"id":"p2","code":"100300","name":{"en":"Valve",
				"nl":"Klep"}}}]`,
			want: "ID  SEGMENT  TYPE     CODE    NAME\np2  s2       PRODUCT  100300  Klep\n",
		},
		{
			args: []string{"segments", "match", "--product", "100200", "--product", "P 2", "--segment", "S1",
				"--site", "AcmeStore_NL", "--allow-write-scope"},
			path:        "/customer-segment/acme1/segments/match",
			scopes:      []string{"customersegment.segment_manage"},
			mode:        getWhole,
			method:      http.MethodPost,
			requestBody: `{"items":[{"id":"100200"},{"id":"P 2"}],"segmentIds":["S1"],"siteCode":"AcmeStore_NL"}`,
			response:    `[{"id":"100200"},{"id":"P 2"}]`,
			want:        "PRODUCT\n100200\nP 2\n",
		},
		{
			args: []string{"segments", "match", "--product", "100200", "--product", "100200", "--segment", "S1",
				"--site", "AcmeStore_NL", "--allow-write-scope"},
			path:        "/customer-segment/acme1/segments/match",
			scopes:      []string{"customersegment.segment_manage"},
			mode:        getWhole,
			method:      http.MethodPost,
			requestBody: `{"items":[{"id":"100200"},{"id":"100200"}],"segmentIds":["S1"],"siteCode":"AcmeStore_NL"}`,
			response:    `[{"id":"100200"},{"id":"100200"}]`,
			want:        "PRODUCT\n100200\n100200\n",
		},
	}
}

func vendorReads() []readCase {
	return []readCase{
		{
			args:    []string{"vendors", "list"},
			path:    "/vendor/acme1/vendors",
			scopes:  []string{"vendor.vendor_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"v1","name":"Pumps Inc","type":"COMPANY","legalInfo":{"legalName":"Pumps Inc.","countryOfRegistration":"DE"}},
		{"id":"v2","name":"Pumps NL","type":"SUBSIDIARY","parentId":"v1"}]`,
			want: "ID  NAME       TYPE        PARENT  LEGAL NAME  COUNTRY\n" +
				"v1  Pumps Inc  COMPANY             Pumps Inc.  DE\n" +
				"v2  Pumps NL   SUBSIDIARY  v1                  \n",
		},
		{
			args:    []string{"vendor-locations", "list"},
			path:    "/vendor/acme1/locations",
			scopes:  []string{"vendor.vendor_read"},
			filters: listFilters{kinds: filterable},
			response: `[
		{"id":"vl1","name":"Depot","type":"WAREHOUSE","contactDetails":{"emails":["depot@example.com"],
			"addressLine1":"Dock 4","postcode":"20457","city":"Hamburg","countryCode":"DE","tags":["pickup"]}}]`,
			want: "ID   NAME   TYPE       ADDRESS  POSTCODE  CITY     COUNTRY  EMAILS             TAGS\n" +
				"vl1  Depot  WAREHOUSE  Dock 4   20457     Hamburg  DE       depot@example.com  pickup\n",
		},
		{
			args:     []string{"vendors", "get", "v/1"},
			path:     "/vendor/acme1/vendors/v%2F1",
			scopes:   []string{"vendor.vendor_read"},
			mode:     getWhole,
			response: `{"id":"v/1","name":"Pumps Inc","type":"COMPANY"}`,
			want: "ID   NAME       TYPE     PARENT  LEGAL NAME  COUNTRY\n" +
				"v/1  Pumps Inc  COMPANY                      \n",
		},
		{
			args:     []string{"vendor-locations", "get", "vl/1"},
			path:     "/vendor/acme1/locations/vl%2F1",
			scopes:   []string{"vendor.vendor_read"},
			mode:     getWhole,
			response: `{"id":"vl/1","name":"Depot","type":"OFFICE"}`,
			want: "ID    NAME   TYPE    ADDRESS  POSTCODE  CITY  COUNTRY  EMAILS  TAGS\n" +
				"vl/1  Depot  OFFICE                                            \n",
		},
	}
}

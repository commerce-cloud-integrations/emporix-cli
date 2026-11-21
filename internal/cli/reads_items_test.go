package cli

func nestedItemReads() []readCase {
	return []readCase{
		{
			args:   []string{"price-lists", "prices", "get", "pl/1", "p/1"},
			path:   "/price/acme1/price-lists/pl%2F1/prices/p%2F1",
			scopes: []string{"price.pricelist_read"},
			mode:   getWhole,
			response: `{"id":"p/1","itemId":{"itemType":"PRODUCT","id":"100200"},"priceModelId":"pm2",
				"tierValues":[{"id":"t1","priceValue":15.90},{"id":"t2","priceValue":14}]}`,
			want: "ID   ITEM    TYPE     MODEL  TIERS\np/1  100200  PRODUCT  pm2    15.90,14\n",
		},
		{
			args: []string{"shipping", "zones", "methods", "get", "z/1", "m 1", "--site", "AcmeStore_NL", "--language",
				"nl"},
			path: "/shipping/acme1/AcmeStore_NL/zones/z%2F1/methods/m%201",
			mode: getWhole,
			response: `{"id":"m 1","name":{"en":"Express","nl":"Snel"},"active":true,"shippingTaxCode":"STANDARD",
				"fees":[{"cost":{"amount":7.50,"currency":"EUR"},"minOrderValue":{"amount":0,"currency":"EUR"}}]}`,
			want: "ID   NAME  ACTIVE  FEES                 MAX ORDER  TAX CODE  GROUP\n" +
				"m 1  Snel  true    7.50 EUR from 0 EUR             STANDARD  \n",
		},
		{
			args:   []string{"delivery-times", "slots", "get", "d/1", "s/1"},
			path:   "/shipping/acme1/delivery-times/d%2F1/slots/s%2F1",
			scopes: []string{"shipping.shipping_read"},
			mode:   getWhole,
			response: `{"id":"s/1","shippingMethod":"standard","deliveryTimeRange":{"timeFrom":"08:00","timeTo":"12:00"},
				"cutOffTime":{"time":"2026-01-01T06:00:00.000Z","cutOffDayShift":1},"capacity":40}`,
			want: "ID   METHOD    FROM   TO     CUT-OFF                   CUT-OFF DAYS  CAPACITY\n" +
				"s/1  standard  08:00  12:00  2026-01-01T06:00:00.000Z  1             40\n",
		},
		{
			args:   []string{"customers", "addresses", "get", "0100/01", "a/1"},
			path:   "/customer/acme1/customers/0100%2F01/addresses/a%2F1",
			scopes: []string{"customer.customer_read"},
			mode:   getWhole,
			response: `{"id":"a/1","contactName":"Ada Example","street":"Main Street","streetNumber":"1","zipCode":"1000 AA",
				"city":"Amsterdam","country":"NL","isDefault":true,"tags":["BILLING","SHIPPING"]}`,
			want: "ID   CONTACT      COMPANY  STREET         ZIP      CITY       COUNTRY  DEFAULT  TAGS\n" +
				"a/1  Ada Example           Main Street 1  1000 AA  Amsterdam  NL       true     BILLING,SHIPPING\n",
		},
	}
}

func segmentMemberReads() []readCase {
	return []readCase{
		{
			args:   []string{"segments", "customers", "get", "s/1", "c-7f3a"},
			path:   "/customer-segment/acme1/segments/s%2F1/customers/c-7f3a",
			scopes: []string{"customersegment.segment_read"},
			mode:   getWhole,
			response: `{"customer":{"id":"c-7f3a","firstName":"Ada","lastName":"Example","email":"ada@example.com"},
				"assignmentType":"MANUAL"}`,
			want: "CUSTOMER  NAME         EMAIL            LEGAL ENTITY  ASSIGNMENT\n" +
				"c-7f3a    Ada Example  ada@example.com                MANUAL\n",
		},
		{
			args:   []string{"segments", "customers", "get-b2b", "s/1", "c 7f3a", "le/1"},
			path:   "/customer-segment/acme1/segments/s%2F1/customers/c%207f3a/le%2F1",
			scopes: []string{"customersegment.segment_read"},
			mode:   getWhole,
			response: `{"customer":{"id":"c 7f3a","firstName":"Ada"},"legalEntity":{"id":"le/1","name":"Acme"},
				"assignmentType":"AUTOMATIC"}`,
			want: "CUSTOMER  NAME  EMAIL  LEGAL ENTITY  ASSIGNMENT\n" +
				"c 7f3a    Ada          le/1          AUTOMATIC\n",
		},
		{
			args:     []string{"segments", "groups", "get", "s/1", "g/1", "--language", "nl"},
			path:     "/customer-segment/acme1/segments/s%2F1/groups/g%2F1",
			scopes:   []string{"customersegment.segment_read"},
			mode:     getWhole,
			response: `{"segmentId":"s/1","group":{"id":"g/1","name":{"en":"Wholesale","nl":"Groothandel"}}}`,
			want:     "GROUP  NAME\ng/1    Groothandel\n",
		},
		{
			args:     []string{"segments", "items", "get", "s/1", "PRODUCT", "p/1", "--language", "nl"},
			path:     "/customer-segment/acme1/segments/s%2F1/items/PRODUCT/p%2F1",
			scopes:   []string{"customersegment.segment_read"},
			mode:     getWhole,
			response: `{"segmentId":"s/1","type":"PRODUCT","item":{"id":"p/1","code":"100200","name":{"en":"Pump"}}}`,
			want:     "ID   SEGMENT  TYPE     CODE    NAME\np/1  s/1      PRODUCT  100200  Pump (en)\n",
		},
	}
}

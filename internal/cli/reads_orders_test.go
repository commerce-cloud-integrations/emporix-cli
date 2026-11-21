package cli

import (
	"net/http"
	"testing"
)

func orderReads() []readCase {
	return []readCase{
		{
			args:    []string{"orders", "list"},
			path:    "/order-v2/acme1/salesorders",
			scopes:  []string{"order.order_read"},
			filters: listFilters{kinds: filterable},
			body: `[
		{"id":"EON1001","orderType":"MASTER_ORDER","status":"CONFIRMED","created":"2026-10-01T10:00:00.000Z",
			"siteCode":"main","customer":{"id":"10001","email":"ada@example.com","firstName":"Ada","lastName":"Example"},
			"entries":[{"id":"e1"},{"id":"e2"}],"currency":"EUR",
			"calculatedPrice":{"finalPrice":{"netValue":100.0,"grossValue":121.00}},"totalPrice":99},
		{"id":"EON1001-1","orderType":"SUB_ORDER","status":"CREATED",
			"customer":{"firstName":"Bo","middleName":"B.","lastName":"Example"},"entries":[],"totalPrice":15.5,
			"currency":"CHF"},
		{"id":"EON1003","customer":{"name":"Acme Ltd"}}]`,
			want: "ID         TYPE          STATUS     CREATED                   CUSTOMER         SITE  ITEMS  TOTAL\n" +
				"EON1001    MASTER_ORDER  CONFIRMED  2026-10-01T10:00:00.000Z  ada@example.com  main  2      121.00 EUR\n" +
				"EON1001-1  SUB_ORDER     CREATED                              Bo B. Example          0      \n" +
				"EON1003                                                       Acme Ltd               0      \n",
		},
		{
			args:   []string{"orders", "get", "EON/1"},
			path:   "/order-v2/acme1/salesorders/EON%2F1",
			scopes: []string{"order.order_read"},
			mode:   getWhole,
			body: `{"id":"EON/1","status":"SHIPPED","entries":[{"id":"e1"}],"currency":"EUR",
				"calculatedPrice":{"finalPrice":{"grossValue":10}}}`,
			want: "ID     TYPE  STATUS   CREATED  CUSTOMER  SITE  ITEMS  TOTAL\n" +
				"EON/1        SHIPPED                           1      10 EUR\n",
		},
		{
			args:   []string{"orders", "transitions", "EON/1"},
			path:   "/order-v2/acme1/salesorders/EON%2F1/transitions",
			scopes: []string{"order.order_read"},
			mode:   getWhole,
			body:   `[{"status":"SHIPPED"},{"status":"DECLINED"}]`,
			want:   "STATUS\nSHIPPED\nDECLINED\n",
		},
		{
			args:   []string{"orders", "history", "EON/1"},
			path:   "/order-v2/acme1/salesorders/EON%2F1/historical-transitions",
			scopes: []string{"order.order_read"},
			mode:   getWhole,
			body: `{"transitions":[{"status":"CREATED","timestamp":"2026-10-01T10:00:00.000Z"},
				{"status":"CONFIRMED","timestamp":"2026-10-01T11:00:00.000Z"}],"metadata":{"version":2}}`,
			want: "STATUS     TIMESTAMP\n" +
				"CREATED    2026-10-01T10:00:00.000Z\n" +
				"CONFIRMED  2026-10-01T11:00:00.000Z\n",
		},
	}
}

func deliveryTimeReads() []readCase {
	return []readCase{
		{
			args:   []string{"delivery-times", "list"},
			path:   "/shipping/acme1/delivery-times",
			scopes: []string{"shipping.shipping_read"},
			mode:   getWhole,
			body: `[
		{"id":"d1","name":"Weekdays","siteCode":"main","zoneId":"z1","isDeliveryDay":true,
			"day":{"weekday":"MONDAY"},"timeZoneId":"Europe/Amsterdam","slots":[{"id":"s1"},{"id":"s2"}]},
		{"id":"d2","siteCode":"main","isForAllZones":true,"isDeliveryDay":false,
			"day":{"singleDate":"2026-12-25T11:00:00.000Z"},"timeZoneId":"Europe/Amsterdam","deliveryDayShift":2},
		{"id":"d3","siteCode":"main","zoneId":"z1",
			"day":{"datePeriod":{"dateFrom":"2026-07-01T10:00:00.000Z","dateTo":"2026-07-31T10:00:00.000Z"}}}]`,
			want: "ID  NAME      SITE  ZONE       DAY                                                " +
				"DELIVERS  TIME ZONE         DAY SHIFT  SLOTS\n" +
				"d1  Weekdays  main  z1         MONDAY                                             " +
				"true      Europe/Amsterdam             2\n" +
				"d2            main  all zones  2026-12-25T11:00:00.000Z                           " +
				"false     Europe/Amsterdam  2          0\n" +
				"d3            main  z1         2026-07-01T10:00:00.000Z/2026-07-31T10:00:00.000Z  " +
				"                                       0\n",
		},
		{
			args:   []string{"delivery-times", "slots", "list", "d/1"},
			path:   "/shipping/acme1/delivery-times/d%2F1/slots",
			scopes: []string{"shipping.shipping_read"},
			mode:   getWhole,
			body: `[
		{"id":"s1","shippingMethod":"standard","deliveryTimeRange":{"timeFrom":"08:00","timeTo":"12:00"},
			"cutOffTime":{"time":"2026-01-01T06:00:00.000Z","cutOffDayShift":1},"capacity":40},
		{"id":"s2","shippingMethod":"express"}]`,
			want: "ID  METHOD    FROM   TO     CUT-OFF                   CUT-OFF DAYS  CAPACITY\n" +
				"s1  standard  08:00  12:00  2026-01-01T06:00:00.000Z  1             40\n" +
				"s2  express                                                         \n",
		},
		{
			args:   []string{"delivery-times", "get", "d/1"},
			path:   "/shipping/acme1/delivery-times/d%2F1",
			scopes: []string{"shipping.shipping_read"},
			mode:   getWhole,
			body: `{"id":"d/1","siteCode":"main","zoneId":"z1","isDeliveryDay":true,"day":{"weekday":"FRIDAY"},
				"timeZoneId":"Europe/Amsterdam","slots":[{"id":"s1","shippingMethod":"standard"}]}`,
			want: "ID   NAME  SITE  ZONE  DAY     DELIVERS  TIME ZONE         DAY SHIFT  SLOTS\n" +
				"d/1        main  z1    FRIDAY  true      Europe/Amsterdam             1\n",
		},
		{
			args: []string{"delivery-times", "windows", "cart/1"},
			path: "/shipping/acme1/actualDeliveryWindows/cart%2F1",
			mode: getWhole,
			body: `[
		{"id":"w1","deliveryDate":"2026-06-20T12:00:00.000Z","deliveryTimeRange":{"startTime":"15:00","endTime":"23:00"},
			"deliveryMethod":"standard","zoneId":"z1","cutOffTime":"2026-06-20T11:00:00.000Z","slotId":"s1",
			"deliveryCycle":"2026-06-20-11"},
		{"deliveryDate":"2026-06-24T12:00:00.000Z","deliveryTimeRange":{"startTime":"12:00","endTime":"16:00"}}]`,
			want: "DATE                      FROM   TO     METHOD    ZONE  CUT-OFF                   SLOT  ID\n" +
				"2026-06-20T12:00:00.000Z  15:00  23:00  standard  z1    2026-06-20T11:00:00.000Z  s1    w1\n" +
				"2026-06-24T12:00:00.000Z  12:00  16:00                                                  \n",
		},
	}
}

func shippingReads() []readCase {
	return []readCase{
		{
			args: []string{"shipping", "zones", "list", "--site", "Acme Store", "--language", "nl"},
			path: "/shipping/acme1/Acme%20Store/zones",
			mode: getWhole,
			body: `[
		{"id":"z1","default":true,"name":{"en":"Standard","nl":"Standaard"},
			"shipTo":[{"country":"NL","postalCode":"*"},{"country":"BE"},{"country":"DE","postalCode":"10*"}]},
		{"id":"z2","name":{"de":"Inseln"}}]`,
			want: "ID  NAME         DEFAULT  SHIPS TO\n" +
				"z1  Standaard    true     NL, BE, DE 10*\n" +
				"z2  Inseln (de)           \n",
		},
		{
			args: []string{"shipping", "zones", "get", "z/1", "--site", "main", "--language", "nl"},
			path: "/shipping/acme1/main/zones/z%2F1",
			mode: getWhole,
			body: `{"id":"z/1","name":{"en":"Standard","nl":"Standaard"},"shipTo":[{"country":"NL"}]}`,
			want: "ID   NAME       DEFAULT  SHIPS TO\nz/1  Standaard           NL\n",
		},
		{
			args:  []string{"shipping", "zones", "methods", "list", "z1", "--site", "main", "--language", "nl"},
			path:  "/shipping/acme1/main/zones/z1/methods",
			query: "expand=fees",
			mode:  getWhole,
			body: `[
			{"id":"m1","name":{"en":"Express","nl":"Snel"},"active":true,"shippingTaxCode":"STANDARD",
				"maxOrderValue":{"amount":1000,"currency":"EUR"},"fees":[
				{"cost":{"amount":7.50,"currency":"EUR"},"minOrderValue":{"amount":0,"currency":"EUR"}},
				{"cost":{"amount":0,"currency":"EUR"},"minOrderValue":{"amount":50,"currency":"EUR"},"shippingGroupId":"vip"}]},
			{"id":"m2","name":"Pickup","active":false,"shippingGroupId":"vip",
				"fees":[{"cost":{"amount":2,"currency":"EUR"}}]}]`,
			want: "ID  NAME    ACTIVE  FEES                                          MAX ORDER  TAX CODE  GROUP\n" +
				"m1  Snel    true    7.50 EUR from 0 EUR, 0 EUR from 50 EUR (vip)  1000 EUR   STANDARD  \n" +
				"m2  Pickup  false   2 EUR                                                              vip\n",
		},
		{
			args: []string{"shipping", "groups", "list", "--site", "main", "--language", "nl"},
			path: "/shipping/acme1/main/groups",
			mode: getWhole,
			body: `[{"id":"vip","description":{"en":"Free shipping","nl":"Gratis verzending"}}]`,
			want: "ID   DESCRIPTION\nvip  Gratis verzending\n",
		},
		{
			args: []string{"shipping", "groups", "get", "vip", "--site", "main", "--language", "nl"},
			path: "/shipping/acme1/main/groups/vip",
			mode: getWhole,
			body: `{"id":"vip","description":{"en":"Free shipping","nl":"Gratis verzending"}}`,
			want: "ID   DESCRIPTION\nvip  Gratis verzending\n",
		},
		{
			args: []string{"shipping", "customers", "list", "--site", "main"},
			path: "/shipping/acme1/main/cgrelations",
			mode: getWhole,
			body: `[{"customerId":"c1","groupId":"vip"}]`,
			want: "CUSTOMER  GROUP\nc1        vip\n",
		},
		{
			args: []string{"shipping", "customers", "get", "c1", "--site", "main"},
			path: "/shipping/acme1/main/cgrelations/c1",
			mode: getWhole,
			body: `{"customerId":"c1","groupId":"vip"}`,
			want: "CUSTOMER  GROUP\nc1        vip\n",
		},
	}
}

const paymentModesBody = `[
	{"id":"pm1","code":"paypal","active":true,"provider":"SPREEDLY","configuration":{
		"Spreedly/GatewayToken":"S1","Spreedly/EnvironmentKey":"env","Spreedly/AccessSecret":"S2",
		"Spreedly/PaymentMethodType":"paypal","Spreedly/SignatureSecret":"S3","Spreedly/SCAProviderToken":"S4"}},
	{"id":"pm2","code":"card","active":false,"provider":"SPREEDLY_SAFERPAY","configuration":{
		"Saferpay/BasicAuth":"S5","Saferpay/CustomerId":"123","Saferpay/RedirectUrl":"https://shop.example"}},
	{"id":"pm3","code":"later","provider":"UNZER","configuration":{"Unzer/PublicKey":"pub","Unzer/PrivateKey":"S6"}}]`

func paymentModeReads() []readCase {
	return []readCase{
		{
			args:   []string{"payment-modes", "list"},
			path:   "/payment-gateway/acme1/paymentmodes/config",
			scopes: []string{"payment-gateway.paymentmodes_read"},
			mode:   getWhole,
			body:   paymentModesBody,
			want: "ID   CODE    PROVIDER           ACTIVE\n" +
				"pm1  paypal  SPREEDLY           true\n" +
				"pm2  card    SPREEDLY_SAFERPAY  false\n" +
				"pm3  later   UNZER              \n",
		},
		{
			args: []string{"payment-modes", "public", "list"},
			path: "/payment-gateway/acme1/paymentmodes/frontend",
			mode: getWhole,
			body: `[{"id":"pm1","code":"paypal","integrationType":"OFFSITE","paymentMethodType":"paypal",
				"environmentKey":"env"}]`,
			want: "ID   CODE    INTEGRATION  METHOD\npm1  paypal  OFFSITE      paypal\n",
		},
		{
			args: []string{"payment-modes", "public", "get", "pm1"},
			path: "/payment-gateway/acme1/paymentmodes/frontend/pm1",
			mode: getWhole,
			body: `[{"id":"pm1","code":"paypal","integrationType":"OFFSITE","paymentMethodType":"paypal"}]`,
			want: "ID   CODE    INTEGRATION  METHOD\npm1  paypal  OFFSITE      paypal\n",
		},
		{
			args:   []string{"payment-modes", "get", "pm 2"},
			path:   "/payment-gateway/acme1/paymentmodes/config/pm%202",
			scopes: []string{"payment-gateway.paymentmodes_read"},
			mode:   getWhole,
			body:   `{"id":"pm 2","code":"card","active":false,"provider":"SPREEDLY_SAFERPAY"}`,
			want:   "ID    CODE  PROVIDER           ACTIVE\npm 2  card  SPREEDLY_SAFERPAY  false\n",
		},
	}
}

const transactionsBody = `[
	{"id":"t1","provider":"SPREEDLY","paymentInformation":{"paymentMode":"pm1","creditCardToken":"S1"},
	 "entries":[
		{"id":"e1","createdDate":"2026-10-14T09:19:42.170Z","type":"AUTHORIZATION","successful":true,
		 "amount":{"amount":36.63,"currency":"EUR"},"summary":"Succeeded!",
		 "remoteRequest":"{\n  \"transaction\" : {\n    \"amount\" : 3663,\n    \"payment_method_token\" : \"S2\"\n  }\n}",
		 "remoteResponse":"{\"transaction\":{\"token\":\"S3\",\"succeeded\":true,\"gateway_token\":\"S4\"}}",
		 "remoteToken":"S5"},
		{"id":"e2","type":"CAPTURE","successful":false,"amount":{"amount":36.63,"currency":"EUR"},
		 "summary":"Declined"}]},
	{"id":"t2","provider":"INVOICE","paymentInformation":{"paymentMode":"pm2"},"entries":[]}]`

func paymentTransactionReads() []readCase {
	return []readCase{
		{
			args:    []string{"payment-transactions", "list"},
			path:    "/payment-gateway/acme1/transactions",
			scopes:  []string{"payment-gateway.paymenttransactions_read"},
			filters: listFilters{kinds: sortable},
			body:    transactionsBody,
			want: "ID  PROVIDER  PAYMENT MODE  ENTRIES\n" +
				"t1  SPREEDLY  pm1           AUTHORIZATION 36.63 EUR, CAPTURE 36.63 EUR failed\n" +
				"t2  INVOICE   pm2           \n",
		},
		{
			args:   []string{"payment-transactions", "get", "t/1"},
			path:   "/payment-gateway/acme1/transactions/t%2F1",
			scopes: []string{"payment-gateway.paymenttransactions_read"},
			mode:   getWhole,
			body:   `{"id":"t/1","provider":"UNZER","entries":[{"type":"REFUND","amount":{"amount":5,"currency":"CHF"}}]}`,
			want:   "ID   PROVIDER  PAYMENT MODE  ENTRIES\nt/1  UNZER                   REFUND 5 CHF\n",
		},
	}
}

func returnReads() []readCase {
	return []readCase{
		{
			args:    []string{"returns", "list"},
			path:    "/return/acme1/returns",
			scopes:  []string{"returns.returns_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"r1","approvalStatus":"PENDING","siteCode":"DE","restriction":"DE","received":false,
				"orders":[{"id":"EON1001","items":[{"id":"1","quantity":2}]},{"id":"EON1002","items":[]}],
				"reason":{"code":"001"},"requestor":{"customerId":"c1","firstName":"Jane","lastName":"Doe",
				"email":"jane@example.com"},"total":{"value":18.90,"currency":"EUR"},
				"metadata":{"createdAt":"2026-10-14T09:00:00.000Z","version":1}},
				{"id":"r2","approvalStatus":"APPROVED","received":true,"orders":[{"id":"EON1003","items":[]}],
				"requestor":{"firstName":"John","lastName":"Roe"}},
				{"id":"r3","requestor":{"customerId":"c3"}},
				{"id":"r4","requestor":{"anonymous":true,"email":"guest@example.com"}}]`,
			want: "ID  STATUS    SITE  RECEIVED  ORDERS            REQUESTOR          TOTAL      CREATED\n" +
				"r1  PENDING   DE    false     EON1001, EON1002  jane@example.com   18.90 EUR  2026-10-14T09:00:00.000Z\n" +
				"r2  APPROVED        true      EON1003           John Roe                      \n" +
				"r3                                              c3                            \n" +
				"r4                                              guest@example.com             \n",
		},
		{
			args:   []string{"returns", "get", "r/1"},
			path:   "/return/acme1/returns/r%2F1",
			scopes: []string{"returns.returns_read"},
			mode:   getWhole,
			body:   `{"id":"r/1","approvalStatus":"CLOSED","siteCode":"NL","orders":[{"id":"EON1001","items":[]}]}`,
			want: "ID   STATUS  SITE  RECEIVED  ORDERS   REQUESTOR  TOTAL  CREATED\n" +
				"r/1  CLOSED  NL              EON1001                    \n",
		},
	}
}

func quoteReads() []readCase {
	return []readCase{
		{
			args:    []string{"quotes", "list"},
			path:    "/quote/acme1/quotes",
			scopes:  []string{"quote.quote_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"q1","status":{"value":"OPEN"},"customer":{"customerId":"c1","firstName":"Jane",
				"lastName":"Doe","contactEmail":"jane@example.com"},"company":{"name":"Acme Ltd"},"siteCode":"main",
				"validTo":"2026-11-01T00:00:00.000Z","totalPrice":{"currency":"EUR","netValue":10,"grossValue":12.10}},
				{"id":"q2","status":{"value":"AWAITING"},"customer":{"firstName":"John","lastName":"Roe"}},
				{"id":"q3","customer":{"customerId":"c3"}}]`,
			want: "ID  STATUS    CUSTOMER          COMPANY   SITE  VALID TO                  TOTAL\n" +
				"q1  OPEN      jane@example.com  Acme Ltd  main  2026-11-01T00:00:00.000Z  12.10 EUR\n" +
				"q2  AWAITING  John Roe                                                    \n" +
				"q3            c3                                                          \n",
		},
		{
			args:   []string{"quotes", "get", "q/1"},
			path:   "/quote/acme1/quotes/q%2F1",
			scopes: []string{"quote.quote_read"},
			mode:   getWhole,
			body:   `{"id":"q/1","status":{"value":"ACCEPTED"},"currency":"CHF","totalPrice":{"grossValue":5}}`,
			want: "ID   STATUS    CUSTOMER  COMPANY  SITE  VALID TO  TOTAL\n" +
				"q/1  ACCEPTED                                     5 CHF\n",
		},
		{
			args: []string{"quotes", "pdf", "Q1001", "--out", "-"}, path: "/quote/acme1/quotes/Q1001/pdf",
			scopes: []string{"quote.quote_read"}, mode: getWhole, method: http.MethodPost,
			media: "application/pdf", body: quotePDF, want: quotePDF,
		},
		{
			args:   []string{"quotes", "pdf", "Q1002", "--out", "-", "--language", "de"},
			path:   "/quote/acme1/quotes/Q1002/pdf",
			scopes: []string{"quote.quote_read"}, mode: getWhole, method: http.MethodPost,
			media: "application/pdf; qs=0.9", body: quotePDF, want: quotePDF,
		},
		{
			args:   []string{"quotes", "history", "q/1"},
			path:   "/quote/acme1/quotes/q%2F1/history",
			scopes: []string{"quote.quote_read"},
			mode:   getWhole,
			body: `[{"id":"h1","op":"REPLACE","path":"/status","newValue":{"value":"OPEN"},
				"previousValue":{"value":"AWAITING"},"userId":"u1","userFirstName":"Robin","userLastName":"Fritz",
				"userType":"EMPLOYEE","modifiedAt":"2026-10-14T13:06:15.280Z"},{"id":"h2","op":"CREATE","path":"/quote"}]`,
			want: "ID  MODIFIED                  BY                    CHANGE\n" +
				"h1  2026-10-14T13:06:15.280Z  EMPLOYEE Robin Fritz  REPLACE /status\n" +
				"h2                                                  CREATE /quote\n",
		},
		{
			args:    []string{"quotes", "reasons", "list", "--language", "de"},
			path:    "/quote/acme1/quote-reasons",
			scopes:  []string{"quote.quote_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"r1","code":"WRONG_MATERIAL","type":"CHANGE","message":{"en":"Wrong material",
				"de":"Falsches Material"}},{"id":"r2","code":"OTHER","type":"DECLINE","message":{"en":"Other"}}]`,
			want: "ID  CODE            TYPE     MESSAGE\n" +
				"r1  WRONG_MATERIAL  CHANGE   Falsches Material\n" +
				"r2  OTHER           DECLINE  Other (en)\n",
		},
		{
			args:   []string{"quotes", "reasons", "get", "r1", "--language", "en"},
			path:   "/quote/acme1/quote-reasons/r1",
			scopes: []string{"quote.quote_read"},
			mode:   getWhole,
			body:   `{"id":"r1","code":"WRONG_MATERIAL","type":"CHANGE","message":{"en":"Wrong material"}}`,
			want:   "ID  CODE            TYPE    MESSAGE\nr1  WRONG_MATERIAL  CHANGE  Wrong material\n",
		},
	}
}

func shoppingListReads() []readCase {
	return []readCase{
		{
			args:   []string{"shopping-lists", "list"},
			path:   "/shoppinglist/acme1/shopping-lists",
			scopes: []string{"shoppinglist.shoppinglist_read"},
			mode:   getWhole,
			body: `[{"id":"c1","metadata":{"version":2,"createdAt":{"epochSecond":1760000000,"nano":0},"mixins":{},
				"schema":"x"},"wishlist":{"customerId":"c1","id":"a1","name":"wishlist","metadata":{"version":1},
				"mixins":{},"items":[{"id":1,"productId":"100200","quantity":1,"mixins":{}}]},
				"default":{"name":"default","items":[{"id":1,"productId":"100200","quantity":7.1},
				{"id":2,"productId":"100201","quantity":1.5}]},"Top Secret":{"name":"Top Secret","items":[]}},
				{"id":"c2","empty":{"items":[]},"unnamed":{"name":"unnamed"},
				"metadata":{"name":"metadata","items":[{"id":1}]}}]`,
			want: "CUSTOMER  LISTS\n" +
				"c1        Top Secret (0 items), default (2 items), wishlist (1 item)\n" +
				"c2        empty (0 items), metadata (1 item), unnamed (0 items)\n",
		},
		{
			args:   []string{"shopping-lists", "get", "c/1"},
			path:   "/shoppinglist/acme1/shopping-lists/c%2F1",
			scopes: []string{"shoppinglist.shoppinglist_read"},
			mode:   getWhole,
			body:   `{"customerId":"c/1","default":{"name":"default","items":[]}}`,
			want:   "CUSTOMER  LISTS\nc/1       default (0 items)\n",
		},
	}
}

func TestReturnsHelpNamesTheirSiteAndAnswers(t *testing.T) {
	assertHelpHolds(t, map[string][]string{
		"returns list": {
			"SITE is the site of the return's orders, which Emporix copies to it (siteCode;",
			"`--query siteCode:DE` those of one site",
			"so a credential that holds only a site-limited one, such as returns.returns_read--DE, fails with exit code 3",
			"and for its writes returns.returns_manage",
		},
		"returns create": {
			"Emporix copies siteCode and restriction from the orders, so the body may leave them out",
			"orders of different sites or restrictions",
			"`\"site does not match order\"`", "`\"orders belong to different sites\"`",
		},
		"returns update": {
			"siteCode and restriction stay as copied from the orders: Emporix ignores them in an update, " +
				"whether the body has them or not",
		},
		"returns patch": {"accepts an operation on /siteCode or /restriction but leaves them as they are, though " +
			"metadata.version still goes up"},
		"returns delete": {"its reference now documents 404 Not Found (exit 4)"},
	})
}

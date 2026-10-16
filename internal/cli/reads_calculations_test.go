package cli

import (
	"net/http"
	"strings"
)

// The read cases of the calculations and checks that Emporix serves as POST
// or PUT with a JSON body but that change nothing: shipping quotes, the sites
// that ship to a postal code, delivery window and coupon validations, unit
// conversions and tax calculations (see TestReads).

// calculationReads returns the calculations of the Shipping, Unit Handling,
// Tax and Coupon services.
func calculationReads() []readCase {
	return append(shippingCalculationReads(),
		readCase{
			args: []string{"delivery-times", "validate", "--delivery-time", "d 1", "--date", "2026-11-02T12:00:00.000Z",
				"--slot", "s1"},
			path:   "/shipping/acme1/deliveryWindowValidation",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"deliveryDate":"2026-11-02T12:00:00.000Z","deliveryWindowId":"d 1","slotId":"s1"}`,
			body:   `true`,
			want:   "VALID\ntrue\n",
		},
		readCase{
			// Emporix answers with the name as bare text (TestDeliveryCycleNames).
			args: []string{"delivery-times", "generate-cycle", "--delivery-time", "d1", "--slot", "s1", "--date",
				"2026-10-19T12:00:00.000Z"},
			path:   "/shipping/acme1/delivery-cycles/generate",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"deliveryDate":"2026-10-19T12:00:00.000Z","deliveryTimeId":"d1","slotId":"s1"}`,
			body:   "2026-10-19-6morning",
			want:   "CYCLE\n2026-10-19-6morning\n",
		},
		readCase{
			// The amount is sent as given; Emporix answers 201 Created (TestOnlyDeclaredReadsTakeCreated).
			args:   []string{"units", "convert", "--from", "g", "--to", "kg", "--amount", "100.50"},
			path:   "/unit-handling/acme1/units/convert-unit-commands",
			mode:   getWhole,
			method: http.MethodPut,
			sent:   `{"input":{"sourceUnit":"g","sourceUnitAmount":100.50,"targetUnit":"kg"}}`,
			body: `{"commandUuid":"u1","input":{"sourceUnitAmount":100.50,"sourceUnit":"g","targetUnit":"kg"},
				"output":{"targetUnitAmount":0.1005,"targetUnit":"kg"}}`,
			want: "FROM      TO\n100.50 g  0.1005 kg\n",
		},
		readCase{
			args:   []string{"units", "conversion-factor", "--from", "us-fl-oz", "--to", "cl"},
			path:   "/unit-handling/acme1/units/conversion-factor-commands",
			mode:   getWhole,
			method: http.MethodPut,
			sent:   `{"input":{"sourceUnit":"us-fl-oz","targetUnit":"cl"}}`,
			body: `{"commandUuid":"u2","input":{"sourceUnit":"us-fl-oz","targetUnit":"cl"},
				"output":{"conversionFactor":2.95735296}}`,
			want: "FROM      TO  FACTOR\nus-fl-oz  cl  2.95735296\n",
		},
		readCase{
			// A net price: includesTax is sent as false.
			args:   []string{"taxes", "calculate", "--price", "1.59", "--country", "DE", "--tax-class", "STANDARD"},
			path:   "/tax/acme1/taxes/calculation-commands",
			mode:   getWhole,
			method: http.MethodPut,
			sent: `{"input":{"includesTax":false,"price":1.59,"targetLocation":{"countryCode":"DE"},` +
				`"targetTaxClass":"STANDARD"}}`,
			body: `{"commandUuid":"t1","input":{"targetLocation":{"countryCode":"DE"},"targetTaxClass":"STANDARD",
				"includesTax":false,"price":1.59},"output":{"netPrice":1.59,"grossPrice":1.892100,"targetTaxRate":19}}`,
			want: "NET   GROSS   SOURCE RATE  TARGET RATE\n1.59  1.8921               19\n",
		},
		readCase{
			// A gross price in one country to a gross price in another.
			args: []string{"taxes", "calculate", "--price", "1.59", "--country", "PL", "--tax-class", "STANDARD", "--gross",
				"--source-country", "DE", "--source-tax-class", "REDUCED"},
			path:   "/tax/acme1/taxes/calculation-commands",
			mode:   getWhole,
			method: http.MethodPut,
			sent: `{"input":{"includesTax":true,"price":1.59,"sourceLocation":{"countryCode":"DE"},` +
				`"sourceTaxClass":"REDUCED","targetLocation":{"countryCode":"PL"},"targetTaxClass":"STANDARD"}}`,
			body: `{"commandUuid":"t2","output":{"netPrice":1.5,"grossPrice":1.845,"sourceTaxRate":6,"targetTaxRate":23}}`,
			want: "NET  GROSS  SOURCE RATE  TARGET RATE\n1.5  1.845  6            23\n",
		},
		readCase{
			// Emporix answers a redeemable coupon with an empty 200 response.
			args: []string{"coupons", "validate", "SAVE 7", "--order-total", "1090 EUR", "--discount", "5.00 EUR",
				"--customer", "C1", "--allow-write-scope"},
			path:   "/coupon/acme1/coupons/SAVE%207/validation",
			scopes: []string{"coupon.coupon_redeem_on_behalf"},
			mode:   getWhole,
			method: http.MethodPost,
			sent: `{"customerNumber":"C1","discount":{"amount":5.00,"currency":"EUR"},` +
				`"orderTotal":{"amount":1090,"currency":"EUR"}}`,
			want: "REDEEMABLE\ntrue\n",
		},
	)
}

// shippingCalculationReads returns the shipping quotes and the sites that
// ship to a postal code.
func shippingCalculationReads() []readCase {
	return []readCase{
		{
			// One row per method of every zone, a zone without methods none.
			args: []string{"shipping", "quote", "--site", "AcmeStore_NL", "--cart-total", "100.50 EUR", "--country", "NL",
				"--postal-code", "1012 AB", "--customer", "c1", "--language", "nl"},
			path:   "/shipping/acme1/AcmeStore_NL/quote",
			mode:   getWhole,
			method: http.MethodPost,
			sent: `{"cartTotal":{"amount":100.50,"currency":"EUR"},"customerId":"c1",` +
				`"shipToAddress":{"country":"NL","zipCode":"1012 AB"}}`,
			body: `[{"zone":{"id":"nl","name":{"en":"Netherlands"}},"methods":[
				{"id":"post","name":{"en":"Post","nl":"Bezorging"},"fee":{"amount":4.99,"currency":"EUR"},
					"shippingTaxCode":"STANDARD"},
				{"id":"pickup","name":"Pickup","fee":{"amount":0,"currency":"EUR"}}]},
				{"zone":{"id":"benelux"},"methods":[]},
				{"zone":{"id":"eu"},"methods":[{"id":"dhl","name":{"en":"DHL"},"fee":{"amount":9.5,"currency":"EUR"}}]}]`,
			want: "ZONE  METHOD  NAME       FEE       TAX CODE\n" +
				"nl    post    Bezorging  4.99 EUR  STANDARD\n" +
				"nl    pickup  Pickup     0 EUR     \n" +
				"eu    dhl     DHL (en)   9.5 EUR   \n",
		},
		{
			args: []string{"shipping", "quote-minimum", "--site", "main", "--cart-total", "100 EUR", "--country", "NL",
				"--postal-code", "1012"},
			path:   "/shipping/acme1/main/quote/minimum",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"cartTotal":{"amount":100,"currency":"EUR"},"shipToAddress":{"country":"NL","zipCode":"1012"}}`,
			body:   `{"fee":{"amount":4.99,"currency":"EUR"},"shippingTaxCode":"STANDARD"}`,
			want:   "FEE       TAX CODE\n4.99 EUR  STANDARD\n",
		},
		{
			// The destination is optional for a slot.
			args: []string{"shipping", "quote-slot", "--site", "main", "--cart-total", "100 EUR", "--delivery-time", "d1",
				"--slot", "s1"},
			path:   "/shipping/acme1/main/quote/slot",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"cartTotal":{"amount":100,"currency":"EUR"},"deliveryWindowId":"d1","slotId":"s1"}`,
			body:   `{"fee":{"amount":2.5,"currency":"EUR"}}`,
			want:   "FEE      TAX CODE\n2.5 EUR  \n",
		},
		{
			// Emporix answers the sites whole, unpaged.
			args:   []string{"shipping", "find-sites", "--postal-code", "1012", "--country", "NL"},
			path:   "/shipping/acme1/findSite",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"country":"NL","postalCode":"1012"}`,
			body: `[{"id":"AcmeStore_NL","defaultZoneId":"nl","zones":[{"id":"nl","methods":[{"id":"post"},{"id":"pickup"}]},
				{"id":"benelux","methods":[{"id":"post"}]}]},{"id":"main","zones":[]}]`,
			want: "SITE          ZONES        METHODS\nAcmeStore_NL  nl, benelux  post, pickup\n" +
				"main" + strings.Repeat(" ", 23) + "\n",
		},
	}
}

package cli

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestCalculationFlagsAreRequiredWithoutBody checks that a calculation's
// required typed flag must be given unless --body gives the body, that the
// two exclude each other, that flags of one group are given together and
// that a flag that applies only with another is refused without it: usage
// errors (exit 2) that send nothing.
func TestCalculationFlagsAreRequiredWithoutBody(t *testing.T) {
	body := writeBody(t, `{"input":{}}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"units", "convert", "--from", "g", "--to", "kg"},
			want: "at least one of the flags in the group [body amount] is required"},
		{args: []string{"shipping", "quote", "--site", "main", "--cart-total", "1 EUR", "--country", "NL"},
			want: "at least one of the flags in the group [body postal-code] is required"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "DE"},
			want: "at least one of the flags in the group [body tax-class] is required"},
		{args: []string{"coupons", "validate", "SAVE7", "--order-total", "1 EUR", "--allow-write-scope"},
			want: "at least one of the flags in the group [body discount] is required"},
		{args: []string{"delivery-times", "validate", "--delivery-time", "d1", "--date", "2026-11-02T12:00:00.000Z"},
			want: "at least one of the flags in the group [body slot] is required"},
		{args: []string{"delivery-times", "generate-cycle", "--delivery-time", "d1", "--slot", "s1"},
			want: "at least one of the flags in the group [body date] is required"},
		{args: []string{"delivery-times", "generate-cycle", "--slot", "s1", "--body", body},
			want: "if any flags in the group [body slot] are set none of the others can be"},
		{args: []string{"taxes", "calculate", "--body", body, "--gross"},
			want: "if any flags in the group [body gross] are set none of the others can be"},
		{args: []string{"shipping", "quote-slot", "--site", "main", "--cart-total", "1 EUR", "--delivery-time", "d1",
			"--slot", "s1", "--country", "NL"},
			want: "if any flags in the group [country postal-code] are set they must all be set; missing [postal-code]"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "DE", "--tax-class", "STANDARD", "--gross",
			"--source-tax-class", "REDUCED"},
			want: "group [source-country source-tax-class] are set they must all be set; missing [source-country]"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "PL", "--tax-class", "STANDARD",
			"--source-country", "DE", "--source-tax-class", "REDUCED"},
			want: "Error: --source-country applies only with --gross\n"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "PL", "--tax-class", "STANDARD",
			"--gross=false", "--source-country", "DE", "--source-tax-class", "REDUCED"},
			want: "Error: --source-country applies only with --gross\n"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "PL", "--tax-class", "STANDARD",
			"--gross", "--gross=false", "--source-country", "DE", "--source-tax-class", "REDUCED"},
			want: "Error: --source-country applies only with --gross\n"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) ||
			len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, requests %q; want exit 2, %q and nothing sent",
				tt.args, code, stdout, stderr, fake.gets, tt.want)
		}
	}
}

// TestCalculationBodyReplacesTheFlags checks that --body sends a whole body
// as given, members the typed flags do not set included.
func TestCalculationBodyReplacesTheFlags(t *testing.T) {
	const given = `{"input":{"targetLocation":{"countryCode":"DE"},"targetTaxClass":"STANDARD","price":2,` +
		`"includesTax":false},"commandUuid":"c1"}`
	fake := newFakeEmporix(t)
	fake.route("/tax/acme1/taxes/calculation-commands", respond(http.StatusOK,
		`{"output":{"netPrice":2,"grossPrice":2.38,"targetTaxRate":19}}`))
	code, stdout, stderr := runCLI(t, fake.env(), "taxes", "calculate", "--body", writeBody(t, given))
	if code != ExitOK || stdout != "NET  GROSS  SOURCE RATE  TARGET RATE\n2    2.38                19\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertSent(t, http.MethodPut, given)
}

// TestCalculationValuesAreCheckedWhileParsing checks that a number or an
// amount of money in another form is a usage error found before anything is
// sent, naming the flag.
func TestCalculationValuesAreCheckedWhileParsing(t *testing.T) {
	convert := []string{"units", "convert", "--from", "g", "--to", "kg", "--amount"}
	quote := []string{"shipping", "quote-minimum", "--site", "main", "--country", "NL", "--postal-code", "1012",
		"--cart-total"}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: append(slices.Clone(convert), "ten"), want: `invalid argument "ten" for "--amount" flag: use a number`},
		{args: append(slices.Clone(convert), ".5"), want: `invalid argument ".5" for "--amount" flag`},
		{args: append(slices.Clone(convert), "0x10"), want: `invalid argument "0x10" for "--amount" flag`},
		{args: append(slices.Clone(convert), "NaN"), want: `invalid argument "NaN" for "--amount" flag`},
		{args: append(slices.Clone(convert), "1 "), want: `invalid argument "1 " for "--amount" flag`},
		{args: append(slices.Clone(convert), ""), want: `invalid argument "" for "--amount" flag`},
		{args: append(slices.Clone(quote), "100"), want: `"100" for "--cart-total" flag: use AMOUNT CURRENCY`},
		{args: append(slices.Clone(quote), "EUR 100"), want: `invalid argument "EUR 100" for "--cart-total" flag`},
		{args: append(slices.Clone(quote), "1 EUR x"), want: `invalid argument "1 EUR x" for "--cart-total" flag`},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) || len(fake.gets) != 0 {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, requests %q; want exit 2 and %q",
				tt.args, code, stdout, stderr, fake.gets, tt.want)
		}
	}
}

// TestOnlyDeclaredReadsTakeCreated checks that a unit conversion reads the
// 201 Created that Emporix answers it with (live), although it creates
// nothing, as generate-cycle does (TestDeliveryCycleNames), while any other
// read with a body fails on it (exit 1): a coupon validation answered so may
// have redeemed the coupon.
func TestOnlyDeclaredReadsTakeCreated(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, body string
		code       int
		want       string // stdout, or what stderr holds when the read fails
	}{
		{args: []string{"units", "conversion-factor", "--from", "g", "--to", "kg"},
			path: "/unit-handling/acme1/units/conversion-factor-commands",
			body: `{"input":{"sourceUnit":"g","targetUnit":"kg"},"output":{"conversionFactor":0.001}}`,
			want: "FROM  TO  FACTOR\ng     kg  0.001\n"},
		{args: []string{"units", "convert", "--from", "g", "--to", "kg", "--amount", "1"},
			path: "/unit-handling/acme1/units/convert-unit-commands",
			body: `{"input":{"sourceUnitAmount":1,"sourceUnit":"g"},"output":{"targetUnitAmount":0.001,"targetUnit":"kg"}}`,
			want: "FROM  TO\n1 g   0.001 kg\n"},
		{args: []string{"taxes", "calculate", "--price", "1", "--country", "DE", "--tax-class", "STANDARD"},
			path: "/tax/acme1/taxes/calculation-commands", body: `{"output":{"netPrice":1}}`, code: ExitError,
			want: "PUT /tax/acme1/taxes/calculation-commands: 201 Created"},
		{args: []string{"coupons", "validate", "SAVE7", "--order-total", "10 EUR", "--discount", "1 EUR",
			"--allow-write-scope"},
			path: "/coupon/acme1/coupons/SAVE7/validation", body: `{"id":"r1"}`, code: ExitError,
			want: "POST /coupon/acme1/coupons/SAVE7/validation: 201 Created"},
		{args: []string{"products", "search"}, path: "/product/acme1/products/search", body: `[]`, code: ExitError,
			want: "POST /product/acme1/products/search: 201 Created"},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusCreated, tt.body))
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != tt.code || (code == ExitOK && stdout != tt.want) ||
			(code != ExitOK && (stdout != "" || !strings.Contains(stderr, tt.want))) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.args, code, stdout, stderr, tt.code,
				tt.want)
		}
	}
}

// TestCouponValidationAnswers checks what a coupon validation prints for each
// answer: a success without a body, 204 No Content included, is redeemable,
// true (also for -o json), one with a body is read as true or false, anything
// else fails, and a failure is the error Emporix gives.
func TestCouponValidationAnswers(t *testing.T) {
	validate := []string{"coupons", "validate", "SAVE7", "--order-total", "10 EUR", "--discount", "1 EUR",
		"--allow-write-scope"}
	for _, tt := range []struct {
		answer http.HandlerFunc
		json   bool // with -o json
		code   int
		want   string // stdout, or what stderr holds when the read fails
	}{
		{answer: respond(http.StatusOK, ""), want: "REDEEMABLE\ntrue\n"},
		{answer: respond(http.StatusOK, " \n"), json: true, want: "true\n"},
		{answer: respond(http.StatusNoContent, ""), json: true, want: "true\n"},
		{answer: respond(http.StatusOK, "true"), json: true, want: "true\n"},
		{answer: respond(http.StatusOK, "false"), want: "REDEEMABLE\nfalse\n"},
		{answer: respond(http.StatusOK, "false"), json: true, want: "false\n"},
		{answer: respond(http.StatusOK, `{"code":"SAVE7"}`), code: ExitError,
			want: "decoding coupon validation response: expected true or false"},
		{answer: respond(http.StatusOK, `{"code":"SAVE7"}`), json: true, code: ExitError,
			want: "decoding coupon validation response: expected true or false"},
		{answer: respond(http.StatusOK, "OK"), code: ExitError,
			want: "decoding coupon validation response: expected true or false"},
		{answer: respond(http.StatusBadRequest, `{"message":"Coupon has expired"}`), code: ExitError,
			want: "POST /coupon/acme1/coupons/SAVE7/validation: 400 Bad Request: Coupon has expired"},
		{answer: respond(http.StatusNotFound, `{"message":"Coupon not found"}`), code: ExitNotFound,
			want: "404 Not Found: Coupon not found"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/coupon/acme1/coupons/SAVE7/validation", tt.answer)
		args := slices.Clone(validate)
		if tt.json {
			args = append(args, "-o", "json")
		}
		code, stdout, stderr := runCLI(t, fake.env(), args...)
		if code != tt.code || (code == ExitOK && stdout != tt.want) ||
			(code != ExitOK && (stdout != "" || !strings.Contains(stderr, tt.want))) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit %d and %q", args, code, stdout, stderr, tt.code, tt.want)
		}
	}
}

// TestCalculationResponsesAreChecked checks that a calculation's answer of
// another shape fails (exit 1), also for -o json, such as a quote's zone or
// method that is not a JSON object, while a delivery window that cannot take
// an order shows as false; a table also rejects a method without its zone or
// its id, as it rejects any item without an identifier.
func TestCalculationResponsesAreChecked(t *testing.T) {
	validate := []string{"delivery-times", "validate", "--delivery-time", "d1", "--date", "2026-11-02T12:00:00.000Z",
		"--slot", "s1"}
	quote := []string{"shipping", "quote", "--site", "main", "--cart-total", "1 EUR", "--country", "NL",
		"--postal-code", "1012"}
	for _, tt := range []struct {
		args       []string
		path, body string
		code       int
		want       string // stdout, or what stderr holds when the read fails
		tableOnly  bool   // -o json prints the response
	}{
		{args: validate, path: "/shipping/acme1/deliveryWindowValidation", body: "false", want: "VALID\nfalse\n"},
		{args: validate, path: "/shipping/acme1/deliveryWindowValidation", body: `{"valid":true}`, code: ExitError,
			want: "decoding delivery window validation response: expected true or false"},
		{args: validate, path: "/shipping/acme1/deliveryWindowValidation", body: "null", code: ExitError,
			want: "expected true or false"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `{"zone":{}}`, code: ExitError,
			want: "decoding shipping quote response: expected a JSON array"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `["nl"]`, code: ExitError,
			want: "decoding shipping quote response: expected zones as JSON objects"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `[null]`, code: ExitError,
			want: "decoding shipping quote response: expected zones as JSON objects"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `[{"zone":{"id":"nl"},"methods":"post"}]`,
			code: ExitError, want: "decoding shipping quote response: expected zones as JSON objects"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `[{"zone":{"id":"nl"},"methods":[null]}]`,
			code: ExitError, want: "decoding shipping quote response: expected shipping methods as JSON objects"},
		{args: quote, path: "/shipping/acme1/main/quote", body: `[{"zone":{"id":"nl"},"methods":[{"name":"Post"}]}]`,
			code: ExitError, want: "decoding shipping quote response: item 1 has no identifier", tableOnly: true},
		{args: quote, path: "/shipping/acme1/main/quote", body: `[{"methods":[{"id":"post"}]}]`, code: ExitError,
			want: "decoding shipping quote response: item 1 has no identifier", tableOnly: true},
	} {
		for _, format := range []string{"table", "json"} {
			if format == "json" && (tt.code == ExitOK || tt.tableOnly) {
				continue
			}
			fake := newFakeEmporix(t)
			fake.route(tt.path, respond(http.StatusOK, tt.body))
			code, stdout, stderr := runCLI(t, fake.env(), append(slices.Clone(tt.args), "-o", format)...)
			if code != tt.code || (code == ExitOK && stdout != tt.want) ||
				(code != ExitOK && (stdout != "" || !strings.Contains(stderr, tt.want))) {
				t.Errorf("%q -o %s on %s: exit %d, stdout %q, stderr %q; want exit %d and %q",
					tt.args, format, tt.body, code, stdout, stderr, tt.code, tt.want)
			}
		}
	}
}

// TestBodyFlagValues checks the flag values of numbers and amounts of money:
// their type names, and that they show what they were set to.
func TestBodyFlagValues(t *testing.T) {
	var number jsonNumber
	var amount moneyValue
	if number.Set("-1.50e2") != nil || amount.Set(" 12.500  EUR ") != nil {
		t.Fatal("valid values rejected")
	}
	if number.String() != "-1.50e2" || number.Type() != "number" || amount.String() != "12.500 EUR" ||
		amount.Type() != "money" {
		t.Fatalf("number %q (%s), amount %q (%s)", number.String(), number.Type(), amount.String(), amount.Type())
	}
}

// TestOverlappingBodyMembersPanic checks that a read declaring two fields
// that set one member, or a member and one inside it, panics while the
// command is built, as its body would depend on the order of a map, and that
// members that only start alike do not.
func TestOverlappingBodyMembersPanic(t *testing.T) {
	for _, tt := range []struct {
		members []string
		want    string // the panic; empty when the members do not overlap
	}{
		{members: []string{"q", "q"}, want: `body members "q" and "q" overlap`},
		{members: []string{"input", "input.price"}, want: `body members "input" and "input.price" overlap`},
		{members: []string{"input.price.amount", "input.price"},
			want: `body members "input.price.amount" and "input.price" overlap`},
		{members: []string{"input", "inputs.price", "input2"}},
	} {
		var fields []bodyField
		for i, member := range tt.members {
			fields = append(fields, stringField(member, "flag"+strconv.Itoa(i), "a member"))
		}
		got := func() (recovered any) {
			defer func() { recovered = recover() }()
			bodyRead{method: http.MethodPost, body: objectBody(fields), fetch: single{items: oneObject}}.bind(&cobra.Command{})
			return nil
		}()
		if (tt.want == "" && got != nil) || (tt.want != "" && got != tt.want) {
			t.Errorf("%q: panic %v; want %q", tt.members, got, tt.want)
		}
	}
}

// TestDeliveryCycleNames checks what generate-cycle prints for each answer:
// the bare text that Emporix answers with 201 Created (live), trimmed, as the
// name, which -o json prints as a JSON string, a JSON string as sent, and
// that any other JSON, a blank name or no answer fails.
func TestDeliveryCycleNames(t *testing.T) {
	args := []string{"delivery-times", "generate-cycle", "--delivery-time", "d1", "--slot", "s1", "--date",
		"2026-10-19T12:00:00.000Z"}
	for _, tt := range []struct {
		body, format string
		code         int
		want         string // stdout, or what stderr holds when the read fails
	}{
		{body: "2026-10-19-6morning\n", format: "table", want: "CYCLE\n2026-10-19-6morning\n"},
		{body: "2026-10-19-6morning", format: "json", want: "\"2026-10-19-6morning\"\n"},
		{body: `"2026-10-19-6 morning"`, format: "table", want: "CYCLE\n2026-10-19-6 morning\n"},
		{body: `"2026-10-19-6morning"`, format: "json", want: "\"2026-10-19-6morning\"\n"},
		{body: `{"name":"2026-10-19-6morning"}`, format: "table", code: ExitError,
			want: "decoding delivery cycle response: expected a name"},
		{body: `{"name":"2026-10-19-6morning"}`, format: "json", code: ExitError, want: "expected a name"},
		{body: `" "`, format: "table", code: ExitError, want: "expected a name"},
		{body: "null", format: "json", code: ExitError, want: "expected a name"},
		{body: "", format: "table", code: ExitNotFound, want: "empty response"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/shipping/acme1/delivery-cycles/generate", respond(http.StatusCreated, tt.body))
		code, stdout, stderr := runCLI(t, fake.env(), append(args, "-o", tt.format)...)
		if code != tt.code || (code == ExitOK && stdout != tt.want) ||
			(code != ExitOK && (stdout != "" || !strings.Contains(stderr, tt.want))) {
			t.Errorf("%q as %s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.body, tt.format, code, stdout,
				stderr, tt.code, tt.want)
		}
	}
}

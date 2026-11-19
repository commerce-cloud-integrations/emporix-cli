package cli

import (
	"net/http"
	"strings"
	"testing"
)

// returnWriteCases are the writes of the Returns Service: returns.
func returnWriteCases() []writeCase {
	const returns = "/return/acme1/returns"
	const back = `{"orders":[{"id":"O1001","items":[{"id":"e1","quantity":1}]}],"reason":{"code":"001"},` +
		`"requestor":{"customerId":"C1"},"submitter":{"userType":"USER","email":"agent@acme.example"}}`
	manageReturns := []string{"returns.returns_manage"}
	return []writeCase{
		{
			args: []string{"returns", "create"}, file: back, method: http.MethodPost, path: returns,
			scopes: manageReturns, status: http.StatusCreated, body: `{"id":"R1"}`, want: "{\n  \"id\": \"R1\"\n}\n",
		},
		{
			args: []string{"returns", "update", "R1"}, file: back, method: http.MethodPut, path: returns + "/R1",
			scopes: manageReturns, status: http.StatusNoContent,
		},
		{
			args: []string{"returns", "patch", "R1"},
			file: `[{"op":"replace","path":"/approvalStatus","value":"APPROVED"}]`, method: http.MethodPatch,
			path: returns + "/R1", scopes: []string{"returns.returns_read", "returns.returns_manage"},
			status: http.StatusNoContent,
		},
		{
			args: []string{"returns", "delete", "R1"}, method: http.MethodDelete, path: returns + "/R1",
			scopes: manageReturns, confirmBy: "R1", status: http.StatusNoContent,
		},
	}
}

// quoteWriteCases are the writes of the Quote Service: quotes (and their
// PDF) and quote reasons.
func quoteWriteCases() []writeCase {
	const quotes, reasons = "/quote/acme1/quotes", "/quote/acme1/quote-reasons"
	const reason = `{"code":"PRICE_TOO_HIGH","type":"CHANGE","message":{"en":"The price is too high"}}`
	manageQuotes := []string{"quote.quote_manage"}
	return []writeCase{
		{
			args: []string{"quotes", "create"}, file: `{"cartId":"cart1"}`, method: http.MethodPost, path: quotes,
			scopes: manageQuotes, status: http.StatusCreated, body: `{"id":"Q1001"}`, want: "{\n  \"id\": \"Q1001\"\n}\n",
		},
		{
			args: []string{"quotes", "patch", "Q1001"}, file: `[{"op":"REPLACE","path":"/status","value":{"value":"OPEN"}}]`,
			method: http.MethodPatch, path: quotes + "/Q1001", scopes: manageQuotes, status: http.StatusNoContent,
		},
		{
			args: []string{"quotes", "delete", "Q1001"}, method: http.MethodDelete, path: quotes + "/Q1001",
			scopes: manageQuotes, confirmBy: "Q1001", status: http.StatusNoContent,
		},
		{
			args: []string{"quotes", "reasons", "create"}, file: reason, method: http.MethodPost, path: reasons,
			scopes: manageQuotes, lang: "*", status: http.StatusCreated, body: `{"id":"r1"}`,
			want: "{\n  \"id\": \"r1\"\n}\n",
		},
		{
			args: []string{"quotes", "reasons", "update", "r1"},
			file: strings.TrimSuffix(reason, "}") + `,"metadata":{"version":2}}`, method: http.MethodPut,
			path: reasons + "/r1", scopes: manageQuotes, lang: "*", status: http.StatusNoContent,
		},
		{
			args: []string{"quotes", "reasons", "delete", "r1"}, method: http.MethodDelete, path: reasons + "/r1",
			scopes: manageQuotes, confirmBy: "r1", status: http.StatusNoContent,
		},
	}
}

// TestQuoteExternalPricesRequestTheirScope checks that quotes create
// requests cart.cart_manage_external_prices besides quote.quote_manage with
// --allow-external-prices only, and that patch, which Emporix refuses an
// external price with that scope too, does not take the flag.
func TestQuoteExternalPricesRequestTheirScope(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/quote/acme1/quotes", respond(http.StatusCreated, `{"id":"Q1"}`))
	args := []string{"quotes", "create", "--file", writeBody(t, `{"cartId":"c1"}`), "--allow-external-prices", "--yes"}
	if code, _, stderr := runIn(t, writeEnv(fake, false, ""), args...); code != ExitOK {
		t.Errorf("quotes create: exit %d, stderr %q", code, stderr)
	}
	fake.assertTokens(t, tokenScope([]string{"quote.quote_manage", "cart.cart_manage_external_prices"}))
	assertRefused(t, []refusedWrite{{args: []string{"quotes", "patch", "Q1", "--allow-external-prices"}, file: `[]`,
		want: "unknown flag: --allow-external-prices"}})
}

// TestReturnsHelpNamesTheirSiteAndAnswers checks that the reads of returns
// say what SITE is, how to filter by site and that a site-limited read scope
// fails, and that the writes say how Emporix treats siteCode and restriction
// and what it answers for an unknown return.
func TestReturnsHelpNamesTheirSiteAndAnswers(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: []string{"list"}, want: []string{
			"SITE is the site of the return's orders, which Emporix copies to it (siteCode;",
			"`--query siteCode:DE` those of one site",
			"so a credential that holds only a site-limited one, such as returns.returns_read--DE, fails with exit code 3",
			"and for its writes returns.returns_manage",
		}},
		{args: []string{"create"}, want: []string{
			"Emporix copies siteCode and restriction from the orders, so the body may leave them out",
			"orders of different sites or restrictions",
			"`\"site does not match order\"`", "`\"orders belong to different sites\"`",
		}},
		{args: []string{"update"}, want: []string{
			"with siteCode and restriction left out: its reference documents that an update does not accept them",
		}},
		{args: []string{"patch"}, want: []string{"a patch does not accept /siteCode or /restriction"}},
		{args: []string{"delete"}, want: []string{"its reference now documents 404 Not Found (exit 4)"}},
	} {
		help := helpText(t, append([]string{"returns"}, tt.args...)...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%q: help %q; want %q in it", tt.args, help, want)
			}
		}
	}
}

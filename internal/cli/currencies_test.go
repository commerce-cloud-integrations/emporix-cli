package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestExchangeRateOfAnotherTypeFails(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/currency/acme1/exchanges/EUR_CHF", respond(http.StatusOK, `{"code":"EUR_CHF","rate":{"v":1}}`))
	code, stdout, stderr := runCLI(t, fake.env(), "exchange-rates", "get", "EUR_CHF")
	// The error names the type, not the value sent.
	wantErr := "rate: json: cannot unmarshal object into Go value of type json.Number"
	if code != ExitError || stdout != "" || !strings.Contains(stderr, wantErr) || strings.Contains(stderr, `"v"`) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

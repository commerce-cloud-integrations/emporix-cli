package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestPriceModelsGetAcceptsTheDocumentedArray checks the opt-in tolerance of
// price models: the API reference documents a one-item array, while Emporix
// returns the model itself (as the read case of price-models get pm1 shows).
func TestPriceModelsGetAcceptsTheDocumentedArray(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/priceModels/as-documented", respond(http.StatusOK, `[{"id":"pm1","name":"Default"}]`))
	fake.route("/price/acme1/priceModels/two", respond(http.StatusOK, `[{"id":"pm1"},{"id":"pm2"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "price-models", "get", "as-documented")
	if code != ExitOK || !strings.Contains(stdout, "\npm1  Default ") {
		t.Fatalf("one-item array: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	// An array of several items is not one model.
	code, stdout, stderr = runCLI(t, fake.env(), "price-models", "get", "two")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "expected one item, got 2") {
		t.Fatalf("two items: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

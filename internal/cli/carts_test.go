package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestCartValidationWithoutResultIsMalformed checks that a validation that
// does not say whether the cart is valid fails as a malformed response, for
// -o json as for tables, and prints nothing.
func TestCartValidationWithoutResultIsMalformed(t *testing.T) {
	for _, body := range []string{`{"itemsValidationDetails":[]}`, `{"isValid":"yes"}`, `{}`, `[]`, `null`} {
		for _, format := range [][]string{nil, {"-o", "json"}} {
			fake := newFakeEmporix(t)
			fake.route("/cart/acme1/carts/c1/validate", respond(http.StatusOK, body))
			args := slices.Concat([]string{"carts", "validate", "c1", "--allow-write-scope"}, format)
			code, stdout, stderr := runCLI(t, fake.env(), args...)
			if code != ExitError || stdout != "" ||
				!strings.Contains(stderr, "decoding cart validation response: expected an object with isValid") {
				t.Errorf("%s %q: exit %d, stdout %q, stderr %q", body, format, code, stdout, stderr)
			}
		}
	}
}

// TestFindWithoutAMatchingCartIsNotFound checks that criteria that no cart
// matches give not found (exit 4), as Emporix answers 404.
func TestFindWithoutAMatchingCartIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t) // serves 404 on every path but the token's
	code, stdout, stderr := runCLI(t, fake.env(), "carts", "find", "--site", "AcmeStore_NL", "--customer", "C0123",
		"--allow-write-scope")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "GET /cart/acme1/carts: 404 Not Found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

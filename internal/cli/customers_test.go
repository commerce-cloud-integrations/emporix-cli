package cli

import (
	"strings"
	"testing"
)

// TestCustomersAreIdentifiedByNumber checks that the customer number, not the
// id, identifies a customer: a profile without one cannot be shown.
func TestCustomersAreIdentifiedByNumber(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/customer/acme1/customers", firstPage(`[{"id":"c-7f3a"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "customers", "list")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "item 1 has no identifier") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

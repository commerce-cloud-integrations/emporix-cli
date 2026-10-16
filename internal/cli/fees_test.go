package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestFeesGetRejectsAnArray checks that get finds a fee wrapped or bare (as
// in the fee read cases), but not in an array: neither the reference nor
// Emporix answers with one.
func TestFeesGetRejectsAnArray(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/fee/acme1/fees/f1", respond(http.StatusOK, `[{"id":"f1","feeType":"ABSOLUTE"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "fees", "get", "f1", "-o", "json")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "expected one item as a JSON object") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestFeesGetReportsAMalformedWrappedFee(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/fee/acme1/fees/f1", respond(http.StatusOK, `{"fee":{"id":"f1","feeType":5}}`))
	code, stdout, stderr := runCLI(t, fake.env(), "fees", "get", "f1")
	// The wrapped fee's own decoding error is reported, not a missing identifier.
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "cannot unmarshal number") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestFeesOfAnItemFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	// Emporix lists nothing for an unknown item, fails on a malformed YRN, and
	// would answer fee ids if expand were not honored.
	fake.route("/fee/acme1/productFees/unknown/fees", respond(http.StatusOK, `[]`))
	fake.route("/fee/acme1/itemFees/malformed/fees", respond(http.StatusInternalServerError,
		`{"message":"Something went wrong"}`))
	fake.route("/fee/acme1/productFees/ids-only/fees", respond(http.StatusOK, `["f1"]`))
	for _, tt := range []struct {
		args         []string
		code         int
		stdout, want string
	}{
		{args: []string{"for-product", "unknown"}, code: ExitOK, stdout: "ID  CODE  NAME  TYPE  AMOUNT  ITEM TYPE  " +
			"SITE  ACTIVE  TAXABLE  FROM  TO\n"},
		{args: []string{"for-item", "malformed"}, code: ExitError, want: "500 Internal Server Error"},
		{args: []string{"for-product", "ids-only"}, code: ExitError, want: "decoding fee response"},
	} {
		code, stdout, stderr := runCLI(t, fake.env(), slices.Concat([]string{"fees"}, tt.args, []string{"--site", "main"})...)
		if code != tt.code || stdout != tt.stdout || !strings.Contains(stderr, tt.want) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", tt.args, code, stdout, stderr)
		}
	}
}

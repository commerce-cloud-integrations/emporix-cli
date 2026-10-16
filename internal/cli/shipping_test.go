package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestShippingErrorsOfAnUnknownGroupOrCustomer pins how Emporix answers a
// group or customer it does not know: 409 Conflict for a group (exit 1) and an
// empty body for a customer without a shipping group (not found, exit 4).
func TestShippingErrorsOfAnUnknownGroupOrCustomer(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/shipping/acme1/main/groups/unknown", respond(http.StatusConflict,
		`{"message":"Group with Id 'unknown' could not be found."}`))
	fake.route("/shipping/acme1/main/cgrelations/c2", respond(http.StatusOK, ``))
	for _, tt := range []struct {
		args []string
		code int
		want string
	}{
		{args: []string{"groups", "get", "unknown"}, code: ExitError, want: "409 Conflict"},
		{args: []string{"customers", "get", "c2"}, code: ExitNotFound, want: "empty response: not found"},
	} {
		args := slices.Concat([]string{"shipping"}, tt.args, []string{"--site", "main"})
		if code, stdout, stderr := runCLI(t, fake.env(), args...); code != tt.code || stdout != "" ||
			!strings.Contains(stderr, tt.want) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", tt.args, code, stdout, stderr)
		}
	}
}

// TestDeliveryWindowsAreIdentifiedByDate checks that the delivery date, the
// one member the reference requires of a delivery window, identifies it: a
// window without one cannot be shown.
func TestDeliveryWindowsAreIdentifiedByDate(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/shipping/acme1/actualDeliveryWindows/cart1", respond(http.StatusOK,
		`[{"deliveryDate":"2026-06-20T12:00:00.000Z"},{"id":"w2","slotId":"s1"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "delivery-times", "windows", "cart1")
	if code != ExitError || stdout != "" ||
		!strings.Contains(stderr, "decoding delivery window response: item 2 has no identifier") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

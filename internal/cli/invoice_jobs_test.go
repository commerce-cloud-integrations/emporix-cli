package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestInvoiceJobErrors pins how Emporix answers a job it does not know, 500
// Internal Server Error (exit 1), and that a job without a status is malformed:
// the reference schema's jobStatus is not read, since Emporix sends status.
func TestInvoiceJobErrors(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/invoice/acme1/jobs/invoices/unknown", respond(http.StatusInternalServerError,
		`{"message":"Internal error"}`))
	fake.route("/invoice/acme1/jobs/invoices/j1", respond(http.StatusOK,
		`{"jobStatus":"DONE","jobType":"MANUAL","invoices":[]}`))
	for id, want := range map[string]string{
		"unknown": "500 Internal Server Error",
		"j1":      "decoding invoice job response: item 1 has no identifier",
	} {
		if code, stdout, stderr := runCLI(t, fake.env(), "invoice-jobs", "get", id); code != ExitError || stdout != "" ||
			!strings.Contains(stderr, want) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", id, code, stdout, stderr)
		}
	}
}

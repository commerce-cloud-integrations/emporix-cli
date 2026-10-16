package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestOrdersHistoryJSONKeepsTheMetadata(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/order-v2/acme1/salesorders/EON1/historical-transitions", respond(http.StatusOK,
		`{"transitions":[{"status":"CREATED","timestamp":"2026-10-01T10:00:00.000Z"}],
		"metadata":{"version":2,"modifiedAt":"2026-10-01T11:00:00.000Z"}}`))
	code, stdout, stderr := runCLI(t, fake.env(), "orders", "history", "EON1", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"modifiedAt"`) || !strings.Contains(stdout, `"CREATED"`) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestOrdersHistoryRejectsAResponseWithoutTransitions(t *testing.T) {
	for _, tt := range []struct {
		body, stdout string
		code         int
	}{
		{body: `{"transitions":[],"metadata":{"version":1}}`, code: ExitOK, stdout: "STATUS  TIMESTAMP\n"},
		{body: `{"metadata":{"version":1}}`, code: ExitError},
		{body: `{"transitions":null}`, code: ExitError},
		{body: `null`, code: ExitError},
	} {
		fake := newFakeEmporix(t)
		fake.route("/order-v2/acme1/salesorders/EON1/historical-transitions", respond(http.StatusOK, tt.body))
		code, stdout, stderr := runCLI(t, fake.env(), "orders", "history", "EON1")
		if code != tt.code || stdout != tt.stdout ||
			tt.code == ExitError && !strings.Contains(stderr, "expected an object with transitions") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", tt.body, code, stdout, stderr)
		}
	}
}

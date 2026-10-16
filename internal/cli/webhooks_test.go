package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestWebhookStatisticsWithoutALimit checks that statistics without a limit,
// or without anything, still show one row rather than failing, while a body
// that is no object fails.
func TestWebhookStatisticsWithoutALimit(t *testing.T) {
	for body, want := range map[string]string{
		`{"monthStatistics":[{"statisticsMonth":"2026-10","succesfullySentWebhooks":3}]}`: "not reported  2026-10: 3",
		`{"webhooksLimit":null}`: "not reported",
		`{}`:                     "not reported",
	} {
		fake := newFakeEmporix(t)
		fake.route("/webhook/acme1/statistics", respond(http.StatusOK, body))
		code, stdout, stderr := runCLI(t, fake.env(), "webhooks", "statistics")
		if code != ExitOK || !strings.Contains(stdout, want) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", body, code, stdout, stderr)
		}
	}
	for _, body := range []string{`[]`, `null`, `"none"`} {
		fake := newFakeEmporix(t)
		fake.route("/webhook/acme1/statistics", respond(http.StatusOK, body))
		code, stdout, stderr := runCLI(t, fake.env(), "webhooks", "statistics", "-o", "json")
		if code != ExitError || stdout != "" || !strings.Contains(stderr, "expected a JSON object") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", body, code, stdout, stderr)
		}
	}
}

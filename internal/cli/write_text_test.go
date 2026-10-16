package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestTextAnswerIsPrintedAsJSON checks that orders recalculate, which Emporix
// answers with bare text labelled JSON (live), prints the text, trimmed, as a
// JSON string, JSON as sent, and nothing for a blank answer, and that its
// help says so.
func TestTextAnswerIsPrintedAsJSON(t *testing.T) {
	recalculate := []string{"orders", "recalculate", "O1001", "--file", writeBody(t, `{"entries":[]}`), "--yes"}
	for body, want := range map[string]string{
		"Order with id: O1001 calculated successfully\n": "\"Order with id: O1001 calculated successfully\"\n",
		`{"id":"O1001"}`: "{\n  \"id\": \"O1001\"\n}\n",
		" \n":            "",
	} {
		fake := newFakeEmporix(t)
		fake.route("/order-v2/acme1/salesorders/O1001/calculations", respond(http.StatusOK, body))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), recalculate...)
		if code != ExitOK || stdout != want || stderr != "" {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want %q", body, code, stdout, stderr, want)
		}
	}
	_, help, _ := runCLI(t, nil, "orders", "recalculate", "--help")
	if !strings.Contains(help, "\n\nEmporix answers with a sentence of text, which epx prints as a JSON string.\n\n") {
		t.Errorf("orders recalculate: help %q; want the text answer explained", help)
	}
}

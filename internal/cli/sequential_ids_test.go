package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestSequentialIDOfAnUnknownIDIsNotFound checks the live behavior the help
// documents: Emporix answers an unknown ID with an empty body, which is not
// found (exit 4).
func TestSequentialIDOfAnUnknownIDIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/sequential-id/acme1/schemas/000000000000000000000000", respond(http.StatusOK, ""))
	code, stdout, stderr := runCLI(t, fake.env(), "sequential-ids", "get", "000000000000000000000000")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestSequentialIDOfAnUnknownIDIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/sequential-id/acme1/schemas/000000000000000000000000", respond(http.StatusOK, ""))
	code, stdout, stderr := runCLI(t, fake.env(), "sequential-ids", "get", "000000000000000000000000")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

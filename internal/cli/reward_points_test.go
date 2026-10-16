package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestRewardPointsBalanceIsOneNumber checks that a balance must be a JSON
// number, also for -o json, and that an empty response is not found.
func TestRewardPointsBalanceIsOneNumber(t *testing.T) {
	for _, body := range []string{`{"points":5}`, `"5"`, `null`, `[5]`, `true`, `5 6`} {
		fake := newFakeEmporix(t)
		fake.route("/reward-points/customer/c1", respond(http.StatusOK, body))
		code, stdout, stderr := runCLI(t, fake.env(), "reward-points", "balance", "c1", "-o", "json")
		if code != ExitError || stdout != "" || !strings.Contains(stderr, "expected a JSON number") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", body, code, stdout, stderr)
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/reward-points/customer/c1", respond(http.StatusOK, `-1.5e3`))
	if code, stdout, stderr := runCLI(t, fake.env(), "reward-points", "balance", "c1", "-o", "json"); code != ExitOK ||
		stdout != "-1.5e3\n" {
		t.Errorf("negative: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.route("/reward-points/customer/c1", respond(http.StatusOK, ``))
	if code, stdout, stderr := runCLI(t, fake.env(), "reward-points", "balance", "c1"); code != ExitNotFound ||
		stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Errorf("empty: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

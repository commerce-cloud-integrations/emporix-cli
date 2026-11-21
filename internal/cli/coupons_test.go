package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestReferralCouponOfACustomerWithoutOneIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/coupon/acme1/referral-coupons/10001", respond(http.StatusNoContent, ``))
	code, stdout, stderr := runCLI(t, fake.env(), "coupons", "referral", "10001")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

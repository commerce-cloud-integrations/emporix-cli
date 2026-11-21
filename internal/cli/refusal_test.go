package cli

import (
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const businessRefusal = `{"type":"business_error","message":"Coupon cannot be redeemed","errorCode":"E-403",` +
	`"details":[{"type":"coupon_redemption_forbidden","message":"This customer is not allowed"}]}`

const refusedMessage = "403 Forbidden: Coupon cannot be redeemed (E-403) (This customer is not allowed)\n"

func refusals() map[string]string {
	return map[string]string{
		businessRefusal: refusedMessage,
		`{"code":403,"message":"User not authorized."}`: "403 Forbidden: User not authorized.\n",
	}
}

func scopeRefusals() []string {
	return []string{`{"message":"Missing one of required scopes '[order.order_update]'"}`, ""}
}

func TestWriteForbiddenAnswers(t *testing.T) {
	for _, tt := range writeCases() {
		fake := newFakeEmporix(t)
		checkRefusals(t, &tt, fake)
		if tt.refuses {
			checkScopeRefusals(t, &tt, fake)
		}
	}
}

func checkRefusals(t *testing.T, tt *writeCase, fake *fakeEmporix) {
	t.Helper()
	for body, ends := range refusals() {
		fake.route(tt.path, tt.serve(http.StatusForbidden, body))
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag)...)
		if tt.refuses && (code != ExitError || !strings.HasSuffix(stderr, ends)) || !tt.refuses && code != ExitAuth {
			t.Errorf("%q refused with %q (declared %v): exit %d, stderr %q", tt.args, body, tt.refuses, code, stderr)
		}
	}
}

func checkScopeRefusals(t *testing.T, tt *writeCase, fake *fakeEmporix) {
	t.Helper()
	for _, body := range scopeRefusals() {
		fake.route(tt.path, tt.serve(http.StatusForbidden, body))
		if code, _, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag)...); code != ExitAuth {
			t.Errorf("%q answered 403 %q: exit %d, stderr %q; want exit 3", tt.args, body, code, stderr)
		}
	}
}

func TestCouponValidationRefusals(t *testing.T) {
	args := []string{"coupons", "validate", "SAVE7", "--order-total", "10 EUR", "--discount", "1 EUR",
		"--customer", "10001", "--allow-write-scope"}
	for body, want := range map[string]int{businessRefusal: ExitError, scopeRefusals()[0]: ExitAuth,
		scopeRefusals()[1]: ExitAuth} {
		fake := newFakeEmporix(t)
		fake.route("/coupon/acme1/coupons/SAVE7/validation", respond(http.StatusForbidden, body))
		code, stdout, stderr := runCLI(t, fake.env(), args...)
		if code != want || stdout != "" || want == ExitError && !strings.HasSuffix(stderr, refusedMessage) {
			t.Errorf("validation answered 403 %q: exit %d, stdout %q, stderr %q; want exit %d", body, code, stdout,
				stderr, want)
		}
	}
}

func TestRefusalsUnnamedInHelpPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("no panic for a refusal that the help does not name")
		}
	}()
	checkRefusalHelp(&cobra.Command{Use: "transition", Long: "Moves the order."})
}

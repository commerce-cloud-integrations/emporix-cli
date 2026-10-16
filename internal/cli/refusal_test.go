package cli

import (
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// businessRefusal is a 403 Forbidden with which Emporix refuses an input
// (live: a coupon that the customer may not redeem).
const businessRefusal = `{"type":"business_error","message":"Coupon cannot be redeemed","errorCode":"E-403",` +
	`"details":[{"type":"coupon_redemption_forbidden","message":"This customer is not allowed"}]}`

// refusedMessage is how the error of businessRefusal ends.
const refusedMessage = "403 Forbidden: Coupon cannot be redeemed (E-403) (This customer is not allowed)\n"

// refusals are the 403 answers with which Emporix refuses an input (live),
// each with how its error ends: businessRefusal, and the Cart Service's for
// a coupon restricted to customers, whose message names no scope (its
// reference gives a missing scope the same body, but epx checks the scope of
// the write before it sends).
func refusals() map[string]string {
	return map[string]string{
		businessRefusal: refusedMessage,
		`{"code":403,"message":"User not authorized."}`: "403 Forbidden: User not authorized.\n",
	}
}

// scopeRefusals are 403 answers about the credential: one that names a
// missing scope and one without a message.
func scopeRefusals() []string {
	return []string{`{"message":"Missing one of required scopes '[order.order_update]'"}`, ""}
}

// TestWriteForbiddenAnswers checks that each refusal answered 403 Forbidden
// to a write that declares one (writeCase.refuses) exits 1 with Emporix's
// message and error code, while a 403 that names a missing scope, or has no
// message, still exits 3, and that every other write exits 3 for any 403.
func TestWriteForbiddenAnswers(t *testing.T) {
	for _, tt := range writeCases() {
		fake := newFakeEmporix(t)
		checkRefusals(t, &tt, fake)
		if tt.refuses {
			checkScopeRefusals(t, &tt, fake)
		}
	}
}

// checkRefusals checks that a write answered 403 with each refusal body exits
// as the write declares it: 1 with the help's sentence where it is a refusal
// of the input, else 3.
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

// checkScopeRefusals checks that a refusal of input that names a missing scope
// exits 3, where the write declares a refusal of its input.
func checkScopeRefusals(t *testing.T, tt *writeCase, fake *fakeEmporix) {
	t.Helper()
	for _, body := range scopeRefusals() {
		fake.route(tt.path, tt.serve(http.StatusForbidden, body))
		if code, _, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag)...); code != ExitAuth {
			t.Errorf("%q answered 403 %q: exit %d, stderr %q; want exit 3", tt.args, body, code, stderr)
		}
	}
}

// TestCouponValidationRefusals checks the same of coupons validate, a read
// that declares a refusal (bodyRead.forbiddenIsRefusal).
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

// TestRefusalsUnnamedInHelpPanic checks that a command declared to take a
// 403 as a refusal, whose help does not say so, is a mistake that panics
// while the tree is built.
func TestRefusalsUnnamedInHelpPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("no panic for a refusal that the help does not name")
		}
	}()
	checkRefusalHelp(&cobra.Command{Use: "transition", Long: "Moves the order."})
}

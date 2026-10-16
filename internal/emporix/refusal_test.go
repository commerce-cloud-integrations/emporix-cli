package emporix

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// forbiddenAnswer is a 403 Forbidden answer of a service, and the error that
// it makes.
type forbiddenAnswer struct {
	name    string
	status  int
	body    string
	refusal bool   // a refusal of the input where the request declares one: no class
	want    string // the error's message, as sent
}

// forbiddenAnswers are the 403 answers seen live, and others that are no
// refusals: a 403 that names a scope, one without a message and a 401.
func forbiddenAnswers() []forbiddenAnswer {
	long := strings.Repeat("x", 320)
	return []forbiddenAnswer{
		{name: "transition", status: http.StatusForbidden, refusal: true,
			body: `{"message":"Order id EON1, with status CREATED, could not be SHIPPED"}`,
			want: "Order id EON1, with status CREATED, could not be SHIPPED"},
		{name: "business error", status: http.StatusForbidden, refusal: true,
			body: `{"type":"business_error","message":"Coupon cannot be redeemed","details":[` +
				`{"type":"coupon_redemption_forbidden","message":"This customer is not allowed to redeem it"}]}`,
			want: "Coupon cannot be redeemed (This customer is not allowed to redeem it)"},
		{name: "plain text", status: http.StatusForbidden, refusal: true,
			body: "direct upload is not enabled for this tenant", want: "direct upload is not enabled for this tenant"},
		{name: "missing scope", status: http.StatusForbidden,
			body: `{"message":"Missing one of required scopes '[order.order_update]'"}`,
			want: "Missing one of required scopes '[order.order_update]'"},
		{name: "scope required", status: http.StatusForbidden,
			body: `{"message":"A scope coupon.coupon_redeem_on_behalf is required in order to set customerNumber"}`,
			want: "A scope coupon.coupon_redeem_on_behalf is required in order to set customerNumber"},
		{name: "scope in capitals", status: http.StatusForbidden,
			body: `{"message":"Required Scope(s): [label.label_delete]"}`, want: "Required Scope(s): [label.label_delete]"},
		{name: "scope past the message shown", status: http.StatusForbidden,
			body: `{"message":"` + long + ` Missing required scope order.order_update"}`, want: long[:300] + "…"},
		{name: "scope in another member", status: http.StatusForbidden,
			body: `{"message":"Not allowed","requiredScopes":["order.order_update"]}`, want: "Not allowed"},
		{name: "deleted coupon", status: http.StatusForbidden, refusal: true,
			body: `{"message":"User not authorized. (The operation is not permitted on coupon with ID SAVE7)"}`,
			want: "User not authorized. (The operation is not permitted on coupon with ID SAVE7)"},
		{name: "insufficient permissions", status: http.StatusForbidden,
			body: `{"message":"Not allowed to access this resource","type":"insufficient_permissions"}`,
			want: "Not allowed to access this resource"},
		{name: "no message", status: http.StatusForbidden, body: `{}`},
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"message":"Order cannot be shipped"}`,
			want: "Order cannot be shipped"},
	}
}

// check fails t unless err is the error of a, of no class where refused and
// ErrAuth otherwise.
func (a forbiddenAnswer) check(t *testing.T, err error, refused bool, route string) {
	t.Helper()
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != a.status || apiErr.Message != a.want {
		t.Fatalf("%s: error %v; want %d with %q", a.name, err, a.status, a.want)
	}
	want := ErrAuth
	if refused {
		want = nil
	}
	if class := classOf(err); !errors.Is(class, want) {
		t.Errorf("%s (%s): error %v of class %v; want %v", a.name, route, err, class, want)
	}
}

// TestWritesDeclaredSoTakeForbiddenAsRefusal checks that a 403 Forbidden to
// a write that declares that its service refuses an input so
// (Write.ForbiddenIsRefusal) is of no class where its message names no scope,
// and ErrAuth otherwise, as every 403 to a write that declares none is.
func TestWritesDeclaredSoTakeForbiddenAsRefusal(t *testing.T) {
	for _, a := range forbiddenAnswers() {
		for _, declared := range []bool{true, false} {
			fake := newFakeEmporix(t)
			fake.route("/order-v2/acme1/orders/EON1/transitions", respond(a.status, a.body))
			_, err := send(fake.client(t), &Write{Method: http.MethodPost,
				Path: "/order-v2/{tenant}/orders/EON1/transitions", Body: jsonBody(t, `{"status":"SHIPPED"}`),
				Answers: Answers{ForbiddenIsRefusal: declared}})
			a.check(t, err, declared && a.refusal, "write")
		}
	}
}

// TestReadsDeclaredSoTakeForbiddenAsRefusal checks the same of reads: those
// of a client made by ForbiddenIsRefusal, a check with a body included, but
// not those of the client that it was made of.
func TestReadsDeclaredSoTakeForbiddenAsRefusal(t *testing.T) {
	const path = "/coupon/{tenant}/coupons/SAVE7/validation"
	for _, a := range forbiddenAnswers() {
		fake := newFakeEmporix(t)
		fake.route("/coupon/acme1/coupons/SAVE7/validation", respond(a.status, a.body))
		checking := withBody(t, fake.client(t), http.MethodPost, `{"customerNumber":"10001"}`)
		_, err := checking.ForbiddenIsRefusal().Check(t.Context(), path, nil)
		a.check(t, err, a.refusal, "declared check")
		_, err = checking.Check(t.Context(), path, nil)
		a.check(t, err, false, "check")
		_, err = fake.client(t).ForbiddenIsRefusal().Get(t.Context(), path, nil)
		a.check(t, err, a.refusal, "declared GET")
	}
}

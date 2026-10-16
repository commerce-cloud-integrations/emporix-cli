package cli

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// couponJSON is a coupon as create takes it.
const couponJSON = `{"code":"SAVE7","name":"Save 7","discountType":"PERCENT","discountPercentage":7}`

// couponWriteCases are the writes of coupons, their redemptions and referral
// coupons.
func couponWriteCases() []writeCase {
	manage, onBehalf := []string{couponManage}, []string{couponRedeemOnBehalf}
	return []writeCase{
		{
			args: []string{"coupons", "create"}, file: couponJSON, method: http.MethodPost, path: "/coupon/acme1/coupons",
			scopes: manage, status: http.StatusCreated, body: `{"id":"SAVE7"}`, want: "{\n  \"id\": \"SAVE7\"\n}\n",
		},
		{
			args: []string{"coupons", "update", "SAVE 7"}, file: `{"name":"Save 7","metadata":{"version":2}}`,
			method: http.MethodPut, path: "/coupon/acme1/coupons/SAVE%207", query: "partial=false", scopes: manage,
			status: http.StatusOK, refuses: true,
		},
		{
			args: []string{"coupons", "patch", "SAVE7"}, file: `{"status":"INACTIVE"}`, method: http.MethodPatch,
			path: "/coupon/acme1/coupons/SAVE7", scopes: manage, status: http.StatusOK, refuses: true,
		},
		{
			args: []string{"coupons", "delete", "SAVE7"}, method: http.MethodDelete, path: "/coupon/acme1/coupons/SAVE7",
			scopes: manage, confirmBy: "SAVE7", status: http.StatusNoContent, refuses: true,
		},
		{
			args: []string{"coupons", "redeem", "SAVE 7", "--order-total", "1090 EUR", "--discount", "5.00 EUR",
				"--customer", "C1", "--order", "o1"},
			sent: `{"customerNumber":"C1","discount":{"amount":5.00,"currency":"EUR"},` +
				`"orderCode":"o1","orderTotal":{"amount":1090,"currency":"EUR"}}`,
			method: http.MethodPost, path: "/coupon/acme1/coupons/SAVE%207/redemptions", scopes: onBehalf,
			status: http.StatusCreated, body: `{"id":"r1"}`, want: "{\n  \"id\": \"r1\"\n}\n", refuses: true,
		},
		{
			args: []string{"coupons", "redeem", "SAVE7"},
			file: `{"legalEntityId":"le1","orderTotal":{"amount":10,"currency":"EUR"},` +
				`"discount":{"amount":1,"currency":"EUR"}}`,
			method: http.MethodPost, path: "/coupon/acme1/coupons/SAVE7/redemptions", scopes: onBehalf,
			status: http.StatusCreated, body: `{"id":"r2"}`, want: "{\n  \"id\": \"r2\"\n}\n", refuses: true,
		},
		{
			args: []string{"coupons", "redemption-delete", "SAVE7", "r 1"}, method: http.MethodDelete,
			path: "/coupon/acme1/coupons/SAVE7/redemptions/r%201", scopes: manage, confirmBy: "r 1",
			status: http.StatusNoContent,
		},
		{
			args: []string{"coupons", "referral-create", "100 01"}, method: http.MethodPost,
			path: "/coupon/acme1/referral-coupons/100%2001", status: http.StatusOK,
			body: `{"code":"REF_30","issuedTo":"100 01","referralCoupon":true}`,
			want: "{\n  \"code\": \"REF_30\",\n  \"issuedTo\": \"100 01\",\n  \"referralCoupon\": true\n}\n",
		},
	}
}

// rewardPointWriteCases are the writes of reward points and redeem options.
func rewardPointWriteCases() []writeCase {
	manage := []string{rewardPointsManage}
	const option = `{"type":"coupon","name":"10 EUR off","points":90,"coupon":{"name":"10 EUR off",` +
		`"discountType":"ABSOLUTE","discountAbsolute":{"amount":10,"currency":"EUR"}}}`
	return []writeCase{
		{
			args:   []string{"reward-points", "create", "C1001", "--site", "AcmeStore_NL", "--points", "20"},
			sent:   `{"customerId":"C1001","startPoints":{"points":20}}`,
			method: http.MethodPost, path: "/reward-points/customer/C1001",
			query: "siteCode=AcmeStore_NL", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"reward-points", "create", "c1", "--site", "main", "--points", "020", "--description",
				"Welcome"},
			sent: `{"customerId":"c1","startPoints":{"description":"Welcome","points":20}}`, method: http.MethodPost,
			path: "/reward-points/customer/c1", query: "siteCode=main", scopes: manage, status: http.StatusOK,
		},
		{
			// Without --site, Emporix applies the settings of site main.
			args: []string{"reward-points", "create", "c 1"}, sent: `{"customerId":"c 1"}`, method: http.MethodPost,
			path: "/reward-points/customer/c%201", scopes: manage, status: http.StatusOK,
		},
		{
			// --file is sent as given, also without customerId (Emporix answers 500 then).
			args: []string{"reward-points", "create", "c1"}, file: `{"startPoints":{"points":20}}`,
			method: http.MethodPost, path: "/reward-points/customer/c1", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"reward-points", "create", "c1", "--site", "main"},
			file: `{"customerId":"c1","startPoints":{"points":20,"description":"Welcome"}}`, method: http.MethodPost,
			path: "/reward-points/customer/c1", query: "siteCode=main", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"reward-points", "delete", "c1"}, method: http.MethodDelete, path: "/reward-points/customer/c1",
			scopes: manage, confirmBy: "c1", status: http.StatusOK,
		},
		{
			args: []string{"reward-points", "add", "c1", "--site", "AcmeStore_NL", "--points", "50", "--description",
				"Goodwill", "--order", "o1", "--valid-until", "2027-12-31T23:59:59Z"},
			sent:   `{"description":"Goodwill","points":50,"reference-id":"o1","validUntil":"2027-12-31T23:59:59Z"}`,
			method: http.MethodPost, path: "/reward-points/customer/c1/addPoints", query: "siteCode=AcmeStore_NL",
			scopes: manage, status: http.StatusCreated,
		},
		{
			args: []string{"reward-points", "add", "c1"}, file: `{"id":"a1","points":5}`,
			method: http.MethodPost, path: "/reward-points/customer/c1/addPoints", scopes: manage,
			status: http.StatusCreated,
		},
		{
			args: []string{"reward-points", "redeem", "c1", "--points", "50"}, sent: `{"points":50}`,
			method: http.MethodPost, path: "/reward-points/customer/c1/redeemPoints", scopes: manage,
			status: http.StatusCreated, conflict: redeemConflict,
		},
		{
			args: []string{"reward-points", "redeem", "c1"}, file: `{"description":"Gift wrapping","points":50}`,
			method: http.MethodPost, path: "/reward-points/customer/c1/redeemPoints", scopes: manage,
			status: http.StatusCreated, conflict: redeemConflict,
		},
		{
			args: []string{"reward-points", "redeem-option-create"}, file: option, method: http.MethodPost,
			path: "/reward-points/acme1/redeemOptions", scopes: manage, status: http.StatusOK,
			body: `[{"id":"o1","type":"coupon","points":90}]`,
			want: "[\n  {\n    \"id\": \"o1\",\n    \"points\": 90,\n    \"type\": \"coupon\"\n  }\n]\n",
		},
		{
			args: []string{"reward-points", "redeem-option-update", "o1"}, file: option, method: http.MethodPut,
			path: "/reward-points/acme1/redeemOptions/o1", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"reward-points", "redeem-option-delete", "o1"}, method: http.MethodDelete,
			path: "/reward-points/acme1/redeemOptions/o1", scopes: manage, confirmBy: "o1",
			status: http.StatusNoContent,
		},
	}
}

// shoppingListWriteCases are the writes of customers' shopping lists.
func shoppingListWriteCases() []writeCase {
	manage := []string{"shoppinglist.shoppinglist_manage"}
	const list = `{"customerId":"c1","name":"default","items":[{"id":1,"productId":"100200","quantity":2}]}`
	return []writeCase{
		{
			args: []string{"shopping-lists", "create"}, file: list, method: http.MethodPost,
			path: "/shoppinglist/acme1/shopping-lists", scopes: manage, status: http.StatusCreated,
			body: `{"id":"c1"}`, want: "{\n  \"id\": \"c1\"\n}\n",
		},
		{
			args: []string{"shopping-lists", "update", "c:1"}, file: `{"name":"default","items":[]}`,
			method: http.MethodPut, path: "/shoppinglist/acme1/shopping-lists/c:1", scopes: manage,
			status: http.StatusNoContent,
		},
		{
			args: []string{"shopping-lists", "delete", "c1", "--name", "week end"}, method: http.MethodDelete,
			path: "/shoppinglist/acme1/shopping-lists/c1", query: "name=week+end",
			scopes: manage, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"shopping-lists", "delete-all", "c1"}, method: http.MethodDelete,
			path: "/shoppinglist/acme1/shopping-lists/c1", scopes: manage, confirmBy: "c1",
			many: "every shopping list of the customer", status: http.StatusNoContent,
		},
	}
}

// TestPromotionWriteUsageErrorsSendNothing checks that the typed writes of
// coupons, reward points and shopping lists fail as a usage error (exit 2)
// that sends nothing, also for a dry run, without a flag that their body or
// request requires, with a value that a flag refuses (a count of points that
// is no whole number above 0), with a flag that applies only with another, or
// with --file and a flag that it replaces.
func TestPromotionWriteUsageErrorsSendNothing(t *testing.T) {
	file := writeBody(t, `{"points":5}`)
	cases := []refusedWrite{
		{args: []string{"coupons", "redeem", "SAVE7", "--discount", "1 EUR"},
			want: "at least one of the flags in the group [file order-total] is required"},
		{args: []string{"coupons", "redeem", "SAVE7", "--order-total", "10", "--discount", "1 EUR"},
			want: `invalid argument "10" for "--order-total" flag: use AMOUNT CURRENCY`},
		{args: []string{"coupons", "redeem", "SAVE7", "--file", file, "--customer", "C1"},
			want: "if any flags in the group [file customer] are set none of the others can be"},
		{args: []string{"coupons", "redemption-delete", "SAVE7"}, want: "accepts 2 arg(s), received 1"},
		{args: []string{"coupons", "referral-create", ".."}, want: `invalid CUSTOMER_NUMBER ".."`},
		{args: []string{"coupons", "delete", "SAVE7", "--file", file}, want: "unknown flag: --file"},
		{args: []string{"reward-points", "add", "c1", "--points", "5", "--site", ".."},
			want: `invalid argument ".." for "--site" flag: invalid site ".."`},
		{args: []string{"reward-points", "add", "c1", "--site", "main"},
			want: "at least one of the flags in the group [file points] is required"},
		{args: []string{"shopping-lists", "update", "c/1", "--file", file},
			want: "a path segment of a write is empty, . or .., or holds /, \\ or %"},
		{args: []string{"reward-points", "create", "c1", "--site", "main", "--file", file, "--points", "1"},
			want: "if any flags in the group [file points] are set none of the others can be"},
		{args: []string{"reward-points", "create", "c1", "--site", "main", "--description", "Welcome"},
			want: "--description applies only with --points"},
		{args: []string{"reward-points", "redeem-option-update", "o1"}, want: `required flag(s) "file" not set`},
		{args: []string{"reward-points", "redeem-option-delete"}, want: "accepts 1 arg(s), received 0"},
		{args: []string{"shopping-lists", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"shopping-lists", "delete", "c1", "--name", ""},
			want: `invalid argument "" for "--name" flag: empty`},
		{args: []string{"shopping-lists", "delete", "c1"}, want: `required flag(s) "name" not set`},
		{args: []string{"shopping-lists", "delete-all", "c1", "--name", "x"}, want: "unknown flag: --name"},
	}
	for _, points := range []string{"many", "1.5", "1e3", "-5", "0", "007x", ""} {
		for _, write := range [][]string{
			{"reward-points", "create", "c1", "--site", "main"},
			{"reward-points", "add", "c1", "--site", "main"},
			{"reward-points", "redeem", "c1"},
		} {
			cases = append(cases, refusedWrite{args: append(slices.Clone(write), "--points="+points),
				want: fmt.Sprintf(`invalid argument %q for "--points" flag: use a whole number above 0`, points)})
		}
	}
	assertRefused(t, cases)
}

// TestPromotionWritesHelp pins what the help of coupon, reward points and
// shopping list writes, and of coupons validate, says of Emporix's answers:
// the exit code of each refusal, the writes that create nothing, the deletes
// of unknown items and the settings the site needs.
func TestPromotionWritesHelp(t *testing.T) {
	assertHelpHolds(t, map[string][]string{
		"coupons redeem": {redemptionRefusals, legalEntityNote, `answers with: {"id": ID}`,
			"Sent twice, it redeems the coupon twice", "Emporix refuses it to a token with coupon.coupon_manage"},
		"coupons validate":          {"may not redeem with " + forbiddenRefusal, legalEntityNote},
		"coupons redemption-delete": {"which then counts one redemption less"},
		"coupons referral-create": {"couponSettings mixin; where the main site has\nnone, it answers 400 Bad Request",
			"A deleted referral coupon stays the customer's", readScopedNote, noScopeSentence},
		"reward-points create": {"couponSettings mixin", "unknown site code, and the points then stay valid 36 months",
			"the default site of the profile in\neffect without it, and site main without either", "so epx prints nothing",
			"must give customerId, CUSTOMER again", "A customer\nwith an entry is a conflict (exit 5)"},
		"reward-points add": {"couponSettings mixin", "reads a time that ends in Z", "so epx prints\nnothing",
			"Sent twice, it adds them twice", "without a reward points entry is not found"},
		"reward-points redeem": {"without a reward points entry is not found", "so epx prints nothing",
			"More points than are active", "Sent twice, it redeems them twice"},
		"reward-points delete":  {"200 OK, also for an unknown customer or one without an entry"},
		"shopping-lists create": {"is a conflict (exit 5)", "unknown customer id"},
		"shopping-lists update": {"or without a list of that name, is not found (exit 4)", "Update creates no list"},
		"shopping-lists delete": {"its other lists stay", "or without a list of that name, is not found"},
	})
}

// assertHelpHolds checks that the help of each command, named by its words,
// holds each part.
func assertHelpHolds(t *testing.T, parts map[string][]string) {
	t.Helper()
	for args, want := range parts {
		_, help, _ := runCLI(t, nil, append(strings.Fields(args), "--help")...)
		for _, part := range want {
			if !strings.Contains(help, part) {
				t.Errorf("%s help %q; want %q", args, help, part)
			}
		}
	}
}

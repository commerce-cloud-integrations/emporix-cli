package cli

import (
	"net/http"
	"strings"
	"testing"
)

// deliveryTimeJSON and deliverySlotJSON are a delivery time and a time slot
// as create takes them.
const (
	deliveryTimeJSON = `{"name":"Weekdays","siteCode":"AcmeStore_NL","zoneId":"nl","isDeliveryDay":true,` +
		`"day":{"weekday":"MONDAY"},"timeZoneId":"Europe/Amsterdam"}`
	deliverySlotJSON = `{"shippingMethod":"post","deliveryTimeRange":{"timeFrom":"08:00","timeTo":"12:00"},` +
		`"cutOffTime":{"time":"2026-01-01T06:00:00.000Z"},"capacity":40}`
)

// deliveryTimeWriteCases are the writes of delivery times and their time
// slots, the count of the orders of a delivery window and its delivery cycle,
// with the scope of the Shipping Service's writes.
func deliveryTimeWriteCases() []writeCase {
	manage := []string{shippingManageScope}
	const window = `{"deliveryDate":"2026-11-02T12:00:00.000Z","deliveryWindowId":"d1","slotId":"s1"}`
	return []writeCase{
		{
			args: []string{"delivery-times", "create"}, file: deliveryTimeJSON, method: http.MethodPost,
			path: "/shipping/acme1/delivery-times", scopes: manage, status: http.StatusCreated, body: `{"id":"d1"}`,
			want: "{\n  \"id\": \"d1\"\n}\n",
		},
		{
			args: []string{"delivery-times", "bulk-create", "--validate-overlap", "false"},
			file: "[" + deliveryTimeJSON + ",\n" + deliveryTimeJSON + "]", method: http.MethodPost,
			path: "/shipping/acme1/delivery-times/bulk", query: "validateOverlap=false", scopes: manage,
			status: http.StatusMultiStatus, body: `[{"id":"d1","index":0,"code":201,"status":"CREATED"},` +
				`{"id":"d2","index":1,"code":201,"status":"CREATED"}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"id\": \"d1\",\n    \"index\": 0,\n    \"status\": \"CREATED\"\n  },\n" +
				"  {\n    \"code\": 201,\n    \"id\": \"d2\",\n    \"index\": 1,\n    \"status\": \"CREATED\"\n  }\n]\n",
		},
		{
			// The ID is escaped as a path segment.
			args: []string{"delivery-times", "update", "d 1"}, file: deliveryTimeJSON, method: http.MethodPut,
			path: "/shipping/acme1/delivery-times/d%201", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"delivery-times", "patch", "d1", "--validate-overlap", "true"},
			file: `[{"op":"replace","path":"/name","value":"Mondays"}]`, method: http.MethodPatch,
			path: "/shipping/acme1/delivery-times/d1", query: "validateOverlap=true", scopes: manage,
			status: http.StatusOK,
		},
		{
			args: []string{"delivery-times", "delete", "d1"}, method: http.MethodDelete,
			path: "/shipping/acme1/delivery-times/d1", scopes: manage, confirmBy: "d1", status: http.StatusNoContent,
		},
		{
			args: []string{"delivery-times", "slots", "create", "d1"}, file: deliverySlotJSON, method: http.MethodPost,
			path: "/shipping/acme1/delivery-times/d1/slots", scopes: manage, status: http.StatusCreated,
			body: `{"id":"s1"}`, want: "{\n  \"id\": \"s1\"\n}\n",
		},
		{
			args: []string{"delivery-times", "slots", "update", "d1", "s 1", "--validate-overlap", "false"},
			file: deliverySlotJSON, method: http.MethodPut, path: "/shipping/acme1/delivery-times/d1/slots/s%201",
			query: "validateOverlap=false", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"delivery-times", "slots", "patch", "d1", "s1"},
			file: `[{"op":"replace","path":"/capacity","value":50}]`, method: http.MethodPatch,
			path: "/shipping/acme1/delivery-times/d1/slots/s1", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"delivery-times", "slots", "delete", "d1", "s1"}, method: http.MethodDelete,
			path: "/shipping/acme1/delivery-times/d1/slots/s1", scopes: manage, confirmBy: "s1",
			status: http.StatusNoContent,
		},
		{
			args: []string{"delivery-times", "slots", "delete-all", "d1"}, method: http.MethodDelete,
			path: "/shipping/acme1/delivery-times/d1/slots", scopes: manage, confirmBy: "d1",
			many: "every time slot of the delivery time", status: http.StatusNoContent,
		},
		{
			args: []string{"delivery-times", "increment-counter", "--delivery-time", "d1", "--slot", "s1", "--date",
				"2026-11-02T12:00:00.000Z"},
			sent: window, method: http.MethodPost, path: "/shipping/acme1/actualDeliveryWindows/incrementCounter",
			scopes: manage, status: http.StatusOK, body: "true", want: "true\n",
		},
		{
			args: []string{"delivery-times", "increment-counter"}, file: window, method: http.MethodPost,
			path: "/shipping/acme1/actualDeliveryWindows/incrementCounter", scopes: manage, status: http.StatusOK,
			body: "true", want: "true\n",
		},
	}
}

// TestDeliveryWindowWriteFlags checks that increment-counter, whose body
// Emporix requires, fails without its flags or --file and with both a flag and
// --file, and that a delivery time write
// fails with an invalid --validate-overlap, as a usage error (exit 2) that
// sends nothing.
func TestDeliveryWindowWriteFlags(t *testing.T) {
	file := writeBody(t, `{"deliveryWindowId":"d1"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"delivery-times", "increment-counter", "--slot", "s1", "--yes"},
			want: "at least one of the flags in the group [file date] is required"},
		{args: []string{"delivery-times", "increment-counter", "--slot", "s1", "--file", file, "--yes"},
			want: "if any flags in the group [file slot] are set none of the others can be"},
		{args: []string{"delivery-times", "create", "--file", file, "--validate-overlap", "no", "--yes"},
			want: `invalid argument "no" for "--validate-overlap" flag: use true or false`},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

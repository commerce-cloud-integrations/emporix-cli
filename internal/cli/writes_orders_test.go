package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// orderWriteCases are the writes of the Order Service: orders and the
// actions on one order.
func orderWriteCases() []writeCase {
	const order = `{"siteCode":"AcmeStore_NL","currency":"EUR","entries":[{"productId":"100200","amount":1}]}`
	const orders = "/order-v2/acme1/salesorders"
	update := []string{"order.order_update"}
	return []writeCase{
		{
			args: []string{"orders", "create"}, file: order, method: http.MethodPost, path: orders,
			scopes: []string{"order.order_create"}, status: http.StatusCreated, body: `{"id":"O1001"}`,
			want: "{\n  \"id\": \"O1001\"\n}\n",
		},
		{
			args: []string{"orders", "update", "O1001"}, file: order, method: http.MethodPut, path: orders + "/O1001",
			query: "recalculate=false", scopes: update, status: http.StatusNoContent,
		},
		{
			args: []string{"orders", "patch", "O1001", "--recalculate"}, file: `{"mixins":{"note":{"text":"gift"}}}`,
			method: http.MethodPatch, path: orders + "/O1001", query: "recalculate=true", scopes: update,
			status: http.StatusNoContent,
		},
		{
			args: []string{"orders", "delete", "O1001"}, method: http.MethodDelete, path: orders + "/O1001",
			scopes: []string{"order.order_delete"}, confirmBy: "O1001", status: http.StatusNoContent,
		},
		{
			args: []string{"orders", "transition", "O1001", "--status", "SHIPPED"}, sent: `{"status":"SHIPPED"}`,
			method: http.MethodPost, path: orders + "/O1001/transitions", scopes: update, status: http.StatusNoContent,
			refuses: true,
		},
		{
			args: []string{"orders", "transition", "O1001"}, file: `{"status":"CONFIRMED"}`, method: http.MethodPost,
			path: orders + "/O1001/transitions", scopes: update, status: http.StatusNoContent, refuses: true,
		},
		{
			args:   []string{"orders", "add-entries", "O1001", "--store"},
			file:   `{"entries":[{"productId":"100200","quantity":2},{"entryId":"e1","quantity":0}]}`,
			method: http.MethodPost, path: orders + "/O1001/entries", query: "store=true", scopes: update,
			status: http.StatusOK, body: `{"id":"O1001","status":"CREATED"}`,
			want: "{\n  \"id\": \"O1001\",\n  \"status\": \"CREATED\"\n}\n",
		},
		{
			args: []string{"orders", "recalculate", "O1001"}, file: `{"entries":[{"entryId":"e1","deliveredAmount":10}]}`,
			method: http.MethodPost, path: orders + "/O1001/calculations", scopes: update, status: http.StatusOK,
			body: "Order with id: O1001 calculated successfully", // bare text, as Emporix answers (live)
			want: "\"Order with id: O1001 calculated successfully\"\n",
		},
		{
			args: []string{"orders", "split", "O1001"}, sent: `{"splitBy":"VENDOR_ID"}`, method: http.MethodPost,
			path: orders + "/O1001/split", scopes: update, status: http.StatusOK,
			body: `{"masterOrder":"O1001","splitBy":"VENDOR_ID","subOrders":[{"id":"O1002","uniqueValue":"V1"}]}`,
			want: "{\n  \"masterOrder\": \"O1001\",\n  \"splitBy\": \"VENDOR_ID\",\n  \"subOrders\": [\n    {\n" +
				"      \"id\": \"O1002\",\n      \"uniqueValue\": \"V1\"\n    }\n  ]\n}\n",
		},
		{
			args: []string{"orders", "split", "O1001"}, file: `{"splitBy":"VENDOR_ID"}`, method: http.MethodPost,
			path: orders + "/O1001/split", scopes: update, status: http.StatusOK,
		},
	}
}

// invoiceJobWriteCases are the writes of the Invoice Service: the creation
// of invoice jobs.
func invoiceJobWriteCases() []writeCase {
	invoice := []string{"invoice.invoice_manage"}
	return []writeCase{
		{
			args: []string{"invoice-jobs", "create", "--type", "MANUAL", "--order", "O1001", "--order", "O1002"},
			sent: `{"jobType":"MANUAL","orderIds":["O1001","O1002"]}`, method: http.MethodPost,
			path: "/invoice/acme1/jobs/invoices", scopes: invoice, status: http.StatusCreated, body: `{"jobId":"J1"}`,
			want: "{\n  \"jobId\": \"J1\"\n}\n",
		},
		{
			args: []string{"invoice-jobs", "create", "--type", "AUTOMATIC"}, sent: `{"jobType":"AUTOMATIC"}`,
			method: http.MethodPost, path: "/invoice/acme1/jobs/invoices", scopes: invoice, status: http.StatusCreated,
		},
		{
			args: []string{"invoice-jobs", "create"}, file: `{"jobType":"MANUAL","orderIds":["O1001"]}`,
			method: http.MethodPost, path: "/invoice/acme1/jobs/invoices", scopes: invoice, status: http.StatusCreated,
		},
	}
}

// TestCompletedOrdersRequestTheirScope checks that the writes that may
// change a COMPLETED order request order.order_update_completed besides
// order.order_update with --completed only, sending no query parameter for
// it, and say so in their help. transition and recalculate do not take it:
// Emporix moves and recalculates no COMPLETED order, whatever the scopes.
func TestCompletedOrdersRequestTheirScope(t *testing.T) {
	for _, tt := range []struct {
		args []string
		path string
	}{
		{args: []string{"orders", "update", "O1001", "--file", writeBody(t, `{}`)}, path: "/O1001"},
		{args: []string{"orders", "patch", "O1001", "--file", writeBody(t, `{}`)}, path: "/O1001"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/order-v2/acme1/salesorders"+tt.path, respond(http.StatusNoContent, ""))
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, "--completed", "--yes")...)
		if code != ExitOK {
			t.Errorf("%q: exit %d, stderr %q", tt.args, code, stderr)
		}
		fake.assertTokens(t, tokenScope([]string{"order.order_update", "order.order_update_completed"}))
		if uri := fake.gets[len(fake.gets)-1]; strings.Contains(uri, "completed") {
			t.Errorf("%q: sent to %q; want no parameter for --completed", tt.args, uri)
		}
		_, help, _ := runCLI(t, nil, slices.Concat(tt.args[:2], []string{"--help"})...)
		if !strings.Contains(help, "\n--completed also requests the order.order_update_completed scope.\n\n"+
			"Requires the order.order_update scope.") || !strings.Contains(help, "for an order that is "+
			"COMPLETED, which Emporix changes only with one more scope (requests order.order_update_completed too)") {
			t.Errorf("%q: help %q; want --completed explained", tt.args[:2], help)
		}
	}
}

// TestOrderWritesRefuseBadInput checks that the typed bodies of orders
// transition and invoice-jobs create refuse what Emporix would refuse, among
// them orders to invoice in an AUTOMATIC job, which selects its own, and a
// MANUAL job without any, and that a body comes from --file or the flags,
// not both: a usage error (exit 2) that sends nothing.
func TestOrderWritesRefuseBadInput(t *testing.T) {
	file := writeBody(t, `{}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"orders", "transition", "O1001", "--status", "LOST"},
			want: `invalid argument "LOST" for "--status" flag: use IN_CHECKOUT, CREATED, CONFIRMED, DECLINED, ` +
				`SHIPPED or COMPLETED`},
		{args: []string{"orders", "transition", "O1001"}, want: "at least one of the flags in the group [file status]"},
		{args: []string{"orders", "transition", "O1001", "--status", "SHIPPED", "--file", file},
			want: "[file status] were all set"},
		{args: []string{"orders", "split", "O1001", "--completed"}, want: "unknown flag: --completed"},
		{args: []string{"orders", "transition", "O1001", "--status", "CONFIRMED", "--completed"},
			want: "unknown flag: --completed"},
		{args: []string{"orders", "recalculate", "O1001", "--file", file, "--completed"},
			want: "unknown flag: --completed"},
		{args: []string{"orders", "add-entries", "O1001"}, want: `required flag(s) "file" not set`},
		{args: []string{"invoice-jobs", "create", "--type", "ONCE"}, want: "use MANUAL or AUTOMATIC"},
		{args: []string{"invoice-jobs", "create", "--order", "O1001"}, want: "[file type]"},
		{args: []string{"invoice-jobs", "create", "--order", "O1001", "--file", file}, want: "[file order] were all set"},
		{args: []string{"invoice-jobs", "create", "--type", "AUTOMATIC", "--order", "O1001"},
			want: "--order applies only with --type MANUAL, which requires it"},
		{args: []string{"invoice-jobs", "create", "--type", "MANUAL"},
			want: "--order applies only with --type MANUAL, which requires it"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, "--yes")...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

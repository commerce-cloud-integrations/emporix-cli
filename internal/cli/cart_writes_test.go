package cli

import (
	"net/http"
	"strings"
	"testing"
)

// cartWriteFailure is a cart write whose response names parts that failed:
// what it prints on standard error and the code it exits with.
type cartWriteFailure struct {
	args     []string // after carts, without --file and --yes
	file     string
	route    string // the path that the fake answers
	status   int
	response string
	code     int
	stderr   string
}

// run sends tt to a fake that answers it and checks that the response is
// printed and the failures listed, sanitized, with the exit code of tt.
func (tt *cartWriteFailure) run(t *testing.T) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route(tt.route, respond(tt.status, tt.response))
	args := append([]string{"carts"}, tt.args...)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(args, "--file", writeBody(t, tt.file),
		"--yes")...)
	if code != tt.code || !strings.HasPrefix(stdout, tt.response[:1]+"\n") || stderr != tt.stderr {
		t.Errorf("%q, %s: exit %d, stdout %q, stderr:\n%s\nwant exit %d and:\n%s", tt.args, tt.response, code, stdout,
			stderr, tt.code, tt.stderr)
	}
}

// cartChainOf3 is a chain of three cart commands.
const cartChainOf3 = `{"commands":[{"type":"AddCartItem"},{"type":"UpdateCartItem"},{"type":"GetCart"}]}`

// TestCartChainsListTheCommandsThatFailed checks that execute, whose
// response names commands that failed, prints the response, lists those
// commands on standard error by position and type with their message,
// sanitized, says how many did not run, and exits with the code that their
// failures share (4 when each was not found, 5 when each was a conflict),
// else 1.
func TestCartChainsListTheCommandsThatFailed(t *testing.T) {
	const execute = "/cart/acme1/carts/c1/execute"
	for _, tt := range []cartWriteFailure{
		{
			// Without --on-error resume, Emporix stops after the first that failed.
			args: []string{"execute", "c1"}, file: cartChainOf3, route: execute, status: http.StatusMultiStatus,
			response: `{"results":[{"index":0,"type":"AddCartItem","code":201},{"index":1,"type":"UpdateCartItem",` +
				`"code":404,"data":{"code":404,"message":"Cart item not found\u001b[2J"}}]}`,
			code: ExitNotFound,
			stderr: "command 1 (UpdateCartItem): 404 Not Found: Cart item not found [2J\n" +
				"Error: POST /cart/acme1/carts/c1/execute: 1 of 3 commands failed, and the last 1 did not run\n",
		},
		{
			args: []string{"execute", "c1", "--on-error", "resume"}, file: cartChainOf3, route: execute,
			status: http.StatusMultiStatus,
			response: `{"results":[{"index":0,"type":"AddCartItem","code":400,"data":{"message":"no price"}},` +
				`{"index":1,"type":"UpdateCartItem","code":409},{"index":2,"type":"GetCart","code":200}]}`,
			code: ExitError,
			stderr: "command 0 (AddCartItem): 400 Bad Request: no price\ncommand 1 (UpdateCartItem): 409 Conflict\n" +
				"Error: POST /cart/acme1/carts/c1/execute: 2 of 3 commands failed\n",
		},
		{
			args: []string{"execute", "c1", "--on-error", "resume"}, file: cartChainOf3, route: execute,
			status: http.StatusMultiStatus,
			response: `{"results":[{"index":0,"type":"AddCartItem","code":409},{"index":1,"type":"UpdateCartItem",` +
				`"code":204},{"index":2,"type":"GetCart","code":200}]}`,
			code: ExitConflict,
			stderr: "command 0 (AddCartItem): 409 Conflict\nError: POST /cart/acme1/carts/c1/execute: 1 of 3 commands " +
				"failed (what each command changes changed since it was read, or exists already: read it " +
				"again before you retry)\n",
		},
	} {
		tt.run(t)
	}
}

// TestCartBatchesListTheItemsThatFailed checks the batches of items of a
// cart: items bulk-create, whose response gives each item's code as status,
// also when Emporix answers a batch that failed in part with 400, and a
// chain whose batch command failed for some of its items, also with 207,
// each list the items that failed (Emporix carried out the others) and exit
// 1.
func TestCartBatchesListTheItemsThatFailed(t *testing.T) {
	const (
		batch  = "/cart/acme1/carts/c1/itemsBatch"
		mixed  = `[{"index":0,"status":201,"id":"1"},{"index":1,"status":400,"errorMessage":"no price for 100200"}]`
		failed = "item 1: 400 Bad Request: no price for 100200\n" +
			"Error: POST /cart/acme1/carts/c1/itemsBatch: 1 of 2 items failed\n"
	)
	items := "[" + cartItemJSON + "," + cartItemJSON + "]"
	for _, tt := range []cartWriteFailure{
		{args: []string{"items", "bulk-create", "c1"}, file: items, route: batch, status: http.StatusOK,
			response: mixed, code: ExitError, stderr: failed},
		{args: []string{"items", "bulk-create", "c1"}, file: items, route: batch, status: http.StatusBadRequest,
			response: mixed, code: ExitError, stderr: failed},
		{
			args: []string{"execute", "c1"}, route: "/cart/acme1/carts/c1/execute", status: http.StatusMultiStatus,
			file: `{"commands":[{"type":"UpdateCartItemsBatch","data":[{"id":"1","quantity":0}]},{"type":"GetCart"}]}`,
			response: `{"results":[{"index":0,"type":"UpdateCartItemsBatch","code":207,"data":[{"index":0,"id":"1",` +
				`"code":400,"message":"Invalid quantity"}]},{"index":1,"type":"GetCart","code":200}]}`,
			code: ExitError,
			stderr: "command 0 (UpdateCartItemsBatch), item 0 (1): 400 Bad Request: Invalid quantity\n" +
				"Error: POST /cart/acme1/carts/c1/execute: 1 of 2 commands failed\n",
		},
	} {
		tt.run(t)
	}
}

// TestCartChainsAreChecked checks that execute sends a JSON object of one to
// 10 commands and takes only the choices of --on-error and --versioning,
// failing otherwise as a usage error (exit 2) that sends nothing.
func TestCartChainsAreChecked(t *testing.T) {
	commands := func(n int) string {
		return `{"commands":[` + strings.Repeat(`{"type":"GetCart"},`, n-1) + `{"type":"GetCart"}]}`
	}
	for _, tt := range []struct {
		file  string
		flags []string
		want  string
	}{
		{file: commands(11), want: "POST /cart/{tenant}/carts/c1/execute: a chain sends at most 10 commands in one " +
			"request, not 11: split them"},
		{file: `[{"type":"GetCart"}]`, want: "a chain sends a JSON object whose commands is an array of one or more"},
		{file: `{"commands":[]}`, want: "a JSON object whose commands is an array of one or more commands"},
		{file: commands(1), flags: []string{"--on-error", "stop"},
			want: `invalid argument "stop" for "--on-error" flag: use fail or resume`},
		{file: commands(1), flags: []string{"--versioning", "lock"},
			want: `invalid argument "lock" for "--versioning" flag: use skip, explicit or follow`},
	} {
		fake := newFakeEmporix(t)
		args := append([]string{"carts", "execute", "c1", "--file", writeBody(t, tt.file), "--yes"}, tt.flags...)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
	code, stdout, _ := runIn(t, writeEnv(newFakeEmporix(t), false, ""), "carts", "execute", "c1", "--file",
		writeBody(t, commands(10)), "--dry-run")
	if code != ExitOK || !strings.HasPrefix(stdout, "POST ") {
		t.Errorf("10 commands: exit %d, stdout %.40q; want the request", code, stdout)
	}
}

// TestExternalPricesAreRequestedOnlyWithTheirFlag checks the help of a write
// whose body may hold external prices: --allow-external-prices adds their
// scope, which the scope sentences of the write and of its group leave out.
func TestExternalPricesAreRequestedOnlyWithTheirFlag(t *testing.T) {
	_, help, _ := runCLI(t, nil, "carts", "items", "create", "--help")
	for _, want := range []string{
		"--allow-external-prices also requests the cart.cart_manage_external_prices scope.\n\n" +
			"Requires the cart.cart_manage scope.\n",
		"--allow-external-prices   allow an item with an external price, product, fee or discount (requests " +
			"cart.cart_manage_external_prices too)",
		"--site site               site code",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help %q; want %q", help, want)
		}
	}
	_, group, _ := runCLI(t, nil, "carts", "items", "--help")
	if !strings.Contains(group, "Its writes require the cart.cart_manage scope.\n") ||
		strings.Contains(group, "cart.cart_manage_external_prices") {
		t.Errorf("group help %q; want the writes to require cart.cart_manage only", group)
	}
}

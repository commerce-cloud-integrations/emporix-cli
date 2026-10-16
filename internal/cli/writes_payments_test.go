package cli

import (
	"net/http"
	"strings"
	"testing"
)

// paymentGatewayWrites are the writes of the Payment Gateway: payment modes,
// and the payment operations, each with the scope that the reference names
// for it.
func paymentGatewayWrites() []writeCase {
	const (
		mode     = `{"code":"invoice","active":true,"provider":"INVOICE","configuration":{}}`
		modeJSON = "{\n  \"active\": true,\n  \"code\": \"invoice\",\n  \"configuration\": {},\n" +
			"  \"id\": \"pm1\",\n  \"provider\": \"INVOICE\"\n}\n"
		modes      = "/payment-gateway/acme1/paymentmodes/config"
		payments   = "/payment-gateway/acme1/payment"
		successful = `{"successful":true,"message":"Success!"}`
		success    = "{\n  \"message\": \"Success!\",\n  \"successful\": true\n}\n"
		captured   = `{"successful":true,"captureId":"c1"}`
		capture    = "{\n  \"captureId\": \"c1\",\n  \"successful\": true\n}\n"
	)
	manage := []string{"payment-gateway.paymentmodes_manage"}
	captureScopes, refundScopes := []string{"payment-gateway.capture_manage"}, []string{"payment-gateway.refund_manage"}
	return []writeCase{
		{
			args: []string{"payment-modes", "create"}, file: mode, method: http.MethodPost, path: modes, scopes: manage,
			status: http.StatusOK, body: `{"id":"pm1",` + mode[1:], want: modeJSON,
		},
		{
			args: []string{"payment-modes", "update", "pm1"}, file: `{"active":true,"configuration":{}}`,
			method: http.MethodPut, path: modes + "/pm1", scopes: manage, status: http.StatusOK,
			body: `{"id":"pm1",` + mode[1:], want: modeJSON,
		},
		{
			args: []string{"payment-modes", "delete", "pm 1"}, method: http.MethodDelete, path: modes + "/pm%201",
			scopes: manage, confirmBy: "pm 1", status: http.StatusOK,
		},
		{
			args:   []string{"payment-transactions", "authorize"},
			file:   `{"order":{"id":"1001"},"paymentModeId":"pm1","amountToAuthorize":3.99,"currency":"EUR"}`,
			method: http.MethodPost, path: payments + "/authorize", scopes: []string{"payment-gateway.authorize_manage"},
			status: http.StatusOK, body: `{"successful":true,"paymentTransactionId":"T1","requiresExternalPayment":false}`,
			want: "{\n  \"paymentTransactionId\": \"T1\",\n  \"requiresExternalPayment\": false,\n  \"successful\": true\n}\n",
		},
		{
			args: []string{"payment-transactions", "capture", "T1", "--amount", "3.990", "--currency", "EUR"},
			sent: `{"amount":3.990,"currency":"EUR"}`, method: http.MethodPost, path: payments + "/T1/capture",
			scopes: captureScopes, status: http.StatusOK, body: captured, want: capture,
		},
		{
			args: []string{"payment-transactions", "capture", "T1"}, file: `{"amount":3.99,"currency":"EUR"}`,
			method: http.MethodPost, path: payments + "/T1/capture", scopes: captureScopes, status: http.StatusOK,
			body: captured, want: capture,
		},
		{
			args: []string{"payment-transactions", "refund", "T 1", "--amount", "1", "--currency", "EUR", "--capture-id",
				"c1"},
			sent: `{"amount":1,"captureId":"c1","currency":"EUR"}`, method: http.MethodPost, path: payments + "/T%201/refund",
			scopes: refundScopes, status: http.StatusOK, body: successful, want: success,
		},
		{
			args: []string{"payment-transactions", "refund", "T1"}, file: `{"amount":1,"currency":"EUR"}`,
			method: http.MethodPost, path: payments + "/T1/refund", scopes: refundScopes, status: http.StatusOK,
			body: successful, want: success,
		},
		{
			args: []string{"payment-transactions", "cancel", "T1"}, method: http.MethodPost, path: payments + "/T1/cancel",
			scopes: []string{"payment-gateway.cancel_manage"}, status: http.StatusOK, body: successful, want: success,
		},
	}
}

// TestPaymentOperationsFailUnlessEmporixSaysTheySucceeded checks that a
// payment operation that Emporix answers with successful false, or with a
// response that does not say whether it succeeded, prints the response and
// exits with 1, naming the request and Emporix's message, if any, sanitized
// for the terminal.
func TestPaymentOperationsFailUnlessEmporixSaysTheySucceeded(t *testing.T) {
	const (
		payments = "/payment-gateway/acme1/payment"
		failed   = ": Emporix answered that it failed (successful false)"
		unknown  = ": the response does not say whether it succeeded (successful true or false), so it may have " +
			"been carried out all the same: read the item before you send it again\n"
	)
	for _, tt := range []struct {
		args       []string
		path, body string
		stdout     string
		stderr     string // the end of standard error
	}{
		{
			args: []string{"capture", "T1", "--amount", "5", "--currency", "EUR"}, path: "/T1/capture",
			body:   `{"successful":false,"message":"Amount exceeds the authorized \u001b[2Jone"}`,
			stdout: "{\n  \"message\": \"Amount exceeds the authorized \\u001b[2Jone\",\n  \"successful\": false\n}\n",
			stderr: failed + ": Amount exceeds the authorized  [2Jone\n",
		},
		{
			args: []string{"cancel", "T1"}, path: "/T1/cancel", body: `{"successful":false}`,
			stdout: "{\n  \"successful\": false\n}\n", stderr: failed + "\n",
		},
		{
			args: []string{"authorize", "--file", writeBody(t, `{}`)}, path: "/authorize",
			body:   `{"successful":false,"message":{"code":"declined"}}`,
			stdout: "{\n  \"message\": {\n    \"code\": \"declined\"\n  },\n  \"successful\": false\n}\n",
			stderr: failed + "\n",
		},
		{
			args: []string{"refund", "T1", "--file", writeBody(t, `{}`)}, path: "/T1/refund", body: `{"message":"queued"}`,
			stdout: "{\n  \"message\": \"queued\"\n}\n", stderr: unknown,
		},
		{
			args: []string{"cancel", "T1"}, path: "/T1/cancel", body: `OK`,
			stderr: "The response (2 bytes) is not JSON, so it is not printed.\nError: POST " + payments + "/T1/cancel" +
				unknown,
		},
		{args: []string{"cancel", "T1"}, path: "/T1/cancel", stderr: unknown},
	} {
		fake := newFakeEmporix(t)
		fake.route(payments+tt.path, respond(http.StatusOK, tt.body))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""),
			append(append([]string{"payment-transactions"}, tt.args...), "--yes")...)
		if want := "Error: POST " + payments + tt.path; code != ExitError || stdout != tt.stdout ||
			!strings.HasSuffix(stderr, tt.stderr) || !strings.Contains(stderr, want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 1, %q and %q", tt.args, code, stdout, stderr,
				tt.stdout, want+"…"+tt.stderr)
		}
		fake.assertRequests(t, payments+tt.path)
	}
}

// TestReportedOutcomeEndsEveryWrite checks that a write whose response says
// whether it succeeded fails where it says that it did not, also where it
// starts a job: the response is printed, and --wait does not follow the job.
func TestReportedOutcomeEndsEveryWrite(t *testing.T) {
	created := `{"id":"j1","status":"PENDING","successful":false}`
	for _, flags := range [][]string{nil, {"--wait", "--poll-interval", "1ms"}} {
		fake := newFakeEmporix(t)
		fake.route(reindexJobsPath, respond(http.StatusCreated, created))
		w := jobCreate()
		w.answers.ReportsOutcome = true
		code, stdout, stderr := runWrite(t.Context(), writeRoot(fake, w),
			append([]string{"create", "--entity-type", "PRODUCT", "--yes"}, flags...)...)
		if want := "Error: POST " + reindexJobsPath + ": Emporix answered that it failed (successful false)\n"; code !=
			ExitError || stdout != asPrinted(t, created) || stderr != want {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 1, the response and %q", flags, code, stdout, stderr,
				want)
		}
		fake.assertRequests(t, reindexJobsPath)
	}
}

// TestPaymentWritesRefuseInvalidInput checks that a payment write without
// the body it needs, with both --file and typed flags, with an amount that
// is not a number, or with a body that Emporix takes none for, is a usage
// error (exit 2) that sends nothing, also for a dry run.
func TestPaymentWritesRefuseInvalidInput(t *testing.T) {
	body := writeBody(t, `{"amount":1,"currency":"EUR"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"payment-modes", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"payment-modes", "update", "pm1"}, want: `required flag(s) "file" not set`},
		{args: []string{"payment-modes", "delete", "pm1", "--file", body}, want: "unknown flag: --file"},
		{args: []string{"payment-modes", "delete", ".."}, want: `invalid ID ".."`},
		{args: []string{"payment-transactions", "authorize"}, want: `required flag(s) "file" not set`},
		{args: []string{"payment-transactions", "capture", "T1"},
			want: "at least one of the flags in the group [file amount] is required"},
		{args: []string{"payment-transactions", "capture", "T1", "--amount", "1"},
			want: "at least one of the flags in the group [file currency] is required"},
		{args: []string{"payment-transactions", "refund", "T1", "--amount", "1", "--currency", "EUR", "--file", body},
			want: "if any flags in the group [file amount] are set none of the others can be"},
		{args: []string{"payment-transactions", "refund", "T1", "--amount", "1,50", "--currency", "EUR"},
			want: `invalid argument "1,50" for "--amount" flag: use a number such as 12.5`},
		{args: []string{"payment-transactions", "capture", "--amount", "1", "--currency", "EUR"},
			want: "accepts 1 arg(s), received 0"},
		{args: []string{"payment-transactions", "cancel", "T1", "--file", body}, want: "unknown flag: --file"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, flag)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, flag, code, stdout, stderr,
					tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

// TestPaymentOperationsSayTheyMoveMoney checks that the help of capture,
// refund and cancel says that they move money, that the help of every
// payment operation says how its outcome sets the exit code, and that their
// confirmation names the transaction.
func TestPaymentOperationsSayTheyMoveMoney(t *testing.T) {
	for _, op := range []string{"authorize", "capture", "refund", "cancel"} {
		_, help, _ := runCLI(t, nil, "payment-transactions", op, "--help")
		if moves := "\nThis moves money: the confirmation names the transaction in the URL that it shows."; op !=
			"authorize" && !strings.Contains(help, moves) || !strings.Contains(help, "\n\n"+outcomeHelp+"\n\n") {
			t.Errorf("%s help %q; want it to say that it moves money and how its outcome sets the exit code", op, help)
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/payment-gateway/acme1/payment/T1/refund", respond(http.StatusOK, `{"successful":true}`))
	code, _, stderr := runIn(t, writeEnv(fake, true, "n\n"), "payment-transactions", "refund", "T1", "--amount", "1",
		"--currency", "EUR")
	if want := "Send POST " + fake.server.URL + "/payment-gateway/acme1/payment/T1/refund to tenant acme1? [y/N] "; code !=
		ExitError || !strings.HasPrefix(stderr, want) {
		t.Errorf("exit %d, stderr %q; want the prompt %q, declined", code, stderr, want)
	}
	fake.assertNothingSent(t)
}

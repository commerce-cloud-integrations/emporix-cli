package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestAPIGetPrintsSortedJSON checks that api get prints the response with
// sorted keys and every value as sent, API keys included.
func TestAPIGetPrintsSortedJSON(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/indexing/acme1/configurations", respond(http.StatusOK,
		`[{"provider":"ALGOLIA","secured":{"writeKey":"algolia-write-sentinel"},"code":"001234"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "api", "get", "/indexing/{tenant}/configurations?expand=x",
		"-p", "pageSize=5", "-p", "q=a,b", "--scope", "indexing.search_view")
	want := "[\n  {\n    \"code\": \"001234\",\n    \"provider\": \"ALGOLIA\",\n    \"secured\": {\n" +
		"      \"writeKey\": \"algolia-write-sentinel\"\n    }\n  }\n]\n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertRequests(t, "/indexing/acme1/configurations?expand=x&pageSize=5&q=a%2Cb")
}

func TestAPIGetFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/empty", respond(http.StatusOK, ""))
	fake.route("/broken", respond(http.StatusInternalServerError, `{"message":"canned failure"}`))
	fake.route("/forbidden", respond(http.StatusForbidden, `{"message":"canned failure"}`))
	fake.route("/escape", respond(http.StatusBadRequest, `{"message":"a\u001b]0;pwned\u0007\u202eb"}`))
	fake.route("/invalid", respond(http.StatusBadRequest, `{"message":"The request violates validation constraints.",`+
		`"details":[{"field":"na\u001bme","message":"must\u202e not be blank","moreInfo":"`+testSecret+`"}]}`))
	tests := []struct {
		args    []string
		code    int
		message string
	}{
		{args: []string{"/missing"}, code: ExitNotFound, message: "GET /missing: 404 Not Found: not found"},
		// An empty item is not found, and its inline query never reaches the message.
		{args: []string{"/empty?token=" + testSecret}, code: ExitNotFound, message: "GET /empty: empty response: not found"},
		{args: []string{"/forbidden"}, code: ExitAuth, message: "403 Forbidden: canned failure"},
		{args: []string{"/broken"}, code: ExitError, message: "500 Internal Server Error: canned failure"},
		{args: []string{"/escape"}, code: ExitError, message: "400 Bad Request: a ]0;pwned  b\n"},
		// Validation details are shown, field and message only, sanitized as the message is.
		{args: []string{"/invalid"}, code: ExitError,
			message: "400 Bad Request: The request violates validation constraints. (na me: must  not be blank)\n"},
		{args: []string{"https://evil.example/x"}, code: ExitUsage, message: `API path "https://evil.example/x"`},
		{args: []string{"/products?q=%zz"}, code: ExitUsage, message: "malformed query"},
		{args: []string{"/sites", "-p", "novalue"}, code: ExitUsage,
			message: `invalid argument "novalue" for "-p, --param" flag: use key=value`},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			code, stdout, stderr := runCLI(t, fake.env(), append([]string{"api", "get"}, tt.args...)...)
			if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.message) {
				t.Fatalf("exit %d (want %d), stdout %q, stderr %q (want %q)", code, tt.code, stdout, stderr, tt.message)
			}
			assertNoSecret(t, stdout, stderr)
		})
	}
}

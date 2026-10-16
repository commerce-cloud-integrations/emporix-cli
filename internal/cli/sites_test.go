package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestSitesGetEscapesTheID checks that path escaping keeps an ID one segment
// and that {tenant} inside an ID is never substituted.
func TestSitesGetEscapesTheID(t *testing.T) {
	fake := newFakeEmporix(t)
	code, _, stderr := runCLI(t, fake.env(), "sites", "get", "a b+{tenant}/c?d")
	escaped := "/site/acme1/sites/a%20b+%7Btenant%7D%2Fc%3Fd"
	if code != ExitNotFound || !strings.Contains(stderr, "GET "+escaped) {
		t.Fatalf("exit %d, stderr %q; want 4 for %s", code, stderr, escaped)
	}
	fake.assertRequests(t, escaped)
}

// TestSiteMixinsNeedAnObject checks the mixins of a site, and one mixin: a
// response that is not a JSON object fails, also for -o json, which prints an
// object whole; an empty object shows no rows.
func TestSiteMixinsNeedAnObject(t *testing.T) {
	for _, tc := range []struct {
		args []string
		path string
	}{
		{[]string{"sites", "mixins", "list", "AcmeStore_NL"}, "/site/acme1/sites/AcmeStore_NL/mixins"},
		{[]string{"sites", "mixins", "get", "AcmeStore_NL", "m1"}, "/site/acme1/sites/AcmeStore_NL/mixins/m1"},
	} {
		for _, body := range []string{`[]`, `null`, `"text"`} {
			checkMixinRefused(t, tc.args, tc.path, body)
		}
		checkMixinsShown(t, tc.args, tc.path)
	}
}

// checkMixinRefused checks that a site's mixins, read from path, fail (exit 1)
// with stdout empty where the response body is not a JSON object, in table and
// JSON output alike.
func checkMixinRefused(t *testing.T, args []string, path, body string) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route(path, respond(http.StatusOK, body))
	for _, format := range []string{"table", "json"} {
		code, stdout, stderr := runCLI(t, fake.env(), slices.Concat(args, []string{"-o", format})...)
		if code != ExitError || stdout != "" ||
			!strings.Contains(stderr, "decoding site mixin response: expected a JSON object") {
			t.Errorf("%q -o %s, body %s: exit %d, stdout %q, stderr %q", args, format, body, code, stdout, stderr)
		}
	}
}

// checkMixinsShown checks that an empty object shows no rows, and that a
// non-empty one prints whole, as JSON, with its members sorted.
func checkMixinsShown(t *testing.T, args []string, path string) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route(path, respond(http.StatusOK, `{}`))
	if code, stdout, stderr := runCLI(t, fake.env(), args...); code != ExitOK || stdout != "NAME  VALUE\n" {
		t.Errorf("%q, empty object: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
	}
	fake.route(path, respond(http.StatusOK, `{"b":{"x":1},"a":"t"}`))
	code, stdout, stderr := runCLI(t, fake.env(), slices.Concat(args, []string{"-o", "json"})...)
	if code != ExitOK || stdout != "{\n  \"a\": \"t\",\n  \"b\": {\n    \"x\": 1\n  }\n}\n" {
		t.Errorf("%q -o json: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
	}
}

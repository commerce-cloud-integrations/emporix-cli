package cli

import (
	"strings"
	"testing"
)

func TestSiteFlagIsRequiredAndValidated(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"prices", "list"}, message: `required flag(s) "site" not set`},
		{args: []string{"prices", "for-product", "p1"}, message: `required flag(s) "site" not set`},
		{args: []string{"availability", "list"}, message: `required flag(s) "site" not set`},
		{args: []string{"availability", "get", "p1"}, message: `required flag(s) "site" not set`},
		{args: []string{"availability", "list", "--site", ".."}, message: `invalid site ".."`},
		{args: []string{"prices", "list", "--site", ""}, message: `invalid site ""`},
		{args: []string{"prices", "list", "--site", " "}, message: `invalid site " "`},
		{args: []string{"prices", "get", "p1", "--site", "main"}, message: "unknown flag: --site"},
		{args: []string{"fees", "list"}, message: `required flag(s) "site" not set`},
		{args: []string{"fees", "for-product", "p1"}, message: `required flag(s) "site" not set`},
		{args: []string{"fees", "for-item", "urn:x"}, message: `required flag(s) "site" not set`},
		{args: []string{"fees", "get", "f1", "--site", "main"}, message: "unknown flag: --site"},
		{args: []string{"shipping", "zones", "list"}, message: `required flag(s) "site" not set`},
		{args: []string{"shipping", "zones", "methods", "list", "z1"}, message: `required flag(s) "site" not set`},
		{args: []string{"shipping", "groups", "get", "g1"}, message: `required flag(s) "site" not set`},
		{args: []string{"shipping", "customers", "list"}, message: `required flag(s) "site" not set`},
		{args: []string{"shipping", "zones", "delete", "z1", "--yes"}, message: `required flag(s) "site" not set`},
		{args: []string{"segments", "match", "--site", "", "--product", "p1", "--segment", "S1", "--allow-write-scope"},
			message: `invalid site ""`},
	} {
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.message) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q (want %q)", tt.args, code, stdout, stderr, tt.message)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("usage errors must not contact the API: tokens %q, requests %q", fake.tokenScopes, fake.gets)
	}
}

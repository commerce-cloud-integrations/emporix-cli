package cli

import (
	"net/http"
	"strings"
	"testing"
)

// sessionContextWriteCases are the writes of session contexts: an update, an
// upsert, and custom attributes added and deleted.
func sessionContextWriteCases() []writeCase {
	manage := []string{"sessioncontext.context_manage"}
	const context = "/session-context/acme1/context/s1"
	const updated = `{"sessionId":"s1","siteCode":"AcmeStore_NL","currency":"EUR","metadata":{"version":3}}`
	return []writeCase{
		{
			args: []string{"session-contexts", "update", "s1"}, file: updated, method: http.MethodPut, path: context,
			query: "upsert=false", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"session-contexts", "update", "s1", "--upsert"}, file: updated, method: http.MethodPut,
			path: context, query: "upsert=true", scopes: manage, status: http.StatusNoContent,
		},
		{
			args:   []string{"session-contexts", "attributes", "create", "s1", "--key", "channel", "--value", "kiosk"},
			sent:   `{"key":"channel","value":"kiosk"}`,
			method: http.MethodPost, path: context + "/attributes", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"session-contexts", "attributes", "create", "s1"}, file: `{"key":"channel","value":{"id":"k1"}}`,
			method: http.MethodPost, path: context + "/attributes", scopes: manage, status: http.StatusNoContent,
		},
		{
			args:   []string{"session-contexts", "attributes", "delete", "s1", "channel"},
			method: http.MethodDelete, path: context + "/attributes/channel", scopes: manage, status: http.StatusNoContent,
			confirmBy: "channel",
		},
	}
}

// TestSessionAttributeNeedsKeyAndValue checks that attributes create without
// --key or --value, and with --file and either, is a usage error (exit 2)
// that sends nothing.
func TestSessionAttributeNeedsKeyAndValue(t *testing.T) {
	file := writeBody(t, `{"key":"channel","value":"kiosk"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"--key", "channel"}, want: "the flags in the group [file value] is required"},
		{args: []string{"--value", "kiosk"}, want: "the flags in the group [file key] is required"},
		{args: []string{"--file", file, "--key", "channel"}, want: "[file key] were all set"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""),
			append([]string{"session-contexts", "attributes", "create", "s1", "--yes"}, tt.args...)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

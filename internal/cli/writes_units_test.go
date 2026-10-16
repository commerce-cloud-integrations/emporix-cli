package cli

import (
	"net/http"
	"strings"
	"testing"
)

// unitWrites are the writes of units, with the scope that the reference
// names for them and the Content-Language that the Unit Service requires.
func unitWrites() []writeCase {
	scopes := []string{"unithandling.unit_manage"}
	return []writeCase{
		{
			args:   []string{"units", "create"},
			file:   `{"code":"g","name":{"en":"gram","de":"Gramm"},"type":"mass","baseUnit":false,"factor":0.0010}`,
			method: http.MethodPost, path: "/unit-handling/acme1/units", scopes: scopes, lang: "*",
			status: http.StatusCreated, body: `{"code":"g"}`, want: "{\n  \"code\": \"g\"\n}\n",
		},
		{
			args: []string{"units", "update", "kg m", "--content-language", "EN"},
			file: "{\"code\":\"kg m\",\"name\":\"kilogram metre\",\"type\":\"torque\",\"baseUnit\":true,\"factor\":1,\n" +
				"  \"metadata\":{\"version\":2}}\n",
			method: http.MethodPut, path: "/unit-handling/acme1/units/kg%20m", scopes: scopes, lang: "en",
			status: http.StatusNoContent,
		},
	}
}

// TestUnitUpdateReportsAStaleVersionAsAConflict checks that the 404 Not
// Found with which the Unit Service answers an update of a stale version
// exits with 5, a conflict, when the body carries metadata.version, and with
// 4, not found, without one.
func TestUnitUpdateReportsAStaleVersionAsAConflict(t *testing.T) {
	const message = "Cannot find the unit with code: g and version: 1. It is possible to update a unit with " +
		"the latest version only. Fetch the unit to get the latest version."
	for _, tt := range []struct {
		file string
		code int
		want string // the end of standard error
	}{
		{file: `{"code":"g","metadata":{"version":1}}`, code: ExitConflict,
			want: "404 Not Found: " + message + " (the item changed since it was read, or exists already: read it " +
				"again before you retry)\n"},
		{file: `{"code":"g"}`, code: ExitNotFound, want: "404 Not Found: " + message + "\n"},
		{file: `{"code":"g","metadata":{"version":null}}`, code: ExitNotFound, want: "404 Not Found: " + message + "\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/unit-handling/acme1/units/g", respond(http.StatusNotFound, `{"code":404,"message":"`+message+`"}`))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "units", "update", "g", "--file",
			writeBody(t, tt.file), "--yes")
		if code != tt.code || stdout != "" || !strings.HasSuffix(stderr, "PUT /unit-handling/acme1/units/g: "+tt.want) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.file, code, stdout, stderr, tt.code,
				tt.want)
		}
		fake.assertRequests(t, "/unit-handling/acme1/units/g")
	}
}

// TestUnitWritesRefuseInvalidInput checks that a unit write without a body,
// with another --content-language than * or a language code, or with a CODE
// that does not address one unit, is a usage error (exit 2) that sends
// nothing, also for a dry run.
func TestUnitWritesRefuseInvalidInput(t *testing.T) {
	const body = `{"code":"g"}`
	assertRefused(t, []refusedWrite{
		{args: []string{"units", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"units", "update", "g"}, want: `required flag(s) "file" not set`},
		{args: []string{"units", "create", "--content-language", "english"}, file: body,
			want: `invalid argument "english" for "--content-language" flag: invalid language "english": use * or a ` +
				"code such as nl or en"},
		{args: []string{"units", "create", "--content-language", ""}, file: body, want: `invalid language ""`},
		{args: []string{"units", "update"}, file: body, want: "accepts 1 arg(s), received 0"},
		{args: []string{"units", "update", ".."}, file: body, want: `invalid CODE ".."`},
		{args: []string{"units", "update", "kg/m"}, file: body,
			want: "PUT /unit-handling/{tenant}/units/kg%2Fm: a path segment of a write is empty, . or .., or holds /"},
	})
}

// TestWritesRequestAWriteCapableScope checks that a write declared without a
// write-capable scope fails as the command tree is built.
func TestWritesRequestAWriteCapableScope(t *testing.T) {
	a := &app{getenv: func(string) string { return "" }}
	for _, scopes := range [][]string{nil, {"unithandling.unit_read"}, {"a.read", "b.view"}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("a write with scopes %q was built", scopes)
				}
			}()
			(&write{use: "create", method: http.MethodPost, req: request{path: "/x"}, scopes: scopes}).command(a)
		}()
	}
}

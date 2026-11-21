package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestUserScopesShapes(t *testing.T) {
	for _, tt := range []struct {
		body, want string
		code       int
	}{
		{body: `{"userId":"x","scopes":"tenant=acme1"}`, want: "SCOPE\n", code: ExitOK},
		{body: `{"userId":"x","scopes":""}`, want: "SCOPE\n", code: ExitOK},
		{body: `{"userId":"x"}`, code: ExitError},
		{body: `{"userId":"x","scopes":null}`, code: ExitError},
		{body: `{"userId":"x","scopes":["a.read"]}`, code: ExitError},
		{body: `["a.read"]`, code: ExitError},
	} {
		fake := newFakeEmporix(t)
		fake.route("/iam/acme1/users/x/scopes", respond(http.StatusOK, tt.body))
		code, stdout, stderr := runCLI(t, fake.env(), "iam", "users", "scopes", "x")
		if code != tt.code || stdout != tt.want ||
			(code != ExitOK && !strings.Contains(stderr, "expected an object with scopes")) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", tt.body, code, stdout, stderr)
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/iam/acme1/users/x/scopes", respond(http.StatusOK, `{"userId":"x","scopes":"tenant=acme1 a.read"}`))
	code, stdout, _ := runCLI(t, fake.env(), "iam", "users", "scopes", "x", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"scopes": "tenant=acme1 a.read"`) {
		t.Fatalf("-o json: exit %d, stdout %s", code, stdout)
	}
}

func TestIAMOnePageReadsFailRatherThanTruncate(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, size string
	}{
		{args: []string{"iam", "users", "list"}, path: "/iam/acme1/users", size: "500"},
		{args: []string{"iam", "users", "access-controls", "u1"}, path: "/iam/acme1/users/u1/access-controls", size: "1000"},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, `[{"id":"x1"},{"id":"x2"}]`, "X-Total-Count", "2"))
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitOK || !strings.Contains(stdout, "x2") || len(fake.gets) != 1 ||
			!strings.Contains(fake.gets[0], "pageSize="+tt.size) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q, requests %q", tt.args, code, stdout, stderr, fake.gets)
		}
		fake = newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, `[{"id":"x1"}]`, "X-Total-Count", "1001"))
		code, stdout, stderr = runCLI(t, fake.env(), tt.args...)
		if code != ExitError || stdout != "" || !strings.Contains(stderr, "returned 1 of 1001 items") {
			t.Errorf("%v cut short: exit %d, stdout %q, stderr %q", tt.args, code, stdout, stderr)
		}
	}
}

func TestUserAccessControlsNeedTheirTotal(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/iam/acme1/users/u1/access-controls", firstPage(`[{"id":"ac1"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "iam", "users", "access-controls", "u1")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "the service reported no total") ||
		len(fake.gets) != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q, requests %q", code, stdout, stderr, fake.gets)
	}
}

func TestIAMPagingNoteIsOnListOnly(t *testing.T) {
	for _, tt := range []struct {
		group  string
		hasGet bool
	}{
		{group: "scopes", hasGet: true},
		{group: "access-controls", hasGet: true},
		{group: "templates"}, // Emporix reads no template by id
		{group: "groups", hasGet: true},
	} {
		if _, stdout, _ := runCLI(t, nil, "iam", tt.group, "list", "--help"); !containsFlat(stdout, iamPagingNote) {
			t.Errorf("iam %s list --help lacks the paging note:\n%s", tt.group, stdout)
		}
		if !tt.hasGet {
			continue
		}
		if _, stdout, _ := runCLI(t, nil, "iam", tt.group, "get", "--help"); containsFlat(stdout, iamPagingNote) {
			t.Errorf("iam %s get --help has the paging note:\n%s", tt.group, stdout)
		}
	}
}

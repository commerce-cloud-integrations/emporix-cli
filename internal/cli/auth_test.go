package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthStatus(t *testing.T) {
	fake := newFakeEmporix(t)
	code, stdout, stderr := runCLI(t, fake.env(), "auth", "status", "--scope", "site.site_read")
	if code != ExitOK || stdout != "TENANT  EXPIRES IN  SCOPES\nacme1   1h0m0s      site.site_read\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	assertNoSecret(t, stdout, stderr)
	code, stdout, _ = runCLI(t, fake.env(), "auth", "status", "-o", "json", "--scope", "b.read,a.read")
	want := "{\n  \"expiresInSeconds\": 3600,\n  \"scopes\": [\n    \"a.read\",\n    \"b.read\"\n  ],\n" +
		"  \"tenant\": \"acme1\"\n}\n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stdout %q", code, stdout)
	}
	code, stdout, _ = runCLI(t, fake.env(), "auth", "status", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"scopes": []`) {
		t.Fatalf("exit %d, stdout %q; want an empty scopes array", code, stdout)
	}
}

func TestAuthStatusFailures(t *testing.T) {
	fake := newFakeEmporix(t, "category.category_read_unpublished")
	with := func(key, value string) map[string]string {
		env := fake.env()
		env[key] = value
		return env
	}
	tests := []struct {
		name    string
		env     map[string]string
		args    []string
		code    int
		message string
	}{
		{name: "missing scope", env: fake.env(), args: []string{"--scope", "category.category_read_unpublished"},
			code: ExitAuth, message: "lacks required scopes: category.category_read_unpublished"},
		{name: "rejected credential", env: with(envClientSecret, "wrong"), code: ExitAuth,
			message: "POST /oauth/token: 401 Unauthorized: Client credentials are invalid"},
		{name: "tenant of another credential", env: with(envTenant, "other"), code: ExitAuth,
			message: "400 Bad Request: The tenant in the scope does not match"},
		{name: "missing configuration", env: map[string]string{}, code: ExitUsage,
			message: "missing configuration: set EMPORIX_TENANT (or --tenant), EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET"},
		{name: "invalid tenant", env: with(envTenant, "a/b"), code: ExitUsage, message: "invalid tenant: use letters"},
		// An empty --tenant (as from an unset shell variable) never falls back to the environment.
		{name: "empty tenant flag", env: fake.env(), args: []string{"--tenant", ""}, code: ExitUsage,
			message: "missing configuration: set a non-empty --tenant"},
		{name: "tenant flag over the environment", env: fake.env(), args: []string{"--tenant", "other"}, code: ExitAuth,
			message: "400 Bad Request: The tenant in the scope does not match"},
		{name: "insecure API URL", env: with(envAPIURL, "http://api.example.com"), code: ExitUsage,
			message: "must use https"},
		{name: "non-positive timeout", env: fake.env(), args: []string{"--timeout", "0s"}, code: ExitUsage,
			message: "timeout must be positive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(t, tt.env, append([]string{"auth", "status"}, tt.args...)...)
			if code != tt.code || !strings.Contains(stderr, tt.message) {
				t.Fatalf("exit %d (want %d), stdout %q, stderr %q (want %q)", code, tt.code, stdout, stderr, tt.message)
			}
			assertNoSecret(t, stdout, stderr)
		})
	}
}

func TestTokenEndpointNotFoundIsNotAMissingResource(t *testing.T) {
	missing := httptest.NewServer(respond(http.StatusNotFound, `{"message":"no route"}`))
	t.Cleanup(missing.Close)
	fake := newFakeEmporix(t)
	env := fake.env()
	env[envAPIURL] = missing.URL
	code, _, stderr := runCLI(t, env, "auth", "status")
	if code != ExitError || !strings.Contains(stderr, "POST /oauth/token: 404 Not Found") {
		t.Fatalf("exit %d (want %d), stderr %q", code, ExitError, stderr)
	}
}

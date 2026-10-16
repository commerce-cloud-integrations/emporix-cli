package emporix

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAuthenticateReportsGrantedToken(t *testing.T) {
	fake := newFakeEmporix(t)
	info, err := fake.client(t, "product.product_read_unpublished").Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	wantScopes := []string{"product.product_read_unpublished", "site.site_read"}
	if info.Tenant != testTenant || !slices.Equal(info.Scopes, wantScopes) || info.ExpiresIn != 14399*time.Second {
		t.Fatalf("info = %+v", info)
	}
}

func TestAuthenticateRejectsUnusableTokens(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		scopes []string
		want   string
		auth   bool // whether the error must match ErrAuth (exit code 3)
	}{
		{name: "missing scope", auth: true, status: 200,
			scopes: []string{"site.site_read", "category.category_read_unpublished"},
			body:   `{"access_token":"t","token_type":"Bearer","scope":"tenant=acme1 site.site_read"}`,
			want:   `credential for tenant "acme1" lacks required scopes: category.category_read_unpublished`},
		{name: "other tenant", auth: true, status: 200,
			body: `{"access_token":"t","token_type":"Bearer","scope":"tenant=other"}`,
			want: `token was issued for tenant "other", not "acme1"; check the credential`},
		{name: "no tenant", auth: true, status: 200, body: `{"access_token":"t","token_type":"Bearer","scope":""}`,
			want: `token was issued for tenant "", not "acme1"; check the credential`},
		{name: "no access token", status: 200, body: `{"token_type":"Bearer"}`, want: "token response has no access token"},
		{name: "malformed json", status: 200, body: `{"access_token":`, want: "token response has no access token"},
		{name: "unsupported type", status: 200, body: `{"access_token":"t","token_type":"mac","scope":"tenant=acme1"}`,
			want: `unsupported token type "mac"`},
		{name: "invalid client", auth: true, status: 401,
			body: `{"ErrorCode":"invalid_client","Error":"Client credentials are invalid"}`,
			want: "POST /oauth/token: 401 Unauthorized: Client credentials are invalid"},
		{name: "tenant mismatch", auth: true, status: 400,
			body: `{"message":"The tenant in the scope does not match the tenant of the client credentials"}`,
			want: "POST /oauth/token: 400 Bad Request: The tenant in the scope does not match " +
				"the tenant of the client credentials"},
		{name: "wrong API URL", status: 404, body: `{"message":"no route"}`,
			want: "POST /oauth/token: 404 Not Found: no route"},
		{name: "server error", status: 500, body: `{"message":"boom"}`,
			want: "POST /oauth/token: 500 Internal Server Error: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.tokenStatus, fake.tokenBody = tt.status, tt.body
			_, err := fake.client(t, tt.scopes...).Authenticate(context.Background())
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if errors.Is(err, ErrAuth) != tt.auth || errors.Is(err, ErrNotFound) {
				t.Fatalf("errors.Is(err, ErrAuth) = %v (want %v); a token failure is never ErrNotFound",
					errors.Is(err, ErrAuth), tt.auth)
			}
			if strings.Contains(err.Error(), `"t"`) || strings.Contains(err.Error(), testSecret) {
				t.Fatalf("error leaks credential material: %v", err)
			}
		})
	}
}

func TestScopeErrorIsTyped(t *testing.T) {
	fake := newFakeEmporix(t)
	_, err := fake.client(t, "brand.brand_manage").Get(context.Background(), "/brand/brands", nil)
	var scopeErr *ScopeError
	if !errors.As(err, &scopeErr) || !slices.Equal(scopeErr.Missing, []string{"brand.brand_manage"}) {
		t.Fatalf("error = %v, want ScopeError for brand.brand_manage", err)
	}
	if !errors.Is(err, ErrAuth) {
		t.Fatalf("ScopeError must match ErrAuth: %v", err)
	}
	if len(fake.requests) != 0 {
		t.Fatalf("an unverified token must not be used; %d API requests sent", len(fake.requests))
	}
}

func TestTokenTransportErrorsPropagate(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	fake.server.Close()
	if _, err := client.Authenticate(context.Background()); err == nil {
		t.Fatal("Authenticate against a closed server succeeded")
	}
}

func TestReadBodyRejectsOversizedResponses(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/big", func(w http.ResponseWriter, _ *http.Request) {
		chunk := make([]byte, 1<<20)
		for range maxBodyBytes/len(chunk) + 1 {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	})
	_, err := fake.client(t).Get(context.Background(), "/big", nil)
	if err == nil || !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("error = %v, want size limit error", err)
	}
}

// TestTokensAreRenewedBeforeTheyExpire checks that a client requests a new
// token once nine tenths of its token's lifetime passed, and not before,
// checking it as it checked the first; and that it keeps a token without a
// lifetime.
func TestTokensAreRenewedBeforeTheyExpire(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/x", respond(http.StatusOK, `{"id":"1"}`))
	client := fake.client(t)
	issued := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	clock := issued
	client.now = func() time.Time { return clock }
	renew := 14399 * time.Second * 9 / 10 // of the fake's tokens, which expire after 14399s
	for _, step := range []struct {
		at     time.Duration // after the first token was issued
		tokens int           // the token requests so far
	}{{0, 1}, {renew - time.Second, 1}, {renew, 2}, {renew + time.Hour, 2}} {
		clock = issued.Add(step.at)
		if _, err := client.Get(context.Background(), "/x", nil); err != nil || fake.tokenCalls != step.tokens {
			t.Fatalf("at %s: Get: %v after %d token requests; want %d", step.at, err, fake.tokenCalls, step.tokens)
		}
	}
	clock = clock.Add(renew)
	fake.tokenBody = `{"access_token":"t","token_type":"Bearer","expires_in":14399,"scope":"tenant=other"}`
	if _, err := client.Get(context.Background(), "/x", nil); !errors.Is(err, ErrAuth) {
		t.Fatalf("Get with a renewed token for another tenant: %v; want ErrAuth", err)
	}
	fake.tokenBody = `{"access_token":"t","token_type":"Bearer","scope":"tenant=acme1"}`
	if _, err := client.Get(context.Background(), "/x", nil); err != nil || fake.tokenCalls != 4 {
		t.Fatalf("Get: %v after %d token requests; want 4", err, fake.tokenCalls)
	}
	clock = clock.Add(100 * 365 * 24 * time.Hour)
	if _, err := client.Get(context.Background(), "/x", nil); err != nil || fake.tokenCalls != 4 {
		t.Fatalf("Get: %v after %d token requests; want a token without a lifetime kept", err, fake.tokenCalls)
	}
}

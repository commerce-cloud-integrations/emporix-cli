package emporix

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

const hostOnly = "scheme and host only"

func TestNewRefusesBadScopes(t *testing.T) {
	valid := Config{BaseURL: DefaultBaseURL, Tenant: "acme1", ClientID: "id", ClientSecret: "secret", Timeout: time.Second}
	for scope, want := range map[string]string{
		"":                "invalid scope: empty",
		" site.site_read": "without spaces",
		"a b":             "without spaces",
		"tenant=other":    "the tenant is set separately",
	} {
		cfg := valid
		if _, err := New(&cfg, scope); !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), want) {
			t.Errorf("New(scope %q) error = %v, want ErrInvalidInput with %q", scope, err, want)
		}
	}
}

func TestNewValidatesConfig(t *testing.T) {
	valid := Config{BaseURL: DefaultBaseURL, Tenant: "acme1", ClientID: "id", ClientSecret: "secret", Timeout: time.Second}
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{name: "https production", mutate: func(*Config) {}},
		{name: "loopback http", mutate: func(c *Config) { c.BaseURL = "http://127.0.0.1:8080" }},
		{name: "localhost http", mutate: func(c *Config) { c.BaseURL = "http://localhost:8080/" }},
		{name: "remote http", mutate: func(c *Config) { c.BaseURL = "http://SENTINEL.emporix.io" }, want: "must use https"},
		{name: "relative url", mutate: func(c *Config) { c.BaseURL = "/api" }, want: "scheme and host only"},
		{name: "url with query", mutate: func(c *Config) { c.BaseURL = "https://x.io/?a=1" }, want: "scheme and host only"},
		{name: "url with path", mutate: func(c *Config) { c.BaseURL = "https://x.io/api" }, want: "scheme and host only"},
		{name: "url with userinfo", mutate: func(c *Config) { c.BaseURL = "https://u:SENTINEL@x.io" }, want: hostOnly},
		{name: "url with secret query", mutate: func(c *Config) { c.BaseURL = "https://x.io?k=SENTINEL" }, want: hostOnly},
		{name: "malformed url", mutate: func(c *Config) { c.BaseURL = "https://x.io/%zzSENTINEL" }, want: hostOnly},
		{name: "url with fragment", mutate: func(c *Config) { c.BaseURL = "https://x.io#f" }, want: "scheme and host only"},
		{name: "empty tenant", mutate: func(c *Config) { c.Tenant = "" }, want: "tenant is empty"},
		{name: "tenant with slash", mutate: func(c *Config) { c.Tenant = "a/SENTINEL" }, want: "invalid tenant"},
		{name: "missing secret", mutate: func(c *Config) { c.ClientSecret = "" }, want: "client secret are required"},
		{name: "zero timeout", mutate: func(c *Config) { c.Timeout = 0 }, want: "timeout must be positive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)
			_, err := New(&cfg)
			if tt.want == "" && err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
			if tt.want != "" && (!errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), tt.want)) {
				t.Fatalf("New() error = %v, want ErrInvalidInput with %q", err, tt.want)
			}
			if err != nil && strings.Contains(err.Error(), "SENTINEL") {
				t.Fatalf("New() error echoes a rejected value: %v", err)
			}
		})
	}
}

func TestGetSendsScopedTokenAndReturnsBody(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/site/acme1/sites", respond(http.StatusOK, `[{"code":"AcmeStore_NL"}]`))
	client := fake.client(t, "site.site_read")

	for range 2 {
		body, err := client.Get(context.Background(), "/site/{tenant}/sites?expand=x", url.Values{"pageSize": {"2"}})
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if string(body) != `[{"code":"AcmeStore_NL"}]` {
			t.Fatalf("body = %s", body)
		}
	}
	if fake.tokenCalls != 1 {
		t.Fatalf("token requested %d times, want once", fake.tokenCalls)
	}
	wantForm := url.Values{
		"grant_type": {"client_credentials"}, "client_id": {testID},
		"client_secret": {testSecret}, "scope": {"tenant=acme1 site.site_read"},
	}
	if fake.tokenForm.Encode() != wantForm.Encode() {
		t.Fatalf("token form = %v, want %v", fake.tokenForm, wantForm)
	}
	req := fake.requests[0]
	for header, want := range map[string]string{
		"Authorization": "Bearer " + testToken, "Accept": "application/json", "User-Agent": "emporix-cli/test",
		"Accept-Language": "*",
	} {
		if got := req.Header.Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	if got := req.URL.Query(); got.Get("expand") != "x" || got.Get("pageSize") != "2" {
		t.Errorf("query = %v, want merged expand and pageSize", got)
	}
	if _, err := client.Get(context.Background(), "/site/{tenant}/sites?a=1&b=2", url.Values{"a": {"3"}}); err != nil {
		t.Fatal(err)
	}
	got := fake.requests[len(fake.requests)-1].URL.Query()
	if !slices.Equal(got["a"], []string{"3"}) || got.Get("b") != "2" {
		t.Errorf("query = %v, want explicit a=3 replacing the inline value and b kept", got)
	}
}

func TestGetPreservesEscapedPathSegments(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products/a b/c", respond(http.StatusOK, `{}`))
	if _, err := fake.client(t).Get(context.Background(), "/product/{tenant}/products/a%20b%2Fc", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got := fake.requests[0].URL.EscapedPath(); got != "/product/acme1/products/a%20b%2Fc" {
		t.Fatalf("escaped path = %q", got)
	}
}

func TestGetRejectsPathsThatLeaveTheAPIHost(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	invalid := []string{
		"products", "//evil.example/x", "https://evil.example/x", "mailto:a@b", "/x#frag", "%zz",
		"/products?q=%zz&pageSize=5", "/products?a=1;b=2",
	}
	for _, path := range invalid {
		_, err := client.Get(context.Background(), path, nil)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("Get(%q) error = %v, want invalid API path", path, err)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("rejected paths must not contact the API (token %d, requests %d)", fake.tokenCalls, len(fake.requests))
	}
}

func TestGetReportsAnEmptyItemAsNotFound(t *testing.T) {
	// Some services answer a read of a missing item with 200 and no body, or with 204 No Content.
	for _, status := range []int{http.StatusOK, http.StatusNoContent} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/x", respond(status, " \n"))
		_, err := fake.client(t).Get(context.Background(), "/product/{tenant}/products/x?token=SENTINEL", nil)
		// An inline query is left out of the message, as it may hold a credential.
		if !errors.Is(err, ErrNotFound) || err.Error() != "GET /product/acme1/products/x: empty response: not found" {
			t.Fatalf("%d: error = %v; want an empty response classified not found, without the query", status, err)
		}
	}
}

func TestGetNamesAnEmptyItemByItsEscapedPath(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products/a b/c?d", respond(http.StatusOK, "")) // routed unescaped
	query := url.Values{"secret": {"SENTINEL2"}}
	_, err := fake.client(t).Get(context.Background(), "/product/{tenant}/products/a%20b%2Fc%3Fd?token=SENTINEL", query)
	const want = "GET /product/acme1/products/a%20b%2Fc%3Fd: empty response: not found"
	if !errors.Is(err, ErrNotFound) || err.Error() != want {
		t.Fatalf("error = %v; want the escaped path without either query", err)
	}
}

func TestGetMapsErrorResponses(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
		want       string
		class      error
	}{
		{name: "service message", status: 404, body: `{"code":404,"message":"The product with id x has not been found."}`,
			want: "GET /product/acme1/products/x: 404 Not Found: The product with id x has not been found.", class: ErrNotFound},
		{name: "gateway fault", status: 401, body: `{"fault":{"faultstring":"Invalid Access Token"}}`,
			want: "401 Unauthorized: Invalid Access Token", class: ErrAuth},
		{name: "plain text", status: 400, body: "  bad request  ", want: "400 Bad Request: bad request"},
		{name: "empty body", status: 403, body: "", want: "403 Forbidden", class: ErrAuth},
		{name: "redirect not followed", status: 302, body: "", want: "302 Found"},
		// A read answered 409 is no conflict of its own: only a write's is ErrConflict.
		{name: "conflict", status: 409, body: `{"message":"no such group"}`, want: "409 Conflict: no such group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/product/acme1/products/x", respond(tt.status, tt.body, "Location", "https://evil.example/"))
			_, err := fake.client(t).Get(context.Background(), "/product/{tenant}/products/x", nil)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Status != tt.status || !strings.HasSuffix(err.Error(), tt.want) {
				t.Fatalf("error = %v, want APIError ending %q", err, tt.want)
			}
			for _, class := range []error{ErrAuth, ErrNotFound, ErrConflict, ErrInvalidInput} {
				if want := errors.Is(tt.class, class); errors.Is(err, class) != want {
					t.Fatalf("errors.Is(err, %v) = %v, want %v", class, !want, want)
				}
			}
		})
	}
}

// cutShort answers with status and a body cut short of the length it
// declares, which cannot be read whole.
func cutShort(status int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"message":"cut`))
	}
}

// TestFailuresWhoseBodyIsCutShortKeepTheirStatus checks that a failure whose
// body cannot be read whole is classified by its status, without a message:
// a write's 409 stays a conflict, a read's 404 not found, and a read's 503 is
// sent again. A success that cannot be read whole fails.
func TestFailuresWhoseBodyIsCutShortKeepTheirStatus(t *testing.T) {
	fake := newFakeEmporix(t)
	for path, status := range map[string]int{
		"/conflict": http.StatusConflict, "/missing": http.StatusNotFound, "/busy": http.StatusServiceUnavailable,
		"/ok": http.StatusOK,
	} {
		fake.route(path, cutShort(status))
	}
	client := fake.client(t)
	_, err := send(client, &Write{Method: http.MethodPut, Path: "/conflict", Body: jsonBody(t, `{}`)})
	if !errors.Is(err, ErrConflict) || err.Error() != "PUT /conflict: 409 Conflict" {
		t.Errorf("write: error %v; want the conflict without a message", err)
	}
	ctx := context.Background()
	if _, err := client.Get(ctx, "/missing", nil); !errors.Is(err, ErrNotFound) ||
		err.Error() != "GET /missing: 404 Not Found" {
		t.Errorf("read: error %v; want not found without a message", err)
	}
	sent := len(fake.requests)
	if _, err := client.Get(ctx, "/busy", nil); err == nil || err.Error() != "GET /busy: 503 Service Unavailable" ||
		len(fake.requests)-sent != 3 {
		t.Errorf("read: error %v after %d requests; want 503 after 3", err, len(fake.requests)-sent)
	}
	if _, err := client.Get(ctx, "/ok", nil); err == nil || !strings.Contains(err.Error(), "reading response") {
		t.Errorf("read: error %v; want the body that cannot be read", err)
	}
}

// TestFailuresToSendLeaveTheQueryOut checks that a failure to send, a write's
// or a read's, names the URL without its query, which may hold values a log
// must not keep, and keeps the failure it wraps.
func TestFailuresToSendLeaveTheQueryOut(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	client.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/oauth/token" {
			return http.DefaultTransport.RoundTrip(req)
		}
		return nil, &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}
	})
	_, write := send(client, &Write{Method: http.MethodPost, Path: "/x?token=SENTINEL", Body: jsonBody(t, `{}`)})
	_, read := client.Get(context.Background(), "/x", url.Values{"token": {"SENTINEL"}})
	for _, err := range []error{write, read} {
		if !errors.Is(err, syscall.ECONNRESET) || strings.Contains(err.Error(), "SENTINEL") ||
			!strings.Contains(err.Error(), `"`+fake.server.URL+`/x"`) {
			t.Errorf("error %v; want the failure to send, naming the URL without its query", err)
		}
	}
}

func TestGetRetriesTransientFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	calls := 0
	fake.route("/site/acme1/sites", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			respond(http.StatusServiceUnavailable, `{"message":"busy"}`, "Retry-After", "0")(w, r)
			return
		}
		respond(http.StatusOK, `[]`)(w, r)
	})
	if _, err := fake.client(t).Get(context.Background(), "/site/{tenant}/sites", nil); err != nil {
		t.Fatalf("Get after retries: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestGetGivesUpAfterMaxAttempts(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/site/acme1/sites", respond(http.StatusTooManyRequests, `{"message":"slow down"}`))
	_, err := fake.client(t).Get(context.Background(), "/site/{tenant}/sites", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTooManyRequests || len(fake.requests) != 3 {
		t.Fatalf("error = %v after %d requests, want 429 after 3", err, len(fake.requests))
	}
}

func TestGetStopsRetryingWhenContextEnds(t *testing.T) {
	fake := newFakeEmporix(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fake.route("/site/acme1/sites", func(w http.ResponseWriter, r *http.Request) {
		cancel()
		respond(http.StatusBadGateway, "")(w, r)
	})
	client := fake.client(t)
	client.retry.wait = time.Hour
	client.retry.maxWait = time.Hour
	if _, err := client.Get(ctx, "/site/{tenant}/sites", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRetryDelay(t *testing.T) {
	policy := retryPolicy{attempts: 3, wait: 100 * time.Millisecond, maxWait: time.Second}
	tests := []struct {
		attempt    int
		retryAfter string
		want       time.Duration
	}{
		{attempt: 1, want: 100 * time.Millisecond},
		{attempt: 2, want: 200 * time.Millisecond},
		{attempt: 5, want: time.Second},
		{attempt: 1, retryAfter: "0", want: 0},
		{attempt: 1, retryAfter: "120", want: time.Second},
		{attempt: 1, retryAfter: "soon", want: 100 * time.Millisecond},
		{attempt: 1, retryAfter: "9223372037", want: time.Second},
		{attempt: 1, retryAfter: time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), want: 0},
		{attempt: 1, retryAfter: time.Now().Add(time.Hour).UTC().Format(http.TimeFormat), want: time.Second},
	}
	for _, tt := range tests {
		header := http.Header{}
		if tt.retryAfter != "" {
			header.Set("Retry-After", tt.retryAfter)
		}
		if got := policy.delay(tt.attempt, header); got != tt.want {
			t.Errorf("delay(%d, %q) = %v, want %v", tt.attempt, tt.retryAfter, got, tt.want)
		}
	}
}

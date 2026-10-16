package emporix

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

const (
	testTenant = "acme1"
	testID     = "client-id"
	testSecret = "s3cret value&="
	testToken  = "tok-123"
)

// fakeEmporix is an httptest server that issues tokens and serves canned
// responses, recording what the client sent.
type fakeEmporix struct {
	t      *testing.T
	server *httptest.Server

	mu          sync.Mutex
	tokenForm   url.Values
	tokenCalls  int
	tokenStatus int
	tokenBody   string
	requests    []*http.Request
	bodies      []string // the body of every request in requests
	routes      map[string]func(http.ResponseWriter, *http.Request)
}

func newFakeEmporix(t *testing.T) *fakeEmporix {
	t.Helper()
	f := &fakeEmporix{
		t:           t,
		tokenStatus: http.StatusOK,
		tokenBody: `{"access_token":"` + testToken + `","token_type":"Bearer","expires_in":14399,` +
			`"scope":"tenant=acme1 site.site_read product.product_read_unpublished"}`,
		routes: map[string]func(http.ResponseWriter, *http.Request){},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeEmporix) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Method == http.MethodPost && r.URL.Path == "/oauth/token" {
		if err := r.ParseForm(); err != nil {
			f.t.Errorf("token form: %v", err)
		}
		f.tokenForm = r.PostForm
		f.tokenCalls++
		w.WriteHeader(f.tokenStatus)
		_, _ = w.Write([]byte(f.tokenBody))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		f.t.Errorf("request body: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	f.requests = append(f.requests, r.Clone(r.Context()))
	f.bodies = append(f.bodies, string(body))
	handler, ok := f.routes[r.URL.Path]
	if !ok {
		http.Error(w, `{"message":"no route"}`, http.StatusNotFound)
		return
	}
	handler(w, r)
}

func (f *fakeEmporix) route(path string, handler func(http.ResponseWriter, *http.Request)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[path] = handler
}

func (f *fakeEmporix) config() Config {
	return Config{
		BaseURL:        f.server.URL,
		Tenant:         testTenant,
		ClientID:       testID,
		ClientSecret:   testSecret,
		Timeout:        5 * time.Second,
		UserAgent:      "emporix-cli/test",
		AcceptLanguage: "*",
	}
}

func (f *fakeEmporix) client(t *testing.T, scopes ...string) *Client {
	t.Helper()
	cfg := f.config()
	client, err := New(&cfg, scopes...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	client.retry = retryPolicy{attempts: 3, wait: time.Millisecond, maxWait: 10 * time.Millisecond}
	return client
}

func respond(status int, body string, headers ...string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		for i := 0; i+1 < len(headers); i += 2 {
			w.Header().Set(headers[i], headers[i+1])
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

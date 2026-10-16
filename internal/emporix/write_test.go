package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
)

const writeBody = "{ \"code\": \"box\",\n  \"factor\": 1.50 }\n"

// send prepares w and sends it with client.
func send(client *Client, w *Write) (json.RawMessage, error) {
	req, err := client.Prepare(w)
	if err != nil {
		return nil, err
	}
	return client.Send(context.Background(), req)
}

// TestSendWritesAndReturnsTheBody checks a write: its method, path with the
// tenant and query (explicit values replacing inline ones), headers and body
// byte for byte, with a token for exactly the tenant and scopes, and the
// response as sent.
func TestSendWritesAndReturnsTheBody(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products", respond(http.StatusCreated, `{"id":"100200"}`))
	fake.tokenBody = `{"access_token":"` + testToken + `","token_type":"Bearer",` +
		`"scope":"tenant=acme1 product.product_manage"}`
	body, err := send(fake.client(t, "product.product_manage"), &Write{
		Method: http.MethodPost, Path: "/product/{tenant}/products?doIndex=true&keep=1",
		Query: url.Values{"doIndex": {"false"}}, Body: jsonBody(t, writeBody),
	})
	if err != nil || string(body) != `{"id":"100200"}` {
		t.Fatalf("body %s, error %v", body, err)
	}
	if got := fake.tokenForm.Get("scope"); got != "tenant=acme1 product.product_manage" {
		t.Fatalf("token scope %q", got)
	}
	req := fake.requests[0]
	if req.Method != http.MethodPost || req.URL.RequestURI() != "/product/acme1/products?doIndex=false&keep=1" ||
		fake.bodies[0] != writeBody {
		t.Fatalf("sent %s %s with body %q", req.Method, req.URL.RequestURI(), fake.bodies[0])
	}
	for header, want := range map[string]string{
		"Authorization": "Bearer " + testToken, "Accept": "application/json", "Accept-Language": "*",
		"Content-Type": "application/json", "User-Agent": "emporix-cli/test",
	} {
		if got := req.Header.Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

// TestSendAcceptsEvery2xx checks that any 2xx answers a write, and that one
// without a body (201 or 204) returns nothing; a write without a body sends
// none and no Content-Type.
func TestSendAcceptsEvery2xx(t *testing.T) {
	for status, answer := range map[int]string{
		http.StatusOK: `[1]`, http.StatusCreated: "", http.StatusAccepted: `{"jobId":"j1"}`,
		http.StatusNoContent: "", http.StatusMultiStatus: `[{"status":201}]`,
	} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/p 1", respond(status, answer))
		body, err := send(fake.client(t), &Write{Method: http.MethodDelete, Path: "/product/{tenant}/products/p%201"})
		if err != nil || string(body) != answer || (answer == "" && body != nil) {
			t.Errorf("%d: body %q, error %v; want %q", status, body, err, answer)
		}
		if req := fake.requests[0]; fake.bodies[0] != "" || req.Header.Get("Content-Type") != "" ||
			req.URL.EscapedPath() != "/product/acme1/products/p%201" {
			t.Errorf("%d: sent %s with body %q and headers %v", status, req.URL.EscapedPath(), fake.bodies[0], req.Header)
		}
	}
}

// TestSendClassifiesFailures checks that a refused write is an APIError of
// its class: a conflict (409) is ErrConflict and nothing else, so callers
// tell it apart without the status.
func TestSendClassifiesFailures(t *testing.T) {
	for status, class := range map[int]error{
		http.StatusConflict: ErrConflict, http.StatusNotFound: ErrNotFound, http.StatusForbidden: ErrAuth,
		http.StatusBadRequest: nil, http.StatusInternalServerError: nil, http.StatusFound: nil,
	} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/p1", respond(status, `{"message":"canned failure"}`))
		_, err := send(fake.client(t), &Write{
			Method: http.MethodPut, Path: "/product/{tenant}/products/p1?token=SENTINEL", Body: jsonBody(t, `{}`),
		})
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != status ||
			!strings.HasPrefix(err.Error(), "PUT /product/acme1/products/p1: ") || strings.Contains(err.Error(), "SENTINEL") {
			t.Errorf("%d: error %v; want an APIError naming the write without its query", status, err)
		}
		for _, other := range []error{ErrConflict, ErrNotFound, ErrAuth, ErrInvalidInput} {
			if want := errors.Is(class, other); errors.Is(err, other) != want {
				t.Errorf("%d: errors.Is(err, %v) = %v, want %v", status, other, !want, want)
			}
		}
	}
}

// TestPrepareShowsTheRequestSent checks the request that a dry run shows: the
// full URL with the tenant and query, and every header but Authorization;
// it requests no token and sends nothing. What it shows is a copy, and the
// request sent is the one prepared, whatever happens to the write after.
func TestPrepareShowsTheRequestSent(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/customer/acme1/customers/c 1", respond(http.StatusNoContent, ""))
	fake.tokenBody = `{"access_token":"` + testToken + `","token_type":"Bearer",` +
		`"scope":"tenant=acme1 customer.customer_manage"}`
	client := fake.client(t, "customer.customer_manage")
	raw := []byte(`{"firstName":"Ann"}`)
	body, err := JSONBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	write := &Write{
		Method: http.MethodPatch, Path: "/customer/{tenant}/customers/c%201", Query: url.Values{"b": {"2"}, "a": {"1"}},
		Body: body,
	}
	req, err := client.Prepare(write)
	if err != nil {
		t.Fatal(err)
	}
	want := http.Header{
		"Accept": {"application/json"}, "Accept-Language": {"*"}, "Content-Type": {"application/json"},
		"User-Agent": {"emporix-cli/test"},
	}
	if req.Method() != http.MethodPatch || req.URL() != fake.server.URL+"/customer/acme1/customers/c%201?a=1&b=2" ||
		string(req.Body()) != `{"firstName":"Ann"}` || !equalHeaders(req.Header(), want) {
		t.Fatalf("request %s %s %v %s", req.Method(), req.URL(), req.Header(), req.Body())
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 || client.Tenant() != testTenant {
		t.Fatalf("token %d, requests %d, tenant %q; want none sent", fake.tokenCalls, len(fake.requests), client.Tenant())
	}
	req.Header().Set("Authorization", "Bearer SENTINEL")
	req.Body()[0] = '['
	raw[0], write.Path = '[', "/elsewhere"
	if _, err := client.Send(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if sent := fake.requests[0]; sent.URL.RequestURI() != "/customer/acme1/customers/c%201?a=1&b=2" ||
		fake.bodies[0] != `{"firstName":"Ann"}` || sent.Header.Get("Authorization") != "Bearer "+testToken {
		t.Fatalf("sent %s with body %q and %q", sent.URL.RequestURI(), fake.bodies[0], sent.Header.Get("Authorization"))
	}
	if body := (&Request{}).Body(); body != nil {
		t.Fatalf("body of a request without one: %q", body)
	}
}

// TestSendOnlyWhatItsClientPrepared checks that a client sends only a
// request that it prepared, with its own host and scopes, and nothing for
// another client's or one that no client prepared.
func TestSendOnlyWhatItsClientPrepared(t *testing.T) {
	fake := newFakeEmporix(t)
	other, err := newFakeEmporix(t).client(t).Prepare(&Write{Method: http.MethodDelete, Path: "/x"})
	if err != nil {
		t.Fatal(err)
	}
	for _, req := range []*Request{other, {method: http.MethodDelete, url: fake.server.URL + "/x"}} {
		if _, err := fake.client(t).Send(context.Background(), req); !errors.Is(err, ErrInvalidInput) ||
			!strings.Contains(err.Error(), "a write is sent by the client that prepared it") {
			t.Errorf("%s %s: error %v; want ErrInvalidInput", req.Method(), req.URL(), err)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("token %d, requests %d; want nothing sent", fake.tokenCalls, len(fake.requests))
	}
}

func equalHeaders(got, want http.Header) bool {
	if len(got) != len(want) {
		return false
	}
	for key, values := range want {
		if strings.Join(got[key], ",") != strings.Join(values, ",") {
			return false
		}
	}
	return true
}

// TestWritesRejectInvalidRequests checks that a write that cannot be sent as
// declared is ErrInvalidInput, found before a token is requested.
func TestWritesRejectInvalidRequests(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	for _, tt := range []struct {
		write Write
		want  string
	}{
		{write: Write{Method: http.MethodGet, Path: "/x"}, want: `"GET": a write is sent as POST, PUT, PATCH or DELETE`},
		{write: Write{Method: "post", Path: "/x"}, want: "a write is sent as"},
		{write: Write{Method: http.MethodPost, Path: "https://evil.example/x"}, want: `API path "https://evil.example/x"`},
		{write: Write{Method: http.MethodPost, Path: "/x?token=SENTINEL&a=%zz"}, want: `API path "/x": malformed query`},
		{write: Write{Method: http.MethodDelete, Path: "/products/a%2F..%2Fb"},
			want: "DELETE /products/a%2F..%2Fb: a path segment of a write is empty, . or .., or holds /, \\ or %"},
		{write: Write{Method: http.MethodDelete, Path: "/products//b"}, want: "a path segment of a write is empty"},
		{write: Write{Method: http.MethodPatch, Path: "/products/%2e%2e"}, want: "a path segment of a write is empty"},
		{write: Write{Method: http.MethodPost, Path: "/chat", Header: http.Header{"Session-Id": {"s1\r\nX-Other: 1"}}},
			want: "POST /chat: the header Session-Id holds a control character"},
		{write: Write{Method: http.MethodPost, Path: "/chat", Header: http.Header{"Session-Id": {"s1", "s\x7f"}}},
			want: "the header Session-Id holds a control character"},
	} {
		if _, err := client.Prepare(&tt.write); !errors.Is(err, ErrInvalidInput) ||
			!strings.Contains(err.Error(), tt.want) || strings.Contains(err.Error(), "SENTINEL") {
			t.Errorf("%s %s: error %v; want ErrInvalidInput with %q", tt.write.Method, tt.write.Path, err, tt.want)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("token %d, requests %d; want nothing sent", fake.tokenCalls, len(fake.requests))
	}
}

// TestHeaderValuesHoldNoControlCharacters checks which header values can be
// sent: any text, with spaces and tabs, but no other control character.
func TestHeaderValuesHoldNoControlCharacters(t *testing.T) {
	for value, sendable := range map[string]bool{
		"0499bef6-a4d2-45b9-86a0-2edf7d70d726": true, "a b\tc": true, "": true, "ü": true,
		"a\nb": false, "a\rb": false, "a\x00": false, "\x7f": false,
	} {
		if err := CheckHeaderValue(value); (err == nil) != sendable {
			t.Errorf("%q: %v; want sendable %v", value, err, sendable)
		}
	}
}

// TestWritesFollowTheRetryRule checks the retry rule: a 429 is sent again
// for any write, and no write after a 502, 503 or 504, which may have
// reached the service, whatever its method (a PUT may act twice too).
func TestWritesFollowTheRetryRule(t *testing.T) {
	for _, tt := range []struct {
		method string
		status int
		sent   int
	}{
		{method: http.MethodPost, status: http.StatusServiceUnavailable, sent: 1},
		{method: http.MethodPatch, status: http.StatusBadGateway, sent: 1},
		{method: http.MethodPost, status: http.StatusGatewayTimeout, sent: 1},
		{method: http.MethodPut, status: http.StatusServiceUnavailable, sent: 1},
		{method: http.MethodDelete, status: http.StatusGatewayTimeout, sent: 1},
		{method: http.MethodPut, status: http.StatusTooManyRequests, sent: 3},
		{method: http.MethodDelete, status: http.StatusTooManyRequests, sent: 3},
		{method: http.MethodPost, status: http.StatusTooManyRequests, sent: 3},
		{method: http.MethodPatch, status: http.StatusTooManyRequests, sent: 3},
		{method: http.MethodPost, status: http.StatusInternalServerError, sent: 1},
		{method: http.MethodPut, status: http.StatusConflict, sent: 1},
	} {
		fake := newFakeEmporix(t)
		fake.route("/x", respond(tt.status, `{"message":"busy"}`, "Retry-After", "0"))
		_, err := send(fake.client(t), &Write{Method: tt.method, Path: "/x", Body: jsonBody(t, `{}`)})
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != tt.status || len(fake.requests) != tt.sent {
			t.Errorf("%s answered %d: error %v after %d requests; want %d", tt.method, tt.status, err,
				len(fake.requests), tt.sent)
		}
	}
}

// answers answers the requests in turn with statuses (each with its own
// message), the last one again for any further request.
func answers(statuses ...int) func(http.ResponseWriter, *http.Request) {
	sent := 0
	return func(w http.ResponseWriter, r *http.Request) {
		status := statuses[min(sent, len(statuses)-1)]
		sent++
		respond(status, `{"message":"answer `+strconv.Itoa(sent)+`"}`, "Retry-After", "0")(w, r)
	}
}

// TestWritesThatMayHaveBeenCarriedOutSaySo checks that a write that fails
// after an attempt that may have been carried out (a 500, 502, 503 or 504,
// or a failure once the connection was made) says so, is not sent again, and
// has no class: what would answer it again (a 409 for the version it set, a
// 404 for the item it deleted, a 403) is never asked. A 429 was not acted
// on, so the answer after it is classified as any; a 4xx or 501 says that
// nothing was done, without the note.
func TestWritesThatMayHaveBeenCarriedOutSaySo(t *testing.T) {
	for _, tt := range []struct {
		method   string
		statuses []int
		sent     int
		want     string // the error, before the note
		class    error
		note     bool
	}{
		{method: http.MethodPut, statuses: []int{503, 409}, sent: 1, want: "503 Service Unavailable: answer 1", note: true},
		{method: http.MethodDelete, statuses: []int{504, 404}, sent: 1, want: "504 Gateway Timeout: answer 1", note: true},
		{method: http.MethodPut, statuses: []int{502, 403}, sent: 1, want: "502 Bad Gateway: answer 1", note: true},
		{method: http.MethodPost, statuses: []int{503}, sent: 1, want: "503 Service Unavailable: answer 1", note: true},
		{method: http.MethodPatch, statuses: []int{502}, sent: 1, want: "502 Bad Gateway: answer 1", note: true},
		{method: http.MethodPost, statuses: []int{429, 409}, sent: 2, want: "409 Conflict: answer 2", class: ErrConflict},
		{method: http.MethodDelete, statuses: []int{429, 404}, sent: 2, want: "404 Not Found: answer 2", class: ErrNotFound},
		{method: http.MethodPut, statuses: []int{429, 403}, sent: 2, want: "403 Forbidden: answer 2", class: ErrAuth},
		{method: http.MethodPost, statuses: []int{500}, sent: 1, want: "500 Internal Server Error: answer 1", note: true},
		{method: http.MethodPost, statuses: []int{501}, sent: 1, want: "501 Not Implemented: answer 1"},
		{method: http.MethodPatch, statuses: []int{400}, sent: 1, want: "400 Bad Request: answer 1"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/x", answers(tt.statuses...))
		_, err := send(fake.client(t), &Write{Method: tt.method, Path: "/x", Body: jsonBody(t, `{}`)})
		want := tt.method + " /x: " + tt.want
		if tt.note {
			want += " (" + carriedOutNote + ")"
		}
		if err == nil || err.Error() != want || len(fake.requests) != tt.sent {
			t.Errorf("%s answered %v: error %v after %d requests; want %q after %d", tt.method, tt.statuses, err,
				len(fake.requests), want, tt.sent)
		}
		for _, class := range []error{ErrConflict, ErrNotFound, ErrAuth} {
			if want := errors.Is(tt.class, class); errors.Is(err, class) != want {
				t.Errorf("%s answered %v: errors.Is(err, %v) = %v", tt.method, tt.statuses, class, !want)
			}
		}
	}
	fake := newFakeEmporix(t)
	client := fake.client(t)
	client.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/oauth/token" {
			return http.DefaultTransport.RoundTrip(req)
		}
		return nil, &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}
	})
	_, err := send(client, &Write{Method: http.MethodPost, Path: "/x", Body: jsonBody(t, `{}`)})
	if !errors.Is(err, syscall.ECONNRESET) || !strings.HasSuffix(err.Error(), "connection reset by peer ("+
		carriedOutNote+")") {
		t.Fatalf("error %v; want the failure to send and the note", err)
	}
}

// failingDials fails the first connections it is asked for but those of
// token requests with err, as a dial that fails, then sends requests as next
// does.
type failingDials struct {
	mu     sync.Mutex
	fail   int // connections still to fail
	failed int
	err    error
	next   http.RoundTripper
}

func (f *failingDials) RoundTrip(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail > 0 && req.URL.Path != "/oauth/token" {
		f.fail--
		f.failed++
		return nil, &net.OpError{Op: "dial", Net: "tcp", Err: f.err}
	}
	return f.next.RoundTrip(req)
}

// TestConnectionsNotMadeAreRetried checks that a connection that could not
// be made, which sent nothing, is tried again for any request, a POST write
// and a read alike, up to the attempts, whatever the operating system calls
// the failure, but not one whose making timed out, and that no other failure
// to send is.
func TestConnectionsNotMadeAreRetried(t *testing.T) {
	refused := os.NewSyscallError("connect", syscall.ECONNREFUSED)
	for _, tt := range []struct {
		fail     int
		err      error
		attempts int // the connections asked for
		sent     bool
	}{
		{fail: 2, err: refused, attempts: 3, sent: true},
		{fail: 2, err: syscall.Errno(10061), attempts: 3, sent: true}, // WSAECONNREFUSED, as Windows refuses one
		{fail: 2, err: &net.DNSError{Err: "no such host", Name: "api.example", IsNotFound: true}, attempts: 3, sent: true},
		{fail: 3, err: refused, attempts: 3},
		{fail: 1, err: os.ErrDeadlineExceeded, attempts: 1},
	} {
		fake := newFakeEmporix(t)
		fake.route("/x", respond(http.StatusOK, `{"id":"x"}`))
		client := fake.client(t)
		transport := &failingDials{fail: tt.fail, err: tt.err, next: http.DefaultTransport}
		client.http.Transport = transport
		_, err := send(client, &Write{Method: http.MethodPost, Path: "/x", Body: jsonBody(t, `{}`)})
		asked := transport.failed + len(fake.requests)
		if asked != tt.attempts || (err == nil) != tt.sent || len(fake.requests) != map[bool]int{true: 1}[tt.sent] ||
			(err != nil && (!errors.Is(err, tt.err) || strings.Contains(err.Error(), carriedOutNote))) {
			t.Errorf("%v failing %d: error %v after %d connections, %d sent; want %d connections, sent %v",
				tt.err, tt.fail, err, asked, len(fake.requests), tt.attempts, tt.sent)
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/x", respond(http.StatusOK, `{"id":"x"}`))
	client := fake.client(t)
	client.http.Transport = &failingDials{fail: 1, err: refused, next: http.DefaultTransport}
	if _, err := client.Get(context.Background(), "/x", nil); err != nil || len(fake.requests) != 1 {
		t.Fatalf("read: error %v, %d requests; want it sent again", err, len(fake.requests))
	}
	attempts := 0
	client.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}
	})
	if _, err := client.Get(context.Background(), "/x", nil); err == nil || attempts != 1 {
		t.Fatalf("error %v after %d attempts; want another failure to send not repeated", err, attempts)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

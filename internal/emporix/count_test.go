package emporix

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// TestCountReadsTheTotalOfAHeadRequest checks that a count sends HEAD,
// without a body (also from a client made by WithBody), asking for the total,
// with the user's filter joined to the read's own q, and returns the total.
func TestCountReadsTheTotalOfAHeadRequest(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/order-v2/acme1/salesorders", respond(http.StatusOK, "", "X-Total-Count", "1250"))
	client := fake.client(t)
	for _, counting := range []*Client{client, withBody(t, client, http.MethodPost, searchBody)} {
		total, err := counting.Count(context.Background(), "/order-v2/{tenant}/salesorders",
			map[string][]string{"q": {"siteCode:main"}}, "status:CREATED")
		if err != nil || total != 1250 {
			t.Fatalf("total %d, error %v; want 1250", total, err)
		}
	}
	for i, req := range fake.requests {
		if req.Method != http.MethodHead || fake.bodies[i] != "" || req.Header.Get("Content-Type") != "" ||
			req.Header.Get("X-Total-Count") != "true" || req.URL.Query().Get("q") != "siteCode:main status:CREATED" {
			t.Fatalf("request %d: %s %s with body %q and headers %v", i+1, req.Method, req.URL.RequestURI(),
				fake.bodies[i], req.Header)
		}
	}
}

// TestCountFailures checks that a count fails on a response without a valid
// total and on a failed request, classified as any read's, naming HEAD.
func TestCountFailures(t *testing.T) {
	for name, tt := range map[string]struct {
		route func(http.ResponseWriter, *http.Request)
		class error
		want  string
	}{
		"no total": {route: respond(http.StatusOK, ""),
			want: "HEAD /orders: the service reported no count in X-Total-Count"},
		"invalid total": {route: respond(http.StatusOK, "", "X-Total-Count", "many"),
			want: "HEAD /orders: the service reported no count"},
		"negative total": {route: respond(http.StatusOK, "", "X-Total-Count", "-1"),
			want: "HEAD /orders: the service reported no count"},
		"not found": {route: respond(http.StatusNotFound, ""), class: ErrNotFound, want: "HEAD /orders: 404 Not Found"},
		"forbidden": {route: respond(http.StatusForbidden, ""), class: ErrAuth, want: "HEAD /orders: 403 Forbidden"},
	} {
		t.Run(name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/orders", tt.route)
			total, err := fake.client(t).Count(context.Background(), "/orders", nil, "")
			if err == nil || total != 0 || !strings.Contains(err.Error(), tt.want) ||
				(tt.class != nil && !errors.Is(err, tt.class)) {
				t.Fatalf("total %d, error %v; want %q", total, err, tt.want)
			}
		})
	}
}

// TestCountRefusesAQueryInItsPath checks that a count sends nothing for a
// path with a query, which a filter would replace instead of narrow, and
// leaves the query out of its message.
func TestCountRefusesAQueryInItsPath(t *testing.T) {
	fake := newFakeEmporix(t)
	_, err := fake.client(t).Count(context.Background(), "/orders?token="+testSecret, nil, "status:CREATED")
	if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "HEAD /orders: pass a collection's query "+
		"separately") || strings.Contains(err.Error(), testSecret) {
		t.Errorf("error = %v; want invalid input naming HEAD /orders", err)
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
}

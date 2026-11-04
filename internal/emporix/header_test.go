package emporix

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
)

const siteHeader = "emporix-site"

// withHeader returns the copy of c that reads with the header name, value
// (WithHeader).
func withHeader(t *testing.T, c *Client, name, value string) *Client {
	t.Helper()
	sending, err := c.WithHeader(name, value)
	if err != nil {
		t.Fatal(err)
	}
	return sending
}

// TestWithHeaderSendsItWithEveryRead checks that every read of the copy
// sends the header, every page of a collection and a search sent with a
// body too, besides the headers of the read's own and of every request, and
// that the client it was made from sends none.
func TestWithHeaderSendsItWithEveryRead(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, 0, ""))
	fake.route("/items/search", pagedRoute(1, 0, "1"))
	fake.route("/items/i1", respond(http.StatusOK, `{"id":"i1"}`))
	client := pagedClient(t, fake)
	sending := withHeader(t, client, siteHeader, "AcmeStore_NL")
	ctx := context.Background()
	if _, err := sending.GetAll(ctx, "/items", nil, ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := sending.Get(ctx, "/items/i1", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := withBody(t, sending, http.MethodPost, searchBody).GetAll(ctx, "/items/search", nil,
		ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if len(fake.requests) != 5 {
		t.Fatalf("%d requests; want 3 pages, one item and one page of the search", len(fake.requests))
	}
	for i, req := range fake.requests {
		if site := req.Header.Values(siteHeader); !slices.Equal(site, []string{"AcmeStore_NL"}) ||
			req.Header.Get("Accept-Language") != "*" || req.Header.Get("Authorization") != "Bearer "+testToken {
			t.Fatalf("request %d: %s %s with headers %v", i+1, req.Method, req.URL.RequestURI(), req.Header)
		}
	}
	if paged := fake.requests[0].Header.Get("X-Total-Count"); paged != "true" {
		t.Fatalf("X-Total-Count %q; want the total asked for besides the header", paged)
	}
	if typed := fake.requests[4].Header.Get("Content-Type"); typed != JSONContentType {
		t.Fatalf("Content-Type %q; want the search's besides the header", typed)
	}
	if _, err := client.Get(ctx, "/items/i1", nil); err != nil {
		t.Fatal(err)
	}
	if last := fake.requests[len(fake.requests)-1]; last.Header.Values(siteHeader) != nil {
		t.Fatalf("the client it was made from sent %s %v", siteHeader, last.Header.Values(siteHeader))
	}
}

// TestWithHeaderKeepsTheHeadersOfEarlierCopies checks that a copy of a copy
// sends both headers, a second value of one header replaces the first, and
// a header of a read's own replaces one of WithHeader named in another case.
func TestWithHeaderKeepsTheHeadersOfEarlierCopies(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(1, 0, "1"))
	first := withHeader(t, pagedClient(t, fake), siteHeader, "main")
	both := withHeader(t, withHeader(t, first, "x-total-count", "false"), "EMPORIX-SITE", "AcmeStore_NL")
	if _, err := both.GetAll(context.Background(), "/items", nil, ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := first.GetAll(context.Background(), "/items", nil, ListOptions{}); err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"AcmeStore_NL", "main"} {
		req := fake.requests[i]
		if site := req.Header.Values(siteHeader); !slices.Equal(site, []string{want}) ||
			!slices.Equal(req.Header.Values("X-Total-Count"), []string{"true"}) {
			t.Fatalf("request %d: headers %v; want %s %s and the read's own X-Total-Count", i+1, req.Header,
				siteHeader, want)
		}
	}
}

// TestHeadersApplyEachExtraInOrder checks that the headers of a request
// take each set of extra headers in its order: one of a later set replaces
// one of an earlier set or of every request, whatever the case of its name,
// one without values removes it, and the keys of one set that name the same
// header apply in their sorted order, so that a request is the same on every
// run.
func TestHeadersApplyEachExtraInOrder(t *testing.T) {
	client := newFakeEmporix(t).client(t)
	got := client.headers(JSONContentType,
		http.Header{"accept": {"text/csv"}, "x-site": {"main"}, "X-SITE": {"AcmeStore_DE"}},
		http.Header{"X-Site": {"AcmeStore_NL"}, "accept-language": nil})
	want := http.Header{"Accept": {"text/csv"}, "Content-Type": {JSONContentType},
		"User-Agent": {"emporix-cli/test"}, "X-Site": {"AcmeStore_NL"}}
	if !maps.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("headers %v; want %v", got, want)
	}
	for range 20 {
		if one := client.headers("", http.Header{"x-site": {"main"}, "X-SITE": {"AcmeStore_DE"}}); !slices.Equal(
			one.Values("X-Site"), []string{"main"}) {
			t.Fatalf("X-Site %q; want the value of the last key in sorted order, main", one.Values("X-Site"))
		}
	}
}

// TestWithHeaderRefusesAValueThatCannotBeSent checks that a value holding a
// control character is ErrInvalidInput, named by its header but not echoed,
// before any token is requested.
func TestWithHeaderRefusesAValueThatCannotBeSent(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, value := range []string{"main\r\nX-Injected: 1", "main\x7f"} {
		_, err := fake.client(t).WithHeader(siteHeader, value)
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "the header "+siteHeader) ||
			strings.Contains(err.Error(), "main") {
			t.Errorf("%q: error %v; want ErrInvalidInput naming the header, not the value", value, err)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("%d token requests and %d reads; want none", fake.tokenCalls, len(fake.requests))
	}
}

// TestWithHeaderLeavesWritesAlone checks that a write prepared by a copy
// sends only the headers of its Write.
func TestWithHeaderLeavesWritesAlone(t *testing.T) {
	fake := newFakeEmporix(t)
	sending := withHeader(t, fake.client(t), siteHeader, "main")
	req, err := sending.Prepare(&Write{Method: http.MethodDelete, Path: "/items/i1"})
	if err != nil {
		t.Fatal(err)
	}
	if site := req.Header().Values(siteHeader); site != nil {
		t.Fatalf("a write sends %s %v; want none", siteHeader, site)
	}
}

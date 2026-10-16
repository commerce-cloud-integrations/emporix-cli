package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
)

const searchBody = `{"q":"code:(A1,A2)"}`

// withBody returns the copy of c that reads with method and body (WithBody).
func withBody(t *testing.T, c *Client, method, body string) *Client {
	t.Helper()
	sending, err := c.WithBody(method, jsonBody(t, body))
	if err != nil {
		t.Fatal(err)
	}
	return sending
}

// TestWithBodyPagesASearch checks that a search reads every page as GetAll
// reads a collection: each page sends the method and body as JSON, with
// pageNumber and pageSize in the query and the total asked for.
func TestWithBodyPagesASearch(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items/search", pagedRoute(5, 0, "5"))
	client := pagedClient(t, fake)
	body, err := withBody(t, client, http.MethodPost, searchBody).GetAll(context.Background(), "/items/search",
		map[string][]string{"showUnpublished": {"true"}}, ListOptions{Sort: "code"})
	if err != nil || !strings.HasPrefix(string(body), `[{"id":"i0"},`) || len(fake.requests) != 3 {
		t.Fatalf("body %s, error %v, %d requests; want 5 items in 3 pages", body, err, len(fake.requests))
	}
	for i, req := range fake.requests {
		query := req.URL.Query()
		if req.Method != http.MethodPost || fake.bodies[i] != searchBody ||
			req.Header.Get("Content-Type") != "application/json" || req.Header.Get("X-Total-Count") != "true" ||
			req.Header.Get("Accept-Language") != "*" || query.Get("pageSize") != "2" ||
			query.Get("showUnpublished") != "true" || query.Get("sort") != "code" {
			t.Fatalf("request %d: %s %s with body %q and headers %v", i+1, req.Method, req.URL.RequestURI(),
				fake.bodies[i], req.Header)
		}
	}
	// The client it was made from still reads with GET, without a body.
	fake.route("/items", pagedRoute(1, 0, "1"))
	if _, err := client.GetAll(context.Background(), "/items", nil, ListOptions{}); err != nil {
		t.Fatal(err)
	}
	last, sent := fake.requests[len(fake.requests)-1], fake.bodies[len(fake.bodies)-1]
	if last.Method != http.MethodGet || sent != "" || last.Header.Get("Content-Type") != "" {
		t.Fatalf("after a search, the client sent %s with body %q", last.Method, sent)
	}
}

// TestWithBodyChecksPagesAsCollections checks that the paging checks of a
// collection hold for a search, naming its method.
func TestWithBodyChecksPagesAsCollections(t *testing.T) {
	for name, tt := range map[string]struct {
		route func(http.ResponseWriter, *http.Request)
		read  func(*Client) (json.RawMessage, error)
		want  string
	}{
		"paging ignored": {route: pagedRoute(3, -1, ""), want: "POST /items/search: page 2 repeats an item of page 1",
			read: func(c *Client) (json.RawMessage, error) {
				return c.GetAll(context.Background(), "/items/search", nil, ListOptions{})
			}},
		"total differs": {route: pagedRoute(3, 0, "4"), want: "POST /items/search: read 3 items but the service reports 4",
			read: func(c *Client) (json.RawMessage, error) {
				return c.GetAll(context.Background(), "/items/search", nil, ListOptions{})
			}},
		"unpaged not an array": {route: respond(http.StatusOK, `{}`), want: "POST /items/search: expected a JSON array",
			read: func(c *Client) (json.RawMessage, error) {
				return c.GetUnpaged(context.Background(), "/items/search", nil, ListOptions{})
			}},
		"rejected body": {route: respond(http.StatusBadRequest, `{"message":"q must not be null"}`),
			want: "POST /items/search: 400 Bad Request: q must not be null",
			read: func(c *Client) (json.RawMessage, error) {
				return c.GetAll(context.Background(), "/items/search", nil, ListOptions{})
			}},
	} {
		t.Run(name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items/search", tt.route)
			if _, err := tt.read(withBody(t, pagedClient(t, fake), http.MethodPost, searchBody)); err == nil ||
				!strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

// TestWithBodyReadsOneResponse checks the single form, for a calculation
// that Emporix serves as PUT and that answers one object; its path may hold
// a query, as a GET's; a failure is classified as for a GET, and an empty
// response is not found, naming the method.
func TestWithBodyReadsOneResponse(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/tax/acme1/calculate", respond(http.StatusOK, `{"grossPrice":12.1}`))
	calculate := func() (json.RawMessage, error) {
		return withBody(t, fake.client(t), http.MethodPut, `[1, 2]`).Get(context.Background(),
			"/tax/{tenant}/calculate?siteCode=main", nil)
	}
	body, err := calculate()
	if err != nil || string(body) != `{"grossPrice":12.1}` {
		t.Fatalf("body %s, error %v", body, err)
	}
	if req := fake.requests[0]; req.Method != http.MethodPut || fake.bodies[0] != `[1, 2]` ||
		req.URL.RawQuery != "siteCode=main" {
		t.Fatalf("sent %s %s with body %q; want PUT with the body as given", req.Method, req.URL.RequestURI(),
			fake.bodies[0])
	}
	for status, tt := range map[int]struct {
		body string
		want string
	}{
		http.StatusNotFound: {body: `{"message":"no such site"}`, want: "PUT /tax/acme1/calculate: 404 Not Found"},
		http.StatusOK:       {body: "", want: "PUT /tax/acme1/calculate: empty response"},
	} {
		fake.route("/tax/acme1/calculate", respond(status, tt.body))
		if _, err := calculate(); !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("error = %v; want not found %q", err, tt.want)
		}
	}
}

// TestWithBodyRetriesWithTheBody checks that a throttled search is sent
// again, body included: it changes nothing, so a retry is safe.
func TestWithBodyRetriesWithTheBody(t *testing.T) {
	fake := newFakeEmporix(t)
	calls := 0
	fake.route("/items/search", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			respond(http.StatusServiceUnavailable, `{"message":"busy"}`, "Retry-After", "0")(w, r)
			return
		}
		respond(http.StatusOK, `{"id":"i1"}`)(w, r)
	})
	search := withBody(t, fake.client(t), http.MethodPost, searchBody)
	if _, err := search.Get(context.Background(), "/items/search", nil); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(fake.bodies, []string{searchBody, searchBody}) {
		t.Fatalf("bodies %q; want the search sent twice", fake.bodies)
	}
}

// TestWithBodyRefusesWhatIsNotARead checks that a client with another method
// or without a body is never made, and that a collection read with a
// body sends nothing for a path with a query or for a q option, which its
// body replaces.
func TestWithBodyRefusesWhatIsNotARead(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	refused := func(method string, body Body) error {
		sending, err := client.WithBody(method, body)
		if sending != nil {
			t.Errorf("%s: got a client", method)
		}
		return err
	}
	list := func(path string) error {
		_, err := withBody(t, client, http.MethodPost, searchBody).GetAll(context.Background(), path, nil,
			ListOptions{Query: "code:A1"})
		return err
	}
	for _, tt := range []struct {
		err  error
		want string
	}{
		{refused(http.MethodDelete, jsonBody(t, `{}`)), `"DELETE": a read with a body is sent as POST or PUT`},
		{refused(http.MethodGet, jsonBody(t, `{}`)), `"GET": a read with a body is sent as POST or PUT`},
		{refused(http.MethodPost, Body{}), "POST: a read with a body sends one"},
		{list("/x"), "POST /x: a read with a body filters with its body, not with a q parameter"},
		{list("/x?token=" + testSecret), "POST /x: pass a collection's query separately"},
	} {
		if !errors.Is(tt.err, ErrInvalidInput) || !strings.Contains(tt.err.Error(), tt.want) ||
			strings.Contains(tt.err.Error(), testSecret) {
			t.Errorf("error = %v; want invalid input %q", tt.err, tt.want)
		}
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
}

// TestWithBodyRefusesPathsThatMayAddressAnotherOperation checks that a read
// with a body sends nothing, not even a token request, to a path with a
// segment that a service or a proxy might decode into another path, while a
// GET may read it; the message names the path as given, without its query.
func TestWithBodyRefusesPathsThatMayAddressAnotherOperation(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	search := withBody(t, client, http.MethodPost, searchBody)
	for _, path := range []string{
		"/search/{tenant}/search/indexes%2Fimport", "/search/{tenant}/search/indexes%2fimport",
		"/search/{tenant}/search/indexes%252Fimport", "/search/{tenant}/search/indexes%5Cimport",
		"/search/{tenant}/search/%2E%2E/import", "/search/{tenant}/search/./x", "/search/{tenant}//x",
		"/search/{tenant}/search/", "/search/{tenant}/search/%25?token=" + testSecret,
	} {
		_, getErr := search.Get(context.Background(), path, nil)
		_, listErr := search.GetAll(context.Background(), withoutQuery(path), nil, ListOptions{})
		for _, err := range []error{getErr, listErr} {
			if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "POST "+withoutQuery(path)+
				": a path segment of a read sent as POST is empty, . or .., or holds /, \\ or %") ||
				strings.Contains(err.Error(), testSecret) {
				t.Errorf("%s: error = %v; want invalid input", path, err)
			}
		}
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
	// Spaces, colons and other escaped characters keep their segment; a GET is not checked.
	fake.route("/search/acme1/search/a b:c", respond(http.StatusOK, `{"id":"i1"}`))
	fake.route("/x/a/b", respond(http.StatusOK, `{"id":"i1"}`))
	if _, err := search.Get(context.Background(), "/search/{tenant}/search/a%20b:c", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get(context.Background(), "/x/a%2Fb", nil); err != nil {
		t.Fatal(err)
	}
}

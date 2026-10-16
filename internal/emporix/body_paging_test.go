package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// bodyPagedRoute serves count items ({"id":"i<n>"}) by the pageNumber and
// pageSize of the request body, ignoring the query, as the Fee Service's
// searches by product page; total, when non-empty, is sent as X-Total-Count.
func bodyPagedRoute(count int, total string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var paging struct {
			PageNumber int `json:"pageNumber"`
			PageSize   int `json:"pageSize"`
		}
		if json.NewDecoder(r.Body).Decode(&paging) != nil || paging.PageNumber < 1 || paging.PageSize < 1 {
			respond(http.StatusBadRequest, `{"message":"must be greater than or equal to 1"}`)(w, r)
			return
		}
		items := []string{}
		for i := (paging.PageNumber - 1) * paging.PageSize; i < min(paging.PageNumber*paging.PageSize, count); i++ {
			items = append(items, fmt.Sprintf(`{"id":"i%d"}`, i))
		}
		if total != "" {
			w.Header().Set("X-Total-Count", total)
		}
		_, _ = w.Write([]byte("[" + strings.Join(items, ",") + "]"))
	}
}

// TestGetAllBodyPagedPagesInTheBody checks that every page sends the body
// with its page and size set, replacing those it held and keeping the other
// members as given, and the query as given, without paging; one token reads
// every page.
func TestGetAllBodyPagedPagesInTheBody(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items/search", bodyPagedRoute(5, ""))
	search := withBody(t, pagedClient(t, fake), http.MethodPost, `{"productId":"100200", "pageNumber":7,"a":"<b&c>"}`)
	body, err := search.GetAllBodyPaged(context.Background(), "/items/search",
		map[string][]string{"siteFallback": {"true"}}, ListOptions{})
	var items []json.RawMessage
	// Without a total, the empty page 4 ends the search.
	if err != nil || json.Unmarshal(body, &items) != nil || len(items) != 5 || len(fake.requests) != 4 {
		t.Fatalf("body %s, error %v, %d requests; want 5 items in 4 pages", body, err, len(fake.requests))
	}
	for i, req := range fake.requests {
		want := fmt.Sprintf(`{"a":"<b&c>","pageNumber":%d,"pageSize":2,"productId":"100200"}`, i+1)
		if req.Method != http.MethodPost || fake.bodies[i] != want || req.URL.RawQuery != "siteFallback=true" ||
			req.Header.Get("X-Total-Count") != "true" || req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("request %d: %s %s with body %q and headers %v; want body %q", i+1, req.Method,
				req.URL.RequestURI(), fake.bodies[i], req.Header, want)
		}
	}
	if fake.tokenCalls != 1 {
		t.Fatalf("%d token requests; want one for every page", fake.tokenCalls)
	}
}

// TestGetAllBodyPagedChecksPagesAsCollections checks that a search paged in
// its body is checked as GetAll checks a collection, and stops at its limit.
func TestGetAllBodyPagedChecksPagesAsCollections(t *testing.T) {
	for name, tt := range map[string]struct {
		route    func(http.ResponseWriter, *http.Request)
		opts     ListOptions
		want     string // the error, or the items when empty
		requests int
	}{
		"paging ignored": {route: pagedRoute(3, -1, ""), requests: 2,
			want: "POST /items/search: page 2 repeats an item of page 1"},
		"total differs": {route: bodyPagedRoute(3, "4"), requests: 3,
			want: "POST /items/search: read 3 items but the service reports 4"},
		"total reached": {route: bodyPagedRoute(3, "3"), requests: 2},
		"limit":         {route: bodyPagedRoute(3, ""), opts: ListOptions{Limit: 1}, requests: 1},
	} {
		t.Run(name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items/search", tt.route)
			_, err := withBody(t, pagedClient(t, fake), http.MethodPost, `{}`).GetAllBodyPaged(context.Background(),
				"/items/search", nil, tt.opts)
			if (tt.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), tt.want)) ||
				len(fake.requests) != tt.requests {
				t.Fatalf("error = %v after %d requests; want %q after %d", err, len(fake.requests), tt.want, tt.requests)
			}
		})
	}
}

// TestGetAllBodyPagedRefusesABodyThatIsNotAnObject checks that a search
// paged in its body needs a JSON object to page: without one, nothing is
// sent, not even a token request.
func TestGetAllBodyPagedRefusesABodyThatIsNotAnObject(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	for name, tt := range map[string]struct {
		client *Client
		want   string
	}{
		"no body": {client: client, want: "GET /items/search: a search paged in its body needs a JSON object"},
		"array":   {client: withBody(t, client, http.MethodPost, `[{"pageNumber":1}]`), want: "POST /items/search: a"},
		"null":    {client: withBody(t, client, http.MethodPut, `null`), want: "PUT /items/search: a search paged"},
	} {
		if _, err := tt.client.GetAllBodyPaged(context.Background(), "/items/search", nil, ListOptions{}); !errors.Is(err,
			ErrInvalidInput) || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: error = %v; want invalid input %q", name, err, tt.want)
		}
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
}

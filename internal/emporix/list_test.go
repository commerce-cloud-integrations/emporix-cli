package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// itemIDs returns the ids of a JSON array of {"id": …} items.
func itemIDs(t *testing.T, body json.RawMessage) []string {
	t.Helper()
	var items []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

func TestGetAllWithLimitStopsEarly(t *testing.T) {
	tests := []struct {
		name         string
		count, limit int
		want         []string
		pageSizes    []string // pageSize of each request
	}{
		// Pages of 2 (the policy size) until the limit is reached; the total is not checked.
		{name: "limit within the collection", count: 7, limit: 3, want: []string{"i0", "i1", "i2"},
			pageSizes: []string{"2", "2"}},
		// A limit below the page size shrinks the page.
		{name: "limit below the page size", count: 7, limit: 1, want: []string{"i0"}, pageSizes: []string{"1"}},
		// A limit beyond the collection reads it whole, and the total is still checked.
		{name: "limit beyond the collection", count: 3, limit: 10, want: []string{"i0", "i1", "i2"},
			pageSizes: []string{"2", "2"}},
		{name: "no limit", count: 3, want: []string{"i0", "i1", "i2"}, pageSizes: []string{"2", "2"}},
		{name: "negative limit", count: 3, limit: -3, want: []string{"i0", "i1", "i2"}, pageSizes: []string{"2", "2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(tt.count, 0, strconv.Itoa(tt.count)))
			body, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, ListOptions{Limit: tt.limit})
			if err != nil || !slices.Equal(itemIDs(t, body), tt.want) {
				t.Fatalf("items %s, error %v; want %v", body, err, tt.want)
			}
			var sizes []string
			for _, request := range fake.requests {
				sizes = append(sizes, request.URL.Query().Get("pageSize"))
			}
			if !slices.Equal(sizes, tt.pageSizes) {
				t.Fatalf("page sizes %q; want %q", sizes, tt.pageSizes)
			}
		})
	}
}

func TestGetAllWithLimitStillChecksAWholeCollection(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, 0, "4")) // the collection ends before its total
	client := pagedClient(t, fake)
	if _, err := client.GetAll(context.Background(), "/items", nil, ListOptions{Limit: 10}); err == nil {
		t.Fatal("want the count mismatch reported")
	}
	// A repeated item is reported even before the limit is reached.
	fake.route("/items", pagedRoute(3, -1, ""))
	if _, err := client.GetAll(context.Background(), "/items", nil, ListOptions{Limit: 5}); err == nil {
		t.Fatal("want the repeated page reported")
	}
}

func TestWholeCollectionReadsAreCutAtTheLimit(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, -1, "3"))
	client, limited := pagedClient(t, fake), ListOptions{Limit: 2}
	body, err := client.GetUnpaged(context.Background(), "/items", nil, limited)
	if err != nil || !slices.Equal(itemIDs(t, body), []string{"i0", "i1"}) {
		t.Fatalf("unpaged: items %s, error %v", body, err)
	}
	fake.route("/items", pagedRoute(3, 0, "3"))
	body, err = client.GetAllInOnePage(context.Background(), "/items", nil, 10, limited)
	if err != nil || !slices.Equal(itemIDs(t, body), []string{"i0", "i1"}) {
		t.Fatalf("one page: items %s, error %v", body, err)
	}
	// A short collection is returned whole, as with a negative limit.
	fake.route("/items", pagedRoute(3, -1, "3"))
	for _, limit := range []int{5, -1} {
		body, err = client.GetUnpaged(context.Background(), "/items", nil, ListOptions{Limit: limit})
		if err != nil || len(itemIDs(t, body)) != 3 {
			t.Fatalf("limit %d: items %s, error %v", limit, body, err)
		}
	}
}

func TestGetAllWithLimitReadsOnAfterAShortPage(t *testing.T) {
	// A service that caps pages at one item: the read cannot stop early, so it
	// reads to the end, checks the total and then cuts.
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(4, 1, "4"))
	body, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, ListOptions{Limit: 3})
	if err != nil || !slices.Equal(itemIDs(t, body), []string{"i0", "i1", "i2"}) || len(fake.requests) != 4 {
		t.Fatalf("items %s, error %v, requests %d; want 3 of 4 items after reading all 4 pages",
			body, err, len(fake.requests))
	}
	// A capped service that still counts offsets by the requested size leaves
	// gaps: the total check reports them instead of returning i0, i2.
	gaps := newFakeEmporix(t)
	gaps.route("/items", func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		size, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		w.Header().Set("X-Total-Count", "4")
		if first := (page - 1) * size; first < 4 {
			_, _ = fmt.Fprintf(w, `[{"id":"i%d"}]`, first)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	if _, err := pagedClient(t, gaps).GetAll(context.Background(), "/items", nil, ListOptions{Limit: 3}); err == nil ||
		!strings.Contains(err.Error(), "read 2 items but the service reports 4") {
		t.Fatalf("error = %v; want the gap reported", err)
	}
}

func TestGetAllWithLimitStopsInAShortPageAfterFullOnes(t *testing.T) {
	// Items of a page are contiguous, so after full pages the limit can be
	// reached in a short page: A, B, C are the first three of four.
	fake := newFakeEmporix(t)
	fake.route("/items", pageSequence([]string{`[{"id":"A"},{"id":"B"}]`, `[{"id":"C"}]`}, "4", "4"))
	body, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, ListOptions{Limit: 3})
	if err != nil || !slices.Equal(itemIDs(t, body), []string{"A", "B", "C"}) || len(fake.requests) != 2 {
		t.Fatalf("items %s, error %v, requests %d; want A B C from two pages", body, err, len(fake.requests))
	}
}

func TestListOptionsReachEveryCollectionRead(t *testing.T) {
	opts := ListOptions{Query: "code:A1", Sort: "name:desc"}
	base := url.Values{"sort": {"code"}, "siteCode": {"main"}}
	for name, read := range map[string]func(*Client) (json.RawMessage, error){
		"GetAll": func(c *Client) (json.RawMessage, error) { return c.GetAll(context.Background(), "/items", base, opts) },
		"GetAllWithoutTotal": func(c *Client) (json.RawMessage, error) {
			return c.GetAllWithoutTotal(context.Background(), "/items", base, opts)
		},
		"GetUnpaged": func(c *Client) (json.RawMessage, error) {
			return c.GetUnpaged(context.Background(), "/items", base, opts)
		},
		"GetAllInOnePage": func(c *Client) (json.RawMessage, error) {
			return c.GetAllInOnePage(context.Background(), "/items", base, 10, opts)
		},
		"GetAllOnFirstPage": func(c *Client) (json.RawMessage, error) {
			return c.GetAllOnFirstPage(context.Background(), "/items", base, 10, opts)
		},
	} {
		t.Run(name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(3, 0, "3"))
			if _, err := read(pagedClient(t, fake)); err != nil {
				t.Fatal(err)
			}
			// Every request carries the filter; the user's sort replaces the read's own.
			for _, request := range fake.requests {
				query := request.URL.Query()
				if query.Get("q") != "code:A1" || query.Get("sort") != "name:desc" || query.Get("siteCode") != "main" {
					t.Fatalf("query %q; want q, the replaced sort and siteCode", request.URL.RawQuery)
				}
			}
		})
	}
	if base.Get("sort") != "code" || base.Has("q") {
		t.Fatalf("base query changed: %q", base)
	}
}

func TestListOptionsApply(t *testing.T) {
	base := url.Values{"sort": {"code"}, "siteCode": {"main"}}
	// Empty options change nothing; a user's filter joins the read's own (both must match).
	if got := (ListOptions{}).apply(base); got.Encode() != base.Encode() {
		t.Fatalf("empty options changed the query to %q", got)
	}
	opts := ListOptions{Query: "code:A1"}
	if got := opts.apply(url.Values{"q": {"legalEntityId:L1"}}).Get("q"); got != "legalEntityId:L1 code:A1" {
		t.Fatalf("q = %q; want the read's own filter and the user's", got)
	}
}

func TestCollectionReadsTakeTheirQuerySeparately(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, 0, "3"))
	client := pagedClient(t, fake)
	// An inline q would be replaced by a filter instead of narrowed, so a query
	// in the path is refused before anything is sent, and the error leaves it out.
	path := "/items?q=legalEntityId:L1&token=" + testSecret
	for name, read := range map[string]func() (json.RawMessage, error){
		"GetAll": func() (json.RawMessage, error) { return client.GetAll(context.Background(), path, nil, ListOptions{}) },
		"GetAllWithoutTotal": func() (json.RawMessage, error) {
			return client.GetAllWithoutTotal(context.Background(), path, nil, ListOptions{})
		},
		"GetUnpaged": func() (json.RawMessage, error) {
			return client.GetUnpaged(context.Background(), path, nil, ListOptions{})
		},
		"GetAllInOnePage": func() (json.RawMessage, error) {
			return client.GetAllInOnePage(context.Background(), path, nil, 10, ListOptions{})
		},
	} {
		if _, err := read(); !errors.Is(err, ErrInvalidInput) || strings.Contains(err.Error(), testSecret) ||
			!strings.Contains(err.Error(), "GET /items: pass a collection's query separately") {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
}

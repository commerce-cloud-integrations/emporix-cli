package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// pagedRoute serves count items ({"id":"i<n>"}) by pageNumber. The page size
// is the requested pageSize capped at maxSize (0: no cap, -1: paging ignored
// and every item returned). total, when non-empty, is sent as X-Total-Count.
func pagedRoute(count, maxSize int, total string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		page, _ := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		if maxSize > 0 {
			size = min(size, maxSize)
		}
		first, last := (page-1)*size, page*size
		if maxSize < 0 {
			first, last = 0, count
		}
		items := []string{}
		for i := first; i < min(last, count); i++ {
			items = append(items, fmt.Sprintf(`{"id":"i%d"}`, i))
		}
		if total != "" {
			w.Header().Set("X-Total-Count", total)
		}
		_, _ = w.Write([]byte("[" + strings.Join(items, ",") + "]"))
	}
}

// pageSequence serves scripted pages by pageNumber (1-based) with optional
// per-page X-Total-Count values; pages beyond the script are empty.
func pageSequence(pages []string, totals ...string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		if page-1 < len(totals) && totals[page-1] != "" {
			w.Header().Set("X-Total-Count", totals[page-1])
		}
		body := "[]"
		if page-1 < len(pages) {
			body = pages[page-1]
		}
		_, _ = w.Write([]byte(body))
	}
}

func pagedClient(t *testing.T, fake *fakeEmporix) *Client {
	t.Helper()
	client := fake.client(t)
	client.paging = pagingPolicy{size: 2, keptPages: 5, streamedPages: 8}
	return client
}

func TestGetAllReadsEveryPage(t *testing.T) {
	tests := []struct {
		name           string
		count, maxSize int
		total          string
		requests       int
	}{
		{name: "total reached on a short page", count: 5, total: "5", requests: 3},
		{name: "total reached on a full page", count: 4, total: "4", requests: 2},
		{name: "without a total, an empty page ends", count: 5, requests: 4},
		{name: "server caps the page size", count: 3, maxSize: 1, requests: 4},
		{name: "empty collection", count: 0, total: "0", requests: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkEveryPage(t, tt.count, tt.maxSize, tt.total, tt.requests)
		})
	}
}

// checkEveryPage checks that GetAll reads a collection of count items, served
// with maxSize as the largest page (0 for none) and total, in requests requests
// with one token, and that the last request asks for the next page.
func checkEveryPage(t *testing.T, count, maxSize int, total string, requests int) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(count, maxSize, total))
	body, err := pagedClient(t, fake).GetAll(context.Background(), "/items", map[string][]string{"q": {"x"}},
		ListOptions{})
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	var items []map[string]string
	if err := json.Unmarshal(body, &items); err != nil || len(items) != count {
		t.Fatalf("body = %s (%v), want %d items", body, err, count)
	}
	if len(fake.requests) != requests || fake.tokenCalls != 1 {
		t.Fatalf("%d requests and %d tokens, want %d requests with one token",
			len(fake.requests), fake.tokenCalls, requests)
	}
	last := fake.requests[len(fake.requests)-1]
	query := last.URL.Query()
	if query.Get("pageNumber") != strconv.Itoa(requests) || query.Get("pageSize") != "2" ||
		query.Get("q") != "x" || last.Header.Get("X-Total-Count") != "true" {
		t.Fatalf("last request %s with X-Total-Count %q", last.URL.RequestURI(), last.Header.Get("X-Total-Count"))
	}
}

func TestGetAllFailures(t *testing.T) {
	tests := []struct {
		name  string
		route func(http.ResponseWriter, *http.Request)
		want  string
	}{
		{name: "total larger than the collection", route: pagedRoute(3, 0, "7"),
			want: "read 3 items but the service reports 7"},
		{name: "total smaller than a page", route: pagedRoute(3, 0, "1"), want: "read 2 items but the service reports 1"},
		{name: "paging ignored", route: pagedRoute(3, -1, ""), want: "page 2 repeats an item of page 1"},
		{name: "later page repeated", route: pageSequence([]string{`["A","B"]`, `["C","D"]`, `["C","D"]`}, "6"),
			want: "page 3 repeats an item of page 2"},
		{name: "item repeated on its page", route: pageSequence([]string{`["A","A"]`}, "2"),
			want: "a response lists an item twice"},
		{name: "repeat serialized differently", route: pageSequence([]string{`[{"id":"A","n":{"de":1,"en":2}},"B"]`,
			`[{"n":{"en":2,"de":1},"id":"A"},"C"]`}, "4"), want: "page 2 repeats an item of page 1"},
		{name: "same id with changed fields", route: pageSequence([]string{`[{"id":"A","n":1},"B"]`,
			`[{"id":"A","n":2},"C"]`}, "4"), want: "page 2 repeats an item of page 1"},
		{name: "total changes", route: pageSequence([]string{`["A","B"]`, `["E","F"]`}, "6", "4"),
			want: "the total changed from 6 to 4"},
		{name: "total disappears", route: pageSequence([]string{`["A","B"]`, `["C","D"]`}, "6"),
			want: "read 4 items but the service reports 6"},
		{name: "null collection", route: pageSequence([]string{`null`}), want: "page 1: expected a JSON array"},
		{name: "null later page", route: pageSequence([]string{`["A","B"]`, `null`}), want: "page 2: expected a JSON array"},
		// Unlike a read of one item, an empty page is malformed, not a missing resource.
		{name: "empty later page", route: pageSequence([]string{`["A","B"]`, ``}), want: "page 2: expected a JSON array"},
		{name: "no content", route: respond(http.StatusNoContent, ``), want: "page 1: expected a JSON array"},
		{name: "endless collection", route: pagedRoute(100, 0, ""), want: "more than 5 pages"},
		{name: "not an array", route: respond(http.StatusOK, `{"items":[]}`), want: "page 1: expected a JSON array"},
		{name: "page error", route: respond(http.StatusForbidden, `{"message":"no"}`), want: "403 Forbidden: no"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", tt.route)
			if _, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, ListOptions{}); err == nil ||
				!strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestGetAllWithoutTotal(t *testing.T) {
	fake := newFakeEmporix(t)
	// Like the reindex-jobs endpoint: asking for the total fails the request.
	paged := pagedRoute(3, 0, "")
	fake.route("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Total-Count") != "" {
			respond(http.StatusInternalServerError, `{"message":"database error"}`)(w, r)
			return
		}
		paged(w, r)
	})
	client := pagedClient(t, fake)
	body, err := client.GetAllWithoutTotal(context.Background(), "/items", nil, ListOptions{})
	if err != nil || string(body) != `[{"id":"i0"},{"id":"i1"},{"id":"i2"}]` || len(fake.requests) != 3 {
		t.Fatalf("body %s, error %v, %d requests; want 3 items in 3 requests", body, err, len(fake.requests))
	}
	fake.route("/items", pagedRoute(3, -1, ""))
	if _, err := client.GetAllWithoutTotal(context.Background(), "/items", nil, ListOptions{}); err == nil ||
		!strings.Contains(err.Error(), "page 2 repeats an item of page 1") {
		t.Fatalf("error = %v; ignored paging must still fail", err)
	}
	// A total sent unasked is still checked.
	fake.route("/items", pagedRoute(3, 0, "7"))
	if _, err := client.GetAllWithoutTotal(context.Background(), "/items", nil, ListOptions{}); err == nil ||
		!strings.Contains(err.Error(), "read 3 items but the service reports 7") {
		t.Fatalf("error = %v; an unasked total must still be checked", err)
	}
}

// TestOnePageReads checks GetAllInOnePage and GetAllOnFirstPage, which read
// one page of 10 asking for its total: a page cut short fails, and without a
// total GetAllInOnePage checks that the next page is empty, while
// GetAllOnFirstPage, whose service never lists a next page, fails.
func TestOnePageReads(t *testing.T) {
	onFirstPage := (*Client).GetAllOnFirstPage
	full := make([]string, 10) // a full page, as a service capped at 10 would send of a longer collection
	for i := range full {
		full[i] = fmt.Sprintf(`{"id":"i%d"}`, i)
	}
	fullPage := pageSequence([]string{"[" + strings.Join(full, ",") + "]"})
	tests := []struct {
		name     string
		read     func(*Client, context.Context, string, url.Values, int, ListOptions) (json.RawMessage, error)
		route    func(http.ResponseWriter, *http.Request)
		requests int
		want     string // error; empty for success with 3 items
	}{
		{name: "total reported", route: pagedRoute(3, 0, "3"), requests: 1},
		{name: "no total, empty next page", route: pagedRoute(3, 0, ""), requests: 2},
		{name: "capped page with a total", route: pagedRoute(12, 5, "12"), requests: 1,
			want: "one page of 10 returned 5 of 12 items; the service may limit the page size"},
		{name: "capped page without a total", route: pagedRoute(12, 5, ""), requests: 2,
			want: "the collection does not fit one page of 10 items"},
		{name: "total smaller than the page", route: pagedRoute(3, 0, "2"), requests: 1,
			want: "read 3 items but the service reports 2"},
		{name: "repeated item", route: pageSequence([]string{`[{"a":1,"b":2},{ "b":2, "a":1 }]`}, "2"), requests: 1,
			want: "a response lists an item twice"},
		{name: "total only on the empty next page", route: pageSequence([]string{`["A","B","C"]`}, "", "12"),
			requests: 2, want: "read 3 items but the service reports 12"},
		{name: "not an array", route: respond(http.StatusOK, `{"items":[]}`), requests: 1,
			want: "page 1: expected a JSON array"},
		{name: "first page only, total reported", read: onFirstPage, route: pagedRoute(3, 0, "3"), requests: 1},
		{name: "first page only, no total", read: onFirstPage, route: fullPage, requests: 1,
			want: "the service reported no total, so one page of 10 may not hold every item"},
		{name: "first page only, capped", read: onFirstPage, route: pagedRoute(12, 5, "12"), requests: 1,
			want: "one page of 10 returned 5 of 12 items"},
		{name: "first page only, total smaller", read: onFirstPage, route: pagedRoute(3, 0, "2"), requests: 1,
			want: "read 3 items but the service reports 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", tt.route)
			read := tt.read
			if read == nil {
				read = (*Client).GetAllInOnePage
			}
			body, err := read(pagedClient(t, fake), context.Background(), "/items", url.Values{"q": {"x"}}, 10,
				ListOptions{})
			if len(fake.requests) != tt.requests {
				t.Fatalf("%d requests, want %d", len(fake.requests), tt.requests)
			}
			query := fake.requests[0].URL.Query()
			if query.Get("pageNumber") != "1" || query.Get("pageSize") != "10" || query.Get("q") != "x" ||
				fake.requests[0].Header.Get("X-Total-Count") != "true" {
				t.Fatalf("first request %s; want page 1 of 10 asking for the total", fake.requests[0].URL.RawQuery)
			}
			assertThreeItemsOr(t, body, err, tt.want)
		})
	}
}

// assertThreeItemsOr checks that a read returned the items i0, i1 and i2 or,
// when want is set, failed with an error that holds it.
func assertThreeItemsOr(t *testing.T, body json.RawMessage, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil || string(body) != `[{"id":"i0"},{"id":"i1"},{"id":"i2"}]` {
			t.Fatalf("body %s, error %v", body, err)
		}
	} else if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestGetUnpaged(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, -1, ""))
	client := pagedClient(t, fake)
	body, err := client.GetUnpaged(context.Background(), "/items", nil, ListOptions{})
	if err != nil || string(body) != `[{"id":"i0"},{"id":"i1"},{"id":"i2"}]` || len(fake.requests) != 1 ||
		fake.requests[0].URL.RawQuery != "" {
		t.Fatalf("body %s, error %v, requests %d; want every item from one unpaged request", body, err, len(fake.requests))
	}
	// An empty collection body is malformed, not a missing resource as for Get.
	for _, malformed := range []string{`null`, `{"items":[]}`, `"x"`, `[1,`, ``} {
		fake.route("/items", respond(http.StatusOK, malformed))
		if _, err := client.GetUnpaged(context.Background(), "/items", nil, ListOptions{}); err == nil ||
			!strings.Contains(err.Error(), "GET /items: expected a JSON array") || errors.Is(err, ErrNotFound) {
			t.Errorf("%q: error = %v", malformed, err)
		}
	}
	fake.route("/items", respond(http.StatusOK, `[{"id":"i0","n":1},{"n":1,"id":"i0"}]`))
	if _, err := client.GetUnpaged(context.Background(), "/items", nil, ListOptions{}); err == nil ||
		!strings.Contains(err.Error(), "GET /items: a response lists an item twice") {
		t.Errorf("repeated item: error = %v", err)
	}
	fake.route("/items", respond(http.StatusNotFound, `{"message":"gone"}`))
	if _, err := client.GetUnpaged(context.Background(), "/items", nil, ListOptions{}); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

// TestCountingTotal checks that each counting reads its own header only.
func TestCountingTotal(t *testing.T) {
	for value, want := range map[string]int{"": -1, "x": -1, "-3": -1, "0": 0, "15": 15} {
		counted, hybris := http.Header{}, http.Header{}
		counted.Set("X-Total-Count", value)
		hybris.Set("hybris-count", value)
		if got := totalCounting().total(counted); got != want {
			t.Errorf("X-Total-Count %q: total %d, want %d", value, got, want)
		}
		if got := hybrisCounting().total(hybris); got != want {
			t.Errorf("hybris-count %q: total %d, want %d", value, got, want)
		}
		if totalCounting().total(hybris) != -1 || hybrisCounting().total(counted) != -1 {
			t.Errorf("%q: a counting read the other's header", value)
		}
	}
}

// hybrisCountedRoute serves count items by pageNumber, with total as the
// hybris-count header when asked with totalCount=true, as sites are served.
func hybrisCountedRoute(count int, total string) func(http.ResponseWriter, *http.Request) {
	paged := pagedRoute(count, 0, "")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("totalCount") == "true" {
			w.Header().Set("hybris-count", total)
		}
		paged(w, r)
	}
}

func TestGetAllHybrisCounted(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", hybrisCountedRoute(3, "3"))
	client := pagedClient(t, fake)
	body, err := client.GetAllHybrisCounted(context.Background(), "/items", nil, ListOptions{Query: "active:true"})
	if err != nil || string(body) != `[{"id":"i0"},{"id":"i1"},{"id":"i2"}]` || len(fake.requests) != 2 {
		t.Fatalf("body %s, error %v, %d requests; want 3 items in 2 requests", body, err, len(fake.requests))
	}
	for _, sent := range fake.requests {
		query := sent.URL.Query()
		if query.Get("totalCount") != "true" || query.Get("q") != "active:true" || sent.Header.Get("X-Total-Count") != "" {
			t.Fatalf("request %s with X-Total-Count %q", sent.URL.RequestURI(), sent.Header.Get("X-Total-Count"))
		}
	}
	fake.route("/items", hybrisCountedRoute(3, "7"))
	if _, err := client.GetAllHybrisCounted(context.Background(), "/items", nil, ListOptions{}); err == nil ||
		!strings.Contains(err.Error(), "read 3 items but the service reports 7") {
		t.Fatalf("error = %v; a wrong hybris-count must fail", err)
	}
}

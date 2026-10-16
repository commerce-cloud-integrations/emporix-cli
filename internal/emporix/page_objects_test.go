package emporix

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// pageObjectRoute serves count items ({"id":"i<n>"}) by page (the first is 1)
// and size, each page an object with the items and, when total is not empty,
// totalElements. With ignorePaging, every page holds every item.
func pageObjectRoute(count int, total string, ignorePaging bool) func(http.ResponseWriter, *http.Request) {
	return pagesFrom(1, "items", count, total, ignorePaging)
}

// pagesFrom is pageObjectRoute for pages numbered from first that hold their
// items in member.
func pagesFrom(first int, member string, count int, total string, ignorePaging bool) func(http.ResponseWriter,
	*http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := strconv.Atoi(r.URL.Query().Get("size"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		start, end := (page-first)*size, (page-first+1)*size
		if ignorePaging {
			start, end = 0, count
		}
		items := []string{}
		for i := start; i < min(end, count); i++ {
			items = append(items, fmt.Sprintf(`{"id":"i%d"}`, i))
		}
		body := `{"` + member + `":[` + strings.Join(items, ",") + `],"page":` + strconv.Itoa(page) + `,"size":` +
			strconv.Itoa(size)
		if total != "" {
			body += `,"totalElements":` + total
		}
		_, _ = w.Write([]byte(body + "}"))
	}
}

func TestGetAllPageObjectsReadsEveryPage(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pageObjectRoute(5, "5", false))
	body, err := pagedClient(t, fake).GetAllPageObjects(context.Background(), "/items",
		map[string][]string{"q": {"type:update"}}, ListOptions{Query: "entity:order"})
	want := `[{"id":"i0"},{"id":"i1"},{"id":"i2"},{"id":"i3"},{"id":"i4"}]`
	if err != nil || string(body) != want || len(fake.requests) != 3 {
		t.Fatalf("body %s, error %v, %d requests; want 5 items in 3 requests", body, err, len(fake.requests))
	}
	last := fake.requests[2]
	query := last.URL.Query()
	if query.Get("page") != "3" || query.Get("size") != "2" || query.Get("q") != "type:update entity:order" ||
		query.Has("pageNumber") || query.Has("pageSize") || last.Header.Get("X-Total-Count") != "" {
		t.Fatalf("last request %s with X-Total-Count %q", last.URL.RequestURI(), last.Header.Get("X-Total-Count"))
	}
}

func TestGetAllPageObjectsFailures(t *testing.T) {
	const malformedItems = `GET /items page 1: expected an object with the items in "items" and an integer ` +
		`totalElements, if any`
	for _, tt := range []struct {
		name  string
		route func(http.ResponseWriter, *http.Request)
		want  string
	}{
		{name: "total larger than the collection", route: pageObjectRoute(3, "7", false),
			want: "read 3 items but the service reports 7"},
		{name: "paging ignored", route: pageObjectRoute(3, "", true), want: "page 2 repeats an item of page 1"},
		{name: "an array", route: respond(http.StatusOK, `[{"id":"i0"}]`), want: malformedItems},
		{name: "items null", route: respond(http.StatusOK, `{"items":null,"totalElements":0}`), want: malformedItems},
		{name: "items left out", route: respond(http.StatusOK, `{"totalElements":0}`), want: malformedItems},
		{name: "items named otherwise", route: respond(http.StatusOK, `{"Items":[],"totalElements":0}`),
			want: malformedItems},
		{name: "total not an integer", route: respond(http.StatusOK, `{"items":[],"totalElements":"3"}`),
			want: malformedItems},
		{name: "page error", route: respond(http.StatusBadRequest, `{"message":"size must be between 1 and 100"}`),
			want: "400 Bad Request: size must be between 1 and 100"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", tt.route)
			_, err := pagedClient(t, fake).GetAllPageObjects(context.Background(), "/items", nil, ListOptions{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

// TestGetAllPageObjectsWithoutATotal checks that pages without a usable total
// are read until an empty page, and that --limit stops early.
func TestGetAllPageObjectsWithoutATotal(t *testing.T) {
	for _, total := range []string{"", "-1", "null"} {
		fake := newFakeEmporix(t)
		fake.route("/items", pageObjectRoute(3, total, false))
		body, err := pagedClient(t, fake).GetAllPageObjects(context.Background(), "/items", nil, ListOptions{})
		if err != nil || string(body) != `[{"id":"i0"},{"id":"i1"},{"id":"i2"}]` || len(fake.requests) != 3 {
			t.Fatalf("total %q: body %s, error %v, %d requests", total, body, err, len(fake.requests))
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/items", pageObjectRoute(5, "5", false))
	body, err := pagedClient(t, fake).GetAllPageObjects(context.Background(), "/items", nil, ListOptions{Limit: 1})
	if err != nil || string(body) != `[{"id":"i0"}]` || len(fake.requests) != 1 {
		t.Fatalf("body %s, error %v, %d requests; want one item in one request", body, err, len(fake.requests))
	}
}

func TestGetAllContentPagesReadsEveryPageFromZero(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/runs", pagesFrom(0, "content", 5, "5", false))
	body, err := pagedClient(t, fake).GetAllContentPages(context.Background(), "/runs",
		map[string][]string{"type": {"product"}}, ListOptions{})
	want := `[{"id":"i0"},{"id":"i1"},{"id":"i2"},{"id":"i3"},{"id":"i4"}]`
	if err != nil || string(body) != want || len(fake.requests) != 3 {
		t.Fatalf("body %s, error %v, %d requests; want 5 items in 3 requests", body, err, len(fake.requests))
	}
	for i, request := range fake.requests {
		query := request.URL.Query()
		if query.Get("page") != strconv.Itoa(i) || query.Get("size") != "2" || query.Get("type") != "product" ||
			query.Has("pageNumber") || request.Header.Get("X-Total-Count") != "" {
			t.Fatalf("request %d: %s with X-Total-Count %q", i, request.URL.RequestURI(), request.Header.Get("X-Total-Count"))
		}
	}
}

func TestGetAllContentPagesFailures(t *testing.T) {
	const malformed = `expected an object with the items in "content" and an integer totalElements, if any`
	for _, tt := range []struct {
		name  string
		route func(http.ResponseWriter, *http.Request)
		want  string
	}{
		// Messages name a page by its number in the request, the first 0.
		{name: "paging ignored", route: pagesFrom(0, "content", 3, "", true),
			want: "GET /runs: page 1 repeats an item of page 0; the collection is inconsistent"},
		{name: "total larger than the collection", route: pagesFrom(0, "content", 3, "7", false),
			want: "GET /runs: read 3 items but the service reports 7"},
		{name: "items instead of content", route: pagesFrom(0, "items", 3, "3", false), want: "page 0: " + malformed},
		{name: "not an object", route: respond(http.StatusOK, `null`), want: "page 0: " + malformed},
		{name: "malformed later page", route: pageSequenceFrom0(`{"content":[{"id":"i0"},{"id":"i1"}]}`, `{}`),
			want: "page 1: " + malformed},
		{name: "endless collection", route: pagesFrom(0, "content", 100, "", false), want: "GET /runs: more than 5 pages"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/runs", tt.route)
			_, err := pagedClient(t, fake).GetAllContentPages(context.Background(), "/runs", nil, ListOptions{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

// pageSequenceFrom0 serves bodies as the pages 0, 1, … that page selects, and
// pages without content after them.
func pageSequenceFrom0(bodies ...string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body := `{"content":[]}`
		if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page < len(bodies) {
			body = bodies[page]
		}
		_, _ = w.Write([]byte(body))
	}
}

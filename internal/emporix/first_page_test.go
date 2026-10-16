package emporix

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestFirstPageReadsOnePage checks that a glance (ListOptions.FirstPage)
// requests one page of every paged read and returns its items as read, where
// a listing would request more pages or fail its checks.
func TestFirstPageReadsOnePage(t *testing.T) {
	glance := ListOptions{FirstPage: true}
	reads := map[string]func(*Client, ListOptions) (json.RawMessage, error){
		"GetAll": func(c *Client, opts ListOptions) (json.RawMessage, error) {
			return c.GetAll(context.Background(), "/items", nil, opts)
		},
		"GetAllWithoutTotal": func(c *Client, opts ListOptions) (json.RawMessage, error) {
			return c.GetAllWithoutTotal(context.Background(), "/items", nil, opts)
		},
		"GetAllInOnePage": func(c *Client, opts ListOptions) (json.RawMessage, error) {
			return c.GetAllInOnePage(context.Background(), "/items", nil, 2, opts)
		},
		"GetAllOnFirstPage": func(c *Client, opts ListOptions) (json.RawMessage, error) {
			return c.GetAllOnFirstPage(context.Background(), "/items", nil, 2, opts)
		},
	}
	tests := []struct {
		name  string
		route func(http.ResponseWriter, *http.Request)
		opts  ListOptions
		want  string
	}{
		{name: "more pages than one", route: pagedRoute(5, 0, "5"), opts: glance, want: `[{"id":"i0"},{"id":"i1"}]`},
		{name: "a short page without a total", route: pagedRoute(1, 0, ""), opts: glance, want: `[{"id":"i0"}]`},
		{name: "a total the page lacks", route: pagedRoute(2, 0, "7"), opts: glance, want: `[{"id":"i0"},{"id":"i1"}]`},
		{name: "a capped page", route: pagedRoute(3, 1, "3"), opts: glance, want: `[{"id":"i0"}]`},
		{name: "a limit", route: pagedRoute(3, 0, "3"), opts: ListOptions{FirstPage: true, Limit: 1},
			want: `[{"id":"i0"}]`},
		{name: "an empty collection", route: pagedRoute(0, 0, ""), opts: glance, want: `[]`},
	}
	for name, read := range reads {
		for _, tt := range tests {
			t.Run(name+"/"+tt.name, func(t *testing.T) {
				fake := newFakeEmporix(t)
				fake.route("/items", tt.route)
				body, err := read(pagedClient(t, fake), tt.opts)
				if err != nil || string(body) != tt.want {
					t.Fatalf("body %s, error %v; want %s", body, err, tt.want)
				}
				if len(fake.requests) != 1 || fake.requests[0].URL.Query().Get("pageNumber") != "1" {
					t.Fatalf("%d requests; want page 1 only", len(fake.requests))
				}
			})
		}
	}
}

// TestFirstPageFailsWhereItsPageFails checks that a glance of a collection
// fails as reading its first page fails.
func TestFirstPageFailsWhereItsPageFails(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	_, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, ListOptions{FirstPage: true})
	if err == nil {
		t.Fatal("no error; want the page's")
	}
}

// TestFirstPageStillRefusesARepeatedItem checks that a glance fails on a page
// that lists an item twice, as a listing does.
func TestFirstPageStillRefusesARepeatedItem(t *testing.T) {
	for name, read := range map[string]func(*Client) error{
		"GetAll": func(c *Client) error {
			_, err := c.GetAll(context.Background(), "/items", nil, ListOptions{FirstPage: true})
			return err
		},
		"GetAllInOnePage": func(c *Client) error {
			_, err := c.GetAllInOnePage(context.Background(), "/items", nil, 2, ListOptions{FirstPage: true})
			return err
		},
	} {
		fake := newFakeEmporix(t)
		fake.route("/items", pageSequence([]string{`["A","A"]`}))
		if err := read(pagedClient(t, fake)); err == nil || !strings.Contains(err.Error(), "lists an item twice") {
			t.Errorf("%s: error %v; want an item listed twice", name, err)
		}
	}
}

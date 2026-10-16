package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

// streamedIDs returns opts streaming into pages, each the ids of its items.
func streamedIDs(t *testing.T, opts ListOptions) (streaming ListOptions, pages *[][]string) {
	t.Helper()
	pages = &[][]string{}
	opts.Stream = func(items []json.RawMessage) error {
		body, err := json.Marshal(items)
		if err != nil {
			return err
		}
		*pages = append(*pages, itemIDs(t, body))
		return nil
	}
	return opts, pages
}

func TestGetAllStreamsPagesAsItChecksThem(t *testing.T) {
	tests := []struct {
		name           string
		maxSize, limit int
		want           [][]string
		reads          int // pages requested
	}{
		// Every page, as read, until the total is reached.
		{name: "whole collection", want: [][]string{{"i0", "i1"}, {"i2", "i3"}, {"i4"}}, reads: 3},
		// At most the limit, stopping early as a kept read does.
		{name: "limit", limit: 3, want: [][]string{{"i0", "i1"}, {"i2"}}, reads: 2},
		// Pages capped at one item: the limit is streamed, and the read goes on to check the total.
		{name: "limit after a short page", maxSize: 1, limit: 2, want: [][]string{{"i0"}, {"i1"}}, reads: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(5, tt.maxSize, "5"))
			opts, pages := streamedIDs(t, ListOptions{Limit: tt.limit})
			body, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, opts)
			if err != nil || body != nil || !slices.EqualFunc(*pages, tt.want, slices.Equal) ||
				len(fake.requests) != tt.reads {
				t.Fatalf("body %s, error %v, pages %q, requests %d; want no body, pages %q, %d requests",
					body, err, *pages, len(fake.requests), tt.want, tt.reads)
			}
		})
	}
}

func TestGetAllStreamFailsAfterThePagesBefore(t *testing.T) {
	tests := []struct {
		name    string
		total   string
		maxSize int
		want    [][]string
		message string
	}{
		// The service ignores paging: page 2 repeats page 1, which was streamed.
		{name: "repeated page", maxSize: -1, want: [][]string{{"i0", "i1", "i2"}}, message: "page 2 repeats"},
		// The collection ends before its total: every page was streamed.
		{name: "count below total", total: "4", want: [][]string{{"i0", "i1"}, {"i2"}},
			message: "read 3 items but the service reports 4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(3, tt.maxSize, tt.total))
			opts, pages := streamedIDs(t, ListOptions{})
			_, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, opts)
			if err == nil || !strings.Contains(err.Error(), tt.message) || !slices.EqualFunc(*pages, tt.want, slices.Equal) {
				t.Fatalf("error %v, pages %q; want %q after pages %q", err, *pages, tt.message, tt.want)
			}
		})
	}
}

func TestReadsStopWhenTheStreamFails(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(5, 0, "5"))
	failed := errors.New("broken pipe")
	failing := ListOptions{Stream: func([]json.RawMessage) error { return failed }}
	client := pagedClient(t, fake)
	_, err := client.GetAll(context.Background(), "/items", nil, failing)
	if !errors.Is(err, failed) || len(fake.requests) != 1 {
		t.Fatalf("error %v after %d requests; want the stream's error after one", err, len(fake.requests))
	}
	fake.route("/items", pagedRoute(5, -1, "5"))
	if _, err := client.GetUnpaged(context.Background(), "/items", nil, failing); !errors.Is(err, failed) {
		t.Fatalf("unpaged: error %v; want the stream's error", err)
	}
}

func TestWholeCollectionReadsStreamOnceChecked(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, -1, "3"))
	client := pagedClient(t, fake)
	opts, pages := streamedIDs(t, ListOptions{Limit: 2})
	body, err := client.GetUnpaged(context.Background(), "/items", nil, opts)
	if err != nil || body != nil || !slices.EqualFunc(*pages, [][]string{{"i0", "i1"}}, slices.Equal) {
		t.Fatalf("unpaged: body %s, error %v, pages %q", body, err, *pages)
	}
	// A page whose count differs from its total fails before anything is streamed.
	fake.route("/items", pagedRoute(3, 0, "2"))
	opts, pages = streamedIDs(t, ListOptions{})
	if _, err := client.GetAllInOnePage(context.Background(), "/items", nil, 10, opts); err == nil || len(*pages) != 0 {
		t.Fatalf("one page: error %v, pages %q; want a failure before streaming", err, *pages)
	}
}

// TestGetAllStreamsNoPageBeyondTheTotal checks that a page holding more items
// than the total is not streamed: the count check runs before it is handed over.
func TestGetAllStreamsNoPageBeyondTheTotal(t *testing.T) {
	tests := []struct {
		name  string
		pages []string
		total string
		want  [][]string
	}{
		{name: "first page", pages: []string{`[{"id":"i0"},{"id":"i1"}]`}, total: "1", want: [][]string{}},
		{name: "later page", pages: []string{`[{"id":"i0"},{"id":"i1"}]`, `[{"id":"i2"},{"id":"i3"}]`}, total: "3",
			want: [][]string{{"i0", "i1"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pageSequence(tt.pages, tt.total, tt.total))
			opts, pages := streamedIDs(t, ListOptions{})
			_, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, opts)
			if err == nil || !strings.Contains(err.Error(), "but the service reports "+tt.total) ||
				!slices.EqualFunc(*pages, tt.want, slices.Equal) {
				t.Fatalf("error %v, pages %q; want the count mismatch after pages %q", err, *pages, tt.want)
			}
		})
	}
}

// TestStreamedReadsTakeMorePages checks that a read that streams may request
// more pages than one that keeps its items (pagedClient: 8, not 5), and that
// both still stop a collection that never ends.
func TestStreamedReadsTakeMorePages(t *testing.T) {
	tests := []struct {
		name     string
		count    int // items served, in pages of 2
		total    string
		stream   bool
		requests int    // pages requested
		streamed int    // pages streamed
		want     string // the error; empty for none
	}{
		{name: "kept read beyond its cap", count: 12, total: "12", requests: 5,
			want: "GET /items: more than 5 pages; a read that streams its items takes up to 8"},
		{name: "streamed read beyond the kept cap", count: 12, total: "12", stream: true, requests: 6, streamed: 6},
		{name: "streamed read of a collection that never ends", count: 100, stream: true, requests: 8, streamed: 8,
			want: "GET /items: more than 8 pages"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(tt.count, 0, tt.total))
			opts, pages := ListOptions{}, &[][]string{}
			if tt.stream {
				opts, pages = streamedIDs(t, opts)
			}
			_, err := pagedClient(t, fake).GetAll(context.Background(), "/items", nil, opts)
			if got := errorText(err); got != tt.want {
				t.Fatalf("error %q; want %q (empty: none)", got, tt.want)
			}
			if len(fake.requests) != tt.requests || len(*pages) != tt.streamed {
				t.Fatalf("%d requests, %d pages streamed; want %d, %d", len(fake.requests), len(*pages), tt.requests,
					tt.streamed)
			}
		})
	}
}

// TestDefaultPageCaps checks the page caps of a client as New makes it: a
// kept read stops after KeptPages, on a service that caps its pages at one
// item, and says that a streamed read takes up to StreamedPages, its cap.
func TestDefaultPageCaps(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(1001, 1, "1001"))
	client := fake.client(t)
	const want = "GET /items: more than 1000 pages; a read that streams its items takes up to 10000"
	if _, err := client.GetAll(context.Background(), "/items", nil, ListOptions{}); errorText(err) != want ||
		len(fake.requests) != 1000 {
		t.Fatalf("kept: error %v after %d requests; want %q after 1000", err, len(fake.requests), want)
	}
	streamed := ListOptions{Stream: func([]json.RawMessage) error { return nil }}
	if got := client.paging.pages(streamed); got != 10000 {
		t.Fatalf("a streamed read takes %d pages; want 10000", got)
	}
}

// errorText returns the message of err, or "" for no error.
func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

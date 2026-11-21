package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func streamedIDs(t *testing.T) (sink func(items []json.RawMessage) error, pages *[][]string) {
	t.Helper()
	pages = &[][]string{}
	return func(items []json.RawMessage) error {
		body, err := json.Marshal(items)
		if err != nil {
			return err
		}
		*pages = append(*pages, itemIDs(t, body))
		return nil
	}, pages
}

func TestGetAllStreamsPagesAsItChecksThem(t *testing.T) {
	tests := []struct {
		name           string
		maxSize, limit int
		want           [][]string
		reads          int
	}{
		{name: "whole collection", want: [][]string{{"i0", "i1"}, {"i2", "i3"}, {"i4"}}, reads: 3},
		{name: "limit", limit: 3, want: [][]string{{"i0", "i1"}, {"i2"}}, reads: 2},
		{name: "limit after a short page", maxSize: 1, limit: 2, want: [][]string{{"i0"}, {"i1"}}, reads: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(5, tt.maxSize, "5"))
			sink, pages := streamedIDs(t)
			body, err := pagedClient(t, fake).Streaming(sink).GetAll(context.Background(), "/items", nil,
				ListOptions{Limit: tt.limit})
			if err != nil || body != nil || !slices.EqualFunc(*pages, tt.want, slices.Equal) ||
				len(fake.requests) != tt.reads {
				t.Fatalf("body %s, error %v, pages %q, requests %d; want no body, pages %q, %d requests",
					body, err, *pages, len(fake.requests), tt.want, tt.reads)
			}
		})
	}
}

func TestClientsDerivedFromAStreamingOneKeepTheirItems(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, 0, "3"))
	sink, pages := streamedIDs(t)
	streaming := pagedClient(t, fake).Streaming(sink)
	derived := map[string]*Client{"WithHeader": withHeader(t, streaming, siteHeader, "main"),
		"AcceptCreated": streaming.AcceptCreated(), "ForbiddenIsRefusal": streaming.ForbiddenIsRefusal()}
	for name, client := range derived {
		body, err := client.GetAll(context.Background(), "/items", nil, ListOptions{})
		if err != nil || !slices.Equal(itemIDs(t, body), []string{"i0", "i1", "i2"}) || len(*pages) != 0 {
			t.Fatalf("%s: body %s, error %v, %d pages streamed; want the items kept, none streamed", name, body, err,
				len(*pages))
		}
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
		{name: "repeated page", maxSize: -1, want: [][]string{{"i0", "i1", "i2"}}, message: "page 2 repeats"},
		{name: "count below total", total: "4", want: [][]string{{"i0", "i1"}, {"i2"}},
			message: "read 3 items but the service reports 4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/items", pagedRoute(3, tt.maxSize, tt.total))
			sink, pages := streamedIDs(t)
			_, err := pagedClient(t, fake).Streaming(sink).GetAll(context.Background(), "/items", nil, ListOptions{})
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
	client := pagedClient(t, fake).Streaming(func([]json.RawMessage) error { return failed })
	_, err := client.GetAll(context.Background(), "/items", nil, ListOptions{})
	if !errors.Is(err, failed) || len(fake.requests) != 1 {
		t.Fatalf("error %v after %d requests; want the stream's error after one", err, len(fake.requests))
	}
	fake.route("/items", pagedRoute(5, -1, "5"))
	if _, err := client.GetUnpaged(context.Background(), "/items", nil, ListOptions{}); !errors.Is(err, failed) {
		t.Fatalf("unpaged: error %v; want the stream's error", err)
	}
}

func TestWholeCollectionReadsStreamOnceChecked(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(3, -1, "3"))
	client := pagedClient(t, fake)
	sink, pages := streamedIDs(t)
	body, err := client.Streaming(sink).GetUnpaged(context.Background(), "/items", nil, ListOptions{Limit: 2})
	if err != nil || body != nil || !slices.EqualFunc(*pages, [][]string{{"i0", "i1"}}, slices.Equal) {
		t.Fatalf("unpaged: body %s, error %v, pages %q", body, err, *pages)
	}
	fake.route("/items", pagedRoute(3, 0, "2"))
	sink, pages = streamedIDs(t)
	if _, err := client.Streaming(sink).GetAllInOnePage(context.Background(), "/items", nil, 10,
		ListOptions{}); err == nil || len(*pages) != 0 {
		t.Fatalf("one page: error %v, pages %q; want a failure before streaming", err, *pages)
	}
}

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
			sink, pages := streamedIDs(t)
			_, err := pagedClient(t, fake).Streaming(sink).GetAll(context.Background(), "/items", nil, ListOptions{})
			if err == nil || !strings.Contains(err.Error(), "but the service reports "+tt.total) ||
				!slices.EqualFunc(*pages, tt.want, slices.Equal) {
				t.Fatalf("error %v, pages %q; want the count mismatch after pages %q", err, *pages, tt.want)
			}
		})
	}
}

func TestStreamedReadsTakeMorePages(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		total    string
		stream   bool
		requests int
		streamed int
		want     string
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
			client, pages := pagedClient(t, fake), &[][]string{}
			if tt.stream {
				var sink func(items []json.RawMessage) error
				sink, pages = streamedIDs(t)
				client = client.Streaming(sink)
			}
			_, err := client.GetAll(context.Background(), "/items", nil, ListOptions{})
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

func TestDefaultPageCaps(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/items", pagedRoute(1001, 1, "1001"))
	client := fake.client(t)
	const want = "GET /items: more than 1000 pages; a read that streams its items takes up to 10000"
	if _, err := client.GetAll(context.Background(), "/items", nil, ListOptions{}); errorText(err) != want ||
		len(fake.requests) != 1000 {
		t.Fatalf("kept: error %v after %d requests; want %q after 1000", err, len(fake.requests), want)
	}
	if got := client.paging.pages(true); got != 10000 {
		t.Fatalf("a streamed read takes %d pages; want 10000", got)
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

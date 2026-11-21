package emporix

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strings"
)

// ListOptions narrow a collection read (GetAll and its siblings): it returns
// at most Limit items (0 or less: every item); Query filters
// with the Emporix q parameter, together with the q of the read's own query
// (both must match, so a filter can only narrow a read); a read with a body
// (WithBody) filters with its body and refuses a Query; Sort orders with the
// sort parameter, replacing the read's own order. Empty values change nothing.
//
// A paged read stops requesting pages once it has Limit items, as long as
// every page before the last one read was full: the items of a page are
// contiguous, so a gap can only follow a short page (the service may cap the
// page size), and after one the read goes on to the end and checks the total.
// A read cut early cannot compare its count with the reported total, while a
// repeated item still fails it. Reads of a whole collection are checked first,
// then cut.
//
// Stream, when set, receives the items as a paged read checks them, page by
// page (at most Limit in all), instead of the read keeping them: the read then
// returns no items (nil). A check that fails on a later page (an item read
// twice, a total that changes or differs from the count) fails the read after
// the items before it were streamed. A read of a whole collection streams its
// items once they are checked. A paged read that streams keeps only a key per
// item, to find one read twice, so it may request more pages than a read that
// keeps its items (StreamedPages, KeptPages).
//
// FirstPage, when set, reads the first page of a paged collection only and
// returns its items, checked for an item listed twice but neither against a
// total nor against the pages after it, which it does not request. A read of a
// whole collection reads its one response as always.
type ListOptions struct {
	Limit     int
	Query     string
	Sort      string
	Stream    func(items []json.RawMessage) error
	FirstPage bool
}

func (o ListOptions) apply(query url.Values) url.Values {
	merged := url.Values{}
	maps.Copy(merged, query)
	if o.Query != "" {
		// Space-separated conditions must all match, so the read's own filter stays.
		merged.Set("q", strings.TrimSpace(merged.Get("q")+" "+o.Query))
	}
	if o.Sort != "" {
		merged.Set("sort", o.Sort)
	}
	return merged
}

func (o ListOptions) streams() bool { return o.Stream != nil }

func (o ListOptions) pageSize(size int) int {
	if o.Limit > 0 {
		return min(size, o.Limit)
	}
	return size
}

func (o ListOptions) deliverFirstPage(read string, items []json.RawMessage) (json.RawMessage, error) {
	if err := (itemPages{}).add(read, 1, items); err != nil {
		return nil, err
	}
	return o.deliver(items)
}

func (o ListOptions) reached(n int) bool { return o.Limit > 0 && n >= o.Limit }

func (o ListOptions) deliver(items []json.RawMessage) (json.RawMessage, error) {
	return (&delivery{opts: o}).last(items)
}

type delivery struct {
	opts   ListOptions
	kept   []json.RawMessage
	handed int
}

func (d *delivery) add(items []json.RawMessage) error {
	if d.opts.Limit > 0 {
		items = items[:min(len(items), d.opts.Limit-d.handed)]
	}
	d.handed += len(items)
	switch {
	case !d.opts.streams():
		d.kept = append(d.kept, items...)
		return nil
	case len(items) == 0:
		return nil
	default:
		return d.opts.Stream(items)
	}
}

func (d *delivery) last(items []json.RawMessage) (json.RawMessage, error) {
	if err := d.add(items); err != nil {
		return nil, err
	}
	return d.result()
}

func (d *delivery) result() (json.RawMessage, error) {
	if d.opts.streams() {
		return nil, nil
	}
	if d.kept == nil {
		return json.RawMessage(`[]`), nil
	}
	return json.Marshal(d.kept)
}

// A read with a body (WithBody) refuses a Query, as Emporix ignores a q
// parameter there.
func (c *Client) collectionQuery(path string, query url.Values, opts ListOptions) (url.Values, error) {
	switch {
	case strings.Contains(path, "?"):
		return nil, fmt.Errorf("%w: %s: pass a collection's query separately, not in its path",
			ErrInvalidInput, c.named(withoutQuery(path)))
	case !c.body.none() && opts.Query != "":
		return nil, fmt.Errorf("%w: %s: a read with a body filters with its body, not with a q parameter",
			ErrInvalidInput, c.named(path))
	}
	return opts.apply(query), nil
}

func withoutQuery(path string) string {
	before, _, _ := strings.Cut(path, "?")
	return before
}

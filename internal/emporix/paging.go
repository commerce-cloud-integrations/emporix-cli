package emporix

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
)

const totalCountHeader = "X-Total-Count"

type counting struct {
	header http.Header
	query  url.Values
	from   string
}

func totalCounting() counting {
	return counting{header: http.Header{totalCountHeader: {"true"}}, from: totalCountHeader}
}

func unaskedTotal() counting { return counting{from: totalCountHeader} }

func hybrisCounting() counting {
	return counting{query: url.Values{"totalCount": {"true"}}, from: "hybris-count"}
}

func (c counting) total(header http.Header) reportedTotal {
	total, err := strconv.Atoi(header.Get(c.from))
	if err != nil || total < 0 {
		return reportedTotal{}
	}
	return reported(total)
}

type reportedTotal struct {
	items int
	known bool
}

func reported(items int) reportedTotal { return reportedTotal{items: items, known: true} }

func (t reportedTotal) exceeds(count int) bool { return t.known && t.items > count }

func (t reportedTotal) reachedBy(count int) bool { return t.known && count >= t.items }

func (t reportedTotal) check(read string, count int) error {
	if t.known && t.items != count {
		return fmt.Errorf("%s: read %d items but the service reports %d; retry the command", read, count, t.items)
	}
	return nil
}

// The page caps of a client as New makes it, with pages of 100: a read that
// keeps its items takes at most KeptPages (100,000 items); a read that streams
// them (Streaming) keeps only the key of each item's duplicate check
// (itemPages, about 100 bytes), so it takes up to StreamedPages (a million
// items, whose checks take about 100 MB).
const (
	KeptPages     = 1000
	StreamedPages = 10000
)

type pagingPolicy struct {
	size          int
	keptPages     int
	streamedPages int
}

func defaultPagingPolicy() pagingPolicy {
	return pagingPolicy{size: 100, keptPages: KeptPages, streamedPages: StreamedPages}
}

func (p pagingPolicy) pages(streamed bool) int {
	if streamed {
		return p.streamedPages
	}
	return p.keptPages
}

func (p pagingPolicy) tooMany(read string, pages int) error {
	if pages < p.streamedPages {
		return fmt.Errorf("%s: more than %d pages; a read that streams its items takes up to %d", read, pages,
			p.streamedPages)
	}
	return fmt.Errorf("%s: more than %d pages", read, pages)
}

// GetAll reads every page of a collection that pages with pageNumber and
// pageSize, using one token, and returns all items as one JSON array. A short
// page does not end the collection, because a service may cap the page size:
// reading stops at the reported X-Total-Count or, without one, at the first
// empty page. Inconsistent paging fails instead of returning duplicated or
// truncated data: an item read twice (for example when the service
// ignores pageNumber), a total that changes between pages, or a final count
// that differs from the total. opts narrow the read (see ListOptions).
func (c *Client) GetAll(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	return c.getAll(ctx, path, query, opts, c.arrayPages(totalCounting()))
}

// GetAllHybrisCounted is GetAll for a collection that reports its total only
// when asked with totalCount=true, in a hybris-count header, as the Site
// Settings Service and the Coupon Service do. Live, that total counts the
// collection as filtered (by q), and leaves out the deleted coupons, which are
// not read; that it leaves out the inactive sites, which are not read either,
// is assumed, as no tenant tested had one. Paging is checked as GetAll checks
// it.
func (c *Client) GetAllHybrisCounted(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	return c.getAll(ctx, path, query, opts, c.arrayPages(hybrisCounting()))
}

// GetAllWithoutTotal is GetAll for a service that fails when asked for its
// total (the reindex-jobs endpoint answers 500 to X-Total-Count): it does not
// request the total, so it reads until the first empty page, but checks one
// that the service sends unasked. It still fails
// when an item is read twice, but without a total it
// cannot notice an item that moved between pages or disappeared meanwhile.
func (c *Client) GetAllWithoutTotal(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	return c.getAll(ctx, path, query, opts, c.arrayPages(unaskedTotal()))
}

// GetAllPageObjects is GetAll for a collection that pages with page (the
// first is 1) and size and answers each page with an object that holds its
// items and their total ({"items": […], "totalElements": n, …}), as the
// Changelog Service does. Paging is checked as GetAll checks it. Pages hold
// the client's page size of 100, which is also that service's maximum.
func (c *Client) GetAllPageObjects(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	return c.getAll(ctx, path, query, opts, c.pageObjects("items"))
}

// GetAllContentPages is GetAllPageObjects for a collection whose pages are
// numbered from 0 and hold their items in content ({"content": […],
// "totalElements": n, …}), as the Import Service pages runs, run errors and
// imported records. Paging is checked as GetAll checks it; messages name the
// pages as requested, the first page 0.
func (c *Client) GetAllContentPages(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	return c.getAllFrom(ctx, path, query, opts, 0, c.pageObjects("content"))
}

// GetUnpaged reads a collection that the service returns whole, ignoring
// pageNumber and pageSize (GetAll would fail on page 2 repeating page 1),
// with one request. Like GetAll, it fails unless the response is a JSON array
// of distinct items; opts filter and sort it, and cut it to the limit.
func (c *Client) GetUnpaged(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	query, err := c.collectionQuery(path, query, opts)
	if err != nil {
		return nil, err
	}
	result, err := c.get(ctx, path, query, nil)
	if err != nil {
		return nil, err
	}
	var items []json.RawMessage
	if err := json.Unmarshal(result.body, &items); err != nil || items == nil {
		return nil, fmt.Errorf("%s: expected a JSON array", c.named(path))
	}
	if err := (itemPages{}).add(c.named(path), 1, items); err != nil {
		return nil, err
	}
	return c.delivery(opts).last(items)
}

// GetAllInOnePage reads a collection whose pages cannot be combined as one
// page of up to size items: the Country Service orders its countries
// differently on every request, even when sorted, so their pages overlap and
// miss items. It fails rather than
// silently truncate: a reported total must match the page, and without a total
// the next page must be empty. opts filter and sort it, and cut the checked
// page to the limit.
func (c *Client) GetAllInOnePage(ctx context.Context, path string, query url.Values, size int,
	opts ListOptions) (json.RawMessage, error) {
	return c.getOnePage(ctx, path, query, size, opts, func(query url.Values) (reportedTotal, error) {
		next, total, err := c.getPage(ctx, path, query, 2, size, totalCounting())
		if err != nil {
			return reportedTotal{}, err
		}
		if len(next) > 0 {
			return reportedTotal{}, fmt.Errorf("%s: the collection does not fit one page of %d items", c.named(path), size)
		}
		return total, nil
	})
}

// GetAllOnFirstPage reads, as one page of up to size items, a collection that
// the service lists on its first page only: every later page is empty,
// whatever its size, as for the access controls of an IAM user. An empty next
// page proves nothing there, so it fails unless the service reports a total,
// which the page must match. opts filter and sort it, and cut the checked page
// to the limit.
func (c *Client) GetAllOnFirstPage(ctx context.Context, path string, query url.Values, size int,
	opts ListOptions) (json.RawMessage, error) {
	return c.getOnePage(ctx, path, query, size, opts, func(url.Values) (reportedTotal, error) {
		return reportedTotal{}, fmt.Errorf("%s: the service reported no total, so one page of %d may not hold every item",
			c.named(path), size)
	})
}

func (c *Client) getOnePage(ctx context.Context, path string, query url.Values, size int, opts ListOptions,
	untotalled func(query url.Values) (reportedTotal, error)) (json.RawMessage, error) {
	query, err := c.collectionQuery(path, query, opts)
	if err != nil {
		return nil, err
	}
	items, total, err := c.getPage(ctx, path, query, 1, size, totalCounting())
	if err != nil {
		return nil, err
	}
	if opts.FirstPage {
		return c.delivery(opts).firstPage(c.named(path), items)
	}
	if total.exceeds(len(items)) {
		return nil, fmt.Errorf("%s: one page of %d returned %d of %d items; the service may limit the page size",
			c.named(path), size, len(items), total.items)
	}
	if !total.known {
		if total, err = untotalled(query); err != nil {
			return nil, err
		}
	}
	if err := (itemPages{}).add(c.named(path), 1, items); err != nil {
		return nil, err
	}
	if err := total.check(c.named(path), len(items)); err != nil {
		return nil, err
	}
	return c.delivery(opts).last(items)
}

type itemPages map[[sha256.Size]byte]int

func (seen itemPages) add(read string, page int, items []json.RawMessage) error {
	for _, item := range items {
		key := sha256.Sum256([]byte(identity(item)))
		first, ok := seen[key]
		switch {
		case !ok:
			seen[key] = page
		case first == page:
			return fmt.Errorf("%s: a response lists an item twice; the collection is inconsistent", read)
		default:
			return fmt.Errorf("%s: page %d repeats an item of page %d; the collection is inconsistent",
				read, page, first)
		}
	}
	return nil
}

func identity(item json.RawMessage) string {
	var named struct {
		ID *string `json:"id"`
	}
	if json.Unmarshal(item, &named) == nil && named.ID != nil {
		return "id\x00" + *named.ID
	}
	return "item\x00" + canonical(item)
}

func canonical(item json.RawMessage) string {
	decoder := json.NewDecoder(bytes.NewReader(item))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return string(item)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return string(item)
	}
	return string(encoded)
}

type readPage func(ctx context.Context, path string, query url.Values, page, size int) (
	items []json.RawMessage, total reportedTotal, err error)

func (c *Client) arrayPages(count counting) readPage {
	return func(ctx context.Context, path string, query url.Values, page, size int) ([]json.RawMessage,
		reportedTotal, error) {
		return c.getPage(ctx, path, query, page, size, count)
	}
}

func (c *Client) getAll(ctx context.Context, path string, query url.Values, opts ListOptions,
	read readPage) (json.RawMessage, error) {
	return c.getAllFrom(ctx, path, query, opts, 1, read)
}

func (c *Client) getAllFrom(ctx context.Context, path string, query url.Values, opts ListOptions, first int,
	read readPage) (json.RawMessage, error) {
	query, err := c.collectionQuery(path, query, opts)
	if err != nil {
		return nil, err
	}
	if opts.FirstPage {
		items, _, err := read(ctx, path, query, first, opts.pageSize(c.paging.size))
		if err != nil {
			return nil, err
		}
		return c.delivery(opts).firstPage(c.named(path), items)
	}
	return c.readPages(ctx, path, query, opts, first, read)
}

type pageReader struct {
	name          string
	opts          ListOptions
	out           *delivery
	seen          itemPages
	total         reportedTotal
	count         int
	size          int
	shortPageSeen bool
}

func (c *Client) readPages(ctx context.Context, path string, query url.Values, opts ListOptions, first int,
	read readPage) (json.RawMessage, error) {
	r := &pageReader{name: c.named(path), opts: opts, out: c.delivery(opts), seen: itemPages{},
		size: opts.pageSize(c.paging.size)}
	pages := c.paging.pages(r.out.streams())
	for page := first; page < first+pages; page++ {
		result, done, err := r.step(ctx, read, path, query, page)
		if done || err != nil {
			return result, err
		}
	}
	return nil, c.paging.tooMany(r.name, pages)
}

func (r *pageReader) step(ctx context.Context, read readPage, path string, query url.Values, page int) (
	result json.RawMessage, done bool, err error) {
	batch, pageTotal, err := read(ctx, path, query, page, r.size)
	if err != nil {
		return nil, false, err
	}
	if r.total, err = keepTotal(r.name, r.total, pageTotal); err != nil {
		return nil, false, err
	}
	if err := r.seen.add(r.name, page, batch); err != nil {
		return nil, false, err
	}
	if r.count += len(batch); len(batch) == 0 || r.total.reachedBy(r.count) {
		if err := r.total.check(r.name, r.count); err != nil {
			return nil, false, err
		}
		result, err := r.out.last(batch)
		return result, true, err
	}
	if err := r.out.add(batch); err != nil {
		return nil, false, err
	}
	if !r.shortPageSeen && r.opts.reached(r.count) {
		result, err := r.out.result()
		return result, true, err
	}
	r.shortPageSeen = r.shortPageSeen || len(batch) < r.size
	return nil, false, nil
}

func (c *Client) named(path string) string { return c.method + " " + path }

func keepTotal(read string, total, pageTotal reportedTotal) (reportedTotal, error) {
	switch {
	case !pageTotal.known || pageTotal == total:
		return total, nil
	case !total.known:
		return pageTotal, nil
	default:
		return total, fmt.Errorf("%s: the total changed from %d to %d while paging; retry the command",
			read, total.items, pageTotal.items)
	}
}

func (c *Client) getPage(ctx context.Context, path string, query url.Values, page, size int, count counting) (
	items []json.RawMessage, total reportedTotal, err error) {
	asked := pageQuery(query, "pageNumber", page, "pageSize", size)
	maps.Copy(asked, count.query)
	return c.arrayPage(ctx, path, asked, page, count)
}

func (c *Client) arrayPage(ctx context.Context, path string, query url.Values, page int, count counting) (
	items []json.RawMessage, total reportedTotal, err error) {
	result, err := c.get(ctx, path, query, count.header)
	if err != nil {
		return nil, reportedTotal{}, err
	}
	if err := json.Unmarshal(result.body, &items); err != nil || items == nil {
		return nil, reportedTotal{}, fmt.Errorf("%s page %d: expected a JSON array", c.named(path), page)
	}
	return items, count.total(result.header), nil
}

func (c *Client) pageObjects(member string) readPage {
	return func(ctx context.Context, path string, query url.Values, page, size int) ([]json.RawMessage,
		reportedTotal, error) {
		result, err := c.get(ctx, path, pageQuery(query, "page", page, "size", size), nil)
		if err != nil {
			return nil, reportedTotal{}, err
		}
		return pageObjectItems(result.body, member, fmt.Sprintf("%s page %d", c.named(path), page))
	}
}

func pageObjectItems(body []byte, member, read string) (items []json.RawMessage, total reportedTotal, err error) {
	malformed := func() error {
		return fmt.Errorf("%s: expected an object with the items in %q and an integer totalElements, if any",
			read, member)
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(body, &members) != nil || json.Unmarshal(members[member], &items) != nil || items == nil {
		return nil, reportedTotal{}, malformed()
	}
	var given *int
	if raw, ok := members["totalElements"]; ok && json.Unmarshal(raw, &given) != nil {
		return nil, reportedTotal{}, malformed()
	}
	if given == nil || *given < 0 {
		return items, reportedTotal{}, nil
	}
	return items, reported(*given), nil
}

func pageQuery(query url.Values, pageName string, page int, sizeName string, size int) url.Values {
	paged := url.Values{}
	maps.Copy(paged, query)
	paged.Set(pageName, strconv.Itoa(page))
	paged.Set(sizeName, strconv.Itoa(size))
	return paged
}

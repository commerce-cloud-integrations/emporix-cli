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

// totalCountHeader asks for (with "true") and carries a collection's total.
const totalCountHeader = "X-Total-Count"

// counting is how a collection read in array pages reports its total: what
// every request adds to ask for it, and the response header that carries it.
type counting struct {
	header http.Header // request headers that ask for the total
	query  url.Values  // query parameters that ask for the total
	from   string      // the response header with the total
}

// totalCounting asks for and reads X-Total-Count, as most services report
// their totals.
func totalCounting() counting {
	return counting{header: http.Header{totalCountHeader: {"true"}}, from: totalCountHeader}
}

// unaskedTotal reads X-Total-Count without asking for it, for a service that
// fails when asked but may still send it.
func unaskedTotal() counting { return counting{from: totalCountHeader} }

// hybrisCounting asks with totalCount=true and reads the hybris-count header.
func hybrisCounting() counting {
	return counting{query: url.Values{"totalCount": {"true"}}, from: "hybris-count"}
}

// total returns the total that header reports, or -1 when it reports no
// valid one.
func (c counting) total(header http.Header) int {
	total, err := strconv.Atoi(header.Get(c.from))
	if err != nil || total < 0 {
		return -1
	}
	return total
}

// The page caps of a client as New makes it, with pages of 100: a read that
// keeps its items takes at most KeptPages (100,000 items); a read that streams
// them (ListOptions.Stream) keeps only the key of each item's duplicate check
// (itemPages, about 100 bytes), so it takes up to StreamedPages (a million
// items, whose checks take about 100 MB).
const (
	KeptPages     = 1000
	StreamedPages = 10000
)

// pagingPolicy controls how GetAll walks pageNumber/pageSize collections. Its
// page caps stop a collection that never ends; one that repeats its pages
// fails earlier, on the first item read twice.
type pagingPolicy struct {
	size          int // items requested per page
	keptPages     int // the page cap of a read that keeps its items
	streamedPages int // the page cap of a read that streams them
}

func defaultPagingPolicy() pagingPolicy {
	return pagingPolicy{size: 100, keptPages: KeptPages, streamedPages: StreamedPages}
}

// pages returns how many pages a read with opts may request.
func (p pagingPolicy) pages(opts ListOptions) int {
	if opts.streams() {
		return p.streamedPages
	}
	return p.keptPages
}

// tooMany returns the error of the read named read (named) that requested
// its cap of pages pages (pages) without reaching the end. Below the streamed
// cap, the error says how far a read that streams its items goes.
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
	return opts.deliver(items)
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
	return c.getOnePage(ctx, path, query, size, opts, func(query url.Values) (int, error) {
		next, total, err := c.getPage(ctx, path, query, 2, size, totalCounting())
		if err != nil {
			return 0, err
		}
		if len(next) > 0 {
			return 0, fmt.Errorf("%s: the collection does not fit one page of %d items", c.named(path), size)
		}
		return total, nil // a total reported only now must match the first page too
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
	return c.getOnePage(ctx, path, query, size, opts, func(url.Values) (int, error) {
		return 0, fmt.Errorf("%s: the service reported no total, so one page of %d may not hold every item",
			c.named(path), size)
	})
}

// getOnePage reads page 1 of a collection in pages of size items, asking for
// its total. Without a total, untotalled is called with the request's query
// and returns the total to check the page against (-1: none) or why the page
// cannot be trusted.
func (c *Client) getOnePage(ctx context.Context, path string, query url.Values, size int, opts ListOptions,
	untotalled func(query url.Values) (int, error)) (json.RawMessage, error) {
	query, err := c.collectionQuery(path, query, opts)
	if err != nil {
		return nil, err
	}
	items, total, err := c.getPage(ctx, path, query, 1, size, totalCounting())
	if err != nil {
		return nil, err
	}
	if opts.FirstPage {
		return opts.deliverFirstPage(c.named(path), items)
	}
	if total > len(items) {
		return nil, fmt.Errorf("%s: one page of %d returned %d of %d items; the service may limit the page size",
			c.named(path), size, len(items), total)
	}
	if total < 0 {
		if total, err = untotalled(query); err != nil {
			return nil, err
		}
	}
	if err := (itemPages{}).add(c.named(path), 1, items); err != nil {
		return nil, err
	}
	if err := checkCount(c.named(path), len(items), total); err != nil {
		return nil, err
	}
	return opts.deliver(items)
}

// itemPages maps each item read to the page it was read on, by the hash of its
// identity, so a long collection keeps a few bytes per item, not the item.
type itemPages map[[sha256.Size]byte]int

// add records the items of page, failing when one is the same item as one
// read before: on an earlier page (as when a service ignores pageNumber) or
// on this one. read names the read in messages (named).
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

// identity returns what makes item the same item on another page: its id, where
// it is an object with one, so an item whose fields changed between pages is
// still recognised; else its canonical form.
func identity(item json.RawMessage) string {
	var named struct {
		ID *string `json:"id"`
	}
	if json.Unmarshal(item, &named) == nil && named.ID != nil {
		return "id\x00" + *named.ID
	}
	return "item\x00" + canonical(item)
}

// canonical returns item re-encoded with sorted keys and no insignificant
// whitespace, so the same value serialized differently compares equal.
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

// readPage reads the page of a collection numbered page, as the service numbers
// them: its items and the reported total (-1 if none).
type readPage func(ctx context.Context, path string, query url.Values, page, size int) (
	items []json.RawMessage, total int, err error)

// arrayPages reads pages that pageNumber and pageSize select, each a JSON
// array, with the total that count asks for (getPage).
func (c *Client) arrayPages(count counting) readPage {
	return func(ctx context.Context, path string, query url.Values, page, size int) ([]json.RawMessage, int, error) {
		return c.getPage(ctx, path, query, page, size, count)
	}
}

// getAll implements GetAll and its siblings, reading each page with read,
// the first numbered 1.
func (c *Client) getAll(ctx context.Context, path string, query url.Values, opts ListOptions,
	read readPage) (json.RawMessage, error) {
	return c.getAllFrom(ctx, path, query, opts, 1, read)
}

// getAllFrom is getAll for a collection whose first page is numbered first:
// read and every message name a page by its number in the request. With
// opts.FirstPage, it reads that page only.
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
		return opts.deliverFirstPage(c.named(path), items)
	}
	return c.readPages(ctx, path, query, opts, first, read)
}

// pageReader carries what readPages keeps from one page to the next.
type pageReader struct {
	name          string // how messages name the read (named)
	opts          ListOptions
	out           delivery
	seen          itemPages
	total, count  int  // the collection total reported so far (-1 if none), and the items read
	size          int  // the page size asked for
	everyPageFull bool // so far; a gap in the items can only follow a short page
}

// readPages implements getAllFrom for every page, its query applied.
func (c *Client) readPages(ctx context.Context, path string, query url.Values, opts ListOptions, first int,
	read readPage) (json.RawMessage, error) {
	r := &pageReader{name: c.named(path), opts: opts, out: delivery{opts: opts}, seen: itemPages{}, total: -1,
		size: opts.pageSize(c.paging.size), everyPageFull: true}
	pages := c.paging.pages(opts)
	for page := first; page < first+pages; page++ {
		result, done, err := r.step(ctx, read, path, query, page)
		if done || err != nil {
			return result, err
		}
	}
	return nil, c.paging.tooMany(r.name, pages)
}

// step reads page of path and adds its items to r. It returns the result of
// the read and done where the read ends with this page: the collection is
// complete, or the options stop it early.
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
	if r.count += len(batch); len(batch) == 0 || (r.total >= 0 && r.count >= r.total) {
		if err := checkCount(r.name, r.count, r.total); err != nil {
			return nil, false, err // before the page is handed over: it holds more than the total
		}
		result, err := r.out.last(batch)
		return result, true, err
	}
	if err := r.out.add(batch); err != nil {
		return nil, false, err
	}
	if r.everyPageFull && r.opts.reached(r.count) {
		result, err := r.out.result() // stopped early: the total cannot be checked
		return result, true, err
	}
	r.everyPageFull = r.everyPageFull && len(batch) >= r.size
	return nil, false, nil
}

// named returns how messages name a read of path: its method and path.
func (c *Client) named(path string) string { return c.method + " " + path }

// keepTotal returns the collection total reported so far; a total that
// changes between pages means the collection changed while it was read.
// read names the read in messages (named).
func keepTotal(read string, total, pageTotal int) (int, error) {
	switch {
	case pageTotal < 0 || pageTotal == total:
		return total, nil
	case total < 0:
		return pageTotal, nil
	default:
		return total, fmt.Errorf("%s: the total changed from %d to %d while paging; retry the command",
			read, total, pageTotal)
	}
}

// getPage reads one page and returns its items and the total that count
// reads (-1 if none).
func (c *Client) getPage(ctx context.Context, path string, query url.Values, page, size int, count counting) (
	items []json.RawMessage, total int, err error) {
	asked := pageQuery(query, "pageNumber", page, "pageSize", size)
	maps.Copy(asked, count.query)
	return c.arrayPage(ctx, path, asked, page, count)
}

// arrayPage reads page number page of a collection, a JSON array, with
// query as given, and returns its items and the total that count reads (-1
// if none).
func (c *Client) arrayPage(ctx context.Context, path string, query url.Values, page int, count counting) (
	items []json.RawMessage, total int, err error) {
	result, err := c.get(ctx, path, query, count.header)
	if err != nil {
		return nil, 0, err
	}
	if err := json.Unmarshal(result.body, &items); err != nil || items == nil {
		return nil, 0, fmt.Errorf("%s page %d: expected a JSON array", c.named(path), page)
	}
	return items, count.total(result.header), nil
}

// pageObjects reads the pages of GetAllPageObjects and GetAllContentPages,
// which page and size select: each an object that holds the page's items in
// member and the total, if it reports one, in totalElements.
func (c *Client) pageObjects(member string) readPage {
	return func(ctx context.Context, path string, query url.Values, page, size int) ([]json.RawMessage, int, error) {
		result, err := c.get(ctx, path, pageQuery(query, "page", page, "size", size), nil)
		if err != nil {
			return nil, 0, err
		}
		return pageObjectItems(result.body, member, fmt.Sprintf("%s page %d", c.named(path), page))
	}
}

// pageObjectItems returns the items of a page, an object that holds them in
// member, and the total it reports in totalElements (-1 if none). read names
// the page in messages, e.g. "GET /changelog page 2".
func pageObjectItems(body []byte, member, read string) (items []json.RawMessage, total int, err error) {
	malformed := func() error {
		return fmt.Errorf("%s: expected an object with the items in %q and an integer totalElements, if any",
			read, member)
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(body, &members) != nil || json.Unmarshal(members[member], &items) != nil || items == nil {
		return nil, 0, malformed()
	}
	var reported *int // none when null or left out
	if raw, ok := members["totalElements"]; ok && json.Unmarshal(raw, &reported) != nil {
		return nil, 0, malformed()
	}
	if reported == nil || *reported < 0 {
		return items, -1, nil
	}
	return items, *reported, nil
}

// pageQuery returns query with a page and its size set under the names the
// service reads them by; query is never changed.
func pageQuery(query url.Values, pageName string, page int, sizeName string, size int) url.Values {
	paged := url.Values{}
	maps.Copy(paged, query)
	paged.Set(pageName, strconv.Itoa(page))
	paged.Set(sizeName, strconv.Itoa(size))
	return paged
}

// checkCount fails when the service reports a total other than the count of
// items read. read names the read in messages (named).
func checkCount(read string, count, total int) error {
	if total >= 0 && total != count {
		return fmt.Errorf("%s: read %d items but the service reports %d; retry the command", read, count, total)
	}
	return nil
}

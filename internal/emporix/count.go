package emporix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Count returns the number of items in the collection at path that the
// service reports for a HEAD request, which reads no items, in X-Total-Count,
// as the Order Service counts orders. filter, an Emporix q filter, narrows the
// count as ListOptions.Query narrows a list, together with the q of query; the
// path holds no query, as a collection read's. A client made by WithBody
// counts with HEAD too, without its body. A response without a valid total
// fails.
func (c *Client) Count(ctx context.Context, path string, query url.Values, filter string) (int, error) {
	head := c.sending(http.MethodHead, Body{})
	query, err := head.collectionQuery(path, query, ListOptions{Query: filter})
	if err != nil {
		return 0, err
	}
	count := totalCounting()
	result, err := head.get(ctx, path, query, count.header)
	if err != nil {
		return 0, err
	}
	total := count.total(result.header)
	if total < 0 {
		return 0, fmt.Errorf("%s: the service reported no count in %s", head.named(path), count.from)
	}
	return total, nil
}

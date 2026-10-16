package emporix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// EncodeBody returns v as the JSON body of a read (WithBody) or a write: as
// json.Marshal encodes it, but with <, > and & left as given, as in a q
// filter's price:>10, where json.Marshal escapes them for HTML.
func EncodeBody(v any) (json.RawMessage, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(body.Bytes(), []byte("\n")), nil
}

// GetAllBodyPaged is GetAll for a search that pages with pageNumber and
// pageSize in its JSON body instead of its query, as the Fee Service's
// searches of item fees by product do: c is a client made by WithBody whose
// body is a JSON object, sent on every page with those two members set
// (replacing any it holds; its other members as given, keys sorted) and the
// query as given. Paging is checked as GetAll checks it, the total asked for
// and read as GetAll does, with one token for every page. A client without a
// body, or with a body that is not a JSON object, is ErrInvalidInput, and
// nothing is sent.
func (c *Client) GetAllBodyPaged(ctx context.Context, path string, query url.Values, opts ListOptions) (
	json.RawMessage, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(c.body.data, &members); err != nil || members == nil {
		return nil, fmt.Errorf("%w: %s: a search paged in its body needs a JSON object as its body",
			ErrInvalidInput, c.named(path))
	}
	// One client sends every page with that page's body; it requests the token on page 1 and keeps it.
	sending := *c
	return c.getAll(ctx, path, query, opts, func(ctx context.Context, path string, query url.Values, page, size int) (
		[]json.RawMessage, int, error) {
		members["pageNumber"] = json.RawMessage(strconv.Itoa(page))
		members["pageSize"] = json.RawMessage(strconv.Itoa(size))
		body, err := JSONBodyOf(members)
		if err != nil {
			return nil, 0, err
		}
		sending.body = body
		return sending.arrayPage(ctx, path, query, page, totalCounting())
	})
}

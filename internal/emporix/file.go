package emporix

import (
	"cmp"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
)

// GetFile reads path as Get does, for a file that Emporix answers with
// instead of JSON, such as the PDF of a quote: it accepts media, the file's
// media type, and JSON, in which Emporix explains a failure, and asks for the
// file in language, or, where it is empty, sends no Accept-Language (the
// client's, such as *, asks for every translation of JSON's localized text;
// without one, Emporix answers in the tenant's default language). It returns
// the file as sent. A language that cannot be sent as a header value
// (CheckHeaderValue) is ErrInvalidInput, and an empty body ErrNotFound, as
// for Get; a success of another media type, parameters aside, fails without
// a class.
func (c *Client) GetFile(ctx context.Context, path string, query url.Values, media, language string) ([]byte, error) {
	header := http.Header{"Accept": {media + ", " + JSONContentType}, "Accept-Language": nil}
	if language != "" {
		header.Set("Accept-Language", language)
	}
	// The query is left out of every message: it may hold values a log must not keep.
	if err := checkHeaders(c.named(withoutQuery(path)), header); err != nil {
		return nil, err
	}
	result, err := c.getNonEmpty(ctx, path, query, header)
	if err != nil {
		return nil, err
	}
	// A malformed parameter leaves the media type, which is all that matters here.
	answered, _, _ := mime.ParseMediaType(result.header.Get("Content-Type"))
	if answered != media {
		return nil, fmt.Errorf("%s: the response is %s, not the %s that it asked for", c.named(result.path),
			cmp.Or(answered, "of no type"), media)
	}
	return result.body, nil
}

package emporix

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
)

// WithHeader returns a copy of c whose reads also send the header name with
// value, such as emporix-site, the site whose overlays the Product and
// Schema Services merge into site-aware mixins; a header of a read's own of
// the same name, such as X-Total-Count, replaces it. The copy keeps c's
// credential and scopes (and its token, once it has one), and c sends no
// more than before. A value that cannot be sent (CheckHeaderValue) is
// ErrInvalidInput. Writes (Prepare) send only the headers of their Write.
func (c *Client) WithHeader(name, value string) (*Client, error) {
	if err := CheckHeaderValue(value); err != nil {
		return nil, fmt.Errorf("%w: the header %s %w", ErrInvalidInput, name, err)
	}
	sending := *c
	sending.header = http.Header{}
	maps.Copy(sending.header, c.header)
	sending.header.Set(name, value)
	return &sending, nil
}

// checkHeaders returns ErrInvalidInput, naming the request named (a write,
// or a read of a file: GetFile), when a value of header cannot be sent
// (CheckHeaderValue); nil when all can.
func checkHeaders(named string, header http.Header) error {
	for name, values := range header {
		for _, value := range values {
			if err := CheckHeaderValue(value); err != nil {
				return fmt.Errorf("%w: %s: the header %s %w", ErrInvalidInput, named, name, err)
			}
		}
	}
	return nil
}

// CheckHeaderValue returns why value cannot be sent as the value of a header,
// which HTTP does not allow to hold a control character other than a tab,
// such as a line break; nil when it can be sent.
func CheckHeaderValue(value string) error {
	if strings.ContainsFunc(value, func(r rune) bool { return r < ' ' && r != '\t' || r == 0x7f }) {
		return errors.New("holds a control character")
	}
	return nil
}

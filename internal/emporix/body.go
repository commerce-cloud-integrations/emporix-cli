package emporix

import (
	"encoding/json"
	"fmt"
	"slices"
)

// JSONContentType is the Content-Type of a JSON body, and of a part of a
// form that holds JSON.
const JSONContentType = "application/json"

// Body is the body of a request as it is sent, with its Content-Type: one
// JSON value (JSONBody) or a multipart/form-data form (FormBody). A write
// sends it as the Body of its Write, a read with a client made by WithBody.
// The zero Body is none.
type Body struct {
	data        []byte
	contentType string
}

// JSONBody returns the body that sends data, one JSON value, byte for byte;
// anything else is ErrInvalidInput. It keeps a copy of data.
func JSONBody(data []byte) (Body, error) {
	if !json.Valid(data) {
		return Body{}, fmt.Errorf("%w: the request body is not JSON", ErrInvalidInput)
	}
	return Body{data: slices.Clone(data), contentType: JSONContentType}, nil
}

// JSONBodyOf returns the JSON body that sends v as EncodeBody encodes it.
func JSONBodyOf(v any) (Body, error) {
	data, err := EncodeBody(v)
	if err != nil {
		return Body{}, err
	}
	return Body{data: data, contentType: JSONContentType}, nil
}

// Len returns the size of b as sent, in bytes: of a form, its parts with
// their headers and boundaries.
func (b Body) Len() int { return len(b.data) }

// none reports whether b is no body, the zero Body.
func (b Body) none() bool { return b.data == nil }

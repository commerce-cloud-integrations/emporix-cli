package emporix

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
)

// Part is a part of a multipart/form-data body (FormBody), such as a file
// that a service stores with the JSON that describes it: a field, or a file
// when it names one.
type Part struct {
	Name        string // the form field, e.g. "file"
	FileName    string // the name of the file whose content it sends; empty for a field
	ContentType string // its content type, e.g. JSONContentType; empty: none sent
	Content     []byte // sent byte for byte
}

// FormBody returns the multipart/form-data body (a form) that sends parts,
// in their order, separated by a random boundary that its Content-Type
// names. A form without parts, or with a part that cannot be sent, is
// ErrInvalidInput.
func FormBody(parts []Part) (Body, error) {
	if len(parts) == 0 {
		return Body{}, fmt.Errorf("%w: a form without parts", ErrInvalidInput)
	}
	var encoded bytes.Buffer
	form := multipart.NewWriter(&encoded)
	for _, part := range parts {
		if err := checkPart(part); err != nil {
			return Body{}, err
		}
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", disposition(part))
		if part.ContentType != "" {
			header.Set("Content-Type", part.ContentType)
		}
		content, err := form.CreatePart(header)
		if err == nil {
			_, err = content.Write(part.Content)
		}
		if err != nil {
			return Body{}, err
		}
	}
	if err := form.Close(); err != nil {
		return Body{}, err
	}
	return Body{data: encoded.Bytes(), contentType: form.FormDataContentType()}, nil
}

func checkPart(part Part) error {
	switch {
	case part.Name == "" || strings.ContainsAny(part.Name, `"\`) || strings.ContainsFunc(part.Name, isControl):
		return fmt.Errorf("%w: invalid form part name %q", ErrInvalidInput, part.Name)
	case strings.ContainsFunc(part.FileName, isControl):
		return fmt.Errorf("%w: invalid file name %q: a header cannot carry its control characters", ErrInvalidInput,
			part.FileName)
	}
	return nil
}

// RFC 7578 rules out the filename* of RFC 5987.
func disposition(part Part) string {
	if part.FileName != "" {
		return multipart.FileContentDisposition(part.Name, part.FileName)
	}
	return `form-data; name="` + part.Name + `"`
}

func isControl(r rune) bool { return r < ' ' || r == 0x7f }

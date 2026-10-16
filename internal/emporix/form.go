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
// names. A form without parts, or with a part that cannot be sent
// (checkPart), is ErrInvalidInput.
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
		// Writes to a bytes.Buffer do not fail; an error would be the encoder's own.
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

// checkPart fails, as ErrInvalidInput, a part whose name is empty or holds a
// character that its header would have to escape, a quote, a backslash or a
// control character (a name is declared, never given), or whose file name
// holds a control character, which no header carries (a strict server
// rejects it). A file name, which comes from a path, is otherwise sent
// escaped as browsers escape it (disposition).
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

// disposition returns the Content-Disposition of part: a field named after
// it, with the name of its file if it sends one, quoted, with a quote or a
// backslash in it escaped (RFC 7578 rules out the filename* of RFC 5987).
func disposition(part Part) string {
	if part.FileName != "" {
		return multipart.FileContentDisposition(part.Name, part.FileName)
	}
	return `form-data; name="` + part.Name + `"`
}

// isControl reports whether r is a control character.
func isControl(r rune) bool { return r < ' ' || r == 0x7f }

package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"mime/multipart"
	"net/textproto"
	"slices"
)

// WriteRequest writes a request as a dry run shows it: the method and URL,
// the headers (sorted by name, each value on its own line), and the body as
// sent after an empty line, followed by a line break unless it ends with one;
// nothing after the headers for a nil body. The caller leaves out what must
// not be shown, such as an Authorization header. A multipart/form-data body,
// as its Content-Type names it, shows as a line for each part, followed by
// its content where it is JSON (formParts).
func WriteRequest(w io.Writer, method, url string, header map[string][]string, body []byte) error {
	var text bytes.Buffer
	fmt.Fprintf(&text, "%s %s\n", method, url)
	for _, name := range slices.Sorted(maps.Keys(header)) {
		for _, value := range header[name] {
			fmt.Fprintf(&text, "%s: %s\n", name, value)
		}
	}
	if body != nil {
		text.WriteByte('\n')
		if err := writeBody(&text, textproto.MIMEHeader(header).Get("Content-Type"), body); err != nil {
			return err
		}
	}
	_, err := w.Write(text.Bytes())
	return err
}

// writeBody writes body, whose Content-Type is contentType: a form as a line
// for each part (formParts), any other body, a JSON one among them, as sent
// (writeSent).
func writeBody(text *bytes.Buffer, contentType string, body []byte) error {
	if media, params, err := mime.ParseMediaType(contentType); err == nil && media == "multipart/form-data" {
		return formParts(text, multipart.NewReader(bytes.NewReader(body), params["boundary"]))
	}
	writeSent(text, body)
	return nil
}

// writeSent writes content as sent, followed by a line break unless it ends
// with one.
func writeSent(text *bytes.Buffer, content []byte) {
	text.Write(content)
	if !bytes.HasSuffix(content, []byte("\n")) {
		text.WriteByte('\n')
	}
}

// formParts writes a line that describes each part of form, in its order: its
// name, the name of its file and its content type where it has them, and the
// size of its content, e.g. `part "file": file "schema.json",
// application/json, 120 bytes`, followed by the content as sent where its
// content type is JSON and it is JSON, as a JSON body shows; the content of
// any other part, which may be binary, never shows, nor that of a file that is
// named as JSON but holds something else, such as terminal controls. Names are
// quoted, so that control characters show escaped. A form that cannot be read
// fails.
func formParts(text *bytes.Buffer, form *multipart.Reader) error {
	for {
		part, err := form.NextRawPart()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading the form: %w", err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			return fmt.Errorf("reading the form: %w", err)
		}
		fmt.Fprintf(text, "part %q:", part.FormName())
		if name := part.FileName(); name != "" {
			fmt.Fprintf(text, " file %q,", name)
		}
		contentType := part.Header.Get("Content-Type")
		if contentType != "" {
			fmt.Fprintf(text, " %s,", contentType)
		}
		fmt.Fprintf(text, " %d bytes\n", len(content))
		if media, _, err := mime.ParseMediaType(contentType); err == nil && media == "application/json" &&
			json.Valid(content) {
			writeSent(text, content)
		}
	}
}

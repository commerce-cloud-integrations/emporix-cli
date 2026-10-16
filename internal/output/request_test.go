package output

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

// TestWriteRequestShowsTheRequestAsSent checks the text of a request: the
// method and URL, the headers sorted by name with every value, and the body
// byte for byte after an empty line, ended by one line break; a request
// without a body ends after its headers.
func TestWriteRequestShowsTheRequestAsSent(t *testing.T) {
	header := map[string][]string{"User-Agent": {"epx"}, "Accept": {"application/json", "text/plain"}}
	for _, tt := range []struct {
		body []byte
		want string
	}{
		{body: []byte("{\"a\": 1}\n"), want: "\n{\"a\": 1}\n"},
		{body: []byte(`[1, 2]`), want: "\n[1, 2]\n"},
		{body: nil, want: ""},
	} {
		var out bytes.Buffer
		if err := WriteRequest(&out, "PUT", "https://api.example/x?a=1", header, tt.body); err != nil {
			t.Fatal(err)
		}
		want := "PUT https://api.example/x?a=1\nAccept: application/json\nAccept: text/plain\nUser-Agent: epx\n" + tt.want
		if out.String() != want {
			t.Errorf("body %q: wrote\n%s\nwant\n%s", tt.body, out.String(), want)
		}
	}
	if err := WriteRequest(failingWriter{}, "DELETE", "https://api.example/x", nil, nil); err == nil {
		t.Fatal("a failed write is no error")
	}
}

// form returns a multipart/form-data body with a JSON file part, whose name
// holds quotes and a letter that is not ASCII, a JSON field, a binary file
// and a plain field, and its Content-Type.
func form(t *testing.T) (body []byte, contentType string) {
	t.Helper()
	var encoded bytes.Buffer
	writer := multipart.NewWriter(&encoded)
	for _, part := range []struct {
		header  textproto.MIMEHeader
		content string
	}{
		{header: textproto.MIMEHeader{"Content-Disposition": {multipart.FileContentDisposition("file", `my "schéma".json`)},
			"Content-Type": {"application/json"}}, content: "{\"type\": \"object\"}\n"},
		{header: textproto.MIMEHeader{"Content-Disposition": {`form-data; name="body"`},
			"Content-Type": {"application/json; charset=utf-8"}}, content: `{"name":"A"}`},
		{header: textproto.MIMEHeader{"Content-Disposition": {multipart.FileContentDisposition("image", "a.png")},
			"Content-Type": {"image/png"}}, content: strings.Repeat("\x00", 120)},
		{header: textproto.MIMEHeader{"Content-Disposition": {`form-data; name="id"`}}, content: "s1"},
		{header: textproto.MIMEHeader{"Content-Disposition": {multipart.FileContentDisposition("upload", "a.json")},
			"Content-Type": {"application/json"}}, content: "\x1b]0;title\x07{"},
	} {
		w, err := writer.CreatePart(part.header)
		if err == nil {
			_, err = w.Write([]byte(part.content))
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes(), writer.FormDataContentType()
}

// TestWriteRequestDescribesTheParts checks the text of a request whose body
// is a form: as that of any request up to its headers, then, after an empty
// line, a line for each part, in order: its name, the name of its file and
// its content type where it has them, and its size, followed by its content
// as sent where it is JSON only (not a file named as JSON that holds terminal
// controls); names are quoted. A form that cannot be read fails, and writes
// nothing.
func TestWriteRequestDescribesTheParts(t *testing.T) {
	body, contentType := form(t)
	var out bytes.Buffer
	err := WriteRequest(&out, "POST", "https://api.example/x", map[string][]string{"Content-Type": {contentType}}, body)
	want := "POST https://api.example/x\nContent-Type: " + contentType + "\n\n" +
		`part "file": file "my \"schéma\".json", application/json, 19 bytes` + "\n" +
		`{"type": "object"}` + "\n" +
		`part "body": application/json; charset=utf-8, 12 bytes` + "\n" +
		`{"name":"A"}` + "\n" +
		`part "image": file "a.png", image/png, 120 bytes` + "\n" +
		`part "id": 2 bytes` + "\n" +
		`part "upload": file "a.json", application/json, 11 bytes` + "\n"
	if err != nil || out.String() != want {
		t.Fatalf("wrote\n%s\nwant\n%s (%v)", out.String(), want, err)
	}
	for _, tt := range []struct {
		contentType string
		body        []byte
		want        string
	}{
		{contentType: "multipart/form-data", body: body, want: "reading the form: multipart: boundary is empty"},
		// Cut within the content of the image.
		{contentType: contentType, body: body[:bytes.IndexByte(body, 0)+1], want: "reading the form: unexpected EOF"},
	} {
		var out bytes.Buffer
		header := map[string][]string{"Content-Type": {tt.contentType}}
		if err := WriteRequest(&out, "POST", "https://api.example/x", header, tt.body); err == nil ||
			err.Error() != tt.want || out.Len() != 0 {
			t.Errorf("%s: wrote %q, error %v; want %q", tt.contentType, out.String(), err, tt.want)
		}
	}
}

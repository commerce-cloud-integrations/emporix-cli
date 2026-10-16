package emporix

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// schemaPart is a part of a form that the tests send: a file, with its name.
func schemaPart() Part {
	return Part{Name: "file", FileName: `my "schema".json`, ContentType: "application/json",
		Content: []byte(`{"type":"object"}`)}
}

// bodyPart is a part of a form that the tests send: a JSON field.
func bodyPart() Part {
	return Part{Name: "body", ContentType: "application/json", Content: []byte(`{"name":{"en":"Specs"}}`)}
}

// receivedForm returns the parts of the form that req sent with body, as
// its Content-Type names its boundary.
func receivedForm(t *testing.T, req *http.Request, body string) []Part {
	t.Helper()
	media, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || media != "multipart/form-data" {
		t.Fatalf("Content-Type %q; want a form", req.Header.Get("Content-Type"))
	}
	form := multipart.NewReader(strings.NewReader(body), params["boundary"])
	var parts []Part
	for {
		part, err := form.NextPart()
		if errors.Is(err, io.EOF) {
			return parts
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, Part{Name: part.FormName(), FileName: part.FileName(),
			ContentType: part.Header.Get("Content-Type"), Content: content})
	}
}

// equalParts reports whether got and want hold the same parts, in order.
func equalParts(got, want []Part) bool {
	return slices.EqualFunc(got, want, func(a, b Part) bool {
		return a.Name == b.Name && a.FileName == b.FileName && a.ContentType == b.ContentType &&
			bytes.Equal(a.Content, b.Content)
	})
}

// formBody returns the form of parts (FormBody).
func formBody(t *testing.T, parts ...Part) Body {
	t.Helper()
	body, err := FormBody(parts)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// TestSendWritesAForm checks a write with a form: it is sent with a
// Content-Type that names the boundary of its parts, each with its name, the
// name of its file, escaped, and its content type, in their order, as the
// request that a dry run shows holds them, whose size Len gives.
func TestSendWritesAForm(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/references", respond(http.StatusCreated, `{"id":"r1"}`))
	client := fake.client(t)
	want := []Part{schemaPart(), bodyPart(), {Name: "id", Content: []byte("r1")}}
	form := formBody(t, want...)
	req, err := client.Prepare(&Write{Method: http.MethodPost, Path: "/schema/{tenant}/references", Body: form})
	if err != nil {
		t.Fatal(err)
	}
	body, err := client.Send(context.Background(), req)
	if err != nil || string(body) != `{"id":"r1"}` {
		t.Fatalf("body %s, error %v", body, err)
	}
	sent := receivedForm(t, fake.requests[0], fake.bodies[0])
	if !equalParts(sent, want) || string(req.Body()) != fake.bodies[0] || form.Len() != len(fake.bodies[0]) ||
		req.Header().Get("Content-Type") != fake.requests[0].Header.Get("Content-Type") {
		t.Fatalf("sent %q; want %q", sent, want)
	}
	if !strings.Contains(fake.bodies[0], `Content-Disposition: form-data; name="file"; filename="my \"schema\".json"`) {
		t.Fatalf("form %q; want the file name escaped", fake.bodies[0])
	}
}

// TestWithBodyReadsWithAForm checks a read with a form: it is sent as POST
// with the form, retried as a read is, and the client it was made from still
// reads with GET.
func TestWithBodyReadsWithAForm(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/schemas/file", answers(http.StatusServiceUnavailable, http.StatusOK))
	client := fake.client(t)
	parsing, err := client.WithBody(http.MethodPost, formBody(t, schemaPart()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parsing.Get(context.Background(), "/schema/{tenant}/schemas/file", nil); err != nil {
		t.Fatal(err)
	}
	for i, req := range fake.requests {
		if req.Method != http.MethodPost || !equalParts(receivedForm(t, req, fake.bodies[i]), []Part{schemaPart()}) {
			t.Fatalf("request %d: %s with %q", i+1, req.Method, fake.bodies[i])
		}
	}
	if len(fake.requests) != 2 || fake.bodies[0] != fake.bodies[1] {
		t.Fatalf("%d requests; want the form sent again after a 503", len(fake.requests))
	}
	fake.route("/x", respond(http.StatusOK, `{"id":"x"}`))
	if _, err := client.Get(context.Background(), "/x", nil); err != nil || fake.requests[2].Method != http.MethodGet ||
		fake.bodies[2] != "" || fake.requests[2].Header.Get("Content-Type") != "" {
		t.Fatalf("after a read with a form, the client sent %s with %q (%v)", fake.requests[2].Method, fake.bodies[2], err)
	}
}

// TestFormsThatCannotBeSent checks that a form without parts, or with a part
// whose name is empty or one that a header would have to escape, or whose
// file name holds a control character, is ErrInvalidInput.
func TestFormsThatCannotBeSent(t *testing.T) {
	for _, tt := range []struct {
		parts []Part
		want  string
	}{
		{parts: nil, want: "a form without parts"},
		{parts: []Part{{Content: []byte("a")}}, want: `invalid form part name ""`},
		{parts: []Part{schemaPart(), {Name: "a\r\nX-Injected: 1"}}, want: `invalid form part name "a\r\nX-Injected: 1"`},
		{parts: []Part{{Name: `a"b`}}, want: `invalid form part name "a\"b"`},
		{parts: []Part{{Name: `a\b`}}, want: "invalid form part name"},
		{parts: []Part{{Name: "a\x7f"}}, want: "invalid form part name"},
		{parts: []Part{{Name: "file", FileName: "a\r\nX-Injected: 1.json"}},
			want: `invalid file name "a\r\nX-Injected: 1.json": a header cannot carry its control characters`},
		{parts: []Part{{Name: "file", FileName: "schema\x1b.json"}}, want: `invalid file name "schema\x1b.json"`},
	} {
		if body, err := FormBody(tt.parts); !errors.Is(err, ErrInvalidInput) || !body.none() ||
			!strings.Contains(err.Error(), tt.want) {
			t.Errorf("%q: error %v; want invalid input %q", tt.parts, err, tt.want)
		}
	}
}

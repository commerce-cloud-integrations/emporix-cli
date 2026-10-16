package emporix

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// filePDF stands in for a file that a read accepts.
const filePDF = "%PDF-1.7\n%%EOF\n"

// pdfPath is the path of a quote's PDF, which Emporix serves as POST.
const pdfPath = "/quote/{tenant}/quotes/Q1/pdf"

// TestGetFileReadsAFile checks a read of a file (GetFile) sent as POST
// without a body (Posting): it asks for the file's type and JSON, in the
// language given or else without the client's Accept-Language, with a token
// of the client's scopes, and returns the file as sent where the response's
// Content-Type, parameters aside, is the file's type.
func TestGetFileReadsAFile(t *testing.T) {
	for _, language := range []string{"", "de"} {
		fake := newFakeEmporix(t)
		fake.route("/quote/acme1/quotes/Q1/pdf", respond(http.StatusOK, filePDF, "Content-Type",
			"Application/PDF; qs=0.9; =x"))
		client := fake.client(t, "site.site_read")
		body, err := client.Posting().GetFile(context.Background(), pdfPath, nil, "application/pdf", language)
		if err != nil || string(body) != filePDF {
			t.Fatalf("%q: body %q, error %v; want the file", language, body, err)
		}
		req := fake.requests[0]
		if req.Method != http.MethodPost || fake.bodies[0] != "" || req.Header.Get("Content-Type") != "" {
			t.Errorf("%q: sent %s with body %q, Content-Type %q; want POST without a body", language, req.Method,
				fake.bodies[0], req.Header.Get("Content-Type"))
		}
		if accept, sent := req.Header.Values("Accept"), req.Header.Values("Accept-Language"); len(accept) != 1 ||
			accept[0] != "application/pdf, application/json" || strings.Join(sent, ",") != language {
			t.Errorf("Accept %q, Accept-Language %q; want the file and JSON, in %q", accept, sent, language)
		}
		if scope := fake.tokenForm.Get("scope"); scope != "tenant=acme1 site.site_read" {
			t.Errorf("%q: token scope %q", language, scope)
		}
	}
}

// TestGetFileFails checks the failures of a read of a file: a success of
// another type, or none, fails without a class; an empty success is not
// found, as for Get; an error status is classified as a read's, and a
// server error is sent again, as the read changes nothing.
func TestGetFileFails(t *testing.T) {
	const named = "POST /quote/acme1/quotes/Q1/pdf: "
	for _, tt := range []struct {
		answer func(http.ResponseWriter, *http.Request)
		class  error
		want   string
		sent   int // the requests sent
	}{
		{answer: respond(http.StatusOK, `{"url":"u"}`, "Content-Type", "application/json"),
			want: named + "the response is application/json, not the application/pdf that it asked for", sent: 1},
		{answer: respond(http.StatusOK, "<html>busy</html>"), want: "the response is text/html", sent: 1},
		{answer: respond(http.StatusOK, filePDF, "Content-Type", "application/x-pdf"),
			want: "the response is application/x-pdf, not", sent: 1},
		{answer: respond(http.StatusOK, filePDF, "Content-Type", ";"), want: "the response is of no type", sent: 1},
		{answer: respond(http.StatusNoContent, ""), class: ErrNotFound, want: named + "empty response", sent: 1},
		{answer: respond(http.StatusOK, " ", "Content-Type", "application/pdf"), class: ErrNotFound,
			want: "empty response", sent: 1},
		{answer: respond(http.StatusNotFound, `{"message":"Quote not found"}`), class: ErrNotFound,
			want: "Quote not found", sent: 1},
		{answer: respond(http.StatusBadRequest, `{"message":"bad"}`), want: "bad", sent: 1},
		{answer: respond(http.StatusServiceUnavailable, `{"message":"busy"}`), want: "busy", sent: 3},
	} {
		fake := newFakeEmporix(t)
		fake.route("/quote/acme1/quotes/Q1/pdf", tt.answer)
		body, err := fake.client(t).Posting().GetFile(context.Background(), pdfPath, nil, "application/pdf", "")
		if body != nil || err == nil || !strings.Contains(err.Error(), tt.want) || len(fake.requests) != tt.sent {
			t.Errorf("body %q, error %v, %d requests; want %q after %d", body, err, len(fake.requests), tt.want,
				tt.sent)
		}
		for _, class := range []error{ErrInvalidInput, ErrAuth, ErrNotFound, ErrConflict} {
			if errors.Is(err, class) != errors.Is(tt.class, class) {
				t.Errorf("%v: is %v %v; want class %v", err, class, errors.Is(err, class), tt.class)
			}
		}
	}
}

// TestGetFileRefusesBeforeSending checks that a read of a file refuses,
// before it requests a token or sends anything, a language that cannot be
// sent as a header value, and, sent as POST, a path with a segment that might
// address another operation once decoded.
func TestGetFileRefusesBeforeSending(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t).Posting()
	for path, language := range map[string]string{
		pdfPath:                         "de\r\nX-Other: 1",
		"/quote/{tenant}/quotes/a%2F..": "",
	} {
		_, err := client.GetFile(context.Background(), path, nil, "application/pdf", language)
		if !errors.Is(err, ErrInvalidInput) || !strings.HasPrefix(err.Error(), "invalid input: POST /quote/") {
			t.Errorf("%s in %q: error %v; want invalid input naming the read", path, language, err)
		}
	}
	if len(fake.requests) != 0 || fake.tokenCalls != 0 {
		t.Fatalf("requests %d, token calls %d; want none", len(fake.requests), fake.tokenCalls)
	}
}

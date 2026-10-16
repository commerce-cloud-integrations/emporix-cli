package emporix

import (
	"errors"
	"net/http"
	"testing"
)

// TestWriteHeadersAreSentAndShown checks that the headers of a write are
// sent and shown with every request's, replacing one of the same name.
func TestWriteHeadersAreSentAndShown(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/unit-handling/acme1/units", respond(http.StatusCreated, `{"code":"g"}`))
	client := fake.client(t)
	req, err := client.Prepare(&Write{
		Method: http.MethodPost, Path: "/unit-handling/{tenant}/units", Body: jsonBody(t, `{"code":"g"}`),
		Header: http.Header{"Content-Language": {"*"}, "Accept-Language": {"de"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := http.Header{
		"Accept": {"application/json"}, "Accept-Language": {"de"}, "Content-Language": {"*"},
		"Content-Type": {"application/json"}, "User-Agent": {"emporix-cli/test"},
	}
	if !equalHeaders(req.Header(), want) {
		t.Fatalf("headers %v; want %v", req.Header(), want)
	}
	if _, err := client.Send(t.Context(), req); err != nil {
		t.Fatal(err)
	}
	if sent := fake.requests[0].Header; sent.Get("Content-Language") != "*" || sent.Get("Accept-Language") != "de" {
		t.Fatalf("sent headers %v", sent)
	}
}

// TestStaleVersionsAnsweredNotFoundAreConflicts checks that where a service
// answers a stale metadata.version with 404 Not Found, a 404 to a body that
// carries a version is ErrConflict, and any other 404 stays ErrNotFound, as
// it does for every other write.
func TestStaleVersionsAnsweredNotFoundAreConflicts(t *testing.T) {
	for _, tt := range []struct {
		body  string
		stale bool // Write.StaleNotFound
		class error
	}{
		{body: `{"code":"g","metadata":{"version":3}}`, stale: true, class: ErrConflict},
		{body: `{"metadata":{"version":"3","createdAt":"2026-01-01T00:00:00Z"}}`, stale: true, class: ErrConflict},
		{body: `{"code":"g","metadata":{"version":3}}`, class: ErrNotFound},
		{body: `{"code":"g"}`, stale: true, class: ErrNotFound},
		{body: `{"code":"g","metadata":{"version":null}}`, stale: true, class: ErrNotFound},
		{body: `{"code":"g","metadata":null}`, stale: true, class: ErrNotFound},
		{body: `{"code":"g","metadata":{"createdAt":"2026-01-01T00:00:00Z"}}`, stale: true, class: ErrNotFound},
		{body: `{"code":"g","Metadata":{"Version":3}}`, stale: true, class: ErrNotFound},
		{body: `{"code":"g","metadata":[{"version":3}]}`, stale: true, class: ErrNotFound},
		{body: `[{"metadata":{"version":3}}]`, stale: true, class: ErrNotFound},
	} {
		fake := newFakeEmporix(t)
		fake.route("/unit-handling/acme1/units/g", respond(http.StatusNotFound, `{"message":"Cannot find the unit"}`))
		_, err := send(fake.client(t), &Write{
			Method: http.MethodPut, Path: "/unit-handling/{tenant}/units/g", Body: jsonBody(t, tt.body),
			Answers: Answers{StaleNotFound: tt.stale},
		})
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound ||
			err.Error() != "PUT /unit-handling/acme1/units/g: 404 Not Found: Cannot find the unit" {
			t.Errorf("%s: error %v; want the 404 as sent", tt.body, err)
		}
		for _, class := range []error{ErrConflict, ErrNotFound} {
			if want := errors.Is(tt.class, class); errors.Is(err, class) != want {
				t.Errorf("%s (stale %v): errors.Is(err, %v) = %v, want %v", tt.body, tt.stale, class, !want, want)
			}
		}
	}
	// A conflict stays one, and a write without a body is no stale version.
	for status, class := range map[int]error{http.StatusConflict: ErrConflict, http.StatusNotFound: ErrNotFound} {
		fake := newFakeEmporix(t)
		fake.route("/unit-handling/acme1/units/g", respond(status, `{}`))
		_, err := send(fake.client(t), &Write{Method: http.MethodDelete, Path: "/unit-handling/{tenant}/units/g",
			Answers: Answers{StaleNotFound: true}})
		if !errors.Is(err, class) {
			t.Errorf("%d: error %v; want %v", status, err, class)
		}
	}
}

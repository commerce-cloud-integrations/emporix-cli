package emporix

import (
	"errors"
	"strings"
	"testing"
)

// jsonBody returns the body that sends text, one JSON value (JSONBody).
func jsonBody(t *testing.T, text string) Body {
	t.Helper()
	body, err := JSONBody([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// TestJSONBodyTakesOneJSONValue checks that a JSON body holds one JSON value,
// byte for byte, a copy of what it was given, with its content type; anything
// else is ErrInvalidInput, and the zero Body is none.
func TestJSONBodyTakesOneJSONValue(t *testing.T) {
	data := []byte("{ \"code\": \"box\" }\n")
	body, err := JSONBody(data)
	data[0] = '['
	if err != nil || string(body.data) != "{ \"code\": \"box\" }\n" || body.contentType != JSONContentType ||
		body.none() {
		t.Fatalf("body %q (%s), error %v", body.data, body.contentType, err)
	}
	for _, text := range []string{`{"a":`, "", `{} {}`, "name=A"} {
		if body, err := JSONBody([]byte(text)); !errors.Is(err, ErrInvalidInput) || !body.none() ||
			!strings.Contains(err.Error(), "the request body is not JSON") {
			t.Errorf("%q: body %q, error %v; want invalid input", text, body.data, err)
		}
	}
	if !(Body{}).none() {
		t.Fatal("the zero Body is a body")
	}
}

// TestJSONBodyOfEncodesAValue checks that the JSON body of a value is encoded
// as EncodeBody encodes it, with its content type, and that a value that JSON
// cannot encode is no body.
func TestJSONBodyOfEncodesAValue(t *testing.T) {
	body, err := JSONBodyOf(map[string]string{"q": "price:>10"})
	if err != nil || string(body.data) != `{"q":"price:>10"}` || body.contentType != JSONContentType {
		t.Fatalf("body %q (%s), error %v", body.data, body.contentType, err)
	}
	if body, err := JSONBodyOf(make(chan int)); err == nil || !body.none() {
		t.Fatalf("body %q, error %v; want an error and no body", body.data, err)
	}
}

package cli

import "testing"

func TestUploadContentTypesAreTheSameEverywhere(t *testing.T) {
	for _, tt := range []struct {
		name, content, want string
	}{
		{name: "order.csv", content: "product,quantity\n", want: "text/csv"},
		{name: "ORDER.TXT", content: "hello", want: "text/plain"},
		{name: "order.pdf", content: "%PDF-1.7\n", want: "application/pdf"},
		{name: "order.text", content: "product 100200\n", want: "text/plain"},
		{name: "ORDER.Text", content: "product 100200\n", want: "text/plain"},
		{name: "order", content: "\x00\x01\x02", want: "application/octet-stream"},
	} {
		if got, err := contentTypeOf(tt.name, []byte(tt.content)); err != nil || got != tt.want {
			t.Errorf("%s: %q, %v; want %q", tt.name, got, err, tt.want)
		}
	}
}

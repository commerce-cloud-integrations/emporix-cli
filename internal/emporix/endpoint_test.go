package emporix

import (
	"errors"
	"testing"
)

// TestCheckEndpointNeedsNoCredential checks the API URL and tenant alone, as
// a profile is checked before its client secret is read.
func TestCheckEndpointNeedsNoCredential(t *testing.T) {
	if err := CheckEndpoint(DefaultBaseURL, "acme1"); err != nil {
		t.Fatalf("valid endpoint: %v", err)
	}
	for _, endpoint := range [][2]string{{"http://api.example.com", "acme1"}, {DefaultBaseURL, "a/b"}, {"x", "acme1"}} {
		if err := CheckEndpoint(endpoint[0], endpoint[1]); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q: %v; want ErrInvalidInput", endpoint, err)
		}
	}
}

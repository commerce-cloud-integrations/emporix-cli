package cli

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestProductsGetJSONKeepsEveryTranslation(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products/001234", allLanguages(respond(http.StatusOK,
		`{"id":"001234","code":"001234","name":{"nl":"Transistor","fr":"Transistors"}}`)))
	code, stdout, _ := runCLI(t, fake.env(), "products", "get", "001234", "-o", "json")
	var got struct {
		Name map[string]string `json:"name"`
	}
	if code != ExitOK || json.Unmarshal([]byte(stdout), &got) != nil ||
		got.Name["nl"] != "Transistor" || got.Name["fr"] != "Transistors" {
		t.Fatalf("exit %d, stdout %q", code, stdout)
	}
}

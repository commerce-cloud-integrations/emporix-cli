package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	algoliaWriteKey  = "algolia-write-sentinel"
	algoliaSearchKey = "algolia-search-sentinel"
)

// providersBody is served for every page, as the service ignores paging.
const providersBody = `[{"provider":"ALGOLIA","active":false,"applicationId":"APP1","indexName":"acme1",` +
	`"searchKey":"` + algoliaSearchKey + `","writeKey":"` + algoliaWriteKey + `","excludedMixinKeys":["x"]}]`

// assertNoAlgoliaKey checks that tables leave the Algolia keys out.
func assertNoAlgoliaKey(t *testing.T, outputs ...string) {
	t.Helper()
	for _, text := range outputs {
		if strings.Contains(text, algoliaWriteKey) || strings.Contains(text, algoliaSearchKey) {
			t.Fatalf("table shows an Algolia key: %q", text)
		}
	}
}

// TestIndexingProvidersShowKeysOnlyInJSON checks that tables show their
// columns only, so the Algolia keys stay out of them, for the tenant's
// providers and a storefront's view, while -o json prints the keys as sent.
func TestIndexingProvidersShowKeysOnlyInJSON(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/indexing/acme1/configurations", respond(http.StatusOK, providersBody))
	fake.route("/indexing/acme1/public/configurations", respond(http.StatusOK,
		`[{"provider":"ALGOLIA","applicationId":"APP1","indexName":"acme1","searchKey":"`+algoliaSearchKey+`"}]`))
	for _, args := range [][]string{{"providers", "list"}, {"providers", "public", "list"}} {
		code, stdout, stderr := runCLI(t, fake.env(), append([]string{"indexing"}, args...)...)
		if code != ExitOK {
			t.Fatalf("%v: exit %d, stderr %q", args, code, stderr)
		}
		assertNoAlgoliaKey(t, stdout, stderr)
	}
	code, stdout, _ := runCLI(t, fake.env(), "indexing", "providers", "list", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"writeKey": "`+algoliaWriteKey+`"`) ||
		!strings.Contains(stdout, `"searchKey": "`+algoliaSearchKey+`"`) || !strings.Contains(stdout, `"excludedMixinKeys"`) {
		t.Fatalf("exit %d, stdout %s", code, stdout)
	}
}

func TestIndexingUnpagedListRejectsMalformedResponses(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, body := range []string{`null`, `{"provider":"ALGOLIA"}`} {
		fake.route("/indexing/acme1/public/configurations", respond(http.StatusOK, body))
		for _, format := range []string{"table", "json"} {
			code, stdout, stderr := runCLI(t, fake.env(), "indexing", "providers", "public", "list", "-o", format)
			if code != ExitError || stdout != "" || !strings.Contains(stderr, "expected a JSON array") {
				t.Errorf("%s as %s: exit %d, stdout %q, stderr %q", body, format, code, stdout, stderr)
			}
		}
	}
}

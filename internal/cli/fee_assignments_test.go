package cli

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestFeeAssignmentSearchSendsTheSiteWithABodyGiven checks that --body
// replaces the body that --item and --site make, while --site, required,
// still goes in the query.
func TestFeeAssignmentSearchSendsTheSiteWithABodyGiven(t *testing.T) {
	const given = `{"itemYrns":["` + paymentYRN + `"],"siteCode":"AcmeStore_NL"}`
	fake := newFakeEmporix(t)
	fake.route("/fee/acme1/itemFees/search", respond(http.StatusOK, `[]`))
	code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader(given),
		"fees", "assignments", "search", "--body", "-", "--site", "AcmeStore_NL")
	if code != ExitOK || stdout != "ITEM  ID  SITE  FEES\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertRequests(t, "/fee/acme1/itemFees/search?siteCode=AcmeStore_NL")
	fake.assertSent(t, http.MethodPost, given)
}

// TestFeeAssignmentSearchesPageTheBodyGiven checks that a search paged in
// its body pages a body given with --body too, replacing its pageNumber and
// pageSize: a page of the --limit, read once.
func TestFeeAssignmentSearchesPageTheBodyGiven(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/fee/acme1/itemFees/searchByProductId", firstBodyPage(`[{"id":"if1","itemYrn":"i1","feeIds":["f1"]}]`))
	code, stdout, stderr := runWithInput(t, fake.env(),
		strings.NewReader(`{"productId":"100200","siteCodes":["main"],"pageNumber":3,"pageSize":5}`),
		"fees", "assignments", "search-by-product", "--body", "-", "--limit", "1")
	if code != ExitOK || stdout != "ITEM  ID   SITE  FEES\ni1    if1        f1\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertSent(t, http.MethodPost, `{"pageNumber":1,"pageSize":1,"productId":"100200","siteCodes":["main"]}`)
}

// TestFeeAssignmentFailures checks the failures of the item-fee reads: usage
// errors that send nothing (exit 2), among them every member that Emporix
// requires in a search's body, missing, and a site that cannot be one; and a
// body given with --body that Emporix rejects (exit 1).
func TestFeeAssignmentFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args  []string
		stdin string
		want  string
	}{
		{args: []string{"list"}, want: `required flag(s) "site" not set`},
		{args: []string{"search", "--item", productYRN}, want: `required flag(s) "site" not set`},
		{args: []string{"search", "--site", "main"}, want: "[body item] is required"},
		{args: []string{"search", "--site", "main", "--body", "-", "--item", productYRN}, stdin: "{}",
			want: "[body item] are set none of the others"},
		{args: []string{"search-by-product", "--site", "main"}, want: "[body product] is required"},
		{args: []string{"search-by-product", "--product", "100200"}, want: "[body site] is required"},
		{args: []string{"search-by-product", "--product", "100200", "--site", "main", "--site", ".."},
			want: `invalid site ".."`},
		{args: []string{"search-by-product", "--body", "-", "--product", "100200"}, stdin: "{}",
			want: "[body product] are set none of the others"},
		{args: []string{"search-by-products", "--site", "main"}, want: "[body products] is required"},
		{args: []string{"search-by-products", "--products", "100200"}, want: "[body site] is required"},
		{args: []string{"search-by-products", "--products", "100200", "--site", ""}, want: `invalid site ""`},
		{args: []string{"search-by-products", "--body", "-", "--site", "main"}, stdin: "{}",
			want: "[body site] are set none of the others"},
		// A body that is no JSON object cannot be paged.
		{args: []string{"search-by-products", "--body", "-"}, stdin: `["100200"]`,
			want: "POST /fee/{tenant}/itemFees/searchByProductIds: a search paged in its body needs a JSON object"},
	} {
		code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader(tt.stdin),
			slices.Concat([]string{"fees", "assignments"}, tt.args)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("tokens %q, requests %q; want none", fake.tokenScopes, fake.gets)
	}
	// Emporix rejects a search without a product (seen live).
	fake.route("/fee/acme1/itemFees/searchByProductIds", respond(http.StatusBadRequest,
		`{"status":400,"message":"Validation failed","type":"validation_violation"}`))
	code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader(`{"siteCode":"main"}`),
		"fees", "assignments", "search-by-products", "--body", "-")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "400 Bad Request: Validation failed") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertSent(t, http.MethodPost, `{"pageNumber":1,"pageSize":100,"siteCode":"main"}`)
}

// TestSiteFieldNeedsTheSiteOfItsRequest checks that a body member that
// repeats the site of its request fails as the command is built, not as it
// runs, when the request names no site.
func TestSiteFieldNeedsTheSiteOfItsRequest(t *testing.T) {
	defer func() {
		if recovered := recover(); !strings.Contains(fmt.Sprint(recovered),
			`search: the body member "siteCode" repeats a site that its request does not name`) {
			t.Fatalf("recovered %v; want the misdeclared search named", recovered)
		}
	}()
	reader[thing]{view: thingView()}.command(&app{}, &command{
		use: "search", req: request{path: "/things/search"}, fetch: listing{}.search(siteField(siteCodeParam)),
	})
}

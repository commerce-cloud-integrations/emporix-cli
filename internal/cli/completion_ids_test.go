package cli

import (
	"bytes"
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// completionDone is what Cobra writes to standard error after every
// completion, which shells discard; anything else there would reach the terminal.
const completionDone = "Completion ended with directive: ShellCompDirectiveNoFileComp\n"

// completionEnded returns what Cobra writes to standard error after a
// completion with directive, as completionDone: NoFileComp, with NoSpace too
// where another ID may follow; none for a directive that epx never gives.
func completionEnded(directive cobra.ShellCompDirective) string {
	if directive == cobra.ShellCompDirectiveNoFileComp|cobra.ShellCompDirectiveNoSpace {
		return "Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp\n"
	}
	if directive == cobra.ShellCompDirectiveNoFileComp {
		return completionDone
	}
	return ""
}

// completionEnv returns the environment of f with completion from the API
// switched on.
func completionEnv(f *fakeEmporix) map[string]string {
	env := f.env()
	env[envCompleteIDs] = "1"
	return env
}

// completion runs `epx __complete args` in env and returns the completions
// it offers and how shells complete them, checking that it succeeds, never
// offers file names and prints nothing else.
func completion(t *testing.T, env map[string]string, args ...string) ([]string, cobra.ShellCompDirective) {
	t.Helper()
	code, stdout, stderr := runCLI(t, env, append([]string{"__complete"}, args...)...)
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	last, isDirective := strings.CutPrefix(lines[len(lines)-1], ":")
	directive, err := strconv.Atoi(last)
	ended := completionEnded(cobra.ShellCompDirective(directive))
	if code != ExitOK || !isDirective || err != nil || ended == "" || stderr != ended {
		t.Fatalf("%v: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
	}
	assertNoSecret(t, stdout, stderr)
	var got []string
	for _, line := range lines[:len(lines)-1] {
		got = append(got, strings.ReplaceAll(line, "\t", "|"))
	}
	return got, cobra.ShellCompDirective(directive)
}

// completed returns the completions that `epx __complete args` offers in env,
// as completion does, checking that shells complete them as file names never
// are, and nothing else.
func completed(t *testing.T, env map[string]string, args ...string) []string {
	t.Helper()
	got, directive := completion(t, env, args...)
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%v: directive %d; want %d", args, directive, cobra.ShellCompDirectiveNoFileComp)
	}
	return got
}

// completionCase is a completion that reads the API: the routes it is
// served, what it offers and how, and the requests and token scopes it sends.
type completionCase struct {
	args      []string
	routes    map[string]http.HandlerFunc
	want      []string                 // each ID|NAME, or ID alone
	directive cobra.ShellCompDirective // besides NoFileComp: NoSpace where another ID may follow
	requests  []string                 // request URIs
	scopes    []string                 // besides the tenant
}

func (tt *completionCase) check(t *testing.T) {
	t.Helper()
	f := newFakeEmporix(t)
	for path, handler := range tt.routes {
		f.route(path, handler)
	}
	got, directive := completion(t, completionEnv(f), tt.args...)
	if !slices.Equal(got, tt.want) || directive != cobra.ShellCompDirectiveNoFileComp|tt.directive {
		t.Fatalf("%v: completions %q, directive %d; want %q, %d", tt.args, got, directive, tt.want,
			cobra.ShellCompDirectiveNoFileComp|tt.directive)
	}
	f.assertRequests(t, tt.requests...)
	f.assertScopes(t, tt.scopes...)
}

const (
	productPage = `[{"id":"100200","code":"SHIRT","name":{"en":"Shirt"}},{"id":"200300","name":{"de":"Hose"}}]`
	sitePage    = `[{"code":"AcmeStore_NL","name":"Acme NL"},{"code":"AcmeStore_DE","name":"Acme DE"}]`
)

func TestCompletionOffersTheIDsOfOnePageOfTheList(t *testing.T) {
	products := "/product/acme1/products?pageNumber=1&pageSize=100"
	for _, tt := range []completionCase{
		{ // get: the resource's own list, described by NAME in the default language
			args:     []string{"products", "get", ""},
			routes:   map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, productPage)},
			want:     []string{"100200|Shirt", "200300|Hose (de)"},
			requests: []string{products}, scopes: []string{productReadUnpublishedScope},
		},
		{ // only the IDs that start with what is typed
			args:     []string{"products", "get", "2"},
			routes:   map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, productPage)},
			want:     []string{"200300|Hose (de)"},
			requests: []string{products}, scopes: []string{productReadUnpublishedScope},
		},
		{ // another resource's list, for an argument that names its items
			args:     []string{"categories", "for-product", ""},
			routes:   map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, productPage)},
			want:     []string{"100200|Shirt", "200300|Hose (de)"},
			requests: []string{products}, scopes: []string{productReadUnpublishedScope},
		},
		{ // a child read: its parent's list
			args: []string{"price-lists", "prices", "list", ""},
			routes: map[string]http.HandlerFunc{
				"/price/acme1/price-lists": respond(http.StatusOK, `[{"id":"PL1","name":{"en":"Retail"}}]`),
			},
			want:     []string{"PL1|Retail"},
			requests: []string{"/price/acme1/price-lists?pageNumber=1&pageSize=100"},
			scopes:   []string{"price.pricelist_read"},
		},
		{ // a child item: the child's list below the parent given
			args: []string{"price-lists", "prices", "get", "PL1", ""},
			routes: map[string]http.HandlerFunc{
				"/price/acme1/price-lists/PL1/prices": respond(http.StatusOK, `[{"id":"P1"},{"id":"P2"}]`),
			},
			want:     []string{"P1", "P2"},
			requests: []string{"/price/acme1/price-lists/PL1/prices?pageNumber=1&pageSize=100"},
			scopes:   []string{"price.pricelist_read"},
		},
		{ // a child item read from one response: the members of a site's mixins, with the child's scopes (none)
			args: []string{"sites", "mixins", "get", "AcmeStore_NL", ""},
			routes: map[string]http.HandlerFunc{
				"/site/acme1/sites/AcmeStore_NL/mixins": respond(http.StatusOK, `{"coupon":{},"customer":{}}`),
			},
			want:     []string{"coupon", "customer"},
			requests: []string{"/site/acme1/sites/AcmeStore_NL/mixins"},
		},
		{ // --site of any command: the site codes, without a scope
			args:     []string{"prices", "list", "--site", "Acme"},
			routes:   map[string]http.HandlerFunc{"/site/acme1/sites": respond(http.StatusOK, sitePage)},
			want:     []string{"AcmeStore_NL|Acme NL", "AcmeStore_DE|Acme DE"},
			requests: []string{"/site/acme1/sites?pageNumber=1&pageSize=100&totalCount=true"},
		},
		{ // a list of a site: the --site given before the ID, the list read whole as its command reads it
			args: []string{"shipping", "zones", "get", "--site", "AcmeStore_NL", ""},
			routes: map[string]http.HandlerFunc{
				"/shipping/acme1/AcmeStore_NL/zones": respond(http.StatusOK, `[{"id":"Z1","name":{"en":"Benelux"}}]`),
			},
			want:     []string{"Z1|Benelux"},
			requests: []string{"/shipping/acme1/AcmeStore_NL/zones"},
		},
		{ // a write: the read scopes of its list, never its write scope
			args: []string{"units", "update", "k"},
			routes: map[string]http.HandlerFunc{
				"/unit-handling/acme1/units": respond(http.StatusOK, `[{"code":"kg","name":{"en":"kilogram"}},{"code":"g"}]`),
			},
			want:     []string{"kg|kilogram"},
			requests: []string{"/unit-handling/acme1/units?pageNumber=1&pageSize=100"},
		},
		{ // the items of the list that get takes: the roots of category trees
			args: []string{"catalogs", "for-category", ""},
			routes: map[string]http.HandlerFunc{"/category/acme1/category-trees": respond(http.StatusOK,
				`[{"id":"C1","name":"Men","subcategories":[{"id":"C2","name":"Shirts"}]},{"id":"C3"}]`)},
			want:     []string{"C1|Men", "C3"},
			requests: []string{"/category/acme1/category-trees?showUnpublished=true"},
			scopes:   []string{"category.category_read_unpublished"},
		},
		{ // a column other than the first: the types of the search indexes, each once, without the NAME of an index
			args: []string{"search", "indexes", "get", ""},
			routes: map[string]http.HandlerFunc{
				"/search/acme1/search/indexes": respond(http.StatusOK, `[{"id":"i1","type":"BRAND","name":"Brands"},`+
					`{"id":"i2","type":"BRAND","name":"Labels"},{"id":"i3","type":"COLOR","name":"Colors"}]`),
			},
			want:     []string{"BRAND", "COLOR"},
			requests: []string{"/search/acme1/search/indexes?pageNumber=1&pageSize=100"},
			scopes:   []string{searchScope},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionLeavesOutWhatNoArgumentTakes(t *testing.T) {
	tt := completionCase{
		args: []string{"products", "get", ""},
		routes: map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, `[
			{"id":"red\u001b[31m"}, {"id":".."}, {"id":""}, {"code":"no-id"}, {"id":7},
			{"id":"ok","name":{"en":"bell\u0007here"}}
		]`)},
		want:     []string{"ok|bell here"},
		requests: []string{"/product/acme1/products?pageNumber=1&pageSize=100"},
		scopes:   []string{productReadUnpublishedScope},
	}
	tt.check(t)
}

// TestCompletionReadsNothingUnlessItCan checks that completion reads nothing
// where it may not or could not, and then offers only the values that need
// no read (want), if any.
func TestCompletionReadsNothingUnlessItCan(t *testing.T) {
	products := map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, productPage)}
	tests := []struct {
		name string
		env  func(f *fakeEmporix) map[string]string
		args []string
		want []string
	}{
		{name: "not opted in", env: (*fakeEmporix).env, args: []string{"products", "get", ""}},
		{name: "opted out", args: []string{"products", "get", ""}, env: func(f *fakeEmporix) map[string]string {
			env := f.env()
			env[envCompleteIDs] = "0"
			return env
		}},
		{name: "a list of a site without --site", env: completionEnv, args: []string{"shipping", "zones", "get", ""}},
		{name: "a list of a site that the read does not take", env: completionEnv, args: []string{"prices", "get", ""}},
		{name: "a list of a site that the read does not take", env: completionEnv, args: []string{"fees", "get", ""}},
		{name: "a parent that names no item", env: completionEnv, args: []string{"price-lists", "prices", "get", "..", ""}},
		{name: "an argument beyond the IDs", env: completionEnv, args: []string{"products", "get", "100200", ""}},
		{name: "a list with a write-capable scope", env: completionEnv,
			args: []string{"coupons", "redemption", "SUMMER", ""}},
		{name: "a collection listed by search", env: completionEnv, args: []string{"carts", "get", ""}},
		{name: "a command without IDs", env: completionEnv, args: []string{"products", "list", ""}},
		{name: "a list whose scopes name a type, below no type", env: completionEnv,
			args: []string{"custom-entities", "instances", "get", "1-not-a-type", ""}},
		{name: "several IDs, not opted in", env: (*fakeEmporix).env, args: []string{"ai", "analytics", "executions", ""}},
		{name: "several IDs, typed, not opted in", env: (*fakeEmporix).env,
			args: []string{"ai", "analytics", "executions", "a"}},
		{name: "several IDs, after a comma, not opted in", env: (*fakeEmporix).env,
			args: []string{"ai", "analytics", "executions", "a1,"}},
		{name: "predefined values, not opted in", env: (*fakeEmporix).env, args: []string{"ai", "rag", "fields", ""},
			want: []string{"PRODUCT", "ORDER"}},
		{name: "predefined values that start with what is typed", env: (*fakeEmporix).env,
			args: []string{"ai", "rag", "reindex", "P"}, want: []string{"PRODUCT"}},
		{name: "no predefined value that starts with what is typed", env: (*fakeEmporix).env,
			args: []string{"ai", "rag", "filter-fields", "X"}},
		{name: "a profile that cannot be read", env: completionEnv,
			args: []string{"ai", "rag", "fields", "--profile", "missing", ""}, want: []string{"PRODUCT", "ORDER"}},
		{name: "an argument beyond the predefined values", env: completionEnv,
			args: []string{"ai", "rag", "fields", "ORDER", ""}},
		{name: "the choices of an argument, which no list holds", env: completionEnv,
			args: []string{"segments", "items", "get", "S1", ""}, want: []string{"PRODUCT", "CATEGORY"}},
		{name: "the choices of an argument of a write", env: (*fakeEmporix).env,
			args: []string{"iam", "groups", "users", "update", "G1", "E"}, want: []string{"EMPLOYEE"}},
		{name: "the choices of an argument of a bulk write", env: (*fakeEmporix).env,
			args: []string{"segments", "items", "bulk-delete", "S1", ""}, want: []string{"PRODUCT", "CATEGORY"}},
	}
	for _, tt := range tests {
		f := newFakeEmporix(t)
		for path, handler := range products {
			f.route(path, handler)
		}
		// Cobra offers a required flag not given, such as --site, before any argument.
		got, _ := completion(t, tt.env(f), tt.args...)
		got = slices.DeleteFunc(got, func(c string) bool { return strings.HasPrefix(c, "--") })
		if !slices.Equal(got, tt.want) || len(f.gets) > 0 || len(f.tokenScopes) > 0 {
			t.Errorf("%s: completions %q after requests %q and tokens %q; want %q and none", tt.name, got, f.gets,
				f.tokenScopes, tt.want)
		}
	}
}

func TestCompletionSwallowsFailures(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"an error":              respond(http.StatusInternalServerError, `{"message":"boom"}`),
		"a response not a list": respond(http.StatusOK, `{"id":"100200"}`),
		"a list of other items": respond(http.StatusOK, `["100200"]`),
	}
	for name, handler := range tests {
		f := newFakeEmporix(t)
		f.route("/product/acme1/products", handler)
		if got := completed(t, completionEnv(f), "products", "get", ""); len(got) > 0 {
			t.Errorf("%s: completions %q; want none", name, got)
		}
	}
	f := newFakeEmporix(t)
	f.route("/site/acme1/sites/AcmeStore_NL/mixins", respond(http.StatusOK, `["coupon"]`))
	if got := completed(t, completionEnv(f), "sites", "mixins", "get", "AcmeStore_NL", ""); len(got) > 0 {
		t.Errorf("a response without the items where they belong: completions %q; want none", got)
	}
	f = newFakeEmporix(t)
	env := completionEnv(f)
	env[envClientSecret] = "wrong"
	if got := completed(t, env, "products", "get", ""); len(got) > 0 || len(f.gets) > 0 {
		t.Errorf("a rejected credential: completions %q after %q; want none", got, f.gets)
	}
	if got := completed(t, map[string]string{envCompleteIDs: "1"}, "products", "get", ""); len(got) > 0 {
		t.Errorf("no credentials: completions %q; want none", got)
	}
}

func TestCompletionGivesUpWithinItsTimeout(t *testing.T) {
	f := newFakeEmporix(t)
	f.route("/product/acme1/products", func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never answers before the client gives up
	})
	env := completionEnv(f)
	const timeout = 50 * time.Millisecond
	a := &app{getenv: func(key string) string { return env[key] }, timeout: defaultTimeout, completionTimeout: timeout}
	var out, errOut bytes.Buffer
	root := newRootCommand(a, &out, &errOut)
	root.SetArgs([]string{"__complete", "products", "get", ""})
	start := time.Now()
	if err := root.ExecuteContext(context.Background()); err != nil || out.String() != ":4\n" {
		t.Fatalf("error %v, completions %q; want none", err, out.String())
	}
	if took := time.Since(start); took < timeout || took > defaultCompletionTimeout {
		t.Fatalf("completion took %s; want about %s", took, timeout)
	}
}

func TestCompletionOfAColumnTheTableLacksPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic")
		}
	}()
	listedIDs(typeNames("unit type"), request{path: "/x"}, single{items: listItems}, columnID)
}

func TestCompletionHelpSaysHowToOptIn(t *testing.T) {
	code, stdout, _ := runCLI(t, nil, "completion", "--help")
	if code != ExitOK || !strings.Contains(stdout, envCompleteIDs+"=1") {
		t.Fatalf("exit %d, help %q; want the opt-in", code, stdout)
	}
}

// TestCompletionTakesTheProfileInEffect checks that completion reads with the
// API key of the profile in effect, its client secret from the keychain, and
// takes its default site as the command would, but reads no profile unless
// the user opted in.
func TestCompletionTakesTheProfileInEffect(t *testing.T) {
	f := newFakeEmporix(t)
	f.route("/shipping/acme1/AcmeStore_NL/zones", respond(http.StatusOK, `[{"id":"Z1","name":{"en":"Benelux"}}]`))
	setup := newProfileSetup(t, twoProfiles(f), map[string]string{"dev": testSecret})
	args := []string{"__complete", "shipping", "zones", "get", ""}
	code, stdout, stderr := runIn(t, setup.env(map[string]string{envCompleteIDs: "1"}, "", false), args...)
	// Cobra offers a required flag not given, such as --site, before any argument.
	if code != ExitOK || !strings.HasSuffix(stdout, "\nZ1\tBenelux\n:4\n") || stderr != completionDone {
		t.Fatalf("exit %d, stdout %q, stderr %q; want Z1", code, stdout, stderr)
	}
	f.assertRequests(t, "/shipping/acme1/AcmeStore_NL/zones")
	env := setup.env(nil, "", false)
	env.Profiles.ConfigDir = func() (string, error) {
		t.Error("read the profiles without the opt-in")
		return setup.dir, nil
	}
	if code, stdout, _ := runIn(t, env, args...); code != ExitOK || strings.Contains(stdout, "Z1") {
		t.Fatalf("not opted in: exit %d, stdout %q; want no IDs", code, stdout)
	}
}

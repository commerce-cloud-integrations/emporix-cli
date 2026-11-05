package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// searchReads returns the reads of the Search Service: indexes and saved
// searches, of every custom entity type or of one, and index jobs.
func searchReads() []readCase {
	const scope = "search.search_read"
	indexes := `[{"id":"vehicles","type":"vehicle","name":{"en":"Vehicles","nl":"Voertuigen"},"status":"ready",
		"fields":[{"path":"name","text":true,"autocomplete":"PREFIX"},{"path":"mixins.power","exact":true}],
		"metadata":{"version":2,"createdAt":"2026-03-12T10:00:00.000Z"}},
		{"id":"parts","type":"part","name":{"de":"Teile"},"status":"building","fields":[]}]`
	queries := `[{"id":"red-cars","type":"vehicle","name":{"en":"Red cars","nl":"Rode auto's"},"indexId":"vehicles",
		"queries":{"or":[{"type":"TEXT","field":"name","query":"red"}]},"metadata":{"version":1}},
		{"id":"all-parts","type":"part","indexId":"parts","filters":{"field":"code","operator":"EXISTS"}}]`
	return []readCase{
		{
			// FIELDS names the indexed fields' paths; how each is searched is only in -o json.
			args:    []string{"search", "indexes", "list", "--language", "nl"},
			path:    "/search/acme1/search/indexes",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    indexes,
			want: "ID        TYPE     NAME        STATUS    FIELDS              VERSION\n" +
				"vehicles  vehicle  Voertuigen  ready     name, mixins.power  2\n" +
				"parts     part     Teile (de)  building                      \n",
		},
		{
			args:    []string{"search", "indexes", "for-type", "vehicle", "--language", "en"},
			path:    "/search/acme1/search/vehicle/indexes",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"vehicles","type":"vehicle","name":{"en":"Vehicles"},"status":"failed","fields":[
				{"path":"name"},{"path":"description"},{"path":"mixins.engine.power"},{"path":"mixins.engine.fuel"}]}]`,
			want: "ID        TYPE     NAME      STATUS  FIELDS                                                      VERSION\n" +
				"vehicles  vehicle  Vehicles  failed  name, description, mixins.engine.power, mixins.engine.fuel  \n",
		},
		{
			// The type and the ID address the index, in that order.
			args:   []string{"search", "indexes", "get", "vehicle", "vehicles", "--language", "de"},
			path:   "/search/acme1/search/vehicle/indexes/vehicles",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"vehicles","type":"vehicle","name":{"en":"Vehicles","de":"Fahrzeuge"},"status":"ready",
				"fields":[{"path":"name"}],"metadata":{"version":3}}`,
			want: "ID        TYPE     NAME       STATUS  FIELDS  VERSION\n" +
				"vehicles  vehicle  Fahrzeuge  ready   name    3\n",
		},
		{
			// Queries and filters are only in -o json.
			args:    []string{"search", "queries", "list", "--language", "nl"},
			path:    "/search/acme1/search/queries",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    queries,
			want: "ID         TYPE     NAME         INDEX     VERSION\n" +
				"red-cars   vehicle  Rode auto's  vehicles  1\n" +
				"all-parts  part                  parts     \n",
		},
		{
			args:    []string{"search", "queries", "for-type", "part", "--language", "en"},
			path:    "/search/acme1/search/part/queries",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    `[{"id":"all-parts","type":"part","name":{"en":"All parts"},"indexId":"parts"}]`,
			want: "ID         TYPE  NAME       INDEX  VERSION\n" +
				"all-parts  part  All parts  parts  \n",
		},
		{
			args:   []string{"search", "queries", "get", "vehicle", "red-cars", "--language", "en"},
			path:   "/search/acme1/search/vehicle/queries/red-cars",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"red-cars","type":"vehicle","name":{"en":"Red cars"},"indexId":"vehicles","metadata":{"version":2}}`,
			want: "ID        TYPE     NAME      INDEX     VERSION\n" +
				"red-cars  vehicle  Red cars  vehicles  2\n",
		},
		{
			// customer and legal-entity are types too, sent as given.
			args:    []string{"search", "indexes", "for-type", "legal-entity", "--language", "en"},
			path:    "/search/acme1/search/legal-entity/indexes",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"companies","type":"legal-entity","name":{"en":"Companies"},"status":"ready",
				"fields":[{"path":"name","text":true}]}]`,
			want: "ID         TYPE          NAME       STATUS  FIELDS  VERSION\n" +
				"companies  legal-entity  Companies  ready   name    \n",
		},
		{
			args:   []string{"search", "queries", "get", "customer", "by-city", "--language", "en"},
			path:   "/search/acme1/search/customer/queries/by-city",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"by-city","type":"customer","name":{"en":"By city"},"indexId":"customers"}`,
			want: "ID       TYPE      NAME     INDEX      VERSION\n" +
				"by-city  customer  By city  customers  \n",
		},
	}
}

// searchJobReads returns the reads of index jobs: a --query is sent as given (a lower-case value matches
// no job, see the help of jobs), and the list and get show what each job recorded.
func searchJobReads() []readCase {
	const scope = "search.search_read"
	return []readCase{
		{
			// --query is sent as given: a lower-case value matches no job (see the help of jobs).
			args:    []string{"search", "jobs", "list", "--query", "status:success"},
			path:    "/search/acme1/jobs",
			query:   "q=status%3Asuccess",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    `[{"id":"66f1a2b3c4d5e6f708192a3b","status":"success","type":"create_index","indexId":"vehicles"}]`,
			want: "ID                        TYPE          STATUS   INDEX     CREATED  MODIFIED  RESULT\n" +
				"66f1a2b3c4d5e6f708192a3b  create_index  success  vehicles                     \n",
		},
		{
			// RESULT is what a job recorded when it finished, cut in tables.
			args:    []string{"search", "jobs", "list"},
			path:    "/search/acme1/jobs",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"66f1a2b3c4d5e6f708192a3b","status":"success","type":"create_index","indexId":"vehicles",
				"message":"Creating index vehicles","response":"Index created.","metadata":{"version":2,
				"createdAt":"2026-03-12T10:00:00.000Z","modifiedAt":"2026-03-12T10:04:00.000Z"}},
				{"id":"66f1a2b3c4d5e6f708192a3c","status":"in_progress","type":"update_index","indexId":"parts",
				"metadata":{"createdAt":"2026-03-13T08:00:00.000Z"}}]`,
			want: "ID                        TYPE          STATUS       INDEX     CREATED                   " +
				"MODIFIED                  RESULT\n" +
				"66f1a2b3c4d5e6f708192a3b  create_index  success      vehicles  2026-03-12T10:00:00.000Z  " +
				"2026-03-12T10:04:00.000Z  Index created.\n" +
				"66f1a2b3c4d5e6f708192a3c  update_index  in_progress  parts     2026-03-13T08:00:00.000Z  " +
				"                          \n",
		},
		{
			args:   []string{"search", "jobs", "get", "66f1a2b3c4d5e6f708192a3b"},
			path:   "/search/acme1/jobs/66f1a2b3c4d5e6f708192a3b",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"66f1a2b3c4d5e6f708192a3b","status":"failure","type":"delete_index","indexId":"parts",
				"response":"The index could not be deleted because the search engine did not answer in time."}`,
			want: "ID                        TYPE          STATUS   INDEX  CREATED  MODIFIED  RESULT\n" +
				"66f1a2b3c4d5e6f708192a3b  delete_index  failure  parts                     " +
				"The index could not be deleted because the search engine di…\n",
		},
	}
}

// searchExportReads returns the opt-in export of search indexes, sent as POST.
func searchExportReads() []readCase {
	return []readCase{
		{
			// Each --index TYPE/ID selects one index; an id holds no slash, so a type may.
			args: []string{"search", "indexes", "export", "--index", "vehicle/vehicles", "--index", "two/part/bikes",
				"--allow-write-scope"},
			path:   "/search/acme1/search/indexes/export",
			scopes: []string{"search.search_manage"},
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `[{"type":"vehicle","id":"vehicles"},{"type":"two/part","id":"bikes"}]`,
			body:   `{"exportedAt":"2026-03-12T10:00:00.000Z","data":"W10="}`,
			want:   "DATA  EXPORTED AT\nW10=  2026-03-12T10:00:00.000Z\n",
		},
	}
}

// TestSearchIndexExportFailures checks an export of search indexes: an
// --index that is not TYPE/ID is a usage error that sends nothing; Emporix's
// answer to an unknown index is not found (exit 4), as help says.
func TestSearchIndexExportFailures(t *testing.T) {
	export := []string{"search", "indexes", "export", "--" + writeScopeFlag}
	fake := newFakeEmporix(t)
	for _, index := range []string{"vehicles", "/vehicles", "vehicle/", ""} {
		code, stdout, stderr := runCLI(t, fake.env(), slices.Concat(export, []string{"--index", index})...)
		want := `invalid argument "` + index + `" for "--index" flag: use TYPE/ID`
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, want) {
			t.Errorf("--index %q: exit %d, stdout %q, stderr %q; want exit 2 and %q", index, code, stdout, stderr, want)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("tokens %q, requests %q; want none", fake.tokenScopes, fake.gets)
	}
	fake.route("/search/acme1/search/indexes/export", respond(http.StatusNotFound, `{"message":"rejected"}`))
	code, stdout, stderr := runCLI(t, fake.env(), slices.Concat(export, []string{"--index", "a/b"})...)
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "rejected") {
		t.Errorf("unknown index: exit %d, stdout %q, stderr %q; want exit 4", code, stdout, stderr)
	}
}

// helpText returns the help of the command that args name, with its words joined by single spaces, so that a
// sentence broken across lines is compared whole.
func helpText(t *testing.T, args ...string) string {
	t.Helper()
	_, help, _ := runCLI(t, nil, append(slices.Clone(args), "--help")...)
	return strings.Join(strings.Fields(help), " ")
}

// TestSearchListHelpNamesTheFilterValues checks that the help of the job and index lists names the upper-case
// values --query matches, checked live, and says that a lower-case one matches nothing.
func TestSearchListHelpNamesTheFilterValues(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{
			args: []string{"search", "jobs", "list"},
			want: []string{"`status:SUCCESS`", "`type:UPDATE_INDEX`", "a lower-case filter, such as",
				"`status:success`, matches no job"},
		},
		{
			args: []string{"search", "indexes", "list"},
			want: []string{"`status:READY`", "`status:ready` matches no index"},
		},
	} {
		text := helpText(t, tt.args...)
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Errorf("%q: help lacks %q:\n%s", tt.args, want, text)
			}
		}
	}
}

// TestSearchHelpNamesTheBuiltInTypes checks that the help of the search group,
// of documents, and of indexes and queries and their commands per type names
// customer and legal-entity as types beside the custom entity types, that
// the help of documents names the limitation of the scopes that it requests,
// and that the help of indexes and queries, their reads and writes per type
// and the export of indexes says how to reach an item kept under another
// spelling of a built-in type.
func TestSearchHelpNamesTheBuiltInTypes(t *testing.T) {
	const types = "customer or legal-entity, exactly as written, or a custom entity type"
	for _, args := range [][]string{
		{"search"}, {"search", "documents"},
		{"search", "indexes"}, {"search", "indexes", "for-type"}, {"search", "indexes", "get"},
		{"search", "indexes", "export"},
		{"search", "indexes", "update"}, {"search", "indexes", "delete"},
		{"search", "queries"}, {"search", "queries", "for-type"}, {"search", "queries", "get"},
		{"search", "queries", "update"}, {"search", "queries", "delete"},
	} {
		if text := helpText(t, args...); !strings.Contains(text, types) {
			t.Errorf("%q: help lacks %q:\n%s", args, types, text)
		}
	}
	const limitation = "Known limitation: for customer or legal-entity, epx requests the read scope that finds " +
		"every document, so a credential that holds only a restricted one, such as customer.customer_read--DE, or " +
		"only the own-read scope, such as customer.customer_read_own, fails with exit code 3."
	if text := helpText(t, "search", "documents"); !strings.Contains(text, limitation) {
		t.Errorf("search documents: help lacks %q:\n%s", limitation, text)
	}
	// What TYPE refuses in another spelling is listed, exported with --body, and reached with epx api.
	const otherwise = " kept under another spelling of a built-in type, such as CUSTOMER, "
	for singular, use := range map[string]string{"search index": "indexes", "saved search": "queries"} {
		path := " /search/{tenant}/search/CUSTOMER/" + use + "/ID --scope search.search_"
		for args, want := range map[[3]string]string{
			{"search", use}: "A " + singular + otherwise + "is listed too; `epx api` reaches it by its path, such as " +
				"`epx api get" + path + "read`.",
			{"search", use, "for-type"}: "`epx api get" + path + "read`.",
			{"search", use, "get"}:      "`epx api get" + path + "read`.",
			{"search", use, "update"}: "A " + singular + otherwise + "is upserted with `epx api put" + path +
				"manage --file FILE`.",
			{"search", use, "delete"}: "A " + singular + otherwise + "is deleted with `epx api delete" + path + "manage`.",
		} {
			if text := helpText(t, args[:]...); !strings.Contains(text, want) {
				t.Errorf("%q: help lacks %q:\n%s", args, want, text)
			}
		}
	}
	const export = "A search index" + otherwise + "is exported with --body."
	if text := helpText(t, "search", "indexes", "export"); !strings.Contains(text, export) {
		t.Errorf("search indexes export: help lacks %q:\n%s", export, text)
	}
}

// TestExactTypesNeedBuiltInTypes checks that saying how to reach an item
// under another spelling of a built-in type, for a service whose collections
// take none, take them in any case, or name the first in upper case, is
// refused while the command tree is built, naming the service (its use).
func TestExactTypesNeedBuiltInTypes(t *testing.T) {
	const want = "things: says how to reach an item under another spelling of a built-in type, but TYPE " +
		"does not refuse the upper-case spelling of its first built-in type, if it has one"
	for name, types := range map[string]builtInTypes{"none": {}, "any case": anyCaseTypes("PRODUCT"),
		"spelled upper": spelledTypes(builtInType{name: "PRODUCT"})} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recovered, _ := recover().(string); recovered != want {
					t.Fatalf("panic %q; want %q", recovered, want)
				}
			}()
			newExactTypes(types, "things")
		})
	}
}

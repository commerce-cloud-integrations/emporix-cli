package cli

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestBulkAssignmentsAreAnsweredPerItem checks that both bulk writes of
// category assignments, each to its own path, read the result of each item:
// a response that names items that failed is printed, those items are listed
// on standard error (one that failed with a server error may have been
// assigned all the same, and is not sent again) and the write exits with the
// code that their failures share, else 1; a response that does not give the
// result of each item exactly once fails (exit 1) too.
func TestBulkAssignmentsAreAnsweredPerItem(t *testing.T) {
	const unknown = "the response does not give the result of each of the 2 items, which may have been carried " +
		"out all the same: read them before you send them again\n"
	for _, tt := range []struct {
		results string
		code    int
		stderr  string // what standard error ends with, after the method and path for the error
	}{
		{
			results: `[{"index":0,"code":201,"status":"CREATED"},
				{"index":1,"code":409,"status":"CONFLICT","message":"assigned already"}]`,
			code: ExitConflict, stderr: "item 1: 409 Conflict: assigned already\nError: {request}: 1 of 2 items failed " +
				"(each changed since it was read, or exists already: read them again before you retry)\n",
		},
		{
			results: `[{"index":1,"code":500,"message":"timed out"},{"index":0,"code":201}]`, code: ExitError,
			stderr: "item 1: 500 Internal Server Error: timed out (it may have been carried out all the same: read the " +
				"item before you send it again)\nError: {request}: 1 of 2 items failed\n",
		},
		{results: `[{"index":0,"code":201},{"index":0,"code":201}]`, code: ExitError, stderr: unknown},
		{results: `[{"code":201},{"index":0,"code":201}]`, code: ExitError, stderr: unknown},
		{results: `[{"index":-1,"code":201},{"code":201}]`, code: ExitError, stderr: unknown},
	} {
		for command, path := range map[string]string{
			"bulk-create": "/category/acme1/categories/c1/assignments/bulk",
			"bulk-update": "/category/acme1/categories/c1/assignments/references/bulk",
		} {
			fake := newFakeEmporix(t)
			fake.route(path, respond(http.StatusMultiStatus, tt.results))
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "categories", "assignments", command, "c1",
				"--file", writeBody(t, assignmentsBody), "--yes")
			method, _, _ := strings.Cut(fake.sent[0], " ")
			want := strings.ReplaceAll(tt.stderr, "{request}", method+" "+path)
			if code != tt.code || !strings.HasPrefix(stdout, "[\n  {\n") || !strings.HasSuffix(stderr, want) ||
				len(fake.sent) != 1 {
				t.Errorf("%s answered %s: exit %d, stdout %q, stderr %q; want exit %d and %q", command, tt.results, code,
					stdout, stderr, tt.code, want)
			}
		}
	}
}

// TestCategoryPublishFlags checks that --publish and --unpublish exclude each
// other, that patch, which Emporix publishes as its body says, takes neither,
// and that --publish=false publishes nothing: the category is sent without the
// publish parameter, with category_manage only.
func TestCategoryPublishFlags(t *testing.T) {
	fake := newFakeEmporix(t)
	body := writeBody(t, `{"published":false}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"update", "c1", "--publish", "--unpublish"}, want: "[publish unpublish] were all set"},
		{args: []string{"patch", "c1", "--publish"}, want: "unknown flag: --publish"},
		{args: []string{"patch", "c1", "--unpublish"}, want: "unknown flag: --unpublish"},
	} {
		args := append(append([]string{"categories"}, tt.args...), "--file", body, "--dry-run")
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
	}
	fake.assertNothingSent(t)
	fake.route("/category/acme1/categories", respond(http.StatusCreated, `{"id":"c1"}`))
	code, _, stderr := runIn(t, writeEnv(fake, false, ""), "categories", "create", "--file", body, "--publish=false",
		"--yes")
	if code != ExitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	fake.assertRequests(t, "/category/acme1/categories")
	fake.assertTokens(t, tokenScope([]string{categoryManageScope}))
}

// TestCatalogWriteUsageErrorsSendNothing checks that the arguments and bodies
// of the writes of catalogs and categories are checked before anything is
// sent: a usage error (exit 2).
func TestCatalogWriteUsageErrorsSendNothing(t *testing.T) {
	body := writeBody(t, `{}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"categories", "assignments", "update", "c1"}, want: "accepts 2 arg(s), received 1"},
		{args: []string{"categories", "assignments", "update", "c1", ".."}, want: `invalid PRODUCT_ID ".."`},
		{args: []string{"categories", "assignments", "update", "c1", "bulk"},
			want: `invalid PRODUCT_ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"categories", "assignments", "update", "c1", "p1", "--file", body},
			want: "unknown flag: --file"},
		{args: []string{"categories", "assignments", "create", "", "--file", body}, want: `invalid CATEGORY_ID ""`},
		{args: []string{"categories", "assignments", "bulk-create", "c1", "--file", body},
			want: "POST /category/{tenant}/categories/c1/assignments/bulk: a bulk write sends a JSON array of one or more " +
				"items"},
		{args: []string{"categories", "assignments", "bulk-update", "c1", "--file", writeBody(t, "[]")},
			want: "a bulk write sends a JSON array of one or more items"},
		{args: []string{"category-trees", "rebuild"}, want: "accepts 1 arg(s), received 0"},
		{args: []string{"catalogs", "patch", "a/b", "--file", body}, want: "PATCH /catalog/{tenant}/catalogs/a%2Fb: " +
			"a path segment of a write"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, "--yes")...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
	// The products assigned to a category are listed with assignments list.
	code, _, stderr := runIn(t, writeEnv(newFakeEmporix(t), false, ""), "categories", "assignments", "c1")
	if want := `unknown command "c1" for "epx categories assignments"`; code != ExitUsage ||
		!strings.Contains(stderr, want) {
		t.Errorf("categories assignments c1: exit %d, stderr %q; want exit 2 and %q", code, stderr, want)
	}
}

// TestCatalogWriteHelp checks the help of the writes of catalogs and
// categories: the scope that each flag of a choice adds, before the scope
// sentence and in its usage; the arguments of a write below a category, in
// order, and where they come from; how a bulk write is answered; and the
// scopes of the reads and writes of each group.
func TestCatalogWriteHelp(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: []string{"categories", "update"}, want: []string{
			"\n\n--publish also requests the category.category_publish scope.\n--unpublish also requests the " +
				"category.category_unpublish scope.\n\nRequires the category.category_manage scope.\n\nUsage:",
			"needs published: false in the body (requests category.category_unpublish too)",
			"the body has the members of create's, published included, but\nid, type and ownClassificationMixins.",
		}},
		{args: []string{"categories", "create"}, want: []string{
			"\n\n--publish also requests the category.category_publish scope.\n\nRequires the category.category_manage " +
				"scope.\n\nUsage:",
		}},
		{args: []string{"catalogs", "patch"}, want: []string{"Emporix requires metadata.version in the body",
			"with 404 Not Found: epx then exits with 5, a conflict"}},
		{args: []string{"categories", "patch"}, want: []string{"with 400 Bad Request (exit 1),\nnot as a conflict.",
			"published in the body publishes or unpublishes the category as sent"}},
		{args: []string{"categories", "delete"}, want: []string{"move to the category above it",
			"Emporix refuses to\ndelete a root category with categories below it"}},
		{args: []string{"categories", "assignments", "update"}, want: []string{
			"not its code.\nCATEGORY_ID is the ID that `epx categories list` shows; an unknown one is not found.\n\n",
			"Usage:\n  epx categories assignments update CATEGORY_ID PRODUCT_ID [flags]",
			"send without asking (required without a terminal)\n",
		}},
		{args: []string{"categories", "assignments", "bulk-update"}, want: []string{
			"as in\n{\"ref\": {\"id\": \"100200\"}}.", "Emporix takes up to 200 items in one request",
		}},
		{args: []string{"categories"}, want: []string{"Its reads require the category.category_read_unpublished " +
			"scope.\nIts writes require different OAuth scopes; the help of each names them.\n\nUsage:",
			" Create a category\n"}},
		{args: []string{"categories", "assignments"}, want: []string{"Its reads require the " +
			"category.category_read_unpublished scope.\nIts writes require different OAuth scopes"}},
		{args: []string{"categories", "assignments", "delete-all"}, want: []string{
			"to the category CATEGORY_ID, published or\nunpublished, as `epx categories assignments list CATEGORY_ID` " +
				"lists them.",
			"Requires the category.category_manage, category.category_read_unpublished scopes.",
		}},
		{args: []string{"category-trees"}, want: []string{"\nIts writes require the category.category_manage, " +
			"category.category_read_unpublished scopes.\n\nUsage:"}},
	} {
		_, help, _ := runCLI(t, nil, append(tt.args, "--help")...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%q: help %q; want %q", tt.args, help, want)
			}
		}
	}
}

// TestChoiceThatItsRequestSetsPanics checks that a write declared with a
// choice that sets a parameter that its request sends already panics as the
// command tree is built.
func TestChoiceThatItsRequestSetsPanics(t *testing.T) {
	defer func() {
		const want = "update ID: its choice sets publish, which its request sends already"
		if recovered, _ := recover().(string); recovered != want {
			t.Errorf("recovered %q; want %q", recovered, want)
		}
	}()
	(&write{
		use: "update ID", method: http.MethodPut, scopes: []string{"thing.thing_manage"}, body: noBody{},
		req: request{path: "/things/" + placeholderID, query: url.Values{"publish": {"false"}}}, choice: publishing(true),
	}).command(&app{})
}

// TestCatalogPatchReportsAStaleVersionAsAConflict checks that the 404 Not
// Found with which the Catalog Service answers a patch of a stale version
// exits with 5, a conflict, when the body carries metadata.version, and with
// 4, not found, without one.
func TestCatalogPatchReportsAStaleVersionAsAConflict(t *testing.T) {
	const message = "Catalog with id 'c1' and version '2' not found"
	for _, tt := range []struct {
		file string
		code int
		want string // the end of standard error
	}{
		{file: `{"name":{"en":"Acme"},"metadata":{"version":2}}`, code: ExitConflict,
			want: "404 Not Found: " + message + " (the item changed since it was read, or exists already: read it " +
				"again before you retry)\n"},
		{file: `{"name":{"en":"Acme"},"metadata":{}}`, code: ExitNotFound, want: "404 Not Found: " + message + "\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/catalog/acme1/catalogs/c1", respond(http.StatusNotFound, `{"code":404,"message":"`+message+`"}`))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "catalogs", "patch", "c1", "--file",
			writeBody(t, tt.file), "--yes")
		if code != tt.code || stdout != "" || !strings.HasSuffix(stderr, "PATCH /catalog/acme1/catalogs/c1: "+tt.want) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.file, code, stdout, stderr, tt.code,
				tt.want)
		}
		fake.assertRequests(t, "/catalog/acme1/catalogs/c1")
	}
}

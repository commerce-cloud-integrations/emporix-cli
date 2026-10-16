package cli

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"strings"
	"testing"
)

// TestBulkWritesListTheItemsThatFailed checks that a bulk write whose
// response names items that failed prints the response, lists the items that
// failed on standard error in the order of the response, sanitized, each
// named by its index or position (and one that failed with a 5xx status as
// possibly carried out), and exits non-zero: with 5 when each failed with 409
// Conflict, 4 when each was not found, and 1 otherwise.
func TestBulkWritesListTheItemsThatFailed(t *testing.T) {
	const items = `[{"id":"p1"},{"id":"p2"},{"id":"p3"}]`
	for _, tt := range []struct {
		results string // the results of the three items
		code    int
		stderr  string
	}{
		{
			results: `[{"index":0,"id":"p1","code":204},` +
				`{"index":1,"id":"p2","code":400,"status":"BAD_REQUEST","message":"unknown site\u001b[2J"},` +
				`{"index":2,"code":409,"status":"CONFLICT","message":"changed"}]`,
			code: ExitError,
			stderr: "item 1 (p2): 400 Bad Request: unknown site [2J\nitem 2: 409 Conflict: changed\n" +
				"Error: PUT /price/acme1/prices/bulk: 2 of 3 items failed\n",
		},
		{
			results: `[{"index":0,"id":"p1","code":409},{"index":1,"id":"p2","code":204},{"id":"p3","code":409}]`,
			code:    ExitConflict,
			stderr: "item 0 (p1): 409 Conflict\nitem 2 (p3): 409 Conflict\nError: PUT /price/acme1/prices/bulk: " +
				"2 of 3 items failed (each changed since it was read, or exists already: read them again before you " +
				"retry)\n",
		},
		{
			results: `[{"index":0,"code":404},{"index":1,"code":404},{"index":2,"code":499,"message":"gone"}]`,
			code:    ExitError,
			stderr: "item 0: 404 Not Found\nitem 1: 404 Not Found\nitem 2: 499: gone\n" +
				"Error: PUT /price/acme1/prices/bulk: 3 of 3 items failed\n",
		},
		{
			results: `[{"index":2,"code":404},{"index":1,"code":201},{"index":0,"code":404}]`,
			code:    ExitNotFound,
			stderr:  "item 2: 404 Not Found\nitem 0: 404 Not Found\nError: PUT /price/acme1/prices/bulk: 2 of 3 items failed\n",
		},
		{
			results: `[{"index":0,"code":204},{"index":1,"id":"p2","code":500,"message":"boom"},{"index":2,"code":204}]`,
			code:    ExitError,
			stderr: "item 1 (p2): 500 Internal Server Error: boom (it may have been carried out all the same: read the " +
				"item before you send it again)\nError: PUT /price/acme1/prices/bulk: 1 of 3 items failed\n",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route("/price/acme1/prices/bulk", respond(http.StatusMultiStatus, tt.results))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "prices", "bulk-update", "--file",
			writeBody(t, items), "--yes")
		if code != tt.code || !strings.HasPrefix(stdout, "[\n  {\n") || stderr != tt.stderr {
			t.Errorf("%s: exit %d, stdout %q, stderr:\n%s\nwant exit %d and:\n%s", tt.results, code, stdout, stderr,
				tt.code, tt.stderr)
		}
	}
}

// TestBulkWritesListTheItemsThatFailedWhenPrintingFails checks that a bulk
// write whose response cannot be printed still lists the items that failed
// on standard error and exits with the code that their failures share, with
// both errors.
func TestBulkWritesListTheItemsThatFailedWhenPrintingFails(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/prices/bulk",
		respond(http.StatusMultiStatus, `[{"id":"p1","code":409},{"id":"p2","code":409}]`))
	env := writeEnv(fake, false, "")
	var errOut bytes.Buffer
	env.Out, env.Err = failingWriter{}, &errOut
	code := Run(context.Background(), []string{"prices", "bulk-update", "--file",
		writeBody(t, `[{"id":"p1"},{"id":"p2"}]`), "--yes"}, env)
	const want = "item 0 (p1): 409 Conflict\nitem 1 (p2): 409 Conflict\nError: PUT /price/acme1/prices/bulk: 2 of 2 " +
		"items failed (each changed since it was read, or exists already: read them again before you retry) disk full\n"
	if code != ExitConflict || errOut.String() != want {
		t.Fatalf("exit %d, stderr:\n%s\nwant exit %d and:\n%s", code, errOut.String(), ExitConflict, want)
	}
}

// TestBulkWritesWithoutTheResultOfEachItemFail checks that a bulk write
// whose response does not give the result of each item sent fails (exit 1),
// saying that the items may have been carried out, after printing what
// Emporix answered, if anything.
func TestBulkWritesWithoutTheResultOfEachItemFail(t *testing.T) {
	const want = "Error: POST /availability/acme1/availability/bulk: the response does not give the result of each " +
		"of the 2 items, which may have been carried out all the same: read them before you send them again\n"
	for _, tt := range []struct {
		status  int
		results string
		stdout  string
	}{
		{status: http.StatusMultiStatus, results: `[{"code":201}]`, stdout: "[\n  {\n    \"code\": 201\n  }\n]\n"},
		{status: http.StatusMultiStatus, results: `[{"code":201},{"status":"Created"}]`},
		{status: http.StatusMultiStatus, results: `[{"code":201},{"code":"201"}]`},
		{status: http.StatusMultiStatus, results: `[{"code":201},{"code":42}]`},
		{status: http.StatusMultiStatus, results: `{"code":201}`},
		{status: http.StatusOK, results: `[{"code":201},[]]`},
		{status: http.StatusNoContent},
	} {
		fake := newFakeEmporix(t)
		fake.route("/availability/acme1/availability/bulk", respond(tt.status, tt.results))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "availability", "bulk-create", "--file",
			writeBody(t, "["+stockItem+","+stockItem+"]"), "--yes")
		if code != ExitError || !strings.HasSuffix(stderr, want) || (tt.stdout != "" && stdout != tt.stdout) ||
			(tt.results == "") != (stdout == "") {
			t.Errorf("%d %s: exit %d, stdout %q, stderr %q", tt.status, tt.results, code, stdout, stderr)
		}
	}
}

// TestBulkWritesTakeUpToTheirLimit checks that a bulk write sends as many
// items as Emporix takes in one request, 200 for prices, and refuses more as
// a usage error (exit 2) that sends nothing, also on a dry run.
func TestBulkWritesTakeUpToTheirLimit(t *testing.T) {
	items := func(n int) string { return "[" + strings.Repeat(priceJSON+",", n-1) + priceJSON + "]" }
	fake := newFakeEmporix(t)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "prices", "bulk-create", "--file",
		writeBody(t, items(200)), "--dry-run")
	if code != ExitOK || !strings.HasPrefix(stdout, "POST ") || stderr != "" {
		t.Errorf("200 items: exit %d, stdout %.40q, stderr %q; want the request", code, stdout, stderr)
	}
	for _, flag := range []string{"--dry-run", "--yes"} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "prices", "bulk-create", "--file",
			writeBody(t, items(201)), flag)
		const want = "POST /price/{tenant}/prices/bulk: a bulk write sends at most 200 items in one request, not 201"
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, want) {
			t.Errorf("201 items %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", flag, code, stdout, stderr, want)
		}
		fake.assertNothingSent(t)
	}
}

// TestWriteBodiesAreChecked checks that a declared write requires --file, a
// bulk write a JSON array of one or more items, a write whose request names
// the site --site, and an ID that addresses one item, not another operation
// of the collection, failing as usage errors (exit 2) that send nothing, also
// for a dry run.
func TestWriteBodiesAreChecked(t *testing.T) {
	assertRefused(t, []refusedWrite{
		{args: []string{"prices", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"prices", "bulk-update"}, want: `required flag(s) "file" not set`},
		{args: []string{"prices", "bulk-create"}, file: priceJSON,
			want: "POST /price/{tenant}/prices/bulk: a bulk write sends a JSON array of one or more items"},
		{args: []string{"availability", "bulk-update"}, file: "[]",
			want: "PUT /availability/{tenant}/availability/bulk: a bulk write sends a JSON array of one or more items"},
		{args: []string{"availability", "bulk-create"}, file: "null", want: "a JSON array of one or more items"},
		{args: []string{"availability", "update", "100200"}, file: stockJSON, want: `required flag(s) "site" not set`},
		{args: []string{"availability", "create", "100200", "--site", ".."}, file: stockJSON, want: `invalid site ".."`},
		{args: []string{"prices", "update", ".."}, file: priceJSON, want: `invalid PRICE_ID ".."`},
		{args: []string{"prices", "update", "bulk"}, file: "[" + priceJSON + "]",
			want: `invalid PRICE_ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"prices", "update"}, file: priceJSON, want: "accepts 1 arg(s), received 0"},
		{args: []string{"availability", "create", "a/b", "--site", "AcmeStore_NL"}, file: stockJSON,
			want: "a path segment of a write"},
	})
}

// TestWriteHelpEndsWithItsScopes checks the help of a declared write: what
// it sends, an upsert named so, for a bulk write how many items Emporix
// takes and how they are answered, how it is sent, and the write scope it
// requests, without the opt-in note of a read; its --file is required and
// says so. A group with writes states the scopes of its reads and writes.
func TestWriteHelpEndsWithItsScopes(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: []string{"prices", "update"}, want: []string{
			"or creates it with that id where none exists\n(an upsert; not with --partial",
			"PRICE_ID is the ID that `epx prices list --site S` shows.\n\nChanges data: on a terminal, epx asks " +
				"before it sends PUT.", "instead of sending it.\n\nRequires the price.price_manage scope.\n\nUsage:",
			"--partial         change only the fields", "sending it then needs --yes; required)"}},
		{args: []string{"availability", "bulk-update"}, want: []string{"(an upsert).\n\n" +
			"Emporix takes up to 200 items in one request (epx refuses more), carries out each on\nits own",
			"(one that failed with 500, 502, 503 or 504\nmay have been carried out all the same)",
			"(for example 5 for a 409 Conflict, 4 for a 404), else 1.\n\n" +
				"Changes data: on a terminal, epx asks before it sends PUT.",
			"instead of sending it.\n\nRequires the availability.availability_manage scope.\n\nUsage:"}},
		{args: []string{"prices", "bulk-create"}, want: []string{"a JSON array of prices", "Emporix takes up to 200 items"}},
		{args: []string{"availability"}, want: []string{"Its reads need no OAuth scope.\nIts writes require the " +
			"availability.availability_manage scope.\n\nUsage:"}},
		{args: []string{"prices"}, want: []string{"Reads and changes prices.", "bulk-create and bulk-update many",
			"bulk-create Create prices in bulk", "bulk-update Upsert prices in bulk",
			"Its reads require different OAuth scopes; the help of each names them.\nIts writes require the " +
				"price.price_manage scope.\n\nUsage:"}},
	} {
		code, help, stderr := runCLI(t, nil, append(tt.args, "--help")...)
		for _, want := range tt.want {
			if code != ExitOK || !strings.Contains(help, want) {
				t.Errorf("%q: exit %d, stderr %q, help %q; want %q", tt.args, code, stderr, help, want)
			}
		}
	}
}

// TestWritesReserveOnlyIDsTheyTake checks that a write that reserves values
// of an ID it does not take panics while the command tree is built, which
// every test of the tree would find, and that one with a write-capable scope
// among read-only ones is built, annotated with every scope it requests.
func TestWritesReserveOnlyIDsTheyTake(t *testing.T) {
	a := &app{getenv: func(string) string { return "" }}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a write that reserves an ID it does not take was built")
			}
		}()
		(&write{use: "create", method: http.MethodPost, req: request{path: "/things"}, body: &payload{},
			scopes: []string{"price.price_manage"}, reserved: map[string]string{bulkSegment: "bulk-create"}}).command(a)
	}()
	cmd := (&write{
		use: "create", method: http.MethodPost, req: request{path: "/things"}, body: &payload{},
		scopes: []string{"price.price_read", "price.price_manage"},
	}).command(a)
	if cmd.Annotations[writeScopesAnnotation] != "price.price_read price.price_manage" {
		t.Fatalf("annotations %q", cmd.Annotations)
	}
}

// TestBulkWritesReserveTheIDsOfTheirPaths checks which IDs of a write of one
// item address a bulk write instead: bulk, where the item is {id} right
// below the collection and the collection has a bulk write of its method.
func TestBulkWritesReserveTheIDsOfTheirPaths(t *testing.T) {
	bulk := &bulkWrites{path: "/things/{parent}/parts", update: "Upserts parts.", delete: "Deletes parts."}
	const part = "/things/{parent}/parts/{id}"
	for _, tt := range []struct {
		bulk   *bulkWrites
		method string
		path   string
		want   map[string]string
	}{
		{bulk: bulk, method: http.MethodPut, path: part, want: map[string]string{"bulk": "bulk-update"}},
		{bulk: bulk, method: http.MethodDelete, path: part, want: map[string]string{"bulk": "bulk-delete"}},
		{bulk: bulk, method: http.MethodPatch, path: part},
		{bulk: bulk, method: http.MethodPut, path: "/things/{parent}/parts/{inner}/{id}"},
		{method: http.MethodPut, path: part},
	} {
		if got := tt.bulk.reserved(tt.method, tt.path); !maps.Equal(got, tt.want) {
			t.Errorf("%s %s: reserved %v; want %v", tt.method, tt.path, got, tt.want)
		}
	}
}

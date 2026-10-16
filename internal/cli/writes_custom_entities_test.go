package cli

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// An instance of a custom entity type as create takes it.
const instanceJSON = `{"id":"123","name":{"en":"Report"},"mixins":{"attributes":{"pages":3}}}`

// customEntityWriteCases are the writes of custom entity types, with the Schema
// Service's manage scope, and of their instances, with the manage scope of
// their type, in lower case, and its read scope where Emporix checks their
// references.
func customEntityWriteCases() []writeCase {
	types, instances := []string{"schema.schema_manage"}, []string{"custom.document_manage"}
	checked := []string{"custom.document_manage", "custom.document_read"}
	return []writeCase{
		{
			args: []string{"custom-entities", "create"}, file: `{"id":"DOCUMENT","name":{"en":"Document"}}`,
			method: http.MethodPost, path: "/schema/acme1/custom-entities", scopes: types, lang: "*",
			status: http.StatusCreated, body: `{"id":"DOCUMENT"}`, want: "{\n  \"id\": \"DOCUMENT\"\n}\n",
		},
		{
			args: []string{"custom-entities", "update", "DOCUMENT"},
			file: `{"name":{"en":"Document"},"metadata":{"version":1}}`, method: http.MethodPut,
			path: "/schema/acme1/custom-entities/DOCUMENT", scopes: types, lang: "*", status: http.StatusNoContent,
		},
		{
			args: []string{"custom-entities", "delete", "DOCUMENT"}, method: http.MethodDelete,
			path: "/schema/acme1/custom-entities/DOCUMENT", scopes: types, confirmBy: "DOCUMENT",
			status: http.StatusNoContent,
		},
		{
			// The export package, as export prints it.
			args: []string{"custom-entities", "import"}, file: `{"exportedAt":"2026-10-01T07:05:18.857Z","data":"W10="}`,
			method: http.MethodPost, path: "/schema/acme1/custom-entities/import", scopes: types,
			status: http.StatusNoContent,
		},
		{
			args: []string{"custom-entities", "instances", "create", "DOCUMENT"}, file: instanceJSON,
			method: http.MethodPost, path: "/schema/acme1/custom-entities/DOCUMENT/instances", scopes: checked,
			lang: "*", status: http.StatusCreated, body: `{"id":"123"}`, want: "{\n  \"id\": \"123\"\n}\n",
		},
		{
			// The ID is escaped as a path segment; --validate-references is sent as given.
			args: []string{"custom-entities", "instances", "update", "DOCUMENT", "12 3", "--validate-references", "false"},
			file: `{"name":{"en":"Report"},"metadata":{"version":2}}`, method: http.MethodPut,
			path: "/schema/acme1/custom-entities/DOCUMENT/instances/12%203", query: "validateReferences=false",
			scopes: instances, lang: "*", status: http.StatusCreated, body: `{"id":"12 3"}`,
			want: "{\n  \"id\": \"12 3\"\n}\n",
		},
		{
			args: []string{"custom-entities", "instances", "patch", "DOCUMENT", "123"},
			file: `[{"op":"add","path":"/mixins/attributes/pages","value":4}]`, method: http.MethodPatch,
			path: "/schema/acme1/custom-entities/DOCUMENT/instances/123", scopes: checked, lang: "*",
			status: http.StatusNoContent,
		},
		{
			args: []string{"custom-entities", "instances", "delete", "DOCUMENT", "123"}, method: http.MethodDelete,
			path: "/schema/acme1/custom-entities/DOCUMENT/instances/123", scopes: instances, confirmBy: "123",
			status: http.StatusNoContent,
		},
		{
			args: []string{"custom-entities", "instances", "bulk-create", "DOCUMENT", "--validate-references", "true"},
			file: "[" + instanceJSON + "]", method: http.MethodPost,
			path: "/schema/acme1/custom-entities/DOCUMENT/instances/bulk", query: "validateReferences=true",
			scopes: checked, lang: "*", status: http.StatusMultiStatus, body: `[{"index":0,"code":201,"status":"Created"}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"index\": 0,\n    \"status\": \"Created\"\n  }\n]\n",
		},
		{
			args: []string{"custom-entities", "instances", "bulk-update", "DOCUMENT"},
			file: `[{"id":"123","name":{"en":"Report"}},{"id":"124","name":{"en":"Manual"}}]`, method: http.MethodPut,
			path: "/schema/acme1/custom-entities/DOCUMENT/instances/bulk", scopes: instances, lang: "*",
			status: http.StatusMultiStatus, body: `[{"index":0,"code":204},{"index":1,"code":201}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"index\": 0\n  },\n  {\n    \"code\": 201,\n    \"index\": 1\n  }\n]\n",
		},
		{
			args:   []string{"custom-entities", "instances", "bulk-patch", "DOCUMENT"},
			file:   `[{"id":"123","data":[{"op":"replace","path":"/name/en","value":"Report"}]}]`,
			method: http.MethodPatch, path: "/schema/acme1/custom-entities/DOCUMENT/instances/bulk", scopes: instances,
			lang: "*", status: http.StatusMultiStatus, body: `[{"index":0,"id":"123","code":204}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"123\",\n    \"index\": 0\n  }\n]\n",
		},
		{
			args: []string{"custom-entities", "instances", "bulk-delete", "DOCUMENT"}, file: `["123","124"]`,
			method: http.MethodDelete, path: "/schema/acme1/custom-entities/DOCUMENT/instances/bulk", scopes: instances,
			confirmBy: "DOCUMENT", many: "the instances of a custom entity type that --file names",
			status: http.StatusMultiStatus, body: `[{"index":0,"code":204},{"index":1,"code":204}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"index\": 0\n  },\n  {\n    \"code\": 204,\n    \"index\": 1\n  }\n]\n",
		},
	}
}

// TestCustomInstanceWritesRefuseInvalidInput checks that an instance write
// whose ID would address the bulk write of its collection, without its type,
// with a --validate-references other than true or false, with
// --content-language (Emporix refuses a string in any language), a
// bulk-delete of no id or of more than 1000, is a usage error (exit 2) that
// sends nothing, also for a dry run.
func TestCustomInstanceWritesRefuseInvalidInput(t *testing.T) {
	ids := func(n int) string { return `[` + strings.Repeat(`"123",`, n-1) + `"123"]` }
	for _, tt := range []struct {
		args []string
		file string
		want string
	}{
		{args: []string{"update", "DOCUMENT", "bulk"}, file: instanceJSON,
			want: `invalid ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"patch", "DOCUMENT", "bulk"}, file: `[]`,
			want: `invalid ID "bulk": it would be sent as bulk-patch, not to one item`},
		{args: []string{"delete", "DOCUMENT", "bulk"}, want: `invalid ID "bulk": it would be sent as bulk-delete`},
		{args: []string{"delete", "DOCUMENT"}, want: "accepts 2 arg(s), received 1"},
		{args: []string{"create"}, file: instanceJSON, want: "accepts 1 arg(s), received 0"},
		{args: []string{"create", ".."}, file: instanceJSON, want: `invalid CUSTOM_ENTITY_ID ".."`},
		{args: []string{"create", "DOCUMENT", "--validate-references", "no"}, file: instanceJSON,
			want: `invalid argument "no" for "--validate-references" flag: use true or false`},
		{args: []string{"update", "DOCUMENT", "123", "--content-language", "en"}, file: instanceJSON,
			want: "unknown flag: --content-language"},
		{args: []string{"bulk-delete", "DOCUMENT"}, file: `[]`, want: "a JSON array of one or more items"},
		{args: []string{"bulk-delete", "DOCUMENT"}, file: ids(1001),
			want: "a bulk write sends at most 1000 items in one request, not 1001"},
		{args: []string{"bulk-create", "DOCUMENT"}, file: "[" + strings.Repeat(instanceJSON+",", 200) + instanceJSON + "]",
			want: "a bulk write sends at most 200 items in one request, not 201"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			args := append(append([]string{"custom-entities", "instances"}, tt.args...), flag)
			if tt.file != "" {
				args = append(args, "--file", writeBody(t, tt.file))
			}
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

// TestCustomInstanceBulkDeleteSendsUpTo1000IDs checks that bulk-delete sends
// as many ids as Emporix takes in one request, more than its other bulk
// writes take.
func TestCustomInstanceBulkDeleteSendsUpTo1000IDs(t *testing.T) {
	const most = 1000
	results := make([]string, most)
	for i := range results {
		results[i] = `{"index":` + strconv.Itoa(i) + `,"code":204}`
	}
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/custom-entities/DOCUMENT/instances/bulk",
		respond(http.StatusMultiStatus, "["+strings.Join(results, ",")+"]"))
	ids := `[` + strings.Repeat(`"123",`, most-1) + `"123"]`
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "custom-entities", "instances", "bulk-delete",
		"DOCUMENT", "--file", writeBody(t, ids), "--yes")
	if code != ExitOK || !strings.Contains(stdout, `"index": 999`) || stderr != "" {
		t.Fatalf("exit %d, stdout %.80q, stderr %q; want exit 0 and the result of each id", code, stdout, stderr)
	}
}

// TestCustomInstanceBulkDeleteListsTheIDsThatFailed checks that a bulk delete
// of instances that Emporix answers with a failure for some of them prints
// the response, lists those on standard error and exits non-zero.
func TestCustomInstanceBulkDeleteListsTheIDsThatFailed(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/custom-entities/DOCUMENT/instances/bulk", respond(http.StatusMultiStatus,
		`[{"index":0,"code":204},{"index":1,"code":404,"status":"Not Found","message":"no instance 124"}]`))
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "custom-entities", "instances", "bulk-delete",
		"DOCUMENT", "--file", writeBody(t, `["123","124"]`), "--yes")
	const want = "item 1: 404 Not Found: no instance 124\n" +
		"Error: DELETE /schema/acme1/custom-entities/DOCUMENT/instances/bulk: 1 of 2 items failed\n"
	if code != ExitNotFound || !strings.HasPrefix(stdout, "[\n  {\n") || stderr != want {
		t.Fatalf("exit %d, stdout %q, stderr:\n%s\nwant exit %d and:\n%s", code, stdout, stderr, ExitNotFound, want)
	}
}

// TestCustomEntityWriteHelp checks the help of the writes of custom entities:
// an upsert is named so, what --validate-references does without it and what
// checking references needs, the scopes requested (those of instances, of
// their type), where the arguments come from, the most items of a bulk write,
// what Emporix does with ids, localized text, mixins, unknown items and the
// scopes of new and deleted types, and that each group names the scopes of
// its reads and writes and the command lines of its reads.
func TestCustomEntityWriteHelp(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: []string{"custom-entities", "create"}, want: []string{"The id has capital letters, digits and\n" +
			"underscores, the first a letter.", "which it grants only minutes\nlater: until then, " +
			"the instance commands of the type exit with code 3.", "sent with Content-Language: *. Emporix refuses " +
			"a string in\nits place."}},
		{args: []string{"custom-entities", "delete"}, want: []string{"answers an unknown ID as deleted, not as not found.",
			"remain granted: the instances of a deleted\ntype then read as an empty list."}},
		{args: []string{"custom-entities", "import"}, want: []string{"the references of its schemas then name the copy."}},
		{args: []string{"custom-entities", "update"}, want: []string{"or creates the type where none\nexists (an upsert)",
			"Requires the schema.schema_manage scope.\n\nUsage:"}},
		{args: []string{"custom-entities", "instances", "update"}, want: []string{"(an upsert)",
			"CUSTOM_ENTITY_ID is the ID that `epx custom-entities list` shows, and ID the one that\n" +
				"`epx custom-entities instances list CUSTOM_ENTITY_ID` shows.",
			"(true without it), requesting custom.<type>_read where it is true or not given: true or false",
			"Emporix checks the items that the references of the instance name (unless\n" +
				"--validate-references false) with epx's token, which reads the instances of CUSTOM_ENTITY_ID\n" +
				"alone: a reference to an instance of another custom entity type is then expected to fail\n" +
				"with 400 Bad Request,",
			"\n\nWhere --validate-references is true or not given, epx also requests\n" +
				"the custom.<type>_read scope.\n\nRequires the custom.<type>_manage scope.\n<type> is the custom " +
				"entity type in lower case (custom.document_manage for DOCUMENT).\n\nUsage:"}},
		{args: []string{"custom-entities", "instances", "delete"}, want: []string{
			"Requires the custom.<type>_manage scope.\n<type> is the custom entity type in lower case " +
				"(custom.document_manage for DOCUMENT).\n\nUsage:"}},
		{args: []string{"custom-entities", "instances", "patch"}, want: []string{"an add whose path ends in\n" +
			"/- appends to an array", "(true without it), requesting custom.<type>_read where it is true or not given"}},
		{args: []string{"custom-entities", "instances", "create"}, want: []string{
			`{"mixins": {"SCHEMA_ID": {…}}, "metadata": {"mixins": {"SCHEMA_ID": "URL"}}}`,
			"A localized\nattribute, unlike the name, is a list of translations, such as\n" +
				`[{"language": "en", "value": "Report"}].`, "sent with Content-Language: *"}},
		{args: []string{"custom-entities", "instances", "bulk-patch"}, want: []string{
			"(false without it, although its reference says true), requesting custom.<type>_read where it is true:",
			"With --validate-references true, Emporix checks",
			"sent with Content-Language: *"}},
		{args: []string{"custom-entities", "instances", "list"}, want: []string{"an unknown one has no scope that " +
			"Emporix grants (exit code 3), and a type created in the last minutes has none yet."}},
		{args: []string{"custom-entities", "instances", "bulk-create"}, want: []string{
			"(false without it), requesting custom.<type>_read where it is true:",
			"With --validate-references true, Emporix checks the items that the references of each\n",
			"instance with a reference to an item that does not exist fails with 500 Internal Server\n" +
				"Error (nothing is created)",
			"\n\nWhere --validate-references is true, epx also requests\nthe custom.<type>_read scope.\n\nRequires the " +
				"custom.<type>_manage scope.",
			"Emporix takes up to 200 items in one request (epx refuses more)"}},
		{args: []string{"custom-entities", "instances", "bulk-delete"}, want: []string{
			"CUSTOM_ENTITY_ID is the ID that `epx custom-entities list` shows.\n",
			"Emporix takes up to 1000 items in one request (epx refuses more), carries out each on\nits own",
			"Emporix answers an unknown id as deleted."}},
		{args: []string{"custom-entities", "instances", "search"}, want: []string{
			"`epx custom-entities instances list CUSTOM_ENTITY_ID`"}},
		{args: []string{"custom-entities", "instances"}, want: []string{"Its reads require the " +
			"custom.<type>_read scope.\nIts writes require the custom.<type>_manage scope.\n<type> is the custom " +
			"entity type in lower case (custom.document_manage, custom.document_read for DOCUMENT).\n\nUsage:"}},
		{args: []string{"custom-entities"}, want: []string{"export exports types with their schemas (opt-in), which\n" +
			"import imports.\n`epx custom-entities instances` reads and changes the instances of a type.\n\n",
			"Its writes require different OAuth scopes"}},
	} {
		code, help, stderr := runCLI(t, nil, append(slices.Clone(tt.args), "--help")...)
		for _, want := range tt.want {
			if code != ExitOK || !strings.Contains(help, want) {
				t.Errorf("%q: exit %d, stderr %q, help %q; want %q", tt.args, code, stderr, help, want)
			}
		}
	}
}

package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// A segment as create takes it, and an assignment's mixins and version.
const (
	segmentJSON    = `{"id":"s1","name":{"en":"Gold"},"siteCode":"AcmeStore_NL","status":"ACTIVE"}`
	assignmentJSON = `{"mixins":{"segmentAttributes":{"tier":"GOLD"}},"metadata":{"version":2}}`
)

// bulkDone is the answer to a bulk write of two items that Emporix carried
// out without content, numbered from 1 as the Customer Segment Service's
// reference numbers them, and doneJSON the answer as epx prints it.
const (
	bulkDone = `[{"index":1,"id":"s1","code":204,"status":"OK"},{"index":2,"id":"s2","code":204,"status":"OK"}]`
	doneJSON = "[\n  {\n    \"code\": 204,\n    \"id\": \"s1\",\n    \"index\": 1,\n    \"status\": \"OK\"\n  },\n" +
		"  {\n    \"code\": 204,\n    \"id\": \"s2\",\n    \"index\": 2,\n    \"status\": \"OK\"\n  }\n]\n"
)

// segmentServiceWrites are the writes of the Customer Segment Service:
// segments and their customer, IAM group and item assignments.
func segmentServiceWrites() []writeCase {
	return slices.Concat(segmentItself(), segmentAssignments())
}

// segmentItself are the writes of segments: those that send localized text
// send maps of translations with Content-Language: *.
func segmentItself() []writeCase {
	manage := []string{segmentManageScope}
	const segments = "/customer-segment/acme1/segments"
	return []writeCase{
		{
			args: []string{"segments", "create"}, file: segmentJSON, method: http.MethodPost, path: segments,
			scopes: manage, lang: "*", status: http.StatusCreated, body: `{"id":"s1"}`, want: "{\n  \"id\": \"s1\"\n}\n",
		},
		{
			args: []string{"segments", "update", "s 1"}, file: `{"name":{"en":"Gold"},"metadata":{"version":3}}`,
			method: http.MethodPut, path: segments + "/s%201", scopes: manage, lang: "*", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "patch", "s1"}, file: `[{"op":"replace","path":"/status","value":"INACTIVE"}]`,
			method: http.MethodPatch, path: segments + "/s1", scopes: manage, lang: "*", status: http.StatusNoContent,
		},
		{
			// Without --with-assignments, Emporix keeps a segment that has assignments.
			args: []string{"segments", "delete", "s1"}, method: http.MethodDelete, path: segments + "/s1",
			query: "forceDelete=false", scopes: manage, confirmBy: "s1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "delete", "s1", "--with-assignments"}, method: http.MethodDelete, path: segments + "/s1",
			query: "forceDelete=true", scopes: manage, confirmBy: "s1",
			many: "the segment and its item, customer and IAM group assignments", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "bulk-create"}, file: "[" + segmentJSON + "," + segmentJSON + "]",
			method: http.MethodPost, path: segments + "/bulk", scopes: manage, lang: "*", status: http.StatusMultiStatus,
			body: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "bulk-update"}, file: "[" + segmentJSON + "," + segmentJSON + "]",
			method: http.MethodPut, path: segments + "/bulk", scopes: manage, lang: "*", status: http.StatusMultiStatus,
			body: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "bulk-delete", "--with-assignments"}, file: `["s1","s2"]`, method: http.MethodDelete,
			path: segments + "/bulk", query: "forceDelete=true", scopes: manage, confirmBy: "acme1",
			many: "the segments that --file names", status: http.StatusMultiStatus,
			body: bulkDone, want: doneJSON,
		},
	}
}

// segmentAssignments are the writes of what is assigned to a segment: an
// assignment sends {} without --file.
func segmentAssignments() []writeCase {
	manage := []string{segmentManageScope}
	const segment = "/customer-segment/acme1/segments/s%201"
	return []writeCase{
		{
			args: []string{"segments", "customers", "update", "s 1", "c1"}, sent: "{}", method: http.MethodPut,
			path: segment + "/customers/c1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "customers", "update", "s 1", "c1"}, file: assignmentJSON, method: http.MethodPut,
			path: segment + "/customers/c1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "customers", "update-b2b", "s 1", "c 1", "le 1"}, sent: "{}",
			method: http.MethodPut, path: segment + "/customers/c%201/le%201", scopes: manage, status: http.StatusCreated,
		},
		{
			args: []string{"segments", "customers", "update-b2b", "s 1", "c1", "le1"}, file: assignmentJSON,
			method: http.MethodPut, path: segment + "/customers/c1/le1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "customers", "delete", "s 1", "c1"}, method: http.MethodDelete,
			path: segment + "/customers/c1", scopes: manage, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "customers", "delete-b2b", "s 1", "c1", "le1"}, method: http.MethodDelete,
			path: segment + "/customers/c1/le1", scopes: manage, confirmBy: "le1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "customers", "bulk-update", "s 1"},
			file: `[{"customerId":"c1"},{"customerId":"c2","legalEntityId":"le1"}]`, method: http.MethodPut,
			path: segment + "/customers/bulk", scopes: manage, status: http.StatusMultiStatus, body: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "customers", "bulk-delete", "s 1"},
			file: `[{"customerId":"c1"},{"customerId":"c2","legalEntityId":"le1"}]`, method: http.MethodDelete,
			path: segment + "/customers/bulk", scopes: manage, confirmBy: "s 1",
			many: "the customer assignments of a segment that --file names", status: http.StatusMultiStatus,
			body: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "groups", "update", "s 1", "g1"}, sent: "{}", method: http.MethodPut,
			path: segment + "/groups/g1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "groups", "update", "s 1", "g1"}, file: assignmentJSON, method: http.MethodPut,
			path: segment + "/groups/g1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "groups", "delete", "s 1", "g1"}, method: http.MethodDelete,
			path: segment + "/groups/g1", scopes: manage, confirmBy: "g1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "items", "update", "s 1", "PRODUCT", "100200"}, sent: "{}", method: http.MethodPut,
			path: segment + "/items/PRODUCT/100200", scopes: manage, status: http.StatusCreated,
		},
		{
			args: []string{"segments", "items", "update", "s 1", "CATEGORY", "c1"}, file: assignmentJSON,
			method: http.MethodPut, path: segment + "/items/CATEGORY/c1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "items", "delete", "s 1", "CATEGORY", "c1"}, method: http.MethodDelete,
			path: segment + "/items/CATEGORY/c1", scopes: manage, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "items", "bulk-update", "s 1", "PRODUCT"}, file: `[{"id":"100200"},{"id":"100300"}]`,
			method: http.MethodPut, path: segment + "/items/PRODUCT/bulk", scopes: manage, status: http.StatusMultiStatus,
			body: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "items", "bulk-delete", "s 1", "PRODUCT"}, file: `["100200","100300"]`,
			method: http.MethodDelete, path: segment + "/items/PRODUCT/bulk", scopes: manage, confirmBy: "PRODUCT",
			many:   "the item assignments of a segment that --file names",
			status: http.StatusMultiStatus, body: bulkDone, want: doneJSON,
		},
	}
}

// TestSegmentWriteUsageErrorsSendNothing checks that the segment writes
// refuse what they cannot send as usage errors (exit 2) that send nothing,
// also for a dry run: an item TYPE other than PRODUCT or CATEGORY, bulk as
// the ID of a write of one item that a bulk write of its collection would
// get (a legal entity, a group or the segment of a patch may be called so),
// a bulk write without a JSON array of items, a patch whose --file is no
// JSON array (Emporix answers an object as done, changing nothing), and
// --content-language, as Emporix refuses a string in any language.
func TestSegmentWriteUsageErrorsSendNothing(t *testing.T) {
	const ids = `["s1"]`
	assertRefused(t, []refusedWrite{
		{args: []string{"segments", "items", "update", "s1", "product", "100200"},
			want: `invalid TYPE "product": use PRODUCT or CATEGORY`},
		{args: []string{"segments", "items", "bulk-delete", "s1", "BRAND"}, file: ids,
			want: `invalid TYPE "BRAND": use PRODUCT or CATEGORY`},
		{args: []string{"segments", "update", "bulk"}, file: ids,
			want: `invalid ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"segments", "delete", "bulk"}, want: `invalid ID "bulk": it would be sent as bulk-delete`},
		{args: []string{"segments", "customers", "update", "s1", "bulk"},
			want: `invalid CUSTOMER "bulk": it would be sent as bulk-update`},
		{args: []string{"segments", "customers", "delete", "s1", "bulk"},
			want: `invalid CUSTOMER "bulk": it would be sent as bulk-delete`},
		{args: []string{"segments", "items", "update", "s1", "PRODUCT", "bulk"},
			want: `invalid ID "bulk": it would be sent as bulk-update`},
		{args: []string{"segments", "items", "delete", "s1", "PRODUCT", "bulk"},
			want: `invalid ID "bulk": it would be sent as bulk-delete`},
		{args: []string{"segments", "bulk-delete"}, file: `"s1"`,
			want: "a bulk write sends a JSON array of one or more items"},
		{args: []string{"segments", "customers", "bulk-update", "s1"}, want: `required flag(s) "file" not set`},
		{args: []string{"segments", "delete", "s1", "--force"}, want: "unknown flag: --force"},
		{args: []string{"segments", "patch", "s1"}, file: `{"status":"INACTIVE"}`, want: "not a JSON array"},
		{args: []string{"segments", "patch", "s1"}, file: `"INACTIVE"`, want: "not a JSON array"},
		{args: []string{"segments", "create", "--content-language", "en"}, file: `{"name":"Gold"}`,
			want: "unknown flag: --content-language"},
		{args: []string{"segments", "bulk-update", "--content-language", "en"}, file: "[" + segmentJSON + "]",
			want: "unknown flag: --content-language"},
	})
}

// TestSegmentWritesTakeBulkWhereNoBulkWriteWouldGetIt checks that the writes
// of one item refuse bulk only as the ID that a bulk write of the collection
// would get: a legal entity, an IAM group and a patched segment may be
// called so, as Emporix sends no bulk write to their paths.
func TestSegmentWritesTakeBulkWhereNoBulkWriteWouldGetIt(t *testing.T) {
	for _, args := range [][]string{
		{"segments", "customers", "update-b2b", "s1", "c1", "bulk"},
		{"segments", "customers", "delete-b2b", "s1", "c1", "bulk"},
		{"segments", "groups", "update", "s1", "bulk"}, {"segments", "groups", "delete", "s1", "bulk"},
		{"segments", "patch", "bulk", "--file", writeBody(t, `[{"op":"remove","path":"/description"}]`)},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(slices.Clone(args), "--dry-run")...)
		if code != ExitOK || !strings.Contains(stdout, "/bulk") || stderr != "" {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want its request printed", args, code, stdout, stderr)
		}
	}
}

// TestSegmentBulkWritesListTheItemsThatFailed checks that a bulk write of the
// Customer Segment Service reads the results of its items numbered from 1, as
// its reference's examples number them: an item that failed is named by its
// position in --file, from 0, and fails the write (exit 1).
func TestSegmentBulkWritesListTheItemsThatFailed(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/customer-segment/acme1/segments/bulk", respond(http.StatusMultiStatus,
		`[{"index":1,"id":"628cd20c6e8b2432b6346ca6","code":201,"status":"Created"},`+
			`{"index":2,"code":400,"status":"Bad Request","message":"Validation problem with request body."}]`))
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "segments", "bulk-create", "--file",
		writeBody(t, "["+segmentJSON+","+segmentJSON+"]"), "--yes")
	want := "item 1: 400 Bad Request: Validation problem with request body.\n" +
		"Error: POST /customer-segment/acme1/segments/bulk: 1 of 2 items failed\n"
	if code != ExitError || !strings.HasPrefix(stdout, "[\n  {\n") || stderr != want {
		t.Fatalf("exit %d, stdout %q, stderr %q; want exit 1 and %q", code, stdout, stderr, want)
	}
}

// TestSegmentGroupsStateTheirScopes checks the help of the segment groups
// with writes, and that the reads of a child collection in a group of its
// own name the list of that group as the source of their arguments.
func TestSegmentGroupsStateTheirScopes(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"segments"}, want: "Its reads require different OAuth scopes; the help of each names them.\n" +
			"Opt-in: epx requests a write-capable scope for a read only when given --allow-write-scope.\n" +
			"Its writes require the customersegment.segment_manage scope.\n\nUsage:"},
		{args: []string{"segments", "items"}, want: "Its reads require the customersegment.segment_read scope.\n" +
			"Its writes require the customersegment.segment_manage scope.\n\nUsage:"},
		{args: []string{"segments", "customers", "get-b2b"}, want: "CUSTOMER and LEGAL_ENTITY are the CUSTOMER and " +
			"LEGAL ENTITY that `epx segments customers list SEGMENT_ID` shows"},
		{args: []string{"segments", "groups", "search"}, want: "`epx segments groups list SEGMENT_ID`"},
		{args: []string{"segments", "delete"}, want: "unless --with-assignments deletes those assignments with it"},
		{args: []string{"segments", "delete"}, want: "a segment with only\nIAM groups assigned is deleted without " +
			"--with-assignments.\nEmporix answers an unknown ID as one that it deleted"},
		{args: []string{"segments", "bulk-delete"}, want: "(its IAM group assignments go either way)"},
		{args: []string{"segments", "create"}, want: "sent with Content-Language: *. Emporix refuses a string " +
			"in\nits place."},
		{args: []string{"segments", "update"}, want: "members it leaves out, such as a description, are removed."},
		{args: []string{"segments", "patch"}, want: "raises the segment's metadata.version."},
		{args: []string{"segments", "patch"}, want: "a path to one translation, such as /name/de, takes a string"},
		{args: []string{"segments", "customers", "update"}, want: "Emporix refuses to assign one of the other kind."},
		{args: []string{"segments", "customers", "bulk-update"}, want: "Emporix refuses to assign one of the other kind."},
		{args: []string{"segments", "groups", "delete"}, want: "Emporix answers one that is not\nassigned as removed."},
		{args: []string{"segments", "items", "update"}, want: "without it, epx sends an empty object"},
		{args: []string{"segments", "items", "update"}, want: "Emporix refuses an ID of no product or category of TYPE"},
		{args: []string{"segments", "items", "bulk-update"}, want: "Emporix refuses an ID of no product or category"},
		{args: []string{"segments", "match"}, want: "a product\nnamed twice is listed twice, and unknown products " +
			"are left out."},
	} {
		code, help, stderr := runCLI(t, nil, append(tt.args, "--help")...)
		if code != ExitOK || !strings.Contains(help, tt.want) {
			t.Errorf("%q: exit %d, stderr %q, help %q; want %q", tt.args, code, stderr, help, tt.want)
		}
	}
}

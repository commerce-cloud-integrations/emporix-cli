package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

const (
	segmentJSON    = `{"id":"s1","name":{"en":"Gold"},"siteCode":"AcmeStore_NL","status":"ACTIVE"}`
	assignmentJSON = `{"mixins":{"segmentAttributes":{"tier":"GOLD"}},"metadata":{"version":2}}`
)

const (
	bulkDone = `[{"index":1,"id":"s1","code":204,"status":"OK"},{"index":2,"id":"s2","code":204,"status":"OK"}]`
	doneJSON = "[\n  {\n    \"code\": 204,\n    \"id\": \"s1\",\n    \"index\": 1,\n    \"status\": \"OK\"\n  },\n" +
		"  {\n    \"code\": 204,\n    \"id\": \"s2\",\n    \"index\": 2,\n    \"status\": \"OK\"\n  }\n]\n"
)

func segmentServiceWrites() []writeCase {
	return slices.Concat(segmentItself(), segmentAssignments())
}

func segmentItself() []writeCase {
	manage := []string{segmentManageScope}
	const segments = "/customer-segment/acme1/segments"
	return []writeCase{
		{
			args: []string{"segments", "create"}, file: segmentJSON, method: http.MethodPost, path: segments,
			scopes: manage, lang: "*", status: http.StatusCreated, response: `{"id":"s1"}`, want: "{\n  \"id\": \"s1\"\n}\n",
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
			args: []string{"segments", "delete", "s1"}, method: http.MethodDelete, path: segments + "/s1",
			query: "forceDelete=false", scopes: manage, confirmBy: "s1", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "delete", "s1", "--with-assignments"}, method: http.MethodDelete, path: segments + "/s1",
			query: "forceDelete=true", scopes: manage, confirmBy: "s1",
			deletedMany: "the segment and its item, customer and IAM group assignments", status: http.StatusNoContent,
		},
		{
			args: []string{"segments", "bulk-create"}, file: "[" + segmentJSON + "," + segmentJSON + "]",
			method: http.MethodPost, path: segments + "/bulk", scopes: manage, lang: "*", status: http.StatusMultiStatus,
			response: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "bulk-update"}, file: "[" + segmentJSON + "," + segmentJSON + "]",
			method: http.MethodPut, path: segments + "/bulk", scopes: manage, lang: "*", status: http.StatusMultiStatus,
			response: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "bulk-delete", "--with-assignments"}, file: `["s1","s2"]`, method: http.MethodDelete,
			path: segments + "/bulk", query: "forceDelete=true", scopes: manage, confirmBy: "acme1",
			deletedMany: "the segments that --file names", status: http.StatusMultiStatus,
			response: bulkDone, want: doneJSON,
		},
	}
}

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
			path: segment + "/customers/bulk", scopes: manage, status: http.StatusMultiStatus, response: bulkDone,
			want: doneJSON,
		},
		{
			args: []string{"segments", "customers", "bulk-delete", "s 1"},
			file: `[{"customerId":"c1"},{"customerId":"c2","legalEntityId":"le1"}]`, method: http.MethodDelete,
			path: segment + "/customers/bulk", scopes: manage, confirmBy: "s 1",
			deletedMany: "the customer assignments of a segment that --file names", status: http.StatusMultiStatus,
			response: bulkDone, want: doneJSON,
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
			response: bulkDone, want: doneJSON,
		},
		{
			args: []string{"segments", "items", "bulk-delete", "s 1", "PRODUCT"}, file: `["100200","100300"]`,
			method: http.MethodDelete, path: segment + "/items/PRODUCT/bulk", scopes: manage, confirmBy: "PRODUCT",
			deletedMany: "the item assignments of a segment that --file names",
			status:      http.StatusMultiStatus, response: bulkDone, want: doneJSON,
		},
	}
}

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

func TestSegmentGroupsStateTheirScopes(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"segments"}, want: "Its reads require different OAuth scopes; the help of each names them. " +
			"Opt-in: epx requests a write-capable scope for a read only when given --allow-write-scope. " +
			"Its writes require the customersegment.segment_manage scope.\n\nUsage:"},
		{args: []string{"segments", "items"}, want: "Its reads require the customersegment.segment_read scope. " +
			"Its writes require the customersegment.segment_manage scope.\n\nUsage:"},
		{args: []string{"segments", "customers", "get-b2b"}, want: "CUSTOMER and LEGAL_ENTITY are the CUSTOMER and " +
			"LEGAL ENTITY that `epx segments customers list SEGMENT_ID` shows"},
		{args: []string{"segments", "groups", "search"}, want: "`epx segments groups list SEGMENT_ID`"},
		{args: []string{"segments", "delete"}, want: "unless --with-assignments deletes those assignments with it"},
		{args: []string{"segments", "delete"}, want: "a segment with only IAM groups assigned is deleted without " +
			"--with-assignments. Emporix answers an unknown ID as one that it deleted"},
		{args: []string{"segments", "bulk-delete"}, want: "(its IAM group assignments go either way)"},
		{args: []string{"segments", "create"}, want: "sent with Content-Language: *. Emporix refuses a string " +
			"in its place."},
		{args: []string{"segments", "update"}, want: "members it leaves out, such as a description, are removed."},
		{args: []string{"segments", "patch"}, want: "raises the segment's metadata.version."},
		{args: []string{"segments", "patch"}, want: "a path to one translation, such as /name/de, takes a string"},
		{args: []string{"segments", "customers", "update"}, want: "Emporix refuses to assign one of the other kind."},
		{args: []string{"segments", "customers", "bulk-update"}, want: "Emporix refuses to assign one of the other kind."},
		{args: []string{"segments", "groups", "delete"}, want: "Emporix answers one that is not assigned as removed."},
		{args: []string{"segments", "items", "update"}, want: "without it, epx sends an empty object"},
		{args: []string{"segments", "items", "update"}, want: "Emporix refuses an ID of no product or category of TYPE"},
		{args: []string{"segments", "items", "bulk-update"}, want: "Emporix refuses an ID of no product or category"},
		{args: []string{"segments", "match"}, want: "a product named twice is listed twice, and unknown products " +
			"are left out."},
	} {
		code, help, stderr := runCLI(t, nil, append(tt.args, "--help")...)
		if code != ExitOK || !containsFlat(help, tt.want) {
			t.Errorf("%q: exit %d, stderr %q, help %q; want %q", tt.args, code, stderr, help, tt.want)
		}
	}
}

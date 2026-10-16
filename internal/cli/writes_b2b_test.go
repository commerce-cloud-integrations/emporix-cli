package cli

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

// A legal entity, a contact assignment, a location and a vendor as create
// takes them, each naming the items it refers to as {"id": ...}, and as
// update takes them: without an id, with the version that a read gave.
const (
	b2bLegalEntityJSON = `{"id":"le1","name":"Acme Corp","type":"COMPANY","entitiesAddresses":[{"id":"loc1"}],` +
		`"approvalGroup":[{"id":"ca1"}]}`
	b2bLegalEntityUpdateJSON = `{"name":"Acme Corp","type":"COMPANY","metadata":{"version":2}}`
	b2bAssignmentJSON        = `{"legalEntity":{"id":"le1"},"customer":{"id":"c1"},"type":"BILLING","primary":true}`
	b2bAssignmentUpdateJSON  = `{"legalEntity":{"id":"le1"},"customer":{"id":"c1"},"type":"CONTACT",` +
		`"metadata":{"version":2}}`
	b2bLocationJSON = `{"name":"Main HQ","type":"HEADQUARTER",` +
		`"contactDetails":{"city":"Utrecht","countryCode":"NL"}}`
	b2bLocationUpdateJSON = `{"type":"WAREHOUSE","contactDetails":{"countryCode":"NL"},"metadata":{"version":2}}`
	b2bVendorJSON         = `{"name":"Acme Supplies","type":"COMPANY","locations":[{"id":"vl1"}]}`
	b2bVendorUpdateJSON   = `{"name":"Acme Supplies","type":"SUBSIDIARY","parentId":"v0","metadata":{"version":2}}`
)

// b2bCreatedJSON is how epx prints the id with which Emporix answers a
// create or an upsert that created the item.
func b2bCreatedJSON(id string) string { return "{\n  \"id\": \"" + id + "\"\n}\n" }

// b2bWrites are the writes of legal entities, contact assignments, customer
// locations, vendors and vendor locations: create, update (an upsert) and
// delete of each, with the manage scope that the reference names for it. A
// legal entity's delete also deletes its subsidiaries, assignments and IAM
// customer groups, which its prompt says.
func b2bWrites() []writeCase {
	entities := b2bWriteCases("legal-entities", "/customer-management/acme1/legal-entities",
		"customermanagement.legalentity_manage", b2bLegalEntityJSON, b2bLegalEntityUpdateJSON, "le1")
	for i := range entities {
		if entities[i].method == http.MethodDelete {
			entities[i].many = "the legal entity with its subsidiaries, assignments and IAM customer groups"
		}
	}
	return slices.Concat(
		entities,
		b2bWriteCases("contact-assignments", "/customer-management/acme1/contact-assignments",
			"customermanagement.contactassignment_manage", b2bAssignmentJSON, b2bAssignmentUpdateJSON, "ca1"),
		b2bWriteCases("customer-locations", "/customer-management/acme1/locations",
			"customermanagement.location_manage", b2bLocationJSON, b2bLocationUpdateJSON, "loc 1"),
		b2bWriteCases("vendors", "/vendor/acme1/vendors", "vendor.vendor_manage", b2bVendorJSON,
			b2bVendorUpdateJSON, "v1"),
		b2bWriteCases("vendor-locations", "/vendor/acme1/locations", "vendor.vendor_manage", b2bLocationJSON,
			b2bLocationUpdateJSON, "vl1"),
	)
}

// b2bWriteCases are the write cases of group, whose items are at path, each
// written with scope: create, with created, of the item id, which Emporix
// answers with its id; update id, which replaces it with updated (204 No
// Content); and delete id, which sends no body and is confirmed by typing id.
func b2bWriteCases(group, path, scope, created, updated, id string) []writeCase {
	scopes := []string{scope}
	item := path + "/" + url.PathEscape(id)
	return []writeCase{
		{
			args: []string{group, createUse}, file: created, method: http.MethodPost, path: path, scopes: scopes,
			status: http.StatusCreated, body: `{"id":"` + id + `"}`, want: b2bCreatedJSON(id),
		},
		{
			args: []string{group, "update", id}, file: updated, method: http.MethodPut, path: item, scopes: scopes,
			status: http.StatusNoContent,
		},
		{
			args: []string{group, "delete", id}, method: http.MethodDelete, path: item, scopes: scopes,
			confirmBy: id, status: http.StatusNoContent,
		},
	}
}

// TestB2BWritesUpsertAndDelete checks that update answers with the id of the
// item that Emporix created where none had the id (201 Created), and that
// delete of an unknown contact assignment is not found (exit 4) and prints
// nothing on standard output, while Emporix answers that of an unknown legal
// entity as one that it deleted (204 No Content, exit 0).
func TestB2BWritesUpsertAndDelete(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/vendor/acme1/vendors/v2", respond(http.StatusCreated, `{"id":"v2"}`))
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "vendors", "update", "v2", "--file",
		writeBody(t, b2bVendorUpdateJSON), "--yes")
	if code != ExitOK || stdout != b2bCreatedJSON("v2") || stderr != "" {
		t.Errorf("upsert: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake = newFakeEmporix(t)
	fake.route("/customer-management/acme1/contact-assignments/ca9", respond(http.StatusNotFound,
		`{"code":404,"status":"Not Found","message":"Contact assignment not found."}`))
	code, stdout, stderr = runIn(t, writeEnv(fake, false, ""), "contact-assignments", "delete", "ca9", "--yes")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "404 Not Found") {
		t.Errorf("unknown: exit %d, stdout %q, stderr %q; want exit 4", code, stdout, stderr)
	}
	fake = newFakeEmporix(t)
	fake.route("/customer-management/acme1/legal-entities/le9", respond(http.StatusNoContent, ""))
	code, stdout, stderr = runIn(t, writeEnv(fake, false, ""), "legal-entities", "delete", "le9", "--yes")
	if code != ExitOK || stdout != "" || stderr != "" {
		t.Errorf("unknown legal entity: exit %d, stdout %q, stderr %q; want exit 0", code, stdout, stderr)
	}
}

// TestB2BWriteBodies checks that create and update, whose body Emporix
// requires, fail without --file as a usage error (exit 2) that sends
// nothing, as do an ID that is no path segment and delete, which sends no
// body, given --file.
func TestB2BWriteBodies(t *testing.T) {
	file := writeBody(t, b2bLocationJSON)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"customer-locations", "create", "--dry-run"}, want: `required flag(s) "file" not set`},
		{args: []string{"vendor-locations", "update", "vl1", "--yes"}, want: `required flag(s) "file" not set`},
		{args: []string{"contact-assignments", "update", "..", "--file", file, "--yes"}, want: `invalid ID ".."`},
		{args: []string{"vendors", "delete", "v1", "--file", file, "--yes"}, want: "unknown flag: --file"},
		{args: []string{"legal-entities", "delete", "--yes"}, want: "accepts 1 arg(s), received 0"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

// TestB2BWritesHelp checks the help of these writes: create names the
// members of its body, how it names the items it refers to, that get prints
// one alike and that Emporix checks that they exist; update what it
// requires, that it creates a missing item and how metadata.version guards
// it, and for a legal entity which customer groups to keep; delete how
// Emporix answers an unknown item and what it deletes along with it; and
// update and delete where ID comes from. Their groups say what the writes
// send.
func TestB2BWritesHelp(t *testing.T) {
	for args, want := range map[string][]string{
		"customer-locations create": {"shaped as\n`epx customer-locations get ID -o json` prints one:\ntype (required",
			"leave out \"phones\": []"},
		"vendors create": {"locations (its vendor locations, each\nas {\"id\": ...} or in full, as get prints it)",
			"each as\n{\"email\": ...}", "parentId and locations must name\nitems that exist (400 Bad Request, exit 1"},
		"legal-entities create": {"the last three each as\n{\"id\": ...} or in full, as get prints it.",
			"must\nname items that exist", "may still make the IAM customer groups"},
		"contact-assignments create": {"A second assignment of a customer to the same legal entity is a conflict",
			"also assigns it to the parent company"},
		"contact-assignments update": {"or creates it with that id where none exists (an upsert); it requires\n" +
			"legalEntity and customer; a type left out becomes CONTACT",
			"With metadata.version as `epx contact-assignments get ID -o json`",
			"without metadata.version, it replaces it whatever its version.\n" +
				"ID is the ID that `epx contact-assignments list` shows.\n"},
		"legal-entities update": {"it requires\nname and type.", "list` shows.\nKeep every customer group",
			"(Admin, Buyer, Requester\nand Contact)", "without the Contact group, Emporix deletes that group"},
		"legal-entities delete": {"Deletes the legal entity ID; it sends no body.\nID is the ID that `epx " +
			"legal-entities list` shows; an unknown one " + string(answeredDeleted) + ".\nThe assignments are",
			"customer locations stay", "with its subsidiaries, assignments and IAM customer groups with one request"},
		"vendors delete": {"vendors list` shows; an unknown one " + string(answeredDeleted) +
			".\nIts subsidiaries stay"},
		"contact-assignments delete": {"list` shows; an unknown one is not found.\n"},
		"customer-locations delete":  {"is not found.\nA legal entity that names it"},
		"vendor-locations delete":    {"is not found.\nA vendor that names it in locations no longer does."},
		"vendor-locations":           {"each as {\"id\": ...}.\n" + versionedWritesNote + "\n"},
	} {
		_, help, _ := runCLI(t, nil, append(strings.Fields(args), "--help")...)
		for _, part := range want {
			if !strings.Contains(help, part) {
				t.Errorf("%s help %q; want %q", args, help, part)
			}
		}
	}
}

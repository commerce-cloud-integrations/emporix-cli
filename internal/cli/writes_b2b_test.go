package cli

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

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

func b2bCreatedJSON(id string) string { return "{\n  \"id\": \"" + id + "\"\n}\n" }

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

func TestB2BWritesHelp(t *testing.T) {
	for args, want := range map[string][]string{
		"customer-locations create": {"shaped as `epx customer-locations get ID -o json` prints one: type (required",
			"leave out \"phones\": []"},
		"vendors create": {"locations (its vendor locations, each as `{\"id\": ...}` or in full, as get prints it)",
			"each as `{\"email\": ...}`", "parentId and locations must name items that exist (400 Bad Request, exit 1"},
		"legal-entities create": {"the last three each as `{\"id\": ...}` or in full, as get prints it.",
			"must name items that exist", "may still make the IAM customer groups"},
		"contact-assignments create": {"A second assignment of a customer to the same legal entity is a conflict",
			"also assigns it to the parent company"},
		"contact-assignments update": {"or creates it with that id where none exists (an upsert); it requires " +
			"legalEntity and customer; a type left out becomes CONTACT",
			"With metadata.version as `epx contact-assignments get ID -o json`",
			"without metadata.version, it replaces it whatever its version. " +
				"ID is the ID that `epx contact-assignments list` shows.\n\n"},
		"legal-entities update": {"it requires name and type.", "list` shows. Keep every customer group",
			"(Admin, Buyer, Requester and Contact)", "without the Contact group, Emporix deletes that group"},
		"legal-entities delete": {"Deletes the legal entity ID; it sends no body. ID is the ID that `epx " +
			"legal-entities list` shows; an unknown one " + string(answeredDeleted) + ". The assignments are",
			"customer locations stay", "with its subsidiaries, assignments and IAM customer groups with one request"},
		"vendors delete": {"vendors list` shows; an unknown one " + string(answeredDeleted) +
			". Its subsidiaries stay"},
		"contact-assignments delete": {"list` shows; an unknown one is not found.\n\n"},
		"customer-locations delete":  {"is not found. A legal entity that names it"},
		"vendor-locations delete":    {"is not found. A vendor that names it in locations no longer does."},
		"vendor-locations":           {"each as `{\"id\": ...}`. " + versionedWritesNote + "\n"},
	} {
		_, help, _ := runCLI(t, nil, append(strings.Fields(args), "--help")...)
		for _, part := range want {
			if !containsFlat(help, part) {
				t.Errorf("%s help %q; want %q", args, help, part)
			}
		}
	}
}

package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	customerJSON = `{"contactEmail":"ada@example.com","firstName":"Ada","lastName":"Example"}`
	addressJSON  = `{"contactName":"Ada Example","street":"Main Street","city":"Amsterdam","country":"NL"}`
)

const (
	bulkCreated = `[{"index":0,"id":"c1","code":201,"status":"Created"},` +
		`{"index":1,"id":"c2","code":201,"status":"Created"}]`
	createdJSON = "[\n  {\n    \"code\": 201,\n    \"id\": \"c1\",\n    \"index\": 0,\n" +
		"    \"status\": \"Created\"\n  },\n  {\n    \"code\": 201,\n    \"id\": \"c2\",\n    \"index\": 1,\n" +
		"    \"status\": \"Created\"\n  }\n]\n"
)

func customerServiceWrites() []writeCase {
	manage := []string{customerManageScope}
	imports := []string{customerImportScope}
	const retention = `{"retentionEndDate":"2027-01-31","emailReminderDate":"2027-01-24","emailNotificationsEnabled":true}`
	return []writeCase{
		{
			args: []string{"customers", "create"}, file: customerJSON, method: http.MethodPost,
			path: "/customer/acme1/customers", query: "sendPasswordResetNotifications=false", scopes: manage,
			status: http.StatusCreated, response: `{"id":"0100"}`, want: "{\n  \"id\": \"0100\"\n}\n",
		},
		{
			args: []string{"customers", "create", "--password-reset-email"}, file: customerJSON, method: http.MethodPost,
			path: "/customer/acme1/customers", query: "sendPasswordResetNotifications=true", scopes: manage,
			status: http.StatusCreated, response: `{"id":"0101"}`, want: "{\n  \"id\": \"0101\"\n}\n",
		},
		{
			args: []string{"customers", "update", "0100"}, file: customerJSON, method: http.MethodPut,
			path: "/customer/acme1/customers/0100", query: "sendPasswordResetNotifications=false", scopes: manage,
			status: http.StatusNoContent,
		},
		{
			args: []string{"customers", "patch", "0100"}, file: `{"active":false}`, method: http.MethodPatch,
			path: "/customer/acme1/customers/0100", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"customers", "delete", "0100"}, method: http.MethodDelete, path: "/customer/acme1/customers/0100",
			scopes: manage, confirmBy: "0100", status: http.StatusAccepted,
		},
		{
			args: []string{"customers", "import"}, file: "[" + customerJSON + ",\n" + customerJSON + "]\n",
			method: http.MethodPost, path: "/customer/acme1/customers/import", scopes: imports,
			status: http.StatusMultiStatus, response: bulkCreated, want: createdJSON,
		},
		{
			args: []string{"customers", "addresses", "create", "0100"}, file: addressJSON, method: http.MethodPost,
			path: "/customer/acme1/customers/0100/addresses", scopes: manage, status: http.StatusCreated,
			response: `{"id":"a1"}`, want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			args: []string{"customers", "addresses", "update", "0100", "a 1"}, file: addressJSON, method: http.MethodPut,
			path: "/customer/acme1/customers/0100/addresses/a%201", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"customers", "addresses", "patch", "0100", "a1"}, file: `{"isDefault":true}`,
			method: http.MethodPatch, path: "/customer/acme1/customers/0100/addresses/a1", scopes: manage,
			status: http.StatusOK,
		},
		{
			args: []string{"customers", "addresses", "delete", "0100", "a1"}, method: http.MethodDelete,
			path: "/customer/acme1/customers/0100/addresses/a1", scopes: manage, confirmBy: "a1",
			status: http.StatusNoContent,
		},
		{
			args:   []string{"customers", "addresses", "add-tags", "0100", "a1", "--tags", "BILLING,SHIPPING"},
			method: http.MethodPost, path: "/customer/acme1/customers/0100/addresses/a1/tags",
			query: "tags=BILLING%2CSHIPPING", scopes: manage, status: http.StatusNoContent,
		},
		{
			args:   []string{"customers", "addresses", "remove-tags", "0100", "a1", "--tags", "BILLING"},
			method: http.MethodDelete, path: "/customer/acme1/customers/0100/addresses/a1/tags", query: "tags=BILLING",
			scopes: manage, confirmBy: "a1", status: http.StatusNoContent,
		},
		{
			args: []string{"customers", "password-migration", "update", "--ends", "2027-01-31", "--reminder", "2027-01-24",
				"--emails"},
			sent:   `{"emailNotificationsEnabled":true,"emailReminderDate":"2027-01-24","retentionEndDate":"2027-01-31"}`,
			method: http.MethodPost,
			path:   "/customer/acme1/config/password-migration-retention", scopes: imports, status: http.StatusOK,
			response: retention, want: "{\n  \"emailNotificationsEnabled\": true,\n  \"emailReminderDate\": \"2027-01-24\",\n" +
				"  \"retentionEndDate\": \"2027-01-31\"\n}\n",
		},
		{
			args: []string{"customers", "password-migration", "update", "--ends", "2027-01-31"},
			sent: `{"emailNotificationsEnabled":false,"retentionEndDate":"2027-01-31"}`, method: http.MethodPost,
			path: "/customer/acme1/config/password-migration-retention", scopes: imports, status: http.StatusOK,
		},
		{
			args: []string{"customers", "password-migration", "update"}, file: retention, method: http.MethodPost,
			path: "/customer/acme1/config/password-migration-retention", scopes: imports, status: http.StatusOK,
		},
		{
			args: []string{"customers", "password-migration", "delete"}, method: http.MethodDelete,
			path: "/customer/acme1/config/password-migration-retention", scopes: imports, confirmBy: "acme1",
			deletedMany: "the password migration retention of the tenant", status: http.StatusNoContent,
		},
	}
}

func TestCustomerWriteUsageErrorsSendNothing(t *testing.T) {
	const retention = `{"retentionEndDate":"2027-01-31"}`
	assertRefused(t, []refusedWrite{
		{args: []string{"customers", "addresses", "add-tags", "0100", "a1"}, want: `required flag(s) "tags" not set`},
		{args: []string{"customers", "addresses", "remove-tags", "0100", "a1", "--tags", ""}, want: `invalid argument ""`},
		{args: []string{"customers", "password-migration", "update"}, want: "at least one of the flags in the group " +
			"[file ends] is required"},
		{args: []string{"customers", "password-migration", "update", "--ends", "2027-01-31"}, file: retention,
			want: "if any flags in the group [file ends] are set none of the others can be"},
		{args: []string{"customers", "password-migration", "update", "--emails"}, file: retention,
			want: "if any flags in the group [file emails] are set none of the others can be"},
		{args: []string{"customers", "import"}, file: "[" + strings.Repeat(customerJSON+",", maxCustomerImport) +
			customerJSON + "]", want: "a bulk write sends at most 200 items in one request, not 201"},
		{args: []string{"customers", "delete", "0100"}, file: retention, want: "unknown flag: --file"},
	})
}

func TestCustomerGroupsStateTheirScopes(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"customers"}, want: "Its reads require different OAuth scopes; the help of each names them. " +
			"Its writes require different OAuth scopes; the help of each names them.\n\nUsage:"},
		{args: []string{"customers", "addresses"}, want: "Its reads require the customer.customer_read scope. " +
			"Its writes require the customer.customer_manage scope.\n\nUsage:"},
		{args: []string{"customers", "password-migration"}, want: "Its reads require the customer.import_read scope. " +
			"Its writes require the customer.import_manage scope.\n\nUsage:"},
		{args: []string{"customers", "create"}, want: "a link to set a password only with --password-reset-email"},
		{args: []string{"customers", "update"}, want: "keeping those it leaves out, or creates the " +
			"profile with that number where none exists (an upsert), as create does: with the email of " +
			"an existing account, that is a conflict (exit 5)."},
		{args: []string{"customers", "update"}, want: "Emporix answers 409 Conflict (exit 5) otherwise. " +
			"On an existing profile, a contactEmail changes its contact email only"},
		{args: []string{"customers", "patch"}, want: "never the email with which the customer signs in"},
		{args: []string{"customers", "delete"}, want: "Emporix answers 202 Accepted, and the profile is not found"},
		{args: []string{"customers", "addresses", "update"}, want: "keeping those it leaves out, such as its tags"},
		{args: []string{"customers", "password-migration", "update"}, want: "(--reminder; without it, 7 days before " +
			"--ends, or tomorrow if that has passed, also where the retention it replaces had another date)"},
		{args: []string{"customers", "password-migration", "update"}, want: "sends emailNotificationsEnabled=false; " +
			"a --file without it gets Emporix's default, true)."},
		{args: []string{"customers", "password-migration", "delete"}, want: "update sets the whole retention, so give " +
			"--reminder and --emails again"},
		{args: []string{"customers", "addresses", "get"}, want: "ID is the ID that `epx customers addresses list " +
			"CUSTOMER_NUMBER` shows"},
	} {
		code, help, stderr := runCLI(t, nil, append(tt.args, "--help")...)
		if code != ExitOK || !containsFlat(help, tt.want) {
			t.Errorf("%q: exit %d, stderr %q, help %q; want %q", tt.args, code, stderr, help, tt.want)
		}
	}
	for _, verb := range []string{"create", "update", "patch", "import"} {
		if _, help, _ := runCLI(t, nil, "customers", verb, "--help"); !containsFlat(help,
			"Emporix ignores a customerNumber that a profile gives (no error).") {
			t.Errorf("customers %s: help %q; want that Emporix ignores a customerNumber", verb, help)
		}
	}
}

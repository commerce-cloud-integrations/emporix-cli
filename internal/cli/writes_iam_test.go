package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func iamWrites() []writeCase {
	assignment := []string{iamAssignmentManage}
	return append([]writeCase{
		{
			args:   []string{"iam", "access-controls", "update", "ac1"},
			file:   `{"name":{"en":"Vehicle managers"},"scopes":["acme.vehicle_manage"],"metadata":{"version":2}}`,
			method: http.MethodPut, path: "/iam/acme1/access-controls/ac1", scopes: []string{iamAccessManage},
			lang: "*", status: http.StatusNoContent,
		},
		{
			args:   []string{"iam", "access-controls", "update", "ac2"},
			file:   `{"name":{"en":"Vehicle readers"},"scopes":["acme.vehicle_read"],"restrictedTo":"EMPLOYEE"}`,
			method: http.MethodPut, path: "/iam/acme1/access-controls/ac2", scopes: []string{iamAccessManage},
			lang: "*", status: http.StatusCreated, response: `{"id":"ac2"}`, want: "{\n  \"id\": \"ac2\"\n}\n",
		},
		{
			args: []string{"iam", "access-controls", "delete", "ac1"}, method: http.MethodDelete,
			path: "/iam/acme1/access-controls/ac1", scopes: []string{iamAccessManage}, status: http.StatusNoContent,
			confirmBy: "ac1",
		},
		{
			args: []string{"iam", "scopes", "update", "acme.vehicle_manage"},
			file: `{"domain":"vehicle","description":{"en":"Manage vehicles"}}`, method: http.MethodPut,
			path: "/iam/acme1/scopes/acme.vehicle_manage", scopes: []string{iamScopeManage},
			status: http.StatusCreated, response: `{"id":"acme.vehicle_manage"}`,
			want: "{\n  \"id\": \"acme.vehicle_manage\"\n}\n",
		},
		{
			args: []string{"iam", "scopes", "delete", "acme.vehicle_manage"}, method: http.MethodDelete,
			path: "/iam/acme1/scopes/acme.vehicle_manage", scopes: []string{iamScopeManage},
			status: http.StatusNoContent, confirmBy: "acme.vehicle_manage",
		},
		{
			args: []string{"iam", "groups", "create"},
			file: `{"id":"vehicle-managers","name":{"en":"Vehicle managers"},"accessControls":["ac1"],` +
				`"userType":"EMPLOYEE"}`,
			method: http.MethodPost, path: "/iam/acme1/groups", scopes: []string{iamGroupManage}, lang: "*",
			status: http.StatusCreated, response: `{"id":"vehicle-managers"}`, want: "{\n  \"id\": \"vehicle-managers\"\n}\n",
		},
		{
			args:   []string{"iam", "groups", "update", "g1"},
			file:   `{"name":{"de":"Fahrzeugverwalter"},"accessControls":["ac1"],"metadata":{"version":3}}`,
			method: http.MethodPut, path: "/iam/acme1/groups/g1", scopes: []string{iamGroupManage}, lang: "*",
			status: http.StatusNoContent,
		},
		{
			args: []string{"iam", "groups", "delete", "g1"}, method: http.MethodDelete, path: "/iam/acme1/groups/g1",
			scopes: []string{iamGroupManage}, status: http.StatusNoContent, confirmBy: "g1",
		},
		{
			args: []string{"iam", "groups", "delete", "g1", "--force-delete"}, method: http.MethodDelete,
			path: "/iam/acme1/groups/g1", query: "forceDelete=true",
			scopes: []string{iamGroupManage, iamAssignmentManage}, status: http.StatusNoContent,
			confirmBy: "g1", deletedMany: "the group and the membership of every user of it",
		},
	}, slices.Concat(iamMembershipWrites(assignment), iamUserWrites())...)
}

func iamMembershipWrites(assignment []string) []writeCase {
	return []writeCase{
		{
			args: []string{"iam", "groups", "users", "create", "g1", "--user", "u1", "--user-type", "EMPLOYEE"},
			sent: `{"userId":"u1","userType":"EMPLOYEE"}`, method: http.MethodPost, path: "/iam/acme1/groups/g1/users",
			scopes: assignment, status: http.StatusCreated, response: `{"id":"m1"}`, want: "{\n  \"id\": \"m1\"\n}\n",
		},
		{
			args: []string{"iam", "groups", "users", "create", "g1", "--user", "u2"}, sent: `{"userId":"u2"}`,
			method: http.MethodPost, path: "/iam/acme1/groups/g1/users", scopes: assignment,
			status: http.StatusCreated, response: `{"id":"m3"}`, want: "{\n  \"id\": \"m3\"\n}\n",
		},
		{
			args: []string{"iam", "groups", "users", "create", "g1"}, file: `{"userId":"c1","userType":"CUSTOMER"}`,
			method: http.MethodPost, path: "/iam/acme1/groups/g1/users", scopes: assignment,
			status: http.StatusCreated, response: `{"id":"m2"}`, want: "{\n  \"id\": \"m2\"\n}\n",
		},
		{
			args: []string{"iam", "groups", "users", "update", "g1", "EMPLOYEE", "u 1"}, method: http.MethodPut,
			path: "/iam/acme1/groups/g1/users/EMPLOYEE/u%201", scopes: assignment, status: http.StatusCreated,
			response: `{"id":"m1"}`, want: "{\n  \"id\": \"m1\"\n}\n",
		},
		{
			args: []string{"iam", "groups", "users", "delete", "g1", "u1"}, method: http.MethodDelete,
			path: "/iam/acme1/groups/g1/users/u1", scopes: assignment, status: http.StatusNoContent,
			confirmBy: "u1",
		},
		{
			args: []string{"iam", "groups", "users", "delete-all", "g1"}, method: http.MethodDelete,
			path: "/iam/acme1/groups/g1/users", scopes: assignment, status: http.StatusNoContent,
			confirmBy: "g1", deletedMany: "the membership of every user of the group",
		},
		{
			args: []string{"iam", "users", "groups", "delete-all", "u1"}, method: http.MethodDelete,
			path: "/iam/acme1/users/u1/groups", scopes: assignment, status: http.StatusNoContent,
			confirmBy: "u1", deletedMany: "the user's membership of every group",
		},
	}
}

func iamUserWrites() []writeCase {
	const user = `{"firstName":"Jane","lastName":"Doe","preferredSite":"AcmeStore_NL","preferredCurrency":"EUR",` +
		`"preferredLanguage":"en","department":"Sales","contactEmail":"jane@example.com","groupIds":["g1"]}`
	return []writeCase{
		{
			args: []string{"iam", "users", "create"}, file: user, method: http.MethodPost, path: "/iam/acme1/users",
			scopes: []string{"iam.user_create"}, status: http.StatusCreated, response: `{"id":"u1"}`,
			want: "{\n  \"id\": \"u1\"\n}\n",
		},
		{
			args:   []string{"iam", "users", "update", "u1"},
			file:   strings.Replace(user, `"contactEmail":"jane@example.com"`, `"backofficeUserNumber":"u1"`, 1),
			method: http.MethodPut, path: "/iam/acme1/users/u1", scopes: []string{"iam.user_update"},
			status: http.StatusNoContent,
		},
		{
			args: []string{"iam", "users", "delete", "u1"}, method: http.MethodDelete, path: "/iam/acme1/users/u1",
			scopes: []string{"iam.user_delete"}, status: http.StatusNoContent, confirmBy: "u1",
		},
	}
}

func TestIAMGroupDeleteForcedFailsWithoutTheAssignmentScope(t *testing.T) {
	fake := newFakeEmporix(t, iamAssignmentManage)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "iam", "groups", "delete", "g1", "--force-delete",
		"--yes")
	if code != ExitAuth || stdout != "" || !strings.Contains(stderr, iamAssignmentManage) {
		t.Fatalf("exit %d, stdout %q, stderr %q; want exit 3 naming %s", code, stdout, stderr, iamAssignmentManage)
	}
	fake.assertRequests(t)
}

func TestIAMWritesRefuseInvalidInput(t *testing.T) {
	file := writeBody(t, `{"userId":"u1"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"groups", "users", "create", "g1"}, want: "at least one of the flags in the group [file user]"},
		{args: []string{"groups", "users", "create", "g1", "--file", file, "--user", "u1"},
			want: "if any flags in the group [file user] are set none of the others can be"},
		{args: []string{"groups", "users", "create", "--user", "u1"}, want: "accepts 1 arg(s), received 0"},
		{args: []string{"groups", "users", "update", "g1", "u1"}, want: "accepts 3 arg(s), received 2"},
		{args: []string{"groups", "users", "update", "g1", "..", "u1"}, want: `invalid USER_TYPE ".."`},
		{args: []string{"groups", "users", "delete", "g1", "."}, want: `invalid USER_ID "."`},
		{args: []string{"groups", "users", "delete-all"}, want: "accepts 1 arg(s), received 0"},
		{args: []string{"groups", "delete", "g1", "--file", file}, want: "unknown flag: --file"},
		{args: []string{"scopes", "update", "acme.vehicle_manage"}, want: `required flag(s) "file" not set`},
		{args: []string{"scopes", "update", "acme.x", "--file", file, "--content-language", "en"},
			want: "unknown flag: --content-language"},
		{args: []string{"access-controls", "update", "ac1", "--file", file, "--content-language", "en"},
			want: "unknown flag: --content-language"},
		{args: []string{"groups", "create", "--file", file, "--content-language", "en"},
			want: "unknown flag: --content-language"},
		{args: []string{"groups", "update", "g1", "--file", file, "--content-language", "en"},
			want: "unknown flag: --content-language"},
		{args: []string{"users", "delete", "u1", "--force-delete"}, want: "unknown flag: --force-delete"},
		{args: []string{"groups", "users", "create", "g1", "--user", "u1", "--user-type", "ALIEN"},
			want: `invalid argument "ALIEN" for "--user-type" flag: use CUSTOMER or EMPLOYEE`},
		{args: []string{"groups", "users", "create", "g1", "--user", "u1", "--user-type", "employee"},
			want: `invalid argument "employee" for "--user-type" flag: use CUSTOMER or EMPLOYEE`},
		{args: []string{"groups", "users", "create", "g1", "--user", "u1", "--user-type="},
			want: `invalid argument "" for "--user-type" flag: empty`},
		{args: []string{"groups", "users", "create", "g1", "--file", file, "--user-type", "EMPLOYEE"},
			want: "if any flags in the group [file user-type] are set none of the others can be"},
		{args: []string{"groups", "users", "update", "g1", "ALIEN", "u1"},
			want: `invalid USER_TYPE "ALIEN": use CUSTOMER or EMPLOYEE`},
		{args: []string{"groups", "users", "update", "g1", "employee", "u1"},
			want: `invalid USER_TYPE "employee": use CUSTOMER or EMPLOYEE`},
		{args: []string{"groups", "users", "update", "g1", "", "u1"}, want: `invalid USER_TYPE ""`},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(append([]string{"iam"}, tt.args...),
				flag)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, flag, code, stdout,
					stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

func TestIAMWritesSayTheyChangeAccess(t *testing.T) {
	var writes int
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		if _, isWrite := cmd.Annotations[writeScopesAnnotation]; isWrite &&
			strings.HasPrefix(cmd.CommandPath(), "epx iam ") {
			writes++
			if !strings.Contains(cmd.Long, accessChange+"\n\n") {
				t.Errorf("%s: help %q; want it to say %q", cmd.CommandPath(), cmd.Long, accessChange)
			}
		}
	})
	if writes == 0 {
		t.Error("no IAM write found")
	}
}

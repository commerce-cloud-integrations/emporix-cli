package cli

import (
	"net/http"
	"strings"
	"testing"
)

// configurationWrites are the writes of tenant configurations and of those of
// clients.
func configurationWrites() []writeCase {
	manage := []string{configurationManage}
	const client = "/configuration/acme1/clients/acme.indexing-client/configurations"
	return []writeCase{
		{
			args: []string{"configurations", "create"},
			file: `[{"key":"storefront.host","value":"shop.acme.example"}]`, method: http.MethodPost,
			path: "/configuration/acme1/configurations", scopes: manage, status: http.StatusCreated,
			body: `[{"key":"storefront.host","secured":false,"value":"shop.acme.example","version":1}]`,
			want: "[\n  {\n    \"key\": \"storefront.host\",\n    \"secured\": false,\n    \"value\": \"shop.acme.example\",\n" +
				"    \"version\": 1\n  }\n]\n",
		},
		{
			args: []string{"configurations", "update", "project_country"},
			file: "{\"key\": \"project_country\", \"value\": \"AT\", \"version\": 1}\n", method: http.MethodPut,
			path: "/configuration/acme1/configurations/project_country", scopes: manage, status: http.StatusOK,
			body: `{"key":"project_country","value":"AT","version":2}`,
			want: "{\n  \"key\": \"project_country\",\n  \"value\": \"AT\",\n  \"version\": 2\n}\n",
		},
		{
			args: []string{"configurations", "client-create", "acme.indexing-client"},
			file: `[{"key":"batchSize","value":100}]`, method: http.MethodPost, path: client, scopes: manage,
			status: http.StatusCreated, body: `[{"client":"acme.indexing-client","key":"batchSize","value":100}]`,
			want: "[\n  {\n    \"client\": \"acme.indexing-client\",\n    \"key\": \"batchSize\",\n    \"value\": 100\n  }\n]\n",
		},
		{
			args: []string{"configurations", "client-update", "acme.indexing-client", "batchSize"},
			file: `{"key":"batchSize","value":250}`, method: http.MethodPut, path: client + "/batchSize", scopes: manage,
			status: http.StatusOK, body: `{"client":"acme.indexing-client","key":"batchSize","value":250}`,
			want: "{\n  \"client\": \"acme.indexing-client\",\n  \"key\": \"batchSize\",\n  \"value\": 250\n}\n",
		},
	}
}

// TestConfigurationWriteUsageErrorsSendNothing checks that a write without
// its body or arguments, or with a key that would address another path,
// fails as a usage error before anything is sent: create and update need
// --file.
func TestConfigurationWriteUsageErrorsSendNothing(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"configurations", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"configurations", "update", "project_country"}, want: `required flag(s) "file" not set`},
		{args: []string{"configurations", "client-update", "acme.indexing-client", "batchSize"},
			want: `required flag(s) "file" not set`},
		{args: []string{"configurations", "create", "--json"}, want: "unknown flag: --json"},
		{args: []string{"configurations", "update", "."}, want: `invalid KEY "."`},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, "--yes")...)
		if code != ExitUsage || stdout != "" || !strings.HasPrefix(stderr, "Error: "+tt.want+"\n") {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

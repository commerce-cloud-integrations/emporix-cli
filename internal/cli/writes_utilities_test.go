package cli

import (
	"net/http"
	"strings"
	"testing"
)

func webhookWrites() []writeCase {
	manage := []string{"webhook.subscription_manage"}
	const config = `{"code":"http","provider":"HTTP","active":true,` +
		`"configuration":{"destinationUrl":"https://hooks.example.com/emporix"}}`
	return []writeCase{
		{
			args: []string{"webhooks", "configurations", "create"}, file: config, method: http.MethodPost,
			path: "/webhook/acme1/config", scopes: manage, status: http.StatusCreated, response: `{"code":"http"}`,
			want: "{\n  \"code\": \"http\"\n}\n",
		},
		{
			args: []string{"webhooks", "configurations", "update", "http"}, file: config, method: http.MethodPut,
			path: "/webhook/acme1/config/http", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"webhooks", "configurations", "patch", "http"},
			file: `[{"op":"UPSERT","path":"/active","value":false}]`, method: http.MethodPatch,
			path: "/webhook/acme1/config/http", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"webhooks", "configurations", "delete", "http", "--force"}, method: http.MethodDelete,
			path: "/webhook/acme1/config/http", query: "force=true", scopes: manage, confirmBy: "http",
			status: http.StatusNoContent,
		},
		{
			args: []string{"webhooks", "configurations", "delete", "svix"}, method: http.MethodDelete,
			path: "/webhook/acme1/config/svix", query: "force=false", scopes: manage, confirmBy: "svix",
			status: http.StatusNoContent,
		},
		{
			args: []string{"webhooks", "subscriptions", "bulk-patch"},
			file: `[{"eventType":"order.created","action":"SUBSCRIBE","metadata":{"version":1}},` +
				`{"eventType":"cart.closed","action":"UNSUBSCRIBE","fieldsToUnsubscribe":["customer"]}]`,
			method: http.MethodPatch, path: "/webhook/acme1/event-subscriptions", scopes: manage,
			status:   http.StatusMultiStatus,
			response: `[{"eventType":"order.created","code":200,"status":"OK"},{"eventType":"cart.closed","code":200}]`,
			want: "[\n  {\n    \"code\": 200,\n    \"eventType\": \"order.created\",\n    \"status\": \"OK\"\n  },\n" +
				"  {\n    \"code\": 200,\n    \"eventType\": \"cart.closed\"\n  }\n]\n",
		},
	}
}

func searchWrites() []writeCase {
	manage := []string{"search.search_manage"}
	const index = `{"name":{"en":"Vehicles"},"fields":[{"path":"make","text":true,"exact":true}]}`
	const pkg = "W3siaWQiOiJ2ZWhpY2xlcyIsInR5cGUiOiJ2ZWhpY2xlIiwiZmllbGRzIjpbeyJwYXRoIjoibWFrZSIsInRleHQiOnRydWUs" +
		"ImV4YWN0Ijp0cnVlfV19XQ=="
	return []writeCase{
		{
			args: []string{"search", "indexes", "update", "vehicle", "vehicles"}, file: index, method: http.MethodPut,
			path: "/search/acme1/search/vehicle/indexes/vehicles", scopes: manage, lang: "*",
			status: http.StatusAccepted, response: `{"id":"job1"}`, want: "{\n  \"id\": \"job1\"\n}\n",
		},
		{
			args:   []string{"search", "indexes", "update", "vehicle", "vehicles"},
			file:   `{"name":{"en":"Vehicles"},"fields":[{"path":"make","text":true,"exact":true}],"metadata":{"version":2}}`,
			method: http.MethodPut, path: "/search/acme1/search/vehicle/indexes/vehicles", scopes: manage, lang: "*",
			status: http.StatusNoContent,
		},
		{
			args: []string{"search", "indexes", "delete", "vehicle", "vehicles"}, method: http.MethodDelete,
			path: "/search/acme1/search/vehicle/indexes/vehicles", scopes: manage, confirmBy: "vehicles",
			status:   http.StatusAccepted,
			response: `{"id":"job2"}`, want: "{\n  \"id\": \"job2\"\n}\n",
		},
		{
			args: []string{"search", "indexes", "import"}, file: `{"exportedAt":"2026-10-01T08:00:00Z","data":"` + pkg + `"}`,
			method: http.MethodPost, path: "/search/acme1/search/indexes/import", scopes: manage, lang: "*",
			status: http.StatusOK, response: `[{"id":"vehicles","type":"vehicle","jobId":"job3","jobType":"create_index"}]`,
			want: "[\n  {\n    \"id\": \"vehicles\",\n    \"jobId\": \"job3\",\n    \"jobType\": \"create_index\",\n" +
				"    \"type\": \"vehicle\"\n  }\n]\n",
		},
		{
			args:   []string{"search", "queries", "update", "vehicle", "by-make"},
			file:   `{"index":"vehicles","queries":[{"type":"TEXT","query":"toyota","field":"make"}]}`,
			method: http.MethodPut, path: "/search/acme1/search/vehicle/queries/by-make", scopes: manage, lang: "*",
			status: http.StatusCreated, response: `{"id":"by-make"}`, want: "{\n  \"id\": \"by-make\"\n}\n",
		},
		{
			args: []string{"search", "queries", "delete", "vehicle", "by-make"}, method: http.MethodDelete,
			path: "/search/acme1/search/vehicle/queries/by-make", scopes: manage, confirmBy: "by-make",
			status: http.StatusNoContent,
		},
		{
			args: []string{"search", "indexes", "update", "legal-entity", "companies"}, method: http.MethodPut,
			file: `{"name":{"en":"Companies"},"fields":[{"path":"name","text":true}]}`,
			path: "/search/acme1/search/legal-entity/indexes/companies", scopes: manage, lang: "*",
			status: http.StatusAccepted, response: `{"id":"job4"}`, want: "{\n  \"id\": \"job4\"\n}\n",
		},
		{
			args: []string{"search", "queries", "delete", "customer", "by-city"}, method: http.MethodDelete,
			path: "/search/acme1/search/customer/queries/by-city", scopes: manage, confirmBy: "by-city",
			status: http.StatusNoContent,
		},
	}
}

func sequentialIDWrites() []writeCase {
	manage, view := []string{"sequentialid.schema_manage"}, []string{"sequentialid.schema_view"}
	const id = "6915e2eb3b4c2e6c7c80711b"
	return []writeCase{
		{
			args: []string{"sequential-ids", "create"},
			file: `{"name":"Orders","schemaType":"orderNoSequence","preText":"ORD-","startValue":1,` +
				`"maxValue":999999,"numberOfDigits":6}`,
			method: http.MethodPost, path: "/sequential-id/acme1/schemas", scopes: manage, status: http.StatusCreated,
			response: `{"id":"` + id + `"}`, want: "{\n  \"id\": \"" + id + "\"\n}\n",
		},
		{
			args:   []string{"sequential-ids", "create"},
			file:   `{"name":"Orders","schemaType":"orderNoSequence","startValue":1,"maxValue":999999,"numberOfDigits":6}`,
			method: http.MethodPost, path: "/sequential-id/acme1/schemas", scopes: manage, status: http.StatusCreated,
			response: id, want: "\"" + id + "\"\n",
		},
		{
			args: []string{"sequential-ids", "activate", id}, method: http.MethodPost,
			path: "/sequential-id/acme1/schemas/" + id + "/setActive", scopes: manage, status: http.StatusOK,
		},
		{
			args: []string{"sequential-ids", "delete", id}, method: http.MethodDelete,
			path: "/sequential-id/acme1/schemas/" + id, scopes: manage, confirmBy: id, status: http.StatusNoContent,
		},
		{
			args:   []string{"sequential-ids", "next-id", "orderNoSequence", "--sequence-key", "2026-10"},
			sent:   `{"sequenceKey":"2026-10"}`,
			method: http.MethodPost, path: "/sequential-id/acme1/schemas/types/orderNoSequence/nextId", scopes: view,
			status: http.StatusCreated, response: `{"id":"ORD-000045"}`, want: "{\n  \"id\": \"ORD-000045\"\n}\n",
		},
		{
			args: []string{"sequential-ids", "next-id", "orderNoSequence", "--site", "AcmeStore_NL"}, sent: `{}`,
			method: http.MethodPost, path: "/sequential-id/acme1/schemas/types/orderNoSequence/nextId",
			query: "siteCode=AcmeStore_NL", scopes: view, status: http.StatusCreated, response: `{"id":"ORD-000046"}`,
			want: "{\n  \"id\": \"ORD-000046\"\n}\n",
		},
		{
			args: []string{"sequential-ids", "next-id", "invoiceNoSequence"},
			file: `{"placeholders":{"__year__":"2026"}}`, method: http.MethodPost,
			path: "/sequential-id/acme1/schemas/types/invoiceNoSequence/nextId", scopes: view,
			status: http.StatusCreated, response: `{"id":"INV-2026-7"}`, want: "{\n  \"id\": \"INV-2026-7\"\n}\n",
		},
		{
			args: []string{"sequential-ids", "next-ids", "--site", "AcmeStore_NL"},
			file: `{"orderNoSequence":{"numberOfIds":2}}`, method: http.MethodPost,
			path: "/sequential-id/sequenceSchemaBatch/nextIds", query: "siteCode=AcmeStore_NL", scopes: view,
			status: http.StatusCreated, response: `{"orderNoSequence":{"ids":["ORD-000047","ORD-000048"]}}`,
			want: "{\n  \"orderNoSequence\": {\n    \"ids\": [\n      \"ORD-000047\",\n      \"ORD-000048\"\n" +
				"    ]\n  }\n}\n",
		},
	}
}

func TestSubscriptionChangesListTheEventsThatFailed(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/webhook/acme1/event-subscriptions", respond(http.StatusMultiStatus,
		`[{"eventType":"cart.closed","code":409,"status":"CONFLICT","message":"Fetch the latest version."},`+
			`{"eventType":"order.created","code":200,"status":"OK"}]`))
	file := writeBody(t, `[{"eventType":"order.created","action":"SUBSCRIBE"},`+
		`{"eventType":"cart.closed","action":"UNSUBSCRIBE","metadata":{"version":1}}]`)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "webhooks", "subscriptions", "bulk-patch", "--file",
		file, "--yes")
	if code != ExitConflict || !strings.Contains(stdout, `"eventType": "cart.closed"`) ||
		!strings.HasPrefix(stderr, "item 1 (cart.closed): 409 Conflict: Fetch the latest version.\n") {
		t.Fatalf("exit %d, stdout %q, stderr %q; want exit 5 naming item 1, cart.closed", code, stdout, stderr)
	}
	changes := func(n int) string {
		return "[" + strings.Repeat(`{"eventType":"order.created","action":"SUBSCRIBE"},`, n-1) +
			`{"eventType":"cart.closed","action":"SUBSCRIBE"}]`
	}
	code, stdout, stderr = runIn(t, writeEnv(fake, false, ""), "webhooks", "subscriptions", "bulk-patch", "--file",
		writeBody(t, changes(200)), "--dry-run")
	if code != ExitOK || !strings.HasPrefix(stdout, "PATCH ") || stderr != "" {
		t.Fatalf("200 changes: exit %d, stdout %.40q, stderr %q; want a dry run", code, stdout, stderr)
	}
	refused := newFakeEmporix(t)
	code, stdout, stderr = runIn(t, writeEnv(refused, false, ""), "webhooks", "subscriptions", "bulk-patch", "--file",
		writeBody(t, changes(201)), "--yes")
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "at most 200 items in one request, not 201") {
		t.Fatalf("201 changes: exit %d, stdout %q, stderr %q; want exit 2", code, stdout, stderr)
	}
	refused.assertNothingSent(t)
}

func TestIDsAreIssuedAsWrites(t *testing.T) {
	for _, use := range []string{"next-id", "next-ids"} {
		_, help, _ := runCLI(t, nil, "sequential-ids", use, "--help")
		if !containsFlat(help, "Each ID issued advances the schema's counter. "+
			"It changes data although Emporix needs no write-capable scope for it.\n\n") {
			t.Errorf("%s: help %q; want it to say that it changes data", use, help)
		}
	}
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"next-id", "orderNoSequence"}, want: "give --yes"},
		{args: []string{"next-id", "orderNoSequence", "--sequence-key", "k", "--file", writeBody(t, `{}`), "--yes"},
			want: "none of the others can be"},
		{args: []string{"next-ids", "--file", writeBody(t, `{}`), "--site", "..", "--yes"}, want: `invalid site ".."`},
	} {
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), append([]string{"sequential-ids"}, tt.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stderr %q; want exit 2 and %q", tt.args, code, stderr, tt.want)
		}
	}
	fake.assertNothingSent(t)
	defer func() {
		if recover() == nil {
			t.Error("a write declared readScoped with a write-capable scope was built")
		}
	}()
	(&write{use: "next-id TYPE", method: http.MethodPost, req: request{path: "/x/{id}"}, body: noBody{},
		scopes: []string{"sequentialid.schema_manage"}, readScoped: true}).command(&app{
		getenv: func(string) string { return "" },
	})
}

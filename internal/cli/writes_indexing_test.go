package cli

import (
	"net/http"
	"strings"
	"testing"
)

const reindexJobBody = `{"id":"j1","status":"IN_PROGRESS","entityType":"PRODUCT","expectedCount":5000,` +
	`"processedCount":0,"metadata":{"createdAt":"2026-06-16T12:32:14.132Z"}}`

const reindexJobJSON = "{\n  \"entityType\": \"PRODUCT\",\n  \"expectedCount\": 5000,\n  \"id\": \"j1\",\n" +
	"  \"metadata\": {\n    \"createdAt\": \"2026-06-16T12:32:14.132Z\"\n  },\n  \"processedCount\": 0,\n" +
	"  \"status\": \"IN_PROGRESS\"\n}\n"

func indexingWrites() []writeCase {
	const provider = `{"provider":"ALGOLIA","active":true,"applicationId":"APP1","indexName":"acme1",` +
		`"searchKey":"s","writeKey":"w"}`
	manage := []string{"indexing.search_manage"}
	return []writeCase{
		{
			args: []string{"indexing", "providers", "create"}, file: provider, method: http.MethodPost,
			path: "/indexing/acme1/configurations", scopes: manage, status: http.StatusCreated,
			response: `{"id":"ALGOLIA"}`, want: "{\n  \"id\": \"ALGOLIA\"\n}\n",
		},
		{
			args: []string{"indexing", "providers", "update", "ALGOLIA"}, file: provider, method: http.MethodPut,
			path: "/indexing/acme1/configurations/ALGOLIA", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"indexing", "reindex"}, sent: `{"mode":"FULL"}`, method: http.MethodPost,
			path: "/indexing/acme1/reindex", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"indexing", "reindex"}, file: `{"mode":"FULL"}`, method: http.MethodPost,
			path: "/indexing/acme1/reindex", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"indexing", "reindex-jobs", "create", "--entity-type", "PRODUCT", "--rag"},
			sent: `{"entityType":"PRODUCT","rag":true}`, method: http.MethodPost, path: "/indexing/acme1/reindex-jobs",
			scopes: manage, status: http.StatusCreated, response: reindexJobBody, want: reindexJobJSON,
		},
		{
			args:   []string{"indexing", "reindex-jobs", "create", "--entity-type", "LEGAL-ENTITY"},
			sent:   `{"entityType":"LEGAL-ENTITY"}`,
			method: http.MethodPost, path: "/indexing/acme1/reindex-jobs", scopes: manage, status: http.StatusCreated,
			response: reindexJobBody, want: reindexJobJSON,
		},
		{
			args: []string{"indexing", "reindex-jobs", "create"}, file: `{"entityType":"vehicle"}`,
			method: http.MethodPost, path: "/indexing/acme1/reindex-jobs", scopes: manage, status: http.StatusOK,
			response: reindexJobBody, want: reindexJobJSON,
		},
		{
			args: []string{"ai", "rag", "reindex", "PRODUCT"}, method: http.MethodPost,
			path: "/ai-rag-indexer/acme1/PRODUCT/reindex", scopes: []string{"ai.agent_manage"},
			status: http.StatusNoContent,
		},
		{
			args: []string{"ai", "rag", "reindex", "customer"}, method: http.MethodPost,
			path: "/ai-rag-indexer/acme1/customer/reindex", scopes: []string{"ai.agent_manage"},
			status: http.StatusNoContent,
		},
	}
}

func TestReindexJobCreateSendsTheEntityTypeAsGiven(t *testing.T) {
	sent := []string{"vehicle"}
	for _, name := range entityBuiltInTypeNames() {
		sent = append(sent, name, strings.ToLower(name))
	}
	for _, entityType := range sent {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""),
			"indexing", "reindex-jobs", "create", "--entity-type", entityType, "--dry-run")
		if want := "\n\n{\"entityType\":\"" + entityType + "\"}\n"; code != ExitOK || !strings.HasSuffix(stdout, want) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want the body %q", entityType, code, stdout, stderr, want)
		}
		fake.assertNothingSent(t)
	}
	help := helpText(t, "indexing", "reindex-jobs", "create")
	for _, want := range []string{
		entityTypeHelp,
		"to reindex: PRODUCT, ORDER, CUSTOMER or LEGAL-ENTITY in any case, or a custom entity type " +
			"(required without --file)",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help lacks %q:\n%s", want, help)
		}
	}
}

func TestIndexingWriteBodies(t *testing.T) {
	file := writeBody(t, `{"provider":"ALGOLIA"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"indexing", "providers", "create", "--dry-run"}, want: `required flag(s) "file" not set`},
		{args: []string{"indexing", "providers", "update", "ALGOLIA", "--yes"}, want: `required flag(s) "file" not set`},
		{args: []string{"ai", "rag", "reindex", "PRODUCT", "--file", file, "--yes"}, want: "unknown flag: --file"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
	_, help, _ := runCLI(t, nil, "ai", "rag", "reindex", "--help")
	if strings.Contains(help, "--file") || !strings.Contains(help, "send without asking (required without a terminal)\n") {
		t.Errorf("help %q; want no --file", help)
	}
}

func TestIndexProviderHelpNamesBatteryIncludedSettings(t *testing.T) {
	settings := []string{"apiUrl", "enrichments", "siteAwareMixinsStrategy", "FEE calls the Fee Service but adds nothing",
		"for ALGOLIA", "an apiUrl of another host"}
	assertHelpHolds(t, map[string][]string{
		"indexing providers create": settings,
		"indexing providers update": append([]string{
			"replaced whole", "goes back to its default", "`epx indexing providers get PROVIDER -o json`",
			"`epx indexing reindex-jobs create --entity-type PRODUCT`",
		}, settings...),
		"indexing providers list": {"apiUrl, enrichments and siteAwareMixinsStrategy, where they are set"},
	})
}

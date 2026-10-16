package cli

import (
	"net/http"
	"strings"
	"testing"
)

// reindexJobBody is a reindex job as Emporix answers its creation.
const reindexJobBody = `{"id":"j1","status":"IN_PROGRESS","entityType":"PRODUCT","expectedCount":5000,` +
	`"processedCount":0,"metadata":{"createdAt":"2026-06-16T12:32:14.132Z"}}`

// reindexJobJSON is reindexJobBody as epx prints it.
const reindexJobJSON = "{\n  \"entityType\": \"PRODUCT\",\n  \"expectedCount\": 5000,\n  \"id\": \"j1\",\n" +
	"  \"metadata\": {\n    \"createdAt\": \"2026-06-16T12:32:14.132Z\"\n  },\n  \"processedCount\": 0,\n" +
	"  \"status\": \"IN_PROGRESS\"\n}\n"

// indexingWrites are the writes of the Indexing Service and of the AI RAG
// Indexer: index provider configurations and reindexing.
func indexingWrites() []writeCase {
	const provider = `{"provider":"ALGOLIA","active":true,"applicationId":"APP1","indexName":"acme1",` +
		`"searchKey":"s","writeKey":"w"}`
	manage := []string{"indexing.search_manage"}
	return []writeCase{
		{
			args: []string{"indexing", "providers", "create"}, file: provider, method: http.MethodPost,
			path: "/indexing/acme1/configurations", scopes: manage, status: http.StatusCreated,
			body: `{"id":"ALGOLIA"}`, want: "{\n  \"id\": \"ALGOLIA\"\n}\n",
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
			scopes: manage, status: http.StatusCreated, body: reindexJobBody, want: reindexJobJSON,
		},
		{
			args: []string{"indexing", "reindex-jobs", "create"}, file: `{"entityType":"vehicle"}`,
			method: http.MethodPost, path: "/indexing/acme1/reindex-jobs", scopes: manage, status: http.StatusOK,
			body: reindexJobBody, want: reindexJobJSON,
		},
		{
			args: []string{"ai", "rag", "reindex", "PRODUCT"}, method: http.MethodPost,
			path: "/ai-rag-indexer/acme1/PRODUCT/reindex", scopes: []string{"ai.agent_manage"},
			status: http.StatusNoContent,
		},
	}
}

// TestIndexingWriteBodies checks that providers create and update, whose
// configuration Emporix requires, fail without --file as a usage error (exit
// 2) that sends nothing, and that ai rag reindex, which Emporix takes no body
// for, has no --file.
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

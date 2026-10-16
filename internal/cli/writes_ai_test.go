package cli

import (
	"net/http"
	"os"
	"slices"
)

// orderCSV returns the content of testdata/order.csv, which a case uploads
// byte for byte; when it cannot be read, every test of the cases fails.
func orderCSV() string {
	content, err := os.ReadFile("testdata/order.csv")
	if err != nil {
		panic(err)
	}
	return string(content)
}

// orderText returns the content of testdata/order.text, as orderCSV does.
func orderText() string {
	content, err := os.ReadFile("testdata/order.text")
	if err != nil {
		panic(err)
	}
	return string(content)
}

// The paths of the AI Service in the cases.
const (
	aiServicePath = "/ai-service/acme1/"
	aiAgenticPath = aiServicePath + "agentic/"
)

// aiWrites are the writes of the AI Service: the agents and what they are
// built from, jobs, and what consumes AI credits.
func aiWrites() []writeCase {
	return slices.Concat(aiAgentWrites(), aiConnectionWrites(), aiCreditWrites())
}

// aiAgentWrites are the writes of agents: one agent's, an agent from a
// template, the export and import of agents and their attachments.
func aiAgentWrites() []writeCase {
	manage, chat := []string{"ai.agent_manage"}, []string{"ai.agentexecution_manage"}
	const agent = `{"name":{"en":"Complaints"},"userPrompt":"Resolve complaints.","triggers":[],"mcpServers":[],` +
		`"llmConfig":{"provider":"emporix_openai","model":"gpt-4o-mini"},"metadata":{"version":2}}`
	const export = `{"checksum":"de56","data":"eyJhZ2VudHMiOltdfQ==","exportedAt":"2026-10-14T10:00:00Z","jobId":"j1"}`
	return []writeCase{
		{
			// Emporix creates a missing agent, and answers with its id.
			args: []string{"ai", "agents", "update", "a1"}, file: agent, method: http.MethodPut,
			path: aiAgenticPath + "agents/a1", scopes: manage, lang: "*", status: http.StatusCreated,
			body: `{"id":"a1"}`, want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			args: []string{"ai", "agents", "patch", "a1"}, file: `[{"op":"REPLACE","path":"/enabled","value":false}]`,
			method: http.MethodPatch, path: aiAgenticPath + "agents/a1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"ai", "agents", "delete", "a1", "--force"}, method: http.MethodDelete,
			path: aiAgenticPath + "agents/a1", query: "force=true", scopes: manage, confirmBy: "a1",
			status: http.StatusNoContent,
		},
		{
			args:   []string{"ai", "agents", "create", "complaint-template"},
			file:   `{"name":"Complaints","description":"Resolves complaints","userPrompt":"Use get-order."}`,
			method: http.MethodPost, path: aiAgenticPath + "templates/complaint-template/agents", scopes: manage,
			lang: "*", status: http.StatusCreated, body: `{"id":"a2"}`, want: "{\n  \"id\": \"a2\"\n}\n",
		},
		{
			// The export needs a read scope only.
			args: []string{"ai", "agents", "export", "--agent", "a1", "--agent", "a 2"},
			sent: `{"agentIds":["a1","a 2"]}`, method: http.MethodPost, path: aiAgenticPath + "agents/export",
			scopes: []string{"ai.agent_read"}, status: http.StatusOK, body: export,
			want: "{\n  \"checksum\": \"de56\",\n  \"data\": \"eyJhZ2VudHMiOltdfQ==\",\n" +
				"  \"exportedAt\": \"2026-10-14T10:00:00Z\",\n  \"jobId\": \"j1\"\n}\n",
		},
		{
			args: []string{"ai", "agents", "export"}, file: `{"agentIds":["a1"]}`, method: http.MethodPost,
			path: aiAgenticPath + "agents/export", scopes: []string{"ai.agent_read"}, status: http.StatusOK,
			body: `{"jobId":"j1"}`, want: "{\n  \"jobId\": \"j1\"\n}\n",
		},
		{
			args: []string{"ai", "agents", "import"}, file: export, method: http.MethodPost,
			path: aiAgenticPath + "agents/import", scopes: manage, status: http.StatusOK,
			body: `{"message":"Imported 1 agents.","jobId":"j2"}`,
			want: "{\n  \"jobId\": \"j2\",\n  \"message\": \"Imported 1 agents.\"\n}\n",
		},
		{
			// The file is sent with its name and the content type of its extension.
			args: []string{"ai", "agents", "attach", "a1", "--upload", "testdata/order.csv"},
			form: []sentPart{{name: "attachment", file: "order.csv", contentType: "text/csv",
				content: orderCSV()}},
			method: http.MethodPost, path: aiAgenticPath + "a1/attachments", scopes: chat,
			status: http.StatusCreated, body: `{"id":"att1","sessionId":"s1"}`,
			want: "{\n  \"id\": \"att1\",\n  \"sessionId\": \"s1\"\n}\n",
		},
		{
			// A .text file is sent with the content type of its extension: text/plain.
			args: []string{"ai", "agents", "attach", "a1", "--upload", "testdata/order.text"},
			form: []sentPart{{name: "attachment", file: "order.text", contentType: "text/plain",
				content: orderText()}},
			method: http.MethodPost, path: aiAgenticPath + "a1/attachments", scopes: chat,
			status: http.StatusCreated, body: `{"id":"att2","sessionId":"s1"}`,
			want: "{\n  \"id\": \"att2\",\n  \"sessionId\": \"s1\"\n}\n",
		},
		{
			args:   []string{"ai", "agents", "attach", "a1", "--media", "m1"},
			form:   []sentPart{{name: "attachmentId", content: "m1"}},
			method: http.MethodPost, path: aiAgenticPath + "a1/attachments", scopes: chat,
			status: http.StatusOK, body: `{"id":"m1","sessionId":"s2"}`,
			want: "{\n  \"id\": \"m1\",\n  \"sessionId\": \"s2\"\n}\n",
		},
		{
			args: []string{"ai", "jobs", "delete", "j1"}, method: http.MethodDelete, path: aiServicePath + "jobs/j1",
			scopes: manage, confirmBy: "j1", status: http.StatusNoContent,
		},
	}
}

// aiConnectionWrites are the writes of what agents connect with: tools,
// tokens, OAuth configurations and MCP servers, whose update and patch take
// --force where Emporix disables what uses them, and whose delete --force.
func aiConnectionWrites() []writeCase {
	manage := []string{"ai.agent_manage"}
	var cases []writeCase
	for _, tt := range []struct {
		use, id, item string
		disables      bool // update and patch take --force
	}{
		{use: "tools", id: "t1", item: `{"type":"rag_emporix","name":"Products","config":{}}`, disables: true},
		{use: "tokens", id: "k1", item: `{"name":"OpenAI key","value":"sk-1"}`},
		{use: "oauths", id: "o1", item: `{"url":"https://idp.acme.test/token","clientId":"c1",` +
			`"grantType":"client_credentials"}`, disables: true},
		{use: "mcp-servers", id: "m1", item: `{"type":"custom","name":"Orders","transport":"sse",` +
			`"config":{"url":"https://mcp.acme.test"}}`, disables: true},
	} {
		path := aiAgenticPath + tt.use + "/" + tt.id
		off, on := "", "" // the query of update and patch without --force and with it
		if tt.disables {
			off, on = "force=false", "force=true"
		}
		patch := []string{"ai", tt.use, "patch", tt.id}
		if tt.disables {
			patch = append(patch, "--force")
		}
		cases = append(cases,
			writeCase{
				args: []string{"ai", tt.use, "update", tt.id}, file: tt.item, method: http.MethodPut, path: path,
				query: off, scopes: manage, status: http.StatusNoContent,
			},
			writeCase{
				args: patch, file: `[{"op":"REPLACE","path":"/name","value":"Renamed"}]`, method: http.MethodPatch,
				path: path, query: on, scopes: manage, status: http.StatusNoContent,
			},
			writeCase{
				args: []string{"ai", tt.use, "delete", tt.id}, method: http.MethodDelete, path: path,
				query: "force=false", scopes: manage, confirmBy: tt.id, status: http.StatusNoContent,
			},
		)
	}
	return cases
}

// aiCreditWrites are the writes that consume AI credits: chats with agents,
// text generation and completions.
func aiCreditWrites() []writeCase {
	chat := []string{"ai.agentexecution_manage"}
	const reply = `[{"agentId":"a1","agentType":"generic","message":"Order EON1243 ships today.","sessionId":"s1"}]`
	return []writeCase{
		{
			args: []string{"ai", "chat", "--agent", "a1", "--message", "Where is order EON1243?"},
			sent: `{"agentId":"a1","message":"Where is order EON1243?"}`, method: http.MethodPost,
			path: aiAgenticPath + "chat", scopes: chat, status: http.StatusOK, body: reply,
			want: "[\n  {\n    \"agentId\": \"a1\",\n    \"agentType\": \"generic\",\n" +
				"    \"message\": \"Order EON1243 ships today.\",\n    \"sessionId\": \"s1\"\n  }\n]\n",
		},
		{
			args: []string{"ai", "chat"}, file: `{"agentId":"a1","message":"Hi","attachments":[{"attachmentId":"att1",` +
				`"caption":"Order"}]}`,
			method: http.MethodPost, path: aiAgenticPath + "chat", scopes: chat, status: http.StatusOK, body: `[]`,
			want: "[]\n",
		},
		{
			args: []string{"ai", "chat-async", "--agent", "a1", "--message", "Hi"}, sent: `{"agentId":"a1","message":"Hi"}`,
			method: http.MethodPost, path: aiAgenticPath + "chat-async", scopes: chat, status: http.StatusCreated,
			body: `{"jobId":"j3"}`, want: "{\n  \"jobId\": \"j3\"\n}\n",
		},
		{
			args: []string{"ai", "chat-async"}, file: `{"agentId":"a1","message":"Hi"}`, method: http.MethodPost,
			path: aiAgenticPath + "chat-async", scopes: chat, status: http.StatusCreated, body: `{"jobId":"j3"}`,
			want: "{\n  \"jobId\": \"j3\"\n}\n",
		},
		{
			args: []string{"ai", "generate-text", "--prompt", "Describe product 100200", "--max-tokens", "500"},
			sent: `{"maxTokens":500,"prompt":"Describe product 100200"}`, method: http.MethodPost,
			path: aiServicePath + "texts", scopes: []string{"ai.text_manage"}, status: http.StatusOK,
			body: `{"result":"A pipe cutter."}`, want: "{\n  \"result\": \"A pipe cutter.\"\n}\n",
		},
		{
			args: []string{"ai", "generate-text"}, file: `{"id":"en","prompt":"Describe product 100200"}`,
			method: http.MethodPost, path: aiServicePath + "texts", scopes: []string{"ai.text_manage"}, status: http.StatusOK,
			body: `{"id":"en","result":"A pipe cutter."}`, want: "{\n  \"id\": \"en\",\n  \"result\": \"A pipe cutter.\"\n}\n",
		},
		{
			// A role is sent upper-cased, the text without the spaces around it.
			args: []string{"ai", "generate-completion", "--message", "SYSTEM:You write product texts.", "--message",
				" user : Describe product 100200: a pipe cutter "},
			sent: `{"messages":[{"role":"SYSTEM","content":"You write product texts."},` +
				`{"role":"USER","content":"Describe product 100200: a pipe cutter"}]}`,
			method: http.MethodPost, path: aiServicePath + "completions", scopes: []string{"ai.completion_manage"},
			status: http.StatusOK, body: `{"result":"Done."}`, want: "{\n  \"result\": \"Done.\"\n}\n",
		},
		{
			args: []string{"ai", "generate-completion"}, file: `{"messages":[{"role":"USER","content":"Hi"}]}`,
			method: http.MethodPost, path: aiServicePath + "completions", scopes: []string{"ai.completion_manage"},
			status: http.StatusOK, body: `{"result":"Hello."}`, want: "{\n  \"result\": \"Hello.\"\n}\n",
		},
	}
}

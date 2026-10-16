package cli

import (
	"net/http"
	"slices"
)

// aiReads returns the reads of the AI Service: agent templates, agents and
// what they are built from, their searches, the LLM models and the commerce
// events; what agents did (reads_ai_activity_test.go) and the AI RAG
// Indexer's metadata (reads_ai_rag_test.go).
func aiReads() []readCase {
	return slices.Concat(aiAgentReads(), aiToolReads(), aiConnectionReads(), aiJobReads(), aiLogReads(),
		aiAnalyticsReads(), aiRAGReads())
}

// aiAgentReads returns the reads of agent templates, agents, LLM models and
// commerce events.
func aiAgentReads() []readCase {
	const scope = "ai.agent_read"
	templates := `[{"id":"complaint-template","type":"complaint","enabled":true,"name":{"en":"Complaints","nl":"Klachten"},
		"llmConfig":{"provider":"emporix_openai","model":"gpt-4.1-mini","temperature":0.2},"userPrompt":"Handle it."},
		{"id":"blank","type":"generic","name":{"de":"Leer"},"llmConfig":{}}]`
	agents := `[{"id":"complaint-agent","type":"complaint","enabled":false,"name":{"en":"Complaints","nl":"Klachten"},
		"llmConfig":{"provider":"anthropic","model":"claude-haiku","token":{"id":"anthropic-key"}},
		"triggers":[{"type":"endpoint"},{"type":"commerce_events","config":{"events":"order.created"}}]},
		{"id":"helper","type":"generic","enabled":true,"name":{"en":"Helper"},"llmConfig":{"provider":"openai",
		"model":"gpt-4.1"},"triggers":[]}]`
	return []readCase{
		{
			// PROVIDER and MODEL name the LLM; prompts are only in -o json.
			args:    []string{"ai", "templates", "list", "--language", "nl"},
			path:    "/ai-service/acme1/agentic/templates",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    templates,
			want: "ID                  TYPE       ENABLED  PROVIDER        MODEL         NAME\n" +
				"complaint-template  complaint  true     emporix_openai  gpt-4.1-mini  Klachten\n" +
				"blank               generic                                           Leer (de)\n",
		},
		{
			// Without --query, the body's q filters nothing; pages are read as list reads them.
			args:    []string{"ai", "templates", "search", "--language", "en"},
			path:    "/ai-service/acme1/agentic/templates/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    templates,
			want: "ID                  TYPE       ENABLED  PROVIDER        MODEL         NAME\n" +
				"complaint-template  complaint  true     emporix_openai  gpt-4.1-mini  Complaints\n" +
				"blank               generic                                           Leer (de)\n",
		},
		{
			// TRIGGERS names what starts an agent; their configuration is only in -o json.
			args:    []string{"ai", "agents", "list", "--language", "nl"},
			path:    "/ai-service/acme1/agentic/agents",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    agents,
			want: "ID               TYPE       ENABLED  PROVIDER   MODEL         TRIGGERS                   NAME\n" +
				"complaint-agent  complaint  false    anthropic  claude-haiku  endpoint, commerce_events  Klachten\n" +
				"helper           generic    true     openai     gpt-4.1                                  Helper (en)\n",
		},
		{
			args:   []string{"ai", "agents", "get", "complaint-agent", "--language", "en"},
			path:   "/ai-service/acme1/agentic/agents/complaint-agent",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"complaint-agent","type":"complaint","enabled":true,"name":{"en":"Complaints"},
				"llmConfig":{"provider":"self_hosted_ollama","model":"llama3"},"triggers":[{"type":"slack"}]}`,
			want: "ID               TYPE       ENABLED  PROVIDER            MODEL   TRIGGERS  NAME\n" +
				"complaint-agent  complaint  true     self_hosted_ollama  llama3  slack     Complaints\n",
		},
		{
			// --query is sent in the body as given.
			args:    []string{"ai", "agents", "search", "--query", "enabled:true type:generic", "--language", "en"},
			path:    "/ai-service/acme1/agentic/agents/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"enabled:true type:generic"}`,
			body:    `[{"id":"helper","type":"generic","enabled":true,"name":{"en":"Helper"},"triggers":[{"type":"teams"}]}]`,
			want: "ID      TYPE     ENABLED  PROVIDER  MODEL  TRIGGERS  NAME\n" +
				"helper  generic  true                      teams     Helper\n",
		},
		{
			// One row per model of every provider; descriptions are only in -o json.
			args:   []string{"ai", "models"},
			path:   "/ai-service/acme1/agentic/models",
			scopes: []string{scope},
			mode:   getWhole,
			body: `[{"provider":"openai","models":[{"id":"gpt-4.1-mini","name":"GPT-4.1 mini","thinking":false},
				{"id":"o3","name":"o3","thinking":true}]},{"provider":"anthropic","models":[]},
				{"provider":"google","models":[{"id":"gemini-flash","name":"Gemini Flash","description":"Fast"}]}]`,
			want: "ID            PROVIDER  THINKING  NAME\n" +
				"gpt-4.1-mini  openai    false     GPT-4.1 mini\n" +
				"o3            openai    true      o3\n" +
				"gemini-flash  google              Gemini Flash\n",
		},
		{
			// Sorted by name, whatever order Emporix sends them in.
			args:   []string{"ai", "commerce-events"},
			path:   "/ai-service/acme1/agentic/commerce-events",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"events":["order.created","customer.updated","cart.closed"]}`,
			want:   "EVENT\ncart.closed\ncustomer.updated\norder.created\n",
		},
	}
}

// aiToolReads returns the reads of tools and tokens, and their searches.
func aiToolReads() []readCase {
	const scope = "ai.agent_read"
	tools := `[{"id":"slack-tool","type":"slack","name":"Slack","enabled":true,"config":{"teamId":"T1"}},
		{"id":"faq","type":"rag_emporix","name":"FAQ","config":{"entityType":"PRODUCT"}}]`
	named := `[{"id":"openai-key","name":"OpenAI key","metadata":{"version":1,"createdAt":"2026-05-02T08:00:00Z",
		"modifiedAt":"2026-05-03T09:00:00Z"}},{"id":"mcp-key","name":"MCP key"}]`
	return []readCase{
		{
			args:    []string{"ai", "tools", "list"},
			path:    "/ai-service/acme1/agentic/tools",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    tools,
			want: "ID          TYPE         ENABLED  NAME\n" +
				"slack-tool  slack        true     Slack\n" +
				"faq         rag_emporix           FAQ\n",
		},
		{
			args:   []string{"ai", "tools", "get", "faq"},
			path:   "/ai-service/acme1/agentic/tools/faq",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"faq","type":"rag_custom","name":"FAQ","enabled":false}`,
			want:   "ID   TYPE        ENABLED  NAME\nfaq  rag_custom  false    FAQ\n",
		},
		{
			args:    []string{"ai", "tools", "search", "--query", "type:slack"},
			path:    "/ai-service/acme1/agentic/tools/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"type:slack"}`,
			body:    `[{"id":"slack-tool","type":"slack","name":"Slack","enabled":true}]`,
			want:    "ID          TYPE   ENABLED  NAME\nslack-tool  slack  true     Slack\n",
		},
		{
			// A token's name and dates; a value Emporix might send is only in -o json.
			args:    []string{"ai", "tokens", "list"},
			path:    "/ai-service/acme1/agentic/tokens",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    named,
			want: "ID          NAME        CREATED               MODIFIED\n" +
				"openai-key  OpenAI key  2026-05-02T08:00:00Z  2026-05-03T09:00:00Z\n" +
				"mcp-key     MCP key                           \n",
		},
		{
			args:   []string{"ai", "tokens", "get", "mcp-key"},
			path:   "/ai-service/acme1/agentic/tokens/mcp-key",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"mcp-key","name":"MCP key","value":"not-shown-in-tables"}`,
			want:   "ID       NAME     CREATED  MODIFIED\nmcp-key  MCP key           \n",
		},
		{
			args:    []string{"ai", "tokens", "search"},
			path:    "/ai-service/acme1/agentic/tokens/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    `[{"id":"mcp-key","name":"MCP key"}]`,
			want:    "ID       NAME     CREATED  MODIFIED\nmcp-key  MCP key           \n",
		},
	}
}

// aiConnectionReads returns the reads of OAuth configurations and MCP
// servers, and their searches.
func aiConnectionReads() []readCase {
	const scope = "ai.agent_read"
	oauths := `[{"id":"llm-auth","url":"https://auth.example.com/token","clientId":"acme-client","grantType":
		"client_credentials","scope":"llm.read","enabled":true,"clientSecretToken":{"id":"llm-secret"}},
		{"id":"other","url":"https://auth.example.org/token","clientId":"c2","grantType":"client_credentials"}]`
	servers := `[{"id":"catalog","type":"custom","name":"Catalog","enabled":true,"transport":"streamable_http",
		"config":{"url":"https://mcp.example.com/mcp","authorizationHeaderName":"Authorization",
		"authorizationHeaderToken":{"id":"mcp-key"}}},{"id":"erp","type":"dynamic","name":"ERP","transport":"sse",
		"tools":[{"name":"stock"}]}]`
	return []readCase{
		{
			// SECRET TOKEN is the ID of the token that holds the client secret; the scope is only in -o json.
			args:    []string{"ai", "oauths", "list"},
			path:    "/ai-service/acme1/agentic/oauths",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    oauths,
			want: "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\n" +
				"llm-auth  true     acme-client  llm-secret    https://auth.example.com/token\n" +
				"other              c2                         https://auth.example.org/token\n",
		},
		{
			args:   []string{"ai", "oauths", "get", "llm-auth"},
			path:   "/ai-service/acme1/agentic/oauths/llm-auth",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"llm-auth","url":"https://auth.example.com/token","clientId":"acme-client","enabled":false,
				"clientSecretToken":{"id":"llm-secret","name":"LLM secret"}}`,
			want: "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\n" +
				"llm-auth  false    acme-client  llm-secret    https://auth.example.com/token\n",
		},
		{
			args:    []string{"ai", "oauths", "search", "--query", "enabled:true"},
			path:    "/ai-service/acme1/agentic/oauths/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"enabled:true"}`,
			body:    `[{"id":"llm-auth","clientId":"acme-client","enabled":true}]`,
			want:    "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\nllm-auth  true     acme-client                \n",
		},
		{
			// URL is a custom server's; a dynamic server's tools are only in -o json.
			args:    []string{"ai", "mcp-servers", "list"},
			path:    "/ai-service/acme1/agentic/mcp-servers",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    servers,
			want: "ID       TYPE     ENABLED  TRANSPORT        NAME     URL\n" +
				"catalog  custom   true     streamable_http  Catalog  https://mcp.example.com/mcp\n" +
				"erp      dynamic           sse              ERP      \n",
		},
		{
			args:   []string{"ai", "mcp-servers", "get", "erp"},
			path:   "/ai-service/acme1/agentic/mcp-servers/erp",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"id":"erp","type":"dynamic","name":"ERP","enabled":true,"transport":"sse"}`,
			want:   "ID   TYPE     ENABLED  TRANSPORT  NAME  URL\nerp  dynamic  true     sse        ERP   \n",
		},
		{
			args:    []string{"ai", "mcp-servers", "search", "--query", "type:custom"},
			path:    "/ai-service/acme1/agentic/mcp-servers/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"type:custom"}`,
			body:    `[{"id":"catalog","type":"custom","name":"Catalog","config":{"url":"https://mcp.example.com/mcp"}}]`,
			want: "ID       TYPE    ENABLED  TRANSPORT  NAME     URL\n" +
				"catalog  custom                      Catalog  https://mcp.example.com/mcp\n",
		},
	}
}

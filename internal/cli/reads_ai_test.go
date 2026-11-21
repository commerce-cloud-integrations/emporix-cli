package cli

import (
	"net/http"
	"slices"
)

func aiReads() []readCase {
	return slices.Concat(aiAgentReads(), aiToolReads(), aiConnectionReads(), aiJobReads(), aiLogReads(),
		aiAnalyticsReads(), aiRAGReads())
}

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
			args:     []string{"ai", "templates", "list", "--language", "nl"},
			path:     "/ai-service/acme1/agentic/templates",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: templates,
			want: "ID                  TYPE       ENABLED  PROVIDER        MODEL         NAME\n" +
				"complaint-template  complaint  true     emporix_openai  gpt-4.1-mini  Klachten\n" +
				"blank               generic                                           Leer (de)\n",
		},
		{
			args:        []string{"ai", "templates", "search", "--language", "en"},
			path:        "/ai-service/acme1/agentic/templates/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":""}`,
			response:    templates,
			want: "ID                  TYPE       ENABLED  PROVIDER        MODEL         NAME\n" +
				"complaint-template  complaint  true     emporix_openai  gpt-4.1-mini  Complaints\n" +
				"blank               generic                                           Leer (de)\n",
		},
		{
			args:     []string{"ai", "agents", "list", "--language", "nl"},
			path:     "/ai-service/acme1/agentic/agents",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: agents,
			want: "ID               TYPE       ENABLED  PROVIDER   MODEL         TRIGGERS                   NAME\n" +
				"complaint-agent  complaint  false    anthropic  claude-haiku  endpoint, commerce_events  Klachten\n" +
				"helper           generic    true     openai     gpt-4.1                                  Helper (en)\n",
		},
		{
			args:   []string{"ai", "agents", "get", "complaint-agent", "--language", "en"},
			path:   "/ai-service/acme1/agentic/agents/complaint-agent",
			scopes: []string{scope},
			mode:   getWhole,
			response: `{"id":"complaint-agent","type":"complaint","enabled":true,"name":{"en":"Complaints"},
				"llmConfig":{"provider":"self_hosted_ollama","model":"llama3"},"triggers":[{"type":"slack"}]}`,
			want: "ID               TYPE       ENABLED  PROVIDER            MODEL   TRIGGERS  NAME\n" +
				"complaint-agent  complaint  true     self_hosted_ollama  llama3  slack     Complaints\n",
		},
		{
			args:        []string{"ai", "agents", "search", "--query", "enabled:true type:generic", "--language", "en"},
			path:        "/ai-service/acme1/agentic/agents/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":"enabled:true type:generic"}`,
			response: `[{"id":"helper","type":"generic","enabled":true,"name":{"en":"Helper"},` +
				`"triggers":[{"type":"teams"}]}]`,
			want: "ID      TYPE     ENABLED  PROVIDER  MODEL  TRIGGERS  NAME\n" +
				"helper  generic  true                      teams     Helper\n",
		},
		{
			args:   []string{"ai", "models"},
			path:   "/ai-service/acme1/agentic/models",
			scopes: []string{scope},
			mode:   getWhole,
			response: `[{"provider":"openai","models":[{"id":"gpt-4.1-mini","name":"GPT-4.1 mini","thinking":false},
				{"id":"o3","name":"o3","thinking":true}]},{"provider":"anthropic","models":[]},
				{"provider":"google","models":[{"id":"gemini-flash","name":"Gemini Flash","description":"Fast"}]}]`,
			want: "ID            PROVIDER  THINKING  NAME\n" +
				"gpt-4.1-mini  openai    false     GPT-4.1 mini\n" +
				"o3            openai    true      o3\n" +
				"gemini-flash  google              Gemini Flash\n",
		},
		{
			args:     []string{"ai", "commerce-events"},
			path:     "/ai-service/acme1/agentic/commerce-events",
			scopes:   []string{scope},
			mode:     getWhole,
			response: `{"events":["order.created","customer.updated","cart.closed"]}`,
			want:     "EVENT\ncart.closed\ncustomer.updated\norder.created\n",
		},
	}
}

const searchTool = `{"id":"search-customer","type":"search","name":"Customer search","enabled":true,` +
	`"config":{"prompt":"Find customers by name or number.","entityType":"customer","maxResults":5}}`

func aiToolReads() []readCase {
	const scope = "ai.agent_read"
	tools := `[{"id":"slack-tool","type":"slack","name":"Slack","enabled":true,"config":{"teamId":"T1"}},
		{"id":"faq","type":"rag_emporix","name":"FAQ","config":{"entityType":"PRODUCT"}},` + searchTool + `]`
	named := `[{"id":"openai-key","name":"OpenAI key","metadata":{"version":1,"createdAt":"2026-05-02T08:00:00Z",
		"modifiedAt":"2026-05-03T09:00:00Z"}},{"id":"mcp-key","name":"MCP key"}]`
	return []readCase{
		{
			args:     []string{"ai", "tools", "list"},
			path:     "/ai-service/acme1/agentic/tools",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: tools,
			want: "ID               TYPE         ENABLED  NAME\n" +
				"slack-tool       slack        true     Slack\n" +
				"faq              rag_emporix           FAQ\n" +
				"search-customer  search       true     Customer search\n",
		},
		{
			args:     []string{"ai", "tools", "get", "faq"},
			path:     "/ai-service/acme1/agentic/tools/faq",
			scopes:   []string{scope},
			mode:     getWhole,
			response: `{"id":"faq","type":"rag_custom","name":"FAQ","enabled":false}`,
			want:     "ID   TYPE        ENABLED  NAME\nfaq  rag_custom  false    FAQ\n",
		},
		{
			args:        []string{"ai", "tools", "search", "--query", "type:slack"},
			path:        "/ai-service/acme1/agentic/tools/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":"type:slack"}`,
			response:    `[{"id":"slack-tool","type":"slack","name":"Slack","enabled":true}]`,
			want:        "ID          TYPE   ENABLED  NAME\nslack-tool  slack  true     Slack\n",
		},
		{
			args:     []string{"ai", "tokens", "list"},
			path:     "/ai-service/acme1/agentic/tokens",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: named,
			want: "ID          NAME        CREATED               MODIFIED\n" +
				"openai-key  OpenAI key  2026-05-02T08:00:00Z  2026-05-03T09:00:00Z\n" +
				"mcp-key     MCP key                           \n",
		},
		{
			args:     []string{"ai", "tokens", "get", "mcp-key"},
			path:     "/ai-service/acme1/agentic/tokens/mcp-key",
			scopes:   []string{scope},
			mode:     getWhole,
			response: `{"id":"mcp-key","name":"MCP key","value":"not-shown-in-tables"}`,
			want:     "ID       NAME     CREATED  MODIFIED\nmcp-key  MCP key           \n",
		},
		{
			args:        []string{"ai", "tokens", "search"},
			path:        "/ai-service/acme1/agentic/tokens/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":""}`,
			response:    `[{"id":"mcp-key","name":"MCP key"}]`,
			want:        "ID       NAME     CREATED  MODIFIED\nmcp-key  MCP key           \n",
		},
	}
}

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
			args:     []string{"ai", "oauths", "list"},
			path:     "/ai-service/acme1/agentic/oauths",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: oauths,
			want: "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\n" +
				"llm-auth  true     acme-client  llm-secret    https://auth.example.com/token\n" +
				"other              c2                         https://auth.example.org/token\n",
		},
		{
			args:   []string{"ai", "oauths", "get", "llm-auth"},
			path:   "/ai-service/acme1/agentic/oauths/llm-auth",
			scopes: []string{scope},
			mode:   getWhole,
			response: `{"id":"llm-auth","url":"https://auth.example.com/token","clientId":"acme-client","enabled":false,
				"clientSecretToken":{"id":"llm-secret","name":"LLM secret"}}`,
			want: "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\n" +
				"llm-auth  false    acme-client  llm-secret    https://auth.example.com/token\n",
		},
		{
			args:        []string{"ai", "oauths", "search", "--query", "enabled:true"},
			path:        "/ai-service/acme1/agentic/oauths/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":"enabled:true"}`,
			response:    `[{"id":"llm-auth","clientId":"acme-client","enabled":true}]`,
			want: "ID        ENABLED  CLIENT ID    SECRET TOKEN  TOKEN URL\n" +
				"llm-auth  true     acme-client                \n",
		},
		{
			args:     []string{"ai", "mcp-servers", "list"},
			path:     "/ai-service/acme1/agentic/mcp-servers",
			scopes:   []string{scope},
			filters:  listFilters{kinds: filterable},
			response: servers,
			want: "ID       TYPE     ENABLED  TRANSPORT        NAME     URL\n" +
				"catalog  custom   true     streamable_http  Catalog  https://mcp.example.com/mcp\n" +
				"erp      dynamic           sse              ERP      \n",
		},
		{
			args:     []string{"ai", "mcp-servers", "get", "erp"},
			path:     "/ai-service/acme1/agentic/mcp-servers/erp",
			scopes:   []string{scope},
			mode:     getWhole,
			response: `{"id":"erp","type":"dynamic","name":"ERP","enabled":true,"transport":"sse"}`,
			want:     "ID   TYPE     ENABLED  TRANSPORT  NAME  URL\nerp  dynamic  true     sse        ERP   \n",
		},
		{
			args:        []string{"ai", "mcp-servers", "search", "--query", "type:custom"},
			path:        "/ai-service/acme1/agentic/mcp-servers/search",
			scopes:      []string{scope},
			filters:     listFilters{kinds: sortable},
			method:      http.MethodPost,
			requestBody: `{"q":"type:custom"}`,
			response:    `[{"id":"catalog","type":"custom","name":"Catalog","config":{"url":"https://mcp.example.com/mcp"}}]`,
			want: "ID       TYPE    ENABLED  TRANSPORT  NAME     URL\n" +
				"catalog  custom                      Catalog  https://mcp.example.com/mcp\n",
		},
	}
}

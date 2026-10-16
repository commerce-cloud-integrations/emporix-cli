package cli

import "net/http"

// aiJobReads returns the reads of the AI Service's jobs and of the
// conversations of agents, and their searches (ai_activity.go).
func aiJobReads() []readCase {
	const scope = "ai.agent_read"
	jobs := `[{"id":"job-1","type":"AGENT_CHAT","status":"success","agentType":"complaint","agentId":"complaint-agent",
		"sessionId":"sess-1","commerceEvent":"order.created","message":"Done.","metadata":{"version":1,
		"createdAt":"2026-09-01T10:30:27Z","modifiedAt":"2026-09-01T10:31:14Z"}},
		{"id":"job-2","type":"EXPORT","status":"in_progress","exportResult":{"data":"e30=","checksum":"c1"}}]`
	conversations := `[{"conversationId":"19:abc@thread.tacv2","conversationName":"Support","contextRef":"order-100200",
		"channelId":"ch-1","agentId":"complaint-agent","sessionId":"sess-1","lastMessageAt":"2026-09-02T08:00:00Z"},
		{"conversationId":"19:def@thread.v2","agentId":"helper"}]`
	return []readCase{
		{
			// The agent's message and an export's result are only in -o json.
			args:    []string{"ai", "jobs", "list"},
			path:    "/ai-service/acme1/jobs",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    jobs,
			want: "ID     TYPE        STATUS       AGENT            SESSION  EVENT          CREATED\n" +
				"job-1  AGENT_CHAT  success      complaint-agent  sess-1   order.created  2026-09-01T10:30:27Z\n" +
				"job-2  EXPORT      in_progress                                           \n",
		},
		{
			args:   []string{"ai", "jobs", "get", "job-2"},
			path:   "/ai-service/acme1/jobs/job-2",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"job-2","type":"IMPORT","status":"failure","importResult":{"message":"1 failed"},
				"metadata":{"createdAt":"2026-09-03T07:00:00Z"}}`,
			want: "ID     TYPE    STATUS   AGENT  SESSION  EVENT  CREATED\n" +
				"job-2  IMPORT  failure                         2026-09-03T07:00:00Z\n",
		},
		{
			args:    []string{"ai", "jobs", "search", "--query", "status:failure"},
			path:    "/ai-service/acme1/jobs/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"status:failure"}`,
			body:    `[{"id":"job-3","type":"AGENT_CHAT","status":"failure","agentId":"helper"}]`,
			want: "ID     TYPE        STATUS   AGENT   SESSION  EVENT  CREATED\n" +
				"job-3  AGENT_CHAT  failure  helper                  \n",
		},
		{
			// ID is the conversation in Teams; its channel and context reference are only in -o json.
			args:    []string{"ai", "conversations", "list"},
			path:    "/ai-service/acme1/agentic/conversations",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    conversations,
			want: "ID                   NAME     AGENT            SESSION  LAST MESSAGE\n" +
				"19:abc@thread.tacv2  Support  complaint-agent  sess-1   2026-09-02T08:00:00Z\n" +
				"19:def@thread.v2              helper                    \n",
		},
		{
			args:    []string{"ai", "conversations", "search"},
			path:    "/ai-service/acme1/agentic/conversations/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    `[{"conversationId":"19:def@thread.v2","conversationName":"Sales","agentId":"helper"}]`,
			want: "ID                NAME   AGENT   SESSION  LAST MESSAGE\n" +
				"19:def@thread.v2  Sales  helper           \n",
		},
	}
}

// aiLogReads returns the reads of the logs of agent requests and sessions,
// and their searches (ai_activity.go).
func aiLogReads() []readCase {
	const scope = "ai.agent_read"
	requests := `[{"id":"log-1","sessionId":"sess-1","requestId":"req-1","triggerAgentId":"complaint-agent",
		"collaborationAgents":["categorizer"],"severity":"WARNING","messages":[{"severity":"INFO","message":"Started"},
		{"severity":"WARNING","message":"Unexpected status"}],"duration":45,"promptTokens":1234,"completionTokens":567,
		"agentVersion":3,"metadata":{"version":1,"createdAt":"2026-09-01T10:30:27Z"}},
		{"id":"log-2","requestId":"req-2","sessionId":"sess-1","messages":[]},
		{"id":"log-3","requestId":"req-3","sessionId":"sess-2","severity":"INFO","messages":null}]`
	sessions := `[{"sessionId":"sess-1","triggerAgentId":"complaint-agent","severity":"INFO","agents":["complaint-agent",
		"categorizer"],"messages":[{"severity":"INFO","message":"Session started"}],"duration":120,"promptTokens":2468,
		"metadata":{"createdAt":"2026-09-01T10:30:27Z"}},{"sessionId":"sess-2","severity":"ERROR"}]`
	return []readCase{
		{
			// The table ID is the log's id, which get takes. MESSAGES counts the messages, and stays empty
			// without them; they, the other agents and the tokens are only in -o json.
			args:    []string{"ai", "logs", "requests", "list"},
			path:    "/ai-service/acme1/agentic/logs/requests",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    requests,
			want: "ID     REQUEST  SESSION  AGENT            SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"log-1  req-1    sess-1   complaint-agent  WARNING   2         45       2026-09-01T10:30:27Z\n" +
				"log-2  req-2    sess-1                              0                  \n" +
				"log-3  req-3    sess-2                    INFO                         \n",
		},
		{
			// The table ID is the log's id, which get takes.
			args:   []string{"ai", "logs", "requests", "get", "log-1"},
			path:   "/ai-service/acme1/agentic/logs/requests/log-1",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"log-1","requestId":"req-1","sessionId":"sess-1","triggerAgentId":"helper","severity":"ERROR",
				"duration":3}`,
			want: "ID     REQUEST  SESSION  AGENT   SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"log-1  req-1    sess-1   helper  ERROR               3        \n",
		},
		{
			args:    []string{"ai", "logs", "requests", "search", "--query", "severity:ERROR"},
			path:    "/ai-service/acme1/agentic/logs/requests/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"severity:ERROR"}`,
			body: `[{"id":"log-4","requestId":"req-4","sessionId":"sess-2","severity":"ERROR",
				"messages":[{"message":"Failed"}]}]`,
			want: "ID     REQUEST  SESSION  AGENT  SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"log-4  req-4    sess-2          ERROR     1                  \n",
		},
		{
			args:    []string{"ai", "logs", "sessions", "list"},
			path:    "/ai-service/acme1/agentic/logs/sessions",
			scopes:  []string{scope},
			filters: listFilters{kinds: filterable},
			body:    sessions,
			want: "ID      AGENT            SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"sess-1  complaint-agent  INFO      1         120      2026-09-01T10:30:27Z\n" +
				"sess-2                   ERROR                        \n",
		},
		{
			args:   []string{"ai", "logs", "sessions", "get", "sess-1"},
			path:   "/ai-service/acme1/agentic/logs/sessions/sess-1",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"sessionId":"sess-1","triggerAgentId":"helper","severity":"INFO","messages":[{},{}],"duration":7}`,
			want: "ID      AGENT   SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"sess-1  helper  INFO      2         7        \n",
		},
		{
			args:    []string{"ai", "logs", "sessions", "search", "--query", "triggerAgentId:helper"},
			path:    "/ai-service/acme1/agentic/logs/sessions/search",
			scopes:  []string{scope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"triggerAgentId:helper"}`,
			body:    `[{"sessionId":"sess-3","triggerAgentId":"helper","severity":"INFO"}]`,
			want: "ID      AGENT   SEVERITY  MESSAGES  SECONDS  CREATED\n" +
				"sess-3  helper  INFO                         \n",
		},
	}
}

// aiAnalyticsReads returns the reads of the analytics of agents
// (ai_analytics.go).
func aiAnalyticsReads() []readCase {
	const scope = "ai.agent_read"
	totals := `{"requests":{"total":120,"bySeverity":{"info":100,"warning":15,"error":5},"errorRate":0.0417},
		"sessions":{"total":40,"bySeverity":{"success":37,"warning":1,"error":2},"errorRate":0.05},
		"resolutionEfficiency":{"requestsPerSession":3,"totalRequests":120,"totalSessions":40},
		"sessionTrend":[{"date":"2026-09-07T00:00:00Z","totalSessions":10,"errorSessions":1,"errorRate":0.1}]}`
	return []readCase{
		{
			// The totals of requests, then of sessions; efficiency and trends are only in -o json.
			args:   []string{"ai", "analytics", "tenant"},
			path:   "/ai-service/acme1/agentic/analytics",
			scopes: []string{scope},
			mode:   getWhole,
			body:   totals,
			want: "LOGS      TOTAL  WARNING  ERROR  ERROR RATE\n" +
				"requests  120    15       5      0.0417\n" +
				"sessions  40     1        2      0.05\n",
		},
		{
			// A metric Emporix leaves out stays empty.
			args:   []string{"ai", "analytics", "agent", "complaint-agent"},
			path:   "/ai-service/acme1/agentic/analytics",
			query:  "agentId=complaint-agent",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `{"requests":{"total":0,"bySeverity":{"error":0}},"sessions":{"total":0}}`,
			want: "LOGS      TOTAL  WARNING  ERROR  ERROR RATE\n" +
				"requests  0               0      \n" +
				"sessions  0                      \n",
		},
		{
			// One row per agent and period, in their order, per quarter (Emporix's default granularity,
			// which epx leaves to it); a null count stays empty.
			args:   []string{"ai", "analytics", "executions", "complaint-agent,helper"},
			path:   "/ai-service/acme1/agentic/analytics/executions",
			query:  "agentIds=complaint-agent%2Chelper",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"granularity":"QUARTER","periods":[{"label":"2026-Q2","start":"2026-04-01T00:00:00Z",
				"end":"2026-07-01T00:00:00Z"},{"label":"2026-Q3","start":"2026-07-01T00:00:00Z","end":"2026-10-01T00:00:00Z"}],
				"series":[{"agentId":"complaint-agent","counts":[12,3]},{"agentId":"helper","counts":[0,null]}]}`,
			want: "AGENT            PERIOD   FROM                  TO                    EXECUTIONS\n" +
				"complaint-agent  2026-Q2  2026-04-01T00:00:00Z  2026-07-01T00:00:00Z  12\n" +
				"complaint-agent  2026-Q3  2026-07-01T00:00:00Z  2026-10-01T00:00:00Z  3\n" +
				"helper           2026-Q2  2026-04-01T00:00:00Z  2026-07-01T00:00:00Z  0\n" +
				"helper           2026-Q3  2026-07-01T00:00:00Z  2026-10-01T00:00:00Z  \n",
		},
		{
			// Per month or week with --granularity, sent as given.
			args:   []string{"ai", "analytics", "executions", "helper", "--granularity", "MONTH"},
			path:   "/ai-service/acme1/agentic/analytics/executions",
			query:  "agentIds=helper&granularity=MONTH",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"granularity":"MONTH","periods":[{"label":"2026-09","start":"2026-09-01T00:00:00Z",
				"end":"2026-10-01T00:00:00Z"}],"series":[{"agentId":"helper","counts":[4]}]}`,
			want: "AGENT   PERIOD   FROM                  TO                    EXECUTIONS\n" +
				"helper  2026-09  2026-09-01T00:00:00Z  2026-10-01T00:00:00Z  4\n",
		},
	}
}

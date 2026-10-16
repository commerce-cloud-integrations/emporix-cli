package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestAIAnalyticsRejectResponsesOfAnotherShape checks the analytics reads,
// whose rows are restructured from their response: a response of another
// shape, such as one with a count that is not a number, fails the read
// (exit 1), also with -o json, which prints it whole, and prints nothing.
func TestAIAnalyticsRejectResponsesOfAnotherShape(t *testing.T) {
	const totals, executions = "/ai-service/acme1/agentic/analytics", "/ai-service/acme1/agentic/analytics/executions"
	tenant := []string{"ai", "analytics", "tenant"}
	executed := []string{"ai", "analytics", "executions", "helper"}
	periods := `"periods":[{"label":"W1"},{"label":"W2"}]`
	for _, tt := range []struct {
		path, body string
		args       []string
		want       string
	}{
		{path: totals, body: `[{"requests":{}}]`, args: tenant,
			want: "decoding agent analytics response: expected an object with requests and sessions"},
		{path: totals, body: `{"requests":{"total":1}}`, args: append(tenant, "-o", "json"),
			want: "expected an object with requests and sessions"},
		{path: totals, body: `{"requests":{"total":1},"sessions":null}`, args: tenant,
			want: "expected an object with requests and sessions"},
		{path: totals, body: `{"requests":[1],"sessions":{}}`, args: tenant,
			want: "expected an object with requests and sessions"},
		{path: totals, body: `{"requests":{"total":{"all":1}},"sessions":{}}`, args: tenant,
			want: "expected an object with requests and sessions"},
		{path: totals, body: `{"requests":{},"sessions":{"bySeverity":{"error":"many"}}}`, args: tenant,
			want: "expected an object with requests and sessions"},
		{path: executions, body: `{` + periods + `}`, args: executed,
			want: "decoding agent executions response: expected an object with periods and series"},
		{path: executions, body: `{"periods":["W1"],"series":[]}`, args: append(executed, "-o", "json"),
			want: "expected an object with periods and series"},
		{path: executions, body: `{` + periods + `,"series":[{"agentId":"helper","counts":[1]}]}`, args: executed,
			want: "expected one count per period for each agent"},
		{path: executions, body: `{` + periods + `,"series":[{"agentId":"helper","counts":[[1],2]}]}`, args: executed,
			want: "expected an object with periods and series"},
		{path: executions, body: `{` + periods + `,"series":[{"counts":[1,2]}]}`, args: executed,
			want: "item 1 has no identifier"},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, tt.body))
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitError || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%v of %s: exit %d, stdout %q, stderr %q; want exit 1 and %q", tt.args, tt.body, code, stdout,
				stderr, tt.want)
		}
	}
}

// TestAIExecutionsOfNoAgent checks that executions without series print only
// the header, and that -o json prints the response as sent.
func TestAIExecutionsOfNoAgent(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/ai-service/acme1/agentic/analytics/executions",
		respond(http.StatusOK, `{"granularity":"QUARTER","periods":[{"label":"Q3"}],"series":[]}`))
	for args, want := range map[string]string{
		"ai analytics executions helper": "AGENT  PERIOD  FROM  TO  EXECUTIONS\n",
		"ai analytics executions helper -o json": "{\n  \"granularity\": \"QUARTER\",\n  \"periods\": [\n" +
			"    {\n      \"label\": \"Q3\"\n    }\n  ],\n  \"series\": []\n}\n",
	} {
		code, stdout, stderr := runCLI(t, fake.env(), strings.Fields(args)...)
		if code != ExitOK || stdout != want {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %q", args, code, stdout, stderr, want)
		}
	}
}

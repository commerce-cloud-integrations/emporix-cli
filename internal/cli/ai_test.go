package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestAIReadsRejectResponsesOfAnotherShape checks the LLM models and the
// commerce events, whose items are inside their response: a response of
// another shape fails the read (exit 1), also with -o json, which prints it
// whole, and prints nothing.
func TestAIReadsRejectResponsesOfAnotherShape(t *testing.T) {
	const models, events = "/ai-service/acme1/agentic/models", "/ai-service/acme1/agentic/commerce-events"
	for _, tt := range []struct {
		path, body string
		args       []string
		want       string
	}{
		{path: models, body: `{"provider":"openai"}`, args: []string{"ai", "models"},
			want: "decoding LLM model response: expected a JSON array"},
		{path: models, body: `[{"provider":"openai"}]`, args: []string{"ai", "models", "-o", "json"},
			want: "expected providers with their models"},
		{path: models, body: `[{"provider":"openai","models":{}}]`, args: []string{"ai", "models"},
			want: "expected providers with their models"},
		{path: models, body: `[{"provider":"openai","models":["o3"]}]`, args: []string{"ai", "models"},
			want: "expected providers with their models"},
		{path: models, body: `[{"provider":"openai","models":[{"name":"No id"}]}]`, args: []string{"ai", "models"},
			want: "item 1 has no identifier"},
		{path: events, body: `["order.created"]`, args: []string{"ai", "commerce-events"},
			want: "decoding commerce event response: expected an object with events"},
		{path: events, body: `{"events":null}`, args: []string{"ai", "commerce-events", "-o", "json"},
			want: "expected an object with events"},
		{path: events, body: `{"events":[1]}`, args: []string{"ai", "commerce-events"},
			want: "expected an object with events"},
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

// TestAIReadsOfNothing checks that a provider without models, and a tenant
// without commerce events, print only the header, and that -o json prints
// the response as sent.
func TestAIReadsOfNothing(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/ai-service/acme1/agentic/models", respond(http.StatusOK, `[{"provider":"google","models":[]}]`))
	fake.route("/ai-service/acme1/agentic/commerce-events", respond(http.StatusOK, `{"events":[]}`))
	for args, want := range map[string]string{
		"ai models":                  "ID  PROVIDER  THINKING  NAME\n",
		"ai commerce-events":         "EVENT\n",
		"ai commerce-events -o json": "{\n  \"events\": []\n}\n",
	} {
		code, stdout, stderr := runCLI(t, fake.env(), strings.Fields(args)...)
		if code != ExitOK || stdout != want {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %q", args, code, stdout, stderr, want)
		}
	}
}

// TestAIModelsJSONKeepsTheProviders checks that -o json prints the models
// response as sent, grouped by provider, not as the table's rows.
func TestAIModelsJSONKeepsTheProviders(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/ai-service/acme1/agentic/models", respond(http.StatusOK,
		`[{"provider":"openai","models":[{"id":"o3","name":"o3","description":"Reasoning"}]}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "ai", "models", "-o", "json")
	var providers []struct {
		Provider string `json:"provider"`
		Models   []struct {
			Description string `json:"description"`
		} `json:"models"`
	}
	if code != ExitOK || json.Unmarshal([]byte(stdout), &providers) != nil || len(providers) != 1 ||
		providers[0].Provider != "openai" || len(providers[0].Models) != 1 ||
		providers[0].Models[0].Description != "Reasoning" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

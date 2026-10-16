package cli

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/spf13/cobra"
)

// Name and paths of the AI Service.
const (
	aiUse       = "ai"
	aiPath      = "/ai-service/{tenant}/"
	agenticPath = aiPath + "agentic/" // the agents, what they are built from and what they did
)

// The scopes of the AI Service and the AI RAG Indexer, by what they carry out.
const (
	// aiScope reads agents, what they are built from and what they did, and
	// the AI RAG Indexer's metadata; it also exports agents.
	aiScope = "ai.agent_read"
	// aiManageScope writes agents and what they are built from, deletes jobs,
	// and starts a reindex of the AI RAG Indexer.
	aiManageScope = "ai.agent_manage"
	// aiChatScope chats with agents and attaches files to a chat (the
	// reference also names ai.agentexecution_manage_own, a customer's).
	aiChatScope = "ai.agentexecution_manage"
	// aiTextScope generates texts.
	aiTextScope = "ai.text_manage"
	// aiCompletionScope generates completions.
	aiCompletionScope = "ai.completion_manage"
)

// Column headers of what agents did (ai_activity.go, ai_analytics.go).
const (
	columnAgent    = "AGENT"   // the agent that ran, as `epx ai agents list` shows its ID
	columnSession  = "SESSION" // an agent's session, as `epx ai logs sessions list` shows its ID
	columnSeverity = "SEVERITY"
)

// llmConfig is the part of an agent's or template's LLM configuration shown
// in tables; its token, limits and temperature are only in -o json.
type llmConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// aiTemplate is the part of an agent template shown in tables. The reference
// documents neither its type nor its LLM configuration, but Emporix sends
// both, as for an agent.
type aiTemplate struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Enabled   *bool     `json:"enabled"`
	LLMConfig llmConfig `json:"llmConfig"`
	Name      localized `json:"name"`
}

// aiAgent is the part of an agent shown in tables: what a template holds and
// what starts the agent.
type aiAgent struct {
	aiTemplate
	Triggers []agentTrigger `json:"triggers"`
}

// agentTrigger is what starts an agent; tables show its type, -o json also
// its configuration.
type agentTrigger struct {
	Type string `json:"type"`
}

// llmModel is an LLM model of a provider; its description is only in -o json.
type llmModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Thinking *bool  `json:"thinking"`
}

// aiModel is an LLM model with its provider: an item of the models read
// (modelItems).
type aiModel struct {
	Provider string   `json:"provider"`
	Model    llmModel `json:"model"`
}

// newAICommand groups the commands of the AI Service: agent templates and
// agents, what they connect with (ai_connections.go), the LLM models they
// can use and the commerce events that can start them, what they did
// (ai_activity.go, ai_analytics.go), and the AI RAG Indexer's metadata and
// reindex (ai_rag.go).
func newAICommand(a *app) *cobra.Command {
	models := reader[aiModel]{view: view[aiModel]{
		singular: "LLM model",
		header:   []string{columnID, columnProvider, "THINKING", columnName},
		row: func(m aiModel, _ language) []string {
			return []string{m.Model.ID, m.Provider, formatBool(m.Model.Thinking), m.Model.Name}
		},
	}, scopes: []string{aiScope}}
	events := reader[string]{view: stringsView("commerce event", "EVENT"), scopes: []string{aiScope}}
	return newGroup(aiUse, "Read and change AI agents and their setup, chat with them, generate texts",
		"Reads the AI Service's agents and what they are built from: agent templates, tools, tokens,\n"+
			"OAuth configurations and MCP servers, the LLM models of the supported providers and the\n"+
			"commerce events that can trigger an agent; what agents did: their jobs, conversations,\n"+
			"request and session logs and analytics; and which fields the AI RAG Indexer embeds.\n"+
			"Upserts, changes and deletes agents and what they are built from, exports and imports\n"+
			"agents, chats with them and generates texts, which consumes AI credits. `epx ai rag\n"+
			"reindex` embeds the entities of a type again.",
		newAITemplatesCommand(a), newAIAgentsCommand(a),
		newAIToolsCommand(a), newAITokensCommand(a), newAIOAuthsCommand(a), newAIMCPServersCommand(a),
		newAIJobsCommand(a), newAIConversationsCommand(a), newAILogsCommand(a), newAIAnalyticsCommand(a),
		newAIRAGCommand(a),
		models.command(a, &command{
			use:   "models",
			short: "List the LLM models that agents can use",
			long: "Reads the LLM models of every provider that agents can use (openai, anthropic and google):\n" +
				"ID is what an agent's LLM configuration names and THINKING whether the model supports\n" +
				"thinking; descriptions are only in -o json.",
			req:   request{path: agenticPath + "models"},
			fetch: single{items: modelItems},
		}),
		events.command(a, &command{
			use:   "commerce-events",
			short: "List the commerce events that can trigger an agent",
			long:  "Reads the commerce events that can start an agent with a commerce_events trigger, sorted.",
			req:   request{path: agenticPath + "commerce-events"},
			fetch: single{items: commerceEventItems},
		}),
		agentChat("chat", "Send a message to an agent and print its reply", "and prints its reply: the messages\n"+
			"of the agent and of those it collaborates with, each with the sessionId that --session takes to\n"+
			"go on.\n"+replyTimeout()+", or use chat-async.").command(a),
		agentChat("chat-async", "Send a message to an agent, which replies in a job", "which replies in a job of type\n"+
			"AGENT_CHAT: Emporix answers with its jobId, which `epx ai jobs get JOB_ID` reads.").command(a),
		textGeneration().command(a), completionGeneration().command(a),
	)
}

// agentic returns r, a collection of the AI Service's agents and what they
// are built from or did, at agentic/<r.use> (aiCollection).
func agentic[T any](r resource[T]) resource[T] { return aiCollection(r, agenticPath) }

// aiCollection returns r, a collection of the AI Service at base + r.use in
// the ai group, which Emporix pages, filters and sorts.
func aiCollection[T any](r resource[T], base string) resource[T] {
	r.under, r.path, r.scopes, r.filters = aiUse, base+r.use, []string{aiScope}, listFilters{kinds: filterable}
	return r
}

// newAIAgentsCommand reads and writes AI agents, attaches files to chats with
// them and exports and imports them.
func newAIAgentsCommand(a *app) *cobra.Command {
	r := aiAgents()
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an agent", language: stringsOrMaps,
		members: "The JSON holds its name and description, userPrompt, llmConfig, triggers and mcpServers,\n" +
			"and its nativeTools, agentCollaborations, enabledMemory and maxRecursionLimit.\n",
		remove: "delete it even where other agents collaborate with it",
	}, agentWrites(a)...)
}

// aiAgents declares the AI agents, whose IDs the analytics of one agent
// complete with too.
func aiAgents() resource[aiAgent] {
	return agentic(resource[aiAgent]{
		use: "agents", noun: "AI agents", short: "Read, create, upsert, change, delete, export and import AI agents",
		long: "Reads AI agents: TYPE is generic, complaint or anti_fraud, PROVIDER and MODEL the LLM an\n" +
			"agent uses and TRIGGERS what starts it (endpoint, slack, teams or commerce_events). Their\n" +
			"prompts, tools and MCP servers are only in -o json.",
		view: view[aiAgent]{
			singular:  "agent",
			header:    []string{columnID, columnType, columnEnabled, columnProvider, columnModel, "TRIGGERS", columnName},
			localized: true,
			row: func(agent aiAgent, lang language) []string {
				return []string{
					agent.ID, agent.Type, formatBool(agent.Enabled), agent.LLMConfig.Provider, agent.LLMConfig.Model,
					joinEach(agent.Triggers, ", ", func(t agentTrigger) string { return t.Type }), agent.Name.in(lang),
				}
			},
		},
	})
}

// newAITemplatesCommand reads agent templates, which Emporix lists and
// searches but does not read one by one.
func newAITemplatesCommand(a *app) *cobra.Command {
	r := aiTemplates()
	return newGroup(r.use, r.summary(), r.long, r.listCommand(a), r.querySearchCommand(a))
}

// aiTemplates declares the agent templates, whose IDs an agent made from a
// template completes with.
func aiTemplates() resource[aiTemplate] {
	return agentic(resource[aiTemplate]{
		use: "templates", noun: "agent templates",
		long: "Reads agent templates, from which agents are made: TYPE is generic, complaint or anti_fraud,\n" +
			"PROVIDER and MODEL the LLM a template uses. Emporix reads no single template:\n" +
			"`epx ai templates list --query id:ID -o json` shows one.",
		view: view[aiTemplate]{
			singular:  "agent template",
			header:    []string{columnID, columnType, columnEnabled, columnProvider, columnModel, columnName},
			localized: true,
			row: func(t aiTemplate, lang language) []string {
				return []string{t.ID, t.Type, formatBool(t.Enabled), t.LLMConfig.Provider, t.LLMConfig.Model, t.Name.in(lang)}
			},
		},
	})
}

// modelItems returns the models of a models response, a list of providers
// each with its models, as one aiModel each.
func modelItems(body json.RawMessage) ([]json.RawMessage, error) {
	providers, err := listItems(body)
	if err != nil {
		return nil, err
	}
	var models []aiModel
	for _, raw := range providers {
		var provider struct {
			Provider string      `json:"provider"`
			Models   *[]llmModel `json:"models"`
		}
		if err := json.Unmarshal(raw, &provider); err != nil || provider.Models == nil {
			return nil, errors.New("expected providers with their models")
		}
		for _, model := range *provider.Models {
			models = append(models, aiModel{Provider: provider.Provider, Model: model})
		}
	}
	return marshalEach(models)
}

// commerceEventItems returns the events of a commerce events response, an
// object holding their names, sorted.
func commerceEventItems(body json.RawMessage) ([]json.RawMessage, error) {
	var answer struct {
		Events *[]string `json:"events"`
	}
	if err := json.Unmarshal(body, &answer); err != nil || answer.Events == nil {
		return nil, errors.New("expected an object with events")
	}
	return marshalEach(slices.Sorted(slices.Values(*answer.Events)))
}

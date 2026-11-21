package cli

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/spf13/cobra"
)

const (
	aiUse       = "ai"
	aiPath      = "/ai-service/{tenant}/"
	agenticPath = aiPath + "agentic/"
)

const (
	aiScope           = "ai.agent_read"
	aiManageScope     = "ai.agent_manage"
	aiChatScope       = "ai.agentexecution_manage"
	aiTextScope       = "ai.text_manage"
	aiCompletionScope = "ai.completion_manage"
)

const (
	columnAgent    = "AGENT"
	columnSession  = "SESSION"
	columnSeverity = "SEVERITY"
)

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

type aiAgent struct {
	aiTemplate
	Triggers []agentTrigger `json:"triggers"`
}

type agentTrigger struct {
	Type string `json:"type"`
}

type llmModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Thinking *bool  `json:"thinking"`
}

type aiModel struct {
	Provider string   `json:"provider"`
	Model    llmModel `json:"model"`
}

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
		"Reads the AI Service's agents and what they are built from: agent templates, tools, tokens, "+
			"OAuth configurations and MCP servers, the LLM models of the supported providers and the "+
			"commerce events that can trigger an agent; what agents did: their jobs, conversations, "+
			"request and session logs and analytics; and which fields the AI RAG Indexer embeds. "+
			"Upserts, changes and deletes agents and what they are built from, exports and imports "+
			"agents, chats with them and generates texts, which consumes AI credits. `epx ai rag "+
			"reindex` embeds the entities of a type again.",
		newAITemplatesCommand(a), newAIAgentsCommand(a),
		newAIToolsCommand(a), newAITokensCommand(a), newAIOAuthsCommand(a), newAIMCPServersCommand(a),
		newAIJobsCommand(a), newAIConversationsCommand(a), newAILogsCommand(a), newAIAnalyticsCommand(a),
		newAIRAGCommand(a),
		models.command(a, &command{
			use:   "models",
			short: "List the LLM models that agents can use",
			long: "Reads the LLM models of every provider that agents can use (openai, anthropic and google): " +
				"ID is what an agent's LLM configuration names and THINKING whether the model supports " +
				"thinking; descriptions are only in `-o json`.",
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
		agentReplyChat().command(a), agentJobChat().command(a),
		textGeneration().command(a), completionGeneration().command(a),
	)
}

func agentic[T any](r resource[T]) resource[T] { return aiCollection(r, agenticPath) }

func aiCollection[T any](r resource[T], base string) resource[T] {
	r.under, r.path, r.scopes, r.filters = aiUse, base+r.use, []string{aiScope}, listFilters{kinds: filterable}
	return r
}

// agentMixinsChecked says which mixins an agent write is refused for, as
// seen live: Emporix checks every mixin against a schema it downloads.
func agentMixinsChecked() string {
	return "Emporix refuses, with 400 Bad Request (exit 1), a mixin without a schema URL in metadata.mixins (" +
		literal(`"No matching schema found"`) + ", although its reference says that skips the check), one whose " +
		"schema it cannot download and a value that its schema does not allow."
}

func newAIAgentsCommand(a *app) *cobra.Command {
	r := aiAgents()
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an agent", language: stringsOrMaps,
		members: "The JSON holds its name and description, userPrompt, llmConfig, triggers and mcpServers, " +
			"and its nativeTools, agentCollaborations, enableMemory (as a read shows it: Emporix ignores " +
			"enabledMemory, the name the reference gives) and maxRecursionLimit (up to 100); the JSON that " +
			"`epx ai agents get ID -o json` prints can be sent back as it is. nativeTools may " +
			"hold file_generator, the built-in file generator, which is not a tool of `epx ai tools`: its ID " +
			"in nativeTools lets the agent create files and images. " + mixinsHelp + " " + agentMixinsChecked(),
		remove: "delete it even where other agents collaborate with it",
	}, agentWrites(a)...)
}

func aiAgents() resource[aiAgent] {
	return agentic(resource[aiAgent]{
		use: "agents", noun: "AI agents", short: "Read, create, upsert, change, delete, export and import AI agents",
		long: "Reads AI agents: TYPE is generic, complaint or anti_fraud, PROVIDER and MODEL the LLM an " +
			"agent uses and TRIGGERS what starts it (endpoint, slack, teams or commerce_events). Their " +
			"prompts, tools and MCP servers are only in `-o json`.",
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

func aiTemplates() resource[aiTemplate] {
	return agentic(resource[aiTemplate]{
		use: "templates", noun: "agent templates",
		long: "Reads agent templates, from which agents are made: TYPE is generic, complaint or anti_fraud, " +
			"PROVIDER and MODEL the LLM a template uses. Emporix reads no single template: " +
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

func commerceEventItems(body json.RawMessage) ([]json.RawMessage, error) {
	var answer struct {
		Events *[]string `json:"events"`
	}
	if err := json.Unmarshal(body, &answer); err != nil || answer.Events == nil {
		return nil, errors.New("expected an object with events")
	}
	return marshalEach(slices.Sorted(slices.Values(*answer.Events)))
}

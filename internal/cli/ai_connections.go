package cli

import "github.com/spf13/cobra"

// The AI Service's connections: what agents connect to (native tools and MCP
// servers) and the credentials they connect with (tokens and OAuth
// configurations).

// aiTool is the part of an agent's native tool shown in tables.
type aiTool struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Enabled *bool  `json:"enabled"`
	Name    string `json:"name"`
}

// aiToken is a token of the AI Service: its name and metadata, as Emporix
// sends it; the reference documents no value in a response.
type aiToken struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Metadata itemMetadata `json:"metadata"`
}

// tokenReference names a token (aiToken) by its ID.
type tokenReference struct {
	ID string `json:"id"`
}

// aiOAuth is the part of an OAuth configuration shown in tables; its scope
// is only in -o json.
type aiOAuth struct {
	ID                string         `json:"id"`
	Enabled           *bool          `json:"enabled"`
	ClientID          string         `json:"clientId"`
	ClientSecretToken tokenReference `json:"clientSecretToken"`
	URL               string         `json:"url"`
}

// aiMCPServer is the part of an MCP server shown in tables.
type aiMCPServer struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Enabled   *bool  `json:"enabled"`
	Transport string `json:"transport"`
	Name      string `json:"name"`
	Config    struct {
		URL string `json:"url"`
	} `json:"config"`
}

// newAIToolsCommand reads and writes the native tools of agents.
func newAIToolsCommand(a *app) *cobra.Command {
	r := agentic(resource[aiTool]{
		use: "tools", noun: "agent tools",
		long: "Reads the native tools that agents can use: TYPE is slack, teams, rag_custom (retrieval-\n" +
			"augmented generation over the tenant's own vector database) or rag_emporix (over Emporix\n" +
			"data); their configuration is only in -o json.",
		view: view[aiTool]{
			singular: "tool",
			header:   []string{columnID, columnType, columnEnabled, columnName},
			row: func(t aiTool, _ language) []string {
				return []string{t.ID, t.Type, formatBool(t.Enabled), t.Name}
			},
		},
	})
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an agent tool",
		members: "The JSON holds its type (slack, teams, rag_custom or rag_emporix), name and config, and\n" +
			"whether it is enabled.\n",
		disable: usedByAgents, remove: removeFromAgents,
	})
}

// newAITokensCommand reads the tokens of the AI Service by name, and writes
// them.
func newAITokensCommand(a *app) *cobra.Command {
	r := agentic(resource[aiToken]{
		use: "tokens", noun: "AI tokens",
		long: "Reads the tokens that hold the API keys and secrets of LLMs, tools, OAuth configurations and\n" +
			"MCP servers, by name: the table shows no value, and the API reference documents none in a\n" +
			"response; -o json prints the response as Emporix sends it.",
		view: view[aiToken]{
			singular: "token",
			header:   []string{columnID, columnName, columnCreated, columnModified},
			row: func(t aiToken, _ language) []string {
				return []string{t.ID, t.Name, t.Metadata.CreatedAt, t.Metadata.ModifiedAt}
			},
		},
	})
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an AI token",
		members: "The JSON holds its name and value, the API key or secret, which cannot change once the\n" +
			"token is created. A dry run prints the body, value included.\n",
		remove: "delete it even where agents or MCP servers use it",
	})
}

// newAIOAuthsCommand reads and writes OAuth configurations, with which agents
// get a token for a self-hosted LLM.
func newAIOAuthsCommand(a *app) *cobra.Command {
	r := agentic(resource[aiOAuth]{
		use: "oauths", noun: "OAuth configurations",
		long: "Reads OAuth configurations (a preview at Emporix): the OAuth 2.0 client credentials with which\n" +
			"agents get a token for a self-hosted LLM from TOKEN URL. SECRET TOKEN is the ID of the token\n" +
			"(`epx ai tokens list`) that holds the client's secret; the scope is only in -o json, which\n" +
			"prints the response as Emporix sends it.",
		view: view[aiOAuth]{
			singular: "OAuth configuration",
			header:   []string{columnID, columnEnabled, "CLIENT ID", "SECRET TOKEN", "TOKEN URL"},
			row: func(o aiOAuth, _ language) []string {
				return []string{o.ID, formatBool(o.Enabled), o.ClientID, o.ClientSecretToken.ID, o.URL}
			},
		},
	})
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an OAuth configuration",
		members: "The JSON holds its url (the token URL), clientId and grantType (client_credentials, the\n" +
			"only one), and its scope, clientSecretToken (the token that holds the secret) and whether it\n" +
			"is enabled.\n",
		disable: usedByAgents, remove: removeFromAgents,
	})
}

// newAIMCPServersCommand reads and writes the tenant's MCP servers.
func newAIMCPServersCommand(a *app) *cobra.Command {
	r := agentic(resource[aiMCPServer]{
		use: "mcp-servers", noun: "MCP servers",
		long: "Reads the tenant's MCP servers, whose tools agents can use: TYPE is custom (a server at URL)\n" +
			"or dynamic (a preview at Emporix: tools that the server's configuration defines, only in\n" +
			"-o json), TRANSPORT streamable_http or sse.",
		view: view[aiMCPServer]{
			singular: "MCP server",
			header:   []string{columnID, columnType, columnEnabled, "TRANSPORT", columnName, columnURL},
			row: func(s aiMCPServer, _ language) []string {
				return []string{s.ID, s.Type, formatBool(s.Enabled), s.Transport, s.Name, s.Config.URL}
			},
		},
	})
	return aiManagedGroup(a, r, aiItemWrites{
		one: "an MCP server",
		members: "The JSON holds its type (custom: config with the server's url; dynamic: its tools), name\n" +
			"and transport, and whether it is enabled.\n",
		disable: usedByAgents, remove: removeFromAgents,
	})
}

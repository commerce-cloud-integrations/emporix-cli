package cli

import (
	"encoding/json"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

type aiJob struct {
	ID            string       `json:"id"`
	Type          string       `json:"type"`
	Status        string       `json:"status"`
	AgentID       string       `json:"agentId"`
	SessionID     string       `json:"sessionId"`
	CommerceEvent string       `json:"commerceEvent"`
	Metadata      itemMetadata `json:"metadata"`
}

type aiConversation struct {
	ID            string `json:"conversationId"`
	Name          string `json:"conversationName"`
	AgentID       string `json:"agentId"`
	SessionID     string `json:"sessionId"`
	LastMessageAt string `json:"lastMessageAt"`
}

type aiLog struct {
	ID             string             `json:"id"`
	RequestID      string             `json:"requestId"`
	SessionID      string             `json:"sessionId"`
	TriggerAgentID string             `json:"triggerAgentId"`
	Severity       string             `json:"severity"`
	Messages       *[]json.RawMessage `json:"messages"`
	Duration       json.Number        `json:"duration"`
	Metadata       itemMetadata       `json:"metadata"`
}

func (l *aiLog) messageCount() string {
	if l.Messages == nil {
		return ""
	}
	return strconv.Itoa(len(*l.Messages))
}

func aiLogView(singular string, ids func(aiLog) []string, idColumns ...string) view[aiLog] {
	return view[aiLog]{
		singular: singular,
		header:   slices.Concat(idColumns, []string{columnAgent, columnSeverity, "MESSAGES", "SECONDS", columnCreated}),
		row: func(l aiLog, _ language) []string {
			return slices.Concat(ids(l),
				[]string{l.TriggerAgentID, l.Severity, l.messageCount(), l.Duration.String(), l.Metadata.CreatedAt})
		},
	}
}

const aiLogsUse = "logs"

const logsLegend = "AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many " +
	"messages were logged (empty when a log comes without them) and SECONDS how long it took; the " +
	"messages, the other agents that took part and the LLM tokens used are only in `-o json`."

func newAIJobsCommand(a *app) *cobra.Command {
	r := aiCollection(resource[aiJob]{
		use: "jobs", noun: "agent jobs", short: "Read and delete agent jobs",
		long: "Reads the jobs of the AI Service: TYPE is AGENT_CHAT (an agent's run), IMPORT or EXPORT (of " +
			"agents), STATUS in_progress, success, failure or skipped, AGENT and SESSION the agent that " +
			"ran and its session (`epx ai logs sessions get`) and EVENT the commerce event that started " +
			"it. The message that started an agent's run (message), the agent's reply (response), the " +
			"provenance of a successful run (" + provenanceMember + "), and an import's or export's result are only " +
			"in `-o json`.",
		view: view[aiJob]{
			singular: "agent job",
			header:   []string{columnID, columnType, columnStatus, columnAgent, columnSession, "EVENT", columnCreated},
			row: func(j aiJob, _ language) []string {
				return []string{j.ID, j.Type, j.Status, j.AgentID, j.SessionID, j.CommerceEvent, j.Metadata.CreatedAt}
			},
		},
	}, aiPath)
	return newResourceCommand(a, r, r.querySearchCommand(a), aiJobDelete(r).command(a))
}

// newAIConversationsCommand reads the conversations of agents, which Emporix
// lists and searches but does not read one by one.
func newAIConversationsCommand(a *app) *cobra.Command {
	r := agentic(resource[aiConversation]{
		use: "conversations", noun: "agent conversations",
		long: "Reads the conversations of agents in Microsoft Teams channels and group chats: ID is the " +
			"conversation in Teams, NAME its channel or chat, SESSION the agent's session (`epx ai logs " +
			"sessions get`) and LAST MESSAGE when it was last active. The channel and the context " +
			"reference are only in `-o json`.",
		view: view[aiConversation]{
			singular: "agent conversation",
			header:   []string{columnID, columnName, columnAgent, columnSession, "LAST MESSAGE"},
			row: func(c aiConversation, _ language) []string {
				return []string{c.ID, c.Name, c.AgentID, c.SessionID, c.LastMessageAt}
			},
		},
	})
	return newGroup(r.use, r.summary(), r.long, r.listCommand(a), r.querySearchCommand(a))
}

func newAILogsCommand(a *app) *cobra.Command {
	requests := aiLogs(resource[aiLog]{
		use: "requests", noun: "agent request logs",
		long: "Reads the logs of the requests to agents, one per request: ID is the log that get takes, " +
			"REQUEST the request it logs and SESSION the session it belongs to. " +
			logsLegend,
		view: aiLogView("agent request log", func(l aiLog) []string { return []string{l.ID, l.RequestID, l.SessionID} },
			columnID, "REQUEST", columnSession),
	})
	sessions := aiLogs(resource[aiLog]{
		use: "sessions", noun: "agent session logs",
		long: "Reads the logs of the sessions of agents, one per session, with the messages of each of its " +
			"requests. " + logsLegend,
		view: aiLogView("agent session log", func(l aiLog) []string { return []string{l.SessionID} }, columnID),
	})
	return newGroup(aiLogsUse, "Read the logs of agent requests and sessions",
		"Reads the logs that agents write: one per request, and one per session, which holds the "+
			"messages of its requests.",
		newResourceCommand(a, requests, requests.querySearchCommand(a)),
		newResourceCommand(a, sessions, sessions.querySearchCommand(a)))
}

func aiLogs(r resource[aiLog]) resource[aiLog] {
	r = aiCollection(r, agenticPath+aiLogsUse+"/")
	r.under += " " + aiLogsUse
	return r
}

package cli

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/spf13/cobra"
)

// The analytics of the AI Service's agents: what Emporix counts of their
// requests and sessions.

// logMetrics is the part of the request or session totals of agent
// analytics shown in tables. Each count is kept as sent, and one that
// Emporix leaves out is omitted from the item (logTotals), so that it stays
// empty instead of becoming 0.
type logMetrics struct {
	Total      json.Number `json:"total,omitempty"`
	BySeverity struct {
		Warning json.Number `json:"warning,omitempty"`
		Error   json.Number `json:"error,omitempty"`
	} `json:"bySeverity"`
	ErrorRate json.Number `json:"errorRate,omitempty"`
}

// logTotals is a row of agent analytics: the totals of requests or of
// sessions (analyticsItems).
type logTotals struct {
	Of      string     `json:"of"`
	Metrics logMetrics `json:"metrics"`
}

// agentExecutions is how often an agent ran in one period: an item of the
// executions read (executionItems).
type agentExecutions struct {
	AgentID string          `json:"agentId"`
	Period  analyticsPeriod `json:"period"`
	Count   json.Number     `json:"count,omitempty"` // kept as sent; omitted, like a logMetrics count, when null
}

// analyticsPeriod is a period of agent executions, from start (inclusive)
// to end (exclusive).
type analyticsPeriod struct {
	Label string `json:"label"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// granularityParam is the query parameter, and the flag, that sets the
// calendar period of agent executions.
const granularityParam = "granularity"

// newAIAnalyticsCommand reads the analytics of agents: request and session
// totals, of every agent or of one, and executions per period.
func newAIAnalyticsCommand(a *app) *cobra.Command {
	totals := reader[logTotals]{view: view[logTotals]{
		singular: "agent analytics",
		header:   []string{"LOGS", columnTotal, "WARNING", "ERROR", "ERROR RATE"},
		legend: "LOGS is requests or sessions, TOTAL how many there were in the last four weeks, WARNING and\n" +
			"ERROR how many of them had that severity and ERROR RATE their share with errors, as Emporix\n" +
			"computes it. The requests per session and the weekly trends are only in -o json.",
		row: func(t logTotals, _ language) []string {
			m := t.Metrics
			return []string{t.Of, m.Total.String(), m.BySeverity.Warning.String(), m.BySeverity.Error.String(),
				m.ErrorRate.String()}
		},
	}, scopes: []string{aiScope}}
	executions := reader[agentExecutions]{view: view[agentExecutions]{
		singular: "agent executions",
		header:   []string{columnAgent, "PERIOD", columnFrom, columnTo, "EXECUTIONS"},
		row: func(e agentExecutions, _ language) []string {
			return []string{e.AgentID, e.Period.Label, e.Period.Start, e.Period.End, e.Count.String()}
		},
	}, scopes: []string{aiScope}}
	const path = agenticPath + "analytics"
	return newGroup("analytics", "Read the analytics of agents",
		"Reads what Emporix counts of the requests and sessions of agents, as its dashboards show them.",
		totals.command(a, &command{
			use:   "tenant",
			short: "Show the request and session totals of every agent",
			long:  "Reads the totals of the requests and sessions of every agent of the tenant.",
			req:   request{path: path},
			fetch: single{items: analyticsItems},
		}),
		totals.command(a, &command{
			use:   "agent AGENT_ID",
			short: "Show the request and session totals of one agent",
			long: "Reads the totals of the requests and sessions of one agent: AGENT_ID is the ID that\n" +
				"`epx ai agents list` shows; an unknown one counts nothing.",
			req:   request{path: path, query: url.Values{"agentId": {placeholderID}}},
			fetch: single{items: analyticsItems},
			ids:   argumentIDs{"AGENT_ID": aiAgents().ids()},
		}),
		executions.command(a, &command{
			use:   "executions AGENT_IDS",
			short: "List how often agents ran per calendar quarter, month or week",
			long: "Reads how often agents ran (their sessions) per calendar quarter, Emporix's default\n" +
				"granularity, or per month or week (--granularity): each PERIOD that Emporix reports runs\n" +
				"from FROM (inclusive) to TO (exclusive), in UTC. AGENT_IDS are the IDs that\n" +
				"`epx ai agents list` shows, separated by commas without spaces (at most 100); an unknown\n" +
				"one ran 0 times.",
			req: request{path: path + "/executions", query: url.Values{"agentIds": {placeholderID}}},
			fetch: &paramRead{
				params: queryFlags{optional: []paramFlag{{
					param: granularityParam, flag: granularityParam, choices: []string{"QUARTER", "MONTH", "WEEK"},
					usage: "the calendar `period` of each count",
				}}},
				fetch: single{items: executionItems},
			},
			ids: argumentIDs{"AGENT_IDS": aiAgents().ids().severalIDs()},
		}),
	)
}

// analyticsItems returns the rows of an analytics response: the totals of
// its requests, then of its sessions (logTotals).
func analyticsItems(body json.RawMessage) ([]json.RawMessage, error) {
	var bundle struct {
		Requests *logMetrics `json:"requests"`
		Sessions *logMetrics `json:"sessions"`
	}
	if err := json.Unmarshal(body, &bundle); err != nil || bundle.Requests == nil || bundle.Sessions == nil {
		return nil, errors.New("expected an object with requests and sessions")
	}
	return marshalEach([]logTotals{
		{Of: "requests", Metrics: *bundle.Requests},
		{Of: "sessions", Metrics: *bundle.Sessions},
	})
}

// executionItems returns the items of an executions response, which holds
// the periods and, per agent, a count for each: one agentExecutions per
// agent and period, in their order.
func executionItems(body json.RawMessage) ([]json.RawMessage, error) {
	var answer struct {
		Periods *[]analyticsPeriod `json:"periods"`
		Series  *[]struct {
			AgentID string        `json:"agentId"`
			Counts  []json.Number `json:"counts"`
		} `json:"series"`
	}
	if err := json.Unmarshal(body, &answer); err != nil || answer.Periods == nil || answer.Series == nil {
		return nil, errors.New("expected an object with periods and series")
	}
	periods := *answer.Periods
	var rows []agentExecutions // grown as counts are checked, not sized from the response up front
	for _, series := range *answer.Series {
		if len(series.Counts) != len(periods) {
			return nil, errors.New("expected one count per period for each agent")
		}
		for i, count := range series.Counts {
			rows = append(rows, agentExecutions{AgentID: series.AgentID, Period: periods[i], Count: count})
		}
	}
	return marshalEach(rows)
}

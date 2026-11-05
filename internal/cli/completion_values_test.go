package cli

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
)

// The routes and first pages of the lists that TYPE of the RAG reads and of
// search documents, and AGENT_IDS of agent executions, complete from.
const (
	customEntitiesPath = "/schema/acme1/custom-entities"
	customEntitiesPage = customEntitiesPath + "?pageNumber=1&pageSize=100"
	agentsPath         = aiAgenticPath + "agents"
	agentsPage         = agentsPath + "?pageNumber=1&pageSize=100"
)

func TestCompletionOffersTheBuiltInRAGTypesAndTheCustomOnes(t *testing.T) {
	custom := map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusOK,
		`[{"id":"CUSTOM_DOCUMENT","name":{"en":"Document"}},{"id":"ORDER"},{"id":"OFFER"}]`)}
	many := map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusOK,
		"["+numbered(99, func(i int) string { return fmt.Sprintf(`{"id":"T%03d"}`, i) })+"]")}
	for _, tt := range []completionCase{
		{ // the built-in types first, then the custom ones, with the scopes of `custom-entities list` (none)
			args: []string{"ai", "rag", "fields", ""}, routes: custom,
			want:     append(entityBuiltInTypeNames(), "CUSTOM_DOCUMENT|Document", "OFFER"), // ORDER once
			requests: []string{customEntitiesPage},
		},
		{ // only those that start with what is typed
			args: []string{"ai", "rag", "filter-fields", "O"}, routes: custom,
			want:     []string{"ORDER", "OFFER"},
			requests: []string{customEntitiesPage},
		},
		{ // a write: the read scopes of the list, never its write scope
			args: []string{"ai", "rag", "reindex", "C"}, routes: custom,
			want:     []string{"CUSTOMER", "CUSTOM_DOCUMENT|Document"},
			requests: []string{customEntitiesPage},
		},
		{ // the built-in types where the list cannot be read
			args:     []string{"ai", "rag", "fields", ""},
			routes:   map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusInternalServerError, `{}`)},
			want:     entityBuiltInTypeNames(),
			requests: []string{customEntitiesPage},
		},
		{ // up to 100 in all: the built-in types, then the first 96 custom ones
			args: []string{"ai", "rag", "fields", ""}, routes: many,
			want:     append(entityBuiltInTypeNames(), keys("T", 96)...),
			requests: []string{customEntitiesPage},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionOffersTheBuiltInSearchTypesAndTheCustomOnes(t *testing.T) {
	custom := map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusOK,
		`[{"id":"VEHICLE","name":{"en":"Vehicle"}},{"id":"LEAFLET"}]`)}
	for _, tt := range []completionCase{
		{ // the built-in types first, then the custom ones, with the scopes of `custom-entities list` (none)
			args: []string{"search", "documents", "--saved-search", "s1", "--text", "x", ""}, routes: custom,
			want:     []string{"customer", "legal-entity", "VEHICLE|Vehicle", "LEAFLET"},
			requests: []string{customEntitiesPage},
		},
		{ // only those that start with what is typed
			args: []string{"search", "documents", "l"}, routes: custom,
			want:     []string{"legal-entity"},
			requests: []string{customEntitiesPage},
		},
		{ // in the case typed
			args: []string{"search", "documents", "L"}, routes: custom,
			want:     []string{"LEAFLET"},
			requests: []string{customEntitiesPage},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionOfSeveralAgentIDsCompletesTheLast(t *testing.T) {
	agents := map[string]http.HandlerFunc{agentsPath: respond(http.StatusOK,
		`[{"id":"a1","name":{"en":"Support"}},{"id":"a2"},{"id":"b1","name":{"en":"Fraud"}},{"id":"b,2"}]`)}
	for _, tt := range []completionCase{
		{ // the first: every agent, except an ID that holds a comma, which the argument could not take
			args: []string{"ai", "analytics", "executions", ""}, routes: agents,
			want:     []string{"a1|Support", "a2", "b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{ // after a comma: those not typed yet, after the IDs typed before
			args: []string{"ai", "analytics", "executions", "a1,"}, routes: agents,
			want:     []string{"a1,a2", "a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{ // the last ID, as far as it is typed
			args: []string{"ai", "analytics", "executions", "a2,a1,b"}, routes: agents,
			want:     []string{"a2,a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{ // an ID typed whole is still offered, to be confirmed
			args: []string{"ai", "analytics", "executions", "a1,a2"}, routes: agents,
			want:     []string{"a1,a2"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
	} {
		// Shells add no space after an ID, so that a comma and another ID may follow.
		tt.directive = cobra.ShellCompDirectiveNoSpace
		tt.check(t)
	}
}

func TestSeveralIDsOfNoListAreNone(t *testing.T) {
	if src := (*idSource)(nil).severalIDs(); src != nil {
		t.Errorf("several IDs of no list: %+v; want none", src)
	}
}

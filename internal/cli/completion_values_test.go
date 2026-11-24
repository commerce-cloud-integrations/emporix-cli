package cli

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
)

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
		{
			name: "the built-in types first, then the custom ones, each once",
			args: []string{"ai", "rag", "fields", ""}, routes: custom,
			want:     append(entityBuiltInTypeNames(), "CUSTOM_DOCUMENT|Document", "OFFER"),
			requests: []string{customEntitiesPage},
		},
		{
			name: "only those that start with what is typed",
			args: []string{"ai", "rag", "filter-fields", "O"}, routes: custom,
			want:     []string{"ORDER", "OFFER"},
			requests: []string{customEntitiesPage},
		},
		{
			name: "a write: the read scopes of the list, never its write scope",
			args: []string{"ai", "rag", "reindex", "C"}, routes: custom,
			want:     []string{"CUSTOMER", "CUSTOM_DOCUMENT|Document"},
			requests: []string{customEntitiesPage},
		},
		{
			name:     "the built-in types where the list cannot be read",
			args:     []string{"ai", "rag", "fields", ""},
			routes:   map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusInternalServerError, `{}`)},
			want:     entityBuiltInTypeNames(),
			requests: []string{customEntitiesPage},
		},
		{
			name: "up to 100 in all: the built-in types, then the first 96 custom ones",
			args: []string{"ai", "rag", "fields", ""}, routes: many,
			want:     append(entityBuiltInTypeNames(), keys("T", 96)...),
			requests: []string{customEntitiesPage},
		},
	} {
		t.Run(tt.name, tt.check)
	}
}

func TestCompletionOffersTheBuiltInSearchTypesAndTheCustomOnes(t *testing.T) {
	custom := map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusOK,
		`[{"id":"VEHICLE","name":{"en":"Vehicle"}},{"id":"LEAFLET"}]`)}
	for _, tt := range []completionCase{
		{
			name: "the built-in types first, then the custom ones",
			args: []string{"search", "documents", "--saved-search", "s1", "--text", "x", ""}, routes: custom,
			want:     []string{"customer", "legal-entity", "VEHICLE|Vehicle", "LEAFLET"},
			requests: []string{customEntitiesPage},
		},
		{
			name: "only those that start with what is typed",
			args: []string{"search", "documents", "l"}, routes: custom,
			want:     []string{"legal-entity"},
			requests: []string{customEntitiesPage},
		},
		{
			name: "in the case typed",
			args: []string{"search", "documents", "L"}, routes: custom,
			want:     []string{"LEAFLET"},
			requests: []string{customEntitiesPage},
		},
	} {
		t.Run(tt.name, tt.check)
	}
}

func TestCompletionOffersTheSavedSearchesOfTheType(t *testing.T) {
	queries := "/search/acme1/search/VEHICLE/queries"
	saved := map[string]http.HandlerFunc{queries: respond(http.StatusOK,
		`[{"id":"by-name","name":{"en":"By name"}},{"id":"by-year"},{"id":"recent"}]`)}
	for _, tt := range []completionCase{
		{
			name: "every saved search of the TYPE typed",
			args: []string{"search", "documents", "VEHICLE", "--saved-search", ""}, routes: saved,
			want:     []string{"by-name|By name", "by-year", "recent"},
			requests: []string{queries + "?pageNumber=1&pageSize=100"}, scopes: []string{searchScope},
		},
		{
			name: "only those that start with what is typed",
			args: []string{"search", "documents", "VEHICLE", "--saved-search", "by"}, routes: saved,
			want:     []string{"by-name|By name", "by-year"},
			requests: []string{queries + "?pageNumber=1&pageSize=100"}, scopes: []string{searchScope},
		},
	} {
		t.Run(tt.name, tt.check)
	}
	f := newFakeEmporix(t)
	if got, _ := completion(t, completionEnv(f), "search", "documents", "--saved-search", ""); len(got) != 0 {
		t.Errorf("--saved-search before a TYPE: completions %q; want none", got)
	}
	f.assertRequests(t)
}

func TestCompletionOfSeveralAgentIDsCompletesTheLast(t *testing.T) {
	agents := map[string]http.HandlerFunc{agentsPath: respond(http.StatusOK,
		`[{"id":"a1","name":{"en":"Support"}},{"id":"a2"},{"id":"b1","name":{"en":"Fraud"}},{"id":"b,2"}]`)}
	for _, tt := range []completionCase{
		{
			name: "the first: every agent, except an ID that holds a comma",
			args: []string{"ai", "analytics", "executions", ""}, routes: agents,
			want:     []string{"a1|Support", "a2", "b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			name: "after a comma: those not typed yet, after the IDs typed before",
			args: []string{"ai", "analytics", "executions", "a1,"}, routes: agents,
			want:     []string{"a1,a2", "a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			name: "the last ID, as far as it is typed",
			args: []string{"ai", "analytics", "executions", "a2,a1,b"}, routes: agents,
			want:     []string{"a2,a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			name: "an ID typed whole, to be confirmed",
			args: []string{"ai", "analytics", "executions", "a1,a2"}, routes: agents,
			want:     []string{"a1,a2"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
	} {
		tt.directive = cobra.ShellCompDirectiveNoSpace
		t.Run(tt.name, tt.check)
	}
}

func TestSeveralIDsOfNoListAreNone(t *testing.T) {
	if src := (*idSource)(nil).severalIDs(); src != nil {
		t.Errorf("several IDs of no list: %+v; want none", src)
	}
}

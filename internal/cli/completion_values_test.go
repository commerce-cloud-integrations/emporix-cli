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
			args: []string{"ai", "rag", "fields", ""}, routes: custom,
			want:     append(entityBuiltInTypeNames(), "CUSTOM_DOCUMENT|Document", "OFFER"),
			requests: []string{customEntitiesPage},
		},
		{
			args: []string{"ai", "rag", "filter-fields", "O"}, routes: custom,
			want:     []string{"ORDER", "OFFER"},
			requests: []string{customEntitiesPage},
		},
		{
			args: []string{"ai", "rag", "reindex", "C"}, routes: custom,
			want:     []string{"CUSTOMER", "CUSTOM_DOCUMENT|Document"},
			requests: []string{customEntitiesPage},
		},
		{
			args:     []string{"ai", "rag", "fields", ""},
			routes:   map[string]http.HandlerFunc{customEntitiesPath: respond(http.StatusInternalServerError, `{}`)},
			want:     entityBuiltInTypeNames(),
			requests: []string{customEntitiesPage},
		},
		{
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
		{
			args: []string{"search", "documents", "--saved-search", "s1", "--text", "x", ""}, routes: custom,
			want:     []string{"customer", "legal-entity", "VEHICLE|Vehicle", "LEAFLET"},
			requests: []string{customEntitiesPage},
		},
		{
			args: []string{"search", "documents", "l"}, routes: custom,
			want:     []string{"legal-entity"},
			requests: []string{customEntitiesPage},
		},
		{
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
		{
			args: []string{"ai", "analytics", "executions", ""}, routes: agents,
			want:     []string{"a1|Support", "a2", "b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			args: []string{"ai", "analytics", "executions", "a1,"}, routes: agents,
			want:     []string{"a1,a2", "a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			args: []string{"ai", "analytics", "executions", "a2,a1,b"}, routes: agents,
			want:     []string{"a2,a1,b1|Fraud"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
		{
			args: []string{"ai", "analytics", "executions", "a1,a2"}, routes: agents,
			want:     []string{"a1,a2"},
			requests: []string{agentsPage}, scopes: []string{aiScope},
		},
	} {
		tt.directive = cobra.ShellCompDirectiveNoSpace
		tt.check(t)
	}
}

func TestSeveralIDsOfNoListAreNone(t *testing.T) {
	if src := (*idSource)(nil).severalIDs(); src != nil {
		t.Errorf("several IDs of no list: %+v; want none", src)
	}
}

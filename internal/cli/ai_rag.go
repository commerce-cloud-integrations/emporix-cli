package cli

import (
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The metadata of the AI RAG Indexer, which embeds entities for the RAG
// tools of agents.

// ragFilterField is a field that a vector search can filter by; its
// deprecated name and description are only in -o json.
type ragFilterField struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

// newAIRAGCommand reads the metadata of the AI RAG Indexer, the fields of an
// entity type it can embed and those a vector search can filter by, and
// embeds the entities of a type again.
func newAIRAGCommand(a *app) *cobra.Command {
	const path = "/ai-rag-indexer/{tenant}/" + placeholderID // the metadata of the entity type TYPE
	// The predefined entity types, besides the custom ones, which the RAG reference names.
	predefined := []string{"PRODUCT", "ORDER"}
	typeList := "\nTYPE is " + alternatives(append(slices.Clone(predefined),
		"a custom entity type (`epx custom-entities list`)")) + "."
	reindexSentence := typeList + " Emporix answers an unknown TYPE with 204 No Content (exit 0),\n" +
		"so check the TYPE first."
	types := argumentIDs{columnType: customEntities().ids()}
	typeValues := map[string][]string{columnType: predefined}
	fieldsSentence := typeList + " The fields come whole, so they are read with one\nrequest."
	whole := listing{read: (*emporix.Client).GetUnpaged}
	fields := reader[string]{view: stringsView("RAG field", "FIELD"), scopes: []string{aiScope}}
	filters := reader[ragFilterField]{view: view[ragFilterField]{
		singular: "RAG filter field",
		header:   []string{"FIELD", columnType},
		row:      func(f ragFilterField, _ language) []string { return []string{f.Key, f.Type} },
	}, scopes: []string{aiScope}}
	reindex := &write{
		use:   "reindex TYPE",
		short: "Embed every entity of a type again for RAG search (deprecated by Emporix)",
		long: "Embeds every entity of a type again in the vector database, replacing its records there;\n" +
			"it sends no body. Emporix has deprecated this endpoint: `epx indexing reindex-jobs create`\n" +
			"with --rag starts a reindex job that embeds products too, which its --wait follows." + reindexSentence,
		example: "  epx ai rag reindex PRODUCT --dry-run",
		method:  http.MethodPost,
		req:     request{path: path + "/reindex"},
		scopes:  []string{aiManageScope},
		body:    noBody{},
		ids:     types, predefined: typeValues,
	}
	return newGroup("rag", "Read which fields the AI RAG Indexer embeds and filters by; reindex",
		"Reads the AI RAG Indexer, which embeds the entities of a type in a vector database for the\n"+
			"retrieval-augmented generation (RAG) tools of agents: the fields it can embed and those a\n"+
			"vector search can filter by; and embeds the entities of a type again.",
		reindex.command(a),
		fields.command(a, &command{
			use:   "fields TYPE",
			short: "List the fields of an entity type that can be embedded for RAG search",
			long:  "Reads the fields of an entity type, by path, that the AI RAG Indexer can embed." + fieldsSentence,
			req:   request{path: path + "/rag-metadata"},
			fetch: whole,
			ids:   types, predefined: typeValues,
		}),
		filters.command(a, &command{
			use:   "filter-fields TYPE",
			short: "List the fields of an entity type that a vector search can filter by",
			long: "Reads the fields of an entity type, by path, that a vector search can filter by before RAG\n" +
				"retrieval, with their TYPE (string, integer, float, boolean, datetime, date, time,\n" +
				"dictionary, list or object)." + fieldsSentence,
			req:   request{path: path + "/filter-metadata"},
			fetch: whole,
			ids:   types, predefined: typeValues,
		}),
	)
}

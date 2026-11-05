package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The metadata of the AI RAG Indexer, which embeds entities for the RAG
// tools of agents.

// ragFilterField is a field that a vector search can filter by, and the
// path of the list of objects it sits in, if any; its deprecated name and
// description are only in -o json.
type ragFilterField struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	Parent string `json:"parent"`
}

// newAIRAGCommand reads the metadata of the AI RAG Indexer, the fields of an
// entity type it can embed and those a vector search can filter by, and
// embeds the entities of a type again.
func newAIRAGCommand(a *app) *cobra.Command {
	const path = "/ai-rag-indexer/{tenant}/" + placeholderID // the metadata of the entity type TYPE
	// The entity types that the RAG reference names besides the custom ones, the same as reindex jobs
	// take, in any case (live: product, order, customer and legal-entity are served too).
	builtIn := entityBuiltInTypes()
	types := argumentIDs{columnType: customEntities().ids()}
	fieldsTypeHelp := "\n" + entityTypeSentences(builtIn) +
		"\nAn unknown TYPE, such as LEGAL_ENTITY, gets 400 Bad Request (exit 1).\n" +
		"The fields come whole, so they are read with one request."
	reindexTypeHelp := "\n" + entityTypeSentences(builtIn) +
		"\nEmporix has answered an unknown TYPE with 204 No Content (exit 0), not the\n" +
		"400 Bad Request its reference gives, so check the TYPE first."
	whole := listing{read: (*emporix.Client).GetUnpaged}
	fields := reader[string]{view: stringsView("RAG field", "FIELD"), scopes: []string{aiScope}}
	filters := reader[ragFilterField]{view: view[ragFilterField]{
		singular: "RAG filter field",
		header:   []string{"FIELD", columnType, columnParent},
		long:     []string{columnParent},
		row:      func(f ragFilterField, _ language) []string { return []string{f.Key, f.Type, f.Parent} },
	}, scopes: []string{aiScope}}
	reindex := &write{
		use:   "reindex TYPE",
		short: "Embed every entity of a type again for RAG search (deprecated by Emporix)",
		long: "Embeds every entity of a type again in the vector database, replacing its records there;\n" +
			"it sends no body. Emporix has deprecated this endpoint in favor of a reindex job:\n" +
			"`epx indexing reindex-jobs create`, which --wait follows. Give it --rag for PRODUCT; for\n" +
			"CUSTOMER and LEGAL-ENTITY, the job updates their embeddings without it." +
			reindexTypeHelp,
		example: "  epx ai rag reindex PRODUCT --dry-run",
		method:  http.MethodPost,
		req:     request{path: path + "/reindex"},
		scopes:  []string{aiManageScope},
		body:    noBody{},
		ids:     types, builtIn: builtIn,
	}
	return newGroup("rag", "Read which fields the AI RAG Indexer embeds and filters by; reindex",
		"Reads the AI RAG Indexer, which embeds the entities of a type in a vector database for the\n"+
			"retrieval-augmented generation (RAG) tools of agents: the fields it can embed and those a\n"+
			"vector search can filter by; and embeds the entities of a type again.",
		reindex.command(a),
		fields.command(a, &command{
			use:   "fields TYPE",
			short: "List the fields of an entity type that can be embedded for RAG search",
			long:  "Reads the fields of an entity type, by path, that the AI RAG Indexer can embed." + fieldsTypeHelp,
			req:   request{path: path + "/rag-metadata"},
			fetch: whole,
			ids:   types, builtIn: builtIn,
		}),
		filters.command(a, &command{
			use:   "filter-fields TYPE",
			short: "List the fields of an entity type that a vector search can filter by",
			long: "Reads the fields of an entity type, by path, that a vector search can filter by before RAG\n" +
				"retrieval, with their data type in the TYPE column (string, integer, float, boolean,\n" +
				"datetime, date, time, dictionary, list or object). PARENT is the path of the list of objects\n" +
				"a field sits in, such as addresses, and empty outside a list; tables cut a long PARENT.\n" +
				"A schema's REFERENCE attribute comes as two fields, its path (object) and that path with .id\n" +
				"(string), the one to filter by." + fieldsTypeHelp,
			req:   request{path: path + "/filter-metadata"},
			fetch: whole,
			ids:   types, builtIn: builtIn,
		}),
	)
}

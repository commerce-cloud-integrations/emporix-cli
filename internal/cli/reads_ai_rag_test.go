package cli

import (
	"strings"
	"testing"
)

// aiRAGReads returns the reads of the AI RAG Indexer's metadata (ai_rag.go).
func aiRAGReads() []readCase {
	const scope = "ai.agent_read"
	return []readCase{
		{
			args:   []string{"ai", "rag", "fields", "PRODUCT"},
			path:   "/ai-rag-indexer/acme1/PRODUCT/rag-metadata",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `["code","name","name.en","mixins.productCustomAttributes.color"]`,
			want:   "FIELD\ncode\nname\nname.en\nmixins.productCustomAttributes.color\n",
		},
		{
			// The deprecated name and description are only in -o json.
			args:   []string{"ai", "rag", "filter-fields", "ORDER"},
			path:   "/ai-rag-indexer/acme1/ORDER/filter-metadata",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `[{"key":"id","type":"string"},{"key":"status","name":"Status","type":"string"},{"key":"totalPrice"}]`,
			want:   "FIELD       TYPE\nid          string\nstatus      string\ntotalPrice  \n",
		},
	}
}

// TestRAGReindexHelpSaysWhatAnUnknownTypeGets checks that the help of reindex gives the answer checked live (204,
// exit 0) and not the refusal that the reference lists only as a common response.
func TestRAGReindexHelpSaysWhatAnUnknownTypeGets(t *testing.T) {
	text := helpText(t, "ai", "rag", "reindex")
	if !strings.Contains(text, "Emporix answers an unknown TYPE with 204 No Content (exit 0)") {
		t.Errorf("help lacks the answer to an unknown TYPE:\n%s", text)
	}
	if strings.Contains(text, "400 Bad Request") {
		t.Errorf("help still names a 400 for an unknown TYPE:\n%s", text)
	}
}

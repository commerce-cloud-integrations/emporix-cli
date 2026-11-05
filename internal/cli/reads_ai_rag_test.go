package cli

import (
	"net/http"
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
			// A built-in type in lower case, sent as given.
			args:   []string{"ai", "rag", "fields", "legal-entity"},
			path:   "/ai-rag-indexer/acme1/legal-entity/rag-metadata",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `["name","legalInfo.legalName"]`,
			want:   "FIELD\nname\nlegalInfo.legalName\n",
		},
		{
			// The deprecated name and description are only in -o json.
			args:   []string{"ai", "rag", "filter-fields", "ORDER"},
			path:   "/ai-rag-indexer/acme1/ORDER/filter-metadata",
			scopes: []string{scope},
			mode:   getWhole,
			body:   `[{"key":"id","type":"string"},{"key":"status","name":"Status","type":"string"},{"key":"totalPrice"}]`,
			want:   "FIELD       TYPE    PARENT\nid          string  \nstatus      string  \ntotalPrice          \n",
		},
		{
			// PARENT, the list of objects a field sits in; empty for a field outside a list.
			args:   []string{"ai", "rag", "filter-fields", "CUSTOMER"},
			path:   "/ai-rag-indexer/acme1/CUSTOMER/filter-metadata",
			scopes: []string{scope},
			mode:   getWhole,
			body:   customerFilterFields,
			want: "FIELD              TYPE    PARENT\n" +
				"customerNumber     string  \n" +
				"addresses.zipCode  string  addresses\n",
		},
	}
}

// customerFilterFields are the filter fields of customers, one of them in
// the list of a customer's addresses.
const customerFilterFields = `[{"key":"customerNumber","type":"string"},
	{"key":"addresses.zipCode","type":"string","parent":"addresses"}]`

// TestRAGFilterFieldsCutALongParentInTables checks that a table cuts a long
// PARENT, the list of objects that a field sits in, which csv and -o json
// print whole, and that the help names it.
func TestRAGFilterFieldsCutALongParentInTables(t *testing.T) {
	parent := "mixins.productCustomAttributes." + strings.Repeat("l", 40)
	body := `[{"key":"` + parent + `.code","type":"string","parent":"` + parent + `"}]`
	for _, tt := range []struct {
		format, want string
	}{
		{format: "table", want: "string  " + parent[:maxValueRunes-1] + "…\n"},
		{format: "csv", want: ",string," + parent + "\n"},
		{format: "json", want: `"parent": "` + parent + `"`},
	} {
		fake := newFakeEmporix(t)
		fake.route("/ai-rag-indexer/acme1/PRODUCT/filter-metadata", respond(http.StatusOK, body))
		code, stdout, stderr := runCLI(t, fake.env(), "ai", "rag", "filter-fields", "PRODUCT", "-o", tt.format)
		if code != ExitOK || !strings.Contains(stdout, tt.want) {
			t.Errorf("-o %s: exit %d, stderr %q, stdout lacks %q:\n%s", tt.format, code, stderr, tt.want, stdout)
		}
	}
	if help := helpText(t, "ai", "rag", "filter-fields"); !strings.Contains(help,
		"PARENT is the path of the list of objects a field sits in, such as addresses, and empty outside a list; "+
			"tables cut a long PARENT.") {
		t.Errorf("help does not name PARENT:\n%s", help)
	}
}

// TestRAGHelpNamesTheBuiltInTypes checks that the help of every RAG command
// with TYPE names the four built-in types, which it completes, and that
// Emporix takes them in either case and LEGAL-ENTITY with a hyphen only; and
// what an unknown TYPE gets: 400 for the reads (checked live), while for
// reindex the 204 seen live beside the reference's 400.
func TestRAGHelpNamesTheBuiltInTypes(t *testing.T) {
	const types = entityTypeHelp + " "
	const read = "An unknown TYPE, such as LEGAL_ENTITY, gets 400 Bad Request (exit 1)."
	for _, tt := range []struct {
		use, want, not string
	}{
		{use: "fields", want: types + read, not: "204"},
		{use: "filter-fields", want: types + read, not: "204"},
		{use: "reindex", want: types + "Emporix has answered an unknown TYPE with 204 No Content (exit 0), not the " +
			"400 Bad Request its reference gives, so check the TYPE first.",
			not: read},
	} {
		text := helpText(t, "ai", "rag", tt.use)
		if !strings.Contains(text, tt.want) {
			t.Errorf("help of %s lacks %q:\n%s", tt.use, tt.want, text)
		}
		if strings.Contains(text, tt.not) {
			t.Errorf("help of %s says %q:\n%s", tt.use, tt.not, text)
		}
	}
}

// TestRAGReadOfAnUnknownTypeFails checks that a read of a TYPE that Emporix
// does not know, which it answers 400 Bad Request, fails with exit 1.
func TestRAGReadOfAnUnknownTypeFails(t *testing.T) {
	for use, metadata := range map[string]string{"fields": "rag-metadata", "filter-fields": "filter-metadata"} {
		fake := newFakeEmporix(t)
		fake.route("/ai-rag-indexer/acme1/LEGAL_ENTITY/"+metadata,
			respond(http.StatusBadRequest, `{"message":"Unsupported type"}`))
		code, stdout, stderr := runCLI(t, fake.env(), "ai", "rag", use, "LEGAL_ENTITY")
		if code != ExitError || stdout != "" || !strings.Contains(stderr, "400 Bad Request") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit 1 and the 400", use, code, stdout, stderr)
		}
	}
}

// TestRAGReindexHelpNamesTheReindexJobs checks that the help of the
// deprecated reindex points to the reindex job that replaces it: with --rag
// for products, without it for customers and legal entities.
func TestRAGReindexHelpNamesTheReindexJobs(t *testing.T) {
	const want = "Emporix has deprecated this endpoint in favor of a reindex job: " +
		"`epx indexing reindex-jobs create`, which --wait follows. Give it --rag for PRODUCT; for " +
		"CUSTOMER and LEGAL-ENTITY, the job updates their embeddings without it."
	if text := helpText(t, "ai", "rag", "reindex"); !strings.Contains(text, want) {
		t.Errorf("help lacks %q:\n%s", want, text)
	}
}

package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestSearchTypesRequestTheirOwnReadScope checks that a search of the
// documents of customer or legal-entity requests the read scope of that type
// in place of that of a custom entity type, besides search.search_read; that
// a credential without it fails with exit code 3 naming it, before any
// search is sent; and that a type that only resembles them, or a read of
// custom entity instances, which takes no built-in type, requests that of a
// custom entity type.
func TestSearchTypesRequestTheirOwnReadScope(t *testing.T) {
	documents := func(typ string) []string {
		return []string{"search", "documents", typ, "--saved-search", "s1", "--text", "x"}
	}
	for _, tt := range []struct {
		args  []string
		path  string
		scope string
	}{
		{documents("customer"), "/search/acme1/search/customer", "customer.customer_read"},
		{documents("legal-entity"), "/search/acme1/search/legal-entity", "customermanagement.legalentity_read"},
		{documents("legal_entity"), "/search/acme1/search/legal_entity", "custom.legal_entity_read"},
		{documents("customers"), "/search/acme1/search/customers", "custom.customers_read"},
		{[]string{"custom-entities", "instances", "list", "customer"}, "/schema/acme1/custom-entities/customer/instances",
			"custom.customer_read"},
		{[]string{"custom-entities", "instances", "list", "CUSTOMER"}, "/schema/acme1/custom-entities/CUSTOMER/instances",
			"custom.customer_read"},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, `[]`))
		code, _, stderr := runCLI(t, fake.env(), tt.args...)
		want := "tenant=acme1 " + tt.scope
		if tt.args[0] == searchUse {
			want = "tenant=acme1 search.search_read " + tt.scope
		}
		if code != ExitOK || !slices.Equal(fake.tokenScopes, []string{want}) {
			t.Errorf("%q: exit %d, stderr %q, token scopes %q; want %q", tt.args, code, stderr, fake.tokenScopes, want)
		}
		denied := newFakeEmporix(t, tt.scope)
		code, stdout, stderr := runCLI(t, denied.env(), tt.args...)
		if code != ExitAuth || stdout != "" || !strings.Contains(stderr, "lacks required scopes: "+tt.scope) ||
			len(denied.gets) != 0 {
			t.Errorf("%q without %s: exit %d, stdout %q, stderr %q, requests %q; want exit 3 and none sent",
				tt.args, tt.scope, code, stdout, stderr, denied.gets)
		}
	}
}

// TestSearchTypesOnlyInTheirSpelling checks every command that takes a type
// of the Search Service, reads and writes, as its TYPE argument or in the
// TYPE/ID of --index: customer and legal-entity in another spelling are a
// usage error that names the spelling Emporix keeps its items under, and
// nothing is sent, not even a token request.
func TestSearchTypesOnlyInTheirSpelling(t *testing.T) {
	search, _, err := newTestRoot().Find([]string{searchUse})
	if err != nil {
		t.Fatal(err)
	}
	var typed []*cobra.Command
	walkCommands(search, func(cmd *cobra.Command) {
		if !cmd.HasSubCommands() && argumentNames(cmd.Use, 1)[0] == columnType {
			typed = append(typed, cmd)
		}
	})
	if len(typed) != 9 { // documents, and for-type, get, update and delete of indexes and of queries
		t.Fatalf("%d commands take a TYPE; want 9", len(typed))
	}
	for typ, name := range map[string]string{"CUSTOMER": "customer", "Customer": "customer",
		"LEGAL-ENTITY": "legal-entity", "Legal-Entity": "legal-entity"} {
		runs := [][]string{{"search", "indexes", "export", "--index", typ + "/i1", "--" + writeScopeFlag}}
		for _, cmd := range typed {
			runs = append(runs, typedArguments(cmd, typ))
		}
		for _, args := range runs {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runCLI(t, fake.env(), args...)
			if want := "use " + name + ", as Emporix spells it"; code != ExitUsage || stdout != "" ||
				!strings.Contains(stderr, typ) || !strings.Contains(stderr, want) {
				t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, want)
			}
			fake.assertNothingSent(t)
		}
	}
}

// unscopedEnd is how typeSentence ends: what an unknown type gives.
const unscopedEnd = "has no scope that Emporix grants (exit code 3), and a type created in the last minutes has " +
	"none yet."

// TestTypedCommandsNameTheirBuiltInTypes checks that a command with built-in
// types names them where it refuses what no type is, in what its help says
// TYPE is, and the scopes that each requests before its scope sentence,
// while its annotation holds its own scopes only.
func TestTypedCommandsNameTheirBuiltInTypes(t *testing.T) {
	fake := newFakeEmporix(t)
	code, stdout, stderr := runCLI(t, fake.env(), "search", "documents", "legal entity", "--saved-search", "s1",
		"--text", "x")
	const refused = `invalid TYPE "legal entity": a custom entity type has letters, digits and underscores, ` +
		"the first a letter; a built-in type is customer or legal-entity"
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, refused) {
		t.Errorf("exit %d, stdout %q, stderr %q; want exit 2 and %q", code, stdout, stderr, refused)
	}
	fake.assertNothingSent(t)
	const sentence = "TYPE is customer or legal-entity, exactly as written, or a custom entity type, " +
		"whose ID `epx custom-entities list` shows; an unknown one " + unscopedEnd
	if got := typeSentence(searchTypes()); got != sentence {
		t.Errorf("type sentence %q; want %q", got, sentence)
	}
	if got, want := typeSentence(builtInTypes{}), "TYPE is a custom entity type, whose ID "+
		"`epx custom-entities list` shows; an unknown one "+unscopedEnd; got != want {
		t.Errorf("type sentence without built-in types %q; want %q", got, want)
	}
	const scopes = " In place of custom.<type>_read, customer requires customer.customer_read; legal-entity " +
		"requires customermanagement.legalentity_read.\n\nRequires the search.search_read, custom.<type>_read " +
		"scopes. <type> is the custom entity type in lower case (custom.document_read for DOCUMENT)."
	cmd, _, err := newTestRoot().Find([]string{"search", "documents"})
	if err != nil || !strings.Contains(cmd.Long, sentence) || !strings.HasSuffix(cmd.Long, scopes) ||
		cmd.Annotations[scopesAnnotation] != "search.search_read custom.<type>_read" {
		t.Errorf("help of search documents %q, annotated %q; want %q and to end with %q", cmd.Long,
			cmd.Annotations[scopesAnnotation], sentence, scopes)
	}
	// Nothing to say where no built-in type requests scopes of its own, or no scope names a custom entity type.
	for _, note := range []string{anyCaseTypes("PRODUCT").inPlaceOf([]string{customTypeReadScope}),
		searchDocumentTypes().inPlaceOf(nil)} {
		if note != "" {
			t.Errorf("note %q; want none", note)
		}
	}
}

// TestBuiltInTypesTakeEveryCaseUnlessExact checks that built-in types that a
// command takes in any case, such as those of the AI RAG Indexer, pass in
// every case, and that exact ones pass only as written; what a custom entity
// type is named is checked only where custom.
func TestBuiltInTypesTakeEveryCaseUnlessExact(t *testing.T) {
	anyCase, exact := anyCaseTypes("PRODUCT", "ORDER"), searchTypes()
	custom, other := []string{customTypeReadScope}, []string{searchScope}
	for _, tt := range []struct {
		types  builtInTypes
		arg    string
		scopes []string // of the command, which name a custom entity type or not
		want   string   // in the error; none when empty
	}{
		{anyCase, "product", custom, ""},
		{anyCase, "Order", other, ""},
		{anyCase, "a-b", other, ""},
		{anyCase, "a-b", custom, "the first a letter; a built-in type is PRODUCT or ORDER"},
		{exact, "customer", custom, ""},
		{exact, "Customer", other, "use customer, as Emporix spells it"},
		{exact, "LEGAL-ENTITY", custom, "use legal-entity, as Emporix spells it"},
		{exact, "a-b", other, ""},
		{builtInTypes{}, "a-b", custom, "the first a letter"},
	} {
		err := newTypeRule(tt.scopes, tt.types).accepts(tt.arg)
		if tt.want == "" && err != nil || tt.want != "" && (err == nil || !strings.HasSuffix(err.Error(), tt.want)) {
			t.Errorf("%v: accepts(%q) with scopes %q = %v; want %q", tt.types.names(), tt.arg, tt.scopes, err, tt.want)
		}
	}
	if got, want := anyCase.kinds(), "PRODUCT or ORDER in any case, or a custom entity type"; got != want {
		t.Errorf("kinds %q; want %q", got, want)
	}
	fake := newFakeEmporix(t)
	fake.route("/ai-rag-indexer/acme1/product/rag-metadata", respond(http.StatusOK, `["name"]`))
	if code, stdout, stderr := runCLI(t, fake.env(), "ai", "rag", "fields", "product"); code != ExitOK ||
		stdout != "FIELD\nname\n" {
		t.Errorf("ai rag fields product: exit %d, stdout %q, stderr %q; want the field", code, stdout, stderr)
	}
}

// TestBuiltInTypesAreDeclaredOnAnArgumentWithReadScopes checks that a read
// with built-in types is refused while the command tree is built where its
// request takes no argument, whose value they would be, where one of them
// requests a write-capable scope, which the read would request without its
// opt-in, or where one requests scopes but no scope of the read names a
// custom entity type, which they would replace.
func TestBuiltInTypesAreDeclaredOnAnArgumentWithReadScopes(t *testing.T) {
	typed := []string{"search.search_read", customTypeReadScope}
	for _, tt := range []struct {
		name    string
		path    string
		scopes  []string
		builtIn builtInTypes
		want    string
	}{
		{"no argument", "/things", typed, anyCaseTypes("PRODUCT"),
			"(a scope that names <type>, or built-in types) is the command's first argument, which its request lacks"},
		{"write-capable", "/things/{id}", typed, spelledTypes(builtInType{name: "customer",
			scopes: []string{"customer.customer_manage"}}), "the built-in type customer requests a write-capable scope"},
		{"nothing to replace", "/things/{id}", []string{"search.search_read"}, searchDocumentTypes(),
			"the built-in type customer requests scopes in place of none that names <type>"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recovered, _ := recover().(string); !strings.Contains(recovered, tt.want) {
					t.Fatalf("panic %q; want the declaration refused with %q", recovered, tt.want)
				}
			}()
			rd := reader[thing]{view: thingView(), scopes: tt.scopes}
			rd.command(&app{}, &command{use: "documents TYPE", req: request{path: tt.path}, fetch: listing{},
				builtIn: tt.builtIn})
		})
	}
}

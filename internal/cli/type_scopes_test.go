package cli

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const typeScopeNote = " <type> is the custom entity type in lower case (custom.document_read for DOCUMENT)."

func instanceReads(typ string) [][]string {
	return [][]string{
		{"custom-entities", "instances", "list", typ},
		{"custom-entities", "instances", "search", typ},
		{"custom-entities", "instances", "get", typ, "123"},
		{"search", "documents", typ, "--saved-search", "manuals", "--text", "diesel"},
	}
}

func TestInstanceReadsRequestTheScopeOfTheirType(t *testing.T) {
	for _, args := range instanceReads("Tech_Sheet2") {
		fake := newFakeEmporix(t)
		code, _, stderr := runCLI(t, fake.env(), args...)
		want := "tenant=acme1 custom.tech_sheet2_read"
		if args[0] == searchUse {
			want = "tenant=acme1 search.search_read custom.tech_sheet2_read"
		}
		if code != ExitOK && code != ExitNotFound || !slices.Equal(fake.tokenScopes, []string{want}) {
			t.Errorf("%q: exit %d, stderr %q, token scopes %q; want %q", args, code, stderr, fake.tokenScopes, want)
		}
	}
}

func TestInstanceReadsWithoutTheScopeOfTheirType(t *testing.T) {
	for _, args := range instanceReads("DOCUMENT") {
		fake := newFakeEmporix(t, "custom.document_read")
		code, stdout, stderr := runCLI(t, fake.env(), args...)
		if code != ExitAuth || stdout != "" || !strings.Contains(stderr, "lacks required scopes: custom.document_read") ||
			len(fake.gets) != 0 {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, requests %q; want exit 3 and none sent",
				args, code, stdout, stderr, fake.gets)
		}
	}
}

func TestTypedCommandsRefuseWhatNoTypeIs(t *testing.T) {
	var typed []string
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		read := strings.Fields(cmd.Annotations[scopesAnnotation])
		if !slices.ContainsFunc(append(read, strings.Fields(cmd.Annotations[writeScopesAnnotation])...), namesType) {
			return
		}
		typed = append(typed, cmd.CommandPath())
		if slices.ContainsFunc(read, namesType) && !strings.Contains(strings.ReplaceAll(cmd.Long, "\n", " "),
			"; an unknown one "+string(unscoped)+".") {
			t.Errorf("%s: help %q; want it to say what an unknown type gives", cmd.CommandPath(), cmd.Long)
		}
		for _, typ := range []string{"DOC MANAGE", "DOC.x", "a-b", "9X", "_X", "DÖC", "DOC%20X", "DOC\n"} {
			args := typedArguments(cmd, typ)
			fake := newFakeEmporix(t)
			code, stdout, stderr := runCLI(t, fake.env(), args...)
			want := fmt.Sprintf("invalid %s %q: a custom entity type has letters, digits and underscores, "+
				"the first a letter", argumentNames(cmd.Use, 1)[0], typ)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, want) {
				t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, want)
			}
			fake.assertNothingSent(t)
		}
	})
	root := newTestRoot()
	for _, args := range instanceReads("DOCUMENT") {
		if cmd, _, err := root.Find(args); err != nil || !slices.Contains(typed, cmd.CommandPath()) {
			t.Errorf("%q: not among the commands whose scopes name a custom entity type %q", args, typed)
		}
	}
}

func typedArguments(cmd *cobra.Command, typ string) []string {
	further := slices.Repeat([]string{"123"}, len(strings.Fields(cmd.Use))-2)
	return slices.Concat(strings.Fields(cmd.CommandPath())[1:], []string{typ}, further)
}

func TestTypeScopeSentences(t *testing.T) {
	read, manage := customTypeReadScope, "custom.<type>_manage"
	manageNote := " <type> is the custom entity type in lower case (custom.document_manage for DOCUMENT)."
	for _, tt := range []struct{ got, want string }{
		{scopeSentence([]string{read}), "Requires the custom.<type>_read scope." + typeScopeNote},
		{scopeSentence([]string{"search.search_read", read}),
			"Requires the search.search_read, custom.<type>_read scopes." + typeScopeNote},
		{scopeSentence([]string{manage}), "Requires the custom.<type>_manage scope." + manageNote + " " + optInSentence},
		{writeScopeSentence([]string{manage}), "Requires the custom.<type>_manage scope." + manageNote},
		{writeScopeSentence([]string{read, manage}), "Requires the custom.<type>_read, custom.<type>_manage scopes. " +
			"<type> is the custom entity type in lower case (custom.document_manage, custom.document_read for DOCUMENT)."},
		{groupScopeSentence(scopedGroup(map[string]string{scopesAnnotation: read})),
			"Requires the custom.<type>_read scope." + typeScopeNote},
		{groupScopeSentence(scopedGroup(map[string]string{scopesAnnotation: read},
			map[string]string{writeScopesAnnotation: manage})),
			"Its reads require the custom.<type>_read scope. Its writes require the custom.<type>_manage scope. " +
				"<type> is the custom entity type in lower case (custom.document_manage, custom.document_read for DOCUMENT)."},
		{groupScopeSentence(scopedGroup(map[string]string{scopesAnnotation: read},
			map[string]string{writeScopesAnnotation: read + " " + manage})),
			"Its reads require the custom.<type>_read scope. Its writes require the custom.<type>_manage, " +
				"custom.<type>_read scopes. <type> is the custom entity type in lower case " +
				"(custom.document_manage, custom.document_read for DOCUMENT)."},
		{groupScopeSentence(scopedGroup(map[string]string{scopesAnnotation: read},
			map[string]string{scopesAnnotation: ""})),
			"Its reads require different OAuth scopes; the help of each names them."},
		{groupScopeSentence(scopedGroup(map[string]string{scopesAnnotation: ""},
			map[string]string{writeScopesAnnotation: manage})),
			"Its reads need no OAuth scope. Its writes require the custom.<type>_manage scope." + manageNote},
	} {
		if tt.got != tt.want {
			t.Errorf("got %q, want %q", tt.got, tt.want)
		}
	}
}

func scopedGroup(annotations ...map[string]string) *cobra.Command {
	group := &cobra.Command{Use: "g"}
	for i, annotation := range annotations {
		for j := range 2 {
			group.AddCommand(&cobra.Command{Use: fmt.Sprintf("c%d%d", i, j), Annotations: annotation})
		}
	}
	return group
}

func TestWriteRequestsTheScopeOfItsType(t *testing.T) {
	create := func() *write {
		return &write{
			use: "create-instance TYPE", method: http.MethodPost, scopes: []string{"custom.<type>_manage"},
			req: request{path: "/schema/{tenant}/custom-entities/{parent}/instances"}, body: noBody{},
		}
	}
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/custom-entities/DOCUMENT/instances", respond(http.StatusCreated, `{"id":"1"}`))
	root := writeRoot(fake, create())
	code, stdout, stderr := runWrite(t.Context(), root, "create-instance", "DOCUMENT", "--yes")
	if code != ExitOK || stdout != "{\n  \"id\": \"1\"\n}\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertScopes(t, "custom.document_manage")
	cmd, _, _ := root.Find([]string{"create-instance"})
	if want := "Requires the custom.<type>_manage scope. <type> is the custom entity type in lower case " +
		"(custom.document_manage for DOCUMENT)."; !strings.HasSuffix(cmd.Long, want) {
		t.Errorf("help %q; want it to end with %q", cmd.Long, want)
	}
	refused := newFakeEmporix(t)
	code, _, stderr = runWrite(t.Context(), writeRoot(refused, create()), "create-instance", "DOC MANAGE", "--yes")
	if code != ExitUsage || !strings.Contains(stderr, `invalid TYPE "DOC MANAGE"`) {
		t.Errorf("exit %d, stderr %q; want exit 2 and the type refused", code, stderr)
	}
	refused.assertNothingSent(t)
}

func TestTypeScopesNeedTheTypeArgument(t *testing.T) {
	defer func() {
		if recovered, _ := recover().(string); !strings.Contains(recovered, "a scope that names <type>") {
			t.Fatalf("panic %q; want the declaration refused", recovered)
		}
	}()
	rd := reader[thing]{view: thingView(), scopes: []string{customTypeReadScope}}
	rd.command(&app{}, &command{use: listUse, req: request{path: "/things"}, fetch: listing{}})
}

type scopedBody struct{ noBody }

func (scopedBody) scopes() []string { return []string{customTypeReadScope} }

func TestOnlyTheirOwnScopesNameAType(t *testing.T) {
	for name, w := range map[string]*write{
		"body": {use: "create-instance TYPE", body: scopedBody{}},
		"choice": {use: "create-instance TYPE", body: noBody{}, choice: queryChoice{param: "publish",
			options: []choiceOption{{flag: "publish", value: "true", scopes: []string{"custom.<type>_publish"}}}}},
		"job": {use: "create-instance TYPE", body: noBody{},
			answer: &jobWait{singular: "job", scopes: []string{customTypeReadScope}, shortest: minWaitDuration}},
	} {
		w.method, w.scopes = http.MethodPost, []string{"custom.<type>_manage"}
		w.req = request{path: "/schema/{tenant}/custom-entities/{parent}/instances"}
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recovered, _ := recover().(string); !strings.Contains(recovered, "only the scopes that the command "+
					"itself requests may name <type>") {
					t.Fatalf("panic %q; want the declaration refused", recovered)
				}
			}()
			w.command(&app{})
		})
	}
}

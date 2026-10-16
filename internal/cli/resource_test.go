package cli

import (
	"bytes"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// thing is a neutral resource used to test the shared read builder.
type thing struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// thingView shows a thing as its ID and NAME.
func thingView() view[thing] {
	return view[thing]{
		singular: "thing",
		header:   []string{"ID", "NAME"},
		row:      func(x thing, _ language) []string { return []string{x.ID, x.Name} },
	}
}

// runThings runs `things <args>` on a root holding only the test resource and
// returns the exit code, stdout and the error text.
func runThings(t *testing.T, fake *fakeEmporix, args ...string) (code int, stdout, errText string) {
	t.Helper()
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	root := &cobra.Command{Use: "epx", SilenceErrors: true, SilenceUsage: true}
	view := thingView()
	things := resource[thing]{use: "things", path: "/things", scopes: []string{"thing.read"}, view: view}
	// Opt-in reads (opt_in_test.go): vaults, a write-capable scope that their children inherit.
	vaults := resource[thing]{
		use: "vaults", under: "things", path: "/vaults", scopes: []string{"vault.manage"}, view: view,
	}
	parts := child[thing]{
		use: "parts", scopes: []string{"part.read"}, fetch: listing{}, view: view,
		query: url.Values{"all": {"true"}}, listQuery: url.Values{"expand": {"tags"}},
	}
	parts.legend = "NAME is a part's name."
	keys := child[thing]{use: "keys", fetch: listing{}, view: view}
	// Children of a thing with their own scopes, the parent's (nil), none (empty) and a
	// write-capable one, one part of a thing (and one key of a vault), read with two arguments,
	// and one part by name, read with three.
	root.AddCommand(newResourceCommand(a, things,
		childCommand(a, things, parts),
		childItemCommand(a, things, parts, &childItem{
			use: "named-part", short: "Show one part by name", long: "Reads one part of a thing by name.",
			columns: []string{"NAME", "ID"},
		}),
		childCommand(a, things, child[thing]{use: "kin", fetch: listing{}, view: view}),
		childCommand(a, things, child[thing]{use: "notes", scopes: []string{}, fetch: listing{}, view: view}),
		childCommand(a, things, child[thing]{
			use: "secrets", scopes: []string{"secret.manage"}, fetch: listing{}, view: view,
		}),
		childItemCommand(a, things, parts, &childItem{use: "part", short: "Show one part of a thing",
			long: "Reads one part of a thing."}),
		newResourceCommand(a, vaults,
			childCommand(a, vaults, keys),
			childItemCommand(a, vaults, keys, &childItem{use: "key", long: "Reads one key of a vault."}),
			childCommand(a, vaults, child[thing]{use: "tags", scopes: []string{"tag.read"}, fetch: listing{}, view: view}),
		)))
	applyExitPolicy(root)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"things"}, args...))
	if err := root.Execute(); err != nil {
		return exitCode(err), out.String(), err.Error()
	}
	return ExitOK, out.String(), ""
}

func TestResourceListAndGet(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things", firstPage(`[{"id":"001234","name":"A"},{"id":"2","name":"B"}]`))
	fake.route("/things/001234", respond(http.StatusOK, `{"id":"001234","name":"A"}`))
	code, stdout, errText := runThings(t, fake, "list")
	if code != ExitOK || stdout != "ID      NAME\n001234  A\n2       B\n" {
		t.Fatalf("list: exit %d, stdout %q, error %q", code, stdout, errText)
	}
	code, stdout, errText = runThings(t, fake, "get", "001234")
	if code != ExitOK || stdout != "ID      NAME\n001234  A\n" {
		t.Fatalf("get: exit %d, stdout %q, error %q", code, stdout, errText)
	}
	// One token per command, also for a list read over two pages, with the declared scopes.
	if !slices.Equal(fake.tokenScopes, []string{"tenant=acme1 thing.read", "tenant=acme1 thing.read"}) {
		t.Fatalf("token scopes %q; list and get must request the declared scopes", fake.tokenScopes)
	}
	// An item the service does not know is not found.
	code, stdout, errText = runThings(t, fake, "get", "unknown")
	if code != ExitNotFound || stdout != "" || !strings.Contains(errText, "GET /things/unknown: 404 Not Found") {
		t.Fatalf("unknown: exit %d, stdout %q, error %q", code, stdout, errText)
	}
}

// TestReadWithTwoArguments checks that a read whose request names {parent}
// and {id} takes their arguments in that order, each escaped as a path segment.
func TestReadWithTwoArguments(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/a%2Fb/parts/007", respond(http.StatusOK, `{"id":"007","name":"Bolt"}`))
	code, stdout, errText := runThings(t, fake, "part", "a/b", "007")
	if code != ExitOK || stdout != "ID   NAME\n007  Bolt\n" {
		t.Fatalf("exit %d, stdout %q, error %q", code, stdout, errText)
	}
}

// TestChildScopes checks that a child read requests the scopes it declares
// instead of its parent's, the parent's when it declares none (nil) and no
// scope when it declares an empty list, and that its help says which.
func TestChildScopes(t *testing.T) {
	for _, tt := range []struct{ use, token, help string }{
		{"parts", "tenant=acme1 part.read", "Requires the part.read scope."},
		{"kin", "tenant=acme1 thing.read", "Requires the thing.read scope."},
		{"notes", "tenant=acme1", "No OAuth scope is needed."},
	} {
		fake := newFakeEmporix(t)
		fake.route("/things/t1/"+tt.use, firstPage(`[{"id":"x1","name":"X"}]`))
		code, stdout, errText := runThings(t, fake, tt.use, "t1")
		if code != ExitOK || stdout != "ID  NAME\nx1  X\n" || !slices.Equal(fake.tokenScopes, []string{tt.token}) {
			t.Errorf("%s: exit %d, stdout %q, error %q, token scopes %q; want %q",
				tt.use, code, stdout, errText, fake.tokenScopes, tt.token)
		}
		if _, help, _ := runThings(t, fake, tt.use, "--help"); !strings.Contains(help, tt.help) {
			t.Errorf("%s: help %q; want %q", tt.use, help, tt.help)
		}
	}
}

func TestResourceRefusedScopeExitsWithAuth(t *testing.T) {
	fake := newFakeEmporix(t, "thing.read")
	for _, args := range [][]string{{"list"}, {"get", "x"}} {
		code, stdout, errText := runThings(t, fake, args...)
		if code != ExitAuth || stdout != "" || !strings.Contains(errText, "lacks required scopes: thing.read") {
			t.Fatalf("%v: exit %d, stdout %q, error %q; want exit 3", args, code, stdout, errText)
		}
	}
	// Without the scope, nothing is read.
	if len(fake.gets) != 0 {
		t.Fatalf("requests %q; want none", fake.gets)
	}
}

func TestResourceRejectsMalformedResponses(t *testing.T) {
	tests := []struct {
		name, path, body string
		args             []string
		code             int
		want             string
	}{
		{name: "empty list", path: "/things", body: `[]`, args: []string{"list"}, code: ExitOK},
		{name: "null item", path: "/things", body: `[null]`, args: []string{"list"},
			code: ExitError, want: "item 1 has no identifier"},
		{name: "null collection", path: "/things", body: `null`, args: []string{"list"},
			code: ExitError, want: "expected a JSON array"},
		{name: "object instead of list", path: "/things", body: `{"unexpected":"object"}`, args: []string{"list"},
			code: ExitError, want: "expected a JSON array"},
		{name: "null item body", path: "/things/x", body: `null`, args: []string{"get", "x"},
			code: ExitError, want: "expected one item as a JSON object"},
		{name: "unexpected object", path: "/things/x", body: `{"unexpected":"object"}`, args: []string{"get", "x"},
			code: ExitError, want: "item 1 has no identifier"},
		{name: "not JSON", path: "/things/x", body: `[1,`, args: []string{"get", "x"},
			code: ExitError, want: "decoding thing response"},
		{name: "array item", path: "/things/x", body: `[{"id":"x"}]`, args: []string{"get", "x"},
			code: ExitError, want: "expected one item as a JSON object"},
		{name: "empty object", path: "/things/x", body: `{}`, args: []string{"get", "x"},
			code: ExitError, want: "expected one item, got an empty object"},
		// JSON output prints the response whole, but only a response of the right shape.
		{name: "JSON of an empty array", path: "/things/x", body: `[]`, args: []string{"get", "x", "-o", "json"},
			code: ExitError, want: "decoding thing response: expected one item as a JSON object"},
		{name: "JSON of a null item", path: "/things/x", body: `null`, args: []string{"get", "x", "-o", "json"},
			code: ExitError, want: "expected one item as a JSON object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route(tt.path, firstPage(tt.body))
			code, stdout, errText := runThings(t, fake, tt.args...)
			if code != tt.code || !strings.Contains(errText, tt.want) {
				t.Fatalf("exit %d (want %d), error %q (want %q)", code, tt.code, errText, tt.want)
			}
			if tt.code == ExitOK && stdout != "ID  NAME\n" || tt.code != ExitOK && stdout != "" {
				t.Fatalf("stdout %q; want the header only on success and nothing on failure", stdout)
			}
		})
	}
}

func TestResourceUsageErrorsSendNothing(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"get"}, message: "accepts 1 arg(s), received 0"},
		{args: []string{"get", ""}, message: `invalid ID ""`},
		{args: []string{"get", "."}, message: `invalid ID "."`},
		{args: []string{"get", ".."}, message: `invalid ID ".."`},
		{args: []string{"part", "1"}, message: "accepts 2 arg(s), received 1"},
		{args: []string{"part", "1", "2", "3"}, message: "accepts 2 arg(s), received 3"},
		{args: []string{"part", "..", "1"}, message: `invalid THING_ID ".."`},
		{args: []string{"part", "1", "."}, message: `invalid ID "."`},
		{args: []string{"list", "extra"}, message: `unknown command "extra"`},
		{args: []string{"list", "-o", "xml"}, message: `invalid output format "xml"`},
	} {
		code, _, errText := runThings(t, fake, tt.args...)
		if code != ExitUsage || !strings.Contains(errText, tt.message) {
			t.Errorf("%v: exit %d, error %q (want %q)", tt.args, code, errText, tt.message)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("usage errors must not contact the API: tokens %q, requests %q", fake.tokenScopes, fake.gets)
	}
}

// TestGroupScopeSentence checks that a group's help states the scopes of the
// commands below it, at any depth: those of one kind, reads or writes, as one
// of them states them when they agree; otherwise a line per kind that names
// them when they agree and says that they differ otherwise, followed for
// reads by the opt-in sentence when one of them is write-capable.
func TestGroupScopeSentence(t *testing.T) {
	a := &app{getenv: func(string) string { return "" }}
	read := func(use string, scopes ...string) *cobra.Command {
		return reader[thing]{view: thingView(), scopes: scopes}.command(a, &command{
			use: use, req: request{path: "/things"}, fetch: listing{},
		})
	}
	change := func(use string, scopes ...string) *cobra.Command {
		w := &write{use: use, method: http.MethodPost, req: request{path: "/things"}, scopes: scopes, body: &payload{}}
		return w.command(a)
	}
	const (
		differ       = "Its reads require different OAuth scopes; the help of each names them."
		writesDiffer = "Its writes require different OAuth scopes; the help of each names them."
	)
	for _, tt := range []struct {
		group *cobra.Command
		want  string
	}{
		{newGroup("g", "", "Long.", read("a", "s.read"), newGroup("h", "", "", read("b", "s.read"))),
			"Long.\n\nRequires the s.read scope."},
		{newGroup("g", "", "Long.", read("a"), newGroup("h", "", "", read("b"))), "Long.\n\nNo OAuth scope is needed."},
		{newGroup("g", "", "Long.", read("a", "s.read"), newGroup("h", "", "", read("b"))), "Long.\n\n" + differ},
		{newGroup("g", "", "Long.", read("a", "s.read", "t.read"), read("b", "s.read")), "Long.\n\n" + differ},
		{newGroup("g", "", "Long.", read("a", "t.read", "s.read"), newGroup("h", "", "", read("b", "s.read", "t.read"))),
			"Long.\n\nRequires the s.read, t.read scopes."},
		{newGroup("g", "", "Long.", &cobra.Command{Use: "plain"}), "Long."},
		{newGroup("g", "", "Long.", read("a", "s.manage"), newGroup("h", "", "", read("b", "s.manage"))),
			"Long.\n\nRequires the s.manage scope.\n" + optInSentence},
		{newGroup("g", "", "Long.", read("a", "s.read"), newGroup("h", "", "", read("b", "s.manage"))),
			"Long.\n\n" + differ + "\n" + optInSentence},
		{newGroup("g", "", "Long.", read("a"), change("c", "s.manage")),
			"Long.\n\nIts reads need no OAuth scope.\nIts writes require the s.manage scope."},
		{newGroup("g", "", "Long.", read("a", "s.read"), change("c", "s.manage"),
			newGroup("h", "", "", change("d", "t.manage", "s.manage"))),
			"Long.\n\nIts reads require the s.read scope.\n" + writesDiffer},
		{newGroup("g", "", "Long.", read("a", "s.manage"), read("b"), change("c", "t.manage", "s.read")),
			"Long.\n\n" + differ + "\n" + optInSentence + "\nIts writes require the s.read, t.manage scopes."},
		{newGroup("g", "", "Long.", newGroup("h", "", "", change("c", "s.manage")), change("d", "s.manage")),
			"Long.\n\nRequires the s.manage scope."},
		{newGroup("g", "", "Long.", change("c", "s.manage", "t.read"), newGroup("h", "", "", change("d", "s.manage"))),
			"Long.\n\n" + writesDiffer},
		{newGroup("g", "", "Long.", read("a"), change("b", "s.manage", "t.manage"), change("c", "t.manage", "s.manage")),
			"Long.\n\nIts reads need no OAuth scope.\nIts writes require the s.manage, t.manage scopes."},
	} {
		if tt.group.Long != tt.want {
			t.Errorf("group help %q, want %q", tt.group.Long, tt.want)
		}
	}
}

// TestGroupsStateTheScopesOfTheirReads checks the command tree: the help of
// every read and every declared write ends with its scopes, and so does every
// group of them, or it says that its reads or its writes differ, so no group
// claims a scope that one of its reads or writes lacks.
func TestGroupsStateTheScopesOfTheirReads(t *testing.T) {
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		want := ""
		if scopes, isRead := cmd.Annotations[scopesAnnotation]; isRead {
			want = scopeSentence(strings.Fields(scopes))
		} else if scopes, isWrite := cmd.Annotations[writeScopesAnnotation]; isWrite {
			want = writeScopeSentence(strings.Fields(scopes))
		} else if cmd.HasParent() {
			want = groupScopeSentence(cmd)
		}
		if want != "" && !strings.HasSuffix(cmd.Long, "\n\n"+want) {
			t.Errorf("%s: help %q; want it to end with %q", cmd.CommandPath(), cmd.Long, want)
		}
	})
}

// TestReadsAnnotateTheScopesTheyRequest checks that every read command
// carries the scopes its read cases request (assertScopes) in its annotation,
// those of a custom entity type with the type of its first argument (withType),
// so no read built outside reader.command escapes the group scope sentences.
func TestReadsAnnotateTheScopesTheyRequest(t *testing.T) {
	root := newTestRoot()
	for path, tt := range casesByCommand(t) {
		cmd, rest, err := root.Find(tt.args)
		scopes, isRead := cmd.Annotations[scopesAnnotation]
		annotated := strings.Fields(scopes)
		if err == nil && cmd.ParseFlags(rest) == nil && cmd.Flags().NArg() > 0 {
			annotated = withType(annotated, cmd.Flags().Arg(0))
		}
		if err != nil || !isRead || !slices.Equal(annotated, tt.scopes) {
			t.Errorf("%s: scopes annotation %q (present %v); its read cases request %q", path, scopes, isRead, tt.scopes)
		}
	}
}

func TestScopeSentence(t *testing.T) {
	for _, tt := range []struct {
		scopes []string
		want   string
	}{
		{want: "No OAuth scope is needed."},
		{scopes: []string{"a.read"}, want: "Requires the a.read scope."},
		{scopes: []string{"a.read", "b.read"}, want: "Requires the a.read, b.read scopes."},
		{scopes: []string{"a.manage"}, want: "Requires the a.manage scope.\n" + optInSentence},
		{scopes: []string{"a.read", "b.manage"}, want: "Requires the a.read, b.manage scopes.\n" + optInSentence},
	} {
		if got := scopeSentence(tt.scopes); got != tt.want {
			t.Errorf("scopeSentence(%q) = %q, want %q", tt.scopes, got, tt.want)
		}
	}
}

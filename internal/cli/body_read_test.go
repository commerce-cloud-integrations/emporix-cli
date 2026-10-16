package cli

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// runWithInput runs epx with stdin as its standard input.
func runWithInput(t *testing.T, env map[string]string, stdin io.Reader, args ...string) (
	code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Run(context.Background(), args, Env{
		In: stdin, Out: &out, Err: &errOut, Getenv: func(key string) string { return env[key] },
	})
	return code, out.String(), errOut.String()
}

// writeBody writes content to a file in a new temporary directory and returns its path.
func writeBody(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return file
}

// TestSearchSendsTheBodyGiven checks that --body sends a file's JSON, or
// standard input's for -, as given instead of the body the flags make, also
// to the search of a child collection below its parent item.
func TestSearchSendsTheBodyGiven(t *testing.T) {
	const given = "{ \"q\": \"code:A1\",\n  \"extra\": [1] }"
	for _, search := range []struct {
		args       []string // the search without --body
		path, body string   // where it is sent, and the answer
		want       string   // in its table
	}{
		{args: []string{"products", "search"}, path: "/product/acme1/products/search",
			body: `[{"id":"001234","code":"A1"}]`, want: "001234"},
		{args: []string{"segments", "items", "search", "s 1"}, path: "/customer-segment/acme1/segments/s%201/items/search",
			body: `[{"segmentId":"s 1","type":"PRODUCT","item":{"id":"p1","code":"100200"}}]`, want: "p1  s 1"},
	} {
		for name, args := range map[string][]string{
			"file":  {"--body", writeBody(t, given+"\n")},
			"stdin": {"--body", "-"},
		} {
			fake := newFakeEmporix(t)
			fake.route(search.path, firstPage(search.body))
			code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader(given),
				slices.Concat(search.args, args)...)
			if code != ExitOK || !strings.Contains(stdout, search.want) {
				t.Fatalf("%q %s: exit %d, stdout %q, stderr %q", search.args, name, code, stdout, stderr)
			}
			fake.assertSent(t, http.MethodPost, given)
		}
	}
}

// TestSearchBodyFailuresSendNothing checks that a body that cannot be read
// or is not JSON is a usage error found before anything is sent.
func TestSearchBodyFailuresSendNothing(t *testing.T) {
	directory := t.TempDir()
	for _, tt := range []struct {
		args  []string
		stdin string
		want  string
	}{
		{args: []string{"--body", writeBody(t, `{"q":`)}, want: `body.json": not JSON: unexpected end of JSON input`},
		{args: []string{"--body", writeBody(t, `{"q":""} {}`)}, want: "invalid character '{' after top-level value"},
		{args: []string{"--body", "-"}, stdin: "", want: `--body "-": not JSON: unexpected end of JSON input`},
		{args: []string{"--body", "-"}, stdin: "q=code:A1", want: `--body "-": not JSON: invalid character 'q'`},
		{args: []string{"--body", "-"}, stdin: strings.Repeat(" ", maxBodyFileBytes) + "{}", want: "larger than 1048576"},
		{args: []string{"--body", filepath.Join(directory, "none.json")}, want: `none.json": no such file or directory`},
		{args: []string{"--body", directory}, want: `": is a directory`},
		{args: []string{"--body", ""}, want: `--body "": no such file or directory`},
		{args: []string{"--body", "-", "--query", "code:A1"}, stdin: "{}", want: "[body query] are set none of the others"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader(tt.stdin),
			slices.Concat([]string{"products", "search"}, tt.args)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Errorf("%q: tokens %q, requests %q; want none", tt.args, fake.tokenScopes, fake.gets)
		}
	}
}

// TestSearchUnreadableStdin checks that a failing standard input is a usage
// error too.
func TestSearchUnreadableStdin(t *testing.T) {
	fake := newFakeEmporix(t)
	code, _, stderr := runWithInput(t, fake.env(), iotest.ErrReader(os.ErrClosed), "categories", "search", "--body", "-")
	if code != ExitUsage || !strings.Contains(stderr, `--body "-": file already closed`) || len(fake.gets) != 0 {
		t.Fatalf("exit %d, stderr %q, requests %q", code, stderr, fake.gets)
	}
}

// TestSearchStdinStopsOnInterrupt checks that an interrupt, which ends the
// command's context, ends a search waiting on a standard input that never
// ends: a runtime error (exit 1) that sends nothing.
func TestSearchStdinStopsOnInterrupt(t *testing.T) {
	fake := newFakeEmporix(t)
	exited, errOut := runUntilReading(t, []string{"products", "search", "--body", "-"}, fake.env(), false)
	select {
	case code := <-exited:
		if code != ExitError || !strings.Contains(errOut.String(), `--body "-": context canceled`) ||
			len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Fatalf("exit %d, stderr %q, tokens %q, requests %q; want exit 1 and nothing sent",
				code, errOut.String(), fake.tokenScopes, fake.gets)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the search still waits on standard input after the interrupt")
	}
}

// priceCommand returns a read with a body sent with method, `price ID`, as a
// calculation of a thing's price that Emporix serves as PUT, run for "t 1".
func priceCommand(fake *fakeEmporix, method string) *cobra.Command {
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	cmd := reader[thing]{view: thingView()}.command(a, &command{
		use: "price ID", long: "Calculates the price of a thing.", req: request{path: "/things/{id}/price"},
		fetch: bodyRead{method: method, fetch: single{items: oneObject}},
	})
	applyExitPolicy(cmd)
	cmd.SetArgs([]string{"t 1"})
	return cmd
}

// TestBodyReadReadsOneResponse checks a read with a body whose answer is one
// object, as a calculation that Emporix serves as PUT: without typed flags,
// its body is an empty object.
func TestBodyReadReadsOneResponse(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/t%201/price", respond(http.StatusOK, `{"id":"t 1","name":"12.10"}`))
	cmd := priceCommand(fake, http.MethodPut)
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil || out.String() != "ID   NAME\nt 1  12.10\n" {
		t.Fatalf("error %v, stdout %q", err, out.String())
	}
	fake.assertSent(t, http.MethodPut, "{}")
	want := "Calculates the price of a thing.\n\nChanges nothing: Emporix serves this read as PUT with a JSON body"
	if !strings.HasPrefix(cmd.Long, want) || !strings.HasSuffix(cmd.Long, "\n\nNo OAuth scope is needed.") {
		t.Fatalf("help %q", cmd.Long)
	}
}

// TestBodyReadSendsNothingWithAnotherMethod checks that a read with a body
// declared with a method that may change data is a usage error (exit 2) that
// sends nothing, not even a token request.
func TestBodyReadSendsNothingWithAnotherMethod(t *testing.T) {
	fake := newFakeEmporix(t)
	cmd := priceCommand(fake, http.MethodDelete)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); exitCode(err) != ExitUsage || len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("error %v, tokens %q, requests %q; want exit 2 and nothing sent", err, fake.tokenScopes, fake.gets)
	}
}

// TestSearchesSayTheyChangeNothing checks the command tree: exactly the
// commands that take --body send a body, as their read cases state, and the
// help of every read sent as POST or PUT, with a body or, as a read of a file
// is, without one, says that it changes nothing.
func TestSearchesSayTheyChangeNothing(t *testing.T) {
	root := newTestRoot()
	searches := commandsWithFlag(bodyFlag)
	for path, tt := range casesByCommand(t) {
		cmd, _, err := root.Find(tt.args)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		sendsBody := tt.sent != ""
		if takesBody := slices.Contains(searches, path); takesBody != sendsBody {
			t.Errorf("%s: takes --body %v, but its read cases send the body %q", path, takesBody, tt.sent)
			continue
		}
		posted := tt.method == http.MethodPost || tt.method == http.MethodPut
		if posted && !strings.Contains(cmd.Long, "Changes nothing: Emporix serves this read as "+tt.method) {
			t.Errorf("%s: help %q; want it to say that it changes nothing", path, cmd.Long)
		}
	}
}

// TestCompletionOffersFilesForBodyOnly checks that a search's --body
// completes file names and its other flags do not.
func TestCompletionOffersFilesForBodyOnly(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string // the completion directive
	}{
		{args: []string{"products", "search", "--body", ""}, want: ":0\n"}, // cobra.ShellCompDirectiveDefault
		{args: []string{"category-trees", "search", "--body", "b"}, want: ":0\n"},
		{args: []string{"products", "search", "--query", ""}, want: ":4\n"}, // cobra.ShellCompDirectiveNoFileComp
		{args: []string{"category-trees", "search", "--category", ""}, want: ":4\n"},
		{args: []string{"products", "search", ""}, want: ":4\n"},
	} {
		code, stdout, _ := runCLI(t, nil, append([]string{cobra.ShellCompRequestCmd}, tt.args...)...)
		if code != ExitOK || !strings.HasSuffix(stdout, tt.want) {
			t.Errorf("%v: exit %d, completions %q; want directive %q", tt.args, code, stdout, tt.want)
		}
	}
	// A flag that adds a value or an object to the body repeats, so completion offers it again.
	for _, args := range [][]string{
		{"category-trees", "search", "--category", "c1", "--cat"},
		{"search", "indexes", "export", "--index", "a/b", "--ind"},
		{"segments", "match", "--product", "p1", "--prod"},
	} {
		flag := args[len(args)-3]
		if _, stdout, _ := runCLI(t, nil, append([]string{cobra.ShellCompRequestCmd}, args...)...); !strings.Contains(
			stdout, flag+"\t") {
			t.Errorf("%q: completions %q; want %s again", args, stdout, flag)
		}
	}
}

// TestBodyReadRequiresItsSelections checks the flags of a body that Emporix
// rejects without them: unless --body gives the body, each is required, as
// its help says, and a missing one is a usage error that sends nothing; with
// --body, none may be given.
func TestBodyReadRequiresItsSelections(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"search", "indexes", "export"}, want: "the flags in the group [body index] is required"},
		{args: []string{"custom-entities", "export"}, want: "the flags in the group [body type] is required"},
		{args: []string{"segments", "match", "--segment", "S1", "--site", "main"}, want: "[body product] is required"},
		{args: []string{"segments", "match", "--product", "p1", "--site", "main"}, want: "[body segment] is required"},
		{args: []string{"segments", "match", "--product", "p1", "--segment", "S1"}, want: "[body site] is required"},
		{args: []string{"search", "indexes", "export", "--index", "a/b", "--body", "-"}, want: "[body index] are set none"},
		{args: []string{"segments", "match", "--site", "main", "--body", "-"}, want: "[body site] are set none"},
		{args: []string{"custom-entities", "export", "--type", "CAR", "--body", "-"}, want: "[body type] are set none"},
	} {
		code, stdout, stderr := runCLI(t, fake.env(), append(tt.args, "--"+writeScopeFlag)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("tokens %q, requests %q; want none", fake.tokenScopes, fake.gets)
	}
	const emptyExport = "Emporix refuses an empty selection (--body with [], exit 1)."
	for _, tt := range []struct{ args, want []string }{
		{args: []string{"search", "indexes", "export"}, want: []string{"(repeatable; required without --body)", emptyExport}},
		{args: []string{"segments", "match"}, want: []string{
			"check (repeatable; required without --body)", "check against (repeatable; required without --body)",
			"to match on, e.g. main (required without --body)",
		}},
		{args: []string{"custom-entities", "export"}, want: []string{"to export (repeatable; required without --body)\n",
			emptyExport}},
	} {
		_, help, _ := runCLI(t, nil, append(tt.args, "--help")...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%q: help %q; want %q", tt.args, help, want)
			}
		}
	}
}

// TestBodyReadsWithoutTypedFlags checks what a read with a body sends
// without its typed flags: the body --body gives, as given, in place of
// required ones.
func TestBodyReadsWithoutTypedFlags(t *testing.T) {
	const exported = `{"data":"W10="}`
	for _, tt := range []struct {
		args              []string
		path, stdin, sent string
		answer            string
	}{
		{
			args: []string{"search", "indexes", "export", "--body", "-"}, path: "/search/acme1/search/indexes/export",
			stdin: `[{"type":"vehicle","id":"vehicles"}]`, sent: `[{"type":"vehicle","id":"vehicles"}]`, answer: exported,
		},
		{
			args: []string{"segments", "match", "--body", "-"}, path: "/customer-segment/acme1/segments/match",
			stdin: `{"siteCode":"main","items":[{"id":"p1"}],"segmentIds":["S1"]}`,
			sent:  `{"siteCode":"main","items":[{"id":"p1"}],"segmentIds":["S1"]}`, answer: `[]`,
		},
		{
			args: []string{"custom-entities", "export", "--body", "-"}, path: "/schema/acme1/custom-entities/export",
			stdin: `["CAR"]`, sent: `["CAR"]`, answer: exported,
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, tt.answer))
		code, _, stderr := runWithInput(t, fake.env(), strings.NewReader(tt.stdin), append(tt.args, "--"+writeScopeFlag)...)
		if code != ExitOK {
			t.Errorf("%q: exit %d, stderr %q", tt.args, code, stderr)
		}
		fake.assertSent(t, http.MethodPost, tt.sent)
	}
}

// TestBodyReadsRequireTheirFields checks that a read with a body whose
// members Emporix requires fails as a usage error that sends nothing without
// their flags, unless --body gives the body whole.
func TestBodyReadsRequireTheirFields(t *testing.T) {
	match := []string{"prices", "match", "--product", "100200", "--site", "main", "--currency", "EUR"}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"prices", "search", "--site", "main"}, want: "[body product] is required"},
		{args: match, want: "[body country] is required"},
		{args: []string{"search", "documents", "DOCUMENT", "--saved-search", "manuals"}, want: "[body text] is required"},
		{args: []string{"search", "documents", "DOCUMENT", "--text", "x", "--body", "-"},
			want: "[body text] are set none of the others"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Errorf("%q: tokens %q, requests %q; want none", tt.args, fake.tokenScopes, fake.gets)
		}
	}
	// --body replaces every flag that makes the body.
	const given = `{"siteCode":"main","targetCurrency":"EUR","targetLocation":{"countryCode":"NL"},"items":[]}`
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/match-prices", respond(http.StatusOK, `[]`))
	code, stdout, stderr := runCLI(t, fake.env(), "prices", "match", "--body", writeBody(t, given))
	if code != ExitOK || !strings.HasPrefix(stdout, "ID  ITEM  PRICE LIST") {
		t.Fatalf("--body: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertSent(t, http.MethodPost, given)
}

// TestPostedReadsRejectArgumentsThatSplitTheirPath checks that an argument
// of a read sent as POST, with a body or without one (a read of a file), may
// not hold what a service, or a proxy before it, might decode into another
// path (/, \ or %, the client refuses them), nor be . or .. (as for any
// read); nothing is sent, not even a token request.
func TestPostedReadsRejectArgumentsThatSplitTheirPath(t *testing.T) {
	const split = ": a path segment of a read sent as POST is empty, . or .., or holds /, \\ or %"
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"price-lists", "prices", "search", "pl/1"},
			want: "POST /price/{tenant}/price-lists/pl%2F1/prices/search" + split},
		{args: []string{"segments", "customers", "search", `s\x`},
			want: "POST /customer-segment/{tenant}/segments/s%5Cx/customers/search" + split},
		{args: []string{"segments", "items", "search", "items%2Fimport"},
			want: "POST /customer-segment/{tenant}/segments/items%252Fimport/items/search" + split},
		// A custom entity type, which also names a scope, has letters, digits and underscores only.
		{args: []string{"custom-entities", "instances", "search", `DOC\x`}, want: `invalid CUSTOM_ENTITY_ID "DOC\\x"`},
		{args: []string{"search", "documents", "indexes%2Fimport", "--body", "-"}, want: `invalid TYPE "indexes%2Fimport"`},
		{args: []string{"search", "documents", "..", "--body", "-"}, want: `invalid TYPE ".."`},
		{args: []string{"search", "documents", "--body", "-"}, want: "accepts 1 arg(s), received 0"},
		// A read of a file, sent as POST without a body, refuses them as well.
		{args: []string{"quotes", "pdf", "Q/1", "--out", "-"}, want: "POST /quote/{tenant}/quotes/Q%2F1/pdf" + split},
		{args: []string{"quotes", "pdf", "..", "--out", "-"}, want: `invalid QUOTE_ID ".."`},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runWithInput(t, fake.env(), strings.NewReader("{}"), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Errorf("%q: tokens %q, requests %q; want none", tt.args, fake.tokenScopes, fake.gets)
		}
	}
}

// TestBodyFieldsLeaveOutFlagsNotGiven checks that a body leaves out the
// members of optional flags not given, and holds those given as given, an
// object member as its object.
func TestBodyFieldsLeaveOutFlagsNotGiven(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{want: `{"q":"","targetLocation":{"countryCode":""}}`},
		{
			args: []string{"--country", "NL", "--site", ""},
			want: `{"q":"","siteCode":"","targetLocation":{"countryCode":"NL"}}`,
		},
	} {
		cmd := &cobra.Command{}
		cmd.Flags().String(bodyFlag, "", "")
		made := objectBody{queryField(), stringField("siteCode", "site", ""), matchLocationField()}.bind(cmd, bodyFlag)
		if err := cmd.Flags().Parse(tt.args); err != nil {
			t.Fatal(err)
		}
		if body, err := emporix.EncodeBody(made()); err != nil || string(body) != tt.want {
			t.Errorf("%q: body %s, error %v; want %s", tt.args, body, err, tt.want)
		}
	}
}

// TestRequiredBodyFlagsSaySo checks that the usage of a flag required unless
// --body is given says so, within the parentheses that end it, if any.
func TestRequiredBodyFlagsSaySo(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"prices", "search"}, want: "a product ID whose prices to read (repeatable, up to 100; required " +
			"without --body)\n"},
		{args: []string{"prices", "match"}, want: "the code of the country to match the price for, e.g. NL (required " +
			"without --body)\n"},
	} {
		code, stdout, stderr := runCLI(t, nil, append(tt.args, "--help")...)
		if code != ExitOK || !strings.Contains(stdout, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want %q", tt.args, code, stdout, stderr, tt.want)
		}
	}
}

// TestSearchedCollectionsNameTheirSearch checks the groups of collections that
// Emporix lists only by search (newSearchedResourceCommand), which have get
// and search but no list: every read of theirs that takes an argument, get
// and the reads below one item, says that search shows the IDs.
func TestSearchedCollectionsNameTheirSearch(t *testing.T) {
	found := 0
	walkCommands(newTestRoot(), func(group *cobra.Command) {
		reads := group.Commands()
		has := func(use string) bool {
			return slices.ContainsFunc(reads, func(cmd *cobra.Command) bool { return cmd.Name() == use })
		}
		if !has("get") || !has(collectionSearch) || has(listUse) {
			return
		}
		found++
		search := "the ID that `" + group.CommandPath() + " " + collectionSearch
		for _, read := range reads {
			if len(strings.Fields(read.Use)) > 1 && !strings.Contains(read.Long, search) {
				t.Errorf("%s: help %q; want %q", read.CommandPath(), read.Long, search)
			}
		}
	})
	if found == 0 {
		t.Fatal("no collection listed only by search found")
	}
}

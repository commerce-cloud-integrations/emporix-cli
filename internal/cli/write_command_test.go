package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestDeclaredWritesStateTheirScopes checks every write command but those of
// api, which take theirs from --scope: its help ends with the scopes that its
// write case requests, those of a read first included and those that its
// flags add (annotated on the flag: of its choice or its parameter, which
// TestWrites checks per case) left out, and so does its
// annotation (groupScopeSentence), those of a custom entity type as declared
// (typeScopes); it is no read (no scopes annotation, no
// --columns) and no opt-in; and it takes its body from --file exactly where
// one of its cases gives a file, which it then requires, as its usage says,
// unless one of its cases sends a body without a file (typed flags), and
// otherwise from its arguments or not at all.
func TestDeclaredWritesStateTheirScopes(t *testing.T) {
	root := newTestRoot()
	withFile, typed := map[*cobra.Command]bool{}, map[*cobra.Command]bool{}
	for _, tt := range writeCases() {
		if cmd, _, err := root.Find(tt.args); err == nil {
			withFile[cmd] = withFile[cmd] || tt.file != ""
			typed[cmd] = typed[cmd] || tt.file == "" && tt.sent != ""
		}
	}
	for _, tt := range writeCases() {
		checkDeclaredWrite(t, root, &tt, withFile, typed)
	}
}

// checkDeclaredWrite checks that the write of the case tt, which api get is not
// part of, states its scopes, and that its --file is required and present
// exactly as its cases give it (withFile, typed).
func checkDeclaredWrite(t *testing.T, root *cobra.Command, tt *writeCase, withFile, typed map[*cobra.Command]bool) {
	t.Helper()
	if tt.args[0] == "api" {
		return
	}
	cmd, rest, err := root.Find(tt.args)
	if err != nil {
		t.Fatalf("%q: %v", tt.args, err)
	}
	fill := typeOfRun(cmd, rest)
	scopes := slices.DeleteFunc(slices.Concat(tt.readScopes, tt.scopes), flagScope(cmd, fill))
	declared := strings.Fields(cmd.Annotations[writeScopesAnnotation])
	filled := fill(declared)
	_, isRead := cmd.Annotations[scopesAnnotation]
	if want := "\n\n" + writeScopeSentence(declared); !strings.HasSuffix(cmd.Long, want) ||
		isRead || cmd.Flags().Lookup(columnsFlag) != nil || cmd.Flags().Lookup(writeScopeFlag) != nil {
		t.Errorf("%s: help %q, a read %v; want it to end with %q, and no --%s or --%s", cmd.CommandPath(),
			cmd.Long, isRead, want, columnsFlag, writeScopeFlag)
	}
	if !slices.Equal(filled, scopes) {
		t.Errorf("%s: write scopes annotation %q; want %q", cmd.CommandPath(), declared, scopes)
	}
	checkFileFlag(t, cmd, withFile[cmd], typed[cmd])
}

// checkFileFlag checks the --file of cmd: present where a case gives a file,
// and required unless a case sends without one (typed).
func checkFileFlag(t *testing.T, cmd *cobra.Command, withFile, typed bool) {
	t.Helper()
	file := cmd.Flags().Lookup(fileFlag)
	if (file != nil) != withFile || file != nil && (strings.HasSuffix(file.Usage, "; required)") == typed ||
		(len(file.Annotations[cobra.BashCompOneRequiredFlag]) == 0) != typed) {
		t.Errorf("%s: --file %v; want it exactly where a case gives a file, required unless a case sends "+
			"without one", cmd.CommandPath(), file)
	}
}

// typeOfRun returns how scopes are filled in for a run of cmd with rest, its
// arguments and flags: those of a custom entity type with the type of its
// first argument (withType), as typeScopes fills them in.
func typeOfRun(cmd *cobra.Command, rest []string) func(scopes []string) []string {
	if cmd.ParseFlags(rest) == nil && cmd.Flags().NArg() > 0 {
		typ := cmd.Flags().Arg(0)
		return func(scopes []string) []string { return withType(scopes, typ) }
	}
	return func(scopes []string) []string { return scopes }
}

// flagScope returns whether a scope is one that a flag of cmd adds, given or
// not, as the flag's annotation names it, filled in by fill (typeOfRun).
func flagScope(cmd *cobra.Command, fill func([]string) []string) func(scope string) bool {
	var added []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) { added = append(added, flag.Annotations[writeScopesAnnotation]...) })
	added = fill(added)
	return func(scope string) bool { return slices.Contains(added, scope) }
}

// writeRoot returns a root holding only the write command of w, which runs
// against fake where standard input is not a terminal.
func writeRoot(fake *fakeEmporix, w *write) *cobra.Command { return writeRootOn(fake, w, false) }

// writeRootOn returns a root holding only the write command of w, which runs
// against fake where standard input is a terminal or not.
func writeRootOn(fake *fakeEmporix, w *write, terminal bool) *cobra.Command {
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout, terminal: terminal}
	root := &cobra.Command{Use: "epx", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(w.command(a))
	applyExitPolicy(root)
	return root
}

// runWrite runs `epx <args>` in ctx on root (writeRoot) and returns the exit
// code, stdout and stderr, which ends with the error, if any, as Run writes it.
func runWrite(ctx context.Context, root *cobra.Command, args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		return exitCode(err), out.String(), errOut.String() + "Error: " + err.Error() + "\n"
	}
	return ExitOK, out.String(), errOut.String()
}

// partPatch is a write whose request names every placeholder: two arguments
// and the site; it sends no body.
func partPatch() *write {
	return &write{
		use: "patch THING_ID ID", method: http.MethodPatch, scopes: []string{"thing.manage"},
		req: request{path: "/things/{parent}/parts/{id}", query: url.Values{"siteCode": {"{site}"}}}, body: noBody{},
	}
}

// TestWriteCommandFillsItsRequest checks that a write fills in its request
// as a read does: its arguments, escaped as path segments, and its required
// --site, in the path and the query.
func TestWriteCommandFillsItsRequest(t *testing.T) {
	fake := newFakeEmporix(t)
	code, stdout, stderr := runWrite(t.Context(), writeRoot(fake, partPatch()), "patch", "t 1", "p:2", "--site",
		"AcmeStore_NL", "--dry-run")
	want := "PATCH " + fake.server.URL + "/things/t%201/parts/p:2?siteCode=AcmeStore_NL\n"
	if code != ExitOK || !strings.HasPrefix(stdout, want) {
		t.Fatalf("exit %d, stderr %q, stdout %q; want it to start with %q", code, stderr, stdout, want)
	}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"patch", "t1", "p1", "--dry-run"}, want: `required flag(s) "site" not set`},
		{args: []string{"patch", "t1", "--site", "AcmeStore_NL", "--dry-run"}, want: "accepts 2 arg(s), received 1"},
		{args: []string{"patch", "..", "p1", "--site", "AcmeStore_NL", "--yes"}, want: `invalid THING_ID ".."`},
		{args: []string{"patch", "t1", "p1", "--site", "AcmeStore_NL"}, want: "which standard input is not"},
	} {
		code, _, stderr := runWrite(t.Context(), writeRoot(fake, partPatch()), tt.args...)
		if code != ExitUsage || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stderr %q; want exit 2 and %q", tt.args, code, stderr, tt.want)
		}
	}
	fake.assertNothingSent(t)
}

// TestDeclaredWriteUsageErrorsSendNothing checks that a declared write
// without its arguments, --file or a required flag, with a value that a flag
// refuses, or with an ID that would leave its path segment, is a usage error
// (exit 2), also for a dry run, that sends nothing.
func TestDeclaredWriteUsageErrorsSendNothing(t *testing.T) {
	const file = `{"name":"Acme"}`
	assertRefused(t, []refusedWrite{
		{args: []string{"brands", "update"}, file: file, want: "accepts 1 arg(s), received 0"},
		{args: []string{"brands", "patch", "b1", "b2"}, file: file, want: "accepts 1 arg(s), received 2"},
		{args: []string{"labels", "update", ".."}, file: file, want: `invalid ID ".."`},
		{args: []string{"labels", "patch", "a/b"}, file: file,
			want: "PATCH /label/labels/a%2Fb: a path segment of a write is empty, . or .., or holds /, \\ or %"},
		{args: []string{"brands", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"brands", "create", "x"}, file: file, want: `unknown command "x" for "epx brands create"`},
		{args: []string{"schemas", "update-types", "s1"}, file: file, want: `required flag(s) "version" not set`},
		{args: []string{"schemas", "update-types", "s1", "--version", "1", "--mode", "SET"}, file: file,
			want: `invalid argument "SET" for "--mode" flag: use REPLACE, ADD or REMOVE`},
		{args: []string{"schemas", "update-types", "s1", "--version", ""}, file: file,
			want: `invalid argument "" for "--version" flag: empty`},
		{args: []string{"schemas", "update-types", "s1", "--version", "abc"}, file: file,
			want: `invalid argument "abc" for "--version" flag: use a whole number, e.g. 2`},
		{args: []string{"schemas", "update-types", "s1", "--version=-1"}, file: file,
			want: `invalid argument "-1" for "--version" flag: use a whole number`},
		{args: []string{"schemas", "update-types", "s1", "--version", "1.5"}, file: file,
			want: `invalid argument "1.5" for "--version" flag: use a whole number`},
	})
}

// TestFormFilesCompleteFileNames checks that a flag that names a file of a
// form completes file names, as --file does.
func TestFormFilesCompleteFileNames(t *testing.T) {
	for _, args := range [][]string{
		{"schemas", "references", "create", "--json-schema", ""},
		{"schemas", "parse", "--json-schema", "s"},
	} {
		code, stdout, _ := runCLI(t, nil, append([]string{cobra.ShellCompRequestCmd}, args...)...)
		if code != ExitOK || !strings.HasSuffix(stdout, ":0\n") { // cobra.ShellCompDirectiveDefault
			t.Errorf("%v: exit %d, completions %q; want file names", args, code, stdout)
		}
	}
}

// TestFileUsageSaysWhetherItIsRequired checks that the --file of a declared
// write says that it is required, after how standard input gives it, and
// that of api, which sends no body without it, does not, nor that of a write
// whose flags make the body, which says that it replaces that body.
func TestFileUsageSaysWhetherItIsRequired(t *testing.T) {
	root := newTestRoot()
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"brands", "create"}, want: "JSON request body in `PATH`, sent byte for byte (- reads standard " +
			"input; sending it then needs --yes; required)"},
		{args: []string{"schemas", "references", "update"}, want: "JSON in `PATH`, sent byte for byte as the body " +
			"field of the form (- reads standard input; sending it then needs --yes; required)"},
		{args: []string{"api", "post"}, want: "JSON request body in `PATH`, sent byte for byte (- reads standard " +
			"input; sending it then needs --yes)"},
		{args: []string{"reward-points", "add"}, want: "JSON request body in `PATH`, sent byte for byte instead " +
			"of the one that epx makes (- reads standard input; sending it then needs --yes)"},
	} {
		cmd, _, err := root.Find(tt.args)
		if err != nil || cmd.Flags().Lookup(fileFlag).Usage != tt.want {
			t.Errorf("%q: --file %q (%v); want %q", tt.args, cmd.Flags().Lookup(fileFlag).Usage, err, tt.want)
		}
	}
}

// TestOptInNamesReadsOnly checks that the opt-in sentence of a group that
// holds opt-in reads and writes names reads only, as writes take no
// --allow-write-scope, and that a write's help ends with its scope alone.
func TestOptInNamesReadsOnly(t *testing.T) {
	_, help, _ := runCLI(t, nil, "schemas", "--help")
	if want := "Its reads require different OAuth scopes; the help of each names them.\nOpt-in: epx requests a " +
		"write-capable scope for a read only when given --allow-write-scope.\nIts writes require the " +
		"schema.schema_manage scope.\n"; !strings.Contains(help, want) {
		t.Errorf("schemas help %q; want %q, whose opt-in names reads only", help, want)
	}
	_, help, _ = runCLI(t, nil, "schemas", "create", "--help")
	if strings.Contains(help, "Opt-in") || !strings.Contains(help, "\n\nRequires the schema.schema_manage scope.\n") {
		t.Errorf("schemas create help %q; want it to end with its scope, without the opt-in of a read", help)
	}
}

// TestArgumentFieldNeedsItsArgument checks that a body member that repeats an
// argument fails as the command is built, not as it runs, when the command's
// usage does not name that argument.
func TestArgumentFieldNeedsItsArgument(t *testing.T) {
	defer func() {
		if recovered := recover(); !strings.Contains(fmt.Sprint(recovered),
			`create ID: the body member "customerId" repeats an argument CUSTOMER that it does not take`) {
			t.Fatalf("recovered %v; want the misdeclared write named", recovered)
		}
	}()
	(&write{
		use: "create ID", method: http.MethodPost, scopes: []string{"thing.manage"}, req: request{path: "/things/{id}"},
		body: &payload{typed: objectBody{argumentField("customerId", "CUSTOMER")}},
	}).command(&app{})
}

// TestArgumentFieldsRepeatTheirArguments checks that members of a typed body
// repeat the arguments that their usage names, each by its name, also the
// second one.
func TestArgumentFieldsRepeatTheirArguments(t *testing.T) {
	fake := newFakeEmporix(t)
	root := writeRoot(fake, &write{
		use: "create TYPE ID", method: http.MethodPost, scopes: []string{"thing.manage"},
		req:  request{path: "/things/{parent}/items/{id}"},
		body: &payload{typed: objectBody{argumentField("id", "ID"), argumentField("type", "TYPE")}},
	})
	code, stdout, stderr := runWrite(t.Context(), root, "create", "t 1", "--dry-run", "--", "-i1")
	if code != ExitOK || !strings.Contains(stdout, "/things/t%201/items/-i1") ||
		!strings.Contains(stdout, `{"id":"-i1","type":"t 1"}`) {
		t.Errorf("exit %d, stdout %q, stderr %q; want the arguments in the path and the body", code, stdout, stderr)
	}
	fake.assertNothingSent(t)
}

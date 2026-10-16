package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestOptInReadsRequireTheFlag checks the opt-in reads of the test resource
// (runThings): a child with a write-capable scope of its own, a resource with
// one, and its child and child item, which inherit it. Without
// --allow-write-scope, also when it is given false, each fails as a usage
// error naming its scope, before a token is requested; with it, each reads
// with exactly its scopes.
func TestOptInReadsRequireTheFlag(t *testing.T) {
	const list, item = `[{"id":"x1","name":"X"}]`, `{"id":"x1","name":"X"}`
	for _, tt := range []struct {
		command, path, body, scope string
		args                       []string // after the command
	}{
		{command: "secrets", args: []string{"t1"}, path: "/things/t1/secrets", body: list, scope: "secret.manage"},
		{command: "vaults list", path: "/vaults", body: list, scope: "vault.manage"},
		{command: "vaults get", args: []string{"x1"}, path: "/vaults/x1", body: item, scope: "vault.manage"},
		{command: "vaults keys", args: []string{"v1"}, path: "/vaults/v1/keys", body: list, scope: "vault.manage"},
		{command: "vaults key", args: []string{"v1", "x1"}, path: "/vaults/v1/keys/x1", body: item, scope: "vault.manage"},
	} {
		run := append(strings.Fields(tt.command), tt.args...)
		fake := newFakeEmporix(t)
		fake.route(tt.path, firstPage(tt.body))
		for _, args := range [][]string{run, append(slices.Clone(run), "--"+writeScopeFlag+"=false")} {
			code, stdout, errText := runThings(t, fake, args...)
			want := "epx things " + tt.command + " requests " + tt.scope + ", which can also change data, only when given --" +
				writeScopeFlag
			if code != ExitUsage || stdout != "" || errText != want {
				t.Errorf("%q: exit %d, stdout %q, error %q; want exit 2 and %q", args, code, stdout, errText, want)
			}
		}
		if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Fatalf("%s: tokens %q, requests %q; want none without --%s",
				tt.command, fake.tokenScopes, fake.gets, writeScopeFlag)
		}
		code, stdout, errText := runThings(t, fake, append(run, "--"+writeScopeFlag)...)
		token := []string{"tenant=acme1 " + tt.scope}
		if code != ExitOK || stdout != "ID  NAME\nx1  X\n" || !slices.Equal(fake.tokenScopes, token) {
			t.Errorf("%s --%s: exit %d, stdout %q, error %q, token scopes %q",
				tt.command, writeScopeFlag, code, stdout, errText, fake.tokenScopes)
		}
	}
}

// TestChildWithItsOwnScopesLeavesTheOptIn checks that a child of an opt-in
// resource that declares its own read scopes is not opt-in: it reads without
// --allow-write-scope and does not take it.
func TestChildWithItsOwnScopesLeavesTheOptIn(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/vaults/v1/tags", firstPage(`[{"id":"x1","name":"X"}]`))
	code, stdout, errText := runThings(t, fake, "vaults", "tags", "v1")
	token := []string{"tenant=acme1 tag.read"}
	if code != ExitOK || stdout != "ID  NAME\nx1  X\n" || !slices.Equal(fake.tokenScopes, token) {
		t.Fatalf("exit %d, stdout %q, error %q, token scopes %q", code, stdout, errText, fake.tokenScopes)
	}
	code, _, errText = runThings(t, fake, "vaults", "tags", "v1", "--"+writeScopeFlag)
	if code != ExitUsage || !strings.Contains(errText, "unknown flag: --"+writeScopeFlag) {
		t.Fatalf("exit %d, error %q; want an unknown flag", code, errText)
	}
}

// TestOptInHelp checks that an opt-in read's help ends with the opt-in
// sentence and that its flag names the scope it requests.
func TestOptInHelp(t *testing.T) {
	_, help, _ := runThings(t, newFakeEmporix(t), "secrets", "--help")
	for _, want := range []string{
		"Requires the secret.manage scope.\n" + optInSentence + "\n",
		"--" + writeScopeFlag, "opt in to requesting secret.manage, which can also change data (required)",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help %q; want %q", help, want)
		}
	}
}

// TestWriteCapable checks which scopes count as write-capable: every scope
// but those named for reading, with read or view as a word after the service.
func TestWriteCapable(t *testing.T) {
	for _, tt := range []struct {
		scopes []string
		want   bool
	}{
		{scopes: nil, want: false},
		{scopes: []string{"coupon.coupon_read"}, want: false},
		{scopes: []string{"indexing.search_view"}, want: false},
		{scopes: []string{"category.category_read_unpublished"}, want: false},
		{scopes: []string{"payment-gateway.paymentmodes_read"}, want: false},
		{scopes: []string{"thing.read"}, want: false},
		{scopes: []string{"coupon.coupon_manage"}, want: true},
		{scopes: []string{"importtool.import_trigger"}, want: true},
		{scopes: []string{"order.order_readascustomer"}, want: true}, // not named for reading: opt-in, to be safe
		{scopes: []string{"read.manage"}, want: true},                // the service is not a word of the name
		{scopes: []string{"read"}, want: true},
		{scopes: []string{"coupon.coupon_read", "coupon.coupon_manage"}, want: true},
	} {
		if got := writeCapable(tt.scopes); got != tt.want {
			t.Errorf("writeCapable(%q) = %v, want %v", tt.scopes, got, tt.want)
		}
	}
}

// TestEveryWriteCapableReadIsOptIn checks the command tree: a read takes
// --allow-write-scope exactly when one of its scopes is write-capable, and
// each such read, run as its read case runs it but without the flag, fails as
// a usage error naming the flag before a token is requested or anything is
// read, and requests exactly its scopes with it. So no write-capable scope,
// such as coupon.coupon_manage, is ever requested implicitly.
func TestEveryWriteCapableReadIsOptIn(t *testing.T) {
	cases := casesByCommand(t)
	optIn := 0
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		scopes, isRead := cmd.Annotations[scopesAnnotation]
		if !isRead {
			return
		}
		writes, hasFlag := writeCapable(strings.Fields(scopes)), cmd.Flags().Lookup(writeScopeFlag) != nil
		if writes != hasFlag {
			t.Errorf("%s: scopes %q, --%s %v; want it exactly for a write-capable scope",
				cmd.CommandPath(), scopes, writeScopeFlag, hasFlag)
		}
		if writes {
			optIn++
			assertOptInEnforced(t, cases[cmd.CommandPath()].args, strings.Fields(scopes))
		}
	})
	if optIn == 0 {
		t.Fatal("no opt-in read found")
	}
}

// assertOptInEnforced runs args, the read case of an opt-in read, which opts
// in with --allow-write-scope: without the flag it must fail as a usage error
// naming it and send nothing; as given, it must request exactly scopes.
func assertOptInEnforced(t *testing.T, args, scopes []string) {
	t.Helper()
	optIn := "--" + writeScopeFlag
	without := slices.DeleteFunc(slices.Clone(args), func(arg string) bool { return arg == optIn })
	if len(without) == len(args) {
		t.Errorf("%q: the read case of an opt-in read must give %s", args, optIn)
		return
	}
	fake := newFakeEmporix(t)
	code, stdout, stderr := runCLI(t, fake.env(), without...)
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "only when given --"+writeScopeFlag+"\n") ||
		len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Errorf("%q: exit %d, stdout %q, stderr %q, tokens %q, requests %q; want exit 2 naming --%s and nothing sent",
			without, code, stdout, stderr, fake.tokenScopes, fake.gets, writeScopeFlag)
	}
	runCLI(t, fake.env(), args...)
	want := []string{strings.Join(append([]string{"tenant=acme1"}, scopes...), " ")}
	if !slices.Equal(fake.tokenScopes, want) {
		t.Errorf("%q: token scopes %q, want %q", args, fake.tokenScopes, want)
	}
}

// TestOptInReadsFailWithUsageBeforeConfiguration checks a real opt-in read:
// without --allow-write-scope it fails as a usage error, pointing to the flag,
// even before the configuration is checked, and sends nothing.
func TestOptInReadsFailWithUsageBeforeConfiguration(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/coupon/acme1/coupons/SAVE7/redemptions", respond(http.StatusOK, `[]`))
	for _, env := range []map[string]string{fake.env(), nil} {
		code, stdout, stderr := runCLI(t, env, "coupons", "redemptions", "SAVE7")
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr,
			"epx coupons redemptions requests coupon.coupon_manage, which can also change data, only when given "+
				"--allow-write-scope\nRun 'epx --help' for usage.") {
			t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("tokens %q, requests %q; want none", fake.tokenScopes, fake.gets)
	}
}

package cli

import (
	"cmp"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type writeCase struct {
	args           []string
	file           string
	form           []sentPart
	read           string
	readScopes     []string
	sent           string
	method         string
	path           string
	query          string
	scopes         []string
	lang           string
	confirmBy      string
	many           string
	status         int
	body           string
	want           string
	conflict       string
	conflictAnswer string
	conflictShown  string
	refuses        bool
}

func writeCases() []writeCase {
	return slices.Concat(apiWrites(), unitWrites(), priceAndAvailabilityWrites(), configurationWrites(),
		indexingWrites(), brandAndLabelWrites(), schemaServiceWrites(), catalogWriteCases(),
		tenantSetupWrites(), webhookWrites(), searchWrites(), sequentialIDWrites(), shippingWrites(),
		deliveryTimeWriteCases(), b2bWrites(), paymentGatewayWrites(), deleteWrites(), iamWrites(), aiWrites(),
		orderWriteCases(), invoiceJobWriteCases(), returnWriteCases(), quoteWriteCases(), couponWriteCases(),
		rewardPointWriteCases(), shoppingListWriteCases(), importWriteCases(), mediaWriteCases(),
		sessionContextWriteCases(), customEntityWriteCases(), priceModelAndListWrites(), feeAndItemFeeWrites(),
		cartWriteCases(), productWriteCases(), customerServiceWrites(), segmentServiceWrites())
}

func apiWrites() []writeCase {
	return []writeCase{
		{
			args: []string{"api", "post", "/product/{tenant}/products?doIndex=true&skipVariantGeneration=true",
				"--scope", "product.product_manage", "-p", "doIndex=false"},
			file: "{ \"id\": \"100200\",\n  \"weight\": 1.50 }\n", method: http.MethodPost, path: "/product/acme1/products",
			query: "doIndex=false&skipVariantGeneration=true", scopes: []string{"product.product_manage"},
			status: http.StatusCreated, body: `{"id":"100200","weight":1.50}`,
			want: "{\n  \"id\": \"100200\",\n  \"weight\": 1.50\n}\n",
		},
		{
			args: []string{"api", "put", "/product/{tenant}/products/100200", "--scope", "product.product_manage"},
			file: `{"id":"100200","metadata":{"version":3}}`, method: http.MethodPut, path: "/product/acme1/products/100200",
			scopes: []string{"product.product_manage"}, status: http.StatusNoContent,
		},
		{
			args: []string{"api", "patch", "/site/{tenant}/sites/AcmeStore_NL", "--scope", "site.site_manage"},
			file: `{"active":false}`, method: http.MethodPatch, path: "/site/acme1/sites/AcmeStore_NL",
			scopes: []string{"site.site_manage"}, status: http.StatusOK, body: `{"code":"AcmeStore_NL","active":false}`,
			want: "{\n  \"active\": false,\n  \"code\": \"AcmeStore_NL\"\n}\n",
		},
		{
			args:   []string{"api", "delete", "/product/{tenant}/products/100200", "--scope", "product.product_manage"},
			method: http.MethodDelete, path: "/product/acme1/products/100200", scopes: []string{"product.product_manage"},
			confirmBy: "/product/{tenant}/products/100200", status: http.StatusNoContent,
		},
	}
}

func writeEnv(fake *fakeEmporix, terminal bool, stdin string) Env {
	env := fake.env()
	return Env{In: strings.NewReader(stdin), Getenv: func(key string) string { return env[key] }, Terminal: terminal}
}

func (tt *writeCase) uri(base string) string {
	if tt.query == "" {
		return base + tt.path
	}
	return base + tt.path + "?" + tt.query
}

func (tt *writeCase) sentBody() string { return cmp.Or(tt.file, tt.sent) }

func (tt *writeCase) with(t *testing.T, flags ...string) []string {
	t.Helper()
	args := slices.Clone(tt.args)
	if tt.file != "" {
		args = append(args, "--"+fileFlag, writeBody(t, tt.file))
	}
	return append(args, flags...)
}

func (tt *writeCase) fake(t *testing.T, status int) *fakeEmporix {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route(tt.path, tt.serve(status, tt.body))
	return fake
}

func (tt *writeCase) serve(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if tt.read != "" && r.Method == http.MethodGet {
			respond(http.StatusOK, tt.read)(w, r)
			return
		}
		respond(status, body)(w, r)
	}
}

func (tt *writeCase) readFirst() (uris, tokens []string) {
	if tt.read == "" {
		return nil, nil
	}
	return []string{tt.path}, []string{tokenScope(tt.readScopes)}
}

func (tt *writeCase) assertSent(t *testing.T, fake *fakeEmporix) {
	t.Helper()
	uris, tokens := tt.readFirst()
	fake.assertRequests(t, append(uris, tt.uri(""))...)
	fake.assertTokens(t, append(tokens, tokenScope(tt.scopes))...)
	last := len(fake.sent) - 1
	if tt.form != nil {
		fake.assertForm(t, last, tt.method, tt.form)
	} else if body := tt.sentBody(); fake.sent[last] != strings.TrimSpace(tt.method+" "+body) ||
		fake.bodies[last] != body {
		t.Fatalf("sent %q with body %q; want %s with %q", fake.sent[last], fake.bodies[last], tt.method, body)
	}
	if fake.languages[last] != tt.lang {
		t.Fatalf("sent %q in %q; want it in %q", fake.sent[last], fake.languages[last], tt.lang)
	}
}

func (tt *writeCase) dryRun(base string) *regexp.Regexp {
	head, tail := tt.method+" "+tt.uri(base)+"\nAccept: application/json\nAccept-Language: *\n", ""
	if tt.lang != "" {
		head += "Content-Language: " + tt.lang + "\n"
	}
	head = regexp.QuoteMeta(head)
	switch body := tt.sentBody(); {
	case tt.form != nil:
		head += regexp.QuoteMeta("Content-Type: multipart/form-data; boundary=") + `\S+\n`
		tail = "\n"
		for _, part := range tt.form {
			tail += part.described()
		}
	case body != "":
		head += regexp.QuoteMeta("Content-Type: application/json\n")
		tail = "\n" + strings.TrimSuffix(body, "\n") + "\n"
	}
	return regexp.MustCompile(`\A` + head + `User-Agent: emporix-cli/\S+\n` + regexp.QuoteMeta(tail) + `\z`)
}

func (tt *writeCase) assertNoWrite(t *testing.T, fake *fakeEmporix) {
	t.Helper()
	uris, tokens := tt.readFirst()
	fake.assertRequests(t, uris...)
	fake.assertTokens(t, tokens...)
}

func tokenScope(scopes []string) string {
	return strings.Join(append([]string{"tenant=acme1"}, scopes...), " ")
}

func (f *fakeEmporix) assertTokens(t *testing.T, scopes ...string) {
	t.Helper()
	if !slices.Equal(f.tokenScopes, scopes) {
		t.Fatalf("token scopes %q; want %q", f.tokenScopes, scopes)
	}
}

func (f *fakeEmporix) assertNothingSent(t *testing.T) {
	t.Helper()
	if len(f.tokenScopes) != 0 || len(f.gets) != 0 {
		t.Fatalf("tokens %q, requests %q; want nothing sent", f.tokenScopes, f.gets)
	}
}

func TestWrites(t *testing.T) {
	for _, tt := range writeCases() {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			fake := tt.fake(t, tt.status)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag)...)
			if code != ExitOK || stdout != tt.want || stderr != "" {
				t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, tt.want)
			}
			tt.assertSent(t, fake)
		})
	}
}

func TestEveryWriteHasACase(t *testing.T) {
	root := newTestRoot()
	writes := commandsWithFlag(dryRunFlag)
	var cased []string
	for _, tt := range writeCases() {
		cmd, _, err := root.Find(tt.args)
		if err != nil || !slices.Contains(writes, cmd.CommandPath()) {
			t.Errorf("write case %q names no write command", tt.args)
			continue
		}
		cased = append(cased, cmd.CommandPath())
	}
	for _, path := range writes {
		if !slices.Contains(cased, path) {
			t.Errorf("%s has no write case", path)
		}
	}
}

func TestWriteDryRunsSendNothing(t *testing.T) {
	for _, tt := range writeCases() {
		fake := tt.fake(t, tt.status)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+dryRunFlag)...)
		want := tt.dryRun(fake.server.URL)
		if code != ExitOK || !want.MatchString(stdout) || stderr != "" {
			t.Errorf("%q: exit %d, stderr %q, stdout:\n%s\nwant to match %s", tt.args, code, stderr, stdout, want)
		}
		tt.assertNoWrite(t, fake)
		assertNoSecret(t, stdout)
	}
}

func TestWritesNeedYesWithoutATerminal(t *testing.T) {
	for _, tt := range writeCases() {
		fake := tt.fake(t, tt.status)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, "y\n"), tt.with(t)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "which standard input is not: give --yes") {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 asking for --yes", tt.args, code, stdout, stderr)
		}
		fake.assertNothingSent(t)
	}
}

func TestWritesAskOnATerminal(t *testing.T) {
	for _, tt := range writeCases() {
		answers, question := tt.confirmation(t)
		for answer, confirmed := range answers {
			checkConfirmedAnswer(t, &tt, answer, confirmed, question)
		}
	}
}

func checkConfirmedAnswer(t *testing.T, tt *writeCase, answer string, confirmed bool, question string) {
	t.Helper()
	fake := tt.fake(t, tt.status)
	code, stdout, stderr := runIn(t, writeEnv(fake, true, answer), tt.with(t)...)
	prompt := "Send " + tt.method + " " + tt.uri(fake.server.URL) + " to tenant acme1? " + question
	if !strings.HasPrefix(stderr, prompt) {
		t.Errorf("%q answered %q: stderr %q; want the prompt %q", tt.args, answer, stderr, prompt)
	}
	if confirmed {
		if code != ExitOK || stdout != tt.want {
			t.Errorf("%q answered %q: exit %d, stdout %q, stderr %q", tt.args, answer, code, stdout, stderr)
		}
		tt.assertSent(t, fake)
		return
	}
	if code != ExitError || stdout != "" || !strings.HasSuffix(stderr, "Error: "+errNotConfirmed.Error()+"\n") {
		t.Errorf("%q answered %q: exit %d, stdout %q, stderr %q; want it declined", tt.args, answer, code, stdout,
			stderr)
	}
	tt.assertNoWrite(t, fake)
}

func TestWriteConflictsExitWith5(t *testing.T) {
	for _, tt := range writeCases() {
		fake := newFakeEmporix(t)
		answer, printed := `{"message":"version mismatch"}`, ""
		if tt.conflictAnswer != "" {
			answer, printed = tt.conflictAnswer, tt.conflictShown
		}
		fake.route(tt.path, tt.serve(http.StatusConflict, answer))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag)...)
		message := ": version mismatch"
		if tt.conflictAnswer != "" {
			message = ""
		}
		want := tt.method + " " + tt.path + ": 409 Conflict" + message + " (" + cmp.Or(tt.conflict,
			"the item changed since it was read, or exists already: read it again before you retry") + ")\n"
		if code != ExitConflict || stdout != printed || !strings.HasSuffix(stderr, want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 5, stdout %q and %q", tt.args, code, stdout,
				stderr, printed, want)
		}
	}
}

func TestCompletionOffersFilesForFileOnly(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"api", "post", "/x", "--file", ""}, want: ":0\n"},
		{args: []string{"api", "delete", "/x", "--file", "b"}, want: ":0\n"},
		{args: []string{"api", "put", ""}, want: ":4\n"},
		{args: []string{"api", "patch", "/x", "--scope", ""}, want: ":4\n"},
		{args: []string{"api", "post", "/x", "-p", ""}, want: ":4\n"},
	} {
		code, stdout, _ := runCLI(t, nil, append([]string{cobra.ShellCompRequestCmd}, tt.args...)...)
		if code != ExitOK || !strings.HasSuffix(stdout, tt.want) {
			t.Errorf("%v: exit %d, completions %q; want directive %q", tt.args, code, stdout, tt.want)
		}
	}
}

func TestRootHelpNamesTheConflictExit(t *testing.T) {
	_, help, _ := runCLI(t, nil, "--help")
	if !containsFlat(help, "4 not found, 5 conflict: a write that Emporix answered with 409 Conflict, as") ||
		!containsFlat(help, "`units update` or `catalogs patch` with metadata.version answered with 404 Not Found") {
		t.Fatalf("help %q; want exit code 5 explained", help)
	}
}

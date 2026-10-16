package cli

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

// TestWriteDryRunShowsTheExactRequest checks the whole text of a dry run,
// which asks nothing, also on a terminal: the method and URL with the tenant,
// every header but Authorization, sorted, and the body byte for byte, or
// none without --file.
func TestWriteDryRunShowsTheExactRequest(t *testing.T) {
	fake := newFakeEmporix(t)
	version := regexp.MustCompile(`(?m)^User-Agent: emporix-cli/\S+$`) // the version of the test binary
	for body, ending := range map[string]string{"{\"name\": \"A\"}\n": "", `[1, 2]`: "\n"} {
		code, stdout, stderr := runIn(t, writeEnv(fake, true, "y\n"), "api", "post", "/brand/brands", "--scope",
			"brand.brand_manage", "--file", writeBody(t, body), "--dry-run")
		stdout = version.ReplaceAllString(stdout, "User-Agent: emporix-cli/VERSION")
		want := "POST " + fake.server.URL + "/brand/brands\nAccept: application/json\nAccept-Language: *\n" +
			"Content-Type: application/json\nUser-Agent: emporix-cli/VERSION\n\n" + body + ending
		if code != ExitOK || stdout != want || stderr != "" {
			t.Errorf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, want)
		}
	}
	code, stdout, _ := runIn(t, writeEnv(fake, false, ""), "api", "delete", "/site/{tenant}/sites/Acme Store",
		"--dry-run")
	stdout = version.ReplaceAllString(stdout, "User-Agent: emporix-cli/VERSION")
	want := "DELETE " + fake.server.URL + "/site/acme1/sites/Acme%20Store\nAccept: application/json\n" +
		"Accept-Language: *\nUser-Agent: emporix-cli/VERSION\n"
	if code != ExitOK || stdout != want {
		t.Errorf("exit %d, stdout:\n%s\nwant:\n%s", code, stdout, want)
	}
	fake.assertNothingSent(t)
}

// TestWriteBodyFailuresSendNothing checks that a body that cannot be read or
// is not JSON, and a body on standard input without --yes, are usage errors
// (exit 2) found before anything is sent, also before standard input is read.
func TestWriteBodyFailuresSendNothing(t *testing.T) {
	directory := t.TempDir()
	untouched := iotest.ErrReader(os.ErrClosed) // fails any read of standard input
	for _, tt := range []struct {
		file  string
		stdin io.Reader
		want  string
	}{
		{file: writeBody(t, `{"name":`), want: `body.json": not JSON: unexpected end`},
		{file: writeBody(t, ""), want: `body.json": not JSON: unexpected end of JSON input`},
		{file: writeBody(t, `{} {}`), want: "invalid character '{' after top-level value"},
		{file: filepath.Join(directory, "none.json"), want: `none.json": no such file`},
		{file: directory, want: `": is a directory`},
		{file: "-", stdin: strings.NewReader("name=A"), want: `--file "-": not JSON`},
		{file: "-", stdin: strings.NewReader(strings.Repeat(" ", maxWriteFileBytes) + "{}"),
			want: "larger than 16777216 bytes"},
		{file: "-", stdin: untouched, want: `--file "-": file already closed`},
	} {
		fake := newFakeEmporix(t)
		env := writeEnv(fake, false, "")
		if tt.stdin != nil {
			env.In = tt.stdin
		}
		code, stdout, stderr := runIn(t, env, "api", "post", "/brand/brands", "--file", tt.file, "--yes")
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("--file %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.file, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
	// Standard input that gives the body cannot confirm it: --yes is required, on a terminal too.
	fake := newFakeEmporix(t)
	env := writeEnv(fake, true, "")
	env.In = untouched
	code, _, stderr := runIn(t, env, "api", "post", "/brand/brands", "--file", "-")
	if code != ExitUsage || !strings.Contains(stderr, "reads the body from standard input, which then cannot "+
		"confirm it: give --yes") {
		t.Errorf("exit %d, stderr %q; want exit 2 asking for --yes", code, stderr)
	}
	fake.assertNothingSent(t)
}

// TestWriteUsageErrorsSendNothing checks that invalid arguments, flags,
// scopes or configuration, and a path that leaves the API host or its
// segments, are usage errors (exit 2), also for a dry run, that send nothing.
func TestWriteUsageErrorsSendNothing(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"delete"}, want: "accepts 1 arg(s), received 0"},
		// A path segment that a service, or a proxy before it, might decode into other segments.
		{args: []string{"delete", "/things/t%2F..%2Fx"}, want: "DELETE /things/t%2F..%2Fx: a path segment of a write"},
		{args: []string{"patch", "/things/.."}, want: "PATCH /things/..: a path segment of a write"},
		{args: []string{"put", "/things/t%252Fx"}, want: "PUT /things/t%252Fx: a path segment of a write"},
		{args: []string{"post", "https://evil.example/x"}, want: `API path "https://evil.example/x"`},
		{args: []string{"post", "/x", "-p", "novalue"}, want: `invalid argument "novalue" for "-p, --param"`},
		{args: []string{"post", "/x", "--scope", "a b"}, want: `invalid scope "a b"`},
		{args: []string{"post", "/x", "--dry-run=maybe"}, want: `invalid argument "maybe"`},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(append([]string{"api"}, tt.args...),
				flag)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, flag, code, stdout,
					stderr, tt.want)
			}
		}
	}
	fake.assertNothingSent(t)
	code, _, stderr := runCLI(t, map[string]string{envTenant: "acme1"}, "api", "delete", "/x", "--dry-run")
	if code != ExitUsage || !strings.Contains(stderr, "missing configuration: set "+envClientID) {
		t.Errorf("exit %d, stderr %q; want missing configuration", code, stderr)
	}
}

// answering answers the requests in turn with statuses, the last one again
// for any further request, each with the message canned and no wait before
// a request is sent again.
func answering(statuses ...int) http.HandlerFunc {
	answers := make([]http.HandlerFunc, len(statuses))
	for i, status := range statuses {
		answers[i] = respond(status, `{"message":"canned"}`, "Retry-After", "0")
	}
	return inTurn(answers...)
}

// TestWriteFailuresKeepTheirExitCodes checks that a write refused for its
// scopes exits with 3, one of an unknown item with 4, that only a conflict
// exits with 5, and that a write that may have been carried out although it
// failed says so and exits with 1: it is not sent again, so whatever would
// answer it then (a 409, 404 or 403) decides nothing.
func TestWriteFailuresKeepTheirExitCodes(t *testing.T) {
	const note = " (it may have been carried out all the same: read the item before you send it again)\n"
	for _, tt := range []struct {
		method   string
		denied   []string
		statuses []int
		code     int
		want     string // the end of standard error
	}{
		{method: "delete", denied: []string{"thing.manage"}, code: ExitAuth, want: "lacks required scopes: thing.manage\n"},
		{method: "delete", statuses: []int{404}, code: ExitNotFound, want: "DELETE /things/t1: 404 Not Found: canned\n"},
		{method: "post", statuses: []int{400}, code: ExitError, want: "POST /things/t1: 400 Bad Request: canned\n"},
		{method: "patch", statuses: []int{409}, code: ExitConflict, want: "409 Conflict: canned (the item changed since"},
		{method: "put", statuses: []int{503, 409}, code: ExitError,
			want: "PUT /things/t1: 503 Service Unavailable: canned" + note},
		{method: "delete", statuses: []int{504, 404}, code: ExitError,
			want: "DELETE /things/t1: 504 Gateway Timeout: canned" + note},
		{method: "put", statuses: []int{502, 403}, code: ExitError, want: "PUT /things/t1: 502 Bad Gateway: canned" + note},
		{method: "post", statuses: []int{503}, code: ExitError, want: "POST /things/t1: 503 Service Unavailable: canned" +
			note},
		{method: "patch", statuses: []int{500}, code: ExitError,
			want: "PATCH /things/t1: 500 Internal Server Error: canned" + note},
	} {
		fake := newFakeEmporix(t, tt.denied...)
		fake.route("/things/t1", answering(append(tt.statuses, http.StatusOK)...))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "api", tt.method, "/things/t1", "--scope",
			"thing.manage", "--yes")
		if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%s answered %v: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.method, tt.statuses, code,
				stdout, stderr, tt.code, tt.want)
		}
	}
}

// TestWriteResponsesThatAreNotJSON checks that a write answered with a body
// that is not JSON succeeds, printing nothing but a note on standard error.
func TestWriteResponsesThatAreNotJSON(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/t1", respond(http.StatusOK, "deleted\x1b]0;"))
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "api", "delete", "/things/t1", "--yes")
	if code != ExitOK || stdout != "" || stderr != "The response (11 bytes) is not JSON, so it is not printed.\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestWriteHelp checks what the help of a write says: what it sends, how it
// is sent, then its scopes, its flags and an example with a scope that the
// reference names for the endpoint.
func TestWriteHelp(t *testing.T) {
	_, help, _ := runCLI(t, nil, "api", "post", "--help")
	for _, want := range []string{
		"Content-Language that the Unit Service requires to create or update a unit.\n\nChanges data: on a " +
			"terminal, epx asks before it sends POST.",
		"asking (required without a terminal, or with --file -); --dry-run prints the request\n" +
			"instead of sending it.\n\nRequests only the scopes given with --scope",
		"epx api post /brand/brands --scope brand.brand_manage --file brand.json", "--file PATH", "--dry-run", "--yes",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help %q; want %q", help, want)
		}
	}
}

// TestWriteConfirmationStopsOnInterrupt checks that an interrupt, which ends
// the command's context, ends a write waiting for its confirmation: an error
// (exit 1) that sends nothing.
func TestWriteConfirmationStopsOnInterrupt(t *testing.T) {
	fake := newFakeEmporix(t)
	exited, errOut := runUntilReading(t, []string{"api", "delete", "/x"}, fake.env(), true)
	select {
	case code := <-exited:
		if code != ExitError || !strings.Contains(errOut.String(), "waiting for confirmation: context canceled") {
			t.Fatalf("exit %d, stderr %q; want exit 1", code, errOut.String())
		}
		fake.assertNothingSent(t)
	case <-time.After(10 * time.Second):
		t.Fatal("the write still waits for its confirmation after the interrupt")
	}
}

// TestMisdeclaredLanguageHelpPanics checks that a write with a languageHelp
// of its own panics as the command tree is built unless it sends strings or
// maps (stringsOrMaps) or maps only (mapsOnly), whose help alone it replaces:
// on another write it would not say how the write sends its localized text.
func TestMisdeclaredLanguageHelpPanics(t *testing.T) {
	const want = "create: a write of headerless maps or without localized text has no language help of its own"
	for _, mode := range []languageMode{notLocalized, headerlessMaps} {
		func() {
			defer func() {
				if recovered, _ := recover().(string); recovered != want {
					t.Errorf("mode %d: recovered %q; want %q", mode, recovered, want)
				}
			}()
			(&write{
				use: createUse, method: http.MethodPost, scopes: []string{"thing.thing_manage"}, body: &payload{},
				req: request{path: "/things"}, language: mode, languageHelp: "Localized text is a map.",
			}).command(&app{})
		}()
	}
}

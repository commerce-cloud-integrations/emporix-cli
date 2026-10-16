package cli

import (
	"cmp"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// The paths of the reindex jobs, as the fake serves them.
const (
	reindexJobsPath = "/indexing/acme1/reindex-jobs"
	reindexJobPath  = reindexJobsPath + "/j1"
)

// jobCreate is `indexing reindex-jobs create`, whose --wait may read the job
// every millisecond, unlike the command's, which reads it every second at
// most.
func jobCreate() *write {
	w := reindexJobCreate(reindexJobs())
	if job, follows := w.answer.(*jobWait); follows {
		job.shortest = time.Millisecond
	}
	return w
}

// waitFor runs `create --entity-type PRODUCT --yes` of jobCreate against
// fake, polling every millisecond, with the flags.
func waitFor(t *testing.T, fake *fakeEmporix, flags ...string) (code int, stdout, stderr string) {
	t.Helper()
	args := []string{"create", "--entity-type", "PRODUCT", "--yes", "--poll-interval", "1ms"}
	return runWrite(t.Context(), writeRoot(fake, jobCreate()), append(args, flags...)...)
}

// asPrinted returns body as epx prints JSON.
func asPrinted(t *testing.T, body string) string {
	t.Helper()
	var printed strings.Builder
	if err := output.WriteJSON(&printed, json.RawMessage(body)); err != nil {
		t.Fatal(err)
	}
	return printed.String()
}

// TestReindexJobWaitFollowsTheJob checks that --wait reads the job that
// create started again, with the read scope too, until it ends, saying on
// standard error how its status changes, and prints it as it ended.
func TestReindexJobWaitFollowsTheJob(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route(reindexJobsPath, respond(http.StatusCreated, `{"id":"j1","status":"PENDING"}`))
	// As Emporix sends it, a job that ended with SUCCESS says only how many items it expected.
	last := `{"id":"j1","status":"SUCCESS","expectedCount":5}`
	fake.route(reindexJobPath, inTurn(respond(http.StatusOK, `{"id":"j1","status":"PENDING"}`),
		respond(http.StatusOK, `{"id":"j1","status":"IN_PROGRESS","expectedCount":5,"processedCount":2}`),
		respond(http.StatusOK, `{"id":"j1","status":"PUBLISHED","expectedCount":5,"processedCount":5}`),
		respond(http.StatusOK, last)))
	code, stdout, stderr := waitFor(t, fake, "--wait")
	progress := "reindex job j1: PENDING\nreindex job j1: IN_PROGRESS (2 of 5)\nreindex job j1: PUBLISHED (5 of 5)\n" +
		"reindex job j1: SUCCESS\n"
	if code != ExitOK || stdout != asPrinted(t, last) || stderr != progress {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertScopes(t, "indexing.search_manage", "indexing.search_view")
	fake.assertRequests(t, reindexJobsPath, reindexJobPath, reindexJobPath, reindexJobPath, reindexJobPath)
	if want := []string{`POST {"entityType":"PRODUCT"}`, "GET", "GET", "GET", "GET"}; !slices.Equal(fake.sent, want) {
		t.Fatalf("sent %q; want %q", fake.sent, want)
	}
}

// TestReindexJobWaitFailsUnlessTheJobSucceeds checks that a job that ends
// otherwise than with SUCCESS, also one that Emporix answers with as it is
// created, is printed and fails (exit 1), naming its status and message.
func TestReindexJobWaitFailsUnlessTheJobSucceeds(t *testing.T) {
	for _, tt := range []struct {
		created, read string // the response of create, and of every read of the job (none when empty)
		code          int
		want          string // the end of standard error
	}{
		{created: `{"id":"j1","status":"IN_PROGRESS"}`, read: `{"id":"j1","status":"FAILURE","message":"no index"}`,
			code: ExitError, want: "Error: reindex job j1 ended with FAILURE: no index\n"},
		{created: `{"id":"j1","status":"IN_PROGRESS"}`, read: `{"id":"j1","status":"CANCELLED"}`,
			code: ExitError, want: "Error: reindex job j1 ended with CANCELLED\n"},
		{created: `{"id":"j1","status":"FAILURE","message":"no index"}`, code: ExitError,
			want: "reindex job j1: FAILURE\nError: reindex job j1 ended with FAILURE: no index\n"},
		{created: `{"id":"j1","status":"SUCCESS"}`, code: ExitOK, want: "reindex job j1: SUCCESS\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(reindexJobsPath, respond(http.StatusOK, tt.created))
		fake.route(reindexJobPath, respond(http.StatusOK, tt.read))
		code, stdout, stderr := waitFor(t, fake, "--wait")
		if code != tt.code || stdout != asPrinted(t, cmp.Or(tt.read, tt.created)) || !strings.HasSuffix(stderr, tt.want) {
			t.Errorf("%s then %s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.created, tt.read, code,
				stdout, stderr, tt.code, tt.want)
		}
		if tt.read == "" && len(fake.gets) != 1 {
			t.Errorf("%s: requests %q; want the job ended as created, without reading it", tt.created, fake.gets)
		}
	}
}

// TestReindexJobWaitTimesOut checks that a job that has not ended when
// --wait-timeout passes is printed as last read and fails (exit 1), saying
// how to follow it.
func TestReindexJobWaitTimesOut(t *testing.T) {
	fake := newFakeEmporix(t)
	running := `{"id":"j1","status":"IN_PROGRESS","expectedCount":5,"processedCount":1}`
	fake.route(reindexJobsPath, respond(http.StatusCreated, running))
	fake.route(reindexJobPath, respond(http.StatusOK, running))
	code, stdout, stderr := waitFor(t, fake, "--wait", "--wait-timeout", "20ms")
	want := "Error: reindex job j1 has not ended within 20ms (last IN_PROGRESS); it goes on: " +
		"`epx indexing reindex-jobs get j1` reads it\n"
	if code != ExitError || stdout != asPrinted(t, running) || !strings.HasSuffix(stderr, want) ||
		strings.Count(stderr, "IN_PROGRESS (1 of 5)") != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestReindexJobWaitFailures checks that a response that names no job to
// wait for, and a read of the job that fails or holds another job, end the
// wait, keeping the exit code of the failure; what was read last is printed.
func TestReindexJobWaitFailures(t *testing.T) {
	const noJob = "Error: sent, but its reindex job cannot be waited for: decoding reindex job response: "
	const notJob = `{"id":"j1","status":7}`
	for _, tt := range []struct {
		created string           // the response of create
		read    http.HandlerFunc // the reads of the job
		stdout  string
		code    int
		want    string // the end of standard error
	}{
		{created: `{"status":"PENDING"}`, stdout: "{\n  \"status\": \"PENDING\"\n}\n", code: ExitError,
			want: noJob + "no id or status\n"},
		{created: `{"id":"..","status":"PENDING"}`, stdout: asPrinted(t, `{"id":"..","status":"PENDING"}`),
			code: ExitError, want: noJob + "no id or status\n"},
		{created: ``, code: ExitError, want: noJob + "expected one item as a JSON object\n"},
		{created: notJob, stdout: asPrinted(t, notJob), code: ExitError,
			want: noJob + "json: cannot unmarshal number into Go struct field reindexJob.status of type string\n"},
		{created: `{"id":"j1","status":"PENDING"}`, read: respond(http.StatusNotFound, `{"message":"gone"}`),
			stdout: asPrinted(t, `{"id":"j1","status":"PENDING"}`), code: ExitNotFound,
			want: "Error: waiting for reindex job j1: GET /indexing/acme1/reindex-jobs/j1: 404 Not Found: gone\n"},
		{created: `{"id":"j1","status":"PENDING"}`, read: respond(http.StatusOK, `{"id":"j1"}`),
			stdout: asPrinted(t, `{"id":"j1","status":"PENDING"}`), code: ExitError,
			want: "Error: waiting for reindex job j1: decoding reindex job response: no id or status\n"},
		{created: `{"id":"j1","status":"PENDING"}`, read: respond(http.StatusOK, `{"id":"j2","status":"SUCCESS"}`),
			stdout: asPrinted(t, `{"id":"j1","status":"PENDING"}`), code: ExitError,
			want: "Error: waiting for reindex job j1: a read of it holds reindex job j2 instead\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(reindexJobsPath, respond(http.StatusCreated, tt.created))
		if tt.read != nil {
			fake.route(reindexJobPath, tt.read)
		}
		code, stdout, stderr := waitFor(t, fake, "--wait")
		if code != tt.code || stdout != tt.stdout || !strings.HasSuffix(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.created, code, stdout, stderr,
				tt.code, tt.want)
		}
	}
}

// TestReindexJobWaitStopsOnInterrupt checks that an interrupt, which ends
// the command's context, ends the wait: an error (exit 1).
func TestReindexJobWaitStopsOnInterrupt(t *testing.T) {
	fake := newFakeEmporix(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fake.route(reindexJobsPath, respond(http.StatusCreated, `{"id":"j1","status":"PENDING"}`))
	fake.route(reindexJobPath, func(w http.ResponseWriter, r *http.Request) {
		cancel()
		respond(http.StatusOK, `{"id":"j1","status":"IN_PROGRESS"}`)(w, r)
	})
	code, stdout, stderr := runWrite(ctx, writeRoot(fake, jobCreate()), "create", "--entity-type", "PRODUCT", "--yes",
		"--wait", "--poll-interval", "1ms")
	if code != ExitError || !strings.Contains(stderr, "Error: waiting for reindex job j1: ") ||
		!strings.HasSuffix(stderr, "context canceled\n") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestReindexJobWaitNeedsTheReadScope checks that --wait requests the scope
// of reading the job with that of the write, so a credential without it
// fails (exit 3) before anything is sent.
func TestReindexJobWaitNeedsTheReadScope(t *testing.T) {
	fake := newFakeEmporix(t, "indexing.search_view")
	code, _, stderr := waitFor(t, fake, "--wait")
	if code != ExitAuth || !strings.Contains(stderr, "lacks required scopes: indexing.search_view") ||
		len(fake.gets) != 0 {
		t.Fatalf("exit %d, stderr %q, requests %q; want exit 3 and nothing sent", code, stderr, fake.gets)
	}
}

// TestReindexJobCreateUsageErrors checks that invalid flags of create are
// usage errors (exit 2) that send nothing: the flags of --wait without it or
// with a duration shorter than a second, and a body given by neither or both
// of --entity-type and --file.
func TestReindexJobCreateUsageErrors(t *testing.T) {
	file := writeBody(t, `{"entityType":"PRODUCT"}`)
	const tooShort = "use a duration of 1s or more, such as 30s or 10m"
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"--entity-type", "PRODUCT", "--wait-timeout", "1m"}, want: "--wait-timeout applies only with --wait"},
		{args: []string{"--entity-type", "PRODUCT", "--poll-interval", "1s", "--wait=false"},
			want: "--poll-interval applies only with --wait"},
		{args: []string{"--entity-type", "PRODUCT", "--wait", "--poll-interval", "0"}, want: tooShort},
		{args: []string{"--entity-type", "PRODUCT", "--wait", "--poll-interval", "5ms"}, want: tooShort},
		{args: []string{"--entity-type", "PRODUCT", "--wait", "--wait-timeout", "-1s"}, want: tooShort},
		{args: []string{"--entity-type", "PRODUCT", "--wait", "--wait-timeout", "soon"}, want: tooShort},
		{args: []string{"--rag"}, want: "at least one of the flags in the group [file entity-type] is required"},
		{args: []string{"--entity-type", "PRODUCT", "--file", file}, want: "[file entity-type] are set none of the others"},
		{args: []string{"--file", file, "--rag"}, want: "[file rag] are set none of the others"},
	} {
		fake := newFakeEmporix(t)
		args := append([]string{"indexing", "reindex-jobs", "create", "--yes"}, tt.args...)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

// TestReindexJobCreateSendsRagOnlyWhenGiven checks the body that the typed
// flags of create make: rag only when --rag is given, also as false.
func TestReindexJobCreateSendsRagOnlyWhenGiven(t *testing.T) {
	for flags, want := range map[string]string{
		"":            `{"entityType":"ORDER"}`,
		"--rag":       `{"entityType":"ORDER","rag":true}`,
		"--rag=false": `{"entityType":"ORDER","rag":false}`,
	} {
		fake := newFakeEmporix(t)
		args := append([]string{"indexing", "reindex-jobs", "create", "--entity-type", "ORDER", "--dry-run"},
			strings.Fields(flags)...)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
		if code != ExitOK || !strings.HasSuffix(stdout, "\n\n"+want+"\n") {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want the body %s", flags, code, stdout, stderr, want)
		}
	}
}

// TestReindexJobWaitOutputFailures checks that a job, or a response without
// one, that cannot be printed fails the run, joined to how the wait ended,
// whose exit code it keeps: 4 for a read of the job that was not found.
func TestReindexJobWaitOutputFailures(t *testing.T) {
	for _, tt := range []struct {
		created string           // the response of create
		read    http.HandlerFunc // the reads of the job
		code    int
		want    []string // what standard error says, the output failure last
	}{
		{created: `{"id":"j1","status":"SUCCESS"}`, code: ExitError, want: []string{"Error: disk full\n"}},
		{created: `{"status":"SUCCESS"}`, code: ExitError,
			want: []string{"Error: sent, but its reindex job cannot be waited for: ", "\ndisk full\n"}},
		{created: `{"id":"j1","status":"PENDING"}`, read: respond(http.StatusOK, `{"id":"j1","status":"FAILURE"}`),
			code: ExitError, want: []string{"Error: reindex job j1 ended with FAILURE\ndisk full\n"}},
		{created: `{"id":"j1","status":"PENDING"}`, read: respond(http.StatusNotFound, `{"message":"gone"}`),
			code: ExitNotFound, want: []string{"Error: waiting for reindex job j1: GET /indexing/acme1/reindex-jobs/j1: " +
				"404 Not Found: gone\ndisk full\n"}},
	} {
		fake := newFakeEmporix(t)
		fake.route(reindexJobsPath, respond(http.StatusCreated, tt.created))
		if tt.read != nil {
			fake.route(reindexJobPath, tt.read)
		}
		root := writeRoot(fake, jobCreate())
		root.SetOut(failingWriter{})
		var stderr strings.Builder
		root.SetErr(&stderr)
		root.SetArgs([]string{"create", "--entity-type", "PRODUCT", "--yes", "--wait", "--poll-interval", "1ms"})
		err := root.ExecuteContext(t.Context())
		got := stderr.String()
		if err != nil {
			got += "Error: " + err.Error() + "\n"
		}
		if code := exitCode(err); code != tt.code || !strings.HasSuffix(got, tt.want[len(tt.want)-1]) ||
			!strings.Contains(got, tt.want[0]) {
			t.Errorf("%s: exit %d, stderr %q; want exit %d and %q", tt.created, code, got, tt.code, tt.want)
		}
	}
}

// TestReindexJobCreateHelp checks what the help of create says about --wait
// and its flags.
func TestReindexJobCreateHelp(t *testing.T) {
	_, help, _ := runCLI(t, nil, "indexing", "reindex-jobs", "create", "--help")
	for _, want := range []string{
		"--wait then reads the job again every --poll-interval, as `epx indexing reindex-jobs get`\n" +
			"does and with the indexing.search_view scope too,",
		"exits 0 if it ended SUCCESS, else 1.", "exits as that read\ndoes, e.g. 4 when the job is not found.",
		"\n\nRequires the indexing.search_manage scope.\n",
		"--wait-timeout duration    how long --wait waits, 1s or more (only with --wait) (default 30m0s)",
		"--poll-interval duration   how often --wait reads the reindex job, 1s or more (only with --wait) (default 5s)",
		"--entity-type TYPE", "(required without --file)",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help %q; want %q", help, want)
		}
	}
}

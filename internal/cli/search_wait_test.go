package cli

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// The paths of a search index, of the import of indexes and of two index
// jobs, as the fake serves them.
const (
	searchIndexPath = "/search/acme1/search/vehicle/indexes/vehicles"
	indexImportPath = "/search/acme1/search/indexes/import"
	indexJob1Path   = "/search/acme1/jobs/j1"
	indexJob2Path   = "/search/acme1/jobs/j2"
)

// indexWaitRoot returns a root holding only `indexes` of search, which runs
// against fake, whose --wait may read a job every millisecond, unlike the
// commands', which read it every second at most.
func indexWaitRoot(fake *fakeEmporix) *cobra.Command {
	job := indexJobWait(indexJobs())
	job.shortest = time.Millisecond
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	root := &cobra.Command{Use: "epx", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(indexesCommand(a, job))
	applyExitPolicy(root)
	return root
}

// waitForIndex runs the index write named use, `indexes <use> --yes --wait
// --poll-interval 1ms` and flags, against fake: update and delete of the
// index vehicles of the type vehicle, update with a file that holds it, or
// import, of a package on standard input.
func waitForIndex(t *testing.T, fake *fakeEmporix, use string, flags ...string) (code int, stdout, stderr string) {
	t.Helper()
	root := indexWaitRoot(fake)
	args := []string{"indexes", use, "vehicle", "vehicles"}
	switch use {
	case "update":
		args = append(args, "--file", writeBody(t, `{"fields":[{"path":"make","text":true}]}`))
	case "import":
		args = []string{"indexes", use, "--file", "-"}
		root.SetIn(strings.NewReader(`{"exportedAt":"2026-10-01T08:00:00Z","data":"W10="}`))
	}
	return runWrite(t.Context(), root, slices.Concat(args, []string{"--yes", "--wait", "--poll-interval", "1ms"},
		flags)...)
}

// indexWritePath returns the path that the index write named use sends to.
func indexWritePath(use string) string {
	if use == "import" {
		return indexImportPath
	}
	return searchIndexPath
}

// TestIndexWaitReadsTheJobByItsID checks that --wait of update and delete,
// which Emporix answers with the id of the job alone, reads the job by that
// id, with the read scope too, then again until it ends, saying on standard
// error how its status changes, and prints it as it ended.
func TestIndexWaitReadsTheJobByItsID(t *testing.T) {
	const last = `{"id":"j1","status":"success","type":"update_index","indexId":"vehicles","response":"Index updated."}`
	for _, use := range []string{"update", "delete"} {
		fake := newFakeEmporix(t)
		fake.route(searchIndexPath, respond(http.StatusAccepted, `{"id":"j1"}`))
		fake.route(indexJob1Path, inTurn(respond(http.StatusOK, `{"id":"j1","status":"in_progress"}`),
			respond(http.StatusOK, last)))
		code, stdout, stderr := waitForIndex(t, fake, use)
		if code != ExitOK || stdout != asPrinted(t, last) || stderr != "index job j1: in_progress\nindex job j1: success\n" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", use, code, stdout, stderr)
		}
		fake.assertScopes(t, "search.search_manage", "search.search_read")
		fake.assertRequests(t, searchIndexPath, indexJob1Path, indexJob1Path)
	}
}

// TestIndexWaitWithoutAJob checks that --wait of a write that started no
// job, an update that Emporix answers with 204 No Content and an import whose
// indexes name none, says so and succeeds, reading no job: it prints the
// response of update, empty, and the jobs of import, none.
func TestIndexWaitWithoutAJob(t *testing.T) {
	const unchanged = `[{"id":"bikes","type":"bike"},{"id":"cars","type":"car","jobId":null}]`
	const noIndexJob = "Emporix started no index job, so there is none to wait for.\n"
	for _, tt := range []struct {
		use    string
		answer http.HandlerFunc
		stdout string
	}{
		{use: "update", answer: respond(http.StatusNoContent, "")},
		{use: "import", answer: respond(http.StatusOK, unchanged), stdout: "[]\n"},
		{use: "import", answer: respond(http.StatusOK, "[]"), stdout: "[]\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(indexWritePath(tt.use), tt.answer)
		code, stdout, stderr := waitForIndex(t, fake, tt.use)
		if code != ExitOK || stdout != tt.stdout || stderr != noIndexJob {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", tt.use, code, stdout, stderr)
		}
		fake.assertRequests(t, indexWritePath(tt.use))
	}
}

// The jobs of an import as the fake reads them: j1 running, then ended with
// success or failure, j2 ended with success, and a read of a job not found.
const (
	indexJob1Running = `{"id":"j1","status":"in_progress"}`
	indexJob1Success = `{"id":"j1","status":"success","indexId":"vehicles"}`
	indexJob1Failure = `{"id":"j1","status":"failure","indexId":"vehicles","response":"make is not allowed"}`
	indexJob2Success = `{"id":"j2","status":"success","indexId":"cars"}`
	indexJobGone     = `{"message":"gone"}`
)

// TestIndexWaitFollowsEachImportedJob checks that --wait of import follows
// each job that its result per index names, in turn, naming each by the
// TYPE/ID of its index, and prints them as a JSON array, as last read,
// leaving out those never read: it succeeds only when each ended with
// success, and fails naming each that did not, still following the jobs
// after it, with the exit code that their failures share, else 1.
func TestIndexWaitFollowsEachImportedJob(t *testing.T) {
	const imported = `[{"id":"vehicles","type":"vehicle","jobId":"j1","jobType":"create_index"},` +
		`{"id":"bikes","type":"bike"},{"id":"cars","type":"car","jobId":"j2","jobType":"update_index"}]`
	ended := func(job string) http.HandlerFunc {
		return inTurn(respond(http.StatusOK, indexJob1Running), respond(http.StatusOK, job))
	}
	gone := respond(http.StatusNotFound, indexJobGone)
	const (
		failed     = "index job j1 of vehicle/vehicles ended with failure: make is not allowed\n"
		j1Gone     = "waiting for index job j1 of vehicle/vehicles: GET " + indexJob1Path + ": 404 Not Found: gone\n"
		j2Gone     = "waiting for index job j2 of car/cars: GET " + indexJob2Path + ": 404 Not Found: gone\n"
		j2Succeeds = "index job j2 of car/cars: success\n"
	)
	for _, tt := range []struct {
		name   string
		j1, j2 http.HandlerFunc // the reads of each job
		code   int
		stdout string // the jobs printed, as JSON
		end    string // the end of standard error
	}{
		{name: "each succeeds", j1: ended(indexJob1Success), j2: respond(http.StatusOK, indexJob2Success),
			code: ExitOK, stdout: "[" + indexJob1Success + "," + indexJob2Success + "]",
			end: "index job j1 of vehicle/vehicles: in_progress\nindex job j1 of vehicle/vehicles: success\n" + j2Succeeds},
		{name: "one fails", j1: ended(indexJob1Failure), j2: respond(http.StatusOK, indexJob2Success),
			code: ExitError, stdout: "[" + indexJob1Failure + "," + indexJob2Success + "]",
			end: "index job j1 of vehicle/vehicles: failure\n" + j2Succeeds + "Error: " + failed},
		{name: "one is not found", j1: gone, j2: respond(http.StatusOK, indexJob2Success), code: ExitNotFound,
			stdout: "[" + indexJob2Success + "]", end: j2Succeeds + "Error: " + j1Gone},
		{name: "each is not found", j1: gone, j2: gone, code: ExitNotFound, stdout: "[]",
			end: "Error: " + j1Gone + j2Gone},
		{name: "they fail differently", j1: ended(indexJob1Failure), j2: gone, code: ExitError,
			stdout: "[" + indexJob1Failure + "]", end: "index job j1 of vehicle/vehicles: failure\nError: " + failed + j2Gone},
	} {
		fake := newFakeEmporix(t)
		fake.route(indexImportPath, respond(http.StatusOK, imported))
		fake.route(indexJob1Path, tt.j1)
		fake.route(indexJob2Path, tt.j2)
		code, stdout, stderr := waitForIndex(t, fake, "import")
		if code != tt.code || stdout != asPrinted(t, tt.stdout) || !strings.HasSuffix(stderr, tt.end) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.name, code, stdout, stderr, tt.code,
				tt.end)
		}
		if gets := fake.gets; !slices.Equal(gets[:2], []string{indexImportPath, indexJob1Path}) ||
			gets[len(gets)-1] != indexJob2Path {
			t.Errorf("%s: requests %q; want the import, then j1 until it ended, then j2", tt.name, fake.gets)
		}
	}
}

// TestIndexWaitTimesOut checks that a job that has not ended when
// --wait-timeout passes is printed as last read and fails (exit 1), saying
// how to follow it, as does each job of an import after it, which it says
// was never read.
func TestIndexWaitTimesOut(t *testing.T) {
	const notEnded = "index job j1%s has not ended within 20ms (last in_progress); it goes on: " +
		"`epx search jobs get j1` reads it\n"
	for _, tt := range []struct {
		use, answer   string
		stdout, wants string // standard output, as JSON, and the end of standard error
	}{
		{use: "update", answer: `{"id":"j1"}`, stdout: indexJob1Running, wants: "Error: " + fmt.Sprintf(notEnded, "")},
		{use: "import",
			answer: `[{"id":"vehicles","type":"vehicle","jobId":"j1"},{"id":"cars","type":"car","jobId":"j2"}]`,
			stdout: "[" + indexJob1Running + "]", wants: "Error: " + fmt.Sprintf(notEnded, " of vehicle/vehicles") +
				"index job j2 of car/cars was not read before --wait-timeout (20ms) passed: " +
				"`epx search jobs get j2` reads it\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(indexWritePath(tt.use), respond(http.StatusOK, tt.answer))
		fake.route(indexJob1Path, respond(http.StatusOK, indexJob1Running))
		code, stdout, stderr := waitForIndex(t, fake, tt.use, "--wait-timeout", "20ms")
		if code != ExitError || stdout != asPrinted(t, tt.stdout) || !strings.HasSuffix(stderr, tt.wants) ||
			strings.Count(stderr, ": in_progress\n") != 1 {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", tt.use, code, stdout, stderr)
		}
	}
}

// TestIndexWaitFailures checks that an answer that does not name the jobs as
// its endpoint does, and a first read of the job that fails or holds another
// job, end the wait, keeping the exit code of the failure, and print the
// response as sent.
func TestIndexWaitFailures(t *testing.T) {
	const cannot = "Error: sent, but its index job cannot be waited for: decoding the response: "
	for _, tt := range []struct {
		use    string
		answer string
		read   http.HandlerFunc // the reads of the job j1
		code   int
		want   string // the end of standard error
	}{
		{use: "update", answer: `{"id":7}`, code: ExitError, want: cannot + "id is not the ID of a job\n"},
		{use: "update", answer: `{"message":"accepted"}`, code: ExitError, want: cannot + "id is not the ID of a job\n"},
		{use: "delete", answer: `{"id":null}`, code: ExitError, want: cannot + "id is not the ID of a job\n"},
		{use: "delete", answer: `{"id":".."}`, code: ExitError, want: cannot + "id is not the ID of a job\n"},
		{use: "delete", answer: `"j1"`, code: ExitError, want: cannot + "expected a JSON object\n"},
		{use: "import", answer: `{"jobId":"j1"}`, code: ExitError, want: cannot + "expected a JSON array\n"},
		{use: "import", answer: `["j1"]`, code: ExitError, want: cannot + "expected a JSON object\n"},
		{use: "import", answer: `[{}]`, code: ExitError, want: cannot + "an item has no type\n"},
		{use: "import", answer: `[{"type":"bike","id":""}]`, code: ExitError, want: cannot + "an item has no id\n"},
		{use: "import", answer: `[{"type":"bike","id":"bikes","jobId":7}]`, code: ExitError,
			want: cannot + "jobId is not the ID of a job\n"},
		{use: "update", answer: `{"id":"j1"}`, read: respond(http.StatusNotFound, indexJobGone),
			code: ExitNotFound, want: "Error: waiting for index job j1: GET /search/acme1/jobs/j1: 404 Not Found: gone\n"},
		{use: "delete", answer: `{"id":"j1"}`, read: respond(http.StatusOK, `{"id":"j1"}`), code: ExitError,
			want: "Error: waiting for index job j1: decoding index job response: no id or status\n"},
		{use: "delete", answer: `{"id":"j1"}`, read: respond(http.StatusOK, `{"id":"j2","status":"success"}`),
			code: ExitError, want: "Error: waiting for index job j1: a read of it holds index job j2 instead\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(indexWritePath(tt.use), respond(http.StatusAccepted, tt.answer))
		if tt.read != nil {
			fake.route(indexJob1Path, tt.read)
		}
		code, stdout, stderr := waitForIndex(t, fake, tt.use)
		if code != tt.code || stdout != asPrinted(t, tt.answer) || !strings.HasSuffix(stderr, tt.want) {
			t.Errorf("%s answered %s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.use, tt.answer, code,
				stdout, stderr, tt.code, tt.want)
		}
	}
}

// TestIndexWaitNeedsTheReadScope checks that --wait requests the scope of
// reading index jobs with that of the write, so a credential without it
// fails (exit 3) before anything is sent.
func TestIndexWaitNeedsTheReadScope(t *testing.T) {
	fake := newFakeEmporix(t, "search.search_read")
	code, _, stderr := waitForIndex(t, fake, "delete")
	if code != ExitAuth || !strings.Contains(stderr, "lacks required scopes: search.search_read") || len(fake.gets) != 0 {
		t.Fatalf("exit %d, stderr %q, requests %q; want exit 3 and nothing sent", code, stderr, fake.gets)
	}
}

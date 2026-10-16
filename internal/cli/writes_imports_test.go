package cli

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// queuedRun is an import run as Emporix answers a write that starts one.
const queuedRun = `{"id":"r2","configId":"c1","status":"QUEUED","mode":"DELTA","queuePosition":0}`

// queuedRunJSON is queuedRun as epx prints it.
const queuedRunJSON = "{\n  \"configId\": \"c1\",\n  \"id\": \"r2\",\n  \"mode\": \"DELTA\",\n" +
	"  \"queuePosition\": 0,\n  \"status\": \"QUEUED\"\n}\n"

// cancelUnknown is the answer of a cancel that Emporix gives an unknown or
// ended run (409 Conflict, live), and cancelUnknownShown what epx prints for it.
const (
	cancelUnknown      = `{"accepted":false,"cancelRequested":false,"force":false,"runId":"r1"}`
	cancelUnknownShown = "{\n  \"accepted\": false,\n  \"cancelRequested\": false,\n" +
		"  \"force\": false,\n  \"runId\": \"r1\"\n}\n"
)

// importWriteCases are the writes of the Import Service: runs started,
// retried and cancelled, and a configuration's schedule set and deleted.
func importWriteCases() []writeCase {
	trigger := []string{"importtool.import_trigger"}
	const configuration, run = "/importtool/acme1/configs/c1", "/importtool/acme1/runs/r1"
	const schedule = `{"cron":"0 0 2 * * *","enabled":true,"timezone":"Europe/Berlin"}`
	return []writeCase{
		{
			args: []string{"imports", "configurations", "trigger", "c1"}, sent: `{}`, method: http.MethodPost,
			path: configuration + "/runs", scopes: trigger, status: http.StatusOK, body: queuedRun, want: queuedRunJSON,
			conflict: activeRunConflict,
		},
		{
			args: []string{"imports", "configurations", "trigger", "c1", "--mode", "FULL"}, sent: `{"mode":"FULL"}`,
			method: http.MethodPost, path: configuration + "/runs", scopes: trigger, status: http.StatusOK, body: queuedRun,
			want: queuedRunJSON, conflict: activeRunConflict,
		},
		{
			args: []string{"imports", "configurations", "trigger", "c1"}, file: `{"mode":"FULL","dryRun":true}`,
			method: http.MethodPost, path: configuration + "/runs", scopes: trigger, status: http.StatusOK, body: queuedRun,
			want: queuedRunJSON, conflict: activeRunConflict,
		},
		{
			args: []string{"imports", "configurations", "update-schedule", "c1", "--cron", "0 0 2 * * *", "--timezone",
				"Europe/Berlin"},
			sent: schedule, method: http.MethodPut, path: configuration + "/schedule", scopes: trigger,
			status: http.StatusOK, body: schedule,
			want: "{\n  \"cron\": \"0 0 2 * * *\",\n  \"enabled\": true,\n  \"timezone\": \"Europe/Berlin\"\n}\n",
		},
		{
			args: []string{"imports", "configurations", "update-schedule", "c1", "--cron", "0 0 2 * * *", "--enabled=false"},
			sent: `{"cron":"0 0 2 * * *","enabled":false}`, method: http.MethodPut, path: configuration + "/schedule",
			scopes: trigger, status: http.StatusOK, body: `{"cron":"0 0 2 * * *","enabled":false}`,
			want: "{\n  \"cron\": \"0 0 2 * * *\",\n  \"enabled\": false\n}\n",
		},
		{
			args: []string{"imports", "configurations", "update-schedule", "c1"}, file: schedule, method: http.MethodPut,
			path: configuration + "/schedule", scopes: trigger, status: http.StatusOK, body: schedule,
			want: "{\n  \"cron\": \"0 0 2 * * *\",\n  \"enabled\": true,\n  \"timezone\": \"Europe/Berlin\"\n}\n",
		},
		{
			args: []string{"imports", "configurations", "delete-schedule", "c1"}, method: http.MethodDelete,
			path: configuration + "/schedule", scopes: trigger, status: http.StatusNoContent, confirmBy: "c1",
		},
		{
			args: []string{"imports", "runs", "cancel", "r1"}, method: http.MethodPost, path: run + "/cancel",
			query: "force=false", scopes: trigger, status: http.StatusAccepted,
			body:     `{"runId":"r1","force":false,"accepted":true}`,
			want:     "{\n  \"accepted\": true,\n  \"force\": false,\n  \"runId\": \"r1\"\n}\n",
			conflict: cancelConflict, conflictAnswer: cancelUnknown, conflictShown: cancelUnknownShown,
		},
		{
			args: []string{"imports", "runs", "cancel", "r1", "--force"}, method: http.MethodPost, path: run + "/cancel",
			query: "force=true", scopes: trigger, status: http.StatusAccepted,
			body:     `{"runId":"r1","force":true,"accepted":false}`,
			want:     "{\n  \"accepted\": false,\n  \"force\": true,\n  \"runId\": \"r1\"\n}\n",
			conflict: cancelConflict, conflictAnswer: cancelUnknown, conflictShown: cancelUnknownShown,
		},
		{
			args: []string{"imports", "runs", "retry", "r1"}, method: http.MethodPost, path: run + "/retry",
			scopes: trigger, status: http.StatusOK, body: queuedRun, want: queuedRunJSON, conflict: activeRunConflict,
		},
	}
}

// runWaits returns trigger and retry, which start a run, whose --wait may
// read the run every millisecond, unlike the commands', which read it every
// second at most.
func runWaits() (trigger, retry *write) {
	runs := importRuns("imports configurations runs CONFIGURATION_ID")
	job := importRunJob(runs)
	job.shortest = time.Millisecond
	return importTrigger(importConfigurations(), job), importRetry(runs, job)
}

// TestImportRunWaitFollowsTheRun checks that --wait reads the run that
// trigger or retry started, which Emporix answers with the run itself, as
// `imports runs get` reads it, in its details, until it ends, saying on
// standard error how its status changes, and that it succeeds only for a run
// that ended SUCCEEDED, naming the status and message of another.
func TestImportRunWaitFollowsTheRun(t *testing.T) {
	const running = `{"run":{"id":"r2","status":"RUNNING","recordsRead":40},"streams":[]}`
	trigger, retry := runWaits()
	for _, tt := range []struct {
		write      *write
		args       []string
		post, last string // the path of the write and the details of the run as it ended
		code       int
		end        string // the last line of standard error
	}{
		{
			write: trigger, args: []string{"trigger", "c1"}, post: "/importtool/acme1/configs/c1/runs",
			last: `{"run":{"id":"r2","status":"SUCCEEDED","recordsRead":90},"streams":[{"streamName":"products"}]}`,
			code: ExitOK, end: "run r2: SUCCEEDED (90 read)\n",
		},
		{
			write: retry, args: []string{"retry", "r1"}, post: "/importtool/acme1/runs/r1/retry",
			last: `{"run":{"id":"r2","status":"PARTIAL","recordsRead":90,"message":"3 records failed"}}`,
			code: ExitError, end: "Error: run r2 ended with PARTIAL: 3 records failed\n",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.post, respond(http.StatusOK, queuedRun))
		fake.route("/importtool/acme1/runs/r2", inTurn(respond(http.StatusOK, running), respond(http.StatusOK, tt.last)))
		code, stdout, stderr := runWrite(t.Context(), writeRoot(fake, tt.write),
			append(tt.args, "--yes", "--wait", "--poll-interval", "1ms")...)
		progress := "run r2: QUEUED\nrun r2: RUNNING (40 read)\n"
		if code != tt.code || stdout != asPrinted(t, tt.last) || !strings.HasPrefix(stderr, progress) ||
			!strings.HasSuffix(stderr, tt.end) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q", tt.args, code, stdout, stderr)
		}
		fake.assertScopes(t, "importtool.import_trigger")
		fake.assertRequests(t, tt.post, "/importtool/acme1/runs/r2", "/importtool/acme1/runs/r2")
	}
}

// TestImportRunWaitNeedsARun checks that --wait fails, printing the response
// as sent, when the run that Emporix answers with has no id or status, also
// when it answers with the run's details, which only a read of a run holds.
func TestImportRunWaitNeedsARun(t *testing.T) {
	trigger, _ := runWaits()
	for _, answer := range []string{`{"status":"QUEUED"}`, `{"run":` + queuedRun + `,"streams":[]}`} {
		fake := newFakeEmporix(t)
		fake.route("/importtool/acme1/configs/c1/runs", respond(http.StatusOK, answer))
		code, stdout, stderr := runWrite(t.Context(), writeRoot(fake, trigger), "trigger", "c1", "--yes", "--wait")
		if code != ExitError || stdout != asPrinted(t, answer) ||
			!strings.Contains(stderr, "sent, but its run cannot be waited for: decoding run response: no id or status") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", answer, code, stdout, stderr)
		}
		fake.assertRequests(t, "/importtool/acme1/configs/c1/runs")
	}
}

// TestImportWriteFlags checks that a --mode other than FULL or DELTA, a
// --mode with --file and update-schedule without --cron are usage errors
// (exit 2) that send nothing.
func TestImportWriteFlags(t *testing.T) {
	file := writeBody(t, `{"mode":"FULL"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"configurations", "trigger", "c1", "--mode", "full"}, want: "use FULL or DELTA"},
		{args: []string{"configurations", "trigger", "c1", "--mode", "FULL", "--file", file}, want: "[file mode] were"},
		{args: []string{"configurations", "update-schedule", "c1"}, want: "the flags in the group [file cron] is required"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(append([]string{"imports"}, tt.args...),
			"--yes")...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

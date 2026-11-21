package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

const queuedRun = `{"id":"r2","configId":"c1","status":"QUEUED","mode":"DELTA","queuePosition":0}`

const queuedRunJSON = "{\n  \"configId\": \"c1\",\n  \"id\": \"r2\",\n  \"mode\": \"DELTA\",\n" +
	"  \"queuePosition\": 0,\n  \"status\": \"QUEUED\"\n}\n"

const (
	cancelUnknown      = `{"accepted":false,"cancelRequested":false,"force":false,"runId":"r1"}`
	cancelUnknownShown = "{\n  \"accepted\": false,\n  \"cancelRequested\": false,\n" +
		"  \"force\": false,\n  \"runId\": \"r1\"\n}\n"
)

func importWriteCases() []writeCase {
	trigger := []string{"importtool.import_trigger"}
	const configuration, run = "/importtool/acme1/configs/c1", "/importtool/acme1/runs/r1"
	const schedule = `{"cron":"0 0 2 * * *","enabled":true,"timezone":"Europe/Berlin"}`
	return slices.Concat([]writeCase{
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
			scopes: trigger, status: http.StatusOK, body: queuedRun, want: queuedRunJSON, conflict: retryConflict,
		},
	}, importFileWriteCases(trigger))
}

func importFileWriteCases(trigger []string) []writeCase {
	const order = `["f3","f2"]`
	reordered := `[{"id":"f3","status":"PENDING","loadRank":1},{"id":"f2","status":"PENDING","loadRank":2}]`
	reorderedJSON := "[\n  {\n    \"id\": \"f3\",\n    \"loadRank\": 1,\n    \"status\": \"PENDING\"\n  },\n" +
		"  {\n    \"id\": \"f2\",\n    \"loadRank\": 2,\n    \"status\": \"PENDING\"\n  }\n]\n"
	return []writeCase{
		{
			args:   []string{"imports", "streams", "files", "upload", "s1", "--upload", "testdata/order.csv"},
			form:   []sentPart{{name: "file", file: "order.csv", contentType: "text/csv", content: orderCSV()}},
			method: http.MethodPost, path: filesPath, scopes: trigger, status: http.StatusOK,
			body: `{"id":"f2","fileName":"order.csv","size":24,"status":"PENDING","loadRank":1}`,
			want: "{\n  \"fileName\": \"order.csv\",\n  \"id\": \"f2\",\n  \"loadRank\": 1,\n  \"size\": 24,\n" +
				"  \"status\": \"PENDING\"\n}\n",
			conflict: uploadConflict(filesList, fileModeShown),
		},
		{
			args: []string{"imports", "streams", "files", "order", "s1", "--id", "f3", "--id", "f2"}, sent: order,
			method: http.MethodPut, path: filesPath + "/order", scopes: trigger, status: http.StatusOK, body: reordered,
			want: reorderedJSON, conflict: fileOrderConflict(filesList),
		},
		{
			args: []string{"imports", "streams", "files", "order", "s1"}, file: order, method: http.MethodPut,
			path: filesPath + "/order", scopes: trigger, status: http.StatusOK, body: reordered, want: reorderedJSON,
			conflict: fileOrderConflict(filesList),
		},
		{
			args: []string{"imports", "streams", "files", "delete", "s1", "f2"}, method: http.MethodDelete,
			path: filesPath + "/f2", scopes: trigger, status: http.StatusNoContent, confirmBy: "f2",
			conflict: fileDeleteConflict(filesList),
		},
		{
			args: []string{"imports", "streams", "files", "requeue", "s1", "f2"}, method: http.MethodPost,
			path: filesPath + "/f2/requeue", scopes: trigger, status: http.StatusOK,
			body: `{"id":"f2","fileName":"order.csv","status":"PENDING","loadRank":3}`,
			want: "{\n  \"fileName\": \"order.csv\",\n  \"id\": \"f2\",\n  \"loadRank\": 3,\n" +
				"  \"status\": \"PENDING\"\n}\n",
			conflict: fileRequeueConflict(filesList),
		},
	}
}

func runWaits() (trigger, retry *write) {
	runs := importRuns("imports configurations runs CONFIGURATION_ID")
	job := importRunJob(runs)
	job.shortest = time.Millisecond
	return importTrigger(importConfigurations(), job), importRetry(runs, job)
}

func TestImportRunWaitFollowsTheRun(t *testing.T) {
	const running = `{"run":{"id":"r2","status":"RUNNING","recordsRead":40},"streams":[]}`
	trigger, retry := runWaits()
	for _, tt := range []struct {
		write      *write
		args       []string
		post, last string
		code       int
		end        string
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

func TestImportRunRefusals(t *testing.T) {
	const runs = "/importtool/acme1/configs/c1/runs"
	for _, tt := range []struct {
		status int
		body   string
		code   int
		want   string
	}{
		{
			status: http.StatusConflict, code: ExitConflict,
			body: `{"resourceId":"c1","status":"Conflict","code":409,"message":"An import is already running for config c1",` +
				`"details":[{"message":"Active run: r9","field":"activeRunId"}]}`,
			want: "POST " + runs + ": 409 Conflict: An import is already running for config c1 (activeRunId: Active " +
				"run: r9)",
		},
		{
			status: http.StatusBadRequest, code: ExitError,
			body: `{"resourceId":null,"status":"Bad Request","code":400,"message":"Nothing to import: no file has been ` +
				`uploaded to 'Vendors'","details":[]}`,
			want: "POST " + runs + ": 400 Bad Request: Nothing to import: no file has been uploaded to 'Vendors'",
		},
		{
			status: http.StatusNotFound, code: ExitNotFound,
			body: `{"resourceId":null,"status":"Not Found","code":404,"message":"config not found","details":[]}`,
			want: "POST " + runs + ": 404 Not Found: config not found",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(runs, respond(tt.status, tt.body))
		code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "imports", "configurations", "trigger", "c1", "--yes")
		if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%d: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.status, code, stdout, stderr, tt.code,
				tt.want)
		}
	}
}

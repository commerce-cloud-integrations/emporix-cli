package cli

import (
	"net/http"
	"strings"
	"testing"
)

// runImports runs `epx imports args… --allow-write-scope` against fake.
func runImports(t *testing.T, fake *fakeEmporix, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	return runCLI(t, fake.env(), append(append([]string{"imports"}, args...), "--"+writeScopeFlag)...)
}

// TestImportScheduleWithoutOneIsNotFound checks that a configuration without
// a schedule, which the service answers with 204 No Content, is not found.
func TestImportScheduleWithoutOneIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/importtool/acme1/configs/c1/schedule", respond(http.StatusNoContent, ""))
	code, stdout, stderr := runImports(t, fake, "configurations", "schedule", "c1")
	if code != ExitNotFound || stdout != "" ||
		!strings.Contains(stderr, "GET /importtool/acme1/configs/c1/schedule: empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q; want not found", code, stdout, stderr)
	}
}

// TestImportShapesRejectMalformedResponses checks the reads whose items the
// response holds in a member, or which restructure it: a response of another
// shape fails (exit 1) and prints nothing, also with -o json.
func TestImportShapesRejectMalformedResponses(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, want string
		bodies     []string
	}{
		{
			args: []string{"configurations", "stream-order", "c1"}, path: "/importtool/acme1/configs/c1/stream-order",
			want:   "decoding stream order response: expected an object with the streams in order",
			bodies: []string{`[]`, `null`, `{}`, `{"order":null}`, `{"order":"vendor"}`, `{"order":["a"],"prereqs":{"a":"b"}}`},
		},
		{
			args: []string{"runs", "get", "r1"}, path: "/importtool/acme1/runs/r1",
			want: "decoding run response: expected an object with a run and its streams",
			bodies: []string{
				`[]`, `null`, `{"streams":[]}`, `{"run":{"id":"r1"},"streams":{}}`, `{"run":null}`, `{"run":{}}`,
				`{"run":[]}`,
			},
		},
		{
			// Streams without their run are malformed too, not a run without progress.
			args: []string{"runs", "streams", "r1"}, path: "/importtool/acme1/runs/r1",
			want: "decoding run stream response: expected an object with a run and its streams",
			bodies: []string{
				`[]`, `null`, `{"streams":[]}`, `{"run":{"id":"r1"},"streams":"none"}`, `{"run":null,"streams":[]}`,
				`{"run":{},"streams":[]}`, `{"run":[],"streams":[]}`,
			},
		},
		{
			args: []string{"runs", "diagnostics", "r1"}, path: "/importtool/acme1/runs/r1/diagnostics",
			want: "decoding run diagnostics response: expected an object with the diagnostic rows",
			bodies: []string{`[]`, `{"returned":0}`, `{"rows":null}`, `{"rows":{}}`, `{"rows":[],"recorded":"0"}`,
				`{"rows":[],"recorded":1.5}`},
		},
		{
			args: []string{"health-thresholds"}, path: "/importtool/acme1/settings/health-thresholds",
			want:   "decoding health thresholds response: expected a JSON object",
			bodies: []string{`[]`, `null`, `"none"`},
		},
	} {
		for _, body := range tt.bodies {
			for _, output := range []string{"table", "json"} {
				fake := newFakeEmporix(t)
				fake.route(tt.path, respond(http.StatusOK, body))
				code, stdout, stderr := runImports(t, fake, append(tt.args, "-o", output)...)
				if code != ExitError || stdout != "" || !strings.Contains(stderr, tt.want) {
					t.Errorf("%q -o %s, body %s: exit %d, stdout %q, stderr %q; want %q",
						tt.args, output, body, code, stdout, stderr, tt.want)
				}
			}
		}
	}
}

// TestImportDiagnosticsFailWithoutEveryRecordedRow checks that diagnostics
// fail (exit 1), also with -o json, when the service recorded more rows than
// it returned: a run that recorded more than epx asks for.
func TestImportDiagnosticsFailWithoutEveryRecordedRow(t *testing.T) {
	for _, output := range []string{"table", "json"} {
		fake := newFakeEmporix(t)
		fake.route("/importtool/acme1/runs/r1/diagnostics", respond(http.StatusOK,
			`{"rows":[{"kind":"REPEATED_KEY","naturalKey":"100200"}],"returned":1,"recorded":2}`))
		code, stdout, stderr := runImports(t, fake, "runs", "diagnostics", "r1", "-o", output)
		const want = "decoding run diagnostics response: the run recorded 2 rows, more than the 1 returned; " +
			"read them per stream, as the help says"
		if code != ExitError || stdout != "" || !strings.Contains(stderr, want) {
			t.Errorf("-o %s: exit %d, stdout %q, stderr %q; want %q", output, code, stdout, stderr, want)
		}
	}
}

// TestImportHelpSaysWhereIDsComeFrom checks that the reads of a run or stream,
// which have no list, take their ID from the reads of a configuration that
// list them, and that a run's columns are explained wherever runs are shown.
func TestImportHelpSaysWhereIDsComeFrom(t *testing.T) {
	const runs = "`epx imports configurations runs CONFIGURATION_ID` shows"
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: []string{"runs", "get"}, want: []string{"a run's ID is the one that " + runs, "STATUS is QUEUED"}},
		{args: []string{"runs", "errors"}, want: []string{"RUN_ID is the ID that " + runs + "; an unknown one lists"}},
		{args: []string{"runs", "streams"}, want: []string{"RUN_ID is the ID that " + runs + "; an unknown one is not"}},
		{args: []string{"runs", "diagnostics"}, want: []string{"RUN_ID is the ID that " + runs, "up to 50000"}},
		{args: []string{"configurations", "runs"}, want: []string{"most recent first", "STATUS is QUEUED"}},
		{args: []string{"streams", "get"}, want: []string{
			"A stream's ID is the one that `epx imports configurations streams CONFIGURATION_ID` shows."}},
		{args: []string{"streams", "records"}, want: []string{
			"STREAM_ID is the ID that `epx imports configurations streams CONFIGURATION_ID` shows; an unknown one lists",
		}},
	} {
		code, help, _ := runCLI(t, nil, append(append([]string{"imports"}, tt.args...), "--help")...)
		for _, want := range tt.want {
			if code != ExitOK || !strings.Contains(help, want) {
				t.Errorf("%q: exit %d, help:\n%s\nwant %q in it", tt.args, code, help, want)
			}
		}
	}
}

// TestImportReadsOfPartialResponses checks what the import reads show for
// responses that leave values out: a queued run has no streams yet, a run
// without diagnostics none to list, a tenant without thresholds of its own
// falls back to the defaults, and statistics and limits without their first
// value still show their row.
func TestImportReadsOfPartialResponses(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, body string
		want       string
	}{
		{args: []string{"runs", "streams", "r1"}, path: "/importtool/acme1/runs/r1",
			body: `{"run":{"id":"r1","status":"QUEUED"}}`,
			want: "STREAM  STATUS  READ  CREATED  UPDATED  SKIPPED  DELETED  FAILED  MESSAGE\n"},
		{args: []string{"runs", "diagnostics", "r1"}, path: "/importtool/acme1/runs/r1/diagnostics",
			body: `{"rows":[],"returned":0,"recorded":0}`,
			want: "KIND  STREAM  KEY  LINKING FIELD  LINKING VALUE  WAITING LINES  OCCURRENCE\n"},
		{args: []string{"configurations", "stream-order", "c1"}, path: "/importtool/acme1/configs/c1/stream-order",
			body: `{"order":["vendor"]}`, want: "STREAM  AFTER\nvendor  \n"},
		{args: []string{"health-thresholds"}, path: "/importtool/acme1/settings/health-thresholds",
			body: `{"thresholds":null,"builtInDefaults":{"failedRunIsRed":false}}`,
			want: "LEVEL             FAILURE RATIO AMBER  FAILURE RATIO RED  STALE HOURS AMBER  STALE HOURS RED  " +
				"FAILED RUN RED\ntenant" + strings.Repeat(" ", 88) + "\nbuilt-in default" + strings.Repeat(" ", 78) + "false\n"},
		{args: []string{"statistics"}, path: "/importtool/acme1/stats", body: `{"summary":null}`,
			want: "RUNS          SUCCEEDED  PARTIAL  FAILED  CANCELLED  ACTIVE  RECORDS READ  RECORDS FAILED\n" +
				"not reported" + strings.Repeat(" ", 63) + "\n"},
		{args: []string{"license"}, path: "/importtool/acme1/license", body: `{"workers":1}`,
			want: "MAX RECORDS PER TYPE  MAX CONCURRENT RUNS  BATCH SIZE  WORKERS\n" +
				"not reported" + strings.Repeat(" ", 43) + "1\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(http.StatusOK, tt.body))
		code, stdout, stderr := runImports(t, fake, tt.args...)
		if code != ExitOK || stdout != tt.want {
			t.Errorf("%q: exit %d, stderr %q, stdout:\n%s\nwant:\n%s", tt.args, code, stderr, stdout, tt.want)
		}
	}
}

// TestImportRunJSONKeepsItsStreams checks that -o json prints a run's whole
// details, the progress of its streams included, which its table leaves out.
func TestImportRunJSONKeepsItsStreams(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/importtool/acme1/runs/r1", respond(http.StatusOK,
		`{"run":{"id":"r1","status":"RUNNING"},"streams":[{"streamName":"Products","recordsRead":120}]}`))
	code, stdout, stderr := runImports(t, fake, "runs", "get", "r1", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"streamName": "Products"`) || !strings.Contains(stdout, `"run": {`) {
		t.Fatalf("exit %d, stderr %q, stdout:\n%s", code, stderr, stdout)
	}
}

// TestImportRecordsArePagedFromZero checks that imported records are read in
// pages numbered from 0 until the reported total, the type kept on each.
func TestImportRecordsArePagedFromZero(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/importtool/acme1/data/records", func(w http.ResponseWriter, r *http.Request) {
		items := map[string]string{"0": `{"id":"d1","naturalKey":"100200"}`, "1": `{"id":"d2","naturalKey":"100201"}`}
		respond(http.StatusOK, `{"content":[`+items[r.URL.Query().Get("page")]+`],"totalElements":2}`)(w, r)
	})
	code, stdout, stderr := runImports(t, fake, "records", "product")
	want := "ID  KEY     TYPE  OUTCOME  IMPORTED\nd1  100200                 \nd2  100201                 \n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, want)
	}
	if len(fake.gets) != 2 || !strings.Contains(fake.gets[1], "page=1&size=100&type=product") {
		t.Fatalf("requests %q; want pages 0 and 1 of type product", fake.gets)
	}
}

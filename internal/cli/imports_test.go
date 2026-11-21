package cli

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func runImports(t *testing.T, fake *fakeEmporix, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	return runCLI(t, fake.env(), append(append([]string{"imports"}, args...), "--"+writeScopeFlag)...)
}

func TestImportScheduleWithoutOneIsNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/importtool/acme1/configs/c1/schedule", respond(http.StatusNoContent, ""))
	code, stdout, stderr := runImports(t, fake, "configurations", "schedule", "c1")
	if code != ExitNotFound || stdout != "" ||
		!strings.Contains(stderr, "GET /importtool/acme1/configs/c1/schedule: empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q; want not found", code, stdout, stderr)
	}
}

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
		{
			args: []string{"streams", "files", "list", "s1"}, path: filesPath,
			want: "decoding stream file response: expected an object with the pending files and the history, and " +
				"an integer pendingTotal and a boolean pendingTruncated, if any",
			bodies: []string{
				`[]`, `null`, `{}`, `{"pending":[]}`, `{"history":[]}`, `{"pending":null,"history":[]}`,
				`{"pending":{},"history":[]}`, `{"pending":[],"history":[],"pendingTotal":"2"}`,
				`{"pending":[],"history":[],"pendingTotal":1.5}`, `{"pending":[],"history":[],"pendingTruncated":"no"}`,
			},
		},
		{
			args: []string{"streams", "files", "limits", "s1"}, path: filesPath + "/limits",
			want:   "decoding file upload limits response: expected a JSON object",
			bodies: []string{`[]`, `null`, `9437184`},
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

func TestImportReadsOfPartialResponses(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, body string
		want       string
	}{
		{args: []string{"runs", "streams", "r1"}, path: "/importtool/acme1/runs/r1",
			body: `{"run":{"id":"r1","status":"QUEUED"}}`,
			want: "STREAM  STATUS  READ  CREATED  UPDATED  SKIPPED  DELETED  FAILED  UNRESOLVED  COLLAPSED  EMBEDDED  " +
				"MESSAGE\n"},
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
			want: "RUNS          SUCCEEDED  PARTIAL  FAILED  CANCELLED  ABORTED  ACTIVE  RECORDS READ  RECORDS FAILED\n" +
				"not reported" + strings.Repeat(" ", 72) + "\n"},
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

func TestImportTablesRejectWhatJSONPrintsAsSent(t *testing.T) {
	type failure struct{ body, want string }
	const runStreams = "decoding run stream response: "
	for _, tt := range []struct {
		args     []string
		path     string
		failures []failure
	}{
		{args: []string{"runs", "streams", "r1"}, path: "/importtool/acme1/runs/r1", failures: []failure{
			{body: `{"run":{"id":"r1"},"streams":[{"streamName":"Lines","embedded":"many"}]}`,
				want: runStreams + "json: cannot unmarshal string"},
			{body: `{"run":{"id":"r1"},"streams":[{"streamName":"Lines","unresolvedParents":{"count":3}}]}`,
				want: runStreams + "json: cannot unmarshal object"},
			{body: `{"run":{"id":"r1"},"streams":[{"streamName":"Lines","collapsed":true}]}`,
				want: runStreams + "json: cannot unmarshal bool"},
			{body: `{"run":{"id":"r1"},"streams":[{"streamName":"Orders","sweep":"yes"}]}`,
				want: runStreams + "json: cannot unmarshal string"},
			{body: `{"run":{"id":"r1"},"streams":[{"sweep":true,"recordsRead":7}]}`,
				want: runStreams + "item 1 has no identifier"},
		}},
		{args: []string{"streams", "get", "s1"}, path: "/importtool/acme1/streams/s1", failures: []failure{
			{body: `{"id":"s1","fileMode":["SNAPSHOT"]}`, want: "decoding stream response: json: cannot unmarshal array"},
		}},
	} {
		for _, c := range tt.failures {
			fake := newFakeEmporix(t)
			fake.route(tt.path, respond(http.StatusOK, c.body))
			code, stdout, stderr := runImports(t, fake, tt.args...)
			if code != ExitError || stdout != "" || !strings.Contains(stderr, c.want) {
				t.Errorf("%q of %s: exit %d, stdout %q, stderr %q; want exit 1 and %q", tt.args, c.body, code, stdout,
					stderr, c.want)
			}
			code, stdout, stderr = runImports(t, fake, append(tt.args, "-o", "json")...)
			if code != ExitOK || stderr != "" || !reflect.DeepEqual(decodedJSON(t, stdout), decodedJSON(t, c.body)) {
				t.Errorf("%q -o json of %s: exit %d, stderr %q, stdout:\n%s\nwant it as sent", tt.args, c.body, code,
					stderr, stdout)
			}
		}
	}
}

func decodedJSON(t *testing.T, text string) any {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decoding %q: %v", text, err)
	}
	return value
}

func TestImportHelpExplainsItsReads(t *testing.T) {
	const runs = "`epx imports configurations runs CONFIGURATION_ID` shows"
	const modes = "SNAPSHOT files or INCREMENTAL files, read as `epx imports streams files` explains"
	const sum = "READ is created + updated + skipped + deleted + failed + unresolvedParents + collapsed + embedded."
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{args: nil, want: []string{"each imports one source, an entity or files, into an Emporix type"}},
		{args: []string{"runs", "get"}, want: []string{
			"a run's ID is the one that " + runs, "STATUS is QUEUED", sum, "`epx imports runs streams` shows.",
			"how it started (trigger: MANUAL, SCHEDULED, or RETRY for a retry of the run that retryOfRunId names)",
			"(streamIds, absent for all of them)",
		}},
		{args: []string{"configurations", "runs"}, want: []string{
			"most recent first", "`epx imports runs streams` shows the counters", "STATUS is QUEUED", sum,
		}},
		{args: []string{"runs", "streams"}, want: []string{
			"RUN_ID is the ID that " + runs + "; an unknown one is not found.", "add up to READ", "UNRESOLVED (",
			"COLLAPSED (", "EMBEDDED (", "STREAM followed by sweep", "duplicateKeys (", "capSkipped (", "keylessSeen (",
			"(filesRead, fileNames)",
			"STATUS is RUNNING, SUCCEEDED, PARTIAL, FAILED, ABORTED (", "or CANCELLED (the run was cancelled while the " +
				"stream ran); older runs may show QUEUED or SKIPPED.",
			"A file-based stream with nothing to read (no file waiting and, for a SNAPSHOT stream, none imported " +
				"before) imports nothing; except in a dry run, a retry or for a file read from a URL, its MESSAGE says " +
				"so and what to do (upload a file, or queue its failed file again).",
		}},
		{args: []string{"runs", "errors"}, want: []string{
			"RUN_ID is the ID that " + runs + "; an unknown one lists nothing.", "VALIDATION (", "NO_KEY (", "UPSERT (",
			"CONFIGURATION (", "EXTERNAL_SYSTEM (", "MERGE or PATCH_EMBED (", "NO_LINKING_VALUE (",
			"MARKER_DELETE, TOMBSTONE_DELETE or TOMBSTONE_NO_KEY (", "REJECTED_BY_SOURCE (", "STREAM_ERROR (",
			"FAN_OUT_REMOVE (", "(EXTRACT reading the source, MAP applying the mappings, VALIDATE checking the mapped " +
				"record or UPSERT writing it; older errors may show other names, such as MAPPING)",
			"as is payload, the source record that failed",
		}},
		{args: []string{"runs", "diagnostics"}, want: []string{"RUN_ID is the ID that " + runs, "up to 50000"}},
		{args: []string{"streams", "get"}, want: []string{
			"one source of a configuration, an entity or files,", modes, "which may name no entity", "sourceDataset",
			"A stream's ID is the one that `epx imports configurations streams CONFIGURATION_ID` shows.",
		}},
		{args: []string{"configurations", "streams"}, want: []string{
			"in the order of their seq", modes, "which may name no entity", "sourceDataset",
			"a FAN_OUT stream has no source of its own, so SOURCE names the stream it fans out from",
			"(targetWriter emporix.customEntity)",
			"PRODUCT for products and CATEGORY for categories by default, set by the service for a FAN_OUT stream",
			"MODE is STANDALONE (a stream that writes its own objects)", "or FAN_OUT for one that writes further objects",
			"(fanOut, fanOutPath, fanOutKey)", "(snapshotFiles)",
		}},
		{args: []string{"configurations", "stream-order"}, want: []string{
			"that of a FAN_OUT stream on the stream it fans out from), their seq breaking ties",
		}},
		{args: []string{"records"}, want: []string{
			"(UPSERTED, DELETED at the source, FAILED or DRY_RUN; older records may show DELETED_TARGET)",
			"natural key (with or without its key name, e.g. ID:SKU-1001; case-sensitive) or Emporix id (in any " +
				"case) starts with a text",
			"(UPSERTED, DELETED, FAILED or DRY_RUN, in upper case)",
		}},
		{args: []string{"statistics"}, want: []string{"ABORTED those the service refused to run for their configuration"}},
		{args: []string{"configurations", "trigger"}, want: []string{
			"It is a conflict (exit 5) where " + runInFlight + ".",
			"is refused with 400 Bad Request (exit 1), whose message names each stream and what to do",
			"a dry run is never refused for this", "An unknown configuration is not found (exit 4).",
			"queue its failed file again (`epx imports streams files requeue`)",
		}},
		{args: []string{"runs", "retry"}, want: []string{
			"its trigger is RETRY and its retryOfRunId the run ID",
			"It is a conflict (exit 5) where " + notRetryable + ", or where " + runInFlight + ".",
		}},
		{args: []string{"runs", "cancel"}, want: []string{"cancelRequested, which Emporix deprecated, repeats accepted"}},
		{args: []string{"streams", "records"}, want: []string{
			"STREAM_ID is the ID that `epx imports configurations streams CONFIGURATION_ID` shows; an unknown one lists",
		}},
	} {
		help := helpText(t, append([]string{"imports"}, tt.args...)...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%q: help:\n%s\nwant %q in it", tt.args, help, want)
			}
		}
	}
}

func TestImportRunJSONKeepsItsStreams(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/importtool/acme1/runs/r1", respond(http.StatusOK,
		`{"run":{"id":"r1","status":"RUNNING"},"streams":[{"streamName":"Products","recordsRead":120}]}`))
	code, stdout, stderr := runImports(t, fake, "runs", "get", "r1", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"streamName": "Products"`) || !strings.Contains(stdout, `"run": {`) {
		t.Fatalf("exit %d, stderr %q, stdout:\n%s", code, stderr, stdout)
	}
}

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

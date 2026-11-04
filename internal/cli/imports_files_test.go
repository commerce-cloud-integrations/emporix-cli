package cli

import (
	"net/http"
	"strings"
	"testing"
)

// filesPath is the file queue of the stream s1 that the tests read and change.
const filesPath = "/importtool/acme1/streams/s1/files"

// filesList is the command line of a stream's file list, as help and the
// recovery of a conflict quote it.
const filesList = "`epx imports streams files list STREAM_ID`"

// streamGet is the command line that prints a stream whole, its fileMode
// included, as help and the recovery of a conflict quote it.
const streamGet = "`epx imports streams get ID -o json`"

// TestImportFilesShowHowManyWait checks that LOAD ORDER places each waiting
// file among all the files waiting, so a table shows when the service lists
// only some of them: of pendingTotal where it is sent, else of more than
// those listed where pendingTruncated says that it lists only some, else of
// those listed; a file of the history has no LOAD ORDER, and an empty queue
// shows the header only.
func TestImportFilesShowHowManyWait(t *testing.T) {
	const files = `"pending":[{"id":"f2","status":"PENDING"},{"id":"f3","status":"PENDING"}],` +
		`"history":[{"id":"f1","status":"FAILED"}]`
	for _, tt := range []struct {
		body, want string
	}{
		{
			body: `{` + files + `,"pendingTruncated":true,"pendingTotal":2500}`,
			want: "ID  LOAD ORDER  STATUS\n" +
				"f2  1 of 2500   PENDING\n" +
				"f3  2 of 2500   PENDING\n" +
				"f1              FAILED\n",
		},
		{
			body: `{` + files + `,"pendingTruncated":true}`,
			want: "ID  LOAD ORDER        STATUS\n" +
				"f2  1 of more than 2  PENDING\n" +
				"f3  2 of more than 2  PENDING\n" +
				"f1                    FAILED\n",
		},
		{
			body: `{` + files + `}`,
			want: "ID  LOAD ORDER  STATUS\n" +
				"f2  1 of 2      PENDING\n" +
				"f3  2 of 2      PENDING\n" +
				"f1              FAILED\n",
		},
		{
			body: `{"pending":[],"history":[],"pendingTruncated":false,"pendingTotal":0}`,
			want: "ID  LOAD ORDER  STATUS\n",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(filesPath, respond(http.StatusOK, tt.body))
		code, stdout, stderr := runImports(t, fake, "streams", "files", "list", "s1", "--columns", "id,load order,status")
		if code != ExitOK || stdout != tt.want {
			t.Errorf("%s: exit %d, stderr %q, stdout:\n%s\nwant:\n%s", tt.body, code, stderr, stdout, tt.want)
		}
	}
}

// TestImportFilesJSONKeepsTheQueue checks that -o json prints the file queue
// as sent, whether the list is cut and the total included, not the rows
// that its table shows.
func TestImportFilesJSONKeepsTheQueue(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route(filesPath, respond(http.StatusOK,
		`{"pending":[{"id":"f2","loadRank":7}],"history":[],"pendingTruncated":true,"pendingTotal":2001}`))
	code, stdout, stderr := runImports(t, fake, "streams", "files", "list", "s1", "-o", "json")
	want := "{\n  \"history\": [],\n  \"pending\": [\n    {\n      \"id\": \"f2\",\n      \"loadRank\": 7\n    }\n  ],\n" +
		"  \"pendingTotal\": 2001,\n  \"pendingTruncated\": true\n}\n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, want)
	}
}

// TestImportFilesRejectAFileWithoutID checks that a table of a file queue
// whose file has no id fails (exit 1), whether it waits or not.
func TestImportFilesRejectAFileWithoutID(t *testing.T) {
	for _, body := range []string{
		`{"pending":[{"fileName":"a.csv"}],"history":[]}`, `{"pending":[],"history":[{"id":"f1"},null]}`,
	} {
		fake := newFakeEmporix(t)
		fake.route(filesPath, respond(http.StatusOK, body))
		code, stdout, stderr := runImports(t, fake, "streams", "files", "list", "s1")
		if code != ExitError || stdout != "" || !strings.Contains(stderr, "decoding stream file response: item") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", body, code, stdout, stderr)
		}
	}
}

// TestImportFileHistoryLimit checks that --history-limit takes a whole number
// from 1 to 500, sent as historyLimit, and refuses any other value while
// flags are parsed (exit 2), before anything is sent.
func TestImportFileHistoryLimit(t *testing.T) {
	for _, limit := range []string{"1", "500"} {
		fake := newFakeEmporix(t)
		fake.route(filesPath, respond(http.StatusOK, `{"pending":[],"history":[]}`))
		code, _, stderr := runImports(t, fake, "streams", "files", "list", "s1", "--history-limit", limit)
		if code != ExitOK {
			t.Fatalf("--history-limit %s: exit %d, stderr %q", limit, code, stderr)
		}
		fake.assertRequests(t, filesPath+"?historyLimit="+limit)
	}
	for _, limit := range []string{"0", "501", "-1", "1.5", "ten", " 5"} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runImports(t, fake, "streams", "files", "list", "s1", "--history-limit", limit)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "use a whole number from 1 to 500") {
			t.Errorf("--history-limit %q: exit %d, stdout %q, stderr %q", limit, code, stdout, stderr)
		}
		fake.assertNothingSent(t)
	}
}

// TestImportFileLimitsOfPartialResponses checks that the upload limits show
// their row when Emporix leaves values out.
func TestImportFileLimitsOfPartialResponses(t *testing.T) {
	const head = "MAX UPLOAD BYTES  GATEWAY LIMITED  UPLOAD SUPPORTED\n"
	for body, want := range map[string]string{
		`{}`:                        head + "not reported" + strings.Repeat(" ", 23) + "\n",
		`{"uploadSupported":false}`: head + "not reported" + strings.Repeat(" ", 23) + "false\n",
	} {
		fake := newFakeEmporix(t)
		fake.route(filesPath+"/limits", respond(http.StatusOK, body))
		code, stdout, stderr := runImports(t, fake, "streams", "files", "limits", "s1")
		if code != ExitOK || stdout != want {
			t.Errorf("%s: exit %d, stderr %q, stdout:\n%s\nwant:\n%s", body, code, stderr, stdout, want)
		}
	}
}

// runFileWrite runs `epx imports streams files args… --yes` against fake.
func runFileWrite(t *testing.T, fake *fakeEmporix, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	return runIn(t, writeEnv(fake, false, ""), append(append([]string{"imports", "streams", "files"}, args...),
		"--yes")...)
}

// TestImportFileUploadFailures checks how an upload that Emporix refuses
// ends: a file larger than the environment takes (413) and an environment
// that cannot store uploads (503) exit 1 with Emporix's message, sent once;
// as epx cannot tell that 503 from a gateway's, it says the file may have
// been queued all the same.
func TestImportFileUploadFailures(t *testing.T) {
	for _, tt := range []struct {
		status        int
		message, want string
	}{
		{
			status: http.StatusRequestEntityTooLarge, message: "The file is larger than this environment's upload limit.",
			want: "POST " + filesPath + ": 413 Request Entity Too Large: The file is larger than this environment's " +
				"upload limit.\n",
		},
		{
			status: http.StatusServiceUnavailable, message: "File upload is not configured in this environment.",
			want: "POST " + filesPath + ": 503 Service Unavailable: File upload is not configured in this environment. " +
				"(it may have been carried out all the same: read the item before you send it again)\n",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(filesPath, respond(tt.status, `{"message":"`+tt.message+`"}`))
		code, stdout, stderr := runFileWrite(t, fake, "upload", "s1", "--upload", "testdata/order.csv")
		if code != ExitError || stdout != "" || !strings.HasSuffix(stderr, tt.want) {
			t.Errorf("%d: exit %d, stdout %q, stderr %q; want exit 1 and %q", tt.status, code, stdout, stderr, tt.want)
		}
		fake.assertRequests(t, filesPath)
	}
}

// TestImportFileWritesRefuseUsage checks that an upload without a file, or
// of standard input, which has no name to send, and an order without the
// files' IDs, with both --id and --file, or with a --file that is not a JSON
// array of one or more IDs, are usage errors (exit 2) that send nothing.
func TestImportFileWritesRefuseUsage(t *testing.T) {
	file := writeBody(t, `["f2"]`)
	const notArray = "not a JSON array of one or more items"
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"upload", "s1"}, want: `required flag(s) "upload" not set`},
		{args: []string{"upload", "s1", "--upload", "-"}, want: `--upload "-": give the path of the file`},
		{args: []string{"order", "s1"}, want: "the flags in the group [file id] is required"},
		{args: []string{"order", "s1", "--id", "f2", "--file", file}, want: "[file id] were all set"},
		{args: []string{"order", "s1", "--file", writeBody(t, `{"ids":["f2"]}`)}, want: notArray},
		{args: []string{"order", "s1", "--file", writeBody(t, `[]`)}, want: notArray},
		{args: []string{"delete", "s1"}, want: "accepts 2 arg(s), received 1"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runFileWrite(t, fake, tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
}

// TestImportFileWritesOfRefusedInput checks that an order that is not
// exactly the waiting files (400) exits 1, a delete of an unknown file (404)
// exits 4, and an upload of content that the stream has (409) exits 5, each
// with Emporix's message; the conflict says what the stream's list shows.
func TestImportFileWritesOfRefusedInput(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		path, want string
		status     int
		code       int
	}{
		{args: []string{"order", "s1", "--id", "f2"}, path: filesPath + "/order", status: http.StatusBadRequest,
			code: ExitError, want: "PUT " + filesPath + "/order: 400 Bad Request: refused\n"},
		{args: []string{"delete", "s1", "f9"}, path: filesPath + "/f9", status: http.StatusNotFound, code: ExitNotFound,
			want: "DELETE " + filesPath + "/f9: 404 Not Found: refused\n"},
		{
			// The list shows files by name; Emporix matches their content.
			args: []string{"upload", "s1", "--upload", "testdata/order.csv"}, path: filesPath,
			status: http.StatusConflict, code: ExitConflict,
			want: "POST " + filesPath + ": 409 Conflict: refused (the stream imported or queued this content already, " +
				"reads no files, or has a run in flight that reads what it would replace: " + filesList + " shows the " +
				"files queued and the latest it dealt with (--history-limit, up to 500), by name, while Emporix " +
				"matches content whatever the name; a stream without fileMode (" + streamGet + ") reads no files and " +
				"takes none; otherwise wait until the run ends)\n",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route(tt.path, respond(tt.status, `{"message":"refused"}`))
		code, stdout, stderr := runFileWrite(t, fake, tt.args...)
		if code != tt.code || stdout != "" || !strings.HasSuffix(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.args, code, stdout, stderr, tt.code,
				tt.want)
		}
	}
}

// TestImportFileHelp checks that the commands of a stream's file queue say
// where STREAM_ID comes from and what an unknown one gives; that the group
// and list say once how a stream reads its waiting files by its fileMode,
// which upload and order refer to; that list says how a table shows a list
// that Emporix cut; that upload says how Emporix may refuse it and what to do
// after a 503; and which ID a delete takes.
func TestImportFileHelp(t *testing.T) {
	const stream = "STREAM_ID is the ID that `epx imports configurations streams CONFIGURATION_ID` shows; an " +
		"unknown one is not found."
	modes := []string{
		"A file-based stream reads its files by its fileMode, SNAPSHOT by default. A SNAPSHOT stream takes each " +
			"file as the whole picture: its run reads only the most recently arrived waiting file",
		"A stream without fileMode reads no files (" + streamGet + " shows it).",
	}
	for _, tt := range []struct {
		args          []string
		want, without []string
	}{
		{want: modes},
		{args: []string{"list"}, want: append([]string{
			stream,
			"LOAD ORDER is a file's place among all the files waiting, e.g. 1 of 3.",
			"Emporix lists up to 2000 waiting files: where more wait, the last LOAD ORDER listed ends below the total, " +
				"e.g. 2000 of 2500, and `epx imports streams files order` cannot reorder them.",
		}, modes...)},
		{args: []string{"limits"}, want: []string{stream}},
		{args: []string{"upload"}, without: modes, want: []string{
			stream, "Which waiting files a run reads depends on the stream's fileMode, as " + filesList + " says.",
			"413 Request Entity Too Large (exit 1). A large file may need a longer --timeout, e.g. --timeout 5m",
			"the delimited file in PATH to queue (up to 30 MB including the form), sent with its name (required)",
			"A 503 Service Unavailable (exit 1) comes from an environment that cannot store uploads (UPLOAD SUPPORTED " +
				"false in `epx imports streams files limits`), which queued nothing, or from a gateway that failed",
			"check " + filesList + " before you send it again; if it was queued, a resend is a conflict (exit 5), not a " +
				"second import.",
		}},
		{args: []string{"order"}, without: modes, want: []string{
			stream, "as " + filesList + " shows its ID;", "The order matters only for an INCREMENTAL stream, as that list says.",
		}},
		{args: []string{"delete"}, want: []string{
			"Only a waiting file is removed",
			stream + " ID is the ID that " + filesList + " shows; an unknown one is not found.",
		}},
	} {
		help := helpText(t, append([]string{"imports", "streams", "files"}, tt.args...)...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%q: help %q; want %q in it", tt.args, help, want)
			}
		}
		for _, unwanted := range tt.without {
			if strings.Contains(help, unwanted) {
				t.Errorf("%q: help %q repeats %q", tt.args, help, unwanted)
			}
		}
	}
}

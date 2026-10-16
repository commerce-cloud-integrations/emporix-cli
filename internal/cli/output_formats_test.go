package cli

import (
	"encoding/csv"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestEveryReadPrintsItsTableAsCSV runs every read case with -o csv: it
// prints the table's rows as CSV records, each with a field per column
// (encoding/csv refuses any other count), under the header, and without it
// with --no-headers; no value is cut, as tables cut long ones (view.long).
// A read that saves a file prints no table (tableReadCases).
func TestEveryReadPrintsItsTableAsCSV(t *testing.T) {
	root := newTestRoot()
	for _, tt := range tableReadCases() {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd, _, err := root.Find(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			checkTableAsCSV(t, cmd, &tt)
		})
	}
}

// checkTableAsCSV checks that the table of the read tt, of cmd, prints as CSV:
// a header and one record per row, whole values, and no header with
// --no-headers.
func checkTableAsCSV(t *testing.T, cmd *cobra.Command, tt *readCase) {
	t.Helper()
	columns, ok := cmd.Flags().Lookup(columnsFlag).Value.(*columnPicks)
	if !ok {
		t.Fatalf("--columns is a %T", cmd.Flags().Lookup(columnsFlag).Value)
	}
	rows := strings.Count(tt.want, "\n") - 1 // the table's lines but its header
	headed := readCSV(t, tt, "-o", "csv")
	if len(headed) != rows+1 || !slices.Equal(headed[0], columns.header) {
		t.Fatalf("records %q; want the header %q and %d rows", headed, columns.header, rows)
	}
	for _, record := range headed {
		if slices.ContainsFunc(record, func(cell string) bool { return strings.HasSuffix(cell, "…") }) {
			t.Fatalf("record %q holds a cut value; want it whole", record)
		}
	}
	if bare := readCSV(t, tt, "-o", "csv", "--"+noHeadersFlag); !slices.EqualFunc(bare, headed[1:], slices.Equal) {
		t.Fatalf("records %q without the header; want %q", bare, headed[1:])
	}
}

// readCSV runs the read case tt with flags and returns the CSV records that
// it prints.
func readCSV(t *testing.T, tt *readCase, flags ...string) [][]string {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route(tt.path, tt.served(t))
	code, stdout, stderr := runCLI(t, fake.env(), slices.Concat(tt.args, flags)...)
	records, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	if code != ExitOK || err != nil {
		t.Fatalf("%q: exit %d, stderr %q, CSV error %v:\n%s", flags, code, stderr, err, stdout)
	}
	return records
}

// TestWritesPrintTheirResponseAsYAML runs every write case whose response
// epx prints with -o yaml: it prints the response as YAML instead of JSON.
func TestWritesPrintTheirResponseAsYAML(t *testing.T) {
	for _, tt := range writeCases() {
		if tt.want == "" {
			continue
		}
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			fake := tt.fake(t, tt.status)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), tt.with(t, "--"+yesFlag, "-o", "yaml")...)
			if want := asYAML(t, tt.want); code != ExitOK || stdout != want || stderr != "" {
				t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, want)
			}
			tt.assertSent(t, fake)
		})
	}
}

// TestResponsesPrintAsYAML checks -o yaml where a response is printed
// otherwise than by a write case: api get, a job that --wait followed, and
// the response of a bulk write that failed; and that their -o takes no
// table format, and fails with --dry-run, which prints no response; nothing
// is sent then.
func TestResponsesPrintAsYAML(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/site/acme1/sites", respond(http.StatusOK, `[{"code":"AcmeStore_NL","active":true}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "api", "get", "/site/{tenant}/sites", "-o", "yaml")
	if want := "- \"active\": true\n  \"code\": \"AcmeStore_NL\"\n"; code != ExitOK || stdout != want {
		t.Fatalf("api get: exit %d, stderr %q, stdout %q; want %q", code, stderr, stdout, want)
	}
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"api", "get", "/x", "-o", "csv"}, message: "use json or yaml"},
		{args: []string{"api", "delete", "/x", "-o", "table"}, message: "use json or yaml"},
		// --dry-run prints the request, so a format of the response cannot apply.
		{args: []string{"api", "delete", "/x", "--dry-run", "-o", "yaml"},
			message: "[dry-run output] are set none of the others"},
		{args: []string{"units", "create", "--file", "-", "--dry-run", "-o", "json"},
			message: "[dry-run output] are set none of the others"},
	} {
		sent := newFakeEmporix(t)
		code, _, stderr = runCLI(t, sent.env(), tt.args...)
		if code != ExitUsage || !strings.Contains(stderr, tt.message) {
			t.Errorf("%q: exit %d, stderr %q; want a usage error with %q", tt.args, code, stderr, tt.message)
		}
		sent.assertNothingSent(t)
	}

	fake = newFakeEmporix(t)
	fake.route(reindexJobsPath, respond(http.StatusCreated, `{"id":"j1","status":"PENDING"}`))
	last := `{"id":"j1","status":"SUCCESS","expectedCount":5}`
	fake.route(reindexJobPath, respond(http.StatusOK, last))
	if code, stdout, stderr = waitFor(t, fake, "--wait", "-o", "yaml"); code != ExitOK || stdout != asYAML(t, last) {
		t.Fatalf("--wait: exit %d, stderr %q, stdout %q", code, stderr, stdout)
	}

	fake = newFakeEmporix(t)
	results := `[{"index":0,"id":"p1","code":204},{"index":1,"id":"p2","code":409}]`
	fake.route("/price/acme1/prices/bulk", respond(http.StatusMultiStatus, results))
	code, stdout, stderr = runIn(t, writeEnv(fake, false, ""), "prices", "bulk-update", "--file",
		writeBody(t, `[{"id":"p1"},{"id":"p2"}]`), "--yes", "-o", "yaml")
	if code != ExitConflict || stdout != asYAML(t, results) || !strings.HasPrefix(stderr, "item 1 (p2): 409 Conflict\n") {
		t.Fatalf("bulk-update: exit %d, stderr %q, stdout %q", code, stderr, stdout)
	}
}

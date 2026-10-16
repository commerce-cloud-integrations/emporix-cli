package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// TestOutputFlagsOfEveryCommand checks which commands take which output
// flags: every read but one that saves a file (--out), auth status and
// profile list and show, which show a table, -o (table, json, yaml or csv),
// --columns, which can name each of its columns, and --no-headers; api get
// and every write, which print a response, -o (json or yaml) only; any other
// command, a read that saves a file included, none.
func TestOutputFlagsOfEveryCommand(t *testing.T) {
	var reads, writes int
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		read, write := checkOutputFlags(t, cmd)
		if read {
			reads++
		}
		if write {
			writes++
		}
	})
	if reads < 100 || writes < 10 {
		t.Fatalf("%d commands show a table and %d print a response; want the whole tree", reads, writes)
	}
}

// checkOutputFlags checks that -o of cmd takes the formats its output has: a
// table for a read, a response for a write and for api get. It reports
// whether cmd is a read, and whether it is a write.
func checkOutputFlags(t *testing.T, cmd *cobra.Command) (read, write bool) {
	t.Helper()
	tables := []string{"table", "json", "yaml", "csv"}
	responses := []string{"json", "yaml"}
	otherTables := []string{"epx auth status", "epx profile list", "epx profile show"}
	_, isRead := cmd.Annotations[scopesAnnotation]
	isRead = isRead && cmd.Flags().Lookup(outFlag) == nil // a read that saves a file prints none
	isWrite := cmd.Flags().Lookup(dryRunFlag) != nil
	var want []string
	switch {
	case isRead || slices.Contains(otherTables, cmd.CommandPath()):
		want, read = tables, true
		assertColumns(t, cmd)
	case isWrite || cmd.CommandPath() == "epx api get":
		want, write = responses, true
		if cmd.Flags().Lookup(columnsFlag) != nil || cmd.Flags().Lookup(noHeadersFlag) != nil {
			t.Errorf("%s takes --%s or --%s, but shows no table", cmd.CommandPath(), columnsFlag, noHeadersFlag)
		}
	}
	var got []string
	if given := cmd.Flags().Lookup(outputFlag); given != nil {
		choice, ok := given.Value.(*formatChoice)
		if !ok {
			t.Fatalf("%s: -o is a %T", cmd.CommandPath(), given.Value)
		}
		got = choice.names()
	}
	if !slices.Equal(got, want) {
		t.Errorf("%s: -o takes %q; want %q", cmd.CommandPath(), got, want)
	}
	return read, write
}

// assertColumns checks that --columns of cmd can name each of its columns:
// no name holds a comma, and no two are the same in any case.
func assertColumns(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	columns, ok := cmd.Flags().Lookup(columnsFlag).Value.(*columnPicks)
	if !ok || cmd.Flags().Lookup(noHeadersFlag) == nil || len(columns.header) == 0 {
		t.Errorf("%s: --columns %v and --no-headers; want both", cmd.CommandPath(), columns)
		return
	}
	for i, name := range columns.header {
		if strings.Contains(name, ",") || slices.IndexFunc(columns.header[:i], func(before string) bool {
			return strings.EqualFold(before, name)
		}) >= 0 {
			t.Errorf("%s: --columns cannot name the column %q of %q", cmd.CommandPath(), name, columns.header)
		}
	}
}

// TestReadsPrintTheFormatAndColumnsTheyAreGiven checks the shared builder's
// output flags: CSV of the table's rows, --columns, which picks and orders
// them by header in any case (repeated, it adds columns), --no-headers, and
// YAML of the whole response, which a list also checks as JSON.
func TestReadsPrintTheFormatAndColumnsTheyAreGiven(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things", firstPage(`[{"id":"a1","name":"Shirt, \"blue\"","size":1.50},{"id":"b2","name":"Pants"}]`))
	fake.route("/things/a1", respond(http.StatusOK, `{"id":"a1","tags":["x"],"name":"yes"}`))
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"list", "-o", "csv"}, want: "ID,NAME\na1,\"Shirt, \"\"blue\"\"\"\nb2,Pants\n"},
		{args: []string{"list", "--columns", "name,Id"}, want: "NAME           ID\nShirt, \"blue\"  a1\nPants          b2\n"},
		{args: []string{"list", "--no-headers"}, want: "a1  Shirt, \"blue\"\nb2  Pants\n"},
		{args: []string{"list", "-o", "csv", "--columns", "NAME", "--columns", " id ", "--no-headers"},
			want: "\"Shirt, \"\"blue\"\"\",a1\nPants,b2\n"},
		{args: []string{"list", "-o", "yaml"}, want: "- \"id\": \"a1\"\n  \"name\": \"Shirt, \\\"blue\\\"\"\n" +
			"  \"size\": 1.50\n- \"id\": \"b2\"\n  \"name\": \"Pants\"\n"},
		{args: []string{"get", "a1", "-o", "yaml"}, want: "\"id\": \"a1\"\n\"name\": \"yes\"\n\"tags\":\n  - \"x\"\n"},
		{args: []string{"get", "a1", "-o", "csv", "--columns", "name"}, want: "NAME\nyes\n"},
	} {
		code, stdout, errText := runThings(t, fake, tt.args...)
		if code != ExitOK || stdout != tt.want {
			t.Errorf("%q: exit %d, error %q, stdout:\n%s\nwant:\n%s", tt.args, code, errText, stdout, tt.want)
		}
	}
}

// TestOutputUsageErrorsSendNothing checks that output flags that cannot
// apply fail as usage errors before anything is sent: a column that the
// table does not have (the error names those it has), one named twice,
// --columns or --no-headers with a format that prints the response whole,
// and with --stream, which excludes every output flag.
func TestOutputUsageErrorsSendNothing(t *testing.T) {
	fake := newFakeEmporix(t)
	const whole = "--columns and --no-headers apply only to -o table and -o csv"
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"list", "--columns", "id,price"},
			message: `invalid argument "id,price" for "--columns" flag: no column "price": use ID, NAME`},
		{args: []string{"list", "--columns", ""}, message: `no column "": use ID, NAME`},
		{args: []string{"list", "--columns", "id,ID"}, message: "column ID named twice"},
		{args: []string{"list", "--columns", "name", "--columns", "Name"}, message: "column NAME named twice"},
		{args: []string{"list", "-o", "json", "--columns", "id"}, message: whole},
		{args: []string{"get", "a1", "-o", "yaml", "--no-headers"}, message: whole},
		{args: []string{"list", "--stream", "-o", "csv"}, message: "[stream output] are set none of the others"},
		{args: []string{"list", "--stream", "--columns", "id"}, message: "[stream columns] are set none of the others"},
		{args: []string{"list", "--stream", "--no-headers"}, message: "[stream no-headers] are set none of the others"},
	} {
		code, stdout, errText := runThings(t, fake, tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(errText, tt.message) {
			t.Errorf("%q: exit %d, stdout %q, error %q; want exit 2 and %q", tt.args, code, stdout, errText, tt.message)
		}
	}
	fake.assertNothingSent(t)
}

func TestAuthStatusPrintsEveryFormat(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"-o", "yaml", "--scope", "b.read,a.read"},
			want: "\"expiresInSeconds\": 3600\n\"scopes\":\n  - \"a.read\"\n  - \"b.read\"\n\"tenant\": \"acme1\"\n"},
		{args: []string{"-o", "csv", "--columns", "scopes,tenant", "--scope", "b.read,a.read"},
			want: "SCOPES,TENANT\na.read b.read,acme1\n"},
		{args: []string{"--no-headers", "--columns", "expires in"}, want: "1h0m0s\n"},
	} {
		code, stdout, stderr := runCLI(t, fake.env(), append([]string{"auth", "status"}, tt.args...)...)
		if code != ExitOK || stdout != tt.want {
			t.Errorf("%q: exit %d, stderr %q, stdout %q; want %q", tt.args, code, stderr, stdout, tt.want)
		}
	}
	fake = newFakeEmporix(t)
	code, _, stderr := runCLI(t, fake.env(), "auth", "status", "-o", "json", "--no-headers")
	if code != ExitUsage || !strings.Contains(stderr, "apply only to -o table and -o csv") {
		t.Fatalf("exit %d, stderr %q; want a usage error", code, stderr)
	}
	fake.assertNothingSent(t)
}

// TestTablesAloneCutLongValues checks that a table cuts the values of the
// columns its view declares long, which CSV prints whole, also when
// --columns picks them in another order, and that a long column the header
// lacks panics as the tree is built.
func TestTablesAloneCutLongValues(t *testing.T) {
	long := strings.Repeat("v", maxValueRunes+10)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{want: "KEY  VALUE\nk1   " + cut(long) + "\n"},
		{args: []string{"-o", "csv"}, want: "KEY,VALUE\nk1," + long + "\n"},
		{args: []string{"-o", "csv", "--columns", "value,key", "--no-headers"}, want: long + ",k1\n"},
	} {
		cmd := &cobra.Command{Use: "values", SilenceErrors: true, SilenceUsage: true}
		printed := addTableOutput(cmd, []string{"KEY", "VALUE"}, "VALUE")
		var out strings.Builder
		cmd.SetOut(&out)
		cmd.RunE = func(c *cobra.Command, _ []string) error {
			return printed.print(c.OutOrStdout(), nil, func() ([][]string, error) { return [][]string{{"k1", long}}, nil })
		}
		cmd.SetArgs(tt.args)
		if err := cmd.Execute(); err != nil || out.String() != tt.want {
			t.Errorf("%q: %q (%v), want %q", tt.args, out.String(), err, tt.want)
		}
	}
	defer func() {
		if recovered := recover(); recovered == nil || !strings.Contains(fmt.Sprint(recovered), "VALUE is not a column") {
			t.Fatalf("recovered %v; want a panic naming the column", recovered)
		}
	}()
	addTableOutput(&cobra.Command{Use: "values"}, []string{"KEY"}, "VALUE")
}

// asYAML returns body as epx prints YAML.
func asYAML(t *testing.T, body string) string {
	t.Helper()
	var printed strings.Builder
	if err := output.WriteYAML(&printed, json.RawMessage(body)); err != nil {
		t.Fatal(err)
	}
	return printed.String()
}

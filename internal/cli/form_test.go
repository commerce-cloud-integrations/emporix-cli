package cli

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// schemaFile is the content of testdata/schema.json, the JSON schema file
// that the cases of forms send.
const schemaFile = `{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object",
  "properties": {"color": {"type": "string"}}}
`

// sentPart is a part of a form as a request sends it.
type sentPart struct {
	name        string
	file        string // the name of the file it sends; empty for a field
	contentType string // empty: none
	content     string
}

// described returns what a dry run shows of p: a line that describes it,
// then its content where it is JSON.
func (p sentPart) described() string {
	text := fmt.Sprintf("part %q:", p.name)
	if p.file != "" {
		text += fmt.Sprintf(" file %q,", p.file)
	}
	if p.contentType != "" {
		text += " " + p.contentType + ","
	}
	text += fmt.Sprintf(" %d bytes\n", len(p.content))
	if p.contentType == emporix.JSONContentType {
		text += strings.TrimSuffix(p.content, "\n") + "\n"
	}
	return text
}

// schemaFilePart is the part of a form that sends testdata/schema.json, the
// file that --json-schema names in the cases.
func schemaFilePart() sentPart {
	return sentPart{name: "file", file: "schema.json", contentType: emporix.JSONContentType, content: schemaFile}
}

// sentForm returns the parts of body, a form whose Content-Type is
// contentType, which names its boundary; any other body fails the test.
func sentForm(t *testing.T, contentType, body string) []sentPart {
	t.Helper()
	media, params, err := mime.ParseMediaType(contentType)
	if err != nil || media != "multipart/form-data" || params["boundary"] == "" {
		t.Fatalf("Content-Type %q; want a form with its boundary", contentType)
	}
	form := multipart.NewReader(strings.NewReader(body), params["boundary"])
	var parts []sentPart
	for {
		part, err := form.NextPart()
		if errors.Is(err, io.EOF) {
			return parts
		}
		if err != nil {
			t.Fatalf("form %q: %v", body, err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, sentPart{
			name: part.FormName(), file: part.FileName(), contentType: part.Header.Get("Content-Type"),
			content: string(content),
		})
	}
}

// assertForm checks that request i was sent with method and a form of the
// parts want, in their order.
func (f *fakeEmporix) assertForm(t *testing.T, i int, method string, want []sentPart) {
	t.Helper()
	if parts := sentForm(t, f.types[i], f.bodies[i]); !strings.HasPrefix(f.sent[i], method+" ") ||
		!slices.Equal(parts, want) {
		t.Fatalf("request %d sent %q with the parts %q; want %s with %q", i+1, f.sent[i], parts, method, want)
	}
}

// TestSchemaFileOfTheCases checks that testdata/schema.json holds schemaFile,
// which the cases expect to be sent.
func TestSchemaFileOfTheCases(t *testing.T) {
	if content, err := os.ReadFile("testdata/schema.json"); err != nil || string(content) != schemaFile {
		t.Fatalf("testdata/schema.json: %q, %v; want %q", content, err, schemaFile)
	}
}

// TestFormFileFailuresSendNothing checks that a file of a form that is not
// given, cannot be read, is not JSON, or is standard input or named with a
// control character (whose name cannot be sent), and an empty field, are
// usage errors (exit 2) found before anything is sent, for a write and for a
// read with a form.
func TestFormFileFailuresSendNothing(t *testing.T) {
	directory := t.TempDir()
	body := writeBody(t, `{"name":{"en":"Specs"},"types":["PRODUCT"]}`)
	unsendable := filepath.Join(directory, "schema"+string(rune(0x1b))+".json") // a header cannot carry its name
	if err := os.WriteFile(unsendable, []byte(schemaFile), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"--json-schema", writeBody(t, `{"type":`)}, want: `body.json": not JSON: unexpected end`},
		{args: []string{"--json-schema", directory}, want: `": is a directory`},
		{args: []string{"--json-schema", directory + "/none.json"}, want: `none.json": no such file`},
		{args: []string{"--json-schema", "-"}, want: `--json-schema "-": give the path of the file, whose name is sent`},
		{args: []string{"--json-schema", unsendable}, want: `invalid file name "schema\x1b.json": a header cannot carry`},
		{args: nil, want: `required flag(s) "json-schema" not set`},
	} {
		for _, command := range [][]string{
			{"schemas", "references", "create", "--file", body, "--yes"},
			{"schemas", "references", "update", "r1", "--file", body, "--dry-run"},
			{"schemas", "parse", "--" + writeScopeFlag},
		} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, `{}`), slices.Concat(command, tt.args)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %q: exit %d, stdout %q, stderr %q; want exit 2 and %q", command, tt.args, code, stdout,
					stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
	for want, args := range map[string][]string{
		`required flag(s) "file" not set`: {"schemas", "references", "create", "--json-schema", "testdata/schema.json",
			"--yes"},
		`invalid argument "" for "--id" flag: empty`: {"schemas", "parse", "--json-schema", "testdata/schema.json",
			"--id", "", "--" + writeScopeFlag},
		`body.json": not JSON`: {"schemas", "references", "update", "r1", "--json-schema", "testdata/schema.json",
			"--file", writeBody(t, `{"name":`), "--yes"},
	} {
		fake := newFakeEmporix(t)
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), args...)
		if code != ExitUsage || !strings.Contains(stderr, want) {
			t.Errorf("%q: exit %d, stderr %q; want exit 2 and %q", args, code, stderr, want)
		}
		fake.assertNothingSent(t)
	}
}

// TestFormReadsChangeNothing checks that a read with a form says that it
// changes nothing, and that its JSON prints the response whole.
func TestFormReadsChangeNothing(t *testing.T) {
	_, help, _ := runCLI(t, nil, "schemas", "parse", "--help")
	if !strings.Contains(help, "Changes nothing: Emporix serves this read as POST with a multipart/form-data\nbody") {
		t.Errorf("help %q; want it to say that it changes nothing", help)
	}
	fake := newFakeEmporix(t)
	fake.route("/schema/acme1/schemas/file", respond(http.StatusOK, `{"attributes":[],"metadata":{"version":1}}`))
	code, stdout, stderr := runCLI(t, fake.env(), "schemas", "parse", "--json-schema", "testdata/schema.json", "-o",
		"json", "--"+writeScopeFlag)
	if code != ExitOK || stdout != "{\n  \"attributes\": [],\n  \"metadata\": {\n    \"version\": 1\n  }\n}\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	for _, body := range []string{`{"metadata":{}}`, `[]`, `{"attributes":null}`} {
		fake.route("/schema/acme1/schemas/file", respond(http.StatusOK, body))
		code, _, stderr := runCLI(t, fake.env(), "schemas", "parse", "--json-schema", "testdata/schema.json",
			"--"+writeScopeFlag)
		if code != ExitError || !strings.Contains(stderr, "decoding parsed schema response: expected a schema with "+
			"attributes") {
			t.Errorf("%s: exit %d, stderr %q; want a malformed response", body, code, stderr)
		}
	}
}

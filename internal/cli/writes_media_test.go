package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// pngFile is the content of testdata/asset.png, the file that --upload names
// in the cases: a PNG signature, then a few bytes.
const pngFile = "\x89PNG\r\n\x1a\nacme"

// The JSON of the media assets that the cases send.
const (
	linkAsset = `{"type":"LINK","access":"PUBLIC","url":"https://example.com/a.pdf"}`
	blobAsset = `{"type":"BLOB","access":"PUBLIC","refIds":[{"type":"PRODUCT","id":"100200"}]}`
)

// uploadForm returns the parts of the form that uploads testdata/asset.png
// with the JSON body.
func uploadForm(body string) []sentPart {
	return []sentPart{
		{name: "file", file: "asset.png", contentType: "image/png", content: pngFile},
		{name: "body", contentType: emporix.JSONContentType, content: body},
	}
}

// mediaWriteCases are the writes of media assets: a link, a file uploaded in
// a form, changes, a delete and the start of an upload straight to storage.
func mediaWriteCases() []writeCase {
	manage := []string{"media.asset_manage"}
	const assets, asset = "/media/acme1/assets", "/media/acme1/assets/a1"
	return []writeCase{
		{
			args: []string{"media", "create"}, file: linkAsset, method: http.MethodPost, path: assets, scopes: manage,
			status: http.StatusCreated, body: `{"id":"a1"}`, want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			args: []string{"media", "create", "--upload", "testdata/asset.png"}, file: blobAsset, form: uploadForm(blobAsset),
			method: http.MethodPost, path: assets, scopes: manage, status: http.StatusCreated, body: `{"id":"a1"}`,
			want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			args: []string{"media", "update", "a1"}, file: linkAsset, method: http.MethodPut, path: asset, scopes: manage,
			status: http.StatusNoContent,
		},
		{
			args: []string{"media", "update", "a1", "--upload", "testdata/asset.png"}, file: blobAsset,
			form: uploadForm(blobAsset), method: http.MethodPut, path: asset, scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"media", "patch", "a1"}, file: `[{"op":"add","path":"/refIds/-","value":{"type":"PRODUCT",` +
				`"id":"100200"}}]`, method: http.MethodPatch, path: asset, scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"media", "delete", "a1"}, method: http.MethodDelete, path: asset, scopes: manage,
			status: http.StatusNoContent, confirmBy: "a1",
		},
		{
			args: []string{"media", "start-upload"}, file: `{"type":"BLOB","access":"PRIVATE","uploadType":"form"}`,
			method: http.MethodPost, path: assets + "/upload-session", scopes: manage, status: http.StatusCreated,
			body: `{"id":"a2","provider":"GCS","status":"PENDING"}`,
			want: "{\n  \"id\": \"a2\",\n  \"provider\": \"GCS\",\n  \"status\": \"PENDING\"\n}\n", refuses: true,
		},
	}
}

// TestUploadFileOfTheCases checks that testdata/asset.png holds pngFile,
// which the cases expect to be sent.
func TestUploadFileOfTheCases(t *testing.T) {
	if content, err := os.ReadFile("testdata/asset.png"); err != nil || string(content) != pngFile {
		t.Fatalf("testdata/asset.png: %q, %v; want %q", content, err, pngFile)
	}
}

// TestUploadSendsAnyFileWithItsContentType checks that --upload sends a file
// whatever it holds, byte for byte, with the content type that its
// extension names in epx's own table, the same on every system, else the one
// its content shows.
func TestUploadSendsAnyFileWithItsContentType(t *testing.T) {
	directory := t.TempDir()
	for name, want := range map[string]string{
		"manual":     "application/pdf", // shown by its content
		"manual.PDF": "application/pdf", // named by its extension, whatever its case
		"sheet.xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"list.csv":   "text/csv",                  // not in Go's own table, so its system's would decide
		"notes.md":   "text/plain; charset=utf-8", // in no table of epx's: shown by its content
		"data":       "application/octet-stream",  // neither
	} {
		content := "%PDF-1.7\n\x00\x01"
		switch name {
		case "data":
			content = "\x00\x01\x02"
		case "notes.md", "list.csv":
			content = "# a,b\n"
		}
		upload := filepath.Join(directory, name)
		if err := os.WriteFile(upload, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		fake := newFakeEmporix(t)
		fake.route("/media/acme1/assets", respond(http.StatusCreated, `{"id":"a1"}`))
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), "media", "create", "--upload", upload, "--file",
			writeBody(t, blobAsset), "--yes")
		if code != ExitOK {
			t.Fatalf("%s: exit %d, stderr %q", name, code, stderr)
		}
		fake.assertForm(t, len(fake.sent)-1, http.MethodPost, []sentPart{
			{name: "file", file: name, contentType: want, content: content},
			{name: "body", contentType: emporix.JSONContentType, content: blobAsset},
		})
	}
}

// TestUploadFailuresSendNothing checks that a file to upload that cannot be
// read, is too large, is standard input or is named with a control
// character, a JSON body that is missing or not JSON, and the JSON of a BLOB
// asset without a file to upload, are usage errors (exit 2) found before
// anything is sent.
func TestUploadFailuresSendNothing(t *testing.T) {
	directory := t.TempDir()
	large := filepath.Join(directory, "large.bin")
	unsendable := filepath.Join(directory, "asset"+string(rune(0x1b))+".png") // a header cannot carry its name
	if err := os.WriteFile(unsendable, []byte(pngFile), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(large, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(large, maxUploadBytes+1); err != nil {
		t.Fatal(err)
	}
	body := writeBody(t, blobAsset)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"--upload", directory + "/none.png", "--file", body}, want: `none.png": no such file`},
		{args: []string{"--upload", large, "--file", body}, want: `large.bin": larger than 31457280 bytes`},
		{args: []string{"--upload", "-", "--file", body}, want: `--upload "-": give the path of the file, whose name`},
		{args: []string{"--upload", unsendable, "--file", body}, want: `invalid file name "asset\x1b.png": a header`},
		{args: []string{"--upload", "testdata/asset.png"}, want: `required flag(s) "file" not set`},
		{args: []string{"--upload", "testdata/asset.png", "--file", writeBody(t, `{"type":`)}, want: `": not JSON`},
		{args: []string{"--file", body}, want: `": a BLOB asset is sent with its file: give --upload PATH`},
	} {
		for _, command := range [][]string{{"media", "create", "--yes"}, {"media", "update", "a1", "--dry-run"}} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, `{}`), slices.Concat(command, tt.args)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %q: exit %d, stdout %q, stderr %q; want exit 2 and %q", command, tt.args, code, stdout,
					stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

// TestUploadDryRunShowsNoTerminalControls checks that a dry run of a file to
// upload that is named as JSON but holds terminal controls describes the
// file's part without showing its content, and shows the JSON of --file.
func TestUploadDryRunShowsNoTerminalControls(t *testing.T) {
	upload := filepath.Join(t.TempDir(), "asset.json")
	if err := os.WriteFile(upload, []byte("\x1b]0;title\x07"), 0o600); err != nil {
		t.Fatal(err)
	}
	fake := newFakeEmporix(t)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "media", "create", "--upload", upload, "--file",
		writeBody(t, blobAsset), "--dry-run")
	if code != ExitOK || strings.Contains(stdout, "\x1b") ||
		!strings.Contains(stdout, `part "file": file "asset.json", application/json, 10 bytes`+"\n"+`part "body"`) ||
		!strings.Contains(stdout, blobAsset) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	fake.assertNothingSent(t)
}

// TestMediaWritesSayHowToUpload checks that create and update name --upload
// and say how a BLOB asset is sent, how large its form may be, and that a
// large file may need a longer --timeout, and that delete has no --file.
func TestMediaWritesSayHowToUpload(t *testing.T) {
	for _, command := range []string{"create", "update"} {
		_, help, _ := runCLI(t, nil, "media", command, "--help")
		if !strings.Contains(help, "--upload PATH") || !strings.Contains(help, "multipart/form-data form, with the JSON") ||
			!strings.Contains(help, "up to 30 MB including the form") ||
			!strings.Contains(help, "may need a longer") || !strings.Contains(help, "e.g. --timeout 5m") {
			t.Errorf("media %s help %q; want --upload explained", command, help)
		}
	}
	if _, help, _ := runCLI(t, nil, "media", "delete", "--help"); strings.Contains(help, "--file") {
		t.Errorf("media delete help %q; want no --file", help)
	}
}

// TestFormsOfUploadsAreBounded checks that a form is bounded by its parts:
// one with a file to upload at maxUploadBytes, one without at nothing; and
// that both bodies that send a form, alone (formBody) and with the JSON of
// --file (formFile), send a form of its bound and refuse one byte more as a
// usage error.
func TestFormsOfUploadsAreBounded(t *testing.T) {
	field := fieldPart("attachmentId", mediaFlag, "")
	if got := (form{field, uploadPart("file", "")}).limit(); got != maxUploadBytes {
		t.Errorf("form with a file to upload bounded at %d; want %d", got, maxUploadBytes)
	}
	if got := (form{field}).limit(); got != 0 {
		t.Errorf("form without a file to upload bounded at %d; want 0", got)
	}
	sources := map[string]func(part formPart) bodySource{
		"formBody": func(part formPart) bodySource {
			return &formBody{form: form{part}, oneOf: []string{uploadFlag}}
		},
		"formFile": func(part formPart) bodySource { return &formFile{form: form{part}, field: "body"} },
	}
	body := writeBody(t, blobAsset)
	for name, source := range sources {
		unbounded, err := fillBounded(t, source, 0, body)
		if err != nil {
			t.Fatal(err)
		}
		size := unbounded.Len()
		if sent, sendErr := fillBounded(t, source, size, body); sendErr != nil || sent.Len() != size {
			t.Errorf("%s bounded at %d: form of %d bytes, %v; want it sent", name, size, sent.Len(), sendErr)
		}
		_, err = fillBounded(t, source, size-1, body)
		want := fmt.Sprintf("the form is %d bytes, larger than the %d bytes that Emporix takes in one request", size,
			size-1)
		if !errors.As(err, new(usageError)) || !strings.Contains(err.Error(), want) {
			t.Errorf("%s bounded at %d: %v; want a usage error %q", name, size-1, err, want)
		}
	}
}

// fillBounded returns the body that source makes of a form with
// testdata/asset.png to upload, bounded at bound, and the JSON in the file
// body where source takes --file.
func fillBounded(t *testing.T, source func(part formPart) bodySource, bound int, body string) (emporix.Body,
	error) {
	t.Helper()
	part := uploadPart("file", "")
	part.bound = bound
	cmd := &cobra.Command{}
	given := source(part)
	given.bind(cmd)
	for flag, value := range map[string]string{uploadFlag: "testdata/asset.png", fileFlag: body} {
		if cmd.Flags().Lookup(flag) != nil {
			if err := cmd.Flags().Set(flag, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	var w emporix.Write
	err := given.fill(context.Background(), &w, nil)
	return w.Body, err
}

// TestUploadsAreBoundedWithTheirForms checks end to end that the form of a
// file that fits alone but not with the rest of the form is a usage error
// (exit 2) found before anything is sent. The file is sparse and read once:
// the test holds a few copies of it at most.
func TestUploadsAreBoundedWithTheirForms(t *testing.T) {
	form, err := emporix.FormBody([]emporix.Part{{Name: "file", FileName: "edge.pdf", ContentType: "application/pdf"},
		{Name: "body", ContentType: emporix.JSONContentType, Content: []byte(blobAsset)}})
	if err != nil {
		t.Fatal(err)
	}
	size := maxUploadBytes - form.Len() + 1
	upload := filepath.Join(t.TempDir(), "edge.pdf")
	if err := os.WriteFile(upload, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(upload, int64(size)); err != nil {
		t.Fatal(err)
	}
	fake := newFakeEmporix(t)
	code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), "media", "create", "--upload", upload, "--file",
		writeBody(t, blobAsset), "--yes")
	want := fmt.Sprintf("the form is %d bytes, larger than the %d bytes", maxUploadBytes+1, maxUploadBytes)
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, want) {
		t.Errorf("file of %d bytes: exit %d, stdout %q, stderr %q; want exit 2 and %q", size, code, stdout, stderr, want)
	}
	fake.assertNothingSent(t)
}

// TestBlobNeedsFileRefusesOnlyABlobAsset checks that the JSON of a media asset
// sent without a file is refused only where it is that of a BLOB asset, as
// Emporix reads its type: exactly so named, the last where it repeats;
// another type, none, or JSON that is no object is sent as it is.
func TestBlobNeedsFileRefusesOnlyABlobAsset(t *testing.T) {
	for asset, refused := range map[string]bool{
		blobAsset: true, linkAsset: false, `{"type":"blob"}`: false, `{}`: false, `[{"type":"BLOB"}]`: false,
		`{"type":1}`: false, `{"Type":"BLOB"}`: false, `{"type":"BLOB","TYPE":"LINK"}`: true,
		`{"type":"LINK","type":"BLOB"}`: true, `{"type":"BLOB","type":"LINK"}`: false,
	} {
		if err := blobNeedsFile([]byte(asset)); (err != nil) != refused {
			t.Errorf("%s: %v; want refused %t", asset, err, refused)
		}
	}
}

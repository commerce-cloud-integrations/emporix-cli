package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// quotePDF stands in for the PDF of a quote, as Emporix answers with it.
const quotePDF = "%PDF-1.7\n1 0 obj << /Type /Catalog >> endobj\n%%EOF\n"

// readEnv returns the environment of a read from fake (writeEnv, without a
// terminal to confirm on), whose standard output is a terminal where
// outTerminal says so.
func readEnv(fake *fakeEmporix, outTerminal bool) Env {
	env := writeEnv(fake, false, "")
	env.OutTerminal = outTerminal
	return env
}

// TestQuotePDFIsSaved checks that quotes pdf, a read, sends POST without a
// body with a token of quote.quote_read, asks for the PDF in the language of
// --language, else without Accept-Language, saves it to the file that --out
// names, replacing one there, readable by its owner only, and says so on
// standard error, printing nothing; it asks for no confirmation.
func TestQuotePDFIsSaved(t *testing.T) {
	for _, tt := range []struct {
		flags    []string
		language string // the Accept-Language that the fake must get
	}{
		{language: ""},
		{flags: []string{"--language", "DE"}, language: "de"},
	} {
		fake := newFakeEmporix(t)
		var accept, language []string
		fake.route("/quote/acme1/quotes/Q1001/pdf", func(w http.ResponseWriter, r *http.Request) {
			accept, language = r.Header.Values("Accept"), r.Header.Values("Accept-Language")
			respond(http.StatusOK, quotePDF, "Content-Type", "application/pdf")(w, r)
		})
		out := filepath.Join(t.TempDir(), "Q1001.pdf")
		if err := os.WriteFile(out, []byte("an older file"), 0o600); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := runIn(t, readEnv(fake, false),
			append([]string{"quotes", "pdf", "Q1001", "--out", out}, tt.flags...)...)
		want := "Saved the PDF (" + strconv.Itoa(len(quotePDF)) + " bytes) to " + out + ".\n"
		if code != ExitOK || stdout != "" || stderr != want {
			t.Fatalf("%q: exit %d, stdout %q, stderr %q; want %q", tt.flags, code, stdout, stderr, want)
		}
		fake.assertScopes(t, "quote.quote_read")
		fake.assertSent(t, http.MethodPost, "")
		if strings.Join(accept, ",") != "application/pdf, application/json" ||
			strings.Join(language, ",") != tt.language {
			t.Errorf("%q: Accept %q, Accept-Language %q; want the PDF in %q", tt.flags, accept, language, tt.language)
		}
		saved, err := os.ReadFile(filepath.Clean(out))
		info, statErr := os.Stat(out)
		if err != nil || statErr != nil || string(saved) != quotePDF || info.Mode().Perm()&0o077 != 0 {
			t.Errorf("%q: saved %q (%v, mode %v); want the PDF, readable by its owner only", tt.flags, saved, err,
				info.Mode())
		}
		if entries, _ := os.ReadDir(filepath.Dir(out)); len(entries) != 1 {
			t.Errorf("%q: %d files beside the PDF; want none left behind", tt.flags, len(entries)-1)
		}
	}
}

// TestQuotePDFRefusesWhereItCannotSave checks that quotes pdf refuses, as
// its flags are parsed and so before anything is sent (exit 2), an --out
// that it cannot save the PDF to: - where standard output is a terminal, an
// empty path, one whose directory does not exist, and a directory, a device
// or a symbolic link, which it would replace rather than write to; and that
// it takes no flag of a write.
func TestQuotePDFRefusesWhereItCannotSave(t *testing.T) {
	dir := t.TempDir()
	target, link := filepath.Join(dir, "target.pdf"), filepath.Join(dir, "link.pdf")
	if err := os.WriteFile(target, []byte("a file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"--out", "-"}, want: "standard output is a terminal: name a file"},
		{args: []string{"--out", ""}, want: `invalid argument "" for "--out" flag: empty`},
		{args: []string{"--out", filepath.Join(dir, "missing", "a.pdf")}, want: "its directory does not exist"},
		{args: []string{"--out", dir}, want: "a directory, not a file: name a file, or use - for standard output"},
		{args: []string{"--out", link}, want: "a symbolic link, not a file"},
		{args: []string{"--out", os.DevNull}, want: "a device, pipe or socket, not a file"},
		{want: `required flag(s) "out" not set`},
		{args: []string{"--out", target, "--yes"}, want: "unknown flag: --yes"},
		{args: []string{"--out", target, "--dry-run"}, want: "unknown flag: --dry-run"},
		{args: []string{"--out", target, "-o", "json"}, want: "unknown shorthand flag: 'o'"},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runIn(t, readEnv(fake, true), append([]string{"quotes", "pdf", "Q1001"}, tt.args...)...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		fake.assertNothingSent(t)
	}
	if saved, err := os.ReadFile(filepath.Clean(target)); err != nil || string(saved) != "a file" {
		t.Errorf("the file that the link names holds %q (%v); want it unchanged", saved, err)
	}
}

// TestQuotePDFFailures checks that quotes pdf saves nothing where Emporix
// answers with no PDF (by its Content-Type; exit 1), with an error status
// (exit 4 for an unknown quote, 1 otherwise), or where the PDF cannot be
// saved (exit 1): here as its directory is removed while Emporix answers.
func TestQuotePDFFailures(t *testing.T) {
	dir := t.TempDir()
	gone := filepath.Join(dir, "gone")
	if err := os.Mkdir(gone, 0o700); err != nil {
		t.Fatal(err)
	}
	removing := func(w http.ResponseWriter, r *http.Request) {
		if err := os.Remove(gone); err != nil {
			t.Error(err)
		}
		respond(http.StatusOK, quotePDF, "Content-Type", "application/pdf")(w, r)
	}
	const notPDF = "Error: POST /quote/acme1/quotes/Q1001/pdf: the response is "
	for _, tt := range []struct {
		answer http.HandlerFunc
		out    string
		code   int
		stderr string
	}{
		{answer: respond(http.StatusOK, `{"url":"https://files.acme.example/Q1001.pdf"}`, "Content-Type",
			"application/json"), out: filepath.Join(dir, "a.pdf"), code: ExitError,
			stderr: notPDF + "application/json, not the application/pdf that it asked for\n"},
		{answer: respond(http.StatusOK, "<html>busy</html>"), out: filepath.Join(dir, "b.pdf"), code: ExitError,
			stderr: notPDF + "text/html, not the application/pdf"},
		{answer: respond(http.StatusNotFound, `{"message":"Quote Q1001 not found"}`), out: filepath.Join(dir, "c.pdf"),
			code: ExitNotFound, stderr: "Error: "},
		{answer: respond(http.StatusBadRequest, `{"message":"bad language"}`), out: filepath.Join(dir, "d.pdf"),
			code: ExitError, stderr: "Error: "},
		{answer: removing, out: filepath.Join(gone, "e.pdf"), code: ExitError, stderr: "Error: the PDF was not saved: "},
	} {
		fake := newFakeEmporix(t)
		fake.route("/quote/acme1/quotes/Q1001/pdf", tt.answer)
		code, stdout, stderr := runIn(t, readEnv(fake, false), "quotes", "pdf", "Q1001", "--out", tt.out)
		if code != tt.code || stdout != "" || !strings.HasPrefix(stderr, tt.stderr) {
			t.Errorf("--out %s: exit %d, stdout %q, stderr %q; want exit %d and %q", tt.out, code, stdout, stderr,
				tt.code, tt.stderr)
		}
		if _, err := os.Lstat(tt.out); err == nil {
			t.Errorf("--out %s: saved; want nothing saved", tt.out)
		}
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Errorf("%d files left in %s (%v); want none", len(entries), dir, err)
	}
}

// TestQuotePDFIsARead checks the help of quotes pdf: a read of
// quote.quote_read that says how it saves the PDF, without a confirmation
// or a claim that it changes data.
func TestQuotePDFIsARead(t *testing.T) {
	_, help, _ := runCLI(t, nil, "quotes", "pdf", "--help")
	for _, want := range []string{
		"\n\nChanges nothing: Emporix serves this read as POST with no body.\n" +
			"Emporix answers with the PDF, which epx saves to the file that --out names",
		"An answer that is no PDF fails\n(exit 1), and nothing is saved.\n\nRequires the quote.quote_read scope.\n",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("quotes pdf: help %q; want %q", help, want)
		}
	}
	for _, unwanted := range []string{"Changes data", "Sent as a write", "--yes", "--dry-run"} {
		if strings.Contains(help, unwanted) {
			t.Errorf("quotes pdf: help %q; want no %q", help, unwanted)
		}
	}
}

package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestLongHelpBackquotesSpacedLiterals checks the long help of every command
// of the tree as written: a literal that the reference shows as code
// (markLiterals) and that holds a space, such as -o json, a JSON value or a
// quoted message, is backquoted, so that the help reflow keeps it on one line
// as the reference keeps it in one code span.
func TestLongHelpBackquotesSpacedLiterals(t *testing.T) {
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		for _, literal := range spacedLiterals(cmd.Long) {
			t.Errorf("%s: help holds %q outside backquotes; backquote it", cmd.CommandPath(), literal)
		}
	})
}

// TestSpacedLiteralsFindsUnquotedLiteralsWithSpaces checks the check of
// TestLongHelpBackquotesSpacedLiterals.
func TestSpacedLiteralsFindsUnquotedLiteralsWithSpaces(t *testing.T) {
	for _, tt := range []struct {
		text string
		want []string
	}{
		{text: "Print it with -o json, or -o yaml.", want: []string{"-o json", "-o yaml"}},
		{text: "Print it with `-o json`, -o, --file or --active=false.", want: nil},
		{text: `Send {"id": ID} or [{"op": "add"}], not {x}.`, want: []string{`{"id": ID}`, `[{"op": "add"}]`}},
		{text: "Send `{\"id\": ID}` and {\"id\":ID}.", want: nil},
		{text: `It answers "The Price must be defined." or "ok".`, want: []string{`"The Price must be defined."`}},
		{text: `As "feeType": "ABSOLUTE" and 'q: v'; don't.`, want: []string{`'q: v'`}},
		{text: `e.g. {"feeType": "ABSOLUTE"}.`, want: []string{`{"feeType": "ABSOLUTE"}`}},
		{text: "A \"split\nmessage\" across lines.", want: []string{"\"split message\""}},
		{text: "One \"a\n\nb\" per paragraph.", want: nil},
		{text: "Run:\n  epx a -o json\n\tepx b -o json", want: nil},
		{text: "an `unmatched -o json", want: []string{"-o json"}},
		{text: "`-o json` and an `unmatched -o yaml", want: []string{"-o yaml"}},
		{text: "#### As -o json:", want: []string{"-o json"}},
	} {
		if got := spacedLiterals(tt.text); !slices.Equal(got, tt.want) {
			t.Errorf("spacedLiterals(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

// spacedLiterals returns the literals of help text that the reference shows
// as code (literalSpans) and that hold a space, in the paragraphs of prose
// that the reflow fills (helpChunks) and outside their backquoted spans,
// which the reflow keeps whole (wordEnd). The text after an unmatched
// backquote counts as prose, as the reflow may break a line inside it.
func spacedLiterals(text string) []string {
	var found []string
	for prose, kept := range helpChunks(text) {
		if kept {
			continue
		}
		parts := strings.Split(prose, "`")
		for i, part := range parts {
			if i%2 == 1 && i < len(parts)-1 {
				continue // a backquoted span; the text after an unmatched backquote is prose
			}
			found = append(found, literalsWithSpace(part)...)
		}
	}
	return found
}

// literalsWithSpace returns the literals of text, prose without backquotes,
// that markLiterals shows as code (literalSpans) and that hold a space or a
// tab.
func literalsWithSpace(text string) []string {
	var found []string
	for start, end := range literalSpans(text) {
		if strings.ContainsAny(text[start:end], " \t") {
			found = append(found, text[start:end])
		}
	}
	return found
}

package cli

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// TestReflowWrapsProseAndKeepsTheRest checks that reflow fills each paragraph
// of prose up to the width, never inside a backquoted span, and keeps every
// line that is not prose as it is.
func TestReflowWrapsProseAndKeepsTheRest(t *testing.T) {
	for _, tt := range []struct{ name, text, want string }{
		{name: "joins and fills", text: "one two\nthree four five six", want: "one two\nthree four\nfive six"},
		{name: "spaces collapse", text: "one  two \n\tthree", want: "one two\n\tthree"},
		{name: "exact width", text: "abcd efghi j", want: "abcd efghi\nj"},
		{name: "span unbroken", text: "run `epx a b` now", want: "run\n`epx a b`\nnow"},
		{name: "span across lines", text: "go `epx\na` now", want: "go `epx a`\nnow"},
		{name: "span with punctuation", text: "as (`a b c`), fine", want: "as\n(`a b c`),\nfine"},
		{name: "wide word alone", text: "a `very long span` b", want: "a\n`very long span`\nb"},
		{name: "paragraphs", text: "one two\n\nthree\nfour\n", want: "one two\n\nthree four\n"},
		{name: "blank with spaces", text: "one\n  \ntwo", want: "one\n  \ntwo"},
		{name: "indented example", text: "run:\n  epx a --site S\nthen", want: "run:\n  epx a --site S\nthen"},
		{name: "unmatched backquote", text: "a `b c d e f", want: "a `b c d e\nf"},
		{name: "runes", text: "…… …… ……", want: "…… …… ……"},
		{name: "empty", text: "", want: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := reflow(tt.text, 10); got != tt.want {
				t.Errorf("reflow(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// TestHelpChunksYieldsKeptLinesAndParagraphs checks helpChunks: kept lines as
// they are and the prose between them joined, in order, until the loop stops.
func TestHelpChunksYieldsKeptLinesAndParagraphs(t *testing.T) {
	var got []string
	for chunk, kept := range helpChunks("one\ntwo\n\n  run\nthree") {
		got = append(got, fmt.Sprintf("%q %v", chunk, kept))
	}
	want := []string{`"one two" false`, `"" true`, `"  run" true`, `"three" false`}
	if !slices.Equal(got, want) {
		t.Errorf("helpChunks = %q, want %q", got, want)
	}
	for stop := range 3 {
		seen := 0
		for range helpChunks("one\n\ntwo") {
			if seen == stop {
				break
			}
			seen++
		}
		if seen != stop {
			t.Errorf("helpChunks went on after the loop stopped at %d", stop)
		}
	}
}

// TestHelpShowsLongReflowedAndLeavesItAsWritten checks the help of commands
// of the tree as epx builds it: it shows the long help reflowed (longHelp)
// and the examples as written, and leaves Long as written, plain prose.
func TestHelpShowsLongReflowedAndLeavesItAsWritten(t *testing.T) {
	for _, tt := range []struct {
		args    []string
		example bool // the command has an example
	}{{args: []string{"sites"}}, {args: []string{"api", "get"}, example: true}} {
		var out strings.Builder
		root := newRootCommand(&app{getenv: func(string) string { return "" }}, &out, io.Discard)
		cmd, _, err := root.Find(tt.args)
		if err != nil || tt.example != (cmd.Example != "") {
			t.Fatalf("%q: %v, example %q", tt.args, err, cmd.Example)
		}
		long := cmd.Long
		root.SetArgs(append(tt.args, "--help"))
		if err := root.Execute(); err != nil || !strings.HasPrefix(out.String(), longHelp(cmd)+"\n") ||
			cmd.Long != long || longHelp(cmd) == long {
			t.Errorf("%q: %v: help %q, then Long %q; want the help to start with %q and Long %q, which help wraps",
				tt.args, err, out.String(), cmd.Long, longHelp(cmd), long)
		}
		if tt.example && !strings.Contains(out.String(), "Examples:\n"+cmd.Example+"\n") {
			t.Errorf("%q: help %q; want the example %q as written", tt.args, out.String(), cmd.Example)
		}
	}
}

// TestLongHelpFitsTheWidth checks the long help of every command of the tree
// as its help shows it (longHelp): no line is wider than helpWidth unless it
// is one word that starts with a backquote, such as a long command line, and
// reflowing it again changes nothing.
func TestLongHelpFitsTheWidth(t *testing.T) {
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		long := longHelp(cmd)
		for _, line := range overWidth(long, helpWidth) {
			t.Errorf("%s: help line of %d characters, more than %d: %q", cmd.CommandPath(),
				utf8.RuneCountInString(line), helpWidth, line)
		}
		if again := reflow(long, helpWidth); again != long {
			t.Errorf("%s: help changes when reflowed again:\n%s\nnot\n%s", cmd.CommandPath(), again, long)
		}
	})
}

// TestLongHelpIsPlainProse checks the long help of every command of the
// tree as written (plainProse). Cobra writes the help of its completion
// commands, wrapped by hand, so they are left out, but for the paragraph
// that epx adds to that of completion (completionHelp).
func TestLongHelpIsPlainProse(t *testing.T) {
	root := newTestRoot()
	completion, _, err := root.Find([]string{"completion"})
	if err != nil {
		t.Fatal(err)
	}
	walkCommands(root, func(cmd *cobra.Command) {
		if cmd != completion && cmd.Parent() != completion && !plainProse(cmd.Long) {
			t.Errorf("%s: help %q is not plain prose", cmd.CommandPath(), cmd.Long)
		}
	})
	if !strings.HasSuffix(completion.Long, "\n\n"+completionHelp()) || !plainProse(completionHelp()) {
		t.Errorf("completion: help %q; want it to end with plain prose %q", completion.Long, completionHelp())
	}
}

// TestPlainProseFindsBreaksAndSeparators checks the check of
// TestLongHelpIsPlainProse.
func TestPlainProseFindsBreaksAndSeparators(t *testing.T) {
	for text, want := range map[string]bool{
		"One. Two.\n\nThree.": true, "Run:\n  epx a\n\tepx b": true, "": true,
		"One.\nTwo.": false, "One.  Two.": false, "One. ": false, "One.\n\n Two.": false, " One.": false,
		"One.Two.": false, "Run `a.B` now.": true, "Exits 1.When": false, "An unknown ID.Next": false,
		"Run `epx products list`.Next": false, "As VENDOR.LOCATION.": true,
	} {
		if got := plainProse(text); got != want {
			t.Errorf("plainProse(%q) = %v, want %v", text, got, want)
		}
	}
}

// plainProse reports whether text, help as written, is plain prose: each
// paragraph one line, its words separated by single spaces (reflow at any
// width wraps it), so no fragment of it carries a line break or a separator
// of its own; a line that starts with one space and then a word is a
// separator too, as an indented line starts with two spaces or a tab. Nor
// does a sentence end joined to the next word outside a backquoted span
// (sentenceJoined): a fragment that lacks its separator.
func plainProse(text string) bool {
	for line := range strings.SplitSeq(text, "\n") {
		if strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "  ") && strings.TrimSpace(line) != "" {
			return false
		}
		if !indented(line) && sentenceJoined.MatchString(backquoted.ReplaceAllString(line, "x")) {
			return false
		}
	}
	return reflow(text, math.MaxInt) == text
}

// sentenceJoined matches the end of a sentence joined to the next one: a
// period after a letter, digit or bracket and before a capital, but between
// two words in capitals (VENDOR.LOCATION). backquoted matches a backquoted
// span, which plainProse replaces with a word first, as a period in it may be
// followed by a capital (`epx api get /x/A.B`).
var (
	sentenceJoined = regexp.MustCompile(`[a-z0-9)\]]\.[A-Z]|[A-Z]\.[A-Z][a-z]`)
	backquoted     = regexp.MustCompile("`[^`]*`")
)

// TestOverWidthFindsWideLines checks the check of TestLongHelpFitsTheWidth:
// a line wider than the width is found unless it is one word that starts with
// a backquote.
func TestOverWidthFindsWideLines(t *testing.T) {
	text := "short\nthis line is wide\n`a wide span`.\n\tan indented wide one\nhttps://example.com/wide"
	got := overWidth(text, 12)
	want := []string{"this line is wide", "\tan indented wide one", "https://example.com/wide"}
	if !slices.Equal(got, want) {
		t.Errorf("overWidth = %q, want %q", got, want)
	}
}

// overWidth returns the lines of text wider than width but those that are
// one word that starts with a backquote (helpWords).
func overWidth(text string, width int) []string {
	var wide []string
	for line := range strings.SplitSeq(text, "\n") {
		words := slices.Collect(helpWords(line))
		if utf8.RuneCountInString(line) > width && (len(words) != 1 || !strings.HasPrefix(words[0], "`")) {
			wide = append(wide, line)
		}
	}
	return wide
}

// flat returns text as paragraphs, which blank lines separate, with the
// white space within each as single spaces.
func flat(text string) string {
	var paragraphs []string
	for paragraph := range strings.SplitSeq(text, "\n\n") {
		paragraphs = append(paragraphs, strings.Join(strings.Fields(paragraph), " "))
	}
	return strings.Join(paragraphs, "\n\n")
}

// containsFlat reports whether help holds want, both flat.
func containsFlat(help, want string) bool { return strings.Contains(flat(help), flat(want)) }

// TestFlatKeepsParagraphs checks flat: lines and spaces of a paragraph
// collapse, paragraphs stay apart.
func TestFlatKeepsParagraphs(t *testing.T) {
	for text, want := range map[string]string{
		"a  b\nc\n \n\n\td\n": "a b c\n\nd", "\n\nRequires\n": "\n\nRequires", "x\n\n": "x\n\n", "\nx": "x",
	} {
		if got := flat(text); got != want {
			t.Errorf("flat(%q) = %q, want %q", text, got, want)
		}
	}
}

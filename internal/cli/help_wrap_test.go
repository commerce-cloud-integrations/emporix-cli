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

func TestHelpShowsLongReflowedAndLeavesItAsWritten(t *testing.T) {
	for _, tt := range []struct {
		args    []string
		example bool
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

// Cobra writes the help of its completion commands, wrapped by hand, so they are
// left out, but for the paragraph that epx adds to that of completion
// (completionHelp).
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

var (
	sentenceJoined = regexp.MustCompile(`[a-z0-9)\]]\.[A-Z]|[A-Z]\.[A-Z][a-z]`)
	backquoted     = regexp.MustCompile("`[^`]*`")
)

func TestOverWidthFindsWideLines(t *testing.T) {
	text := "short\nthis line is wide\n`a wide span`.\n\tan indented wide one\nhttps://example.com/wide"
	got := overWidth(text, 12)
	want := []string{"this line is wide", "\tan indented wide one", "https://example.com/wide"}
	if !slices.Equal(got, want) {
		t.Errorf("overWidth = %q, want %q", got, want)
	}
}

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

// pflag takes the first backquoted word of a flag usage for the name of its value
// and drops the backquotes, so a flag usage has no spans.
func overWidthUsage(lines []string, width int) []string {
	var wide []string
	for _, line := range lines {
		if utf8.RuneCountInString(line) > width && len(strings.Fields(line)) != 1 {
			wide = append(wide, line)
		}
	}
	return wide
}

func cobraHeading(line string) bool {
	switch line {
	case "Usage:", "Aliases:", "Examples:", "Available Commands:", "Additional Commands:",
		"Flags:", "Global Flags:", "Additional help topics:":
		return true
	}
	return false
}

func helpSections(help string) map[string][]string {
	sections := map[string][]string{}
	name, usage := "", false
	for line := range strings.SplitSeq(help, "\n") {
		switch {
		case cobraHeading(line) && (usage || line == "Usage:"):
			name, usage = line, true
		case line == "" && usage:
			name = ""
		default:
			sections[name] = append(sections[name], line)
		}
	}
	return sections
}

func flagEntries(lines []string) []string {
	var entries []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		starts := strings.HasPrefix(trimmed, "-") && len(line)-len(trimmed) <= len("      ")
		switch {
		case trimmed == "":
		case starts || len(entries) == 0:
			entries = append(entries, trimmed)
		default:
			entries[len(entries)-1] += " " + trimmed
		}
	}
	for i, entry := range entries {
		entries[i] = strings.Join(strings.Fields(entry), " ")
	}
	return entries
}

func flagUsage(help, flag string) (string, bool) {
	for _, entry := range flagEntries(helpSections(help)["Flags:"]) {
		if len(entry) > 4 && entry[0] == '-' && entry[2:4] == ", " {
			entry = entry[4:]
		}
		if usage, found := strings.CutPrefix(entry, flag+" "); found {
			return usage, true
		}
	}
	return "", false
}

func flat(text string) string {
	var paragraphs []string
	for paragraph := range strings.SplitSeq(text, "\n\n") {
		paragraphs = append(paragraphs, strings.Join(strings.Fields(paragraph), " "))
	}
	return strings.Join(paragraphs, "\n\n")
}

func containsFlat(help, want string) bool { return strings.Contains(flat(help), flat(want)) }

func TestFlatKeepsParagraphs(t *testing.T) {
	for text, want := range map[string]string{
		"a  b\nc\n \n\n\td\n": "a b c\n\nd", "\n\nRequires\n": "\n\nRequires", "x\n\n": "x\n\n", "\nx": "x",
	} {
		if got := flat(text); got != want {
			t.Errorf("flat(%q) = %q, want %q", text, got, want)
		}
	}
}

func assertHelpHolds(t *testing.T, parts map[string][]string) {
	t.Helper()
	for args, want := range parts {
		_, help, _ := runCLI(t, nil, append(strings.Fields(args), "--help")...)
		for _, part := range want {
			if !containsFlat(help, part) {
				t.Errorf("%s help %q; want %q", args, help, part)
			}
		}
	}
}

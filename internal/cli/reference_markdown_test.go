package cli

import (
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
)

// Help text becomes Markdown that shows the same text as --help: prose is
// escaped, backquoted spans become code spans and indented lines text blocks.

// helpBlocks renders help text as Markdown blocks showing the same text:
// paragraphs of prose lines, and runs of indented lines as text blocks.
func helpBlocks(text string, root *cobra.Command) []string {
	var blocks, run []string
	indented := false
	flush := func() {
		switch {
		case len(run) == 0:
		case indented:
			blocks = append(blocks, fence(dedent(run)))
		default:
			blocks = append(blocks, paragraph(strings.Join(run, "\n"), root))
		}
		run = nil
	}
	for line := range strings.Lines(text) {
		line = strings.TrimRight(line, " \t\r\n")
		lineIndented := strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
		if line == "" || lineIndented != indented {
			flush()
		}
		if title, isHeading := helpHeading(line); isHeading {
			flush()
			blocks = append(blocks, "**"+markLiterals(title)+"**")
			continue
		}
		if line != "" {
			indented = lineIndented
			run = append(run, line)
		}
	}
	flush()
	return blocks
}

// helpHeading returns the title of a Markdown heading line in help text, such
// as Cobra's "#### Linux:", which the site shows as a bold line.
func helpHeading(line string) (string, bool) {
	level := len(line) - len(strings.TrimLeft(line, "#"))
	if level == 0 || level > 6 || !strings.HasPrefix(line[level:], " ") {
		return "", false
	}
	return strings.TrimSpace(line[level:]), true
}

// paragraph renders prose as Markdown that shows the same text: backquoted
// spans become code spans (linked to the command they cite, if any), literals
// such as --file or a JSON value become code spans too (markLiterals), and
// every other character Markdown or HTML would interpret is escaped.
func paragraph(text string, root *cobra.Command) string {
	parts := strings.Split(text, "`")
	var b strings.Builder
	for i, part := range parts {
		switch {
		case i%2 == 0:
			b.WriteString(markLiterals(part))
		case i == len(parts)-1:
			b.WriteString("\\`" + escapeMarkdown(part)) // an unmatched backquote
		case part == "":
			b.WriteString("\\`\\`")
		default:
			b.WriteString(citation(strings.ReplaceAll(part, "\n", " "), root))
		}
	}
	lines := strings.Split(b.String(), "\n")
	for i, line := range lines {
		lines[i] = escapeLineStart(line)
	}
	return strings.Join(lines, "\n")
}

// cell renders text for a table cell, on one line.
func cell(text string, root *cobra.Command) string {
	return strings.ReplaceAll(paragraph(text, root), "\n", " ")
}

// citation renders a backquoted span of help as a code span, linked to the
// section of the command it cites when it is a command line of root.
func citation(code string, root *cobra.Command) string {
	span := codeSpan(code)
	words := strings.Fields(code)
	if len(words) < 2 || words[0] != root.Name() {
		return span
	}
	cmd := citedCommand(root, words[1:])
	if cmd == nil {
		return span
	}
	return "[" + span + "](" + pageURL(cmd) + "#" + anchorOf(cmd) + ")"
}

// citedCommand returns the command that words, a command line of root without
// the program name, cites: each word names a subcommand until an argument (a
// flag, a quoted or an uppercase word) or a command without subcommands. It
// returns nil when a word names no subcommand or none is named.
func citedCommand(root *cobra.Command, words []string) *cobra.Command {
	cmd := root
	for _, word := range words {
		if !cmd.HasSubCommands() || strings.ContainsAny(word[:1], "-'\"") || unicode.IsUpper(rune(word[0])) {
			break
		}
		next := slices.IndexFunc(cmd.Commands(), func(sub *cobra.Command) bool {
			return sub.Name() == word || sub.HasAlias(word)
		})
		if next < 0 {
			return nil
		}
		cmd = cmd.Commands()[next]
	}
	if cmd == root {
		return nil
	}
	return cmd
}

// markLiterals renders prose with its literals as code spans and the rest
// escaped. The site sets prose in typography, turning -- into an em dash and
// straight quotes into curly ones, which would change what a reader copies:
// flags (--file, --active=false, -o), a bare --, JSON values and quoted
// values are therefore code.
func markLiterals(text string) string {
	var b strings.Builder
	last := 0
	for i := 0; i < len(text); i++ {
		end := literalEnd(text, i)
		if end <= i {
			continue
		}
		b.WriteString(escapeMarkdown(text[last:i]) + codeSpan(strings.ReplaceAll(text[i:end], "\n", " ")))
		last, i = end, end-1
	}
	return b.String() + escapeMarkdown(text[last:])
}

// literalEnd returns the end of the literal that starts at text[i], or i when
// none does; a literal never starts inside a word.
func literalEnd(text string, i int) int {
	if i > 0 && isWordByte(text[i-1]) {
		return i
	}
	switch text[i] {
	case '-':
		return flagEnd(text, i)
	case '{', '[':
		return jsonEnd(text, i)
	case '"', '\'':
		return quotedEnd(text, i)
	}
	return i
}

// flagEnd returns the end of a flag at text[i]: --name with an optional
// =value, a bare -- (or one before a digit), or a one-letter -x.
func flagEnd(text string, i int) int {
	switch {
	case strings.HasPrefix(text[i:], "--"):
		end := i + 2
		if end < len(text) && isLower(text[end]) {
			for end < len(text) && (isLower(text[end]) || isDigit(text[end]) || text[end] == '-') {
				end++
			}
			end = valueEnd(text, end)
		}
		return end
	case i+1 < len(text) && isLetter(text[i+1]) && (i+2 == len(text) || !isWordByte(text[i+2])):
		return formatEnd(text, i+2)
	}
	return i
}

// formatEnd returns the end of the output format that follows -o at text[end]
// (as in -o json), so that the two read as one; end when none follows.
func formatEnd(text string, end int) int {
	if text[end-1] != 'o' {
		return end
	}
	for _, format := range []string{"json", "yaml", "csv", "table"} {
		after := end + 1 + len(format)
		if strings.HasPrefix(text[end:], " "+format) && (after == len(text) || !isWordByte(text[after])) {
			return after
		}
	}
	return end
}

// valueEnd returns the end of the =value that may follow a flag ending at
// text[end]: up to a space or a , ; or ), without a final period.
func valueEnd(text string, end int) int {
	if end >= len(text) || text[end] != '=' {
		return end
	}
	value := end + 1
	for value < len(text) && !strings.ContainsRune(" \t\n,;)", rune(text[value])) {
		value++
	}
	for value > end+1 && text[value-1] == '.' {
		value--
	}
	return value
}

// jsonEnd returns the end of a JSON object or array at text[i], one that
// starts with {" [{ or [" and closes before the text ends; i otherwise.
func jsonEnd(text string, i int) int {
	if i+1 >= len(text) || !strings.ContainsRune(map[byte]string{'{': "\"", '[': "{\""}[text[i]], rune(text[i+1])) {
		return i
	}
	depth, inString := 0, false
	for end := i; end < len(text); end++ {
		switch c := text[end]; {
		case inString:
			inString = c != '"' || text[end-1] == '\\'
		case c == '"':
			inString = true
		case c == '{' || c == '[':
			depth++
		case c == '}' || c == ']':
			depth--
			if depth == 0 {
				return end + 1
			}
		}
	}
	return i
}

// quotedEnd returns the end of a quoted value at text[i]: a non-empty run in
// double or single quotes on one line, the closing quote not followed by a
// letter (an apostrophe); i otherwise.
func quotedEnd(text string, i int) int {
	quote := text[i]
	closing := strings.IndexByte(text[i+1:], quote)
	if closing <= 0 || strings.Contains(text[i+1:i+1+closing], "\n") {
		return i
	}
	end := i + 1 + closing + 1
	if end < len(text) && isLetter(text[end]) {
		return i
	}
	return end
}

func isLetter(c byte) bool { return isLower(c) || ('A' <= c && c <= 'Z') }

func isLower(c byte) bool { return 'a' <= c && c <= 'z' }

func isDigit(c byte) bool { return '0' <= c && c <= '9' }

// isWordByte reports whether c continues a word or a flag: a letter, digit, "_" or "-".
func isWordByte(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '-'
}

// escapeMarkdown escapes the characters of text that Markdown or HTML would interpret.
func escapeMarkdown(text string) string {
	var b strings.Builder
	for _, r := range text {
		switch {
		case r == '<':
			b.WriteString("&lt;")
		case r == '&':
			b.WriteString("&amp;")
		case strings.ContainsRune("\\`*_{}[]#|>", r):
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// escapeLineStart escapes what would start a block at the beginning of a
// line: a list item, a setext underline, an admonition or a numbered item.
func escapeLineStart(line string) string {
	switch {
	case line == "":
		return line
	case strings.ContainsRune("-+!", rune(line[0])):
		return "\\" + line
	case line[0] == '=':
		return "&#61;" + line[1:]
	case line[0] == '?':
		return "&#63;" + line[1:]
	}
	digits := len(line) - len(strings.TrimLeft(line, "0123456789"))
	if digits > 0 && digits < len(line) && line[digits] == '.' {
		return line[:digits] + "\\" + line[digits:]
	}
	return line
}

// codeSpan renders text as a code span, delimited by more backquotes than it holds.
func codeSpan(text string) string {
	ticks := "`"
	for strings.Contains(text, ticks) {
		ticks += "`"
	}
	if len(ticks) > 1 {
		text = " " + text + " "
	}
	return ticks + text + ticks
}

func codeList(items []string) string {
	spans := make([]string, len(items))
	for i, item := range items {
		spans[i] = codeSpan(item)
	}
	return strings.Join(spans, ", ")
}

// fence renders text as a fenced text block, delimited by more backquotes than it holds.
func fence(text string) string { return fenced("text", text) }

// shellFence renders commands as a fenced shell block.
func shellFence(text string) string { return fenced("sh", text) }

func fenced(language, text string) string {
	marks := "```"
	for strings.Contains(text, marks) {
		marks += "`"
	}
	return marks + language + "\n" + text + "\n" + marks
}

// dedent joins lines without the leading white space they all share (blank lines aside).
func dedent(lines []string) string {
	text := slices.DeleteFunc(slices.Clone(lines), isBlank)
	prefix := ""
	if len(text) > 0 {
		prefix = text[0][:len(text[0])-len(strings.TrimLeft(text[0], " \t"))]
	}
	for _, line := range text {
		for !strings.HasPrefix(line, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = strings.TrimPrefix(line, prefix)
	}
	return strings.Join(out, "\n")
}

func isBlank(line string) bool { return strings.TrimSpace(line) == "" }

// TestReferenceRendersHelpFaithfully checks that help text shows as written:
// Markdown and HTML characters are escaped, backquoted command lines link to
// their section and indented lines keep their layout.
func TestReferenceRendersHelpFaithfully(t *testing.T) {
	root := newTestRoot()
	for _, tt := range []struct{ help, want string }{
		{help: `a <b> & *c* _d_ [e](f) #g {h} |i| \j`, want: `a &lt;b\> &amp; \*c\* \_d\_ \[e\](f) \#g \{h\} \|i\| \\j`},
		{help: "x\n- y\n= z\n? w\n12. v\n!!! u\n3 t", want: "x\n\\- y\n&#61; z\n&#63; w\n12\\. v\n\\!!! u\n3 t"},
		{
			help: "--yes, given --dry-run; --active=false. -o json, e-mail, a--b, -- -1, --9",
			want: "`--yes`, given `--dry-run`; `--active=false`. `-o json`, e-mail, a--b, `--` -1, `--`9",
		},
		{
			help: "as {\"a\": [1, \"]\"]}, [{\"op\": 1}], {x}, \"100 EUR\", 'q:\"v\"' and don't",
			want: "as `{\"a\": [1, \"]\"]}`, `[{\"op\": 1}]`, \\{x\\}, `\"100 EUR\"`, `'q:\"v\"'` and don't",
		},
		{help: "#### Linux:\n\ntext", want: "**Linux:**\n\ntext"},
		{
			help: "run `epx sites\nlist`, `epx nope`, `epx sites nope`, `x`, `` and `odd",
			want: "run [`epx sites list`](/commands/sites#epx-sites-list), `epx nope`, `epx sites nope`, `x`, \\`\\` and \\`odd",
		},
		{
			help: "`epx sites get main`, `epx sites CODE`, `epx sites --site S`",
			want: "[`epx sites get main`](/commands/sites#epx-sites-get), [`epx sites CODE`](/commands/sites#epx-sites), " +
				"[`epx sites --site S`](/commands/sites#epx-sites)",
		},
		{help: "intro:\n\n    a\n      b\n\tc ```\nend", want: "intro:\n\n````text\n    a\n      b\n\tc ```\n````\n\nend"},
		{help: "  one\n\n  two", want: "```text\none\n```\n\n```text\ntwo\n```"},
	} {
		if got := strings.Join(helpBlocks(tt.help, root), "\n\n"); got != tt.want {
			t.Errorf("help %q renders as\n%s\nwant\n%s", tt.help, got, tt.want)
		}
	}
	if got := codeSpan("a`b"); got != "`` a`b ``" {
		t.Errorf("codeSpan = %q", got)
	}
}

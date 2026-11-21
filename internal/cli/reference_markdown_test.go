package cli

import (
	"iter"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
)

func helpBlocks(text string, root *cobra.Command) []string {
	var blocks []string
	for _, run := range helpRuns(text) {
		if run.indented {
			blocks = append(blocks, fence(dedent(run.lines)))
		} else {
			blocks = append(blocks, proseBlock(run.lines, root))
		}
	}
	return blocks
}

type helpRun struct {
	lines    []string
	indented bool
}

func helpRuns(text string) []helpRun {
	var runs []helpRun
	var run helpRun
	flush := func() {
		if len(run.lines) > 0 {
			runs = append(runs, run)
		}
		run = helpRun{}
	}
	for line := range strings.Lines(text) {
		line = strings.TrimRight(line, " \t\r\n")
		if line == "" || indented(line) != run.indented {
			flush()
		}
		if line != "" {
			run.indented = indented(line)
			run.lines = append(run.lines, line)
		}
	}
	flush()
	return runs
}

func proseBlock(run []string, root *cobra.Command) string {
	if len(run) == 1 {
		if title, isHeading := helpHeading(run[0]); isHeading {
			return "**" + markLiterals(title) + "**"
		}
	}
	return paragraph(strings.Join(run, " "), root)
}

func helpHeading(line string) (string, bool) {
	level := len(line) - len(strings.TrimLeft(line, "#"))
	if level == 0 || level > 6 || !strings.HasPrefix(line[level:], " ") {
		return "", false
	}
	return strings.TrimSpace(line[level:]), true
}

func paragraph(text string, root *cobra.Command) string {
	parts := strings.Split(text, "`")
	var b strings.Builder
	for i, part := range parts {
		switch {
		case i%2 == 0:
			b.WriteString(markLiterals(part))
		case i == len(parts)-1:
			b.WriteString("\\`" + escapeMarkdown(part))
		case part == "":
			b.WriteString("\\`\\`")
		default:
			b.WriteString(citation(part, root))
		}
	}
	return escapeLineStart(b.String())
}

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

func markLiterals(text string) string {
	var b strings.Builder
	last := 0
	for start, end := range literalSpans(text) {
		b.WriteString(escapeMarkdown(text[last:start]) + codeSpan(text[start:end]))
		last = end
	}
	return b.String() + escapeMarkdown(text[last:])
}

func literalSpans(text string) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		for i := 0; i < len(text); i++ {
			end := literalEnd(text, i)
			if end <= i {
				continue
			}
			if !yield(i, end) {
				return
			}
			i = end - 1
		}
	}
}

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

func valueEnd(text string, end int) int {
	if end >= len(text) || text[end] != '=' {
		return end
	}
	value := end + 1
	for value < len(text) && !strings.ContainsRune(" \t,;)", rune(text[value])) {
		value++
	}
	for value > end+1 && text[value-1] == '.' {
		value--
	}
	return value
}

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

func quotedEnd(text string, i int) int {
	quote := text[i]
	closing := strings.IndexByte(text[i+1:], quote)
	if closing <= 0 {
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

func isWordByte(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '-'
}

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

func escapeLineStart(text string) string {
	switch {
	case text == "":
		return text
	case strings.ContainsRune("-+!", rune(text[0])):
		return "\\" + text
	case text[0] == '=':
		return "&#61;" + text[1:]
	case text[0] == '?':
		return "&#63;" + text[1:]
	}
	digits := len(text) - len(strings.TrimLeft(text, "0123456789"))
	if digits > 0 && digits < len(text) && text[digits] == '.' {
		return text[:digits] + "\\" + text[digits:]
	}
	return text
}

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

func fence(text string) string { return fenced("text", text) }

func shellFence(text string) string { return fenced("sh", text) }

func fenced(language, text string) string {
	marks := "```"
	for strings.Contains(text, marks) {
		marks += "`"
	}
	return marks + language + "\n" + text + "\n" + marks
}

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

func TestReferenceRendersHelpFaithfully(t *testing.T) {
	root := newTestRoot()
	for _, tt := range []struct{ help, want string }{
		{help: `a <b> & *c* _d_ [e](f) #g {h} |i| \j`, want: `a &lt;b\> &amp; \*c\* \_d\_ \[e\](f) \#g \{h\} \|i\| \\j`},
		{
			help: "x\n\n- y\n\n= z\n\n? w\n\n12. v\n\n!!! u\n\n3 t",
			want: "x\n\n\\- y\n\n&#61; z\n\n&#63; w\n\n12\\. v\n\n\\!!! u\n\n3 t",
		},
		{help: "a\n# b\n* c\n- d\n#### e", want: "a \\# b \\* c - d \\#\\#\\#\\# e"},
		{help: "x \"a\nb\" y, -o\njson", want: "x `\"a b\"` y, `-o json`"},
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

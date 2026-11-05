package cli

import (
	"iter"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// Help text is written as plain prose and wrapped in one place: the help of
// every command shows its long help reflowed (longHelp, reflowHelp), so the
// fragments that builders join never carry line breaks of their own.

// helpWidth is the most characters that a line of long help takes, unless it
// is one word, which no width splits. Lines filled up to 95 characters look
// as the help did when it was wrapped by hand, and fit a terminal of 100
// columns.
const helpWidth = 95

// longHelp returns the long help of cmd as its help shows it: reflowed to
// helpWidth (reflow).
func longHelp(cmd *cobra.Command) string { return reflow(cmd.Long, helpWidth) }

// reflowHelp makes the help of root, which every command below it uses,
// show the long help reflowed (longHelp); Long stays as written. The help is
// reflowed only when it is shown, as a command's long help is finished by
// then and a run that shows none spends nothing on it.
func reflowHelp(root *cobra.Command) {
	show := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		long := cmd.Long
		cmd.Long = longHelp(cmd)
		show(cmd, args)
		cmd.Long = long
	})
}

// reflow returns text with each paragraph of prose wrapped to width: its
// lines are joined and split again between words (helpWords), never inside a
// backquoted span, so a word wider than width stands on a line of its own.
// Blank lines, which separate paragraphs, and lines that start with a space
// or a tab, such as examples, stay as they are and where they are (keptLine).
func reflow(text string, width int) string {
	out := helpLines{width: width}
	out.text.Grow(len(text))
	var prose []string // the lines of the paragraph so far
	for line := range strings.SplitSeq(text, "\n") {
		if !keptLine(line) {
			prose = append(prose, line)
			continue
		}
		out.fill(strings.Join(prose, " "))
		prose = prose[:0]
		out.line(line)
	}
	out.fill(strings.Join(prose, " "))
	return out.text.String()
}

// keptLine reports whether line of help text is kept as it is by reflow: a
// blank line or an indented one. No line that reflow fills is one, as a word
// never starts with white space.
func keptLine(line string) bool {
	return strings.TrimSpace(line) == "" || indented(line)
}

// indented reports whether line of help text starts with a space or a tab,
// as an example does: help shows it as written, and the documentation site
// as a text block.
func indented(line string) bool { return line != "" && isBlankByte(line[0]) }

// helpLines is help text that reflow writes line by line, the lines
// separated by line breaks.
type helpLines struct {
	text  strings.Builder
	width int // the most characters a line of several words takes
	lines int // the lines written
}

// line writes a line.
func (h *helpLines) line(line string) {
	if h.lines > 0 {
		h.text.WriteByte('\n')
	}
	h.text.WriteString(line)
	h.lines++
}

// fill writes the words of paragraph on lines of as many words as fit in
// h.width characters, and at least one.
func (h *helpLines) fill(paragraph string) {
	used := 0 // the characters of the line being filled
	for word := range helpWords(paragraph) {
		size := utf8.RuneCountInString(word)
		if used > 0 && used+1+size <= h.width {
			h.text.WriteByte(' ')
			h.text.WriteString(word)
			used += 1 + size
		} else {
			h.line(word)
			used = size
		}
	}
}

// helpWords yields the words of help text, which spaces and tabs separate
// (wordEnd).
func helpWords(text string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for start := 0; start < len(text); {
			if isBlankByte(text[start]) {
				start++
				continue
			}
			end := wordEnd(text, start)
			if !yield(text[start:end]) {
				return
			}
			start = end
		}
	}
}

// wordEnd returns the end of the word of help text that starts at text[i]:
// the next space or tab, where a backquoted span belongs to the word it is
// part of, its spaces included, and a backquote that no other one closes is
// a character of its word.
func wordEnd(text string, i int) int {
	for ; i < len(text) && !isBlankByte(text[i]); i++ {
		if text[i] != '`' {
			continue
		}
		if closing := strings.IndexByte(text[i+1:], '`'); closing >= 0 {
			i += closing + 1
		}
	}
	return i
}

// isBlankByte reports whether c is a space or a tab.
func isBlankByte(c byte) bool { return c == ' ' || c == '\t' }

// joinSentences joins parts, sentences of help text, with a space between each
// two that are not empty: the separator belongs to the joining, not to the
// fragments. Use it where a part may be empty; elsewhere a space inside the
// literal joins two sentences, which TestLongHelpIsPlainProse checks.
func joinSentences(parts ...string) string { return joinNonEmpty(parts, " ") }

// joinParagraphs joins parts, paragraphs of help text, with a blank line
// between each two that are not empty, as joinSentences joins sentences.
func joinParagraphs(parts ...string) string { return joinNonEmpty(parts, "\n\n") }

// joinNonEmpty joins those of parts that are not empty with separator.
func joinNonEmpty(parts []string, separator string) string {
	var joined strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if joined.Len() > 0 {
			joined.WriteString(separator)
		}
		joined.WriteString(part)
	}
	return joined.String()
}

package cli

import (
	"iter"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

const helpWidth = 95

func longHelp(cmd *cobra.Command) string { return reflow(cmd.Long, helpWidth) }

func wrapHelp(root *cobra.Command) {
	show := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		long := cmd.Long
		cmd.Long = longHelp(cmd)
		show(cmd, args)
		cmd.Long = long
	})
	root.SetUsageTemplate(flagUsageTemplate(root.UsageTemplate()))
}

const (
	flagUsages        = ".FlagUsages "
	flagUsagesWrapped = ".FlagUsagesWrapped "
)

// pflag breaks a usage between words and aligns its lines under the first; a
// word wider than the room left stays whole, and so does the rest of the usage
// after it, on the same line. pflag counts bytes and keeps a line one short of
// the width, so the lines of a usage take at most helpWidth-1 characters.
func flagUsageTemplate(usage string) string {
	return strings.ReplaceAll(usage, flagUsages, flagUsagesWrapped+strconv.Itoa(helpWidth)+" ")
}

func reflow(text string, width int) string {
	out := helpLines{width: width}
	out.text.Grow(len(text))
	for chunk, kept := range helpChunks(text) {
		if kept {
			out.line(chunk)
		} else {
			out.fill(chunk)
		}
	}
	return out.text.String()
}

func helpChunks(text string) iter.Seq2[string, bool] {
	return func(yield func(string, bool) bool) {
		var prose []string
		for line := range strings.SplitSeq(text, "\n") {
			if !keptLine(line) {
				prose = append(prose, line)
				continue
			}
			if len(prose) > 0 && !yield(strings.Join(prose, " "), false) {
				return
			}
			prose = prose[:0]
			if !yield(line, true) {
				return
			}
		}
		if len(prose) > 0 {
			yield(strings.Join(prose, " "), false)
		}
	}
}

func keptLine(line string) bool {
	return strings.TrimSpace(line) == "" || indented(line)
}

func indented(line string) bool { return line != "" && isBlankByte(line[0]) }

type helpLines struct {
	text  strings.Builder
	width int
	lines int
}

func (h *helpLines) line(line string) {
	if h.lines > 0 {
		h.text.WriteByte('\n')
	}
	h.text.WriteString(line)
	h.lines++
}

func (h *helpLines) fill(paragraph string) {
	used := 0
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

func isBlankByte(c byte) bool { return c == ' ' || c == '\t' }

func literal(s string) string { return "`" + s + "`" }

func joinSentences(parts ...string) string { return joinNonEmpty(parts, " ") }

func joinParagraphs(parts ...string) string { return joinNonEmpty(parts, "\n\n") }

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

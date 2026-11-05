package cli

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The man pages are rendered from the command tree like the command reference
// (referencePages): epx.1 holds the root command, its global flags and the
// top-level commands, and each top-level command has a page, such as
// epx-units.1, with its help and a subsection per command below it. Pages are
// roff with the man macros; their text is the help text, set in paragraphs,
// with backquoted spans and literals such as --file in bold and indented
// lines kept as written.

// manDir holds the generated man pages, relative to this package.
const manDir = "../../man"

// manExt is the extension of the man pages: section 1, user commands.
const manExt = ".1"

// regenerateMan is the command that rewrites the man pages.
const regenerateMan = "go test ./internal/cli -run TestManPages -update"

// siteURL is the address of the documentation site.
const siteURL = "https://commerce-cloud-integrations.github.io/emporix-cli"

// TestManPages checks that the man pages match the command tree; with -update
// it rewrites them and removes pages no longer rendered.
func TestManPages(t *testing.T) {
	generated{dir: manDir, ext: manExt, regenerate: regenerateMan}.check(t, manPages(newTestRoot()))
}

// manPages renders the man pages of root by file name. Like referencePages,
// it first adds -h to every command, as Cobra does before it shows help.
func manPages(root *cobra.Command) map[string]string {
	walkCommands(root, (*cobra.Command).InitDefaultHelpFlag)
	pages := map[string]string{manName(root) + manExt: rootManPage(root)}
	for _, top := range shownCommands(root) {
		pages[manName(top)+manExt] = commandManPage(top)
	}
	return pages
}

// manName returns the name of cmd's man page: its path joined with "-".
func manName(cmd *cobra.Command) string { return anchorOf(cmd) }

// rootManPage renders the root command: its help, the global flags and the
// top-level commands, each named by its page.
func rootManPage(root *cobra.Command) string {
	var m manWriter
	m.title(root)
	m.section("SYNOPSIS")
	m.preformatted(usage(root))
	m.section("DESCRIPTION")
	m.help(longHelp(root))
	m.section("GLOBAL OPTIONS")
	m.paragraph("Every command takes these options.")
	m.flags(root.LocalFlags())
	m.section("COMMANDS")
	for _, top := range shownCommands(root) {
		m.item(manRef(top), roffProse(sentence(top.Short)))
	}
	m.section("SEE ALSO")
	m.paragraph("The documentation, with guides and the Emporix API coverage of each command: " + siteURL + "/")
	return m.String()
}

// commandManPage renders a top-level command: its help, then a subsection per
// command below it, in the order of help.
func commandManPage(top *cobra.Command) string {
	var m manWriter
	m.title(top)
	m.section("SYNOPSIS")
	m.preformatted(usage(top))
	m.section("DESCRIPTION")
	m.command(top)
	var below []*cobra.Command
	walkShown(top, func(cmd *cobra.Command) { below = append(below, cmd) })
	if len(below) > 1 {
		m.section("COMMANDS")
		for _, cmd := range below[1:] {
			m.heading("SS", cmd.CommandPath())
			m.paragraph(roffProse(sentence(cmd.Short)))
			m.preformatted(usage(cmd))
			m.command(cmd)
		}
	}
	m.section("SEE ALSO")
	m.paragraph(manRef(top.Root()) + " for the global options, the credentials epx reads and its exit codes; " +
		siteURL + pageURL(top))
	return m.String()
}

// manRef names cmd's man page as man pages cite each other: epx-units(1).
func manRef(cmd *cobra.Command) string {
	return bold(escapeRoff(manName(cmd))) + "(" + strings.TrimPrefix(manExt, ".") + ")"
}

// manWriter writes a man page. headed reports whether the last line written
// is a heading, after which a paragraph needs no macro of its own.
type manWriter struct {
	b      strings.Builder
	headed bool
}

func (m *manWriter) String() string { return m.b.String() }

// title writes the title line and the NAME section of cmd's page: its name
// and summary, which whatis and apropos read.
func (m *manWriter) title(cmd *cobra.Command) {
	name := manName(cmd)
	program := cmd.Root().Name()
	m.macro("TH", strings.ToUpper(name), strings.TrimPrefix(manExt, "."), "", program, program+" manual")
	m.section("NAME")
	m.line(escapeRoff(name) + ` \- ` + escapeRoff(summary(cmd.Short)))
}

// command writes the help of cmd, besides its summary and usage lines: its
// long help, aliases, flags (besides the global ones and -h), OAuth scopes
// and examples.
func (m *manWriter) command(cmd *cobra.Command) {
	m.help(longHelp(cmd))
	if len(cmd.Aliases) > 0 {
		m.paragraph(bold("Aliases:") + " " + escapeRoff(strings.Join(cmd.Aliases, ", ")))
	}
	if flags := shownFlags(cmd.LocalFlags(), "help"); len(flags) > 0 {
		m.paragraph(bold("Options:"))
		m.flags(cmd.LocalFlags(), "help")
	}
	if line := manScopes(cmd); line != "" {
		m.paragraph(line)
	}
	if cmd.Example != "" {
		m.paragraph(bold("Examples:"))
		m.preformatted(strings.Split(dedent(strings.Split(strings.TrimRight(cmd.Example, "\n"), "\n")), "\n"))
	}
}

// manScopes names the OAuth scopes of a read or a write (shownScopes), none
// where it requests none, and marks an opt-in read.
func manScopes(cmd *cobra.Command) string {
	scopes, optIn, shown := shownScopes(cmd)
	switch {
	case !shown:
		return ""
	case len(scopes) == 0:
		return bold("OAuth scopes:") + " none"
	case optIn:
		return bold("OAuth scopes:") + " " + escapeRoff(strings.Join(scopes, ", ")) +
			" (opt-in with " + bold(escapeRoff("--"+writeScopeFlag)) + ")"
	}
	return bold("OAuth scopes:") + " " + escapeRoff(strings.Join(scopes, ", "))
}

// help writes help text: runs of prose as paragraphs, a Markdown heading
// line such as Cobra's "#### Linux:" as a bold line, and runs of indented
// lines as written.
func (m *manWriter) help(text string) {
	for _, run := range helpRuns(text) {
		title, isHeading := helpHeading(run.lines[0])
		switch {
		case run.indented:
			m.preformatted(strings.Split(dedent(run.lines), "\n"))
		case len(run.lines) == 1 && isHeading:
			m.paragraph(bold(escapeRoff(title)))
		default:
			m.paragraph(roffProse(strings.Join(run.lines, " ")))
		}
	}
}

// flags writes the flags in flags that help shows, but those named in skip,
// as a tagged list: each flag, then its description and default.
func (m *manWriter) flags(flags *pflag.FlagSet, skip ...string) {
	for _, f := range shownFlags(flags, skip...) {
		text := roffProse(sentence(f.text))
		if f.value != "" {
			text += " Default: " + bold(escapeRoff(f.value)) + "."
		}
		m.item(bold(escapeRoff(f.name)), text)
	}
}

// item writes a tagged paragraph: term, then text indented below it.
func (m *manWriter) item(term, text string) {
	m.macro("TP")
	m.line(term)
	m.line(text)
}

func (m *manWriter) section(name string) { m.heading("SH", name) }

// heading writes a section (SH) or subsection (SS) heading.
func (m *manWriter) heading(macro, title string) {
	m.macro(macro, title)
	m.headed = true
}

// paragraph writes text, already roff, as a paragraph.
func (m *manWriter) paragraph(text string) {
	m.startParagraph()
	m.line(text)
}

// startParagraph starts a paragraph, unless a heading has just started one.
func (m *manWriter) startParagraph() {
	if !m.headed {
		m.macro("PP")
	}
}

// preformatted writes lines as written, indented, without filling.
func (m *manWriter) preformatted(lines []string) {
	m.startParagraph()
	m.macro("RS", "4")
	m.macro("nf")
	for _, line := range lines {
		m.line(escapeRoff(line))
	}
	m.macro("fi")
	m.macro("RE")
}

// macro writes a macro line, each argument quoted.
func (m *manWriter) macro(name string, args ...string) {
	m.b.WriteString("." + name)
	for _, arg := range args {
		m.b.WriteString(` "` + strings.ReplaceAll(escapeRoff(arg), `"`, `\(dq`) + `"`)
	}
	m.b.WriteString("\n")
	m.headed = false
}

// line writes a text line, roff already; a line that starts with a period or
// an apostrophe, which roff would read as a request, starts with \& instead.
func (m *manWriter) line(text string) {
	if strings.HasPrefix(text, ".") || strings.HasPrefix(text, "'") {
		text = `\&` + text
	}
	m.b.WriteString(text + "\n")
	m.headed = false
}

// summary returns a command's summary as the NAME section shows it: without
// a final period, and starting lower case unless it starts with an acronym.
func summary(short string) string {
	text := strings.TrimSuffix(strings.TrimSpace(short), ".")
	if len(text) > 1 && unicode.IsUpper(rune(text[0])) && !unicode.IsUpper(rune(text[1])) {
		text = string(unicode.ToLower(rune(text[0]))) + text[1:]
	}
	return text
}

// roffProse renders prose as roff that shows the same text: backquoted spans
// in bold, without their backquotes, literals such as --file in bold
// (markLiterals) and the rest escaped. Prose with an unmatched backquote shows
// its backquotes as written.
func roffProse(text string) string {
	parts := strings.Split(text, "`")
	if len(parts)%2 == 0 {
		return roffLiterals(text)
	}
	var b strings.Builder
	for i, part := range parts {
		switch {
		case i%2 == 0:
			b.WriteString(roffLiterals(part))
		case part == "":
			b.WriteString(escapeRoff("``"))
		default:
			b.WriteString(bold(escapeRoff(part)))
		}
	}
	return b.String()
}

// roffLiterals renders prose with its literals (literalSpans) in bold and the rest escaped.
func roffLiterals(text string) string {
	var b strings.Builder
	last := 0
	for start, end := range literalSpans(text) {
		b.WriteString(escapeRoff(text[last:start]) + bold(escapeRoff(text[start:end])))
		last = end
	}
	return b.String() + escapeRoff(text[last:])
}

func bold(text string) string { return `\fB` + text + `\fR` }

// escapeRoff escapes the characters of text that roff would interpret or
// change: a backslash, a hyphen (which roff may set as a hyphen, not the
// minus of a flag), a backquote and an apostrophe (which it may set as
// typographic quotes), and every character outside ASCII, named by its code
// point.
func escapeRoff(text string) string {
	var b strings.Builder
	for _, r := range text {
		switch {
		case r == '\\':
			b.WriteString(`\e`)
		case r == '-':
			b.WriteString(`\-`)
		case r == '`':
			b.WriteString(`\(ga`)
		case r == '\'':
			b.WriteString(`\(aq`)
		case r > unicode.MaxASCII:
			fmt.Fprintf(&b, `\[u%04X]`, r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TestManRendersHelpFaithfully checks that help text shows in a man page as
// written: roff's special characters are escaped, backquoted spans and
// literals are bold and indented lines keep their layout.
func TestManRendersHelpFaithfully(t *testing.T) {
	for _, tt := range []struct{ help, want string }{
		{help: `a \b c-d 'e' – é`, want: ".PP\na \\eb c\\-d \\fB\\(aqe\\(aq\\fR \\[u2013] \\[u00E9]\n"},
		{help: "run `epx sites\nlist` with --yes, -o json and\n{\"a\": 1}", want: ".PP\nrun \\fBepx sites list\\fR with " +
			"\\fB\\-\\-yes\\fR, \\fB\\-o json\\fR and \\fB{\"a\": 1}\\fR\n"},
		{help: "odd ` quote, ``", want: ".PP\nodd \\(ga quote, \\(ga\\(ga\n"},
		{help: "a `` b `c`", want: ".PP\na \\(ga\\(ga b \\fBc\\fR\n"},
		{help: ".hidden request\n\n'quoted", want: ".PP\n\\&.hidden request\n.PP\n\\(aqquoted\n"},
		{help: "#### Linux:\n\ntext", want: ".PP\n\\fBLinux:\\fR\n.PP\ntext\n"},
		{help: "intro:\n\n    a\n      .b", want: ".PP\nintro:\n.PP\n.RS \"4\"\n.nf\na\n  .b\n.fi\n.RE\n"},
	} {
		var m manWriter
		m.help(tt.help)
		if got := m.String(); got != tt.want {
			t.Errorf("help %q renders as\n%s\nwant\n%s", tt.help, got, tt.want)
		}
	}
	for short, want := range map[string]string{"Read units.": "read units", "IAM users": "IAM users", "x": "x"} {
		if got := summary(short); got != want {
			t.Errorf("summary(%q) = %q, want %q", short, got, want)
		}
	}
}

// TestManPagesNameEveryCommand checks that the pages cover the command tree:
// every shown command has a subsection or a page, and the root page names the
// page of every top-level command.
func TestManPagesNameEveryCommand(t *testing.T) {
	root := newTestRoot()
	pages := manPages(root)
	walkShown(root, func(cmd *cobra.Command) {
		top := topCommand(cmd)
		switch {
		case top == nil:
		case cmd == top:
			if !strings.Contains(pages[manName(root)+manExt], manRef(cmd)) {
				t.Errorf("%s does not name %s", manName(root)+manExt, manName(cmd)+manExt)
			}
		case !strings.Contains(pages[manName(top)+manExt], `.SS "`+escapeRoff(cmd.CommandPath())+`"`):
			t.Errorf("%s has no subsection for %s", manName(top)+manExt, cmd.CommandPath())
		}
	})
}

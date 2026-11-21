package cli

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const manDir = "../../man"

const manExt = ".1"

const regenerateMan = "go test ./internal/cli -run TestManPages -update"

const siteURL = "https://commerce-cloud-integrations.github.io/emporix-cli"

func TestManPages(t *testing.T) {
	generated{dir: manDir, ext: manExt, regenerate: regenerateMan}.check(t, manPages(newTestRoot()))
}

func manPages(root *cobra.Command) map[string]string {
	walkCommands(root, (*cobra.Command).InitDefaultHelpFlag)
	pages := map[string]string{manName(root) + manExt: rootManPage(root)}
	for _, top := range shownCommands(root) {
		pages[manName(top)+manExt] = commandManPage(top)
	}
	return pages
}

func manName(cmd *cobra.Command) string { return anchorOf(cmd) }

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

func manRef(cmd *cobra.Command) string {
	return bold(escapeRoff(manName(cmd))) + "(" + strings.TrimPrefix(manExt, ".") + ")"
}

type manWriter struct {
	b      strings.Builder
	headed bool
}

func (m *manWriter) String() string { return m.b.String() }

func (m *manWriter) title(cmd *cobra.Command) {
	name := manName(cmd)
	program := cmd.Root().Name()
	m.macro("TH", strings.ToUpper(name), strings.TrimPrefix(manExt, "."), "", program, program+" manual")
	m.section("NAME")
	m.line(escapeRoff(name) + ` \- ` + escapeRoff(summary(cmd.Short)))
}

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

func (m *manWriter) flags(flags *pflag.FlagSet, skip ...string) {
	for _, f := range shownFlags(flags, skip...) {
		text := roffProse(sentence(f.text))
		if f.value != "" {
			text += " Default: " + bold(escapeRoff(f.value)) + "."
		}
		m.item(bold(escapeRoff(f.name)), text)
	}
}

func (m *manWriter) item(term, text string) {
	m.macro("TP")
	m.line(term)
	m.line(text)
}

func (m *manWriter) section(name string) { m.heading("SH", name) }

func (m *manWriter) heading(macro, title string) {
	m.macro(macro, title)
	m.headed = true
}

func (m *manWriter) paragraph(text string) {
	m.startParagraph()
	m.line(text)
}

func (m *manWriter) startParagraph() {
	if !m.headed {
		m.macro("PP")
	}
}

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

func (m *manWriter) macro(name string, args ...string) {
	m.b.WriteString("." + name)
	for _, arg := range args {
		m.b.WriteString(` "` + strings.ReplaceAll(escapeRoff(arg), `"`, `\(dq`) + `"`)
	}
	m.b.WriteString("\n")
	m.headed = false
}

func (m *manWriter) line(text string) {
	if strings.HasPrefix(text, ".") || strings.HasPrefix(text, "'") {
		text = `\&` + text
	}
	m.b.WriteString(text + "\n")
	m.headed = false
}

func summary(short string) string {
	text := strings.TrimSuffix(strings.TrimSpace(short), ".")
	if len(text) > 1 && unicode.IsUpper(rune(text[0])) && !unicode.IsUpper(rune(text[1])) {
		text = string(unicode.ToLower(rune(text[0]))) + text[1:]
	}
	return text
}

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

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const referenceDir = "../../site/commands"

const pageExt = ".mdx"

const referenceURL = "/commands"

const regenerateReference = "go test ./internal/cli -run TestCommandReference -update"

const updateFlag = "update"

const globalFlagsLink = "[global flags](" + referenceURL + "#global-flags)"

type generated struct {
	dir, ext, regenerate string
}

func TestCommandReference(t *testing.T) {
	generated{dir: referenceDir, ext: pageExt, regenerate: regenerateReference}.check(t, referencePages(newTestRoot()))
}

func (g generated) check(t *testing.T, pages map[string]string) {
	t.Helper()
	if flag.Lookup(updateFlag).Value.String() == "true" {
		g.write(t, pages)
		return
	}
	if stale := g.stale(t, pages); len(stale) > 0 {
		t.Fatalf("%s is out of date:\n  %s\nregenerate it with: %s",
			g.shown(""), strings.Join(stale, "\n  "), g.regenerate)
	}
}

func (g generated) stale(t *testing.T, pages map[string]string) []string {
	t.Helper()
	dir := os.DirFS(g.dir)
	var stale []string
	for _, name := range slices.Sorted(maps.Keys(pages)) {
		got, err := fs.ReadFile(dir, name)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			stale = append(stale, "missing "+g.shown(name))
		case err != nil:
			t.Fatal(err)
		case string(got) != pages[name]:
			stale = append(stale, "different "+g.shown(name))
		}
	}
	for _, name := range g.extra(t, pages) {
		stale = append(stale, "extra "+g.shown(name))
	}
	return stale
}

func (g generated) shown(name string) string {
	return path.Join(strings.TrimPrefix(g.dir, "../../"), name)
}

func (g generated) write(t *testing.T, pages map[string]string) {
	t.Helper()
	if err := os.MkdirAll(g.dir, 0o750); err != nil {
		t.Fatal(err)
	}
	dir, err := os.OpenRoot(g.dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = dir.Close() }()
	for _, name := range g.extra(t, pages) {
		if err := dir.Remove(name); err != nil {
			t.Fatal(err)
		}
	}
	for name, page := range pages {
		if err := dir.WriteFile(name, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func (g generated) extra(t *testing.T, pages map[string]string) []string {
	t.Helper()
	entries, err := os.ReadDir(g.dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	var extra []string
	for _, entry := range entries {
		if _, rendered := pages[entry.Name()]; !rendered && !entry.IsDir() && strings.HasSuffix(entry.Name(), g.ext) {
			extra = append(extra, entry.Name())
		}
	}
	return extra
}

func referencePages(root *cobra.Command) map[string]string {
	walkCommands(root, (*cobra.Command).InitDefaultHelpFlag)
	pages := map[string]string{"index" + pageExt: rootPage(root)}
	for _, top := range shownCommands(root) {
		blocks := []string{frontMatter(top.CommandPath(), top.Name(), sentence(top.Short))}
		walkShown(top, func(cmd *cobra.Command) { blocks = append(blocks, section(cmd)...) })
		pages[referencePage(top)] = strings.Join(blocks, "\n\n") + "\n"
	}
	return pages
}

func rootPage(root *cobra.Command) string {
	blocks := []string{
		frontMatter("CLI reference", "Overview",
			"Every "+root.Name()+" command and flag, with the OAuth scopes each one requests."),
		"This reference documents every command and flag of " + root.Name() + ". The same text is in " +
			codeSpan(root.Name()+" <command> --help") + ".",
	}
	blocks = append(blocks, helpBlocks(longHelp(root), root)...)
	blocks = append(blocks,
		"## Global flags",
		"Every command takes these flags.",
		flagList(flagItems(root.LocalFlags(), root)),
		"## Commands",
		commandList(shownCommands(root), pageURL))
	return strings.Join(blocks, "\n\n") + "\n"
}

func section(cmd *cobra.Command) []string {
	root := cmd.Root()
	blocks := []string{"## " + cmd.CommandPath()}
	if cmd.Parent() != root {
		blocks = append(blocks, paragraph(sentence(cmd.Short), root))
	}
	blocks = append(blocks, helpBlocks(longHelp(cmd), root)...)
	blocks = append(blocks, "**Command:** "+codeList(usage(cmd)))
	if len(cmd.Aliases) > 0 {
		blocks = append(blocks, "**Aliases:** "+codeList(cmd.Aliases))
	}
	blocks = append(blocks, flagBlocks(cmd)...)
	if line := scopesLine(cmd); line != "" {
		blocks = append(blocks, line)
	}
	if cmd.Example != "" {
		blocks = append(blocks, "**Examples**", shellFence(dedent(strings.Split(strings.TrimRight(cmd.Example, "\n"), "\n"))))
	}
	if subs := shownCommands(cmd); len(subs) > 0 {
		anchor := func(sub *cobra.Command) string { return "#" + anchorOf(sub) }
		blocks = append(blocks, "**Subcommands**", commandList(subs, anchor))
	}
	return blocks
}

func flagBlocks(cmd *cobra.Command) []string {
	items := flagItems(cmd.LocalFlags(), cmd.Root(), "help")
	if len(items) == 0 {
		return []string{"**Flags:** only the " + globalFlagsLink + "."}
	}
	return []string{"**Flags**", flagList(items), "Also takes the " + globalFlagsLink + "."}
}

func usage(cmd *cobra.Command) []string {
	var lines []string
	if cmd.Runnable() {
		lines = append(lines, cmd.UseLine())
	}
	if cmd.HasAvailableSubCommands() {
		lines = append(lines, cmd.CommandPath()+" [command]")
	}
	return lines
}

func scopesLine(cmd *cobra.Command) string {
	scopes, optIn, shown := shownScopes(cmd)
	switch {
	case !shown:
		return ""
	case len(scopes) == 0:
		return "**OAuth scopes:** none"
	case optIn:
		return "**OAuth scopes:** " + codeList(scopes) + " (opt-in with `--" + writeScopeFlag + "`)"
	}
	return "**OAuth scopes:** " + codeList(scopes)
}

func shownScopes(cmd *cobra.Command) (scopes []string, optIn, shown bool) {
	value, isRead := cmd.Annotations[scopesAnnotation]
	sentence := scopeSentence
	if !isRead {
		var isWrite bool
		if value, isWrite = cmd.Annotations[writeScopesAnnotation]; !isWrite {
			return nil, false, false
		}
		sentence = writeScopeSentence
	}
	scopes = strings.Fields(value)
	if strings.HasSuffix(strings.TrimSpace(cmd.Long), sentence(scopes)) {
		return nil, false, false
	}
	return scopes, isRead && len(scopes) > 0 && writeCapable(scopes), true
}

func flagItems(flags *pflag.FlagSet, root *cobra.Command, skip ...string) []string {
	var items []string
	for _, f := range shownFlags(flags, skip...) {
		items = append(items, listItem(codeSpan(f.name), paragraph(sentence(f.text), root)+defaultNote(f.value)))
	}
	return items
}

type shownFlag struct{ name, text, value string }

func shownFlags(flags *pflag.FlagSet, skip ...string) []shownFlag {
	var shown []shownFlag
	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || slices.Contains(skip, f.Name) {
			return
		}
		name := "--" + f.Name
		if f.Shorthand != "" {
			name = "-" + f.Shorthand + ", " + name
		}
		kind, text := pflag.UnquoteUsage(f)
		if kind != "" {
			name += " <" + kind + ">"
		}
		shown = append(shown, shownFlag{name: name, text: text, value: shownDefault(f.DefValue)})
	})
	return shown
}

func defaultNote(value string) string {
	if value == "" {
		return ""
	}
	return " Default: " + codeSpan(value) + "."
}

func shownDefault(value string) string {
	if slices.Contains([]string{"", "0", "0s", "false", "[]", "<nil>"}, value) {
		return ""
	}
	if strings.HasSuffix(value, "m0s") && strings.Trim(value, "0123456789hms") == "" {
		value = strings.TrimSuffix(strings.TrimSuffix(value, "0s"), "h0m")
	}
	return value
}

func listItem(term, description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return "- " + term
	}
	return "- " + term + "\\\n  " + description
}

func flagList(items []string) string { return strings.Join(items, "\n") }

func commandList(cmds []*cobra.Command, link func(*cobra.Command) string) string {
	items := make([]string, len(cmds))
	for i, cmd := range cmds {
		term := fmt.Sprintf("[%s](%s)", codeSpan(cmd.CommandPath()), link(cmd))
		items[i] = listItem(term, paragraph(sentence(cmd.Short), cmd.Root()))
	}
	return strings.Join(items, "\n")
}

func sentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	if 'a' <= text[0] && text[0] <= 'z' {
		text = string(text[0]-'a'+'A') + text[1:]
	}
	if !strings.ContainsAny(text[len(text)-1:], ".!?:") {
		text += "."
	}
	return text
}

func shownCommands(cmd *cobra.Command) []*cobra.Command {
	var shown []*cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && sub.Name() != "help" {
			shown = append(shown, sub)
		}
	}
	return shown
}

func walkShown(cmd *cobra.Command, visit func(*cobra.Command)) {
	visit(cmd)
	for _, sub := range shownCommands(cmd) {
		walkShown(sub, visit)
	}
}

func referencePage(cmd *cobra.Command) string {
	if top := topCommand(cmd); top != nil {
		return top.Name() + pageExt
	}
	return "index" + pageExt
}

func pageURL(cmd *cobra.Command) string {
	if top := topCommand(cmd); top != nil {
		return referenceURL + "/" + top.Name()
	}
	return referenceURL
}

func topCommand(cmd *cobra.Command) *cobra.Command {
	for ; cmd.HasParent(); cmd = cmd.Parent() {
		if !cmd.Parent().HasParent() {
			return cmd
		}
	}
	return nil
}

func frontMatter(title, sidebarTitle, description string) string {
	return fmt.Sprintf("---\ntitle: %s\nsidebarTitle: %s\ndescription: %s\n---",
		strconv.Quote(title), strconv.Quote(sidebarTitle), strconv.Quote(description))
}

func anchorOf(cmd *cobra.Command) string {
	return strings.ReplaceAll(cmd.CommandPath(), " ", "-")
}

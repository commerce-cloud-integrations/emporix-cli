package cli

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// usageSections returns the sections of help that Cobra's usage template
// writes and that TestUsageFitsTheWidth checks: the usage lines and the
// flags. The other sections hold text that epx writes, checked where it is
// written (the long help and its examples) or as written (the summaries of
// subcommands).
func usageSections() []string { return []string{"Usage:", "Flags:", "Global Flags:"} }

// TestFlagUsageTemplateWrapsOnlyTheFlags checks that the usage template of
// epx is Cobra's with both lists of flags wrapped to helpWidth, and nothing
// else changed.
func TestFlagUsageTemplateWrapsOnlyTheFlags(t *testing.T) {
	cobraUsage := (&cobra.Command{}).UsageTemplate()
	usage := newTestRoot().UsageTemplate()
	wrapped := fmt.Sprintf(".FlagUsagesWrapped %d ", helpWidth)
	if usage != flagUsageTemplate(cobraUsage) || strings.Count(usage, wrapped+"|") != 2 ||
		!strings.Contains(usage, ".LocalFlags"+wrapped) || !strings.Contains(usage, ".InheritedFlags"+wrapped) ||
		strings.Contains(usage, ".FlagUsages |") || strings.ReplaceAll(usage, wrapped, ".FlagUsages ") != cobraUsage {
		t.Errorf("usage template:\n%s\nwant Cobra's with both flag lists wrapped to %d:\n%s", usage, helpWidth, cobraUsage)
	}
}

// TestUsageFitsTheWidth checks the help of every command of the tree as
// --help shows it, with the flag -h, --help that Cobra adds when it runs: no
// line of its usage, flags or global flags is wider than helpWidth unless it
// is one word (overWidthUsage), and each flag shows the words of its usage as
// pflag writes them unwrapped, in the same order.
func TestUsageFitsTheWidth(t *testing.T) {
	root := newTestRoot()
	var out strings.Builder
	root.SetOut(&out)
	walkCommands(root, func(cmd *cobra.Command) {
		out.Reset()
		cmd.InitDefaultHelpFlag()
		cmd.HelpFunc()(cmd, nil)
		sections := helpSections(out.String())
		for _, name := range usageSections() {
			for _, line := range overWidthUsage(sections[name], helpWidth) {
				t.Errorf("%s: %s line of %d characters, more than %d: %q", cmd.CommandPath(), name,
					utf8.RuneCountInString(line), helpWidth, line)
			}
		}
		lists := map[string]*pflag.FlagSet{"Flags:": cmd.LocalFlags(), "Global Flags:": cmd.InheritedFlags()}
		for name, flags := range lists {
			got, want := flagEntries(sections[name]), flagEntries(strings.Split(flags.FlagUsages(), "\n"))
			if !slices.Equal(got, want) {
				t.Errorf("%s: %s %q; want the words of %q", cmd.CommandPath(), name, got, want)
			}
		}
	})
}

// TestUsageWrapsAWideFlagList checks the help of a command whose flags are
// wider than helpWidth: each usage wraps under itself, and a flag that fits
// stays on one line. A word wider than the room left stays whole: at the end
// of a usage on a line of its own, within one with the rest of the usage on
// its line, as pflag wraps, which overWidthUsage reports.
func TestUsageWrapsAWideFlagList(t *testing.T) {
	root := &cobra.Command{Use: "epx"}
	root.PersistentFlags().String("tenant", "", strings.Repeat("tenant words ", 10))
	wrapHelp(root)
	cmd := &cobra.Command{Use: "wide", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("short", "", "fits on its line")
	cmd.Flags().String("long", "", strings.Repeat("several words ", 10)+strings.Repeat("x", 80))
	cmd.Flags().String("mid", "", "some words "+strings.Repeat("y", 80)+strings.Repeat(" and the rest", 3))
	root.AddCommand(cmd)
	var out strings.Builder
	root.SetOut(&out)
	if err := cmd.Usage(); err != nil {
		t.Fatal(err)
	}
	sections := helpSections(out.String())
	want := map[string][]string{
		"Flags:": {
			"      --long string    several words several words several words several words several",
			"                       words several words several words several words several words",
			"                       several words",
			"                       " + strings.Repeat("x", 80),
			"      --mid string     some words",
			"                       " + strings.Repeat("y", 80) + " and the rest and the rest and the rest",
			"      --short string   fits on its line",
		},
		"Global Flags:": {
			"      --tenant string   tenant words tenant words tenant words tenant words tenant words",
			"                        tenant words tenant words tenant words tenant words tenant words",
		},
	}
	for name, lines := range want {
		if got := sections[name]; !slices.Equal(got, lines) {
			t.Errorf("%s:\n%s\nwant\n%s", name, strings.Join(got, "\n"), strings.Join(lines, "\n"))
		}
	}
	wide := slices.Concat(overWidthUsage(want["Flags:"], helpWidth), overWidthUsage(want["Global Flags:"], helpWidth))
	if !slices.Equal(wide, want["Flags:"][5:6]) {
		t.Errorf("lines over the width %q; want only the line of a wide word within a usage", wide)
	}
}

// TestOverWidthUsageFindsWideLines checks the check of TestUsageFitsTheWidth.
func TestOverWidthUsageFindsWideLines(t *testing.T) {
	word := strings.Repeat("x", 40)
	for _, tt := range []struct {
		name string
		line string
		wide bool
	}{
		{name: "fits", line: "  --a string   one two", wide: false},
		{name: "exact width", line: "  --a string   " + strings.Repeat("y", 15), wide: false},
		{name: "one over", line: "  --a string   " + strings.Repeat("y", 16), wide: true},
		{name: "runes", line: "      " + strings.Repeat("…", 20) + " …", wide: false},
		{name: "words over", line: "  --a string   one two three four five", wide: true},
		{name: "one word over", line: "                " + word, wide: false},
		{name: "one word then another", line: "  " + word + " x", wide: true},
		{name: "flag and wide word", line: "  --a " + word, wide: true},
		{name: "empty", line: "", wide: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := overWidthUsage([]string{tt.line}, 30); (len(got) == 1) != tt.wide {
				t.Errorf("overWidthUsage(%q) = %q, want wide %v", tt.line, got, tt.wide)
			}
		})
	}
}

// TestHelpSectionsSplitsAtHeadings checks the parsing of TestUsageFitsTheWidth.
func TestHelpSectionsSplitsAtHeadings(t *testing.T) {
	help := "Long help.\n\nUsage:\n  epx a [flags]\n\nFlags:\n  -h, --help   help for a\n" +
		"      --x int  an x\n               more\n\nGlobal Flags:\n      --tenant string   tenant\n\n" +
		"Use \"epx a [command] --help\" for more.\n"
	sections := helpSections(help)
	want := map[string][]string{
		"":              {"Long help.", "", "Use \"epx a [command] --help\" for more."},
		"Usage:":        {"  epx a [flags]"},
		"Flags:":        {"  -h, --help   help for a", "      --x int  an x", "               more"},
		"Global Flags:": {"      --tenant string   tenant"},
	}
	if len(sections) != len(want) {
		t.Errorf("helpSections = %q, want %q", sections, want)
	}
	for name, lines := range want {
		if !slices.Equal(sections[name], lines) {
			t.Errorf("section %q = %q, want %q", name, sections[name], lines)
		}
	}
	entries := flagEntries(sections["Flags:"])
	if !slices.Equal(entries, []string{"-h, --help help for a", "--x int an x more"}) {
		t.Errorf("flagEntries = %q", entries)
	}
	for flag, want := range map[string]string{"--help": "help for a", "--x": "int an x more", "--tenant": ""} {
		if usage, shown := flagUsage(help, flag); usage != want || shown != (want != "") {
			t.Errorf("flagUsage(%s) = %q, %v; want %q", flag, usage, shown, want)
		}
	}
	if got := flagEntries([]string{"  stray line"}); !slices.Equal(got, []string{"stray line"}) {
		t.Errorf("flagEntries of a line before any flag = %q", got)
	}
}

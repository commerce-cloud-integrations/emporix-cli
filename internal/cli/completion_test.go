package cli

import (
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestCompletionScripts(t *testing.T) {
	for shell, marker := range map[string]string{
		"bash": "__start_epx", "zsh": "#compdef epx", "fish": "complete -c epx",
		"powershell": "# powershell completion for epx",
	} {
		code, stdout, stderr := runCLI(t, nil, "completion", shell)
		if code != ExitOK || !strings.Contains(stdout, marker) || stderr != "" {
			t.Errorf("%s: exit %d, stderr %q, script without %q", shell, code, stderr, marker)
		}
	}
}

func directiveLine(directive cobra.ShellCompDirective) string {
	return ":" + strconv.Itoa(int(directive)) + "\n"
}

func TestCompletionNeverSuggestsFiles(t *testing.T) {
	noFiles := directiveLine(cobra.ShellCompDirectiveNoFileComp)
	tests := []struct {
		args []string
		want []string
	}{
		{args: []string{"products", "list", "-o", ""}, want: []string{"table\n", "json\n", "yaml\n", "csv\n"}},
		{args: []string{"auth", "status", "--output", "j"}, want: []string{"table\n", "json\n"}},
		{args: []string{"api", "get", "/x", "-o", ""}, want: []string{"json\n", "yaml\n"}},
		{args: []string{"products", "list", "--columns", ""}},
		{args: []string{"s"}, want: []string{"schemas\t", "sites\t"}},
		{args: []string{"indexing", ""}, want: []string{"providers\t", "reindex-jobs\t"}},
		{args: []string{"help", ""}, want: []string{"products\t", "sites\t"}},
		{args: []string{"help", "indexing", "re"}, want: []string{"reindex-jobs\t"}},
		{args: []string{"products", "get", "--help="}},
		{args: []string{"help", "indexing", "bogus", ""}},
		{args: []string{"auth", "foo", ""}},
		{args: []string{"products", "get", ""}},
		{args: []string{"api", "get", ""}},
		{args: []string{"--tenant", ""}},
		{args: []string{"prices", "list", "--site", ""}},
		{args: []string{"products", "list", "--language", ""}},
		{args: []string{"ai", "analytics", "executions", "helper", "--granularity", ""},
			want: []string{"QUARTER\n", "MONTH\n", "WEEK\n"}},
		{args: []string{"category-trees", "list", "--type", ""}, want: []string{"STANDARD\n", "CLASSIFICATION\n"}},
		{args: []string{"indexing", "reindex-jobs", "create", "--entity-type", ""},
			want: entityBuiltInTypeNames()},
		// --param repeats, so completion offers it again.
		{args: []string{"api", "get", "/x", "-p", "a=1", "--pa"}, want: []string{"--param\t"}},
	}
	for _, tt := range tests {
		code, stdout, _ := runCLI(t, nil, append([]string{"__complete"}, tt.args...)...)
		if code != ExitOK || !strings.Contains(stdout, noFiles) {
			t.Errorf("%v: exit %d, completions %q; want no file completion", tt.args, code, stdout)
		}
		if tt.want == nil && strings.Contains(strings.TrimSuffix(stdout, noFiles), "\t") {
			t.Errorf("%v: completions %q; want none", tt.args, stdout)
		}
		for _, want := range tt.want {
			if !strings.Contains(stdout, want) {
				t.Errorf("%v: completions %q lack %q", tt.args, stdout, want)
			}
		}
	}
}

func TestOnlyACompletionRequestRegistersFlagCompletions(t *testing.T) {
	for _, args := range [][]string{
		{"version"}, {"products", "list", "-o", "json"}, {"completion", "bash"}, {"products", "get", "__complete"},
		{"products", "search", "--body", "__completeNoDesc"}, {"api", "get", "/x", "--", "__complete"},
		{"help", "__complete"},
	} {
		if got, _ := registeredAfter(args...); len(got) > 0 {
			t.Errorf("%q: registered %d flag completions, first %s; want none", args, len(got), got[0])
		}
	}
	for _, request := range []string{cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd} {
		got, want := registeredAfter("--timeout", "5s", request, "")
		if !slices.Equal(got, want) {
			t.Errorf("%s: registered %q; want %q", request, got, want)
		}
		for _, flag := range []string{"epx --profile", "epx products list --output", "epx prices list --site",
			"epx api post --file", "epx products search --body", "epx quotes pdf --out"} {
			if !slices.Contains(got, flag) {
				t.Errorf("%s: no completion of %s", request, flag)
			}
		}
	}
}

func registeredAfter(args ...string) (registered, completing []string) {
	a := &app{getenv: func(string) string { return "" }, ProfileEnv: *noProfiles()}
	root := newRootCommand(a, io.Discard, io.Discard)
	root.SetArgs(args)
	_ = root.Execute()
	registered = flagsWhere(root, func(cmd *cobra.Command, flag *pflag.Flag) bool {
		_, ok := cmd.GetFlagCompletionFunc(flag.Name)
		return ok
	})
	completing = flagsWhere(root, func(_ *cobra.Command, flag *pflag.Flag) bool {
		return flagCompletion(a, flag, sites().ids()) != nil
	})
	return registered, completing
}

func TestCompletionAfterRootFlags(t *testing.T) {
	code, stdout, _ := runCLI(t, nil, "--timeout=5s", cobra.ShellCompNoDescRequestCmd, "quotes", "pdf", "Q1", "--out", "")
	if code != ExitOK || stdout != directiveLine(cobra.ShellCompDirectiveDefault) {
		t.Errorf("--out: exit %d, completions %q; want file names", code, stdout)
	}
	code, stdout, _ = runCLI(t, nil, "--timeout", "5s", cobra.ShellCompRequestCmd, "products", "list", "-o", "")
	if code != ExitOK || stdout != "table\njson\nyaml\ncsv\n"+directiveLine(cobra.ShellCompDirectiveNoFileComp) {
		t.Errorf("-o: exit %d, completions %q; want the formats", code, stdout)
	}
}

func TestPersistentFlagsComplete(t *testing.T) {
	root := &cobra.Command{Use: "epx"}
	group := &cobra.Command{Use: "group"}
	root.AddCommand(group)
	group.PersistentFlags().String("format", "", "")
	completeChoices(group, "format", []string{"a", "b"})
	group.PersistentFlags().String("in", "", "")
	completeFileNames(group, "in")
	if err := completeFlags(&app{}, root, sites().ids()); err != nil {
		t.Fatal(err)
	}
	for flag, want := range map[string][]string{"format": {"a", "b"}, "in": nil} {
		complete, ok := group.GetFlagCompletionFunc(flag)
		if !ok {
			t.Fatalf("--%s: no completion", flag)
		}
		if got, _ := complete(group, nil, ""); !slices.Equal(got, want) {
			t.Errorf("--%s: completions %q; want %q", flag, got, want)
		}
	}
}

func TestFlagCompletionThatFailsIsAnError(t *testing.T) {
	root := &cobra.Command{Use: "epx"}
	request := &cobra.Command{Use: cobra.ShellCompRequestCmd}
	root.AddCommand(request)
	root.Flags().String("format", "", "")
	completeChoices(root, "format", []string{"a", "b"})
	if err := completeFlagsFor(&app{}, request); err != nil {
		t.Fatalf("first completion request: %v", err)
	}
	if err := completeFlagsFor(&app{}, request); err == nil || !strings.HasPrefix(err.Error(), "epx: ") {
		t.Errorf("registering --format twice: error %v; want one naming epx", err)
	}
}

func flagsWhere(root *cobra.Command, keep func(*cobra.Command, *pflag.Flag) bool) []string {
	var flags []string
	walkCommands(root, func(cmd *cobra.Command) {
		cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			if keep(cmd, flag) {
				flags = append(flags, cmd.CommandPath()+" --"+flag.Name)
			}
		})
	})
	return flags
}

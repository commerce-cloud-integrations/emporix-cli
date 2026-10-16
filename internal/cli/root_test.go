package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runCLI runs the CLI with the given environment variables.
func runCLI(t *testing.T, env map[string]string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	return runIn(t, Env{Getenv: func(key string) string { return env[key] }}, args...)
}

// runIn runs the CLI in env, whose output it captures.
func runIn(t *testing.T, env Env, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	env.Out, env.Err = &out, &errOut
	if env.Profiles == nil {
		env.Profiles = noProfiles()
	}
	code = Run(context.Background(), args, env)
	return code, out.String(), errOut.String()
}

func run(out, errOut io.Writer, env map[string]string, args ...string) int {
	getenv := func(key string) string { return env[key] }
	return Run(context.Background(), args, Env{Out: out, Err: errOut, Getenv: getenv, Profiles: noProfiles()})
}

func TestHelpPathsSucceed(t *testing.T) {
	for _, args := range [][]string{
		nil, {"--help"}, {"auth"}, {"api"}, {"help", "api", "get"}, {"api", "get", "-h"}, {"sites"}, {"sites", "list", "-h"},
		{"prices"}, {"availability"}, {"availability", "get", "-h"}, {"shipping"}, {"shipping", "zones"},
		{"payment-modes", "public"},
	} {
		code, stdout, stderr := runCLI(t, nil, args...)
		if code != ExitOK || !strings.Contains(stdout, "Usage:") || stderr != "" {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
		}
	}
}

func TestGetHelpNamesTheIdentifier(t *testing.T) {
	for _, tt := range []struct {
		args  []string
		usage string
	}{
		{args: []string{"sites", "get", "-h"}, usage: "epx sites get CODE"},
		{args: []string{"units", "get", "-h"}, usage: "epx units get CODE"},
		{args: []string{"configurations", "get", "-h"}, usage: "epx configurations get KEY"},
		{args: []string{"indexing", "providers", "get", "-h"}, usage: "epx indexing providers get PROVIDER"},
		{args: []string{"products", "get", "-h"}, usage: "epx products get ID"},
		{args: []string{"taxes", "get", "-h"}, usage: "epx taxes get LOCATION"},
		{args: []string{"exchange-rates", "get", "-h"}, usage: "epx exchange-rates get CODE"},
		{args: []string{"payment-modes", "public", "get", "-h"}, usage: "epx payment-modes public get ID"},
	} {
		if code, stdout, _ := runCLI(t, nil, tt.args...); code != ExitOK || !strings.Contains(stdout, tt.usage+" [flags]") {
			t.Errorf("%v: exit %d, usage %q not in %q", tt.args, code, tt.usage, stdout)
		}
	}
}

// TestHelpCitesCommandsThatExist checks that every command line a help cites
// (`epx …`), such as the read that shows the IDs an argument takes, names
// commands that exist: a group is cited alone, or with one of its commands.
func TestHelpCitesCommandsThatExist(t *testing.T) {
	cited := regexp.MustCompile("`epx ([^`]+)`")
	root := newTestRoot()
	walkCommands(root, func(cmd *cobra.Command) {
		for _, line := range cited.FindAllStringSubmatch(cmd.Long, -1) {
			found, rest, err := root.Find(strings.Fields(line[1]))
			if err != nil || (found.HasSubCommands() && len(rest) > 0) {
				t.Errorf("%s: its help cites `epx %s`, which names no command", cmd.CommandPath(), line[1])
			}
		}
	})
}

func TestRunReportsUsageErrors(t *testing.T) {
	tests := []struct {
		args    []string
		message string
	}{
		{args: []string{"nope"}, message: `unknown command "nope" for "epx"`},
		{args: []string{"--nope"}, message: "unknown flag: --nope"},
		{args: []string{"version", "extra"}, message: `unknown command "extra" for "epx version"`},
		{args: []string{"auth", "nope"}, message: `unknown command "nope" for "epx auth"`},
		{args: []string{"completion", "nope"}, message: `unknown command "nope" for "epx completion"`},
		{args: []string{"help", "nope"}, message: `unknown help topic "nope"`},
		{args: []string{"api", "get"}, message: "accepts 1 arg(s), received 0"},
		{args: []string{"auth", "status", "-o", "xml"}, message: `invalid output format "xml": use table, json, yaml or csv`},
		{args: []string{"--timeout", "soon", "version"}, message: `invalid argument "soon"`},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			code, stdout, stderr := runCLI(t, nil, tt.args...)
			if code != ExitUsage || stdout != "" {
				t.Fatalf("exit = %d, stdout = %q; want %d and no output", code, stdout, ExitUsage)
			}
			if !strings.Contains(stderr, tt.message) || !strings.Contains(stderr, "epx --help") {
				t.Fatalf("stderr = %q, want %q and a help hint", stderr, tt.message)
			}
		})
	}
}

func TestOutputWriteFailuresAreErrors(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--help"}, {"auth"}, {"completion", "bash"}} {
		var errOut bytes.Buffer
		code := run(failingWriter{}, &errOut, nil, args...)
		if code != ExitError || !strings.Contains(errOut.String(), "disk full") {
			t.Errorf("%v: exit %d, stderr %q; want %d and the write error", args, code, errOut.String(), ExitError)
		}
	}
}

func TestVersionPrintsModuleVersion(t *testing.T) {
	code, stdout, _ := runCLI(t, nil, "version")
	if code != ExitOK || !strings.HasPrefix(stdout, "epx ") || !strings.HasSuffix(stdout, ")\n") {
		t.Fatalf("exit %d, stdout %q", code, stdout)
	}
	if strings.Count(stdout, "dirty") > 1 {
		t.Fatalf("dirty marker repeated: %q", stdout)
	}
}

func TestRunHookErrorsAreRuntimeErrors(t *testing.T) {
	failing := func(*cobra.Command, []string) error { return errors.New("hook failed") }
	for name, configure := range map[string]func(*cobra.Command){
		"persistent pre-run": func(c *cobra.Command) { c.PersistentPreRunE = failing },
		"pre-run":            func(c *cobra.Command) { c.PreRunE = failing },
		"post-run":           func(c *cobra.Command) { c.PostRunE = failing },
	} {
		leaf := &cobra.Command{Use: "leaf", RunE: func(*cobra.Command, []string) error { return nil }}
		configure(leaf)
		root := &cobra.Command{Use: "root", SilenceErrors: true, SilenceUsage: true}
		root.AddCommand(leaf)
		applyExitPolicy(root)
		root.SetArgs([]string{"leaf"})
		if code := exitCode(root.Execute()); code != ExitError {
			t.Errorf("%s: exit %d, want %d", name, code, ExitError)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// newTestRoot returns the command tree without configuration, to inspect it.
func newTestRoot() *cobra.Command {
	return newRootCommand(&app{getenv: func(string) string { return "" }}, io.Discard, io.Discard)
}

// walkCommands calls visit for cmd and every command below it.
func walkCommands(cmd *cobra.Command, visit func(*cobra.Command)) {
	visit(cmd)
	for _, sub := range cmd.Commands() {
		walkCommands(sub, visit)
	}
}

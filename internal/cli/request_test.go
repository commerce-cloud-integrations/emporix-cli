package cli

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRequestFill(t *testing.T) {
	req := request{
		path:  "/x/{tenant}/" + placeholderSite + "/items/" + placeholderID,
		query: url.Values{"itemId": {placeholderID}, "site": {placeholderSite}, "fixed": {"a", "b"}},
	}
	// IDs are escaped in the path and kept in the query; a placeholder inside a value stays as given.
	path, query := req.fill([]string{"a/b {site}"}, "Acme Store")
	if path != "/x/{tenant}/Acme%20Store/items/a%2Fb%20%7Bsite%7D" ||
		query.Encode() != "fixed=a&fixed=b&itemId=a%2Fb+%7Bsite%7D&site=Acme+Store" {
		t.Fatalf("fill = %q, %q", path, query.Encode())
	}
	// A placeholder counts wherever fill fills it in, also inside a query value.
	embedded := request{path: "/x", query: url.Values{"q": {"siteCode:" + placeholderSite}}}
	if !req.names(placeholderSite) || !req.names(placeholderID) || !embedded.names(placeholderSite) ||
		embedded.names(placeholderID) || (request{path: "/x/" + placeholderID}).names(placeholderSite) {
		t.Fatal("names must report a placeholder in the path or anywhere in a query value only")
	}
	if _, query := embedded.fill(nil, "main"); query.Get("q") != "siteCode:main" {
		t.Fatalf("embedded site filled as %q", query.Get("q"))
	}
}

func TestRequestFillsParentThenID(t *testing.T) {
	// The arguments fill {parent}, then {id}, wherever each is named; one holding the other's
	// placeholder stays as given.
	req := request{
		path:  "/x/{tenant}/types/" + placeholderParent + "/items/" + placeholderID,
		query: url.Values{"q": {"type:" + placeholderParent}},
	}
	if got := req.arguments(); !slices.Equal(got, []string{placeholderParent, placeholderID}) {
		t.Fatalf("arguments %q", got)
	}
	path, query := req.fill([]string{"a {id}", "b/c"}, "")
	if path != "/x/{tenant}/types/a%20%7Bid%7D/items/b%2Fc" || query.Get("q") != "type:a {id}" {
		t.Fatalf("fill = %q, %q", path, query.Encode())
	}
	if got := (request{path: "/x/" + placeholderID}).arguments(); !slices.Equal(got, []string{placeholderID}) {
		t.Fatalf("arguments of a read of one item %q", got)
	}
	if got := (request{path: "/x"}).arguments(); len(got) != 0 {
		t.Fatalf("arguments of a collection read %q", got)
	}
}

func TestRequestFillsThreeArguments(t *testing.T) {
	// {parent}, {inner} and {id} take the arguments in that order, wherever the path names them.
	req := request{path: "/x/{tenant}/" + placeholderParent + "/items/" + placeholderInner + "/" + placeholderID}
	if got := req.arguments(); !slices.Equal(got, []string{placeholderParent, placeholderInner, placeholderID}) {
		t.Fatalf("arguments %q", got)
	}
	if path, _ := req.fill([]string{"s/1", "{id}", "p 1"}, ""); path != "/x/{tenant}/s%2F1/items/%7Bid%7D/p%201" {
		t.Fatalf("fill = %q", path)
	}
	inner := request{path: "/x/" + placeholderInner + "/" + placeholderID}
	if got := inner.arguments(); !slices.Equal(got, []string{placeholderInner, placeholderID}) {
		t.Fatalf("arguments without a parent %q", got)
	}
}

// TestHelpNamesExistingCommands checks that help text quotes every command
// it names (`epx ...`) and that each quoted command exists: its leading words
// name a command, and what follows (arguments, flags) is not a further
// command name.
func TestHelpNamesExistingCommands(t *testing.T) {
	root := newTestRoot()
	walkCommands(root, func(cmd *cobra.Command) {
		if strings.HasPrefix(cmd.CommandPath(), "epx completion") {
			return // Cobra writes that help, with commands in shell examples
		}
		for _, problem := range helpReferenceProblems(root, cmd.Long) {
			t.Errorf("%s: %s", cmd.CommandPath(), problem)
		}
	})
	// The check itself fails on an unquoted or unknown command.
	if problems := helpReferenceProblems(root, "see epx sites list or `epx nope list`"); len(problems) != 2 {
		t.Fatalf("problems %q; want an unquoted and an unknown command", problems)
	}
}

// helpReferenceProblems returns what is wrong with the commands that text names.
func helpReferenceProblems(root *cobra.Command, text string) []string {
	quoted := regexp.MustCompile("`epx\\s+([^`]*)`")
	unquoted := regexp.MustCompile("(?:^|[^`])epx\\s+([a-z][a-z0-9-]*)")
	commandWord := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	var problems []string
	for _, match := range unquoted.FindAllStringSubmatch(text, -1) {
		if found, _, err := root.Find(match[1:2]); err == nil && found != root {
			problems = append(problems, "names epx "+match[1]+" without quoting it")
		}
	}
	for _, match := range quoted.FindAllStringSubmatch(text, -1) {
		names := strings.Fields(match[1])
		if i := slices.IndexFunc(names, func(w string) bool { return !commandWord.MatchString(w) }); i >= 0 {
			names = names[:i]
		}
		if found, rest, err := root.Find(names); err != nil || len(rest) > 0 || found == root {
			problems = append(problems, "quotes `epx "+match[1]+"`, which is not a command")
		}
	}
	return problems
}

// TestArgumentsMatchUsage checks that a command names as many arguments in
// its usage as it takes: a read takes one where its request names {id} and
// another where it names {parent}, so the two cannot drift apart.
func TestArgumentsMatchUsage(t *testing.T) {
	walkCommands(newTestRoot(),
		func(cmd *cobra.Command) {
			if names, takes := len(strings.Fields(cmd.Use))-1, argumentsTaken(cmd); names != takes {
				t.Errorf("%s: usage %q names %d arguments, takes %d", cmd.CommandPath(), cmd.Use, names, takes)
			}
		})
}

// argumentsTaken returns how many arguments cmd takes: 0 without a validator,
// else the one count, up to three, that it accepts, or -1 when it accepts
// none or several counts, which no usage matches. A count whose arguments
// are refused as an invalid value (one that only some values may take, such
// as TYPE) is accepted.
func argumentsTaken(cmd *cobra.Command) int {
	if cmd.Args == nil {
		return 0
	}
	var accepted []int
	for n := range 4 {
		if err := cmd.Args(cmd, slices.Repeat([]string{"x"}, n)); err == nil || strings.HasPrefix(err.Error(), "invalid ") {
			accepted = append(accepted, n)
		}
	}
	if len(accepted) != 1 {
		return -1
	}
	return accepted[0]
}

func TestArgumentNames(t *testing.T) {
	for _, tt := range []struct {
		use  string
		n    int
		want []string
	}{
		{use: "get TYPE ID", n: 2, want: []string{"TYPE", "ID"}},
		{use: "get ID", n: 1, want: []string{"ID"}},
		{use: "get ID", n: 2, want: []string{"ID", "argument"}},
		{use: "", n: 1, want: []string{"argument"}},
	} {
		if got := argumentNames(tt.use, tt.n); !slices.Equal(got, tt.want) {
			t.Errorf("argumentNames(%q, %d) = %q, want %q", tt.use, tt.n, got, tt.want)
		}
	}
}

func TestArgumentsTaken(t *testing.T) {
	for _, tt := range []struct {
		args cobra.PositionalArgs
		want int
	}{
		{args: nil, want: 0},
		{args: cobra.NoArgs, want: 0},
		{args: requestArgs([]string{"ID"}, 1, nil), want: 1},
		// Arguments refused for their value only still count.
		{args: requestArgs([]string{"TYPE", "ID"}, 2, map[string][]string{"TYPE": {"PRODUCT"}}), want: 2},
		{args: requestArgs([]string{"KEY", "VALUE"}, 1, nil), want: 2},
		{args: cobra.ExactArgs(2), want: 2},
		{args: cobra.MinimumNArgs(1), want: -1},
		{args: cobra.RangeArgs(0, 1), want: -1},
		{args: cobra.ExactArgs(5), want: -1},
	} {
		if got := argumentsTaken(&cobra.Command{Use: "x A B C D E", Args: tt.args}); got != tt.want {
			t.Errorf("argumentsTaken = %d, want %d", got, tt.want)
		}
	}
}

// TestAlternatives checks how choices are named in messages and help.
func TestAlternatives(t *testing.T) {
	for want, values := range map[string][]string{
		"PRODUCT":                {"PRODUCT"},
		"PRODUCT or CATEGORY":    {"PRODUCT", "CATEGORY"},
		"QUARTER, MONTH or WEEK": {"QUARTER", "MONTH", "WEEK"},
	} {
		if got := alternatives(values); got != want {
			t.Errorf("alternatives(%q) = %q, want %q", values, got, want)
		}
	}
}

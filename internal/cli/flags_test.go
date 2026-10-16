package cli

import (
	"slices"
	"strings"
	"testing"
)

// TestReadFlagsFollowTheCases checks that --language, --site and
// --allow-write-scope appear exactly on the commands whose read cases pass
// them: --language on every read whose table shows translations (each case
// shows one), --site on every read whose request names the site,
// --allow-write-scope on every opt-in read.
func TestReadFlagsFollowTheCases(t *testing.T) {
	for _, flag := range []string{"language", "site", writeScopeFlag} {
		var want []string
		for path, cases := range casesOfCommands(t) {
			if slices.ContainsFunc(cases, func(c readCase) bool { return slices.Contains(c.args, "--"+flag) }) {
				want = append(want, path)
			}
		}
		slices.Sort(want)
		// A write whose request names the site takes --site too, which its write cases pass.
		reads := readCommands()
		got := slices.DeleteFunc(commandsWithFlag(flag), func(path string) bool { return !slices.Contains(reads, path) })
		if !slices.Equal(got, want) {
			t.Errorf("--%s on:\n%s\nwant:\n%s", flag, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}
}

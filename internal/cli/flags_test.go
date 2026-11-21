package cli

import (
	"slices"
	"strings"
	"testing"
)

func TestReadFlagsFollowTheCases(t *testing.T) {
	for _, flag := range []string{"language", "site", mixinSiteFlag, writeScopeFlag} {
		var want []string
		for path, cases := range casesOfCommands(t) {
			if slices.ContainsFunc(cases, func(c readCase) bool { return slices.Contains(c.args, "--"+flag) }) {
				want = append(want, path)
			}
		}
		slices.Sort(want)
		reads := readCommands()
		got := slices.DeleteFunc(commandsWithFlag(flag), func(path string) bool { return !slices.Contains(reads, path) })
		if !slices.Equal(got, want) {
			t.Errorf("--%s on:\n%s\nwant:\n%s", flag, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}
}

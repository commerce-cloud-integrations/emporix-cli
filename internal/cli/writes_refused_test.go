package cli

import (
	"slices"
	"strings"
	"testing"
)

type refusedWrite struct {
	args []string
	file string
	want string
}

func assertRefused(t *testing.T, cases []refusedWrite) {
	t.Helper()
	for _, tt := range cases {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			args := append(slices.Clone(tt.args), flag)
			if tt.file != "" {
				args = append(args, "--file", writeBody(t, tt.file))
			}
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

func TestPatchesOfOperationsRefuseAnythingButAnArray(t *testing.T) {
	var cases []refusedWrite
	for _, args := range [][]string{
		{"ai", "agents", "patch", "a1"}, {"ai", "mcp-servers", "patch", "m1"}, {"ai", "oauths", "patch", "o1"},
		{"ai", "tokens", "patch", "t1"}, {"ai", "tools", "patch", "t1"},
		{"custom-entities", "instances", "patch", "DOCUMENT", "123"},
		{"delivery-times", "patch", "d1"}, {"delivery-times", "slots", "patch", "d1", "s1"},
		{"media", "patch", "m1"}, {"quotes", "patch", "q1"}, {"returns", "patch", "r1"}, {"segments", "patch", "s1"},
		{"webhooks", "configurations", "patch", "http"},
	} {
		for _, file := range []string{`{"op":"replace","path":"/name","value":"A"}`, `"replace"`, "null", "[]"} {
			cases = append(cases, refusedWrite{args: args, file: file, want: "not a JSON array of one or more items"})
		}
	}
	assertRefused(t, cases)
}

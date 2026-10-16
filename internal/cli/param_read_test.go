package cli

import (
	"slices"
	"strings"
	"testing"
)

// TestParamReadFailuresSendNothing checks the query parameters that flags set:
// on `carts find`, exactly one of --customer and --session is required and
// none may be empty, also when the read is opted in to; on
// `ai analytics executions`, --granularity takes only the granularities that
// Emporix names, in capitals. Any other use is a usage error found while
// flags are parsed, before a token is requested or anything is read.
func TestParamReadFailuresSendNothing(t *testing.T) {
	find := []string{"carts", "find", "--site", "AcmeStore_NL", "--allow-write-scope"}
	executions := []string{"ai", "analytics", "executions", "helper"}
	const choices = `flag: use QUARTER, MONTH or WEEK`
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: find, want: "at least one of the flags in the group [customer session] is required"},
		{args: slices.Concat(find, []string{"--legal-entity", "le1", "--type", "shopping"}),
			want: "at least one of the flags in the group [customer session] is required"},
		{args: slices.Concat(find, []string{"--customer", "C0123", "--session", "s-1"}),
			want: "[customer session] were all set"},
		{args: slices.Concat(find, []string{"--customer", ""}), want: `invalid argument "" for "--customer" flag: empty`},
		{args: slices.Concat(find, []string{"--session", "s-1", "--type="}),
			want: `invalid argument "" for "--type" flag: empty`},
		{args: slices.Concat(executions, []string{"--granularity", "DAY"}),
			want: `invalid argument "DAY" for "--granularity" ` + choices},
		{args: slices.Concat(executions, []string{"--granularity", "week"}),
			want: `invalid argument "week" for "--granularity" ` + choices},
		{args: slices.Concat(executions, []string{"--granularity", "Month"}),
			want: `invalid argument "Month" for "--granularity" ` + choices},
		{args: slices.Concat(executions, []string{"--granularity="}),
			want: `invalid argument "" for "--granularity" flag: empty`},
	} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, code, stdout, stderr, tt.want)
		}
		if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
			t.Errorf("%q: tokens %q, requests %q; want none", tt.args, fake.tokenScopes, fake.gets)
		}
	}
}

// TestParamFlagsNameTheirChoices checks the help of the flags that set query
// parameters: a flag with choices names them after its own help, from the
// same list that it accepts and completes; any other keeps its help as
// declared.
func TestParamFlagsNameTheirChoices(t *testing.T) {
	root := newTestRoot()
	for _, tt := range []struct {
		args       []string
		flag, want string
	}{
		{args: []string{"ai", "analytics", "executions"}, flag: granularityParam,
			want: "the calendar `period` of each count: QUARTER, MONTH or WEEK"},
		{args: []string{"carts", "find"}, flag: cartCustomerFlag, want: "the customer `ID` whose cart to read"},
	} {
		cmd, _, err := root.Find(tt.args)
		if err != nil {
			t.Fatalf("%q: %v", tt.args, err)
		}
		if flag := cmd.Flags().Lookup(tt.flag); flag == nil || flag.Usage != tt.want {
			t.Errorf("%q --%s: %+v; want usage %q", tt.args, tt.flag, flag, tt.want)
		}
	}
}

// TestParamValueWithoutChoicesTakesAnyValue checks that a flag whose
// choices are empty, not only nil, takes any value but an empty one.
func TestParamValueWithoutChoicesTakesAnyValue(t *testing.T) {
	for _, choices := range [][]string{nil, {}} {
		v := paramValue{paramFlag: paramFlag{choices: choices}}
		if err := v.Set("any"); err != nil || v.String() != "any" {
			t.Errorf("choices %q: Set(any) = %v, value %q; want any taken", choices, err, v.String())
		}
		if err := v.Set(""); err == nil || v.String() != "any" {
			t.Errorf("choices %q: Set(\"\") = %v, value %q; want an error and any kept", choices, err, v.String())
		}
	}
}

package cli

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestParamScopesFollowTheQuery(t *testing.T) {
	flag := paramFlag{param: "check", flag: "check", choices: []string{queryTrue, queryFalse}, usage: "whether to check"}
	for _, tt := range []struct {
		byDefault bool
		query     url.Values
		want      bool
	}{
		{byDefault: true, query: url.Values{}, want: true},
		{byDefault: true, query: url.Values{"check": {queryTrue}}, want: true},
		{byDefault: true, query: url.Values{"check": {queryFalse}}, want: false},
		{byDefault: false, query: url.Values{}, want: false},
		{byDefault: false, query: url.Values{"check": {queryTrue}}, want: true},
		{byDefault: false, query: url.Values{"check": {queryFalse}}, want: false},
	} {
		p := paramScopes{flag: flag, neededValue: queryTrue, byDefault: tt.byDefault, scopes: []string{"a.a_read"}}
		if got := p.needed(tt.query); got != tt.want {
			t.Errorf("byDefault %v, query %v: needed %v; want %v", tt.byDefault, tt.query, got, tt.want)
		}
	}
}

func TestParamScopesNameTheirFlag(t *testing.T) {
	check := paramFlag{param: "check", flag: "check", choices: []string{queryTrue, queryFalse}, usage: "whether to check"}
	cmd := (&write{
		use: "create", method: http.MethodPost, scopes: []string{"a.a_manage"}, req: request{path: "/a"}, body: noBody{},
		paramScopes: paramScopes{
			flag: check, neededValue: queryTrue, byDefault: true, scopes: []string{"a.a_read"},
		},
	}).command(&app{})
	flag := cmd.Flags().Lookup("check")
	const usage = "whether to check, requesting a.a_read where it is true or not given: true or false"
	const help = "\n\nWhere --check is true or not given, epx also requests the a.a_read scope.\n\n" +
		"Requires the a.a_manage scope."
	if got := flag.Annotations[writeScopesAnnotation]; !slices.Equal(got, []string{"a.a_read"}) || flag.Usage != usage ||
		!strings.HasSuffix(cmd.Long, help) {
		t.Errorf("--check annotated %q, usage %q, help %q; want a.a_read, %q and the help to end with %q",
			got, flag.Usage, cmd.Long, usage, help)
	}
}

package cli

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

type paramScopes struct {
	flag        paramFlag
	neededValue string
	byDefault   bool
	scopes      []string
}

func (p *paramScopes) bind(cmd *cobra.Command) func(query url.Values) []string {
	if len(p.scopes) == 0 {
		return func(url.Values) []string { return nil }
	}
	flag := p.flag
	flag.usage += ", requesting " + strings.Join(p.scopes, ", ") + " where " + p.when("it is")
	set := queryFlags{optional: []paramFlag{flag}}.bind(cmd)
	annotate(cmd, flag.flag, writeScopesAnnotation, p.scopes)
	return func(query url.Values) []string {
		set(query)
		if p.needed(query) {
			return p.scopes
		}
		return nil
	}
}

func (p *paramScopes) needed(query url.Values) bool {
	if !query.Has(p.flag.param) {
		return p.byDefault
	}
	return query.Get(p.flag.param) == p.neededValue
}

func (p *paramScopes) when(is string) string {
	if p.byDefault {
		return is + " " + p.neededValue + " or not given"
	}
	return is + " " + p.neededValue
}

func (p *paramScopes) note() string {
	if len(p.scopes) == 0 {
		return ""
	}
	return "Where --" + p.when(p.flag.flag+" is") + ", epx also requests " + theScopes(p.scopes) + "."
}

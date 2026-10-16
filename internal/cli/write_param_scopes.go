package cli

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// paramScopes is a query parameter of a write, which its flag sets, with
// scopes that the write requests too only where the query that it sends
// needs them: where it sends the parameter as neededValue or, byDefault,
// does not send it, as Emporix then acts as with that value. E.g. the read
// scope of a custom entity type, with which Emporix checks the references of
// its instances where validateReferences is true, or not sent for the writes
// whose references it checks by default. Unlike those of a choice, its scopes
// may name a custom entity type, which typeScopes fills in. Its flag is
// annotated with them (writeScopesAnnotation), as the flags of a choice are,
// and its usage and the help say when the write requests them (note).
type paramScopes struct {
	flag        paramFlag // the flag that sets the parameter, besides the write's params
	neededValue string    // the value of the parameter with which the write needs the scopes, e.g. true
	byDefault   bool      // whether it needs them where it does not send the parameter
	scopes      []string  // none: the write has no such parameter
}

// bind defines p's flag on cmd, annotated with p's scopes, its usage ending
// with when the write requests them, and returns how a run sets the parameter
// that the flag gives in query and returns the scopes that query then needs:
// p's or none. Without scopes, it defines nothing and returns none.
func (p *paramScopes) bind(cmd *cobra.Command) func(query url.Values) []string {
	if len(p.scopes) == 0 {
		return func(url.Values) []string { return nil }
	}
	flag := p.flag
	flag.usage += ", requesting " + strings.Join(p.scopes, ", ") + " where " + p.when("it is")
	set := queryFlags{optional: []paramFlag{flag}}.bind(cmd)
	// SetAnnotation fails only for an undefined flag, and this one was just defined.
	_ = cmd.Flags().SetAnnotation(flag.flag, writeScopesAnnotation, p.scopes)
	return func(query url.Values) []string {
		set(query)
		if p.needed(query) {
			return p.scopes
		}
		return nil
	}
}

// needed reports whether a write that sends query needs p's scopes.
func (p *paramScopes) needed(query url.Values) bool {
	if !query.Has(p.flag.param) {
		return p.byDefault
	}
	return query.Get(p.flag.param) == p.neededValue
}

// when says when a write needs p's scopes, of its flag after is, e.g. "it is
// true or not given".
func (p *paramScopes) when(is string) string {
	if p.byDefault {
		return is + " " + p.neededValue + " or not given"
	}
	return is + " " + p.neededValue
}

// note returns the paragraph that precedes the scope sentence of a write with
// p, as that of its choice does (queryChoice.scopeNote), which says when it
// requests p's scopes too; none without any.
func (p *paramScopes) note() string {
	if len(p.scopes) == 0 {
		return ""
	}
	return "\n\nWhere --" + p.when(p.flag.flag+" is") + ", epx also requests\n" + theScopes(p.scopes) + "."
}

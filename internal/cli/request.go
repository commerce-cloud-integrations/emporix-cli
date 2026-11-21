package cli

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	placeholderParent = "{parent}"
	placeholderInner  = "{inner}"
	placeholderID     = "{id}"
	placeholderSite   = "{site}"
)

const (
	queryTrue  = "true"
	queryFalse = "false"
)

const partialParam = "partial"

const forceDeleteParam = "forceDelete"

const typeParam = "type"

const expandParam = "expand"

type request struct {
	path  string
	query url.Values
}

func withQueries(path string, queries ...url.Values) request {
	query := url.Values{}
	for _, members := range queries {
		maps.Copy(query, members)
	}
	return request{path: path, query: query}
}

func (req request) arguments() []string {
	all := []string{placeholderParent, placeholderInner, placeholderID}
	return slices.DeleteFunc(all, func(placeholder string) bool {
		return !req.names(placeholder)
	})
}

func (req request) bind(cmd *cobra.Command, choices map[string][]string) (*siteCode, cobra.PositionalArgs) {
	n := len(req.arguments())
	var accept cobra.PositionalArgs
	if names := argumentNames(cmd.Use, max(n, argumentCount(cmd.Use))); len(names) > 0 {
		accept = requestArgs(names, n, choices)
	}
	site := new(siteCode)
	if req.names(placeholderSite) {
		addSiteFlag(cmd, site)
	}
	return site, accept
}

type argumentBinding struct {
	req      request
	choices  map[string][]string
	builtIn  builtInTypes
	ids      argumentIDs
	reserved map[string]string
}

func (b *argumentBinding) bind(a *app, cmd *cobra.Command, scopes []string) (*siteCode, typeRule) {
	site, accept := b.req.bind(cmd, b.choices)
	if len(b.reserved) > 0 {
		accept = refuseReserved(accept, b.reserved)
	}
	rule := typeScopes(cmd.Use, b.req, scopes, b.builtIn)
	cmd.Args = rule.args(accept)
	completeArguments(a, cmd, site, b.ids, b.choices, rule)
	return site, rule
}

func (req request) fill(args []string, site siteCode) (string, url.Values) {
	inPath := []string{placeholderSite, url.PathEscape(string(site))}
	inQuery := []string{placeholderSite, string(site)}
	for i, placeholder := range req.arguments() {
		var arg string
		if i < len(args) {
			arg = args[i]
		}
		inPath = append(inPath, placeholder, url.PathEscape(arg))
		inQuery = append(inQuery, placeholder, arg)
	}
	path, values := strings.NewReplacer(inPath...), strings.NewReplacer(inQuery...)
	query := url.Values{}
	for name, given := range req.query {
		for _, value := range given {
			query.Add(name, values.Replace(value))
		}
	}
	return path.Replace(req.path), query
}

func (req request) names(placeholder string) bool {
	if strings.Contains(req.path, placeholder) {
		return true
	}
	for _, values := range req.query {
		for _, value := range values {
			if strings.Contains(value, placeholder) {
				return true
			}
		}
	}
	return false
}

func requestArgs(names []string, n int, choices map[string][]string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(len(names))(cmd, args); err != nil {
			return err
		}
		for i, arg := range args[:n] {
			if !isPathSegment(arg) {
				return fmt.Errorf("invalid %s %q", names[i], arg)
			}
			if values, closed := choices[names[i]]; closed && !slices.Contains(values, arg) {
				return fmt.Errorf("invalid %s %q: use %s", names[i], arg, alternatives(values))
			}
		}
		return nil
	}
}

func alternatives(values []string) string {
	last := len(values) - 1
	if last < 1 {
		return strings.Join(values, "")
	}
	return strings.Join(values[:last], ", ") + " or " + values[last]
}

func argumentNames(use string, n int) []string {
	names := slices.Repeat([]string{"argument"}, n)
	words := strings.Fields(use)
	copy(names, words[min(1, len(words)):])
	return names
}

func declaredArguments(use string) []string {
	return argumentNames(use, argumentCount(use))
}

func argumentCount(use string) int {
	return max(0, len(strings.Fields(use))-1)
}

func isPathSegment(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".."
}

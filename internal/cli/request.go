package cli

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// Placeholders of a request, filled in when the command runs.
const (
	placeholderParent = "{parent}" // the command's first argument, before ID: the item that ID belongs to
	placeholderInner  = "{inner}"  // the argument between {parent} and ID, such as the type of the item ID
	placeholderID     = "{id}"     // the command's ID argument, its last
	placeholderSite   = "{site}"   // the command's --site, which it then requires
)

// queryTrue and queryFalse are the values of a query parameter that switches
// something on or off, e.g. showUnpublished=true or partial=false.
const (
	queryTrue  = "true"
	queryFalse = "false"
)

// partialParam is the query parameter of a PUT that says whether it changes
// only the fields that the body gives (true) or replaces the whole item, as
// the Price, Coupon, Fee and Cart Services take it, and the flag that sets
// it (partialSwitch).
const partialParam = "partial"

// forceDeleteParam is the query parameter of a delete that, true, also deletes
// what refers to the item, where Emporix would otherwise refuse to delete it.
const forceDeleteParam = "forceDelete"

// typeParam is the name of the query parameter, and of the flag, that selects
// a type, e.g. of a cart, of imported records or of exported custom entities.
const typeParam = "type"

// expandParam is the name of the query parameter that names what Emporix
// expands in a response or, in a site's update, the site settings that it
// changes, such as payment:all.
const expandParam = "expand"

// request is what a read or a write sends: a path and query that may name
// {parent}, {inner}, {id} and {site}. {tenant} is left to the Emporix client.
type request struct {
	path  string
	query url.Values
}

// withQueries returns the request of path with the members of queries; one of
// a later query replaces the member of the same name of an earlier one.
func withQueries(path string, queries ...url.Values) request {
	query := url.Values{}
	for _, members := range queries {
		maps.Copy(query, members)
	}
	return request{path: path, query: query}
}

// arguments returns the placeholders of req that a command's arguments fill
// in, in their order: {parent}, {inner}, then {id}.
func (req request) arguments() []string {
	all := []string{placeholderParent, placeholderInner, placeholderID}
	return slices.DeleteFunc(all, func(placeholder string) bool {
		return !req.names(placeholder)
	})
}

// bind makes cmd, a read or a write that sends req, take the arguments that
// its usage names: those first fill in the placeholders that req names
// (arguments), each addressing one item (one of its choices if it has any),
// and the others are the command's own to check; and require --site where req names {site}. It
// returns the site that --site gives once flags are parsed.
func (req request) bind(cmd *cobra.Command, choices map[string][]string) *siteCode {
	n := len(req.arguments())
	if names := argumentNames(cmd.Use, max(n, argumentCount(cmd.Use))); len(names) > 0 {
		cmd.Args = requestArgs(names, n, choices)
	}
	site := new(siteCode)
	if req.names(placeholderSite) {
		addSiteFlag(cmd, site)
	}
	return site
}

// fill returns the path and query of req for a command's arguments and site:
// escaped as path segments in the path, as given in query values.
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
	// A replacer works in one pass, so an argument holding "{site}" or "{id}" stays as given.
	path, values := strings.NewReplacer(inPath...), strings.NewReplacer(inQuery...)
	query := url.Values{}
	for name, given := range req.query {
		for _, value := range given {
			query.Add(name, values.Replace(value))
		}
	}
	return path.Replace(req.path), query
}

// names reports whether req names placeholder, in its path or in a query
// value, as fill would fill it in.
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

// requestArgs accepts exactly one argument per name, of which the first n
// each address a single item and, where choices names one, are one of its
// values; an invalid one is named, e.g. as ID or TYPE.
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

// alternatives names values as choices, e.g. "PRODUCT or CATEGORY" or
// "QUARTER, MONTH or WEEK".
func alternatives(values []string) string {
	last := len(values) - 1
	if last < 1 {
		return strings.Join(values, "")
	}
	return strings.Join(values[:last], ", ") + " or " + values[last]
}

// argumentNames returns the names of a command's n arguments as its usage
// names them after the command, e.g. TYPE and ID in "get TYPE ID"; one the
// usage leaves out is "argument" (TestArgumentsMatchUsage rules that out).
func argumentNames(use string, n int) []string {
	names := slices.Repeat([]string{"argument"}, n)
	words := strings.Fields(use)
	copy(names, words[min(1, len(words)):])
	return names
}

// declaredArguments returns the names of the arguments that a usage declares,
// one per word after the command name (argumentNames).
func declaredArguments(use string) []string {
	return argumentNames(use, argumentCount(use))
}

// argumentCount returns how many arguments a usage declares: the words after
// the command name, none for an empty usage.
func argumentCount(use string) int {
	return max(0, len(strings.Fields(use))-1)
}

// isPathSegment reports whether value addresses one item once escaped as a
// path segment: empty, "." and ".." would address the collection or its
// parent, and nothing but spaces no item (nor could a delete of it be
// confirmed by typing it).
func isPathSegment(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".."
}

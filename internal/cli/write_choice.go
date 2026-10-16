package cli

import (
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// queryChoice is a query parameter of a write that one of its flags sets, to
// that flag's value, with the scopes that Emporix requires for it, such as
// publish, which --publish sets to true with the category.category_publish
// scope; without one of its flags, the parameter is not sent and none of
// those scopes is requested. Its flags exclude each other, as the parameter
// takes one value. A choice without a parameter only adds scopes, those that
// Emporix requires of some items or bodies, which epx never requests
// implicitly: e.g. --completed adds order.order_update_completed, which
// Emporix requires to change an order that is COMPLETED, and
// --allow-external-prices adds cart.cart_manage_external_prices, which it
// requires of a cart body with an external price. Its flags may be given
// together, each adding its own, as one write may need several: e.g.
// --allow-publish and --allow-unpublish for a bulk write of products.
type queryChoice struct {
	param   string         // e.g. "publish"; none for a choice that only adds scopes
	options []choiceOption // a flag per value; no choice without any
}

// choiceOption is a flag of a queryChoice and the value that it sets.
type choiceOption struct {
	flag   string   // its name, e.g. "publish"
	usage  string   // what it does; the scopes it adds follow
	value  string   // the parameter's value, e.g. "true"; none without a parameter
	scopes []string // the scopes it adds
	// refused is what Emporix refuses with 403 Forbidden without the flag,
	// e.g. "a product that the write publishes", which the help says; nothing
	// when empty.
	refused string
}

// bind adds the flags of c's options to cmd, which exclude each other where c
// has a parameter, each annotated with the scopes it adds
// (writeScopesAnnotation), and returns how a run sets the parameter of the
// flag given, if any, in its query (where c has one) and adds the scopes of
// each flag given to scopes, which typeScopes does not fill in, so none may
// name a custom entity type (refuseTypeScopes).
func (c queryChoice) bind(cmd *cobra.Command) func(query url.Values, scopes []string) []string {
	given := make([]*bool, len(c.options))
	flags := make([]string, len(c.options))
	for i, option := range c.options {
		refuseTypeScopes(cmd.Use, option.scopes)
		given[i] = cmd.Flags().Bool(option.flag, false, option.usage+" (requests "+
			strings.Join(option.scopes, ", ")+" too)")
		// SetAnnotation fails only for an undefined flag, and this one was just defined.
		_ = cmd.Flags().SetAnnotation(option.flag, writeScopesAnnotation, option.scopes)
		flags[i] = option.flag
	}
	if c.param != "" {
		cmd.MarkFlagsMutuallyExclusive(flags...)
	}
	return func(query url.Values, scopes []string) []string {
		for i, option := range c.options {
			if !*given[i] {
				continue
			}
			if c.param != "" {
				query.Set(c.param, option.value)
			}
			scopes = slices.Concat(scopes, option.scopes)
		}
		return scopes
	}
}

// scopeNote returns the paragraph that precedes the scope sentence of a write
// with c: a line per flag, naming the scopes that it adds and what Emporix
// refuses without it, if it says; none without options.
func (c queryChoice) scopeNote() string {
	note := ""
	for _, option := range c.options {
		note += "\n--" + option.flag + " also requests " + theScopes(option.scopes) + "."
		if option.refused != "" {
			note += " Without it, Emporix\nrefuses " + option.refused + " (403 Forbidden, exit 3)."
		}
	}
	if note == "" {
		return ""
	}
	return "\n" + note
}

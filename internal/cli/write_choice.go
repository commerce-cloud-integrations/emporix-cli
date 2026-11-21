package cli

import (
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type queryChoice struct {
	param   string
	options []choiceOption
}

type choiceOption struct {
	flag    string
	usage   string
	value   string
	scopes  []string
	refused string
}

func (c queryChoice) bind(cmd *cobra.Command) func(query url.Values, scopes []string) []string {
	given := make([]*bool, len(c.options))
	flags := make([]string, len(c.options))
	for i, option := range c.options {
		refuseTypeScopes(cmd.Use, option.scopes)
		given[i] = cmd.Flags().Bool(option.flag, false, option.usage+" (requests "+
			strings.Join(option.scopes, ", ")+" too)")
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

func (c queryChoice) scopeNote() string {
	notes := make([]string, 0, len(c.options))
	for _, option := range c.options {
		note := "--" + option.flag + " also requests " + theScopes(option.scopes) + "."
		if option.refused != "" {
			note += " Without it, Emporix refuses " + option.refused + " (403 Forbidden, exit 3)."
		}
		notes = append(notes, note)
	}
	return joinSentences(notes...)
}

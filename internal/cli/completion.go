package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func completeCommandNames(c *cobra.Command, args []string, toComplete string) (
	[]cobra.Completion, cobra.ShellCompDirective) {
	target, rest, err := c.Root().Find(args)
	if err != nil || len(rest) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []cobra.Completion
	for _, sub := range target.Commands() {
		if sub.IsAvailableCommand() && strings.HasPrefix(sub.Name(), toComplete) {
			names = append(names, cobra.CompletionWithDesc(sub.Name(), sub.Short))
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

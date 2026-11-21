package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// How shells complete the values of flags. Cobra keeps the completion
// function of every flag in a map of its package that is never cleared, which
// keeps the flag's whole command tree in memory for the life of the process.
// So the tree records what each flag completes with as annotations of the
// flag, and the functions are registered only for a completion request
// (completeFlagsFor), never for a run.

const choicesAnnotation = "epx_completes_choices"

const fileNamesAnnotation = "epx_completes_files"

func completeChoices(cmd *cobra.Command, flag string, values []string) {
	annotate(cmd.Flag(flag), choicesAnnotation, values)
}

func completeFileNames(cmd *cobra.Command, flag string) {
	annotate(cmd.Flag(flag), fileNamesAnnotation, nil)
}

func completeFlagsFor(a *app, cmd *cobra.Command) {
	if cmd.Name() == cobra.ShellCompRequestCmd {
		completeFlags(a, cmd.Root(), sites().ids())
	}
}

func completeFlags(a *app, cmd *cobra.Command, codes *idSource) {
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if complete := flagCompletion(a, flag, codes); complete != nil {
			_ = cmd.RegisterFlagCompletionFunc(flag.Name, complete)
		}
	})
	for _, sub := range cmd.Commands() {
		completeFlags(a, sub, codes)
	}
}

func flagCompletion(a *app, flag *pflag.Flag, codes *idSource) cobra.CompletionFunc {
	if values, ok := flag.Annotations[choicesAnnotation]; ok {
		return cobra.FixedCompletions(values, cobra.ShellCompDirectiveNoFileComp)
	}
	if _, ok := flag.Annotations[fileNamesAnnotation]; ok {
		return cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault)
	}
	switch flag.Name {
	case profileFlag:
		return a.completeProfiles
	case siteFlag, mixinSiteFlag:
		return func(c *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return offerArgument(nil, codes, a.listedItems(c.Context(), codes, nil, ""), toComplete)
		}
	default:
		return nil
	}
}

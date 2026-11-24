package cli

import (
	"errors"
	"fmt"

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
	annotate(cmd, flag, choicesAnnotation, values)
}

func completeFileNames(cmd *cobra.Command, flag string) {
	annotate(cmd, flag, fileNamesAnnotation, nil)
}

func completeFlagsFor(a *app, cmd *cobra.Command) error {
	if cmd.Name() != cobra.ShellCompRequestCmd {
		return nil
	}
	return completeFlags(a, cmd.Root(), sites().ids())
}

func completeFlags(a *app, cmd *cobra.Command, codes *idSource) error {
	var failures []error
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if complete := flagCompletion(a, flag, codes); complete != nil {
			if err := cmd.RegisterFlagCompletionFunc(flag.Name, complete); err != nil {
				failures = append(failures, fmt.Errorf("%s: %w", cmd.Use, err))
			}
		}
	})
	for _, sub := range cmd.Commands() {
		failures = append(failures, completeFlags(a, sub, codes))
	}
	return errors.Join(failures...)
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
	case savedSearchFlag:
		src := savedSearchIDs()
		return func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return offerArgument(nil, src, a.listedItems(c.Context(), src, args, ""), toComplete)
		}
	default:
		return nil
	}
}

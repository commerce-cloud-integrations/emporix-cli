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

// choicesAnnotation is the annotation of a flag that completes with fixed
// values: the values (completeChoices).
const choicesAnnotation = "epx_completes_choices"

// fileNamesAnnotation is the annotation of a flag that completes with file
// names (completeFileNames).
const fileNamesAnnotation = "epx_completes_files"

// completeChoices makes shells complete --flag, which cmd has just added, with
// values only, never with file names.
func completeChoices(cmd *cobra.Command, flag string, values []string) {
	annotate(cmd.Flag(flag), choicesAnnotation, values)
}

// completeFileNames makes shells complete --flag, which cmd has just added,
// with file names: completion suggests none elsewhere (newRootCommand).
func completeFileNames(cmd *cobra.Command, flag string) {
	annotate(cmd.Flag(flag), fileNamesAnnotation, nil)
}

// completeFlagsFor registers how shells complete the flags of the tree of
// cmd, about to run, when cmd is Cobra's completion request, the hidden
// command that the completion scripts run (`epx __complete ARGS`), and never
// otherwise.
func completeFlagsFor(a *app, cmd *cobra.Command) {
	if cmd.Name() == cobra.ShellCompRequestCmd {
		completeFlags(a, cmd.Root(), sites().ids())
	}
}

// completeFlags registers, for a completion request, how shells complete the
// flags, local or persistent, that cmd, a command of a, and every command
// below it define (flagCompletion).
func completeFlags(a *app, cmd *cobra.Command, codes *idSource) {
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if complete := flagCompletion(a, flag, codes); complete != nil {
			// Registration fails only for an unknown or already registered flag; each flag
			// is defined, and registered once, by the command that defines it.
			_ = cmd.RegisterFlagCompletionFunc(flag.Name, complete)
		}
	})
	for _, sub := range cmd.Commands() {
		completeFlags(a, sub, codes)
	}
}

// flagCompletion returns how shells complete flag, or nil where the default
// applies (no file names): with the values or file names that it records,
// --profile with the names of the profiles, and --site, wherever a command
// has one, with the site codes that codes lists, as listedItems reads them.
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
	case siteFlag:
		return func(c *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return codes.offer(a.listedItems(c.Context(), codes, nil, ""), toComplete), cobra.ShellCompDirectiveNoFileComp
		}
	default:
		return nil
	}
}

package cli

import "github.com/spf13/cobra"

type answerer interface {
	bind(a *app, cmd *cobra.Command) endingPlanner
}

type endingPlanner interface {
	plan(cmd *cobra.Command, write []string) ([]string, answered)
}

type printing struct{}

func (printing) plan(_ *cobra.Command, write []string) ([]string, answered) {
	return write, asAnswered
}

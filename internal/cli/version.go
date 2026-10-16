package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the epx version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "epx %s (%s)\n", a.version, a.goVersion)
			return err
		},
	}
}

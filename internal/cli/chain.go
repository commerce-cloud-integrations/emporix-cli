package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type chain struct {
	payload
	most int
}

func (c *chain) bind(cmd *cobra.Command) string {
	stdin := c.payload.bind(cmd)
	cmd.Long += "\n\nEmporix takes up to " + strconv.Itoa(c.most) + " commands in one request (epx refuses more), " +
		"runs them in order and answers with the result of each that ran: when one failed, epx prints the " +
		"response, lists the commands that failed on standard error (of a command that adds or updates items " +
		"in a batch, the items that failed; Emporix carried out the others) and " + bulkExitSentence
	return stdin
}

func (c *chain) fill(ctx context.Context, w *emporix.Write, values []string) error {
	err := c.payload.fill(ctx, w, values)
	w.Parts = emporix.Commands(c.most)
	return err
}

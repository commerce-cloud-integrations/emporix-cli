package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// chain is the body of a write that runs a chain of commands, such as the
// Cart Service's execute: the JSON object of --file, sent as a payload is,
// whose member commands holds up to most commands, which Emporix runs in
// order and answers with the result of each that ran (emporix.Commands).
type chain struct {
	payload
	most int // the most commands that Emporix takes in one request, as its reference documents
}

// bind adds --file as a payload does and says in the help how many commands
// Emporix takes and how it answers.
func (c *chain) bind(cmd *cobra.Command) string {
	stdin := c.payload.bind(cmd)
	cmd.Long += "\n\nEmporix takes up to " + strconv.Itoa(c.most) + " commands in one request (epx refuses more), " +
		"runs them in\norder and answers with the result of each that ran: when one failed, epx prints the\n" +
		"response, lists the commands that failed on standard error (of a command that adds\nor updates items " +
		"in a batch, the items that failed; Emporix carried out the others)\nand " + bulkExitSentence
	return stdin
}

func (c *chain) fill(ctx context.Context, w *emporix.Write, values []string) error {
	err := c.payload.fill(ctx, w, values)
	w.Parts = emporix.Commands(c.most)
	return err
}

package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// Emporix answers some writes, such as an order's recalculation, with bare text labelled application/json.
type textAnswer struct {
	what string
}

func (t textAnswer) bind(_ *app, cmd *cobra.Command) endingPlanner {
	cmd.Long += "\n\nEmporix answers with " + t.what + ", which epx prints as a JSON string."
	return t
}

func (textAnswer) plan(_ *cobra.Command, write []string) ([]string, answered) {
	return write, func(_ *emporix.Client, response []byte) (json.RawMessage, error) {
		return textAsJSON(response)
	}
}

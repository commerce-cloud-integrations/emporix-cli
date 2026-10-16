package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// textAnswer declares that Emporix answers a write with bare text, although
// labelled JSON, such as the sentence with which it confirms that it
// recalculated an order (live): epx prints the text as a JSON string
// (textAsJSON), as -o json or yaml, where it would otherwise only say that
// the response is not JSON. An empty answer prints nothing.
type textAnswer struct {
	what string // the text in help, e.g. "a sentence of text"
}

// bind implements answerer: it says in the help of cmd how the answer is
// printed, and returns how a run of cmd ends.
func (t textAnswer) bind(_ *app, cmd *cobra.Command) ending {
	cmd.Long += "\n\nEmporix answers with " + t.what + ", which epx prints as a JSON string."
	return t
}

// plan implements ending: the run requests the write's scopes alone and
// prints Emporix's answer as JSON (textAsJSON).
func (textAnswer) plan(_ *cobra.Command, write []string) ([]string, answered) {
	return write, func(_ *emporix.Client, response []byte) (json.RawMessage, error) {
		return textAsJSON(response)
	}
}

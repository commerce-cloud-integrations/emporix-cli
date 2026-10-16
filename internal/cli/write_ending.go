package cli

import "github.com/spf13/cobra"

// answerer declares how a run of a write ends once Emporix answered it
// (write.answer), other than by printing the response; each implementation
// says how.
type answerer interface {
	// bind adds the flags of the answer to cmd, a write command of a, says in
	// its help what they do, and returns how a run of cmd ends.
	bind(a *app, cmd *cobra.Command) ending
}

// ending is how a run of a write ends once Emporix answered it, as its
// answerer declares, or by printing the response (printing).
type ending interface {
	// plan returns the scopes of a run of cmd, a write that requests write,
	// and what the run does with Emporix's answer.
	plan(cmd *cobra.Command, write []string) ([]string, answered)
}

// printing ends a write without an answerer: it prints the response and
// requests the write's scopes alone.
type printing struct{}

func (printing) plan(_ *cobra.Command, write []string) ([]string, answered) {
	return write, asAnswered
}

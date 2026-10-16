package cli

import (
	"context"
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether f, the standard input or output, is a terminal
// a person reads and answers on: neither a file, a pipe nor the null device,
// which a script's empty input often is.
func IsTerminal(f *os.File) bool { return term.IsTerminal(int(f.Fd())) }

// ReadHidden returns a reader of one line from the terminal f that does not
// show what is typed, for ProfileEnv.ReadHidden. An interrupt ends the wait
// and restores the terminal, which the read left without echo.
func ReadHidden(f *os.File) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		fd := int(f.Fd())
		state, err := term.GetState(fd)
		if err != nil {
			return "", err
		}
		defer func() { _ = term.Restore(fd, state) }()
		line, err := untilDone(ctx, "waiting for input", func() ([]byte, error) { return term.ReadPassword(fd) })
		return string(line), err
	}
}

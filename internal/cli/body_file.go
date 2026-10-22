package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// bodyFile is a flag that gives a request body whole: one JSON value in a
// file, or on standard input for - (readJSON). It is --body of a read with a
// body and --file of a write; the flag of a file that a form sends is one too
// (filePartOf).
type bodyFile struct {
	flag  string // its name
	limit int    // the largest body it reads, in bytes
	array bool   // it gives a JSON array of one or more items, such as JSON Patch operations: else a usage error
}

// bind adds the flag to cmd with usage, completing file names, and returns
// the body that it gives in a run.
func (f bodyFile) bind(cmd *cobra.Command, usage string) *givenBody {
	given := &givenBody{bodyFile: f, cmd: cmd}
	cmd.Flags().StringVar(&given.file, f.flag, "", usage)
	completeFileNames(cmd, f.flag)
	return given
}

// givenBody is the body that the flag of a bodyFile gives in a run of cmd.
type givenBody struct {
	bodyFile
	cmd  *cobra.Command
	file string // the flag's value: a path, or - for standard input
}

// given reports whether the flag is given.
func (g *givenBody) given() bool { return g.cmd.Flags().Changed(g.flag) }

// fromStdin reports whether standard input gives the body (-).
func (g *givenBody) fromStdin() bool { return g.given() && g.file == "-" }

// read returns the content of the file, as load reads it, unless ctx ends
// first (an interrupt): standard input or a named pipe may never end, and the
// read left waiting on one ends with the process.
func (g *givenBody) read(ctx context.Context) ([]byte, error) {
	return untilDone(ctx, g.named(g.file), func() ([]byte, error) { return g.load(g.file, g.cmd.InOrStdin()) })
}

// readJSON returns the body in the file, as read reads it, which must be one
// JSON value (isJSON), an array where the bodyFile says so: one that is not is
// a usage error, as nothing was sent.
func (g *givenBody) readJSON(ctx context.Context) (json.RawMessage, error) {
	data, err := g.read(ctx)
	if err != nil {
		return nil, err
	}
	if err := isJSON(data); err != nil {
		return nil, g.failure(g.file, err)
	}
	if g.array && !isJSONArray(data) {
		return nil, g.failure(g.file, errors.New("not a JSON array of one or more items"))
	}
	return data, nil
}

// isJSON returns why data is not one JSON value; nil when it is.
func isJSON(data []byte) error {
	if err := json.Unmarshal(data, new(json.RawMessage)); err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return nil
}

// isJSONArray reports whether data, one JSON value, is an array of one or
// more items (null is none, although it decodes into a nil slice).
func isJSONArray(data []byte) bool {
	var items []json.RawMessage
	return json.Unmarshal(data, &items) == nil && len(items) > 0
}

// fileName returns the name of the file that g names, which a form sends
// with its content: the last element of its path. Standard input has no
// name to send: a usage error.
func (g *givenBody) fileName() (string, error) {
	if g.fromStdin() {
		return "", usageError{fmt.Errorf("%s: give the path of the file, whose name is sent with it", g.named(g.file))}
	}
	return filepath.Base(g.file), nil
}

// untilDone returns what read returns, unless ctx ends first; then the read
// is left waiting, and the error names it as what, e.g. `--body "-"`.
func untilDone[T any](ctx context.Context, what string, read func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}
	done := make(chan result, 1) // buffered: a read left waiting never blocks on it once it ends
	go func() {
		value, err := read()
		done <- result{value: value, err: err}
	}()
	select {
	case got := <-done:
		return got.value, got.err
	case <-ctx.Done():
		var none T
		return none, fmt.Errorf("%s: %w", what, context.Cause(ctx))
	}
}

// load returns the content of file, or of stdin for "-", whatever it holds,
// up to f.limit bytes. Failures are usage errors, as nothing was sent.
func (f bodyFile) load(file string, stdin io.Reader) ([]byte, error) {
	in := stdin
	if file != "-" {
		// The user names the file to send, as with any file argument; it is only read.
		opened, err := os.Open(file) //nolint:gosec // G304: the path is the user's own --body or --file argument.
		if err != nil {
			return nil, f.failure(file, err)
		}
		defer func() { _ = opened.Close() }()
		in = opened
	}
	data, err := io.ReadAll(io.LimitReader(in, int64(f.limit)+1))
	switch {
	case err != nil:
		return nil, f.failure(file, err)
	case len(data) > f.limit:
		return nil, f.failure(file, fmt.Errorf("larger than %d bytes", f.limit))
	}
	return data, nil
}

// named names the flag given file in messages, e.g. `--body "-"`.
func (f bodyFile) named(file string) string { return fmt.Sprintf("--%s %q", f.flag, file) }

// failure is the usage error of a file that cannot be sent, naming it once (a
// file error names it too).
func (f bodyFile) failure(file string, err error) error {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}
	return usageError{fmt.Errorf("%s: %w", f.named(file), err)}
}

// boundBody is the body of a request once its flags are bound: the one
// that the flag of a bodyFile gives whole, else the one that typed flags
// make, if any; none without either.
type boundBody struct {
	file  *givenBody // the flag that gives it whole: --body of a read, --file of a write
	typed func() any // the body that typed flags make, once flags are parsed; none when nil
}

// bindWith adds the flag of f to cmd with usage, as bind does, then the flags
// of typed, if any, each of which the flag of f excludes (bodyBinder.bind),
// and returns the body of a run.
func (f bodyFile) bindWith(cmd *cobra.Command, usage string, typed bodyBinder) boundBody {
	body := boundBody{file: f.bind(cmd, usage)}
	if typed != nil {
		body.typed = typed.bind(cmd, f.flag)
	}
	return body
}

// make returns the JSON body of a run: the one the flag gives, which must be
// JSON, else the one that typed flags make, if any; none otherwise (the zero
// emporix.Body).
func (b boundBody) make(ctx context.Context) (emporix.Body, error) {
	switch {
	case b.file.given():
		data, err := b.file.readJSON(ctx)
		if err != nil {
			return emporix.Body{}, err
		}
		return emporix.JSONBody(data)
	case b.typed != nil:
		return emporix.JSONBodyOf(b.typed())
	default:
		return emporix.Body{}, nil
	}
}

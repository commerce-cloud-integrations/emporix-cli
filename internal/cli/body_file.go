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

type bodyFile struct {
	flag          string
	limit         int
	nonEmptyArray bool
}

func (f bodyFile) bind(cmd *cobra.Command, usage string) *givenBody {
	given := &givenBody{bodyFile: f, cmd: cmd}
	cmd.Flags().StringVar(&given.file, f.flag, "", usage)
	completeFileNames(cmd, f.flag)
	return given
}

type givenBody struct {
	bodyFile
	cmd  *cobra.Command
	file string
}

func (g *givenBody) given() bool { return g.cmd.Flags().Changed(g.flag) }

func (g *givenBody) fromStdin() bool { return g.given() && g.file == "-" }

func (g *givenBody) read(ctx context.Context) ([]byte, error) {
	return untilDone(ctx, g.named(g.file), func() ([]byte, error) { return g.load(g.file, g.cmd.InOrStdin()) })
}

func (g *givenBody) readJSON(ctx context.Context) (json.RawMessage, error) {
	data, err := g.read(ctx)
	if err != nil {
		return nil, err
	}
	if err := isJSON(data); err != nil {
		return nil, g.failure(g.file, err)
	}
	if g.nonEmptyArray && !isNonEmptyJSONArray(data) {
		return nil, g.failure(g.file, errors.New("not a JSON array of one or more items"))
	}
	return data, nil
}

func isJSON(data []byte) error {
	if err := json.Unmarshal(data, new(json.RawMessage)); err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return nil
}

func isNonEmptyJSONArray(data []byte) bool {
	var items []json.RawMessage
	return json.Unmarshal(data, &items) == nil && len(items) > 0
}

func (g *givenBody) fileName() (string, error) {
	if g.fromStdin() {
		return "", usageError{fmt.Errorf("%s: give the path of the file, whose name is sent with it", g.named(g.file))}
	}
	return filepath.Base(g.file), nil
}

func untilDone[T any](ctx context.Context, what string, read func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}
	done := make(chan result, 1)
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

func (f bodyFile) load(file string, stdin io.Reader) ([]byte, error) {
	in := stdin
	if file != "-" {
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

func (f bodyFile) named(file string) string { return fmt.Sprintf("--%s %q", f.flag, file) }

func (f bodyFile) failure(file string, err error) error {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}
	return usageError{fmt.Errorf("%s: %w", f.named(file), err)}
}

type boundBody struct {
	file  *givenBody
	typed func() any
}

func (f bodyFile) bindWith(cmd *cobra.Command, usage string, typed bodyBinder) boundBody {
	body := boundBody{file: f.bind(cmd, usage)}
	if typed != nil {
		body.typed = typed.bind(cmd, f.flag)
	}
	return body
}

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

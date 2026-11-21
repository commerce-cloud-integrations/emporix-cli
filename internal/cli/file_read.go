package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const outFlag = "out"

const standardOutput = "-"

type fileRead struct {
	command
	example string
	noun    string
	media   string
}

func (rd reader[T]) fileCommand(a *app, f *fileRead) *cobra.Command {
	cmd, target := newRead(a, &f.command, rd.scopes)
	cmd.Example = f.example
	out, lang := &outPath{terminal: a.outTerminal}, language("")
	cmd.Flags().Var(out, outFlag, "save the "+f.noun+" to the file `PATH`, or with - write it to standard output")
	requireFlag(cmd, outFlag)
	completeFileNames(cmd, outFlag)
	cmd.Flags().Var(&lang, languageFlag, "the `LANGUAGE` of the "+f.noun+"'s text, e.g. de, sent as typed: "+
		"write it as the tenant configures it (default: the tenant's default language)")
	cmd.Long += "\n\n" + changesNothing(http.MethodPost, "no body.") + " Emporix answers with the " + f.noun +
		", which epx saves to the file that --" + outFlag + " names (readable by its owner only; a file there is " +
		"replaced, but not a directory, device or link), or writes to standard output with --" + outFlag +
		" -, unless that is a terminal. An answer that is no " + f.noun + " fails (exit 1), and nothing is " +
		"saved."
	cmd.Long = target.rule.helpEnding(cmd.Long, scopeSentence(rd.scopes))
	cmd.RunE = func(cc *cobra.Command, args []string) error {
		client, path, query, err := target.open(cc.Context(), args)
		if err != nil {
			return err
		}
		file, err := client.Posting().GetFile(cc.Context(), path, query, f.media, string(lang))
		if err != nil {
			return err
		}
		return out.save(cc, f.noun, file)
	}
	return cmd
}

func replaceFile(path string, body []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	_, err = temp.Write(body)
	if closed := temp.Close(); err == nil {
		err = closed
	}
	if err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

type outPath struct {
	path     string
	terminal bool
}

func (o *outPath) Set(value string) error {
	switch {
	case value == "":
		return errors.New("empty") // the flag parser names the flag and the value
	case value == standardOutput && o.terminal:
		return errors.New("standard output is a terminal: name a file, or redirect standard output")
	case value != standardOutput:
		if err := savable(value); err != nil {
			return err
		}
	}
	o.path = value
	return nil
}

func (o *outPath) String() string { return o.path }

func (o *outPath) Type() string { return "path" }

func (o *outPath) save(cmd *cobra.Command, noun string, body []byte) error {
	if o.path == standardOutput {
		_, err := cmd.OutOrStdout().Write(body)
		return err
	}
	if err := replaceFile(o.path, body); err != nil {
		return fmt.Errorf("the %s was not saved: %w", noun, err)
	}
	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Saved the %s (%d bytes) to %s.\n", noun, len(body), o.path)
	return err
}

func savable(path string) error {
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if dir, statErr := os.Stat(filepath.Dir(path)); statErr != nil || !dir.IsDir() {
			return errors.New("its directory does not exist")
		}
		return nil
	case err != nil:
		return err
	case !info.Mode().IsRegular():
		return fmt.Errorf("a %s, not a file: name a file, or use %s for standard output", fileKind(info.Mode()),
			standardOutput)
	default:
		return nil
	}
}

func fileKind(mode fs.FileMode) string {
	switch {
	case mode.IsDir():
		return "directory"
	case mode&fs.ModeSymlink != 0:
		return "symbolic link"
	default:
		return "device, pipe or socket"
	}
}

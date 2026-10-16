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

// outFlag is the flag of a read that Emporix answers with a file
// (fileRead), which names where the file is saved.
const outFlag = "out"

// standardOutput is the --out that writes the file to standard output.
const standardOutput = "-"

// fileRead declares a read that Emporix answers with a file instead of JSON,
// such as the PDF of a quote, which Emporix serves as POST without a body
// although it changes nothing (emporix.Client.Posting): its command
// (reader.fileCommand) saves the file where --out names instead of printing
// it, in the language of --language. Its fetch is none: it reads the file.
type fileRead struct {
	command
	example string // command lines that show it, each indented by two spaces
	noun    string // the file in help and messages, e.g. "PDF"
	media   string // its media type, e.g. application/pdf
}

// fileCommand returns the read command f of rd's scopes (newRead), with
// --out, required, and --language, which reads the file
// (emporix.Client.GetFile) and saves it where --out names (outPath.save).
func (rd reader[T]) fileCommand(a *app, f *fileRead) *cobra.Command {
	cmd, target := newRead(a, &f.command, rd.scopes)
	cmd.Example = f.example
	out, lang := &outPath{terminal: a.outTerminal}, language("")
	cmd.Flags().Var(out, outFlag, "save the "+f.noun+" to the file `PATH`, or with - write it to standard output")
	requireFlag(cmd, outFlag)
	completeFileNames(cmd, outFlag)
	cmd.Flags().Var(&lang, languageFlag, "the `LANGUAGE` of the "+f.noun+"'s text, e.g. de, as the tenant "+
		"configures it (default: the tenant's default language)")
	cmd.Long += "\n\n" + changesNothing(http.MethodPost, "no body.") + "\nEmporix answers with the " + f.noun +
		", which epx saves to the file that --" + outFlag + " names (readable by\nits owner only; a file there is " +
		"replaced, but not a directory, device or link), or writes to\nstandard output with --" + outFlag +
		" -, unless that is a terminal. An answer that is no " + f.noun + " fails\n(exit 1), and nothing is " +
		"saved.\n\n" + scopeSentence(rd.scopes)
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

// replaceFile writes body to the file at path, replacing any file there only
// once body is written whole: it writes a new file in the same directory,
// readable by its owner only, and renames it to path.
func replaceFile(path string, body []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }() // fails once renamed, as nothing is left to remove
	_, err = temp.Write(body)
	if closed := temp.Close(); err == nil {
		err = closed
	}
	if err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

// outPath is the value of --out: - for standard output, which is refused
// where standard output is a terminal, or the path of a file that replaceFile
// can save (savable); it implements pflag.Value, so another value is
// rejected while flags are parsed, before anything is sent.
type outPath struct {
	path     string
	terminal bool // standard output is a terminal
}

// Set implements pflag.Value.
func (o *outPath) Set(value string) error {
	switch {
	case value == "":
		return errors.New("empty") // the flag parser names the flag and the value
	case value == standardOutput && o.terminal:
		return errors.New("standard output is a terminal: name a file, or redirect standard output")
	case value != standardOutput:
		if err := savable(value); err != nil {
			return err // the flag parser names the flag and the value
		}
	}
	o.path = value
	return nil
}

func (o *outPath) String() string { return o.path }

// Type implements pflag.Value.
func (o *outPath) Type() string { return "path" }

// save saves body, the noun that Emporix answered with, such as a PDF, where
// o names: to standard output, or to a file, after which it says so on
// standard error.
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

// savable fails for a path where replaceFile would not save a file in place:
// one whose directory does not exist, or that names something other than a
// regular file, such as a directory, a device or a symbolic link, which it
// would replace (or fail to replace) rather than write to.
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

// fileKind names what mode, not that of a regular file, is the mode of.
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

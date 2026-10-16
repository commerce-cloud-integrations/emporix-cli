package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// The flags that pick how a command prints its result: -o/--output, its
// format, and for a table, --columns and --no-headers.
const (
	outputFlag    = "output"
	columnsFlag   = "columns"
	noHeadersFlag = "no-headers"
)

// printer prints the result of a command that shows it as a table, as its
// flags pick (addTableOutput): its rows as a table or CSV, in the columns of
// --columns, under the header unless --no-headers is given, or the result
// whole as JSON or YAML.
type printer struct {
	format    formatChoice
	columns   columnPicks
	noHeaders bool
	long      []int // the columns of long values, by their index in the header, which a table cuts (cut)
}

// addTableOutput adds -o (table, json, yaml or csv), --columns and
// --no-headers to cmd, which shows its result as a table of header, and
// returns how it prints: a table cuts the values of the columns that long
// names (cut), which CSV prints whole. --columns or --no-headers given with
// a format that prints the result whole fails before cmd runs (check). A
// long column that header lacks is a mistake in the declaration of cmd, so
// it panics while the command tree is built, which every test of the tree
// does.
func addTableOutput(cmd *cobra.Command, header []string, long ...string) *printer {
	p := &printer{
		format:  formatChoice{offered: []output.Format{output.Table, output.JSON, output.YAML, output.CSV}},
		columns: columnPicks{header: header},
	}
	for _, name := range long {
		column := slices.Index(header, name)
		if column < 0 {
			panic(fmt.Sprintf("%s: %s is not a column of %q", cmd.Use, name, header))
		}
		p.long = append(p.long, column)
	}
	p.format.add(cmd, "output format: %s (csv: the table's rows; json and yaml: the whole response)")
	cmd.Flags().Var(&p.columns, columnsFlag, "print only these columns of table and csv, in this order: "+
		"comma-separated `names` (any case) among "+strings.Join(header, ", "))
	cmd.Flags().BoolVar(&p.noHeaders, noHeadersFlag, false, "print table and csv without the header row")
	addPreRunCheck(cmd, p.check)
	return p
}

// addResponseOutput adds -o (json or yaml) to cmd, which prints a response
// whole in the format it returns.
func addResponseOutput(cmd *cobra.Command) *formatChoice {
	format := &formatChoice{offered: []output.Format{output.JSON, output.YAML}}
	format.add(cmd, "format of the response: %s")
	return format
}

// check fails, as a usage error, --columns or --no-headers given with a
// format that prints the result whole, without columns.
func (p *printer) check() error {
	if p.format.chosen().Whole() && (len(p.columns.picked) > 0 || p.noHeaders) {
		return usageError{fmt.Errorf("--%s and --%s apply only to -o %s and -o %s", columnsFlag, noHeadersFlag,
			output.Table, output.CSV)}
	}
	return nil
}

// print writes a result to w: whole in a format that prints it whole, else
// the rows that rows returns, their long values cut in a table, in the
// columns that --columns picks, under the header unless --no-headers is
// given.
func (p *printer) print(w io.Writer, whole any, rows func() ([][]string, error)) error {
	format := p.format.chosen()
	if format.Whole() {
		return output.Write(w, format, whole)
	}
	all, err := rows()
	if err != nil {
		return err
	}
	if format == output.Table {
		p.cutLong(all)
	}
	header, picked := p.columns.pick(all)
	if p.noHeaders {
		header = nil
	}
	return output.WriteRows(w, format, header, picked)
}

// cutLong cuts the long values of rows, as a table shows them (cut).
func (p *printer) cutLong(rows [][]string) {
	for _, row := range rows {
		for _, column := range p.long {
			row[column] = cut(row[column])
		}
	}
}

// formatChoice is -o/--output: one of the formats that a command offers, its
// first unless another is given. It implements pflag.Value, so a format that
// the command does not offer fails while flags are parsed.
type formatChoice struct {
	offered []output.Format
	given   output.Format // the format given; none when -o is not given
}

// add adds f to cmd as -o/--output, whose help is usage with the formats
// offered in place of its %s; shells complete them.
func (f *formatChoice) add(cmd *cobra.Command, usage string) {
	cmd.Flags().VarP(f, outputFlag, "o", fmt.Sprintf(usage, alternatives(f.names())))
	completeChoices(cmd, outputFlag, f.names())
}

// chosen returns the format that f gives: the one given, else its first.
func (f *formatChoice) chosen() output.Format {
	if f.given == "" {
		return f.offered[0]
	}
	return f.given
}

// names returns the formats that f offers, as -o takes them.
func (f *formatChoice) names() []string {
	names := make([]string, len(f.offered))
	for i, format := range f.offered {
		names[i] = string(format)
	}
	return names
}

// Set implements pflag.Value.
func (f *formatChoice) Set(value string) error {
	if !slices.Contains(f.offered, output.Format(value)) {
		return fmt.Errorf("invalid output format %q: use %s", value, alternatives(f.names()))
	}
	f.given = output.Format(value)
	return nil
}

func (f *formatChoice) String() string { return string(f.chosen()) }

// Type implements pflag.Value.
func (f *formatChoice) Type() string { return "format" }

// columnPicks is --columns: the columns of a table that a command prints, in
// the order given, each named by its header in any case; every column when
// it is not given. It implements pflag.Value, so a name that is no column,
// or a column named twice, fails while flags are parsed.
type columnPicks struct {
	header []string
	picked []int // the picked columns, by their index in header, in the order given
}

// Set implements pflag.Value. Each --columns adds its columns to those of
// the --columns before it.
func (c *columnPicks) Set(value string) error {
	for name := range strings.SplitSeq(value, ",") {
		name = strings.TrimSpace(name)
		column := slices.IndexFunc(c.header, func(header string) bool { return strings.EqualFold(header, name) })
		if column < 0 {
			return fmt.Errorf("no column %q: use %s", name, strings.Join(c.header, ", "))
		}
		if slices.Contains(c.picked, column) {
			return fmt.Errorf("column %s named twice", c.header[column])
		}
		c.picked = append(c.picked, column)
	}
	return nil
}

func (c *columnPicks) String() string { return strings.Join(c.cells(c.header), ",") }

// Type implements pflag.Value.
func (c *columnPicks) Type() string { return "names" }

// pick returns the header and the cells of rows in the picked columns, in
// their order: every column when none is picked.
func (c *columnPicks) pick(rows [][]string) (header []string, picked [][]string) {
	if len(c.picked) == 0 {
		return c.header, rows
	}
	picked = make([][]string, len(rows))
	for i, row := range rows {
		picked[i] = c.cells(row)
	}
	return c.cells(c.header), picked
}

// cells returns the cells of row in the picked columns, in their order.
func (c *columnPicks) cells(row []string) []string {
	cells := make([]string, len(c.picked))
	for i, column := range c.picked {
		cells[i] = row[column]
	}
	return cells
}

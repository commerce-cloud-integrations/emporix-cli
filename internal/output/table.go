package output

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// WriteTable writes rows under header, unless it is empty, as space-aligned
// columns. Cell text is sanitized: API data must not inject tabs, newlines or
// terminal escapes.
func WriteTable(w io.Writer, header []string, rows [][]string) error {
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, row := range sanitized(header, rows) {
		if _, err := fmt.Fprintln(table, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return table.Flush()
}

// WriteCSV writes rows under header, unless it is empty, as CSV records
// (RFC 4180 quoting, each record ending in a line feed), with cell text
// sanitized as WriteTable sanitizes it. A record of one empty cell is written
// as "", not as the empty line that encoding/csv writes and CSV readers skip.
func WriteCSV(w io.Writer, header []string, rows [][]string) error {
	var out bytes.Buffer
	records := csv.NewWriter(&out)
	for _, record := range sanitized(header, rows) {
		if len(record) == 1 && record[0] == "" {
			records.Flush()
			out.WriteString(`""` + "\n")
			continue
		}
		if err := records.Write(record); err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	}
	records.Flush()
	if _, err := w.Write(out.Bytes()); err != nil {
		return fmt.Errorf("writing CSV: %w", err)
	}
	return nil
}

// sanitized returns rows, after header unless it is empty, with every cell
// sanitized for the terminal.
func sanitized(header []string, rows [][]string) [][]string {
	if len(header) > 0 {
		rows = append([][]string{header}, rows...)
	}
	cells := make([][]string, len(rows))
	for i, row := range rows {
		cells[i] = make([]string, len(row))
		for j, cell := range row {
			cells[i][j] = Sanitize(cell)
		}
	}
	return cells
}

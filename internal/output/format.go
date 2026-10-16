// Package output renders command results as indented JSON or YAML, aligned
// tables or CSV, and keeps API text from acting on the terminal.
package output

import (
	"fmt"
	"io"
)

// Format is how a command prints its result: the rows of its table (Table,
// CSV), or the result whole, every value as sent (JSON, YAML).
type Format string

// Supported formats.
const (
	Table Format = "table"
	JSON  Format = "json"
	YAML  Format = "yaml"
	CSV   Format = "csv"
)

// Whole reports whether f prints a result whole rather than the rows of a table.
func (f Format) Whole() bool { return f == JSON || f == YAML }

// Write writes v whole in f, JSON (WriteJSON) or YAML (WriteYAML); a format
// that prints the rows of a table prints nothing whole, so it fails.
func Write(w io.Writer, f Format, v any) error {
	switch f {
	case JSON:
		return WriteJSON(w, v)
	case YAML:
		return WriteYAML(w, v)
	default:
		return fmt.Errorf("output format %q prints no result whole", f)
	}
}

// WriteRows writes rows under header, if any, in f, a table (WriteTable) or
// CSV (WriteCSV); a format that prints a result whole has no rows, so it
// fails.
func WriteRows(w io.Writer, f Format, header []string, rows [][]string) error {
	switch f {
	case Table:
		return WriteTable(w, header, rows)
	case CSV:
		return WriteCSV(w, header, rows)
	default:
		return fmt.Errorf("output format %q prints no rows", f)
	}
}

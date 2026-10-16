package output

import (
	"bytes"
	"encoding/csv"
	"slices"
	"strings"
	"testing"
)

func TestWriteTableAlignsAndSanitizes(t *testing.T) {
	var out bytes.Buffer
	rows := [][]string{{"AcmeStore_NL", "AcmeStore\tNL\x1b[31m"}, {"main", "Default\nconfig\u202e"}}
	if err := WriteTable(&out, []string{"CODE", "NAME"}, rows); err != nil {
		t.Fatal(err)
	}
	want := "CODE          NAME\nAcmeStore_NL  AcmeStore NL [31m\nmain          Default config \n"
	if out.String() != want {
		t.Fatalf("WriteTable =\n%q\nwant\n%q", out.String(), want)
	}
	if strings.ContainsRune(out.String(), '\x1b') {
		t.Fatal("escape character reached the terminal")
	}
}

func TestWriteTableWithoutHeader(t *testing.T) {
	var out bytes.Buffer
	if err := WriteTable(&out, nil, [][]string{{"AcmeStore_NL", "1"}, {"main", "22"}}); err != nil {
		t.Fatal(err)
	}
	if want := "AcmeStore_NL  1\nmain          22\n"; out.String() != want {
		t.Fatalf("WriteTable = %q, want %q", out.String(), want)
	}
}

func TestWriteTableReportsWriteErrors(t *testing.T) {
	if err := WriteTable(failingWriter{}, []string{"A"}, nil); err == nil {
		t.Fatal("write error ignored")
	}
}

// TestWriteCSVQuotesAndSanitizes checks that CSV quotes what needs quoting
// (commas, quotes, a leading space), keeps cells as given (leading zeros,
// spaces), and sanitizes what a terminal could act on, as tables do.
func TestWriteCSVQuotesAndSanitizes(t *testing.T) {
	var out bytes.Buffer
	rows := [][]string{
		{"100200", `Shirt, "blue"`, " 001 "},
		{"100201", "line\nbreak\x1b[31m\u202e", ""},
	}
	if err := WriteCSV(&out, []string{"ID", "NAME", "CODE"}, rows); err != nil {
		t.Fatal(err)
	}
	want := "ID,NAME,CODE\n100200,\"Shirt, \"\"blue\"\"\",\" 001 \"\n100201,line break [31m ,\n"
	if out.String() != want {
		t.Fatalf("WriteCSV =\n%q\nwant\n%q", out.String(), want)
	}
	records, err := csv.NewReader(strings.NewReader(out.String())).ReadAll()
	if err != nil || len(records) != 3 || !slices.Equal(records[1], []string{"100200", `Shirt, "blue"`, " 001 "}) {
		t.Fatalf("CSV reads back as %q (%v)", records, err)
	}
}

func TestWriteCSVWithoutHeaderAndErrors(t *testing.T) {
	var out bytes.Buffer
	if err := WriteCSV(&out, nil, [][]string{{"a", "b"}}); err != nil || out.String() != "a,b\n" {
		t.Fatalf("WriteCSV = %q (%v), want one record", out.String(), err)
	}
	if err := WriteCSV(failingWriter{}, []string{"A"}, nil); err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("write error %v; want it reported", err)
	}
}

// TestWriteCSVKeepsEmptyRecords checks that a record of one empty cell, which
// a single column can give, reads back as a record, with and without the
// header, where encoding/csv alone writes an empty line that readers skip.
func TestWriteCSVKeepsEmptyRecords(t *testing.T) {
	rows := [][]string{{""}, {"B"}, {""}, {""}}
	for _, tt := range []struct {
		header []string
		want   string
	}{
		{header: []string{"CODE"}, want: "CODE\n\"\"\nB\n\"\"\n\"\"\n"},
		{want: "\"\"\nB\n\"\"\n\"\"\n"},
	} {
		var out bytes.Buffer
		if err := WriteCSV(&out, tt.header, rows); err != nil || out.String() != tt.want {
			t.Fatalf("WriteCSV = %q (%v), want %q", out.String(), err, tt.want)
		}
		records, err := csv.NewReader(strings.NewReader(out.String())).ReadAll()
		wanted := rows
		if len(tt.header) > 0 {
			wanted = append([][]string{tt.header}, rows...)
		}
		if err != nil || !slices.EqualFunc(records, wanted, slices.Equal) {
			t.Fatalf("CSV reads back as %q (%v), want %q", records, err, wanted)
		}
	}
}

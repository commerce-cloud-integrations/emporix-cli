package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteWritesTheFormat(t *testing.T) {
	body := json.RawMessage(`{"code":"AcmeStore_NL"}`)
	header, rows := []string{"CODE", "NAME"}, [][]string{{"AcmeStore_NL", "Acme, NL"}}
	for _, tt := range []struct {
		format Format
		whole  bool
		want   string
	}{
		{format: Table, want: "CODE          NAME\nAcmeStore_NL  Acme, NL\n"},
		{format: CSV, want: "CODE,NAME\nAcmeStore_NL,\"Acme, NL\"\n"},
		{format: JSON, whole: true, want: "{\n  \"code\": \"AcmeStore_NL\"\n}\n"},
		{format: YAML, whole: true, want: "\"code\": \"AcmeStore_NL\"\n"},
	} {
		var out bytes.Buffer
		var err error
		if tt.format.Whole() != tt.whole {
			t.Fatalf("%s: Whole() = %v", tt.format, !tt.whole)
		}
		if tt.whole {
			err = Write(&out, tt.format, body)
		} else {
			err = WriteRows(&out, tt.format, header, rows)
		}
		if err != nil || out.String() != tt.want {
			t.Errorf("%s: %q (%v), want %q", tt.format, out.String(), err, tt.want)
		}
	}
}

// TestWriteRefusesAFormatOfTheOtherKind checks that a format that prints
// rows writes nothing whole, and one that prints a result whole no rows.
func TestWriteRefusesAFormatOfTheOtherKind(t *testing.T) {
	var out bytes.Buffer
	for _, format := range []Format{Table, CSV, "xml"} {
		if err := Write(&out, format, json.RawMessage(`{}`)); err == nil {
			t.Errorf("Write in %q succeeded", format)
		}
	}
	for _, format := range []Format{JSON, YAML, "xml"} {
		if err := WriteRows(&out, format, []string{"A"}, [][]string{{"a"}}); err == nil {
			t.Errorf("WriteRows in %q succeeded", format)
		}
	}
	if out.Len() > 0 {
		t.Fatalf("wrote %q", out.String())
	}
}

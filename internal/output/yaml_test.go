package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestWriteYAMLWritesTheDataOfJSON checks the block layout of every kind of
// value: keys sorted, maps and lists nested in each other, empty
// collections in flow style, numbers as given and every string quoted, so
// that none reads as another type.
func TestWriteYAMLWritesTheDataOfJSON(t *testing.T) {
	var out bytes.Buffer
	body := json.RawMessage(`{"z":1.50,"a":{"id":"001234","tags":[]},"list":[{"b":true,"a":null},["x",["y"]],"s",{}],
		"no":"no","empty":{},"n":-1e+400,"html":"<b>&é"}`)
	if err := WriteYAML(&out, body); err != nil {
		t.Fatal(err)
	}
	want := `"a":
  "id": "001234"
  "tags": []
"empty": {}
"html": "<b>&é"
"list":
  - "a": null
    "b": true
  - - "x"
    - - "y"
  - "s"
  - {}
"n": -1e+400
"no": "no"
"z": 1.50
`
	if out.String() != want {
		t.Fatalf("WriteYAML =\n%s\nwant\n%s", out.String(), want)
	}
}

func TestWriteYAMLWritesValuesAndScalars(t *testing.T) {
	for _, tt := range []struct {
		value any
		want  string
	}{
		{value: json.RawMessage(`"yes"`), want: "\"yes\"\n"},
		{value: json.RawMessage(`null`), want: "null\n"},
		{value: json.RawMessage(`[]`), want: "[]\n"},
		{value: json.RawMessage(`{}`), want: "{}\n"},
		{value: json.RawMessage(`[[],{},[1]]`), want: "- []\n- {}\n- - 1\n"},
		{value: struct {
			Scopes []string `json:"scopes"`
		}{Scopes: []string{"a"}}, want: "\"scopes\":\n  - \"a\"\n"},
	} {
		var out bytes.Buffer
		if err := WriteYAML(&out, tt.value); err != nil || out.String() != tt.want {
			t.Errorf("WriteYAML(%v) = %q (%v), want %q", tt.value, out.String(), err, tt.want)
		}
	}
}

// TestWriteYAMLEscapes checks that strings escape as JSON escapes them, but
// with \U for a character beyond U+FFFF, which YAML takes instead of
// surrogates, and also escape what YAML would not read as text.
func TestWriteYAMLEscapes(t *testing.T) {
	text := "q\"b\\n\nt\tr\ra\u202eb\u009b[31m\u007fc\u200bd\U000e0041\u001be\u2028\ufffe\U0001f600"
	var out bytes.Buffer
	if err := WriteYAML(&out, map[string]string{"name\u202e": text}); err != nil {
		t.Fatal(err)
	}
	want := `"name\u202e": "q\"b\\n\nt\tr\ra\u202eb\u009b[31m\u007fc\u200bd\U000e0041\u001be\u2028\ufffe` +
		"\U0001f600\"\n"
	if out.String() != want {
		t.Fatalf("WriteYAML =\n%q\nwant\n%q", out.String(), want)
	}
	for _, r := range out.String() {
		if r != '\n' && isTerminalActive(r) {
			t.Fatalf("%U reached the terminal", r)
		}
	}
}

// TestWriteYAMLWritesALongKeyAsAnExplicitKey checks that a key longer than
// YAML allows before a colon is written after a question mark, also in a
// list and with a block value.
func TestWriteYAMLWritesALongKeyAsAnExplicitKey(t *testing.T) {
	long := strings.Repeat("k", maxImplicitKey-1) // with its quotes, one character too many
	fits := strings.Repeat("k", maxImplicitKey-2)
	var out bytes.Buffer
	value := []map[string]any{{long: []int{1}, fits: 2}}
	if err := WriteYAML(&out, value); err != nil {
		t.Fatal(err)
	}
	want := `- "` + fits + `": 2` + "\n" + `  ? "` + long + `"` + "\n  :\n    - 1\n"
	if out.String() != want {
		t.Fatalf("WriteYAML =\n%s\nwant\n%s", out.String(), want)
	}
}

func TestWriteYAMLErrors(t *testing.T) {
	if err := WriteYAML(&bytes.Buffer{}, json.RawMessage(`{"a":`)); err == nil {
		t.Fatal("invalid raw JSON accepted")
	}
	if err := WriteYAML(&bytes.Buffer{}, func() {}); err == nil {
		t.Fatal("unencodable value accepted")
	}
	if err := WriteYAML(failingWriter{}, json.RawMessage(`{}`)); err == nil {
		t.Fatal("write error ignored")
	}
}

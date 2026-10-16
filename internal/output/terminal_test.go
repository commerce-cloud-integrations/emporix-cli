package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONEscapesInvisibleCharacters(t *testing.T) {
	text := "a\u202eb\u009b[31m\u007fc\u200bd\U000e0041\u001be"
	var out bytes.Buffer
	if err := WriteJSON(&out, map[string]string{"name\u202e": text}); err != nil {
		t.Fatal(err)
	}
	want := `{` + "\n" + `  "name\u202e": "a\u202eb\u009b[31m\u007fc\u200bd\udb40\udc41\u001be"` + "\n" + `}` + "\n"
	if out.String() != want {
		t.Fatalf("WriteJSON = %q, want %q", out.String(), want)
	}
	var decoded map[string]string
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil || decoded["name\u202e"] != text {
		t.Fatalf("escaped output decodes to %q (%v), want the original text", decoded, err)
	}
}

package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestWriteJSONSortsKeysAndKeepsNumberAndStringText(t *testing.T) {
	var out bytes.Buffer
	if err := WriteJSON(&out, json.RawMessage(`{"z":1,"a":{"id":"001234"},"n":1.50,"html":"<b>&é"}`)); err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": {\n    \"id\": \"001234\"\n  },\n  \"html\": \"<b>&é\",\n  \"n\": 1.50,\n  \"z\": 1\n}\n"
	if out.String() != want {
		t.Fatalf("WriteJSON = %q, want %q", out.String(), want)
	}
}

func TestWriteJSONEncodesValues(t *testing.T) {
	var out bytes.Buffer
	if err := WriteJSON(&out, struct {
		Scopes []string `json:"scopes"`
	}{Scopes: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "{\n  \"scopes\": [\n    \"a\"\n  ]\n}\n" {
		t.Fatalf("WriteJSON = %q", out.String())
	}
}

func TestWriteJSONErrors(t *testing.T) {
	if err := WriteJSON(&bytes.Buffer{}, json.RawMessage(`{"a":`)); err == nil {
		t.Fatal("invalid raw JSON accepted")
	}
	if err := WriteJSON(&bytes.Buffer{}, func() {}); err == nil {
		t.Fatal("unencodable value accepted")
	}
	if err := WriteJSON(failingWriter{}, json.RawMessage(`{}`)); err == nil {
		t.Fatal("write error ignored")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestWriteJSONLinesWritesOneSortedItemPerLine(t *testing.T) {
	var out bytes.Buffer
	items := []json.RawMessage{json.RawMessage(`{"z":1.50,"a":"<b>"}`), json.RawMessage(` [ "x" ] `),
		json.RawMessage(`"bidi` + "\u202e" + `"`)}
	if err := WriteJSONLines(&out, items); err != nil {
		t.Fatal(err)
	}
	if want := "{\"a\":\"<b>\",\"z\":1.50}\n[\"x\"]\n\"bidi\\u202e\"\n"; out.String() != want {
		t.Fatalf("WriteJSONLines = %q, want %q", out.String(), want)
	}
	if err := WriteJSONLines(&bytes.Buffer{}, []json.RawMessage{json.RawMessage(`{"a":`)}); err == nil {
		t.Fatal("invalid item accepted")
	}
	if err := WriteJSONLines(failingWriter{}, items); err == nil {
		t.Fatal("write error ignored")
	}
}

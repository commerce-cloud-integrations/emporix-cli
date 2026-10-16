package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON writes v (a Go value or a json.RawMessage) as indented JSON with
// sorted object keys, so output is stable and diffable. Values are written as
// given; numbers keep their exact text; characters a terminal could act on are
// escaped (see escapeInvisible).
func WriteJSON(w io.Writer, v any) error {
	data, err := sortedJSON(v, "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(escapeInvisible(data))
	return err
}

// WriteJSONLines writes items as JSON Lines: each on one line, with sorted
// object keys and escaped as WriteJSON escapes them.
func WriteJSONLines(w io.Writer, items []json.RawMessage) error {
	var lines bytes.Buffer
	for _, item := range items {
		data, err := sortedJSON(item, "")
		if err != nil {
			return err
		}
		lines.Write(data)
	}
	_, err := w.Write(escapeInvisible(lines.Bytes()))
	return err
}

// sortedJSON returns v as JSON with sorted object keys and exact numbers,
// indented by indent (none when empty), ending in a newline.
func sortedJSON(v any, indent string) ([]byte, error) {
	tree, err := treeOf(v)
	if err != nil {
		return nil, err
	}
	return encode(tree, indent)
}

// treeOf returns v, a Go value or a json.RawMessage, as the JSON value it
// encodes to: maps, slices, strings, json.Numbers that keep their exact text,
// booleans and nils.
func treeOf(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encoding JSON output: %w", err)
	}
	tree, err := decodeTree(data)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON output: %w", err)
	}
	return tree, nil
}

// encode returns tree as JSON indented by indent, without HTML escaping.
func encode(tree any, indent string) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetIndent("", indent)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(tree); err != nil {
		return nil, fmt.Errorf("encoding JSON output: %w", err)
	}
	return out.Bytes(), nil
}

func decodeTree(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var tree any
	err := decoder.Decode(&tree)
	return tree, err
}

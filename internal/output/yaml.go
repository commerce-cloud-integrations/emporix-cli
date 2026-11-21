package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxImplicitKey is the most characters, quotes included, of a YAML key
// written before its colon; a longer one is written as an explicit key (?).
const maxImplicitKey = 1024

// WriteYAML writes v (a Go value or a json.RawMessage) as YAML 1.2: the data
// that WriteJSON writes, with object keys sorted and numbers as given, in
// block style. Every string, keys included, is a double-quoted scalar, so
// that none reads as another type (such as "no" or "1.0"); it escapes as
// JSON escapes, characters a terminal could act on included, but writes a
// character beyond U+FFFF as \UXXXXXXXX, which YAML takes instead of a pair
// of surrogates.
func WriteYAML(w io.Writer, v any) error {
	tree, err := treeOf(v)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	writeYAMLNode(&out, tree, "")
	_, err = w.Write(out.Bytes())
	return err
}

func writeYAMLNode(out *bytes.Buffer, node any, indent string) {
	switch n := node.(type) {
	case map[string]any:
		writeYAMLMap(out, n, indent)
	case []any:
		writeYAMLList(out, n, indent)
	default:
		out.WriteString(yamlScalar(n))
		out.WriteByte('\n')
	}
}

func writeYAMLMap(out *bytes.Buffer, m map[string]any, indent string) {
	if len(m) == 0 {
		out.WriteString("{}\n")
		return
	}
	for i, key := range slices.Sorted(maps.Keys(m)) {
		if i > 0 {
			out.WriteString(indent)
		}
		writeYAMLKey(out, key, indent)
		out.WriteByte(':')
		if isYAMLBlock(m[key]) {
			out.WriteString("\n" + indent + "  ")
			writeYAMLNode(out, m[key], indent+"  ")
		} else {
			out.WriteByte(' ')
			writeYAMLNode(out, m[key], indent)
		}
	}
}

func writeYAMLKey(out *bytes.Buffer, key, indent string) {
	quoted := yamlString(key)
	if utf8.RuneCountInString(quoted) > maxImplicitKey {
		out.WriteString("? " + quoted + "\n" + indent)
		return
	}
	out.WriteString(quoted)
}

func writeYAMLList(out *bytes.Buffer, items []any, indent string) {
	if len(items) == 0 {
		out.WriteString("[]\n")
		return
	}
	for i, item := range items {
		if i > 0 {
			out.WriteString(indent)
		}
		out.WriteString("- ")
		writeYAMLNode(out, item, indent+"  ")
	}
}

func isYAMLBlock(node any) bool {
	switch n := node.(type) {
	case map[string]any:
		return len(n) > 0
	case []any:
		return len(n) > 0
	default:
		return false
	}
}

func yamlScalar(v any) string {
	switch s := v.(type) {
	case string:
		return yamlString(s)
	case json.Number:
		return s.String()
	case bool:
		return strconv.FormatBool(s)
	default:
		return "null"
	}
}

func yamlString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case isTerminalActive(r) || !isYAMLText(r):
			b.WriteString(escapeRune(r, yamlBeyondBMP))
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func yamlBeyondBMP(r rune) string { return fmt.Sprintf(`\U%08x`, r) }

// isYAMLText reports whether YAML parsers read r, which no terminal acts on,
// as text in a quoted scalar: U+2028 and U+2029 break lines in YAML 1.1, and
// U+FFFE and U+FFFF are not printable.
func isYAMLText(r rune) bool {
	return r != '\u2028' && r != '\u2029' && r != '\ufffe' && r != '\uffff'
}

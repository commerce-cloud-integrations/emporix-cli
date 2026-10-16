package output

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"
)

// Sanitize replaces the characters a terminal may act on (isTerminalActive:
// tabs, newlines, escapes, bidirectional overrides) with spaces. Apply it to
// any API-supplied text written to a terminal.
func Sanitize(text string) string {
	return strings.Map(func(r rune) rune {
		if isTerminalActive(r) {
			return ' '
		}
		return r
	}, text)
}

// isTerminalActive reports whether a terminal may act on r or hide text with
// it: control characters (C0, DEL, C1) and format characters such as
// bidirectional overrides and zero-width spaces.
func isTerminalActive(r rune) bool { return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) }

// escapeInvisible escapes, in encoded JSON, the characters that encoding/json
// writes raw although a terminal may act on them or hide text with them: DEL,
// C1 controls and format characters such as bidirectional overrides and
// zero-width spaces. They can only occur inside strings, so the escaped
// document decodes to the same data.
func escapeInvisible(data []byte) []byte {
	var out bytes.Buffer
	for _, r := range string(data) {
		// encoding/json escapes C0 controls itself, and the layout's newlines stay.
		if r < 0x7f || !isTerminalActive(r) {
			out.WriteRune(r)
			continue
		}
		out.WriteString(escapeRune(r, jsonBeyondBMP))
	}
	return out.Bytes()
}

// escapeRune returns r as an escape in a quoted string of JSON or YAML:
// \uXXXX, or for a character beyond U+FFFF, which takes more than four hex
// digits, what beyondBMP returns, the escape of the format.
func escapeRune(r rune, beyondBMP func(rune) string) string {
	if r > 0xffff {
		return beyondBMP(r)
	}
	return fmt.Sprintf(`\u%04x`, r)
}

// jsonBeyondBMP escapes a character beyond U+FFFF as JSON takes it: a pair of
// UTF-16 surrogates, each as \uXXXX.
func jsonBeyondBMP(r rune) string {
	high, low := utf16.EncodeRune(r)
	return fmt.Sprintf(`\u%04x\u%04x`, high, low)
}

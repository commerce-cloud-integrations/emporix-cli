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

func isTerminalActive(r rune) bool { return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) }

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

func escapeRune(r rune, beyondBMP func(rune) string) string {
	if r > 0xffff {
		return beyondBMP(r)
	}
	return fmt.Sprintf(`\u%04x`, r)
}

func jsonBeyondBMP(r rune) string {
	high, low := utf16.EncodeRune(r)
	return fmt.Sprintf(`\u%04x\u%04x`, high, low)
}

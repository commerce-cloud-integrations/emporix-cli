package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLanguageFlag(t *testing.T) {
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"products", "list", "--language", "dutch!"}, message: `invalid language "dutch!"`},
		{args: []string{"products", "list", "--language", "en-a"}, message: `invalid language "en-a"`},
		{args: []string{"products", "list", "--language", "en-123456789"}, message: `invalid language "en-123456789"`},
		// A table without translations takes no --language (where it appears is checked by
		// TestLanguageAndSiteFlagsFollowTheCases), and nothing takes it globally.
		{args: []string{"sites", "list", "--language", "nl"}, message: "unknown flag: --language"},
		{args: []string{"version", "--language", "nl"}, message: "unknown flag: --language"},
	} {
		if code, _, stderr := runCLI(t, nil, tt.args...); code != ExitUsage || !strings.Contains(stderr, tt.message) {
			t.Errorf("%v: exit %d, stderr %q (want %q)", tt.args, code, stderr, tt.message)
		}
	}
}

func TestLanguageAcceptsWellFormedTags(t *testing.T) {
	for _, tag := range []string{"nl", "EN-gb", "zh-Hant-TW", "es-419", "de-1996", "sl-rozaj"} {
		var value language
		if err := value.Set(tag); err != nil || value.String() != strings.ToLower(tag) {
			t.Errorf("Set(%q) = %v, value %q", tag, err, value)
		}
	}
}

func TestLocalized(t *testing.T) {
	tests := []struct {
		json, language, want string
	}{
		{json: `"Plain"`, language: "nl", want: "Plain"},
		{json: `{"nl":"Klep","en":"Valve"}`, language: "nl", want: "Klep"},
		{json: `{"nl":"Klep","fr":"Clapet"}`, language: "en", want: "Clapet (fr)"},
		{json: `{"en":null,"nl":"Klep"}`, language: "en", want: "Klep (nl)"},
		{json: `{"de":"","fr":null,"nl":"Klep"}`, language: "de", want: "Klep (nl)"},
		{json: `""`, language: "en", want: ""},
		{json: `{}`, language: "en", want: ""},
		{json: `null`, language: "en", want: ""},
		{json: `{"de":"Ventil","en-US":"Valve"}`, language: "en-us", want: "Valve"},
		{json: `{"EN":"Upper","en":"Lower"}`, language: "en", want: "Upper"},
	}
	for _, tt := range tests {
		var text localized
		if err := json.Unmarshal([]byte(tt.json), &text); err != nil {
			t.Fatalf("%s: %v", tt.json, err)
		}
		if got := text.in(language(tt.language)); got != tt.want {
			t.Errorf("%s in %s = %q, want %q", tt.json, tt.language, got, tt.want)
		}
	}
	var text localized
	if err := json.Unmarshal([]byte(`{"nl":1}`), &text); err == nil {
		t.Fatal("non-string translations accepted")
	}
}

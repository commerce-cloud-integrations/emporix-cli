package cli

import (
	"slices"
	"strings"
	"testing"
)

// TestLocalizedWritesSayHowTheySendMaps checks that the help of a write
// whose body holds localized text holds the help of its languageMode and of
// no other, with what the write adds to it, if anything, and that it takes
// --content-language, with its usage, only where its mode lets it name one
// language. What they send, the write cases check.
func TestLocalizedWritesSayHowTheySendMaps(t *testing.T) {
	const (
		stringsOrMapsUsage  = "or * for maps of translations (default *)"
		headerlessMapsUsage = "for maps of translations, sent without Content-Language"
	)
	modes := []string{contentLanguageHelp, headerlessMapsHelp, mapsOnlyHelp, feeLanguageHelp}
	for _, tt := range []struct {
		args  []string
		mode  string // the help of its languageMode, one of modes
		extra string // what its help adds to that of its mode; nothing when empty
		usage string // the usage of its --content-language; no such flag when empty
	}{
		{args: []string{"shipping", "zones", "create"}, mode: headerlessMapsHelp, usage: headerlessMapsUsage},
		{args: []string{"shipping", "zones", "methods", "patch"}, mode: headerlessMapsHelp, usage: headerlessMapsUsage},
		{args: []string{"shipping", "groups", "update"}, mode: headerlessMapsHelp, usage: headerlessMapsUsage},
		{args: []string{"currencies", "create"}, mode: contentLanguageHelp, usage: stringsOrMapsUsage},
		{args: []string{"units", "update"}, mode: contentLanguageHelp, usage: stringsOrMapsUsage,
			extra: "\nSuch a string replaces the unit's other translations of that text; send the map to keep them."},
		{args: []string{"iam", "access-controls", "update"}, mode: mapsOnlyHelp},
		{args: []string{"iam", "groups", "create"}, mode: mapsOnlyHelp},
		{args: []string{"iam", "groups", "update"}, mode: mapsOnlyHelp},
		{args: []string{"brands", "create"}, mode: mapsOnlyHelp},
		{args: []string{"brands", "update"}, mode: mapsOnlyHelp},
		{args: []string{"brands", "patch"}, mode: mapsOnlyHelp},
		{args: []string{"quotes", "reasons", "create"}, mode: mapsOnlyHelp},
		{args: []string{"quotes", "reasons", "update"}, mode: mapsOnlyHelp},
		{args: []string{"schemas", "create"}, mode: mapsOnlyHelp},
		{args: []string{"schemas", "update"}, mode: mapsOnlyHelp},
		{args: []string{"schemas", "references", "create"}, mode: mapsOnlyHelp},
		{args: []string{"schemas", "references", "update"}, mode: mapsOnlyHelp},
		{args: []string{"search", "indexes", "update"}, mode: mapsOnlyHelp},
		{args: []string{"search", "indexes", "import"}, mode: mapsOnlyHelp},
		{args: []string{"search", "queries", "update"}, mode: mapsOnlyHelp},
		{args: []string{"fees", "create"}, mode: feeLanguageHelp},
		{args: []string{"fees", "update"}, mode: feeLanguageHelp},
	} {
		_, help, _ := runCLI(t, nil, append(tt.args, "--help")...)
		long, flags, _ := strings.Cut(help, "\n\nUsage:")
		held := slices.DeleteFunc(slices.Clone(modes), func(mode string) bool { return !strings.Contains(long, mode) })
		if !slices.Equal(held, []string{tt.mode}) || !strings.Contains(long, tt.mode+tt.extra) {
			t.Errorf("%q: help %q; want only the language help %q, followed by %q", tt.args, long, tt.mode, tt.extra)
		}
		usage, given := flagUsage(flags, "--"+contentLanguageFlag)
		if given != (tt.usage != "") || !strings.Contains(usage, tt.usage) {
			t.Errorf("%q: flags %q; want --%s with %q", tt.args, flags, contentLanguageFlag, tt.usage)
		}
	}
}

// flagUsage returns the rest of the line of flags, as the help of a command
// prints them, that names flag, and whether one does.
func flagUsage(flags, flag string) (string, bool) {
	for line := range strings.Lines(flags) {
		if _, usage, found := strings.Cut(line, flag+" "); found {
			return usage, true
		}
	}
	return "", false
}

// TestMapsOnlyWritesRefuseALanguage checks that the brand and schema
// writes, whose services refuse a string for localized text, and the writes
// of fees, whose service ignores its language, take no --content-language.
func TestMapsOnlyWritesRefuseALanguage(t *testing.T) {
	const want = "unknown flag: --" + contentLanguageFlag
	var cases []refusedWrite
	for _, args := range [][]string{
		{"brands", "create"}, {"brands", "update", "b1"}, {"brands", "patch", "b1"},
		{"schemas", "create"}, {"schemas", "update", "s1"},
		{"schemas", "references", "create", "--json-schema", "testdata/schema.json"},
		{"schemas", "references", "update", "r1", "--json-schema", "testdata/schema.json"},
		{"fees", "create"}, {"fees", "update", "f1"},
		{"quotes", "reasons", "create"}, {"quotes", "reasons", "update", "r1"},
		{"search", "indexes", "update", "vehicle", "vehicles"}, {"search", "indexes", "import"},
		{"search", "queries", "update", "vehicle", "by-make"},
	} {
		cases = append(cases, refusedWrite{args: append(args, "--"+contentLanguageFlag, "en"), file: `{"name":"Acme"}`,
			want: want})
	}
	assertRefused(t, cases)
}

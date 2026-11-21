package cli

import (
	"slices"
	"strings"
	"testing"
)

func TestLocalizedWritesSayHowTheySendMaps(t *testing.T) {
	const (
		stringsOrMapsUsage  = "or * for maps of translations (default *)"
		headerlessMapsUsage = "for maps of translations, sent without Content-Language"
	)
	modes := []string{contentLanguageHelp(), headerlessMapsHelp(), mapsOnlyHelp(), feeLanguageHelp()}
	for _, tt := range []struct {
		args  []string
		mode  string
		extra string
		usage string
	}{
		{args: []string{"shipping", "zones", "create"}, mode: headerlessMapsHelp(), usage: headerlessMapsUsage},
		{args: []string{"shipping", "zones", "methods", "patch"}, mode: headerlessMapsHelp(), usage: headerlessMapsUsage},
		{args: []string{"shipping", "groups", "update"}, mode: headerlessMapsHelp(), usage: headerlessMapsUsage},
		{args: []string{"currencies", "create"}, mode: contentLanguageHelp(), usage: stringsOrMapsUsage},
		{args: []string{"units", "update"}, mode: contentLanguageHelp(), usage: stringsOrMapsUsage,
			extra: " Such a string replaces the unit's other translations of that text; send the map to keep them."},
		{args: []string{"iam", "access-controls", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"iam", "groups", "create"}, mode: mapsOnlyHelp()},
		{args: []string{"iam", "groups", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"brands", "create"}, mode: mapsOnlyHelp()},
		{args: []string{"brands", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"brands", "patch"}, mode: mapsOnlyHelp()},
		{args: []string{"quotes", "reasons", "create"}, mode: mapsOnlyHelp()},
		{args: []string{"quotes", "reasons", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"schemas", "create"}, mode: mapsOnlyHelp()},
		{args: []string{"schemas", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"schemas", "references", "create"}, mode: mapsOnlyHelp()},
		{args: []string{"schemas", "references", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"search", "indexes", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"search", "indexes", "import"}, mode: mapsOnlyHelp()},
		{args: []string{"search", "queries", "update"}, mode: mapsOnlyHelp()},
		{args: []string{"fees", "create"}, mode: feeLanguageHelp()},
		{args: []string{"fees", "update"}, mode: feeLanguageHelp()},
	} {
		_, help, _ := runCLI(t, nil, append(tt.args, "--help")...)
		long := strings.Join(helpSections(help)[""], "\n")
		held := slices.DeleteFunc(slices.Clone(modes), func(mode string) bool { return !containsFlat(long, mode) })
		if !slices.Equal(held, []string{tt.mode}) || !containsFlat(long, tt.mode+tt.extra) {
			t.Errorf("%q: help %q; want only the language help %q, followed by %q", tt.args, long, tt.mode, tt.extra)
		}
		usage, given := flagUsage(help, "--"+contentLanguageFlag)
		if given != (tt.usage != "") || !strings.Contains(usage, tt.usage) {
			t.Errorf("%q: help %q; want --%s with %q", tt.args, help, contentLanguageFlag, tt.usage)
		}
	}
}

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

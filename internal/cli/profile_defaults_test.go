package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
)

func TestAProfileGivesTheDefaultSiteAndLanguage(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/prices", allLanguages(firstPage(pricesBody)))
	fake.route("/brand/brands", allLanguages(firstPage(brandsBody)))
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret, "prod": testSecret})
	for _, tt := range []struct {
		args  []string
		query string // of the first request
		want  string
	}{
		{args: []string{"prices", "list"}, query: "siteCode=AcmeStore_NL", want: pricesTable},
		{args: []string{"prices", "list", "--site", "main"}, query: "siteCode=main", want: pricesTable},
		{args: []string{"--profile", "prod", "prices", "list"}, query: "siteCode=main", want: pricesTable},
		{args: []string{"brands", "list"}, want: "ID  NAME\nb1  Merk een\nb2  NORTHWIND TRADE\nb3  Plain only\n"},
		{args: []string{"--profile", "prod", "brands", "list"}, want: "ID  NAME\nb1  Brand one\nb2  NORTHWIND TRADE (nl)\n" +
			"b3  Plain only\n"},
		{args: []string{"brands", "list", "--language", "en"}, want: "ID  NAME\nb1  Brand one\nb2  NORTHWIND TRADE (nl)\n" +
			"b3  Plain only\n"},
	} {
		fake.gets = nil
		code, stdout, stderr := runIn(t, setup.env(nil, "", false), tt.args...)
		if code != ExitOK || stdout != tt.want || !strings.Contains(fake.gets[0], tt.query) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, requests %q", tt.args, code, stdout, stderr, fake.gets)
		}
	}
}

// TestAProfileSiteAppliesOnlyToItsKey checks that a profile's default site
// applies only where the profile gives the API key and the tenant is its
// own: never with the environment's key, nor on another tenant.
func TestAProfileSiteAppliesOnlyToItsKey(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	file := writeBody(t, `[{"stockLevel":5}]`)
	for _, tt := range []struct {
		name string
		vars map[string]string
		args []string
		want string // on standard output, else the error
	}{
		{name: "the profile's key", want: "POST " + fake.server.URL + "/availability/acme1/availability/P1/AcmeStore_NL\n"},
		{name: "the environment's key", vars: fake.env(), want: `required flag(s) "site" not set`},
		{name: "another tenant", args: []string{"--tenant", "acme2"}, want: `required flag(s) "site" not set`},
	} {
		args := slices.Concat(tt.args, []string{"availability", "create", "P1", "--file", file, "--dry-run"})
		code, stdout, stderr := runIn(t, setup.env(tt.vars, "", false), args...)
		if !strings.Contains(stdout+stderr, tt.want) || (code == ExitOK) != strings.HasPrefix(tt.want, "POST") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %q", tt.name, code, stdout, stderr, tt.want)
		}
	}
	fake.assertNothingSent(t)
}

// TestAProfileSiteGivesAnOptionalSiteOfSettings checks that the optional
// --site of reward points create and add, which selects the site whose
// settings apply, takes a profile's default site, and that without either
// no siteCode is sent (Emporix then applies those of site main).
func TestAProfileSiteGivesAnOptionalSiteOfSettings(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	addPoints := fake.server.URL + "/reward-points/customer/C1/addPoints"
	for _, tt := range []struct {
		vars map[string]string
		args []string
		want string // the request line of the dry run
	}{
		{want: "POST " + addPoints + "?siteCode=AcmeStore_NL\n"},
		{args: []string{"--site", "main"}, want: "POST " + addPoints + "?siteCode=main\n"},
		{vars: fake.env(), want: "POST " + addPoints + "\n"},
	} {
		args := slices.Concat([]string{"reward-points", "add", "C1", "--points", "5", "--dry-run"}, tt.args)
		code, stdout, stderr := runIn(t, setup.env(tt.vars, "", false), args...)
		if code != ExitOK || !strings.HasPrefix(stdout, tt.want) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want %q", tt.args, code, stdout, stderr, tt.want)
		}
	}
	fake.assertNothingSent(t)
}

func TestAProfileSiteLeavesABodyGivenWhole(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/customer-segment/acme1/segments/match", respond(http.StatusOK, `[{"id":"100200"}]`))
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	whole := `{"items":[{"id":"100200"}],"segmentIds":["S1"],"siteCode":"main"}`
	for _, tt := range []struct {
		args []string
		sent string
	}{
		{args: []string{"--body", "-"}, sent: whole},
		{args: []string{"--product", "100200", "--segment", "S1"},
			sent: `{"items":[{"id":"100200"}],"segmentIds":["S1"],"siteCode":"AcmeStore_NL"}`},
	} {
		fake.bodies = nil
		args := append([]string{"segments", "match", "--allow-write-scope"}, tt.args...)
		code, stdout, stderr := runIn(t, setup.env(nil, whole, false), args...)
		if code != ExitOK || stdout != "PRODUCT\n100200\n" || len(fake.bodies) != 1 || fake.bodies[0] != tt.sent {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, sent %q", tt.args, code, stdout, stderr, fake.bodies)
		}
	}
}

// TestAProfileSiteLeavesAnOptionalSiteUnset checks that a search whose site
// is an optional filter (all sites without it) keeps searching every site.
func TestAProfileSiteLeavesAnOptionalSiteUnset(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/prices/search", respond(http.StatusOK, `[{"itemId":"100200","prices":[]}]`))
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	code, _, stderr := runIn(t, setup.env(nil, "", false), "prices", "search", "--product", "100200")
	if code != ExitOK || len(fake.bodies) != 1 || fake.bodies[0] != `{"itemIds":["100200"]}` {
		t.Errorf("exit %d, stderr %q, sent %q; want no siteCode", code, stderr, fake.bodies)
	}
}

// TestAProfileSiteKeepsTheErrorOfAFlagGroup checks that a site filled in
// from a profile leaves the error of another flag group as Cobra reports it.
func TestAProfileSiteKeepsTheErrorOfAFlagGroup(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	code, _, stderr := runIn(t, setup.env(nil, "", false), "shipping", "quote-slot", "--cart-total", "100 EUR",
		"--delivery-time", "D", "--slot", "S", "--country", "DE")
	if code != ExitUsage || !strings.Contains(stderr, "[country postal-code]") || strings.Contains(stderr, `"site"`) {
		t.Errorf("exit %d, stderr %q; want the error of --country without --postal-code", code, stderr)
	}
	fake.assertNothingSent(t)
}

func TestAnInvalidProfileDefaultIsNotQuoted(t *testing.T) {
	for _, tt := range []struct {
		setting string
		profile config.Profile
		args    []string
	}{
		{setting: siteFlag, profile: config.Profile{Site: ".."}, args: []string{"prices", "list"}},
		{setting: languageFlag, profile: config.Profile{Language: "NOT-A-LANGUAGE-S3CR3T"}, args: []string{"brands", "list"}},
	} {
		tt.profile.Tenant, tt.profile.ClientID = "acme1", "client"
		setup := newProfileSetup(t, &config.File{Default: "dev", Profiles: map[string]config.Profile{"dev": tt.profile}}, nil)
		code, stdout, stderr := runIn(t, setup.env(nil, "", false), tt.args...)
		want := "profile dev has an invalid default " + tt.setting
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, want) || strings.Contains(stderr, "S3CR3T") ||
			strings.Contains(stderr, `".."`) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %q", tt.setting, code, stdout, stderr, want)
		}
	}
}

// TestOnlyTheRootHasAPersistentHook checks that the root's persistent hook,
// which fills in the defaults of profiles, runs for every command: Cobra
// runs only the nearest one.
func TestOnlyTheRootHasAPersistentHook(t *testing.T) {
	root := newTestRoot()
	walkCommands(root, func(cmd *cobra.Command) {
		if cmd != root && (cmd.PersistentPreRunE != nil || cmd.PersistentPreRun != nil) {
			t.Errorf("%s has a persistent hook of its own", cmd.CommandPath())
		}
	})
}

// TestFlagsTakeProfileDefaultsByDeclaration checks which flags take a
// profile's default: a site that a command needs or whose settings apply,
// and the language of its table, never an optional filter or the settings of
// a profile itself.
func TestFlagsTakeProfileDefaultsByDeclaration(t *testing.T) {
	root := newTestRoot()
	for _, tt := range []struct {
		command, flag, setting string
	}{
		{command: "prices list", flag: siteFlag, setting: siteFlag},
		{command: "brands list", flag: languageFlag, setting: languageFlag},
		{command: "segments match", flag: siteFlag, setting: siteFlag},
		{command: "prices match", flag: siteFlag, setting: siteFlag},
		{command: "fees assignments search-by-products", flag: siteFlag, setting: siteFlag},
		{command: "reward-points create", flag: siteFlag, setting: siteFlag},
		{command: "reward-points add", flag: siteFlag, setting: siteFlag},
		{command: "prices search", flag: siteFlag},
		{command: "profile add", flag: siteFlag},
		{command: "profile add", flag: languageFlag},
	} {
		cmd, _, err := root.Find(strings.Fields(tt.command))
		if err != nil || cmd.Flags().Lookup(tt.flag) == nil {
			t.Fatalf("%s --%s: %v", tt.command, tt.flag, err)
		}
		var setting string
		if takes := cmd.Flags().Lookup(tt.flag).Annotations[profileDefault]; len(takes) > 0 {
			setting = takes[0]
		}
		if setting != tt.setting {
			t.Errorf("%s --%s takes the default %q; want %q", tt.command, tt.flag, setting, tt.setting)
		}
	}
}

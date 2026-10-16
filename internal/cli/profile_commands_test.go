package cli

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
)

// devProfile is the profile that the add cases add as dev.
func devProfile() config.Profile {
	return config.Profile{Tenant: "acme1", ClientID: "client", Site: "AcmeStore_NL", Language: "nl"}
}

// addDev is the command line that adds devProfile.
func addDev() []string {
	return []string{"profile", "add", "dev", "--tenant", "acme1", "--client-id", "client", "--site", "AcmeStore_NL",
		"--language", "NL"}
}

func TestProfileAddStoresTheSecretInTheKeychainOnly(t *testing.T) {
	setup := newProfileSetup(t, nil, nil)
	code, stdout, stderr := runIn(t, setup.env(nil, " "+testSecret+"\r\nnext line\n", false), addDev()...)
	if code != ExitOK || stdout != "" ||
		stderr != "Added profile dev (the default); its client secret is in the OS keychain.\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	want := config.File{Default: "dev", Profiles: map[string]config.Profile{"dev": devProfile()}}
	if got := setup.load(t); !reflect.DeepEqual(got, want) {
		t.Errorf("saved %+v; want %+v", got, want)
	}
	if !maps.Equal(setup.keychain.secrets, map[string]string{"dev": testSecret}) {
		t.Errorf("keychain %q; want the secret of dev", setup.keychain.secrets)
	}
	saved, err := os.ReadFile(setup.store().Path())
	if err != nil || strings.Contains(string(saved), testSecret) {
		t.Errorf("configuration file %q, %v; want no secret in it", saved, err)
	}
	// A second profile is not the default; one added again is replaced.
	code, _, stderr = runIn(t, setup.env(nil, "other-secret\n", false),
		"profile", "add", "prod", "--tenant", "acme2", "--client-id", "c2", "--api-url", "https://api.example.com")
	if code != ExitOK || stderr != "Added profile prod; its client secret is in the OS keychain.\n" {
		t.Fatalf("second add: exit %d, stderr %q", code, stderr)
	}
	code, _, stderr = runIn(t, setup.env(nil, "new-secret", false), "profile", "add", "prod", "--tenant", "acme2",
		"--client-id", "c3")
	if code != ExitOK || stderr != "Replaced profile prod; its client secret is in the OS keychain.\n" {
		t.Fatalf("replace: exit %d, stderr %q", code, stderr)
	}
	got := setup.load(t)
	if got.Default != "dev" || got.Profiles["prod"] != (config.Profile{Tenant: "acme2", ClientID: "c3"}) ||
		setup.keychain.secrets["prod"] != "new-secret" {
		t.Errorf("after replace %+v, keychain %q", got, setup.keychain.secrets)
	}
}

func TestProfileAddAsksForTheSecretOnATerminal(t *testing.T) {
	setup := newProfileSetup(t, nil, nil)
	env := setup.env(nil, "", true)
	env.Profiles.ReadHidden = func(context.Context) (string, error) { return testSecret, nil }
	code, stdout, stderr := runIn(t, env, addDev()...)
	if code != ExitOK || stdout != "" || !strings.HasPrefix(stderr, "Client secret of profile dev: \nAdded profile dev") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	assertNoSecret(t, stdout, stderr)
	if setup.keychain.secrets["dev"] != testSecret {
		t.Errorf("keychain %q", setup.keychain.secrets)
	}
	env.Profiles.ReadHidden = func(context.Context) (string, error) { return "", context.Canceled }
	if code, _, _ := runIn(t, env, "profile", "add", "other", "--tenant", "acme1", "--client-id", "c"); code != ExitError {
		t.Errorf("interrupted: exit %d; want %d", code, ExitError)
	}
	if _, added := setup.load(t).Profiles["other"]; added {
		t.Error("an interrupted add saved the profile")
	}
}

func TestProfileAddFailures(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		stdin   string
		fail    error
		code    int
		message string
	}{
		{name: "empty secret", args: addDev(), stdin: "  \n", code: ExitUsage, message: "no client secret given"},
		{name: "no tenant", args: []string{"profile", "add", "dev", "--client-id", "c"}, stdin: "s", code: ExitUsage,
			message: `required flag(s) "tenant" not set`},
		{name: "invalid tenant", args: []string{"profile", "add", "dev", "--tenant", "a/b", "--client-id", "c"},
			stdin: "s", code: ExitUsage, message: "invalid tenant: use letters"},
		{name: "API URL without https", args: append(addDev(), "--api-url", "http://api.example.com"), stdin: "s",
			code: ExitUsage, message: "must use https"},
		{name: "API URL with a path", args: append(addDev(), "--api-url", "https://api.example.com/x"), stdin: "s",
			code: ExitUsage, message: "API URL must be a scheme and host only"},
		{name: "invalid name", args: []string{"profile", "add", "-x", "--tenant", "acme1"}, code: ExitUsage,
			message: "unknown shorthand flag"},
		{name: "invalid name argument", args: []string{"profile", "add", "a b", "--tenant", "acme1", "--client-id", "c"},
			code: ExitUsage, message: `invalid NAME "a b": a profile name is`},
		{name: "name in capitals", args: []string{"profile", "add", "Dev", "--tenant", "acme1", "--client-id", "c"},
			code: ExitUsage, message: `invalid NAME "Dev": a profile name is a lowercase letter`},
		{name: "invalid site", args: append(addDev(), "--site", ".."), code: ExitUsage, message: `invalid site ".."`},
		{name: "keychain failure", args: addDev(), stdin: "s", fail: errors.New("no keychain"), code: ExitError,
			message: "storing the client secret in the OS keychain: no keychain"},
	} {
		setup := newProfileSetup(t, nil, nil)
		setup.keychain.fail = map[string]error{"SetSecret": tt.fail}
		code, stdout, stderr := runIn(t, setup.env(nil, tt.stdin, false), tt.args...)
		if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.message) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %d and %q", tt.name, code, stdout, stderr, tt.code, tt.message)
		}
		if _, err := os.Stat(setup.store().Path()); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s: a configuration file was written: %v", tt.name, err)
		}
	}
}

// TestProfileAddKeepsTheKeychainAsItWasWhenTheProfileCannotBeSaved checks
// that a profile that cannot be saved leaves the keychain as it was: without
// the secret of a new profile, with the previous one of a replaced profile,
// and that a keychain that cannot be restored is reported.
func TestProfileAddKeepsTheKeychainAsItWasWhenTheProfileCannotBeSaved(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory takes new files on Windows")
	}
	for _, tt := range []struct {
		name    string
		profile string
		fail    map[string]error
		want    map[string]string
		message string
	}{
		{name: "new", profile: "ci", want: map[string]string{"dev": "old"}},
		{name: "replaced", profile: "dev", want: map[string]string{"dev": "old"}},
		{name: "not restored", profile: "ci", fail: map[string]error{"DeleteSecret": errors.New("locked")},
			want: map[string]string{"dev": "old", "ci": "new"}, message: "restoring the client secret of profile ci " +
				"in the OS keychain (service epx, account ci): locked"},
	} {
		setup := newProfileSetup(t, &config.File{Default: "dev", Profiles: map[string]config.Profile{"dev": devProfile()}},
			map[string]string{"dev": "old"})
		readOnly(t, setup)
		setup.keychain.fail = tt.fail
		code, _, stderr := runIn(t, setup.env(nil, "new", false), "profile", "add", tt.profile, "--tenant", "acme2",
			"--client-id", "c2")
		if code != ExitError || !strings.Contains(stderr, "saving the profile") || !strings.Contains(stderr, tt.message) ||
			!maps.Equal(setup.keychain.secrets, tt.want) || setup.load(t).Profiles["dev"] != devProfile() {
			t.Errorf("%s: exit %d, stderr %q, keychain %q; want %q", tt.name, code, stderr, setup.keychain.secrets, tt.want)
		}
	}
}

// readOnly makes the configuration directory of setup, which holds its
// configuration file and lock, one where no file can be added.
func readOnly(t *testing.T, setup *profileSetup) {
	t.Helper()
	unlock, err := setup.store().Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	dir := filepath.Dir(setup.store().Path())
	//nolint:gosec // G302 checks the mode of a file; a directory needs its x bit to be searched, and 0500 keeps it.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // G302: as above, the directory's own mode.
}

func TestProfileListAndShowNeverPrintTheSecret(t *testing.T) {
	setup := newProfileSetup(t, &config.File{Default: "dev", Profiles: map[string]config.Profile{
		"dev": {
			Tenant: "acme1", APIURL: "https://api.example.com", ClientID: "client", Site: "AcmeStore_NL", Language: "nl",
		},
		"prod": {Tenant: "acme2", ClientID: "c2", Site: "main"},
	}}, map[string]string{"dev": testSecret})
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"profile", "list"}, want: "" +
			"NAME  DEFAULT  TENANT  CLIENT ID  API URL                  SITE          LANGUAGE\n" +
			"dev   yes      acme1   client     https://api.example.com  AcmeStore_NL  nl\n" +
			"prod           acme2   c2                                  main          \n"},
		{args: []string{"profile", "list", "-o", "json"}, want: `[
  {
    "apiUrl": "https://api.example.com",
    "clientId": "client",
    "default": true,
    "language": "nl",
    "name": "dev",
    "site": "AcmeStore_NL",
    "tenant": "acme1"
  },
  {
    "clientId": "c2",
    "default": false,
    "name": "prod",
    "site": "main",
    "tenant": "acme2"
  }
]
`},
		{args: []string{"profile", "show", "prod"}, want: "" +
			"NAME  DEFAULT  TENANT  CLIENT ID  API URL  SITE  LANGUAGE  CLIENT SECRET\n" +
			"prod           acme2   c2                  main            not stored\n"},
		{args: []string{"profile", "show", "dev", "-o", "json"}, want: `{
  "apiUrl": "https://api.example.com",
  "clientId": "client",
  "default": true,
  "language": "nl",
  "name": "dev",
  "secretStored": true,
  "site": "AcmeStore_NL",
  "tenant": "acme1"
}
`},
	} {
		code, stdout, stderr := runIn(t, setup.env(nil, "", false), tt.args...)
		if code != ExitOK || stdout != tt.want || stderr != "" {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want %q", tt.args, code, stdout, stderr, tt.want)
		}
		assertNoSecret(t, stdout, stderr)
	}
	// Without a configuration, there are no profiles.
	empty := newProfileSetup(t, nil, nil).env(nil, "", false)
	if code, stdout, _ := runIn(t, empty, "profile", "list", "-o", "json"); code != ExitOK || stdout != "[]\n" {
		t.Errorf("no profiles: exit %d, stdout %q", code, stdout)
	}
}

func TestProfileShowFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), nil)
	if code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "show", "nope"); code != ExitUsage ||
		!strings.Contains(stderr, `no profile "nope"; `+"`epx profile list` lists them") {
		t.Errorf("unknown: exit %d, stderr %q", code, stderr)
	}
	setup.keychain.fail = map[string]error{"Secret": errors.New("locked")}
	if code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "show", "dev"); code != ExitError ||
		!strings.Contains(stderr, "OS keychain: locked") {
		t.Errorf("keychain failure: exit %d, stderr %q", code, stderr)
	}
	writeConfiguration(t, setup, "{")
	for _, args := range [][]string{{"profile", "show", "dev"}, {"profile", "list"}, {"profile", "use", "dev"}} {
		if code, _, stderr := runIn(t, setup.env(nil, "", false), args...); code != ExitUsage ||
			!strings.Contains(stderr, "is not an epx configuration") {
			t.Errorf("%q of an invalid file: exit %d, stderr %q", args, code, stderr)
		}
	}
}

package cli

import (
	"cmp"
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
)

// fakeKeychain holds client secrets in memory, never in a real keychain.
type fakeKeychain struct {
	mu      sync.Mutex
	secrets map[string]string
	fail    map[string]error // by call (Secret, SetSecret or DeleteSecret); under "*", every call fails
	block   chan struct{}    // where set, every call waits until it is closed
	waiting chan struct{}    // where set, a call blocked by block first tells it here
}

// failure returns the error of call, if it fails, once the keychain stops
// blocking; it holds the lock until done is called.
func (k *fakeKeychain) failure(call string) (done func(), err error) {
	if k.block != nil {
		select {
		case k.waiting <- struct{}{}:
		case <-k.block:
		}
		<-k.block
	}
	k.mu.Lock()
	return k.mu.Unlock, cmp.Or(k.fail[call], k.fail["*"])
}

func (k *fakeKeychain) Secret(profile string) (string, error) {
	done, err := k.failure("Secret")
	defer done()
	if err != nil {
		return "", err
	}
	secret, ok := k.secrets[profile]
	if !ok {
		return "", config.ErrNoSecret
	}
	return secret, nil
}

func (k *fakeKeychain) SetSecret(profile, secret string) error {
	done, err := k.failure("SetSecret")
	defer done()
	if err == nil {
		k.secrets[profile] = secret
	}
	return err
}

func (k *fakeKeychain) DeleteSecret(profile string) error {
	done, err := k.failure("DeleteSecret")
	defer done()
	if err != nil {
		return err
	}
	if _, ok := k.secrets[profile]; !ok {
		return config.ErrNoSecret
	}
	delete(k.secrets, profile)
	return nil
}

// noProfiles is the ProfileEnv of a run without profiles: no configuration
// directory, and a keychain that holds no secret.
func noProfiles() *ProfileEnv {
	return &ProfileEnv{
		ConfigDir:  func() (string, error) { return "", errors.New("$HOME is not defined") },
		Keychain:   &fakeKeychain{secrets: map[string]string{}},
		ReadHidden: func(context.Context) (string, error) { return "", errors.New("no terminal") },
	}
}

// profileSetup is the configuration directory, a temporary one, and the
// fake keychain of a test.
type profileSetup struct {
	dir      string
	keychain *fakeKeychain
}

// newProfileSetup saves file, unless nil, in a new configuration directory
// whose keychain holds secrets.
func newProfileSetup(t *testing.T, file *config.File, secrets map[string]string) *profileSetup {
	t.Helper()
	s := &profileSetup{dir: t.TempDir(), keychain: &fakeKeychain{secrets: map[string]string{}}}
	for profile, secret := range secrets {
		s.keychain.secrets[profile] = secret
	}
	if file != nil {
		if err := s.store().Save(*file); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func (s *profileSetup) store() config.Store { return config.In(s.dir) }

// env returns the environment of a run with vars, stdin as its standard input
// (a terminal or not), the setup's configuration directory and keychain.
func (s *profileSetup) env(vars map[string]string, stdin string, terminal bool) Env {
	profiles := noProfiles()
	profiles.ConfigDir, profiles.Keychain = func() (string, error) { return s.dir, nil }, s.keychain
	return Env{
		In: strings.NewReader(stdin), Getenv: func(key string) string { return vars[key] }, Terminal: terminal,
		Profiles: profiles,
	}
}

// load returns the saved configuration.
func (s *profileSetup) load(t *testing.T) config.File {
	t.Helper()
	file, err := s.store().Load()
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// twoProfiles returns a configuration whose default profile dev, and prod,
// address fake (API URL), each with its own default site.
func twoProfiles(fake *fakeEmporix) *config.File {
	return &config.File{Default: "dev", Profiles: map[string]config.Profile{
		"dev":  {Tenant: "acme1", APIURL: fake.server.URL, ClientID: "client", Site: "AcmeStore_NL", Language: "nl"},
		"prod": {Tenant: "acme1", APIURL: fake.server.URL, ClientID: "client", Site: "main"},
	}}
}

func TestAProfileGivesTheCredentials(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret, "prod": testSecret})
	for _, tt := range []struct {
		name string
		vars map[string]string
		args []string
	}{
		{name: "default profile"},
		{name: "EPX_PROFILE", vars: map[string]string{envProfile: "prod"}},
		{name: "--profile over EPX_PROFILE", vars: map[string]string{envProfile: "nope"}, args: []string{"--profile", "dev"}},
		{name: "--tenant over the profile", args: []string{"--tenant", "acme1"}},
	} {
		code, stdout, stderr := runIn(t, setup.env(tt.vars, "", false), append(tt.args, "auth", "status")...)
		if code != ExitOK || !strings.Contains(stdout, "acme1") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", tt.name, code, stdout, stderr)
		}
		assertNoSecret(t, stdout, stderr)
	}
	if len(fake.tokenScopes) != 4 {
		t.Errorf("%d token requests at the profiles' API URL; want 4", len(fake.tokenScopes))
	}
}

// TestTheEnvironmentOrAProfileGivesTheWholeKey checks that the settings of an
// API key come from the environment, where it sets an EMPORIX_ variable, or
// from a profile, never from both: no client secret goes to an API URL or
// with a client ID of the other, and naming a profile while the environment
// gives the key is refused.
func TestTheEnvironmentOrAProfileGivesTheWholeKey(t *testing.T) {
	envFake, profileFake := newFakeEmporix(t), newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(profileFake), map[string]string{"dev": testSecret})
	only := func(name, value string) map[string]string { return map[string]string{name: value} }
	envWith := func(name, value string) map[string]string {
		vars := envFake.env()
		vars[name] = value
		return vars
	}
	for _, tt := range []struct {
		name    string
		vars    map[string]string
		args    []string
		code    int
		message string
		tokens  int // token requests at the environment's API URL
	}{
		{name: "environment", vars: envFake.env(), code: ExitOK, tokens: 1},
		{name: "--profile", vars: envFake.env(), args: []string{"--profile", "dev"}, code: ExitUsage,
			message: "--profile and EMPORIX_TENANT are both set: a profile applies only where no EMPORIX_ variable is set"},
		{name: "EPX_PROFILE", vars: envWith(envProfile, "dev"), code: ExitUsage,
			message: "EPX_PROFILE and EMPORIX_TENANT are both set"},
		{name: "client secret only", vars: only(envClientSecret, testSecret), code: ExitUsage,
			message: "missing configuration: set EMPORIX_TENANT (or --tenant), EMPORIX_CLIENT_ID\n"},
		{name: "client ID only", vars: only(envClientID, "client"), code: ExitUsage,
			message: "missing configuration: set EMPORIX_TENANT (or --tenant), EMPORIX_CLIENT_SECRET\n"},
		{name: "API URL only", vars: only(envAPIURL, envFake.server.URL), code: ExitUsage,
			message: "missing configuration: set EMPORIX_TENANT (or --tenant), EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET\n"},
		{name: "tenant only", vars: only(envTenant, "acme1"), code: ExitUsage,
			message: "missing configuration: set EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET\n"},
	} {
		envFake.tokenScopes = nil
		code, stdout, stderr := runIn(t, setup.env(tt.vars, "", false), append(tt.args, "auth", "status")...)
		if code != tt.code || !strings.Contains(stderr, tt.message) || len(envFake.tokenScopes) != tt.tokens {
			t.Errorf("%s: exit %d, stdout %q, stderr %q, %d token requests; want %d, %q and %d", tt.name, code, stdout,
				stderr, len(envFake.tokenScopes), tt.code, tt.message, tt.tokens)
		}
	}
	if len(profileFake.tokenScopes) != 0 {
		t.Errorf("the profile's API URL got %d token requests; want none", len(profileFake.tokenScopes))
	}
}

func TestProfileCredentialFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	broken := errors.New("keychain locked")
	for _, tt := range []struct {
		name    string
		secrets map[string]string
		fail    error
		vars    map[string]string
		args    []string
		code    int
		message string
	}{
		{name: "no secret stored", code: ExitUsage, message: "missing configuration: set the client secret of " +
			"profile dev in the OS keychain (`epx profile add dev`)\n"},
		{name: "empty --tenant", secrets: map[string]string{"dev": testSecret}, args: []string{"--tenant", ""},
			code: ExitUsage, message: "missing configuration: set a non-empty --tenant\n"},
		{name: "keychain failure", fail: broken, code: ExitError,
			message: "reading the client secret of profile dev from the OS keychain: keychain locked"},
		{name: "unknown --profile", args: []string{"--profile", "nope"}, code: ExitUsage,
			message: "--profile names no profile; `epx profile list` lists them"},
		{name: "unknown EPX_PROFILE", vars: map[string]string{envProfile: "nope"}, code: ExitUsage,
			message: "EPX_PROFILE names no profile"},
		{name: "invalid EPX_PROFILE", vars: map[string]string{envProfile: "s3cr3t value"}, code: ExitUsage,
			message: "EPX_PROFILE is not a profile name\n"},
		{name: "invalid --profile", args: []string{"--profile", "a/b"}, code: ExitUsage,
			message: "a profile name is a lowercase letter or digit"},
		{name: "another tenant", secrets: map[string]string{"dev": testSecret}, args: []string{"--tenant", "other"},
			code: ExitAuth, message: "The tenant in the scope does not match"},
	} {
		setup := newProfileSetup(t, twoProfiles(fake), tt.secrets)
		setup.keychain.fail = map[string]error{"*": tt.fail}
		code, stdout, stderr := runIn(t, setup.env(tt.vars, "", false), append(tt.args, "auth", "status")...)
		if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.message) || strings.Contains(stderr, "s3cr3t") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want %d and %q", tt.name, code, stdout, stderr, tt.code, tt.message)
		}
	}
}

func TestProfilesNeedAConfigurationDirectoryOnlyWhenChosen(t *testing.T) {
	code, _, stderr := runCLI(t, nil, "auth", "status")
	if code != ExitUsage || !strings.Contains(stderr, "missing configuration: set EMPORIX_TENANT (or --tenant), ") {
		t.Errorf("without a profile: exit %d, stderr %q", code, stderr)
	}
	for _, args := range [][]string{
		{"--profile", "dev", "auth", "status"}, {"profile", "list"}, {"profile", "use", "dev"},
	} {
		code, _, stderr = runCLI(t, nil, args...)
		if code != ExitUsage || !strings.Contains(stderr, "epx has no configuration directory here: $HOME is not defined") {
			t.Errorf("%q: exit %d, stderr %q", args, code, stderr)
		}
	}
}

func TestAConfigurationEpxRefusesFailsEveryCommandThatReadsIt(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, nil, nil)
	writeConfiguration(t, setup, `{"default":"dev","profiles":{}}`)
	// The environment gives the key, so no profile is read.
	if code, _, stderr := runIn(t, setup.env(fake.env(), "", false), "auth", "status"); code != ExitOK {
		t.Fatalf("with the environment's key: exit %d, stderr %q", code, stderr)
	}
	code, _, stderr := runIn(t, setup.env(nil, "", false), "auth", "status")
	if code != ExitUsage || !strings.Contains(stderr, "names a default profile that it does not hold") {
		t.Fatalf("exit %d, stderr %q; want a refusal", code, stderr)
	}
}

// writeConfiguration writes content as the configuration file of setup.
func writeConfiguration(t *testing.T, setup *profileSetup, content string) {
	t.Helper()
	if err := setup.store().Save(config.File{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(setup.store().Path(), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestProfileNamesComplete(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), nil)
	for _, args := range [][]string{
		{"__complete", "profile", "use", ""},
		{"__complete", "--profile", ""},
	} {
		code, stdout, _ := runIn(t, setup.env(nil, "", false), args...)
		if code != ExitOK || !strings.HasPrefix(stdout, "dev\nprod\n:4\n") {
			t.Errorf("%q: exit %d, stdout %q; want the profile names", args, code, stdout)
		}
	}
	// Without a configuration, nothing is offered.
	code, stdout, _ := runCLI(t, nil, "__complete", "profile", "use", "")
	if code != ExitOK || !strings.HasPrefix(stdout, ":4\n") {
		t.Errorf("exit %d, stdout %q; want no names", code, stdout)
	}
}

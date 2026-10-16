package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// envProfile names the profile to use when --profile is not given.
const envProfile = "EPX_PROFILE"

// profileFlag is the global flag that names the profile to use.
const profileFlag = "profile"

// profileName is a --profile value or a profile argument; it implements
// pflag.Value, so a name that cannot name a profile is rejected while flags
// are parsed.
type profileName string

// Set implements pflag.Value.
func (p *profileName) Set(value string) error {
	if err := config.CheckName(value); err != nil {
		return err
	}
	*p = profileName(value)
	return nil
}

func (p *profileName) String() string { return string(*p) }

// Type implements pflag.Value.
func (p *profileName) Type() string { return "name" }

// apiKey is the API key a command uses. Its settings come from the
// environment where it sets an EMPORIX_ variable (envSettings), else from
// the profile in effect: never from both, so a key's client secret goes only
// to its own API URL, and a profile's default site only to its own tenant.
type apiKey struct {
	profile  string         // the profile that gives the key; "" for the environment
	settings config.Profile // tenant, API URL and client ID; a profile's default site and language
	secret   string         // the environment's; the OS keychain holds a profile's (secretOf)
}

// resolvedKey is the API key of a command, resolved once (app.key).
type resolvedKey struct {
	once sync.Once
	key  apiKey
	err  error
}

// envSettings returns the environment variables that give an API key.
func envSettings() []string { return []string{envTenant, envAPIURL, envClientID, envClientSecret} }

// key returns the API key of the command (resolveKey), which it resolves
// once: the profile defaults of flags and the client read the same.
func (a *app) key() (apiKey, error) {
	a.resolved.once.Do(func() { a.resolved.key, a.resolved.err = a.resolveKey() })
	return a.resolved.key, a.resolved.err
}

// resolveKey returns the API key of the environment where it sets an
// EMPORIX_ variable; else that of the profile that --profile names, else
// EPX_PROFILE, else the default profile; else none. A profile chosen where
// the environment gives the key, a chosen profile that does not exist and a
// configuration file epx refuses are usage errors.
func (a *app) resolveKey() (apiKey, error) {
	name, chosenBy := string(a.profile), "--"+profileFlag
	if name == "" {
		name, chosenBy = a.getenv(envProfile), envProfile
	}
	if set := slices.IndexFunc(envSettings(), func(v string) bool { return a.getenv(v) != "" }); set >= 0 {
		if name != "" {
			return apiKey{}, usageError{fmt.Errorf("%s and %s are both set: a profile applies only where no "+
				"EMPORIX_ variable is set", chosenBy, envSettings()[set])}
		}
		return apiKey{settings: config.Profile{
			Tenant: a.getenv(envTenant), APIURL: a.getenv(envAPIURL), ClientID: a.getenv(envClientID),
		}, secret: a.getenv(envClientSecret)}, nil
	}
	if name != "" && config.CheckName(name) != nil {
		return apiKey{}, usageError{fmt.Errorf("%s is not a profile name", envProfile)}
	}
	file, err := a.loadProfiles()
	switch {
	case errors.Is(err, errNoConfigDir) && name == "":
		return apiKey{}, nil // without a configuration directory, no profile is the default
	case err != nil:
		return apiKey{}, err
	}
	name = cmp.Or(name, file.Default)
	if name == "" {
		return apiKey{}, nil
	}
	profile, err := file.Lookup(name)
	if err != nil {
		return apiKey{}, usageError{fmt.Errorf("%s names no profile; `epx profile list` lists them", chosenBy)}
	}
	return apiKey{profile: name, settings: profile}, nil
}

// profileSnapshot sets the settings and client secret of key, a profile's key,
// read as one state: under the configuration file's lock, which
// editProfiles holds while it replaces a profile's settings and secret. Read
// apart, a replacement in between would pair the old host with the new secret.
func (a *app) profileSnapshot(ctx context.Context, key *apiKey) error {
	store, err := a.store()
	if err != nil {
		return err
	}
	unlock, err := store.Lock(ctx)
	if err != nil {
		return fmt.Errorf("locking the configuration file: %w", err)
	}
	defer unlock()
	file, err := a.loadProfiles()
	if err != nil {
		return err
	}
	profile, err := file.Lookup(key.profile)
	if err != nil {
		return usageError{fmt.Errorf("profile %s no longer exists: `epx profile list` lists them", key.profile)}
	}
	key.settings = profile
	key.secret, err = a.secretOf(ctx, key.profile)
	return err
}

// settings returns the Emporix settings of the command's API key (key),
// with the tenant of --tenant where given and, for a profile, the client
// secret the OS keychain holds. Missing settings are a usage error.
func (a *app) settings(ctx context.Context) (emporix.Config, error) {
	key, err := a.key()
	if err == nil && key.profile != "" {
		err = a.profileSnapshot(ctx, &key)
	}
	if err != nil {
		return emporix.Config{}, err
	}
	cfg := emporix.Config{
		BaseURL:  cmp.Or(key.settings.APIURL, emporix.DefaultBaseURL),
		Tenant:   a.tenant.or(key.settings.Tenant),
		ClientID: key.settings.ClientID, ClientSecret: key.secret,
	}
	if missing := a.missing(key.profile, &cfg); len(missing) > 0 {
		return emporix.Config{}, usageError{fmt.Errorf("missing configuration: set %s", strings.Join(missing, ", "))}
	}
	return cfg, nil
}

// missing returns the settings that cfg lacks, named as the user sets them:
// in the environment, or in the profile that gives the key.
func (a *app) missing(profile string, cfg *emporix.Config) []string {
	tenant, clientID, secret := envTenant+" (or --tenant)", envClientID, envClientSecret
	if profile != "" {
		tenant, clientID = "the tenant of profile "+profile, "the client ID of profile "+profile
		secret = "the client secret of profile " + profile + " in the OS keychain (`epx profile add " + profile + "`)"
	}
	if a.tenant.given {
		tenant = "a non-empty --tenant" // an empty one never falls back to the environment or a profile
	}
	var missing []string
	for _, setting := range [][2]string{{tenant, cfg.Tenant}, {clientID, cfg.ClientID}, {secret, cfg.ClientSecret}} {
		if setting[1] == "" {
			missing = append(missing, setting[0])
		}
	}
	return missing
}

// errNoConfigDir is what a command that needs the configuration file returns
// where the system names no configuration directory.
var errNoConfigDir = errors.New("epx has no configuration directory here")

// store returns the configuration file in the user's configuration
// directory; without one, a usage error.
func (a *app) store() (config.Store, error) {
	dir, err := a.ConfigDir()
	if err != nil {
		return config.Store{}, usageError{fmt.Errorf("%w: %w", errNoConfigDir, err)}
	}
	return config.In(dir), nil
}

// loadProfiles returns what the configuration file holds. Its failures are
// usage errors, as nothing was sent.
func (a *app) loadProfiles() (config.File, error) {
	store, err := a.store()
	if err != nil {
		return config.File{}, err
	}
	file, err := store.Load()
	if err != nil {
		return config.File{}, usageError{err}
	}
	return file, nil
}

// editProfiles calls edit with the configuration file and what it holds
// while it holds the file's lock: another epx that changes profiles waits
// until edit returns, or an interrupt ends the wait.
func (a *app) editProfiles(ctx context.Context, edit func(config.Store, config.File) error) error {
	store, err := a.store()
	if err != nil {
		return err
	}
	unlock, err := store.Lock(ctx)
	if err != nil {
		return fmt.Errorf("locking the configuration file: %w", err)
	}
	defer unlock()
	file, err := store.Load()
	if err != nil {
		return usageError{err}
	}
	return edit(store, file)
}

// waitingForKeychain is what a call of the OS keychain that an interrupt
// ends was waiting for (untilDone): a keychain may wait for a prompt that
// the user cannot reach.
const waitingForKeychain = "waiting for the OS keychain"

// secretOf returns the client secret that the OS keychain holds for profile,
// or "" when it holds none.
func (a *app) secretOf(ctx context.Context, profile string) (string, error) {
	secret, err := untilDone(ctx, waitingForKeychain, func() (string, error) { return a.Keychain.Secret(profile) })
	if errors.Is(err, config.ErrNoSecret) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading the client secret of profile %s from the OS keychain: %w", profile, err)
	}
	return secret, nil
}

// changeKeychain makes change to the OS keychain; an interrupt ends the wait.
func (a *app) changeKeychain(ctx context.Context, change func(config.Keychain) error) error {
	_, err := untilDone(ctx, waitingForKeychain, func() (struct{}, error) { return struct{}{}, change(a.Keychain) })
	return err
}

// applyProfileDefaults gives the flags of cmd that take a profile's default
// (takeProfileDefault) and were not given the default site or language of
// the profile in effect, where its tenant is the command's. It runs before
// Cobra checks the required flags, so a profile's site gives the site that a
// read requires.
func (a *app) applyProfileDefaults(cmd *cobra.Command) error {
	var pending []*pflag.Flag
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		takes := flag.Annotations[profileDefault]
		if len(takes) > 0 && !flag.Changed && !slices.ContainsFunc(takes[1:], cmd.Flags().Changed) {
			pending = append(pending, flag)
		}
	})
	if len(pending) == 0 {
		return nil
	}
	key, err := a.key()
	if err != nil || key.profile == "" || a.tenant.or(key.settings.Tenant) != key.settings.Tenant {
		return err
	}
	defaults := map[string]string{siteFlag: key.settings.Site, languageFlag: key.settings.Language}
	for _, flag := range pending {
		setting := flag.Annotations[profileDefault][0]
		if value := defaults[setting]; value != "" && cmd.Flags().Set(flag.Name, value) != nil {
			// The flag's error would quote the value, which comes from the configuration.
			return usageError{fmt.Errorf("profile %s has an invalid default %s; `epx profile add` replaces the profile",
				key.profile, setting)}
		}
	}
	return nil
}

// profileNames returns the sorted names of the profiles, for completion.
func (a *app) profileNames() []string {
	file, _ := a.loadProfiles() // a configuration that cannot be read has none
	return slices.Sorted(maps.Keys(file.Profiles))
}

// completeProfiles completes the name of a profile, never a file name.
func (a *app) completeProfiles(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return a.profileNames(), cobra.ShellCompDirectiveNoFileComp
}

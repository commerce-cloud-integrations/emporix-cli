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

const envProfile = "EPX_PROFILE"

const profileFlag = "profile"

type profileName string

func (p *profileName) Set(value string) error {
	if err := config.CheckName(value); err != nil {
		return err
	}
	*p = profileName(value)
	return nil
}

func (p *profileName) String() string { return string(*p) }

func (p *profileName) Type() string { return "name" }

type apiKey struct {
	profile  string
	settings config.Profile
	secret   string
}

type resolvedKey struct {
	once sync.Once
	key  apiKey
	err  error
}

func envSettings() []string { return []string{envTenant, envAPIURL, envClientID, envClientSecret} }

func (a *app) key() (apiKey, error) {
	a.resolved.once.Do(func() { a.resolved.key, a.resolved.err = a.resolveKey() })
	return a.resolved.key, a.resolved.err
}

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
		return apiKey{}, nil
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

func (a *app) missing(profile string, cfg *emporix.Config) []string {
	tenant, clientID, secret := envTenant+" (or --tenant)", envClientID, envClientSecret
	if profile != "" {
		tenant, clientID = "the tenant of profile "+profile, "the client ID of profile "+profile
		secret = "the client secret of profile " + profile + " in the OS keychain (`epx profile add " + profile + "`)"
	}
	if a.tenant.given {
		tenant = "a non-empty --tenant"
	}
	var missing []string
	for _, setting := range [][2]string{{tenant, cfg.Tenant}, {clientID, cfg.ClientID}, {secret, cfg.ClientSecret}} {
		if setting[1] == "" {
			missing = append(missing, setting[0])
		}
	}
	return missing
}

var errNoConfigDir = errors.New("epx has no configuration directory here")

func (a *app) store() (config.Store, error) {
	dir, err := a.ConfigDir()
	if err != nil {
		return config.Store{}, usageError{fmt.Errorf("%w: %w", errNoConfigDir, err)}
	}
	return config.In(dir), nil
}

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

const waitingForKeychain = "waiting for the OS keychain"

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

func (a *app) changeKeychain(ctx context.Context, change func(config.Keychain) error) error {
	_, err := untilDone(ctx, waitingForKeychain, func() (struct{}, error) { return struct{}{}, change(a.Keychain) })
	return err
}

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
	for _, flag := range pending {
		if err := key.setFromProfile(cmd.Flags(), flag); err != nil {
			return err
		}
	}
	return nil
}

func (k *apiKey) setFromProfile(flags *pflag.FlagSet, flag *pflag.Flag) error {
	setting := flag.Annotations[profileDefault][0]
	value := map[string]string{siteFlag: k.settings.Site, languageFlag: k.settings.Language}[setting]
	if value == "" || flags.Set(flag.Name, value) == nil {
		return nil
	}
	return usageError{fmt.Errorf("profile %s has an invalid default %s; `epx profile add` replaces the profile",
		k.profile, setting)}
}

func (a *app) profileNames() []string {
	file, _ := a.loadProfiles()
	return slices.Sorted(maps.Keys(file.Profiles))
}

func (a *app) completeProfiles(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return a.profileNames(), cobra.ShellCompDirectiveNoFileComp
}

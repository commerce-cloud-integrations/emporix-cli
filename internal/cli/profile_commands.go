package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

func newProfileCommand(a *app) *cobra.Command {
	group := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles: named settings and credentials of a tenant",
		Long: "A profile keeps the settings of one API key: its tenant, API URL and client ID, and a default\n" +
			"site and language. Profiles are in epx/config.json in your configuration directory (~/.config on\n" +
			"Linux, ~/Library/Application Support on macOS, %AppData% on Windows), which only you may read;\n" +
			"the client secret of each is in the OS keychain (the macOS Keychain, the Windows Credential\n" +
			"Manager or a Secret Service on Linux), never in that file.\n\n" +
			"A command uses the profile that --profile names, else the one EPX_PROFILE names, else the\n" +
			"default (`epx profile use`), but only where no EMPORIX_ variable is set: those give the whole\n" +
			"API key, never mixed with a profile's. --tenant overrides the profile's tenant. These commands\n" +
			"change only the configuration file and the keychain: they send nothing to Emporix.",
	}
	group.AddCommand(newProfileAddCommand(a), newProfileListCommand(a), newProfileShowCommand(a),
		newProfileUseCommand(a), newProfileRemoveCommand(a))
	return group
}

// profileArgs accepts one argument, a profile name.
func profileArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}
	if err := config.CheckName(args[0]); err != nil {
		return fmt.Errorf("invalid NAME %q: %w", args[0], err)
	}
	return nil
}

// noProfile is the usage error of a profile argument that names no profile
// (config.ErrNoProfile); any other error is returned as it is.
func noProfile(err error) error {
	if errors.Is(err, config.ErrNoProfile) {
		return usageError{fmt.Errorf("%w; `epx profile list` lists them", err)}
	}
	return err
}

func newProfileAddCommand(a *app) *cobra.Command {
	var profile config.Profile
	var site siteCode
	var lang language
	cmd := &cobra.Command{
		Use:   "add NAME",
		Short: "Add a profile, or replace one, with its client secret in the OS keychain",
		Long: "Adds the profile NAME, or replaces the profile of that name, and stores its client secret in\n" +
			"the OS keychain. On a terminal, epx asks for the secret without showing it; otherwise it reads\n" +
			"the first line of standard input. It never takes the secret from a flag. Where no profile is\n" +
			"the default, as for the first one, it becomes the default. Nothing is sent to Emporix: `epx\n" +
			"auth status` with --profile NAME checks the API key.",
		Example: "  epx profile add dev --tenant acme1 --client-id CLIENT_ID --site AcmeStore_NL --language nl\n" +
			"  printf '%s\\n' \"$CLIENT_SECRET\" | epx profile add ci --tenant acme1 --client-id CLIENT_ID",
		Args: profileArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile.Site, profile.Language = string(site), string(lang)
			return a.addProfile(cmd, args[0], &profile)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&profile.Tenant, "tenant", "", "Emporix tenant of the API key")
	flags.StringVar(&profile.ClientID, "client-id", "", "client ID of the API key")
	flags.StringVar(&profile.APIURL, "api-url", "", "API host (default "+emporix.DefaultBaseURL+")")
	flags.Var(&site, siteFlag, "default site code, for commands that need a site")
	flags.Var(&lang, languageFlag, "default language of localized table text, for commands that take --language")
	requireFlag(cmd, "tenant")
	requireFlag(cmd, "client-id")
	return cmd
}

// addProfile checks the settings of profile as a client checks them, reads
// its client secret (readSecret), then, under the configuration file's lock,
// stores the secret in the OS keychain and saves the profile as name
// (saveWithSecret).
func (a *app) addProfile(cmd *cobra.Command, name string, profile *config.Profile) error {
	if err := emporix.CheckEndpoint(cmp.Or(profile.APIURL, emporix.DefaultBaseURL), profile.Tenant); err != nil {
		return err
	}
	if _, err := a.store(); err != nil {
		return err // before the secret is asked for
	}
	secret, err := a.readSecret(cmd, name)
	if err != nil {
		return err
	}
	done := "Added profile " + name
	err = a.editProfiles(cmd.Context(), func(store config.Store, file config.File) error {
		if file.Put(name, profile) {
			done = "Replaced profile " + name
		}
		if file.Default == name {
			done += " (the default)"
		}
		return a.saveWithSecret(cmd.Context(), store, file, name, secret)
	})
	if err == nil {
		_, err = fmt.Fprintln(cmd.ErrOrStderr(), done+"; its client secret is in the OS keychain.")
	}
	return err
}

// saveWithSecret stores secret in the OS keychain as the client secret of
// the profile name, then saves file, which holds that profile. Where the
// save fails, the keychain gets back the secret it held before, or none, so
// the file and the keychain still agree; a failure to restore it is
// reported too.
func (a *app) saveWithSecret(ctx context.Context, store config.Store, file config.File, name, secret string) error {
	previous, err := a.secretOf(ctx, name)
	if err != nil {
		return err
	}
	if err = a.changeKeychain(ctx, func(k config.Keychain) error { return k.SetSecret(name, secret) }); err != nil {
		return fmt.Errorf("storing the client secret in the OS keychain: %w", err)
	}
	if err = store.Save(file); err == nil {
		return nil
	}
	err = fmt.Errorf("saving the profile: %w", err)
	restore := func(k config.Keychain) error { return k.DeleteSecret(name) }
	if previous != "" {
		restore = func(k config.Keychain) error { return k.SetSecret(name, previous) }
	}
	if undo := a.changeKeychain(ctx, restore); undo != nil {
		return errors.Join(err, fmt.Errorf("restoring the client secret of profile %s in the OS keychain "+
			"(service %s, account %s): %w", name, config.KeychainService, name, undo))
	}
	return err
}

// readSecret reads the client secret of profile name: on a terminal, where
// it asks for it, without showing it; otherwise the first line of standard
// input. Spaces around it are dropped; an empty secret is a usage error.
func (a *app) readSecret(cmd *cobra.Command, name string) (string, error) {
	var secret string
	var err error
	if a.terminal {
		secret, err = a.askSecret(cmd, name)
	} else {
		secret, err = readLine(cmd.Context(), cmd.InOrStdin(), "reading the client secret")
	}
	if err != nil {
		return "", err
	}
	if secret = strings.TrimSpace(secret); secret == "" {
		return "", usageError{errors.New("no client secret given: type it when asked, or give it on standard input")}
	}
	return secret, nil
}

// askSecret asks for the client secret of profile name on the terminal and
// reads it without showing it.
func (a *app) askSecret(cmd *cobra.Command, name string) (string, error) {
	errOut := cmd.ErrOrStderr()
	if _, err := fmt.Fprintf(errOut, "Client secret of profile %s: ", name); err != nil {
		return "", err
	}
	secret, err := a.ReadHidden(cmd.Context())
	_, newline := fmt.Fprintln(errOut) // the newline typed was not shown
	return secret, cmp.Or(err, newline)
}

// profileView is a profile as list prints it. Its client secret is never
// shown.
type profileView struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	config.Profile
}

// profileHeader is the table header of list.
func profileHeader() []string {
	return []string{"NAME", "DEFAULT", "TENANT", "CLIENT ID", "API URL", "SITE", "LANGUAGE"}
}

// row returns the table cells of v.
func (v *profileView) row() []string {
	isDefault := ""
	if v.Default {
		isDefault = "yes"
	}
	return []string{v.Name, isDefault, v.Tenant, v.ClientID, v.APIURL, v.Site, v.Language}
}

// profileDetail is a profile as show prints it: also whether the OS keychain
// holds its client secret, never the secret.
type profileDetail struct {
	profileView
	SecretStored bool `json:"secretStored"`
}

// row returns the table cells of d, those of list and CLIENT SECRET.
func (d *profileDetail) row() []string {
	stored := "not stored"
	if d.SecretStored {
		stored = "stored"
	}
	return append(d.profileView.row(), stored)
}

func newProfileListCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the profiles",
		Long:  "Lists the profiles and which one is the default. Client secrets are never shown.",
	}
	printed := addTableOutput(cmd, profileHeader())
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		file, err := a.loadProfiles()
		if err != nil {
			return err
		}
		views := []profileView{}
		var rows [][]string
		for _, name := range slices.Sorted(maps.Keys(file.Profiles)) {
			views = append(views, profileView{Name: name, Default: name == file.Default, Profile: file.Profiles[name]})
			rows = append(rows, views[len(views)-1].row())
		}
		return printed.print(cmd.OutOrStdout(), views, func() ([][]string, error) { return rows, nil })
	}
	return cmd
}

func newProfileShowCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show NAME",
		Short: "Show a profile and whether its client secret is stored",
		Long: "Shows the settings of the profile NAME and whether the OS keychain holds its client secret.\n" +
			"The secret itself is never shown.",
		Args:              profileArgs,
		ValidArgsFunction: a.completeProfiles,
	}
	printed := addTableOutput(cmd, append(profileHeader(), "CLIENT SECRET"))
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		d, err := a.showProfile(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printed.print(cmd.OutOrStdout(), d, func() ([][]string, error) { return [][]string{d.row()}, nil })
	}
	return cmd
}

// showProfile returns the profile name and whether the OS keychain holds its
// client secret.
func (a *app) showProfile(ctx context.Context, name string) (*profileDetail, error) {
	file, err := a.loadProfiles()
	if err != nil {
		return nil, err
	}
	profile, err := file.Lookup(name)
	if err != nil {
		return nil, noProfile(err)
	}
	secret, err := a.secretOf(ctx, name)
	if err != nil {
		return nil, err
	}
	view := profileView{Name: name, Default: name == file.Default, Profile: profile}
	return &profileDetail{profileView: view, SecretStored: secret != ""}, nil
}

func newProfileUseCommand(a *app) *cobra.Command {
	return &cobra.Command{
		Use:               "use NAME",
		Short:             "Make a profile the default",
		Long:              "Makes the profile NAME the default: the one a command uses without --profile and EPX_PROFILE.",
		Args:              profileArgs,
		ValidArgsFunction: a.completeProfiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			err := a.editProfiles(cmd.Context(), func(store config.Store, file config.File) error {
				if err := file.Use(name); err != nil {
					return noProfile(err)
				}
				return store.Save(file)
			})
			if err == nil {
				_, err = fmt.Fprintf(cmd.ErrOrStderr(), "%s is the default profile.\n", name)
			}
			return err
		},
	}
}

func newProfileRemoveCommand(a *app) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove NAME",
		Short: "Remove a profile and its client secret",
		Long: "Removes the profile NAME and deletes its client secret from the OS keychain. On a terminal,\n" +
			"epx asks first; --" + yesFlag + " removes it without asking, as scripts must. Removing the default\n" +
			"profile leaves no default. Where the keychain cannot delete the secret, the profile is removed\n" +
			"all the same, and the error names the keychain entry (service " + config.KeychainService +
			", account NAME) that still holds it.",
		Args:              profileArgs,
		ValidArgsFunction: a.completeProfiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.removeProfile(cmd, args[0], yes)
		},
	}
	cmd.Flags().BoolVar(&yes, yesFlag, false, "remove without asking (required without a terminal)")
	return cmd
}

// removeProfile removes the profile name, once confirmed, then deletes its
// client secret, both under the configuration file's lock: a secret the
// keychain cannot delete is reported once the profile is removed, so a
// keychain that cannot be reached never keeps a profile.
func (a *app) removeProfile(cmd *cobra.Command, name string, yes bool) error {
	if !yes && !a.terminal {
		return usageError{fmt.Errorf("%s asks for confirmation on a terminal, which standard input is not: "+
			"give --%s to remove the profile", cmd.CommandPath(), yesFlag)}
	}
	if !yes {
		if err := a.confirmRemoval(cmd, name); err != nil {
			return err
		}
	}
	err := a.editProfiles(cmd.Context(), func(store config.Store, file config.File) error {
		if err := file.Remove(name); err != nil {
			return noProfile(err)
		}
		if err := store.Save(file); err != nil {
			return err
		}
		err := a.changeKeychain(cmd.Context(), func(k config.Keychain) error { return k.DeleteSecret(name) })
		if err != nil && !errors.Is(err, config.ErrNoSecret) {
			return fmt.Errorf("removed profile %s, but the OS keychain may still hold its client secret "+
				"(service %s, account %s): %w", name, config.KeychainService, name, err)
		}
		return nil
	})
	if err == nil {
		_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Removed profile %s and its client secret.\n", name)
	}
	return err
}

// confirmRemoval asks whether to remove the profile name, which must exist.
func (a *app) confirmRemoval(cmd *cobra.Command, name string) error {
	file, err := a.loadProfiles()
	if err != nil {
		return err
	}
	if _, err = file.Lookup(name); err != nil {
		return noProfile(err)
	}
	question := "Remove profile " + name + " and its client secret?"
	ok, err := confirmed(cmd.Context(), cmd.ErrOrStderr(), cmd.InOrStdin(), question)
	if err == nil && !ok {
		return errors.New("not confirmed; nothing was removed")
	}
	return err
}

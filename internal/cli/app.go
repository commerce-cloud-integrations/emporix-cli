package cli

import (
	"context"
	"time"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// Environment variables read by epx. An API key comes from the environment,
// else from a profile, whose client secret is in the OS keychain: a client
// secret never comes from a flag or a file.
const (
	envAPIURL       = "EMPORIX_API_URL"
	envTenant       = "EMPORIX_TENANT"
	envClientID     = "EMPORIX_CLIENT_ID"
	envClientSecret = "EMPORIX_CLIENT_SECRET" //nolint:gosec // G101: an environment variable name, not a credential.
)

const defaultTimeout = 30 * time.Second

// app carries the process environment, build information and global flags
// to every command.
type app struct {
	getenv    func(string) string
	terminal  bool   // standard input is a terminal, where a write asks for confirmation
	version   string // module version stamped by the Go toolchain (embeds commit and +dirty)
	goVersion string
	tenant    tenantFlag
	profile   profileName
	timeout   time.Duration
	ProfileEnv
	resolved resolvedKey // the API key of the command (key)
	// outTerminal says that standard output is a terminal, which gets no file
	// that a read saves (fileRead).
	outTerminal bool
	// completionTimeout bounds what one completion reads (listedItems).
	completionTimeout time.Duration
}

// client returns an Emporix client whose token is limited to scopes, with
// the settings of the command's API key (settings): those of the
// environment, else of the profile in effect, whose client secret it reads
// from the OS keychain until ctx ends. Missing configuration is a usage
// error; emporix.New marks invalid values ErrInvalidInput.
func (a *app) client(ctx context.Context, scopes ...string) (*emporix.Client, error) {
	cfg, err := a.settings(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Timeout, cfg.UserAgent = a.timeout, "emporix-cli/"+a.version
	// Every translation is requested; tables pick --language, JSON keeps them all.
	cfg.AcceptLanguage = "*"
	return emporix.New(&cfg, scopes...)
}

// tenantFlag is --tenant. The environment's tenant applies only when the flag
// is not given: an empty --tenant "" (as from an unset shell variable) stays
// empty and fails as missing configuration rather than reading another tenant.
type tenantFlag struct {
	value string
	given bool
}

// Set implements pflag.Value.
func (t *tenantFlag) Set(value string) error {
	t.value, t.given = value, true
	return nil
}

func (t *tenantFlag) String() string { return t.value }

// Type implements pflag.Value.
func (t *tenantFlag) Type() string { return "string" }

// or returns the given tenant, else fallback.
func (t *tenantFlag) or(fallback string) string {
	if t.given {
		return t.value
	}
	return fallback
}

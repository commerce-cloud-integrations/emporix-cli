package cli

import (
	"context"
	"time"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const (
	envAPIURL       = "EMPORIX_API_URL"
	envTenant       = "EMPORIX_TENANT"
	envClientID     = "EMPORIX_CLIENT_ID"
	envClientSecret = "EMPORIX_CLIENT_SECRET" //nolint:gosec // G101: an environment variable name, not a credential.
)

const defaultTimeout = 30 * time.Second

type app struct {
	getenv     func(string) string
	inTerminal bool
	version    string
	goVersion  string
	tenant     tenantFlag
	profile    profileName
	timeout    time.Duration
	ProfileEnv
	resolved          resolvedKey
	outTerminal       bool
	completionTimeout time.Duration
}

func (a *app) client(ctx context.Context, scopes ...string) (*emporix.Client, error) {
	cfg, err := a.settings(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Timeout, cfg.UserAgent = a.timeout, "emporix-cli/"+a.version
	cfg.AcceptLanguage = "*"
	return emporix.New(&cfg, scopes...)
}

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

func (t *tenantFlag) or(fallback string) string {
	if t.given {
		return t.value
	}
	return fallback
}

package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// TokenInfo describes an issued service token without its secret value.
type TokenInfo struct {
	Tenant string
	// Scopes are the granted scopes other than the tenant=<name> entry, sorted.
	Scopes    []string
	ExpiresIn time.Duration
}

type token struct {
	value string
	info  TokenInfo
	renew time.Time // when a new token replaces it, before it expires (renewAt); never when zero
}

// Authenticate obtains the client's service token, and a new one in its place
// once it is about to expire, so a long run, such as a wait for a job, never
// sends an expired one; it verifies that each was issued for the configured
// tenant with every requested scope. Emporix silently drops scopes a
// credential lacks, so this check is what makes failures explicit.
func (c *Client) Authenticate(ctx context.Context) (TokenInfo, error) {
	if c.token == nil || c.token.due(c.now()) {
		issued := c.now()
		tok, err := c.requestToken(ctx)
		if err != nil {
			return TokenInfo{}, err
		}
		tok.renew = renewAt(issued, tok.info.ExpiresIn)
		c.token = tok
	}
	return c.token.info, nil
}

// renewAt returns when a token issued at issued (requested then, at the
// latest) that expires after lifetime is replaced: once nine tenths of its
// lifetime passed, which leaves the requests sent with it time to end. A
// token without a lifetime is never replaced.
func renewAt(issued time.Time, lifetime time.Duration) time.Time {
	if lifetime <= 0 {
		return time.Time{}
	}
	return issued.Add(lifetime - lifetime/10)
}

// due reports whether t is to be replaced at now.
func (t *token) due(now time.Time) bool { return !t.renew.IsZero() && !now.Before(t.renew) }

func (c *Client) ensureToken(ctx context.Context) (string, error) {
	if _, err := c.Authenticate(ctx); err != nil {
		return "", err
	}
	return c.token.value, nil
}

func (c *Client) requestToken(ctx context.Context) (*token, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.cfg.ClientID},
		"client_secret": {c.cfg.ClientSecret},
		"scope":         {strings.Join(append([]string{"tenant=" + c.cfg.Tenant}, c.scopes...), " ")},
	}.Encode()
	endpoint := c.base.JoinPath("oauth", "token").String()
	// A token request changes no data, so it is idempotent.
	result, err := c.doWithRetry(ctx, idempotent, func() (*http.Request, error) {
		return newTokenRequest(ctx, endpoint, form, c.cfg.UserAgent)
	})
	if err != nil {
		return nil, err
	}
	if result.status != http.StatusOK {
		return nil, newAPIError(http.MethodPost, "/oauth/token", result.status, result.body,
			tokenErrorClass(result.status))
	}
	return parseToken(result.body, c.cfg.Tenant, c.scopes)
}

func newTokenRequest(ctx context.Context, endpoint, form, userAgent string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// parseToken validates a token response. Errors never include the body,
// because it carries the access token.
func parseToken(body []byte, tenant string, requested []string) (*token, error) {
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.AccessToken == "" {
		return nil, errors.New("token response has no access token")
	}
	if !strings.EqualFold(payload.TokenType, "bearer") {
		return nil, fmt.Errorf("unsupported token type %q", payload.TokenType)
	}
	info := TokenInfo{ExpiresIn: time.Duration(payload.ExpiresIn) * time.Second}
	info.Tenant, info.Scopes = SplitScopes(payload.Scope)
	var missing []string
	for _, scope := range requested {
		if !slices.Contains(info.Scopes, scope) {
			missing = append(missing, scope)
		}
	}
	if info.Tenant != tenant || len(missing) > 0 {
		return nil, &ScopeError{Tenant: tenant, GrantedTenant: info.Tenant, Missing: missing}
	}
	return &token{value: payload.AccessToken, info: info}, nil
}

// SplitScopes splits an OAuth scope text, scopes separated by spaces as
// Emporix grants them, into its tenant entry (tenant=NAME) and its other
// scopes, sorted.
func SplitScopes(text string) (tenant string, scopes []string) {
	for _, scope := range strings.Fields(text) {
		if name, ok := strings.CutPrefix(scope, "tenant="); ok {
			tenant = name
		} else {
			scopes = append(scopes, scope)
		}
	}
	slices.Sort(scopes)
	return tenant, scopes
}

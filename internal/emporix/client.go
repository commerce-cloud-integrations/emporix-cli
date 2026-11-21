// Package emporix is the HTTP client for the Emporix Commerce APIs. It owns
// OAuth service tokens, request execution, retries and error mapping, so
// commands never build HTTP requests themselves.
package emporix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
)

// DefaultBaseURL is the Emporix production API endpoint.
const DefaultBaseURL = "https://api.emporix.io"

const maxBodyBytes = 32 << 20

var tenantPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// Config identifies the API endpoint, tenant and credential a client acts with.
type Config struct {
	BaseURL      string
	Tenant       string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
	UserAgent    string
	// AcceptLanguage is sent with every request when set; "*" asks for every
	// translation of localized fields instead of the tenant's default language.
	AcceptLanguage string
}

// Client performs requests with a service token limited to its scopes: reads
// (GETs, or the reads with a body of a client made by WithBody) and
// writes (Prepare and Send). A Client is not safe for concurrent use.
type Client struct {
	cfg     Config
	base    *url.URL
	scopes  []string
	http    *http.Client
	retry   retryPolicy
	paging  pagingPolicy
	token   *token
	now     func() time.Time
	method  string
	body    Body
	created bool
	refusal bool
	header  http.Header
}

// New validates cfg and returns a client whose token requests exactly the
// tenant plus scopes. Plain HTTP is accepted only for loopback test servers.
func New(cfg *Config, scopes ...string) (*Client, error) {
	if err := CheckEndpoint(cfg.BaseURL, cfg.Tenant); err != nil {
		return nil, err
	}
	base, _ := url.Parse(cfg.BaseURL)
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("%w: client ID and client secret are required", ErrInvalidInput)
	}
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("%w: timeout must be positive", ErrInvalidInput)
	}
	for _, scope := range scopes {
		if err := checkScope(scope); err != nil {
			return nil, err
		}
	}
	base.Path = ""
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		// Never follow redirects: a redirect could forward the bearer token.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return &Client{
		cfg: *cfg, base: base, scopes: scopes, http: httpClient,
		retry: defaultRetryPolicy(), paging: defaultPagingPolicy(), now: time.Now, method: http.MethodGet,
	}, nil
}

// WithBody returns a copy of c whose reads send method, POST or PUT, with
// body (JSONBody, or FormBody for a form) instead of a GET: for a search, a
// calculation or the parsing of a file that Emporix serves so but that
// changes no data. The copy keeps c's credential and scopes (and its token,
// once it has one), retries and error classes: GetAll and its siblings page
// it as they page a collection (pageNumber and pageSize in the query, as
// Emporix pages most searches, or in the body with GetAllBodyPaged, with the
// same total and duplicate checks, limit and stream), Get reads one that
// answers one object, and Check one that answers with its status alone. Only
// 200 OK or 204 No Content answers such a read, unless AcceptCreated says
// otherwise. Any other method, or no body, is ErrInvalidInput, and so is a
// read of a path with a segment that might address another operation once
// decoded (singleSegments): one sent with a body could reach one that changes
// data, such as an import next to a search.
func (c *Client) WithBody(method string, body Body) (*Client, error) {
	switch {
	case method != http.MethodPost && method != http.MethodPut:
		return nil, fmt.Errorf("%w: %q: a read with a body is sent as POST or PUT", ErrInvalidInput, method)
	case body.none():
		return nil, fmt.Errorf("%w: %s: a read with a body sends one", ErrInvalidInput, method)
	}
	return c.sending(method, body), nil
}

// Posting returns a copy of c whose reads send POST without a body instead
// of a GET, for a read that Emporix serves so although it changes no data,
// such as the PDF of a quote (GetFile). As for a read with a body, a read of
// a path with a segment that might address another operation once decoded
// (singleSegments) is ErrInvalidInput.
func (c *Client) Posting() *Client { return c.sending(http.MethodPost, Body{}) }

func (c *Client) sending(method string, body Body) *Client {
	copied := *c
	copied.method, copied.body = method, body
	return &copied
}

// AcceptCreated returns a copy of c whose reads also take 201 Created as their
// answer, as 200 OK, for a calculation that Emporix answers so although it
// creates nothing (unit conversions sent as PUT, seen live; their reference
// documents 201 and no resource). Every other read fails on 201 Created, the
// sign that it may have created something.
func (c *Client) AcceptCreated() *Client {
	accepting := *c
	accepting.created = true
	return &accepting
}

// ForbiddenIsRefusal returns a copy of c whose reads take a 403 Forbidden
// that refuses their input (refusesInput) as of no class, as
// Write.ForbiddenIsRefusal says of a write: for a check that Emporix answers
// so, such as whether a customer may redeem a coupon. Every other read takes
// a 403 as ErrAuth.
func (c *Client) ForbiddenIsRefusal() *Client {
	refusing := *c
	refusing.refusal = true
	return &refusing
}

func checkScope(scope string) error {
	switch {
	case scope == "":
		return fmt.Errorf("%w: invalid scope: empty", ErrInvalidInput)
	case strings.ContainsFunc(scope, unicode.IsSpace):
		return fmt.Errorf("%w: invalid scope %q: one scope per value, without spaces", ErrInvalidInput, scope)
	case strings.HasPrefix(scope, "tenant="):
		return fmt.Errorf("%w: invalid scope %q: the tenant is set separately", ErrInvalidInput, scope)
	}
	return nil
}

// CheckEndpoint checks the API URL and tenant of a Config as New does:
// baseURL is a scheme and host only, https (plain HTTP only for a loopback
// host), and tenant a tenant name. Its errors are ErrInvalidInput.
func CheckEndpoint(baseURL, tenant string) error {
	base, err := url.Parse(baseURL)
	if err != nil || !isOrigin(base) {
		return fmt.Errorf("%w: API URL must be a scheme and host only, e.g. %s", ErrInvalidInput, DefaultBaseURL)
	}
	if base.Scheme != "https" && (base.Scheme != "http" || !isLoopback(base.Hostname())) {
		return fmt.Errorf("%w: API URL must use https (plain http only for a loopback host)", ErrInvalidInput)
	}
	if tenant == "" {
		return fmt.Errorf("%w: tenant is empty", ErrInvalidInput)
	}
	if !tenantPattern.MatchString(tenant) {
		return fmt.Errorf("%w: invalid tenant: use letters, digits, _ and -, starting with a letter or digit",
			ErrInvalidInput)
	}
	return nil
}

func isOrigin(u *url.URL) bool {
	extras := u.User != nil || u.RawQuery != "" || u.Fragment != ""
	return u.Host != "" && strings.Trim(u.Path, "/") == "" && !extras
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Get reads path, which may contain a {tenant} placeholder and a query string;
// values in query replace inline values with the same key (collection reads,
// unlike Get, take their query separately). It returns the JSON
// body. The host always comes from the configured API URL. An empty body is
// ErrNotFound: some services answer a read of a missing item with 200 and
// nothing else, or with 204 No Content. Collections are read with GetAll and
// its siblings instead, which reject an empty body as malformed. A client made
// by WithBody sends its method and body instead of a GET, for a search or
// calculation that answers one object.
func (c *Client) Get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	result, err := c.getNonEmpty(ctx, path, query, nil)
	return result.body, err
}

func (c *Client) getNonEmpty(ctx context.Context, path string, query url.Values, header http.Header) (reply, error) {
	result, err := c.get(ctx, path, query, header)
	if err == nil && result.empty() {
		return reply{}, fmt.Errorf("%s: empty response: %w", c.named(result.path), ErrNotFound)
	}
	return result, err
}

// Check reads path as Get does, for a check that Emporix answers with its
// status alone, such as whether a coupon can be redeemed: a success may come
// without a body. It returns the body, nil when it is empty, and fails as Get
// fails otherwise (the check failing included, which Emporix answers with an
// error status).
func (c *Client) Check(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	result, err := c.get(ctx, path, query, nil)
	if result.empty() {
		return nil, err
	}
	return result.body, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, header http.Header) (reply, error) {
	target, err := c.locate(path, query)
	if err == nil && (c.method == http.MethodPost || c.method == http.MethodPut) {
		err = keepSegments(target, c.named(withoutQuery(path)), "a read sent as "+c.method)
	}
	if err != nil {
		return reply{}, err
	}
	bearer, err := c.ensureToken(ctx)
	if err != nil {
		return reply{}, err
	}
	headers := c.headers(c.body.contentType, c.header, header)
	result, err := c.doWithRetry(ctx, idempotent, func() (*http.Request, error) {
		return newRequest(ctx, c.method, target.String(), bearer, headers, c.body.data)
	})
	if err != nil {
		return reply{}, err
	}
	if !c.answered(result.status) {
		return reply{}, newAPIError(c.method, result.path, result.status, result.body,
			c.failureClass(result.status, result.body))
	}
	return result, nil
}

func (c *Client) failureClass(status int, body []byte) error {
	if c.refusal && refusesInput(status, body) {
		return nil
	}
	return resourceErrorClass(status)
}

func (c *Client) answered(status int) bool {
	return status == http.StatusOK || status == http.StatusNoContent ||
		(c.created && status == http.StatusCreated)
}

type reply struct {
	status    int
	header    http.Header
	body      []byte
	path      string
	uncertain bool
}

func (r reply) empty() bool { return len(bytes.TrimSpace(r.body)) == 0 }

func succeeded(status int) bool { return status >= 200 && status <= 299 }

func (c *Client) exchange(req *http.Request) (reply, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return reply{}, queryLeftOut(err)
	}
	defer func() { _ = resp.Body.Close() }()
	result := reply{status: resp.StatusCode, header: resp.Header, path: req.URL.EscapedPath()}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	switch {
	case err == nil && len(body) <= maxBodyBytes:
		result.body = body
		return result, nil
	case !succeeded(result.status):
		return result, nil
	case err != nil:
		return reply{}, fmt.Errorf("reading response: %w", err)
	default:
		return reply{}, fmt.Errorf("response exceeds %d bytes", maxBodyBytes)
	}
}

func queryLeftOut(err error) error {
	var failed *url.Error
	if errors.As(err, &failed) {
		failed.URL = withoutQuery(failed.URL)
	}
	return err
}

func (c *Client) headers(content string, extras ...http.Header) http.Header {
	header := http.Header{}
	header.Set("Accept", JSONContentType)
	header.Set("User-Agent", c.cfg.UserAgent)
	if c.cfg.AcceptLanguage != "" {
		header.Set("Accept-Language", c.cfg.AcceptLanguage)
	}
	if content != "" {
		header.Set("Content-Type", content)
	}
	for _, extra := range extras {
		for _, key := range slices.Sorted(maps.Keys(extra)) {
			header.Del(key)
			for _, value := range extra[key] {
				header.Add(key, value)
			}
		}
	}
	return header
}

func newRequest(ctx context.Context, method, target, bearer string, header http.Header, body []byte) (
	*http.Request, error) {
	var content io.Reader = http.NoBody
	if body != nil {
		content = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, content)
	if err != nil {
		return nil, err
	}
	req.Header = header.Clone()
	req.Header.Set("Authorization", "Bearer "+bearer)
	return req, nil
}

func (c *Client) locate(path string, query url.Values) (*url.URL, error) {
	ref, err := url.Parse(strings.ReplaceAll(path, "{tenant}", url.PathEscape(c.cfg.Tenant)))
	if err != nil || !isRelativePath(path, ref) {
		return nil, fmt.Errorf("%w: API path %q: use a path such as /product/{tenant}/products", ErrInvalidInput,
			pathOnly(path))
	}
	merged, err := url.ParseQuery(ref.RawQuery)
	if err != nil {
		// Neither the query nor the parser's message is named: a bad escape may hold a credential.
		return nil, fmt.Errorf("%w: API path %q: malformed query", ErrInvalidInput, pathOnly(path))
	}
	target := *c.base
	target.Path, target.RawPath = ref.Path, ref.RawPath
	for key, values := range query {
		merged[key] = values
	}
	target.RawQuery = merged.Encode()
	return &target, nil
}

func pathOnly(path string) string {
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		return path[:i]
	}
	return path
}

func keepSegments(target *url.URL, named, what string) error {
	if singleSegments(target.EscapedPath()) {
		return nil
	}
	return fmt.Errorf("%w: %s: a path segment of %s is empty, . or .., or holds /, \\ or %%", ErrInvalidInput, named,
		what)
}

func singleSegments(escaped string) bool {
	for _, segment := range strings.Split(strings.TrimPrefix(escaped, "/"), "/") {
		decoded, err := url.PathUnescape(segment)
		if err != nil || !oneSegment(decoded) {
			return false
		}
	}
	return true
}

func oneSegment(decoded string) bool {
	switch decoded {
	case "", ".", "..":
		return false
	}
	return !strings.ContainsAny(decoded, `/\%`)
}

func isRelativePath(raw string, ref *url.URL) bool {
	leavesHost := ref.Scheme != "" || ref.Host != "" || strings.HasPrefix(raw, "//")
	return !leavesHost && ref.Fragment == "" && strings.HasPrefix(ref.Path, "/")
}

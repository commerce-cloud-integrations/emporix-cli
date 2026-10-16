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
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// DefaultBaseURL is the Emporix production API endpoint.
const DefaultBaseURL = "https://api.emporix.io"

// maxBodyBytes bounds every response read into memory.
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
	now     func() time.Time // the clock that tells when the token is to be replaced (Authenticate)
	method  string           // the method of every read: GET, or POST or PUT with body (WithBody)
	body    Body             // the body of every read; none for a GET
	created bool             // 201 Created answers a read as 200 OK does (AcceptCreated)
	refusal bool             // a 403 Forbidden may refuse a read's input (ForbiddenIsRefusal, refusesInput)
}

// New validates cfg and returns a client whose token requests exactly the
// tenant plus scopes. Plain HTTP is accepted only for loopback test servers.
func New(cfg *Config, scopes ...string) (*Client, error) {
	if err := CheckEndpoint(cfg.BaseURL, cfg.Tenant); err != nil {
		return nil, err
	}
	base, _ := url.Parse(cfg.BaseURL) // CheckEndpoint parsed it
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

// sending returns a copy of c whose reads send method with body (the zero
// Body: none).
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

// checkScope rejects a scope that is empty, holds spaces (one scope per value)
// or names the tenant, which New adds itself.
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
	// Rejected values are not echoed: they may hold a credential set in the
	// wrong place.
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

// isOrigin reports whether u is only a scheme and host, without credentials,
// path, query or fragment.
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

// getNonEmpty reads path as get does, with header, and fails an empty body as
// ErrNotFound (Get).
func (c *Client) getNonEmpty(ctx context.Context, path string, query url.Values, header http.Header) (reply, error) {
	result, err := c.get(ctx, path, query, header)
	if err == nil && result.empty() {
		// The query is left out of the message: it may hold values a log must not keep.
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
		return nil, err // a failed read has no body either
	}
	return result.body, nil
}

// get performs an authenticated read with optional extra headers: a GET, or
// the method and body of a client made by WithBody, or a POST without a body
// of one made by Posting. Responses that do not
// answer a read (answered) become classified APIErrors.
func (c *Client) get(ctx context.Context, path string, query url.Values, header http.Header) (reply, error) {
	target, err := c.locate(path, query)
	if err == nil && (c.method == http.MethodPost || c.method == http.MethodPut) {
		// One sent so could reach an operation that changes data, such as an import next to a search.
		err = keepSegments(target, c.named(withoutQuery(path)), "a read sent as "+c.method)
	}
	if err != nil {
		return reply{}, err
	}
	bearer, err := c.ensureToken(ctx)
	if err != nil {
		return reply{}, err
	}
	headers := c.headers(header, c.body.contentType)
	// A read changes no data, so it is idempotent, also when sent as POST or PUT.
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

// failureClass classifies the failure status of a read of c, answered with
// body, as a resource request's (resourceErrorClass), but a 403 Forbidden
// that refuses its input where c says that the service answers so (c.refusal,
// refusesInput) as of no class.
func (c *Client) failureClass(status int, body []byte) error {
	if c.refusal && refusesInput(status, body) {
		return nil
	}
	return resourceErrorClass(status)
}

// answered reports whether status answers a read: 200 OK, 204 No Content
// (which has no body), or 201 Created for a client made by AcceptCreated.
func (c *Client) answered(status int) bool {
	return status == http.StatusOK || status == http.StatusNoContent ||
		(c.created && status == http.StatusCreated)
}

// reply is a completed HTTP exchange whose body has been read and closed.
type reply struct {
	status    int
	header    http.Header
	body      []byte
	path      string // the path requested, escaped, without the query (as error messages name it)
	uncertain bool   // an attempt may have been carried out without an answer that says so (doWithRetry)
}

// empty reports whether the reply has no body but white space, as a check
// that passed or a missing item may be answered.
func (r reply) empty() bool { return len(bytes.TrimSpace(r.body)) == 0 }

// succeeded reports whether status is a success (2xx).
func succeeded(status int) bool { return status >= 200 && status <= 299 }

// exchange sends req and reads at most maxBodyBytes of the response. The body
// of a failure only explains it: one that cannot be read whole is left out,
// so that its status still classifies the failure (a 409 stays a conflict, a
// 404 not found) and decides whether it is sent again.
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

// queryLeftOut leaves the query out of the URL that a failure to send names
// (*url.Error), as every error of the client leaves it out: it may hold
// values a log must not keep. The failure it wraps is kept.
func queryLeftOut(err error) error {
	var failed *url.Error
	if errors.As(err, &failed) {
		failed.URL = withoutQuery(failed.URL)
	}
	return err
}

// headers returns the headers of an API request but its Authorization:
// Accept (JSON), User-Agent, Accept-Language when the client sends one and,
// with a body, its Content-Type, content (empty: no body); a header of extra
// replaces the one of the same name, one of extra without values removes it
// (as a read of a file removes Accept-Language: GetFile), and
// its others are added.
func (c *Client) headers(extra http.Header, content string) http.Header {
	header := http.Header{}
	header.Set("Accept", JSONContentType)
	header.Set("User-Agent", c.cfg.UserAgent)
	if c.cfg.AcceptLanguage != "" {
		header.Set("Accept-Language", c.cfg.AcceptLanguage)
	}
	if content != "" {
		header.Set("Content-Type", content)
	}
	for key, values := range extra {
		header.Del(key)
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header
}

// newRequest returns a request of target with method, header and the bearer
// token, and with body unless it is nil.
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

// locate builds the request URL from a relative API path, which may contain
// {tenant} and a query.
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
	// Explicit query values replace inline ones with the same key, as
	// `epx api get PATH -p key=value` documents. (Collection reads refuse an
	// inline query, so their paging and filters are always the explicit ones.)
	for key, values := range query {
		merged[key] = values
	}
	target.RawQuery = merged.Encode()
	return &target, nil
}

// pathOnly returns path without its query or fragment, for error messages: a
// query may hold a filter value or a credential.
func pathOnly(path string) string {
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		return path[:i]
	}
	return path
}

// keepSegments fails, as ErrInvalidInput, a request to target, which named
// names (e.g. "POST /x") and what describes (e.g. "a write"), unless every
// segment of its path stays one once decoded (singleSegments).
func keepSegments(target *url.URL, named, what string) error {
	if singleSegments(target.EscapedPath()) {
		return nil
	}
	return fmt.Errorf("%w: %s: a path segment of %s is empty, . or .., or holds /, \\ or %%", ErrInvalidInput, named,
		what)
}

// singleSegments reports whether every segment of escaped, a path as sent,
// stays one segment that addresses one item once decoded (oneSegment).
func singleSegments(escaped string) bool {
	for _, segment := range strings.Split(strings.TrimPrefix(escaped, "/"), "/") {
		decoded, err := url.PathUnescape(segment)
		if err != nil || !oneSegment(decoded) {
			return false
		}
	}
	return true
}

// oneSegment reports whether decoded, a decoded path segment, addresses one
// item: it is not empty, "." or "..", and holds no "/" or "\" that a service,
// or a proxy before it, might take for a separator, nor a "%" that it might
// decode again.
func oneSegment(decoded string) bool {
	switch decoded {
	case "", ".", "..":
		return false
	}
	return !strings.ContainsAny(decoded, `/\%`)
}

// isRelativePath reports whether raw (parsed as ref) is an absolute path on the
// configured host: no scheme, no host (including "//host" forms), no fragment.
func isRelativePath(raw string, ref *url.URL) bool {
	leavesHost := ref.Scheme != "" || ref.Host != "" || strings.HasPrefix(raw, "//")
	return !leavesHost && ref.Fragment == "" && strings.HasPrefix(ref.Path, "/")
}

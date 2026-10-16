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
)

// Answers says how the service answers a write beyond its status, as the
// policies of a write declare it; the zero value is the usual answer. A Write
// carries them as its own fields, promoted from Answers.
type Answers struct {
	// StaleNotFound says that the service answers a stale metadata.version
	// with 404 Not Found instead of 409 Conflict, as the Unit Service does: a
	// 404 to a body that carries metadata.version is then ErrConflict.
	StaleNotFound bool
	// ForbiddenIsRefusal says that the service answers an input of w that it
	// will not act on with 403 Forbidden, although the token holds every scope
	// that it needs, as the Order Service answers a transition to a status that
	// the order cannot take: such a 403 is then a failure of no class
	// (refusesInput).
	ForbiddenIsRefusal bool
	// ReportsOutcome says that the service answers w, when it answers 2xx,
	// with a JSON object whose boolean member successful says whether it was
	// carried out, as the payment operations of the Payment Gateway do (Send).
	ReportsOutcome bool
	// FailureAnswered says that the service answers a conflict of w with a body
	// that says why, as the Import Service answers a cancel of an unknown run
	// with 409 Conflict and "accepted": false; Send returns that body with the
	// error (Request.failureAnswered), so that the write can print it.
	FailureAnswered bool
}

// Write is a request that may change data: a create, update, patch, delete
// or action. Prepare checks it and returns the Request that Send sends.
type Write struct {
	Method string      // http.MethodPost, MethodPut, MethodPatch or MethodDelete
	Path   string      // a path on the API host, which may contain {tenant} and a query, as Get takes it
	Query  url.Values  // query values, which replace inline ones of the same key
	Header http.Header // headers that the service requires besides every request's, e.g. Content-Language
	Body   Body        // JSONBody or FormBody; the zero Body: none
	// Parts makes w a write of parts that Emporix answers one by one, the
	// items of a bulk write (Items) or the commands of a chain (Commands),
	// which Body then holds (Send); nil for one value.
	Parts *Parts
	// AnsweredWhole says that Emporix answers the parts of w as a whole, as
	// it answers a write of one value (e.g. 204 No Content), not with the
	// result of each: Prepare still checks them as Parts says.
	AnsweredWhole bool
	// Batched lets a bulk write (Items) hold more items than Emporix takes
	// in one request, which Split then sends in batches of up to that many,
	// one request each.
	Batched bool
	// Answers says how the service answers w beyond its status (Answers).
	Answers
	// first is the position of the first item of the body among the items of
	// the write that Split made w a batch of (ItemFailure.Index); 0 for any
	// other.
	first int
}

// Request is a write that Prepare checked, as Send sends it: what a dry run
// shows and a confirmation names. Only Prepare makes one, and only its client
// sends it, so the request shown is the one sent: to the configured API host,
// never redirected, with a token limited to the client's scopes.
type Request struct {
	by     *Client // the client that prepared it, the only one that sends it
	method string
	url    string      // the full URL, with the tenant and the query
	header http.Header // every header but Authorization, which carries the token
	body   Body
	// answers are the answers of the service that Send checks, as the write
	// declares them; StaleNotFound is set only where the body carries a
	// metadata.version (carriesVersion), so that a stale version can be told.
	answers Answers
	// parts are the parts of a write of parts (Write.Parts), whose results
	// Send reads; nil for one value.
	parts *sentParts
}

// Method returns the method of r.
func (r *Request) Method() string { return r.method }

// URL returns the full URL of r, with the tenant and the query.
func (r *Request) URL() string { return r.url }

// Header returns a copy of the headers of r but Authorization, which carries
// the token that Send requests.
func (r *Request) Header() http.Header { return r.header.Clone() }

// Body returns a copy of the body of r as sent, nil for none; its
// Content-Type is among the headers.
func (r *Request) Body() []byte { return slices.Clone(r.body.data) }

// Items returns the number of parts that r, a write of parts, sends; 0 for
// any other write, and for one that Emporix answers as a whole.
func (r *Request) Items() int {
	if r.parts == nil {
		return 0
	}
	return len(r.parts.parts)
}

// writeMethods are the methods of a write.
func writeMethods() []string {
	return []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
}

// Prepare checks w and returns the request that Send sends for it, without
// requesting a token or sending anything. A header of w.Header replaces the
// one of the same name that every request sends. A method other than POST,
// PUT, PATCH or DELETE, a header value that cannot be sent (CheckHeaderValue),
// a path that leaves the API host or holds a segment that might address
// another item once decoded (singleSegments), or a write of parts whose body
// does not hold one to as many parts as Emporix takes in one request (Parts;
// also one answered as a whole) is ErrInvalidInput.
func (c *Client) Prepare(w *Write) (*Request, error) {
	named := w.Method + " " + withoutQuery(w.Path)
	if !slices.Contains(writeMethods(), w.Method) {
		return nil, fmt.Errorf("%w: %q: a write is sent as POST, PUT, PATCH or DELETE", ErrInvalidInput, w.Method)
	}
	if err := checkHeaders(named, w.Header); err != nil {
		return nil, err
	}
	target, err := c.locate(w.Path, w.Query)
	if err == nil {
		err = keepSegments(target, named, "a write")
	}
	if err != nil {
		return nil, err
	}
	parts, err := w.Parts.split(named, w.Body.data, target.Query())
	if err != nil {
		return nil, err
	}
	switch {
	case w.AnsweredWhole:
		parts = nil // checked, but answered as one value
	case parts != nil:
		parts.first = w.first
	}
	checked := w.Answers
	checked.StaleNotFound = w.StaleNotFound && carriesVersion(w.Body.data)
	return &Request{by: c, method: w.Method, url: target.String(), header: c.headers(w.Header, w.Body.contentType),
		body: w.Body, parts: parts, answers: checked}, nil
}

// checkHeaders returns ErrInvalidInput, naming the request named (a write,
// or a read of a file: GetFile), when a value of header cannot be sent
// (CheckHeaderValue); nil when all can.
func checkHeaders(named string, header http.Header) error {
	for name, values := range header {
		for _, value := range values {
			if err := CheckHeaderValue(value); err != nil {
				return fmt.Errorf("%w: %s: the header %s %w", ErrInvalidInput, named, name, err)
			}
		}
	}
	return nil
}

// CheckHeaderValue returns why value cannot be sent as the value of a header,
// which HTTP does not allow to hold a control character other than a tab,
// such as a line break; nil when it can be sent.
func CheckHeaderValue(value string) error {
	if strings.ContainsFunc(value, func(r rune) bool { return r < ' ' && r != '\t' || r == 0x7f }) {
		return errors.New("holds a control character")
	}
	return nil
}

// carriedOutNote ends the error of a write that may have been carried out
// although it failed (Send).
const carriedOutNote = "it may have been carried out all the same: read the item before you send it again"

// Send sends r, which Prepare of c made (ErrInvalidInput otherwise), with the
// client's token, and returns the response body as sent, nil when it is empty
// (as a 201 Created or 204 No Content may be). Any 2xx status is a success.
// Failures are classified as those of reads are; one that conflicts with the
// current state of the item (409) is ErrConflict. A write that may have
// reached the service is never sent again (the retry rule): one that fails
// after it may have been carried out (a 500, 502, 503 or 504, or a failure
// once its connection was made) says so. Where the service answers a stale
// metadata.version with 404 (Write.StaleNotFound), a 404 to a body that
// carries one is ErrConflict, and where it refuses an input with 403
// Forbidden (Write.ForbiddenIsRefusal), such a refusal is of no class
// (refusesInput). Of a write of parts, Emporix answers with the
// result of each part (Parts), also in the body of a response that failed:
// when one failed, Send returns the response with an *ItemsError, and with an
// error when the response does not give the result of each part that ran. Of
// a write whose response says whether it was carried out
// (Write.ReportsOutcome), Send returns the response with an error where it
// does not say that it was (outcome). Of a write that declares that its
// conflicts are answered with a body (Write.FailureAnswered), Send returns that
// body with the error.
func (c *Client) Send(ctx context.Context, r *Request) ([]byte, error) {
	if r.by != c {
		return nil, fmt.Errorf("%w: a write is sent by the client that prepared it", ErrInvalidInput)
	}
	bearer, err := c.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.doWithRetry(ctx, onceOnly, func() (*http.Request, error) {
		return newRequest(ctx, r.method, r.url, bearer, r.header, r.body.data)
	})
	if err == nil && !succeeded(result.status) {
		err = newAPIError(r.method, result.path, result.status, result.body, r.failureClass(result.status, result.body))
	}
	if err != nil && result.uncertain {
		err = fmt.Errorf("%w (%s)", err, carriedOutNote)
	}
	switch {
	case r.parts != nil:
		return r.parts.check(r.method, result, err)
	case err != nil && r.answers.FailureAnswered && result.status == http.StatusConflict:
		return responseBody(result), err
	case err != nil:
		return nil, err
	case r.answers.ReportsOutcome:
		body := responseBody(result)
		return body, outcome(r.method, result.path, body)
	default:
		return responseBody(result), nil
	}
}

// responseBody returns the body of result, nil when it is empty.
func responseBody(result reply) []byte {
	if result.empty() {
		return nil
	}
	return result.body
}

// failureClass classifies the failure status of r, answered with body, as a
// write's (writeErrorClass), but a 404 Not Found that may answer a stale
// metadata.version (r.answers.StaleNotFound) as a conflict, as the item may have
// changed since it was read (reading it again tells that apart from an
// unknown item), and a 403 Forbidden that refuses its input where its service
// answers so (r.answers.ForbiddenIsRefusal, refusesInput) as of no class.
func (r *Request) failureClass(status int, body []byte) error {
	switch {
	case status == http.StatusNotFound && r.answers.StaleNotFound:
		return ErrConflict
	case r.answers.ForbiddenIsRefusal && refusesInput(status, body):
		return nil
	default:
		return writeErrorClass(status)
	}
}

// carriesVersion reports whether body is a JSON object whose member
// metadata, an object, has a version other than null (names match exactly,
// as Emporix matches them).
func carriesVersion(body []byte) bool {
	var item, metadata map[string]json.RawMessage
	if json.Unmarshal(body, &item) != nil || json.Unmarshal(item["metadata"], &metadata) != nil {
		return false
	}
	version, found := metadata["version"]
	return found && string(version) != "null"
}

// Tenant returns the tenant that the client acts for, which its token names.
func (c *Client) Tenant() string { return c.cfg.Tenant }

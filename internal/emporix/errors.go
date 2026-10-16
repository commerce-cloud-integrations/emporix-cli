package emporix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxMessageBytes bounds error text taken from a response body.
const maxMessageBytes = 300

// Error classes. Every error returned by this package that callers should act
// on matches one of these with errors.Is, so callers never inspect HTTP statuses.
var (
	// ErrInvalidInput reports invalid configuration or request input; nothing was sent.
	ErrInvalidInput = errors.New("invalid input")
	// ErrAuth reports a credential that cannot be used: the token endpoint
	// rejected it, or the issued token lacks the requested tenant or scopes.
	ErrAuth = errors.New("authentication failed")
	// ErrNotFound reports a resource the API does not know.
	ErrNotFound = errors.New("not found")
	// ErrConflict reports a write that Emporix answered with 409 Conflict, as
	// it answers one that conflicts with the current state of the item: the
	// item changed since it was read (optimistic locking through
	// metadata.version), one with the same identifier exists, or a job on it
	// is still running; or with 404 Not Found where a service answers a stale
	// metadata.version so (Write.StaleNotFound). Reading the item again shows
	// its current state.
	ErrConflict = errors.New("conflict")
)

// APIError is a non-success response from the Emporix API.
type APIError struct {
	Method  string
	Path    string
	Status  int
	Message string
	class   error // ErrAuth, ErrNotFound, ErrConflict or nil
}

func (e *APIError) Error() string {
	text := fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.Status, http.StatusText(e.Status))
	if e.Message != "" {
		text += ": " + e.Message
	}
	return text
}

// Unwrap exposes the error class.
func (e *APIError) Unwrap() error { return e.class }

// ScopeError reports a token issued for another tenant or without requested
// scopes. It matches ErrAuth with errors.Is.
type ScopeError struct {
	Tenant        string
	GrantedTenant string
	Missing       []string
}

func (e *ScopeError) Error() string {
	if e.GrantedTenant != e.Tenant {
		return fmt.Sprintf("token was issued for tenant %q, not %q; check the credential", e.GrantedTenant, e.Tenant)
	}
	return fmt.Sprintf("credential for tenant %q lacks required scopes: %s", e.Tenant, strings.Join(e.Missing, " "))
}

func (e *ScopeError) Unwrap() error { return ErrAuth }

func newAPIError(method, path string, status int, body []byte, class error) *APIError {
	return &APIError{Method: method, Path: path, Status: status, Message: errorMessage(body), class: class}
}

// refusesInput reports whether a 403 Forbidden answered with body refuses
// the input of a request whose service answers so an input that it will not
// act on (Write.ForbiddenIsRefusal, Client.ForbiddenIsRefusal): one with a
// message, e.g. "Order id …, with status CREATED, could not be SHIPPED", whose
// body names no scope or permission (namesScopeOrPermission). Such a refusal
// is of no class, as a 400 Bad Request is: epx checked that the token holds
// every scope that it requested, so the credential is fine.
//
// Any other 403 stays ErrAuth, as the credential lacks a scope or permission
// that Emporix needs: one whose body names a scope anywhere, also past the
// part that the message shows ("Missing one of required scopes '[…]'", "A
// scope … is required in order to set customerNumber"), or a permission (the
// type insufficient_permissions of the Coupon Service's documented 403), and
// one without a message.
func refusesInput(status int, body []byte) bool {
	return status == http.StatusForbidden && errorMessage(body) != "" && !namesScopeOrPermission(body)
}

// namesScopeOrPermission reports whether body holds "scope" or "permission",
// in whatever case.
func namesScopeOrPermission(body []byte) bool {
	lower := bytes.ToLower(body)
	return bytes.Contains(lower, []byte("scope")) || bytes.Contains(lower, []byte("permission"))
}

// resourceErrorClass classifies a failed resource request.
func resourceErrorClass(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrAuth
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return nil
	}
}

// writeErrorClass classifies a failed write: a conflict (409) is ErrConflict,
// any other failure is classified as a read's. A read answered 409 is no
// conflict of its own, as the Shipping Service answers a read of an unknown
// group so. A write is never sent again after an attempt that may have been
// carried out (the retry rule), so the status classified answers the only
// attempt that may have reached the service.
func writeErrorClass(status int) error {
	if status == http.StatusConflict {
		return ErrConflict
	}
	return resourceErrorClass(status)
}

// tokenErrorClass classifies a failed token request. Emporix answers 400 when
// the requested tenant does not belong to the credential.
func tokenErrorClass(status int) error {
	switch status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		return ErrAuth
	default:
		return nil
	}
}

// errorMessage extracts the human-readable message from an error body: of a
// JSON body, its message, error code and details (jsonMessage), e.g. "The
// request violates validation constraints. (ENTITY-400001) (name: must not
// be null)", and none of its other members, which may carry tokens; a
// plain-text body as it is. The result is bounded; terminal sanitizing is the
// caller's job.
func errorMessage(body []byte) string {
	text := string(body)
	if json.Valid(body) {
		text = jsonMessage(body)
	}
	text = strings.TrimSpace(text)
	if len(text) <= maxMessageBytes {
		return text
	}
	cut := maxMessageBytes
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "…"
}

// jsonMessage returns the text of a JSON error body. It is made of:
//   - its message: the first non-blank string of message, errorMessage (an
//     item of the Cart Service's itemsBatch), OAuth's Error, error_description
//     and error, the gateway's fault.faultstring, and problem+json's detail
//     and title (the Reward Points Service);
//   - its errorCode (the Shipping Service and others), in parentheses after
//     the message, unless the message names it already;
//   - its details, in parentheses: those of its details member, of
//     Micronaut's _embedded.errors (the Brand, Price, Schema and Unit
//     Services) and of the Schema Service's fileDetails, in that order.
//
// Without a message, the code and the details stand first, as in
// "ENTITY-400001 (name: must not be blank)" or "name: must not be blank".
func jsonMessage(body []byte) string {
	var fields map[string]any
	if json.Unmarshal(body, &fields) != nil {
		return ""
	}
	if fault, ok := fields["fault"].(map[string]any); ok {
		fields = fault
	}
	message := firstText(fields, "message", "errorMessage", "Error", "error_description", "error", "faultstring",
		"detail", "title")
	summary := message
	if code := firstText(fields, "errorCode"); code != "" && !namesWord(message, code) {
		summary = appendNote(message, code)
	}
	embedded, _ := fields["_embedded"].(map[string]any)
	var texts []string
	for _, member := range []any{fields["details"], embedded["errors"], fields["fileDetails"]} {
		texts = append(texts, detailTexts(member, message)...)
	}
	return appendNote(summary, strings.Join(texts, "; "))
}

// appendNote returns text followed by note in parentheses; either alone
// where the other is empty.
func appendNote(text, note string) string {
	switch {
	case note == "":
		return text
	case text == "":
		return note
	default:
		return text + " (" + note + ")"
	}
}

// namesWord reports whether text holds word as a word of its own, not as a
// part of a longer one, e.g. an error code "E-1" in "failed (E-1)" but not in
// "failed (E-10)".
func namesWord(text, word string) bool {
	for from := 0; ; {
		at := strings.Index(text[from:], word)
		if at < 0 {
			return false
		}
		at += from
		if !continuesWord(text[:at], true) && !continuesWord(text[at+len(word):], false) {
			return true
		}
		from = at + 1
	}
}

// continuesWord reports whether the text next to a match continues its
// word: the last character of the text before it (atEnd) or the first one
// of the text after it is a letter, a digit or a hyphen.
func continuesWord(text string, atEnd bool) bool {
	if text == "" {
		return false
	}
	var r rune
	if atEnd {
		r, _ = utf8.DecodeLastRuneInString(text)
	} else {
		r, _ = utf8.DecodeRuneInString(text)
	}
	return r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// firstText returns the first non-blank string among the members keys of
// object, trimmed; none where it has none.
func firstText(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, _ := object[key].(string); strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

// detailTexts returns the text of each detail in member, a details-like
// member of an error body, that holds one and does not merely repeat
// message. Each takes one of the shapes that the references document for a
// details member, which Micronaut's _embedded.errors ({"message","path"})
// and fileDetails take too:
//   - an array of objects: {"field","type","message","moreInfo"} or {"message"};
//   - an array of strings;
//   - one such object (the Customer Segment Service).
//
// Another shape gives none.
func detailTexts(member any, message string) []string {
	var items []any
	switch d := member.(type) {
	case []any:
		items = d
	case map[string]any:
		items = []any{d}
	}
	var texts []string
	for _, item := range items {
		if text := detailText(item); text != "" && text != message {
			texts = append(texts, text)
		}
	}
	return texts
}

// detailText returns the text of one detail:
//   - of a string, the string;
//   - of an object, "field: message", where field is its field or else its
//     path; its message alone where it names neither or the message starts
//     with the field already (the Availability Service's "productId: must
//     not be blank"); none of its other members, which may hold anything;
//   - of anything else, none.
func detailText(detail any) string {
	switch d := detail.(type) {
	case string:
		return strings.TrimSpace(d)
	case map[string]any:
		text, field := firstText(d, "message"), firstText(d, "field", "path")
		if text == "" || field == "" || startsWithField(text, field) {
			return text
		}
		return field + ": " + text
	default:
		return ""
	}
}

// startsWithField reports whether a detail's message starts with its field
// as a word of its own, followed by a colon, a space or nothing.
func startsWithField(message, field string) bool {
	rest, found := strings.CutPrefix(message, field)
	return found && (rest == "" || rest[0] == ':' || rest[0] == ' ')
}

package emporix

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// A bulk write sends a JSON array of items, which Emporix carries out one by
// one: it answers 207 Multi-Status with a JSON array that holds the result
// of each item, such as
// {"index": 1, "id": "…", "code": 400, "status": "Bad Request", "message": "…"}.
// index names the item by its position in the body sent, from 0, or, where
// each result names an index and none names 0, from 1 (firstIndex); a result
// without one is the result of the item at its own position, as Emporix's
// guide to bulk requests says it answers. Where neither index nor id names
// the item, a member of both may (Items), such as the eventType of a change
// of a webhook subscription. The Cart Service's itemsBatch gives the code as
// status, a number, with errorMessage: {"index": 1, "status": 400}, and
// answers a batch that failed in part with 400 or 500 and that array.

// ItemFailure is the result of one part of a write sent in parts (Parts), an
// item or a command, that Emporix did not carry out, or, with a 500, 502,
// 503 or 504, may have carried out all the same (as a write answered so may
// have been, Send); or of one item of a command that sends items, such as a
// batch of a chain.
type ItemFailure struct {
	Index   int    // the part's position in the body sent, from 0, as its result names it; of a batch, in the write split
	ID      string // the item's id as Emporix names it, or its value of the key of Items; empty when it names none
	Type    string // a command's type, such as AddCartItem; empty for an item
	Status  int    // the status of its result
	Message string // what Emporix says of it, bounded as an APIError's message
	// Item is, for a command that sends items (a batch), the item of it that
	// failed, whose status classifies the failure; nil otherwise.
	Item *ItemFailure
}

// String names the failure as an item's (ItemsError.Lines names the parts of
// a chain as commands): its position and id, its status and message, and, for
// a status after which it may have been carried out, says so.
func (f ItemFailure) String() string { return f.line(ItemNoun) }

func (f ItemFailure) line(noun string) string {
	text := noun + " " + strconv.Itoa(f.Index)
	if name := cmp.Or(f.Type, f.ID); name != "" {
		text += " (" + name + ")"
	}
	if f.Item != nil {
		return text + ", " + f.Item.String()
	}
	text += ": " + strings.TrimSpace(fmt.Sprintf("%d %s", f.Status, http.StatusText(f.Status)))
	if f.Message != "" {
		text += ": " + f.Message
	}
	if unanswered(f.Status, nil) {
		text += " (" + carriedOutNote + ")"
	}
	return text
}

func (f ItemFailure) status() int {
	if f.Item != nil {
		return f.Item.status()
	}
	return f.Status
}

// ItemsError reports a write sent in parts (Parts) that Emporix carried out
// for some of its parts only, items or commands: the others failed, each as
// its result says, or did not run. It matches the class that every failure
// shares, as a write's status classifies it (a conflict, an item not found, a
// scope missing), and none when they differ.
type ItemsError struct {
	Method string
	Path   string
	Noun   string        // what a part is called: ItemNoun (also when empty) or CommandNoun
	Items  int           // the parts sent
	NotRun int           // the last parts, which did not run, as a chain stops after a command that failed
	Failed []ItemFailure // the parts that failed, in the order of the response; a command's items each apart
	class  error
}

func (e *ItemsError) Error() string {
	text := fmt.Sprintf("%s %s: %d of %d %ss failed", e.Method, e.Path, e.failedParts(), e.Items, e.noun())
	if e.NotRun > 0 {
		text += fmt.Sprintf(", and the last %d did not run", e.NotRun)
	}
	return text
}

// Lines returns a line per failure, naming the part as String names an
// item's, in the order of Failed.
func (e *ItemsError) Lines() []string {
	lines := make([]string, len(e.Failed))
	for i, failure := range e.Failed {
		lines[i] = failure.line(e.noun())
	}
	return lines
}

func (e *ItemsError) noun() string { return cmp.Or(e.Noun, ItemNoun) }

func (e *ItemsError) failedParts() int {
	parts := 0
	for i, failure := range e.Failed {
		if i == 0 || failure.Index != e.Failed[i-1].Index {
			parts++
		}
	}
	return parts
}

// Unwrap exposes the class that every failure shares, if any.
func (e *ItemsError) Unwrap() error { return e.class }

type sentItems struct {
	count int
	key   string
	names []string
}

func namedItems(items []json.RawMessage, key string) sentItems {
	sent := sentItems{count: len(items), key: key}
	if key != "" {
		sent.names = make([]string, len(items))
		for i, item := range items {
			sent.names[i] = memberString(item, key)
		}
	}
	return sent
}

func memberString(raw json.RawMessage, key string) string {
	var object map[string]json.RawMessage
	var value string
	if json.Unmarshal(raw, &object) != nil || json.Unmarshal(object[key], &value) != nil {
		return ""
	}
	return value
}

type itemResult struct {
	Index  *int            `json:"index"` // the item's position in the body sent (firstIndex); none: the result's own
	ID     string          `json:"id"`
	Code   int             `json:"code"`
	Status json.RawMessage `json:"status"` // the code where code is missing, as a number; else its text
}

func (r itemResult) code() int {
	var status int
	if r.Code == 0 && json.Unmarshal(r.Status, &status) == nil {
		return status
	}
	return r.Code
}

func (s sentItems) item(r itemResult, raw json.RawMessage, position, first int, read []bool) (item int,
	name string, isItem bool) {
	if s.key == "" {
		if r.Index != nil {
			position = *r.Index - first
		}
		return position, r.ID, position >= 0 && position < s.count && !read[position]
	}
	name = memberString(raw, s.key)
	for i, sent := range s.names {
		if sent == name && !read[i] {
			return i, name, true
		}
	}
	return 0, name, false
}

func readItems(sent *sentParts, results []json.RawMessage) ([]ItemFailure, int, bool) {
	failed, ok := itemFailures(namedItems(sent.parts, sent.key), results)
	return failed, 0, ok
}

func firstIndex(results []json.RawMessage) int {
	for _, raw := range results {
		var result itemResult
		if json.Unmarshal(raw, &result) != nil || result.Index == nil || *result.Index == 0 {
			return 0
		}
	}
	return 1
}

func itemFailures(items sentItems, results []json.RawMessage) ([]ItemFailure, bool) {
	if len(results) != items.count {
		return nil, false
	}
	var failed []ItemFailure
	first := firstIndex(results)
	read := make([]bool, items.count)
	for position, raw := range results {
		var result itemResult
		if json.Unmarshal(raw, &result) != nil || !isStatus(result.code()) {
			return nil, false
		}
		item, name, isItem := items.item(result, raw, position, first, read)
		if !isItem {
			return nil, false
		}
		read[item] = true
		if code := result.code(); !succeeded(code) {
			failed = append(failed, ItemFailure{Index: item, ID: name, Status: code, Message: errorMessage(raw)})
		}
	}
	return failed, true
}

func isStatus(code int) bool { return code >= 100 && code <= 599 }

func sharedClass(failures []ItemFailure) error {
	class := writeErrorClass(failures[0].status())
	for _, failure := range failures[1:] {
		if !errors.Is(writeErrorClass(failure.status()), class) {
			return nil
		}
	}
	return class
}

// Split returns the writes that send w: w alone, unless w is a bulk write
// (Items) that may be sent in batches (Batched) whose body is a JSON array of
// more items than Emporix takes in one request; then a copy of w for each
// batch of up to that many of them, in order, which sends them as given and
// names the failures of its items by their position among those of w
// (ItemFailure.Index). A body that Prepare refuses is left whole, for
// Prepare to refuse.
func (w *Write) Split() []*Write {
	if !w.Batched || w.Parts == nil || w.Parts.member != "" {
		return []*Write{w}
	}
	most := w.Parts.most
	items, isArray := within(w.Body.data, "")
	if !isArray || len(items) <= most {
		return []*Write{w}
	}
	batches := make([]*Write, 0, (len(items)+most-1)/most)
	for first := 0; first < len(items); first += most {
		batch := *w
		batch.Body = Body{data: jsonArray(items[first:min(first+most, len(items))]), contentType: JSONContentType}
		batch.first = first
		batches = append(batches, &batch)
	}
	return batches
}

func jsonArray(items []json.RawMessage) []byte {
	array := []byte{'['}
	for i, item := range items {
		if i > 0 {
			array = append(array, ',')
		}
		array = append(array, item...)
	}
	return append(array, ']')
}

// JoinItems returns the failures of the batches of one bulk write (Split),
// of items in all, as one *ItemsError, named as the first of failed names
// the write: their failures in order, matching the class that every one of
// them shares, if any (as an ItemsError of one request does); nil where none
// of them failed.
func JoinItems(items int, failed []*ItemsError) error {
	if len(failed) == 0 {
		return nil
	}
	joined := &ItemsError{Method: failed[0].Method, Path: failed[0].Path, Noun: failed[0].Noun, Items: items}
	for _, batch := range failed {
		joined.Failed = append(joined.Failed, batch.Failed...)
	}
	joined.class = sharedClass(joined.Failed)
	return joined
}

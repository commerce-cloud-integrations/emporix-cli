package emporix

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Parts says that a write sends several parts in one JSON body, which Emporix
// carries out one by one and answers with the result of each: the items of a
// bulk write (Items) or the commands of a chain (Commands). Prepare refuses a
// body without one to most parts, and Send reads the result of each part: an
// *ItemsError lists those that failed.
type Parts struct {
	noun    string // what a part is called in messages, e.g. "item"
	sender  string // what sends the parts in messages, e.g. "a bulk write"
	most    int    // the most parts that Emporix takes in one request
	member  string // the member of the body's JSON object that holds the parts; the body is their array when empty
	results string // the member of the response's JSON object that holds their results; alike
	key     string // the member of each item and of its result that names the item (Items); none when empty
	// read reads results, the results of the parts of sent: the parts that
	// failed, in the order of the response, and how many did not run; false
	// when results do not give the result of each part that ran.
	read func(sent *sentParts, results []json.RawMessage) (failed []ItemFailure, notRun int, ok bool)
}

// What the parts of a bulk write and of a chain are called in messages, as
// ItemsError.Noun names them.
const (
	ItemNoun    = "item"
	CommandNoun = "command"
)

// Items returns the Parts of a bulk write: a JSON array of one to most items,
// which Emporix answers with a JSON array of the result of each item
// (readItems), named by its index, else by its position. key, unless empty,
// names instead the member of each item and of its result that names the
// item where the results give neither index nor id, such as eventType: a
// result is then its item's by that member, whatever their order, and a
// failure names its item by it.
func Items(most int, key string) *Parts {
	return &Parts{noun: ItemNoun, sender: "a bulk write", most: most, key: key, read: readItems}
}

// Commands returns the Parts of a chain of commands: a JSON object whose
// member commands is an array of one to most commands, which Emporix answers
// with a JSON object whose member results is an array of the results of the
// commands that ran (readCommands).
func Commands(most int) *Parts {
	return &Parts{noun: CommandNoun, sender: "a chain", most: most, member: "commands", results: "results",
		read: readCommands}
}

// sentParts is a body of parts as Prepare checked it, whose results Send
// reads.
type sentParts struct {
	*Parts
	parts []json.RawMessage // the parts as sent
	query url.Values        // the query sent with them, such as the onError of a chain
	first int               // the position of the first among the items of the write split (Write.Split); else 0
}

// split returns the parts of body, sent with query by the request that named
// names, or ErrInvalidInput when it does not hold one to most of them; nil
// for no Parts, a write of one value.
func (p *Parts) split(named string, body []byte, query url.Values) (*sentParts, error) {
	if p == nil {
		return nil, nil
	}
	parts, ok := within(body, p.member)
	if !ok || len(parts) == 0 {
		shape := "a JSON array"
		if p.member != "" {
			shape = "a JSON object whose " + p.member + " is an array"
		}
		return nil, fmt.Errorf("%w: %s: %s sends %s of one or more %ss", ErrInvalidInput, named, p.sender, shape,
			p.noun)
	}
	if len(parts) > p.most {
		return nil, fmt.Errorf("%w: %s: %s sends at most %d %ss in one request, not %d: split them",
			ErrInvalidInput, named, p.sender, p.most, p.noun, len(parts))
	}
	return &sentParts{Parts: p, parts: parts, query: query}, nil
}

// within returns the JSON array that data is or, with member, that the member
// of its JSON object holds (named exactly, as Emporix names it); false for
// anything else.
func within(data []byte, member string) ([]json.RawMessage, bool) {
	if member != "" {
		var object map[string]json.RawMessage
		if json.Unmarshal(data, &object) != nil {
			return nil, false
		}
		data = object[member]
	}
	var array []json.RawMessage
	return array, json.Unmarshal(data, &array) == nil && array != nil
}

// check returns the body of result, the response to s sent as method, nil
// when it is empty, and how it failed, with failure, the error of a response
// that failed (nil for a success), which may still give the result of each
// part, as the Cart Service answers a batch of items that failed in part with
// 400 or 500: an *ItemsError when a part failed; failure where the response
// failed without the result of each part; an error where it succeeded
// without them, as then which ones Emporix carried out is unknown; else
// failure (where every part succeeded, yet the response failed).
func (s *sentParts) check(method string, result reply, failure error) (json.RawMessage, error) {
	body, path := responseBody(result), result.path
	results, given := within(body, s.results)
	var failed []ItemFailure
	notRun, ok := 0, false
	if given {
		failed, notRun, ok = s.read(s, results)
	}
	for i := range failed {
		failed[i].Index += s.first // its position in the write that was split, if any
	}
	switch {
	case failure != nil && !ok:
		return nil, failure
	case !ok:
		return body, fmt.Errorf("%s %s: the response does not give the result of each of the %d %ss, which "+
			"may have been carried out all the same: read them before you send them again", method, path,
			len(s.parts), s.noun)
	case len(failed) > 0:
		return body, &ItemsError{Method: method, Path: path, Noun: s.noun, Items: len(s.parts), NotRun: notRun,
			Failed: failed, class: sharedClass(failed)}
	default:
		return body, failure
	}
}

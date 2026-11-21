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
	noun    string
	sender  string
	most    int
	member  string
	results string
	key     string
	read    func(sent *sentParts, results []json.RawMessage) (failed []ItemFailure, notRun int, ok bool)
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

type sentParts struct {
	*Parts
	parts []json.RawMessage
	query url.Values
	first int
}

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

// The Cart Service answers a batch of items that failed in part with 400 or
// 500 and still gives the result of each part.
func (s *sentParts) check(method string, result reply, failure error) (json.RawMessage, error) {
	body, path := responseBody(result), result.path
	results, given := within(body, s.results)
	var failed []ItemFailure
	notRun, ok := 0, false
	if given {
		failed, notRun, ok = s.read(s, results)
	}
	for i := range failed {
		failed[i].Index += s.first
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

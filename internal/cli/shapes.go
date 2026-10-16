package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// shape returns the items of a response, failing on a response of another
// shape (also for -o json, which prints the response whole). A shape may
// restructure the items (category trees flatten theirs).
type shape func(body json.RawMessage) ([]json.RawMessage, error)

// itemShape returns s, or oneItem when s is nil: how a read of one item that
// declares no shape finds the item in its response.
func itemShape(s shape) shape {
	if s == nil {
		return oneItem
	}
	return s
}

// listItems returns the items of a list, a JSON array.
func listItems(body json.RawMessage) ([]json.RawMessage, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil || items == nil {
		return nil, errors.New("expected a JSON array")
	}
	return items, nil
}

// jsonObject returns the members of body, a JSON object, by name; any other
// JSON, null included, is not one.
func jsonObject(body json.RawMessage) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil || object == nil {
		return nil, errors.New("expected a JSON object")
	}
	return object, nil
}

// oneItem returns the item that get reads: a JSON object with members.
func oneItem(body json.RawMessage) ([]json.RawMessage, error) {
	object, err := jsonObject(body)
	if err != nil {
		return nil, errors.New("expected one item as a JSON object")
	}
	if len(object) == 0 {
		return nil, errors.New("expected one item, got an empty object")
	}
	return []json.RawMessage{body}, nil
}

// oneObject returns the item of a response that is one JSON object, even an
// empty one, such as statistics with nothing to report.
func oneObject(body json.RawMessage) ([]json.RawMessage, error) {
	if _, err := jsonObject(body); err != nil {
		return nil, err
	}
	return []json.RawMessage{body}, nil
}

// namedValue is a member of a JSON object as a table row: its name and value
// (members).
type namedValue struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// members returns the members of a response that is one JSON object as items
// (namedValue), sorted by name; an empty object has none. Each value is kept
// as sent (json.Marshal would escape <, > and & in it).
func members(body json.RawMessage) ([]json.RawMessage, error) {
	object, err := jsonObject(body)
	if err != nil {
		return nil, err
	}
	items := make([]json.RawMessage, 0, len(object))
	for _, name := range slices.Sorted(maps.Keys(object)) {
		quoted, err := json.Marshal(name)
		if err != nil {
			return nil, err
		}
		items = append(items, slices.Concat([]byte(`{"name":`), quoted, []byte(`,"value":`), object[name], []byte(`}`)))
	}
	return items, nil
}

// marshalEach returns each value encoded as JSON: the items of a shape that
// restructures a response.
func marshalEach[T any](values []T) ([]json.RawMessage, error) {
	items := make([]json.RawMessage, len(values))
	for i, value := range values {
		item, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}
	return items, nil
}

// oneNumber returns the item of a response that is one JSON number, such as
// a customer's reward points. It checks the first byte because decoding into
// a json.Number would also accept a quoted number such as "5".
func oneNumber(body json.RawMessage) ([]json.RawMessage, error) {
	value := bytes.TrimSpace(body)
	if !json.Valid(value) || (value[0] != '-' && (value[0] < '0' || value[0] > '9')) {
		return nil, errors.New("expected a JSON number")
	}
	return []json.RawMessage{value}, nil
}

// oneItemOrArray is oneItem for a get that the API reference documents as a
// one-item array (price models, storefront payment modes) while Emporix has
// been seen to return the object itself: it accepts both.
func oneItemOrArray(body json.RawMessage) ([]json.RawMessage, error) {
	var items []json.RawMessage
	if json.Unmarshal(body, &items) != nil || items == nil {
		return oneItem(body)
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("expected one item, got %d", len(items))
	}
	return oneItem(items[0])
}

// textAsJSON returns answer, the body of a response that Emporix labels JSON
// although it may be bare text, as JSON: JSON as sent, else the text,
// trimmed, as a JSON string, or nil where it is blank.
func textAsJSON(answer []byte) (json.RawMessage, error) {
	if json.Valid(answer) {
		return answer, nil
	}
	text := strings.TrimSpace(string(answer))
	if text == "" {
		return nil, nil
	}
	return json.Marshal(text)
}

// oneName returns the answer of a read that is a name, a JSON string that is
// not blank, as its item.
func oneName(body json.RawMessage) ([]json.RawMessage, error) {
	var name *string
	if json.Unmarshal(body, &name) != nil || name == nil || strings.TrimSpace(*name) == "" {
		return nil, errors.New("expected a name")
	}
	return []json.RawMessage{body}, nil
}

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

type shape func(body json.RawMessage) ([]json.RawMessage, error)

func itemShape(s shape) shape {
	if s == nil {
		return oneItem
	}
	return s
}

func listItems(body json.RawMessage) ([]json.RawMessage, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil || items == nil {
		return nil, errors.New("expected a JSON array")
	}
	return items, nil
}

func jsonObject(body json.RawMessage) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil || object == nil {
		return nil, errors.New("expected a JSON object")
	}
	return object, nil
}

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

func oneObject(body json.RawMessage) ([]json.RawMessage, error) {
	if _, err := jsonObject(body); err != nil {
		return nil, err
	}
	return []json.RawMessage{body}, nil
}

type namedValue struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

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

func oneName(body json.RawMessage) ([]json.RawMessage, error) {
	var name *string
	if json.Unmarshal(body, &name) != nil || name == nil || strings.TrimSpace(*name) == "" {
		return nil, errors.New("expected a name")
	}
	return []json.RawMessage{body}, nil
}

package cli

import (
	"encoding/json"
	"errors"
	"strconv"
)

func answerView(singular, column string) view[bool] {
	return view[bool]{
		singular: singular,
		header:   []string{column},
		row:      func(passed bool, _ language) []string { return []string{strconv.FormatBool(passed)} },
	}
}

func oneBoolean(body json.RawMessage) ([]json.RawMessage, error) {
	var answer *bool
	if json.Unmarshal(body, &answer) != nil || answer == nil {
		return nil, errors.New("expected true or false")
	}
	return []json.RawMessage{body}, nil
}

func statusCheck() single { return single{items: oneBoolean, read: emptyAs("true")} }

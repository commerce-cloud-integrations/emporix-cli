package cli

import (
	"encoding/json"
	"errors"
	"strconv"
)

// The reads of checks that Emporix serves as POST without changing data, such
// as whether a delivery window or a coupon can still be used: their answer is
// true or false, and a check that fails is an error that names why.

// answerView shows the answer of a check, true or false, in one column.
func answerView(singular, column string) view[bool] {
	return view[bool]{
		singular: singular,
		header:   []string{column},
		row:      func(passed bool, _ language) []string { return []string{strconv.FormatBool(passed)} },
	}
}

// oneBoolean returns the answer of a check that is a JSON boolean as its item.
func oneBoolean(body json.RawMessage) ([]json.RawMessage, error) {
	var answer *bool
	if json.Unmarshal(body, &answer) != nil || answer == nil {
		return nil, errors.New("expected true or false")
	}
	return []json.RawMessage{body}, nil
}

// statusCheck reads a check that Emporix answers with its status alone
// (emporix.Client.Check), such as whether a coupon can be redeemed: a success
// without a body is the answer true, as if Emporix had sent it, so that
// -o json prints it, and an answer with a body is read as a check that
// answers true or false (oneBoolean) is: anything else fails.
func statusCheck() single { return single{items: oneBoolean, read: emptyAs("true")} }

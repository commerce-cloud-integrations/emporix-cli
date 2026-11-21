package emporix

import (
	"encoding/json"
	"errors"
	"fmt"
)

func outcome(method, path string, body []byte) error {
	var members map[string]json.RawMessage
	var successful *bool
	if json.Unmarshal(body, &members) != nil || json.Unmarshal(members["successful"], &successful) != nil ||
		successful == nil {
		return fmt.Errorf("%s %s: the response does not say whether it succeeded (successful true or false), so %s",
			method, path, carriedOutNote)
	}
	if *successful {
		return nil
	}
	text := fmt.Sprintf("%s %s: Emporix answered that it failed (successful false)", method, path)
	if message := errorMessage(body); message != "" {
		text += ": " + message
	}
	return errors.New(text)
}

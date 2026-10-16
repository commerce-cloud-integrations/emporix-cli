package emporix

import (
	"encoding/json"
	"errors"
	"fmt"
)

// A write whose response says whether it was carried out
// (Write.ReportsOutcome) is answered 2xx with a JSON object such as
// {"successful": false, "message": "…"}, as the Payment Gateway answers a
// payment operation that the payment provider declined: successful says
// whether it was carried out, and message, where it was not, why not.

// outcome checks the response body of a write that reports its outcome, sent
// as method to path: nil where it says that the write was carried out
// (successful true); an error with Emporix's message, if any, bounded as an
// APIError's, where it says that it was not (successful false); and an error
// where it does not say (no JSON object with a boolean successful), as the
// write may then have been carried out all the same.
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

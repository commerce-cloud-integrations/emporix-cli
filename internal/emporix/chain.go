package emporix

import (
	"encoding/json"
)

// A chain of commands, such as the Cart Service's execute, sends a JSON
// object whose member commands is an array of commands, which Emporix runs
// one after another: it answers 207 Multi-Status with the result of each
// command that ran, in order, such as
// {"results": [{"index": 0, "type": "AddCartItem", "code": 201, "status": "Created", "data": {…}}]},
// where data holds the error of a command that failed. Without
// onError=resume, Emporix stops after the first command that failed: those
// after it do not run. A command whose data is a list of items (a batch, such
// as AddCartItemsBatch) is answered with the result of each item as data, as
// a bulk write is, also with a code that succeeded (207) when items failed.

const (
	onErrorParam  = "onError"
	resumeOnError = "resume"
)

type commandResult struct {
	Index *int            `json:"index"`
	Type  string          `json:"type"`
	Code  int             `json:"code"`
	Data  json.RawMessage `json:"data"`
}

func readCommands(sent *sentParts, results []json.RawMessage) ([]ItemFailure, int, bool) {
	commands, ran := len(sent.parts), len(results)
	resume := sent.query.Get(onErrorParam) == resumeOnError
	if ran == 0 || ran > commands || resume && ran < commands {
		return nil, 0, false
	}
	var failed []ItemFailure
	stopped := false
	for position, raw := range results {
		var result commandResult
		if stopped || json.Unmarshal(raw, &result) != nil || !result.at(position) {
			return nil, 0, false
		}
		failures, ok := result.failures(position, sent.parts[position])
		if !ok {
			return nil, 0, false
		}
		failed = append(failed, failures...)
		stopped = !resume && !succeeded(result.Code)
	}
	if ran < commands && !stopped {
		return nil, 0, false
	}
	return failed, commands - ran, true
}

func (r commandResult) at(position int) bool {
	return r.Index != nil && *r.Index == position && isStatus(r.Code)
}

func (r commandResult) failures(position int, command json.RawMessage) ([]ItemFailure, bool) {
	items, isBatch := within(command, "data")
	entries, isList := within(r.Data, "")
	switch {
	case isBatch && len(items) > 0 && isList:
		return r.itemFailures(position, namedItems(items, ""), entries)
	case isBatch && len(items) > 0 && succeeded(r.Code):
		return nil, false
	case succeeded(r.Code):
		return nil, true
	default:
		return []ItemFailure{r.failure(position)}, true
	}
}

func (r commandResult) itemFailures(position int, items sentItems, entries []json.RawMessage) ([]ItemFailure,
	bool) {
	failed, ok := itemFailures(items, entries)
	switch {
	case !ok:
		return nil, false
	case len(failed) == 0 && !succeeded(r.Code):
		return []ItemFailure{r.failure(position)}, true
	}
	lines := make([]ItemFailure, len(failed))
	for i := range failed {
		lines[i] = ItemFailure{Index: position, Type: r.Type, Status: r.Code, Item: &failed[i]}
	}
	return lines, true
}

func (r commandResult) failure(position int) ItemFailure {
	return ItemFailure{Index: position, Type: r.Type, Status: r.Code, Message: errorMessage(r.Data)}
}

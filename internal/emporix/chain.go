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

// The query parameter of a chain, and its value, with which Emporix runs
// every command, also after one that failed.
const (
	onErrorParam  = "onError"
	resumeOnError = "resume"
)

// commandResult is what readCommands reads of the result of one command of
// a chain.
type commandResult struct {
	Index *int            `json:"index"` // the command's position in the body sent, from 0
	Type  string          `json:"type"`
	Code  int             `json:"code"`
	Data  json.RawMessage `json:"data"` // the response of the command, its error where it failed
}

// readCommands reads the results of the commands of a chain, sent: the
// result of each command that ran, in order, each naming it by its index:
// with onError=resume, of every command; else up to the first that failed,
// the last that ran then. False for anything else, as then which commands
// ran is unknown.
func readCommands(sent *sentParts, results []json.RawMessage) ([]ItemFailure, int, bool) {
	commands, ran := len(sent.parts), len(results)
	resume := sent.query.Get(onErrorParam) == resumeOnError
	if ran == 0 || ran > commands || resume && ran < commands {
		return nil, 0, false
	}
	var failed []ItemFailure
	stopped := false // a command failed, after which Emporix ran none without resume
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
		return nil, 0, false // the chain stopped although no command failed
	}
	return failed, commands - ran, true
}

// at reports whether r is the result of the command at position: it names
// that index and gives a status code.
func (r commandResult) at(position int) bool {
	return r.Index != nil && *r.Index == position && isStatus(r.Code)
}

// failures returns the failures of r, the result of command, sent at
// position: none when it succeeded; for a command whose data is a list of
// items (batchItems), whose result then gives the result of each item as
// data, the items that failed; else the command. False when the result of a
// batch that succeeded does not give the result of each item, as then which
// ones Emporix carried out is unknown.
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

// itemFailures returns the failures of r, the result of a command at
// position that sends items, whose data, entries, gives the result of each
// (itemFailures): each item that failed, or the command where it failed
// although none of them did; false when entries do not give the result of
// each item.
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

// failure returns the failure of r, the result of the command at position
// that failed, with the message of its data.
func (r commandResult) failure(position int) ItemFailure {
	return ItemFailure{Index: position, Type: r.Type, Status: r.Code, Message: errorMessage(r.Data)}
}

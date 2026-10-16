package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// The answers of writes that start jobs, as --wait finds the jobs in them
// (jobWait.answer).

// answering returns how --wait follows a job of j that a write starts whose
// answer gives the jobs that it started as answer does.
func (j *jobWait) answering(answer jobAnswer) *jobWait {
	answered := *j
	answered.answer = answer
	return &answered
}

// jobAnswer is how the answer of a write gives the jobs that it started:
// holding the one job (heldJob), or naming them by their IDs alone (jobIDs).
type jobAnswer interface {
	// started returns the jobs that body, the answer, gives, in order: none
	// where the write started none.
	started(body json.RawMessage) ([]startedJob, error)
	// printed returns what --wait prints of the jobs that body gave, each as
	// last read in last, in order: nil for a job never read.
	printed(last []*jobRead, body json.RawMessage) json.RawMessage
	// help says how --wait finds the jobs in the answer, to follow the help of
	// the write, done being the status of a job that ended well; none where
	// that help says it all.
	help(done string) string
}

// startedJob is a job that a write started, as its answer gives it: its ID,
// what it is of where the answer names it (e.g. "vehicle/vehicles", the
// TYPE/ID of an index), and the job as first read where the answer holds it.
type startedJob struct {
	id    string
	of    string
	first *jobRead // nil where the answer names the job alone
}

// name returns how messages name j, e.g. "j1" or "j1 of vehicle/vehicles".
func (j *startedJob) name() string {
	if j.of == "" {
		return j.id
	}
	return j.id + " of " + j.of
}

// heldJob is the answer of a write that holds the one job that it started,
// whose state it takes.
type heldJob func(body json.RawMessage) (jobState, error)

// started implements jobAnswer: the job as the answer holds it.
func (h heldJob) started(body json.RawMessage) ([]startedJob, error) {
	first, err := readJob(h, body)
	if err != nil {
		return nil, err
	}
	return []startedJob{{id: first.state.ID, first: &first}}, nil
}

// printed implements jobAnswer: the job as last read (oneJob).
func (heldJob) printed(last []*jobRead, body json.RawMessage) json.RawMessage {
	return oneJob(last, body)
}

// help implements jobAnswer: the help of the write says it all.
func (heldJob) help(string) string { return "" }

// oneJob returns what --wait prints of the one job that a write started, as
// last read: the job, or body, the answer, where it was never read or the
// write started none.
func oneJob(last []*jobRead, body json.RawMessage) json.RawMessage {
	if len(last) == 1 && last[0] != nil {
		return last[0].raw
	}
	return body
}

// jobIDs is the answer of a write that names the jobs that it started by
// their IDs alone, each in field (a jobAnswer): one JSON object, which names
// the one job, unless list; else a JSON array with an item per thing that
// the write changed, which its members of name, each naming the job that it
// started, or none, without field or with null. An empty answer, such as 204
// No Content, names no job; any other must name its jobs so.
type jobIDs struct {
	field string   // the member that holds the ID of a job, e.g. "id"
	list  bool     // the answer is a JSON array of items, each naming a job or none; else one object that names one
	of    []string // of a list, the members of each item, strings, that name what it changed, e.g. "type", "id"
}

// started implements jobAnswer.
func (n jobIDs) started(body json.RawMessage) ([]startedJob, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}
	items, err := n.items(body)
	if err != nil {
		return nil, fmt.Errorf("decoding the response: %w", err)
	}
	var jobs []startedJob
	for _, item := range items {
		job, err := n.job(item)
		if err != nil {
			return nil, fmt.Errorf("decoding the response: %w", err)
		}
		if job.id != "" {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// items returns the items of body, the answer: its items, or it, an object.
func (n jobIDs) items(body json.RawMessage) ([]json.RawMessage, error) {
	if n.list {
		return listItems(body)
	}
	return oneObject(body)
}

// job returns the job that item names, of what its members of name; one
// without an ID for an item of a list that names none.
func (n jobIDs) job(item json.RawMessage) (startedJob, error) {
	object, err := jsonObject(item)
	if err != nil {
		return startedJob{}, err
	}
	of := make([]string, len(n.of))
	for i, member := range n.of {
		if json.Unmarshal(object[member], &of[i]) != nil || of[i] == "" {
			return startedJob{}, errors.New("an item has no " + member)
		}
	}
	notAJob := fmt.Errorf("%s is not the ID of a job", n.field)
	var id *string
	if raw, names := object[n.field]; names && json.Unmarshal(raw, &id) != nil {
		return startedJob{}, notAJob
	}
	switch {
	case id == nil && n.list:
		return startedJob{}, nil // an item that started no job
	case id == nil || !isPathSegment(*id):
		return startedJob{}, notAJob
	}
	return startedJob{id: *id, of: strings.Join(of, "/")}, nil
}

// printed implements jobAnswer: the one job, as oneJob prints it; of a list,
// the jobs as last read, in order, as a JSON array, those never read left
// out.
func (n jobIDs) printed(last []*jobRead, body json.RawMessage) json.RawMessage {
	if !n.list {
		return oneJob(last, body)
	}
	read := make([][]byte, 0, len(last))
	for _, job := range last {
		if job != nil {
			read = append(read, job.raw)
		}
	}
	return slices.Concat([]byte("["), bytes.Join(read, []byte(",")), []byte("]"))
}

// help implements jobAnswer.
func (n jobIDs) help(done string) string {
	if !n.list {
		return "\n\nAs Emporix answers with the id of the job alone, --wait reads the job by that id first.\n" +
			"Where Emporix answers with none, having started no job, --wait prints the response and says\n" +
			"so (exit 0); where no read of the job succeeds, it prints the response."
	}
	of := strings.Join(n.of, "/")
	return "\n\nAs Emporix answers with the id of each job alone, in a result per item that names its\n" +
		of + ", --wait reads each job by its id first and follows the jobs in turn, all within\n" +
		"--wait-timeout, naming each by that " + of + ". It prints the jobs as last read in place of\n" +
		"the results, as a JSON array in their order, leaving out each job never read (standard\n" +
		"error says why), and exits 0 only if each ended " + done + "; where jobs failed with different\n" +
		"exit codes, 1. Where Emporix started no job, it prints [] and says so (exit 0)."
}

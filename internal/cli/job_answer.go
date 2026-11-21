package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

func (j *jobWait) answering(answer jobAnswer) *jobWait {
	answered := *j
	answered.answer = answer
	return &answered
}

type jobAnswer interface {
	started(body json.RawMessage) ([]startedJob, error)
	printed(last []*jobRead, body json.RawMessage) json.RawMessage
	help(done string) string
}

type startedJob struct {
	id    string
	of    string
	first *jobRead
}

func (j *startedJob) name() string {
	if j.of == "" {
		return j.id
	}
	return j.id + " of " + j.of
}

type heldJob func(body json.RawMessage) (jobState, error)

func (h heldJob) started(body json.RawMessage) ([]startedJob, error) {
	first, err := readJob(h, body)
	if err != nil {
		return nil, err
	}
	return []startedJob{{id: first.state.ID, first: &first}}, nil
}

func (heldJob) printed(last []*jobRead, body json.RawMessage) json.RawMessage {
	return oneJob(last, body)
}

func (heldJob) help(string) string { return "" }

func oneJob(last []*jobRead, body json.RawMessage) json.RawMessage {
	if len(last) == 1 && last[0] != nil {
		return last[0].raw
	}
	return body
}

type jobIDs struct {
	field string
	list  bool
	of    []string
}

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

func (n jobIDs) items(body json.RawMessage) ([]json.RawMessage, error) {
	if n.list {
		return listItems(body)
	}
	return oneObject(body)
}

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
		return startedJob{}, nil
	case id == nil || !isPathSegment(*id):
		return startedJob{}, notAJob
	}
	return startedJob{id: *id, of: strings.Join(of, "/")}, nil
}

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

func (n jobIDs) help(done string) string {
	if !n.list {
		return "As Emporix answers with the id of the job alone, --wait reads the job by that id first. " +
			"Where Emporix answers with none, having started no job, --wait prints the response and says " +
			"so (exit 0); where no read of the job succeeds, it prints the response."
	}
	of := strings.Join(n.of, "/")
	return "As Emporix answers with the id of each job alone, in a result per item that names its " +
		of + ", --wait reads each job by its id first and follows the jobs in turn, all within " +
		"--wait-timeout, naming each by that " + of + ". It prints the jobs as last read in place of " +
		"the results, as a JSON array in their order, leaving out each job never read (standard " +
		"error says why), and exits 0 only if each ended " + done + "; where jobs failed with different " +
		"exit codes, 1. Where Emporix started no job, it prints [] and says so (exit 0)."
}

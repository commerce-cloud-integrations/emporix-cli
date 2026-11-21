package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

type jobWait struct {
	singular string
	get      request
	scopes   []string
	getLine  string
	decode   func(body json.RawMessage) (jobState, error)
	answer   jobAnswer
	running  []string
	done     string
	shortest time.Duration
}

type jobState struct {
	ID       string
	Status   string
	Progress string
	Failure  string
}

const minWaitDuration = time.Second

func followJob[T any](r resource[T], state func(*T) jobState, running []string, done string) *jobWait {
	get := r.itemRequest()
	decode := jobDecoder(r, r.item, state)
	return &jobWait{
		singular: r.singular, get: get, scopes: r.scopes, getLine: r.commandLine("get", get),
		running: running, done: done, shortest: minWaitDuration, decode: decode, answer: heldJob(decode),
	}
}

func jobDecoder[T any](r resource[T], item shape, state func(*T) jobState) func(body json.RawMessage) (jobState,
	error) {
	return func(body json.RawMessage) (jobState, error) {
		items, err := itemShape(item)(body)
		if err == nil && len(items) != 1 {
			err = fmt.Errorf("expected one item, got %d", len(items))
		}
		if err != nil {
			return jobState{}, r.decodeError(err)
		}
		var job T
		if err := json.Unmarshal(items[0], &job); err != nil {
			return jobState{}, r.decodeError(err)
		}
		if s := state(&job); isPathSegment(s.ID) && s.Status != "" {
			return s, nil
		}
		return jobState{}, r.decodeError(errors.New("no id or status"))
	}
}

const (
	waitFlag         = "wait"
	waitTimeoutFlag  = "wait-timeout"
	pollIntervalFlag = "poll-interval"
)

const (
	defaultWaitTimeout  = 30 * time.Minute
	defaultPollInterval = 5 * time.Second
)

type waiting struct {
	job      *jobWait
	timeout  waitDuration
	interval waitDuration
}

func (j *jobWait) bind(_ *app, cmd *cobra.Command) endingPlanner {
	refuseTypeScopes(cmd.Use, j.scopes)
	w := &waiting{job: j, timeout: waitDuration{value: defaultWaitTimeout, shortest: j.shortest},
		interval: waitDuration{value: defaultPollInterval, shortest: j.shortest}}
	least := ", " + j.shortest.String() + " or more"
	cmd.Flags().Bool(waitFlag, false, "wait until the "+j.singular+" ends, then print it")
	cmd.Flags().Var(&w.timeout, waitTimeoutFlag, "how long --"+waitFlag+" waits"+least)
	cmd.Flags().Var(&w.interval, pollIntervalFlag, "how often --"+waitFlag+" reads the "+j.singular+least)
	onlyWithFlag(cmd, waitTimeoutFlag, waitFlag)
	onlyWithFlag(cmd, pollIntervalFlag, waitFlag)
	cmd.Long += fmt.Sprintf("\n\n--%s then reads the job again every --%s, as `epx %s` "+
		"does and with %s too, saying on standard error how its status "+
		"changes, until it ends: then it prints the job and exits 0 if it ended %s, else 1. "+
		"When --%s passes first, it prints the job as last read and exits 1; the job goes "+
		"on. When a read of the job fails, it prints the job as last read and exits as that read "+
		"does, e.g. 4 when the job is not found.",
		waitFlag, pollIntervalFlag, j.getLine, theScopes(j.scopes), j.done, waitTimeoutFlag)
	cmd.Long = joinParagraphs(cmd.Long, j.answer.help(j.done))
	return w
}

func (w *waiting) plan(cmd *cobra.Command, write []string) ([]string, answered) {
	if !flagOn(cmd.Flags(), waitFlag) {
		return printing{}.plan(cmd, write)
	}
	read := slices.DeleteFunc(slices.Clone(w.job.scopes), func(scope string) bool { return slices.Contains(write, scope) })
	return slices.Concat(write, read), func(client *emporix.Client, response []byte) (json.RawMessage, error) {
		return w.follow(cmd, client, response)
	}
}

var errWaitTimedOut = errors.New("the wait timed out")

func (w *waiting) follow(cmd *cobra.Command, client *emporix.Client, response json.RawMessage) (
	json.RawMessage, error) {
	jobs, err := w.job.answer.started(response)
	switch {
	case err != nil:
		return response, fmt.Errorf("sent, but its %s cannot be waited for: %w", w.job.singular, err)
	case len(jobs) == 0:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Emporix started no %s, so there is none to wait for.\n", w.job.singular)
		return w.job.answer.printed(nil, response), nil
	}
	ctx, cancel := context.WithTimeoutCause(cmd.Context(), w.timeout.value, errWaitTimedOut)
	defer cancel()
	last, outcomes := make([]*jobRead, len(jobs)), make([]error, len(jobs))
	for i := range jobs {
		read, err := w.poll(ctx, cmd.ErrOrStderr(), client, &jobs[i])
		last[i], outcomes[i] = read, w.outcome(ctx, &jobs[i], read, err)
	}
	return w.job.answer.printed(last, response), joinOutcomes(outcomes)
}

func joinOutcomes(outcomes []error) error {
	err := errors.Join(outcomes...)
	for _, outcome := range outcomes {
		if outcome != nil && exitCode(runError{outcome}) != exitCode(runError{err}) {
			return errors.New(err.Error())
		}
	}
	return err
}

func (w *waiting) outcome(ctx context.Context, job *startedJob, last *jobRead, err error) error {
	switch {
	case err == nil && last.state.Status == w.job.done:
		return nil
	case err == nil:
		return fmt.Errorf("%s %s ended with %s", w.job.singular, job.name(), last.state.outcome())
	case !errors.Is(context.Cause(ctx), errWaitTimedOut):
		return fmt.Errorf("waiting for %s %s: %w", w.job.singular, job.name(), err)
	case last == nil:
		return fmt.Errorf("%s %s was not read before --%s (%s) passed: `epx %s %s` reads it", w.job.singular,
			job.name(), waitTimeoutFlag, w.timeout.value, w.job.getLine, job.id)
	default:
		return fmt.Errorf("%s %s has not ended within %s (last %s); it goes on: `epx %s %s` reads it",
			w.job.singular, job.name(), w.timeout.value, last.state.Status, w.job.getLine, job.id)
	}
}

type jobRead struct {
	state jobState
	raw   json.RawMessage
}

func readJob(decode func(body json.RawMessage) (jobState, error), body json.RawMessage) (jobRead, error) {
	state, err := decode(body)
	return jobRead{state: state, raw: body}, err
}

func (w *waiting) poll(ctx context.Context, errOut io.Writer, client *emporix.Client, job *startedJob) (*jobRead,
	error) {
	watched := &watch{wait: w.job, errOut: errOut, job: job}
	path, query := w.job.get.fill([]string{job.id}, "")
	ended, err := watched.first(ctx, client, path, query)
	if !ended && err == nil {
		err = client.Poll(ctx, path, query, w.interval.value, watched.take)
	}
	return watched.last, err
}

type watch struct {
	wait   *jobWait
	errOut io.Writer
	job    *startedJob
	last   *jobRead
	said   string
}

func (v *watch) first(ctx context.Context, client *emporix.Client, path string, query url.Values) (bool, error) {
	if v.job.first != nil {
		return v.seen(v.job.first)
	}
	body, err := client.Get(ctx, path, query)
	if err != nil {
		return false, err
	}
	return v.take(body)
}

func (v *watch) take(body json.RawMessage) (bool, error) {
	read, err := readJob(v.wait.decode, body)
	if err != nil {
		return false, err
	}
	return v.seen(&read)
}

func (v *watch) seen(read *jobRead) (bool, error) {
	if read.state.ID != v.job.id {
		return false, fmt.Errorf("a read of it holds %s %s instead", v.wait.singular, read.state.ID)
	}
	v.last = read
	if state := read.state.String(); state != v.said {
		_, _ = fmt.Fprintf(v.errOut, "%s %s: %s\n", v.wait.singular, output.Sanitize(v.job.name()), output.Sanitize(state))
		v.said = state
	}
	return !slices.Contains(v.wait.running, read.state.Status), nil
}

func (s *jobState) String() string {
	if s.Progress == "" {
		return s.Status
	}
	return s.Status + " (" + s.Progress + ")"
}

func (s *jobState) outcome() string {
	if s.Failure == "" {
		return s.Status
	}
	return s.Status + ": " + s.Failure
}

type waitDuration struct {
	value    time.Duration
	shortest time.Duration
}

func (d *waitDuration) Set(value string) error {
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < d.shortest {
		return fmt.Errorf("use a duration of %s or more, such as 30s or 10m", d.shortest)
	}
	d.value = parsed
	return nil
}

func (d *waitDuration) String() string { return d.value.String() }

func (d *waitDuration) Type() string { return "duration" }

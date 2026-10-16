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

// jobWait declares the job that a write starts, which --wait follows: it
// reads the job again, as the read of one job does, until its status is no
// longer one of running, then prints it, and fails unless it ended with done.
// followJob makes one from the resource of the jobs, for a write whose answer
// holds the job; answering, one whose write answers otherwise.
type jobWait struct {
	singular string                                       // the job in messages, e.g. "reindex job"
	get      request                                      // the read of one job: {id} in its path
	scopes   []string                                     // the scopes of that read
	getLine  string                                       // the command line, without epx, of that read
	decode   func(body json.RawMessage) (jobState, error) // the state of the job that a read of it holds
	answer   jobAnswer                                    // how the write's answer gives the jobs that it started
	running  []string                                     // the statuses of a job that has not ended
	done     string                                       // the status of a job that ended well; it failed with another
	shortest time.Duration                                // the shortest --wait-timeout and --poll-interval
}

// jobState is what --wait follows of a job, whatever its type: its ID, which
// addresses it, its status, how far it got and why it failed, where it says.
type jobState struct {
	ID       string
	Status   string
	Progress string // e.g. "2000 of 5000"; none when the job does not say
	Failure  string // why it failed; none when the job does not say
}

// minWaitDuration is the shortest --wait-timeout and --poll-interval: a job
// is not read more often than once a second.
const minWaitDuration = time.Second

// followJob returns how --wait follows a job of r that a write starts: it
// reads the job as `get` of r does, and finds it in the write's answer as in
// a read (jobDecoder). A job whose status is one of running has not ended;
// one that ended with done succeeded. A write whose answer gives the job
// otherwise is answering it.
func followJob[T any](r resource[T], state func(*T) jobState, running []string, done string) *jobWait {
	get := r.itemRequest()
	decode := jobDecoder(r, r.item, state)
	return &jobWait{
		singular: r.singular, get: get, scopes: r.scopes, getLine: r.commandLine("get", get),
		running: running, done: done, shortest: minWaitDuration, decode: decode, answer: heldJob(decode),
	}
}

// jobDecoder returns the state of the job of r that a response holds: it
// finds the job as item finds an item (oneItem when nil), decodes it as a T
// and takes its state with state, which must give an ID and a status.
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

// The flags of a write that starts a job, which follow the job.
const (
	waitFlag         = "wait"
	waitTimeoutFlag  = "wait-timeout"
	pollIntervalFlag = "poll-interval"
)

// The defaults of --wait-timeout and --poll-interval.
const (
	defaultWaitTimeout  = 30 * time.Minute
	defaultPollInterval = 5 * time.Second
)

// waiting ends a write that starts a job: it follows the job with --wait,
// and prints the response otherwise.
type waiting struct {
	job      *jobWait
	timeout  waitDuration
	interval waitDuration
}

// bind implements answerer: it adds --wait, --wait-timeout and
// --poll-interval to cmd, a write that starts j, which the latter two apply
// only with (onlyWithFlag), and says in its help how --wait follows the job.
func (j *jobWait) bind(_ *app, cmd *cobra.Command) ending {
	refuseTypeScopes(cmd.Use, j.scopes)
	w := &waiting{job: j, timeout: waitDuration{value: defaultWaitTimeout, shortest: j.shortest},
		interval: waitDuration{value: defaultPollInterval, shortest: j.shortest}}
	least := ", " + j.shortest.String() + " or more"
	cmd.Flags().Bool(waitFlag, false, "wait until the "+j.singular+" ends, then print it")
	cmd.Flags().Var(&w.timeout, waitTimeoutFlag, "how long --"+waitFlag+" waits"+least)
	cmd.Flags().Var(&w.interval, pollIntervalFlag, "how often --"+waitFlag+" reads the "+j.singular+least)
	onlyWithFlag(cmd, waitTimeoutFlag, waitFlag)
	onlyWithFlag(cmd, pollIntervalFlag, waitFlag)
	cmd.Long += fmt.Sprintf("\n\n--%s then reads the job again every --%s, as `epx %s`\n"+
		"does and with %s too, saying on standard error how its status\n"+
		"changes, until it ends: then it prints the job and exits 0 if it ended %s, else 1.\n"+
		"When --%s passes first, it prints the job as last read and exits 1; the job goes\n"+
		"on. When a read of the job fails, it prints the job as last read and exits as that read\n"+
		"does, e.g. 4 when the job is not found.",
		waitFlag, pollIntervalFlag, j.getLine, theScopes(j.scopes), j.done, waitTimeoutFlag)
	cmd.Long += j.answer.help(j.done)
	return w
}

// plan implements ending: with --wait, the run also requests the scopes of
// reading the job, which then fail it before anything is sent when the
// credential lacks them, and follows the job; without, it prints the
// response.
func (w *waiting) plan(cmd *cobra.Command, write []string) ([]string, answered) {
	if !flagOn(cmd.Flags(), waitFlag) {
		return printing{}.plan(cmd, write)
	}
	read := slices.DeleteFunc(slices.Clone(w.job.scopes), func(scope string) bool { return slices.Contains(write, scope) })
	return slices.Concat(write, read), func(client *emporix.Client, response []byte) (json.RawMessage, error) {
		return w.follow(cmd, client, response)
	}
}

// errWaitTimedOut is the cause of the end of a wait that --wait-timeout ended.
var errWaitTimedOut = errors.New("the wait timed out")

// follow follows the jobs that response, the write's answer, gives
// (jobAnswer.started) with client, in turn, each until it ends or
// --wait-timeout passes, then returns them as last read, to be printed
// (jobAnswer.printed): it fails unless each ended with done. A response that
// gives no job, having started none, is printed as the answer prints it,
// saying so on standard error; one that cannot give the jobs, as sent. The
// sender prints what it returns also when it fails, with the outcome's class
// setting the exit code (sendThen).
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

// joinOutcomes returns the outcomes of the jobs of one wait as one error: nil
// where each succeeded, else their failures, in order, with the exit code of
// each where they share one, else with their messages alone (exit 1), as
// batchStopped keeps them.
func joinOutcomes(outcomes []error) error {
	err := errors.Join(outcomes...)
	for _, outcome := range outcomes {
		if outcome != nil && exitCode(runError{outcome}) != exitCode(runError{err}) {
			return errors.New(err.Error())
		}
	}
	return err
}

// outcome returns how a wait in ctx ends for job, last read as last (nil
// where it was never read), after a poll that ended with err: nil when the
// job ended with done, else why not.
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

// jobRead is one read of a job: its state and the response that holds it.
type jobRead struct {
	state jobState
	raw   json.RawMessage
}

// readJob returns the read of the job that body holds, whose state decode
// takes.
func readJob(decode func(body json.RawMessage) (jobState, error), body json.RawMessage) (jobRead, error) {
	state, err := decode(body)
	return jobRead{state: state, raw: body}, err
}

// poll follows job with client until it ended or ctx ends: it reads the job
// by its ID first where the answer named it alone, then again every
// --poll-interval (emporix.Client.Poll) while it runs, and returns it as last
// read, nil where no read of it succeeded.
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

// watch is a job that --wait follows as wait declares it: the job as last
// read, nil before its first read, and its state as last written to errOut.
type watch struct {
	wait   *jobWait
	errOut io.Writer
	job    *startedJob
	last   *jobRead
	said   string
}

// first takes the job as the answer holds it, else as client reads it now
// at path with query (take), and returns whether it ended.
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

// take takes the job that body, a read of it, holds (seen).
func (v *watch) take(body json.RawMessage) (bool, error) {
	read, err := readJob(v.wait.decode, body)
	if err != nil {
		return false, err
	}
	return v.seen(&read)
}

// seen takes read as the job's last read, writing its state to errOut where
// it changed, and returns whether the job ended; a read of another job fails.
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

// String returns the status of s, and how far it got where it says so, e.g.
// "IN_PROGRESS (2000 of 5000)".
func (s *jobState) String() string {
	if s.Progress == "" {
		return s.Status
	}
	return s.Status + " (" + s.Progress + ")"
}

// outcome returns the status of s, an ended job, with why it failed, if it
// says, e.g. "FAILURE: the index is not configured".
func (s *jobState) outcome() string {
	if s.Failure == "" {
		return s.Status
	}
	return s.Status + ": " + s.Failure
}

// waitDuration is the value of --wait-timeout or --poll-interval: a duration
// such as 30s or 10m, shortest or more; it implements pflag.Value, so another
// value is rejected while flags are parsed.
type waitDuration struct {
	value    time.Duration
	shortest time.Duration
}

// Set implements pflag.Value.
func (d *waitDuration) Set(value string) error {
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < d.shortest {
		return fmt.Errorf("use a duration of %s or more, such as 30s or 10m", d.shortest)
	}
	d.value = parsed
	return nil
}

func (d *waitDuration) String() string { return d.value.String() }

// Type implements pflag.Value.
func (d *waitDuration) Type() string { return "duration" }

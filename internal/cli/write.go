package cli

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// Write commands change data. Every one sends through a sender, which gives
// it the behavior that every write shares: its body from one source (--file,
// - for standard input, typed flags that --file replaces, or its arguments;
// none where Emporix takes none), --dry-run, which prints the request
// instead of sending it, a confirmation on a terminal that --yes gives
// instead (confirmation: for a delete, the identifier of what it deletes,
// typed), the exit code of a conflict and the response printed as JSON, or,
// with --wait, the job that it started, followed until it ends (jobWait).
// Write commands are named after what they send: `create` (POST to a
// collection), `update ID` (PUT, which replaces the item; some services also
// create a missing one), `patch ID` (PATCH), `delete ID`, `bulk-create`,
// `bulk-update`, `bulk-patch` and `bulk-delete`, an action's own verb, such
// as `rebuild`, `trigger` or `capture`, or `set` for one member of an item.
// A setting of which a tenant has one at most, such as the password migration
// retention, is set with `update` (sent as Emporix takes it: a POST that
// creates or replaces it) and removed with `delete`.

// deleteUse is the name of the write that deletes one item: a DELETE of it.
const deleteUse = "delete"

// createUse is the name of the write that creates an item: a POST to its
// collection.
const createUse = "create"

// importUse is the name of the write that imports items into a collection: what
// an export of the same service gives, or items made outside Emporix.
const importUse = "import"

// recalculate names a write that recalculates (products recalculate), and the
// query parameter and flag of the order writes that ask Emporix to
// recalculate the order.
const recalculate = "recalculate"

// fileFlag is the flag of a write that gives its body (payload).
const fileFlag = "file"

// maxWriteFileBytes bounds the body that --file reads, above the batches that
// Emporix takes in one request.
const maxWriteFileBytes = 16 << 20

// stdinFileNote ends the usage of --file: how standard input gives the file.
const stdinFileNote = " (- reads standard input; sending it then needs --" + yesFlag + ")"

// The flags of every write that decide whether it is sent.
const (
	dryRunFlag = "dry-run"
	yesFlag    = "yes"
)

// bodySource is where the body of a write comes from: --file or typed flags
// (payload), a member of which may repeat an argument of the request
// (argumentField), or nothing (noBody). Each write takes its body from one.
type bodySource interface {
	// bind adds the flags of the body to cmd and what its help says of the
	// body, and returns the flag that gives the body on standard input, e.g.
	// "--file -"; none where it never reads it.
	bind(cmd *cobra.Command) (stdin string)
	// scopes returns the OAuth scopes with which fill reads, if it reads.
	scopes() []string
	// fill sets the body of w, a write to w.Path, for values, the arguments
	// after those of its request: no body, a JSON value, or the parts of a
	// bulk write or chain (w.Parts). Values or a file that cannot be sent fail
	// as a usage error, before anything is sent.
	fill(ctx context.Context, w *emporix.Write, values []string) error
	// fromStdin reports whether fill reads standard input, which then cannot
	// confirm the write.
	fromStdin() bool
}

// payload is the body that --file gives a write, sent byte for byte
// (bodyFile): one JSON value, or the items of a bulk write; or, without
// --file, the JSON object that its typed flags make, if it has any. --file is
// required unless the body is optional or typed. Each write takes its own,
// &payload{} for one JSON value, required.
type payload struct {
	optional bool       // --file may be left out, as an `api` write then sends no body
	array    bool       // --file gives one JSON array, such as the operations of a patch (operationsFile)
	maxItems int        // a bulk write's: the most items of its JSON array, answered per item; 0 for one value
	itemKey  string     // a bulk write's: the member naming an item in it and its result (emporix.Items)
	whole    bool       // a bulk write's items are answered as a whole (a 204), not per item
	batched  bool       // a bulk write's: more items are sent in requests of up to maxItems (emporix.Write.Split)
	typed    bodyBinder // the body that typed flags make, each of which --file excludes; none when nil
	body     boundBody  // --file and the typed flags, once bound
}

// bind adds --file and the typed flags, if any, with --file required unless
// the body is optional or typed (then --file says that it replaces the body
// that epx makes), and for a bulk write says in the help how many items
// Emporix takes and how it answers.
func (p *payload) bind(cmd *cobra.Command) string {
	usage := "JSON request body in `PATH`, sent byte for byte"
	if p.typed != nil {
		usage += " instead of the one that epx makes"
	}
	p.body = bodyFile{flag: fileFlag, limit: maxWriteFileBytes, array: p.array}.bindWith(cmd, usage+stdinFileNote,
		p.typed)
	if !p.optional && p.typed == nil {
		requireFlag(cmd, fileFlag)
	}
	switch {
	case p.maxItems > 0 && p.whole:
		cmd.Long += "\n\nEmporix takes up to " + strconv.Itoa(p.maxItems) + " items in one request (epx refuses " +
			"more) and answers for\nall of them at once, not item by item."
	case p.maxItems > 0:
		cmd.Long += "\n\n" + p.itemsSentence() + "\nits own and answers with the result of each: when one failed, epx " +
			"prints the response,\nlists the items that failed on standard error (one that failed with 500, 502, 503 " +
			"or 504\nmay have been carried out all the same) and " + bulkExitSentence
	}
	return "--" + fileFlag + " -"
}

// bulkExitSentence says which exit code a bulk write or a chain gives when
// items failed: the code that their failures share.
const bulkExitSentence = "exits with the code that their failures\nshare (for example 5 for a 409 Conflict, 4 for " +
	"a 404), else 1."

// bulkItemsDefault is the most items that Emporix takes in one bulk request
// where the reference of the endpoint names no other limit, as its guide to
// bulk request limits says.
const bulkItemsDefault = 200

// itemsSentence says how many items of a bulk write Emporix takes in one
// request and how epx sends more, up to "carries out each on", which what
// Emporix does with them follows.
func (p *payload) itemsSentence() string {
	most := strconv.Itoa(p.maxItems)
	if p.batched {
		return "epx sends the items in requests of up to " + most + ", in order, confirmed once, and stops at a\n" +
			"request that fails as a whole, sending none after it. Emporix carries out each on"
	}
	return "Emporix takes up to " + most + " items in one request (epx refuses more), carries out each on"
}

func (*payload) scopes() []string { return nil }

func (p *payload) fill(ctx context.Context, w *emporix.Write, _ []string) error {
	if p.maxItems > 0 {
		w.Parts = emporix.Items(p.maxItems, p.itemKey)
	}
	w.AnsweredWhole, w.Batched = p.whole, p.batched
	body, err := p.body.make(ctx)
	w.Body = body
	return err
}

func (p *payload) fromStdin() bool { return p.body.file.fromStdin() }

// emptyObjectOrFile returns the body of a write whose members are all
// optional: the JSON object of --file, or without it {}, an object without
// members (the typed body of no flag). Where a reference marks such a body
// optional without saying what a request without one does, epx sends {}.
func emptyObjectOrFile() *payload { return &payload{typed: objectBody{}} }

// operationsFile returns the body of a patch: the JSON array of one or more
// operations of --file, such as JSON Patch (RFC 6902), which epx refuses
// otherwise, as a service may answer anything else as done without applying
// it (the Customer Segment Service does).
func operationsFile() *payload { return &payload{array: true} }

// noBody is the body of a write that Emporix takes no body for: it adds no
// flag and sends none.
type noBody struct{}

func (noBody) bind(*cobra.Command) string { return "" }

func (noBody) scopes() []string { return nil }

func (noBody) fill(context.Context, *emporix.Write, []string) error { return nil }

func (noBody) fromStdin() bool { return false }

// sender sends the writes of one command, with method, as every write is
// sent: with the body that its source gives and the flags that bind adds,
// once confirmed as its confirmation asks.
type sender struct {
	a        *app
	method   string
	body     bodySource
	confirm  confirmation
	conflict string // what a 409 Conflict means and how to recover (finish); that the item changed when empty
	dryRun   bool
	yes      bool
	format   *formatChoice // -o: how it prints the response
}

// bind adds to cmd the flags of its body and those of every write, --dry-run,
// --yes and -o for its response, which --dry-run excludes, and says in its
// help how it is sent and confirmed.
func (s *sender) bind(cmd *cobra.Command) {
	required := "required without a terminal"
	if stdin := s.body.bind(cmd); stdin != "" {
		required += ", or with " + stdin
	}
	cmd.Flags().BoolVar(&s.dryRun, dryRunFlag, false, "print the request instead of sending it (excludes -o)")
	cmd.Flags().BoolVar(&s.yes, yesFlag, false, "send without asking ("+required+")")
	s.format = addResponseOutput(cmd)
	cmd.MarkFlagsMutuallyExclusive(dryRunFlag, outputFlag)
	cmd.Long += "\n\n" + s.confirm.help(s.method, required)
}

// answered is what a write does once Emporix answered it without failing,
// with the response body as sent (nil for an empty one) and the client that
// sent it: it returns what the write then prints (writeResponse; nil for
// nothing), and an error that fails the write, which prints what it returned
// all the same.
type answered func(client *emporix.Client, response []byte) (json.RawMessage, error)

// asAnswered is what a write without an answerer does with Emporix's
// answer: it prints the response.
func asAnswered(_ *emporix.Client, response []byte) (json.RawMessage, error) {
	return response, nil
}

// send sends a write as sendThen does, and prints the response (asAnswered).
func (s *sender) send(cmd *cobra.Command, scopes []string, target *emporix.Write, values []string,
	about subject) error {
	return s.sendThen(cmd, scopes, target, values, about, asAnswered)
}

// sendThen sends a write to where target names (its path and query, and the
// headers and StaleNotFound it adds; it leaves target unchanged), with the
// sender's method, the body that its source gives for values, the arguments
// after those of its request, and a token of scopes: it fails as a usage
// error where it would have to ask without a terminal to ask on
// (confirmable), then makes the body, which must be JSON, and prepares the
// request; --dry-run prints it, and otherwise, unless --yes is given, it is
// sent only once confirmed as s.confirm asks, naming about. Once Emporix
// answered, it hands the response to then and prints what then returns
// (writeResponse), or, where the write failed, prints the response as finish
// does.
func (s *sender) sendThen(cmd *cobra.Command, scopes []string, target *emporix.Write, values []string,
	about subject, then answered) error {
	if err := s.confirmable(cmd); err != nil {
		return err
	}
	w := *target
	w.Method = s.method
	if err := s.body.fill(cmd.Context(), &w, values); err != nil {
		return err
	}
	client, err := s.a.client(cmd.Context(), scopes...)
	if err != nil {
		return err
	}
	requests, err := prepare(client, w.Split())
	switch {
	case err != nil:
		return err
	case s.dryRun:
		return writeRequests(cmd.OutOrStdout(), requests)
	case !s.yes:
		prompt, confirms := s.confirm.question(about, client.Tenant())
		if declined := confirm(cmd.Context(), cmd.ErrOrStderr(), cmd.InOrStdin(), requests, client.Tenant(), prompt,
			confirms); declined != nil {
			return declined
		}
	}
	if len(requests) > 1 {
		return s.sendBatches(cmd, client, requests, then)
	}
	response, err := client.Send(cmd.Context(), requests[0])
	if err != nil {
		return s.finish(cmd, response, err)
	}
	return s.answer(cmd, client, response, then)
}

// answer hands response to then and prints what it returns, and returns the
// error of either.
func (s *sender) answer(cmd *cobra.Command, client *emporix.Client, response []byte, then answered) error {
	printed, err := then(client, response)
	return errors.Join(err, s.writeResponse(cmd, printed))
}

// prepare returns the requests that client prepares for writes, in order.
func prepare(client *emporix.Client, writes []*emporix.Write) ([]*emporix.Request, error) {
	requests := make([]*emporix.Request, len(writes))
	for i, w := range writes {
		req, err := client.Prepare(w)
		if err != nil {
			return nil, err
		}
		requests[i] = req
	}
	return requests, nil
}

// writeRequests prints requests as a dry run shows them, an empty line
// between two.
func writeRequests(out io.Writer, requests []*emporix.Request) error {
	for i, req := range requests {
		if i > 0 {
			if _, err := io.WriteString(out, "\n"); err != nil {
				return err
			}
		}
		if err := output.WriteRequest(out, req.Method(), req.URL(), req.Header(), req.Body()); err != nil {
			return err
		}
	}
	return nil
}

// sendBatches sends requests, the batches of one bulk write
// (emporix.Write.Split), with client, in order: it hands the response of
// each to then and prints what it returns, or, where items of a batch
// failed, prints the response and lists them (listFailures), and goes on. A
// batch that fails as a whole stops the run (batchStopped), and so does a
// response that cannot be printed; the batches after it are not sent. Once
// every batch is answered, it fails where an item failed, with the class
// that every failure shares (emporix.JoinItems).
func (s *sender) sendBatches(cmd *cobra.Command, client *emporix.Client, requests []*emporix.Request,
	then answered) error {
	var failed []*emporix.ItemsError
	sent := 0 // the items of the batches answered
	for i, req := range requests {
		response, err := client.Send(cmd.Context(), req)
		var items *emporix.ItemsError
		switch {
		case err == nil:
			err = s.answer(cmd, client, response, then)
		case errors.As(err, &items):
			failed = append(failed, items)
			err = s.listFailures(cmd, response, items)
		default:
			return batchStopped(i, len(requests), sent, failed, s.finish(cmd, response, err))
		}
		if err != nil {
			return fmt.Errorf("request %d of %d was answered, but %w; those after it were not sent", i+1,
				len(requests), err)
		}
		sent += req.Items()
	}
	return s.withRecovery(emporix.JoinItems(sent, failed))
}

// batchStopped returns the error of batch i of n, which failed as a whole
// with err, after the batches before it, of sent items, were answered, with
// failed those whose items failed. err sets the exit code only where it
// shares the class of every failure before it.
func batchStopped(i, n, sent int, failed []*emporix.ItemsError, err error) error {
	listed := ""
	var before *emporix.ItemsError
	if errors.As(emporix.JoinItems(sent, failed), &before) {
		listed = fmt.Sprintf(", and %d of them failed, as listed above", len(before.Failed))
		if class := errors.Unwrap(before); class == nil || !errors.Is(err, class) {
			err = errors.New(err.Error()) // its message, without its class
		}
	}
	return fmt.Errorf("request %d of %d, from item %d: %w (the %d items before it were sent%s; those after it "+
		"were not)", i+1, n, sent, err, sent, listed)
}

// finish prints the response of a write that failed with err, if any, and
// returns err, which Send returns with a response only for a write of parts
// (a bulk write or a chain) and for one whose response says whether it was
// carried out (emporix.Write.ReportsOutcome). Of a write of parts, it lists the
// parts that failed on standard error, one per line, also when printing the
// response failed, whose error it joins to err, so that err's class still
// sets the exit code. A conflict says how to recover: as the sender's
// conflict says, where the write declares what its conflicts mean, else that
// the item, each item or what each command changes changed since it was read
// or exists already.
func (s *sender) finish(cmd *cobra.Command, response json.RawMessage, err error) error {
	var items *emporix.ItemsError
	errors.As(err, &items)
	return errors.Join(s.withRecovery(err), s.listFailures(cmd, response, items))
}

// listFailures prints response, if any, and lists the parts of a write of
// parts that failed, if any, on standard error, one per line
// (ItemsError.Lines), also when printing the response failed, whose error it
// returns.
func (s *sender) listFailures(cmd *cobra.Command, response json.RawMessage, items *emporix.ItemsError) error {
	printed := s.writeResponse(cmd, response)
	if items != nil {
		for _, line := range items.Lines() {
			// The failures are API-supplied text, so they are sanitized for the terminal.
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output.Sanitize(line))
		}
	}
	return printed
}

// withRecovery returns err, saying how to recover where it is a conflict: as
// the sender's conflict says, where the write declares what its conflicts
// mean, else that the item, each item or what each command changes changed
// since it was read or exists already.
func (s *sender) withRecovery(err error) error {
	if !errors.Is(err, emporix.ErrConflict) {
		return err
	}
	changed, again := "the item", "it"
	var items *emporix.ItemsError
	if errors.As(err, &items) {
		changed, again = "each", "them"
		if items.Noun == emporix.CommandNoun {
			changed, again = "what each command changes", "it"
		}
	}
	return fmt.Errorf("%w (%s)", err, cmp.Or(s.conflict, fmt.Sprintf(
		"%s changed since it was read, or exists already: read %s again before you retry", changed, again)))
}

// confirmable fails, as a usage error, a write that would have to ask for
// confirmation without --yes where it cannot: standard input gives its body
// (--file -) or is not a terminal. A dry run asks nothing.
func (s *sender) confirmable(cmd *cobra.Command) error {
	switch {
	case s.dryRun || s.yes:
		return nil
	case s.body.fromStdin():
		return usageError{fmt.Errorf("%s reads the body from standard input, which then cannot confirm it: "+
			"give --%s to send it", cmd.CommandPath(), yesFlag)}
	case !s.a.terminal:
		return usageError{fmt.Errorf("%s asks for confirmation on a terminal, which standard "+
			"input is not: give --%s to send it, or --%s to print it", cmd.CommandPath(), yesFlag, dryRunFlag)}
	default:
		return nil
	}
}

// writeResponse prints the response of a write: its JSON body with sorted
// keys and every value as sent, as JSON or YAML (-o), nothing for an empty
// one, and for one that is not JSON a note on standard error: the write
// succeeded all the same.
func (s *sender) writeResponse(cmd *cobra.Command, body json.RawMessage) error {
	switch {
	case body == nil:
		return nil
	case !json.Valid(body):
		_, err := fmt.Fprintf(cmd.ErrOrStderr(), "The response (%d bytes) is not JSON, so it is not printed.\n", len(body))
		return err
	default:
		return output.Write(cmd.OutOrStdout(), s.format.chosen(), body)
	}
}

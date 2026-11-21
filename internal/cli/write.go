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

const deleteUse = "delete"

const createUse = "create"

const importUse = "import"

const recalculate = "recalculate"

const fileFlag = "file"

const maxWriteFileBytes = 16 << 20

const stdinFileNote = " (- reads standard input; sending it then needs --" + yesFlag + ")"

const (
	dryRunFlag = "dry-run"
	yesFlag    = "yes"
)

type bodySource interface {
	bind(cmd *cobra.Command) (stdin string)
	scopes() []string
	fill(ctx context.Context, w *emporix.Write, values []string) error
	fromStdin() bool
}

type payload struct {
	optional bool
	array    bool
	maxItems int
	itemKey  string
	whole    bool
	batched  bool
	typed    bodyBinder
	body     boundBody
}

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
			"more) and answers for all of them at once, not item by item."
	case p.maxItems > 0:
		cmd.Long += "\n\n" + p.itemsSentence() + " its own and answers with the result of each: when one failed, epx " +
			"prints the response, lists the items that failed on standard error (one that failed with 500, 502, 503 " +
			"or 504 may have been carried out all the same) and " + bulkExitSentence
	}
	return "--" + fileFlag + " -"
}

const bulkExitSentence = "exits with the code that their failures share (for example 5 for a 409 Conflict, 4 for " +
	"a 404), else 1."

// Emporix's guide to bulk request limits: 200 items where the endpoint's reference names no other limit.
const bulkItemsDefault = 200

func (p *payload) itemsSentence() string {
	most := strconv.Itoa(p.maxItems)
	if p.batched {
		return "epx sends the items in requests of up to " + most + ", in order, confirmed once, and stops at a " +
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

func emptyObjectOrFile() *payload { return &payload{typed: objectBody{}} }

// The Customer Segment Service answers a patch body that is not an array of operations as done without applying it.
func operationsFile() *payload { return &payload{array: true} }

type noBody struct{}

func (noBody) bind(*cobra.Command) string { return "" }

func (noBody) scopes() []string { return nil }

func (noBody) fill(context.Context, *emporix.Write, []string) error { return nil }

func (noBody) fromStdin() bool { return false }

type sender struct {
	a        *app
	method   string
	body     bodySource
	confirm  confirmation
	conflict string
	dryRun   bool
	yes      bool
	format   *formatChoice
}

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

type answered func(client *emporix.Client, response []byte) (json.RawMessage, error)

func asAnswered(_ *emporix.Client, response []byte) (json.RawMessage, error) {
	return response, nil
}

func (s *sender) send(cmd *cobra.Command, scopes []string, target *emporix.Write, values []string,
	about subject) error {
	return s.sendThen(cmd, scopes, target, values, about, asAnswered)
}

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

func (s *sender) answer(cmd *cobra.Command, client *emporix.Client, response []byte, then answered) error {
	printed, err := then(client, response)
	return errors.Join(err, s.writeResponse(cmd, printed))
}

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

func (s *sender) sendBatches(cmd *cobra.Command, client *emporix.Client, requests []*emporix.Request,
	then answered) error {
	var failed []*emporix.ItemsError
	sent := 0
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

func batchStopped(i, n, sent int, failed []*emporix.ItemsError, err error) error {
	listed := ""
	var before *emporix.ItemsError
	if errors.As(emporix.JoinItems(sent, failed), &before) {
		listed = fmt.Sprintf(", and %d of them failed, as listed above", len(before.Failed))
		if class := errors.Unwrap(before); class == nil || !errors.Is(err, class) {
			err = errors.New(err.Error())
		}
	}
	return fmt.Errorf("request %d of %d, from item %d: %w (the %d items before it were sent%s; those after it "+
		"were not)", i+1, n, sent, err, sent, listed)
}

func (s *sender) finish(cmd *cobra.Command, response json.RawMessage, err error) error {
	var items *emporix.ItemsError
	errors.As(err, &items)
	return errors.Join(s.withRecovery(err), s.listFailures(cmd, response, items))
}

func (s *sender) listFailures(cmd *cobra.Command, response json.RawMessage, items *emporix.ItemsError) error {
	printed := s.writeResponse(cmd, response)
	if items != nil {
		for _, line := range items.Lines() {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output.Sanitize(line))
		}
	}
	return printed
}

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

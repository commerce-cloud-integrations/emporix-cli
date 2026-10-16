package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// errNotConfirmed is what a write returns when its confirmation is declined.
var errNotConfirmed = errors.New("not confirmed; nothing was sent")

// confirmation is how a terminal user confirms a write before epx sends it
// (sender): by answering y (yesNo) or, for a delete, which cannot be undone,
// by typing what it deletes (typed). confirmationFor chooses a write's once,
// when its command is built.
type confirmation interface {
	// help says in the write's help how it is confirmed: on a terminal before
	// it sends method, or with --yes, required as required says.
	help(method, required string) string
	// question returns what epx asks, after "Send METHOD URL to tenant T? ",
	// before it sends a write of target to tenant, and whether an answer
	// confirms it.
	question(target subject, tenant string) (prompt string, confirms func(answer string) bool)
}

// subject is what the confirmation of one write names: word, the identifier
// that a terminal user types to confirm a delete, as given (none where the
// command names no item: the tenant), and many, what the delete deletes with
// the same request besides, if it deletes more than one item (deletion).
type subject struct {
	word string
	many string
}

// confirmationFor returns the confirmation of a write of method: typed for a
// delete, whose usage names the identifier to type as word ("the tenant"
// where it names none) and which deletes what d says, else yesNo.
func confirmationFor(method, word string, d deletion) confirmation {
	if method == http.MethodDelete {
		return typed{word: word, deletes: d}
	}
	return yesNo{}
}

// yesNo confirms a write that changes data: y or yes, in any case, sends it.
type yesNo struct{}

func (yesNo) help(method, required string) string {
	return "Changes data: on a terminal, epx asks before it sends " + method + ". --" + yesFlag +
		" sends without\nasking (" + required + "); --" + dryRunFlag + " prints the request\n" +
		"instead of sending it."
}

func (yesNo) question(subject, string) (prompt string, confirms func(answer string) bool) {
	return "[y/N] ", func(answer string) bool {
		answer = strings.ToLower(strings.TrimSpace(answer))
		return answer == "y" || answer == "yes"
	}
}

// typed confirms a delete, which cannot be undone: typing the identifier of
// what it deletes, which its usage names as word, sends it.
type typed struct {
	word    string
	deletes deletion
}

func (t typed) help(method, required string) string {
	return t.deletes.sentence() + "\nOn a terminal, epx sends " + method + " only once you type " + t.word +
		" at its prompt.\n--" + yesFlag + " sends without asking (" + required + ");\n--" + dryRunFlag +
		" prints the request instead of sending it."
}

// question asks for target's word, or the tenant where it has none, as the
// terminal shows it (output.Sanitize): an answer confirms the delete when it
// shows the same, spaces around either aside. An empty answer never does.
func (typed) question(target subject, tenant string) (prompt string, confirms func(answer string) bool) {
	word := target.word
	if word == "" {
		word = tenant
	}
	shown := output.Sanitize(word)
	deletes := "It cannot be undone"
	if target.many != "" {
		deletes = "It deletes " + deletesMany(target.many) + " and cannot be undone"
	}
	want := strings.TrimSpace(shown)
	return deletes + ": type " + shown + " to confirm: ", func(answer string) bool {
		answer = strings.TrimSpace(output.Sanitize(answer))
		return answer != "" && answer == want
	}
}

// confirm asks on errOut whether to send requests, one or the batches of one
// write, which share their method and URL, to tenant, followed by prompt, and
// reads the answer (ask): one that confirms sends them; another, or no answer,
// declines (errNotConfirmed).
func confirm(ctx context.Context, errOut io.Writer, in io.Reader, requests []*emporix.Request, tenant, prompt string,
	confirms func(answer string) bool) error {
	req, batches := requests[0], ""
	if len(requests) > 1 {
		batches = fmt.Sprintf(" in %d requests", len(requests))
	}
	question := fmt.Sprintf("Send %s %s to tenant %s%s? %s", req.Method(), req.URL(), tenant, batches, prompt)
	ok, err := ask(ctx, errOut, in, question, confirms)
	if err == nil && !ok {
		return errNotConfirmed
	}
	return err
}

// confirmed asks question on errOut, followed by the prompt of yesNo, and
// reads the answer (ask): y or yes (in any case) confirms; anything else, or
// no answer, declines.
func confirmed(ctx context.Context, errOut io.Writer, in io.Reader, question string) (bool, error) {
	prompt, confirms := yesNo{}.question(subject{}, "")
	return ask(ctx, errOut, in, question+" "+prompt, confirms)
}

// ask writes question on errOut and reads the answer, one line, from in,
// reporting whether confirms accepts it. An interrupt ends the wait.
func ask(ctx context.Context, errOut io.Writer, in io.Reader, question string,
	confirms func(answer string) bool) (bool, error) {
	if _, err := fmt.Fprint(errOut, question); err != nil {
		return false, err
	}
	answer, err := readLine(ctx, in, "waiting for confirmation")
	if err != nil {
		return false, err
	}
	return confirms(answer), nil
}

// readLine returns the first line of in, without its line ending; an
// interrupt ends the wait, which what names.
func readLine(ctx context.Context, in io.Reader, what string) (string, error) {
	return untilDone(ctx, what, func() (string, error) {
		line := bufio.NewScanner(in)
		line.Scan()
		return line.Text(), line.Err()
	})
}

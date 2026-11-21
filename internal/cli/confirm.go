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

var errNotConfirmed = errors.New("not confirmed; nothing was sent")

type confirmation interface {
	help(method, required string) string
	question(target subject, tenant string) (prompt string, confirms func(answer string) bool)
}

type subject struct {
	word string
	many string
}

func confirmationFor(method, word string, d deletion) confirmation {
	if method == http.MethodDelete {
		return typed{word: word, deletes: d}
	}
	return yesNo{}
}

type yesNo struct{}

func (yesNo) help(method, required string) string {
	return "Changes data: on a terminal, epx asks before it sends " + method + ". --" + yesFlag +
		" sends without asking (" + required + "); --" + dryRunFlag + " prints the request " +
		"instead of sending it."
}

func (yesNo) question(subject, string) (prompt string, confirms func(answer string) bool) {
	return "[y/N] ", func(answer string) bool {
		answer = strings.ToLower(strings.TrimSpace(answer))
		return answer == "y" || answer == "yes"
	}
}

type typed struct {
	word    string
	deletes deletion
}

func (t typed) help(method, required string) string {
	return t.deletes.sentence() + " On a terminal, epx sends " + method + " only once you type " + t.word +
		" at its prompt. --" + yesFlag + " sends without asking (" + required + "); --" + dryRunFlag +
		" prints the request instead of sending it."
}

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

func confirmed(ctx context.Context, errOut io.Writer, in io.Reader, question string) (bool, error) {
	prompt, confirms := yesNo{}.question(subject{}, "")
	return ask(ctx, errOut, in, question+" "+prompt, confirms)
}

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

func readLine(ctx context.Context, in io.Reader, what string) (string, error) {
	return untilDone(ctx, what, func() (string, error) {
		line := bufio.NewScanner(in)
		line.Scan()
		return line.Text(), line.Err()
	})
}

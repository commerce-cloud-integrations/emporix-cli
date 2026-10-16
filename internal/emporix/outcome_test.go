package emporix

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

const (
	capturePath     = "/payment-gateway/{tenant}/payment/T1/capture"
	capturedRequest = "POST /payment-gateway/acme1/payment/T1/capture: "
)

// sendCapture sends a capture that reports its outcome to a fake that answers
// it with status and response, and returns what Send returns.
func sendCapture(t *testing.T, status int, response string, reportsOutcome bool) (json.RawMessage, error) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route("/payment-gateway/acme1/payment/T1/capture", respond(status, response))
	return send(fake.client(t), &Write{Method: http.MethodPost, Path: capturePath,
		Body: jsonBody(t, `{"amount":1,"currency":"EUR"}`), Answers: Answers{ReportsOutcome: reportsOutcome}})
}

// TestSendReadsTheOutcome checks a write that reports its outcome: a response
// that says that it succeeded is returned as sent; one that says that it
// failed is returned with an unclassified error that names the request and
// Emporix's message, if a string, bounded as an APIError's.
func TestSendReadsTheOutcome(t *testing.T) {
	const failed = "Emporix answered that it failed (successful false)"
	long := strings.Repeat("x", maxMessageBytes+10)
	for response, want := range map[string]string{
		`{"successful":true,"captureId":"c1"}`:                "",
		`{"message":"Success!","successful":true}`:            "",
		`{"successful":false}`:                                failed,
		`{"successful":false,"message":"declined"}`:           failed + ": declined",
		`{"successful":false,"message":{"code":"declined"}}`:  failed,
		`{"successful":false,"message":["declined"]}`:         failed,
		`{"successful":false,"message":"` + long + `"}`:       failed + ": " + long[:maxMessageBytes] + "…",
		`{"successful":false,"message":"","error":"expired"}`: failed + ": expired",
	} {
		body, err := sendCapture(t, http.StatusOK, response, true)
		switch {
		case string(body) != response:
			t.Errorf("%s: body %s; want it as sent", response, body)
		case want == "" && err != nil:
			t.Errorf("%s: error %v; want none", response, err)
		case want != "" && (err == nil || err.Error() != capturedRequest+want || errors.Unwrap(err) != nil):
			t.Errorf("%s: error %v; want %q, unclassified", response, err, capturedRequest+want)
		}
	}
}

// TestSendWithoutTheOutcome checks that a write that reports its outcome
// fails, unclassified, with the response, if any, where the response does
// not say whether it succeeded: it may have been carried out all the same.
func TestSendWithoutTheOutcome(t *testing.T) {
	want := capturedRequest + "the response does not say whether it succeeded (successful true or false), so " +
		carriedOutNote
	for _, tt := range []struct {
		status   int
		response string
	}{
		{http.StatusOK, `{"message":"queued"}`}, {http.StatusOK, `{"successful":null}`},
		{http.StatusOK, `{"successful":"false"}`}, {http.StatusOK, `{"successful":0}`}, {http.StatusOK, `null`},
		{http.StatusOK, `[{"successful":true}]`}, {http.StatusOK, `OK`}, {http.StatusOK, `true`},
		{http.StatusNoContent, ""}, {http.StatusAccepted, `{}`},
	} {
		body, err := sendCapture(t, tt.status, tt.response, true)
		if err == nil || err.Error() != want || errors.Unwrap(err) != nil || string(body) != tt.response ||
			(tt.response == "" && body != nil) {
			t.Errorf("%d %s: body %q, error %v; want %q", tt.status, tt.response, body, err, want)
		}
	}
}

// TestSendLeavesTheOutcomeOfOtherWrites checks that a write that does not
// report its outcome succeeds with any 2xx response, successful false
// included, and that a failure status of one that does is an APIError, as
// any write's.
func TestSendLeavesTheOutcomeOfOtherWrites(t *testing.T) {
	if body, err := sendCapture(t, http.StatusOK, `{"successful":false}`, false); err != nil ||
		string(body) != `{"successful":false}` {
		t.Errorf("body %s, error %v; want it as sent", body, err)
	}
	body, err := sendCapture(t, http.StatusNotFound, `{"successful":false,"message":"unknown transaction"}`, true)
	var api *APIError
	if !errors.As(err, &api) || !errors.Is(err, ErrNotFound) || api.Message != "unknown transaction" || body != nil {
		t.Errorf("body %s, error %#v; want a 404 APIError", body, err)
	}
}

// TestSendReturnsAFailureAnswer checks that a conflict of a write that
// declares FailureAnswered comes with its body, which says why, and that
// one that does not declare it, or another failure of one that does, comes
// without a body.
func TestSendReturnsAFailureAnswer(t *testing.T) {
	const answer = `{"accepted":false,"runId":"r1"}`
	for _, tt := range []struct {
		status   int
		declared bool
		body     bool
	}{
		{http.StatusConflict, true, true},
		{http.StatusConflict, false, false},
		{http.StatusNotFound, true, false},
	} {
		fake := newFakeEmporix(t)
		fake.route("/importtool/acme1/runs/r1/cancel", respond(tt.status, answer))
		body, err := send(fake.client(t), &Write{Method: http.MethodPost,
			Path: "/importtool/{tenant}/runs/r1/cancel", Answers: Answers{FailureAnswered: tt.declared}})
		if err == nil {
			t.Fatalf("status %d declared %t: no error", tt.status, tt.declared)
		}
		switch {
		case tt.body && string(body) != answer:
			t.Errorf("status %d: body %s; want the answer as sent", tt.status, body)
		case !tt.body && body != nil:
			t.Errorf("status %d declared %t: body %s; want none", tt.status, tt.declared, body)
		}
	}
}

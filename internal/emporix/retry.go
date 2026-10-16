package emporix

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

// retryPolicy bounds how often and how long a request is sent again (the retry rule).
type retryPolicy struct {
	attempts int
	wait     time.Duration // first backoff; doubles on every further attempt
	maxWait  time.Duration // cap for backoff and Retry-After
}

func defaultRetryPolicy() retryPolicy {
	return retryPolicy{attempts: 3, wait: 500 * time.Millisecond, maxWait: 30 * time.Second}
}

// The retry rule, for every request the client sends: a request is sent
// again only when that cannot make the service act on it twice.
//
//   - 429 Too Many Requests: the service refused it without acting on it, so
//     any request is sent again.
//   - 502, 503 and 504: the service may have acted on it before the gateway
//     answered, so only an idempotent request is sent again: a read (also one
//     sent as POST or PUT with a body, which changes no data) or a token
//     request. A write never is, whatever its method: a POST could create
//     something twice, and even a PUT may act twice, as a product template or
//     schema gets a new version with every update.
//   - A connection that could not be made sent nothing, so any request is sent
//     again (notSent), unless making it timed out: each attempt would wait as
//     long. Any other failure to send may have reached the service, and none is.
//
// An attempt that got 500, 502, 503 or 504, or that failed once its
// connection was made, may have been carried out without an answer that says
// so (unanswered): a write then reports that it may have been (Send).

// idempotence says whether sending a request twice has the effect of sending
// it once, so that the retry rule may send it again after a 502, 503 or 504.
type idempotence bool

const (
	idempotent idempotence = true  // a read or a token request
	onceOnly   idempotence = false // a write
)

// retryable reports whether the retry rule sends a request again after an
// attempt that failed to send with err, or else got status.
func retryable(status int, err error, repeat idempotence) bool {
	switch {
	case err != nil:
		return notSent(err) && !timedOut(err)
	case status == http.StatusTooManyRequests:
		return true
	default:
		return bool(repeat) && gatewayFailure(status)
	}
}

// gatewayFailure reports whether status is a 502, 503 or 504, which a gateway
// may answer after the service acted on the request.
func gatewayFailure(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

// notSent reports whether err is a connection that could not be made (a
// failed dial: refused, unreachable, not resolved or timed out), so nothing
// was sent, on every operating system.
func notSent(err error) bool {
	var op *net.OpError
	return errors.As(err, &op) && op.Op == "dial"
}

// timedOut reports whether err is a timeout.
func timedOut(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// unanswered reports whether an attempt that failed to send with err, or
// else got status, may have been carried out without an answer that says
// so: a 500 (the service may have failed after it acted), a 502, 503 or 504,
// or a failure once its connection was made.
func unanswered(status int, err error) bool {
	if err != nil {
		return !notSent(err)
	}
	return status == http.StatusInternalServerError || gatewayFailure(status)
}

// delay prefers the server's Retry-After over exponential backoff and stays
// within [0, maxWait].
func (p retryPolicy) delay(attempt int, header http.Header) time.Duration {
	wait := p.wait << (attempt - 1)
	if after, ok := retryAfter(header.Get("Retry-After"), p.maxWait); ok {
		wait = after
	}
	return max(0, min(wait, p.maxWait))
}

// retryAfter parses delay-seconds or an HTTP-date. Seconds are clamped to
// limit before conversion so huge values cannot overflow time.Duration.
func retryAfter(value string, limit time.Duration) (time.Duration, bool) {
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		return time.Duration(min(seconds, int64(limit/time.Second))) * time.Second, true
	}
	if when, err := http.ParseTime(value); err == nil {
		return time.Until(when), true
	}
	return 0, false
}

// doWithRetry sends the request built by newRequest, rebuilding it for each
// attempt, and returns the final exchange, or the failure to send it. It
// sends it again as the retry rule allows (retryable), up to the policy's
// attempts. The reply is uncertain, also with an error, once an attempt may
// have been carried out without an answer that says so (unanswered).
func (c *Client) doWithRetry(ctx context.Context, repeat idempotence, newRequest func() (*http.Request, error)) (
	reply, error) {
	uncertain := false
	for attempt := 1; ; attempt++ {
		req, err := newRequest()
		if err != nil {
			return reply{uncertain: uncertain}, err
		}
		result, err := c.exchange(req)
		uncertain = uncertain || unanswered(result.status, err)
		result.uncertain = uncertain
		if attempt >= c.retry.attempts || !retryable(result.status, err, repeat) {
			return result, err
		}
		if err := sleep(ctx, c.retry.delay(attempt, result.header)); err != nil {
			return reply{uncertain: uncertain}, err
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

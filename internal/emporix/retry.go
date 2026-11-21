package emporix

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

type retryPolicy struct {
	attempts int
	wait     time.Duration
	maxWait  time.Duration
}

func defaultRetryPolicy() retryPolicy {
	return retryPolicy{attempts: 3, wait: 500 * time.Millisecond, maxWait: 30 * time.Second}
}

type idempotence bool

// A write is never sent again, whatever its method: even a PUT may act twice,
// as a product template or schema gets a new version with every update.
const (
	idempotent idempotence = true
	onceOnly   idempotence = false
)

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

func gatewayFailure(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func notSent(err error) bool {
	var op *net.OpError
	return errors.As(err, &op) && op.Op == "dial"
}

func timedOut(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func unanswered(status int, err error) bool {
	if err != nil {
		return !notSent(err)
	}
	return status == http.StatusInternalServerError || gatewayFailure(status)
}

func (p retryPolicy) delay(attempt int, header http.Header) time.Duration {
	wait := p.wait << (attempt - 1)
	if after, ok := retryAfter(header.Get("Retry-After"), p.maxWait); ok {
		wait = after
	}
	return max(0, min(wait, p.maxWait))
}

func retryAfter(value string, limit time.Duration) (time.Duration, bool) {
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		return time.Duration(min(seconds, int64(limit/time.Second))) * time.Second, true
	}
	if when, err := http.ParseTime(value); err == nil {
		return time.Until(when), true
	}
	return 0, false
}

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

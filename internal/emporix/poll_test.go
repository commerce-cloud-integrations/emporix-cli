package emporix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"testing"
	"time"
)

// TestPollReadsUntilEnded checks that Poll reads the path again until ended
// says that a response ends it, handing it every response in turn.
func TestPollReadsUntilEnded(t *testing.T) {
	fake := newFakeEmporix(t)
	sent := 0
	fake.route("/jobs/j1", func(w http.ResponseWriter, r *http.Request) {
		sent++
		respond(http.StatusOK, `{"n":`+strconv.Itoa(sent)+`}`)(w, r)
	})
	var seen []string
	err := fake.client(t).Poll(context.Background(), "/jobs/j1", nil, time.Millisecond,
		func(body json.RawMessage) (bool, error) {
			seen = append(seen, string(body))
			return len(seen) == 3, nil
		})
	if want := []string{`{"n":1}`, `{"n":2}`, `{"n":3}`}; err != nil || !slices.Equal(seen, want) {
		t.Fatalf("Poll: %v, handed %q; want %q", err, seen, want)
	}
}

// TestPollFailures checks that a failure of ended, of a read, or the end of
// the context ends Poll with that failure, reading nothing more.
func TestPollFailures(t *testing.T) {
	errEnded := errors.New("not a job")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tt := range []struct {
		name   string
		ctx    context.Context
		status int
		want   error
		reads  int
	}{
		{name: "ended fails", ctx: context.Background(), status: http.StatusOK, want: errEnded, reads: 1},
		{name: "read fails", ctx: context.Background(), status: http.StatusNotFound, want: ErrNotFound, reads: 1},
		{name: "context ends", ctx: canceled, status: http.StatusOK, want: context.Canceled},
	} {
		fake := newFakeEmporix(t)
		fake.route("/jobs/j1", respond(tt.status, `{"message":"gone"}`))
		err := fake.client(t).Poll(tt.ctx, "/jobs/j1", nil, time.Millisecond, func(json.RawMessage) (bool, error) {
			return false, errEnded
		})
		if !errors.Is(err, tt.want) || len(fake.requests) != tt.reads {
			t.Errorf("%s: Poll: %v after %d reads; want %v after %d", tt.name, err, len(fake.requests), tt.want, tt.reads)
		}
	}
}

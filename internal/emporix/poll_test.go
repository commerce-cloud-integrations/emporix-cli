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

func TestPollFailures(t *testing.T) {
	errEnded := errors.New("not a job")
	for _, tt := range []struct {
		name     string
		canceled bool
		status   int
		want     error
		reads    int
	}{
		{name: "ended fails", status: http.StatusOK, want: errEnded, reads: 1},
		{name: "read fails", status: http.StatusNotFound, want: ErrNotFound, reads: 1},
		{name: "context ends", canceled: true, status: http.StatusOK, want: context.Canceled},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		if tt.canceled {
			cancel()
		}
		fake := newFakeEmporix(t)
		fake.route("/jobs/j1", respond(tt.status, `{"message":"gone"}`))
		err := fake.client(t).Poll(ctx, "/jobs/j1", nil, time.Millisecond, func(json.RawMessage) (bool, error) {
			return false, errEnded
		})
		cancel()
		if !errors.Is(err, tt.want) || len(fake.requests) != tt.reads {
			t.Errorf("%s: Poll: %v after %d reads; want %v after %d", tt.name, err, len(fake.requests), tt.want, tt.reads)
		}
	}
}

package emporix

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

// Poll reads path with query again and again, as Get reads it, every interval,
// and hands each response to ended, until ended reports that it ends the
// polling or fails, which Poll then returns; a read that fails ends it too,
// with that read's error, and so does ctx, with ctx's error, also while Poll
// waits. It waits interval before its first read: the caller has read once
// already, such as the job that a write started and answered with.
func (c *Client) Poll(ctx context.Context, path string, query url.Values, interval time.Duration,
	ended func(body json.RawMessage) (bool, error)) error {
	for {
		if err := sleep(ctx, interval); err != nil {
			return err
		}
		body, err := c.Get(ctx, path, query)
		if err != nil {
			return err
		}
		if done, err := ended(body); done || err != nil {
			return err
		}
	}
}

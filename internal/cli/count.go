package cli

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type counted struct {
	filters listFilters
}

func (c counted) bind(cmd *cobra.Command) boundRead {
	var opts emporix.ListOptions
	c.filters.without(sortable).registerFilters(&opts, cmd)
	count := func(client *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		n, err := client.Count(ctx, path, query, opts.Query)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(strconv.Itoa(n)), nil
	}
	return boundRead{read: count, items: func(body json.RawMessage) ([]json.RawMessage, error) {
		return []json.RawMessage{body}, nil
	}}
}

func (r resource[T]) countCommand(a *app, long string) *cobra.Command {
	rd := reader[json.Number]{scopes: r.scopes, view: view[json.Number]{
		singular: r.singular + " count",
		header:   []string{"COUNT"},
		row:      func(n json.Number, _ language) []string { return []string{n.String()} },
	}}
	return rd.command(a, &command{
		use: "count", short: "Count " + r.nouns(), long: long, req: r.listRequest(),
		fetch: counted{filters: r.filters},
	})
}

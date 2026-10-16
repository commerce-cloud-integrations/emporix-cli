package cli

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// counted reads the number of items of a collection that a HEAD request
// reports (emporix.Client.Count) as one JSON number, with the list's --query
// where its service filters the list: a count orders nothing, so it takes no
// --sort.
type counted struct {
	filters listFilters // the filters of the collection's list
}

func (c counted) bind(cmd *cobra.Command) (readFunc, shape) {
	var opts emporix.ListOptions
	c.filters.without(sortable).registerFilters(&opts, cmd)
	count := func(client *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		n, err := client.Count(ctx, path, query, opts.Query)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(strconv.Itoa(n)), nil
	}
	return count, func(body json.RawMessage) ([]json.RawMessage, error) { return []json.RawMessage{body}, nil }
}

// countCommand returns `count`, which prints the number of r's items, those
// --query finds where the service filters them, that Emporix reports without
// sending them: a HEAD request to the collection with list's query.
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

package cli

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type listFilters struct {
	kinds    filterKinds
	sortHelp string
}

type filterKinds int

const (
	noFilters filterKinds = 0
	sortable  filterKinds = 1 << iota
	queryable
	filterable = sortable | queryable
)

func (f listFilters) sorts() bool { return f.kinds&sortable != 0 }

func (f listFilters) queries() bool { return f.kinds&queryable != 0 }

func (f listFilters) without(kinds filterKinds) listFilters {
	f.kinds &^= kinds
	return f
}

const sortParameter = "sort"

func (f listFilters) register(opts *emporix.ListOptions, cmd *cobra.Command) {
	cmd.Flags().Var((*readLimit)(&opts.Limit), "limit", "print at most `n` items; paged reads stop early where they "+
		"safely can (0: every item)")
	f.registerFilters(opts, cmd)
}

const defaultSortUsage = "Emporix sort `order`, passed as given, e.g. code or name.en:desc,code; fields mostly " +
	"as in -o json (Emporix may ignore an unknown one)"

func (f listFilters) registerFilters(opts *emporix.ListOptions, cmd *cobra.Command) {
	if f.sorts() {
		sortHelp := cmp.Or(f.sortHelp, defaultSortUsage)
		cmd.Flags().StringVar(&opts.Sort, "sort", "", sortHelp)
	}
	if f.queries() {
		cmd.Flags().StringVar(&opts.Query, "query", "", queryUsage("passed as given"))
	}
}

func queryUsage(how string) string {
	return "Emporix q `filter`, " + how + ", e.g. code:(A1,A2); space-separated conditions must all match " +
		"(fields mostly as in -o json)"
}

const streamFlag = "stream"

func addStreamFlag(cmd *cobra.Command, stream *bool) {
	cmd.Flags().BoolVar(stream, streamFlag, false, fmt.Sprintf("print items as JSON Lines as soon as they are "+
		"checked, page by page on paged lists, reading up to %d pages where a list without it stops after %d: a "+
		"check failing on a later page exits 1 after the pages before it (excludes -o, --columns and --no-headers)",
		emporix.StreamedPages, emporix.KeptPages))
	for _, flag := range []string{outputFlag, columnsFlag, noHeadersFlag} {
		cmd.MarkFlagsMutuallyExclusive(streamFlag, flag)
	}
}

type readLimit int

// Set implements pflag.Value.
func (l *readLimit) Set(value string) error {
	limit, err := strconv.Atoi(value)
	if errors.Is(err, strconv.ErrRange) {
		return fmt.Errorf("invalid limit %q: too large", value)
	}
	if err != nil || limit < 0 {
		return fmt.Errorf("invalid limit %q: use a whole number, 0 for every item", value)
	}
	*l = readLimit(limit)
	return nil
}

func (l *readLimit) String() string { return strconv.Itoa(int(*l)) }

// Type implements pflag.Value.
func (l *readLimit) Type() string { return "n" }

package cli

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// listFilters is the declaration of a list's filters: what its service supports
// besides --limit, which every list has, and the help of --sort where it differs
// from defaultSortUsage. Listing passes the options to its collectionRead as
// emporix.ListOptions, so a declaration never handles them.
type listFilters struct {
	kinds    filterKinds // the filters the list takes, e.g. sortable
	sortHelp string      // the help of --sort; defaultSortUsage when empty
}

// filterKinds is a set of list filters, which combine: --limit only, and
// --sort, --query or both.
type filterKinds int

// The list filter kinds.
const (
	noFilters  filterKinds            = 0
	sortable   filterKinds            = 1 << iota // --sort: the Emporix sort parameter
	queryable                                     // --query: the Emporix q parameter
	filterable = sortable | queryable             // both
)

// sorts reports whether the list takes --sort.
func (f listFilters) sorts() bool { return f.kinds&sortable != 0 }

// queries reports whether the list takes --query.
func (f listFilters) queries() bool { return f.kinds&queryable != 0 }

// without returns f without the filters of kinds, e.g. a count's filters
// without --sort.
func (f listFilters) without(kinds filterKinds) listFilters {
	f.kinds &^= kinds
	return f
}

// sortParameter is the query parameter Emporix lists sort by.
const sortParameter = "sort"

// register adds --limit and the flags of f to cmd, setting opts.
func (f listFilters) register(opts *emporix.ListOptions, cmd *cobra.Command) {
	cmd.Flags().Var((*readLimit)(&opts.Limit), "limit", "print at most `n` items; paged reads stop early where they "+
		"safely can (0: every item)")
	f.registerFilters(opts, cmd)
}

// defaultSortUsage is the help of --sort for a list whose service may ignore
// an unknown field; a list that Emporix refuses one for declares its own.
const defaultSortUsage = "Emporix sort `order`, passed as given, e.g. code or name.en:desc,code; fields mostly " +
	"as in -o json (Emporix may ignore an unknown one)"

// registerFilters adds the flags of f, without --limit, to cmd, setting opts.
func (f listFilters) registerFilters(opts *emporix.ListOptions, cmd *cobra.Command) {
	if f.sorts() {
		sortHelp := cmp.Or(f.sortHelp, defaultSortUsage)
		cmd.Flags().StringVar(&opts.Sort, "sort", "", sortHelp)
	}
	if f.queries() {
		cmd.Flags().StringVar(&opts.Query, "query", "", queryUsage("passed as given"))
	}
}

// queryUsage returns the help of a --query flag, an Emporix q filter sent as
// how says.
func queryUsage(how string) string {
	return "Emporix q `filter`, " + how + ", e.g. code:(A1,A2); space-separated conditions must all match " +
		"(fields mostly as in -o json)"
}

// addStreamFlag adds --stream, which prints a list as JSON Lines while it is
// read (emporix.ListOptions.Stream) and so excludes the output flags
// (addTableOutput).
func addStreamFlag(cmd *cobra.Command, stream *bool) {
	cmd.Flags().BoolVar(stream, "stream", false, fmt.Sprintf("print items as JSON Lines as soon as they are "+
		"checked, page by page on paged lists, reading up to %d pages where a list without it stops after %d: a "+
		"check failing on a later page exits 1 after the pages before it (excludes -o, --columns and --no-headers)",
		emporix.StreamedPages, emporix.KeptPages))
	for _, flag := range []string{outputFlag, columnsFlag, noHeadersFlag} {
		cmd.MarkFlagsMutuallyExclusive("stream", flag)
	}
}

// readLimit is a --limit value: the most items a list prints, 0 for every
// item. It implements pflag.Value, so a negative or non-numeric limit is
// rejected while flags are parsed.
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

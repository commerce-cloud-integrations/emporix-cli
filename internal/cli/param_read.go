package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// paramRead reads as fetch reads, with query parameters that flags set
// (queryFlags), such as the customer that, with the site, identifies one
// cart: the query counterpart of bodyRead.
type paramRead struct {
	params queryFlags // the query parameters that flags set
	fetch  fetcher    // how the response is read: a listing or single
}

func (p *paramRead) bind(cmd *cobra.Command) (readFunc, shape) {
	setParams := p.params.bind(cmd)
	read, items := p.fetch.bind(cmd)
	return func(client *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		setParams(query) // the query that request.fill made for this run
		return read(client, ctx, path, query)
	}, items
}

// queryFlags are query parameters of a read or a write, or headers of a
// write, that flags set, each sent when its flag is given: the flags of
// required must be given, and exactly one of those of oneOf, if any, which
// is checked while flags are parsed.
type queryFlags struct {
	required []paramFlag // parameters that are always sent
	oneOf    []paramFlag // parameters of which exactly one is sent
	optional []paramFlag // parameters sent when their flag is given
}

// valueSetter is what a run sets the parameters of flags in: the url.Values
// of a query or the http.Header of a request.
type valueSetter interface{ Set(key, value string) }

// bind adds the flags of q to cmd and returns how a run sets in its query or
// headers the parameter of every flag given.
func (q queryFlags) bind(cmd *cobra.Command) func(into valueSetter) {
	params := slices.Concat(q.required, q.oneOf, q.optional)
	values := make([]paramValue, len(params))
	for i, param := range params {
		values[i] = paramValue{paramFlag: param}
		values[i].register(cmd)
	}
	for _, param := range q.required {
		requireFlag(cmd, param.flag)
	}
	oneOf := make([]string, len(q.oneOf))
	for i, param := range q.oneOf {
		oneOf[i] = param.flag
	}
	cmd.MarkFlagsOneRequired(oneOf...)
	cmd.MarkFlagsMutuallyExclusive(oneOf...)
	return func(into valueSetter) {
		for _, v := range values {
			if v.given != "" { // given: an empty value is refused while flags are parsed
				into.Set(v.param, v.given)
			}
		}
	}
}

// paramFlag is a parameter of a request that a flag sets: a query parameter
// or a header (queryFlags), or a field of a form (fieldPart).
type paramFlag struct {
	param   string                   // the parameter, e.g. customerId
	flag    string                   // the flag that sets it, e.g. customer
	usage   string                   // the flag's help, without the choices; a `word` in it names the value
	choices []string                 // the only values that the flag takes, e.g. QUARTER, MONTH, WEEK; any when empty
	check   func(value string) error // refuses a value that the parameter cannot take (wholeNumber); any when nil
	// profileSetting names the profile setting whose default the flag takes where it
	// is not given, e.g. siteFlag (takeProfileDefault); empty when it takes none.
	profileSetting string
}

// wholeNumber refuses a value that is not a whole number, 0 or more, written
// in digits, such as the version of an item (paramFlag.check).
func wholeNumber(value string) error {
	if _, err := strconv.ParseUint(value, 10, 63); err != nil {
		return errors.New("use a whole number, e.g. 2")
	}
	return nil
}

// paramValue is a paramFlag with the value given to it; it implements
// pflag.Value, so an empty value, which would identify nothing, one that is
// not among the flag's choices and one that its check refuses are rejected
// while flags are parsed.
type paramValue struct {
	paramFlag
	given string // the value given; empty when the flag is not given
}

// register adds v's flag to cmd, taking the default of its profileSetting where it has
// one. A flag with choices names them in its help, e.g. "the calendar period of each
// count: QUARTER, MONTH or WEEK", and shells complete them.
func (v *paramValue) register(cmd *cobra.Command) {
	if len(v.choices) == 0 {
		cmd.Flags().Var(v, v.flag, v.usage)
	} else {
		cmd.Flags().Var(v, v.flag, v.usage+": "+alternatives(v.choices))
		completeChoices(cmd, v.flag, v.choices)
	}
	if v.profileSetting != "" {
		takeProfileDefault(cmd, v.flag, v.profileSetting)
	}
}

// Set implements pflag.Value.
func (v *paramValue) Set(value string) error {
	if value == "" {
		return errors.New("empty") // the flag parser names the flag and the value
	}
	if len(v.choices) > 0 && !slices.Contains(v.choices, value) {
		return fmt.Errorf("use %s", alternatives(v.choices))
	}
	if v.check != nil {
		if err := v.check(value); err != nil {
			return err // the flag parser names the flag and the value
		}
	}
	v.given = value
	return nil
}

func (v *paramValue) String() string { return v.given }

// Type implements pflag.Value.
func (v *paramValue) Type() string { return "string" }

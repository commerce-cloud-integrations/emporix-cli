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

type paramRead struct {
	params queryFlags
	fetch  fetcher
}

func (p *paramRead) bind(cmd *cobra.Command) boundRead {
	setParams := p.params.bind(cmd)
	bound := p.fetch.bind(cmd)
	read := bound.read
	bound.read = func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		setParams(query)
		return read(c, ctx, path, query)
	}
	return bound
}

type queryFlags struct {
	required []paramFlag
	oneOf    []paramFlag
	optional []paramFlag
}

type valueSetter interface{ Set(key, value string) }

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
			if v.given != "" {
				into.Set(v.param, v.given)
			}
		}
	}
}

type paramFlag struct {
	param          string
	flag           string
	usage          string
	choices        []string
	check          func(value string) error
	profileSetting string
}

func wholeNumber(value string) error {
	if _, err := strconv.ParseUint(value, 10, 63); err != nil {
		return errors.New("use a whole number, e.g. 2")
	}
	return nil
}

type paramValue struct {
	paramFlag
	given string
}

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

func (v *paramValue) Set(value string) error {
	if value == "" {
		return errors.New("empty") // the flag parser names the flag and the value
	}
	if len(v.choices) > 0 && !slices.Contains(v.choices, value) {
		return fmt.Errorf("use %s", alternatives(v.choices))
	}
	if v.check != nil {
		if err := v.check(value); err != nil {
			return err
		}
	}
	v.given = value
	return nil
}

func (v *paramValue) String() string { return v.given }

func (v *paramValue) Type() string { return "string" }

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func requireAll(fields []bodyField) []bodyField {
	all := make([]bodyField, len(fields))
	for i, f := range fields {
		all[i] = f.require()
	}
	return all
}

func (f bodyField) afterAdd(hook func(cmd *cobra.Command, flag string)) bodyField {
	add, flag := f.value.add, f.value.flag
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		hook(cmd, flag)
		return value
	}
	return f
}

func (f bodyField) onlyWith(flag string) bodyField {
	return f.afterAdd(func(cmd *cobra.Command, own string) { onlyWithFlag(cmd, own, flag) })
}

func (f bodyField) suggesting(values ...string) bodyField {
	return f.afterAdd(func(cmd *cobra.Command, flag string) { completeChoices(cmd, flag, values) })
}

func onlyWithFlag(cmd *cobra.Command, own, flag string) {
	cmd.Flags().Lookup(own).Usage += " (only with --" + flag + ")"
	addPreRunCheck(cmd, func() error {
		if cmd.Flags().Changed(own) && !flagOn(cmd.Flags(), flag) {
			return usageError{fmt.Errorf("--%s applies only with --%s", own, flag)}
		}
		return nil
	})
}

func flagOn(fs *pflag.FlagSet, name string) bool {
	given := fs.Lookup(name)
	return given != nil && given.Changed && (given.Value.Type() != "bool" || given.Value.String() == "true")
}

func valueField(member, flag, usage string, newValue func() pflag.Value) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		value := newValue()
		fs.Var(value, flag, usage)
		return unlessGiven(fs, flag, func() any { return value })
	})
}

func numberField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(jsonNumber) })
}

func versionField(noun string) bodyField {
	return valueField("metadata.version", versionParam, "the `VERSION` of the "+noun+" as read (metadata.version)",
		func() pflag.Value { return new(versionValue) })
}

func moneyField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(moneyValue) })
}

func boolField(member, flag, usage string) bodyField { return sentBool(member, flag, usage, false) }

func onField(member, flag, usage string) bodyField { return sentBool(member, flag, usage, true) }

func sentBool(member, flag, usage string, value bool) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		on := fs.Bool(flag, value, usage)
		return func() any { return *on }
	})
}

func switchField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		on := fs.Bool(flag, false, usage)
		return unlessGiven(fs, flag, func() any { return *on })
	})
}

func choiceField(member, flag, usage string, choices ...string) bodyField {
	return bodyField{member: member, value: commandValue(flag, func(cmd *cobra.Command) func() any {
		choice := &paramValue{paramFlag: paramFlag{param: member, flag: flag, usage: usage, choices: choices}}
		choice.register(cmd)
		return unlessGiven(cmd.Flags(), flag, func() any { return choice.given })
	})}
}

func (f bodyField) optional() bodyField {
	add, flag := f.value.add, f.value.flag
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		return unlessGiven(cmd.Flags(), flag, add(cmd, whole))
	}
	return f
}

func (f bodyField) onlyWithChoice(flag, choice string) bodyField {
	return f.afterAdd(func(cmd *cobra.Command, own string) {
		cmd.Flags().Lookup(own).Usage += " (only with --" + flag + " " + choice + ", which requires it)"
		addPreRunCheck(cmd, func() error {
			fs := cmd.Flags()
			if chosen := fs.Lookup(flag); !chosen.Changed || (chosen.Value.String() == choice) == fs.Changed(own) {
				return nil
			}
			return usageError{fmt.Errorf("--%s applies only with --%s %s, which requires it", own, flag, choice)}
		})
	})
}

func fixedField(member string, value any) bodyField {
	return bodyField{member: member, value: bodyValue{add: func(*cobra.Command, string) func() any {
		return func() any { return value }
	}}}
}

func argumentField(member, name string) bodyField {
	return bodyField{member: member, value: bodyValue{add: func(cmd *cobra.Command, _ string) func() any {
		i := slices.Index(declaredArguments(cmd.Use), name)
		if i < 0 {
			panic(fmt.Sprintf("%s: the body member %q repeats an argument %s that it does not take", cmd.Use, member,
				name))
		}
		return func() any { return cmd.Flags().Arg(i) }
	}}}
}

type jsonNumber string

// Set implements pflag.Value.
func (n *jsonNumber) Set(value string) error {
	if !isJSONNumber(value) {
		return errors.New("use a number such as 12.5")
	}
	*n = jsonNumber(value)
	return nil
}

func (n *jsonNumber) String() string { return string(*n) }

// Type implements pflag.Value.
func (n *jsonNumber) Type() string { return "number" }

// MarshalJSON returns the number as given.
func (n jsonNumber) MarshalJSON() ([]byte, error) { return []byte(n), nil }

type versionValue uint64

// Set implements pflag.Value.
func (v *versionValue) Set(value string) error {
	n, err := strconv.ParseUint(value, 10, 63)
	if err != nil || n == 0 {
		return errors.New("use the version as read, a whole number such as 3")
	}
	*v = versionValue(n)
	return nil
}

func (v *versionValue) String() string { return strconv.FormatUint(uint64(*v), 10) }

// Type implements pflag.Value.
func (v *versionValue) Type() string { return "version" }

type moneyValue money

// Set implements pflag.Value.
func (m *moneyValue) Set(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 || !isJSONNumber(parts[0]) {
		return errors.New(`use AMOUNT CURRENCY, e.g. "100 EUR"`)
	}
	*m = moneyValue{Amount: json.Number(parts[0]), Currency: parts[1]}
	return nil
}

// String implements pflag.Value.
func (m *moneyValue) String() string { return spaced(m.Amount.String(), m.Currency) }

// Type implements pflag.Value.
func (m *moneyValue) Type() string { return "money" }

func isJSONNumber(value string) bool {
	return value != "" && strings.ContainsRune("-0123456789", rune(value[0])) &&
		strings.TrimSpace(value) == value && json.Valid([]byte(value))
}

func countField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(countValue) })
}

type countValue uint64

// Set implements pflag.Value.
func (c *countValue) Set(value string) error {
	n, err := strconv.ParseUint(value, 10, 63)
	if err != nil || n == 0 {
		return errors.New("use a whole number above 0, e.g. 50")
	}
	*c = countValue(n)
	return nil
}

func (c *countValue) String() string { return strconv.FormatUint(uint64(*c), 10) }

// Type implements pflag.Value.
func (c *countValue) Type() string { return "count" }

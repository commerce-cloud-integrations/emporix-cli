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

// The typed members of a body, such as a calculation's (bodyRead) or a
// write's: numbers, amounts of money and switches, and the constraints
// between their flags.

// requireAll returns fields, each required unless the flag that gives the
// body whole gives it (bodyField.require).
func requireAll(fields []bodyField) []bodyField {
	all := make([]bodyField, len(fields))
	for i, f := range fields {
		all[i] = f.require()
	}
	return all
}

// onlyWith returns f, a member that applies only with --flag (onlyWithFlag).
func (f bodyField) onlyWith(flag string) bodyField {
	add, own := f.value.add, f.value.flag
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		onlyWithFlag(cmd, own, flag)
		return value
	}
	return f
}

// onlyWithFlag makes --own of cmd apply only with --flag (flagOn): given
// without it, the command fails before anything is sent, a usage error (exit
// 2). The check runs before cmd (addPreRunCheck), and the usage of --own
// says so.
func onlyWithFlag(cmd *cobra.Command, own, flag string) {
	cmd.Flags().Lookup(own).Usage += " (only with --" + flag + ")"
	addPreRunCheck(cmd, func() error {
		if cmd.Flags().Changed(own) && !flagOn(cmd.Flags(), flag) {
			return usageError{fmt.Errorf("--%s applies only with --%s", own, flag)}
		}
		return nil
	})
}

// flagOn reports whether --name was given, and a switch as true: --gross=false,
// or --gross --gross=false, is no --gross.
func flagOn(fs *pflag.FlagSet, name string) bool {
	given := fs.Lookup(name)
	return given != nil && given.Changed && (given.Value.Type() != "bool" || given.Value.String() == "true")
}

// valueField is a member of a body set by --flag to a value that newValue
// makes, which checks the flag as it is parsed and is sent as JSON; it is
// left out of the body unless the flag is given.
func valueField(member, flag, usage string, newValue func() pflag.Value) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		value := newValue()
		fs.Var(value, flag, usage)
		return unlessGiven(fs, flag, func() any { return value })
	})
}

// numberField is a member of a body that is a number, set by --flag; it keeps
// the number as given, e.g. 1.590.
func numberField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(jsonNumber) })
}

// versionField is the member metadata.version of a body, the version of the
// item (noun) as read, which Emporix requires to change it, set by --version:
// a whole number, 1 or more (versionValue).
func versionField(noun string) bodyField {
	return valueField("metadata.version", versionParam, "the `VERSION` of the "+noun+" as read (metadata.version)",
		func() pflag.Value { return new(versionValue) })
}

// moneyField is a member of a body that is an amount in a currency, set by
// --flag as "AMOUNT CURRENCY", e.g. "100 EUR", and sent as {amount, currency}.
func moneyField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(moneyValue) })
}

// boolField is a member of a body that is true when --flag is given and
// false otherwise; it is always sent.
func boolField(member, flag, usage string) bodyField { return sentBool(member, flag, usage, false) }

// onField is a member of a body that is true unless --flag=false is given;
// it is always sent, so that Emporix's default never decides it.
func onField(member, flag, usage string) bodyField { return sentBool(member, flag, usage, true) }

// sentBool is a member of a body that --flag sets, value when it is not
// given; it is always sent.
func sentBool(member, flag, usage string, value bool) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		on := fs.Bool(flag, value, usage)
		return func() any { return *on }
	})
}

// switchField is a member of a body that is true when --flag is given, or
// as given (--flag=false), and left out of the body unless it is given.
func switchField(member, flag, usage string) bodyField {
	return flagField(member, flag, func(fs *pflag.FlagSet) func() any {
		on := fs.Bool(flag, false, usage)
		return unlessGiven(fs, flag, func() any { return *on })
	})
}

// choiceField is a member of a body that is one of choices, set by --flag,
// whose help names them and which shells complete (paramValue, whose
// parameter is the member); another value is rejected while flags are
// parsed. It is left out unless the flag is given.
func choiceField(member, flag, usage string, choices ...string) bodyField {
	return bodyField{member: member, value: commandValue(flag, func(cmd *cobra.Command) func() any {
		choice := &paramValue{paramFlag: paramFlag{param: member, flag: flag, usage: usage, choices: choices}}
		choice.register(cmd)
		return unlessGiven(cmd.Flags(), flag, func() any { return choice.given })
	})}
}

// optional returns f left out of the body unless its flag is given, such as
// a list that is no empty array then.
func (f bodyField) optional() bodyField {
	add, flag := f.value.add, f.value.flag
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		return unlessGiven(cmd.Flags(), flag, add(cmd, whole))
	}
	return f
}

// onlyWithChoice returns f, a member that applies only where --flag is
// choice, which then requires it: --own given with another value of --flag,
// or that value without --own, fails as a usage error (exit 2) before anything
// is sent (addPreRunCheck).
// Where --flag is not given, as where another flag gives the body whole,
// neither is checked. The usage of --own says so.
func (f bodyField) onlyWithChoice(flag, choice string) bodyField {
	add, own := f.value.add, f.value.flag
	f.value.add = func(cmd *cobra.Command, whole string) func() any {
		value := add(cmd, whole)
		cmd.Flags().Lookup(own).Usage += " (only with --" + flag + " " + choice + ", which requires it)"
		addPreRunCheck(cmd, func() error {
			fs := cmd.Flags()
			if chosen := fs.Lookup(flag); !chosen.Changed || (chosen.Value.String() == choice) == fs.Changed(own) {
				return nil
			}
			return usageError{fmt.Errorf("--%s applies only with --%s %s, which requires it", own, flag, choice)}
		})
		return value
	}
	return f
}

// fixedField is a member of a body that no flag sets: always value, such as
// the mode FULL of a reindex, the only one that Emporix offers.
func fixedField(member string, value any) bodyField {
	return bodyField{member: member, value: bodyValue{add: func(*cobra.Command, string) func() any {
		return func() any { return value }
	}}}
}

// argumentField is a member of a write's body that no flag sets: the command
// argument called name in its usage (e.g. CUSTOMER), such as the customer of
// its path, which Emporix requires in the body too. A body given with --file
// holds the member as given. The usage must name the argument: argumentField
// panics as the command is built otherwise, as siteField does for a site.
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

// jsonNumber is a number flag kept as given, so the body sends it exactly. It
// implements pflag.Value: a value that is not a JSON number is rejected while
// flags are parsed.
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

// MarshalJSON returns the number as given, which Set checked to be JSON.
func (n jsonNumber) MarshalJSON() ([]byte, error) { return []byte(n), nil }

// versionValue is the version of an item as read, a whole number, 1 or more,
// sent as a JSON number. It implements pflag.Value: another value is rejected
// while flags are parsed.
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

// moneyValue is an amount in a currency given as "AMOUNT CURRENCY", e.g.
// "100 EUR", as tables show amounts, and sent as money is. It implements
// pflag.Value: a value of another form is rejected while flags are parsed;
// Emporix checks the currency.
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

// String implements pflag.Value: the value as it was set, untrimmed (unlike
// money's table cell).
func (m *moneyValue) String() string { return spaced(m.Amount.String(), m.Currency) }

// Type implements pflag.Value.
func (m *moneyValue) Type() string { return "money" }

// isJSONNumber reports whether value is a JSON number, such as -1, 12.5 or
// 1e3 (not .5, 0x10 or NaN, which Go's parser would take), without spaces.
func isJSONNumber(value string) bool {
	return value != "" && strings.ContainsRune("-0123456789", rune(value[0])) &&
		strings.TrimSpace(value) == value && json.Valid([]byte(value))
}

// countField is a member of a body that is a count of things, such as reward
// points, set by --flag: a whole number above 0 (countValue).
func countField(member, flag, usage string) bodyField {
	return valueField(member, flag, usage, func() pflag.Value { return new(countValue) })
}

// countValue is a count of things, a whole number above 0 written in digits,
// sent as a JSON integer. It implements pflag.Value: another value, such as
// 0, -5, 1.5 or 1e3, is rejected while flags are parsed.
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

package cli

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const placeholderType = "<type>"

const customTypeReadScope = "custom." + placeholderType + "_read"

const customTypeManageScope = "custom." + placeholderType + "_manage"

const exampleType = "DOCUMENT"

var typeIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func namesType(scope string) bool { return strings.Contains(scope, placeholderType) }

func typeScopes(use string, req request, scopes []string, builtIn builtInTypes) typeRule {
	rule := newTypeRule(scopes, builtIn)
	if !rule.checks() {
		return rule
	}
	if len(req.arguments()) == 0 {
		panic(fmt.Sprintf("%s: its type (a scope that names %s, or built-in types) is the command's first "+
			"argument, which its request lacks", use, placeholderType))
	}
	for _, t := range builtIn.types {
		switch {
		case writeCapable(t.scopes):
			panic(fmt.Sprintf("%s: the built-in type %s requests a write-capable scope", use, t.name))
		case len(t.scopes) > 0 && !rule.custom():
			panic(fmt.Sprintf("%s: the built-in type %s requests scopes in place of none that names %s", use,
				t.name, placeholderType))
		}
	}
	return rule
}

type typeRule struct {
	builtIn builtInTypes
	typed   []string
}

func newTypeRule(scopes []string, builtIn builtInTypes) typeRule {
	return typeRule{builtIn: builtIn, typed: typedScopes(scopes)}
}

func (r typeRule) custom() bool { return len(r.typed) > 0 }

func (r typeRule) checks() bool { return r.custom() || len(r.builtIn.types) > 0 }

func (r typeRule) args(accept cobra.PositionalArgs) cobra.PositionalArgs {
	if !r.checks() {
		return accept
	}
	return func(cmd *cobra.Command, args []string) error {
		if err := accept(cmd, args); err != nil {
			return err
		}
		if err := r.accepts(args[0]); err != nil {
			return fmt.Errorf("invalid %s %q: %w", argumentNames(cmd.Use, 1)[0], args[0], err)
		}
		return nil
	}
}

func (r typeRule) accepts(arg string) error {
	if err := r.builtIn.spelling(arg); err != nil {
		return err
	}
	if _, found := r.builtIn.find(arg); !r.custom() || found || typeIDPattern.MatchString(arg) {
		return nil
	}
	others := ""
	if len(r.builtIn.types) > 0 {
		others = "; a built-in type is " + alternatives(r.builtIn.names())
	}
	return fmt.Errorf("a custom entity type has letters, digits and underscores, the first a letter%s", others)
}

func (r typeRule) fill(args, requested []string) []string {
	if !r.custom() {
		return requested
	}
	return r.builtIn.fill(requested, args[0])
}

func (r typeRule) helpEnding(help, sentence string) string {
	return joinSentences(help, r.builtIn.inPlaceOf(r.typed)) + "\n\n" + sentence
}

func refuseTypeScopes(use string, scopes []string) {
	if slices.ContainsFunc(scopes, namesType) {
		panic(fmt.Sprintf("%s: only the scopes that the command itself requests may name %s", use, placeholderType))
	}
}

func withType(scopes []string, typ string) []string {
	filled := make([]string, len(scopes))
	for i, scope := range scopes {
		filled[i] = strings.ReplaceAll(scope, placeholderType, strings.ToLower(typ))
	}
	return filled
}

func unknownFor(scopes []string, otherwise unknownParent) unknownParent {
	if slices.ContainsFunc(scopes, namesType) {
		return unscoped
	}
	return otherwise
}

func typeSentence(builtIn builtInTypes) string {
	return builtIn.clause("TYPE") + "; an unknown one " + unscoped.String() + "."
}

func typedScopes(scopes []string) []string {
	typed := slices.DeleteFunc(slices.Clone(scopes), func(scope string) bool { return !namesType(scope) })
	slices.Sort(typed)
	return slices.Compact(typed)
}

func typeNote(scopes []string) string {
	typed := typedScopes(scopes)
	if len(typed) == 0 {
		return ""
	}
	return fmt.Sprintf("%s is the custom entity type in lower case (%s for %s).", placeholderType,
		strings.Join(withType(typed, exampleType), ", "), exampleType)
}

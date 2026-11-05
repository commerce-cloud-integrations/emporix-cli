package cli

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// placeholderType in a scope stands for the custom entity type whose
// instances a command reads or changes, in lower case, as Emporix names the
// scopes that it creates for each type: custom.<type>_read is
// custom.document_read for the type DOCUMENT. A command whose scopes name it
// takes the type as its first argument (typeScopes).
const placeholderType = "<type>"

// customTypeReadScope is the scope that reads the instances of one custom
// entity type, narrower than the tenant-wide schema.custominstance_read.
const customTypeReadScope = "custom." + placeholderType + "_read"

// customTypeManageScope is the scope that changes the instances of one custom
// entity type, narrower than the tenant-wide schema.custominstance_manage.
const customTypeManageScope = "custom." + placeholderType + "_manage"

// exampleType is the custom entity type whose scopes help text shows to say
// what placeholderType stands for (typeNote).
const exampleType = "DOCUMENT"

// unscoped is what a command whose scopes name a custom entity type gives for
// a type that does not exist: Emporix grants no scope for it, which epx
// reports before it sends anything; nor for a new type, whose scopes Emporix
// grants only minutes after it created it (2 to 3 in a test). Its parts,
// unscopedLead and unscopedRest, serve typeSentence, which puts them on two
// lines, until help text is reflowed in one place.
const (
	unscopedLead               = "has no scope that"
	unscopedRest               = "Emporix grants (exit code 3), and a type created in the last minutes has none yet"
	unscoped     unknownParent = unscopedLead + " " + unscopedRest
)

// typeIDPattern matches what can be the ID of a custom entity type, which
// the Schema Service documents as uppercase letters, digits and underscores,
// the first a letter; epx also accepts lower case, which Emporix may refuse.
var typeIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// namesType reports whether scope names a custom entity type (placeholderType).
func namesType(scope string) bool { return strings.Contains(scope, placeholderType) }

// typeScopes makes the first argument of cmd, a command that sends req and
// may request scopes, a type where scopes name a custom entity type or
// builtIn holds types, and returns the rule of what that argument takes
// (newTypeRule), which also fills in the scopes that a run requests
// (typeRule.fill).
//
// A run fails as a usage error, before anything is sent, for a first
// argument that the rule does not accept (typeRule.accepts). typeScopes wraps
// the argument rule that cmd has, so it is called after request.bind, which
// sets that rule.
//
// A command with a type but no argument, with built-in types that request a
// write-capable scope, which a read would request without its opt-in, or
// with built-in types that request scopes where none of its own names a
// custom entity type, which they would replace, is a mistake in its
// declaration, so it panics while the command tree is built,
// which every test of the tree does.
func typeScopes(cmd *cobra.Command, req request, scopes []string, builtIn builtInTypes) typeRule {
	rule := newTypeRule(scopes, builtIn)
	if !rule.custom() && len(builtIn.types) == 0 {
		return rule
	}
	if len(req.arguments()) == 0 {
		panic(fmt.Sprintf("%s: its type (a scope that names %s, or built-in types) is the command's first "+
			"argument, which its request lacks", cmd.Use, placeholderType))
	}
	for _, t := range builtIn.types {
		switch {
		case writeCapable(t.scopes):
			panic(fmt.Sprintf("%s: the built-in type %s requests a write-capable scope", cmd.Use, t.name))
		case len(t.scopes) > 0 && !rule.custom():
			panic(fmt.Sprintf("%s: the built-in type %s requests scopes in place of none that names %s", cmd.Use,
				t.name, placeholderType))
		}
	}
	accept := cmd.Args
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if err := accept(cmd, args); err != nil {
			return err
		}
		if err := rule.accepts(args[0]); err != nil {
			return fmt.Errorf("invalid %s %q: %w", argumentNames(cmd.Use, 1)[0], args[0], err)
		}
		return nil
	}
	return rule
}

// typeRule is what the first argument of a command takes as its type
// (typeScopes): one of builtIn, as builtIn takes them, or another type,
// which, where the command is custom (typed holds those of its scopes that
// name a custom entity type, typedScopes), must be what a custom entity type
// can be named, as it becomes part of a scope.
type typeRule struct {
	builtIn builtInTypes
	typed   []string
}

// newTypeRule returns the rule of the first argument of a command that
// requests scopes and takes builtIn.
func newTypeRule(scopes []string, builtIn builtInTypes) typeRule {
	return typeRule{builtIn: builtIn, typed: typedScopes(scopes)}
}

// custom reports whether a scope of r's command names a custom entity type,
// which its first argument then fills in.
func (r typeRule) custom() bool { return len(r.typed) > 0 }

// accepts returns why arg cannot be the type that r takes: it writes a type
// of r's exact built-in types otherwise (builtInTypes.spelling); or, where r
// is custom, it is none of the built-in types and no custom entity type can
// be named so (typeIDPattern). Nil where arg can be.
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

// fill returns requested, the scopes of a command of r, as a run with the
// arguments args requests them: where r is custom, for the type that args[0]
// names (builtInTypes.fill); as requested otherwise.
func (r typeRule) fill(args, requested []string) []string {
	if !r.custom() {
		return requested
	}
	return r.builtIn.fill(requested, args[0])
}

// helpEnding returns what ends the help of a command whose type argument
// takes what r says: the note on the scopes that r's built-in types request
// in place of r's typed scopes, which a run fills in alike (fill), if any
// (builtInTypes.inPlaceOf), then, after a blank line, sentence, the
// command's scope sentence (scopeSentence, writeScopeSentence).
func (r typeRule) helpEnding(sentence string) string {
	return r.builtIn.inPlaceOf(r.typed) + "\n\n" + sentence
}

// refuseTypeScopes refuses scopes that use, a command, requests as declared,
// without passing them to typeScopes, such as those of a write's body,
// of its choice's flags or of the job it starts: one that names a custom
// entity type would be sent with the placeholder, so it is a mistake in the
// declaration of use, which panics while the command tree is built, which
// every test of the tree does.
func refuseTypeScopes(use string, scopes []string) {
	if slices.ContainsFunc(scopes, namesType) {
		panic(fmt.Sprintf("%s: only the scopes that the command itself requests may name %s", use, placeholderType))
	}
}

// withType returns scopes with the custom entity type typ filled in where
// they name one (placeholderType), in lower case.
func withType(scopes []string, typ string) []string {
	filled := make([]string, len(scopes))
	for i, scope := range scopes {
		filled[i] = strings.ReplaceAll(scope, placeholderType, strings.ToLower(typ))
	}
	return filled
}

// unknownFor returns what a command with scopes gives for an unknown parent
// item: unscoped where one of them names a custom entity type, which the
// parent is (its first argument, typeScopes), otherwise otherwise.
func unknownFor(scopes []string, otherwise unknownParent) unknownParent {
	if slices.ContainsFunc(scopes, namesType) {
		return unscoped
	}
	return otherwise
}

// typeSentence says that TYPE, the first argument of a command whose scopes
// name a custom entity type, is one of builtIn or the ID of a custom entity
// type (builtInTypes.clause), and what an unknown one gives (unscoped, its
// parts on two lines).
func typeSentence(builtIn builtInTypes) string {
	return builtIn.clause("TYPE") + "; an unknown one " + unscopedLead + "\n" + unscopedRest + "."
}

// typedScopes returns those of scopes that name a custom entity type
// (placeholderType), sorted, each once.
func typedScopes(scopes []string) []string {
	typed := slices.DeleteFunc(slices.Clone(scopes), func(scope string) bool { return !namesType(scope) })
	slices.Sort(typed)
	return slices.Compact(typed)
}

// typeNote returns what follows a scope sentence naming scopes where they
// name a custom entity type (placeholderType): a new line saying what it
// stands for, with those scopes of an example type; else nothing. A command
// whose built-in types request scopes of their own in their place says so
// before its scope sentence (builtInTypes.inPlaceOf).
func typeNote(scopes []string) string {
	typed := typedScopes(scopes)
	if len(typed) == 0 {
		return ""
	}
	return fmt.Sprintf("\n%s is the custom entity type in lower case (%s for %s).", placeholderType,
		strings.Join(withType(typed, exampleType), ", "), exampleType)
}

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
// grants only minutes after it created it (2 to 3 in a test).
const unscoped unknownParent = "has no scope that Emporix grants (exit code 3), and a type created in " +
	"the last minutes has none yet"

// typeIDPattern matches what can be the ID of a custom entity type, which
// the Schema Service documents as uppercase letters, digits and underscores,
// the first a letter; epx also accepts lower case, which Emporix may refuse.
var typeIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// namesType reports whether scope names a custom entity type (placeholderType).
func namesType(scope string) bool { return strings.Contains(scope, placeholderType) }

// typeScopes makes the first argument of cmd, a command that sends req and
// may request scopes, a type where scopes name a custom entity type or
// builtIn holds types. It returns, where scopes name a custom entity type,
// the tokens of builtIn's scopes (builtInTypes.tokens), which the scope
// annotation and sentence of cmd hold after its own, none otherwise; and how
// a run with the arguments args fills in the scopes that it requests: where
// scopes name a custom entity type, for the type that args[0] names
// (builtInTypes.fill); as requested otherwise.
//
// A run fails as a usage error, before anything is sent, for a first
// argument that writes a type of an exact builtIn otherwise
// (builtInTypes.spelling), and, where scopes name a custom entity type, for
// one that is none of builtIn and that no custom entity type can be named
// (typeIDPattern), as it would become part of a scope (typeArgument).
// typeScopes wraps the argument rule that cmd has, so it is called after
// request.bind, which sets that rule.
//
// A command with a type but no argument, or with built-in types that request
// a write-capable scope, which a read would request without its opt-in, is a
// mistake in its declaration, so it panics while the command tree is built,
// which every test of the tree does.
func typeScopes(cmd *cobra.Command, req request, scopes []string, builtIn builtInTypes) (
	tokens []string, fill func(args, requested []string) []string) {
	custom := slices.ContainsFunc(scopes, namesType)
	asRequested := func(_, requested []string) []string { return requested }
	if !custom && len(builtIn.types) == 0 {
		return nil, asRequested
	}
	if len(req.arguments()) == 0 {
		panic(fmt.Sprintf("%s: its type (a scope that names %s, or built-in types) is the command's first "+
			"argument, which its request lacks", cmd.Use, placeholderType))
	}
	for _, t := range builtIn.types {
		if writeCapable(t.scopes) {
			panic(fmt.Sprintf("%s: the built-in type %s requests a write-capable scope", cmd.Use, t.name))
		}
	}
	accept := cmd.Args
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if err := accept(cmd, args); err != nil {
			return err
		}
		if err := typeArgument(args[0], custom, builtIn); err != nil {
			return fmt.Errorf("invalid %s %q: %w", argumentNames(cmd.Use, 1)[0], args[0], err)
		}
		return nil
	}
	if !custom {
		return nil, asRequested
	}
	return builtIn.tokens(), func(args, requested []string) []string { return builtIn.fill(requested, args[0]) }
}

// typeArgument returns why arg cannot be the type that a command's first
// argument names (typeScopes): it writes a type of builtIn otherwise
// (builtInTypes.spelling); or, where custom, it is none of builtIn and no
// custom entity type can be named so.
func typeArgument(arg string, custom bool, builtIn builtInTypes) error {
	if err := builtIn.spelling(arg); err != nil {
		return err
	}
	if _, found := builtIn.find(arg); !custom || found || typeIDPattern.MatchString(arg) {
		return nil
	}
	others := ""
	if len(builtIn.types) > 0 {
		others = "; a built-in type is " + alternatives(builtIn.names())
	}
	return fmt.Errorf("a custom entity type has letters, digits and underscores, the first a letter%s", others)
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

// typeSentence says that arg, the first argument of a command whose scopes
// name a custom entity type, is one of builtIn or the ID of a custom entity
// type (builtInTypes.clause), and what an unknown one gives.
func typeSentence(arg string, builtIn builtInTypes) string {
	return builtIn.clause(arg) + "; an unknown one " + string(unscoped) + "."
}

// typeNote returns what follows a scope sentence naming scopes where they
// name a custom entity type (placeholderType): a new line saying what it
// stands for, with those scopes of an example type, then, where some of
// types, the built-in types of the scope sentence (parseScopeTokens), request
// scopes of their own, a sentence on a new line, one built-in type per line,
// naming those that each requests in their place; else nothing.
func typeNote(scopes []string, types []builtInType) string {
	typed := slices.DeleteFunc(slices.Clone(scopes), func(scope string) bool { return !namesType(scope) })
	if len(typed) == 0 {
		return ""
	}
	slices.Sort(typed)
	typed = slices.Compact(typed)
	note := fmt.Sprintf("\n%s is the custom entity type in lower case (%s for %s).", placeholderType,
		strings.Join(withType(typed, exampleType), ", "), exampleType)
	var each []string
	for _, t := range types {
		if len(t.scopes) > 0 {
			each = append(each, t.name+" requires "+strings.Join(t.scopes, ", "))
		}
	}
	if len(each) > 0 {
		note += "\nIn place of " + strings.Join(typed, ", ") + ", " + strings.Join(each, ";\n") + "."
	}
	return note
}

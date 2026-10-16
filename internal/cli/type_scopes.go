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

// typeScopes binds scopes, every scope that cmd, a command that sends req,
// may request, and returns how a run fills in the scopes that it requests
// for its arguments: as declared, or, where scopes name a custom entity type
// (placeholderType), with the type that the first argument names filled in,
// in lower case. cmd then accepts as its first argument only an ID that a
// custom entity type can have (customType), so that a scope holds nothing
// else: it wraps the argument rule that cmd has, so it is called after
// everything that sets one (request.bind). A command whose scopes name a type
// but whose request takes no argument is a mistake in its declaration, so it
// panics while the command tree is built, which every test of the tree does.
func typeScopes(cmd *cobra.Command, req request, scopes []string) func(args, requested []string) []string {
	if !slices.ContainsFunc(scopes, namesType) {
		return func(_, requested []string) []string { return requested }
	}
	if len(req.arguments()) == 0 {
		panic(fmt.Sprintf("%s: a scope that names %s takes the type from the command's first argument, "+
			"which its request lacks", cmd.Use, placeholderType))
	}
	cmd.Args = customType(cmd.Args)
	return func(args, requested []string) []string { return withType(requested, args[0]) }
}

// refuseTypeScopes refuses scopes that use, a command, requests as declared,
// without passing them to typeScopes, such as those of a write's body, of
// its choice's flags or of the job it starts: one that names a custom entity type would be sent with
// the placeholder, so it is a mistake in the declaration of use, which
// panics while the command tree is built, which every test of the tree does.
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

// customType returns accept, which takes the arguments of a command, refusing
// also a first argument that no custom entity type has as its ID
// (typeIDPattern).
func customType(accept cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := accept(cmd, args); err != nil {
			return err
		}
		if !typeIDPattern.MatchString(args[0]) {
			return fmt.Errorf("invalid %s %q: a custom entity type has letters, digits and underscores, "+
				"the first a letter", argumentNames(cmd.Use, 1)[0], args[0])
		}
		return nil
	}
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
// name a custom entity type, is the ID of one, and what an unknown one gives.
func typeSentence(arg string) string {
	return fmt.Sprintf("%s is a custom entity type, as `epx custom-entities list` shows its ID; an unknown one %s.",
		arg, unscoped)
}

// typeNote returns what follows a scope sentence naming scopes: where they
// name a custom entity type (placeholderType), a new line saying what it
// stands for, with those scopes of an example type; else nothing.
func typeNote(scopes []string) string {
	typed := slices.DeleteFunc(slices.Clone(scopes), func(scope string) bool { return !namesType(scope) })
	if len(typed) == 0 {
		return ""
	}
	slices.Sort(typed)
	return fmt.Sprintf("\n%s is the custom entity type in lower case (%s for %s).", placeholderType,
		strings.Join(slices.Compact(withType(typed, exampleType)), ", "), exampleType)
}

package cli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// The help of every read and write command ends with the OAuth scopes that it
// requests, and that of every group with those of the commands below it.

// Annotations that hold the scopes of a read command (scopesAnnotation) and
// of a write command, separated by spaces, so that a group can say which
// scopes its reads and its writes request.
const (
	scopesAnnotation      = "scopes"
	writeScopesAnnotation = "writeScopes"
)

// newGroup returns a command group holding reads and writes (and groups of
// them), whose help ends with the scopes they request (groupScopeSentence).
func newGroup(use, short, long string, commands ...*cobra.Command) *cobra.Command {
	group := &cobra.Command{Use: use, Short: short, Long: long}
	group.AddCommand(commands...)
	if sentence := groupScopeSentence(group); sentence != "" {
		group.Long += "\n\n" + sentence
	}
	return group
}

// scopeKind is a kind of command below a group whose scopes the group's help
// names: its reads or its writes.
type scopeKind struct {
	name     string                // how the group names them, e.g. "reads"
	sets     []string              // their distinct scope sets (scopeSets); none: the group holds none
	sentence func([]string) string // the scope sentence of one of them
	optIn    bool                  // one of them with a write-capable scope is opt-in (optInNote)
}

// groupScopeSentence says which OAuth scopes the commands below group
// request, at any depth, by kind (scopeKind): where they are of one kind and
// their scopes agree, the scope sentence of one of them; otherwise a line per
// kind, its reads first, that names their scopes where they agree and says
// that they differ otherwise, followed by optInSentence where reads are
// opt-in, and by one typeNote for the scopes the lines name. Without reads or
// writes, nothing.
func groupScopeSentence(group *cobra.Command) string {
	kinds := slices.DeleteFunc([]scopeKind{
		{name: "reads", sets: scopeSets(group, scopesAnnotation), sentence: scopeSentence, optIn: true},
		{name: "writes", sets: scopeSets(group, writeScopesAnnotation), sentence: writeScopeSentence},
	}, func(kind scopeKind) bool { return len(kind.sets) == 0 })
	if len(kinds) == 1 && len(kinds[0].sets) == 1 {
		return kinds[0].sentence(strings.Fields(kinds[0].sets[0]))
	}
	lines := make([]string, len(kinds))
	var named []string // the scopes that the lines name, those of kinds whose scopes agree
	for i, kind := range kinds {
		lines[i] = "Its " + kind.name + " " + groupRequirement(kind.sets)
		if kind.optIn {
			lines[i] += optInNote(strings.Fields(strings.Join(kind.sets, " ")))
		}
		if len(kind.sets) == 1 {
			named = append(named, strings.Fields(kind.sets[0])...)
		}
	}
	return strings.Join(lines, "\n") + typeNote(named)
}

// groupRequirement says what the commands of one or more scope sets require,
// after "Its reads" or "Its writes".
func groupRequirement(sets []string) string {
	switch {
	case len(sets) > 1:
		return "require different OAuth scopes; the help of each names them."
	case sets[0] == "":
		return "need no OAuth scope."
	default:
		return "require " + theScopes(strings.Fields(sets[0])) + "."
	}
}

// scopeSets returns the distinct scope sets of the commands below cmd, at any
// depth, that carry annotation, each sorted and separated by spaces (their
// order does not matter).
func scopeSets(cmd *cobra.Command, annotation string) []string {
	var all []string
	for _, sub := range cmd.Commands() {
		if scopes, annotated := sub.Annotations[annotation]; annotated {
			all = append(all, strings.Join(slices.Sorted(slices.Values(strings.Fields(scopes))), " "))
		} else {
			all = append(all, scopeSets(sub, annotation)...)
		}
	}
	slices.Sort(all)
	return slices.Compact(all)
}

// scopeSentence says which OAuth scopes a read requests, followed by
// typeNote where they name a custom entity type and optInSentence when one
// of them is write-capable.
func scopeSentence(scopes []string) string {
	return writeScopeSentence(scopes) + optInNote(scopes)
}

// noScopeSentence ends the help of a command that requests no scope.
const noScopeSentence = "No OAuth scope is needed."

// writeScopeSentence says which OAuth scopes a write requests, followed by
// typeNote where they name a custom entity type; noScopeSentence where it
// requests none. A read's scope sentence (scopeSentence) builds on it.
func writeScopeSentence(scopes []string) string {
	if len(scopes) == 0 {
		return noScopeSentence
	}
	return "Requires " + theScopes(scopes) + "." + typeNote(scopes)
}

// theScopes names one or more scopes, e.g. "the a.read scope" or "the
// a.read, b.read scopes".
func theScopes(scopes []string) string {
	if len(scopes) == 1 {
		return "the " + scopes[0] + " scope"
	}
	return "the " + strings.Join(scopes, ", ") + " scopes"
}

// optInSentence follows the scope sentence of a read with a write-capable
// scope, and of a group that holds one.
const optInSentence = "Opt-in: epx requests a write-capable scope for a read only when given --" +
	writeScopeFlag + "."

// optInNote returns what follows a scope sentence naming scopes: a new line
// with optInSentence when one of them is write-capable, else nothing.
func optInNote(scopes []string) string {
	if !writeCapable(scopes) {
		return ""
	}
	return "\n" + optInSentence
}

// writeCapable reports whether one of scopes may also change data. Only a
// scope named for reading counts as read-only: read or view is one of the
// words (separated by _) after its service, as in coupon.coupon_read,
// indexing.search_view or category.category_read_unpublished. Every other
// scope, such as coupon.coupon_manage, is write-capable, and the reads that
// request it are opt-in (requireOptIn).
func writeCapable(scopes []string) bool {
	return slices.ContainsFunc(scopes, func(scope string) bool {
		_, name, _ := strings.Cut(scope, ".")
		words := strings.Split(name, "_")
		return !slices.Contains(words, "read") && !slices.Contains(words, "view")
	})
}

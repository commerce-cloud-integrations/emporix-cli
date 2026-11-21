package cli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	scopesAnnotation      = "scopes"
	writeScopesAnnotation = "writeScopes"
)

func newGroup(use, short, long string, commands ...*cobra.Command) *cobra.Command {
	group := &cobra.Command{Use: use, Short: short, Long: long}
	group.AddCommand(commands...)
	if sentence := groupScopeSentence(group); sentence != "" {
		group.Long += "\n\n" + sentence
	}
	return group
}

type scopeKind struct {
	name     string
	sets     []string
	sentence func([]string) string
	optIn    bool
}

func groupScopeSentence(group *cobra.Command) string {
	kinds := slices.DeleteFunc([]scopeKind{
		{name: "reads", sets: scopeSets(group, scopesAnnotation), sentence: scopeSentence, optIn: true},
		{name: "writes", sets: scopeSets(group, writeScopesAnnotation), sentence: writeScopeSentence},
	}, func(kind scopeKind) bool { return len(kind.sets) == 0 })
	if len(kinds) == 1 && len(kinds[0].sets) == 1 {
		return kinds[0].sentence(strings.Fields(kinds[0].sets[0]))
	}
	sentences := make([]string, len(kinds))
	var named []string
	for i, kind := range kinds {
		sentences[i] = "Its " + kind.name + " " + groupRequirement(kind.sets)
		if kind.optIn {
			sentences[i] = joinSentences(sentences[i], optInNote(strings.Fields(strings.Join(kind.sets, " "))))
		}
		if len(kind.sets) == 1 {
			named = append(named, strings.Fields(kind.sets[0])...)
		}
	}
	return joinSentences(strings.Join(sentences, " "), typeNote(named))
}

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

func scopeSentence(scopes []string) string {
	return joinSentences(writeScopeSentence(scopes), optInNote(scopes))
}

const noScopeSentence = "No OAuth scope is needed."

func writeScopeSentence(scopes []string) string {
	if len(scopes) == 0 {
		return noScopeSentence
	}
	return joinSentences("Requires "+theScopes(scopes)+".", typeNote(scopes))
}

func theScopes(scopes []string) string {
	if len(scopes) == 1 {
		return "the " + scopes[0] + " scope"
	}
	return "the " + strings.Join(scopes, ", ") + " scopes"
}

const optInSentence = "Opt-in: epx requests a write-capable scope for a read only when given --" +
	writeScopeFlag + "."

func optInNote(scopes []string) string {
	if !writeCapable(scopes) {
		return ""
	}
	return optInSentence
}

func writeCapable(scopes []string) bool {
	return slices.ContainsFunc(scopes, func(scope string) bool {
		_, name, _ := strings.Cut(scope, ".")
		words := strings.Split(name, "_")
		return !slices.Contains(words, "read") && !slices.Contains(words, "view")
	})
}

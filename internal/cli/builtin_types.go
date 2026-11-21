package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type builtInType struct {
	name   string
	scopes []string
}

type builtInTypes struct {
	types []builtInType
	exact bool
}

func spelledTypes(types ...builtInType) builtInTypes { return builtInTypes{types: types, exact: true} }

func anyCaseTypes(names ...string) builtInTypes { return namedTypes(false, names...) }

func namedTypes(exact bool, names ...string) builtInTypes {
	types := make([]builtInType, len(names))
	for i, name := range names {
		types[i] = builtInType{name: name}
	}
	return builtInTypes{types: types, exact: exact}
}

func (b builtInTypes) names() []string {
	names := make([]string, len(b.types))
	for i, t := range b.types {
		names[i] = t.name
	}
	return names
}

func (b builtInTypes) kinds() string {
	switch {
	case len(b.types) == 0:
		return "a custom entity type"
	case b.exact:
		return alternatives(b.names()) + ", exactly as written, or a custom entity type"
	default:
		return alternatives(b.names()) + " in any case, or a custom entity type"
	}
}

func (b builtInTypes) clause(subject string) string {
	return subject + " is " + b.kinds() + ", whose ID `epx custom-entities list` shows"
}

func (b builtInTypes) sentence() string { return b.clause("A type") + "." }

func (b builtInTypes) find(arg string) (builtInType, bool) {
	at := slices.IndexFunc(b.types, func(t builtInType) bool {
		return t.name == arg || !b.exact && strings.EqualFold(t.name, arg)
	})
	if at < 0 {
		return builtInType{}, false
	}
	return b.types[at], true
}

func (b builtInTypes) spelling(arg string) error {
	if !b.exact {
		return nil
	}
	for _, t := range b.types {
		if t.name != arg && strings.EqualFold(t.name, arg) {
			return fmt.Errorf("use %s, as Emporix spells it", t.name)
		}
	}
	return nil
}

func (b builtInTypes) fill(scopes []string, typ string) []string {
	t, found := b.find(typ)
	if !found {
		return withType(scopes, typ)
	}
	return slices.Concat(slices.DeleteFunc(slices.Clone(scopes), namesType), t.scopes)
}

func (b builtInTypes) inPlaceOf(typed []string) string {
	var each []string
	for _, t := range b.types {
		if len(t.scopes) > 0 {
			each = append(each, t.name+" requires "+strings.Join(t.scopes, ", "))
		}
	}
	if len(typed) == 0 || len(each) == 0 {
		return ""
	}
	return "In place of " + strings.Join(typed, ", ") + ", " + strings.Join(each, "; ") + "."
}

func (b builtInTypes) withoutScopes() builtInTypes { return namedTypes(b.exact, b.names()...) }

func (b builtInTypes) values(use string, choices map[string][]string) map[string][]string {
	if len(b.types) == 0 {
		return choices
	}
	first := argumentNames(use, 1)[0]
	values := maps.Clone(choices)
	if values == nil {
		values = map[string][]string{}
	}
	values[first] = slices.Concat(values[first], b.names())
	return values
}

package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// builtInType is a type that Emporix defines, such as customer in the Search
// Service, by its name, with the scopes that it requests in place of those
// that name a custom entity type where a command's scopes name one.
type builtInType struct {
	name   string
	scopes []string
}

// builtInTypes are the built-in types that the first argument of a command
// takes besides the custom entity types, in the order that its help and
// completion name them, and how it takes them: exactly as written (a
// service that keeps what it stores under a type as written, so that
// CUSTOMER would address other items than customer), or in any case.
type builtInTypes struct {
	types []builtInType
	exact bool
}

// spelledTypes returns built-in types that a command takes exactly as
// written.
func spelledTypes(types ...builtInType) builtInTypes { return builtInTypes{types: types, exact: true} }

// anyCaseTypes returns built-in types, by their names, that a command takes
// in any case and that request no scopes of their own.
func anyCaseTypes(names ...string) builtInTypes {
	types := make([]builtInType, len(names))
	for i, name := range names {
		types[i] = builtInType{name: name}
	}
	return builtInTypes{types: types}
}

// names returns the names of b's types, in their order.
func (b builtInTypes) names() []string {
	names := make([]string, len(b.types))
	for i, t := range b.types {
		names[i] = t.name
	}
	return names
}

// kinds says what a type argument that takes b is, for help: one of b, and
// how it is written where b is exact, or a custom entity type.
func (b builtInTypes) kinds() string {
	if b.exact {
		return alternatives(b.names()) + ", exactly as written, or a custom entity type"
	}
	return alternatives(append(b.names(), "a custom entity type"))
}

// clause says, for help, that subject, such as TYPE, is what kinds says,
// and where the IDs of the custom entity types are; a sentence ends it.
func (b builtInTypes) clause(subject string) string {
	return subject + " is " + b.kinds() + ",\nwhose ID `epx custom-entities list` shows"
}

// sentence says, for help, what a type is (clause).
func (b builtInTypes) sentence() string { return b.clause("A type") + "." }

// find returns the type of b that arg names, as b takes it, and whether one
// does.
func (b builtInTypes) find(arg string) (builtInType, bool) {
	at := slices.IndexFunc(b.types, func(t builtInType) bool {
		return t.name == arg || !b.exact && strings.EqualFold(t.name, arg)
	})
	if at < 0 {
		return builtInType{}, false
	}
	return b.types[at], true
}

// spelling returns an error that names the type of an exact b that arg
// writes otherwise, such as CUSTOMER for customer; nil for any other arg.
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

// fill returns scopes, which a command requests for the type typ, its first
// argument: where typ is one of b, without those that name a custom entity
// type and with typ's own after the others; otherwise with typ filled in
// where they name one (withType).
func (b builtInTypes) fill(scopes []string, typ string) []string {
	t, found := b.find(typ)
	if !found {
		return withType(scopes, typ)
	}
	return slices.Concat(slices.DeleteFunc(slices.Clone(scopes), namesType), t.scopes)
}

// builtInSeparator separates, in a token of a scope annotation or of the
// scopes that a scope sentence names, a built-in type from a scope that it
// requests in place of those that name a custom entity type, as in
// customer:customer.customer_read (builtInTypes.tokens, parseScopeTokens).
const builtInSeparator = ":"

// tokens returns a token for each scope that one of b's types requests,
// TYPE:SCOPE (builtInSeparator).
func (b builtInTypes) tokens() []string {
	var tokens []string
	for _, t := range b.types {
		for _, scope := range t.scopes {
			tokens = append(tokens, t.name+builtInSeparator+scope)
		}
	}
	return tokens
}

// parseScopeTokens returns tokens, the scopes that a command's annotation or
// scope sentence names, as the scopes that the command requests as declared
// and, apart, the built-in types whose scopes its tokens hold
// (builtInTypes.tokens), each once, with each of its scopes once, in their
// order. Tokens do not say how a command takes the types (builtInTypes.exact).
func parseScopeTokens(tokens []string) (scopes []string, types []builtInType) {
	for _, token := range tokens {
		name, scope, isBuiltIn := strings.Cut(token, builtInSeparator)
		if !isBuiltIn {
			scopes = append(scopes, token)
			continue
		}
		at := slices.IndexFunc(types, func(t builtInType) bool { return t.name == name })
		if at < 0 {
			at = len(types)
			types = append(types, builtInType{name: name})
		}
		if !slices.Contains(types[at].scopes, scope) {
			types[at].scopes = append(types[at].scopes, scope)
		}
	}
	return scopes, types
}

// values returns choices, the values that the arguments of a command whose
// usage is use take by their names, with the names of b's types as values of
// its first argument too (completeArguments).
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

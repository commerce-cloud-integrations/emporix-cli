package cli

import (
	"cmp"
	"fmt"

	"github.com/spf13/cobra"
)

// languageMode is whether a write's body holds localized text and, where it
// does, how the write tells Emporix the language of that text, as its service
// reads it: with --content-language, which names it, and the Content-Language
// header, which says so to Emporix.
type languageMode int

const (
	// notLocalized: the body holds no localized text; no Content-Language.
	notLocalized languageMode = iota
	// stringsOrMaps: --content-language LANGUAGE with strings in that
	// language, or everyLanguage, the default, with maps of translations; the
	// Content-Language header names it.
	stringsOrMaps
	// headerlessMaps: as stringsOrMaps, but maps of translations go without a
	// Content-Language header, as the service's reference documents: it
	// refuses them sent with Content-Language: *.
	headerlessMaps
	// mapsOnly: maps of translations only, always sent with Content-Language:
	// *, and no --content-language, for a service that refuses a string in
	// any language or stores it in none.
	mapsOnly
)

// bindLanguage binds --content-language to cmd where w's language lets its
// localized text be given in one language, adds what its help says of that
// text, and returns the language that it gives, everyLanguage unless it says
// otherwise. A languageHelp of its own replaces only the help of
// stringsOrMaps or mapsOnly: on another write it would not say how the write
// sends its text, a mistake in its declaration, so it panics.
func (w *write) bindLanguage(cmd *cobra.Command) *contentLanguage {
	lang := contentLanguage(everyLanguage)
	if w.languageHelp != "" && (w.language == notLocalized || w.language == headerlessMaps) {
		panic(w.use + ": a write of headerless maps or without localized text has no language help of its own")
	}
	switch w.language {
	case notLocalized: // no flag, and nothing in its help
	case mapsOnly:
		cmd.Long += "\n\n" + cmp.Or(w.languageHelp, mapsOnlyHelp)
	case headerlessMaps:
		cmd.Flags().Var(&lang, contentLanguageFlag, languageUsage+", sent without Content-Language")
		cmd.Long += "\n\n" + headerlessMapsHelp
	case stringsOrMaps:
		cmd.Flags().Var(&lang, contentLanguageFlag, languageUsage)
		cmd.Long += "\n\n" + cmp.Or(w.languageHelp, contentLanguageHelp)
	}
	return &lang
}

// languageHeader returns the Content-Language that w sends with lang, its
// --content-language (everyLanguage where w takes none): none where its body
// holds no localized text, or where lang is everyLanguage and w sends maps of
// translations without one (headerlessMaps); lang otherwise.
func (w *write) languageHeader(lang contentLanguage) string {
	switch {
	case w.language == notLocalized, w.language == headerlessMaps && lang == everyLanguage:
		return ""
	default:
		return string(lang)
	}
}

// languageUsage is the usage of --content-language.
const languageUsage = "the `LANGUAGE` of localized text in the body, or " + everyLanguage + " for maps of translations"

// contentLanguageFlag is the flag of a write whose body holds localized text,
// which Emporix reads in the language that its Content-Language header names.
const contentLanguageFlag = "content-language"

// everyLanguage is the --content-language of a body that gives each localized
// text as a map of translations, the default.
const everyLanguage = "*"

// localizedMaps begins what the help of a write whose body holds localized
// text says of it.
const localizedMaps = "Localized text in the body, such as a name, is a map of translations, e.g.\n"

// contentLanguageHelp follows what a write whose body holds localized text
// sends, in its help.
const contentLanguageHelp = localizedMaps +
	`{"en": "gram", "de": "Gramm"}, as -o json of a read prints it, or with --` + contentLanguageFlag +
	"\nLANGUAGE a string in that language, e.g. \"gram\" with --" + contentLanguageFlag + " en."

// headerlessMapsHelp follows what a write of headerlessMaps sends,
// in its help: its service reads maps of translations sent without a
// Content-Language header, and refuses them with Content-Language: *.
const headerlessMapsHelp = localizedMaps +
	`{"en": "Netherlands", "nl": "Nederland"}, as -o json of a read prints it, sent without a` + "\n" +
	"Content-Language header (--" + contentLanguageFlag + " " + everyLanguage + ", the default), or with\n--" +
	contentLanguageFlag + " LANGUAGE a string in that language, sent with that Content-Language,\ne.g. " +
	`"Netherlands" with --` + contentLanguageFlag + " en."

// mapsSent says how a write of mapsOnly sends localized text, which the help
// of each such write begins with. It names no member, unlike localizedMaps: a
// member such as a brand's name may be plain text, which the service takes as
// a string.
const mapsSent = "Localized text in the body is a map of translations, e.g. " +
	`{"en": "Acme", "de": "Acme"},` + "\nas -o json of a read prints it, sent with Content-Language: " +
	everyLanguage + "."

// mapsOnlyHelp follows what a write of mapsOnly sends, in its help, where its
// service refuses a string.
const mapsOnlyHelp = mapsSent + " Emporix refuses a string in\nits place."

// contentLanguage is a --content-language value: everyLanguage, or a language
// code, lower-cased as a --language value is; it implements pflag.Value, so
// another value is rejected while flags are parsed.
type contentLanguage string

// Set implements pflag.Value.
func (l *contentLanguage) Set(value string) error {
	if value == everyLanguage {
		*l = everyLanguage
		return nil
	}
	var code language
	if err := code.Set(value); err != nil {
		return fmt.Errorf("invalid language %q: use %s or a code such as nl or en", value, everyLanguage)
	}
	*l = contentLanguage(code)
	return nil
}

func (l *contentLanguage) String() string { return string(*l) }

// Type implements pflag.Value.
func (l *contentLanguage) Type() string { return "language" }

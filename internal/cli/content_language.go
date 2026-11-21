package cli

import (
	"cmp"
	"fmt"

	"github.com/spf13/cobra"
)

type languageMode int

const (
	notLocalized languageMode = iota
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

func (w *write) bindLanguage(cmd *cobra.Command) *contentLanguage {
	lang := contentLanguage(everyLanguage)
	if w.languageHelp != "" && (w.language == notLocalized || w.language == headerlessMaps) {
		panic(w.use + ": a write of headerless maps or without localized text has no language help of its own")
	}
	switch w.language {
	case notLocalized:
	case mapsOnly:
		cmd.Long += "\n\n" + cmp.Or(w.languageHelp, mapsOnlyHelp())
	case headerlessMaps:
		cmd.Flags().Var(&lang, contentLanguageFlag, languageUsage+", sent without Content-Language")
		cmd.Long += "\n\n" + headerlessMapsHelp()
	case stringsOrMaps:
		cmd.Flags().Var(&lang, contentLanguageFlag, languageUsage)
		cmd.Long += "\n\n" + cmp.Or(w.languageHelp, contentLanguageHelp())
	}
	return &lang
}

func (w *write) languageHeader(lang contentLanguage) string {
	switch {
	case w.language == notLocalized, w.language == headerlessMaps && lang == everyLanguage:
		return ""
	default:
		return string(lang)
	}
}

const languageUsage = "the `LANGUAGE` of localized text in the body, sent as typed: write it as the tenant's " +
	"language configuration does; or " + everyLanguage + " for maps of translations"

const contentLanguageFlag = "content-language"

const everyLanguage = "*"

const localizedMaps = "Localized text in the body, such as a name, is a map of translations, e.g."

func contentLanguageHelp() string {
	return localizedMaps + " " +
		literal(`{"en": "gram", "de": "Gramm"}`) + ", as `-o json` of a read prints it, or with --" + contentLanguageFlag +
		" LANGUAGE a string in that language, e.g. \"gram\" with --" + contentLanguageFlag + " en."
}

func headerlessMapsHelp() string {
	return localizedMaps + " " +
		literal(`{"en": "Netherlands", "nl": "Nederland"}`) + ", as `-o json` of a read prints it, sent without a " +
		"Content-Language header (--" + contentLanguageFlag + " " + everyLanguage + ", the default), or with --" +
		contentLanguageFlag + " LANGUAGE a string in that language, sent with that Content-Language, e.g. " +
		`"Netherlands" with --` + contentLanguageFlag + " en."
}

func mapsSent() string {
	return "Localized text in the body is a map of translations, e.g. " +
		literal(`{"en": "Acme", "de": "Acme"}`) + "," + " as `-o json` of a read prints it, sent with Content-Language: " +
		everyLanguage + "."
}

func mapsOnlyHelp() string {
	return mapsSent() + " Emporix refuses a string in its place."
}

type contentLanguage string

func (l *contentLanguage) Set(value string) error {
	if value == everyLanguage {
		*l = everyLanguage
		return nil
	}
	var code language
	if err := code.Set(value); err != nil {
		return fmt.Errorf("invalid language %q: use %s or %s", value, everyLanguage, languageShape)
	}
	*l = contentLanguage(code)
	return nil
}

func (l *contentLanguage) String() string { return string(*l) }

func (l *contentLanguage) Type() string { return "language" }

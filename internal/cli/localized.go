package cli

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// languagePattern matches language codes as Emporix checks them, in any
// letter case: a language of 2 or 3 ASCII letters, then any subtags of 2 to 8
// ASCII letters or digits, each after a hyphen, such as "nl", "en-GB",
// "zh-Hant-TW" or a custom "de-CH-ZH"; whether the language or region exists
// is not checked. It checks the value as typed; where epx sends a code
// (--content-language, quotes pdf --language), it sends it as typed.
var languagePattern = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// languageShape says what a language code is, in an error message.
const languageShape = "a code such as nl, en or en-GB: 2 or 3 letters a to z, then any parts of 2 to 8 " +
	"letters a to z or digits, each after a hyphen"

// translationKey returns the key under which localized keeps the translation
// of code: lower-cased, so a table finds it in any letter case.
func translationKey(code string) string { return strings.ToLower(code) }

// localized is text that Emporix returns either as a plain string or, when
// every translation is requested, as a map from language code to text.
type localized map[string]string

// UnmarshalJSON accepts both shapes; a plain string is stored under "".
// Language codes are keyed by translationKey (on a clash the alphabetically
// first original key wins), and null or empty translations are dropped, so
// they never hide another one.
func (l *localized) UnmarshalJSON(data []byte) error {
	var texts map[string]string
	if err := json.Unmarshal(data, &texts); err != nil {
		var text string
		if json.Unmarshal(data, &text) != nil {
			return err
		}
		texts = map[string]string{"": text}
	}
	*l = localized{}
	for _, key := range slices.Sorted(maps.Keys(texts)) {
		code := translationKey(key)
		if _, taken := (*l)[code]; !taken && texts[key] != "" {
			(*l)[code] = texts[key]
		}
	}
	return nil
}

// in returns the text for language in any letter case (translationKey), else
// the plain text, else the alphabetically first translation labeled with its
// language, so a substitute is never mistaken for the requested language.
func (l localized) in(lang language) string {
	if text := cmp.Or(l[translationKey(string(lang))], l[""]); text != "" {
		return text
	}
	if len(l) == 0 {
		return ""
	}
	first := slices.Min(slices.Collect(maps.Keys(l)))
	return l[first] + " (" + first + ")"
}

// plainAndLocalizedName is the name of services that keep a plain name (one
// translation in an unknown language) next to a map of localized names.
type plainAndLocalizedName struct {
	Name          string    `json:"name"`
	LocalizedName localized `json:"localizedName"`
}

// displayName returns the localized name in lang (labeled when another
// language substitutes it); the plain name is only a last resort.
func (n plainAndLocalizedName) displayName(lang language) string {
	return cmp.Or(n.LocalizedName.in(lang), n.Name)
}

// language is a --language value, as typed; it implements pflag.Value, so
// codes are validated while flags are parsed.
type language string

// defaultLanguage is the --language default of every resource that has one.
const defaultLanguage language = "en"

// Set implements pflag.Value.
func (l *language) Set(value string) error {
	if !languagePattern.MatchString(value) {
		return fmt.Errorf("invalid language %q: use %s", value, languageShape)
	}
	*l = language(value)
	return nil
}

func (l *language) String() string { return string(*l) }

// Type implements pflag.Value.
func (l *language) Type() string { return "language" }

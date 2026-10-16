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

// languagePattern matches lower-cased BCP-47 language tags (language, optional
// script, region and variants), such as "nl", "en-gb" or "zh-hant-tw".
var languagePattern = regexp.MustCompile(
	`^[a-z]{2,3}(-[a-z]{4})?(-([a-z]{2}|[0-9]{3}))?(-([a-z0-9]{5,8}|[0-9][a-z0-9]{3}))*$`)

// localized is text that Emporix returns either as a plain string or, when
// every translation is requested, as a map from language code to text.
type localized map[string]string

// UnmarshalJSON accepts both shapes; a plain string is stored under "".
// Language codes are lower-cased (tags compare case-insensitively; on a clash
// the alphabetically first original key wins), and null or empty translations
// are dropped, so they never hide another one.
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
		lower := strings.ToLower(key)
		if _, taken := (*l)[lower]; !taken && texts[key] != "" {
			(*l)[lower] = texts[key]
		}
	}
	return nil
}

// in returns the text for language, else the plain text, else the
// alphabetically first translation labeled with its language, so a
// substitute is never mistaken for the requested language.
func (l localized) in(lang language) string {
	if text := cmp.Or(l[string(lang)], l[""]); text != "" {
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

// language is a --language value; it implements pflag.Value, so codes are
// validated and lower-cased while flags are parsed.
type language string

// defaultLanguage is the --language default of every resource that has one.
const defaultLanguage language = "en"

// Set implements pflag.Value.
func (l *language) Set(value string) error {
	value = strings.ToLower(value)
	if !languagePattern.MatchString(value) {
		return fmt.Errorf("invalid language %q: use a code such as nl or en", value)
	}
	*l = language(value)
	return nil
}

func (l *language) String() string { return string(*l) }

// Type implements pflag.Value.
func (l *language) Type() string { return "language" }

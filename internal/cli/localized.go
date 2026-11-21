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

// Emporix checks only the shape of a language code, in any letter case;
// whether the language or region exists is not checked.
var languagePattern = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

const languageShape = "a code such as nl, en or en-GB: 2 or 3 letters a to z, then any parts of 2 to 8 " +
	"letters a to z or digits, each after a hyphen"

func translationKey(code string) string { return strings.ToLower(code) }

// Emporix returns localized text either as a plain string or, when every
// translation is requested, as a map from language code to text.
type localized map[string]string

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

type plainAndLocalizedName struct {
	Name          string    `json:"name"`
	LocalizedName localized `json:"localizedName"`
}

func (n plainAndLocalizedName) displayName(lang language) string {
	return cmp.Or(n.LocalizedName.in(lang), n.Name)
}

type language string

const defaultLanguage language = "en"

func (l *language) Set(value string) error {
	if !languagePattern.MatchString(value) {
		return fmt.Errorf("invalid language %q: use %s", value, languageShape)
	}
	*l = language(value)
	return nil
}

func (l *language) String() string { return string(*l) }

func (l *language) Type() string { return "language" }

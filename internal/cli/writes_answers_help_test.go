package cli

import (
	"strings"
	"testing"
)

const unknownIDDeleted = "Emporix answers an unknown ID as a deleted one (204), so epx exits with 0."

func TestWriteHelpSaysHowEmporixAnswers(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want []string
	}{
		{[]string{"sites", "mixins", "update"}, []string{
			"the mixin's own fields, as `epx sites mixins get SITE_CODE NAME -o json` prints them",
			"metadata.schema is kept when left out",
			"Emporix stores that object as one more field of the mixin",
		}},
		{[]string{"sites", "mixins", "patch"}, []string{"a nested object merges member by member"}},
		{[]string{"sites", "update"}, []string{
			"Emporix replaces only the active ones: an inactive one in the body is dropped",
			"--expand NAME:all, e.g. payment:all, replaces the whole list of NAME",
			"active and none change it as no --expand does",
			"only all changes inactive ones",
		}},
		{[]string{"sites", "create"}, []string{"refuses more than one active tax configuration (400 Bad Request)"}},
		{[]string{"ai", "jobs", "delete"}, []string{"shows. " + unknownIDDeleted}},
		{[]string{"ai", "tools", "delete"}, []string{unknownIDDeleted}},
		{[]string{"indexing", "reindex-jobs", "create"}, []string{
			"accepts a TYPE that it does not know too and ends that job SUCCESS with nothing indexed",
		}},
		{[]string{"schemas", "references", "delete"}, []string{
			"with every version of it. The JSON schema files of its versions stay where Emporix stored them",
		}},
		{[]string{"products", "get"}, []string{
			"arrays, a localized attribute's list of translations too, and other values are replaced, not merged",
		}},
	} {
		help := helpText(t, tt.args...)
		for _, want := range tt.want {
			if !strings.Contains(help, want) {
				t.Errorf("%s: help lacks %q:\n%s", strings.Join(tt.args, " "), want, help)
			}
		}
	}
}

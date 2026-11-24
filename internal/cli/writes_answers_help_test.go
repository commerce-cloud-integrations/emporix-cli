package cli

import (
	"strings"
	"testing"
)

const unknownIDDeleted = "Emporix answers an unknown ID as a deleted one (204), so epx exits with 0."

func TestHelpSaysHowEmporixAnswers(t *testing.T) {
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
		{[]string{"sites", "mixins", "list"}, []string{
			"Emporix serves the mixins of an inactive site only with the write-capable site.site_manage scope",
			"for an inactive site it answers 403 (exit 3), also for an unknown NAME",
		}},
		{[]string{"sites", "mixins", "get"}, []string{"for an inactive site it answers 403 (exit 3)"}},
		{[]string{"ai", "rag", "reindex"}, []string{
			"starts a reindex job for it that `epx indexing reindex-jobs list` shows and that ends SUCCESS with nothing indexed",
		}},
		{[]string{"schemas", "references", "list"}, []string{
			"--query finds a reference by ID as _id:ID (id:ID finds none)", "Emporix ignores a sort by id",
		}},
		{[]string{"labels", "create"}, []string{
			"Its name is a string (Emporix refuses a map of translations, 400 Bad Request)",
		}},
		{[]string{"labels", "update"}, []string{"description is a string or a map of translations"}},
		{[]string{"custom-entities", "instances", "create"}, []string{
			"to an order fails with 400 Bad Request whether the item exists or not",
			"one to a product finds published products only",
		}},
		{[]string{"custom-entities", "instances", "bulk-create"}, []string{
			"to an order or to an unpublished product fails with 500 Internal Server Error (Emporix does not write it)",
		}},
		{[]string{"fees", "assignments", "delete-fee"}, []string{"from the item fee of the item YRN on the site of the fee"}},
		{[]string{"search", "queries", "update"}, []string{
			"without it Emporix refuses an existing ID (400 Bad Request, exit 1)", "A read prints the index as indexId",
		}},
		{[]string{"search", "indexes", "delete"}, []string{"An unknown one is not found (exit 4)"}},
		{[]string{"search", "documents"}, []string{
			"--sort sorts only by fields that the index holds with exact true",
			"Emporix orders documents of the same SCORE differently on each request",
			"can repeat one on a later page: epx then fails with exit 1 (with --stream, after printing the pages before it)",
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

func TestListFilterHelpOnlyOnTheList(t *testing.T) {
	if help := helpText(t, "schemas", "references", "get"); strings.Contains(help, "_id:ID") {
		t.Errorf("get help names a --query filter that get does not take:\n%s", help)
	}
}

package cli

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type iamScope struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	Predefined  *bool     `json:"predefined"`
	Description localized `json:"description"`
}

type iamAccessControl struct {
	ID         string            `json:"id"`
	Name       localized         `json:"name"`
	Scopes     []json.RawMessage `json:"scopes"`
	Predefined *bool             `json:"predefined"`
}

type iamTemplate struct {
	ID             string            `json:"id"`
	Name           localized         `json:"name"`
	AccessControls []json.RawMessage `json:"accessControls"`
}

const accessControlColumns = "SCOPES counts the scopes an access control grants (`-o json` lists them)."

const iamPagingNote = "list sorts by id unless --sort says otherwise; end another order with ,id:asc (e.g. " +
	"name.en:asc,id:asc), as Emporix wants a direction on every field, so pages do not overlap."

// Unsorted, Emporix orders scopes, templates and groups differently on every
// request, so pages overlap and miss items. Sorted by id (live, also for access
// controls, whose reference documents no sort) every IAM list pages cleanly.
func iamList[T any](use, scope, long string, filters listFilters, v view[T]) resource[T] {
	return resource[T]{
		use:       use,
		noun:      "IAM " + strings.ReplaceAll(use, "-", " "),
		under:     iamUse,
		long:      long,
		listNote:  iamPagingNote,
		path:      "/iam/{tenant}/" + use,
		listQuery: url.Values{sortParameter: {"id:asc"}},
		scopes:    []string{scope},
		filters:   filters,
		view:      v,
	}
}

func iamScopes() resource[iamScope] {
	return iamList("scopes", "iam.scope_read",
		"Reads the OAuth scopes the tenant knows: ID is the scope, DOMAIN the service it belongs to "+
			"and PREDEFINED whether Emporix defines it.", listFilters{kinds: filterable},
		view[iamScope]{
			singular:  "scope",
			header:    []string{columnID, "DOMAIN", "PREDEFINED", columnDescription},
			localized: true,
			row: func(s iamScope, lang language) []string {
				return []string{s.ID, s.Domain, formatBool(s.Predefined), s.Description.in(lang)}
			},
		})
}

func iamAccessControls() resource[iamAccessControl] {
	return iamList("access-controls", "iam.access_read", "Reads IAM access controls. "+accessControlColumns,
		listFilters{kinds: filterable},
		view[iamAccessControl]{
			singular:  "access control",
			header:    []string{columnID, columnName, "SCOPES", "PREDEFINED"},
			localized: true,
			row: func(c iamAccessControl, lang language) []string {
				return []string{c.ID, c.Name.in(lang), strconv.Itoa(len(c.Scopes)), formatBool(c.Predefined)}
			},
		})
}

func newIAMTemplatesCommand(a *app) *cobra.Command {
	templates := iamList("templates", "iam.template_read",
		"Reads IAM access control templates: sets of access controls that groups are built from "+
			"(`-o json` lists them).", listFilters{kinds: sortable},
		view[iamTemplate]{
			singular:  "template",
			header:    []string{columnID, columnName, "ACCESS CONTROLS"},
			localized: true,
			row: func(t iamTemplate, lang language) []string {
				return []string{t.ID, t.Name.in(lang), strconv.Itoa(len(t.AccessControls))}
			},
		})
	return newGroup(templates.use, "Read "+templates.nouns(), templates.long, templates.listCommand(a))
}

package cli

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// iamRole is the part of an IAM role shown in tables.
type iamRole struct {
	ID          string            `json:"id"`
	Name        localized         `json:"name"`
	Permissions []json.RawMessage `json:"permissions"` // counted only
}

// iamPermission is the part of an IAM permission shown in tables.
type iamPermission struct {
	ID                  string            `json:"id"`
	Code                string            `json:"code"`
	Name                localized         `json:"name"`
	ApplicableResources []json.RawMessage `json:"applicableResources"` // counted only
}

// iamScope is the part of an IAM scope shown in tables; its id is its name.
type iamScope struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	Predefined  *bool     `json:"predefined"`
	Description localized `json:"description"`
}

// iamResource is the part of an IAM resource shown in tables.
type iamResource struct {
	ID   string    `json:"id"`
	Code string    `json:"code"`
	Name localized `json:"name"`
}

// iamAccessControl is the part of an IAM access control shown in tables.
type iamAccessControl struct {
	ID         string            `json:"id"`
	Name       localized         `json:"name"`
	ResourceID string            `json:"resourceId"`
	RoleID     string            `json:"roleId"`
	Scopes     []json.RawMessage `json:"scopes"` // counted only
	Predefined *bool             `json:"predefined"`
}

// iamTemplate is the part of an IAM access control template shown in tables.
type iamTemplate struct {
	ID             string            `json:"id"`
	Name           localized         `json:"name"`
	AccessControls []json.RawMessage `json:"accessControls"` // counted only
}

// What the columns of an IAM permission and access control hold, also in the
// reads of one group or user.
const (
	permissionColumns    = "RESOURCES counts the resources a permission applies to."
	accessControlColumns = "SCOPES counts the scopes an access control grants (-o json lists them); RESOURCE\n" +
		"and ROLE are the ids of the resource and role older access controls refer to."
)

// deprecatedNote ends the help of the reads whose endpoints the API reference
// marks as deprecated.
const deprecatedNote = "Emporix has deprecated the endpoints this reads; they may be removed."

// iamPagingNote ends the help of the IAM lists that sort by id.
const iamPagingNote = "\nlist sorts by id unless --sort says otherwise; end another order with ,id:asc (e.g.\n" +
	"name.en:asc,id:asc), as Emporix wants a direction on every field, so pages do not overlap."

// iamList returns the IAM collection use, read with scope. Unsorted, Emporix
// orders roles, permissions, scopes and templates differently on every
// request, so their pages overlap and miss items; sorted by id (live, also for
// access controls, whose reference documents no sort) every IAM list pages
// cleanly, so each sorts by id unless --sort says otherwise.
func iamList[T any](use, scope, long string, filters listFilters, v view[T]) resource[T] {
	return resource[T]{
		use:       use,
		noun:      "IAM " + strings.ReplaceAll(use, "-", " "),
		under:     iamUse,
		long:      long + iamPagingNote,
		path:      "/iam/{tenant}/" + use,
		listQuery: url.Values{sortParameter: {"id:asc"}},
		scopes:    []string{scope},
		filters:   filters,
		view:      v,
	}
}

// iamRoles returns IAM roles.
func iamRoles() resource[iamRole] {
	return iamList("roles", "iam.role_read",
		"Reads IAM roles: PERMISSIONS counts the permissions a role grants (-o json lists them).\n"+deprecatedNote,
		listFilters{kinds: sortable},
		view[iamRole]{
			singular:  "role",
			header:    []string{columnID, columnName, "PERMISSIONS"},
			localized: true,
			row: func(r iamRole, lang language) []string {
				return []string{r.ID, r.Name.in(lang), strconv.Itoa(len(r.Permissions))}
			},
		})
}

// iamPermissions returns IAM permissions.
func iamPermissions() resource[iamPermission] {
	return iamList("permissions", "iam.permission_read", "Reads IAM permissions: "+permissionColumns+"\n"+deprecatedNote,
		listFilters{kinds: sortable},
		view[iamPermission]{
			singular:  "permission",
			header:    []string{columnID, columnCode, columnName, "RESOURCES"},
			localized: true,
			row: func(p iamPermission, lang language) []string {
				return []string{p.ID, p.Code, p.Name.in(lang), strconv.Itoa(len(p.ApplicableResources))}
			},
		})
}

// iamScopes returns the OAuth scopes the tenant knows.
func iamScopes() resource[iamScope] {
	return iamList("scopes", "iam.scope_read",
		"Reads the OAuth scopes the tenant knows: ID is the scope, DOMAIN the service it belongs to\n"+
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

// iamResources returns IAM resources.
func iamResources() resource[iamResource] {
	return iamList("resources", "iam.resource_read",
		"Reads IAM resources, what older access controls grant a role on.\n"+deprecatedNote, listFilters{kinds: sortable},
		view[iamResource]{
			singular:  "resource",
			header:    []string{columnID, columnCode, columnName},
			localized: true,
			row:       func(r iamResource, lang language) []string { return []string{r.ID, r.Code, r.Name.in(lang)} },
		})
}

// iamAccessControls returns IAM access controls.
func iamAccessControls() resource[iamAccessControl] {
	return iamList("access-controls", "iam.access_read", "Reads IAM access controls.\n"+accessControlColumns,
		listFilters{kinds: filterable},
		view[iamAccessControl]{
			singular:  "access control",
			header:    []string{columnID, columnName, "SCOPES", "PREDEFINED", "RESOURCE", "ROLE"},
			localized: true,
			row: func(c iamAccessControl, lang language) []string {
				return []string{
					c.ID, c.Name.in(lang), strconv.Itoa(len(c.Scopes)), formatBool(c.Predefined), c.ResourceID, c.RoleID,
				}
			},
		})
}

// newIAMTemplatesCommand lists access control templates, which Emporix only
// lists (it reads none by id).
func newIAMTemplatesCommand(a *app) *cobra.Command {
	templates := iamList("templates", "iam.template_read",
		"Reads IAM access control templates: sets of access controls that groups are built from\n"+
			"(-o json lists them).", listFilters{kinds: sortable},
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

package cli

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"
)

// iamGroup is the part of an IAM user group shown in tables.
type iamGroup struct {
	ID             string            `json:"id"`
	Code           string            `json:"code"`
	Name           localized         `json:"name"`
	UserType       string            `json:"userType"`
	AccessControls []json.RawMessage `json:"accessControls"` // counted only
}

// groupMember is a user's membership of a group.
type groupMember struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	UserType string `json:"userType"`
}

// iamGroups returns the tenant's user groups, which a user's groups are read
// as too.
func iamGroups() resource[iamGroup] {
	return resource[iamGroup]{
		use:   groupsUse,
		noun:  "IAM groups",
		short: "Read, create, upsert and delete IAM groups and their users",
		under: iamUse,
		long: "Reads the tenant's user groups. TYPE is the kind of user a group is for (e.g. EMPLOYEE,\n" +
			"CUSTOMER), and ACCESS CONTROLS counts the access controls it grants (-o json lists their\n" +
			"ids, `epx iam groups access-controls` reads them).",
		about: "Reads the tenant's user groups, their users and the access controls they grant, and\n" +
			"changes groups and their users." + writesChangeAccess,
		path:    "/iam/{tenant}/groups",
		scopes:  []string{"iam.group_read"},
		filters: listFilters{kinds: filterable},
		view: view[iamGroup]{
			singular:  "group",
			header:    []string{columnID, columnCode, columnName, columnType, "ACCESS CONTROLS"},
			localized: true,
			row: func(g iamGroup, lang language) []string {
				return []string{g.ID, g.Code, g.Name.in(lang), g.UserType, strconv.Itoa(len(g.AccessControls))}
			},
		},
	}
}

// newIAMGroupsCommand reads, creates, upserts and deletes the tenant's user
// groups, reads and changes their members and reads the access controls they
// grant.
func newIAMGroupsCommand(a *app) *cobra.Command {
	groups, accessControls := iamGroups(), iamAccessControls().reader()
	extra := append(iamGroupWrites(a, groups), newGroupUsersCommand(a, groups),
		childCommand(a, groups, child[iamAccessControl]{
			use:   "access-controls",
			short: "List the access controls of a group",
			long: "Reads the access controls a group grants its users.\n" + accessControlColumns + "\n" +
				"A group's ACCESS CONTROLS in `epx iam groups list` may also count removed access controls,\n" +
				"which this leaves out.",
			scopes: accessControls.scopes,
			fetch:  listing{},
			view:   accessControls.view,
		}))
	return newResourceCommand(a, groups, extra...)
}

// newGroupUsersCommand returns `users`, the group of the reads and writes of
// the members of a group of groups.
func newGroupUsersCommand(a *app, groups resource[iamGroup]) *cobra.Command {
	membersOf := groups.nested("users")
	members := child[groupMember]{
		use:     listUse,
		sub:     membersOf.use,
		short:   "List the users of a group",
		long:    "Reads the members of a group: USER is a user's id and TYPE the kind of user.",
		scopes:  []string{"iam.user_read"},
		fetch:   listing{},
		unknown: listsNothing,
		view: view[groupMember]{
			singular: "group member",
			header:   []string{columnID, "USER", columnType},
			row: func(m groupMember, _ language) []string {
				return []string{m.ID, m.UserID, m.UserType}
			},
		},
	}
	return newGroup(membersOf.use, "Read the users of a group, add and remove them",
		"Reads the users of a group, adds users to it and removes them from it."+writesChangeAccess,
		append([]*cobra.Command{childCommand(a, membersOf, members)}, groupUserWrites(a, membersOf, members)...)...)
}

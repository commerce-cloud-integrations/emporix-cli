package cli

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// iamUser is the part of an IAM user (an employee) shown in tables.
type iamUser struct {
	personName
	ID              string   `json:"id"`
	ContactEmail    string   `json:"contactEmail"`
	Status          string   `json:"status"`
	IsAccountLocked *bool    `json:"isAccountLocked"`
	GroupIDs        []string `json:"groupIds"`
}

// iamUsersPage is the page the IAM users are read in: the IAM Service orders
// them differently on every request and fails on most second pages, so they
// are read as one page, which accepts up to 500 users. Its X-Total-Count is the
// number of users (live: also with a page of 1), so a page cut short fails.
const iamUsersPage = 500

// userAccessControlsPage is the page a user's access controls are read in:
// Emporix lists them on one page only (live: every later page is empty,
// whatever its size), so they are read as one page of up to 1000, more than a
// tenant's access controls (a few hundred live, the predefined ones
// included). Its X-Total-Count counts them (live), and the read requires it,
// as an empty next page proves nothing there: a page cut short fails.
const userAccessControlsPage = 1000

// newIAMUsersCommand reads the tenant's users, their groups, what they may do
// and the users of a vendor.
func newIAMUsersCommand(a *app) *cobra.Command {
	users := resource[iamUser]{
		use:   "users",
		noun:  "IAM users",
		short: "Read, create, replace and delete IAM users; read what they may do",
		under: iamUse,
		long: "Reads the tenant's users (such as employees), personal data included. STATUS and LOCKED\n" +
			"tell whether a user can sign in, and GROUPS counts the groups a user belongs to\n" +
			"(`epx iam users groups list` lists them). Emporix orders users differently on every\n" +
			"request and fails on most second pages, so list reads them as " + onePage(iamUsersPage) + ": the\n" +
			"order, and the users --limit keeps, vary per run, and a tenant with more users fails.",
		path:   "/iam/{tenant}/users",
		scopes: []string{"iam.user_read"},
		about: "Reads the tenant's users (such as employees), personal data included, their groups and what\n" +
			"they may do, and the users of a vendor; creates, replaces and deletes users and removes them\n" +
			"from their groups." + writesChangeAccess,
		read: inOnePage(iamUsersPage),
		view: view[iamUser]{
			singular: "user",
			header:   []string{columnID, columnName, columnEmail, columnStatus, "LOCKED", "GROUPS"},
			row: func(u iamUser, _ language) []string {
				return []string{
					u.ID, u.fullName(), u.ContactEmail, u.Status, formatBool(u.IsAccountLocked),
					strconv.Itoa(len(u.GroupIDs)),
				}
			},
		},
	}
	return newResourceCommand(a, users, slices.Concat(userWrites(a, users),
		[]*cobra.Command{newUserGroupsCommand(a, users)}, userAccessReads(a, users),
		[]*cobra.Command{vendorUsers(a, users)})...)
}

// newUserGroupsCommand returns `groups`, the group of the reads of the groups
// a user of users belongs to, all of them and one, and of the write that
// removes the user from all of them.
func newUserGroupsCommand(a *app, users resource[iamUser]) *cobra.Command {
	groups := iamGroups()
	groupsOf := users.nested(groups.use)
	userGroups := child[iamGroup]{
		use:     listUse,
		sub:     groupsOf.use,
		short:   "List the groups of a user",
		long:    "Reads the groups a user belongs to.",
		scopes:  groups.scopes,
		fetch:   listing{},
		unknown: listsNothing,
		view:    groups.view,
	}
	return newGroup(groupsOf.use, "Read the groups of a user, remove a user from them",
		"Reads the groups a user belongs to, all or one, and removes a user from every group."+writesChangeAccess,
		childCommand(a, groupsOf, userGroups),
		childItemCommand(a, groupsOf, userGroups, &childItem{use: getUse, short: "Show one group of a user",
			long: "Reads one group a user belongs to; one the user does not belong to is not found."}),
		userGroupsWrite(a, groupsOf, userGroups))
}

// userAccessReads returns the reads of what a user may do: the scopes, all
// access controls and those on one resource that their groups grant them, and
// their permissions on one resource.
func userAccessReads(a *app, users resource[iamUser]) []*cobra.Command {
	accessControls, permissions, resources := iamAccessControls().reader(), iamPermissions().reader(), iamResources()
	user := users.path + "/" + placeholderParent
	onResource := users.itemArgumentsSentence(resources.argument(), resources.listLine(),
		"an unknown one, or a resource the user has no access to, lists nothing")
	// Paging parameters do not page the answer there (live: pageNumber 2, and
	// pageSize 1 on page 1, list nothing), which comes whole without them.
	whole := listing{read: (*emporix.Client).GetUnpaged}
	onResourceIDs := argumentIDs{users.argument(): users.ids(), resources.argument(): resources.ids()}
	return []*cobra.Command{
		childCommand(a, users, child[string]{
			use:   "scopes",
			short: "List the scopes of a user",
			long: "Reads the scopes a user is granted through their groups, sorted, without the tenant\n" +
				"entry; a user without scopes lists nothing, like an unknown one.",
			scopes:  []string{"iam.scope_read"},
			fetch:   single{items: userScopeItems},
			unknown: listsNothing,
			view:    stringsView("scope", "SCOPE"),
		}),
		childCommand(a, users, child[iamAccessControl]{
			use:   "access-controls",
			short: "List the access controls of a user",
			long: "Reads the access controls a user is granted through their groups.\n" + accessControlColumns + "\n" +
				"Emporix lists them on one page only, so they are read as " + onePage(userAccessControlsPage) +
				": a user\nwith more fails.",
			scopes:  accessControls.scopes,
			fetch:   listing{read: onFirstPage(userAccessControlsPage)},
			unknown: listsNothing,
			view:    accessControls.view,
		}),
		accessControls.command(a, &command{
			use:   "resource-access-controls " + users.argument() + " " + resources.argument(),
			short: "List the access controls of a user on a resource",
			long: "Reads the access controls a user is granted on one resource.\n" + accessControlColumns + "\n" +
				onResource + "\n" + deprecatedNote,
			req:   request{path: user + "/access-controls/" + placeholderID},
			fetch: whole,
			ids:   onResourceIDs,
		}),
		permissions.command(a, &command{
			use:   "resource-permissions " + users.argument() + " " + resources.argument(),
			short: "List the permissions of a user on a resource",
			long: "Reads the permissions a user has on one resource through the roles that their groups'\n" +
				"access controls grant there: " + permissionColumns + "\n" + onResource + "\n" + deprecatedNote,
			req:   request{path: user + "/permissions/" + placeholderID},
			fetch: whole,
			ids:   onResourceIDs,
		}),
	}
}

// vendorUsers returns `for-vendor VENDOR_ID`, the users of a vendor. The
// reference documents no paging for them, unlike the tenant's users, and the
// tenant tested has no vendor to check it on, so they are read with GetAll,
// which fails rather than repeat or miss users wherever paging misbehaves.
func vendorUsers(a *app, users resource[iamUser]) *cobra.Command {
	sellers := vendors()
	return users.reader().command(a, &command{
		use:   "for-vendor " + sellers.argument(),
		short: "List the users of a vendor",
		long: "Reads the users of a vendor (`epx vendors list`), personal data included, with the columns\n" +
			"of `epx iam users list`.\n" +
			sellers.argumentSentence(listsNothing),
		req:   request{path: users.path + "/vendors/" + placeholderID},
		fetch: listing{},
		ids:   argumentIDs{sellers.argument(): sellers.ids()},
	})
}

// userScopeItems returns the scopes of a user's scopes response, a text of
// scopes separated by spaces, as items, sorted; the tenant entry is left out.
func userScopeItems(body json.RawMessage) ([]json.RawMessage, error) {
	var granted struct {
		Scopes *string `json:"scopes"`
	}
	if err := json.Unmarshal(body, &granted); err != nil || granted.Scopes == nil {
		return nil, errors.New("expected an object with scopes")
	}
	_, scopes := emporix.SplitScopes(*granted.Scopes)
	return marshalEach(scopes)
}

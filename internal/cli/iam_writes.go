package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// The scopes of the IAM writes, as the reference names them.
const (
	iamAccessManage     = "iam.access_manage"
	iamScopeManage      = "iam.scope_manage"
	iamAssignmentManage = "iam.assignment_manage"
)

// iamGroupManage is the scope of the writes of groups. The reference names it
// for deleting a group, but iam.group_read, which cannot change data, for
// creating and upserting one; Emporix documents iam.group_manage as the scope
// that manages groups, and the tenant knows it (live), so every group write
// requests it.
const iamGroupManage = "iam.group_manage"

// accessChange follows what an IAM write sends, in its help (accessChanging).
const accessChange = "\nIt changes who may do what in the tenant."

// writesChangeAccess ends the help of a group of IAM reads and writes, after
// what it reads and changes.
const writesChangeAccess = "\nIts writes change who may do what in the tenant."

// upsertTerms says, in the help of an IAM upsert, how metadata.version locks
// it (the IAM Service checks it where the body gives it), and what Emporix
// answers.
const upsertTerms = "With metadata.version as read, Emporix changes it only if that is still its version.\n" +
	"Emporix answers with the id of one it created, and with nothing when it replaced one."

// unknownUpdate says, in the help of users update, what an unknown ID gets: Emporix answers 204 and
// makes the ID a member of the groups of groupIds (live), so no refusal is sent by epx: the service
// decides, and a read first would be a second request that a concurrent delete can outdate.
const unknownUpdate = "An unknown ID is not refused: Emporix answers 204 No Content and adds that ID as a member of\n" +
	"each group of groupIds, so those groups take effect for it. Check the ID with `epx iam users get`\n" +
	"first; `epx iam users groups delete-all ID` removes it from every group it is in."

// iamUserTypes returns the only kinds of user that a group's users have, in
// capitals as the reference names them.
func iamUserTypes() []string { return []string{"CUSTOMER", "EMPLOYEE"} }

// accessChanging returns w, an IAM write, whose help then says that it
// changes who may do what (accessChange): every IAM write is declared
// through it.
func accessChanging(w *write) *write {
	w.long += accessChange
	return w
}

// newIAMAccessControlsCommand returns the group of access controls: their
// reads, and update ID (an upsert) and delete ID.
func newIAMAccessControlsCommand(a *app) *cobra.Command {
	controls := iamAccessControls()
	controls.short = "Read, upsert and delete access controls"
	controls.about = "Reads IAM access controls, the scopes they grant the users of the groups that hold them, and\n" +
		"upserts and deletes them." + writesChangeAccess
	scopes := []string{iamAccessManage}
	return newResourceCommand(a, controls,
		accessChanging(&write{
			use: controls.itemUse("update"), short: "Upsert an access control", method: http.MethodPut,
			long: "Replaces the access control ID with the JSON in --file, or creates it with that id where\n" +
				"none exists (an upsert): the scopes it grants (at least one), and optionally its name and\n" +
				"description, domains and restrictedTo (CUSTOMER or EMPLOYEE: the only kind of group that may\n" +
				"grant it, set when it is created only).\n" + upsertTerms,
			req: request{path: controls.itemPath()}, scopes: scopes, body: &payload{}, language: mapsOnly,
			ids: controls.itemIDs(),
		}).command(a),
		accessChanging(&write{
			use: controls.itemUse("delete"), short: "Delete an access control", method: http.MethodDelete,
			long: "Deletes the access control ID. Emporix deletes only those made for the tenant, not the\n" +
				"predefined ones (PREDEFINED in `epx iam access-controls list`).",
			req: request{path: controls.itemPath()}, scopes: scopes, body: noBody{}, ids: controls.itemIDs(),
		}).command(a))
}

// newIAMScopesCommand returns the group of scopes: their reads, and update ID
// (an upsert) and delete ID of the tenant's custom scopes. Unlike its other
// upserts, the reference names no Content-Language for a scope, whose
// description is a map of translations, so none is sent.
func newIAMScopesCommand(a *app) *cobra.Command {
	scopes := iamScopes()
	scopes.short = "Read, upsert and delete OAuth scopes"
	scopes.about = "Reads the OAuth scopes the tenant knows, and upserts and deletes its custom scopes." +
		writesChangeAccess
	manage := []string{iamScopeManage}
	return newResourceCommand(a, scopes,
		accessChanging(&write{
			use: scopes.itemUse("update"), short: "Upsert a custom scope", method: http.MethodPut,
			long: "Replaces the custom scope ID, e.g. acme.vehicle_manage, with the JSON in --file, or creates\n" +
				"it where none exists (an upsert): its domain and description, a map of translations such\n" +
				`as {"en": "Manage vehicles"}.` + "\n" + upsertTerms + "\nAccess controls " +
				"grant scopes (`epx iam access-controls update`).",
			req: request{path: scopes.itemPath()}, scopes: manage, body: &payload{}, ids: scopes.itemIDs(),
		}).command(a),
		accessChanging(&write{
			use: scopes.itemUse("delete"), short: "Delete a custom scope", method: http.MethodDelete,
			long: "Deletes the custom scope ID: it can no longer be used, though API keys made before may\n" +
				"still show it. Emporix deletes only custom scopes, not the predefined ones (PREDEFINED in\n" +
				"`epx iam scopes list`).",
			req: request{path: scopes.itemPath()}, scopes: manage, body: noBody{}, ids: scopes.itemIDs(),
		}).command(a))
}

// iamGroupWrites returns the writes of groups: create, update ID (an upsert)
// and delete ID, which with --force-delete sends forceDelete=true, also
// removing the group's users, and then needs the scope that manages them
// too, as the reference documents; without it, it sends no forceDelete,
// which Emporix takes as false.
func iamGroupWrites(a *app, groups resource[iamGroup]) []*cobra.Command {
	scopes := []string{iamGroupManage}
	return []*cobra.Command{
		accessChanging(&write{
			use: createUse, short: "Create a group", method: http.MethodPost,
			long: "Creates the group in --file, a JSON object with its name and the ids of the access controls\n" +
				"that its users get (accessControls, from `epx iam access-controls list`), and optionally\n" +
				"its id (Emporix makes one otherwise), description, userType (CUSTOMER, or EMPLOYEE, the\n" +
				"default), templates, b2b.legalEntityId, restrictions and mixins. Emporix answers with the\n" +
				"group's id; a group with the id exists already: a conflict (exit 5).",
			req: request{path: groups.path}, scopes: scopes, body: &payload{}, language: mapsOnly,
		}).command(a),
		accessChanging(&write{
			use: groups.itemUse("update"), short: "Upsert a group", method: http.MethodPut,
			long: "Replaces the group ID with the JSON in --file, as create takes it without id, or\n" +
				"creates it with that id where none exists (an upsert): give its name and every other\n" +
				"member to keep, as `epx iam groups get ID -o json` prints it.\n" + upsertTerms,
			req: request{path: groups.itemPath()}, scopes: scopes, body: &payload{}, language: mapsOnly,
			ids: groups.itemIDs(),
		}).command(a),
		accessChanging(&write{
			use: groups.itemUse("delete"), short: "Delete a group", method: http.MethodDelete,
			long: "Deletes the group ID. Emporix refuses to delete a group that users belong to (400 Bad\n" +
				"Request, exit 1; `epx iam groups users list ID` lists them) unless --force-delete is\n" +
				"given, which removes every user from it too.",
			req: request{path: groups.itemPath()}, scopes: scopes, body: noBody{}, ids: groups.itemIDs(),
			choice: queryChoice{param: forceDeleteParam, options: []choiceOption{{
				flag: "force-delete", value: queryTrue, scopes: []string{iamAssignmentManage},
				usage: "delete the group also when users belong to it, removing every user from it",
			}}},
			deletes: deletion{many: "the group and the membership of every user of it", with: "force-delete"},
		}).command(a),
	}
}

// groupUserWrites returns the writes of members, the users of a group of
// membersOf (groups nested in `iam groups users`): create GROUP_ID, update
// GROUP_ID USER_TYPE USER_ID (an upsert), delete GROUP_ID USER_ID and
// delete-all GROUP_ID.
func groupUserWrites(a *app, membersOf resource[iamGroup], members child[groupMember]) []*cobra.Command {
	scopes := []string{iamAssignmentManage}
	group, collection := membersOf.argument(), members.below(membersOf.path)
	const user = "USER_ID" // a member, as the USER of the group's members
	source := "\n" + membersOf.argumentSentence(notFound)
	removal := "\n" + membersOf.argumentSentence(answeredDeleted)
	removeOne := "\n" + membersOf.argumentSource() + "; an unknown GROUP_ID or USER_ID " + string(answeredDeleted) + "."
	ofGroup := argumentIDs{group: membersOf.ids()}
	gets := "\nThe user then gets the scopes of the group's access controls; a group of a vendor takes\n" +
		"only the vendor's employees."
	return []*cobra.Command{
		accessChanging(&write{
			use: createUse + " " + group, short: "Add a user to a group", method: http.MethodPost,
			long: "Adds the user that --user names, of the kind that --user-type names (EMPLOYEE when left\n" +
				"out), to the group " + group + ", or the one in --file, a JSON object with its userId and\n" +
				"userType. Emporix answers with the id of the membership; a user already in the group is\n" +
				"answered with a new id that is not the listed membership's, and no second membership is made\n" +
				"(read the members with `epx iam groups users list`)." + gets + source,
			req: request{path: collection}, scopes: scopes, ids: ofGroup,
			body: &payload{typed: objectBody{
				stringField("userId", "user", "the `USER_ID` of the user to add: the ID that `epx iam users list` "+
					"shows, or a customer's id").require(),
				choiceField("userType", "user-type", "the `TYPE` of user", iamUserTypes()...),
			}},
		}).command(a),
		accessChanging(&write{
			use:   childItemUse(membersOf, "update", []string{"USER_TYPE", user}),
			short: "Upsert the membership of a user in a group", method: http.MethodPut,
			long: "Adds the user USER_ID, of the kind USER_TYPE (" + alternatives(iamUserTypes()) + "), to the group " +
				group + "\nunless it belongs to it already (an upsert): Emporix answers with the id of the membership\n" +
				"it made, and with nothing when there was one, which it leaves as it is." + gets + source,
			req:    request{path: collection + "/" + placeholderInner + "/" + placeholderID},
			scopes: scopes, body: noBody{}, ids: ofGroup,
			choices: map[string][]string{"USER_TYPE": iamUserTypes()},
		}).command(a),
		accessChanging(&write{
			use:   childItemUse(membersOf, "delete", []string{user}),
			short: "Remove a user from a group", method: http.MethodDelete,
			long: "Removes the user USER_ID from the group " + group + ": the user no longer gets the scopes of\n" +
				"its access controls." + removeOne,
			req: request{path: collection + "/" + placeholderID}, scopes: scopes, body: noBody{},
			ids: argumentIDs{group: membersOf.ids(), user: childIDs(membersOf, members, "USER")},
		}).command(a),
		accessChanging(&write{
			use: "delete-all " + group, short: "Remove every user from a group", method: http.MethodDelete,
			long: "Removes every user from the group " + group + ", which stays." + removal,
			req:  request{path: collection}, scopes: scopes, body: noBody{}, ids: ofGroup,
			deletes: deletion{many: "the membership of every user of the group"},
		}).command(a),
	}
}

// userWrites returns the writes of users (employees): create, update ID and
// delete ID. The reference's forceDelete of a delete, which also deletes the
// user's Developer Portal account, is left out: only a Developer Portal
// admin, a person, may send it.
func userWrites(a *app, users resource[iamUser]) []*cobra.Command {
	const members = "firstName, lastName, preferredSite, preferredCurrency, preferredLanguage, department"
	return []*cobra.Command{
		accessChanging(&write{
			use: createUse, short: "Create a user", method: http.MethodPost,
			long: "Creates a Management Dashboard user (an employee) from the JSON in --file, which gives each\n" +
				"of " + members + ",\ncontactEmail and groupIds (the groups that the user belongs to, from " +
				"`epx iam groups list`).\nEmporix emails the user to confirm the address and set a password. " +
				"An e-mail that a user of the\ntenant has already is a conflict (exit 5): Emporix creates no user. " +
				"It answers with the user's id.",
			req: request{path: users.path}, scopes: []string{"iam.user_create"}, body: &payload{},
		}).command(a),
		accessChanging(&write{
			use: users.itemUse("update"), short: "Replace a user", method: http.MethodPut,
			long: "Replaces the user ID, an employee, with the JSON in --file, which gives each of\n" +
				members + ",\nbackofficeUserNumber and groupIds: Emporix removes the user from every group and " +
				"adds them\nto those of groupIds.\n" + unknownUpdate,
			req: request{path: users.itemPath()}, scopes: []string{"iam.user_update"}, body: &payload{},
			ids: users.itemIDs(),
		}).command(a),
		accessChanging(&write{
			use: users.itemUse("delete"), short: "Delete a user", method: http.MethodDelete,
			long: "Deletes the user ID from the tenant: Emporix removes them from every group and blocks\n" +
				"their access to the Management Dashboard; their Developer Portal account stays.",
			req: request{path: users.itemPath()}, scopes: []string{"iam.user_delete"}, body: noBody{},
			ids: users.itemIDs(),
		}).command(a),
	}
}

// userGroupsWrite returns `delete-all USER_ID` of userGroups, the groups of a
// user of groupsOf (users nested in `iam users groups`), which removes the
// user from every group.
func userGroupsWrite(a *app, groupsOf resource[iamUser], userGroups child[iamGroup]) *cobra.Command {
	user := groupsOf.argument()
	return accessChanging(&write{
		use: "delete-all " + user, short: "Remove a user from every group", method: http.MethodDelete,
		long: "Removes the user " + user + " from every group: the user keeps their account, but no\n" +
			"group grants them access any more.\n" + groupsOf.argumentSentence(answeredDeleted),
		req:    request{path: userGroups.below(groupsOf.path)},
		scopes: []string{iamAssignmentManage}, body: noBody{}, ids: argumentIDs{user: groupsOf.ids()},
		deletes: deletion{many: "the user's membership of every group"},
	}).command(a)
}

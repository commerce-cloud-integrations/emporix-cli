package cli

import "github.com/spf13/cobra"

// iamUse is the name of the group of IAM reads.
const iamUse = "iam"

// newIAMCommand groups the reads and writes of the IAM Service: users,
// groups and what they may do.
func newIAMCommand(a *app) *cobra.Command {
	return newGroup(iamUse, "Read and change IAM users, groups and access rights",
		"Reads the tenant's IAM (identity and access management) data: user accounts, the groups they\n"+
			"belong to and what these may do: roles, permissions, scopes, resources, access controls and\n"+
			"templates. Its writes change users, groups and their users, access controls and custom\n"+
			"scopes, and with them who may do what in the tenant.",
		newIAMUsersCommand(a), newIAMGroupsCommand(a), newResourceCommand(a, iamRoles()),
		newResourceCommand(a, iamPermissions()), newIAMScopesCommand(a), newResourceCommand(a, iamResources()),
		newIAMAccessControlsCommand(a), newIAMTemplatesCommand(a))
}

package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// passwordMigration is the tenant's password migration retention: until when
// customers imported with a legacy password hash are migrated as they sign in.
type passwordMigration struct {
	RetentionEndDate          string `json:"retentionEndDate"`
	EmailReminderDate         string `json:"emailReminderDate"`
	EmailNotificationsEnabled *bool  `json:"emailNotificationsEnabled"`
}

// passwordMigrationPath is the path of the password migration retention.
const passwordMigrationPath = "/customer/{tenant}/config/password-migration-retention"

// newPasswordMigrationCommand returns the group of the password migration
// retention, which the Customer Service keeps for customer imports: get
// reads it, update sets it and delete removes it.
func newPasswordMigrationCommand(a *app) *cobra.Command {
	retention := request{path: passwordMigrationPath}
	scopes := []string{customerImportScope}
	get := reader[passwordMigration]{view: view[passwordMigration]{
		singular: "password migration retention",
		header:   []string{"ENDS", "REMINDER", "EMAILS"},
		row: func(p passwordMigration, _ language) []string {
			return []string{p.RetentionEndDate, p.EmailReminderDate, formatBool(p.EmailNotificationsEnabled)}
		},
	}, scopes: []string{"customer.import_read"}}.command(a, &command{
		use:   getUse,
		short: "Show until when imported customers' passwords are migrated",
		long: "Reads the password migration retention of customer imports: a customer imported with a\n" +
			"password hash from another system gets an Emporix one on signing in until ENDS; after\n" +
			"that, the customers not yet migrated must reset their password. REMINDER is when Emporix\n" +
			"asks them by email to sign in once, and EMAILS whether customers get these reminder and\n" +
			"reset emails. A tenant without this setting is not found (exit 4).",
		req:   retention,
		fetch: single{items: oneItem},
	})
	update := &write{
		use: "update", short: "Set until when imported customers' passwords are migrated", method: http.MethodPost,
		long: "Sets the whole password migration retention, or creates it where the tenant has none\n" +
			"(an upsert), from its flags or the JSON object of --file: until when customers imported\n" +
			"with legacyAuth get an Emporix password on signing in (--ends), when Emporix asks those\n" +
			"not yet migrated by email to sign in once (--reminder; without it, 7 days before --ends,\n" +
			"or tomorrow if that has passed, also where the retention it replaces had another date)\n" +
			"and whether customers get these reminder and reset emails (--emails: without it, epx\n" +
			"sends emailNotificationsEnabled=false; a --file without it gets Emporix's default, true).\n" +
			"Emporix answers with the retention.",
		example: "  epx customers password-migration update --ends 2027-01-31 --reminder 2027-01-24 --emails\n" +
			"  epx customers password-migration update --ends 2027-01-31 --yes",
		req: retention, scopes: scopes,
		body: &payload{typed: objectBody{
			stringField("retentionEndDate", "ends", "the `DATE` the migration ends, e.g. 2027-01-31").require(),
			stringField("emailReminderDate", "reminder", "the `DATE` of the reminder email, before --ends"),
			boolField("emailNotificationsEnabled", "emails", "email customers the reminder and the password reset"),
		}},
	}
	remove := &write{
		use: "delete", short: "Remove the password migration retention", method: http.MethodDelete,
		long: "Removes the password migration retention. Emporix refuses it (400 Bad Request, exit 1)\n" +
			"while a customer still has a legacy password hash to migrate: move --ends of update\n" +
			"earlier instead to end the migration sooner. update sets the whole retention, so give\n" +
			"--reminder and --emails again where it should keep them.",
		req: retention, scopes: scopes, body: noBody{},
		deletes: deletion{many: "the password migration retention of the tenant"},
	}
	return newGroup("password-migration", "Read and set until when imported customers' passwords are migrated",
		"Reads, sets and removes the password migration retention of customer imports, until when\n"+
			"customers imported with a password hash from another system get an Emporix one.",
		get, update.command(a), remove.command(a))
}

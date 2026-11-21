package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

const customerManageScope = "customer.customer_manage"

const customerImportScope = "customer.import_manage"

const maxCustomerImport = 200

func passwordResetEmail() paramFlag {
	return paramFlag{param: "sendPasswordResetNotifications", flag: "password-reset-email",
		usage: "email the customer a link to set a password once the account is created"}
}

const contactEmailNote = "On an existing profile, a contactEmail changes its contact email only, never the email " +
	"with which the customer signs in (that of its account)."

const passwordResetNote = "Emporix emails a new customer a link to set a password only with --password-reset-email " +
	"(without it, epx sends sendPasswordResetNotifications=false, where Emporix's default is true)."

const customerNumberNote = "Emporix ignores a customerNumber that a profile gives (no error)."

func customerWrites(a *app, customers resource[customer]) []*cobra.Command {
	scopes := []string{customerManageScope}
	item, ids := request{path: customers.itemPath()}, customers.itemIDs()
	number := customers.idSource() + "."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a customer", method: http.MethodPost,
			long: "Creates the customer profile in --file, a JSON object with at least contactEmail, such as " +
				"`epx customers get NUMBER -o json` prints one; Emporix makes the customer number unless " +
				"the body gives it as id, and answers with it. An account with that email that exists " +
				"already is a conflict (exit 5). " + customerNumberNote + " " + passwordResetNote,
			req: request{path: customers.path}, scopes: scopes, body: &payload{},
			switches: []paramFlag{passwordResetEmail()},
		}).command(a),
		(&write{
			use: customers.itemUse("update"), short: "Upsert a customer", method: http.MethodPut,
			long: "Changes the members of the customer profile NUMBER that --file gives, a JSON object as " +
				"`epx customers get NUMBER -o json` prints it, keeping those it leaves out, or creates the " +
				"profile with that number where none exists (an upsert), as create does: with the email of " +
				"an existing account, that is a conflict (exit 5). " + number + " " + customerNumberNote + " " +
				ifUnchangedNote + " " + contactEmailNote + " " + passwordResetNote,
			req: item, scopes: scopes, body: &payload{}, switches: []paramFlag{passwordResetEmail()}, ids: ids,
		}).command(a),
		(&write{
			use: customers.itemUse("patch"), short: "Change a customer", method: http.MethodPatch,
			long: "Changes the members of the customer profile NUMBER that --file gives, a JSON object such " +
				"as " + literal(`{"active": false}`) + "; Emporix keeps the others. " + number + " " + customerNumberNote + " " +
				ifUnchangedNote + " " + contactEmailNote,
			req: item, scopes: scopes, body: &payload{}, ids: ids,
		}).command(a),
		(&write{
			use: customers.itemUse("delete"), short: "Delete a customer", method: http.MethodDelete,
			long: "Deletes the customer profile NUMBER and its accounts. Emporix answers 202 Accepted, and " +
				"the profile is not found from then on. " + number,
			req: item, scopes: scopes, body: noBody{}, ids: ids,
		}).command(a),
		(&write{
			use: importUse, short: "Import customers with their passwords", method: http.MethodPost,
			long: "Imports the customer profiles in --file, a JSON array of profiles as create takes them, " +
				"each with an account that holds its email and exactly one of passwordHash, a hash that " +
				"Emporix checks as its own, or legacyAuth, the hash of another system and its algorithm, " +
				"which Emporix replaces with its own as the customer signs in. Customers with legacyAuth " +
				"need a password migration retention: `epx customers password-migration update` sets it. " +
				customerNumberNote,
			req: request{path: customers.path + "/import"}, scopes: []string{customerImportScope},
			body: &payload{maxItems: maxCustomerImport},
		}).command(a),
	}
}

// The reference names no scope for adding tags; Emporix requires the scope with which it removes them,
// customer_manage, as for the other writes.
func addressWrites(a *app, customers resource[customer], addresses child[customerAddress]) []*cobra.Command {
	scopes := []string{customerManageScope}
	arguments := customers.argumentSource() + ", and ID the ID that `epx " + childListLine(customers, addresses) +
		"` shows."
	address := func(verb string) childTarget { return childItemTarget(customers, addresses, verb, nil) }
	update, patch, remove, addTags, removeTags := address("update"), address("patch"), address("delete"),
		address("add-tags"), address("remove-tags")
	ofCustomer := argumentIDs{customers.argument(): customers.ids()}
	ofAddress := childItemIDs(customers, addresses, &update)
	tags := func(verb string) queryFlags {
		return queryFlags{required: []paramFlag{{param: "tags", flag: "tags",
			usage: "the `TAGS` to " + verb + ", separated by commas, e.g. BILLING,SHIPPING"}}}
	}
	return []*cobra.Command{
		(&write{
			use: createUse + " " + customers.argument(), short: "Add an address to a customer", method: http.MethodPost,
			long: "Adds the address in --file, a JSON object with at least contactName, to the customer " +
				"CUSTOMER_NUMBER; Emporix makes its id unless the body gives one, and answers with it. " +
				customers.argumentSource() + ".",
			req: request{path: addresses.below(customers.path)}, scopes: scopes, body: &payload{}, ids: ofCustomer,
		}).command(a),
		(&write{
			use: update.use, short: "Upsert an address of a customer", method: http.MethodPut,
			long: "Changes the members of the address ID of the customer CUSTOMER_NUMBER that --file gives, " +
				"a JSON object as `epx customers addresses get CUSTOMER_NUMBER ID -o json` prints it, " +
				"keeping those it leaves out, such as its tags, or creates the address with that id where " +
				"none exists (an upsert). " + arguments,
			req: request{path: update.path}, scopes: scopes, body: &payload{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: patch.use, short: "Change an address of a customer", method: http.MethodPatch,
			long: "Changes the members of the address ID of the customer CUSTOMER_NUMBER that --file gives, " +
				"a JSON object such as " + literal(`{"isDefault": true}`) + "; Emporix keeps the others. " + arguments,
			req: request{path: patch.path}, scopes: scopes, body: &payload{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: remove.use, short: "Delete an address of a customer", method: http.MethodDelete,
			long: "Deletes the address ID of the customer CUSTOMER_NUMBER. " + arguments,
			req:  request{path: remove.path}, scopes: scopes, body: noBody{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: addTags.use, short: "Tag an address of a customer", method: http.MethodPost,
			long: "Adds the tags that --tags names, such as BILLING or SHIPPING, to the address ID of the " +
				"customer CUSTOMER_NUMBER. " + arguments,
			req: request{path: addTags.path + "/tags"}, scopes: scopes, body: noBody{}, params: tags("add"),
			ids: ofAddress,
		}).command(a),
		(&write{
			use: removeTags.use, short: "Remove tags from an address of a customer", method: http.MethodDelete,
			long: "Removes the tags that --tags names from the address ID of the customer CUSTOMER_NUMBER. " +
				arguments,
			req: request{path: removeTags.path + "/tags"}, scopes: scopes, body: noBody{}, params: tags("remove"),
			ids: ofAddress,
		}).command(a),
	}
}

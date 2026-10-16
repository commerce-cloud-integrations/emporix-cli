package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// customerManageScope is the scope of the Customer Service's writes of
// customer profiles and their addresses (customer_manage_own is that of a
// customer's own profile).
const customerManageScope = "customer.customer_manage"

// customerImportScope is the scope of customer imports and of the password
// migration retention that they need.
const customerImportScope = "customer.import_manage"

// maxCustomerImport is the most customers that Emporix imports in one request.
const maxCustomerImport = 200

// passwordResetEmail is the switch of a write that may create a customer
// account: Emporix emails the customer a link to set a password unless told
// not to, which epx tells it unless the flag is given.
func passwordResetEmail() paramFlag {
	return paramFlag{param: "sendPasswordResetNotifications", flag: "password-reset-email",
		usage: "email the customer a link to set a password once the account is created"}
}

// contactEmailNote follows what a write that changes a customer profile
// sends, in its help: the account's email is not its contactEmail.
const contactEmailNote = "On an existing profile, a contactEmail changes its contact email only, never the email\n" +
	"with which the customer signs in (that of its account)."

// passwordResetNote follows what a write that may create a customer account
// sends, in its help.
const passwordResetNote = "Emporix emails a new customer a link to set a password only with --password-reset-email\n" +
	"(without it, epx sends sendPasswordResetNotifications=false, where Emporix's default is true)."

// customerNumberNote states, in the help of a write that sends customer
// profiles, that Emporix ignores a customerNumber in one: the customer number
// is the id that Emporix makes or the profile or path gives.
const customerNumberNote = "Emporix ignores a customerNumber that a profile gives (no error)."

// customerWrites returns the writes of customer profiles: create, update (an
// upsert), patch, delete and import, which creates customers with their
// passwords.
func customerWrites(a *app, customers resource[customer]) []*cobra.Command {
	scopes := []string{customerManageScope}
	item, ids := request{path: customers.itemPath()}, customers.itemIDs()
	number := customers.idSource() + "."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a customer", method: http.MethodPost,
			long: "Creates the customer profile in --file, a JSON object with at least contactEmail, such as\n" +
				"`epx customers get NUMBER -o json` prints one; Emporix makes the customer number unless\n" +
				"the body gives it as id, and answers with it. An account with that email that exists\n" +
				"already is a conflict (exit 5).\n" + customerNumberNote + "\n" + passwordResetNote,
			req: request{path: customers.path}, scopes: scopes, body: &payload{},
			switches: []paramFlag{passwordResetEmail()},
		}).command(a),
		(&write{
			use: customers.itemUse("update"), short: "Upsert a customer", method: http.MethodPut,
			long: "Changes the members of the customer profile NUMBER that --file gives, a JSON object as\n" +
				"`epx customers get NUMBER -o json` prints it, keeping those it leaves out, or creates the\n" +
				"profile with that number where none exists (an upsert), as create does: with the email of\n" +
				"an existing account, that is a conflict (exit 5).\n" + number + "\n" + customerNumberNote + "\n" +
				ifUnchangedNote + "\n" + contactEmailNote + "\n" + passwordResetNote,
			req: item, scopes: scopes, body: &payload{}, switches: []paramFlag{passwordResetEmail()}, ids: ids,
		}).command(a),
		(&write{
			use: customers.itemUse("patch"), short: "Change a customer", method: http.MethodPatch,
			long: "Changes the members of the customer profile NUMBER that --file gives, a JSON object such\n" +
				"as {\"active\": false}; Emporix keeps the others.\n" + number + "\n" + customerNumberNote + "\n" +
				ifUnchangedNote + "\n" + contactEmailNote,
			req: item, scopes: scopes, body: &payload{}, ids: ids,
		}).command(a),
		(&write{
			use: customers.itemUse("delete"), short: "Delete a customer", method: http.MethodDelete,
			long: "Deletes the customer profile NUMBER and its accounts. Emporix answers 202 Accepted, and\n" +
				"the profile is not found from then on.\n" + number,
			req: item, scopes: scopes, body: noBody{}, ids: ids,
		}).command(a),
		(&write{
			use: importUse, short: "Import customers with their passwords", method: http.MethodPost,
			long: "Imports the customer profiles in --file, a JSON array of profiles as create takes them,\n" +
				"each with an account that holds its email and exactly one of passwordHash, a hash that\n" +
				"Emporix checks as its own, or legacyAuth, the hash of another system and its algorithm,\n" +
				"which Emporix replaces with its own as the customer signs in. Customers with legacyAuth\n" +
				"need a password migration retention: `epx customers password-migration update` sets it.\n" +
				customerNumberNote,
			req: request{path: customers.path + "/import"}, scopes: []string{customerImportScope},
			body: &payload{maxItems: maxCustomerImport},
		}).command(a),
	}
}

// addressWrites returns the writes of the addresses of a customer of
// customers, in the collection addresses: create, update (an upsert), patch
// and delete, and add-tags and remove-tags, which change the tags of one,
// each addressing its address as a read of it does (childItemTarget). The
// reference names no scope for adding tags; Emporix requires the scope with
// which it removes them, customer_manage, as for the other writes.
func addressWrites(a *app, customers resource[customer], addresses child[customerAddress]) []*cobra.Command {
	scopes := []string{customerManageScope}
	arguments := customers.argumentSource() + ", and ID the ID that\n`epx " + childListLine(customers, addresses) +
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
			long: "Adds the address in --file, a JSON object with at least contactName, to the customer\n" +
				"CUSTOMER_NUMBER; Emporix makes its id unless the body gives one, and answers with it.\n" +
				customers.argumentSource() + ".",
			req: request{path: addresses.below(customers.path)}, scopes: scopes, body: &payload{}, ids: ofCustomer,
		}).command(a),
		(&write{
			use: update.use, short: "Upsert an address of a customer", method: http.MethodPut,
			long: "Changes the members of the address ID of the customer CUSTOMER_NUMBER that --file gives,\n" +
				"a JSON object as `epx customers addresses get CUSTOMER_NUMBER ID -o json` prints it,\n" +
				"keeping those it leaves out, such as its tags, or creates the address with that id where\n" +
				"none exists (an upsert).\n" + arguments,
			req: request{path: update.path}, scopes: scopes, body: &payload{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: patch.use, short: "Change an address of a customer", method: http.MethodPatch,
			long: "Changes the members of the address ID of the customer CUSTOMER_NUMBER that --file gives,\n" +
				"a JSON object such as {\"isDefault\": true}; Emporix keeps the others.\n" + arguments,
			req: request{path: patch.path}, scopes: scopes, body: &payload{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: remove.use, short: "Delete an address of a customer", method: http.MethodDelete,
			long: "Deletes the address ID of the customer CUSTOMER_NUMBER.\n" + arguments,
			req:  request{path: remove.path}, scopes: scopes, body: noBody{}, ids: ofAddress,
		}).command(a),
		(&write{
			use: addTags.use, short: "Tag an address of a customer", method: http.MethodPost,
			long: "Adds the tags that --tags names, such as BILLING or SHIPPING, to the address ID of the\n" +
				"customer CUSTOMER_NUMBER.\n" + arguments,
			req: request{path: addTags.path + "/tags"}, scopes: scopes, body: noBody{}, params: tags("add"),
			ids: ofAddress,
		}).command(a),
		(&write{
			use: removeTags.use, short: "Remove tags from an address of a customer", method: http.MethodDelete,
			long: "Removes the tags that --tags names from the address ID of the customer CUSTOMER_NUMBER.\n" +
				arguments,
			req: request{path: removeTags.path + "/tags"}, scopes: scopes, body: noBody{}, params: tags("remove"),
			ids: ofAddress,
		}).command(a),
	}
}

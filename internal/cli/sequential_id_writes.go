package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// issuesIDs says, in the help of a write that issues IDs, what it changes,
// which Emporix allows with the read scope (write.readScoped).
const issuesIDs = "Each ID issued advances the schema's counter."

// issuedForSite is the optional --site of a write that issues IDs.
func issuedForSite() queryFlags {
	return queryFlags{optional: []paramFlag{
		optionalSiteParam("the site `CODE` whose data fills in the placeholders that the body leaves out"),
	}}
}

// sequenceSchemaWrites returns the writes of sequence schemas: create, delete
// ID and activate ID, and the writes that issue IDs, next-id TYPE and
// next-ids.
func sequenceSchemaWrites(a *app, schemas resource[sequenceSchema]) []*cobra.Command {
	manage := []string{sequenceManageScope}
	id := "ID is the ID that `epx " + schemas.listLine() + "` shows."
	writes := []*write{
		{
			use: createUse, short: "Create a sequential ID schema", method: http.MethodPost,
			long: "Creates the sequence schema in --file, a JSON object with its name, schemaType (such as " +
				"orderNoSequence or invoiceNoSequence; letters and digits only), startValue, maxValue and " +
				"numberOfDigits, and optionally preText and postText, the text around the number, and " +
				"placeholders: each placeholder in that text, named with __ at both ends such as __year__, " +
				"with required and default, or sitePath, where in the data of a site its value is. " +
				"`epx sequential-ids activate ID` makes the schema the one that numbers its type.",
			example: "  epx sequential-ids create --file order-numbers.json --dry-run",
			req:     request{path: schemas.path}, scopes: manage, body: &payload{},
			answer: textAnswer{what: "the new schema's ID as bare text"},
		},
		{
			use: schemas.itemUse("activate"), short: "Activate a sequential ID schema", method: http.MethodPost,
			long: "Makes the sequence schema ID the active one, which numbers the records of its type. " + id,
			req:  request{path: schemas.itemPath() + "/setActive"}, scopes: manage, body: noBody{},
			ids: schemas.itemIDs(),
		},
		{
			use: schemas.itemUse("delete"), short: "Delete a sequential ID schema", method: http.MethodDelete,
			long: "Deletes the sequence schema ID. " + id,
			req:  request{path: schemas.itemPath()}, scopes: manage, body: noBody{}, ids: schemas.itemIDs(),
		},
		{
			use: "next-id TYPE", short: "Issue the next ID of a sequential ID schema type", method: http.MethodPost,
			long: "Issues the next ID of the active schema of TYPE, such as orderNoSequence (`epx " +
				"sequential-ids list` shows the types), and prints it. --sequence-key numbers in a sequence " +
				"of its own, such as one per month, instead of the default one; --file gives the body " +
				"whole, with the values of placeholders, e.g. " + literal(`{"sequenceKey": "2026-10", `+
				`"placeholders": {"__year__": "2026"}}`) + ". With --site, Emporix fills in the placeholders " +
				"that the body leaves out from that site's data. " + issuesIDs,
			example: "  epx sequential-ids next-id orderNoSequence --sequence-key 2026-10 --dry-run",
			req:     request{path: schemas.path + "/types/" + placeholderID + "/nextId"},
			ids:     argumentIDs{columnType: schemas.idsIn(columnType)},
			scopes:  schemas.scopes, readScoped: true, params: issuedForSite(),
			body: &payload{typed: objectBody{stringField("sequenceKey", "sequence-key",
				"the `KEY` of the sequence to number in, such as 2026-10")}},
		},
		{
			use: "next-ids", short: "Issue IDs of several sequential ID schema types", method: http.MethodPost,
			long: "Issues IDs of the active schemas of several types in one request and prints them, by " +
				"type. --file gives a JSON object with a member per type, such as orderNoSequence, each " +
				"with numberOfIds (how many), and optionally sequenceKey and placeholders, as next-id " +
				"takes them. With --site, Emporix fills in the placeholders that the body leaves out from " +
				"that site's data. " + issuesIDs,
			example: "  epx sequential-ids next-ids --file ids.json --site AcmeStore_NL --dry-run",
			req:     request{path: "/sequential-id/sequenceSchemaBatch/nextIds"},
			scopes:  schemas.scopes, readScoped: true, params: issuedForSite(), body: &payload{},
		},
	}
	commands := make([]*cobra.Command, len(writes))
	for i, w := range writes {
		commands[i] = w.command(a)
	}
	return commands
}

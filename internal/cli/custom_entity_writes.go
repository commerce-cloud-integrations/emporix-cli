package cli

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

// The writes of the Schema Service's custom entities: the types, with the
// service's manage scope, which its reference names for each, and their
// instances (customInstanceWrites).

// customEntityWrites returns the writes of custom entity types, entities:
// create, update ID (an upsert), delete ID and import.
func customEntityWrites(a *app, entities resource[customEntity]) []*cobra.Command {
	manage := []string{schemaManage}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a custom entity type", method: http.MethodPost,
			long: "Creates the custom entity type in --file, a JSON object with its id and name, such as\n" +
				`{"id": "CUSTOM_DOCUMENT", "name": {"en": "Document"}}, and prints the id that Emporix answers` +
				"\nwith; one that exists already is a conflict. The id has capital letters, digits and\n" +
				"underscores, the first a letter. Emporix also creates the scopes that read and manage the\n" +
				"instances of the type, such as custom.custom_document_manage, which it grants only minutes\n" +
				"later: until then, the instance commands of the type exit with code 3.\n" +
				"Built-in types, such as AVAILABILITY and LOCATION, cannot be created.",
			req: request{path: entities.path}, scopes: manage, body: &payload{}, language: mapsOnly,
		}).command(a),
		(&write{
			use: entities.itemUse("update"), short: "Upsert a custom entity type", method: http.MethodPut,
			long: "Replaces the name of the custom entity type ID with the one in --file, a JSON object such\n" +
				`as {"name": {"en": "Document"}, "metadata": {"version": 1}}, or creates the type where none` +
				"\nexists (an upsert). With metadata.version, as `epx custom-entities get ID -o json` prints\n" +
				"it, Emporix replaces only that version: an older one is a conflict.",
			req: request{path: entities.itemPath()}, scopes: manage, body: &payload{}, language: mapsOnly,
			ids: entities.itemIDs(),
		}).command(a),
		(&write{
			use: entities.itemUse("delete"), short: "Delete a custom entity type", method: http.MethodDelete,
			long: "Deletes the custom entity type ID. Emporix refuses to delete a type that schemas or\n" +
				"instances use (400 Bad Request), and answers an unknown ID as deleted, not as not found.\n" +
				"The scopes that Emporix created with the type remain granted: the instances of a deleted\n" +
				"type then read as an empty list.",
			req: request{path: entities.itemPath()}, scopes: manage, body: noBody{}, ids: entities.itemIDs(),
		}).command(a),
		(&write{
			use: importUse, short: "Import custom entity types with their schemas", method: http.MethodPost,
			long: "Imports the custom entity types and their schemas of the export package in --file, as\n" +
				"`epx custom-entities export -o json` prints it, e.g. on another tenant. A type or schema\n" +
				"that exists already is created again, its id followed by _COPY, such as\n" +
				"CUSTOM_DOCUMENT_COPY, and the references of its schemas then name the copy.",
			req: request{path: entities.path + "/import"}, scopes: manage, body: &payload{},
		}).command(a),
	}
}

// maxBulkInstances is the most custom instances that Emporix takes in one
// bulk write, as its reference documents for bulk-create, bulk-update and
// bulk-patch.
const maxBulkInstances = 200

// maxBulkDeleteInstances is the most ids of custom instances that Emporix
// takes in one bulk-delete, for which its reference gives no limit: it
// refuses more (400 Bad Request, "Maximum size of the payload array is
// 1000.").
const maxBulkDeleteInstances = 1000

// customInstanceWrites returns the writes of instances, the instances of a
// custom entity type of typesOf (the types nested as the parent of the
// instances group): create, update ID (an upsert), patch ID and delete ID one
// (instanceWrites.one), and each in bulk (instanceWrites.bulk), with the
// manage scope of the type, and its read scope too where Emporix checks the
// references of the instances (declareReferenceCheck).
func customInstanceWrites(a *app, typesOf resource[customEntity], instances child[customInstance]) []*cobra.Command {
	typeSource, get := "\n"+typesOf.argumentSource(), childItemTarget(typesOf, instances, getUse, nil)
	w := &instanceWrites{
		path: instances.below(typesOf.path), typeArg: typesOf.argument(), id: instances.header[0],
		manage: []string{customTypeManageScope}, ofType: typeSource + ".",
		ofItem: typeSource + ", and " + instances.header[0] + " the one that\n`epx " +
			childListLine(typesOf, instances) + "` shows.",
		typeIDs: argumentIDs{typesOf.argument(): typesOf.ids()},
		itemIDs: childItemIDs(typesOf, instances, &get),
	}
	return writeCommands(a, slices.Concat(w.one(), w.bulk()))
}

// instanceWrites declares the writes of the instances of a custom entity
// type, each with the manage scope of the type and help that ends saying
// where its arguments come from, followed, where Emporix may check the
// references of the instances, by what that check needs (declareReferenceCheck).
type instanceWrites struct {
	path    string      // the collection of the instances of the type {parent}
	typeArg string      // the argument that names the type, {parent}
	id      string      // the argument that names one instance, {id}
	manage  []string    // the scopes of each, besides those of checking references
	ofType  string      // ends the help of a write that takes typeArg alone
	ofItem  string      // ends the help of a write that takes typeArg and id
	typeIDs argumentIDs // what typeArg completes with
	itemIDs argumentIDs // what typeArg and id complete with
}

// one returns the writes of one instance: create, update ID (an upsert) and
// patch ID, whose references Emporix checks unless told otherwise, and
// delete ID.
func (in *instanceWrites) one() []write {
	item := request{path: in.path + "/" + placeholderID}
	itemUse := func(verb string) string { return verb + " " + in.typeArg + " " + in.id }
	writes := []write{
		{
			use: createUse + " " + in.typeArg, short: "Create an instance of a custom entity type", method: http.MethodPost,
			long: "Creates an instance of the custom entity type " + in.typeArg + " from the JSON object in\n" +
				`--file, such as {"id": "123", "name": {"en": "Report"}, "mixins": {…}}, and prints the id` + "\n" +
				"that Emporix answers with; it makes one unless the object names it, and an id that exists\n" +
				`already is a conflict. An owner, such as {"type": "EMPLOYEE", "userId": "…"}, cannot change` +
				"\nonce set. The attributes of a schema go under its id, which metadata.mixins maps to its\n" +
				`URL: {"mixins": {"SCHEMA_ID": {…}}, "metadata": {"mixins": {"SCHEMA_ID": "URL"}}}, the URL` +
				"\nbeing the metadata.url that `epx schemas get SCHEMA_ID -o json` prints. A localized\n" +
				`attribute, unlike the name, is a list of translations, such as` + "\n" +
				`[{"language": "en", "value": "Report"}].` + in.ofType,
			req: request{path: in.path}, scopes: in.manage, body: &payload{}, language: mapsOnly, ids: in.typeIDs,
		},
		{
			use: itemUse("update"), short: "Upsert an instance of a custom entity type", method: http.MethodPut,
			long: "Replaces the instance ID of the custom entity type " + in.typeArg + " with the JSON object in\n" +
				"--file, as create takes it without the id, or creates it with that id where none exists\n" +
				"(an upsert). With metadata.version, as `epx custom-entities instances get " + in.typeArg + " ID\n" +
				"-o json` prints it, Emporix replaces only that version: an older one is a conflict." + in.ofItem,
			req: item, scopes: in.manage, body: &payload{}, language: mapsOnly,
			reserved: map[string]string{bulkSegment: bulkUpdateUse}, ids: in.itemIDs,
		},
		{
			use: itemUse("patch"), short: "Change an instance of a custom entity type", method: http.MethodPatch,
			long: "Changes the instance ID of the custom entity type " + in.typeArg + " with the JSON array of\n" +
				"JSON Patch (RFC 6902) operations in --file, such as\n" +
				`[{"op": "replace", "path": "/name", "value": {"en": "Report"}}]; an add whose path ends in` +
				"\n/- appends to an array. An unknown ID is not found." + in.ofItem,
			req: item, scopes: in.manage, body: operationsFile(), language: mapsOnly,
			reserved: map[string]string{bulkSegment: bulkPatchUse}, ids: in.itemIDs,
		},
	}
	for i := range writes {
		in.declareReferenceCheck(&writes[i], true)
	}
	return append(writes, write{
		use: itemUse("delete"), short: "Delete an instance of a custom entity type", method: http.MethodDelete,
		long: "Deletes the instance ID of the custom entity type " + in.typeArg + "; Emporix answers an\n" +
			"unknown ID as deleted, not as not found." + in.ofItem,
		req: item, scopes: in.manage, body: noBody{}, reserved: map[string]string{bulkSegment: bulkDeleteUse},
		ids: in.itemIDs,
	})
}

// bulk returns the bulk writes of instances (bulkWrites): bulk-create,
// bulk-update (an upsert), bulk-patch and bulk-delete. Their defaults of
// validateReferences differ, and all but bulk-delete send their maps of
// translations with Content-Language: *: the reference lists the header for
// bulk-patch only, but Emporix refuses the maps of bulk-create and
// bulk-update items without it. bulk-delete takes more items than the others
// (maxBulkDeleteInstances).
func (in *instanceWrites) bulk() []write {
	writes := (&bulkWrites{
		path: in.path, args: in.typeArg, nouns: "instances of a custom entity type", scopes: in.manage,
		maxItems: maxBulkInstances, language: mapsOnly,
		create: "Creates instances of the custom entity type " + in.typeArg + " from --file, a JSON array of\n" +
			"instances, each as create takes it." + in.ofType,
		update: "Replaces the instances of the custom entity type " + in.typeArg + " in --file, a JSON array of\n" +
			"instances, each with its id and as update takes it, or creates those that do not exist\n" +
			"(an upsert)." + in.ofType,
		patch: "Changes instances of the custom entity type " + in.typeArg + " as --file says: a JSON array\n" +
			"with the id of each instance and the JSON Patch operations that change it, as patch takes\n" +
			`them, such as` + "\n" + `[{"id": "123", "data": [{"op": "replace", "path": "/name/en", "value": "Report"}]}].` +
			in.ofType,
		delete: "Deletes the instances of the custom entity type " + in.typeArg + " whose ids --file lists, a\n" +
			`JSON array of strings, such as ["123", "124"]; Emporix answers an unknown id as deleted.` + in.ofType,
	}).writes()
	create, update, patch, remove := &writes[0], &writes[1], &writes[2], &writes[3]
	for _, w := range []*write{create, update, patch} {
		in.declareReferenceCheck(w, false)
	}
	patch.paramScopes.flag.usage = referencesUsage + " (false without it, although its reference says true)"
	patch.language = mapsOnly
	remove.body = &payload{maxItems: maxBulkDeleteInstances}
	for i := range writes {
		writes[i].ids = in.typeIDs
	}
	return writes
}

// validateReferencesParam is the query parameter of a write of custom
// instances that says whether Emporix checks that what their references name
// exists.
const validateReferencesParam = "validateReferences"

// referencesUsage is what the flag that sets validateReferencesParam does;
// what Emporix does without it follows.
const referencesUsage = "whether Emporix checks that the items that references in the body name exist"

// declareReferenceCheck declares w as a write of instances whose references
// Emporix checks where it sends validateReferencesParam true, or does not
// send it and byDefault says that Emporix then checks them. Emporix reads the
// items that references name with the write's token, so w takes the flag
// that sets the parameter with the read scope of the type, which it requests
// where Emporix checks references (paramScopes), and its help says that the
// token reads the instances of the type alone, so that a reference to another
// type is expected to fail the check.
func (in *instanceWrites) declareReferenceCheck(w *write, byDefault bool) {
	w.paramScopes = paramScopes{
		flag: paramFlag{
			param: validateReferencesParam, flag: "validate-references", choices: []string{queryTrue, queryFalse},
			usage: referencesUsage + " (" + strconv.FormatBool(byDefault) + " without it)",
		},
		neededValue: queryTrue, byDefault: byDefault, scopes: []string{customTypeReadScope},
	}
	if byDefault {
		w.long += "\n\nEmporix checks the items that the references of the instance name (unless\n" +
			"--validate-references false) with epx's token, which reads the instances of " + in.typeArg + "\n" +
			"alone: a reference to an instance of another custom entity type is then expected to fail\n" +
			"with 400 Bad Request, as one to the type fails without that scope, and one to another kind\n" +
			"of item, such as a product, may fail too."
		return
	}
	w.long += "\n\nWith --validate-references true, Emporix checks the items that the references of each\n" +
		"instance name with epx's token, which reads the instances of " + in.typeArg + " alone: an\n" +
		"instance with a reference to an item that does not exist fails with 500 Internal Server\n" +
		"Error (nothing is created), as one with a reference to an instance of\n" +
		"another custom entity type is expected to, and one with a reference to another kind of\n" +
		"item, such as a product, may."
}

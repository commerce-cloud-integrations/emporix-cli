package cli

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// schema is the part of a Schema Service schema shown in tables.
type schema struct {
	ID         string            `json:"id"`
	Name       localized         `json:"name"`
	Types      []string          `json:"types"`
	Attributes []json.RawMessage `json:"attributes"`
	Metadata   itemMetadata      `json:"metadata"`
}

// schemaManage is the scope of every write of the Schema Service, and of the
// parsing of a JSON schema file (newSchemaParseCommand).
const schemaManage = "schema.schema_manage"

// versionParam is the query parameter, and the flag, that names a version of
// a schema.
const versionParam = "version"

// newSchemasCommand reads, creates, changes and deletes the tenant's custom schemas
// (mixin definitions) and schema references, reads the entity types that have
// schemas and parses JSON schema files.
func newSchemasCommand(a *app) *cobra.Command {
	schemas := resource[schema]{
		use:   "schemas",
		short: "Read, create, change and delete custom schemas and schema references",
		long: "Reads custom schemas, the attribute sets that extend Emporix entities. TYPES lists the\n" +
			"entity types a schema applies to and ATTRIBUTES counts its top-level attributes (-o json\n" +
			"prints their definitions). `epx schemas references` reads schema references and\n" +
			"`epx schemas types` the entity types that have schemas.",
		path:    "/schema/{tenant}/schemas",
		filters: listFilters{kinds: filterable},
		view: view[schema]{
			singular:  "schema",
			header:    []string{columnID, columnName, "TYPES", "ATTRIBUTES", columnVersion},
			localized: true,
			row: func(s schema, lang language) []string {
				return []string{
					s.ID, s.Name.in(lang), strings.Join(s.Types, ","), strconv.Itoa(len(s.Attributes)),
					s.Metadata.Version.String(),
				}
			},
		},
	}
	types := typeNames("schema type").command(a, &command{
		use:   typesUse,
		short: "List the entity types that have schemas",
		long:  "Reads the entity types that have at least one schema, such as PRODUCT or CUSTOM_ENTITY.",
		req:   request{path: "/schema/{tenant}/types"},
		fetch: listing{read: (*emporix.Client).GetUnpaged},
	})
	more := append([]*cobra.Command{types, newSchemaReferencesCommand(a), newSchemaParseCommand(a)},
		schemaWriteCommands(a, schemas)...)
	return newResourceCommand(a, schemas, more...)
}

// schemaTypesHelp follows what the help of a write of schemas or schema
// references says of the entity types that the JSON names.
const schemaTypesHelp = "\nName entity types by their ids, with _, such as VENDOR_LOCATION."

// unknownDeletedHelp ends the help of the deletes of schemas and schema
// references, which Emporix answers alike for an unknown ID.
const unknownDeletedHelp = "\nEmporix answers an unknown ID as a deleted one (204), so epx exits with 0."

// schemaWriteCommands returns the writes of schemas: create, update ID,
// update-types ID and delete ID.
func schemaWriteCommands(a *app, schemas resource[schema]) []*cobra.Command {
	manage := []string{schemaManage}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a schema", method: http.MethodPost,
			long: "Creates a schema from the JSON in --file: its name (translations), the entity types it\n" +
				"extends and its attributes, such as `epx schemas parse -o json` prints them once a name and\n" +
				"types are added. It prints the id that Emporix answers with; one that exists already is a\n" +
				"conflict." + schemaTypesHelp,
			req: request{path: schemas.path}, scopes: manage, body: &payload{}, language: mapsOnly,
		}).command(a),
		(&write{
			use: schemas.itemUse("update"), short: "Replace a schema", method: http.MethodPut,
			long: "Replaces the schema ID with the JSON in --file: its name, types and attributes and, optionally,\n" +
				"as metadata.version, the version that it replaces (VERSION in `epx schemas get ID`; an older\n" +
				"one is a conflict). Every update, with a version or without, releases a new version of the\n" +
				"schema. An unknown ID is not found." + schemaTypesHelp,
			req: request{path: schemas.itemPath()}, scopes: manage, body: &payload{}, language: mapsOnly,
			ids: schemas.itemIDs(),
		}).command(a),
		(&write{
			use: schemas.itemUse("update-types"), short: "Change the entity types of a schema", method: http.MethodPut,
			long: "Sets the entity types of the schema ID to those of the JSON array in --file, such as\n" +
				`["PRODUCT", "ORDER"], in the version that --version names: the latest is VERSION in` + "\n" +
				"`epx schemas get ID`, and an earlier one may be named too. --mode ADD adds the types instead,\n" +
				"REMOVE removes them. It releases no new version of the schema. An unknown ID is not found." +
				schemaTypesHelp,
			req: request{path: schemas.itemPath() + "/types"}, scopes: manage, body: &payload{},
			ids: schemas.itemIDs(),
			params: queryFlags{
				required: []paramFlag{{param: versionParam, flag: versionParam, check: wholeNumber,
					usage: "`VERSION` of the schema whose types change, a whole number"}},
				optional: []paramFlag{{param: "mode", flag: "mode", choices: []string{"REPLACE", "ADD", "REMOVE"},
					usage: "`MODE` in which the types given change the schema's (REPLACE without it)"}},
			},
		}).command(a),
		itemDelete(schemas, "Delete a schema", "Deletes the schema ID, every version of it."+unknownDeletedHelp,
			manage).command(a),
	}
}

// newSchemaReferencesCommand reads, creates, changes and deletes schema references:
// schemas uploaded as a JSON schema file instead of defined by attributes.
// Each write uploads the file in a form with the JSON that describes the
// reference (referenceForm).
func newSchemaReferencesCommand(a *app) *cobra.Command {
	references := resource[schema]{
		use:   "references",
		noun:  "schema references",
		short: "Read, create, change and delete schema references",
		under: "schemas",
		long: "Reads schema references: schemas uploaded as a JSON schema file, for attribute types that\n" +
			"schemas cannot define. TYPES lists the entity types a reference applies to; the file's\n" +
			"URL is only in -o json.",
		path:    "/schema/{tenant}/references",
		filters: listFilters{kinds: filterable},
		view: view[schema]{
			singular:  "reference",
			header:    []string{columnID, columnName, "TYPES", columnVersion},
			localized: true,
			row: func(s schema, lang language) []string {
				return []string{s.ID, s.Name.in(lang), strings.Join(s.Types, ","), s.Metadata.Version.String()}
			},
		},
	}
	manage := []string{schemaManage}
	const (
		sent               = "\nIt sends both as a multipart/form-data body, the file as its file field."
		referenceTypesHelp = schemaTypesHelp + " A read of a reference shows that\ntype as VENDOR.LOCATION; a write " +
			"that sends it so makes Emporix answer reads of references\nwith 500."
	)
	return newResourceCommand(a, references,
		(&write{
			use: createUse, short: "Create a schema reference", method: http.MethodPost,
			long: "Creates a schema reference: the JSON schema file of --json-schema with the JSON in\n" +
				"--file, which names the reference and the entity types it applies to, such as\n" +
				`{"name": {"en": "Specifications"}, "types": ["PRODUCT"]}. It prints the id that Emporix answers` +
				"\nwith; one that exists already is a conflict." + sent + referenceTypesHelp,
			req: request{path: references.path}, scopes: manage, body: referenceForm(), language: mapsOnly,
		}).command(a),
		(&write{
			use: references.itemUse("update"), short: "Replace a schema reference", method: http.MethodPut,
			long: "Replaces the schema reference ID: the JSON schema file of --json-schema with the JSON\n" +
				"in --file, which names the reference, its types and, as metadata.version, the version that it\n" +
				"replaces (VERSION in `epx schemas references get ID`). An unknown ID is not found." + sent +
				referenceTypesHelp,
			req: request{path: references.itemPath()}, scopes: manage, body: referenceForm(), language: mapsOnly,
			ids: references.itemIDs(),
		}).command(a),
		itemDelete(references, "Delete a schema reference", "Deletes the schema reference ID. The JSON schema "+
			"file that it holds stays\nwhere Emporix stored it."+unknownDeletedHelp, manage).command(a),
	)
}

// referenceForm is the body of a write of a schema reference: a form of the
// JSON schema file of --json-schema and, as its body field, the JSON of
// --file.
func referenceForm() *formFile {
	return &formFile{
		form: form{jsonSchemaFile("JSON schema file in `PATH` that the reference holds")}, field: "body",
		usage: "JSON in `PATH`, sent byte for byte as the body field of the form",
	}
}

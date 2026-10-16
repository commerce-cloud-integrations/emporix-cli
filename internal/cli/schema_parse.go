package cli

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spf13/cobra"
)

// schemaAttribute is the part of an attribute of a parsed JSON schema file
// shown in tables.
type schemaAttribute struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Metadata struct {
		Required  *bool `json:"required"`
		Localized *bool `json:"localized"`
		ReadOnly  *bool `json:"readOnly"`
	} `json:"metadata"`
}

// jsonSchemaFile is the part of a form that sends a JSON schema file, named
// by --json-schema, described by usage.
func jsonSchemaFile(usage string) formPart { return filePart("file", "json-schema", usage) }

// newSchemaParseCommand returns `schemas parse`, which parses a JSON schema
// file into the schema that Emporix would make of it and stores nothing: an
// opt-in read, as Emporix allows it only with schema.schema_manage.
func newSchemaParseCommand(a *app) *cobra.Command {
	return reader[schemaAttribute]{
		view: view[schemaAttribute]{
			singular: "parsed schema",
			header:   []string{columnKey, columnType, "REQUIRED", "LOCALIZED", "READ ONLY"},
			row: func(s schemaAttribute, _ language) []string {
				return []string{
					s.Key, s.Type, formatBool(s.Metadata.Required), formatBool(s.Metadata.Localized),
					formatBool(s.Metadata.ReadOnly),
				}
			},
		},
		scopes: []string{schemaManage},
	}.command(a, &command{
		use:   "parse",
		short: "Parse a JSON schema file into a schema",
		long: "Parses the JSON schema file of --json-schema into the schema that Emporix makes of it, and\n" +
			"stores nothing: the table lists its attributes; -o json prints the schema, which\n" +
			"`epx schemas create` takes once its missing members, such as its name and types, are added,\n" +
			"or, with --id, `epx schemas update ID`.",
		req: request{path: "/schema/{tenant}/schemas/file"},
		fetch: bodyRead{method: http.MethodPost, fetch: single{items: schemaAttributes}, body: form{
			jsonSchemaFile("JSON schema file in `PATH` to parse"),
			fieldPart("id", "id", "`ID` of the schema that the file updates"),
		}},
	})
}

// schemaAttributes returns the attributes of the schema that the parsing of
// a JSON schema file answers with; a response without them is no schema.
func schemaAttributes(body json.RawMessage) ([]json.RawMessage, error) {
	var parsed struct {
		Attributes *[]json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Attributes == nil {
		return nil, errors.New("expected a schema with attributes")
	}
	return *parsed.Attributes, nil
}

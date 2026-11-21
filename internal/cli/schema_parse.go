package cli

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spf13/cobra"
)

type schemaAttribute struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Metadata struct {
		Required  *bool `json:"required"`
		Localized *bool `json:"localized"`
		ReadOnly  *bool `json:"readOnly"`
	} `json:"metadata"`
}

func jsonSchemaFile(usage string) formPart { return filePart("file", "json-schema", usage) }

// Emporix allows the parse only with schema.schema_manage.
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
		long: "Parses the JSON schema file of --json-schema into the schema that Emporix makes of it, and " +
			"stores nothing: the table lists its attributes; `-o json` prints the schema, which " +
			"`epx schemas create` takes once its missing members, such as its name and types, are added, " +
			"or, with --id, `epx schemas update ID`.",
		req: request{path: "/schema/{tenant}/schemas/file"},
		fetch: bodyRead{method: http.MethodPost, fetch: single{items: schemaAttributes}, body: form{
			jsonSchemaFile("JSON schema file in `PATH` to parse"),
			fieldPart("id", "id", "`ID` of the schema that the file updates"),
		}},
	})
}

func schemaAttributes(body json.RawMessage) ([]json.RawMessage, error) {
	var parsed struct {
		Attributes *[]json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Attributes == nil {
		return nil, errors.New("expected a schema with attributes")
	}
	return *parsed.Attributes, nil
}

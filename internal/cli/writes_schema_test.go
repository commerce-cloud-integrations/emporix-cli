package cli

import (
	"net/http"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const referenceBody = `{"name":{"en":"Specifications"},"types":["PRODUCT"],"metadata":{"version":1}}`

func schemaServiceWrites() []writeCase {
	manage := []string{"schema.schema_manage"}
	return []writeCase{
		{
			args: []string{"schemas", "create"}, file: `{"name":{"en":"Specs"},"types":["PRODUCT"],"attributes":[]}`,
			method: http.MethodPost, path: "/schema/acme1/schemas", scopes: manage, lang: "*", status: http.StatusCreated,
			body: `{"id":"s1"}`, want: "{\n  \"id\": \"s1\"\n}\n",
		},
		{
			args:   []string{"schemas", "update", "s1"},
			file:   `{"name":{"en":"Specs"},"types":["PRODUCT"],"attributes":[],"metadata":{"version":2}}`,
			method: http.MethodPut, path: "/schema/acme1/schemas/s1", scopes: manage, lang: "*", status: http.StatusNoContent,
		},
		{
			args: []string{"schemas", "update-types", "s1", "--version", "2", "--mode", "ADD"}, file: `["ORDER"]`,
			method: http.MethodPut, path: "/schema/acme1/schemas/s1/types", query: "mode=ADD&version=2", scopes: manage,
			status: http.StatusNoContent,
		},
		{
			args:   []string{"schemas", "references", "create", "--json-schema", "testdata/schema.json"},
			file:   referenceBody,
			form:   []sentPart{schemaFilePart(), {name: "body", contentType: emporix.JSONContentType, content: referenceBody}},
			method: http.MethodPost, path: "/schema/acme1/references", scopes: manage, lang: "*", status: http.StatusCreated,
			body: `{"id":"r1"}`, want: "{\n  \"id\": \"r1\"\n}\n",
		},
		{
			args:   []string{"schemas", "references", "update", "r1", "--json-schema", "testdata/schema.json"},
			file:   referenceBody,
			form:   []sentPart{schemaFilePart(), {name: "body", contentType: emporix.JSONContentType, content: referenceBody}},
			method: http.MethodPut, path: "/schema/acme1/references/r1", scopes: manage, lang: "*", status: http.StatusNoContent,
		},
	}
}

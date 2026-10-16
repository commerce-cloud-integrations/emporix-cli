package cli

import "net/http"

// exportedData is an export package's base64 data, longer than a table cell.
const exportedData = "W3siaWQiOiJ2ZWhpY2xlcyIsInR5cGUiOiJ2ZWhpY2xlIiwiZmllbGRzIjpbeyJwYXRoIjoiY29kZSJ9XX1d"

// schemaAndUnitReads returns the Schema and Unit Handling Service reads.
func schemaAndUnitReads() []readCase {
	return []readCase{
		{
			// FACTOR keeps the number as sent (1.0, not 1); a missing baseUnit or factor stays empty.
			args:    []string{"units", "list", "--language", "nl"},
			path:    "/unit-handling/acme1/units",
			filters: listFilters{kinds: sortable},
			body: `[
		{"code":"box","name":{"en":"Carton","nl":"Doos"},"type":"box_count","baseUnit":true,"factor":1.0},
		{"code":"g","name":{"en":"gram"},"type":"mass","symbol":"g","baseUnit":false,"factor":0.001},
		{"code":"x","name":{"nl":"Iets"}}]`,
			want: "CODE  NAME       TYPE       SYMBOL  BASE   FACTOR\n" +
				"box   Doos       box_count          true   1.0\n" +
				"g     gram (en)  mass       g       false  0.001\n" +
				"x     Iets                                 \n",
		},
		{
			args:    []string{"schemas", "list", "--language", "nl"},
			path:    "/schema/acme1/schemas",
			filters: listFilters{kinds: filterable},
			body: `[
		{"id":"acmestore-import-v1","name":{"en":"Import","nl":"Importeren"},"types":["PRODUCT","CATEGORY"],
		 "attributes":[{"key":"identifiers","type":"ARRAY"},{"key":"weight","type":"DECIMAL"}],
		 "metadata":{"version":2,"url":"https://example.invalid/schema.json"}},
		{"id":"bare"}]`,
			want: "ID                   NAME        TYPES             ATTRIBUTES  VERSION\n" +
				"acmestore-import-v1  Importeren  PRODUCT,CATEGORY  2           2\n" +
				"bare                                               0           \n",
		},
		{
			args: []string{"schemas", "get", "s1", "--language", "nl"},
			path: "/schema/acme1/schemas/s1",
			mode: getWhole,
			body: `{"id":"s1","name":{"en":"Import","nl":"Importeren"},"types":["PRODUCT"],"metadata":{"version":2}}`,
			want: "ID  NAME        TYPES    ATTRIBUTES  VERSION\ns1  Importeren  PRODUCT  0           2\n",
		},
		{
			args: []string{"units", "get", "g", "--language", "de"},
			path: "/unit-handling/acme1/units/g",
			mode: getWhole,
			body: `{"code":"g","name":{"en":"gram","de":"Gramm"},"type":"mass","symbol":"g","factor":0.001}`,
			want: "CODE  NAME   TYPE  SYMBOL  BASE  FACTOR\ng     Gramm  mass  g             0.001\n",
		},
		{
			// The types are strings, sent whole; they keep the order Emporix sends.
			args: []string{"units", "types"},
			path: "/unit-handling/acme1/types",
			mode: getWhole,
			body: `["mass","box_count","length"]`,
			want: "TYPE\nmass\nbox_count\nlength\n",
		},
	}
}

// schemaServiceReads returns the Schema Service reads besides schemas: schema
// types, schema references, custom entity types and their instances.
func schemaServiceReads() []readCase {
	return []readCase{
		{
			// The types keep the order Emporix sends.
			args: []string{"schemas", "types"},
			path: "/schema/acme1/types",
			mode: getWhole,
			body: `["PRODUCT","CUSTOM_ENTITY"]`,
			want: "TYPE\nPRODUCT\nCUSTOM_ENTITY\n",
		},
		{
			args:    []string{"schemas", "references", "list", "--language", "nl"},
			path:    "/schema/acme1/references",
			filters: listFilters{kinds: filterable},
			body: `[{"id":"r1","name":{"en":"Product attributes","nl":"Productkenmerken"},"types":["PRODUCT","CART"],
				"metadata":{"version":1,"url":"https://example.invalid/r1_v1.json"}},{"id":"bare"}]`,
			want: "ID    NAME              TYPES         VERSION\n" +
				"r1    Productkenmerken  PRODUCT,CART  1\n" +
				"bare                                  \n",
		},
		{
			args: []string{"schemas", "references", "get", "r/1", "--language", "nl"},
			path: "/schema/acme1/references/r%2F1",
			mode: getWhole,
			body: `{"id":"r/1","name":{"en":"Product attributes","nl":"Productkenmerken"},"types":["PRODUCT"],
				"metadata":{"version":2}}`,
			want: "ID   NAME              TYPES    VERSION\nr/1  Productkenmerken  PRODUCT  2\n",
		},
		{
			args:    []string{"custom-entities", "list", "--language", "de"},
			path:    "/schema/acme1/custom-entities",
			filters: listFilters{kinds: filterable},
			body: `[{"id":"DOCUMENT","name":{"en":"Documents","de":"Dokumente"},"metadata":{"version":2}},
				{"id":"CUSTOM_PRODUCT","name":{"en":"Custom product"}}]`,
			want: "ID              NAME                 VERSION\n" +
				"DOCUMENT        Dokumente            2\n" +
				"CUSTOM_PRODUCT  Custom product (en)  \n",
		},
		{
			args: []string{"custom-entities", "get", "DOCUMENT", "--language", "de"},
			path: "/schema/acme1/custom-entities/DOCUMENT",
			mode: getWhole,
			body: `{"id":"DOCUMENT","name":{"en":"Documents","de":"Dokumente"},"metadata":{"version":1}}`,
			want: "ID        NAME       VERSION\nDOCUMENT  Dokumente  1\n",
		},
		{
			// The scope of the type, in lower case, is requested (TestChildScopes: instead of the
			// parent's); OWNER joins type and id (the legal entity of a customer stays in -o json).
			args:    []string{"custom-entities", "instances", "list", "DOCUMENT", "--language", "nl"},
			path:    "/schema/acme1/custom-entities/DOCUMENT/instances",
			scopes:  []string{"custom.document_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"123","name":{"en":"Report","nl":"Rapport"},"type":"DOCUMENT",
				"owner":{"type":"EMPLOYEE","userId":"u1"},"mixins":{"attributes":{"pages":3}},"metadata":{"version":3}},
				{"id":"124","name":{"en":"Invoice"},"owner":{"type":"CUSTOMER","userId":"c1","legalEntityId":"le1"}}]`,
			want: "ID   NAME          OWNER        VERSION\n" +
				"123  Rapport       EMPLOYEE u1  3\n" +
				"124  Invoice (en)  CUSTOMER c1  \n",
		},
		{
			// The type, then the ID, address one instance; the type's scope keeps its digits and
			// underscores.
			args:   []string{"custom-entities", "instances", "get", "TECH_SHEET2", "123", "--language", "nl"},
			path:   "/schema/acme1/custom-entities/TECH_SHEET2/instances/123",
			scopes: []string{"custom.tech_sheet2_read"},
			mode:   getWhole,
			body: `{"id":"123","name":{"en":"Report","nl":"Rapport"},"type":"DOCUMENT",
				"owner":{"type":"SERVICE","userId":"client-1"},"metadata":{"version":4}}`,
			want: "ID   NAME     OWNER             VERSION\n" +
				"123  Rapport  SERVICE client-1  4\n",
		},
		{
			// Opt-in, sent as POST: the body is the array of the types given; DATA is cut in the table.
			args:   []string{"custom-entities", "export", "--type", "CAR", "--type", "MANUFACTURER", "--allow-write-scope"},
			path:   "/schema/acme1/custom-entities/export",
			scopes: []string{"schema.schema_manage"},
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `["CAR","MANUFACTURER"]`,
			body:   `{"exportedAt":"2026-10-01T07:05:18.857Z","data":"` + exportedData + `"}`,
			want: "DATA                                                          EXPORTED AT\n" +
				"W3siaWQiOiJ2ZWhpY2xlcyIsInR5cGUiOiJ2ZWhpY2xlIiwiZmllbGRzIjp…  2026-10-01T07:05:18.857Z\n",
		},
	}
}

// schemaParseReads returns the parsing of a JSON schema file, a read with a
// form, without and with the schema that the file updates.
func schemaParseReads() []readCase {
	return []readCase{
		{
			// Opt-in, sent as POST with a form: the JSON schema file with its name; the table lists the
			// attributes that Emporix parsed.
			args:   []string{"schemas", "parse", "--json-schema", "testdata/schema.json", "--allow-write-scope"},
			path:   "/schema/acme1/schemas/file",
			scopes: []string{"schema.schema_manage"},
			mode:   getWhole,
			method: http.MethodPost,
			form:   []sentPart{schemaFilePart()},
			body: `{"attributes":[{"key":"color","type":"TEXT","metadata":{"required":false,"localized":true}},
				{"key":"size","type":"NUMBER"}],"metadata":{"version":1}}`,
			want: "KEY    TYPE    REQUIRED  LOCALIZED  READ ONLY\n" +
				"color  TEXT    false     true       \n" +
				"size   NUMBER                       \n",
		},
		{
			// --id adds the schema that the file updates as a field.
			args: []string{"schemas", "parse", "--json-schema", "testdata/schema.json", "--id", "s1",
				"--allow-write-scope"},
			path:   "/schema/acme1/schemas/file",
			scopes: []string{"schema.schema_manage"},
			mode:   getWhole,
			method: http.MethodPost,
			form:   []sentPart{schemaFilePart(), {name: "id", content: "s1"}},
			body:   `{"id":"s1","attributes":[{"key":"color","type":"TEXT","metadata":{"readOnly":true}}]}`,
			want:   "KEY    TYPE  REQUIRED  LOCALIZED  READ ONLY\ncolor  TEXT                       true\n",
		},
	}
}

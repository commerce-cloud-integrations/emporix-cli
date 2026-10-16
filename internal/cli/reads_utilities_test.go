package cli

// utilityReads returns the reads of the utility services: sequential IDs and
// the change history.
func utilityReads() []readCase {
	const scope = "sequentialid.schema_view"
	schemas := `[{"id":"6512a0f3c1d2e3f4a5b6c7d8","name":"Orders","schemaType":"orderNoSequence","active":true,
		"preText":"EC-__year__-","postText":"-D","startValue":1,"maxValue":999999999,"numberOfDigits":9,"counter":45684},
		{"id":"6512a0f3c1d2e3f4a5b6c7d9","name":"Customers","schemaType":"customerNoSequence","active":false,
		"preText":null,"postText":null,"numberOfDigits":4,"counter":109,"maxValue":null,"metadata":null,"placeholders":null}]`
	return []readCase{
		{
			// The schemas come whole; PREFIX and SUFFIX keep placeholders as sent, null as empty.
			args:   []string{"sequential-ids", "list"},
			path:   "/sequential-id/acme1/schemas",
			scopes: []string{scope},
			mode:   getWhole,
			body:   schemas,
			want: "ID                        TYPE                NAME       ACTIVE  PREFIX        DIGITS  SUFFIX  COUNTER  " +
				"MAX\n" +
				"6512a0f3c1d2e3f4a5b6c7d8  orderNoSequence     Orders     true    EC-__year__-  9       -D      45684    " +
				"999999999\n" +
				"6512a0f3c1d2e3f4a5b6c7d9  customerNoSequence  Customers  false                 4               109      " +
				"\n",
		},
		{
			args:   []string{"sequential-ids", "get", "6512a0f3c1d2e3f4a5b6c7d8"},
			path:   "/sequential-id/acme1/schemas/6512a0f3c1d2e3f4a5b6c7d8",
			scopes: []string{scope},
			mode:   getWhole,
			body: `{"id":"6512a0f3c1d2e3f4a5b6c7d8","name":"Orders","schemaType":"orderNoSequence","active":true,
				"preText":"EC-","numberOfDigits":9,"counter":0,"maxValue":99}`,
			want: "ID                        TYPE             NAME    ACTIVE  PREFIX  DIGITS  SUFFIX  COUNTER  MAX\n" +
				"6512a0f3c1d2e3f4a5b6c7d8  orderNoSequence  Orders  true    EC-     9               0        99\n",
		},
		{
			args:   []string{"sequential-ids", "for-type", "orderNoSequence"},
			path:   "/sequential-id/acme1/schemas/types/orderNoSequence",
			scopes: []string{scope},
			mode:   getWhole,
			body: `[{"id":"6512a0f3c1d2e3f4a5b6c7d8","name":"Orders","schemaType":"orderNoSequence","active":true,
				"numberOfDigits":9,"counter":7}]`,
			want: "ID                        TYPE             NAME    ACTIVE  PREFIX  DIGITS  SUFFIX  COUNTER  MAX\n" +
				"6512a0f3c1d2e3f4a5b6c7d8  orderNoSequence  Orders  true            9               7        \n",
		},
		{
			// The changes come in pages that hold their total; FIELDS names the changed fields, sorted
			// (their values are only in -o json).
			args:    []string{"changelog", "list"},
			path:    "/changelog/acme1/changelogs",
			scopes:  []string{"changelog.changelog_read"},
			mode:    getAllPageObjects,
			filters: listFilters{kinds: queryable},
			body: `{"items":[{"at":"2026-10-01T13:01:29.123Z","type":"update","entity":"order","entityId":"EON1001",
				"paths":{"status":{"before":"CREATED","after":"CONFIRMED"},"mixins.payment.token":{"after":"S1"}},
				"actor":"Jane Doe"},{"at":"2026-10-02T10:00:00.000Z","type":"create","entity":"product","entityId":"100200",
				"name":"Monkey wrench","actor":"system"}],"page":1,"size":100,"totalElements":2,"totalPages":1}`,
			want: "AT                        TYPE    ENTITY   ENTITY ID  NAME           ACTOR     FIELDS\n" +
				"2026-10-01T13:01:29.123Z  update  order    EON1001                   Jane Doe  mixins.payment.token, status\n" +
				"2026-10-02T10:00:00.000Z  create  product  100200     Monkey wrench  system    \n",
		},
	}
}

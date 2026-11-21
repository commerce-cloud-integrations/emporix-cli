package cli

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func numbered(n int, item func(i int) string) string {
	items := make([]string, n)
	for i := range items {
		items[i] = item(i)
	}
	return strings.Join(items, ",")
}

func keys(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s%03d", prefix, i)
	}
	return out
}

func TestCompletionOffersUpTo100IDsThatStartWithWhatIsTyped(t *testing.T) {
	settings := "[" + numbered(120, func(i int) string { return fmt.Sprintf(`{"key":"a%03d"}`, i) }) + "," +
		numbered(30, func(i int) string { return fmt.Sprintf(`{"key":"b%03d"}`, i) }) + "]"
	configurations := map[string]http.HandlerFunc{"/configuration/acme1/configurations": respond(http.StatusOK, settings)}
	mixins := "{" + numbered(101, func(i int) string { return fmt.Sprintf(`"m%03d":{}`, i) }) + "}"
	for _, tt := range []completionCase{
		{
			args: []string{"configurations", "get", "b"}, routes: configurations,
			want:     keys("b", 30),
			requests: []string{"/configuration/acme1/configurations"}, scopes: []string{configurationScope},
		},
		{
			args: []string{"configurations", "get", ""}, routes: configurations,
			want:     keys("a", 100),
			requests: []string{"/configuration/acme1/configurations"}, scopes: []string{configurationScope},
		},
		{
			args: []string{"sites", "mixins", "get", "AcmeStore_NL", ""},
			routes: map[string]http.HandlerFunc{
				"/site/acme1/sites/AcmeStore_NL/mixins": respond(http.StatusOK, mixins),
			},
			want:     keys("m", 100),
			requests: []string{"/site/acme1/sites/AcmeStore_NL/mixins"},
		},
		{
			args: []string{"iam", "users", "get", ""},
			routes: map[string]http.HandlerFunc{
				"/iam/acme1/users": respond(http.StatusOK, `[{"id":"u1"},{"id":"u2"}]`),
			},
			want:     []string{"u1", "u2"},
			requests: []string{"/iam/acme1/users?pageNumber=1&pageSize=500"}, scopes: []string{"iam.user_read"},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionDescribesOnlyTheIDsOfItems(t *testing.T) {
	tt := completionCase{
		args: []string{"sequential-ids", "for-type", ""},
		routes: map[string]http.HandlerFunc{"/sequential-id/acme1/schemas": respond(http.StatusOK,
			`[{"id":"s1","schemaType":"orderNoSequence","name":"Orders NL"},`+
				`{"id":"s2","schemaType":"orderNoSequence","name":"Orders DE"},`+
				`{"id":"s3","schemaType":"invoiceNoSequence","name":"Invoices"}]`)},
		want:     []string{"orderNoSequence", "invoiceNoSequence"},
		requests: []string{"/sequential-id/acme1/schemas"},
		scopes:   []string{"sequentialid.schema_view"},
	}
	tt.check(t)
}

func TestCompletionOfWritesAndOtherReadsReadsTheirLists(t *testing.T) {
	products := map[string]http.HandlerFunc{"/product/acme1/products": respond(http.StatusOK, productPage)}
	for _, tt := range []completionCase{
		{
			args: []string{"brands", "update", "B"}, // typed, as Cobra offers the required --file before it
			routes: map[string]http.HandlerFunc{
				"/brand/brands": respond(http.StatusOK, `[{"id":"B1","name":"Acme"}]`),
			},
			want:     []string{"B1|Acme"},
			requests: []string{"/brand/brands?pageNumber=1&pageSize=100"},
		},
		{
			args: []string{"vendors", "delete", ""},
			routes: map[string]http.HandlerFunc{
				"/vendor/acme1/vendors": respond(http.StatusOK, `[{"id":"v1","name":"Acme Supply"}]`),
			},
			want:     []string{"v1|Acme Supply"},
			requests: []string{"/vendor/acme1/vendors?pageNumber=1&pageSize=100"},
			scopes:   []string{vendorScope},
		},
		{
			args: []string{"categories", "assignments", "update", "C1", ""}, routes: products,
			want:     []string{"100200|Shirt", "200300|Hose (de)"},
			requests: []string{"/product/acme1/products?pageNumber=1&pageSize=100"},
			scopes:   []string{productReadUnpublishedScope},
		},
		{
			args:     []string{"sites", "mixins", "delete", ""},
			routes:   map[string]http.HandlerFunc{"/site/acme1/sites": respond(http.StatusOK, sitePage)},
			want:     []string{"AcmeStore_NL|Acme NL", "AcmeStore_DE|Acme DE"},
			requests: []string{"/site/acme1/sites?pageNumber=1&pageSize=100&totalCount=true"},
		},
		{
			args: []string{"search", "queries", "delete", ""},
			routes: map[string]http.HandlerFunc{
				"/search/acme1/search/queries": respond(http.StatusOK, `[{"id":"q1","type":"vehicle","name":"By make"}]`),
			},
			want:     []string{"customer", "legal-entity", "vehicle"},
			requests: []string{"/search/acme1/search/queries?pageNumber=1&pageSize=100"},
			scopes:   []string{searchScope},
		},
		{
			args: []string{"reward-points", "balance", ""},
			routes: map[string]http.HandlerFunc{
				"/reward-points/summaryBatch": respond(http.StatusOK, `[{"customerId":"c1"},{"customerId":"c2"}]`),
			},
			want:     []string{"c1", "c2"},
			requests: []string{"/reward-points/summaryBatch"},
			scopes:   []string{"rewardspoints.rewardpoints_view"},
		},
		{
			args: []string{"shipping", "zones", "methods", "create", "--site", "AcmeStore_NL", "Z"},
			routes: map[string]http.HandlerFunc{
				"/shipping/acme1/AcmeStore_NL/zones": respond(http.StatusOK, `[{"id":"Z1","name":{"en":"Benelux"}}]`),
			},
			want:     []string{"Z1|Benelux"},
			requests: []string{"/shipping/acme1/AcmeStore_NL/zones"},
		},
		{
			args: []string{"category-trees", "rebuild", ""},
			routes: map[string]http.HandlerFunc{"/category/acme1/category-trees": respond(http.StatusOK,
				`[{"id":"C1","name":"Men","subcategories":[{"id":"C2","name":"Shirts"}]}]`)},
			want:     []string{"C1|Men"},
			requests: []string{"/category/acme1/category-trees?showUnpublished=true"},
			scopes:   []string{"category.category_read_unpublished"},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionOfDeletesAndActionsReadsTheirLists(t *testing.T) {
	for _, tt := range []completionCase{
		{
			args: []string{"categories", "assignments", "delete", "C1", ""},
			routes: map[string]http.HandlerFunc{"/category/acme1/categories/C1/assignments": respond(http.StatusOK,
				`[{"id":"A1","ref":{"id":"100200","type":"PRODUCT","localizedName":{"en":"Shirt"}}}]`)},
			want:     []string{"100200"},
			requests: []string{"/category/acme1/categories/C1/assignments?pageNumber=1&pageSize=100&showUnpublished=true"},
			scopes:   []string{"category.category_read_unpublished"},
		},
		{
			args: []string{"availability", "delete", "--site", "AcmeStore_NL", ""},
			routes: map[string]http.HandlerFunc{"/availability/acme1/availability/site/AcmeStore_NL": respond(
				http.StatusOK, `[{"productId":"100200","site":"AcmeStore_NL"}]`)},
			want:     []string{"100200"},
			requests: []string{"/availability/acme1/availability/site/AcmeStore_NL?pageNumber=1&pageSize=100"},
		},
		{
			args: []string{"payment-transactions", "cancel", ""},
			routes: map[string]http.HandlerFunc{"/payment-gateway/acme1/transactions": respond(http.StatusOK,
				`[{"id":"T1","provider":"INVOICE"}]`)},
			want:     []string{"T1"},
			requests: []string{"/payment-gateway/acme1/transactions?pageNumber=1&pageSize=100"},
			scopes:   []string{"payment-gateway.paymenttransactions_read"},
		},
		{
			args: []string{"iam", "groups", "delete", ""},
			routes: map[string]http.HandlerFunc{"/iam/acme1/groups": respond(http.StatusOK,
				`[{"id":"g1","code":"BO_ADMIN","name":{"en":"Administrators"}}]`)},
			want:     []string{"g1|Administrators"},
			requests: []string{"/iam/acme1/groups?pageNumber=1&pageSize=100&sort=id%3Aasc"},
			scopes:   []string{"iam.group_read"},
		},
		{
			args: []string{"iam", "groups", "users", "delete", "G1", ""},
			routes: map[string]http.HandlerFunc{"/iam/acme1/groups/G1/users": respond(http.StatusOK,
				`[{"id":"m1","userId":"u1","userType":"EMPLOYEE"}]`)},
			want:     []string{"u1"},
			requests: []string{"/iam/acme1/groups/G1/users?pageNumber=1&pageSize=100"},
			scopes:   []string{"iam.user_read"},
		},
		{
			args: []string{"ai", "agents", "create", "t"},
			routes: map[string]http.HandlerFunc{"/ai-service/acme1/agentic/templates": respond(http.StatusOK,
				`[{"id":"t1","name":{"en":"Complaints"}}]`)},
			want:     []string{"t1|Complaints"},
			requests: []string{"/ai-service/acme1/agentic/templates?pageNumber=1&pageSize=100"},
			scopes:   []string{aiScope},
		},
		{
			args: []string{"custom-entities", "instances", "delete", "CUSTOM_DOCUMENT", ""},
			routes: map[string]http.HandlerFunc{"/schema/acme1/custom-entities/CUSTOM_DOCUMENT/instances": respond(
				http.StatusOK, `[{"id":"123","name":{"en":"Report"}}]`)},
			want:     []string{"123|Report"},
			requests: []string{"/schema/acme1/custom-entities/CUSTOM_DOCUMENT/instances?pageNumber=1&pageSize=100"},
			scopes:   []string{"custom.custom_document_read"},
		},
		{
			args: []string{"segments", "customers", "delete", "s1", ""},
			routes: map[string]http.HandlerFunc{"/customer-segment/acme1/segments/s1/customers": respond(http.StatusOK,
				`[{"customer":{"id":"c-7f3a","firstName":"Ada","lastName":"Example"}}]`)},
			want:     []string{"c-7f3a|Ada Example"},
			requests: []string{"/customer-segment/acme1/segments/s1/customers?pageNumber=1&pageSize=100&sort=id"},
			scopes:   []string{"customersegment.segment_read"},
		},
		{
			args: []string{"customers", "addresses", "delete", "0100", ""},
			routes: map[string]http.HandlerFunc{"/customer/acme1/customers/0100/addresses": respond(http.StatusOK,
				`[{"id":"a1","contactName":"Ada Example"}]`)},
			want:     []string{"a1"},
			requests: []string{"/customer/acme1/customers/0100/addresses"},
			scopes:   []string{"customer.customer_read"},
		},
		{
			args: []string{"orders", "split", ""},
			routes: map[string]http.HandlerFunc{"/order-v2/acme1/salesorders": respond(http.StatusOK,
				`[{"id":"O1"}]`)},
			want:     []string{"O1"},
			requests: []string{"/order-v2/acme1/salesorders?pageNumber=1&pageSize=100"},
			scopes:   []string{"order.order_read"},
		},
	} {
		tt.check(t)
	}
}

func TestCompletionThatCouldNeverReadPanics(t *testing.T) {
	things := reader[thing]{view: thingView()}
	ofSite := listedIDs(things, request{path: "/sites/" + placeholderSite + "/things"}, listing{}, columnID)
	for name, c := range map[string]*command{
		"an argument the command lacks": {
			ids: argumentIDs{"THING_ID": listedIDs(things, request{path: "/things"}, listing{}, columnID)},
		},
		"a list of a site, without --site":        {ids: argumentIDs{columnID: ofSite}},
		"values of an argument the command lacks": {choices: map[string][]string{"THING_ID": {"T1"}}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			c.use, c.req, c.fetch = "get ID", request{path: "/things/" + placeholderID}, single{items: oneItem}
			things.command(&app{}, c)
		}()
	}
}

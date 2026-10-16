package cli

import (
	"net/http"
	"strings"
	"testing"
)

// productSwitches is the query of a write of products without its flags:
// the switches, sent false.
const productSwitches = "skipRelatedItemsValidation=false&skipVariantGeneration=false"

// productWriteCases are the writes of products and product templates, with
// the scopes that the reference names for them and the Content-Language of
// their localized text.
func productWriteCases() []writeCase {
	manage := []string{productManageScope}
	const products, bulk = "/product/acme1/products", "/product/acme1/products/bulk"
	const two = `[{"id":"100200","name":"Chair"},{"id":"100201","name":"Table"}]`
	return []writeCase{
		{
			args:   []string{"products", "create"},
			file:   "{\"id\": \"100200\", \"code\": \"100200\", \"productType\": \"BASIC\"}\n",
			method: http.MethodPost, path: products, query: productSwitches, scopes: manage, lang: "*",
			status: http.StatusCreated, body: `{"id":"100200","yrn":"urn:yaas:saasag:caasproduct:product:acme1;100200"}`,
			want: "{\n  \"id\": \"100200\",\n  \"yrn\": \"urn:yaas:saasag:caasproduct:product:acme1;100200\"\n}\n",
		},
		{
			args: []string{"products", "create", "--allow-publish", "--content-language", "en", "--do-index", "false",
				"--skip-variant-generation", "--skip-related-items-validation"},
			file: `{"id":"100200","name":"Chair","published":true}`, method: http.MethodPost, path: products,
			query:  "doIndex=false&skipRelatedItemsValidation=true&skipVariantGeneration=true",
			scopes: []string{productManageScope, "product.product_publish"}, lang: "en",
			status: http.StatusCreated, body: `{"id":"100200"}`, want: "{\n  \"id\": \"100200\"\n}\n",
		},
		{
			args: []string{"products", "update", "100200"}, file: `{"code":"100200","metadata":{"version":3}}`,
			method: http.MethodPut, path: products + "/100200", query: productSwitches, scopes: manage, lang: "*",
			status: http.StatusNoContent,
		},
		{
			args: []string{"products", "update", "100200", "--allow-unpublish", "--allow-publish"},
			file: `{"code":"100200","published":false}`, method: http.MethodPut, path: products + "/100200",
			query: productSwitches, lang: "*",
			scopes: []string{productManageScope, "product.product_publish", "product.product_unpublish"},
			status: http.StatusCreated, body: `{"id":"100200"}`, want: "{\n  \"id\": \"100200\"\n}\n",
		},
		{
			args: []string{"products", "patch", "100 200"}, file: `{"name":{"en":"Chair"}}`, method: http.MethodPatch,
			path: products + "/100%20200", query: productSwitches, scopes: manage, lang: "*",
			status: http.StatusNoContent,
		},
		{
			args: []string{"products", "bulk-create"}, file: two, method: http.MethodPost, path: bulk,
			query: productSwitches, scopes: manage, lang: "*", status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"100200","code":201},{"index":1,"id":"100201","code":201}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"id\": \"100200\",\n    \"index\": 0\n  },\n" +
				"  {\n    \"code\": 201,\n    \"id\": \"100201\",\n    \"index\": 1\n  }\n]\n",
		},
		{
			args: []string{"products", "bulk-update"}, file: "[\n  {\"id\": \"100200\"}\n]\n", method: http.MethodPut,
			path: bulk, query: productSwitches, scopes: manage, lang: "*", status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"100200","code":204}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"100200\",\n    \"index\": 0\n  }\n]\n",
		},
		{
			args: []string{"products", "import"}, file: two, method: http.MethodPut, path: bulk, query: productSwitches,
			scopes: manage, lang: "*", status: http.StatusMultiStatus, body: `[{"code":204},{"code":201}]`,
			want: "[\n  {\n    \"code\": 204\n  },\n  {\n    \"code\": 201\n  }\n]\n",
		},
		{
			args: []string{"products", "recalculate"}, file: `{"productIds":["100200","100201"]}`,
			method: http.MethodPost, path: products + "/recalculate", scopes: manage, status: http.StatusAccepted,
			body: `{"jobs":[{"id":"job-1","parentRoot":"100200","status":"PENDING"}],"skippedProductIds":["100201"]}`,
			want: "{\n  \"jobs\": [\n    {\n      \"id\": \"job-1\",\n      \"parentRoot\": \"100200\",\n" +
				"      \"status\": \"PENDING\"\n    }\n  ],\n  \"skippedProductIds\": [\n    \"100201\"\n  ]\n}\n",
		},
		{
			args:   []string{"products", "recalculate", "--product", "100200", "--product", "100 201"},
			sent:   `{"productIds":["100200","100 201"]}`,
			method: http.MethodPost, path: products + "/recalculate", scopes: manage, status: http.StatusAccepted,
			body: `{"jobs":[],"skippedProductIds":[]}`, want: "{\n  \"jobs\": [],\n  \"skippedProductIds\": []\n}\n",
		},
		{
			args: []string{"product-templates", "create"}, file: `{"name":{"en":"Furniture"},"attributes":[]}`,
			method: http.MethodPost, path: "/product/acme1/product-templates", lang: "*",
			scopes: []string{"product.product_template_manage"},
			status: http.StatusCreated, body: `{"id":"tmpl-1"}`, want: "{\n  \"id\": \"tmpl-1\"\n}\n",
		},
		{
			args: []string{"product-templates", "update", "tmpl-1", "--content-language", "de"},
			file: `{"name":"Möbel","attributes":[],"metadata":{"version":2}}`, method: http.MethodPut,
			path: "/product/acme1/product-templates/tmpl-1", scopes: []string{"product.product_template_manage"},
			lang: "de", status: http.StatusNoContent,
		},
	}
}

// TestProductWritesStateTheirOptionalScopes checks the help of the product
// writes: each that may publish a product names --allow-publish and the
// scope it adds, those that replace or change products --allow-unpublish
// too, with what Emporix refuses without them, before the scope that each
// requests, and so do the flags' usages;
// recalculate names neither. The products group states the scopes of its
// reads, with their opt-in, and those of its writes.
func TestProductWritesStateTheirOptionalScopes(t *testing.T) {
	const publish = "\n\n--allow-publish also requests the product.product_publish scope. Without it, Emporix\n" +
		"refuses a product that the write publishes (403 Forbidden, exit 3)."
	const unpublish = "\n--allow-unpublish also requests the product.product_unpublish scope. Without it, Emporix\n" +
		"refuses a product that the write unpublishes (403 Forbidden, exit 3)."
	const requires = "\n\nRequires the product.product_manage scope.\n\nUsage:"
	const usage = "allow the write to publish products, which needs a scope of its own (requests " +
		"product.product_publish too)"
	for _, tt := range []struct {
		use       string
		unpublish bool
	}{{"create", false}, {"bulk-create", false}, {"update", true}, {"patch", true}, {"bulk-update", true},
		{"import", true}} {
		want := publish + requires
		if tt.unpublish {
			want = publish + unpublish + requires
		}
		code, help, _ := runCLI(t, nil, "products", tt.use, "--help")
		if code != ExitOK || !strings.Contains(help, want) || !strings.Contains(help, usage) ||
			strings.Contains(help, unpublish) != tt.unpublish {
			t.Errorf("products %s: exit %d, help:\n%s\nwant %q", tt.use, code, help, want)
		}
	}
	code, help, _ := runCLI(t, nil, "products", "recalculate", "--help")
	if code != ExitOK || strings.Contains(help, "--allow-") || !strings.Contains(help, "sending it."+requires) {
		t.Errorf("products recalculate: exit %d, help:\n%s", code, help)
	}
	code, help, _ = runCLI(t, nil, "products", "--help")
	const group = "Its reads require different OAuth scopes; the help of each names them.\n" + optInSentence +
		"\nIts writes require the product.product_manage scope.\n\nUsage:"
	if code != ExitOK || !strings.HasPrefix(help, "Reads and changes products.") || !strings.Contains(help, group) {
		t.Errorf("products: exit %d, help:\n%s\nwant %q", code, help, group)
	}
}

// TestProductWritesRefuseInvalidInput checks that a product write without a
// body, recalculate without --product or --file, an update whose ID is the
// bulk endpoint, --allow-unpublish on a create and an invalid --do-index
// are usage errors (exit 2) that send nothing, also for a dry run.
func TestProductWritesRefuseInvalidInput(t *testing.T) {
	body := writeBody(t, `{"id":"100200"}`)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"products", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"products", "recalculate"}, want: "at least one of the flags in the group [file product] is " +
			"required"},
		{args: []string{"products", "update", "bulk", "--file", body},
			want: `invalid ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"products", "bulk-create", "--file", body, "--allow-unpublish"},
			want: "unknown flag: --allow-unpublish"},
		{args: []string{"products", "patch", "100200", "--file", body, "--do-index", "no"},
			want: `invalid argument "no" for "--do-index" flag: use true or false`},
		{args: []string{"product-templates", "update", "--file", body}, want: "accepts 1 arg(s), received 0"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(tt.args, flag)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, flag, code, stdout, stderr,
					tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

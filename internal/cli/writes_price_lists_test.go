package cli

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// A price model as create takes it, a price list and a price of a price list.
const (
	priceModelJSON = `{"name":{"en":"Per kilogram"},"includesTax":true,"includesMarkup":false,` +
		`"measurementUnit":{"quantity":1,"unitCode":"kg"},` +
		`"tierDefinition":{"tierType":"BASIC","tiers":[{"minQuantity":{"quantity":0,"unitCode":"kg"}}]}}`
	priceListJSON = `{"name":{"en":"Wholesale"},"currency":"EUR","siteCode":"AcmeStore_NL","countries":["NL"],` +
		`"validity":{"from":"2026-01-01T00:00:00.000Z"}}`
	listPriceJSON = `{"itemId":{"itemType":"PRODUCT","id":"100200"},"priceModelId":"pm1",` +
		`"tierValues":[{"id":"t1","priceValue":15.90}]}`
)

// priceModelAndListWrites are the writes of price models, price lists and
// the prices of a price list.
func priceModelAndListWrites() []writeCase {
	models, lists := []string{priceModelManage}, []string{priceListManage}
	const prices = "/price/acme1/price-lists/pl1/prices"
	return []writeCase{
		{
			args: []string{"price-models", "create"}, file: priceModelJSON, method: http.MethodPost,
			path: "/price/acme1/priceModels", scopes: models, lang: "*", status: http.StatusCreated,
			body: `{"id":"pm1"}`, want: "{\n  \"id\": \"pm1\"\n}\n",
		},
		{
			args: []string{"price-models", "update", "pm 1", "--content-language", "en"},
			file: `{"name":"Per kilogram","metadata":{"version":2}}`, method: http.MethodPut,
			path: "/price/acme1/priceModels/pm%201", scopes: models, lang: "en", status: http.StatusNoContent,
		},
		{
			args: []string{"price-models", "delete", "pm1"}, method: http.MethodDelete, path: "/price/acme1/priceModels/pm1",
			scopes: models, confirmBy: "pm1", status: http.StatusNoContent,
		},
		{
			// A model that prices use is deleted, with its prices, only with --force and the admin scope.
			args: []string{"price-models", "delete", "pm1", "--force"}, method: http.MethodDelete,
			path: "/price/acme1/priceModels/pm1", query: "forceDelete=true", scopes: []string{priceModelManage, priceModelAdmin},
			confirmBy: "pm1", many: "the price model and every price that uses it", status: http.StatusNoContent,
		},
		{
			args: []string{"price-lists", "create"}, file: priceListJSON, method: http.MethodPost,
			path: "/price/acme1/price-lists", scopes: lists, lang: "*", status: http.StatusCreated,
			body: `{"id":"pl1"}`, want: "{\n  \"id\": \"pl1\"\n}\n",
		},
		{
			// An upsert of an unknown price list creates it.
			args: []string{"price-lists", "update", "pl 1"}, file: priceListJSON, method: http.MethodPut,
			path: "/price/acme1/price-lists/pl%201", scopes: lists, lang: "*", status: http.StatusCreated,
			body: `{"id":"pl 1"}`, want: "{\n  \"id\": \"pl 1\"\n}\n",
		},
		{
			args: []string{"price-lists", "delete", "pl1"}, method: http.MethodDelete, path: "/price/acme1/price-lists/pl1",
			scopes: lists, confirmBy: "pl1", many: "the price list and every price in it", status: http.StatusNoContent,
		},
		{
			args: []string{"price-lists", "prices", "create", "pl1"}, file: listPriceJSON, method: http.MethodPost,
			path: prices, scopes: lists, status: http.StatusCreated, body: `{"id":"p1"}`,
			want: "{\n  \"id\": \"p1\"\n}\n",
		},
		{
			args: []string{"price-lists", "prices", "update", "pl1", "p 1"}, file: listPriceJSON,
			method: http.MethodPut, path: prices + "/p%201", scopes: lists, status: http.StatusNoContent,
		},
		{
			args: []string{"price-lists", "prices", "delete", "pl1", "p1"}, method: http.MethodDelete,
			path: prices + "/p1", scopes: lists, confirmBy: "p1", status: http.StatusNoContent,
		},
		{
			args: []string{"price-lists", "prices", "bulk-create", "pl1"}, file: "[" + listPriceJSON + "]",
			method: http.MethodPost, path: prices + "/bulk", scopes: lists, status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"p1","code":201,"status":"CREATED"}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"id\": \"p1\",\n    \"index\": 0,\n    \"status\": \"CREATED\"\n  }\n]\n",
		},
		{
			args: []string{"price-lists", "prices", "bulk-update", "pl1"}, file: `[{"id":"p1","tierValues":[]},{"id":"p2"}]`,
			method: http.MethodPut, path: prices + "/bulk", scopes: lists, status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"p1","code":204},{"index":1,"id":"p2","code":201}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"p1\",\n    \"index\": 0\n  },\n" +
				"  {\n    \"code\": 201,\n    \"id\": \"p2\",\n    \"index\": 1\n  }\n]\n",
		},
		{
			// Emporix answers a bulk delete with 204 No Content, not item by item.
			args: []string{"price-lists", "prices", "bulk-delete", "pl1"}, file: `["p1","p2"]`,
			method: http.MethodDelete, path: prices + "/bulk", scopes: lists, confirmBy: "pl1",
			many: "the prices of a price list that --file names", status: http.StatusNoContent,
		},
	}
}

// feeAndItemFeeWrites are the writes of fees and of item fees, each with the
// scope of its own operation.
func feeAndItemFeeWrites() []writeCase {
	const (
		feeJSON = `{"code":"deposit","name":{"en":"Deposit"},"feeType":"ABSOLUTE","siteCode":"AcmeStore_NL",` +
			`"active":true}`
		itemFees    = "/fee/acme1/itemFees/urn:yaas:saasag:caasproduct:product:acme1%3B100200/fees"
		productFees = "/fee/acme1/productFees/100200/fees"
		itemFeeBody = `{"id":"if1","yrn":"urn:yaas:saasag:fee:id:acme1;if1"}`
		itemFeeJSON = "{\n  \"id\": \"if1\",\n  \"yrn\": \"urn:yaas:saasag:fee:id:acme1;if1\"\n}\n"
	)
	update, remove := []string{itemFeeUpdate}, []string{itemFeeDelete}
	return []writeCase{
		{
			args: []string{"fees", "create"}, file: feeJSON, method: http.MethodPost, path: "/fee/acme1/fees",
			scopes: []string{feeCreate}, lang: "*", status: http.StatusCreated,
			body: `{"id":"f1","yrn":"urn:yaas:saasag:fee:id:acme1;f1"}`,
			want: "{\n  \"id\": \"f1\",\n  \"yrn\": \"urn:yaas:saasag:fee:id:acme1;f1\"\n}\n",
		},
		{
			args: []string{"fees", "update", "f1"}, file: feeJSON, method: http.MethodPut, path: "/fee/acme1/fees/f1",
			query: "partial=false", scopes: []string{feeUpdate}, lang: "*", status: http.StatusOK, body: `{"id":"f1"}`,
			want: "{\n  \"id\": \"f1\"\n}\n", conflict: feeUpdateConflict,
		},
		{
			args:   []string{"fees", "update", "f1", "--partial"},
			file:   `{"code":"deposit","siteCode":"main","active":true,"feeType":"ABSOLUTE","name":{"en":"Deposit"}}`,
			method: http.MethodPut, path: "/fee/acme1/fees/f1", query: "partial=true", scopes: []string{feeUpdate},
			lang: "*", status: http.StatusOK, conflict: feeUpdateConflict,
		},
		{
			args: []string{"fees", "delete", "f1"}, method: http.MethodDelete, path: "/fee/acme1/fees/f1",
			scopes: []string{feeDelete}, confirmBy: "f1", status: http.StatusNoContent,
		},
		{
			args: []string{"fees", "assignments", "create", "--item", productYRN, "--site", "AcmeStore_NL", "--fee", "f1",
				"--fee", "f2"},
			sent:   `{"feeIds":["f1","f2"],"itemYrn":"` + productYRN + `","siteCode":"AcmeStore_NL"}`,
			method: http.MethodPost, path: "/fee/acme1/itemFees", scopes: []string{itemFeeCreate},
			status: http.StatusCreated, body: itemFeeBody, want: itemFeeJSON,
		},
		{
			args: []string{"fees", "assignments", "create"}, file: `{"itemYrn":"` + paymentYRN + `","siteCode":"main",` +
				`"feeIds":["f1"]}`,
			method: http.MethodPost, path: "/fee/acme1/itemFees", scopes: []string{itemFeeCreate},
			status: http.StatusCreated, body: itemFeeBody, want: itemFeeJSON,
		},
		{
			args: []string{"fees", "assignments", "update", productYRN, "--fee", "f1"}, sent: `{"feeIds":["f1"]}`,
			method: http.MethodPut, path: itemFees, query: "partial=false", scopes: update, status: http.StatusOK,
			body: `{"feeIds":["f1"]}`, want: "{\n  \"feeIds\": [\n    \"f1\"\n  ]\n}\n",
		},
		{
			args: []string{"fees", "assignments", "update", paymentYRN, "--partial"}, file: `{"feeIds":["f2"]}`,
			method: http.MethodPut, path: "/fee/acme1/itemFees/urn:yaas:saasag:paymentmode:code:acme1%3Binvoice/fees",
			query: "partial=true", scopes: update, status: http.StatusOK,
		},
		{
			args: []string{"fees", "assignments", "update-product", "100200", "--site", "AcmeStore_NL", "--fee", "f1",
				"--fee", "f2"},
			sent: `{"feeIds":["f1","f2"]}`, method: http.MethodPut, path: productFees,
			query: "partial=false&siteCode=AcmeStore_NL", scopes: update, status: http.StatusOK,
		},
		{
			args: []string{"fees", "assignments", "update-product", "100200", "--site", "AcmeStore_NL", "--partial"},
			file: `{"feeIds":["f2"]}`, method: http.MethodPut, path: productFees,
			query: "partial=true&siteCode=AcmeStore_NL", scopes: update, status: http.StatusOK,
		},
		{
			// Without --site, Emporix acts on the item fee on site main; --site names another.
			args: []string{"fees", "assignments", "update", productYRN, "--fee", "f1", "--site", "AcmeStore_DE"},
			sent: `{"feeIds":["f1"]}`, method: http.MethodPut, path: itemFees, query: "partial=false&siteCode=AcmeStore_DE",
			scopes: update, status: http.StatusOK,
		},
		{
			args: []string{"fees", "assignments", "delete", productYRN}, method: http.MethodDelete, path: itemFees,
			scopes: remove, confirmBy: productYRN, many: "every fee of the item YRN", status: http.StatusNoContent,
		},
		{
			args:   []string{"fees", "assignments", "delete", productYRN, "--site", "AcmeStore_DE"},
			method: http.MethodDelete, path: itemFees, query: "siteCode=AcmeStore_DE", scopes: remove,
			confirmBy: productYRN, many: "every fee of the item YRN", status: http.StatusNoContent,
		},
		{
			args: []string{"fees", "assignments", "delete-product", "100200"}, method: http.MethodDelete,
			path: productFees, scopes: remove, confirmBy: "100200", many: "every fee of the product PRODUCT_ID",
			status: http.StatusNoContent,
		},
		{
			args:   []string{"fees", "assignments", "delete-product", "100200", "--site", "AcmeStore_DE"},
			method: http.MethodDelete, path: productFees, query: "siteCode=AcmeStore_DE", scopes: remove,
			confirmBy: "100200", many: "every fee of the product PRODUCT_ID", status: http.StatusNoContent,
		},
		{
			args: []string{"fees", "assignments", "delete-fee", productYRN, "f1"}, method: http.MethodDelete,
			path: itemFees + "/f1", scopes: remove, confirmBy: "f1", status: http.StatusNoContent,
		},
		{
			args: []string{"fees", "assignments", "delete-product-fee", "100200", "f1"}, method: http.MethodDelete,
			path: productFees + "/f1", scopes: remove, confirmBy: "f1", status: http.StatusNoContent,
		},
	}
}

// TestPriceListAndFeeWritesRefuseInvalidInput checks that the writes of
// price lists, their prices and item fees refuse, as usage errors (exit 2)
// that send nothing, also for a dry run: an ID that addresses the bulk
// operation of a price list's prices, missing arguments, a bulk body that is
// no array of one to 200 items (also that of bulk-delete, which Emporix
// answers as a whole), a body that a delete does not take, and item fees
// without the flags or the file that give their body, or with both.
func TestPriceListAndFeeWritesRefuseInvalidInput(t *testing.T) {
	const bulk = "DELETE /price/{tenant}/price-lists/pl1/prices/bulk: a bulk write sends "
	ids := strings.Repeat(`"p",`, 200) + `"p"`
	for _, tt := range []struct {
		args []string
		file string
		want string
	}{
		{args: []string{"price-lists", "prices", "update", "pl1", "bulk"}, file: "[" + listPriceJSON + "]",
			want: `invalid ID "bulk": it would be sent as bulk-update, not to one item`},
		{args: []string{"price-lists", "prices", "delete", "pl1", "bulk"},
			want: `invalid ID "bulk": it would be sent as bulk-delete, not to one item`},
		{args: []string{"price-lists", "prices", "create"}, file: listPriceJSON, want: "accepts 1 arg(s), received 0"},
		{args: []string{"price-lists", "prices", "bulk-create", "pl1"}, file: listPriceJSON,
			want: "POST /price/{tenant}/price-lists/pl1/prices/bulk: a bulk write sends a JSON array of one or more items"},
		{args: []string{"price-lists", "prices", "bulk-delete", "pl1"}, want: `required flag(s) "file" not set`},
		{args: []string{"price-lists", "prices", "bulk-delete", "pl1"}, file: `{}`,
			want: bulk + "a JSON array of one or more items"},
		{args: []string{"price-lists", "prices", "bulk-delete", "pl1"}, file: `[]`,
			want: bulk + "a JSON array of one or more items"},
		{args: []string{"price-lists", "prices", "bulk-delete", "pl1"}, file: "[" + ids + "]",
			want: bulk + "at most 200 items in one request, not 201: split them"},
		{args: []string{"price-models", "delete", "pm1", "--force=maybe"}, want: `invalid argument "maybe"`},
		{args: []string{"price-models", "delete", "pm1"}, file: "{}", want: "unknown flag: --file"},
		{args: []string{"fees", "assignments", "create", "--item", productYRN, "--site", "main"},
			want: "at least one of the flags in the group [file fee] is required"},
		{args: []string{"fees", "assignments", "update", productYRN, "--fee", "f1"}, file: `{"feeIds":["f1"]}`,
			want: "if any flags in the group [file fee] are set none of the others can be"},
		{args: []string{"fees", "assignments", "update-product", "100200", "--fee", "f1"},
			want: `required flag(s) "site" not set`},
		{args: []string{"fees", "assignments", "delete-fee", productYRN}, want: "accepts 2 arg(s), received 1"},
		{args: []string{"fees", "assignments", "delete", productYRN, "--site", " "}, want: `invalid site " "`},
		{args: []string{"fees", "assignments", "delete-fee", productYRN, "f1", "--site", "main"},
			want: "unknown flag: --site"},
		{args: []string{"fees", "assignments", "delete-product-fee", "100200", "f1", "--site", "main"},
			want: "unknown flag: --site"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			args := append(slices.Clone(tt.args), flag)
			if tt.file != "" {
				args = append(args, "--file", writeBody(t, tt.file))
			}
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), args...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q: exit %d, stdout %q, stderr %q; want exit 2 and %q", args, code, stdout, stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
}

// TestPriceWritesSayHowEmporixReadsThem checks the help of writes that the
// Price Service reads in its own way: the localized text of a price list,
// a map of translations also with --content-language LANGUAGE, unlike that
// of a price model; bulk-delete, which Emporix answers as a whole, not item
// by item; and price-models delete, which deletes a model that prices use
// only with --force, which requests the admin scope too.
func TestPriceWritesSayHowEmporixReadsThem(t *testing.T) {
	for _, tt := range []struct {
		args          []string
		want, without string
	}{
		{args: []string{"price-lists", "create"}, want: priceListLanguageHelp, without: contentLanguageHelp},
		{args: []string{"price-lists", "update"}, want: priceListLanguageHelp, without: contentLanguageHelp},
		{args: []string{"price-models", "create"}, want: contentLanguageHelp, without: priceListLanguageHelp},
		{args: []string{"price-lists", "prices", "bulk-delete"},
			want:    "Emporix takes up to 200 items in one request (epx refuses more) and answers for\nall of them at once",
			without: "the result of each"},
		{args: []string{"price-models", "delete"},
			want: "\n\n--force also requests the price.pricemodel_manage_admin scope.\n\n" +
				"Requires the price.pricemodel_manage scope.",
			without: "epx api"},
	} {
		_, help, _ := runCLI(t, nil, append(tt.args, "--help")...)
		long, _, _ := strings.Cut(help, "\n\nUsage:")
		if !strings.Contains(long, tt.want) || strings.Contains(long, tt.without) {
			t.Errorf("%q: help %q; want %q, without %q", tt.args, long, tt.want, tt.without)
		}
	}
}

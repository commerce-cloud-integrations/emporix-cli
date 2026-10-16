package cli

import (
	"net/http"
)

// A price as create takes it and an availability record, with the product
// and site that a bulk write names it by.
const (
	priceJSON = `{"itemId":{"itemType":"PRODUCT","id":"100200"},"currency":"EUR","location":{"countryCode":"NL"},` +
		`"priceModelId":"pm1","restrictions":{"siteCodes":["AcmeStore_NL"]},"tierValues":[{"id":"t1","priceValue":15.90}]}`
	stockJSON = `{"stockLevel":10,"available":true,"distributionChannel":"ASSORTMENT"}`
	stockItem = `{"productId":"100200","site":"AcmeStore_NL","stockLevel":10,"available":true}`
)

// priceAndAvailabilityWrites are the writes of prices and availability.
func priceAndAvailabilityWrites() []writeCase {
	scopes := []string{priceManageScope}
	stockScopes := []string{availabilityManageScope}
	return []writeCase{
		{
			args: []string{"prices", "create"}, file: priceJSON, method: http.MethodPost, path: "/price/acme1/prices",
			scopes: scopes, status: http.StatusCreated, body: `{"id":"p1"}`, want: "{\n  \"id\": \"p1\"\n}\n",
		},
		{
			args: []string{"prices", "update", "p1"}, file: `{"id":"p1","metadata":{"version":3}}`, method: http.MethodPut,
			path: "/price/acme1/prices/p1", query: "partial=false", scopes: scopes, status: http.StatusNoContent,
		},
		{
			args: []string{"prices", "update", "p 1", "--partial"}, file: `{"tierValues":[{"id":"t1","priceValue":14.90}]}`,
			method: http.MethodPut, path: "/price/acme1/prices/p%201", query: "partial=true", scopes: scopes,
			status: http.StatusCreated, body: `{"id":"p 1"}`, want: "{\n  \"id\": \"p 1\"\n}\n",
		},
		{
			args: []string{"prices", "bulk-create"}, file: "[" + priceJSON + ",\n" + priceJSON + "]\n",
			method: http.MethodPost, path: "/price/acme1/prices/bulk", scopes: scopes, status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"p1","code":201,"status":"CREATED"},{"index":1,"id":"p2","code":201,"status":"CREATED"}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"id\": \"p1\",\n    \"index\": 0,\n    \"status\": \"CREATED\"\n  },\n" +
				"  {\n    \"code\": 201,\n    \"id\": \"p2\",\n    \"index\": 1,\n    \"status\": \"CREATED\"\n  }\n]\n",
		},
		{
			args: []string{"prices", "bulk-update"}, file: `[{"id":"p1","currency":"EUR"}]`, method: http.MethodPut,
			path: "/price/acme1/prices/bulk", scopes: scopes, status: http.StatusMultiStatus,
			body: `[{"index":0,"id":"p1","code":204,"status":"NO_CONTENT"}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"p1\",\n    \"index\": 0,\n    \"status\": \"NO_CONTENT\"\n  }\n]\n",
		},
		{
			args: []string{"availability", "create", "100200", "--site", "AcmeStore_NL"}, file: stockJSON,
			method: http.MethodPost, path: "/availability/acme1/availability/100200/AcmeStore_NL", scopes: stockScopes,
			status: http.StatusCreated, body: `{"id":"AcmeStore_NL:100200"}`,
			want: "{\n  \"id\": \"AcmeStore_NL:100200\"\n}\n",
		},
		{
			args: []string{"availability", "update", "100200", "--site", "AcmeStore_NL"}, file: stockJSON,
			method: http.MethodPut, path: "/availability/acme1/availability/100200/AcmeStore_NL", scopes: stockScopes,
			status: http.StatusNoContent,
		},
		{
			args: []string{"availability", "bulk-create"}, file: "[" + stockItem + "]", method: http.MethodPost,
			path: "/availability/acme1/availability/bulk", scopes: stockScopes, status: http.StatusMultiStatus,
			body: `[{"id":"AcmeStore_NL:100200","index":0,"code":201,"status":"Created"}]`,
			want: "[\n  {\n    \"code\": 201,\n    \"id\": \"AcmeStore_NL:100200\",\n    \"index\": 0,\n" +
				"    \"status\": \"Created\"\n  }\n]\n",
		},
		{
			args: []string{"availability", "bulk-update"}, file: "[" + stockItem + "," + stockItem + "]",
			method: http.MethodPut, path: "/availability/acme1/availability/bulk", scopes: stockScopes,
			status: http.StatusMultiStatus, body: `[{"id":"AcmeStore_NL:100200","index":0,"code":204},` +
				`{"id":"AcmeStore_NL:100200","index":1,"code":201}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"AcmeStore_NL:100200\",\n    \"index\": 0\n  },\n" +
				"  {\n    \"code\": 201,\n    \"id\": \"AcmeStore_NL:100200\",\n    \"index\": 1\n  }\n]\n",
		},
	}
}

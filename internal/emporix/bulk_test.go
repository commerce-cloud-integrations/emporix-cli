package emporix

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
)

const bulkPath = "/price/{tenant}/prices/bulk"

func TestPrepareChecksTheItemsOfABulkWrite(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	form, formErr := FormBody([]Part{{Name: "items", Content: []byte(`[1]`)}})
	if formErr != nil {
		t.Fatal(formErr)
	}
	for _, body := range []Body{{}, jsonBody(t, `[]`), jsonBody(t, `{"id":"p1"}`), jsonBody(t, `null`),
		jsonBody(t, `"p1"`), {data: []byte(`[1,`), contentType: JSONContentType}, jsonBody(t, `[1,2,3]`), form} {
		_, err := client.Prepare(&Write{Method: http.MethodPut, Path: bulkPath, Body: body, Parts: Items(2, "")})
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("body %q: error %v; want ErrInvalidInput", body.data, err)
		}
	}
	_, err := client.Prepare(&Write{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[1,2,3]`),
		Parts: Items(2, "")})
	if want := "PUT /price/{tenant}/prices/bulk: a bulk write sends at most 2 items in one request, not 3: split " +
		"them"; err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Errorf("error %v; want it to end with %q", err, want)
	}
	for body, items := range map[string]int{`[1]`: 1, `[1,2]`: 2} {
		if req, err := client.Prepare(&Write{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, body),
			Parts: Items(2, "")}); err != nil || len(req.parts.parts) != items {
			t.Errorf("body %s: request %+v, error %v; want it prepared", body, req, err)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("token %d, requests %d; want none", fake.tokenCalls, len(fake.requests))
	}
}

func TestSendReadsTheResultOfEachItem(t *testing.T) {
	for _, tt := range []struct {
		results string
		failed  []ItemFailure
		class   error
	}{
		{results: `[{"index":0,"id":"p1","code":201},{"index":1,"id":"p2","code":204}]`},
		{results: `[{"index":2,"id":"p2","code":204,"status":"OK"},{"index":1,"id":"p1","code":204,"status":"OK"}]`},
		{
			results: `[{"index":1,"id":"p1","code":201,"status":"Created"},` +
				`{"index":2,"code":400,"status":"Bad Request","message":"Validation problem with request body."}]`,
			failed: []ItemFailure{
				{Index: 1, Status: http.StatusBadRequest, Message: "Validation problem with request body."},
			},
		},
		{
			results: `[{"index":0,"id":"p1","code":409,"status":"CONFLICT","message":"changed"},{"id":"p2","code":409}]`,
			failed: []ItemFailure{
				{Index: 0, ID: "p1", Status: http.StatusConflict, Message: "changed"},
				{Index: 1, ID: "p2", Status: http.StatusConflict},
			},
			class: ErrConflict,
		},
		{
			results: `[{"index":0,"code":201},{"code":403,"message":"not your price"}]`,
			failed:  []ItemFailure{{Index: 1, Status: http.StatusForbidden, Message: "not your price"}},
			class:   ErrAuth,
		},
		{
			results: `[{"index":1,"id":"p2","code":409},{"index":0,"id":"p1","code":204}]`,
			failed:  []ItemFailure{{Index: 1, ID: "p2", Status: http.StatusConflict}},
			class:   ErrConflict,
		},
		{
			results: `[{"index":0,"id":"p1","code":500,"message":"boom"},{"index":1,"code":503}]`,
			failed: []ItemFailure{
				{Index: 0, ID: "p1", Status: http.StatusInternalServerError, Message: "boom"},
				{Index: 1, Status: http.StatusServiceUnavailable},
			},
		},
		{
			results: `[{"code":404},{"code":400,"details":["unknown site"]}]`,
			failed: []ItemFailure{{Index: 0, Status: http.StatusNotFound},
				{Index: 1, Status: http.StatusBadRequest, Message: "unknown site"}},
		},
		{
			results: `[{"index":0,"id":"1","status":201},{"index":1,"status":400,"errorMessage":"no price"}]`,
			failed:  []ItemFailure{{Index: 1, Status: http.StatusBadRequest, Message: "no price"}},
		},
		{results: `[{"index":0,"code":204,"status":"No Content"},{"index":1,"status":201,"code":0}]`},
	} {
		fake := newFakeEmporix(t)
		fake.route("/price/acme1/prices/bulk", respond(http.StatusMultiStatus, tt.results))
		body, err := send(fake.client(t), &Write{
			Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[{"id":"p1"},{"id":"p2"}]`), Parts: Items(200, ""),
		})
		if string(body) != tt.results {
			t.Errorf("%s: body %s", tt.results, body)
		}
		var items *ItemsError
		switch {
		case tt.failed == nil && err != nil:
			t.Errorf("%s: error %v; want none", tt.results, err)
		case tt.failed == nil:
		case !errors.As(err, &items) || !slices.Equal(items.Failed, tt.failed) || items.Items != 2 ||
			!errors.Is(errors.Unwrap(err), tt.class):
			t.Errorf("%s: error %#v; want failures %v of class %v", tt.results, err, tt.failed, tt.class)
		case err.Error() != "PUT /price/acme1/prices/bulk: "+items.summary():
			t.Errorf("%s: error %q", tt.results, err)
		}
	}
}

func (e *ItemsError) summary() string {
	return strings.TrimPrefix(e.Error(), e.Method+" "+e.Path+": ")
}

func TestSendWithoutTheResultOfEachItem(t *testing.T) {
	for _, tt := range []struct {
		status  int
		results string
	}{
		{http.StatusMultiStatus, `[{"code":201}]`}, {http.StatusOK, `[{"code":201},{"code":0}]`},
		{http.StatusCreated, `{"id":"p1"}`}, {http.StatusNoContent, ""}, {http.StatusAccepted, `[{"code":201},{"code":600}]`},
		{http.StatusMultiStatus, `[{"index":0,"code":201},{"index":0,"code":409}]`},
		{http.StatusMultiStatus, `[{"code":201},{"index":0,"code":201}]`},
		{http.StatusMultiStatus, `[{"index":1,"code":201},{"index":1,"code":201}]`},
		{http.StatusMultiStatus, `[{"index":2,"code":201},{"index":3,"code":201}]`},
		{http.StatusMultiStatus, `[{"index":1,"code":201},{"code":201}]`},
		{http.StatusMultiStatus, `[{"index":-1,"code":201},{"index":1,"code":201}]`},
		{http.StatusMultiStatus, `[{"index":0,"code":201},{"index":"1","code":201}]`},
		{http.StatusOK, `[{"index":0,"status":201},{"index":1,"status":"201"}]`},
		{http.StatusOK, `[{"index":0,"status":201},{"index":1,"status":600}]`},
	} {
		status, results := tt.status, tt.results
		fake := newFakeEmporix(t)
		fake.route("/price/acme1/prices/bulk", respond(status, results))
		body, err := send(fake.client(t), &Write{
			Method: http.MethodPost, Path: bulkPath, Body: jsonBody(t, `[{"id":"p1"},{"id":"p2"}]`), Parts: Items(200, ""),
		})
		want := "POST /price/acme1/prices/bulk: the response does not give the result of each of the 2 items"
		if err == nil || !strings.HasPrefix(err.Error(), want) || errors.Unwrap(err) != nil ||
			string(body) != results || (results == "" && body != nil) {
			t.Errorf("%d %s: body %q, error %v", status, results, body, err)
		}
	}
}

func TestSendMatchesResultsByTheirItemKey(t *testing.T) {
	const sent = `[{"eventType":"order.created"},{"eventType":"cart.closed"},{"eventType":"cart.closed"},` +
		`{"action":"SUBSCRIBE"}]`
	for _, tt := range []struct {
		results string
		failed  []ItemFailure
		ok      bool
	}{
		{
			results: `[{"code":204},{"eventType":"cart.closed","code":200},{"eventType":"cart.closed","code":200},` +
				`{"eventType":"order.created","code":200}]`,
			ok: true,
		},
		{
			results: `[{"eventType":"cart.closed","code":409,"message":"stale"},{"code":400},` +
				`{"eventType":"order.created","code":200},{"eventType":"cart.closed","code":200}]`,
			failed: []ItemFailure{
				{Index: 1, ID: "cart.closed", Status: http.StatusConflict, Message: "stale"},
				{Index: 3, Status: http.StatusBadRequest},
			},
		},
		{results: `[{"eventType":"order.created","code":200},{"eventType":"cart.closed","code":200},` +
			`{"eventType":"order.created","code":200},{"code":200}]`},
		{results: `[{"eventType":"order.created","code":200},{"eventType":"cart.closed","code":200},` +
			`{"eventType":"cart.closed","code":200},{"eventType":"product.updated","code":200}]`},
	} {
		fake := newFakeEmporix(t)
		fake.route("/webhook/acme1/event-subscriptions", respond(http.StatusMultiStatus, tt.results))
		_, err := send(fake.client(t), &Write{Method: http.MethodPatch, Path: "/webhook/{tenant}/event-subscriptions",
			Body: jsonBody(t, sent), Parts: Items(200, "eventType")})
		var items *ItemsError
		switch {
		case tt.ok && err != nil:
			t.Errorf("%s: error %v; want none", tt.results, err)
		case tt.ok:
		case tt.failed == nil && (err == nil ||
			!strings.Contains(err.Error(), "the response does not give the result of each of the 4 items")):
			t.Errorf("%s: error %v; want the results unknown", tt.results, err)
		case tt.failed != nil && (!errors.As(err, &items) || !slices.Equal(items.Failed, tt.failed)):
			t.Errorf("%s: error %#v; want failures %v", tt.results, err, tt.failed)
		}
	}
}

func TestItemFailureNamesTheItem(t *testing.T) {
	const carriedOut = " (" + carriedOutNote + ")"
	for want, failure := range map[string]ItemFailure{
		"item 0 (p1): 409 Conflict: changed": {Index: 0, ID: "p1", Status: http.StatusConflict, Message: "changed"},
		"item 3: 400 Bad Request":            {Index: 3, Status: http.StatusBadRequest},
		"item 1 (AcmeStore_NL:100200): 499":  {Index: 1, ID: "AcmeStore_NL:100200", Status: 499},
		"item 2 (p3): 500 Internal Server Error: boom" + carriedOut: {
			Index: 2, ID: "p3", Status: http.StatusInternalServerError, Message: "boom",
		},
		"item 0: 502 Bad Gateway" + carriedOut:         {Status: http.StatusBadGateway},
		"item 0: 503 Service Unavailable" + carriedOut: {Status: http.StatusServiceUnavailable},
		"item 0: 504 Gateway Timeout" + carriedOut:     {Status: http.StatusGatewayTimeout},
		"item 0: 501 Not Implemented":                  {Status: http.StatusNotImplemented},
	} {
		if got := failure.String(); got != want {
			t.Errorf("%#v: %q, want %q", failure, got, want)
		}
	}
}

func TestBulkWriteAnsweredAsAWhole(t *testing.T) {
	const path = "/price/{tenant}/price-lists/pl1/prices/bulk"
	fake := newFakeEmporix(t)
	client := fake.client(t)
	for _, tt := range []struct{ body, want string }{
		{`{}`, "a bulk write sends a JSON array of one or more items"},
		{`[]`, "a bulk write sends a JSON array of one or more items"},
		{`["p1","p2","p3"]`, "a bulk write sends at most 2 items in one request, not 3: split them"},
	} {
		_, err := client.Prepare(&Write{Method: http.MethodDelete, Path: path, Body: jsonBody(t, tt.body),
			Parts: Items(2, ""), AnsweredWhole: true})
		if !errors.Is(err, ErrInvalidInput) || !strings.HasSuffix(err.Error(), tt.want) {
			t.Errorf("body %s: error %v; want ErrInvalidInput ending with %q", tt.body, err, tt.want)
		}
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("token %d, requests %d; want none", fake.tokenCalls, len(fake.requests))
	}
	fake.route("/price/acme1/price-lists/pl1/prices/bulk", respond(http.StatusNoContent, ""))
	body, err := send(client, &Write{Method: http.MethodDelete, Path: path, Body: jsonBody(t, `["p1","p2"]`),
		Parts: Items(2, ""), AnsweredWhole: true})
	if err != nil || body != nil {
		t.Fatalf("body %q, error %v; want none", body, err)
	}
}

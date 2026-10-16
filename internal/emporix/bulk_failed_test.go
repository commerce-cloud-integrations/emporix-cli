package emporix

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const cartBatchPath = "/cart/{tenant}/carts/c1/itemsBatch"

// TestSendReadsTheItemsOfABatchThatFailed checks a bulk write that Emporix
// answers with a failure status, as the Cart Service answers a batch of
// items that failed in part with 400 or 500: where the body gives the result
// of each item, Send returns it with an *ItemsError listing the items that
// failed (the others were carried out), as for a 207; otherwise, or where no
// item failed, the failure, classified as any write's.
func TestSendReadsTheItemsOfABatchThatFailed(t *testing.T) {
	const mixed = `[{"index":0,"status":201,"id":"1"},{"index":1,"status":400,"errorMessage":"no price for 100200"}]`
	for _, tt := range []struct {
		status   int
		response string
		body     bool          // the response is returned
		failed   []ItemFailure // the items that failed; the failure of the batch when nil
		class    error
		err      string
	}{
		{
			status: http.StatusBadRequest, response: mixed, body: true,
			failed: []ItemFailure{{Index: 1, Status: http.StatusBadRequest, Message: "no price for 100200"}},
			err:    "1 of 2 items failed",
		},
		{
			status: http.StatusInternalServerError, body: true,
			response: `[{"index":0,"status":500,"errorMessage":"boom"},{"index":1,"status":404}]`,
			failed: []ItemFailure{
				{Index: 0, Status: http.StatusInternalServerError, Message: "boom"}, {Index: 1, Status: http.StatusNotFound},
			},
			err: "2 of 2 items failed",
		},
		{
			status: http.StatusNotFound, response: `{"code":404,"message":"Cart not found"}`, class: ErrNotFound,
			err: "404 Not Found: Cart not found",
		},
		{status: http.StatusBadRequest, response: `[{"index":0,"status":201}]`, err: "400 Bad Request"},
		{
			status: http.StatusInternalServerError, response: `[{"index":0,"status":201},{"index":1,"status":201}]`,
			body: true, err: "500 Internal Server Error (" + carriedOutNote + ")",
		},
	} {
		fake := newFakeEmporix(t)
		fake.route("/cart/acme1/carts/c1/itemsBatch", respond(tt.status, tt.response))
		body, err := send(fake.client(t), &Write{
			Method: http.MethodPost, Path: cartBatchPath, Body: jsonBody(t, `[{"quantity":1},{"quantity":2}]`),
			Parts: Items(200, ""),
		})
		var items *ItemsError
		isItems := errors.As(err, &items)
		switch {
		case (body != nil) != tt.body || tt.body && string(body) != tt.response:
			t.Errorf("%d %s: body %s", tt.status, tt.response, body)
		case err == nil || err.Error() != "POST /cart/acme1/carts/c1/itemsBatch: "+tt.err:
			t.Errorf("%d %s: error %v; want %q", tt.status, tt.response, err, tt.err)
		case isItems != (tt.failed != nil) || isItems && !reflect.DeepEqual(items.Failed, tt.failed):
			t.Errorf("%d %s: error %#v; want failures %v", tt.status, tt.response, err, tt.failed)
		case !isItems && !errors.Is(classOf(err), tt.class):
			t.Errorf("%d %s: error %v; want class %v", tt.status, tt.response, err, tt.class)
		}
	}
}

// classOf returns the class of err, nil for none.
func classOf(err error) error {
	for _, class := range []error{ErrInvalidInput, ErrAuth, ErrNotFound, ErrConflict} {
		if errors.Is(err, class) {
			return class
		}
	}
	return nil
}

// TestItemsErrorNamesItsParts checks the lines and summary of an
// *ItemsError: its parts are items unless it names them, the items of a
// command each on a line of their own, which count as one failed command.
func TestItemsErrorNamesItsParts(t *testing.T) {
	item := ItemFailure{Index: 1, ID: "2", Status: http.StatusBadRequest, Message: "Invalid quantity"}
	failed := &ItemsError{Method: http.MethodPost, Path: "/x", Noun: CommandNoun, Items: 3, NotRun: 1}
	failed.Failed = []ItemFailure{
		{Index: 0, Type: "UpdateCartItemsBatch", Status: http.StatusMultiStatus, Item: &item},
		{Index: 0, Type: "UpdateCartItemsBatch", Status: http.StatusMultiStatus, Item: &ItemFailure{Status: 409}},
		{Index: 1, Type: "GetCart", Status: http.StatusNotFound, Message: "gone"},
	}
	want := []string{
		"command 0 (UpdateCartItemsBatch), item 1 (2): 400 Bad Request: Invalid quantity",
		"command 0 (UpdateCartItemsBatch), item 0: 409 Conflict",
		"command 1 (GetCart): 404 Not Found: gone",
	}
	if got := failed.Lines(); !reflect.DeepEqual(got, want) {
		t.Errorf("lines %q; want %q", got, want)
	}
	if got := failed.Error(); got != "POST /x: 2 of 3 commands failed, and the last 1 did not run" {
		t.Errorf("error %q", got)
	}
	bulk := &ItemsError{Method: http.MethodPut, Path: "/y", Items: 2, Failed: []ItemFailure{item}}
	if got := strings.Join(bulk.Lines(), "\n") + "\n" + bulk.Error(); got != "item 1 (2): 400 Bad Request: "+
		"Invalid quantity\nPUT /y: 1 of 2 items failed" {
		t.Errorf("bulk %q", got)
	}
}

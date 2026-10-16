package emporix

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

const chainPath = "/cart/{tenant}/carts/c1/execute"

// chainBody is a chain of three commands.
const chainBody = `{"commands":[{"type":"AddCartItem","data":{"quantity":1}},{"type":"UpdateCartItem"},` +
	`{"type":"GetCart"}]}`

// batchChainBody is a chain of a batch of two items to add, a batch of one
// item to update and a read.
const batchChainBody = `{"commands":[{"type":"AddCartItemsBatch","data":[{"quantity":1},{"quantity":2}]},` +
	`{"type":"UpdateCartItemsBatch","data":[{"id":"1","quantity":0}]},{"type":"GetCart"}]}`

// TestPrepareChecksTheCommandsOfAChain checks that a chain's body must be a
// JSON object whose commands, named exactly, are an array of one to as many
// commands as its Parts take: anything else is ErrInvalidInput, found before
// a token is requested or anything is sent.
func TestPrepareChecksTheCommandsOfAChain(t *testing.T) {
	fake := newFakeEmporix(t)
	client := fake.client(t)
	for _, body := range []Body{{}, jsonBody(t, `{}`), jsonBody(t, `{"commands":[]}`), jsonBody(t, `{"commands":null}`),
		jsonBody(t, `{"Commands":[1]}`), jsonBody(t, `[1]`), jsonBody(t, `{"commands":{"type":"GetCart"}}`),
		jsonBody(t, `{"commands":[1,2,3]}`)} {
		_, err := client.Prepare(&Write{Method: http.MethodPost, Path: chainPath, Body: body, Parts: Commands(2)})
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("body %q: error %v; want ErrInvalidInput", body.data, err)
		}
	}
	_, err := client.Prepare(&Write{Method: http.MethodPost, Path: chainPath, Body: jsonBody(t, chainBody),
		Parts: Commands(2)})
	if want := "POST /cart/{tenant}/carts/c1/execute: a chain sends at most 2 commands in one request, not 3: " +
		"split them"; err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Errorf("error %v; want it to end with %q", err, want)
	}
	_, err = client.Prepare(&Write{Method: http.MethodPost, Path: chainPath, Body: jsonBody(t, `{"commands":[]}`),
		Parts: Commands(2)})
	if want := "a chain sends a JSON object whose commands is an array of one or more commands"; err == nil ||
		!strings.HasSuffix(err.Error(), want) {
		t.Errorf("error %v; want it to end with %q", err, want)
	}
	if req, err := client.Prepare(&Write{Method: http.MethodPost, Path: chainPath, Body: jsonBody(t, chainBody),
		Parts: Commands(10)}); err != nil || len(req.parts.parts) != 3 {
		t.Errorf("request %+v, error %v; want it prepared with 3 commands", req, err)
	}
	if fake.tokenCalls != 0 || len(fake.requests) != 0 {
		t.Fatalf("token %d, requests %d; want none", fake.tokenCalls, len(fake.requests))
	}
}

// chainCase is a chain sent with onError (none when empty) and answered with
// results: the commands that failed, how many did not run, and the class
// that their failures share, if any.
type chainCase struct {
	body    string
	onError string
	results string
	failed  []ItemFailure
	notRun  int
	class   error
	summary string
}

// sendChain sends the chain of tt to a fake that answers 207 with its
// results.
func (tt *chainCase) send(t *testing.T) ([]byte, error) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route("/cart/acme1/carts/c1/execute", respond(http.StatusMultiStatus, tt.results))
	query := url.Values{}
	if tt.onError != "" {
		query.Set(onErrorParam, tt.onError)
	}
	return send(fake.client(t), &Write{
		Method: http.MethodPost, Path: chainPath, Query: query, Body: jsonBody(t, tt.body), Parts: Commands(10),
	})
}

// TestSendReadsTheResultOfEachCommand checks a chain: a response whose
// commands all ran and succeeded is returned as sent; one with failures is
// returned with an *ItemsError listing the commands that failed by position
// and type, with the message of their data, how many did not run, and the
// class that they share, if any.
func TestSendReadsTheResultOfEachCommand(t *testing.T) {
	for _, tt := range []chainCase{
		{body: chainBody, results: `{"results":[{"index":0,"type":"AddCartItem","code":201,"data":{"itemId":"3"}},` +
			`{"index":1,"type":"UpdateCartItem","code":204},{"index":2,"type":"GetCart","code":200,"data":{"id":"c1"}}]}`},
		{
			// onError=fail: the chain stops after the first that failed.
			body: chainBody,
			results: `{"results":[{"index":0,"type":"AddCartItem","code":201},{"index":1,"type":"UpdateCartItem",` +
				`"code":404,"status":"Not Found","data":{"code":404,"message":"Cart item not found"}}]}`,
			failed: []ItemFailure{
				{Index: 1, Type: "UpdateCartItem", Status: http.StatusNotFound, Message: "Cart item not found"},
			},
			notRun: 1, class: ErrNotFound, summary: "1 of 3 commands failed, and the last 1 did not run",
		},
		{
			// onError=resume: every command ran.
			body: chainBody, onError: resumeOnError,
			results: `{"results":[{"index":0,"type":"AddCartItem","code":409,"data":{"message":"version"}},` +
				`{"index":1,"type":"UpdateCartItem","code":204},{"index":2,"code":500}]}`,
			failed: []ItemFailure{
				{Index: 0, Type: "AddCartItem", Status: http.StatusConflict, Message: "version"},
				{Index: 2, Status: http.StatusInternalServerError},
			},
			summary: "2 of 3 commands failed",
		},
		{
			// onError=fail, whose last command failed: none was left to run.
			body: chainBody, onError: "fail",
			results: `{"results":[{"index":0,"code":201},{"index":1,"code":204},{"index":2,"type":"GetCart","code":404}]}`,
			failed:  []ItemFailure{{Index: 2, Type: "GetCart", Status: http.StatusNotFound}},
			class:   ErrNotFound, summary: "1 of 3 commands failed",
		},
	} {
		tt.assert(t)
	}
}

// TestSendReadsTheItemsOfTheBatchesOfAChain checks a chain with commands
// that send items: their data gives the result of each item, as a bulk
// write's response does, also with a code that succeeded (207), and each
// item that failed is listed under its command, which counts once.
func TestSendReadsTheItemsOfTheBatchesOfAChain(t *testing.T) {
	addFailed := &ItemFailure{Index: 1, Status: http.StatusBadRequest, Message: "no price for 100200"}
	updateFailed := &ItemFailure{Index: 0, ID: "1", Status: http.StatusBadRequest, Message: "Invalid quantity"}
	for _, tt := range []chainCase{
		{
			// UpdateCartItemsBatch answers 207 even when every item failed, so the chain goes on.
			body: batchChainBody,
			results: `{"results":[{"index":0,"type":"AddCartItemsBatch","code":200,"data":[{"index":0,"status":201},` +
				`{"index":1,"status":201}]},{"index":1,"type":"UpdateCartItemsBatch","code":207,"data":[{"index":0,` +
				`"id":"1","code":400,"message":"Invalid quantity"}]},{"index":2,"type":"GetCart","code":200}]}`,
			failed: []ItemFailure{
				{Index: 1, Type: "UpdateCartItemsBatch", Status: http.StatusMultiStatus, Item: updateFailed},
			},
			summary: "1 of 3 commands failed",
		},
		{
			// AddCartItemsBatch answers a batch that failed in part with 400, which stops the chain.
			body: batchChainBody,
			results: `{"results":[{"index":0,"type":"AddCartItemsBatch","code":400,"data":[{"index":0,"status":201},` +
				`{"index":1,"status":400,"errorMessage":"no price for 100200"}]}]}`,
			failed: []ItemFailure{
				{Index: 0, Type: "AddCartItemsBatch", Status: http.StatusBadRequest, Item: addFailed},
			},
			notRun: 2, summary: "1 of 3 commands failed, and the last 2 did not run",
		},
		{
			// A batch refused whole: its data is the error.
			body: batchChainBody, onError: resumeOnError,
			results: `{"results":[{"index":0,"type":"AddCartItemsBatch","code":400,"data":{"message":"too many"}},` +
				`{"index":1,"type":"UpdateCartItemsBatch","code":500,"data":[{"index":0,"code":204}]},` +
				`{"index":2,"type":"GetCart","code":200}]}`,
			failed: []ItemFailure{
				{Index: 0, Type: "AddCartItemsBatch", Status: http.StatusBadRequest, Message: "too many"},
				{Index: 1, Type: "UpdateCartItemsBatch", Status: http.StatusInternalServerError},
			},
			summary: "2 of 3 commands failed",
		},
	} {
		tt.assert(t)
	}
}

// assert sends tt and checks that the response is returned as sent with the
// failures of tt, if any.
func (tt *chainCase) assert(t *testing.T) {
	t.Helper()
	body, err := tt.send(t)
	if string(body) != tt.results {
		t.Errorf("%s: body %s", tt.results, body)
	}
	var items *ItemsError
	switch {
	case tt.failed == nil && err != nil:
		t.Errorf("%s: error %v; want none", tt.results, err)
	case tt.failed == nil:
	case !errors.As(err, &items) || !reflect.DeepEqual(items.Failed, tt.failed) || items.Items != 3 ||
		items.NotRun != tt.notRun || !errors.Is(classOf(err), tt.class):
		t.Errorf("%s: error %#v; want failures %v of class %v", tt.results, err, tt.failed, tt.class)
	case items.summary() != tt.summary:
		t.Errorf("%s: error %q; want %q", tt.results, err, tt.summary)
	}
}

// TestSendWithoutTheResultOfEachCommand checks that a chain whose response
// does not give the result of each command that ran, in order, fails,
// unclassified, with the response, if any: which commands ran is unknown.
// Without onError=resume, the results stop only after a command that failed;
// with it, they give each command.
func TestSendWithoutTheResultOfEachCommand(t *testing.T) {
	for _, tt := range []chainCase{
		{body: chainBody, results: `{"results":[{"index":0,"code":201},{"index":1,"code":204}]}`}, // none failed
		{body: chainBody, onError: resumeOnError, results: `{"results":[{"index":0,"code":201},{"index":1,"code":404}]}`},
		{body: chainBody, results: `{"results":[{"index":0,"code":404},{"index":1,"code":200}]}`}, // ran after one failed
		{body: chainBody, results: `{"results":[]}`}, {body: chainBody, results: `{"results":null}`},
		{body: chainBody, results: `[{"index":0,"code":201}]`}, {body: chainBody, results: ""},
		{body: chainBody, results: `{"results":[{"index":0,"code":201},{"index":1,"code":201},{"index":2,"code":201},` +
			`{"index":3,"code":201}]}`},
		{body: chainBody, results: `{"results":[{"index":1,"code":201},{"index":0,"code":201},{"index":2,"code":201}]}`},
		{body: chainBody, results: `{"results":[{"code":201},{"index":1,"code":201},{"index":2,"code":201}]}`},
		{body: chainBody, results: `{"results":[{"index":0,"code":201},{"index":1,"code":42},{"index":2,"code":201}]}`},
		{body: chainBody, results: `{"results":[{"index":0,"code":201},{"index":1,"code":"404"}]}`},
		// A batch that succeeded without the result of each of its items.
		{body: batchChainBody, results: `{"results":[{"index":0,"code":200},{"index":1,"code":207,"data":[]},` +
			`{"index":2,"code":200}]}`},
		{body: batchChainBody, results: `{"results":[{"index":0,"code":200,"data":[{"index":0,"status":201}]},` +
			`{"index":1,"code":207,"data":[{"index":0,"code":204}]},{"index":2,"code":200}]}`},
	} {
		body, err := tt.send(t)
		want := "POST /cart/acme1/carts/c1/execute: the response does not give the result of each of the 3 commands"
		if err == nil || !strings.HasPrefix(err.Error(), want) || errors.Unwrap(err) != nil ||
			string(body) != tt.results || (tt.results == "" && body != nil) {
			t.Errorf("%s %s: body %q, error %v", tt.onError, tt.results, body, err)
		}
	}
}

// TestChainThatFailsWhole checks that a chain that Emporix refuses whole, as
// it answers one of more than 10 commands, fails as any write does.
func TestChainThatFailsWhole(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/cart/acme1/carts/c1/execute", respond(http.StatusBadRequest, `{"code":400,"message":"too many"}`))
	body, err := send(fake.client(t), &Write{
		Method: http.MethodPost, Path: chainPath, Body: jsonBody(t, chainBody), Parts: Commands(10),
	})
	var api *APIError
	if body != nil || !errors.As(err, &api) || api.Status != http.StatusBadRequest || api.Message != "too many" {
		t.Errorf("body %s, error %v; want a 400 Bad Request", body, err)
	}
}

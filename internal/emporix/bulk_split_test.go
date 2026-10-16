package emporix

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// TestSplitSendsBatchesOfUpToMaxItems checks that a bulk write that may be
// sent in batches is split into writes of up to MaxItems items, in order,
// each item as given, and that any other write, or a body that Prepare
// refuses, is left whole.
func TestSplitSendsBatchesOfUpToMaxItems(t *testing.T) {
	whole := []*Write{
		{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[1,2,3]`), Parts: Items(2, "")},
		{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[1,2]`), Parts: Items(2, ""), Batched: true},
		{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `{"a":[1,2,3]}`), Parts: Items(2, ""), Batched: true},
		{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[1,2,3]`), Batched: true},
		{Method: http.MethodPost, Path: bulkPath, Body: jsonBody(t, `{"commands":[1,2,3]}`), Parts: Commands(2),
			Batched: true},
	}
	for _, w := range whole {
		if split := w.Split(); len(split) != 1 || split[0] != w {
			t.Errorf("%s: split into %d; want it whole", w.Body.data, len(split))
		}
	}
	w := &Write{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, "[ {\"id\": 1},\n2 ,3,4,\"5\" ]"),
		Parts: Items(2, ""), Batched: true, Header: http.Header{"Content-Language": {"en"}}}
	split := w.Split()
	var bodies []string
	for _, batch := range split {
		bodies = append(bodies, string(batch.Body.data))
		if batch.Method != w.Method || batch.Path != w.Path || batch.Header.Get("Content-Language") != "en" ||
			batch.Parts != w.Parts || batch.Body.contentType != JSONContentType {
			t.Errorf("batch %+v; want a copy of the write but for its body", batch)
		}
	}
	if want := []string{`[{"id": 1},2]`, `[3,4]`, `["5"]`}; !slices.Equal(bodies, want) {
		t.Errorf("bodies %q; want %q", bodies, want)
	}
	if string(w.Body.data) != "[ {\"id\": 1},\n2 ,3,4,\"5\" ]" {
		t.Errorf("the write split was changed: %s", w.Body.data)
	}
}

// TestBatchesNameTheirItemsInTheWholeWrite checks that the failures of a
// batch name each item by its position among the items of the write split,
// and its request the items it sends.
func TestBatchesNameTheirItemsInTheWholeWrite(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/price/acme1/prices/bulk",
		respond(http.StatusMultiStatus, `[{"index":0,"code":204},{"id":"p4","code":409}]`))
	client := fake.client(t)
	batches := (&Write{Method: http.MethodPut, Path: bulkPath, Body: jsonBody(t, `[1,2,3,4,5]`),
		Parts: Items(2, ""), Batched: true}).Split()
	req, err := client.Prepare(batches[1])
	if err != nil || req.Items() != 2 {
		t.Fatalf("request %+v, error %v; want one of 2 items", req, err)
	}
	_, err = fake.client(t).Send(t.Context(), req)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("sent by another client: error %v; want ErrInvalidInput", err)
	}
	_, err = client.Send(t.Context(), req)
	var items *ItemsError
	if !errors.As(err, &items) || !slices.Equal(items.Failed, []ItemFailure{{Index: 3, ID: "p4",
		Status: http.StatusConflict}}) || items.Items != 2 || !errors.Is(err, ErrConflict) {
		t.Errorf("error %#v; want item 3 failed", err)
	}
	if last, err := client.Prepare(batches[2]); err != nil || last.Items() != 1 {
		t.Errorf("request %+v, error %v; want one of 1 item", last, err)
	}
	if one, err := client.Prepare(&Write{Method: http.MethodPost, Path: bulkPath, Body: jsonBody(t, `{}`)}); err != nil ||
		one.Items() != 0 {
		t.Errorf("request %+v, error %v; want no items", one, err)
	}
}

// TestJoinItemsJoinsTheFailuresOfBatches checks that the failures of the
// batches of one write are joined in order, as those of one write of every
// item, of the class that they all share, if any; none failed, nothing.
func TestJoinItemsJoinsTheFailuresOfBatches(t *testing.T) {
	if err := JoinItems(400, nil); err != nil {
		t.Errorf("no failures: %v; want nil", err)
	}
	conflict := func(index int) ItemFailure { return ItemFailure{Index: index, Status: http.StatusConflict} }
	first := &ItemsError{Method: http.MethodPut, Path: "/price/acme1/prices/bulk", Items: 200,
		Failed: []ItemFailure{conflict(7)}, class: ErrConflict}
	second := &ItemsError{Method: http.MethodPut, Path: "/price/acme1/prices/bulk", Items: 200,
		Failed: []ItemFailure{conflict(203), conflict(250)}, class: ErrConflict}
	err := JoinItems(450, []*ItemsError{first, second})
	var joined *ItemsError
	if !errors.As(err, &joined) || !slices.Equal(joined.Failed, []ItemFailure{conflict(7), conflict(203),
		conflict(250)}) || !errors.Is(err, ErrConflict) ||
		err.Error() != "PUT /price/acme1/prices/bulk: 3 of 450 items failed" {
		t.Errorf("error %#v (%v); want 3 of 450 conflicts", err, err)
	}
	third := &ItemsError{Method: http.MethodPut, Path: "/price/acme1/prices/bulk", Items: 50,
		Failed: []ItemFailure{{Index: 420, Status: http.StatusNotFound}}, class: ErrNotFound}
	err = JoinItems(450, []*ItemsError{first, third})
	if errors.Unwrap(err) != nil || !strings.HasSuffix(err.Error(), ": 2 of 450 items failed") {
		t.Errorf("error %v of class %v; want 2 of 450 of no class", err, errors.Unwrap(err))
	}
}

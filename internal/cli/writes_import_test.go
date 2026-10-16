package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// productsBulk is the request URI of the Product Service's bulk writes, as
// sent without their flags.
const productsBulk = "/product/acme1/products/bulk?" + productSwitches

// productsFile returns a JSON array of n products, 100000 and on, one per
// line with the spaces of a file written by hand, and the products as given.
func productsFile(n int) (file string, items []string) {
	items = make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf(`{"id": "%d", "name": "Product %d"}`, 100000+i, i)
	}
	return "[\n  " + strings.Join(items, ",\n  ") + "\n]\n", items
}

// answerItems answers a bulk write item by item, as Emporix does (207): the
// result of each item at its position, with the code that codes gives its
// id, 204 for any other.
func answerItems(t *testing.T, codes map[string]int) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var items []struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			t.Errorf("a bulk write sent a body that is not an array of items: %v", err)
		}
		results := make([]string, len(items))
		for i, item := range items {
			code := http.StatusNoContent
			if failed, ok := codes[item.ID]; ok {
				code = failed
			}
			results[i] = fmt.Sprintf(`{"index":%d,"id":%q,"code":%d,"message":"canned"}`, i, item.ID, code)
		}
		respond(http.StatusMultiStatus, "["+strings.Join(results, ",")+"]")(w, r)
	}
}

// runImport runs `products import` of a file of n products against fake,
// with flags, where standard input is not a terminal.
func runImport(t *testing.T, fake *fakeEmporix, n int, flags ...string) (code int, stdout, stderr string) {
	t.Helper()
	file, _ := productsFile(n)
	return runIn(t, writeEnv(fake, false, ""), append([]string{"products", "import", "--file", writeBody(t, file)},
		flags...)...)
}

// TestProductImportSendsBatchesInOrder checks that an import of more products than
// one request takes sends them in requests of up to 200, in order, each
// product as given, with one token, prints each answer and succeeds when
// every product did.
func TestProductImportSendsBatchesInOrder(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products/bulk", answerItems(t, nil))
	code, stdout, stderr := runImport(t, fake, 450, "--yes")
	if code != ExitOK || stderr != "" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	fake.assertRequests(t, productsBulk, productsBulk, productsBulk)
	fake.assertTokens(t, tokenScope([]string{productManageScope}))
	_, items := productsFile(450)
	for i, batch := range [][]string{items[:200], items[200:400], items[400:]} {
		if want := "[" + strings.Join(batch, ",") + "]"; fake.bodies[i] != want || fake.languages[i] != "*" {
			t.Errorf("request %d sent %.80q… in %q; want %.80q… in *", i+1, fake.bodies[i], fake.languages[i], want)
		}
	}
	if answers, results := strings.Count(stdout, "\n]\n"), strings.Count(stdout, `"code": 204`); answers != 3 ||
		results != 450 {
		t.Errorf("printed %d answers of %d results; want 3 of 450 in all", answers, results)
	}
}

// TestProductImportShowsAndConfirmsEveryRequest checks that a dry run of an import
// in batches prints every request, an empty line between two, and that its
// confirmation names how many requests it sends: declined, it sends none;
// confirmed, it sends them all.
func TestProductImportShowsAndConfirmsEveryRequest(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/product/acme1/products/bulk", answerItems(t, nil))
	code, stdout, _ := runImport(t, fake, 201, "--dry-run")
	request := "PUT " + fake.server.URL + productsBulk + "\n"
	if code != ExitOK || strings.Count(stdout, request) != 2 || !strings.Contains(stdout, "}]\n\n"+request) {
		t.Errorf("exit %d, dry run:\n%s\nwant two requests apart", code, stdout)
	}
	fake.assertNothingSent(t)
	file, _ := productsFile(201)
	path := writeBody(t, file)
	for answer, sent := range map[string]int{"n\n": 0, "y\n": 2} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/bulk", answerItems(t, nil))
		code, _, stderr := runIn(t, writeEnv(fake, true, answer), "products", "import", "--file", path)
		prompt := "Send PUT " + fake.server.URL + productsBulk + " to tenant acme1 in 2 requests? [y/N] "
		if (code == ExitOK) != (sent > 0) || !strings.HasPrefix(stderr, prompt) || len(fake.gets) != sent {
			t.Errorf("answered %q: exit %d, stderr %q, %d sent; want %d sent after %q", answer, code, stderr,
				len(fake.gets), sent, prompt)
		}
	}
}

// TestProductImportListsTheProductsThatFailed checks that an import goes on after a
// batch whose products failed, lists each of them on standard error by its
// position in --file, and then fails with the class that every failure
// shares: 5 when each was a conflict, 1 otherwise.
func TestProductImportListsTheProductsThatFailed(t *testing.T) {
	for _, tt := range []struct {
		codes map[string]int
		code  int
		want  string
	}{
		{codes: map[string]int{"100203": http.StatusConflict, "100420": http.StatusBadRequest}, code: ExitError,
			want: "item 203 (100203): 409 Conflict: canned\nitem 420 (100420): 400 Bad Request: canned\n" +
				"Error: PUT /product/acme1/products/bulk: 2 of 450 items failed\n"},
		{codes: map[string]int{"100007": http.StatusConflict, "100449": http.StatusConflict}, code: ExitConflict,
			want: "item 7 (100007): 409 Conflict: canned\nitem 449 (100449): 409 Conflict: canned\n" +
				"Error: PUT /product/acme1/products/bulk: 2 of 450 items failed (each changed since it was read, or " +
				"exists already: read them again before you retry)\n"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/bulk", answerItems(t, tt.codes))
		code, stdout, stderr := runImport(t, fake, 450, "--yes")
		if code != tt.code || stderr != tt.want || strings.Count(stdout, "\n]\n") != 3 {
			t.Errorf("%v: exit %d, stderr %q; want exit %d and %q", tt.codes, code, stderr, tt.code, tt.want)
		}
		fake.assertRequests(t, productsBulk, productsBulk, productsBulk)
	}
}

// TestProductImportStopsAtARequestThatFails checks that a batch of an import that
// fails as a whole ends the import, which sends no further batch and says
// where it stopped; its failure sets the exit code only where every product
// that failed before it shares its class.
func TestProductImportStopsAtARequestThatFails(t *testing.T) {
	const where = "Error: request 2 of 3, from item 200: PUT /product/acme1/products/bulk: "
	const before = " (the 200 items before it were sent, and 1 of them failed, as listed above; those after it " +
		"were not)\n"
	const conflict = "409 Conflict: canned (the item changed since it was read, or exists already: read it " +
		"again before you retry)"
	for _, tt := range []struct {
		first  int // the status of product 100007 in the first batch; 0: it succeeds
		status int // the status of the second batch
		code   int
		want   string
	}{
		{first: http.StatusBadRequest, status: http.StatusConflict, code: ExitError,
			want: "item 7 (100007): 400 Bad Request: canned\n" + where + conflict + before},
		{first: http.StatusConflict, status: http.StatusConflict, code: ExitConflict,
			want: "item 7 (100007): 409 Conflict: canned\n" + where + conflict + before},
		{status: http.StatusNotFound, code: ExitNotFound, want: where + "404 Not Found: canned (the 200 items " +
			"before it were sent; those after it were not)\n"},
		{status: http.StatusInternalServerError, code: ExitError, want: where + "500 Internal Server Error: canned " +
			"(it may have been carried out all the same: read the item before you send it again) (the 200 items"},
	} {
		fake := newFakeEmporix(t)
		first := map[string]int{}
		if tt.first != 0 {
			first["100007"] = tt.first
		}
		fake.route("/product/acme1/products/bulk", inTurn(answerItems(t, first),
			respond(tt.status, `{"message":"canned"}`), answerItems(t, nil)))
		code, _, stderr := runImport(t, fake, 450, "--yes")
		if code != tt.code || !strings.Contains(stderr, tt.want) {
			t.Errorf("%d then %d: exit %d, stderr %q; want exit %d and %q", tt.first, tt.status, code, stderr, tt.code,
				tt.want)
		}
		fake.assertRequests(t, productsBulk, productsBulk)
	}
}

// TestProductImportStopsWhenAnAnswerCannotBePrinted checks that an import whose
// answer to a batch cannot be printed, one whose products all succeeded or
// one that lists a failure, sends no further batch and says so.
func TestProductImportStopsWhenAnAnswerCannotBePrinted(t *testing.T) {
	file, _ := productsFile(450)
	for _, failed := range []string{"100000", "100300"} {
		fake := newFakeEmporix(t)
		fake.route("/product/acme1/products/bulk", answerItems(t, map[string]int{failed: http.StatusConflict}))
		env := writeEnv(fake, false, "")
		var errOut bytes.Buffer
		env.Out, env.Err = failingWriter{}, &errOut
		code := Run(context.Background(), []string{"products", "import", "--file", writeBody(t, file), "--yes"}, env)
		const want = "Error: request 1 of 3 was answered, but disk full; those after it were not sent\n"
		if code != ExitError || !strings.HasSuffix(errOut.String(), want) {
			t.Errorf("%s failed: exit %d, stderr %q; want exit 1 and %q", failed, code, errOut.String(), want)
		}
		fake.assertRequests(t, productsBulk)
	}
}

// failingAfter is a writer that takes writes until it took ok of them, then
// fails each.
type failingAfter struct{ ok int }

func (f *failingAfter) Write(p []byte) (int, error) {
	if f.ok == 0 {
		return failingWriter{}.Write(p)
	}
	f.ok--
	return len(p), nil
}

// TestProductImportDryRunFailsWhenItCannotPrint checks that a dry run of an
// import in batches whose requests cannot be printed, the first or the empty
// line before the second, fails (exit 1) and sends nothing.
func TestProductImportDryRunFailsWhenItCannotPrint(t *testing.T) {
	file, _ := productsFile(201)
	for _, ok := range []int{0, 1, 2} {
		fake := newFakeEmporix(t)
		env := writeEnv(fake, false, "")
		var errOut bytes.Buffer
		env.Out, env.Err = &failingAfter{ok: ok}, &errOut
		code := Run(context.Background(), []string{"products", "import", "--file", writeBody(t, file), "--dry-run"}, env)
		if code != ExitError || errOut.String() != "Error: disk full\n" {
			t.Errorf("failing after %d writes: exit %d, stderr %q; want exit 1 and the failure", ok, code, errOut.String())
		}
		fake.assertNothingSent(t)
	}
}

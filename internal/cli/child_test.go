package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestChildItemRead(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/a%2Fb/parts/big%20one/007", respond(http.StatusOK, `{"id":"007","name":"Bolt"}`))
	code, stdout, errText := runThings(t, fake, "named-part", "a/b", "big one", "007")
	if code != ExitOK || stdout != "ID   NAME\n007  Bolt\n" ||
		!slices.Equal(fake.tokenScopes, []string{"tenant=acme1 part.read"}) {
		t.Fatalf("exit %d, stdout %q, error %q, token scopes %q", code, stdout, errText, fake.tokenScopes)
	}
	fake.assertRequests(t, "/things/a%2Fb/parts/big%20one/007?all=true")
	code, stdout, errText = runThings(t, fake, "named-part", "a/b", "big one", "008")
	if code != ExitNotFound || stdout != "" || !strings.Contains(errText, "GET /things/a%2Fb/parts/big%20one/008") {
		t.Fatalf("unknown: exit %d, stdout %q, error %q", code, stdout, errText)
	}
}

func TestChildItemReadOnASite(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/stores/main/s%2F1/racks/x1", respond(http.StatusOK, `{"id":"x1","name":"X"}`))
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	stores := resource[thing]{
		use: "stores", under: "things", path: "/stores/{site}", scopes: []string{"store.read"}, view: thingView(),
	}
	shelves := child[thing]{use: "shelves", sub: "racks", query: url.Values{"expand": {"bins"}}, view: thingView()}
	shelf := childItemCommand(a, stores, shelves, &childItem{use: "shelf", long: "Reads one shelf of a store."})
	if want := "shelf THING_ID ID"; shelf.Use != want {
		t.Errorf("use %q, want %q", shelf.Use, want)
	}
	if want := "Reads one shelf of a store. " +
		"THING_ID is the ID that `epx things stores list --site S` shows; an unknown one is not found. " +
		"ID is the ID that `epx things stores shelves THING_ID --site S` shows; an unknown one is not found.\n\n" +
		"Requires the store.read scope."; shelf.Long != want {
		t.Errorf("help %q, want %q", shelf.Long, want)
	}
	root := &cobra.Command{Use: "epx", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newResourceCommand(a, stores, shelf))
	applyExitPolicy(root)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"stores", "shelf", "s/1", "x1", "--site", "main"})
	if err := root.Execute(); err != nil || out.String() != "ID  NAME\nx1  X\n" {
		t.Fatalf("error %v, stdout %q", err, out.String())
	}
	fake.assertRequests(t, "/stores/main/s%2F1/racks/x1?expand=bins")
	if !slices.Equal(fake.tokenScopes, []string{"tenant=acme1 store.read"}) {
		t.Fatalf("token scopes %q; want the parent's", fake.tokenScopes)
	}
}

func TestChildSearch(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/stores/main/s%201/racks/search", firstPage(`[{"id":"x1","name":"X"}]`))
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	stores := resource[thing]{
		use: "stores", under: "things", path: "/stores/{site}", scopes: []string{"store.read"}, view: thingView(),
	}
	shelves := listedChildOf(listing{filters: listFilters{kinds: filterable}}, child[thing]{
		use: "shelves", sub: "racks", query: url.Values{"expand": {"bins"}}, view: thingView(),
	})
	search := childSearchCommand(a, stores, shelves, &childSearch{short: "Search shelves",
		long: "Reads the shelves a q filter finds.", fields: []bodyField{queryField()}})
	if want := "search-shelves THING_ID"; search.Use != want {
		t.Errorf("use %q, want %q", search.Use, want)
	}
	if want := "Reads the shelves a q filter finds. " +
		"It reads and shows them as `epx things stores shelves THING_ID --site S` does. " +
		"THING_ID is the ID that `epx things stores list --site S` shows; an unknown one is not found.\n\n" +
		"Changes nothing:"; !strings.HasPrefix(search.Long, want) {
		t.Errorf("help %q, want it to start with %q", search.Long, want)
	}
	root := &cobra.Command{Use: "epx", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newResourceCommand(a, stores, search))
	applyExitPolicy(root)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"stores", "search-shelves", "s 1", "--site", "main", "--query", "id:x1", "--sort", "name"})
	if err := root.Execute(); err != nil || out.String() != "ID  NAME\nx1  X\n" {
		t.Fatalf("error %v, stdout %q", err, out.String())
	}
	assertQueries(t, fake.gets, "/stores/main/s%201/racks/search", url.Values{"expand": {"bins"}, "sort": {"name"}})
	fake.assertSent(t, http.MethodPost, `{"q":"id:x1"}`)
	fake.assertScopes(t, "store.read")
}

func TestChildItemReadsWithTheChildsScopes(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/t1/parts/x1", respond(http.StatusOK, `{"id":"x1","name":"X"}`))
	code, stdout, errText := runThings(t, fake, "part", "t1", "x1")
	token := []string{"tenant=acme1 part.read"}
	if code != ExitOK || stdout != "ID  NAME\nx1  X\n" || !slices.Equal(fake.tokenScopes, token) {
		t.Fatalf("exit %d, stdout %q, error %q, token scopes %q", code, stdout, errText, fake.tokenScopes)
	}
}

func TestChildListSendsBothQueries(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/things/t%2F1/parts", firstPage(`[{"id":"x1","name":"X"}]`))
	if code, _, errText := runThings(t, fake, "parts", "t/1"); code != ExitOK {
		t.Fatalf("exit %d, error %q", code, errText)
	}
	assertQueries(t, fake.gets, "/things/t%2F1/parts", url.Values{"all": {"true"}, "expand": {"tags"}})
}

func TestChildItemHelp(t *testing.T) {
	_, help, _ := runThings(t, newFakeEmporix(t), "named-part", "--help")
	want := "Reads one part of a thing by name. " +
		"THING_ID is the ID that `epx things list` shows; an unknown one is not found. " +
		"NAME and ID are the NAME and ID that `epx things parts THING_ID` shows; an unknown one is not found. " +
		"NAME is a part's name.\n\n" +
		"Requires the part.read scope.\n\nUsage:\n  epx things named-part THING_ID NAME ID"
	if !strings.HasPrefix(flat(help), flat(want)) {
		t.Fatalf("help:\n%s\nwant it to start with:\n%s", help, want)
	}
}

func TestArgumentOfAGroupWithoutList(t *testing.T) {
	runs := resource[thing]{use: "runs", under: "things", view: thingView()}
	runs.singular = "run"
	if got, want := runs.argumentSentence(notFound), "RUN_ID is the ID that `epx things runs list` shows; "+
		"an unknown one is not found."; got != want {
		t.Errorf("with list: %q, want %q", got, want)
	}
	runs.listedBy = "things runs-of THING_ID"
	if got, want := runs.argumentSentence(listsNothing), "RUN_ID is the ID that `epx things runs-of THING_ID` "+
		"shows; an unknown one lists nothing."; got != want {
		t.Errorf("without list: %q, want %q", got, want)
	}
}

func TestNestedParentNamesItsGroup(t *testing.T) {
	stores := resource[thing]{use: "stores", under: "things", path: "/stores", view: thingView()}
	shelves := stores.nested("shelves")
	c := child[thing]{use: listUse, sub: shelves.use, view: thingView()}
	if got, want := childListLine(shelves, c), "things stores shelves list THING_ID"; got != want {
		t.Errorf("list line %q, want %q", got, want)
	}
	if got, want := shelves.argumentSource(), "THING_ID is the ID that `epx things stores list` shows"; got != want {
		t.Errorf("argument source %q, want %q", got, want)
	}
	target := childItemTarget(shelves, c, "get", nil)
	if got, want := target.use, "get THING_ID ID"; got != want {
		t.Errorf("item use %q, want %q", got, want)
	}
	if want := "/stores/{parent}/shelves/{id}"; target.path != want {
		t.Errorf("item path %q, want %q", target.path, want)
	}
	if got, want := childListLine(stores, child[thing]{use: "shelves", sub: "racks", view: thingView()}),
		"things stores shelves THING_ID"; got != want {
		t.Errorf("not nested: %q, want %q", got, want)
	}
}

func TestItemArgumentsNameColumns(t *testing.T) {
	c := child[thing]{view: view[thing]{singular: "part", header: []string{"ID", "LEGAL ENTITY"}}}
	args, key := c.itemArguments("x", []string{"LEGAL ENTITY", "ID"})
	if !slices.Equal(args, []string{"LEGAL_ENTITY", "ID"}) || key != placeholderInner+"/"+placeholderID {
		t.Fatalf("two columns: arguments %q, key %q", args, key)
	}
	if args, key := c.itemArguments("x", []string{"ID"}); !slices.Equal(args, []string{"ID"}) || key != placeholderID {
		t.Fatalf("one column: arguments %q, key %q", args, key)
	}
	for _, tt := range []struct {
		columns []string
		want    string
	}{
		{columns: []string{"KIND", "ID"}, want: "x: KIND is not a column of part"},
		{columns: []string{"ID", "ID", "ID"}, want: "x: an item takes at most 2 arguments after its parent's"},
	} {
		func() {
			defer func() {
				if got := fmt.Sprint(recover()); got != tt.want {
					t.Errorf("columns %q: panic %q, want %q", tt.columns, got, tt.want)
				}
			}()
			c.itemArguments("x", tt.columns)
		}()
	}
}

func TestChildItemUsageErrorsSendNothing(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		args    []string
		message string
	}{
		{args: []string{"named-part", "1", "2"}, message: "accepts 3 arg(s), received 2"},
		{args: []string{"named-part", "1", "2", "3", "4"}, message: "accepts 3 arg(s), received 4"},
		{args: []string{"named-part", "1", "..", "3"}, message: `invalid NAME ".."`},
		{args: []string{"named-part", "1", "2", ""}, message: `invalid ID ""`},
	} {
		code, _, errText := runThings(t, fake, tt.args...)
		if code != ExitUsage || !strings.Contains(errText, tt.message) {
			t.Errorf("%v: exit %d, error %q (want %q)", tt.args, code, errText, tt.message)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("usage errors must not contact the API: tokens %q, requests %q", fake.tokenScopes, fake.gets)
	}
}

func TestSegmentItemTypeIsAChoice(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, kind := range []string{"product", "Category", "BRAND"} {
		code, _, stderr := runCLI(t, fake.env(), "segments", "items", "get", "s1", kind, "100200")
		if want := fmt.Sprintf("invalid TYPE %q: use PRODUCT or CATEGORY", kind); code != ExitUsage ||
			!strings.Contains(stderr, want) {
			t.Errorf("%s: exit %d, error %q (want %q)", kind, code, stderr, want)
		}
	}
	if len(fake.tokenScopes) != 0 || len(fake.gets) != 0 {
		t.Fatalf("usage errors must not contact the API: tokens %q, requests %q", fake.tokenScopes, fake.gets)
	}
	fake.route("/customer-segment/acme1/segments/s1/items/CATEGORY/c1",
		allLanguages(respond(http.StatusOK, `{"segmentId":"s1","type":"CATEGORY","item":{"id":"c1"}}`)))
	if code, stdout, stderr := runCLI(t, fake.env(), "segments", "items", "get", "s1", "CATEGORY", "c1"); code != ExitOK ||
		!strings.Contains(stdout, "c1  s1       CATEGORY") {
		t.Fatalf("CATEGORY: exit %d, stdout %q, error %q", code, stdout, stderr)
	}
}

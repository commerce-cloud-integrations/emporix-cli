package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestLimitIsOfferedOnEveryListAndNoSingleRead(t *testing.T) {
	limited := commandsWithFlag("limit")
	// Single reads print one object: every get, an order's history (one object
	// holding its status changes), an asset's download URL, a user's scopes, the
	// webhook statistics, one setting of a client, one coupon redemption, a
	// referral coupon, a reward points balance, a cart found by its criteria,
	// its delivery restrictions and its validation, the AI Service's LLM
	// models, commerce events and analytics, the webhook dashboard access, an
	// export package and the import reads of one object: a schedule, a stream
	// order, a run's streams and diagnostics, the statistics, health thresholds
	// and limits (one response each); or one response: the order count, the
	// prices and price matches of given products, the products a segment
	// match finds, the calculations and checks, and a parsed JSON schema file.
	single := []string{
		"epx orders history", "epx media download-url", "epx iam users scopes", "epx webhooks statistics",
		"epx configurations client-setting",
		"epx coupons redemption", "epx coupons referral",
		"epx reward-points balance", "epx carts find", "epx carts delivery-restrictions", "epx carts validate",
		"epx ai models", "epx ai commerce-events", "epx orders count", "epx prices search", "epx prices match",
		"epx ai analytics tenant", "epx ai analytics agent", "epx ai analytics executions",
		"epx webhooks dashboard-access", "epx custom-entities export", "epx search indexes export", "epx segments match",
		"epx imports configurations schedule", "epx imports configurations stream-order", "epx imports runs streams",
		"epx imports runs diagnostics", "epx imports statistics", "epx imports health-thresholds", "epx imports license",
		"epx shipping quote", "epx shipping quote-minimum", "epx shipping quote-slot", "epx delivery-times validate",
		"epx delivery-times generate-cycle",
		"epx units convert", "epx units conversion-factor", "epx taxes calculate", "epx coupons validate",
		"epx schemas parse", "epx quotes pdf",
	}
	// One item of a nested list, and the mixins of a site (one object).
	single = append(single, "epx segments customers get-b2b", "epx sites mixins list")
	for _, path := range readCommands() {
		want := !strings.HasSuffix(path, " get") && !slices.Contains(single, path)
		if slices.Contains(limited, path) != want {
			t.Errorf("%s: --limit offered %v, want %v", path, !want, want)
		}
	}
}

func TestLimitStopsReading(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/brand/brands", respond(http.StatusOK, `[{"id":"b1","name":"Alpha"}]`, "X-Total-Count", "250"))
	code, stdout, stderr := runCLI(t, fake.env(), "brands", "list", "--limit", "1")
	if code != ExitOK || stdout != "ID  NAME\nb1  Alpha\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	// One page of one item, although the total is 250.
	assertQueries(t, fake.gets, "/brand/brands", url.Values{})
	if sent, err := url.Parse(fake.gets[0]); err != nil || len(fake.gets) != 1 || sent.Query().Get("pageSize") != "1" {
		t.Fatalf("requests %q; want one page of one item", fake.gets)
	}
	// A list read whole is cut.
	fake.route("/configuration/acme1/configurations", respond(http.StatusOK,
		`[{"key":"a","value":"1"},{"key":"b","value":"2"},{"key":"c","value":"3"}]`))
	code, stdout, stderr = runCLI(t, fake.env(), "configurations", "list", "--limit", "2", "-o", "json")
	var settings []struct {
		Key string `json:"key"`
	}
	if code != ExitOK || json.Unmarshal([]byte(stdout), &settings) != nil || len(settings) != 2 ||
		settings[1].Key != "b" {
		t.Fatalf("configurations: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestLimitRejectsInvalidValues(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, value := range []string{"-1", "x", "1.5", "", "99999999999999999999"} {
		code, _, stderr := runCLI(t, fake.env(), "brands", "list", "--limit", value)
		if code != ExitUsage || !strings.Contains(stderr, `invalid limit "`+value+`"`) {
			t.Errorf("--limit %q: exit %d, stderr %q", value, code, stderr)
		}
	}
	if len(fake.gets) != 0 {
		t.Fatalf("requests %q; want none", fake.gets)
	}
}

func TestQueryFilteredTotalAndMalformedFilter(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/brand/brands", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("q") {
		case "name:~a":
			// The total counts the filtered items, so the read stops after them.
			respond(http.StatusOK, `[{"id":"b1","name":"Alpha"},{"id":"b2","name":"Gamma"}]`, "X-Total-Count", "2")(w, r)
		case "name:((":
			respond(http.StatusBadRequest, `{"message":"Missing closing parenthesis in value '((' of key 'name'"}`)(w, r)
		default:
			respond(http.StatusOK, `[{"id":"b1","name":"Alpha"},{"id":"b2","name":"Gamma"},{"id":"b3","name":"Ox"}]`,
				"X-Total-Count", "3")(w, r)
		}
	})
	code, stdout, stderr := runCLI(t, fake.env(), "brands", "list", "--query", "name:~a")
	if code != ExitOK || stdout != "ID  NAME\nb1  Alpha\nb2  Gamma\n" || len(fake.gets) != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q, requests %q", code, stdout, stderr, fake.gets)
	}
	code, stdout, stderr = runCLI(t, fake.env(), "brands", "list", "--query", "name:((")
	if code != ExitError || stdout != "" || !strings.Contains(stderr, "400 Bad Request: Missing closing parenthesis") {
		t.Fatalf("malformed: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestListFiltersAreOfferedExactlyWhereTheCasesDeclareThem checks --query and
// --sort against the filters of the read cases: offered exactly where the
// Emporix endpoint documents q and sort.
func TestListFiltersAreOfferedExactlyWhereTheCasesDeclareThem(t *testing.T) {
	var withQuery, withSort []string
	for path, tt := range casesByCommand(t) {
		if tt.filters.sorts() {
			withSort = append(withSort, path)
		}
		if tt.filters.queries() {
			withQuery = append(withQuery, path)
		}
	}
	slices.Sort(withQuery)
	slices.Sort(withSort)
	// A search's --query fills its body instead, as its read cases state.
	searches := commandsWithFlag(bodyFlag)
	if got := slices.DeleteFunc(commandsWithFlag("query"), func(path string) bool {
		return slices.Contains(searches, path)
	}); !slices.Equal(got, withQuery) {
		t.Errorf("--query on %q; want exactly %q", got, withQuery)
	}
	if got := commandsWithFlag("sort"); !slices.Equal(got, withSort) {
		t.Errorf("--sort on %q; want exactly %q", got, withSort)
	}
}

// TestListFiltersReachEveryRequest runs every filterable or sortable list
// case with --query and --sort, as it takes them: every request (each page)
// carries them on top of the case's own query, and the user's sort replaces a
// fixed one.
func TestListFiltersReachEveryRequest(t *testing.T) {
	const q, order = `name.en:~(Big shirt) published:true`, "name.en:desc,code"
	for path, tt := range casesByCommand(t) {
		if tt.filters.kinds == noFilters {
			continue
		}
		t.Run(path, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route(tt.path, tt.served(t))
			want, err := url.ParseQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			args := slices.Concat(tt.args, []string{"-o", "json"})
			if tt.filters.sorts() {
				args = append(args, "--sort", order)
				want.Set("sort", order)
			}
			if tt.filters.queries() {
				args = append(args, "--query", q)
				want.Set("q", q)
			}
			if code, _, stderr := runCLI(t, fake.env(), args...); code != ExitOK {
				t.Fatalf("exit %d, stderr %q", code, stderr)
			}
			fake.assertMode(t, tt.mode, tt.page, tt.path, want.Encode())
		})
	}
}

// TestStreamIsOfferedOnListsOfItems checks that every list takes --stream but
// category trees (list and search), whose shape restructures their items.
func TestStreamIsOfferedOnListsOfItems(t *testing.T) {
	want := slices.DeleteFunc(commandsWithFlag("limit"), func(path string) bool {
		return path == "epx category-trees list" || path == "epx category-trees search"
	})
	if got := commandsWithFlag("stream"); !slices.Equal(got, want) {
		t.Fatalf("--stream on:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestStreamPrintsJSONLines(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/brand/brands", firstPage(`[{"name":"Alpha","id":"b1"},{"id":"b2","name":"Beta"}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "brands", "list", "--stream")
	if code != ExitOK || stdout != "{\"id\":\"b1\",\"name\":\"Alpha\"}\n{\"id\":\"b2\",\"name\":\"Beta\"}\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, fake.env(), "brands", "list", "--stream", "--limit", "1")
	if code != ExitOK || stdout != "{\"id\":\"b1\",\"name\":\"Alpha\"}\n" {
		t.Fatalf("--limit 1: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestStreamFailsAfterThePagesBefore checks that a check failing on a later
// page exits 1 after the items of the pages before were printed.
func TestStreamFailsAfterThePagesBefore(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/brand/brands", respond(http.StatusOK, `[{"id":"b1"}]`)) // every page repeats page 1
	code, stdout, stderr := runCLI(t, fake.env(), "brands", "list", "--stream")
	if code != ExitError || stdout != "{\"id\":\"b1\"}\n" ||
		!strings.Contains(stderr, "page 2 repeats an item of page 1") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	// Without --stream, nothing is printed.
	if code, stdout, _ = runCLI(t, fake.env(), "brands", "list", "-o", "json"); code != ExitError || stdout != "" {
		t.Fatalf("buffered: exit %d, stdout %q", code, stdout)
	}
}

func TestStreamExcludesOutput(t *testing.T) {
	fake := newFakeEmporix(t)
	code, _, stderr := runCLI(t, fake.env(), "brands", "list", "--stream", "-o", "json")
	if code != ExitUsage || !strings.Contains(stderr, "stream") || len(fake.gets) != 0 {
		t.Fatalf("exit %d, stderr %q, requests %q", code, stderr, fake.gets)
	}
}

// TestStreamBelongsToOneRun checks that a command run again without --stream
// prints a table, not JSON Lines into the writer of the run before.
func TestStreamBelongsToOneRun(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/brand/brands", firstPage(`[{"id":"b1","name":"Alpha"}]`))
	a := &app{getenv: func(key string) string { return fake.env()[key] }, timeout: defaultTimeout}
	cmd := newBrandsCommand(a)
	applyExitPolicy(cmd)
	var first, second bytes.Buffer
	for _, run := range []struct {
		out  *bytes.Buffer
		args []string
	}{{&first, []string{"list", "--stream"}}, {&second, []string{"list", "--stream=false"}}} {
		cmd.SetOut(run.out)
		cmd.SetArgs(run.args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	if first.String() != "{\"id\":\"b1\",\"name\":\"Alpha\"}\n" || second.String() != "ID  NAME\nb1  Alpha\n" {
		t.Fatalf("first run %q, second run %q", first.String(), second.String())
	}
}

// TestSortHelpStatesWhatEmporixAnswers checks that the help of --sort says what
// the list's service does with an unknown field: payment transactions are refused
// it (400), and the other lists keep the shared help.
func TestSortHelpStatesWhatEmporixAnswers(t *testing.T) {
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		flag := cmd.Flags().Lookup("sort")
		if flag == nil {
			return
		}
		if strings.HasSuffix(cmd.CommandPath(), "payment-transactions list") {
			if flag.Usage != paymentTransactionsSortUsage {
				t.Errorf("%s --sort help %q; want %q", cmd.CommandPath(), flag.Usage, paymentTransactionsSortUsage)
			}
			return
		}
		if flag.Usage != defaultSortUsage {
			t.Errorf("%s --sort help %q; want the shared help", cmd.CommandPath(), flag.Usage)
		}
	})
}

package cli

import (
	"cmp"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// readCase is a read command, the response it gets, and what it must request
// and print. It checks what a declaration states (path, query, scopes, read
// mode and table); the shared builder's behavior is tested in resource_test.go.
type readCase struct {
	args    []string    // command, arguments and flags
	path    string      // the request path as escaped on the wire, which the fake serves
	query   string      // the request query, paging aside, encoded with sorted keys
	scopes  []string    // the scopes besides the tenant
	mode    readMode    // how the command reads (getAll, the default, pages the collection)
	page    string      // the page size of a getAllInOnePage or getAllOnFirstPage read, which only they name
	filters listFilters // what a list takes besides --limit (TestListFiltersReachEveryRequest)
	method  string      // the method of every request; GET when empty
	sent    string      // the body of every request, e.g. a search's ({page}: its page); none when empty
	form    []sentPart  // the parts of the form of every request, in place of sent; none when nil
	body    string      // the response; a paged read gets it on page 1 and nothing after
	total   string      // the X-Total-Count served instead of a body, as to a HEAD request; none when empty
	media   string      // the Content-Type of a file served as body, in the language that args name (fileRead)
	want    string      // the table printed, or the file
}

// readCases returns the read cases of every service.
func readCases() []readCase {
	return slices.Concat(productReads(),
		categoryReads(),
		categoryTreeReads(),
		catalogReads(),
		labelReads(),
		mediaReads(),
		schemaAndUnitReads(),
		schemaServiceReads(),
		schemaParseReads(),
		priceReads(),
		priceModelReads(),
		priceListReads(),
		feeReads(),
		feeAssignmentReads(),
		availabilityReads(),
		siteReads(),
		currencyReads(),
		countryReads(),
		taxAndConfigurationReads(),
		indexingReads(),
		customerReads(),
		sessionContextReads(),
		legalEntityReads(),
		customerManagementReads(),
		segmentReads(),
		segmentMemberReads(),
		nestedItemReads(),
		vendorReads(),
		orderReads(),
		invoiceJobReads(),
		deliveryTimeReads(),
		shippingReads(),
		paymentModeReads(),
		paymentTransactionReads(),
		returnReads(),
		quoteReads(),
		shoppingListReads(),
		cartReads(),
		couponReads(),
		rewardPointReads(),
		iamReads(),
		iamAccessReads(),
		iamUserAccessReads(),
		webhookReads(),
		searchReads(),
		searchJobReads(),
		searchExportReads(),
		catalogSearchReads(),
		customerSearchReads(),
		aiReads(),
		commerceSearchReads(),
		instanceAndDocumentSearchReads(),
		calculationReads(),
		utilityReads(),
		importReads())
}

// tableReadCases returns the read cases of the reads that print a table:
// all but those of a read that saves the file that Emporix answers with
// (media).
func tableReadCases() []readCase {
	all := readCases()
	var tables []readCase
	for i := range all {
		if all[i].media == "" {
			tables = append(tables, all[i])
		}
	}
	return tables
}

// TestReads runs every read case. The fake serves the body only when every
// translation is requested, as the client always asks (Accept-Language: *).
func TestReads(t *testing.T) {
	for _, tt := range readCases() {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route(tt.path, tt.served(t))
			code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
			if code != ExitOK || stdout != tt.want {
				t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, tt.want)
			}
			fake.assertScopes(t, tt.scopes...)
			fake.assertMode(t, tt.mode, tt.page, tt.path, tt.query)
			if tt.form != nil {
				for i := range fake.sent {
					fake.assertForm(t, i, tt.method, tt.form)
				}
				return
			}
			fake.assertSent(t, cmp.Or(tt.method, http.MethodGet), tt.sent)
		})
	}
}

// TestEveryReadHasACase checks that every read command has a read case, so a
// new command cannot ship without its path, scopes and table being pinned.
func TestEveryReadHasACase(t *testing.T) {
	cases := casesOfCommands(t)
	reads := readCommands()
	for _, path := range reads {
		if len(cases[path]) == 0 {
			t.Errorf("%s has no read case", path)
		}
	}
	// The other way round, so the check cannot pass with no read commands found.
	for path := range cases {
		if !slices.Contains(reads, path) {
			t.Errorf("read case of %s, which is not a read command", path)
		}
	}
}

// casesOfCommands returns the read cases of every command, by command path.
func casesOfCommands(t *testing.T) map[string][]readCase {
	t.Helper()
	root := newTestRoot()
	byPath := map[string][]readCase{}
	cases := readCases()
	for i := range cases {
		cmd, _, err := root.Find(cases[i].args)
		if err != nil || cmd.HasSubCommands() {
			t.Fatalf("read case %q names no read command", cases[i].args)
		}
		byPath[cmd.CommandPath()] = append(byPath[cmd.CommandPath()], cases[i])
	}
	return byPath
}

// casesByCommand returns the first read case of every command; the cases of
// one command must agree on its list filters.
func casesByCommand(t *testing.T) map[string]readCase {
	t.Helper()
	first := map[string]readCase{}
	for path, cases := range casesOfCommands(t) {
		for i := range cases {
			if cases[i].filters != cases[0].filters {
				t.Fatalf("%s: read cases disagree on the list filters", path)
			}
		}
		first[path] = cases[0]
	}
	return first
}

// readCommands returns the sorted paths of the read commands: every command
// that a reader builds, which notes its scopes (scopesAnnotation).
func readCommands() []string {
	var paths []string
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		if _, isRead := cmd.Annotations[scopesAnnotation]; isRead {
			paths = append(paths, cmd.CommandPath())
		}
	})
	slices.Sort(paths)
	return paths
}

// commandsWithFlag returns the sorted paths of the commands that define flag.
func commandsWithFlag(flag string) []string {
	var paths []string
	walkCommands(newTestRoot(), func(cmd *cobra.Command) {
		if cmd.Flags().Lookup(flag) != nil {
			paths = append(paths, cmd.CommandPath())
		}
	})
	slices.Sort(paths)
	return paths
}

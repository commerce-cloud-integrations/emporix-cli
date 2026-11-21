package cli

import (
	"cmp"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type readCase struct {
	args    []string
	path    string
	query   string
	scopes  []string
	mode    readMode
	page    string
	filters listFilters
	method  string
	sent    string
	form    []sentPart
	body    string
	total   string
	media   string
	site    string
	want    string
}

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
		mixinSiteReads(),
		calculationReads(),
		utilityReads(),
		importReads())
}

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
			fake.assertHeader(t, siteHeader, tt.site)
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

func TestEveryReadHasACase(t *testing.T) {
	cases := casesOfCommands(t)
	reads := readCommands()
	for _, path := range reads {
		if len(cases[path]) == 0 {
			t.Errorf("%s has no read case", path)
		}
	}
	for path := range cases {
		if !slices.Contains(reads, path) {
			t.Errorf("read case of %s, which is not a read command", path)
		}
	}
}

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

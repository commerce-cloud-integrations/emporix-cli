package cli

import "net/http"

// The read cases of item fees, which say which fees apply to which item: the
// list, read with GET, and the searches that Emporix serves as POST with a
// JSON body but which change nothing (see TestReads).

// Item YRNs of a product and of a payment type, and item fees that assign
// fees to them.
const (
	productYRN     = "urn:yaas:saasag:caasproduct:product:acme1;100200"
	paymentYRN     = "urn:yaas:saasag:paymentmode:code:acme1;invoice"
	feeAssignments = `[{"id":"if1","itemYrn":"` + productYRN + `","siteCode":"AcmeStore_NL","feeIds":["f1","f2"]},
		{"id":"if2","itemYrn":"` + paymentYRN + `","siteCode":"AcmeStore_NL","feeIds":[]}]`
	assignmentHeader = "ITEM                                              ID   SITE          FEES\n"
	assignmentTable  = assignmentHeader +
		productYRN + "  if1  AcmeStore_NL  f1,f2\n" +
		paymentYRN + "    if2  AcmeStore_NL  \n"
)

// feeAssignmentReads returns the reads of the Fee Service's item fees.
func feeAssignmentReads() []readCase {
	return []readCase{
		{
			args:  []string{"fees", "assignments", "list", "--site", "AcmeStore_NL"},
			path:  "/fee/acme1/itemFees",
			query: "siteCode=AcmeStore_NL",
			body:  feeAssignments,
			want:  assignmentTable,
		},
		{
			// The site goes in the query and in the body; Emporix returns the search whole. The
			// reference requires no id: an item fee without one is identified by its item.
			args: []string{
				"fees", "assignments", "search", "--item", productYRN, "--item", paymentYRN, "--site", "AcmeStore_NL",
			},
			path:   "/fee/acme1/itemFees/search",
			query:  "siteCode=AcmeStore_NL",
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"itemYrns":["` + productYRN + `","` + paymentYRN + `"],"siteCode":"AcmeStore_NL"}`,
			body: `[{"id":"if1","itemYrn":"` + productYRN + `","siteCode":"AcmeStore_NL","feeIds":["f1","f2"]},
				{"itemYrn":"` + paymentYRN + `","siteCode":"AcmeStore_NL","feeIds":[]}]`,
			want: assignmentHeader +
				productYRN + "  if1  AcmeStore_NL  f1,f2\n" +
				paymentYRN + "         AcmeStore_NL  \n",
		},
		{
			// Each --site adds one site; the pages are in the body, not in the query.
			args: []string{
				"fees", "assignments", "search-by-product", "--product", "100200", "--site", "AcmeStore_NL",
				"--site", "AcmeStore_BE",
			},
			path:   "/fee/acme1/itemFees/searchByProductId",
			mode:   getAllBodyPaged,
			method: http.MethodPost,
			sent: `{"pageNumber":{page},"pageSize":100,"productId":"100200",` +
				`"siteCodes":["AcmeStore_NL","AcmeStore_BE"]}`,
			body: `[{"id":"if1","itemYrn":"` + productYRN + `","siteCode":"AcmeStore_NL","feeIds":["f1"],
				"fees":[{"id":"f1","code":"deposit"}]}]`,
			want: assignmentHeader + productYRN + "  if1  AcmeStore_NL  f1\n",
		},
		{
			args: []string{
				"fees", "assignments", "search-by-products", "--products", "100200,100201", "--site", "AcmeStore_NL",
			},
			path:   "/fee/acme1/itemFees/searchByProductIds",
			mode:   getAllBodyPaged,
			method: http.MethodPost,
			sent:   `{"pageNumber":{page},"pageSize":100,"productIds":"100200,100201","siteCode":"AcmeStore_NL"}`,
			body:   feeAssignments,
			want:   assignmentTable,
		},
	}
}

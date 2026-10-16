package cli

import (
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

// invoiceJob is the part of an Invoice Service job shown in tables. Emporix
// sends the job's status as status and its orders as invoices, each with its
// own status (the reference's schema names them jobStatus, orders and
// orderStatus).
type invoiceJob struct {
	Status   string `json:"status"`
	JobType  string `json:"jobType"`
	Invoices []struct {
		Status string `json:"status"`
	} `json:"invoices"`
}

// count returns how many of the job's orders have status, as text.
func (j *invoiceJob) count(status string) string {
	n := 0
	for _, o := range j.Invoices {
		if o.Status == status {
			n++
		}
	}
	return strconv.Itoa(n)
}

// newInvoiceJobsCommand creates and reads invoice jobs, which generate the
// invoices of orders. Emporix lists no jobs, so a job is read by the id its
// creation returned. The reference declares no scope for either, but Emporix
// refuses a token without scopes (403 Forbidden). IAM knows two invoice
// scopes: invoice.invoice_view, the read scope, with which Emporix gets past
// authorization (an unknown job then answers 500), and
// invoice.invoice_manage, which can change data and so is the one that
// creating a job requests.
func newInvoiceJobsCommand(a *app) *cobra.Command {
	jobs := reader[invoiceJob]{view: view[invoiceJob]{
		singular: "invoice job",
		header:   []string{columnStatus, columnType, "ORDERS", "INVOICED", columnFailed, "PENDING"},
		row: func(j invoiceJob, _ language) []string {
			return []string{
				j.Status, j.JobType, strconv.Itoa(len(j.Invoices)), j.count("SUCCESS"), j.count("FAILURE"),
				j.count("PENDING"),
			}
		},
	}, scopes: []string{"invoice.invoice_view"}}
	return newGroup("invoice-jobs", "Create and read invoice jobs",
		"Creates and reads the jobs that generate the invoices of orders.",
		invoiceJobCreate().command(a),
		jobs.command(a, &command{
			use:   "get ID",
			short: "Show the status of an invoice job",
			long: "Reads an invoice job: STATUS is IN_PROGRESS or DONE and TYPE MANUAL (orders chosen when\n" +
				"the job was created) or AUTOMATIC (orders that the tenant's configuration selects). ORDERS\n" +
				"counts the orders the job processed, of which INVOICED got an invoice, FAILED did not and\n" +
				"PENDING wait for one. Each order's entry in invoices (orderId and status; errorStep and\n" +
				"message for a failure, e.g. an order not ready for invoicing; for a success, the reference\n" +
				"documents the invoice number and download link) is only in -o json.\n" +
				"ID is the job id that creating the job returned. Emporix answers an unknown job with\n" +
				"500 Internal Server Error (the reference documents 404), so get fails with exit code 1.",
			req:   request{path: "/invoice/{tenant}/jobs/invoices/" + placeholderID},
			fetch: single{items: oneItem},
		}))
}

// invoiceJobCreate starts an invoice job: MANUAL, for the orders that --order
// names, or AUTOMATIC, for those that the tenant's configuration selects.
func invoiceJobCreate() *write {
	return &write{
		use:   createUse,
		short: "Start an invoice job",
		long: "Starts a job that generates the invoices of orders: of those that --order names\n" +
			"(--type MANUAL, which requires one at least), or of those ready for an invoice that the\n" +
			"tenant's invoice configuration selects (--type AUTOMATIC, which takes no --order).\n" +
			"Without an invoice configuration of the tenant's own, an order is ready for an invoice when\n" +
			"its mixins.generalAttributes.extendedOrderStatus is 70, 75 or 76 (`epx orders create`\n" +
			"stores 10; `epx orders patch` changes it); the job's entry of an order with 10 fails as not\n" +
			"ready. Invoicing a ready order calls the tenant's Pick-Pack service: where that service is\n" +
			"not reachable, the entry fails at the VALIDATION step.\n" +
			"Emporix answers with the job's id (jobId), which `epx invoice-jobs get ID` reads.",
		example: "  epx invoice-jobs create --type MANUAL --order O1001 --order O1002\n" +
			"  epx invoice-jobs create --type AUTOMATIC --yes",
		method: http.MethodPost,
		req:    request{path: "/invoice/{tenant}/jobs/invoices"},
		scopes: []string{"invoice.invoice_manage"},
		body: &payload{typed: objectBody{
			choiceField("jobType", typeParam, "the `TYPE` of job", "MANUAL", "AUTOMATIC").require(),
			listField("orderIds", "order", "the `ID` of an order to invoice, repeatable").optional().
				onlyWithChoice(typeParam, "MANUAL"),
		}},
	}
}

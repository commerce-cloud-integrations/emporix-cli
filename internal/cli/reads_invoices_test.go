package cli

// invoiceJobReads returns the Invoice Service reads.
func invoiceJobReads() []readCase {
	return []readCase{
		{
			// The orders are counted by their status: the job and a failed entry in the shape
			// Emporix sends, the SUCCESS entry as the reference documents it (V77).
			args:   []string{"invoice-jobs", "get", "j/1"},
			path:   "/invoice/acme1/jobs/invoices/j%2F1",
			scopes: []string{"invoice.invoice_view"},
			mode:   getWhole,
			body: `{"status":"DONE","jobType":"MANUAL","invoices":[
				{"orderId":"O1","status":"FAILURE","errorStep":"VALIDATION","message":"Order O1 has not ready invoice status"},
				{"orderId":"O2","status":"SUCCESS"}]}`,
			want: "STATUS  TYPE    ORDERS  INVOICED  FAILED  PENDING\n" +
				"DONE    MANUAL  2       1         1       0\n",
		},
		{
			// A running job: the type is optional in the table.
			args:   []string{"invoice-jobs", "get", "j2"},
			path:   "/invoice/acme1/jobs/invoices/j2",
			scopes: []string{"invoice.invoice_view"},
			mode:   getWhole,
			body:   `{"status":"IN_PROGRESS","invoices":[{"orderId":"O4","status":"PENDING"}]}`,
			want: "STATUS       TYPE  ORDERS  INVOICED  FAILED  PENDING\n" +
				"IN_PROGRESS        1       0         0       1\n",
		},
	}
}

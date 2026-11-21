package cli

func invoiceJobReads() []readCase {
	return []readCase{
		{
			args:   []string{"invoice-jobs", "get", "j/1"},
			path:   "/invoice/acme1/jobs/invoices/j%2F1",
			scopes: []string{"invoice.invoice_view"},
			mode:   getWhole,
			response: `{"status":"DONE","jobType":"MANUAL","invoices":[
				{"orderId":"O1","status":"FAILURE","errorStep":"VALIDATION","message":"Order O1 has not ready invoice status"},
				{"orderId":"O2","status":"SUCCESS"}]}`,
			want: "STATUS  TYPE    ORDERS  INVOICED  FAILED  PENDING\n" +
				"DONE    MANUAL  2       1         1       0\n",
		},
		{
			args:     []string{"invoice-jobs", "get", "j2"},
			path:     "/invoice/acme1/jobs/invoices/j2",
			scopes:   []string{"invoice.invoice_view"},
			mode:     getWhole,
			response: `{"status":"IN_PROGRESS","invoices":[{"orderId":"O4","status":"PENDING"}]}`,
			want: "STATUS       TYPE  ORDERS  INVOICED  FAILED  PENDING\n" +
				"IN_PROGRESS        1       0         0       1\n",
		},
	}
}

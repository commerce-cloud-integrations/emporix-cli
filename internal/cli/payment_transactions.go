package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// paymentTransaction is the part of a Payment Gateway transaction shown in
// tables; its payment details and the provider's requests are left to -o json.
type paymentTransaction struct {
	ID                 string `json:"id"`
	Provider           string `json:"provider"`
	PaymentInformation struct {
		PaymentMode string `json:"paymentMode"`
	} `json:"paymentInformation"`
	Entries []transactionEntry `json:"entries"`
}

// transactionEntry is one operation of a transaction, such as a capture.
type transactionEntry struct {
	Type       string `json:"type"`
	Successful *bool  `json:"successful"`
	Amount     money  `json:"amount"`
}

// String returns the type and amount of e, marked when it failed.
func (e transactionEntry) String() string {
	failed := ""
	if e.Successful != nil && !*e.Successful {
		failed = "failed"
	}
	return spaced(e.Type, e.Amount.String(), failed)
}

// newPaymentTransactionsCommand reads the payment transactions that payment
// providers ran through the Payment Gateway, and authorizes, captures,
// refunds and cancels payments.
func newPaymentTransactionsCommand(a *app) *cobra.Command {
	transactions := resource[paymentTransaction]{
		use:   "payment-transactions",
		short: "Read payment transactions; authorize, capture, refund and cancel payments",
		about: "Reads payment transactions, personal data included, and runs payments through the\n" +
			"tenant's payment providers: authorize starts a transaction for an order, capture collects\n" +
			"an authorized amount, refund returns a captured one and cancel voids an authorization.\n" +
			"Capture, refund and cancel move money.",
		long: "Reads payment transactions, personal data included. PAYMENT MODE is the id of the payment\n" +
			"mode used, and ENTRIES the operations the provider ran, in the order sent: type and amount,\n" +
			"marked failed when they failed. Payment details (such as bank details and pay-by-link URLs),\n" +
			"card tokens and the provider's requests and responses are only in -o json. list's --sort\n" +
			"takes only createdDate (e.g. createdDate:desc); Emporix rejects other and several fields.",
		path:    "/payment-gateway/{tenant}/transactions",
		scopes:  []string{"payment-gateway.paymenttransactions_read"},
		filters: listFilters{kinds: sortable, sortHelp: paymentTransactionsSortUsage},
		view: view[paymentTransaction]{
			singular: "payment transaction",
			header:   []string{columnID, columnProvider, "PAYMENT MODE", "ENTRIES"},
			row: func(p paymentTransaction, _ language) []string {
				return []string{
					p.ID, p.Provider, p.PaymentInformation.PaymentMode, joinEach(p.Entries, ", ", transactionEntry.String),
				}
			},
		},
	}
	return newResourceCommand(a, transactions, paymentWrites(a, transactions)...)
}

// paymentTransactionsSortUsage is the help of list's --sort: Emporix refuses
// any field but createdDate, and several fields.
const paymentTransactionsSortUsage = "Emporix sort `order`, passed as given: createdDate only, e.g. " +
	"createdDate:desc; another field, or several, is refused (400, exit 1)"

// paymentWrites returns the payment operations of the Payment Gateway, each
// with the one scope that the reference names for it: authorize, and
// capture, refund and cancel of a transaction. Emporix answers each with
// whether it succeeded (answers.ReportsOutcome).
func paymentWrites(a *app, transactions resource[paymentTransaction]) []*cobra.Command {
	const (
		payments    = "/payment-gateway/{tenant}/payment"
		transaction = payments + "/" + placeholderID
		// moves is the second paragraph of the help of capture, refund and cancel.
		moves = "This moves money: the confirmation names the transaction in the URL that it shows. ID is\n" +
			"the ID that `epx payment-transactions list` shows; an unknown one is not found (exit 4)."
	)
	amount := func(verb string) []bodyField {
		return requireAll([]bodyField{
			numberField("amount", "amount", "the `AMOUNT` to "+verb+", e.g. 3.99"),
			stringField("currency", "currency", "the `CURRENCY` of --amount, e.g. EUR"),
		})
	}
	return []*cobra.Command{
		(&write{
			use: "authorize", short: "Authorize a payment", method: http.MethodPost,
			long: "Authorizes the payment of an order with a payment mode, as --file, a JSON object, gives\n" +
				"them: order.id, paymentModeId, amountToAuthorize and currency, and what the provider\n" +
				"needs, such as creditCardToken. The provider reserves the amount, which capture then\n" +
				"collects. Emporix answers with the paymentTransactionId and, where the customer must\n" +
				"finish the payment with the provider, where to send them (externalPaymentRedirectURL).",
			example: "  epx payment-transactions authorize --file authorize.json --dry-run",
			req:     request{path: payments + "/authorize"}, scopes: []string{"payment-gateway.authorize_manage"},
			body: &payload{}, answers: emporix.Answers{ReportsOutcome: true},
		}).command(a),
		(&write{
			use: transactions.itemUse("capture"), short: "Capture an authorized payment", method: http.MethodPost,
			long: "Captures --amount in --currency of the authorized payment of transaction ID, at most the\n" +
				"amount authorized: the provider collects it from the customer. Emporix answers with the\n" +
				"provider's captureId, which refund --capture-id takes.\n\n" + moves,
			example: "  epx payment-transactions capture T1 --amount 3.99 --currency EUR --dry-run",
			req:     request{path: transaction + "/capture"}, scopes: []string{"payment-gateway.capture_manage"},
			body: &payload{typed: objectBody(amount("capture"))}, ids: transactions.itemIDs(),
			answers: emporix.Answers{ReportsOutcome: true},
		}).command(a),
		(&write{
			use: transactions.itemUse("refund"), short: "Refund a captured payment", method: http.MethodPost,
			long: "Refunds --amount in --currency of the captured payment of transaction ID, at most the\n" +
				"amount captured: the provider returns it to the customer. Where the payment was captured\n" +
				"more than once, --capture-id names the capture to refund.\n\n" + moves,
			example: "  epx payment-transactions refund T1 --amount 3.99 --currency EUR --dry-run",
			req:     request{path: transaction + "/refund"}, scopes: []string{"payment-gateway.refund_manage"},
			body: &payload{typed: objectBody(append(amount("refund"),
				stringField("captureId", "capture-id", "the `ID` of the capture to refund, as capture answered it")))},
			answers: emporix.Answers{ReportsOutcome: true}, ids: transactions.itemIDs(),
		}).command(a),
		(&write{
			use: transactions.itemUse("cancel"), short: "Cancel an authorized payment", method: http.MethodPost,
			long: "Cancels the authorized payment of transaction ID: the provider releases the amount that\n" +
				"it reserved.\n\n" + moves,
			example: "  epx payment-transactions cancel T1 --dry-run",
			req:     request{path: transaction + "/cancel"}, scopes: []string{"payment-gateway.cancel_manage"},
			body: noBody{}, answers: emporix.Answers{ReportsOutcome: true}, ids: transactions.itemIDs(),
		}).command(a),
	}
}

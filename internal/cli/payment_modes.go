package cli

import (
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// paymentMode is the part of a Payment Gateway payment mode shown in tables;
// its provider configuration holds credentials and is left to -o json.
type paymentMode struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Provider string `json:"provider"`
	Active   *bool  `json:"active"`
}

// publicPaymentMode is the part of a payment mode as a storefront sees it.
type publicPaymentMode struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	IntegrationType   string `json:"integrationType"`
	PaymentMethodType string `json:"paymentMethodType"`
}

// newPaymentModesCommand reads the tenant's payment modes, which the service
// returns whole, and the view a storefront gets of them.
func newPaymentModesCommand(a *app) *cobra.Command {
	storefront := resource[publicPaymentMode]{
		use:  "public",
		noun: "payment modes as a storefront sees them",
		long: "Reads payment modes as a storefront sees them, without provider secrets. INTEGRATION is\n" +
			"OFFSITE, EXTERNAL, SEPA or TOKENIZED and METHOD the payment method type, e.g. paypal; public\n" +
			"keys, tokens and script URLs are only in -o json.",
		path: "/payment-gateway/{tenant}/paymentmodes/frontend",
		read: (*emporix.Client).GetUnpaged,
		item: oneItemOrArray,
		view: view[publicPaymentMode]{
			singular: "storefront payment mode",
			header:   []string{columnID, columnCode, "INTEGRATION", columnMethod},
			row: func(m publicPaymentMode, _ language) []string {
				return []string{m.ID, m.Code, m.IntegrationType, m.PaymentMethodType}
			},
		},
	}
	modes := resource[paymentMode]{
		use:   "payment-modes",
		short: "Read, create, update and delete payment modes",
		about: "Reads the tenant's payment modes, and creates, updates and deletes them (create and\n" +
			"update send the JSON of --file). Provider configurations hold credentials, which the reads\n" +
			"show only in -o json and the writes send and print as given. `epx payment-modes public`\n" +
			"shows what a storefront sees and needs no scope.",
		long: "Reads the tenant's payment modes. PROVIDER is INVOICE, CASH_ON_DELIVERY, SPREEDLY,\n" +
			"SPREEDLY_SAFERPAY or UNZER. The provider configuration, credentials included (such as\n" +
			"Spreedly/AccessSecret or Saferpay/BasicAuth), is only in -o json. `epx payment-modes public`\n" +
			"shows what a storefront sees and needs no scope.",
		path:   "/payment-gateway/{tenant}/paymentmodes/config",
		scopes: []string{"payment-gateway.paymentmodes_read"},
		read:   (*emporix.Client).GetUnpaged,
		view: view[paymentMode]{
			singular: "payment mode",
			header:   []string{columnID, columnCode, columnProvider, columnActive},
			row: func(m paymentMode, _ language) []string {
				return []string{m.ID, m.Code, m.Provider, formatBool(m.Active)}
			},
		},
	}
	return newResourceCommand(a, modes,
		slices.Concat([]*cobra.Command{newResourceCommand(a, storefront)}, paymentModeWrites(a, modes))...)
}

// paymentModeWrites returns the writes of payment modes, with the scope that
// the reference names for them: create, update ID (which Emporix answers
// with 404 Not Found for an unknown mode) and delete ID.
func paymentModeWrites(a *app, modes resource[paymentMode]) []*cobra.Command {
	manage := []string{"payment-gateway.paymentmodes_manage"}
	// unknown is the second paragraph of the help of update and delete.
	const unknown = "\n\nID is the ID that `epx payment-modes list` shows; an unknown one is not found (exit 4)."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a payment mode", method: http.MethodPost,
			long: "Creates the payment mode in --file, a JSON object with its code and provider (both\n" +
				"required; PROVIDER as list shows it), active, and the provider's configuration: a map of\n" +
				"string values such as Spreedly/GatewayToken, credentials included. Emporix answers with\n" +
				"the payment mode, its configuration included, or, for a code that another mode has, with\n" +
				"400 Bad Request (exit 1); the reference documents 409 Conflict (exit 5) for that case instead.",
			example: "  epx payment-modes create --file invoice.json --dry-run",
			req:     request{path: modes.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: modes.itemUse("update"), short: "Update a payment mode", method: http.MethodPut,
			long: "Sets active and the configuration of the payment mode ID to those in --file, a JSON\n" +
				"object with active and configuration, the map of the provider's values that create takes;\n" +
				"the code and provider stay as created. Emporix answers with the payment mode, its\n" +
				"configuration included, and creates no missing mode." + unknown,
			req: request{path: modes.itemPath()}, scopes: manage, body: &payload{}, ids: modes.itemIDs(),
		}).command(a),
		itemDelete(modes, "Delete a payment mode",
			"Deletes the payment mode ID, which storefronts then no longer offer."+unknown, manage).command(a),
	}
}

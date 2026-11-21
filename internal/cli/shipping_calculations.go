package cli

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type shippingQuote struct {
	Zone struct {
		ID string `json:"id"`
	} `json:"zone"`
	Method struct {
		ID              string    `json:"id"`
		Name            localized `json:"name"`
		Fee             money     `json:"fee"`
		ShippingTaxCode string    `json:"shippingTaxCode"`
	} `json:"method"`
}

func (q *shippingQuote) zone() string {
	if q.Method.ID == "" {
		return ""
	}
	return q.Zone.ID
}

type shippingFeeQuote struct {
	Fee             money  `json:"fee"`
	ShippingTaxCode string `json:"shippingTaxCode"`
}

type shippingSite struct {
	ID    string             `json:"id"`
	Zones []shippingSiteZone `json:"zones"`
}

type shippingSiteZone struct {
	ID      string `json:"id"`
	Methods []struct {
		ID string `json:"id"`
	} `json:"methods"`
}

func (s *shippingSite) zones() string {
	return joinEach(s.Zones, ", ", func(z shippingSiteZone) string { return z.ID })
}

func (s *shippingSite) methods() string {
	var ids []string
	for _, zone := range s.Zones {
		for _, method := range zone.Methods {
			if !slices.Contains(ids, method.ID) {
				ids = append(ids, method.ID)
			}
		}
	}
	return strings.Join(ids, ", ")
}

const (
	columnFee     = "FEE"
	columnTaxCode = "TAX CODE"
)

const quoteSentence = "Emporix answers a destination or cart total that no method serves, or an " +
	"unknown site, with 400 Bad Request (exit 1)."

func newShippingQuoteCommand(a *app) *cobra.Command {
	quote := reader[shippingQuote]{view: view[shippingQuote]{
		singular:  "shipping quote",
		header:    []string{columnZone, columnMethod, columnName, columnFee, columnTaxCode},
		localized: true,
		row: func(q shippingQuote, lang language) []string {
			m := q.Method
			return []string{q.zone(), m.ID, m.Name.in(lang), m.Fee.String(), m.ShippingTaxCode}
		},
	}}
	return quote.command(a, &command{
		use:   "quote",
		short: "Calculate the shipping fee of each method to a destination",
		long: "Calculates what each shipping method of a site's zones that ship to a destination costs for " +
			"a cart total: one row per method, with its zone, its FEE and its shipping TAX CODE. With " +
			"--customer, the fees of the customer's shipping group apply (`epx shipping customers`). " +
			quoteSentence,
		req:   request{path: shippingPath("quote")},
		fetch: bodyRead{method: http.MethodPost, body: destinationQuoteBody(), fetch: single{items: quoteMethods}},
	})
}

func shippingFeeReader() reader[shippingFeeQuote] {
	return reader[shippingFeeQuote]{view: view[shippingFeeQuote]{
		singular: "shipping fee",
		header:   []string{columnFee, columnTaxCode},
		row: func(q shippingFeeQuote, _ language) []string {
			return []string{q.Fee.String(), q.ShippingTaxCode}
		},
	}}
}

func newShippingMinimumQuoteCommand(a *app) *cobra.Command {
	return shippingFeeReader().command(a, &command{
		use:   "quote-minimum",
		short: "Calculate the lowest shipping fee to a destination",
		long: "Calculates the lowest shipping fee of a site to a destination for a cart total, as a cart " +
			"shows it before a delivery window is chosen. TAX CODE is the method's shipping tax code. " +
			quoteSentence,
		req:   request{path: shippingPath("quote/minimum")},
		fetch: bodyRead{method: http.MethodPost, body: destinationQuoteBody(), fetch: single{items: oneItem}},
	})
}

func newShippingSlotQuoteCommand(a *app) *cobra.Command {
	slot := shippingFeeReader().command(a, &command{
		use:   "quote-slot",
		short: "Calculate the shipping fee of a delivery time slot",
		long: "Calculates the shipping fee of a delivery time slot for a cart total: --delivery-time and " +
			"--slot are the IDs that `epx delivery-times list` and `epx delivery-times slots list` show. " +
			"The destination is optional; --country and --postal-code go together. An unknown delivery " +
			"time is not found (exit 4).",
		req: request{path: shippingPath("quote/slot")},
		fetch: bodyRead{method: http.MethodPost, body: objectBody(slices.Concat(cartFields(), destinationFields(),
			deliverySlotFields(deliveryWindowMember))), fetch: single{items: oneItem}},
	})
	slot.MarkFlagsRequiredTogether(countryFlag, postalCodeFlag)
	return slot
}

func newFindSitesCommand(a *app) *cobra.Command {
	sites := reader[shippingSite]{view: view[shippingSite]{
		singular: "shipping site",
		header:   []string{columnSite, "ZONES", "METHODS"},
		row: func(s shippingSite, _ language) []string {
			return []string{s.ID, s.zones(), s.methods()}
		},
	}}
	return sites.command(a, &command{
		use:   "find-sites",
		short: "List the sites that ship to a postal code",
		long: "Lists the sites with a shipping zone that ships to a postal code: ZONES lists the zones that " +
			"Emporix returns with each site, and METHODS their shipping methods (`-o json` adds the " +
			"zones' destinations and fees). Emporix matched no site without --country on the tenant " +
			"tested.",
		req: request{path: "/shipping/{tenant}/findSite"},
		fetch: bodyRead{method: http.MethodPost, body: objectBody{
			stringField("postalCode", postalCodeFlag, "the `POSTAL_CODE` to ship to").require(),
			stringField("country", countryFlag, "the `COUNTRY` to ship to, a country code such as DE"),
		}, fetch: listing{read: (*emporix.Client).GetUnpaged}},
	})
}

const (
	countryFlag    = "country"
	postalCodeFlag = "postal-code"
)

func cartFields() []bodyField {
	return []bodyField{
		moneyField("cartTotal", "cart-total", "the cart's total as `AMOUNT CURRENCY`, e.g. \"100 EUR\"").require(),
		stringField("customerId", "customer", "the `CUSTOMER` id, whose shipping group's fees apply"),
	}
}

func destinationQuoteBody() objectBody {
	return objectBody(slices.Concat(cartFields(), requireAll(destinationFields())))
}

func destinationFields() []bodyField {
	return []bodyField{
		stringField("shipToAddress.country", countryFlag, "the destination `COUNTRY`, a country code such as DE"),
		stringField("shipToAddress.zipCode", postalCodeFlag, "the destination `POSTAL_CODE`"),
	}
}

func quoteMethods(body json.RawMessage) ([]json.RawMessage, error) {
	zones, err := listItems(body)
	if err != nil {
		return nil, err
	}
	var items []json.RawMessage
	for _, raw := range zones {
		zone, err := jsonObject(raw)
		var methods []json.RawMessage
		if err != nil || (zone["methods"] != nil && json.Unmarshal(zone["methods"], &methods) != nil) {
			return nil, errors.New("expected zones as JSON objects")
		}
		for _, method := range methods {
			if _, err := jsonObject(method); err != nil {
				return nil, errors.New("expected shipping methods as JSON objects")
			}
			item, err := json.Marshal(map[string]json.RawMessage{"zone": zone["zone"], "method": method})
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	}
	return items, nil
}

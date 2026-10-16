package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// exchangeRate is the part of a Currency Service exchange rate shown in tables.
type exchangeRate struct {
	Code           string   `json:"code"`
	SourceCurrency string   `json:"sourceCurrency"`
	TargetCurrency string   `json:"targetCurrency"`
	Rate           rateText `json:"rate"`
}

// rateText is an exchange rate kept as sent. The API reference types it as a
// string, with a decimal comma in some descriptions and a dot in its examples,
// so any string is accepted as well as a JSON number.
type rateText string

// UnmarshalJSON implements json.Unmarshaler.
func (r *rateText) UnmarshalJSON(data []byte) error {
	var text string
	if json.Unmarshal(data, &text) == nil {
		*r = rateText(text)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("rate: %w", err)
	}
	*r = rateText(number)
	return nil
}

// currencyManageScope is the scope of every write of the Currency Service:
// currencies and exchange rates.
const currencyManageScope = "currency.currency_manage"

// newCurrenciesCommand reads, creates, changes and deletes the tenant's
// currencies.
func newCurrenciesCommand(a *app) *cobra.Command {
	currencies := resource[codeName]{
		use:     "currencies",
		short:   "Read, create, change and delete currencies",
		long:    "Reads the tenant's currencies.",
		path:    "/currency/{tenant}/currencies",
		filters: listFilters{kinds: sortable},
		view:    codeNameView("currency"),
	}
	return newResourceCommand(a, currencies, currencyWrites(a, currencies)...)
}

// currencyWrites returns the writes of currencies, with the scope that the
// reference names for them: create, update CODE and delete CODE.
func currencyWrites(a *app, currencies resource[codeName]) []*cobra.Command {
	manage := []string{currencyManageScope}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a currency", method: http.MethodPost,
			long: "Creates a currency from the JSON in --file: its code (ISO 4217) and name, e.g.\n" +
				`{"code": "CHF", "name": {"en": "Swiss franc", "de": "Schweizer Franken"}}, in languages` + "\n" +
				"that the tenant has. Emporix answers with the code; a code that exists already is a\n" +
				"conflict (exit 5).",
			req: request{path: currencies.path}, scopes: manage, body: &payload{}, language: stringsOrMaps,
		}).command(a),
		(&write{
			use: currencies.itemUse("update"), short: "Rename a currency", method: http.MethodPut,
			long: "Replaces the name of the currency CODE with the one in the JSON of --file, which also\n" +
				"needs the currency's metadata.version as read (`epx currencies get CODE -o json`), e.g.\n" +
				`{"name": {"en": "Swiss franc"}, "metadata": {"version": 1}}; the code cannot change. A` + "\n" +
				"stale version and an unknown CODE are both not found (exit 4): Emporix answers 404 to both.",
			req: request{path: currencies.itemPath()}, scopes: manage, body: &payload{}, language: stringsOrMaps,
			ids: currencies.itemIDs(),
		}).command(a),
		(&write{
			use: currencies.itemUse("delete"), short: "Delete a currency", method: http.MethodDelete,
			long: "Deletes the currency CODE; it sends no body. The reference says Emporix also deletes the\n" +
				"prices in the currency, asynchronously (not checked here). The exchange rates stay, and a site\n" +
				"that lists the currency keeps listing it. It sends no forceDelete, which the reference documents\n" +
				"(default false) without saying what it changes." +
				unknownAsDeleted("CODE"),
			req: request{path: currencies.itemPath()}, scopes: manage, body: noBody{}, ids: currencies.itemIDs(),
		}).command(a),
	}
}

// newExchangeRatesCommand reads, creates, changes and deletes currency
// exchange rates, whose small bodies typed flags make (or --file gives).
func newExchangeRatesCommand(a *app) *cobra.Command {
	rates := resource[exchangeRate]{
		use:     "exchange-rates",
		short:   "Read, create, change and delete currency exchange rates",
		long:    "Reads currency exchange rates. CODE names a rate, e.g. EUR_CHF.",
		path:    "/currency/{tenant}/exchanges",
		filters: listFilters{kinds: sortable},
		scopes:  []string{"currency.currency_read"},
		view: view[exchangeRate]{
			singular: "exchange rate",
			header:   []string{columnCode, "SOURCE", "TARGET", "RATE"},
			row: func(r exchangeRate, _ language) []string {
				return []string{r.Code, r.SourceCurrency, r.TargetCurrency, string(r.Rate)}
			},
		},
	}
	return newResourceCommand(a, rates, exchangeRateWrites(a, rates)...)
}

// exchangeRateWrites returns the writes of exchange rates, with the scope
// that the reference names for them: create, update CODE and delete CODE,
// whose small bodies typed flags make (or --file gives).
func exchangeRateWrites(a *app, rates resource[exchangeRate]) []*cobra.Command {
	manage := []string{currencyManageScope}
	rate := stringField("rate", "rate", "the `RATE`: what one unit of the source currency is worth in the "+
		"target currency, e.g. 0.95 (sent as a string)").require()
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create an exchange rate", method: http.MethodPost,
			long: "Creates the exchange rate from the currency --source into --target (both defined in the\n" +
				"tenant, as `epx currencies list` shows them), or the one in the JSON of --file. Emporix\n" +
				"answers with its code, SOURCE_TARGET; a rate that exists already is a conflict (exit 5).",
			example: "  epx exchange-rates create --source EUR --target CHF --rate 0.95",
			req:     request{path: rates.path}, scopes: manage, body: &payload{typed: objectBody{
				stringField("sourceCurrency", "source", "the `CODE` of the currency to convert from").require(),
				stringField("targetCurrency", "target", "the `CODE` of the currency to convert into").require(),
				rate,
			}},
		}).command(a),
		(&write{
			use: rates.itemUse("update"), short: "Change an exchange rate", method: http.MethodPut,
			long: "Changes the rate of the exchange rate CODE (its currencies cannot change) to --rate, with\n" +
				"its metadata.version as read (--version, as `epx exchange-rates get CODE -o json` prints\n" +
				"it), or to the JSON of --file. A stale version and an unknown CODE are both not found (exit 4):\n" +
				"Emporix answers 404 to both.",
			example: "  epx exchange-rates update EUR_CHF --rate 0.96 --version 2",
			ids:     rates.itemIDs(),
			req:     request{path: rates.itemPath()}, scopes: manage, body: &payload{typed: objectBody{
				rate,
				versionField("rate").require(),
			}},
		}).command(a),
		(&write{
			use: rates.itemUse("delete"), short: "Delete an exchange rate", method: http.MethodDelete,
			long: "Deletes the exchange rate CODE; it sends no body." + unknownAsDeleted("CODE"),
			req:  request{path: rates.itemPath()}, scopes: manage, body: noBody{}, ids: rates.itemIDs(),
		}).command(a),
	}
}

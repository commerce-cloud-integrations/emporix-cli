package cli

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/spf13/cobra"
)

// taxConfiguration is the part of a Tax Service configuration shown in tables.
// locationCode is optional in the spec, whose examples identify a location by
// its country code only.
type taxConfiguration struct {
	LocationCode string `json:"locationCode"`
	Location     struct {
		CountryCode string `json:"countryCode"`
	} `json:"location"`
	TaxClasses taxClasses `json:"taxClasses"`
}

// taxClasses are the tax classes of a location, in the order Emporix sends them.
type taxClasses []taxClass

// taxClass is a tax class of a location.
type taxClass struct {
	Code      string      `json:"code"`
	Rate      json.Number `json:"rate"` // percent, kept as sent
	IsDefault bool        `json:"isDefault"`
}

// String shows the class with its rate and marks the default class.
func (c taxClass) String() string {
	text := c.Code + " " + c.Rate.String()
	if c.IsDefault {
		text += " (default)"
	}
	return text
}

// String joins the classes.
func (t taxClasses) String() string { return joinEach(t, ", ", taxClass.String) }

// location returns the location code, or the country code when it is missing.
func (t taxConfiguration) location() string { return cmp.Or(t.LocationCode, t.Location.CountryCode) }

// newTaxesCommand reads, creates, changes and deletes tax configurations, one
// per location, and calculates prices with them.
func newTaxesCommand(a *app) *cobra.Command {
	taxes := resource[taxConfiguration]{
		use:   "taxes",
		short: "Read and change tax configurations, and calculate net and gross prices",
		long: "Reads tax configurations, one per location. CLASSES lists each tax class with its rate\n" +
			"in percent (-o json adds their names). `epx taxes calculate` applies them to a price.",
		path:    "/tax/{tenant}/taxes",
		filters: listFilters{kinds: sortable},
		scopes:  []string{"tax.tax_read"},
		view: view[taxConfiguration]{
			singular: "tax configuration",
			header:   []string{"LOCATION", columnCountry, "CLASSES"},
			row: func(t taxConfiguration, _ language) []string {
				return []string{t.location(), t.Location.CountryCode, t.TaxClasses.String()}
			},
		},
	}
	return newResourceCommand(a, taxes, slices.Concat([]*cobra.Command{newTaxCalculationCommand(a)},
		taxWrites(a, taxes))...)
}

// taxCalculationSegment is the path segment, below the tax configurations,
// of the tax calculation, which a PUT sends (calculate).
const taxCalculationSegment = "calculation-commands"

// taxWrites returns the writes of tax configurations, with the scope that the
// reference names for them: create, update LOCATION (which creates no missing
// one) and delete LOCATION. update refuses the LOCATION calculation-commands,
// whose PUT is the calculation that calculate sends.
func taxWrites(a *app, taxes resource[taxConfiguration]) []*cobra.Command {
	manage := []string{"tax.tax_manage"}
	item := request{path: taxes.itemPath()}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a tax configuration", method: http.MethodPost,
			long: "Creates the tax configuration of a country from the JSON in --file: location.countryCode\n" +
				"and taxClasses, each with its code, name, order, rate in percent and, for one, isDefault,\n" +
				`e.g. {"location": {"countryCode": "NL"}, "taxClasses": [{"code": "STANDARD", "name":` + "\n" +
				`{"en": "Standard"}, "order": 0, "rate": 21, "isDefault": true}]}. A country that has one` + "\n" +
				"already is a conflict (exit 5).",
			req: request{path: taxes.path}, scopes: manage, body: &payload{}, language: stringsOrMaps,
		}).command(a),
		(&write{
			use: taxes.itemUse("update"), short: "Replace a tax configuration", method: http.MethodPut,
			long: "Replaces the tax configuration of LOCATION with the JSON in --file, shaped as\n" +
				"`epx taxes get LOCATION -o json` prints it, with its metadata.version as read: a stale\n" +
				"version and an unknown LOCATION are both not found (exit 4): Emporix answers 404 to both. The\n" +
				"country code cannot change.",
			req: item, scopes: manage, body: &payload{}, language: stringsOrMaps, ids: taxes.itemIDs(),
			reserved: map[string]string{taxCalculationSegment: "calculate"},
		}).command(a),
		(&write{
			use: taxes.itemUse("delete"), short: "Delete a tax configuration", method: http.MethodDelete,
			long: "Deletes the tax configuration of LOCATION; it sends no body.",
			req:  item, scopes: manage, body: noBody{}, ids: taxes.itemIDs(),
		}).command(a),
	}
}

// taxCalculation is the part of a tax calculation shown in tables.
type taxCalculation struct {
	Output struct {
		NetPrice      json.Number `json:"netPrice"`
		GrossPrice    json.Number `json:"grossPrice"`
		SourceTaxRate json.Number `json:"sourceTaxRate"`
		TargetTaxRate json.Number `json:"targetTaxRate"`
	} `json:"output"`
}

// The flags of a tax calculation that name a gross price and the tax it
// includes: the source is given whole, and only for a gross price.
const (
	grossFlag          = "gross"
	sourceCountryFlag  = "source-country"
	sourceTaxClassFlag = "source-tax-class"
)

// newTaxCalculationCommand returns `calculate`, the net and gross value of a
// price: a calculation that Emporix serves as PUT without changing data,
// with no scope (live; the reference names none).
func newTaxCalculationCommand(a *app) *cobra.Command {
	calculation := reader[taxCalculation]{view: view[taxCalculation]{
		singular: "tax calculation",
		header:   []string{"NET", "GROSS", "SOURCE RATE", "TARGET RATE"},
		row: func(c taxCalculation, _ language) []string {
			out := c.Output
			return []string{trimAmount(out.NetPrice.String()), trimAmount(out.GrossPrice.String()), out.SourceTaxRate.String(),
				out.TargetTaxRate.String()}
		},
	}}
	calculate := calculation.command(a, &command{
		use:   "calculate",
		short: "Calculate the net and gross value of a price",
		long: "Calculates the net and gross value of a price with the tax configurations that\n" +
			"`epx taxes list` shows: a net --price gets the rate of --tax-class in --country; with\n" +
			"--gross, a gross price is reduced by that rate, or, with --source-country and\n" +
			"--source-tax-class, by theirs before the rate of --country is added. SOURCE RATE and\n" +
			"TARGET RATE are the rates applied, in percent. A country without a tax configuration is\n" +
			"not found (exit 4).",
		req: request{path: "/tax/{tenant}/taxes/" + taxCalculationSegment},
		fetch: bodyRead{method: http.MethodPut, body: objectBody{
			numberField("input.price", "price", "the `PRICE` to calculate from, e.g. 12.99").require(),
			stringField("input.targetLocation.countryCode", "country", "the `COUNTRY` code to calculate for").require(),
			stringField("input.targetTaxClass", "tax-class", "the tax `CLASS` to apply, e.g. STANDARD").require(),
			boolField("input.includesTax", grossFlag, "the price is gross, tax included (default: net)"),
			stringField("input.sourceLocation.countryCode", sourceCountryFlag, "the `COUNTRY` code whose tax the "+
				"gross price includes").onlyWith(grossFlag),
			stringField("input.sourceTaxClass", sourceTaxClassFlag, "the tax `CLASS` that the gross price "+
				"includes").onlyWith(grossFlag),
		}, fetch: single{items: oneItem}},
	})
	calculate.MarkFlagsRequiredTogether(sourceCountryFlag, sourceTaxClassFlag)
	return calculate
}

package cli

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// unit is the part of a Unit Handling Service unit shown in tables.
type unit struct {
	Code     string      `json:"code"`
	Name     localized   `json:"name"`
	Type     string      `json:"type"`
	Symbol   string      `json:"symbol"`
	BaseUnit *bool       `json:"baseUnit"`
	Factor   json.Number `json:"factor"` // kept as sent, e.g. 1.0
}

// newUnitsCommand reads, creates, updates and deletes the tenant's units of
// measurement, and reads the unit types.
func newUnitsCommand(a *app) *cobra.Command {
	units := resource[unit]{
		use:   "units",
		short: "Read and change units of measurement; read unit types and convert between units",
		long: "Reads units of measurement. TYPE groups units that convert into each other, BASE marks\n" +
			"the base unit of a type, and FACTOR converts a unit into its type's base unit, as sent by\n" +
			"Emporix (g: 0.001 when kg is the base). `epx units types` lists the types;\n" +
			"`epx units convert` and `conversion-factor` convert between units of a type.\n" +
			"`epx units create`, `update`, `delete` and `bulk-delete` change the units.",
		path:    "/unit-handling/{tenant}/units",
		filters: listFilters{kinds: sortable},
		view: view[unit]{
			singular:  "unit",
			header:    []string{columnCode, columnName, columnType, "SYMBOL", "BASE", "FACTOR"},
			localized: true,
			row: func(u unit, lang language) []string {
				return []string{u.Code, u.Name.in(lang), u.Type, u.Symbol, formatBool(u.BaseUnit), u.Factor.String()}
			},
		},
	}
	types := typeNames("unit type").command(a, &command{
		use:   typesUse,
		short: "List unit types",
		long:  "Reads the unit types: the values that the TYPE of a unit takes.",
		req:   request{path: "/unit-handling/{tenant}/types"},
		fetch: listing{read: (*emporix.Client).GetUnpaged},
	})
	return newResourceCommand(a, units, slices.Concat([]*cobra.Command{types, newUnitConvertCommand(a),
		newUnitConversionFactorCommand(a)}, unitWriteCommands(a, units))...)
}

// unitManageScope is the scope of every write of units, the one that the
// reference names for them.
const unitManageScope = "unithandling.unit_manage"

// unitLanguageHelp is what the help of units update says of the localized
// text of its body: Emporix replaces a name with a string in one language,
// so that the unit's other translations are gone.
const unitLanguageHelp = contentLanguageHelp + "\nSuch a string replaces the unit's other translations of " +
	"that text; send the map to keep them."

// unitWriteCommands returns the writes of units, with the scope that the
// reference names for them: `create` (POST to the collection), `update CODE`
// (PUT to the unit), which replaces a unit, `delete CODE` and `bulk-delete`,
// which deletes the units whose codes a JSON array names. Emporix creates no
// missing unit, and answers a stale metadata.version as an unknown unit,
// with 404 Not Found: a conflict when the body carries one.
func unitWriteCommands(a *app, units resource[unit]) []*cobra.Command {
	manage := []string{unitManageScope}
	const visibility = "\nOther services see the change after up to 5 minutes."
	return []*cobra.Command{
		(&write{
			use:   createUse,
			short: "Create a unit",
			long: "Creates the unit that the JSON body of --file describes: its code, name, type, baseUnit\n" +
				"(whether it is the base unit of its type) and factor (what converts it into the base unit,\n" +
				"greater than zero), and its symbol if any. Emporix answers with the unit's code; a unit\n" +
				"with the code exists already: a conflict (exit 5)." + visibility,
			method: http.MethodPost, req: request{path: units.path}, scopes: manage, body: &payload{}, language: stringsOrMaps,
		}).command(a),
		(&write{
			use:   units.itemUse("update"),
			short: "Update a unit",
			long: "Replaces the unit CODE (the CODE that `epx units list` shows) with the unit that the JSON\n" +
				"body of --file describes, as create does, with the same code; Emporix creates no missing\n" +
				"unit. With metadata.version in the body, as `epx units get CODE -o json` prints it, Emporix\n" +
				"replaces only that version of the unit, and answers an older one as it answers an unknown\n" +
				"unit, with 404 Not Found: epx then exits with 5, a conflict (read the unit again before\n" +
				"you retry), and without a version with 4, not found." + visibility,
			method: http.MethodPut, req: request{path: units.itemPath()}, scopes: manage, body: &payload{},
			language: stringsOrMaps, languageHelp: unitLanguageHelp, ids: units.itemIDs(),
			answers: emporix.Answers{StaleNotFound: true},
		}).command(a),
		itemDelete(units, "Delete a unit", "Deletes the unit CODE, as `epx units list` shows it."+visibility,
			manage).command(a),
		(&write{
			use: bulkDeleteUse, short: "Delete units by their codes", method: http.MethodDelete,
			long: "Deletes the units whose codes the JSON array of --file names, such as [\"g\", \"kg\"]." +
				visibility,
			req: request{path: units.path}, scopes: manage, body: &payload{},
			deletes: deletion{many: "the units whose codes --file names"},
		}).command(a),
	}
}

// unitConversion is the part of a unit conversion, or of its factor, shown
// in tables.
type unitConversion struct {
	Input struct {
		SourceUnitAmount json.Number `json:"sourceUnitAmount"`
		SourceUnit       string      `json:"sourceUnit"`
		TargetUnit       string      `json:"targetUnit"`
	} `json:"input"`
	Output struct {
		TargetUnitAmount json.Number `json:"targetUnitAmount"`
		TargetUnit       string      `json:"targetUnit"`
		ConversionFactor json.Number `json:"conversionFactor"`
	} `json:"output"`
}

// unitConversionHelp ends the help of the unit conversions.
const unitConversionHelp = "\nAn unknown unit is not found (exit 4); Emporix answers units of different types with\n" +
	"400 Bad Request (exit 1)."

// newUnitConvertCommand returns `convert`, an amount of one unit in
// another: a calculation that Emporix serves as PUT, answering 201 Created
// without changing data, with no scope (live; the reference names none).
func newUnitConvertCommand(a *app) *cobra.Command {
	return reader[unitConversion]{view: view[unitConversion]{
		singular: "unit conversion",
		header:   []string{columnFrom, columnTo},
		row: func(c unitConversion, _ language) []string {
			in, out := c.Input, c.Output
			return []string{spaced(in.SourceUnitAmount.String(), in.SourceUnit), spaced(out.TargetUnitAmount.String(),
				out.TargetUnit)}
		},
	}}.command(a, &command{
		use:   "convert",
		short: "Convert an amount from one unit into another",
		long: "Converts an --amount of one unit into another of the same type (`epx units list` shows\n" +
			"their codes and types): FROM is the amount given, TO what it converts to." +
			unitConversionHelp,
		req: request{path: "/unit-handling/{tenant}/units/convert-unit-commands"},
		fetch: bodyRead{method: http.MethodPut, created: true, body: objectBody(append(unitPairFields(),
			numberField("input.sourceUnitAmount", "amount", "the `AMOUNT` to convert, in --from units").require())),
			fetch: single{items: oneItem}},
	})
}

// newUnitConversionFactorCommand returns `conversion-factor`, the factor from
// one unit into another, a calculation that Emporix serves as convert's.
func newUnitConversionFactorCommand(a *app) *cobra.Command {
	return reader[unitConversion]{view: view[unitConversion]{
		singular: "unit conversion factor",
		header:   []string{columnFrom, columnTo, "FACTOR"},
		row: func(c unitConversion, _ language) []string {
			return []string{c.Input.SourceUnit, c.Input.TargetUnit, c.Output.ConversionFactor.String()}
		},
	}}.command(a, &command{
		use:   "conversion-factor",
		short: "Calculate the factor that converts one unit into another",
		long: "Calculates the FACTOR that converts one unit into another of the same type\n" +
			"(`epx units list` shows their codes and types): an amount in FROM units times FACTOR is\n" +
			"the amount in TO units." +
			unitConversionHelp,
		req: request{path: "/unit-handling/{tenant}/units/conversion-factor-commands"},
		fetch: bodyRead{method: http.MethodPut, created: true, body: objectBody(unitPairFields()),
			fetch: single{items: oneItem}},
	})
}

// unitPairFields are the members of a unit conversion's body that name its
// units.
func unitPairFields() []bodyField {
	return []bodyField{
		stringField("input.sourceUnit", "from", "the `CODE` of the unit to convert from").require(),
		stringField("input.targetUnit", "to", "the `CODE` of the unit to convert to").require(),
	}
}

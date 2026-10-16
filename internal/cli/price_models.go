package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

// priceModel is the part of a Price Service price model shown in tables.
type priceModel struct {
	ID             string    `json:"id"`
	Name           localized `json:"name"`
	Default        *bool     `json:"default"`
	IncludesTax    *bool     `json:"includesTax"`
	IncludesMarkup *bool     `json:"includesMarkup"`
	TierDefinition struct {
		TierType string     `json:"tierType"`
		Tiers    modelTiers `json:"tiers"`
	} `json:"tierDefinition"`
	MeasurementUnit quantity `json:"measurementUnit"`
}

// quantity is an amount in a unit, such as 50 g.
type quantity struct {
	Amount   json.Number `json:"quantity"` // kept as sent, e.g. 1.0
	UnitCode string      `json:"unitCode"`
}

// String joins the amount and the unit code; a missing part is left out.
func (q quantity) String() string {
	return spaced(q.Amount.String(), q.UnitCode)
}

// modelTiers are the tiers of a price model, in the model's order.
type modelTiers []modelTier

// modelTier is one tier of a price model.
type modelTier struct {
	MinQuantity quantity `json:"minQuantity"`
}

func (t modelTier) String() string { return t.MinQuantity.String() }

// String lists the minimum quantity of each tier, joined with ", " because
// each quantity already holds a space.
func (t modelTiers) String() string { return joinEach(t, ", ", modelTier.String) }

// newPriceModelsCommand reads and changes price models, the tier structures
// that prices follow. Emporix returns one model for get, although the API
// reference documents a one-item array; get accepts both.
func newPriceModelsCommand(a *app) *cobra.Command {
	models := resource[priceModel]{
		use:   "price-models",
		short: "Read and change price models",
		about: "Reads and changes price models, the tier structure that prices follow. create, update and\n" +
			"delete change one; delete --force also deletes a model that prices use, with its prices.",
		long: "Reads price models, the tier structure that prices follow. DEFAULT marks the tenant's\n" +
			"default model, TYPE is the tier type (BASIC, TIERED or VOLUME), TIERS the minimum quantity\n" +
			"of each tier, UNIT the quantity a price applies to, and TAX and MARKUP whether prices of\n" +
			"the model include them (-o json adds descriptions and the tier ids that price tier values\n" +
			"refer to).",
		path:    "/price/{tenant}/priceModels",
		filters: listFilters{kinds: sortable},
		scopes:  []string{"price.pricemodel_read"},
		item:    oneItemOrArray,
		view: view[priceModel]{
			singular:  "price model",
			header:    []string{columnID, columnName, columnDefault, columnType, columnTiers, "UNIT", "TAX", "MARKUP"},
			localized: true,
			row: func(m priceModel, lang language) []string {
				return []string{
					m.ID, m.Name.in(lang), formatBool(m.Default), m.TierDefinition.TierType, m.TierDefinition.Tiers.String(),
					m.MeasurementUnit.String(), formatBool(m.IncludesTax), formatBool(m.IncludesMarkup),
				}
			},
		},
	}
	return newResourceCommand(a, models, priceModelWrites(a, models)...)
}

// priceModelManage is the scope of the writes of price models. A delete of a
// model that prices use, with its prices (forceDelete, --force), also needs
// priceModelAdmin, which the reference names beside it.
const (
	priceModelManage = "price.pricemodel_manage"
	priceModelAdmin  = "price.pricemodel_manage_admin"
)

// priceModelWrites returns the writes of price models: create, update ID (an
// upsert) and delete ID, which deletes a model that prices use only with
// --force.
func priceModelWrites(a *app, models resource[priceModel]) []*cobra.Command {
	at := models.at()
	create, update, remove := (&collectionWrites{
		at: at, one: "a price model", scopes: []string{priceModelManage}, language: stringsOrMaps, upsert: true,
		create: "Creates the price model in --file, a JSON object with name, includesTax, includesMarkup,\n" +
			"measurementUnit and tierDefinition: tierType (BASIC, TIERED or VOLUME) and the minQuantity of\n" +
			"each tier, e.g. " + `{"name": {"en": "Per kilogram"}, "includesTax": true,` + "\n" +
			`"includesMarkup": false, "measurementUnit": {"quantity": 1, "unitCode": "kg"},` + "\n" +
			`"tierDefinition": {"tierType": "BASIC", "tiers": [{"minQuantity": {"quantity": 0,` + "\n" +
			`"unitCode": "kg"}}]}}` + ". Without an id, Emporix makes one, which the response gives.",
		update: "Replaces the price model ID with the one in --file, a JSON object as\n" +
			"`epx " + at.get + " -o json` prints it, or creates it with that id where none exists\n" +
			"(an upsert).",
		remove: "Deletes the price model ID. Emporix refuses a model that prices use (400 Bad Request,\n" +
			"exit 1) unless --force is given, which needs both scopes below and deletes those prices\n" +
			"too: Emporix deletes them after the model, in the background." + unknownAsDeleted("ID"),
	}).writes()
	remove.choice = queryChoice{param: forceDeleteParam, options: []choiceOption{{
		flag: forceParam, value: queryTrue, scopes: []string{priceModelAdmin},
		usage: "delete a model that prices use, with its prices: sends forceDelete=true",
	}}}
	remove.deletes = deletion{many: "the price model and every price that uses it", with: forceParam}
	return writeCommands(a, []write{*create, *update, *remove})
}

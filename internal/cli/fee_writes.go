package cli

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/spf13/cobra"
)

const (
	feeCreate     = "fee.fee_create"
	feeUpdate     = "fee.fee_update"
	feeDelete     = "fee.fee_delete"
	itemFeeCreate = "fee.item_create"
	itemFeeUpdate = "fee.item_update"
	itemFeeDelete = "fee.item_delete"
)

func feeLanguageHelp() string {
	return mapsSent() + " The Fee Service ignores Content-Language: it stores a string as sent, " +
		"in no language, in place of any translations."
}

const feeUpdateConflict = "another fee has the code already, or the fee ID does not exist (Emporix answers that " +
	"with 409, not 404): check it with `epx fees get ID`"

func feeWrites(a *app, fees resource[fee]) []*cobra.Command {
	create, update, remove := (&collectionWrites{
		at: fees.tenantCollection(), one: "a fee", language: mapsOnly, languageHelp: feeLanguageHelp(),
		createHelp: "Creates the fee in --file, a JSON object with code (unique in the tenant), name, feeType " +
			"(PERCENT, ABSOLUTE or ABSOLUTE_MULTIPLY_ITEMQUANTITY), feePercentage or feeAbsolute (amount " +
			"and currency), siteCode and active, and optionally itemType (PRODUCT or PAYMENTTYPE), " +
			"taxable with taxCode, and activeTimespan, e.g. " + literal(`{"code": "deposit", "name": {"en": "Deposit"}, `+
			`"feeType": "ABSOLUTE", "feeAbsolute": {"amount": 0.25, "currency": "EUR"}, `+
			`"siteCode": "AcmeStore_NL", "active": true}`) + ". Without an id, Emporix makes one; the response " +
			"gives it and its YRN. A code that the tenant has already is a conflict (exit 5).",
		updateHelp: "Replaces the fee ID with the one in --file, a JSON object as create takes it, or with " +
			"--partial changes only the members that it gives, which still include code, siteCode, " +
			"active and feeType (Emporix refuses the body otherwise). Emporix answers an unknown ID " +
			"with 409 Conflict (exit 5), not as not found.",
		deleteHelp: "Deletes the fee ID, which no item has then. " + unknownAsDeleted("ID"),
	}).writes()
	create.scopes, update.scopes, remove.scopes = []string{feeCreate}, []string{feeUpdate}, []string{feeDelete}
	update.short, update.switches = "Replace or change a fee", []paramFlag{partialSwitch("fee")}
	update.conflict = feeUpdateConflict
	return writeCommands(a, []write{*create, *update, *remove})
}

func feeIDsField() bodyField {
	return listField("feeIds", "fee", "the `ID` of a fee, as `epx fees list` shows it (repeatable)").require()
}

type feeOwner struct {
	segment     string
	arg         string
	suffix      string
	summaryNoun string
	helpNoun    string
	argSource   string
	site        bool
	ids         *idSource
}

func feeOwners() (item, product feeOwner) {
	return feeOwner{segment: "itemFees", arg: "YRN", summaryNoun: "an item", helpNoun: "the item YRN",
			argSource: "YRN names the item, " + yrnNote + "."},
		feeOwner{segment: "productFees", arg: "PRODUCT_ID", suffix: "-product", summaryNoun: "a product",
			helpNoun: "the product PRODUCT_ID", argSource: productIDNote, site: true,
			ids: products().ids()}
}

func (o *feeOwner) fees(placeholder string) string {
	return feeService + "/" + o.segment + "/" + placeholder + "/fees"
}

func siteOrMain() queryFlags {
	return queryFlags{optional: []paramFlag{
		optionalSiteParam("the site `CODE` of the item fee, sent as siteCode"),
	}}
}

const mainUnlessSite = "Without --site, Emporix acts on the item fee on site main."

func (o *feeOwner) writes() []*write {
	update := request{path: o.fees(placeholderID)}
	updateSite := siteOrMain()
	onSite, short, siteNote := "", "", mainUnlessSite
	if o.site {
		update.query = url.Values{siteCodeParam: {placeholderSite}}
		updateSite = queryFlags{}
		onSite = " on the site that --site names"
		short = " on a site"
		siteNote = ""
	}
	const feeID = "FEE_ID is the ID that `epx fees list` shows."
	ids := argumentIDs{o.arg: o.ids}
	return []*write{
		{
			use: "update" + o.suffix + " " + o.arg, short: "Add fees to " + o.summaryNoun + short, method: http.MethodPut,
			long: joinSentences("Adds the fees that --fee names, or the feeIds of --file, a JSON object such as "+
				literal(`{"feeIds": ["f1", "f2"]}`)+", to those of "+o.helpNoun+onSite+", creating its item fee where it "+
				"has none. "+
				"Emporix removes no fee, with or without --partial, and an empty list changes nothing: delete"+
				o.suffix+"-fee removes one.", siteNote, o.argSource),
			req: update, params: updateSite, scopes: []string{itemFeeUpdate},
			body: &payload{typed: objectBody{feeIDsField()}}, switches: []paramFlag{{param: partialParam, flag: partialParam,
				usage: "send partial=true instead of partial=false, which Emporix treats alike"}}, ids: ids,
		},
		{
			use: deleteUse + o.suffix + " " + o.arg, short: "Delete the item fee of " + o.summaryNoun, method: http.MethodDelete,
			long: "Deletes the item fee of " + o.helpNoun + ", which then has no fees. " + mainUnlessSite +
				" " + unknownAsDeleted(o.arg) + " " + o.argSource,
			req:    request{path: o.fees(placeholderID)},
			params: siteOrMain(), scopes: []string{itemFeeDelete}, body: noBody{}, ids: ids,
			deletes: deletion{many: "every fee of " + o.helpNoun},
		},
		{
			use: deleteUse + o.suffix + "-fee " + o.arg + " FEE_ID", short: "Remove a fee from " + o.summaryNoun,
			method: http.MethodDelete,
			long: "Removes the fee FEE_ID from the item fee of " + o.helpNoun + " on the site of the fee; Emporix deletes " +
				"the item fee once it has no fees. " + unknownAsDeleted("FEE_ID") + " " + o.argSource + " " + feeID,
			req:    request{path: o.fees(placeholderParent) + "/" + placeholderID},
			scopes: []string{itemFeeDelete}, body: noBody{}, ids: ids,
		},
	}
}

func feeAssignmentWrites(a *app) []*cobra.Command {
	item, product := feeOwners()
	commands := []*cobra.Command{
		(&write{
			use: createUse, short: "Assign fees to an item on a site", method: http.MethodPost,
			long: "Assigns the fees that --fee names to the item that --item names on the site that --site " +
				"names, or those of --file, a JSON object with itemYrn, siteCode and feeIds, and prints " +
				"the id and YRN of the item fee. --item takes the YRN of an item, " + yrnNote + ".",
			example: "  epx fees assignments create --item 'urn:yaas:saasag:caasproduct:product:acme1;100200' \\\n" +
				"    --site AcmeStore_NL --fee f1 --fee f2",
			req:    request{path: feeService + "/" + item.segment},
			scopes: []string{itemFeeCreate},
			body: &payload{typed: objectBody{
				stringField("itemYrn", "item", "the `YRN` of the item to assign the fees to").require(),
				siteCodeField(siteCodeParam, "the `code` of the site that the fees apply on").require(),
				feeIDsField(),
			}},
		}).command(a),
	}
	for _, w := range slices.Concat(item.writes(), product.writes()) {
		commands = append(commands, w.command(a))
	}
	return commands
}

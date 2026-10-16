package cli

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/spf13/cobra"
)

// The writes of the Fee Service: fees, and which items have which fees (item
// fees), which it also addresses by product. Each operation has a scope of
// its own; the deletes of a product's item fees name fee.fee_delete and
// fee.item_delete, of which fee.item_delete, the one for item fees, is
// requested.
const (
	feeCreate     = "fee.fee_create"
	feeUpdate     = "fee.fee_update"
	feeDelete     = "fee.fee_delete"
	itemFeeCreate = "fee.item_create"
	itemFeeUpdate = "fee.item_update"
	itemFeeDelete = "fee.item_delete"
)

// feeLanguageHelp says, in the help of the writes of fees, how the Fee
// Service stores localized text: it ignores Content-Language, so a string is
// stored as sent, in no language, in place of the translations.
const feeLanguageHelp = mapsSent + " The Fee Service ignores\nContent-Language: it stores a string as sent, " +
	"in no language, in place of any\ntranslations."

// feeUpdateConflict is what a 409 Conflict of fees update means: a code that
// another fee has, as the reference documents, or an unknown ID, which
// Emporix answers so, not with 404 Not Found.
const feeUpdateConflict = "another fee has the code already, or the fee ID does not exist (Emporix answers that " +
	"with 409, not 404): check it with `epx fees get ID`"

// feeWrites returns the writes of fees: create, update ID (with --partial)
// and delete ID, each with the scope of its own operation. Their maps of
// translations go with Content-Language: *, which the Fee Service ignores
// (mapsOnly, feeLanguageHelp).
func feeWrites(a *app, fees resource[fee]) []*cobra.Command {
	create, update, remove := (&collectionWrites{
		at: fees.at(), one: "a fee", language: mapsOnly, languageHelp: feeLanguageHelp,
		create: "Creates the fee in --file, a JSON object with code (unique in the tenant), name, feeType\n" +
			"(PERCENT, ABSOLUTE or ABSOLUTE_MULTIPLY_ITEMQUANTITY), feePercentage or feeAbsolute (amount\n" +
			"and currency), siteCode and active, and optionally itemType (PRODUCT or PAYMENTTYPE),\n" +
			"taxable with taxCode, and activeTimespan, e.g. " + `{"code": "deposit", "name": {"en": "Deposit"},` +
			"\n" + `"feeType": "ABSOLUTE", "feeAbsolute": {"amount": 0.25, "currency": "EUR"},` + "\n" +
			`"siteCode": "AcmeStore_NL", "active": true}` + ". Without an id, Emporix makes one; the response\n" +
			"gives it and its YRN. A code that the tenant has already is a conflict (exit 5).",
		update: "Replaces the fee ID with the one in --file, a JSON object as create takes it, or with\n" +
			"--partial changes only the members that it gives, which still include code, siteCode,\n" +
			"active and feeType (Emporix refuses the body otherwise). Emporix answers an unknown ID\n" +
			"with 409 Conflict (exit 5), not as not found.",
		remove: "Deletes the fee ID, which no item has then." + unknownAsDeleted("ID"),
	}).writes()
	create.scopes, update.scopes, remove.scopes = []string{feeCreate}, []string{feeUpdate}, []string{feeDelete}
	update.short, update.switches = "Replace or change a fee", []paramFlag{partialSwitch("fee")}
	update.conflict = feeUpdateConflict
	return writeCommands(a, []write{*create, *update, *remove})
}

// feeIDsField is the body of a write that gives an item its fees: their ids,
// each set by one --fee.
func feeIDsField() bodyField {
	return listField("feeIds", "fee", "the `ID` of a fee, as `epx fees list` shows it (repeatable)").require()
}

// feeOwner is what item fees are of, as the Fee Service addresses them: an
// item, by its YRN, or a product, by its id. Each has the same writes, named
// after it (suffix) and taking its argument first.
type feeOwner struct {
	segment string    // the path segment of its item fees: itemFees or productFees
	arg     string    // its argument, e.g. YRN
	suffix  string    // of the names of its writes, e.g. -product in update-product
	one     string    // it in summaries, e.g. "an item"
	the     string    // it in help, e.g. "the item YRN"
	note    string    // where its argument comes from: a line of help
	site    bool      // its update requires --site, as the reference names siteCode for it
	feeSite string    // where delete-fee removes a fee, in its help; none where not known
	ids     *idSource // what its argument completes with; none for an item's YRN
}

// feeOwners returns the owners of item fees: an item and a product.
func feeOwners() (item, product feeOwner) {
	return feeOwner{segment: "itemFees", arg: "YRN", one: "an item", the: "the item YRN",
			note: "\nYRN names the item, " + yrnNote + "."},
		feeOwner{segment: "productFees", arg: "PRODUCT_ID", suffix: "-product", one: "a product",
			the: "the product PRODUCT_ID", note: productIDNote, site: true, feeSite: " on the site of the fee",
			ids: products().ids()}
}

// fees returns the path of the fees of the item fee of o that placeholder
// names, e.g. /fee/{tenant}/itemFees/{id}/fees.
func (o *feeOwner) fees(placeholder string) string {
	return feeService + "/" + o.segment + "/" + placeholder + "/fees"
}

// siteOrMain is the optional --site of a write of item fees that Emporix
// carries out on site main unless the siteCode query parameter, which the
// reference does not document, names another site.
func siteOrMain() queryFlags {
	return queryFlags{optional: []paramFlag{
		optionalSiteParam("the site `CODE` of the item fee, sent as siteCode"),
	}}
}

// mainUnlessSite says, in the help of a write that takes siteOrMain, which
// site it acts on without --site.
const mainUnlessSite = "\nWithout --site, Emporix acts on the item fee on site main."

// writes returns the writes of o's item fees: update, which adds fees (with
// partial=true or, without --partial, false), delete, and delete-fee, which
// removes one fee. Update requires
// --site where the reference names siteCode for it (o.site) and otherwise
// takes it optionally (siteOrMain); delete always takes it optionally;
// delete-fee takes none.
func (o *feeOwner) writes() []*write {
	update := request{path: o.fees(placeholderID)}
	updateSite := siteOrMain()
	onSite, short, siteNote := "", "", mainUnlessSite
	if o.site {
		update.query = url.Values{siteCodeParam: {placeholderSite}}
		updateSite = queryFlags{}
		onSite = "\non the site that --site names"
		short = " on a site"
		siteNote = ""
	}
	const feeID = "\nFEE_ID is the ID that `epx fees list` shows."
	ids := argumentIDs{o.arg: o.ids}
	return []*write{
		{
			use: "update" + o.suffix + " " + o.arg, short: "Add fees to " + o.one + short, method: http.MethodPut,
			long: "Adds the fees that --fee names, or the feeIds of --file, a JSON object such as\n" +
				`{"feeIds": ["f1", "f2"]}, to those of ` + o.the + onSite + ", creating its item fee where it\n" +
				"has none. Emporix removes no fee, with or without --partial, and an empty list changes\n" +
				"nothing: delete" + o.suffix + "-fee removes one." + siteNote + o.note,
			req: update, params: updateSite, scopes: []string{itemFeeUpdate},
			body: &payload{typed: objectBody{feeIDsField()}}, switches: []paramFlag{{param: partialParam, flag: partialParam,
				usage: "send partial=true instead of partial=false, which Emporix treats alike"}}, ids: ids,
		},
		{
			use: deleteUse + o.suffix + " " + o.arg, short: "Delete the item fee of " + o.one, method: http.MethodDelete,
			long: "Deletes the item fee of " + o.the + ", which then has no fees." + mainUnlessSite +
				unknownAsDeleted(o.arg) + o.note,
			req:    request{path: o.fees(placeholderID)},
			params: siteOrMain(), scopes: []string{itemFeeDelete}, body: noBody{}, ids: ids,
			deletes: deletion{many: "every fee of " + o.the},
		},
		{
			use: deleteUse + o.suffix + "-fee " + o.arg + " FEE_ID", short: "Remove a fee from " + o.one,
			method: http.MethodDelete,
			long: "Removes the fee FEE_ID from the item fee of " + o.the + o.feeSite + ";\nEmporix deletes " +
				"the item fee once it has no fees." + unknownAsDeleted("FEE_ID") + o.note + feeID,
			req:    request{path: o.fees(placeholderParent) + "/" + placeholderID},
			scopes: []string{itemFeeDelete}, body: noBody{}, ids: ids,
		},
	}
}

// feeAssignmentWrites returns the writes of item fees, which assign fees to
// an item, named by its YRN, or to a product: create, and the writes of each
// owner (feeOwner.writes).
func feeAssignmentWrites(a *app) []*cobra.Command {
	item, product := feeOwners()
	commands := []*cobra.Command{
		(&write{
			use: createUse, short: "Assign fees to an item on a site", method: http.MethodPost,
			long: "Assigns the fees that --fee names to the item that --item names on the site that --site\n" +
				"names, or those of --file, a JSON object with itemYrn, siteCode and feeIds, and prints\n" +
				"the id and YRN of the item fee. --item takes the YRN of an item,\n" + yrnNote + ".",
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

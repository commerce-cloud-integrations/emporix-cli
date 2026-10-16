package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// The writes of delivery times and their time slots, each with the scope of
// the Shipping Service's writes (shippingManageScope), which the reference
// names for every one; those of their delivery windows are with the reads of
// these (delivery_window.go).

// validateOverlap returns the query parameter of the writes of delivery
// times and slots that turns off the check that the time ranges of slots of
// one shipping method do not overlap.
func validateOverlap() queryFlags {
	return queryFlags{optional: []paramFlag{{
		param: "validateOverlap", flag: "validate-overlap", choices: []string{queryTrue, queryFalse},
		usage: "whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless " +
			"given false)",
	}}}
}

// jsonPatchHelp says what the body of a patch of a delivery time or slot is,
// with an example of an operation.
func jsonPatchHelp(example string) string {
	return "the JSON Patch (RFC 6902) in\n" +
		"--file: an array of operations whose op is add, remove or replace, such as\n[" + example + "]."
}

// deliveryTimeWrites returns the writes of delivery times: create,
// bulk-create, update ID, patch ID and delete ID.
func deliveryTimeWrites(a *app, times resource[deliveryTime]) []*cobra.Command {
	manage := []string{shippingManageScope}
	item, ids := request{path: times.itemPath()}, times.itemIDs()
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a delivery time", method: http.MethodPost,
			long: "Creates a delivery time from the JSON in --file: its siteCode and zoneId (or isForAllZones,\n" +
				"for a day without delivery on every zone of the site), its day (a weekday, singleDate or\n" +
				"datePeriod), isDeliveryDay, timeZoneId and deliveryDayShift, the slots of a delivery day\n" +
				"and optionally the delivery time's id and name, as `epx delivery-times get ID -o json`\n" +
				"prints one; a slot with a cutOffDayShift needs a delivery time without deliveryDayShift.\n" +
				"Emporix answers with its id, and refuses a delivery time that overlaps another of its\n" +
				"zone, or is not valid otherwise, with 400 Bad Request (exit 1).",
			req: request{path: times.path}, scopes: manage, body: &payload{}, params: validateOverlap(),
		}).command(a),
		(&write{
			use: bulkCreateUse, short: "Create delivery times in bulk", method: http.MethodPost,
			long: "Creates the delivery times in --file, a JSON array of delivery times, each as create\n" +
				"takes it.",
			req: request{path: times.path + "/" + bulkSegment}, scopes: manage, body: &payload{maxItems: 50},
			params: validateOverlap(),
		}).command(a),
		(&write{
			use: times.itemUse("update"), short: "Replace a delivery time", method: http.MethodPut,
			long: "Replaces the delivery time ID with the one in --file, as create takes it. With\n" +
				"metadata.version in the body, as `epx delivery-times get ID -o json` prints it, Emporix\n" +
				"refuses the write as a conflict (exit 5) when the delivery time changed since it was read;\n" +
				"without it, Emporix replaces it whatever its version. An unknown ID is not found (exit 4).",
			req: item, scopes: manage, body: &payload{}, params: validateOverlap(), ids: ids,
		}).command(a),
		(&write{
			use: times.itemUse("patch"), short: "Change a delivery time", method: http.MethodPatch,
			long: "Changes the delivery time ID with " +
				jsonPatchHelp(`{"op": "replace", "path": "/name", "value": "Mondays"}`) +
				"\nAn unknown ID is not found (exit 4).",
			req: item, scopes: manage, body: operationsFile(), params: validateOverlap(), ids: ids,
		}).command(a),
		(&write{
			use: times.itemUse("delete"), short: "Delete a delivery time", method: http.MethodDelete,
			long: "Deletes the delivery time ID with its time slots." + unknownAsDeleted("ID"),
			req:  item, scopes: manage, body: noBody{}, ids: ids,
		}).command(a),
	}
}

// slotWrites returns the writes of the time slots of a delivery time, below
// slotsOf (the delivery times nested in the slots group), slots being their
// list: create and delete-all DELIVERY_TIME_ID, and update, patch and delete
// DELIVERY_TIME_ID ID, which address one slot as get reads it.
func slotWrites(a *app, slotsOf resource[deliveryTime], slots child[deliveryTimeSlot]) []*cobra.Command {
	manage := []string{shippingManageScope}
	slot := func(verb string) childTarget { return childItemTarget(slotsOf, slots, verb, nil) }
	get := slot(getUse)
	item, collection := request{path: get.path}, request{path: slots.below(slotsOf.path)}
	deliveryTime := slotsOf.argument()
	ofItem, ofCollection := childItemIDs(slotsOf, slots, &get), argumentIDs{deliveryTime: slotsOf.ids()}
	itemHelp := childItemHelp(slotsOf, slots, &get)
	return []*cobra.Command{
		(&write{
			use: createUse + " " + deliveryTime, short: "Create a time slot of a delivery time", method: http.MethodPost,
			long: "Creates a time slot of a delivery day from the JSON in --file: its shippingMethod (a method\n" +
				"of the delivery time's zone), deliveryTimeRange (timeFrom and timeTo, local times such as\n" +
				"08:00), cutOffTime (its time, an ISO 8601 timestamp whose time of day counts, and for a\n" +
				"weekday its cutOffDayShift, which the delivery time must not combine with a\n" +
				"deliveryDayShift), capacity (the orders it takes) and optionally its id, as\n" +
				"`epx delivery-times slots get DELIVERY_TIME_ID ID -o json` prints one. Emporix answers\n" +
				"with its id, and a slot id that exists already with 400 Bad Request (exit 1), not as a\n" +
				"conflict.\n" + slotsOf.argumentSentence(notFound),
			req: collection, scopes: manage, body: &payload{}, params: validateOverlap(), ids: ofCollection,
		}).command(a),
		(&write{
			use: slot("update").use, short: "Replace a time slot of a delivery time", method: http.MethodPut,
			long: "Replaces the time slot ID of a delivery time with the one in --file, as create takes it." +
				itemHelp,
			req: item, scopes: manage, body: &payload{}, params: validateOverlap(), ids: ofItem,
		}).command(a),
		(&write{
			use: slot("patch").use, short: "Change a time slot of a delivery time", method: http.MethodPatch,
			long: "Changes the time slot ID of a delivery time with " +
				jsonPatchHelp(`{"op": "replace", "path": "/capacity", "value": 50}`) + itemHelp,
			req: item, scopes: manage, body: operationsFile(), params: validateOverlap(), ids: ofItem,
		}).command(a),
		(&write{
			use: slot("delete").use, short: "Delete a time slot of a delivery time",
			method: http.MethodDelete,
			long: "Deletes the time slot ID of a delivery time." +
				childItemAnswerHelp(slotsOf, slots, &get, answeredDeleted),
			req: item, scopes: manage, body: noBody{}, ids: ofItem,
		}).command(a),
		(&write{
			use: "delete-all " + deliveryTime, short: "Delete every time slot of a delivery time",
			method: http.MethodDelete,
			long:   "Deletes every time slot of a delivery time.\n" + slotsOf.argumentSentence(notFound),
			req:    collection, scopes: manage, body: noBody{}, ids: ofCollection,
			deletes: deletion{many: "every time slot of the delivery time"},
		}).command(a),
	}
}

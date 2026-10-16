package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// The writes of a site's shipping configuration: its zones, the methods of a
// zone, its shipping groups and the shipping group of each customer. Each
// sends the JSON of --file (a delete sends none) to the site that --site
// names, with the scope that the reference names for every one of them. The
// writes of zones, methods and groups send their maps of translations
// without a Content-Language header, as the reference documents: Emporix
// refuses them sent with Content-Language: * (headerlessMaps).

// shippingManageScope is the scope of the Shipping Service's writes.
const shippingManageScope = "shipping.shipping_manage"

// unknownGroup is what a 409 Conflict to an update of a shipping group means:
// Emporix answers an unknown group so (live), and a group has no version that
// could conflict.
const unknownGroup = "the site has no shipping group with this ID: `epx shipping groups list --site S` lists them"

// noGroup is what a 409 Conflict to an update of a customer's shipping group
// means: Emporix answers so (live) for a customer without one, as update
// creates no assignment, and an assignment has no version that could conflict.
const noGroup = "the customer has no shipping group on the site: assign it one with `epx shipping customers create`"

// zoneWrites returns the writes of zones: create, update ID, patch ID and
// delete ID.
func zoneWrites(a *app, zones resource[shippingZone]) []*cobra.Command {
	manage := []string{shippingManageScope}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a shipping zone", method: http.MethodPost,
			long: "Creates a shipping zone of the site --site names from the JSON in --file: its id, name,\n" +
				"shipTo (the countries it ships to, each with a postal code, or a prefix followed by *, where\n" +
				"it ships to some only) and whether it is the site's default zone, as in\n" +
				`{"id": "nl", "name": {"en": "Netherlands"}, "shipTo": [{"country": "NL"}], "default": true}.` + "\n" +
				"The first zone of a site becomes its default zone whatever default says.\n" +
				"Emporix answers with the zone's id; a zone with that id exists already: a conflict (exit 5).",
			req: request{path: zones.path}, scopes: manage, body: &payload{}, language: headerlessMaps,
		}).command(a),
		(&write{
			use: zones.itemUse("update"), short: "Replace a shipping zone", method: http.MethodPut,
			long: "Replaces the shipping zone ID of the site --site names with the zone in --file, as create\n" +
				"takes it: a member that it leaves out is removed, and a name given as a string in one\n" +
				"language (--content-language LANGUAGE) replaces every translation of it. Emporix refuses\n" +
				"to make the default zone of a site not default (400 Bad Request, exit 1). An unknown ID is\n" +
				"not found (exit 4).",
			req: request{path: zones.itemPath()}, scopes: manage, body: &payload{}, language: headerlessMaps,
			ids: zones.itemIDs(),
		}).command(a),
		(&write{
			use: zones.itemUse("patch"), short: "Change members of a shipping zone", method: http.MethodPatch,
			long: "Changes the members of the shipping zone ID of the site --site names that the JSON in\n" +
				"--file names, such as name or shipTo; the others keep their values, and a name merges into\n" +
				"the zone's translations, per language. An unknown ID is not found (exit 4).",
			req: request{path: zones.itemPath()}, scopes: manage, body: &payload{}, language: headerlessMaps,
			ids: zones.itemIDs(),
		}).command(a),
		(&write{
			use: zones.itemUse("delete"), short: "Delete a shipping zone", method: http.MethodDelete,
			long: "Deletes the shipping zone ID of the site --site names with its shipping methods. Delivery\n" +
				"times of the zone stay and still name it: delete them with `epx delivery-times delete ID`.\n" +
				"An unknown ID is not found (exit 4).",
			req: request{path: zones.itemPath()}, scopes: manage, body: noBody{},
			ids: zones.itemIDs(),
		}).command(a),
	}
}

// methodWrites returns the writes of the shipping methods of a zone, below
// methodsOf (the zones nested in the methods group), methods being their list:
// create SHIPPING_ZONE_ID, and update, patch and delete SHIPPING_ZONE_ID ID,
// which address one method as get reads it.
func methodWrites(a *app, methodsOf resource[shippingZone], methods child[shippingMethod]) []*cobra.Command {
	manage := []string{shippingManageScope}
	method := func(verb string) childTarget { return childItemTarget(methodsOf, methods, verb, nil) }
	get := method(getUse)
	item, ids := request{path: get.path}, childItemIDs(methodsOf, methods, &get)
	zoneHelp, itemHelp := "\n"+methodsOf.argumentSentence(notFound), childItemHelp(methodsOf, methods, &get)
	return []*cobra.Command{
		(&write{
			use: createUse + " " + methodsOf.argument(), short: "Create a shipping method of a zone", method: http.MethodPost,
			long: "Creates a shipping method of a zone from the JSON in --file: its id, name, active, fees\n" +
				"(each a cost from a minOrderValue, optionally for the customers of a shippingGroupId),\n" +
				"maxOrderValue and shippingTaxCode, as\n" +
				"`epx shipping zones methods get SHIPPING_ZONE_ID ID -o json` prints one. Emporix answers\n" +
				"with the method's id; a method with that id exists already: a conflict (exit 5)." + zoneHelp,
			req: request{path: methods.below(methodsOf.path)}, scopes: manage, body: &payload{},
			language: headerlessMaps, ids: argumentIDs{methodsOf.argument(): methodsOf.ids()},
		}).command(a),
		(&write{
			use: method("update").use, short: "Replace a shipping method of a zone", method: http.MethodPut,
			long: "Replaces the shipping method ID of a zone with the method in --file, as create takes it: a\n" +
				"member that it leaves out is removed." + itemHelp,
			req: item, scopes: manage, body: &payload{}, language: headerlessMaps, ids: ids,
		}).command(a),
		(&write{
			use: method("patch").use, short: "Change members of a shipping method of a zone",
			method: http.MethodPatch,
			long: "Changes the members of the shipping method ID of a zone that the JSON in --file names, such\n" +
				"as active or fees; the others keep their values." + itemHelp,
			req: item, scopes: manage, body: &payload{}, language: headerlessMaps, ids: ids,
		}).command(a),
		(&write{
			use: method("delete").use, short: "Delete a shipping method of a zone", method: http.MethodDelete,
			long: "Deletes the shipping method ID of a zone. The time slots of delivery times that name it\n" +
				"stay and still name it." + itemHelp,
			req: item, scopes: manage, body: noBody{}, ids: ids,
		}).command(a),
	}
}

// groupWrites returns the writes of shipping groups: create, update ID and
// delete ID.
func groupWrites(a *app, groups resource[shippingGroup]) []*cobra.Command {
	manage := []string{shippingManageScope}
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a shipping group", method: http.MethodPost,
			long: "Creates a shipping group of the site --site names from the JSON in --file: its id and\n" +
				`description, as in {"id": "vip", "description": {"en": "Free shipping"}}. A fee of a` + "\n" +
				"shipping method applies to the customers of a group (`epx shipping customers create`) when\n" +
				"it names the group as its shippingGroupId. Emporix answers with the group's id; a group\n" +
				"with that id exists already: a conflict (exit 5).",
			req: request{path: groups.path}, scopes: manage, body: &payload{}, language: headerlessMaps,
		}).command(a),
		(&write{
			use: groups.itemUse("update"), short: "Replace a shipping group", method: http.MethodPut,
			long: "Replaces the shipping group ID of the site --site names with the group in --file, as create\n" +
				"takes it. Emporix answers an unknown ID with 409 Conflict (exit 5):\n" + unknownGroup + ".",
			req: request{path: groups.itemPath()}, scopes: manage, body: &payload{}, language: headerlessMaps,
			ids: groups.itemIDs(), conflict: unknownGroup,
		}).command(a),
		(&write{
			use: groups.itemUse("delete"), short: "Delete a shipping group", method: http.MethodDelete,
			long: "Deletes the shipping group ID of the site --site names. Fees of shipping methods that name\n" +
				"it as their shippingGroupId stay and still name it, and customers assigned to it stay assigned\n" +
				"to it (`epx shipping customers`)." + unknownAsDeleted("ID"),
			req: request{path: groups.itemPath()}, scopes: manage, body: noBody{},
			ids: groups.itemIDs(),
		}).command(a),
	}
}

// customerGroupWrites returns the writes of the shipping group of each
// customer (customer-group relations): create, update CUSTOMER and delete
// CUSTOMER.
func customerGroupWrites(a *app, customers resource[shippingCustomer]) []*cobra.Command {
	manage := []string{shippingManageScope}
	const relation = `{"customerId": "C001", "groupId": "vip"}`
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Assign a customer to a shipping group", method: http.MethodPost,
			long: "Assigns a customer to a shipping group of the site --site names: the JSON in --file names\n" +
				"the customer id and the group's, as in " + relation + ". A customer\n" +
				"that belongs to a group already is a conflict (exit 5): change its group with update.\n" +
				"Emporix checks neither the customer nor the group: unknown ids are assigned too.",
			req: request{path: customers.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: customers.itemUse("update"), short: "Change the shipping group of a customer", method: http.MethodPut,
			long: "Assigns the customer CUSTOMER (the customer id) to the shipping group that the JSON in --file\n" +
				"names, as create takes it, in place of the group it belongs to on the site --site names; a\n" +
				"customerId in the JSON is ignored. Emporix answers a customer without a shipping group with\n" +
				"409 Conflict (exit 5):\n" + noGroup + ".",
			req: request{path: customers.itemPath()}, scopes: manage, body: &payload{},
			ids: customers.itemIDs(), conflict: noGroup,
		}).command(a),
		(&write{
			use: customers.itemUse("delete"), short: "Remove a customer from its shipping group",
			method: http.MethodDelete,
			long: "Removes the customer CUSTOMER (the customer id) from its shipping group on the site --site\n" +
				"names. Emporix answers a CUSTOMER without a shipping group, or an unknown one, as one that\n" +
				"it removed (204 No Content).",
			req: request{path: customers.itemPath()}, scopes: manage, body: noBody{},
			ids: customers.itemIDs(),
		}).command(a),
	}
}

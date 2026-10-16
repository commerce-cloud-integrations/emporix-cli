package cli

import (
	"net/http"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// orderUpdateScope is the scope of the Order Service's writes that change an
// order; creating and deleting one have scopes of their own, and changing
// one that is COMPLETED takes one more (completedOrder). The *_by_vendor
// scopes are those of a vendor's own orders.
const orderUpdateScope = "order.order_update"

// orderStatuses are the statuses of an order, which `orders transition`
// moves it to.
func orderStatuses() []string {
	return []string{"IN_CHECKOUT", "CREATED", "CONFIRMED", "DECLINED", "SHIPPED", "COMPLETED"}
}

// completedOrder is the flag of a write that may change a COMPLETED order,
// which Emporix allows only with order.order_update_completed besides
// order.order_update (live: neither alone): epx requests it only when told
// that the order is COMPLETED. Only update and patch take it: Emporix moves a
// COMPLETED order to no other status and recalculates none, whatever the
// scopes (live).
func completedOrder() queryChoice {
	return queryChoice{options: []choiceOption{{flag: "completed", scopes: []string{"order.order_update_completed"},
		usage: "for an order that is COMPLETED, which Emporix changes only with one more scope"}}}
}

// recalculateOrder is the query switch with which Emporix recalculates the
// whole order that a write replaces or changes.
func recalculateOrder() []paramFlag {
	return []paramFlag{{param: recalculate, flag: recalculate,
		usage: "have Emporix recalculate the whole order, coupon discounts included"}}
}

// orderWrites returns the writes of orders: create, update ID, patch ID and
// delete ID, and the actions on one order (orderActions).
func orderWrites(orders resource[order]) []*write {
	item := request{path: orders.itemPath()}
	update := []string{orderUpdateScope}
	unversioned := "Emporix does not check the order's metadata.version: it applies the change whatever version\n" +
		"the body names, and does not answer a stale one with 409 Conflict; an unknown order is not\n" +
		"found (exit 4).\n" +
		orders.idSource() + "."
	writes := []*write{
		{
			use: createUse, short: "Create an order",
			long: "Creates an order on behalf of a customer from the JSON object in --file, shaped as\n" +
				"`epx orders get ID -o json` prints one (entries, customer, siteCode, currency and\n" +
				"calculatedPrice among others). Emporix makes the order's id and answers with it: an id in\n" +
				"the body is ignored, so sending the body again creates another order (with checkout\n" +
				"true, at most one per cartId). It stores the status CREATED for IN_CHECKOUT, and sets\n" +
				"mixins.generalAttributes (extendedOrderStatus \"10\", orderNumber and customerFirstOrder)\n" +
				"whatever the body says; `epx orders patch` can change them afterwards, e.g. to make the\n" +
				"order ready for `epx invoice-jobs create`. It also adds an address of the order to the\n" +
				"customer's addresses, as a default one tagged shipping and billing, so a customer who had\n" +
				"a default address then has two default addresses.\n\n" +
				"Give the order's and its entries' metadata.mixins the URLs of the schemas of their mixins,\n" +
				"as orders placed in a checkout carry them: without them, `epx orders add-entries --store`,\n" +
				"`recalculate`, `patch --recalculate` and `split` fail on the order with 400 Bad Request\n" +
				"(exit 1).",
			method: http.MethodPost, req: request{path: orders.path}, scopes: []string{"order.order_create"},
			body: &payload{},
		},
		{
			use: orders.itemUse("update"), short: "Replace an order",
			long: "Replaces the order ID with the order in --file, a JSON object as\n" +
				"`epx orders get ID -o json` prints it.\n" + unversioned,
			method: http.MethodPut, req: item, scopes: update, body: &payload{}, switches: recalculateOrder(),
			choice: completedOrder(), ids: orders.itemIDs(),
		},
		{
			use: orders.itemUse("patch"), short: "Change members of an order",
			long: "Changes the members of the order ID that the JSON object in --file gives, such as its\n" +
				"shippingAddress or mixins, and keeps the others.\n" + unversioned,
			method: http.MethodPatch, req: item, scopes: update, body: &payload{}, switches: recalculateOrder(),
			choice: completedOrder(), ids: orders.itemIDs(),
		},
		orders.deleteWrite([]string{"order.order_delete"}, "Deletes the order ID; it sends no body.\n"+
			orders.idSource()+"."),
	}
	return append(writes, orderActions(orders)...)
}

// orderActions returns the actions on one order: transition, add-entries,
// recalculate and split.
func orderActions(orders resource[order]) []*write {
	id := orders.argument()
	ofOrder := argumentIDs{id: orders.ids()}
	below := func(segment string) request { return request{path: orders.itemPath() + "/" + segment} }
	update := []string{orderUpdateScope}
	return []*write{
		{
			use: "transition " + id, short: "Move an order to another status",
			long: "Moves the order " + id + " to the status that --status names, one that\n" +
				"`epx orders transitions " + id + "` lists: CONFIRMED or DECLINED from CREATED, COMPLETED or\n" +
				"DECLINED from CONFIRMED, and none from COMPLETED or DECLINED. Emporix refuses another with\n" +
				forbiddenRefusal + " and says why, and answers the order's own status as done,\n" +
				"without an entry in its history. " + orders.argumentSource() + ".",
			method: http.MethodPost, req: below("transitions"), scopes: update, ids: ofOrder,
			answers: emporix.Answers{ForbiddenIsRefusal: true},
			body: &payload{typed: objectBody{
				choiceField("status", "status", "the `STATUS` to move the order to", orderStatuses()...).require(),
			}},
		},
		{
			use: "add-entries " + id, short: "Add products to an order, or change its entries",
			long: "Adds products to the order " + id + ", or changes the quantity of its entries, and\n" +
				"recalculates it: the JSON object in --file holds entries, each with the productId of a\n" +
				"product to add or the entryId of an entry to change (an id in the order's entries, as\n" +
				"`epx orders get ID -o json` prints them), and its quantity; 0 removes the entry.\n" +
				"Emporix answers with the order recalculated, which it stores only with --store: without\n" +
				"it, the order stays as it is. Emporix refuses to change the entries of a COMPLETED order\n" +
				"with 400 Bad Request (exit 1).",
			method: http.MethodPost, req: below("entries"), scopes: update, body: &payload{}, ids: ofOrder,
			switches: []paramFlag{{param: "store", flag: "store",
				usage: "store the recalculated order; without it, Emporix only answers with it"}},
		},
		{
			use: "recalculate " + id, short: "Set the delivered amounts of an order and recalculate it",
			long: "Sets the amounts delivered of entries of the order " + id + " and recalculates the\n" +
				"order: the JSON object in --file holds entries, each with its entryId (an id in the\n" +
				"order's entries, as `epx orders get ID -o json` prints them) and deliveredAmount, and\n" +
				"optionally deliveredAmountInPc and unit, the unit of deliveredAmount (e.g. GRM).\n" +
				"Emporix refuses to recalculate a COMPLETED order with 400 Bad Request (exit 1), whatever\n" +
				"the scopes.",
			method: http.MethodPost, req: below("calculations"), scopes: update, body: &payload{}, ids: ofOrder,
			answer: textAnswer{what: "a sentence of text"},
		},
		{
			use: "split " + id, short: "Split an order into sub-orders by vendor",
			long: "Splits the order " + id + " into sub-orders, one per vendor of its entries, sending\n" +
				"splitBy VENDOR_ID, the only criterion that Emporix offers, unless --file gives the body.\n" +
				"Emporix splits only an order with status CREATED that was not split before and has no\n" +
				"discounts: it becomes the MASTER_ORDER of the SUB_ORDERs that the response names. An order\n" +
				"whose entries name no vendor stays as it is (the response names no subOrders); an order\n" +
				"split before, or a sub-order, is refused with 400 Bad Request (exit 1).",
			method: http.MethodPost, req: below("split"), scopes: update, ids: ofOrder,
			body: &payload{typed: objectBody{fixedField("splitBy", "VENDOR_ID")}},
		},
	}
}

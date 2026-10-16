package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The delivery windows of delivery times: those open to a cart, whether one
// can take an order, the count of its orders and its delivery cycle. A
// delivery window is a delivery time on a date in one of its slots, which
// the flags of deliveryWindowBody name.

// deliveryWindow is the part of a cart's delivery window shown in tables.
type deliveryWindow struct {
	ID                string `json:"id"`
	DeliveryDate      string `json:"deliveryDate"` // the one member the reference requires
	DeliveryTimeRange struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"deliveryTimeRange"`
	DeliveryMethod string `json:"deliveryMethod"`
	ZoneID         string `json:"zoneId"`
	CutOffTime     string `json:"cutOffTime"`
	SlotID         string `json:"slotId"`
}

// newDeliveryWindowsCommand reads the delivery windows open to a cart, which
// Emporix computes from the delivery times and returns whole. The reference
// declares no scope for them, and Emporix answers a token without scopes
// (live: the windows of a cart with a postal code, and 404 for an unknown
// cart).
func newDeliveryWindowsCommand(a *app) *cobra.Command {
	return reader[deliveryWindow]{view: view[deliveryWindow]{
		singular: "delivery window",
		header:   []string{"DATE", columnFrom, columnTo, columnMethod, columnZone, "CUT-OFF", "SLOT", columnID},
		row: func(w deliveryWindow, _ language) []string {
			return []string{
				w.DeliveryDate, w.DeliveryTimeRange.StartTime, w.DeliveryTimeRange.EndTime, w.DeliveryMethod,
				w.ZoneID, w.CutOffTime, w.SlotID, w.ID,
			}
		},
	}}.command(a, &command{
		use:   "windows CART_ID",
		short: "List the delivery windows open to a cart",
		long: "Reads the delivery windows in which a cart can be delivered to the country and postal code\n" +
			"set on the cart. DATE is the delivery day, FROM and TO the delivery time range on it,\n" +
			"METHOD the shipping method, ZONE the shipping zone, CUT-OFF the latest time to order for\n" +
			"the window, SLOT the time slot of its delivery time, which\n" +
			"`epx delivery-times slots list DELIVERY_TIME_ID` lists, and ID the id of that delivery\n" +
			"time (`epx delivery-times get ID`); the delivery cycle is only in -o json.\n" +
			"For another postal code than the cart's, run\n" +
			"`epx api get /shipping/{tenant}/actualDeliveryWindows/CART_ID -p postalCode=CODE`.\n" +
			"CART_ID is the id of a cart; an unknown one is not found (exit 4).",
		req:   request{path: "/shipping/{tenant}/actualDeliveryWindows/" + placeholderID},
		fetch: listing{read: (*emporix.Client).GetUnpaged},
	})
}

// newDeliveryWindowValidationCommand returns `validate`, which checks whether
// a delivery window can still take an order: a check that Emporix serves as
// POST without changing data, with no scope (live; the reference names none).
func newDeliveryWindowValidationCommand(a *app) *cobra.Command {
	check := reader[bool]{view: answerView("delivery window validation", "VALID")}
	return check.command(a, &command{
		use:   "validate",
		short: "Check whether a delivery window can take an order",
		long: "Checks whether a delivery window can take an order.\n" + deliveryWindowHelp + "\n" +
			"VALID is true when it can; Emporix answers a date in the past or in another format, or a\n" +
			"window that is full or not offered that day, with 400 Bad Request (exit 1), and an unknown\n" +
			"delivery time as not found (exit 4).",
		req: request{path: "/shipping/{tenant}/deliveryWindowValidation"},
		fetch: bodyRead{method: http.MethodPost, body: deliveryWindowBody(deliveryWindowMember),
			fetch: single{items: oneBoolean}},
	})
}

// newDeliveryWindowCountCommand returns `increment-counter`, which counts one
// more order in a delivery window, as the Checkout Service does for each
// order with a delivery window that it places. The reference names no scope
// for it, so it requests the scope of the Shipping Service's writes.
func newDeliveryWindowCountCommand(a *app) *cobra.Command {
	return (&write{
		use:   "increment-counter",
		short: "Count one more order in a delivery window",
		long: "Counts one more order in a delivery window that has capacity left.\n" + deliveryWindowHelp +
			"\nThe Checkout Service counts each order that it places with a delivery window; only a\n" +
			"checkout of your own needs this. Emporix answers true once it counted the order, and a\n" +
			"window that is full or not offered that day, or a date in the past or in another format,\n" +
			"with 400 Bad Request (exit 1). " + wholeBodyNote,
		method: http.MethodPost, req: request{path: "/shipping/{tenant}/actualDeliveryWindows/incrementCounter"},
		scopes: []string{shippingManageScope}, body: &payload{typed: deliveryWindowBody(deliveryWindowMember)},
	}).command(a)
}

// newDeliveryCycleCommand returns `generate-cycle`, which names the delivery
// cycle of a delivery window: a calculation that Emporix serves as POST and
// answers with 201 Created, but that changes nothing (live: the delivery time
// and its slots stay as they were, and the same window gets the same name
// again). It answers with the name as bare text, although labelled JSON
// (cycleName). The reference names no scope for it; Emporix refuses it to a
// token with shipping.shipping_read but not to one without any scope (live),
// so it requests none; to that token it answers an unknown delivery time with
// 403 Forbidden (live).
func newDeliveryCycleCommand(a *app) *cobra.Command {
	cycle := reader[string]{view: stringsView("delivery cycle", "CYCLE")}
	return cycle.command(a, &command{
		use:   "generate-cycle",
		short: "Name the delivery cycle of a delivery window",
		long: "Prints the name of the delivery cycle of a delivery window, as Emporix names it in the\n" +
			"delivery windows of a cart and in orders (deliveryCycle in\n" +
			"`epx delivery-times windows CART_ID -o json`): the delivery date, the hour (UTC) of the\n" +
			"slot's cut-off time and the deliveryCycleName of its cutOffTime, such as 2026-06-20-6morning.\n" +
			deliveryWindowHelp + "\nEmporix answers an unknown delivery time with 403 Forbidden (exit 3), not as not found.",
		req: request{path: "/shipping/{tenant}/delivery-cycles/generate"},
		fetch: bodyRead{method: http.MethodPost, created: true, body: deliveryWindowBody("deliveryTimeId"),
			fetch: cycleName{}},
	})
}

// cycleName reads the name of a delivery cycle, which Emporix answers as
// bare text labelled JSON, such as 2026-06-20-6morning: the text, trimmed,
// becomes a JSON string (textAsJSON), which -o json prints. A JSON string is
// taken as sent; any other JSON, or no text, fails. A name starts with its
// date, which is never valid JSON, so no bare name is mistaken for JSON.
type cycleName struct{}

func (cycleName) bind(*cobra.Command) (readFunc, shape) {
	read := func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		answer, err := c.Get(ctx, path, query)
		if err != nil {
			return answer, err
		}
		return textAsJSON(answer)
	}
	return read, oneName
}

// deliveryWindowHelp says which delivery window the flags of
// deliveryWindowBody name.
const deliveryWindowHelp = "The delivery window is the delivery time --delivery-time on --date, an ISO 8601\n" +
	"date and time such as 2026-11-02T12:00:00.000Z, in its --slot: `epx delivery-times list`\n" +
	"and `epx delivery-times slots list` show their IDs."

// wholeBodyNote ends the help of a write whose body flags make
// (deliveryWindowBody).
const wholeBodyNote = "The body may also come whole from --file."

// deliveryWindowMember is the member of most bodies that name a delivery
// window, or a slot, that names its delivery time; the body of a delivery
// cycle names it deliveryTimeId.
const deliveryWindowMember = "deliveryWindowId"

// deliveryWindowBody is the body that names a delivery window: a delivery
// time (its member timeMember), one of its slots and the delivery date.
func deliveryWindowBody(timeMember string) objectBody {
	return append(deliverySlotFields(timeMember),
		stringField("deliveryDate", "date", "the delivery `DATE` and time, ISO 8601").require())
}

// deliverySlotFields are the members of a body that name a delivery time
// (its member timeMember) and one of its slots. Both are required, the slot
// too (seen live; the reference makes it optional for a validation).
func deliverySlotFields(timeMember string) []bodyField {
	return []bodyField{
		stringField(timeMember, "delivery-time", "the delivery time's `ID`").require(),
		stringField("slotId", "slot", "the `ID` of the delivery time's slot").require(),
	}
}

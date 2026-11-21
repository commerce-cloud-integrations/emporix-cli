package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type deliveryWindow struct {
	ID                string `json:"id"`
	DeliveryDate      string `json:"deliveryDate"`
	DeliveryTimeRange struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"deliveryTimeRange"`
	DeliveryMethod string `json:"deliveryMethod"`
	ZoneID         string `json:"zoneId"`
	CutOffTime     string `json:"cutOffTime"`
	SlotID         string `json:"slotId"`
}

// The reference declares no scope for delivery windows, and Emporix answers a
// token without scopes (live).
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
		long: "Reads the delivery windows in which a cart can be delivered to the country and postal code " +
			"set on the cart. DATE is the delivery day, FROM and TO the delivery time range on it, " +
			"METHOD the shipping method, ZONE the shipping zone, CUT-OFF the latest time to order for " +
			"the window, SLOT the time slot of its delivery time, which " +
			"`epx delivery-times slots list DELIVERY_TIME_ID` lists, and ID the id of that delivery " +
			"time (`epx delivery-times get ID`); the delivery cycle is only in `-o json`. " +
			"For another postal code than the cart's, run " +
			"`epx api get /shipping/{tenant}/actualDeliveryWindows/CART_ID -p postalCode=CODE`. " +
			"CART_ID is the id of a cart; an unknown one is not found (exit 4).",
		req:   request{path: "/shipping/{tenant}/actualDeliveryWindows/" + placeholderID},
		fetch: listing{read: (*emporix.Client).GetUnpaged},
	})
}

// Emporix serves the validation as POST without changing data, with no scope
// (live; the reference names none).
func newDeliveryWindowValidationCommand(a *app) *cobra.Command {
	check := reader[bool]{view: answerView("delivery window validation", "VALID")}
	return check.command(a, &command{
		use:   "validate",
		short: "Check whether a delivery window can take an order",
		long: "Checks whether a delivery window can take an order. " + deliveryWindowHelp + " " +
			"VALID is true when it can; Emporix answers a date in the past or in another format, or a " +
			"window that is full or not offered that day, with 400 Bad Request (exit 1), and an unknown " +
			"delivery time as not found (exit 4).",
		req: request{path: "/shipping/{tenant}/deliveryWindowValidation"},
		fetch: bodyRead{method: http.MethodPost, body: deliveryWindowBody(deliveryWindowMember),
			fetch: single{items: oneBoolean}},
	})
}

// The reference names no scope for incrementing the counter, so it requests
// the scope of the Shipping Service's writes.
func newDeliveryWindowCountCommand(a *app) *cobra.Command {
	return (&write{
		use:   "increment-counter",
		short: "Count one more order in a delivery window",
		long: "Counts one more order in a delivery window that has capacity left. " + deliveryWindowHelp +
			" The Checkout Service counts each order that it places with a delivery window; only a " +
			"checkout of your own needs this. Emporix answers true once it counted the order, and a " +
			"window that is full or not offered that day, or a date in the past or in another format, " +
			"with 400 Bad Request (exit 1). " + wholeBodyNote,
		method: http.MethodPost, req: request{path: "/shipping/{tenant}/actualDeliveryWindows/incrementCounter"},
		scopes: []string{shippingManageScope}, body: &payload{typed: deliveryWindowBody(deliveryWindowMember)},
	}).command(a)
}

// Emporix serves the delivery cycle as POST and answers with 201 Created, but
// changes nothing (live). The reference names no scope for it; Emporix refuses
// it to a token with shipping.shipping_read but not to one without any scope
// (live), so it requests none; to that token it answers an unknown delivery
// time with 403 Forbidden (live).
func newDeliveryCycleCommand(a *app) *cobra.Command {
	cycle := reader[string]{view: stringsView("delivery cycle", "CYCLE")}
	return cycle.command(a, &command{
		use:   "generate-cycle",
		short: "Name the delivery cycle of a delivery window",
		long: "Prints the name of the delivery cycle of a delivery window, as Emporix names it in the " +
			"delivery windows of a cart and in orders (deliveryCycle in " +
			"`epx delivery-times windows CART_ID -o json`): the delivery date, the hour (UTC) of the " +
			"slot's cut-off time and the deliveryCycleName of its cutOffTime, such as 2026-06-20-6morning. " +
			deliveryWindowHelp + " Emporix answers an unknown delivery time with 403 Forbidden (exit 3), not as not found.",
		req: request{path: "/shipping/{tenant}/delivery-cycles/generate"},
		fetch: bodyRead{method: http.MethodPost, created: true, body: deliveryWindowBody("deliveryTimeId"),
			fetch: cycleName{}},
	})
}

// Emporix answers the name of a delivery cycle as bare text labelled JSON,
// such as 2026-06-20-6morning.
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

const deliveryWindowHelp = "The delivery window is the delivery time --delivery-time on --date, an ISO 8601 " +
	"date and time such as 2026-11-02T12:00:00.000Z, in its --slot: `epx delivery-times list` " +
	"and `epx delivery-times slots list` show their IDs."

const wholeBodyNote = "The body may also come whole from --file."

const deliveryWindowMember = "deliveryWindowId"

func deliveryWindowBody(timeMember string) objectBody {
	return append(deliverySlotFields(timeMember),
		stringField("deliveryDate", "date", "the delivery `DATE` and time, ISO 8601").require())
}

// Both members are required, the slot too (seen live; the reference makes it
// optional for a validation).
func deliverySlotFields(timeMember string) []bodyField {
	return []bodyField{
		stringField(timeMember, "delivery-time", "the delivery time's `ID`").require(),
		stringField("slotId", "slot", "the `ID` of the delivery time's slot").require(),
	}
}

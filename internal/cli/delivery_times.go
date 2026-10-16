package cli

import (
	"cmp"
	"encoding/json"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// deliveryTime is the part of a Shipping Service delivery time shown in tables.
type deliveryTime struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	SiteCode         string            `json:"siteCode"`
	ZoneID           string            `json:"zoneId"`
	IsForAllZones    bool              `json:"isForAllZones"`
	IsDeliveryDay    *bool             `json:"isDeliveryDay"`
	Day              deliveryDay       `json:"day"`
	TimeZoneID       string            `json:"timeZoneId"`
	DeliveryDayShift json.Number       `json:"deliveryDayShift"`
	Slots            []json.RawMessage `json:"slots"` // counted only
}

// deliveryDay is the day of a delivery time: a weekday, a single date or a
// period of dates.
type deliveryDay struct {
	Weekday    string             `json:"weekday"`
	SingleDate string             `json:"singleDate"`
	DatePeriod deliveryDatePeriod `json:"datePeriod"`
}

// deliveryDatePeriod is the period of dates of a delivery day.
type deliveryDatePeriod struct {
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
}

// zone returns the delivery zone, or "all zones" for a day without delivery
// on every zone of the site (Emporix allows isForAllZones only there).
func (d *deliveryTime) zone() string {
	if d.IsForAllZones {
		return "all zones"
	}
	return d.ZoneID
}

// day returns the weekday, the date or the date period (ISO 8601 from/to)
// the delivery time applies to; Emporix sets only one of them.
func (d *deliveryTime) day() string {
	period := d.Day.DatePeriod
	if period.DateFrom != "" || period.DateTo != "" {
		return period.DateFrom + "/" + period.DateTo
	}
	return cmp.Or(d.Day.Weekday, d.Day.SingleDate)
}

// deliveryTimeSlot is the part of a delivery time slot shown in tables.
type deliveryTimeSlot struct {
	ID                string `json:"id"`
	ShippingMethod    string `json:"shippingMethod"`
	DeliveryTimeRange struct {
		TimeFrom string `json:"timeFrom"`
		TimeTo   string `json:"timeTo"`
	} `json:"deliveryTimeRange"`
	CutOffTime struct {
		Time           string      `json:"time"`
		CutOffDayShift json.Number `json:"cutOffDayShift"`
	} `json:"cutOffTime"`
	Capacity json.Number `json:"capacity"`
}

// newDeliveryTimesCommand reads delivery times, which the service returns
// whole, the time slots of one and the delivery windows of a cart, and
// changes delivery times and their slots.
func newDeliveryTimesCommand(a *app) *cobra.Command {
	times := resource[deliveryTime]{
		use:   "delivery-times",
		short: "Read and change delivery times and their time slots; read the delivery windows of a cart",
		long: "Reads delivery times, the days on which a site delivers to a zone. ZONE is a zone id\n" +
			"(`epx shipping zones list --site S`), or \"all zones\" for a day without delivery on every\n" +
			"zone; DAY is a weekday, a date or a date period (from/to); DELIVERS is false for a day\n" +
			"without delivery; DAY SHIFT is how many days the delivery follows an order placed before\n" +
			"the cut-off, when the delivery time sets it (a weekday's slots may set their own); SLOTS\n" +
			"counts the time slots, which `epx delivery-times slots list DELIVERY_TIME_ID` lists.",
		about: "Reads and changes delivery times, the days on which a site delivers to a zone, and their\n" +
			"time slots; reads the delivery windows open to a cart, checks and counts the orders of a\n" +
			"delivery window and names its delivery cycle.",
		path:   deliveryTimesPath,
		scopes: []string{"shipping.shipping_read"},
		read:   (*emporix.Client).GetUnpaged,
		view: view[deliveryTime]{
			singular: "delivery time",
			header: []string{
				columnID, columnName, columnSite, columnZone, "DAY", "DELIVERS", "TIME ZONE", "DAY SHIFT", "SLOTS",
			},
			row: func(d deliveryTime, _ language) []string {
				return []string{
					d.ID, d.Name, d.SiteCode, d.zone(), d.day(), formatBool(d.IsDeliveryDay), d.TimeZoneID,
					d.DeliveryDayShift.String(), strconv.Itoa(len(d.Slots)),
				}
			},
		},
	}
	return newResourceCommand(a, times, slices.Concat([]*cobra.Command{
		newDeliveryTimeSlotsCommand(a, times), newDeliveryWindowsCommand(a), newDeliveryWindowValidationCommand(a),
		newDeliveryWindowCountCommand(a), newDeliveryCycleCommand(a),
	}, deliveryTimeWrites(a, times))...)
}

// deliveryTimesPath is the path of the delivery times of the tenant.
const deliveryTimesPath = "/shipping/{tenant}/delivery-times"

// newDeliveryTimeSlotsCommand returns the group of the time slots of a
// delivery time of times: list and get, and their writes (slotWrites).
func newDeliveryTimeSlotsCommand(a *app, times resource[deliveryTime]) *cobra.Command {
	slotsOf := times.nested("slots")
	slots := child[deliveryTimeSlot]{
		use:   listUse,
		sub:   slotsOf.use,
		short: "List the time slots of a delivery time",
		long:  "Reads the time slots of a delivery time.",
		fetch: listing{read: (*emporix.Client).GetUnpaged},
		view: view[deliveryTimeSlot]{
			singular: "delivery time slot",
			header:   []string{columnID, columnMethod, columnFrom, columnTo, "CUT-OFF", "CUT-OFF DAYS", "CAPACITY"},
			legend: "METHOD is the shipping method of the slot (`epx shipping zones methods list`), FROM and TO\n" +
				"its delivery window in the delivery time's TIME ZONE, CUT-OFF the latest time of day to\n" +
				"order for the slot (sent as an ISO 8601 timestamp whose date Emporix ignores), CUT-OFF DAYS\n" +
				"how many days the delivery follows such an order (weekday delivery times only; otherwise\n" +
				"see DAY SHIFT in `epx delivery-times list`), and CAPACITY how many orders the slot takes.",
			row: func(s deliveryTimeSlot, _ language) []string {
				return []string{
					s.ID, s.ShippingMethod, s.DeliveryTimeRange.TimeFrom, s.DeliveryTimeRange.TimeTo,
					s.CutOffTime.Time, s.CutOffTime.CutOffDayShift.String(), s.Capacity.String(),
				}
			},
		},
	}
	return newChildGroup(a, slotsOf, slots, &childGroup{
		short: "Read and change the time slots of a delivery time",
		long:  "Reads and changes the time slots of a delivery time.",
		items: []*childItem{{
			use:   getUse,
			short: "Show one time slot of a delivery time",
			long:  "Reads one time slot of a delivery time.",
		}},
		writes: slotWrites(a, slotsOf, slots),
	})
}

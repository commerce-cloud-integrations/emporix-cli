package cli

import (
	"cmp"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// order is the part of an Order Service sales order shown in tables.
type order struct {
	ID        string `json:"id"`
	OrderType string `json:"orderType"`
	Status    string `json:"status"`
	Created   string `json:"created"`
	Customer  struct {
		personName
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"customer"`
	SiteCode        string            `json:"siteCode"`
	Entries         []json.RawMessage `json:"entries"` // counted only
	CalculatedPrice calculatedPrice   `json:"calculatedPrice"`
	Currency        string            `json:"currency"`
}

// buyer returns the customer's email, or else their name.
func (o *order) buyer() string {
	c := o.Customer
	return cmp.Or(c.Email, c.fullName(), c.Name)
}

// orderStatus is a status an order can move to.
type orderStatus struct {
	Status string `json:"status"`
}

// orderStatusChange is a status change in an order's history.
type orderStatusChange struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// orderHistoryItems finds the status changes of an order's history, which
// Emporix returns as the transitions of an object with its metadata; a
// response without them is not a history.
func orderHistoryItems(body json.RawMessage) ([]json.RawMessage, error) {
	var history struct {
		Transitions *[]json.RawMessage `json:"transitions"`
	}
	if err := json.Unmarshal(body, &history); err != nil || history.Transitions == nil {
		return nil, errors.New("expected an object with transitions")
	}
	return *history.Transitions, nil
}

// newOrdersCommand reads sales orders, which hold personal data, with their
// possible and past status changes.
func newOrdersCommand(a *app) *cobra.Command {
	orders := resource[order]{
		use:   "orders",
		short: "Read, create and change orders and their status",
		long: "Reads sales orders, personal data included. TYPE is MASTER_ORDER or SUB_ORDER for an order\n" +
			"split, for example by vendor: the master and its sub-orders are all listed, so do not add\n" +
			"up their totals. STATUS is " + alternatives(orderStatuses()) + ",\n" +
			"CUSTOMER the customer's email (or name), ITEMS the number of order entries and TOTAL the\n" +
			"gross final price in the order's currency (empty when Emporix sends only the deprecated\n" +
			"totalPrice, which -o json shows, as for an order that `epx orders add-entries --store`,\n" +
			"`recalculate` or `split` rewrote, and for sub-orders). The orders of a legal entity:\n" +
			"`epx orders list --query legalEntityId:ID`.",
		about: "Reads, creates and changes sales orders, personal data included, and their status:\n" +
			"`epx orders transition` moves an order to another status, `add-entries` and `recalculate`\n" +
			"change its entries, and `split` splits it into sub-orders by vendor.",
		path:    "/order-v2/{tenant}/salesorders",
		scopes:  []string{"order.order_read"},
		filters: listFilters{kinds: filterable},
		view: view[order]{
			singular: "order",
			header: []string{
				columnID, columnType, columnStatus, columnCreated, columnCustomer, columnSite, "ITEMS", columnTotal,
			},
			row: func(o order, _ language) []string {
				return []string{
					o.ID, o.OrderType, o.Status, o.Created, o.buyer(), o.SiteCode, strconv.Itoa(len(o.Entries)),
					o.CalculatedPrice.in(o.Currency),
				}
			},
		},
	}
	reads := []*cobra.Command{
		orders.searchCommand(a, "Reads the sales orders, personal data included, that an Emporix q filter finds\n"+
			"(--query; without it, every order).", queryField()),
		orders.countCommand(a, "Prints the number of sales orders, or of those that an Emporix q filter finds\n"+
			"(--query), which Emporix reports in a header without sending the orders: a master order\n"+
			"and each of its sub-orders count once, as list lists them. -o json prints the number alone.\n"+
			"Emporix counts 0 for a filter on id or customer.id although list finds the orders (seen\n"+
			"live): list them instead."),
		childCommand(a, orders, child[orderStatus]{
			use:   "transitions",
			short: "List the statuses an order can move to next",
			long:  "Reads the statuses that an order can move to next.",
			fetch: listing{read: (*emporix.Client).GetUnpaged},
			view: view[orderStatus]{
				singular: "order transition",
				header:   []string{columnStatus},
				row:      func(s orderStatus, _ language) []string { return []string{s.Status} },
			},
		}),
		// The history is one object holding the status changes, not a list.
		childCommand(a, orders, child[orderStatusChange]{
			use:   "history",
			sub:   "historical-transitions",
			short: "Show the status changes of an order",
			long: "Reads the status changes of an order, in the order Emporix sends them. -o json adds the\n" +
				"history's metadata.",
			fetch: single{items: orderHistoryItems},
			view: view[orderStatusChange]{
				singular: "order history",
				header:   []string{columnStatus, "TIMESTAMP"},
				row:      func(c orderStatusChange, _ language) []string { return []string{c.Status, c.Timestamp} },
			},
		}),
	}
	return newResourceCommand(a, orders, append(reads, commandsOf(a, orderWrites(orders))...)...)
}

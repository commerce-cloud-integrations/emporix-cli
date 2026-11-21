package cli

import (
	"net/url"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type shippingZone struct {
	ID      string                `json:"id"`
	Name    localized             `json:"name"`
	Default *bool                 `json:"default"`
	ShipTo  []shippingDestination `json:"shipTo"`
}

type shippingDestination struct {
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
}

// String shows the country, and the postal code unless it matches every code.
func (d shippingDestination) String() string {
	if d.PostalCode == "" || d.PostalCode == "*" {
		return d.Country
	}
	return d.Country + " " + d.PostalCode
}

type shippingMethod struct {
	ID              string        `json:"id"`
	Name            localized     `json:"name"`
	Active          *bool         `json:"active"`
	Fees            []shippingFee `json:"fees"`
	MaxOrderValue   money         `json:"maxOrderValue"`
	ShippingTaxCode string        `json:"shippingTaxCode"`
	ShippingGroupID string        `json:"shippingGroupId"`
}

type shippingFee struct {
	Cost            money  `json:"cost"`
	MinOrderValue   money  `json:"minOrderValue"`
	ShippingGroupID string `json:"shippingGroupId"`
}

// String shows the cost from its minimum order value, and the shipping group
// the fee is limited to.
func (f shippingFee) String() string {
	text := f.Cost.String()
	if minimum := f.MinOrderValue.String(); minimum != "" {
		text += " from " + minimum
	}
	if f.ShippingGroupID != "" {
		text += " (" + f.ShippingGroupID + ")"
	}
	return text
}

type shippingGroup struct {
	ID          string    `json:"id"`
	Description localized `json:"description"`
}

type shippingCustomer struct {
	CustomerID string `json:"customerId"`
	GroupID    string `json:"groupId"`
}

// Zones and methods were shown to ignore pageNumber and pageSize, and groups
// and customer assignments are documented without paging, so every shipping
// read uses GetUnpaged.
func shippingPath(collection string) string {
	return "/shipping/{tenant}/" + placeholderSite + "/" + collection
}

func newShippingCommand(a *app) *cobra.Command {
	return newGroup("shipping", "Read and change shipping zones, methods and groups, and calculate shipping fees",
		"Reads and changes a site's shipping configuration, calculates its shipping fees and finds the "+
			"sites that ship to a postal code. Every command but find-sites requires --site; for an "+
			"unknown site, the lists show nothing and the quotes fail.",
		newShippingZonesCommand(a), newShippingGroupsCommand(a), newShippingCustomersCommand(a),
		newShippingQuoteCommand(a), newShippingMinimumQuoteCommand(a), newShippingSlotQuoteCommand(a),
		newFindSitesCommand(a))
}

func newShippingZonesCommand(a *app) *cobra.Command {
	zones := resource[shippingZone]{
		use:   "zones",
		noun:  "shipping zones",
		short: "Read and change shipping zones and their methods",
		long: "Reads the shipping zones of a site. SHIPS TO lists the countries of the zone, each with " +
			"its postal codes when they are limited (a trailing * matches every code that starts that " +
			"way).",
		about: "Reads and changes the shipping zones of a site and the shipping methods of a zone.",
		under: "shipping",
		path:  shippingPath("zones"),
		read:  (*emporix.Client).GetUnpaged,
		view: view[shippingZone]{
			singular:  "shipping zone",
			header:    []string{columnID, columnName, columnDefault, columnShipsTo},
			localized: true,
			row: func(z shippingZone, lang language) []string {
				return []string{
					z.ID, z.Name.in(lang), formatBool(z.Default), joinEach(z.ShipTo, ", ", shippingDestination.String),
				}
			},
		},
	}
	return newResourceCommand(a, zones,
		slices.Concat([]*cobra.Command{newShippingMethodsCommand(a, zones)}, zoneWrites(a, zones))...)
}

func newShippingMethodsCommand(a *app, zones resource[shippingZone]) *cobra.Command {
	methodsOf := zones.nested("methods")
	methods := child[shippingMethod]{
		use:   listUse,
		sub:   methodsOf.use,
		short: "List the shipping methods of a zone",
		long:  "Reads the shipping methods of a zone.",
		// Emporix leaves out the fees of each listed method unless asked to
		// expand them, and sends those of one method without being asked.
		listQuery: url.Values{"expand": {"fees"}},
		fetch:     listing{read: (*emporix.Client).GetUnpaged},
		view: view[shippingMethod]{
			singular:  "shipping method",
			header:    []string{columnID, columnName, columnActive, "FEES", "MAX ORDER", columnTaxCode, columnGroup},
			localized: true,
			legend: "FEES lists each fee as its cost from a minimum order value, with the shipping group it is " +
				"limited to in parentheses; MAX ORDER is the highest order value the method accepts.",
			row: func(m shippingMethod, lang language) []string {
				return []string{
					m.ID, m.Name.in(lang), formatBool(m.Active), joinEach(m.Fees, ", ", shippingFee.String),
					m.MaxOrderValue.String(), m.ShippingTaxCode, m.ShippingGroupID,
				}
			},
		},
	}
	return newChildGroup(a, methodsOf, methods, &childGroup{
		short: "Read and change the shipping methods of a zone",
		long:  "Reads and changes the shipping methods of a zone of a site.",
		items: []*childItem{{
			use:   getUse,
			short: "Show one shipping method of a zone",
			long:  "Reads one shipping method of a zone.",
		}},
		writes: methodWrites(a, methodsOf, methods),
	})
}

func newShippingGroupsCommand(a *app) *cobra.Command {
	groups := resource[shippingGroup]{
		use:   groupsUse,
		noun:  "shipping groups",
		short: "Read and change shipping groups",
		long: "Reads the shipping groups of a site, which limit shipping fees to the customers assigned " +
			"to them (`epx shipping customers`). Emporix answers an unknown group with 409 Conflict, so " +
			"get fails with exit code 1.",
		about: "Reads and changes the shipping groups of a site, which limit shipping fees to the customers " +
			"assigned to them (`epx shipping customers`).",
		path: shippingPath("groups"),
		read: (*emporix.Client).GetUnpaged,
		view: view[shippingGroup]{
			singular:  "shipping group",
			header:    []string{columnID, columnDescription},
			localized: true,
			row:       func(g shippingGroup, lang language) []string { return []string{g.ID, g.Description.in(lang)} },
		},
	}
	return newResourceCommand(a, groups, groupWrites(a, groups)...)
}

func newShippingCustomersCommand(a *app) *cobra.Command {
	customers := resource[shippingCustomer]{
		use:   customersCommand,
		noun:  "shipping groups of customers",
		short: "Read and change which shipping group each customer belongs to",
		long: "Reads which shipping group each customer of a site belongs to. CUSTOMER is the customer " +
			"id. Emporix answers a customer without a shipping group with an empty response, so get " +
			"reports it as not found.",
		about: "Reads and changes which shipping group each customer of a site belongs to.",
		path:  shippingPath("cgrelations"),
		read:  (*emporix.Client).GetUnpaged,
		view: view[shippingCustomer]{
			singular: "customer's shipping group",
			header:   []string{columnCustomer, columnGroup},
			row:      func(c shippingCustomer, _ language) []string { return []string{c.CustomerID, c.GroupID} },
		},
	}
	return newResourceCommand(a, customers, customerGroupWrites(a, customers)...)
}

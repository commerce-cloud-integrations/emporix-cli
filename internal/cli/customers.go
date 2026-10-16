package cli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// customer is the part of a Customer Service customer shown in tables.
type customer struct {
	CustomerNumber string `json:"customerNumber"`
	personName
	ContactEmail  string `json:"contactEmail"` // its account's email is only in accounts, which epx does not expand
	Company       string `json:"company"`
	BusinessModel string `json:"businessModel"`
	Active        *bool  `json:"active"`
	OnHold        *bool  `json:"onHold"`
}

// customerAddress is the part of a customer address shown in tables.
type customerAddress struct {
	ID           string   `json:"id"`
	ContactName  string   `json:"contactName"`
	CompanyName  string   `json:"companyName"`
	Street       string   `json:"street"`
	StreetNumber string   `json:"streetNumber"`
	ZipCode      string   `json:"zipCode"`
	City         string   `json:"city"`
	Country      string   `json:"country"`
	IsDefault    *bool    `json:"isDefault"`
	Tags         []string `json:"tags"`
}

// customersCommand names the reads of customers: of the tenant, of a segment
// and of the shipping groups of a site.
const customersCommand = "customers"

// customers declares the customer profiles, whose numbers the
// CUSTOMER_NUMBER of another read completes with too.
func customers() resource[customer] {
	return resource[customer]{
		use:   customersCommand,
		short: "Read and change customers and their addresses",
		about: "Reads and changes customer profiles, personal data included, and their addresses,\n" +
			"imports customers with their passwords and sets until when imported passwords are\n" +
			"migrated.",
		long: "Reads customer profiles, personal data included. NUMBER is the customer number that get\n" +
			"takes, EMAIL the contact email (that of the account, with which the customer signs in,\n" +
			"may differ), BUSINESS MODEL is B2B (assigned to a company) or B2C, and ON HOLD marks a\n" +
			"suspended profile. B2B data and the default address are only in -o json; Emporix leaves\n" +
			"out a profile's addresses, accounts and mixins unless asked to expand them, which epx\n" +
			"does not: `epx customers addresses list` lists the addresses.",
		path:    "/customer/{tenant}/customers",
		scopes:  []string{"customer.customer_read"},
		filters: listFilters{kinds: filterable},
		read:    (*emporix.Client).GetAllHybrisCounted, // the total, as filtered, comes only as hybris-count
		view: view[customer]{
			singular: "customer",
			header:   []string{"NUMBER", columnName, columnEmail, columnCompany, "BUSINESS MODEL", columnActive, "ON HOLD"},
			row: func(c customer, _ language) []string {
				return []string{
					c.CustomerNumber, c.fullName(), c.ContactEmail,
					c.Company, c.BusinessModel, formatBool(c.Active), formatBool(c.OnHold),
				}
			},
		},
	}
}

// newCustomersCommand reads and changes customer profiles, which hold
// personal data, also those a search finds, imports them, and reads and
// changes the addresses of one and the password migration retention of
// imports.
func newCustomersCommand(a *app) *cobra.Command {
	profiles := customers()
	reads := []*cobra.Command{
		newCustomerAddressesCommand(a, profiles), newPasswordMigrationCommand(a), profiles.searchCommand(a,
			"Reads the customer profiles, personal data included, that an Emporix q filter finds\n"+
				"(--query; without it, every profile).", queryField()),
	}
	return newResourceCommand(a, profiles, slices.Concat(reads, customerWrites(a, profiles))...)
}

// newCustomerAddressesCommand returns the group of the addresses of a
// customer of customers: list and get, and the writes that change them
// (addressWrites).
func newCustomerAddressesCommand(a *app, customers resource[customer]) *cobra.Command {
	in := customers.nested("addresses")
	// A customer's addresses are documented without paging.
	addresses := child[customerAddress]{
		use:   listUse,
		sub:   in.use,
		short: "List the addresses of a customer",
		long:  "Reads the addresses of a customer, personal data included.",
		fetch: listing{read: (*emporix.Client).GetUnpaged},
		view: view[customerAddress]{
			singular: "customer address",
			header: []string{
				columnID, "CONTACT", columnCompany, "STREET", "ZIP", "CITY", columnCountry, columnDefault, "TAGS",
			},
			legend: "DEFAULT marks the default address; TAGS are labels such as BILLING or SHIPPING. Phone\n" +
				"numbers, states and extra address lines are only in -o json.",
			row: func(d customerAddress, _ language) []string {
				return []string{
					d.ID, d.ContactName, d.CompanyName, spaced(d.Street, d.StreetNumber), d.ZipCode, d.City, d.Country,
					formatBool(d.IsDefault), strings.Join(d.Tags, ","),
				}
			},
		},
	}
	return newChildGroup(a, in, addresses, &childGroup{
		short: "Read and change the addresses of a customer",
		long:  "Reads and changes the addresses of a customer, personal data included, and their tags.",
		items: []*childItem{{
			use: getUse, short: "Show one address of a customer",
			long: "Reads one address of a customer, personal data included.",
		}},
		writes: addressWrites(a, in, addresses),
	})
}

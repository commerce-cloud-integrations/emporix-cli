package cli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type priceList struct {
	ID             string    `json:"id"`
	Name           localized `json:"name"`
	Currency       string    `json:"currency"`
	SiteCode       string    `json:"siteCode"`
	Countries      []string  `json:"countries"`
	Regions        []string  `json:"regions"`
	CustomerGroups []string  `json:"customerGroups"`
	LegalEntityID  string    `json:"legalEntityId"`
	Validity       validity  `json:"validity"`
}

type priceListPrice struct {
	ID           string     `json:"id"`
	ItemID       priceItem  `json:"itemId"`
	PriceModelID string     `json:"priceModelId"`
	TierValues   priceTiers `json:"tierValues"`
}

func newPriceListsCommand(a *app) *cobra.Command {
	lists := resource[priceList]{
		use:   "price-lists",
		short: "Read and change price lists and their prices",
		long: "Reads price lists. A price list holds prices in one currency for one site that apply " +
			"between FROM and TO to its countries or regions, and to its customer groups (a preview " +
			"feature of Emporix) or to one legal entity and its subsidiaries. Empty COUNTRIES and " +
			"REGIONS apply everywhere; empty CUSTOMER GROUPS and LEGAL ENTITY apply to every customer.",
		about: "Reads and changes price lists and, in `prices`, the prices of one. A price list holds " +
			"prices in one currency for one site that apply between their validity dates to its " +
			"countries or regions, and to its customer groups or to one legal entity.",
		path:    "/price/{tenant}/price-lists",
		scopes:  []string{"price.pricelist_read"},
		filters: listFilters{kinds: filterable},
		view: view[priceList]{
			singular: "price list",
			header: []string{
				columnID, columnName, columnCurrency, columnSite, "COUNTRIES", "REGIONS", "CUSTOMER GROUPS",
				columnLegalEntity, columnFrom, columnTo,
			},
			localized: true,
			row: func(l priceList, lang language) []string {
				return []string{
					l.ID, l.Name.in(lang), l.Currency, l.SiteCode, strings.Join(l.Countries, ","),
					strings.Join(l.Regions, ","), strings.Join(l.CustomerGroups, ","), l.LegalEntityID,
					l.Validity.From, l.Validity.To,
				}
			},
		},
	}
	return newResourceCommand(a, lists, slices.Concat([]*cobra.Command{lists.querySearchCommand(a)},
		priceListWrites(a, lists), []*cobra.Command{newPriceListPricesCommand(a, lists)})...)
}

func priceListPrices(pricesOf resource[priceList]) child[priceListPrice] {
	return child[priceListPrice]{
		use:     listUse,
		sub:     pricesOf.use,
		short:   "List the prices of a price list",
		long:    "Reads the prices of a price list, in its currency.",
		fetch:   listing{filters: listFilters{kinds: filterable}},
		unknown: listsNothing,
		view: view[priceListPrice]{
			singular: "price list price",
			header:   []string{columnID, columnItem, columnType, "MODEL", columnTiers},
			legend: "TIERS lists each tier value as sent; `-o json` pairs it with the id of its tier in the price " +
				"model MODEL (`epx price-models get MODEL`; an empty MODEL is the default model).",
			row: func(p priceListPrice, _ language) []string {
				return []string{p.ID, p.ItemID.ID, p.ItemID.ItemType, p.PriceModelID, p.TierValues.String()}
			},
		},
	}
}

func newPriceListPricesCommand(a *app, lists resource[priceList]) *cobra.Command {
	pricesOf := lists.nested("prices")
	prices := priceListPrices(pricesOf)
	reads := []*cobra.Command{
		childCommand(a, pricesOf, prices),
		childSearchCommand(a, pricesOf, prices, "Search the prices of a price list",
			"Reads the prices of a price list, in its currency, that an Emporix q filter finds "+
				"(--query; without it, every one).", queryField()),
		childItemCommand(a, pricesOf, prices, &childItem{
			use:   getUse,
			short: "Show one price of a price list",
			long:  "Reads one price of a price list, in its currency.",
		}),
	}
	return newGroup(pricesOf.use, "Read and change the prices of a price list",
		"Reads and changes the prices of a price list, in its currency. "+lists.argumentSource()+".",
		slices.Concat(reads, priceListPriceWrites(a, pricesOf, prices))...)
}

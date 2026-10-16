package cli

import (
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// site is the part of a Site Settings Service site shown in tables.
type site struct {
	Code            string       `json:"code"`
	Name            string       `json:"name"`
	Default         *bool        `json:"default"` // nil when the service omits it
	Languages       []string     `json:"languages"`
	Currency        string       `json:"currency"`
	HomeBase        siteHomeBase `json:"homeBase"`
	ShipToCountries []string     `json:"shipToCountries"`
}

// siteHomeBase is the home base of a site, which places it by its address.
type siteHomeBase struct {
	Address siteAddress `json:"address"`
}

// siteAddress is the address of a site's home base; its country is shown.
type siteAddress struct {
	Country string `json:"country"`
}

// siteManageScope is the write-capable scope without which Emporix leaves out
// inactive sites.
const siteManageScope = "site.site_manage"

// siteReadScope is the scope that the codes of inactive sites need beside siteManageScope.
const siteReadScope = "site.site_read"

// sites declares the sites, which every --site completes with
// (completeSites).
func sites() resource[site] {
	return resource[site]{
		use:   "sites",
		short: "Read, create and change sites and their mixins",
		long: "Reads active sites. Inactive sites are not included: Emporix returns them only with the\n" +
			"write-capable " + siteManageScope + " scope, so `get CODE` of an inactive site is refused (403,\n" +
			"exit 3) without it. `epx sites codes` lists the codes of every site, inactive ones too (opt-in);\n" +
			"to read inactive sites whole, run\n" +
			"`epx api get '/site/{tenant}/sites?includeInactive=true' --scope " + siteManageScope + "`.",
		path:    "/site/{tenant}/sites",
		filters: listFilters{kinds: filterable},
		read:    (*emporix.Client).GetAllHybrisCounted, // the total comes only as hybris-count
		view: view[site]{
			singular: "site",
			header:   []string{columnCode, columnName, columnDefault, "LANGUAGES", columnCurrency, "HOME", columnShipsTo},
			row: func(s site, _ language) []string {
				return []string{
					s.Code, s.Name, formatBool(s.Default), strings.Join(s.Languages, ","), s.Currency,
					s.HomeBase.Address.Country, strings.Join(s.ShipToCountries, ","),
				}
			},
		},
	}
}

// newSitesCommand reads and changes site settings and the mixins of a site.
// Inactive sites need the write-capable site.site_manage scope, so only the
// read of every site's code, which exists to show them, requests it, with
// site.site_read (opt-in).
func newSitesCommand(a *app) *cobra.Command {
	r := sites()
	codes := reader[string]{view: stringsView("site code", columnCode), scopes: []string{siteReadScope, siteManageScope}}
	reads := []*cobra.Command{
		newSiteMixinsCommand(a, r),
		codes.command(a, &command{
			use:   "codes",
			short: "List the codes of every site, inactive ones too",
			long: "Reads the code of every site, active or not. Emporix lists inactive sites only with the\n" +
				"write-capable " + siteManageScope + " and " + siteReadScope + " scopes together (each alone\n" +
				"answers 403), so this read is opt-in; `epx sites list` shows active sites whole without a scope.\n" +
				"Emporix sends no total for the codes, so their pages are read until an empty one.",
			req:   request{path: "/site/{tenant}/siteslist", query: url.Values{"includeInactive": {queryTrue}}},
			fetch: listing{filters: listFilters{kinds: filterable}},
		}),
	}
	return newResourceCommand(a, r, slices.Concat(reads, siteWrites(a, r))...)
}

// siteWrites returns the writes of sites, with the scope that the reference
// names for them: create, update CODE (which creates no missing site), patch
// CODE and delete CODE.
func siteWrites(a *app, sites resource[site]) []*cobra.Command {
	manage := []string{siteManageScope}
	item := request{path: sites.itemPath()}
	const unknown = "\nAn unknown CODE is not found (exit 4)."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a site", method: http.MethodPost,
			long: "Creates a site from the JSON in --file, shaped as `epx sites get CODE -o json` prints one.\n" +
				"The reference requires code, name, defaultLanguage, languages, currency, homeBase (an\n" +
				"address with country and zipCode), shipToCountries and taxDeterminationBasedOn\n" +
				"(BILLING_ADDRESS or SHIPPING_ADDRESS); active, default, includesTax,\n" +
				"availableCurrencies, payment, shipping, tax, mixins and cartCalculationScale are optional.",
			req: request{path: sites.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: sites.itemUse("update"), short: "Replace a site", method: http.MethodPut,
			long: "Replaces the site CODE with the JSON in --file, shaped as create takes it: a member that it\n" +
				"leaves out is removed. It changes the payment, shipping and tax configurations that the body\n" +
				"gives too; Emporix changes an inactive one only with --expand naming it, e.g. payment:all." +
				unknown,
			req: item, scopes: manage, body: &payload{}, ids: sites.itemIDs(),
			params: queryFlags{optional: []paramFlag{{param: expandParam, flag: expandParam,
				usage: "the `SETTINGS` to expand, e.g. payment:all,shipping:all,tax:all (all, active or none each)"}}},
		}).command(a),
		(&write{
			use: sites.itemUse("patch"), short: "Change members of a site", method: http.MethodPatch,
			long: "Changes the members of the site CODE that the JSON in --file gives, such as\n" +
				`{"languages": ["en", "nl"]}; the others keep their values.` + unknown,
			req: item, scopes: manage, body: &payload{}, ids: sites.itemIDs(),
		}).command(a),
		(&write{
			use: sites.itemUse("delete"), short: "Delete a site", method: http.MethodDelete,
			long: "Deletes the site CODE; it sends no body. Emporix refuses to delete the default site." +
				unknown,
			req: item, scopes: manage, body: noBody{}, ids: sites.itemIDs(),
		}).command(a),
	}
}

// newSiteMixinsCommand returns the group of the mixins of a site of sites,
// the custom settings that extend it: list SITE_CODE and get SITE_CODE NAME,
// which Emporix serves without a scope (live), and the writes, which need
// site.site_manage (siteMixinWrites).
func newSiteMixinsCommand(a *app, sites resource[site]) *cobra.Command {
	mixinsOf := sites.nested("mixins")
	mixins := child[namedValue]{
		use:   listUse,
		sub:   mixinsOf.use,
		short: "List the mixins of a site",
		long: "Reads the mixins of a site, the custom settings that extend it (such as coupon or customer\n" +
			"settings), by NAME, each VALUE cut (-o json prints them whole). A site without mixins lists none.",
		// Emporix documents site.site_manage for mixins, but serves them without a scope (live), so
		// none is requested, whatever the scopes of sites.
		scopes: []string{},
		// Emporix sends an empty body for a site without mixins, which is no mixin, not a missing site.
		fetch: single{items: members, read: emptyAs("{}")},
		view: view[namedValue]{
			singular: "site mixin",
			header:   []string{columnName, columnValue},
			long:     []string{columnValue},
			row:      func(m namedValue, _ language) []string { return []string{m.Name, valueCell(m.Value)} },
		},
	}
	return newChildGroup(a, mixinsOf, mixins, &childGroup{
		short: "Read and change the mixins of a site",
		long: "Reads the mixins of a site, the custom settings that extend it, and creates, changes and\n" +
			"deletes them. " + sites.argumentSource() + ".",
		items: []*childItem{{
			use:   getUse,
			short: "Show one mixin of a site",
			long:  "Reads one mixin of a site: its fields by NAME, each VALUE cut (-o json prints them whole).",
			items: members, // a mixin is an object too: the table shows its fields
		}},
		writes: siteMixinWrites(a, mixinsOf, mixins),
	})
}

// siteMixinWrites returns the writes of the mixins of a site, below
// mixinsOf, with the scope that the reference names for them: create
// SITE_CODE, which adds mixins, and update, patch and delete SITE_CODE NAME,
// which address one mixin as get reads it.
func siteMixinWrites(a *app, mixinsOf resource[site], mixins child[namedValue]) []*cobra.Command {
	siteCode := mixinsOf.argument()
	mixin := func(verb string) childTarget { return childItemTarget(mixinsOf, mixins, verb, nil) }
	get := mixin(getUse)
	item, ids := request{path: get.path}, childItemIDs(mixinsOf, mixins, &get)
	manage := []string{siteManageScope}
	const unknown = "\nAn unknown site or mixin is not found (exit 4)."
	return []*cobra.Command{
		(&write{
			use: createUse + " " + siteCode, short: "Add mixins to a site", method: http.MethodPost,
			long: "Adds mixins to the site " + siteCode + ": the JSON object in --file maps the name of each to its\n" +
				`fields, e.g. {"generalShopSettings": {"logoUrl": "https://shop.acme.example/logo.png"}}.` + "\n" +
				"Emporix checks each against the schema that its metadata.schema names; a mixin that the\n" +
				"site has already is a conflict (exit 5).",
			req: request{path: mixins.below(mixinsOf.path)}, scopes: manage, body: &payload{},
			ids: argumentIDs{siteCode: mixinsOf.ids()},
		}).command(a),
		(&write{
			use: mixin("update").use, short: "Replace a mixin of a site", method: http.MethodPut,
			long: "Replaces the mixin NAME of the site " + siteCode + " with the JSON object in --file; the\n" +
				"reference's example gives it as create does, in an object named after the mixin." + unknown,
			req: item, scopes: manage, body: &payload{}, ids: ids,
		}).command(a),
		(&write{
			use: mixin("patch").use, short: "Change fields of a mixin of a site",
			method: http.MethodPatch,
			long: "Changes the fields of the mixin NAME of the site " + siteCode + " that the JSON object in\n" +
				`--file gives, such as {"active": true}; the others keep their values.`,
			req: item, scopes: manage, body: &payload{}, ids: ids,
		}).command(a),
		(&write{
			use: mixin("delete").use, short: "Delete a mixin of a site", method: http.MethodDelete,
			long: "Deletes the mixin NAME of the site " + siteCode + "; it sends no body." + unknown,
			req:  item, scopes: manage, body: noBody{}, ids: ids,
		}).command(a),
	}
}

package cli

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// countriesPage holds every country a tenant can have (ISO 3166 lists 249).
const countriesPage = 1000

// country is the part of a Country Service country shown in tables.
type country struct {
	Code    string    `json:"code"`
	Name    localized `json:"name"`
	Active  *bool     `json:"active"`
	Regions []string  `json:"regions"`
}

// newCountriesCommand reads the countries known to the tenant, active or not,
// and activates or deactivates one (patch). The Country Service orders
// countries differently on every request, even when sorted, so its pages
// overlap; list reads them as one page instead, sorted by code so the output
// is the same on every run (a --sort of the user's replaces it).
func newCountriesCommand(a *app) *cobra.Command {
	r := resource[country]{
		use:   "countries",
		short: "Read countries, and activate or deactivate them",
		long: "Reads all countries known to the tenant, active or not (ACTIVE shows which ones the\n" +
			"tenant uses). REGIONS lists region codes (see `epx regions`).",
		path:      "/country/{tenant}/countries",
		listQuery: url.Values{sortParameter: {"code"}},
		filters:   listFilters{kinds: sortable},
		read:      inOnePage(countriesPage),
		view: view[country]{
			singular:  "country",
			header:    []string{columnCode, columnName, columnActive, "REGIONS"},
			localized: true,
			row: func(c country, lang language) []string {
				return []string{c.Code, c.Name.in(lang), formatBool(c.Active), strings.Join(c.Regions, ",")}
			},
		},
	}
	return newResourceCommand(a, r, countryWrites(a, r)...)
}

// countryWrites returns the write of countries, with the scope that the
// reference names for it: patch CODE, which activates or deactivates a
// country, the only change that Emporix allows, from typed flags or --file,
// with the version that Emporix requires.
func countryWrites(a *app, countries resource[country]) []*cobra.Command {
	return []*cobra.Command{(&write{
		use:   countries.itemUse("patch"),
		short: "Activate or deactivate a country",
		long: "Activates the country CODE for the tenant (--active) or deactivates it (--active=false):\n" +
			"Emporix lets a PATCH change active only. epx takes the country's metadata.version as read\n" +
			"(--version, as `epx countries get CODE -o json` prints it), which the reference asks for;\n" +
			"Emporix does not check it, so a stale version is accepted and applied. --file gives the JSON body\n" +
			`instead, e.g. {"active": true, "metadata": {"version": 1}}. An unknown CODE is not found` + "\n" +
			"(exit 4).",
		example: "  epx countries patch NL --active --version 1",
		method:  http.MethodPatch,
		req:     request{path: countries.itemPath()},
		ids:     countries.itemIDs(),
		scopes:  []string{"country.country_manage"},
		body: &payload{typed: objectBody{
			switchField("active", "active", "activate the country, or with --active=false deactivate it").require(),
			versionField("country").require(),
		}},
	}).command(a)}
}

// newRegionsCommand reads the regions countries belong to. Unsorted, the
// Country Service orders them differently on every request too, but unlike
// countries it sorts regions before paging, so list pages them sorted by code
// (a --sort of the user's replaces it).
func newRegionsCommand(a *app) *cobra.Command {
	return newResourceCommand(a, resource[codeName]{
		use:       "regions",
		long:      "Reads the regions that group countries.",
		path:      "/country/{tenant}/regions",
		listQuery: url.Values{sortParameter: {"code"}},
		filters:   listFilters{kinds: sortable},
		view:      codeNameView("region"),
	})
}

package cli

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// countriesPage holds every country a tenant can have (ISO 3166 lists 249).
const countriesPage = 1000

type country struct {
	Code    string    `json:"code"`
	Name    localized `json:"name"`
	Active  *bool     `json:"active"`
	Regions []string  `json:"regions"`
}

// The Country Service orders countries differently on every request, even when
// sorted, so its pages overlap.
func newCountriesCommand(a *app) *cobra.Command {
	r := resource[country]{
		use:   "countries",
		short: "Read countries, and activate or deactivate them",
		long: "Reads all countries known to the tenant, active or not (ACTIVE shows which ones the " +
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

func countryWrites(a *app, countries resource[country]) []*cobra.Command {
	return []*cobra.Command{(&write{
		use:   countries.itemUse("patch"),
		short: "Activate or deactivate a country",
		long: "Activates the country CODE for the tenant (--active) or deactivates it (--active=false): " +
			"Emporix lets a PATCH change active only. epx takes the country's metadata.version as read " +
			"(--version, as `epx countries get CODE -o json` prints it), which the reference asks for; " +
			"Emporix does not check it, so a stale version is accepted and applied. --file gives the JSON body " +
			"instead, e.g. " + literal(`{"active": true, "metadata": {"version": 1}}`) + ". An unknown CODE is not found " +
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

// Unsorted, the Country Service orders regions differently on every request too,
// but unlike countries it sorts regions before paging.
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

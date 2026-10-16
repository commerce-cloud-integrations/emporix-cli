package cli

import (
	"net/http"
	"testing"
)

// siteJSON is a site as create takes it.
const siteJSON = `{"code":"AcmeStore_BE","name":"AcmeStore BE","defaultLanguage":"nl","languages":["nl","fr"],` +
	`"currency":"EUR","homeBase":{"address":{"country":"BE","zipCode":"1000"}},"shipToCountries":["BE"],` +
	`"taxDeterminationBasedOn":"BILLING_ADDRESS"}`

// tenantSetupWrites are the writes of sites and their mixins, countries,
// currencies, exchange rates and tax configurations, each with the manage
// scope that the reference names for its service.
func tenantSetupWrites() []writeCase {
	return append(siteWriteCases(), currencyAndTaxWrites()...)
}

// siteWriteCases are the writes of sites, of their mixins and of countries.
func siteWriteCases() []writeCase {
	sites, country := []string{siteManageScope}, []string{"country.country_manage"}
	const mixin = "/site/acme1/sites/AcmeStore_NL/mixins/generalShopSettings"
	return []writeCase{
		{
			args: []string{"sites", "create"}, file: siteJSON, method: http.MethodPost, path: "/site/acme1/sites",
			scopes: sites, status: http.StatusCreated, body: `{"id":"AcmeStore_BE"}`,
			want: "{\n  \"id\": \"AcmeStore_BE\"\n}\n",
		},
		{
			args: []string{"sites", "update", "AcmeStore_BE", "--expand", "payment:all,tax:all"}, file: siteJSON,
			method: http.MethodPut, path: "/site/acme1/sites/AcmeStore_BE", query: "expand=payment%3Aall%2Ctax%3Aall",
			scopes: sites, status: http.StatusOK,
		},
		{
			// The CODE is escaped as a path segment.
			args: []string{"sites", "patch", "AcmeStore NL"}, file: `{"languages":["en","nl"]}`,
			method: http.MethodPatch, path: "/site/acme1/sites/AcmeStore%20NL", scopes: sites, status: http.StatusOK,
		},
		{
			args: []string{"sites", "delete", "AcmeStore_BE"}, method: http.MethodDelete,
			path: "/site/acme1/sites/AcmeStore_BE", scopes: sites, confirmBy: "AcmeStore_BE", status: http.StatusNoContent,
		},
		{
			args: []string{"sites", "mixins", "create", "AcmeStore_NL"},
			file: `{"generalShopSettings":{"logoUrl":"https://shop.acme.example/logo.png"}}`, method: http.MethodPost,
			path: "/site/acme1/sites/AcmeStore_NL/mixins", scopes: sites, status: http.StatusCreated,
			body: `{"id":"generalShopSettings"}`, want: "{\n  \"id\": \"generalShopSettings\"\n}\n",
		},
		{
			args: []string{"sites", "mixins", "update", "AcmeStore_NL", "generalShopSettings"},
			file: `{"generalShopSettings":{"logoUrl":"https://shop.acme.example/logo.svg"}}`, method: http.MethodPut,
			path: mixin, scopes: sites, status: http.StatusOK,
		},
		{
			args: []string{"sites", "mixins", "patch", "AcmeStore_NL", "generalShopSettings"},
			file: `{"active":true}`, method: http.MethodPatch, path: mixin, scopes: sites, status: http.StatusOK,
		},
		{
			args:   []string{"sites", "mixins", "delete", "AcmeStore_NL", "generalShopSettings"},
			method: http.MethodDelete, path: mixin, scopes: sites, confirmBy: "generalShopSettings",
			status: http.StatusNoContent,
		},
		{
			// The flags make the body: deactivate the country at the version read.
			args: []string{"countries", "patch", "NL", "--active=false", "--version", "3"},
			sent: `{"active":false,"metadata":{"version":3}}`, method: http.MethodPatch,
			path: "/country/acme1/countries/NL", scopes: country,
			status: http.StatusNoContent,
		},
		{
			args: []string{"countries", "patch", "BE"}, file: `{"active":true,"metadata":{"version":1}}`,
			method: http.MethodPatch, path: "/country/acme1/countries/BE", scopes: country,
			status: http.StatusNoContent,
		},
	}
}

// The tax configuration of a country, as create takes it and, with the
// version read, as update takes it.
const (
	taxLocationJSON = `{"location":{"countryCode":"NL"},"taxClasses":[{"code":"STANDARD","name":{"en":"Standard"},` +
		`"order":0,"rate":21,"isDefault":true}]`
	taxJSON          = taxLocationJSON + `}`
	taxVersionedJSON = taxLocationJSON + `,"metadata":{"version":1}}`
)

// currencyAndTaxWrites are the writes of currencies, exchange rates and tax
// configurations.
func currencyAndTaxWrites() []writeCase {
	currency, tax := []string{currencyManageScope}, []string{"tax.tax_manage"}
	return []writeCase{
		{
			args: []string{"currencies", "create"}, file: `{"code":"CHF","name":{"en":"Swiss franc"}}`,
			method: http.MethodPost, path: "/currency/acme1/currencies", scopes: currency, lang: "*",
			status: http.StatusCreated, body: `{"code":"CHF"}`, want: "{\n  \"code\": \"CHF\"\n}\n",
		},
		{
			args: []string{"currencies", "update", "CHF", "--content-language", "en"},
			file: `{"name":"Swiss franc","metadata":{"version":1}}`, method: http.MethodPut,
			path: "/currency/acme1/currencies/CHF", scopes: currency, lang: "en", status: http.StatusNoContent,
		},
		{
			args: []string{"currencies", "delete", "CHF"}, method: http.MethodDelete,
			path: "/currency/acme1/currencies/CHF", scopes: currency, confirmBy: "CHF", status: http.StatusNoContent,
		},
		{
			// The rate is sent as a string, as the reference types it.
			args: []string{"exchange-rates", "create", "--source", "EUR", "--target", "CHF", "--rate", "0.95"},
			sent: `{"rate":"0.95","sourceCurrency":"EUR","targetCurrency":"CHF"}`, method: http.MethodPost,
			path: "/currency/acme1/exchanges", scopes: currency, status: http.StatusCreated, body: `{"code":"EUR_CHF"}`,
			want: "{\n  \"code\": \"EUR_CHF\"\n}\n",
		},
		{
			args: []string{"exchange-rates", "create"}, file: `{"sourceCurrency":"EUR","targetCurrency":"USD","rate":"1.17"}`,
			method: http.MethodPost, path: "/currency/acme1/exchanges", scopes: currency, status: http.StatusCreated,
		},
		{
			args: []string{"exchange-rates", "update", "EUR_CHF", "--rate", "0.96", "--version", "2"},
			sent: `{"metadata":{"version":2},"rate":"0.96"}`, method: http.MethodPut,
			path: "/currency/acme1/exchanges/EUR_CHF", scopes: currency, status: http.StatusNoContent,
		},
		{
			args: []string{"exchange-rates", "update", "EUR_CHF"}, file: `{"rate":"0.97","metadata":{"version":3}}`,
			method: http.MethodPut, path: "/currency/acme1/exchanges/EUR_CHF", scopes: currency,
			status: http.StatusNoContent,
		},
		{
			args: []string{"exchange-rates", "delete", "EUR_CHF"}, method: http.MethodDelete,
			path: "/currency/acme1/exchanges/EUR_CHF", scopes: currency, confirmBy: "EUR_CHF",
			status: http.StatusNoContent,
		},
		{
			args: []string{"taxes", "create"}, file: taxJSON, method: http.MethodPost, path: "/tax/acme1/taxes",
			scopes: tax, lang: "*", status: http.StatusCreated, body: `{"locationCode":"NL"}`,
			want: "{\n  \"locationCode\": \"NL\"\n}\n",
		},
		{
			args: []string{"taxes", "update", "NL"}, file: taxVersionedJSON,
			method: http.MethodPut, path: "/tax/acme1/taxes/NL", scopes: tax, lang: "*", status: http.StatusNoContent,
		},
		{
			args: []string{"taxes", "delete", "NL"}, method: http.MethodDelete, path: "/tax/acme1/taxes/NL",
			scopes: tax, confirmBy: "NL", status: http.StatusNoContent,
		},
	}
}

// TestTenantSetupWritesRefuseInvalidInput checks that a write of the tenant
// setup without its body, the flags that make it or its arguments, with
// flags that --file excludes or it does not take, or with an ID that
// addresses another operation, is a usage error (exit 2) that sends nothing,
// also for a dry run.
func TestTenantSetupWritesRefuseInvalidInput(t *testing.T) {
	const country = `{"active":true,"metadata":{"version":1}}`
	assertRefused(t, []refusedWrite{
		{args: []string{"sites", "create"}, want: `required flag(s) "file" not set`},
		{args: []string{"sites", "delete"}, want: "accepts 1 arg(s), received 0"},
		{args: []string{"sites", "update", "AcmeStore_NL", "--expand", ""}, file: siteJSON,
			want: `invalid argument "" for "--expand" flag: empty`},
		{args: []string{"sites", "delete", "AcmeStore_NL"}, file: country, want: "unknown flag: --file"},
		{args: []string{"sites", "mixins", "update", "AcmeStore_NL"}, file: country,
			want: "accepts 2 arg(s), received 1"},
		{args: []string{"sites", "mixins", "delete", "AcmeStore_NL", ".."}, want: `invalid NAME ".."`},
		{args: []string{"countries", "patch", "NL", "--version", "1"},
			want: "at least one of the flags in the group [file active] is required"},
		{args: []string{"countries", "patch", "NL", "--active"},
			want: "at least one of the flags in the group [file version] is required"},
		{args: []string{"countries", "patch", "NL", "--active", "--version", "one"},
			want: `invalid argument "one" for "--version" flag: use the version as read, a whole number such as 3`},
		{args: []string{"countries", "patch", "NL", "--active", "--version", "1.5"},
			want: `invalid argument "1.5" for "--version" flag: use the version as read`},
		{args: []string{"countries", "patch", "NL", "--active", "--version", "0"},
			want: `invalid argument "0" for "--version" flag: use the version as read`},
		{args: []string{"countries", "patch", "NL", "--active"}, file: country,
			want: "if any flags in the group [file active] are set none of the others can be"},
		{args: []string{"exchange-rates", "create", "--source", "EUR", "--rate", "0.95"},
			want: "at least one of the flags in the group [file target] is required"},
		{args: []string{"exchange-rates", "update", "EUR_CHF", "--rate", "0.95"},
			want: "at least one of the flags in the group [file version] is required"},
		{args: []string{"exchange-rates", "update", "EUR_CHF", "--rate", "0.95", "--version=-2"},
			want: `invalid argument "-2" for "--version" flag: use the version as read`},
		{args: []string{"currencies", "create", "--content-language", "Swiss"}, file: country,
			want: `invalid argument "Swiss" for "--content-language" flag`},
		// forceDelete is not sent: the reference does not say what it changes.
		{args: []string{"currencies", "delete", "CHF", "--force-delete"}, want: "unknown flag: --force-delete"},
		{args: []string{"taxes", "update", taxCalculationSegment}, file: taxJSON,
			want: `invalid LOCATION "calculation-commands": it would be sent as calculate, not to one item`},
	})
}

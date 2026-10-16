package cli

// The read cases of the Site Settings, Currency, Country, Tax, Configuration and Indexing services (see TestReads).

// sitesBody holds two sites; the second has mixins, which only -o json shows.
const sitesBody = `[
	{"code":"main","name":"Default","active":true,"default":false,"languages":["en","de"],"currency":"EUR",
	 "homeBase":{"address":{"country":"DE"}},"shipToCountries":["DE","AT"]},
	{"code":"AcmeStore_NL","name":"AcmeStore","active":true,"default":true,"languages":["nl"],"currency":"EUR",
	 "homeBase":{"address":{"country":"NL"}},"shipToCountries":["NL"],"mixins":{"apiKey":"site-secret"}}
]`

// siteReads returns the Site Settings Service reads.
func siteReads() []readCase {
	return []readCase{
		{
			// Sites send their total only as hybris-count, when asked with totalCount=true.
			args:    []string{"sites", "list"},
			path:    "/site/acme1/sites",
			mode:    getAllHybrisCounted,
			filters: listFilters{kinds: filterable},
			body:    sitesBody,
			want: "CODE          NAME       DEFAULT  LANGUAGES  CURRENCY  HOME  SHIPS TO\n" +
				"main          Default    false    en,de      EUR       DE    DE,AT\n" +
				"AcmeStore_NL  AcmeStore  true     nl         EUR       NL    NL\n",
		},
		{
			args: []string{"sites", "get", "AcmeStore_NL"},
			path: "/site/acme1/sites/AcmeStore_NL",
			mode: getWhole,
			body: `{"code":"AcmeStore_NL","name":"AcmeStore","active":true,"default":true,"languages":["nl"],"currency":"EUR",
				"homeBase":{"address":{"country":"NL"}},"shipToCountries":["NL"]}`,
			want: "CODE          NAME       DEFAULT  LANGUAGES  CURRENCY  HOME  SHIPS TO\n" +
				"AcmeStore_NL  AcmeStore  true     nl         EUR       NL    NL\n",
		},
		{
			// A boolean the service leaves out stays empty instead of looking like false.
			args: []string{"sites", "get", "x"},
			path: "/site/acme1/sites/x",
			mode: getWhole,
			body: `{"code":"x","name":"X"}`,
			want: "CODE  NAME  DEFAULT  LANGUAGES  CURRENCY  HOME  SHIPS TO\n" +
				"x     X                                         \n",
		},
		{
			// Opt-in: inactive sites need the write-capable scope. Emporix sends no total for the codes.
			args:    []string{"sites", "codes", "--allow-write-scope"},
			path:    "/site/acme1/siteslist",
			query:   "includeInactive=true",
			scopes:  []string{"site.site_read", "site.site_manage"},
			filters: listFilters{kinds: filterable},
			body:    `["main","AcmeStore_NL","AcmeStore_old"]`,
			want:    "CODE\nmain\nAcmeStore_NL\nAcmeStore_old\n",
		},
		{
			// The mixins are the members of one object, sorted by name; a value is kept as sent (text as
			// is, other JSON compacted, null empty) and cut.
			args: []string{"sites", "mixins", "list", "AcmeStore NL"},
			path: "/site/acme1/sites/AcmeStore%20NL/mixins",
			mode: getWhole,
			body: `{"orderSettings":{"minimumOrderValue":20,"metadata":{"schema":"https://example.com/orderSettings.v1"}},
				"customerSettings":{"regExSettings":{"name":"^[a-z&<>]{0,70}$"}},"note":"plain text","none":null}`,
			want: "NAME              VALUE\n" +
				"customerSettings  {\"regExSettings\":{\"name\":\"^[a-z&<>]{0,70}$\"}}\n" +
				"none              \n" +
				"note              plain text\n" +
				"orderSettings     {\"minimumOrderValue\":20,\"metadata\":{\"schema\":\"https://examp…\n",
		},
		{
			// One mixin shows its fields.
			args: []string{"sites", "mixins", "get", "AcmeStore_NL", "customerSettings"},
			path: "/site/acme1/sites/AcmeStore_NL/mixins/customerSettings",
			mode: getWhole,
			body: `{"regExSettings":{"phone":"^[0-9]{7,}$"},"metadata":{"schema":"https://example.com/customerSettings.v3"}}`,
			want: "NAME           VALUE\n" +
				"metadata       {\"schema\":\"https://example.com/customerSettings.v3\"}\n" +
				"regExSettings  {\"phone\":\"^[0-9]{7,}$\"}\n",
		},
	}
}

// currencyReads returns the Currency Service reads.
func currencyReads() []readCase {
	return []readCase{
		{
			args:    []string{"currencies", "list", "--language", "nl"},
			path:    "/currency/acme1/currencies",
			filters: listFilters{kinds: sortable},
			body: `[
		{"code":"EUR","name":{"en":"Euro"}},{"code":"CHF","name":{"de":"Schweizer Franken","nl":"Zwitserse frank"}}]`,
			want: "CODE  NAME\nEUR   Euro (en)\nCHF   Zwitserse frank\n",
		},
		{
			args: []string{"currencies", "get", "CHF", "--language", "nl"},
			path: "/currency/acme1/currencies/CHF",
			mode: getWhole,
			body: `{"code":"CHF","name":{"en":"Swiss franc","nl":"Zwitserse frank"}}`,
			want: "CODE  NAME\nCHF   Zwitserse frank\n",
		},
		{
			// A rate is kept as sent: a string with a dot or a comma, a number, or nothing.
			args:    []string{"exchange-rates", "list"},
			path:    "/currency/acme1/exchanges",
			scopes:  []string{"currency.currency_read"},
			filters: listFilters{kinds: sortable},
			body: `[
		{"code":"EUR_CHF","sourceCurrency":"EUR","targetCurrency":"CHF","rate":"0.9412"},
		{"code":"EUR_GBP","sourceCurrency":"EUR","targetCurrency":"GBP","rate":0.87},
		{"code":"EUR_USD","sourceCurrency":"EUR","targetCurrency":"USD","rate":"1,0712"},
		{"code":"EUR_PLN","sourceCurrency":"EUR","targetCurrency":"PLN","rate":""},
		{"code":"EUR_SEK","sourceCurrency":"EUR","targetCurrency":"SEK","rate":null}]`,
			want: `CODE     SOURCE  TARGET  RATE
EUR_CHF  EUR     CHF     0.9412
EUR_GBP  EUR     GBP     0.87
EUR_USD  EUR     USD     1,0712
EUR_PLN  EUR     PLN     
EUR_SEK  EUR     SEK     
`,
		},
		{
			args:   []string{"exchange-rates", "get", "EUR_CHF"},
			path:   "/currency/acme1/exchanges/EUR_CHF",
			scopes: []string{"currency.currency_read"},
			mode:   getWhole,
			body:   `{"code":"EUR_CHF","sourceCurrency":"EUR","targetCurrency":"CHF","rate":"0.9412"}`,
			want:   "CODE     SOURCE  TARGET  RATE\nEUR_CHF  EUR     CHF     0.9412\n",
		},
	}
}

// countryReads returns the Country Service reads.
func countryReads() []readCase {
	return []readCase{
		{
			// Sorted by code, so the output is the same on every run.
			args:    []string{"regions", "list", "--language", "de"},
			path:    "/country/acme1/regions",
			query:   "sort=code",
			filters: listFilters{kinds: sortable},
			body:    `[{"code":"Europe","name":{"de":"Europa","en":"Europe"}}]`,
			want:    "CODE    NAME\nEurope  Europa\n",
		},
		{
			// The Country Service orders countries differently on every request: one page of 1000,
			// sorted by code.
			args:    []string{"countries", "list", "--language", "de"},
			path:    "/country/acme1/countries",
			query:   "sort=code",
			mode:    getAllInOnePage,
			page:    "1000",
			filters: listFilters{kinds: sortable},
			body: `[
		{"code":"NL","name":{"de":"Niederlande","en":"Netherlands"},"active":true,"regions":["Europe"]},
		{"code":"AQ","name":{"en":"Antarctica"},"active":false,"regions":[]}]`,
			want: "CODE  NAME             ACTIVE  REGIONS\n" +
				"NL    Niederlande      true    Europe\n" +
				"AQ    Antarctica (en)  false   \n",
		},
		{
			args: []string{"countries", "get", "NL", "--language", "nl"},
			path: "/country/acme1/countries/NL",
			mode: getWhole,
			body: `{"code":"NL","name":{"en":"Netherlands","nl":"Nederland"},"active":true,"regions":["Europe"]}`,
			want: "CODE  NAME       ACTIVE  REGIONS\nNL    Nederland  true    Europe\n",
		},
		{
			args: []string{"regions", "get", "Europe", "--language", "nl"},
			path: "/country/acme1/regions/Europe",
			mode: getWhole,
			body: `{"code":"Europe","name":{"en":"Europe","nl":"Europa"}}`,
			want: "CODE    NAME\nEurope  Europa\n",
		},
	}
}

// taxAndConfigurationReads returns the Tax and Configuration Service reads.
func taxAndConfigurationReads() []readCase {
	return []readCase{
		{
			// Rates keep their text (21.0, not 21); the default class is marked.
			args:    []string{"taxes", "list"},
			path:    "/tax/acme1/taxes",
			scopes:  []string{"tax.tax_read"},
			filters: listFilters{kinds: sortable},
			body: `[
		{"locationCode":"NL","location":{"countryCode":"NL"},"taxClasses":[
			{"code":"STANDARD","name":{"en":"Standard"},"rate":21.0,"isDefault":true},
			{"code":"REDUCED","name":{"en":"Reduced"},"rate":9.0},{"code":"ZERO","rate":0}]},
		{"locationCode":"EMPTY","location":{"countryCode":"XX"}},
		{"location":{"countryCode":"BE"},"taxClasses":[{"code":"STANDARD","rate":21,"isDefault":true}]}]`,
			want: "LOCATION  COUNTRY  CLASSES\n" +
				"NL        NL       STANDARD 21.0 (default), REDUCED 9.0, ZERO 0\n" +
				"EMPTY     XX       \n" +
				"BE        BE       STANDARD 21 (default)\n",
		},
		{
			args:   []string{"taxes", "get", "NL"},
			path:   "/tax/acme1/taxes/NL",
			scopes: []string{"tax.tax_read"},
			mode:   getWhole,
			body: `{"locationCode":"NL","location":{"countryCode":"NL"},"taxClasses":[{"code":"STANDARD","rate":21.0,
				"isDefault":true}]}`,
			want: "LOCATION  COUNTRY  CLASSES\nNL        NL       STANDARD 21.0 (default)\n",
		},
		{
			// The service ignores paging, so the settings are read with one request.
			args:   []string{"configurations", "list"},
			path:   "/configuration/acme1/configurations",
			scopes: []string{"configuration.configuration_view"},
			mode:   getWhole,
			body:   `[{"key":"project_lang","value":"[{\"id\":\"nl\"}]","version":2,"secured":false,"readOnly":false}]`,
			want: "KEY           VALUE          VERSION  SECURED  READ ONLY\n" +
				"project_lang  [{\"id\":\"nl\"}]  2        false    false\n",
		},
		{
			args:   []string{"configurations", "get", "project_lang"},
			path:   "/configuration/acme1/configurations/project_lang",
			scopes: []string{"configuration.configuration_view"},
			mode:   getWhole,
			body:   `{"key":"project_lang","value":"[{\"id\":\"nl\"}]","version":2,"secured":false,"readOnly":true}`,
			want: "KEY           VALUE          VERSION  SECURED  READ ONLY\n" +
				"project_lang  [{\"id\":\"nl\"}]  2        false    true\n",
		},
		{
			// Global settings need no scope; a secured one shows its value as sent.
			args: []string{"configurations", "global"},
			path: "/configuration/acme1/global-configurations",
			mode: getWhole,
			body: `[{"key":"ext-version","value":"e851c72","version":0,"secured":false,"readOnly":false,
				"securedValue":null},{"key":"sealed","value":"S1","secured":true,"securedValue":"S2"}]`,
			want: "KEY          VALUE    VERSION  SECURED  READ ONLY\n" +
				"ext-version  e851c72  0        false    false\n" +
				"sealed       S1                true     \n",
		},
		{
			args:   []string{"configurations", "clients"},
			path:   "/configuration/acme1/clients",
			scopes: []string{"configuration.configuration_view"},
			mode:   getWhole,
			body:   `["storefront","acme.indexing-client"]`,
			want:   "CLIENT\nstorefront\nacme.indexing-client\n",
		},
		{
			args:   []string{"configurations", "for-client", "acme.indexing-client"},
			path:   "/configuration/acme1/clients/acme.indexing-client/configurations",
			scopes: []string{"configuration.configuration_view"},
			mode:   getWhole,
			body:   `[{"_id":"x1","client":"acme.indexing-client","key":"batchSize","value":100,"version":1}]`,
			want:   "KEY        VALUE  VERSION  SECURED  READ ONLY\nbatchSize  100    1                 \n",
		},
		{
			// The client, then the key, address one setting.
			args:   []string{"configurations", "client-setting", "acme.indexing-client", "batchSize"},
			path:   "/configuration/acme1/clients/acme.indexing-client/configurations/batchSize",
			scopes: []string{"configuration.configuration_view"},
			mode:   getWhole,
			body: `{"_id":"x1","client":"acme.indexing-client","key":"batchSize","value":{"size":100},"version":2,
				"secured":false,"readOnly":true}`,
			want: "KEY        VALUE         VERSION  SECURED  READ ONLY\nbatchSize  {\"size\":100}  2        false    true\n",
		},
	}
}

// indexingReads returns the Indexing Service reads.
func indexingReads() []readCase {
	return []readCase{
		{
			// Providers come whole, without their keys (see TestIndexingProvidersHideKeys).
			args:   []string{"indexing", "providers", "list"},
			path:   "/indexing/acme1/configurations",
			scopes: []string{"indexing.search_view"},
			mode:   getWhole,
			body:   providersBody,
			want:   "PROVIDER  ACTIVE  APPLICATION ID  INDEX\nALGOLIA   false   APP1            acme1\n",
		},
		{
			args:   []string{"indexing", "providers", "get", "ALGOLIA"},
			path:   "/indexing/acme1/configurations/ALGOLIA",
			scopes: []string{"indexing.search_view"},
			mode:   getWhole,
			body:   `{"provider":"ALGOLIA","active":true,"applicationId":"APP1","indexName":"acme1"}`,
			want:   "PROVIDER  ACTIVE  APPLICATION ID  INDEX\nALGOLIA   true    APP1            acme1\n",
		},
		{
			// A storefront's view needs no scope.
			args: []string{"indexing", "providers", "public", "list"},
			path: "/indexing/acme1/public/configurations",
			mode: getWhole,
			body: `[{"provider":"ALGOLIA","applicationId":"APP1","indexName":"acme1"}]`,
			want: "PROVIDER  ACTIVE  APPLICATION ID  INDEX\nALGOLIA           APP1            acme1\n",
		},
		{
			args: []string{"indexing", "providers", "public", "get", "ALGOLIA"},
			path: "/indexing/acme1/public/configurations/ALGOLIA",
			mode: getWhole,
			body: `{"provider":"ALGOLIA","applicationId":"APP1","indexName":"acme1"}`,
			want: "PROVIDER  ACTIVE  APPLICATION ID  INDEX\nALGOLIA           APP1            acme1\n",
		},
		{
			// Emporix fails when asked for the jobs' total, so their pages are read without it.
			args:    []string{"indexing", "reindex-jobs", "list"},
			path:    "/indexing/acme1/reindex-jobs",
			scopes:  []string{"indexing.search_view"},
			mode:    getAllWithoutTotal,
			filters: listFilters{kinds: filterable},
			body: `[{"id":"j1","entityType":"PRODUCT","status":"FAILURE",
		"processedCount":3,"expectedCount":15,"message":"timeout","metadata":{"createdAt":"2026-10-13T16:00:00Z",
		"modifiedAt":"2026-10-13T16:05:00Z"}}]`,
			want: "ID  ENTITY   STATUS   PROCESSED  EXPECTED  CREATED               MODIFIED              MESSAGE\n" +
				"j1  PRODUCT  FAILURE  3          15        2026-10-13T16:00:00Z  2026-10-13T16:05:00Z  timeout\n",
		},
		{
			args:   []string{"indexing", "reindex-jobs", "get", "j1"},
			path:   "/indexing/acme1/reindex-jobs/j1",
			scopes: []string{"indexing.search_view"},
			mode:   getWhole,
			body:   `{"id":"j1","entityType":"PRODUCT","status":"SUCCESS","processedCount":15,"expectedCount":15}`,
			want: "ID  ENTITY   STATUS   PROCESSED  EXPECTED  CREATED  MODIFIED  MESSAGE\n" +
				"j1  PRODUCT  SUCCESS  15         15                           \n",
		},
	}
}

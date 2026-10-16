# epx sites

> Read, create and change sites and their mixins.

## epx sites

Reads active sites. Inactive sites are not included: Emporix returns them only with the
write-capable site.site_manage scope, so `get CODE` of an inactive site is refused (403,
exit 3) without it. [`epx sites codes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-codes) lists the codes of every site, inactive ones too (opt-in);
to read inactive sites whole, run
[`epx api get '/site/{tenant}/sites?includeInactive=true' --scope site.site_manage`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the site.site_manage scope.

**Command:** `epx sites [flags]`, `epx sites [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx sites codes`](#epx-sites-codes)\
  List the codes of every site, inactive ones too.
- [`epx sites create`](#epx-sites-create)\
  Create a site.
- [`epx sites delete`](#epx-sites-delete)\
  Delete a site.
- [`epx sites get`](#epx-sites-get)\
  Show one site.
- [`epx sites list`](#epx-sites-list)\
  List sites.
- [`epx sites mixins`](#epx-sites-mixins)\
  Read and change the mixins of a site.
- [`epx sites patch`](#epx-sites-patch)\
  Change members of a site.
- [`epx sites update`](#epx-sites-update)\
  Replace a site.

## epx sites codes

List the codes of every site, inactive ones too.

Reads the code of every site, active or not. Emporix lists inactive sites only with the
write-capable site.site_manage and site.site_read scopes together (each alone
answers 403), so this read is opt-in; [`epx sites list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-list) shows active sites whole without a scope.
Emporix sends no total for the codes, so their pages are read until an empty one.

Requires the site.site_read, site.site_manage scopes.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx sites codes [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting site.site_read, site.site_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites create

Create a site.

Creates a site from the JSON in `--file`, shaped as [`epx sites get CODE -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-get) prints one.
The reference requires code, name, defaultLanguage, languages, currency, homeBase (an
address with country and zipCode), shipToCountries and taxDeterminationBasedOn
(BILLING_ADDRESS or SHIPPING_ADDRESS); active, default, includesTax,
availableCurrencies, payment, shipping, tax, mixins and cartCalculationScale are optional.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites delete

Delete a site.

Deletes the site CODE; it sends no body. Emporix refuses to delete the default site.
An unknown CODE is not found (exit 4).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CODE at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites delete CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites get

Show one site.

Reads active sites. Inactive sites are not included: Emporix returns them only with the
write-capable site.site_manage scope, so `get CODE` of an inactive site is refused (403,
exit 3) without it. [`epx sites codes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-codes) lists the codes of every site, inactive ones too (opt-in);
to read inactive sites whole, run
[`epx api get '/site/{tenant}/sites?includeInactive=true' --scope site.site_manage`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).

No OAuth scope is needed.

**Command:** `epx sites get CODE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, DEFAULT, LANGUAGES, CURRENCY, HOME, SHIPS TO.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites list

List sites.

Reads active sites. Inactive sites are not included: Emporix returns them only with the
write-capable site.site_manage scope, so `get CODE` of an inactive site is refused (403,
exit 3) without it. [`epx sites codes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-codes) lists the codes of every site, inactive ones too (opt-in);
to read inactive sites whole, run
[`epx api get '/site/{tenant}/sites?includeInactive=true' --scope site.site_manage`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).

No OAuth scope is needed.

**Command:** `epx sites list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, DEFAULT, LANGUAGES, CURRENCY, HOME, SHIPS TO.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins

Read and change the mixins of a site.

Reads the mixins of a site, the custom settings that extend it, and creates, changes and
deletes them. SITE_CODE is the CODE that [`epx sites list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-list) shows.

Its reads need no OAuth scope.
Its writes require the site.site_manage scope.

**Command:** `epx sites mixins [flags]`, `epx sites mixins [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx sites mixins create`](#epx-sites-mixins-create)\
  Add mixins to a site.
- [`epx sites mixins delete`](#epx-sites-mixins-delete)\
  Delete a mixin of a site.
- [`epx sites mixins get`](#epx-sites-mixins-get)\
  Show one mixin of a site.
- [`epx sites mixins list`](#epx-sites-mixins-list)\
  List the mixins of a site.
- [`epx sites mixins patch`](#epx-sites-mixins-patch)\
  Change fields of a mixin of a site.
- [`epx sites mixins update`](#epx-sites-mixins-update)\
  Replace a mixin of a site.

## epx sites mixins create

Add mixins to a site.

Adds mixins to the site SITE_CODE: the JSON object in `--file` maps the name of each to its
fields, e.g. `{"generalShopSettings": {"logoUrl": "https://shop.acme.example/logo.png"}}`.
Emporix checks each against the schema that its metadata.schema names; a mixin that the
site has already is a conflict (exit 5).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites mixins create SITE_CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins delete

Delete a mixin of a site.

Deletes the mixin NAME of the site SITE_CODE; it sends no body.
An unknown site or mixin is not found (exit 4).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type NAME at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites mixins delete SITE_CODE NAME [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins get

Show one mixin of a site.

Reads one mixin of a site: its fields by NAME, each VALUE cut (`-o json` prints them whole).
SITE_CODE is the CODE that [`epx sites list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-list) shows; an unknown one is not found.
NAME is the NAME that [`epx sites mixins list SITE_CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-list) shows; an unknown one is not found.

No OAuth scope is needed.

**Command:** `epx sites mixins get SITE_CODE NAME [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among NAME, VALUE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins list

List the mixins of a site.

Reads the mixins of a site, the custom settings that extend it (such as coupon or customer
settings), by NAME, each VALUE cut (`-o json` prints them whole). A site without mixins lists none.
SITE_CODE is the CODE that [`epx sites list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-list) shows; an unknown one is not found.

No OAuth scope is needed.

**Command:** `epx sites mixins list SITE_CODE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among NAME, VALUE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins patch

Change fields of a mixin of a site.

Changes the fields of the mixin NAME of the site SITE_CODE that the JSON object in
`--file` gives, such as `{"active": true}`; the others keep their values.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites mixins patch SITE_CODE NAME [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites mixins update

Replace a mixin of a site.

Replaces the mixin NAME of the site SITE_CODE with the JSON object in `--file`; the
reference's example gives it as create does, in an object named after the mixin.
An unknown site or mixin is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites mixins update SITE_CODE NAME [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites patch

Change members of a site.

Changes the members of the site CODE that the JSON in `--file` gives, such as
`{"languages": ["en", "nl"]}`; the others keep their values.
An unknown CODE is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites patch CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sites update

Replace a site.

Replaces the site CODE with the JSON in `--file`, shaped as create takes it: a member that it
leaves out is removed. It changes the payment, shipping and tax configurations that the body
gives too; Emporix changes an inactive one only with `--expand` naming it, e.g. payment:all.
An unknown CODE is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the site.site_manage scope.

**Command:** `epx sites update CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--expand <SETTINGS>`\
  The SETTINGS to expand, e.g. payment:all,shipping:all,tax:all (all, active or none each).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

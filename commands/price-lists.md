# epx price-lists

> Read and change price lists and their prices.

## epx price-lists

Reads and changes price lists and, in `prices`, the prices of one. A price list holds
prices in one currency for one site that apply between their validity dates to its
countries or regions, and to its customer groups or to one legal entity.

Its reads require the price.pricelist_read scope.
Its writes require the price.pricelist_manage scope.

**Command:** `epx price-lists [flags]`, `epx price-lists [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx price-lists create`](#epx-price-lists-create)\
  Create a price list.
- [`epx price-lists delete`](#epx-price-lists-delete)\
  Delete a price list.
- [`epx price-lists get`](#epx-price-lists-get)\
  Show one price list.
- [`epx price-lists list`](#epx-price-lists-list)\
  List price lists.
- [`epx price-lists prices`](#epx-price-lists-prices)\
  Read and change the prices of a price list.
- [`epx price-lists search`](#epx-price-lists-search)\
  Search price lists.
- [`epx price-lists update`](#epx-price-lists-update)\
  Upsert a price list.

## epx price-lists create

Create a price list.

Creates the price list in `--file`, a JSON object with name, currency, siteCode and validity
(from and to), and optionally countries or regions, and customerGroups or legalEntityId,
e.g. `{"name": {"en": "Wholesale"}, "currency": "EUR", "siteCode": "AcmeStore_NL", "countries": ["NL"], "validity": {"from": "2026-01-01T00:00:00.000Z"}}`. Without an id, Emporix
makes one, which the response gives.

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "Wholesale", "nl": "Groothandel"}`, as `-o json` of a read prints it, also with `--content-language`
LANGUAGE: then a map of that language only, e.g. `{"en": "Wholesale"}` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists create [flags]`

**Flags**

- `--content-language <LANGUAGE>`\
  The LANGUAGE of localized text in the body, or \* for maps of translations. Default: `*`.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists delete

Delete a price list.

Deletes the price list ID and the prices in it, which Emporix deletes after it, in the
background.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Deletes the price list and every price in it with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists get

Show one price list.

Reads price lists. A price list holds prices in one currency for one site that apply
between FROM and TO to its countries or regions, and to its customer groups (a preview
feature of Emporix) or to one legal entity and its subsidiaries. Empty COUNTRIES and
REGIONS apply everywhere; empty CUSTOMER GROUPS and LEGAL ENTITY apply to every customer.

Requires the price.pricelist_read scope.

**Command:** `epx price-lists get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CURRENCY, SITE, COUNTRIES, REGIONS, CUSTOMER GROUPS, LEGAL ENTITY, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists list

List price lists.

Reads price lists. A price list holds prices in one currency for one site that apply
between FROM and TO to its countries or regions, and to its customer groups (a preview
feature of Emporix) or to one legal entity and its subsidiaries. Empty COUNTRIES and
REGIONS apply everywhere; empty CUSTOMER GROUPS and LEGAL ENTITY apply to every customer.

Requires the price.pricelist_read scope.

**Command:** `epx price-lists list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CURRENCY, SITE, COUNTRIES, REGIONS, CUSTOMER GROUPS, LEGAL ENTITY, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
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

## epx price-lists prices

Read and change the prices of a price list.

Reads and changes the prices of a price list, in its currency. PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.

Its reads require the price.pricelist_read scope.
Its writes require the price.pricelist_manage scope.

**Command:** `epx price-lists prices [flags]`, `epx price-lists prices [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx price-lists prices bulk-create`](#epx-price-lists-prices-bulk-create)\
  Create prices of a price list in bulk.
- [`epx price-lists prices bulk-delete`](#epx-price-lists-prices-bulk-delete)\
  Delete prices of a price list in bulk.
- [`epx price-lists prices bulk-update`](#epx-price-lists-prices-bulk-update)\
  Upsert prices of a price list in bulk.
- [`epx price-lists prices create`](#epx-price-lists-prices-create)\
  Create a price of a price list.
- [`epx price-lists prices delete`](#epx-price-lists-prices-delete)\
  Delete a price of a price list.
- [`epx price-lists prices get`](#epx-price-lists-prices-get)\
  Show one price of a price list.
- [`epx price-lists prices list`](#epx-price-lists-prices-list)\
  List the prices of a price list.
- [`epx price-lists prices search`](#epx-price-lists-prices-search)\
  Search the prices of a price list.
- [`epx price-lists prices update`](#epx-price-lists-prices-update)\
  Upsert a price of a price list.

## epx price-lists prices bulk-create

Create prices of a price list in bulk.

Adds the prices in `--file` to the price list PRICE_LIST_ID: a JSON array of prices, each
as create takes it.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices bulk-create PRICE_LIST_ID [flags]`

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

## epx price-lists prices bulk-delete

Delete prices of a price list in bulk.

Deletes the prices of the price list PRICE_LIST_ID whose ids `--file` gives, a JSON array of
ids, e.g. `["p1", "p2"]`.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.

Emporix takes up to 200 items in one request (epx refuses more) and answers for
all of them at once, not item by item.

Deletes the prices of a price list that `--file` names with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRICE_LIST_ID at its prompt.
`--yes` sends without asking (required without a terminal, or with `--file` -);
`--dry-run` prints the request instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices bulk-delete PRICE_LIST_ID [flags]`

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

## epx price-lists prices bulk-update

Upsert prices of a price list in bulk.

Replaces the prices in `--file` in the price list PRICE_LIST_ID, a JSON array of prices,
each with its id, or creates those that do not exist (an upsert).
An item that failed with 500 may be one whose metadata.version is stale: Emporix
reports it so, not as 409 Conflict. Read it again before you retry it.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices bulk-update PRICE_LIST_ID [flags]`

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

## epx price-lists prices create

Create a price of a price list.

Adds the price in `--file` to the price list PRICE_LIST_ID, in its currency: a JSON object
with itemId (itemType and id) and
tierValues: the id of each tier of the price model that priceModelId names, as
[`epx price-models get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-models#epx-price-models-get) prints it, with a priceValue, e.g.
`{"itemId": {"itemType": "PRODUCT", "id": "100200"}, "priceModelId": "pm1", "tierValues": [{"id": "t1", "priceValue": 15.90}]}`.
Without an id, Emporix makes one, which the response gives.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices create PRICE_LIST_ID [flags]`

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

## epx price-lists prices delete

Delete a price of a price list.

Deletes the price ID of the price list PRICE_LIST_ID.
Emporix answers an unknown ID as one that it deleted (204 No Content).
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.
ID is the ID that [`epx price-lists prices list PRICE_LIST_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-prices-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices delete PRICE_LIST_ID ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists prices get

Show one price of a price list.

Reads one price of a price list, in its currency.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows; an unknown one is not found.
ID is the ID that [`epx price-lists prices list PRICE_LIST_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-prices-list) shows; an unknown one is not found.
TIERS lists each tier value as sent; `-o json` pairs it with the id of its tier in the price
model MODEL ([`epx price-models get MODEL`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-models#epx-price-models-get); an empty MODEL is the default model).

Requires the price.pricelist_read scope.

**Command:** `epx price-lists prices get PRICE_LIST_ID ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, MODEL, TIERS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists prices list

List the prices of a price list.

Reads the prices of a price list, in its currency.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows; an unknown one lists nothing.
TIERS lists each tier value as sent; `-o json` pairs it with the id of its tier in the price
model MODEL ([`epx price-models get MODEL`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-models#epx-price-models-get); an empty MODEL is the default model).

Requires the price.pricelist_read scope.

**Command:** `epx price-lists prices list PRICE_LIST_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, MODEL, TIERS.
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

## epx price-lists prices search

Search the prices of a price list.

Reads the prices of a price list, in its currency, that an Emporix q filter finds
(`--query`; without it, every one).
It reads and shows them as [`epx price-lists prices list PRICE_LIST_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-prices-list) does.
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows; an unknown one lists nothing.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
TIERS lists each tier value as sent; `-o json` pairs it with the id of its tier in the price
model MODEL ([`epx price-models get MODEL`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-models#epx-price-models-get); an empty MODEL is the default model).

Requires the price.pricelist_read scope.

**Command:** `epx price-lists prices search PRICE_LIST_ID [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, MODEL, TIERS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists prices update

Upsert a price of a price list.

Replaces the price ID of the price list PRICE_LIST_ID with the one in `--file`, a JSON object
as [`epx price-lists prices get PRICE_LIST_ID ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-prices-get) prints it, or creates it with
that id where none exists (an upsert).
PRICE_LIST_ID is the ID that [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) shows.
ID is the ID that [`epx price-lists prices list PRICE_LIST_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-prices-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists prices update PRICE_LIST_ID ID [flags]`

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

## epx price-lists search

Search price lists.

Reads the price lists that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx price-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the price.pricelist_read scope.

**Command:** `epx price-lists search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CURRENCY, SITE, COUNTRIES, REGIONS, CUSTOMER GROUPS, LEGAL ENTITY, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx price-lists update

Upsert a price list.

Replaces the price list ID with the one in `--file`, a JSON object as
[`epx price-lists get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists#epx-price-lists-get) prints it, or creates it with that id where none exists
(an upsert).

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "Wholesale", "nl": "Groothandel"}`, as `-o json` of a read prints it, also with `--content-language`
LANGUAGE: then a map of that language only, e.g. `{"en": "Wholesale"}` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.pricelist_manage scope.

**Command:** `epx price-lists update ID [flags]`

**Flags**

- `--content-language <LANGUAGE>`\
  The LANGUAGE of localized text in the body, or \* for maps of translations. Default: `*`.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

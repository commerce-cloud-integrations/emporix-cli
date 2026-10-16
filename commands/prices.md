# epx prices

> Read and change prices.

## epx prices

Reads and changes prices. Without a site, Emporix lists the prices of every site, so list
and for-product require `--site` and read that site's prices only; an unknown site lists
nothing. create and update change one price, bulk-create and bulk-update many, each from
`--file`, and delete deletes one.

Its reads require different OAuth scopes; the help of each names them.
Its writes require the price.price\_manage scope.

**Command:** `epx prices [flags]`, `epx prices [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx prices bulk-create`](#epx-prices-bulk-create)\
  Create prices in bulk.
- [`epx prices bulk-update`](#epx-prices-bulk-update)\
  Upsert prices in bulk.
- [`epx prices create`](#epx-prices-create)\
  Create a price.
- [`epx prices delete`](#epx-prices-delete)\
  Delete a price.
- [`epx prices for-product`](#epx-prices-for-product)\
  List the prices of a product on a site.
- [`epx prices get`](#epx-prices-get)\
  Show one price.
- [`epx prices list`](#epx-prices-list)\
  List the prices of a site.
- [`epx prices match`](#epx-prices-match)\
  Match the best price of products.
- [`epx prices search`](#epx-prices-search)\
  Search the prices of products.
- [`epx prices update`](#epx-prices-update)\
  Upsert a price.

## epx prices bulk-create

Create prices in bulk.

Creates the prices in `--file`, a JSON array of prices, each as create takes it.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.price\_manage scope.

**Command:** `epx prices bulk-create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices bulk-update

Upsert prices in bulk.

Replaces the prices in `--file`, a JSON array of prices, each with its id, or creates
those that do not exist (an upsert).
An item that failed with 500 may be one whose metadata.version is stale: Emporix
reports it so, not as 409 Conflict. Read it again before you retry it.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.price\_manage scope.

**Command:** `epx prices bulk-update [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices create

Create a price.

Creates the price in `--file`, a JSON object with at least itemId, currency, location and
tierValues: the id of each tier of the price model that priceModelId names, as
[`epx price-models get ID -o json`](/emporix-cli/commands/price-models#epx-price-models-get) prints it, with a priceValue. Without an id, Emporix
makes one, which the response gives. Emporix does not refuse an id that a price has
already: it stores a second price with it, and each delete of that id removes one.
Leave id out, or change a price with update.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.price\_manage scope.

**Command:** `epx prices create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices delete

Delete a price.

Deletes the price PRICE\_ID.
Emporix answers an unknown PRICE\_ID as one that it deleted (204 No Content).
PRICE\_ID is the ID that [`epx prices list --site S`](/emporix-cli/commands/prices#epx-prices-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRICE\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the price.price\_manage scope.

**Command:** `epx prices delete PRICE_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices for-product

List the prices of a product on a site.

Reads prices. Without a site, Emporix lists the prices of every site, so list and
for-product require `--site` and read that site's prices only; an unknown site lists
nothing. TIERS lists the value of each price tier as sent; tier quantities and net or
gross come from the price model ([`epx price-models get`](/emporix-cli/commands/price-models#epx-price-models-get) with the priceModelId of `-o json`).
Customer and price-list restrictions and sale prices are only in `-o json`.
PRODUCT\_ID is the product id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.
An unknown id lists nothing.

No OAuth scope is needed.

**Command:** `epx prices for-product PRODUCT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, CURRENCY, COUNTRY, SITES, TIERS, FROM, TO.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--site <site>`\
  Site code, e.g. main (required).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices get

Show one price.

Reads prices. Without a site, Emporix lists the prices of every site, so list and
for-product require `--site` and read that site's prices only; an unknown site lists
nothing. TIERS lists the value of each price tier as sent; tier quantities and net or
gross come from the price model ([`epx price-models get`](/emporix-cli/commands/price-models#epx-price-models-get) with the priceModelId of `-o json`).
Customer and price-list restrictions and sale prices are only in `-o json`.

No OAuth scope is needed.

**Command:** `epx prices get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, CURRENCY, COUNTRY, SITES, TIERS, FROM, TO.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices list

List the prices of a site.

Reads prices. Without a site, Emporix lists the prices of every site, so list and
for-product require `--site` and read that site's prices only; an unknown site lists
nothing. TIERS lists the value of each price tier as sent; tier quantities and net or
gross come from the price model ([`epx price-models get`](/emporix-cli/commands/price-models#epx-price-models-get) with the priceModelId of `-o json`).
Customer and price-list restrictions and sale prices are only in `-o json`.

No OAuth scope is needed.

**Command:** `epx prices list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, CURRENCY, COUNTRY, SITES, TIERS, FROM, TO.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--site <site>`\
  Site code, e.g. main (required).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices match

Match the best price of products.

Matches the price of one unit of each product that `--product` names (the product id as in
[`epx products list`](/emporix-cli/commands/products#epx-products-list)) that Emporix would charge on the site `--site` names, for the country
`--country` names, in the currency `--currency` names: of the prices that apply, the lowest
after discounts, converted with the tenant's exchange rates and tax classes where the price
is in another currency or for another country. ID is the matched price (of the price list
PRICE LIST, if any), ORIGINAL and EFFECTIVE (discounted) the price of one unit, TOTAL that
of QUANTITY, gross where TAX is true. Other quantities and units, a customer or a legal
entity go in `--body`.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the price.price\_read scope.

**Command:** `epx prices match [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, PRICE LIST, SITE, COUNTRY, QUANTITY, ORIGINAL, EFFECTIVE, TOTAL, TAX.
- `--country <code>`\
  The code of the country to match the price for, e.g. NL (required without `--body`).
- `--currency <code>`\
  The code of the currency to match the price in, e.g. EUR (required without `--body`).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--product <ID>`\
  A product ID whose price to match (repeatable; required without `--body`).
- `--site <code>`\
  The code of the site whose price to match (required without `--body`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices search

Search the prices of products.

Reads the prices of up to 100 products (`--product`, the product id as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), on the site and in the currency that `--site` and `--currency` name, else on every
site and in every currency, shown as list shows prices; `-o json` also lists the products
without a price. Other criteria (a customer, a date, net or gross values) go in `--body`.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the price.price\_read scope.

**Command:** `epx prices search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ITEM, TYPE, CURRENCY, COUNTRY, SITES, TIERS, FROM, TO.
- `--currency <code>`\
  The code of the currency whose prices to read, e.g. EUR.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--product <ID>`\
  A product ID whose prices to read (repeatable, up to 100; required without `--body`).
- `--site <code>`\
  The code of the site whose prices to read.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx prices update

Upsert a price.

Replaces the price PRICE\_ID with the one in `--file`, a JSON object as
[`epx prices get PRICE_ID -o json`](/emporix-cli/commands/prices#epx-prices-get) prints it, or creates it with that id where none exists
(an upsert; not with `--partial`: Emporix answers a missing price with 404 Not Found, exit
4). Send the price's metadata.version as read to change it only if nobody changed it
since: Emporix answers 409 Conflict otherwise.
PRICE\_ID is the ID that [`epx prices list --site S`](/emporix-cli/commands/prices#epx-prices-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the price.price\_manage scope.

**Command:** `epx prices update PRICE_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--partial`\
  Change only the fields that the body gives; without it, the body replaces the whole price.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

# epx availability

> Read and change availability.

## epx availability

Reads product availability (stock), the tenant's own availability: on a site with list and
get, which require `--site`, and on every site with search. create, update and delete change
the availability of a product on a site, bulk-create, bulk-update and bulk-delete that of
many.

Its reads need no OAuth scope.
Its writes require the availability.availability_manage scope.

**Command:** `epx availability [flags]`, `epx availability [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx availability bulk-create`](#epx-availability-bulk-create)\
  Create availability in bulk.
- [`epx availability bulk-delete`](#epx-availability-bulk-delete)\
  Delete availability in bulk.
- [`epx availability bulk-update`](#epx-availability-bulk-update)\
  Upsert availability in bulk.
- [`epx availability create`](#epx-availability-create)\
  Create the availability of a product on a site.
- [`epx availability delete`](#epx-availability-delete)\
  Delete the availability of a product on a site.
- [`epx availability get`](#epx-availability-get)\
  Show the availability of a product on a site.
- [`epx availability list`](#epx-availability-list)\
  List the availability of every product on a site.
- [`epx availability search`](#epx-availability-search)\
  Search the availability of products.
- [`epx availability update`](#epx-availability-update)\
  Upsert the availability of a product on a site.

## epx availability bulk-create

Create availability in bulk.

Creates the availability records in `--file`, a JSON array of records, each with the
productId and site it is for and the fields that create takes. Emporix advises against a
record on a site named main, which aggregates the availability of the other sites.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability bulk-create [flags]`

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

## epx availability bulk-delete

Delete availability in bulk.

Deletes the availability records that `--file` names, a JSON array of objects with the
productId and site of each, such as `[{"productId": "100200", "site": "AcmeStore_NL"}]`.
The availability of a bundle is that of its scarcest product and cannot be deleted.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Deletes the availability that `--file` names with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type the tenant at its prompt.
`--yes` sends without asking (required without a terminal, or with `--file` -);
`--dry-run` prints the request instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability bulk-delete [flags]`

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

## epx availability bulk-update

Upsert availability in bulk.

Replaces the availability records in `--file`, a JSON array of records as bulk-create
takes them, or creates those that do not exist (an upsert).

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability bulk-update [flags]`

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

## epx availability create

Create the availability of a product on a site.

Creates the availability of the product PRODUCT_ID on the site `--site` names, from `--file`:
a JSON object with stockLevel, available, popularity and distributionChannel (ASSORTMENT,
HOME_DELIVERY or PICKUP), as [`epx availability get PRODUCT_ID --site S -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/availability#epx-availability-get) prints
them. Emporix answers 409 Conflict when the product has one on that site already. It
advises against a record on a site named main, which aggregates the availability of the
other sites.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability create PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx availability delete

Delete the availability of a product on a site.

Deletes the availability of the product PRODUCT_ID on the site `--site` names.
The availability of a bundle is that of its scarcest product and cannot be deleted.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRODUCT_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability delete PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx availability get

Show the availability of a product on a site.

Reads product availability (stock) on the site `--site` names, the tenant's own
availability.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.
The availability of a bundle's products is only in `-o json` (bundleAvailabilities).

No OAuth scope is needed.

**Command:** `epx availability get PRODUCT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PRODUCT, SITE, STOCK, AVAILABLE, CHANNEL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx availability list

List the availability of every product on a site.

Reads product availability (stock) on the site `--site` names, the tenant's own
availability.
A bundle is listed as a whole and per product; an unknown site lists nothing.

No OAuth scope is needed.

**Command:** `epx availability list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PRODUCT, SITE, STOCK, AVAILABLE, CHANNEL.
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

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx availability search

Search the availability of products.

Reads the availability records, on every site, that an Emporix q filter finds (`--query`;
without it, every record), shown as list shows them, bundles left out. `--body` with a JSON
array of product ids reads theirs instead, bundles included.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

No OAuth scope is needed.

**Command:** `epx availability search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PRODUCT, SITE, STOCK, AVAILABLE, CHANNEL.
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

## epx availability update

Upsert the availability of a product on a site.

Replaces the availability of the product PRODUCT_ID on the site `--site` names with the one
in `--file`, a JSON object as create takes it, or creates it where there is none (an
upsert). The availability of a bundle is that of its scarcest product and cannot be set.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the availability.availability_manage scope.

**Command:** `epx availability update PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

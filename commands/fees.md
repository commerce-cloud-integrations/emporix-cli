# epx fees

> Read and change fees and which items have them.

## epx fees

Reads and changes fees, surcharges on items such as products or payment types, and in
`assignments` which items have which fees. Emporix filters fees by site and documents main as
the default site, so list, for-product and for-item require `--site`. create, update and delete
change one fee.

Its reads need no OAuth scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx fees [flags]`, `epx fees [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx fees assignments`](#epx-fees-assignments)\
  Read and change which items have which fees.
- [`epx fees create`](#epx-fees-create)\
  Create a fee.
- [`epx fees delete`](#epx-fees-delete)\
  Delete a fee.
- [`epx fees for-item`](#epx-fees-for-item)\
  List the fees of an item on a site.
- [`epx fees for-product`](#epx-fees-for-product)\
  List the fees of a product on a site.
- [`epx fees get`](#epx-fees-get)\
  Show one fee.
- [`epx fees list`](#epx-fees-list)\
  List the fees of a site.
- [`epx fees update`](#epx-fees-update)\
  Replace or change a fee.

## epx fees assignments

Read and change which items have which fees.

Reads item fees: which fees apply to which item (a product or a payment type) on a site.
Its writes assign fees to an item, by its YRN, or to a product, by its id.

Its reads need no OAuth scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx fees assignments [flags]`, `epx fees assignments [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx fees assignments create`](#epx-fees-assignments-create)\
  Assign fees to an item on a site.
- [`epx fees assignments delete`](#epx-fees-assignments-delete)\
  Delete the item fee of an item.
- [`epx fees assignments delete-fee`](#epx-fees-assignments-delete-fee)\
  Remove a fee from an item.
- [`epx fees assignments delete-product`](#epx-fees-assignments-delete-product)\
  Delete the item fee of a product.
- [`epx fees assignments delete-product-fee`](#epx-fees-assignments-delete-product-fee)\
  Remove a fee from a product.
- [`epx fees assignments list`](#epx-fees-assignments-list)\
  List the item fees of a site.
- [`epx fees assignments search`](#epx-fees-assignments-search)\
  Search the item fees of items.
- [`epx fees assignments search-by-product`](#epx-fees-assignments-search-by-product)\
  Search the item fees of a product on sites.
- [`epx fees assignments search-by-products`](#epx-fees-assignments-search-by-products)\
  Search the item fees of products on a site.
- [`epx fees assignments update`](#epx-fees-assignments-update)\
  Add fees to an item.
- [`epx fees assignments update-product`](#epx-fees-assignments-update-product)\
  Add fees to a product on a site.

## epx fees assignments create

Assign fees to an item on a site.

Assigns the fees that `--fee` names to the item that `--item` names on the site that `--site`
names, or those of `--file`, a JSON object with itemYrn, siteCode and feeIds, and prints
the id and YRN of the item fee. `--item` takes the YRN of an item,
e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the fee.item\_create scope.

**Command:** `epx fees assignments create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--fee <ID>`\
  The ID of a fee, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) shows it (repeatable; required without `--file`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--item <YRN>`\
  The YRN of the item to assign the fees to (required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <code>`\
  The code of the site that the fees apply on (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx fees assignments create --item 'urn:yaas:saasag:caasproduct:product:acme1;100200' \
  --site AcmeStore_NL --fee f1 --fee f2
```

## epx fees assignments delete

Delete the item fee of an item.

Deletes the item fee of the item YRN, which then has no fees.
Without `--site`, Emporix acts on the item fee on site main.
Emporix answers an unknown YRN as one that it deleted (204 No Content).
YRN names the item, e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;).

Deletes every fee of the item YRN with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type YRN at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the fee.item\_delete scope.

**Command:** `epx fees assignments delete YRN [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <CODE>`\
  The site CODE of the item fee, sent as siteCode.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments delete-fee

Remove a fee from an item.

Removes the fee FEE\_ID from the item fee of the item YRN;
Emporix deletes the item fee once it has no fees.
Emporix answers an unknown FEE\_ID as one that it deleted (204 No Content).
YRN names the item, e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;).
FEE\_ID is the ID that [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type FEE\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the fee.item\_delete scope.

**Command:** `epx fees assignments delete-fee YRN FEE_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments delete-product

Delete the item fee of a product.

Deletes the item fee of the product PRODUCT\_ID, which then has no fees.
Without `--site`, Emporix acts on the item fee on site main.
Emporix answers an unknown PRODUCT\_ID as one that it deleted (204 No Content).
PRODUCT\_ID is the product id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.

Deletes every fee of the product PRODUCT\_ID with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRODUCT\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the fee.item\_delete scope.

**Command:** `epx fees assignments delete-product PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <CODE>`\
  The site CODE of the item fee, sent as siteCode.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments delete-product-fee

Remove a fee from a product.

Removes the fee FEE\_ID from the item fee of the product PRODUCT\_ID on the site of the fee;
Emporix deletes the item fee once it has no fees.
Emporix answers an unknown FEE\_ID as one that it deleted (204 No Content).
PRODUCT\_ID is the product id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.
FEE\_ID is the ID that [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type FEE\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the fee.item\_delete scope.

**Command:** `epx fees assignments delete-product-fee PRODUCT_ID FEE_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments list

List the item fees of a site.

Reads item fees: which fees apply to which item (a product or a payment type) on a site.
Emporix documents main as the default site, so `--site` is required; an unknown site lists
nothing.
ITEM is the item's YRN, ID the item fee's id and FEES the ids of its fees, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list)
shows them; `-o json` shows the response as sent.

No OAuth scope is needed.

**Command:** `epx fees assignments list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ITEM, ID, SITE, FEES.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments search

Search the item fees of items.

Reads item fees: which fees apply to which item (a product or a payment type) on a site.
Reads the item fees of the items that `--item` names. Emporix requires a site, which `--site`
names in the query and in the body the flags make, but does not filter by it: the item
fees of other sites come too. `--item` takes the YRN of an item,
e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;).

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
ITEM is the item's YRN, ID the item fee's id and FEES the ids of its fees, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list)
shows them; `-o json` shows the response as sent.

No OAuth scope is needed.

**Command:** `epx fees assignments search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ITEM, ID, SITE, FEES.
- `--item <YRN>`\
  The YRN of an item whose fees to read (repeatable; required without `--body`).
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments search-by-product

Search the item fees of a product on sites.

Reads item fees: which fees apply to which item (a product or a payment type) on a site.
Reads the item fees of the product that `--product` names on the sites that `--site` names: its
id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.
Emporix pages this search in its body, so epx sets pageNumber and pageSize there, also in
the body of `--body`.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
ITEM is the item's YRN, ID the item fee's id and FEES the ids of its fees, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list)
shows them; `-o json` shows the response as sent.

No OAuth scope is needed.

**Command:** `epx fees assignments search-by-product [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ITEM, ID, SITE, FEES.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--product <ID>`\
  The ID of the product whose fees to read (required without `--body`).
- `--site <code>`\
  A site code to read them on, e.g. main (repeatable; required without `--body`).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments search-by-products

Search the item fees of products on a site.

Reads item fees: which fees apply to which item (a product or a payment type) on a site.
Reads the item fees of the products that `--products` names on the site that `--site` names:
their ids (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not their codes, separated by commas.
Emporix pages this search in its body, so epx sets pageNumber and pageSize there, also in
the body of `--body`.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
ITEM is the item's YRN, ID the item fee's id and FEES the ids of its fees, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list)
shows them; `-o json` shows the response as sent.

No OAuth scope is needed.

**Command:** `epx fees assignments search-by-products [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ITEM, ID, SITE, FEES.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--products <IDs>`\
  The IDs of the products whose fees to read, separated by commas, e.g. 100200,100201 (required without `--body`).
- `--site <code>`\
  The site code to read them on, e.g. main (required without `--body`).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments update

Add fees to an item.

Adds the fees that `--fee` names, or the feeIds of `--file`, a JSON object such as
`{"feeIds": ["f1", "f2"]}`, to those of the item YRN, creating its item fee where it
has none. Emporix removes no fee, with or without `--partial`, and an empty list changes
nothing: delete-fee removes one.
Without `--site`, Emporix acts on the item fee on site main.
YRN names the item, e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the fee.item\_update scope.

**Command:** `epx fees assignments update YRN [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--fee <ID>`\
  The ID of a fee, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) shows it (repeatable; required without `--file`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--partial`\
  Send partial=true instead of partial=false, which Emporix treats alike.
- `--site <CODE>`\
  The site CODE of the item fee, sent as siteCode.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees assignments update-product

Add fees to a product on a site.

Adds the fees that `--fee` names, or the feeIds of `--file`, a JSON object such as
`{"feeIds": ["f1", "f2"]}`, to those of the product PRODUCT\_ID
on the site that `--site` names, creating its item fee where it
has none. Emporix removes no fee, with or without `--partial`, and an empty list changes
nothing: delete-product-fee removes one.
PRODUCT\_ID is the product id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the fee.item\_update scope.

**Command:** `epx fees assignments update-product PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--fee <ID>`\
  The ID of a fee, as [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) shows it (repeatable; required without `--file`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--partial`\
  Send partial=true instead of partial=false, which Emporix treats alike.
- `--site <site>`\
  Site code, e.g. main (required).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees create

Create a fee.

Creates the fee in `--file`, a JSON object with code (unique in the tenant), name, feeType
(PERCENT, ABSOLUTE or ABSOLUTE\_MULTIPLY\_ITEMQUANTITY), feePercentage or feeAbsolute (amount
and currency), siteCode and active, and optionally itemType (PRODUCT or PAYMENTTYPE),
taxable with taxCode, and activeTimespan, e.g. `{"code": "deposit", "name": {"en": "Deposit"}, "feeType": "ABSOLUTE", "feeAbsolute": {"amount": 0.25, "currency": "EUR"}, "siteCode": "AcmeStore_NL", "active": true}`. Without an id, Emporix makes one; the response
gives it and its YRN. A code that the tenant has already is a conflict (exit 5).

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. The Fee Service ignores
Content-Language: it stores a string as sent, in no language, in place of any
translations.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the fee.fee\_create scope.

**Command:** `epx fees create [flags]`

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

## epx fees delete

Delete a fee.

Deletes the fee ID, which no item has then.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the fee.fee\_delete scope.

**Command:** `epx fees delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees for-item

List the fees of an item on a site.

Reads fees, surcharges on items such as products or payment types. Emporix filters fees
by site and documents main as the default site, so list, for-product and for-item require
`--site`; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or
ABSOLUTE\_MULTIPLY\_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage
or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or
PAYMENTTYPE), and FROM and TO the active time span.
YRN names the item, e.g. `'urn:yaas:saasag:caasproduct:product:TENANT;PRODUCT_ID'`
(quoted, because the shell would end the command at ;). An unknown item lists nothing;
Emporix fails on a malformed YRN.

No OAuth scope is needed.

**Command:** `epx fees for-item YRN [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, NAME, TYPE, AMOUNT, ITEM TYPE, SITE, ACTIVE, TAXABLE, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees for-product

List the fees of a product on a site.

Reads fees, surcharges on items such as products or payment types. Emporix filters fees
by site and documents main as the default site, so list, for-product and for-item require
`--site`; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or
ABSOLUTE\_MULTIPLY\_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage
or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or
PAYMENTTYPE), and FROM and TO the active time span.
PRODUCT\_ID is the product id (as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)), not its code.
An unknown id lists nothing.

No OAuth scope is needed.

**Command:** `epx fees for-product PRODUCT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, NAME, TYPE, AMOUNT, ITEM TYPE, SITE, ACTIVE, TAXABLE, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees get

Show one fee.

Reads fees, surcharges on items such as products or payment types. Emporix filters fees
by site and documents main as the default site, so list, for-product and for-item require
`--site`; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or
ABSOLUTE\_MULTIPLY\_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage
or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or
PAYMENTTYPE), and FROM and TO the active time span.

No OAuth scope is needed.

**Command:** `epx fees get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, NAME, TYPE, AMOUNT, ITEM TYPE, SITE, ACTIVE, TAXABLE, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees list

List the fees of a site.

Reads fees, surcharges on items such as products or payment types. Emporix filters fees
by site and documents main as the default site, so list, for-product and for-item require
`--site`; an unknown site lists nothing. TYPE is PERCENT, ABSOLUTE or
ABSOLUTE\_MULTIPLY\_ITEMQUANTITY (the amount times the item quantity), AMOUNT the percentage
or the amount in its currency, ITEM TYPE the type of item the fee applies to (PRODUCT or
PAYMENTTYPE), and FROM and TO the active time span.

No OAuth scope is needed.

**Command:** `epx fees list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, NAME, TYPE, AMOUNT, ITEM TYPE, SITE, ACTIVE, TAXABLE, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--site <site>`\
  Site code, e.g. main (required).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx fees update

Replace or change a fee.

Replaces the fee ID with the one in `--file`, a JSON object as create takes it, or with
`--partial` changes only the members that it gives, which still include code, siteCode,
active and feeType (Emporix refuses the body otherwise). Emporix answers an unknown ID
with 409 Conflict (exit 5), not as not found.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. The Fee Service ignores
Content-Language: it stores a string as sent, in no language, in place of any
translations.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the fee.fee\_update scope.

**Command:** `epx fees update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--partial`\
  Change only the fields that the body gives; without it, the body replaces the whole fee.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

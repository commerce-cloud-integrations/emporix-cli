# epx carts

> Read and change carts and their items and discounts.

## epx carts

Reads shopping carts, personal data included, with their items, discounts, delivery
restrictions and validation; creates, changes and deletes carts, their items and discounts.

Its reads require the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the cart.cart_manage scope.

**Command:** `epx carts [flags]`, `epx carts [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx carts change-currency`](#epx-carts-change-currency)\
  Change the currency of a cart.
- [`epx carts change-site`](#epx-carts-change-site)\
  Move a cart to another site.
- [`epx carts create`](#epx-carts-create)\
  Create a customer's cart.
- [`epx carts delete`](#epx-carts-delete)\
  Delete a cart.
- [`epx carts delivery-restrictions`](#epx-carts-delivery-restrictions)\
  Show the lead time and non-delivery days of a cart.
- [`epx carts discounts`](#epx-carts-discounts)\
  Read, apply and remove the discounts of a cart.
- [`epx carts execute`](#epx-carts-execute)\
  Run a chain of cart commands.
- [`epx carts find`](#epx-carts-find)\
  Show the cart of a customer or of an anonymous session on a site.
- [`epx carts get`](#epx-carts-get)\
  Show one cart.
- [`epx carts items`](#epx-carts-items)\
  Read and change the items of a cart.
- [`epx carts merge`](#epx-carts-merge)\
  Merge anonymous carts into a customer's cart.
- [`epx carts refresh`](#epx-carts-refresh)\
  Refresh the prices of a cart's items.
- [`epx carts search`](#epx-carts-search)\
  Search carts.
- [`epx carts update`](#epx-carts-update)\
  Update a cart.
- [`epx carts validate`](#epx-carts-validate)\
  Check the prices of a cart's items.

## epx carts change-currency

Change the currency of a cart.

Changes the currency of the cart CART_ID to the one that `--currency` names; Emporix
matches the prices of the cart's items in it.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts change-currency CART_ID [flags]`

**Flags**

- `--currency <code>`\
  The three-letter code of the currency, e.g. USD (required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts change-site

Move a cart to another site.

Moves the cart CART_ID to the site that `--site` names, whose language, currency,
shipping, tax and payment the cart then takes; where the site's currency differs, Emporix
matches the prices of the cart's items in it.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts change-site CART_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <code>`\
  The code of the site to move the cart to (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts create

Create a customer's cart.

Creates a customer's cart from the JSON object of `--file`: its customerId, siteCode,
currency and type (such as shopping or wishlist; a customer may hold one open cart of
each type: another is a conflict, exit 5), and optionally legalEntityId, channel,
deliveryWindow, addresses and mixins. Emporix answers with the cartId. A cart of an
anonymous customer needs, instead of a customerId, the session-id header of the
customer's session, which epx does not send: without either, Emporix answers 400 Bad
Request.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts create [flags]`

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

## epx carts delete

Delete a cart.

Deletes the cart CART_ID with its items. A CLOSED cart, such as one merged into
another, is not found.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CART_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts delete CART_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts delivery-restrictions

Show the lead time and non-delivery days of a cart.

Reads what limits the delivery of a cart's items: LEAD HOURS is the time needed to prepare
them for delivery, the longest leadTime of the items' own mixins.productCustomAttributes,
and NON-DELIVERY DAYS the days on which they cannot be delivered, the union of their
nonDelivery days, as numbers from 0 (Monday). An item gets these mixins where it is added
([`epx carts items create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-items-create) takes them); Emporix ignores those of the item's
product and of its product member. The cart itself also carries leadTime and nonDelivery.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts delivery-restrictions CART_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among LEAD HOURS, NON-DELIVERY DAYS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts discounts

Read, apply and remove the discounts of a cart.

Reads the discounts (coupons) applied to a cart; applies and removes them.

Its reads require the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the cart.cart_manage scope.

**Command:** `epx carts discounts [flags]`, `epx carts discounts [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx carts discounts create`](#epx-carts-discounts-create)\
  Apply a discount to a cart.
- [`epx carts discounts delete`](#epx-carts-discounts-delete)\
  Remove a discount from a cart.
- [`epx carts discounts delete-all`](#epx-carts-discounts-delete-all)\
  Remove the discounts of a cart.
- [`epx carts discounts list`](#epx-carts-discounts-list)\
  List the discounts of a cart.

## epx carts discounts create

Apply a discount to a cart.

Applies a discount to the cart CART_ID: the coupon that `--code` names, or the JSON
object of `--file`, with the coupon's code and optionally its name, amount or discountRate,
currency and discountCalculationType (SUBTOTAL or TOTAL). One discount per request.
Emporix may limit the coupons per cart (a tenant setting; one tenant allows one: a second is
refused with 400 Bad Request, exit 1). It refuses a coupon restricted to customers
(restrictions.validFor), even to the cart's own customer, with 403 Forbidden (exit 1),
`"User not authorized."`, when an API key credential such as epx's applies it.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts discounts create CART_ID [flags]`

**Flags**

- `--code <code>`\
  The code of the coupon to apply (required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts discounts delete

Remove a discount from a cart.

Removes the discount at INDEX from the cart CART_ID. The discounts after it then move
up by one: list them again before you remove another by INDEX, or remove discounts by
code with delete-all `--codes`.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
INDEX is the INDEX that [`epx carts discounts list CART_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-discounts-list) shows; an unknown one is not found.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type INDEX at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts discounts delete CART_ID INDEX [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts discounts delete-all

Remove the discounts of a cart.

Removes every discount from the cart CART_ID, or with `--codes` only those with the codes
it names.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Deletes every discount of the cart (or those that `--codes` names) with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type CART_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts discounts delete-all CART_ID [flags]`

**Flags**

- `--codes <CODES>`\
  Remove only the discounts with these CODES, separated by commas.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts discounts list

List the discounts of a cart.

Reads the discounts (coupons) applied to a cart: AMOUNT is a fixed amount, RATE a percentage
of the price, APPLIES TO is SUBTOTAL or TOTAL, and INDEX its position among the cart's
discounts. Whether Emporix counts a discount when it calculates the cart (valid) is not
sent here, only with the cart's own discounts: [`epx carts get CART_ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-get).
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts discounts list CART_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, AMOUNT, RATE, APPLIES TO, INDEX.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts execute

Run a chain of cart commands.

Runs a chain of commands on the cart CART_ID in one request, from the JSON object of
`--file`: `{"commands": [...]}`, each with its type (AddCartItem, UpdateCartItem,
DeleteCartItem, DeleteCartItems, GetCart, AddCartItemsBatch, UpdateCartItemsBatch,
UpdateCart, ApplyCartDiscount, GetCartDiscounts, DeleteCartDiscounts, DeleteCartDiscount,
RefreshCart or ValidateCart), the body of that operation as data, and the rest of its
request as options (itemId, partial, expandCalculation, resourceVersion, codes,
discountIndex, …). epx prints the result of each command that ran, the data of each
as its operation answers. Without `--on-error` resume, Emporix runs no command after one
that failed, but keeps what those before it changed. With `--versioning` explicit, Emporix
answers a stale resourceVersion as a failed command with code 500 (its message names 409
CONFLICT), so epx exits 1 and says it may have been carried out; it was not.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Emporix takes up to 10 commands in one request (epx refuses more), runs them in
order and answers with the result of each that ran: when one failed, epx prints the
response, lists the commands that failed on standard error (of a command that adds
or updates items in a batch, the items that failed; Emporix carried out the others)
and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts execute CART_ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow a command with an external price, product, fee or discount (requests cart.cart_manage_external_prices too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--on-error <string>`\
  Whether Emporix stops after a command that failed (fail, the default) or runs the rest (resume): fail or resume.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--versioning <string>`\
  How the writes check the cart's version (default skip): skip checks none, explicit each command's options.resourceVersion, follow the first's, then the version each write makes: skip, explicit or follow.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx carts execute 6a86b20c2f5961330a8b3eb6 --file add-then-get.json --dry-run
epx carts execute 6a86b20c2f5961330a8b3eb6 --file chain.json --on-error resume --yes
```

## epx carts find

Show the cart of a customer or of an anonymous session on a site.

Reads the one cart of a site that the criteria identify, personal data included: a
customer's (`--customer`) or an anonymous session's (`--session`). Without `--legal-entity`, it
is the cart without a legal entity, and without `--type` the cart without a type. epx never
asks Emporix to create a missing cart; when none matches, Emporix answers not found (exit 4).
STATUS is OPEN or CLOSED (sent by search only), CUSTOMER the customer id
(empty for an anonymous cart, whose sessionId `-o json` shows), ITEMS the number of items and
TOTAL the gross final price (items, shipping and payment fee) in the cart's currency.
A CLOSED cart, such as one merged into another, is listed by search but not found by get.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts find [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, STATUS, SITE, TYPE, CUSTOMER, LEGAL ENTITY, ITEMS, TOTAL, MODIFIED.
- `--customer <ID>`\
  The customer ID whose cart to read.
- `--legal-entity <ID>`\
  The legal entity ID of the cart.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--session <ID>`\
  The ID of the anonymous session whose cart to read.
- `--site <site>`\
  Site code, e.g. main (required).
- `--type <type>`\
  The cart's type.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts get

Show one cart.

Reads shopping carts, personal data included.
ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
STATUS is OPEN or CLOSED (sent by search only), CUSTOMER the customer id
(empty for an anonymous cart, whose sessionId `-o json` shows), ITEMS the number of items and
TOTAL the gross final price (items, shipping and payment fee) in the cart's currency.
A CLOSED cart, such as one merged into another, is listed by search but not found by get.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts get ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, STATUS, SITE, TYPE, CUSTOMER, LEGAL ENTITY, ITEMS, TOTAL, MODIFIED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items

Read and change the items of a cart.

Reads the items of a cart; adds items to it, changes and removes them.

Its reads require the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the cart.cart_manage scope.

**Command:** `epx carts items [flags]`, `epx carts items [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx carts items bulk-create`](#epx-carts-items-bulk-create)\
  Add products to a cart.
- [`epx carts items bulk-update`](#epx-carts-items-bulk-update)\
  Update items of a cart.
- [`epx carts items create`](#epx-carts-items-create)\
  Add a product to a cart.
- [`epx carts items delete`](#epx-carts-items-delete)\
  Remove an item from a cart.
- [`epx carts items delete-all`](#epx-carts-items-delete-all)\
  Remove every item from a cart.
- [`epx carts items get`](#epx-carts-items-get)\
  Show one item of a cart.
- [`epx carts items list`](#epx-carts-items-list)\
  List the items of a cart.
- [`epx carts items update`](#epx-carts-items-update)\
  Update an item of a cart.

## epx carts items bulk-create

Add products to a cart.

Adds the items in `--file`, a JSON array of items, each as create takes it, to the cart
CART_ID.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts items bulk-create CART_ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow an item with an external price, product, fee or discount (requests cart.cart_manage_external_prices too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items bulk-update

Update items of a cart.

Updates the items in `--file`, a JSON array of items, each with the id of an item of the
cart CART_ID and the whole item, as update takes it without `--partial`. Emporix
refuses an entry with only quantity for that item (400 "The Price must be
defined."). [`epx carts execute`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-execute) changes some members of several items
(UpdateCartItemsBatch with options.partial).
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Emporix takes up to 50 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts items bulk-update CART_ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow an item with an external price, product, fee or discount (requests cart.cart_manage_external_prices too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items create

Add a product to a cart.

Adds a product to the cart CART_ID as a new item, from the JSON object of `--file`: its
product (its id) or itemYrn, quantity and price (priceId, originalAmount,
effectiveAmount and currency, as [`epx prices match`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/prices#epx-prices-match) finds them), and optionally
keepAsSeparateLineItem, externalFees, externalDiscounts and mixins.
`--site` names the cart's site. Emporix answers 201 with the itemId (e.g. 0 for the first
item) and yrn. A product that the cart holds already, added without
keepAsSeparateLineItem, raises the quantity of its item (the same itemId), so a create
sent twice adds twice. A product object
in the body is stored as sent, in place of the one Emporix would store. The item's
mixins.productCustomAttributes (leadTime, nonDelivery) feed
[`epx carts delivery-restrictions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-delivery-restrictions).
Emporix refuses a product that is not published with 400 Bad Request (exit 1), and an
item with an external price without `--allow-external-prices` with 403 Forbidden
(exit 3).
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts items create CART_ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow an item with an external price, product, fee or discount (requests cart.cart_manage_external_prices too).
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

## epx carts items delete

Remove an item from a cart.

Removes an item from the cart CART_ID.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
ID is the ID that [`epx carts items list CART_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-items-list) shows; an unknown one is not found.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts items delete CART_ID ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items delete-all

Remove every item from a cart.

Removes every item from the cart CART_ID.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Deletes every item of the cart with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type CART_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts items delete-all CART_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items get

Show one item of a cart.

Reads one item of a cart.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
ID is the ID that [`epx carts items list CART_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-items-list) shows; an unknown one is not found.
PRODUCT is the product id, NAME the product's name, PRICE the price that matched the item,
per unit and as the price states it (net or gross), and GROSS the gross value of the item's
quantity before discounts, one per tax line. Emporix does not always send an item's gross
unit price and total (after discounts, with fees) here: only for the items that the cart
held when it was last calculated (by a read of the cart or a discount, for example), not
for those added since. The cart always has them: [`epx carts get CART_ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-get).

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts items get CART_ID ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PRODUCT, NAME, QUANTITY, PRICE, GROSS.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items list

List the items of a cart.

Reads the items of a cart.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
PRODUCT is the product id, NAME the product's name, PRICE the price that matched the item,
per unit and as the price states it (net or gross), and GROSS the gross value of the item's
quantity before discounts, one per tax line. Emporix does not always send an item's gross
unit price and total (after discounts, with fees) here: only for the items that the cart
held when it was last calculated (by a read of the cart or a discount, for example), not
for those added since. The cart always has them: [`epx carts get CART_ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-get).

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts items list CART_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PRODUCT, NAME, QUANTITY, PRICE, GROSS.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts items update

Update an item of a cart.

Replaces an item of the cart CART_ID with the JSON object of `--file`, whose members are
those that create takes, or with `--partial` changes only the members it gives.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.
ID is the ID that [`epx carts items list CART_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-items-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts items update CART_ID ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow an item with an external price, product, fee or discount (requests cart.cart_manage_external_prices too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--partial`\
  Change only the members that the body gives; without it, the body replaces the whole item.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts merge

Merge anonymous carts into a customer's cart.

Merges the anonymous carts that `--cart` names into the cart CART_ID, which must belong
to a customer. Emporix answers with no body and closes the merged carts rather than
deleting them: [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) still lists them as CLOSED, but get and delete no
longer find them.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts merge CART_ID [flags]`

**Flags**

- `--cart <ID>`\
  The ID of an anonymous cart to merge (repeatable; required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts refresh

Refresh the prices of a cart's items.

Refreshes the cart CART_ID: Emporix reads the prices of its items again and assigns
those that changed. It sends no body.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the cart.cart_manage scope.

**Command:** `epx carts refresh CART_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts search

Search carts.

Reads the carts, personal data included, that an Emporix q filter finds (`--query`; without
it, every cart), e.g. `--query` `'status:OPEN siteCode:AcmeStore_NL'`.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
STATUS is OPEN or CLOSED (sent by search only), CUSTOMER the customer id
(empty for an anonymous cart, whose sessionId `-o json` shows), ITEMS the number of items and
TOTAL the gross final price (items, shipping and payment fee) in the cart's currency.
A CLOSED cart, such as one merged into another, is listed by search but not found by get.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts search [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, STATUS, SITE, TYPE, CUSTOMER, LEGAL ENTITY, ITEMS, TOTAL, MODIFIED.
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

## epx carts update

Update a cart.

Updates the cart CART_ID with the JSON object of `--file`: customerId, currency, type,
status (OPEN or CLOSED), deliveryWindow, zipCode and countryCode (both or neither: one
alone is answered 400 Bad Request, exit 1), addresses, mixins and externalDiscounts, as
[`epx carts get CART_ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-get) shows them. Emporix creates no missing cart, and
ignores metadata.version in the body: only [`epx carts execute --versioning`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-execute) checks the
cart's version.
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--allow-external-prices` also requests the cart.cart_manage_external_prices scope.

Requires the cart.cart_manage scope.

**Command:** `epx carts update CART_ID [flags]`

**Flags**

- `--allow-external-prices`\
  Allow external discounts (externalDiscounts) (requests cart.cart_manage_external_prices too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx carts validate

Check the prices of a cart's items.

Reads whether a cart's items pass Emporix's checks, which find prices that are no longer
valid and prices duplicated for a product; it changes nothing. VALID is true or false and
ERRORS the error codes by item id (`-o json` adds each error's message and field).
CART_ID is the ID that [`epx carts search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts#epx-carts-search) shows; an unknown one is not found.

Requires the cart.cart_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx carts validate CART_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting cart.cart_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among VALID, ERRORS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

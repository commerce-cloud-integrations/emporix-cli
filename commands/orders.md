# epx orders

> Read, create and change orders and their status.

## epx orders

Reads, creates and changes sales orders, personal data included, and their status:
[`epx orders transition`](/emporix-cli/commands/orders#epx-orders-transition) moves an order to another status, `add-entries` and `recalculate`
change its entries, and `split` splits it into sub-orders by vendor.

Its reads require the order.order\_read scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx orders [flags]`, `epx orders [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx orders add-entries`](#epx-orders-add-entries)\
  Add products to an order, or change its entries.
- [`epx orders count`](#epx-orders-count)\
  Count orders.
- [`epx orders create`](#epx-orders-create)\
  Create an order.
- [`epx orders delete`](#epx-orders-delete)\
  Delete one order.
- [`epx orders get`](#epx-orders-get)\
  Show one order.
- [`epx orders history`](#epx-orders-history)\
  Show the status changes of an order.
- [`epx orders list`](#epx-orders-list)\
  List orders.
- [`epx orders patch`](#epx-orders-patch)\
  Change members of an order.
- [`epx orders recalculate`](#epx-orders-recalculate)\
  Set the delivered amounts of an order and recalculate it.
- [`epx orders search`](#epx-orders-search)\
  Search orders.
- [`epx orders split`](#epx-orders-split)\
  Split an order into sub-orders by vendor.
- [`epx orders transition`](#epx-orders-transition)\
  Move an order to another status.
- [`epx orders transitions`](#epx-orders-transitions)\
  List the statuses an order can move to next.
- [`epx orders update`](#epx-orders-update)\
  Replace an order.

## epx orders add-entries

Add products to an order, or change its entries.

Adds products to the order ORDER\_ID, or changes the quantity of its entries, and
recalculates it: the JSON object in `--file` holds entries, each with the productId of a
product to add or the entryId of an entry to change (an id in the order's entries, as
[`epx orders get ID -o json`](/emporix-cli/commands/orders#epx-orders-get) prints them), and its quantity; 0 removes the entry.
Emporix answers with the order recalculated, which it stores only with `--store`: without
it, the order stays as it is. Emporix refuses to change the entries of a COMPLETED order
with 400 Bad Request (exit 1).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the order.order\_update scope.

**Command:** `epx orders add-entries ORDER_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--store`\
  Store the recalculated order; without it, Emporix only answers with it.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders count

Count orders.

Prints the number of sales orders, or of those that an Emporix q filter finds
(`--query`), which Emporix reports in a header without sending the orders: a master order
and each of its sub-orders count once, as list lists them. `-o json` prints the number alone.
Emporix counts 0 for a filter on id or customer.id although list finds the orders (seen
live): list them instead.

Requires the order.order\_read scope.

**Command:** `epx orders count [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among COUNT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders create

Create an order.

Creates an order on behalf of a customer from the JSON object in `--file`, shaped as
[`epx orders get ID -o json`](/emporix-cli/commands/orders#epx-orders-get) prints one (entries, customer, siteCode, currency and
calculatedPrice among others). Emporix makes the order's id and answers with it: an id in
the body is ignored, so sending the body again creates another order (with checkout
true, at most one per cartId). It stores the status CREATED for IN\_CHECKOUT, and sets
mixins.generalAttributes (extendedOrderStatus `"10"`, orderNumber and customerFirstOrder)
whatever the body says; [`epx orders patch`](/emporix-cli/commands/orders#epx-orders-patch) can change them afterwards, e.g. to make the
order ready for [`epx invoice-jobs create`](/emporix-cli/commands/invoice-jobs#epx-invoice-jobs-create). It also adds an address of the order to the
customer's addresses, as a default one tagged shipping and billing, so a customer who had
a default address then has two default addresses.

Give the order's and its entries' metadata.mixins the URLs of the schemas of their mixins,
as orders placed in a checkout carry them: without them, [`epx orders add-entries --store`](/emporix-cli/commands/orders#epx-orders-add-entries),
`recalculate`, `patch --recalculate` and `split` fail on the order with 400 Bad Request
(exit 1).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the order.order\_create scope.

**Command:** `epx orders create [flags]`

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

## epx orders delete

Delete one order.

Deletes the order ID; it sends no body.
ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the order.order\_delete scope.

**Command:** `epx orders delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders get

Show one order.

Reads sales orders, personal data included. TYPE is MASTER\_ORDER or SUB\_ORDER for an order
split, for example by vendor: the master and its sub-orders are all listed, so do not add
up their totals. STATUS is IN\_CHECKOUT, CREATED, CONFIRMED, DECLINED, SHIPPED or COMPLETED,
CUSTOMER the customer's email (or name), ITEMS the number of order entries and TOTAL the
gross final price in the order's currency (empty when Emporix sends only the deprecated
totalPrice, which `-o json` shows, as for an order that [`epx orders add-entries --store`](/emporix-cli/commands/orders#epx-orders-add-entries),
`recalculate` or `split` rewrote, and for sub-orders). The orders of a legal entity:
[`epx orders list --query legalEntityId:ID`](/emporix-cli/commands/orders#epx-orders-list).

Requires the order.order\_read scope.

**Command:** `epx orders get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, CREATED, CUSTOMER, SITE, ITEMS, TOTAL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders history

Show the status changes of an order.

Reads the status changes of an order, in the order Emporix sends them. `-o json` adds the
history's metadata.
ORDER\_ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows; an unknown one is not found.

Requires the order.order\_read scope.

**Command:** `epx orders history ORDER_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among STATUS, TIMESTAMP.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders list

List orders.

Reads sales orders, personal data included. TYPE is MASTER\_ORDER or SUB\_ORDER for an order
split, for example by vendor: the master and its sub-orders are all listed, so do not add
up their totals. STATUS is IN\_CHECKOUT, CREATED, CONFIRMED, DECLINED, SHIPPED or COMPLETED,
CUSTOMER the customer's email (or name), ITEMS the number of order entries and TOTAL the
gross final price in the order's currency (empty when Emporix sends only the deprecated
totalPrice, which `-o json` shows, as for an order that [`epx orders add-entries --store`](/emporix-cli/commands/orders#epx-orders-add-entries),
`recalculate` or `split` rewrote, and for sub-orders). The orders of a legal entity:
[`epx orders list --query legalEntityId:ID`](/emporix-cli/commands/orders#epx-orders-list).

Requires the order.order\_read scope.

**Command:** `epx orders list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, CREATED, CUSTOMER, SITE, ITEMS, TOTAL.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders patch

Change members of an order.

Changes the members of the order ID that the JSON object in `--file` gives, such as its
shippingAddress or mixins, and keeps the others.
Emporix does not check the order's metadata.version: it applies the change whatever version
the body names, and does not answer a stale one with 409 Conflict; an unknown order is not
found (exit 4).
ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--completed` also requests the order.order\_update\_completed scope.

Requires the order.order\_update scope.

**Command:** `epx orders patch ID [flags]`

**Flags**

- `--completed`\
  For an order that is COMPLETED, which Emporix changes only with one more scope (requests order.order\_update\_completed too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--recalculate`\
  Have Emporix recalculate the whole order, coupon discounts included.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders recalculate

Set the delivered amounts of an order and recalculate it.

Sets the amounts delivered of entries of the order ORDER\_ID and recalculates the
order: the JSON object in `--file` holds entries, each with its entryId (an id in the
order's entries, as [`epx orders get ID -o json`](/emporix-cli/commands/orders#epx-orders-get) prints them) and deliveredAmount, and
optionally deliveredAmountInPc and unit, the unit of deliveredAmount (e.g. GRM).
Emporix refuses to recalculate a COMPLETED order with 400 Bad Request (exit 1), whatever
the scopes.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Emporix answers with a sentence of text, which epx prints as a JSON string.

Requires the order.order\_update scope.

**Command:** `epx orders recalculate ORDER_ID [flags]`

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

## epx orders search

Search orders.

Reads the sales orders, personal data included, that an Emporix q filter finds
(`--query`; without it, every order).
It reads and shows them as [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the order.order\_read scope.

**Command:** `epx orders search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, CREATED, CUSTOMER, SITE, ITEMS, TOTAL.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders split

Split an order into sub-orders by vendor.

Splits the order ORDER\_ID into sub-orders, one per vendor of its entries, sending
splitBy VENDOR\_ID, the only criterion that Emporix offers, unless `--file` gives the body.
Emporix splits only an order with status CREATED that was not split before and has no
discounts: it becomes the MASTER\_ORDER of the SUB\_ORDERs that the response names. An order
whose entries name no vendor stays as it is (the response names no subOrders); an order
split before, or a sub-order, is refused with 400 Bad Request (exit 1).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the order.order\_update scope.

**Command:** `epx orders split ORDER_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders transition

Move an order to another status.

Moves the order ORDER\_ID to the status that `--status` names, one that
[`epx orders transitions ORDER_ID`](/emporix-cli/commands/orders#epx-orders-transitions) lists: CONFIRMED or DECLINED from CREATED, COMPLETED or
DECLINED from CONFIRMED, and none from COMPLETED or DECLINED. Emporix refuses another with
403 Forbidden (exit 1) and says why, and answers the order's own status as done,
without an entry in its history. ORDER\_ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the order.order\_update scope.

**Command:** `epx orders transition ORDER_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--status <STATUS>`\
  The STATUS to move the order to: IN\_CHECKOUT, CREATED, CONFIRMED, DECLINED, SHIPPED or COMPLETED (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders transitions

List the statuses an order can move to next.

Reads the statuses that an order can move to next.
ORDER\_ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows; an unknown one is not found.

Requires the order.order\_read scope.

**Command:** `epx orders transitions ORDER_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among STATUS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx orders update

Replace an order.

Replaces the order ID with the order in `--file`, a JSON object as
[`epx orders get ID -o json`](/emporix-cli/commands/orders#epx-orders-get) prints it.
Emporix does not check the order's metadata.version: it applies the change whatever version
the body names, and does not answer a stale one with 409 Conflict; an unknown order is not
found (exit 4).
ID is the ID that [`epx orders list`](/emporix-cli/commands/orders#epx-orders-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--completed` also requests the order.order\_update\_completed scope.

Requires the order.order\_update scope.

**Command:** `epx orders update ID [flags]`

**Flags**

- `--completed`\
  For an order that is COMPLETED, which Emporix changes only with one more scope (requests order.order\_update\_completed too).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--recalculate`\
  Have Emporix recalculate the whole order, coupon discounts included.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

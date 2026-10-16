# epx shopping-lists

> Read, create, replace and delete customers' shopping lists.

## epx shopping-lists

Reads customers' shopping lists, personal data included: each customer's lists by name,
with how many items they hold. Items are only in `-o json`. CUSTOMER is the customer id the
lists are kept under (the reference calls it the customer number; for a customer that
Emporix numbered, the two are the same); get takes it, and a customer without shopping
lists is not found.
create, update and delete create, replace and delete a customer's lists.

Its reads require the shoppinglist.shoppinglist\_read scope.
Its writes require the shoppinglist.shoppinglist\_manage scope.

**Command:** `epx shopping-lists [flags]`, `epx shopping-lists [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx shopping-lists create`](#epx-shopping-lists-create)\
  Create a shopping list of a customer.
- [`epx shopping-lists delete`](#epx-shopping-lists-delete)\
  Delete a shopping list of a customer.
- [`epx shopping-lists delete-all`](#epx-shopping-lists-delete-all)\
  Delete every shopping list of a customer.
- [`epx shopping-lists get`](#epx-shopping-lists-get)\
  Show one customer's shopping lists.
- [`epx shopping-lists list`](#epx-shopping-lists-list)\
  List customers' shopping lists.
- [`epx shopping-lists update`](#epx-shopping-lists-update)\
  Replace a shopping list of a customer.

## epx shopping-lists create

Create a shopping list of a customer.

Creates the shopping list in `--file`, a JSON object with the customerId of the customer it
belongs to, the list's name and its items (each with its id, productId and quantity), and
prints what Emporix answers with: the customer's id. A list of that name that the customer
has already is a conflict (exit 5). Emporix does not check the customer: it creates the list
of an unknown customer id too.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shoppinglist.shoppinglist\_manage scope.

**Command:** `epx shopping-lists create [flags]`

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

**Examples**

```sh
epx shopping-lists create --file list.json --dry-run
```

## epx shopping-lists delete

Delete a shopping list of a customer.

Deletes the shopping list `--name` of CUSTOMER; its other lists stay.
A customer without shopping lists, or without a list of that name, is not found (exit 4).
[`epx shopping-lists get CUSTOMER`](/emporix-cli/commands/shopping-lists#epx-shopping-lists-get) shows the lists that are left.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CUSTOMER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the shoppinglist.shoppinglist\_manage scope.

**Command:** `epx shopping-lists delete CUSTOMER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--name <NAME>`\
  The NAME of the list to delete (required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx shopping-lists delete-all

Delete every shopping list of a customer.

Deletes every shopping list of CUSTOMER.
A customer without shopping lists is not found (exit 4).

Deletes every shopping list of the customer with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type CUSTOMER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the shoppinglist.shoppinglist\_manage scope.

**Command:** `epx shopping-lists delete-all CUSTOMER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx shopping-lists get

Show one customer's shopping lists.

Reads customers' shopping lists, personal data included: each customer's lists by name,
with how many items they hold. Items are only in `-o json`. CUSTOMER is the customer id the
lists are kept under (the reference calls it the customer number; for a customer that
Emporix numbered, the two are the same); get takes it, and a customer without shopping
lists is not found.

Requires the shoppinglist.shoppinglist\_read scope.

**Command:** `epx shopping-lists get CUSTOMER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, LISTS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx shopping-lists list

List customers' shopping lists.

Reads customers' shopping lists, personal data included: each customer's lists by name,
with how many items they hold. Items are only in `-o json`. CUSTOMER is the customer id the
lists are kept under (the reference calls it the customer number; for a customer that
Emporix numbered, the two are the same); get takes it, and a customer without shopping
lists is not found.

Requires the shoppinglist.shoppinglist\_read scope.

**Command:** `epx shopping-lists list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, LISTS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx shopping-lists update

Replace a shopping list of a customer.

Replaces the shopping list of CUSTOMER that the JSON in `--file` names with it: an object
with the list's name and its items (each with its id, productId and quantity). Send its
metadata.version as read to replace it only if nobody changed it since: Emporix answers
409 Conflict (exit 5) otherwise.
A customer without shopping lists, or without a list of that name, is not found (exit 4). Update creates no list; create does.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shoppinglist.shoppinglist\_manage scope.

**Command:** `epx shopping-lists update CUSTOMER [flags]`

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

# epx categories

> Read and change categories and the products assigned to them.

## epx categories

Reads categories, published and unpublished.

Its reads require the category.category_read_unpublished scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx categories [flags]`, `epx categories [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx categories assignments`](#epx-categories-assignments)\
  Read and change the products assigned to a category.
- [`epx categories create`](#epx-categories-create)\
  Create a category.
- [`epx categories delete`](#epx-categories-delete)\
  Delete a category.
- [`epx categories for-product`](#epx-categories-for-product)\
  List the categories a product is assigned to.
- [`epx categories get`](#epx-categories-get)\
  Show one category.
- [`epx categories list`](#epx-categories-list)\
  List categories.
- [`epx categories parents`](#epx-categories-parents)\
  List the categories above a category.
- [`epx categories patch`](#epx-categories-patch)\
  Change members of a category.
- [`epx categories search`](#epx-categories-search)\
  Search categories.
- [`epx categories subcategories`](#epx-categories-subcategories)\
  List the categories below a category.
- [`epx categories update`](#epx-categories-update)\
  Upsert a category.

## epx categories assignments

Read and change the products assigned to a category.

Reads the products assigned to a category, assigns products to it and unassigns them.

Its reads require the category.category_read_unpublished scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx categories assignments [flags]`, `epx categories assignments [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx categories assignments bulk-create`](#epx-categories-assignments-bulk-create)\
  Assign products to a category.
- [`epx categories assignments bulk-update`](#epx-categories-assignments-bulk-update)\
  Upsert the assignments of products to a category.
- [`epx categories assignments create`](#epx-categories-assignments-create)\
  Assign a product to a category.
- [`epx categories assignments delete`](#epx-categories-assignments-delete)\
  Unassign a product from a category.
- [`epx categories assignments delete-all`](#epx-categories-assignments-delete-all)\
  Unassign every product from a category.
- [`epx categories assignments delete-assignment`](#epx-categories-assignments-delete-assignment)\
  Delete an assignment of a product to a category.
- [`epx categories assignments delete-product-everywhere`](#epx-categories-assignments-delete-product-everywhere)\
  Unassign a product from every category.
- [`epx categories assignments list`](#epx-categories-assignments-list)\
  List the products assigned to a category.
- [`epx categories assignments update`](#epx-categories-assignments-update)\
  Upsert the assignment of a product to a category.

## epx categories assignments bulk-create

Assign products to a category.

Assigns products to a category: the JSON in `--file` is an array of assignments, each as
in `{"ref": {"id": "100200", "type": "PRODUCT"}}`.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments bulk-create CATEGORY_ID [flags]`

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

## epx categories assignments bulk-update

Upsert the assignments of products to a category.

Assigns products to a category, or updates the assignments that exist (an upsert): the
JSON in `--file` is an array of assignments, each naming a product by its id, as in
`{"ref": {"id": "100200"}}`.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments bulk-update CATEGORY_ID [flags]`

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

## epx categories assignments create

Assign a product to a category.

Assigns a product to a category: the JSON in `--file` names it by its id, as in
`{"ref": {"id": "100200", "type": "PRODUCT"}}`. Emporix answers with the id of the assignment.
[`epx categories assignments update CATEGORY_ID PRODUCT_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-update) assigns one without a body.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments create CATEGORY_ID [flags]`

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

## epx categories assignments delete

Unassign a product from a category.

Deletes the assignment of the product PRODUCT_ID to the category CATEGORY_ID: the PRODUCT
that [`epx categories assignments list CATEGORY_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-list) shows, as update assigns it. The product and the category stay.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRODUCT_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments delete CATEGORY_ID PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories assignments delete-all

Unassign every product from a category.

Deletes the assignments of every product to the category CATEGORY_ID, published or
unpublished, as [`epx categories assignments list CATEGORY_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-list) lists them. The products and the category stay.

Deletes the assignments of every product to the category with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type CATEGORY_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the category.category_manage, category.category_read_unpublished scopes.

**Command:** `epx categories assignments delete-all CATEGORY_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories assignments delete-assignment

Delete an assignment of a product to a category.

Deletes the assignment ASSIGNMENT_ID of a product to the category CATEGORY_ID: the ID that
[`epx categories assignments list CATEGORY_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-list) shows. The product and the category stay.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ASSIGNMENT_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments delete-assignment CATEGORY_ID ASSIGNMENT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories assignments delete-product-everywhere

Unassign a product from every category.

Deletes the assignments of the product PRODUCT_ID to every category, as
[`epx categories for-product PRODUCT_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-for-product) lists them. The product and the categories stay.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.

Deletes the assignments of the product to every category with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type PRODUCT_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments delete-product-everywhere PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories assignments list

List the products assigned to a category.

Reads the products assigned to a category, published and unpublished. PRODUCT is the
product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)) and TYPE the kind of item, a product.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories assignments list CATEGORY_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PRODUCT, TYPE, PUBLISHED, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories assignments update

Upsert the assignment of a product to a category.

Assigns the product PRODUCT_ID to a category, or updates its assignment when it exists (an
upsert). It sends no body.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories assignments update CATEGORY_ID PRODUCT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories create

Create a category.

Creates a category from the JSON in `--file`, shaped as [`epx categories get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-get)
shows one: localizedName (a map of languages to text, as are the other localized
members), published, which Emporix requires, parentId (a root category without one)
and the other members; Emporix generates the id unless the body gives one, and answers
with it. To create it published, give published: true in the body and `--publish`, which
sends the publish parameter that Emporix requires to match it, with the scope it requires
for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--publish` also requests the category.category_publish scope.

Requires the category.category_manage scope.

**Command:** `epx categories create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--publish`\
  Publish the category: sends publish=true, which needs published: true in the body (requests category.category_publish too).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories delete

Delete a category.

Deletes the category ID. Without `--with-subcategories`, the categories directly below it
move to the category above it and keep the categories below them; Emporix refuses to
delete a root category with categories below it (400 Bad Request, exit 1). With it,
Emporix deletes every category below it too.

Deletes data, which cannot be undone.
With `--with-subcategories`, it deletes the category and every category below it with one request.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--with-subcategories`\
  Delete every category below the category too.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories for-product

List the categories a product is assigned to.

Reads categories, published and unpublished.
PRODUCT_ID is the product id (as in [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list)), not its code.
An unknown id lists nothing, because Emporix does not distinguish it
from a product without assignments.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories for-product PRODUCT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, TYPE, PUBLISHED, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories get

Show one category.

Reads categories, published and unpublished.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, TYPE, PUBLISHED, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories list

List categories.

Reads categories, published and unpublished.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, TYPE, PUBLISHED, NAME.
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

## epx categories parents

List the categories above a category.

Reads the categories above a category, published and unpublished, from its root down
unless `--sort` says otherwise.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories parents CATEGORY_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, TYPE, PUBLISHED, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories patch

Change members of a category.

Changes the members of the category ID that the JSON in `--file` gives, and keeps the
others. Emporix requires metadata.version in the body, as [`epx categories get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-get)
shows it, and refuses a version that is not the category's with 400 Bad Request (exit 1),
not as a conflict. published in the body publishes or unpublishes the category as sent,
without the publish scopes. A move (parentId) shows in [`epx categories parents`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-parents) at once,
in [`epx category-trees`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees#epx-category-trees) a few seconds later.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the category.category_manage scope.

**Command:** `epx categories patch ID [flags]`

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

## epx categories search

Search categories.

Reads the categories, published and unpublished, that an Emporix q filter finds
(`--query`; without it, every category).
It reads and shows them as [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CODE, TYPE, PUBLISHED, NAME.
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

## epx categories subcategories

List the categories below a category.

Reads every category below a category, at all levels, published and unpublished; PARENT
is the id of the category each one lies directly below.
CATEGORY_ID is the ID that [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) shows; an unknown one is not found.

Requires the category.category_read_unpublished scope.

**Command:** `epx categories subcategories CATEGORY_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PARENT, CODE, TYPE, PUBLISHED, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx categories update

Upsert a category.

Replaces the category ID with the JSON in `--file`, or creates it with that id when
there is none (an upsert); the body has the members of create's, published included, but
id, type and ownClassificationMixins. Without parentId, the category becomes a root
category.
With metadata.version, as [`epx categories get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-get) shows it, Emporix refuses
the write as a conflict (exit 5) when the category changed since it was read.
To publish or unpublish it, give published in the body and `--publish` or `--unpublish`,
which sends the publish parameter that Emporix requires to match it, with the scope it
requires for it. Emporix requires `--publish` for any body with published: true, also of a
category that is published already.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--publish` also requests the category.category_publish scope.
`--unpublish` also requests the category.category_unpublish scope.

Requires the category.category_manage scope.

**Command:** `epx categories update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--publish`\
  Publish the category: sends publish=true, which needs published: true in the body (requests category.category_publish too).
- `--unpublish`\
  Unpublish the category: sends publish=false, which needs published: false in the body (requests category.category_unpublish too).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

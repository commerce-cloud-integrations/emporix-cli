# epx product-templates

> Read, create, update and delete product templates.

## epx product-templates

Reads and changes product templates, the attribute sets shared by products based on them:
create makes one, update replaces one, each from `--file`; delete deletes one.

Its reads require the product.product\_template\_read scope.
Its writes require the product.product\_template\_manage scope.

**Command:** `epx product-templates [flags]`, `epx product-templates [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx product-templates create`](#epx-product-templates-create)\
  Create a product template.
- [`epx product-templates delete`](#epx-product-templates-delete)\
  Delete a product template.
- [`epx product-templates get`](#epx-product-templates-get)\
  Show one product template.
- [`epx product-templates list`](#epx-product-templates-list)\
  List product templates.
- [`epx product-templates update`](#epx-product-templates-update)\
  Replace a product template.

## epx product-templates create

Create a product template.

Creates the product template in the JSON object of `--file` (name, attributes) and prints its id.

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "gram", "de": "Gramm"}`, as `-o json` of a read prints it, or with `--content-language`
LANGUAGE a string in that language, e.g. `"gram"` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the product.product\_template\_manage scope.

**Command:** `epx product-templates create [flags]`

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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx product-templates delete

Delete a product template.

Deletes the product template ID. Emporix deletes only a template that no product and no
parent variant is based on: it refuses one in use with 400 Bad Request (exit 1). An unknown
ID is no error: Emporix answers as for a deleted template.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the product.product\_template\_manage scope.

**Command:** `epx product-templates delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx product-templates get

Show one product template.

Reads product templates, the attribute sets shared by products based on them.
ATTRIBUTES counts the template's attributes (`-o json` prints their definitions).

Requires the product.product\_template\_read scope.

**Command:** `epx product-templates get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, ATTRIBUTES, VERSION.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx product-templates list

List product templates.

Reads product templates, the attribute sets shared by products based on them.
ATTRIBUTES counts the template's attributes (`-o json` prints their definitions).

Requires the product.product\_template\_read scope.

**Command:** `epx product-templates list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, ATTRIBUTES, VERSION.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx product-templates update

Replace a product template.

Replaces product template ID with the JSON object of `--file`, which names its
metadata.version: Emporix refuses it as a conflict (exit 5) when the template changed since.
Every update makes a new version of the template; the products based on an earlier version keep
it until they are updated to the new one. An unknown ID is not found (exit 4).

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "gram", "de": "Gramm"}`, as `-o json` of a read prints it, or with `--content-language`
LANGUAGE a string in that language, e.g. `"gram"` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the product.product\_template\_manage scope.

**Command:** `epx product-templates update ID [flags]`

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

Also takes the [global flags](/emporix-cli/commands#global-flags).

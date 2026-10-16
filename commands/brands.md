# epx brands

> Read, create, change and delete brands.

## epx brands

Reads brands.

Its reads need no OAuth scope.
Its writes require the brand.brand\_manage scope.

**Command:** `epx brands [flags]`, `epx brands [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx brands create`](#epx-brands-create)\
  Create a brand.
- [`epx brands delete`](#epx-brands-delete)\
  Delete a brand.
- [`epx brands get`](#epx-brands-get)\
  Show one brand.
- [`epx brands list`](#epx-brands-list)\
  List brands.
- [`epx brands patch`](#epx-brands-patch)\
  Change members of a brand.
- [`epx brands update`](#epx-brands-update)\
  Replace a brand.

## epx brands create

Create a brand.

Creates a brand from the JSON in `--file`, such as
`{"name": "Acme", "localizedName": {"en": "Acme", "de": "Acme"}}`, and prints the brand that
Emporix answers with, including its id. Emporix generates the id unless the JSON names one;
an id that exists already is a conflict, while a name that exists already makes a second
brand.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the brand.brand\_manage scope.

**Command:** `epx brands create [flags]`

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

## epx brands delete

Delete a brand.

Deletes the brand ID.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the brand.brand\_manage scope.

**Command:** `epx brands delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx brands get

Show one brand.

Reads brands.

No OAuth scope is needed.

**Command:** `epx brands get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx brands list

List brands.

Reads brands.

No OAuth scope is needed.

**Command:** `epx brands list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME.
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

## epx brands patch

Change members of a brand.

Changes the members of the brand ID that the JSON in `--file` names, such as name or
localizedDescription; the others keep their values. A map of translations in it, such as
localizedName, adds to the brand's: the languages it names change, the others stay. Unlike
update, Emporix does not check a metadata.version in it. An unknown ID is not found. It prints the brand that Emporix answers with.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the brand.brand\_manage scope.

**Command:** `epx brands patch ID [flags]`

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

## epx brands update

Replace a brand.

Replaces the brand ID with the JSON in `--file`: a member that it leaves out is removed.
An unknown ID is not found. It prints the brand that Emporix answers with.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the brand.brand\_manage scope.

**Command:** `epx brands update ID [flags]`

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

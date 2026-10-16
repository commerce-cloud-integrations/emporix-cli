# epx labels

> Read, create, change and delete labels.

## epx labels

Reads labels, the badges shown on products. OVERLAY is the label's overlay flag and
POSITION its overlay position, as sent; MEDIA is the id of its image (image URLs are only
in `-o json`), and DESCRIPTION is often HTML.

Its reads need no OAuth scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx labels [flags]`, `epx labels [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx labels create`](#epx-labels-create)\
  Create a label.
- [`epx labels delete`](#epx-labels-delete)\
  Delete a label.
- [`epx labels get`](#epx-labels-get)\
  Show one label.
- [`epx labels list`](#epx-labels-list)\
  List labels.
- [`epx labels patch`](#epx-labels-patch)\
  Change members of a label.
- [`epx labels update`](#epx-labels-update)\
  Replace a label.

## epx labels create

Create a label.

Creates a label from the JSON in `--file`, such as
`{"name": "Free shipping", "description": "<p>Ships free</p>", "overlay": {"isTrue": true, "position": 1}}`,
and prints the label that Emporix answers with. Emporix generates its id unless the JSON names one;
an id that exists already is a conflict.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the label.label_manage scope.

**Command:** `epx labels create [flags]`

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

## epx labels delete

Delete a label.

Deletes the label ID.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the label.label_delete scope.

**Command:** `epx labels delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx labels get

Show one label.

Reads labels, the badges shown on products. OVERLAY is the label's overlay flag and
POSITION its overlay position, as sent; MEDIA is the id of its image (image URLs are only
in `-o json`), and DESCRIPTION is often HTML.

No OAuth scope is needed.

**Command:** `epx labels get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, OVERLAY, POSITION, MEDIA, DESCRIPTION.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx labels list

List labels.

Reads labels, the badges shown on products. OVERLAY is the label's overlay flag and
POSITION its overlay position, as sent; MEDIA is the id of its image (image URLs are only
in `-o json`), and DESCRIPTION is often HTML.

No OAuth scope is needed.

**Command:** `epx labels list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, OVERLAY, POSITION, MEDIA, DESCRIPTION.
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

## epx labels patch

Change members of a label.

Changes the members of the label ID that the JSON in `--file` names, which may be only
overlay; the others, such as name and description, keep their values. Unlike update,
Emporix does not check a metadata.version in it. An unknown ID is not found. It prints the label that Emporix answers with.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the label.label_manage scope.

**Command:** `epx labels patch ID [flags]`

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

## epx labels update

Replace a label.

Replaces the label ID with the JSON in `--file`, which must name its name and description.
An unknown ID is not found. It prints the label that Emporix answers with.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the label.label_manage scope.

**Command:** `epx labels update ID [flags]`

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

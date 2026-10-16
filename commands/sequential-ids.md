# epx sequential-ids

> Read, create, activate and delete sequential ID schemas; issue IDs.

## epx sequential-ids

Reads, creates, activates and deletes the schemas that number orders, invoices, quotes,
customers and similar records, and issues IDs from them.

Its reads require the sequentialid.schema_view scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx sequential-ids [flags]`, `epx sequential-ids [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx sequential-ids activate`](#epx-sequential-ids-activate)\
  Activate a sequential ID schema.
- [`epx sequential-ids create`](#epx-sequential-ids-create)\
  Create a sequential ID schema.
- [`epx sequential-ids delete`](#epx-sequential-ids-delete)\
  Delete a sequential ID schema.
- [`epx sequential-ids for-type`](#epx-sequential-ids-for-type)\
  List the sequential ID schemas of a type.
- [`epx sequential-ids get`](#epx-sequential-ids-get)\
  Show one sequential ID schema.
- [`epx sequential-ids list`](#epx-sequential-ids-list)\
  List sequential ID schemas.
- [`epx sequential-ids next-id`](#epx-sequential-ids-next-id)\
  Issue the next ID of a sequential ID schema type.
- [`epx sequential-ids next-ids`](#epx-sequential-ids-next-ids)\
  Issue IDs of several sequential ID schema types.

## epx sequential-ids activate

Activate a sequential ID schema.

Makes the sequence schema ID the active one, which numbers the records of its type.
ID is the ID that [`epx sequential-ids list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids#epx-sequential-ids-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the sequentialid.schema_manage scope.

**Command:** `epx sequential-ids activate ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sequential-ids create

Create a sequential ID schema.

Creates the sequence schema in `--file`, a JSON object with its name, schemaType (such as
orderNoSequence or invoiceNoSequence; letters and digits only), startValue, maxValue and
numberOfDigits, and optionally preText and postText, the text around the number, and
placeholders: each placeholder in that text, named with __ at both ends such as __year__,
with required and default, or sitePath, where in the data of a site its value is.
[`epx sequential-ids activate ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids#epx-sequential-ids-activate) makes the schema the one that numbers its type.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Emporix answers with the new schema's ID as bare text, which epx prints as a JSON string.

Requires the sequentialid.schema_manage scope.

**Command:** `epx sequential-ids create [flags]`

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

**Examples**

```sh
epx sequential-ids create --file order-numbers.json --dry-run
```

## epx sequential-ids delete

Delete a sequential ID schema.

Deletes the sequence schema ID. ID is the ID that [`epx sequential-ids list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids#epx-sequential-ids-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the sequentialid.schema_manage scope.

**Command:** `epx sequential-ids delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sequential-ids for-type

List the sequential ID schemas of a type.

Reads the schemas of one TYPE, such as orderNoSequence, invoiceNoSequence or
customerNoSequence ([`epx sequential-ids list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids#epx-sequential-ids-list) shows the tenant's types); an unknown type
lists nothing.
TYPE is what a schema numbers, PREFIX and SUFFIX the text around the number
(placeholders such as __year__ are filled in when an ID is issued), DIGITS the number's
length, COUNTER the schema's current counter (it starts at the schema's start value) and
MAX the highest number it can reach.

Requires the sequentialid.schema_view scope.

**Command:** `epx sequential-ids for-type TYPE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, ACTIVE, PREFIX, DIGITS, SUFFIX, COUNTER, MAX.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sequential-ids get

Show one sequential ID schema.

Reads the schemas that number orders, invoices, quotes, customers and similar records:
TYPE is what a schema numbers, PREFIX and SUFFIX the text around the number
(placeholders such as __year__ are filled in when an ID is issued), DIGITS the number's
length, COUNTER the schema's current counter (it starts at the schema's start value) and
MAX the highest number it can reach.
The service returns the schemas whole, so they are read with one request. An unknown
ID is not found (exit 4); Emporix answers an ID that is not 24 hexadecimal digits with an
internal error (500 Internal Server Error, exit 1).

Requires the sequentialid.schema_view scope.

**Command:** `epx sequential-ids get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, ACTIVE, PREFIX, DIGITS, SUFFIX, COUNTER, MAX.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sequential-ids list

List sequential ID schemas.

Reads the schemas that number orders, invoices, quotes, customers and similar records:
TYPE is what a schema numbers, PREFIX and SUFFIX the text around the number
(placeholders such as __year__ are filled in when an ID is issued), DIGITS the number's
length, COUNTER the schema's current counter (it starts at the schema's start value) and
MAX the highest number it can reach.
The service returns the schemas whole, so they are read with one request. An unknown
ID is not found (exit 4); Emporix answers an ID that is not 24 hexadecimal digits with an
internal error (500 Internal Server Error, exit 1).

Requires the sequentialid.schema_view scope.

**Command:** `epx sequential-ids list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, ACTIVE, PREFIX, DIGITS, SUFFIX, COUNTER, MAX.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx sequential-ids next-id

Issue the next ID of a sequential ID schema type.

Issues the next ID of the active schema of TYPE, such as orderNoSequence ([`epx sequential-ids list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids#epx-sequential-ids-list) shows the types), and prints it. `--sequence-key` numbers in a sequence
of its own, such as one per month, instead of the default one; `--file` gives the body
whole, with the values of placeholders, e.g. `{"sequenceKey": "2026-10", "placeholders": {"__year__": "2026"}}`. With `--site`, Emporix fills in the placeholders
that the body leaves out from that site's data.
Each ID issued advances the schema's counter.
It changes data although Emporix needs no write-capable scope for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the sequentialid.schema_view scope.

**Command:** `epx sequential-ids next-id TYPE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--sequence-key <KEY>`\
  The KEY of the sequence to number in, such as 2026-10.
- `--site <CODE>`\
  The site CODE whose data fills in the placeholders that the body leaves out.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx sequential-ids next-id orderNoSequence --sequence-key 2026-10 --dry-run
```

## epx sequential-ids next-ids

Issue IDs of several sequential ID schema types.

Issues IDs of the active schemas of several types in one request and prints them, by
type. `--file` gives a JSON object with a member per type, such as orderNoSequence, each
with numberOfIds (how many), and optionally sequenceKey and placeholders, as next-id
takes them. With `--site`, Emporix fills in the placeholders that the body leaves out from
that site's data.
Each ID issued advances the schema's counter.
It changes data although Emporix needs no write-capable scope for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the sequentialid.schema_view scope.

**Command:** `epx sequential-ids next-ids [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--site <CODE>`\
  The site CODE whose data fills in the placeholders that the body leaves out.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx sequential-ids next-ids --file ids.json --site AcmeStore_NL --dry-run
```

# epx indexing

> Read and configure search index providers; reindex.

## epx indexing

Reads and configures the search index providers (e.g. Algolia) and starts and follows
reindex jobs.

Its reads require different OAuth scopes; the help of each names them.
Its writes require the indexing.search_manage scope.

**Command:** `epx indexing [flags]`, `epx indexing [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx indexing providers`](#epx-indexing-providers)\
  Read, create, update and delete index providers.
- [`epx indexing reindex`](#epx-indexing-reindex)\
  Reindex every product (deprecated by Emporix).
- [`epx indexing reindex-jobs`](#epx-indexing-reindex-jobs)\
  Read and start reindex jobs.

## epx indexing providers

Read, create, update and delete index providers.

Reads index provider configurations. Search and write keys are only in `-o json`;
[`epx indexing providers public`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-public) shows what a storefront sees and needs no scope.

Its reads require different OAuth scopes; the help of each names them.
Its writes require the indexing.search_manage scope.

**Command:** `epx indexing providers [flags]`, `epx indexing providers [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx indexing providers create`](#epx-indexing-providers-create)\
  Create an index provider configuration.
- [`epx indexing providers delete`](#epx-indexing-providers-delete)\
  Delete an index provider configuration.
- [`epx indexing providers get`](#epx-indexing-providers-get)\
  Show one provider.
- [`epx indexing providers list`](#epx-indexing-providers-list)\
  List index providers.
- [`epx indexing providers public`](#epx-indexing-providers-public)\
  Read index providers as a storefront sees them.
- [`epx indexing providers update`](#epx-indexing-providers-update)\
  Replace an index provider configuration.

## epx indexing providers create

Create an index provider configuration.

Creates the configuration of an index provider, whose ID is the provider's name.
The JSON in `--file` holds provider (ALGOLIA or BATTERY_INCLUDED, required), active,
applicationId, indexName, searchKey, writeKey and, for BATTERY_INCLUDED, excludedMixinKeys
or includedMixinPaths. Only one provider can be active at a time. Emporix checks the keys
of BATTERY_INCLUDED before it stores them: it refuses invalid ones (400 Bad Request, exit 1)
and stores nothing when it cannot check them (502 Bad Gateway).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the indexing.search_manage scope.

**Command:** `epx indexing providers create [flags]`

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
epx indexing providers create --file algolia.json --dry-run
```

## epx indexing providers delete

Delete an index provider configuration.

Deletes the configuration of PROVIDER, as [`epx indexing providers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-list) shows it.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type PROVIDER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the indexing.search_manage scope.

**Command:** `epx indexing providers delete PROVIDER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing providers get

Show one provider.

Reads index provider configurations. Search and write keys are only in `-o json`;
[`epx indexing providers public`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-public) shows what a storefront sees and needs no scope.

Requires the indexing.search_view scope.

**Command:** `epx indexing providers get PROVIDER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PROVIDER, ACTIVE, APPLICATION ID, INDEX.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing providers list

List index providers.

Reads index provider configurations. Search and write keys are only in `-o json`;
[`epx indexing providers public`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-public) shows what a storefront sees and needs no scope.

Requires the indexing.search_view scope.

**Command:** `epx indexing providers list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PROVIDER, ACTIVE, APPLICATION ID, INDEX.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing providers public

Read index providers as a storefront sees them.

Reads the public index provider configurations, as a storefront sees them. The search key
is only in `-o json`.

No OAuth scope is needed.

**Command:** `epx indexing providers public [flags]`, `epx indexing providers public [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx indexing providers public get`](#epx-indexing-providers-public-get)\
  Show one public provider.
- [`epx indexing providers public list`](#epx-indexing-providers-public-list)\
  List index providers as a storefront sees them.

## epx indexing providers public get

Show one public provider.

Reads the public index provider configurations, as a storefront sees them. The search key
is only in `-o json`.

No OAuth scope is needed.

**Command:** `epx indexing providers public get PROVIDER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PROVIDER, ACTIVE, APPLICATION ID, INDEX.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing providers public list

List index providers as a storefront sees them.

Reads the public index provider configurations, as a storefront sees them. The search key
is only in `-o json`.

No OAuth scope is needed.

**Command:** `epx indexing providers public list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PROVIDER, ACTIVE, APPLICATION ID, INDEX.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing providers update

Replace an index provider configuration.

Replaces the configuration of PROVIDER, as [`epx indexing providers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-list) shows it. The
reference documents not found (exit 4) for an unknown one, as update creates none; for
BATTERY_INCLUDED Emporix checks the write key first, so an invalid configuration is answered
400 before that (the not-found answer was not checked with a valid key).
The JSON in `--file` holds provider (ALGOLIA or BATTERY_INCLUDED, required), active,
applicationId, indexName, searchKey, writeKey and, for BATTERY_INCLUDED, excludedMixinKeys
or includedMixinPaths. Only one provider can be active at a time. Emporix checks the keys
of BATTERY_INCLUDED before it stores them: it refuses invalid ones (400 Bad Request, exit 1)
and stores nothing when it cannot check them (502 Bad Gateway).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the indexing.search_manage scope.

**Command:** `epx indexing providers update PROVIDER [flags]`

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
epx indexing providers update ALGOLIA --file algolia.json
```

## epx indexing reindex

Reindex every product (deprecated by Emporix).

Reindexes every product of the tenant, sending the mode FULL, the only one there is, unless
`--file` gives the body. Emporix has deprecated this endpoint: [`epx indexing reindex-jobs create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-create)
starts a reindex job instead, which its `--wait` follows. For BATTERY_INCLUDED, Emporix checks
the stored keys first (400 Bad Request when they are invalid, 502 when it cannot check them).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the indexing.search_manage scope.

**Command:** `epx indexing reindex [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx indexing reindex --dry-run
```

## epx indexing reindex-jobs

Read and start reindex jobs.

Reads reindex jobs and their progress. Emporix fails when asked for the number of jobs,
so list cannot check it: a job that moves between pages while they are read can be shown
twice or missed.

Its reads require the indexing.search_view scope.
Its writes require the indexing.search_manage scope.

**Command:** `epx indexing reindex-jobs [flags]`, `epx indexing reindex-jobs [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx indexing reindex-jobs create`](#epx-indexing-reindex-jobs-create)\
  Start a reindex job.
- [`epx indexing reindex-jobs get`](#epx-indexing-reindex-jobs-get)\
  Show one reindex job.
- [`epx indexing reindex-jobs list`](#epx-indexing-reindex-jobs-list)\
  List reindex jobs.

## epx indexing reindex-jobs create

Start a reindex job.

Starts a job that indexes every entity of a type again: PRODUCT, ORDER or a custom entity
type ([`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list)); with `--rag`, it also embeds products for RAG search.
Emporix answers with the job, or with the job of that type that is IN_PROGRESS already.
A job is PENDING while others take its turn, PUBLISHED once every entity was sent to the
index and SUCCESS once the index took them all. For BATTERY_INCLUDED, Emporix checks the
stored keys before it starts a PRODUCT job (400 Bad Request when they are invalid, 502 when
it cannot check them).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx indexing reindex-jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-get)
does and with the indexing.search_view scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended SUCCESS, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

Requires the indexing.search_manage scope.

**Command:** `epx indexing reindex-jobs create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--entity-type <TYPE>`\
  The TYPE of the entities to reindex: PRODUCT, ORDER or a custom entity type (required without `--file`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the reindex job, 1s or more (only with `--wait`). Default: `5s`.
- `--rag`\
  Embed products for RAG search too (PRODUCT only).
- `--wait`\
  Wait until the reindex job ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx indexing reindex-jobs create --entity-type PRODUCT --rag --wait
epx indexing reindex-jobs create --file reindex.json --yes
```

## epx indexing reindex-jobs get

Show one reindex job.

Reads reindex jobs and their progress. Emporix fails when asked for the number of jobs,
so list cannot check it: a job that moves between pages while they are read can be shown
twice or missed.

Requires the indexing.search_view scope.

**Command:** `epx indexing reindex-jobs get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ENTITY, STATUS, PROCESSED, EXPECTED, CREATED, MODIFIED, MESSAGE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx indexing reindex-jobs list

List reindex jobs.

Reads reindex jobs and their progress. Emporix fails when asked for the number of jobs,
so list cannot check it: a job that moves between pages while they are read can be shown
twice or missed.

Requires the indexing.search_view scope.

**Command:** `epx indexing reindex-jobs list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ENTITY, STATUS, PROCESSED, EXPECTED, CREATED, MODIFIED, MESSAGE.
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

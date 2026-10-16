# epx search

> Search documents; read and change search indexes and saved searches.

## epx search

Reads the Search Service, which searches the instances of custom entity types (a preview
at Emporix): the documents a search finds, its indexes and saved searches, kept per type
([`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list) shows the types), and the jobs that create, update and delete
the indexes; changes the indexes and saved searches.

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the search.search_manage scope.

**Command:** `epx search [flags]`, `epx search [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx search documents`](#epx-search-documents)\
  Search the documents of a custom entity type.
- [`epx search indexes`](#epx-search-indexes)\
  Read, upsert, delete, export and import search indexes.
- [`epx search jobs`](#epx-search-jobs)\
  Read index jobs.
- [`epx search queries`](#epx-search-queries)\
  Read, upsert and delete saved searches.

## epx search documents

Search the documents of a custom entity type.

Reads the documents of a custom entity type, its instances as a search index holds them,
that the saved search `--saved-search` finds with the text of `--text` ([`epx search queries for-type TYPE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-for-type) lists them). SCORE is how well a document matches; its other fields are
only in `-o json`. `--body` gives a search of an index by its own queries and filters
instead.
TYPE is a custom entity type, as [`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list) shows its ID; an unknown one has no scope that Emporix grants (exit code 3), and a type created in the last minutes has none yet.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the search.search_read, custom.&lt;type\>_read scopes.
&lt;type\> is the custom entity type in lower case (custom.document_read for DOCUMENT).

**Command:** `epx search documents TYPE [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, SCORE.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--saved-search <ID>`\
  The ID of the saved search to run (required without `--body`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).
- `--text <text>`\
  The text to search for in every term of the saved search (required without `--body`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search indexes

Read, upsert, delete, export and import search indexes.

Reads search indexes: which fields of a custom entity type's instances can be searched.
STATUS is building, ready or failed and FIELDS the indexed fields (`-o json` shows how each is
searched). `--query` matches STATUS in upper case (`status:READY`; `status:ready` matches no index).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the search.search_manage scope.

**Command:** `epx search indexes [flags]`, `epx search indexes [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx search indexes delete`](#epx-search-indexes-delete)\
  Delete a search index.
- [`epx search indexes export`](#epx-search-indexes-export)\
  Export search indexes.
- [`epx search indexes for-type`](#epx-search-indexes-for-type)\
  List the search indexes of a custom entity type.
- [`epx search indexes get`](#epx-search-indexes-get)\
  Show one search index.
- [`epx search indexes import`](#epx-search-indexes-import)\
  Import search indexes.
- [`epx search indexes list`](#epx-search-indexes-list)\
  List search indexes.
- [`epx search indexes update`](#epx-search-indexes-update)\
  Upsert a search index.

## epx search indexes delete

Delete a search index.

Starts the job that deletes the search index ID of the custom entity type TYPE: Emporix
answers with the job's id ([`epx search jobs get ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-jobs-get) reads the job), and with 409 Conflict
(exit 5) while another job of the index is running.
TYPE and ID are a search index's custom entity type and ID, as [`epx search indexes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-list)
shows them.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx search jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-jobs-get)
does and with the search.search_read scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended success, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

As Emporix answers with the id of the job alone, `--wait` reads the job by that id first.
Where Emporix answers with none, having started no job, `--wait` prints the response and says
so (exit 0); where no read of the job succeeds, it prints the response.

Requires the search.search_manage scope.

**Command:** `epx search indexes delete TYPE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the index job, 1s or more (only with `--wait`). Default: `5s`.
- `--wait`\
  Wait until the index job ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search indexes export

Export search indexes.

Reads an export package of the search indexes that `--index` names, as TYPE/ID (the TYPE and
ID that [`epx search indexes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-list) shows): their configuration, without status or metadata.
Emporix answers an unknown index as not found (exit 4).
Emporix refuses an empty selection (`--body` with \[\], exit 1).

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
DATA is the package, base64-encoded and cut in the table; `-o json` prints the package
whole, as an import on another tenant takes it.

Requires the search.search_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx search indexes export [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting search.search_manage, which can also change data (required).
- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among DATA, EXPORTED AT.
- `--index <TYPE/ID>`\
  A search index to export, as TYPE/ID (repeatable; required without `--body`).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search indexes for-type

List the search indexes of a custom entity type.

Reads search indexes: which fields of a custom entity type's instances can be searched.
STATUS is building, ready or failed and FIELDS the indexed fields (`-o json` shows how each is
searched). `--query` matches STATUS in upper case (`status:READY`; `status:ready` matches no index).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).
TYPE is the custom entity type of search indexes, as [`epx search indexes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-list) shows it; an
unknown type lists nothing.

Requires the search.search_read scope.

**Command:** `epx search indexes for-type TYPE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, STATUS, FIELDS, VERSION.
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

## epx search indexes get

Show one search index.

Reads search indexes: which fields of a custom entity type's instances can be searched.
STATUS is building, ready or failed and FIELDS the indexed fields (`-o json` shows how each is
searched). `--query` matches STATUS in upper case (`status:READY`; `status:ready` matches no index).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).
TYPE and ID are a search index's custom entity type and ID, as [`epx search indexes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-list)
shows them; an unknown one is not found (exit 4).

Requires the search.search_read scope.

**Command:** `epx search indexes get TYPE ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, STATUS, FIELDS, VERSION.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search indexes import

Import search indexes.

Creates or updates each search index of the export package in `--file`, as update does. The
package is the whole response of [`epx search indexes export -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-export), with its exportedAt
(Emporix answers 400 without it). Emporix answers with a
result per index, in the order of the package, with the id of the job that builds it
where one started. When an index fails, those before it stay created or updated, and
Emporix answers with the error of the one that failed (409 Conflict, exit 5, while a job
of it is running or when its metadata.version is stale).

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx search jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-jobs-get)
does and with the search.search_read scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended success, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

As Emporix answers with the id of each job alone, in a result per item that names its
type/id, `--wait` reads each job by its id first and follows the jobs in turn, all within
`--wait-timeout`, naming each by that type/id. It prints the jobs as last read in place of
the results, as a JSON array in their order, leaving out each job never read (standard
error says why), and exits 0 only if each ended success; where jobs failed with different
exit codes, 1. Where Emporix started no job, it prints \[\] and says so (exit 0).

Requires the search.search_manage scope.

**Command:** `epx search indexes import [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the index job, 1s or more (only with `--wait`). Default: `5s`.
- `--wait`\
  Wait until the index job ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx search indexes import --file indexes.json --dry-run
epx search indexes import --file indexes.json --wait --yes
```

## epx search indexes list

List search indexes.

Reads search indexes: which fields of a custom entity type's instances can be searched.
STATUS is building, ready or failed and FIELDS the indexed fields (`-o json` shows how each is
searched). `--query` matches STATUS in upper case (`status:READY`; `status:ready` matches no index).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Requires the search.search_read scope.

**Command:** `epx search indexes list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, STATUS, FIELDS, VERSION.
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

## epx search indexes update

Upsert a search index.

Creates the search index ID of the custom entity type TYPE, or replaces it (an upsert),
with the one in `--file`, a JSON object with its fields (each with its path and how it is
searched) and optionally name and description. A job builds a new index, and one whose
fields changed or that is building or failed: Emporix then answers 202 Accepted with the
job's id ([`epx search jobs get ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-jobs-get) reads the job), else 204 No Content. With
metadata.version, as [`epx search indexes get TYPE ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-get) prints it, Emporix replaces
only that version; another, or a job of the index that is still running, is a conflict
(exit 5).
TYPE and ID are a search index's custom entity type and ID, as [`epx search indexes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-indexes-list)
shows them, or a new ID.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx search jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-jobs-get)
does and with the search.search_read scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended success, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

As Emporix answers with the id of the job alone, `--wait` reads the job by that id first.
Where Emporix answers with none, having started no job, `--wait` prints the response and says
so (exit 0); where no read of the job succeeds, it prints the response.

Requires the search.search_manage scope.

**Command:** `epx search indexes update TYPE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the index job, 1s or more (only with `--wait`). Default: `5s`.
- `--wait`\
  Wait until the index job ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx search indexes update vehicle vehicles --file index.json --dry-run
epx search indexes update vehicle vehicles --file index.json --wait --yes
```

## epx search jobs

Read index jobs.

Reads the jobs that create, update and delete search indexes: TYPE is create_index,
update_index or delete_index, STATUS in_progress, success or failure as the table and JSON
show them, and RESULT what the job recorded when it finished. `--query` matches the upper-case
form (`status:SUCCESS` and `type:UPDATE_INDEX` are checked; a lower-case filter, such as
`status:success`, matches no job).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Requires the search.search_read scope.

**Command:** `epx search jobs [flags]`, `epx search jobs [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx search jobs get`](#epx-search-jobs-get)\
  Show one index job.
- [`epx search jobs list`](#epx-search-jobs-list)\
  List index jobs.

## epx search jobs get

Show one index job.

Reads the jobs that create, update and delete search indexes: TYPE is create_index,
update_index or delete_index, STATUS in_progress, success or failure as the table and JSON
show them, and RESULT what the job recorded when it finished. `--query` matches the upper-case
form (`status:SUCCESS` and `type:UPDATE_INDEX` are checked; a lower-case filter, such as
`status:success`, matches no job).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Requires the search.search_read scope.

**Command:** `epx search jobs get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, INDEX, CREATED, MODIFIED, RESULT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search jobs list

List index jobs.

Reads the jobs that create, update and delete search indexes: TYPE is create_index,
update_index or delete_index, STATUS in_progress, success or failure as the table and JSON
show them, and RESULT what the job recorded when it finished. `--query` matches the upper-case
form (`status:SUCCESS` and `type:UPDATE_INDEX` are checked; a lower-case filter, such as
`status:success`, matches no job).
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Requires the search.search_read scope.

**Command:** `epx search jobs list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, INDEX, CREATED, MODIFIED, RESULT.
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

## epx search queries

Read, upsert and delete saved searches.

Reads saved searches: queries and filters (only in `-o json`) stored to run against the
search INDEX of a custom entity type.
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Its reads require the search.search_read scope.
Its writes require the search.search_manage scope.

**Command:** `epx search queries [flags]`, `epx search queries [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx search queries delete`](#epx-search-queries-delete)\
  Delete a saved search.
- [`epx search queries for-type`](#epx-search-queries-for-type)\
  List the saved searches of a custom entity type.
- [`epx search queries get`](#epx-search-queries-get)\
  Show one saved search.
- [`epx search queries list`](#epx-search-queries-list)\
  List saved searches.
- [`epx search queries update`](#epx-search-queries-update)\
  Upsert a saved search.

## epx search queries delete

Delete a saved search.

Deletes the saved search ID of the custom entity type TYPE; an unknown one is not found
(exit 4).
TYPE and ID are a saved search's custom entity type and ID, as [`epx search queries list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-list)
shows them.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the search.search_manage scope.

**Command:** `epx search queries delete TYPE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search queries for-type

List the saved searches of a custom entity type.

Reads saved searches: queries and filters (only in `-o json`) stored to run against the
search INDEX of a custom entity type.
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).
TYPE is the custom entity type of saved searches, as [`epx search queries list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-list) shows it; an
unknown type lists nothing.

Requires the search.search_read scope.

**Command:** `epx search queries for-type TYPE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, INDEX, VERSION.
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

## epx search queries get

Show one saved search.

Reads saved searches: queries and filters (only in `-o json`) stored to run against the
search INDEX of a custom entity type.
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).
TYPE and ID are a saved search's custom entity type and ID, as [`epx search queries list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-list)
shows them; an unknown one is not found (exit 4).

Requires the search.search_read scope.

**Command:** `epx search queries get TYPE ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, INDEX, VERSION.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx search queries list

List saved searches.

Reads saved searches: queries and filters (only in `-o json`) stored to run against the
search INDEX of a custom entity type.
Emporix answers a list's malformed `--query` with an internal error (500 Internal Server
Error, exit 1).

Requires the search.search_read scope.

**Command:** `epx search queries list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, INDEX, VERSION.
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

## epx search queries update

Upsert a saved search.

Creates the saved search ID of the custom entity type TYPE, or replaces it (an upsert),
with the one in `--file`, a JSON object with index (the ID of a search index of TYPE),
queries, filters or both, and optionally name and description. Emporix answers a new
one with its id (201 Created), and a replaced one with 204 No Content. Replacing one
needs its metadata.version, as [`epx search queries get TYPE ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-get) prints it; a
stale one is a conflict (exit 5). An unknown index is not found (exit 4).
TYPE and ID are a saved search's custom entity type and ID, as [`epx search queries list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search#epx-search-queries-list)
shows them, or a new ID.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the search.search_manage scope.

**Command:** `epx search queries update TYPE ID [flags]`

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
epx search queries update vehicle by-make --file by-make.json --dry-run
```

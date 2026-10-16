# epx imports

> Read import configurations, runs and imported records; start and schedule runs.

## epx imports

Reads the Import Service, which imports master data from external sources into Emporix:
configurations, their streams (each imports one source entity into an Emporix type),
schedules and runs, the errors and diagnostics of a run, the records imported, statistics,
job groups, health thresholds and the tenant's import limits. Emporix allows these reads
only with importtool.import\_trigger, with which epx also starts, retries, cancels and schedules runs.

Its reads require the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the importtool.import\_trigger scope.

**Command:** `epx imports [flags]`, `epx imports [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx imports configurations`](#epx-imports-configurations)\
  Read import configurations; start and schedule their runs.
- [`epx imports health-thresholds`](#epx-imports-health-thresholds)\
  Show the thresholds that rate the health of import streams.
- [`epx imports job-groups`](#epx-imports-job-groups)\
  List the job groups that scope the import statistics.
- [`epx imports license`](#epx-imports-license)\
  Show the tenant's import limits.
- [`epx imports records`](#epx-imports-records)\
  List the imported records of a type.
- [`epx imports runs`](#epx-imports-runs)\
  Read, cancel and retry import runs.
- [`epx imports statistics`](#epx-imports-statistics)\
  Show the import statistics of the last 30 days.
- [`epx imports streams`](#epx-imports-streams)\
  Read import streams and the records they imported.
- [`epx imports types`](#epx-imports-types)\
  List the types that hold imported records.

## epx imports configurations

Read import configurations; start and schedule their runs.

Reads import configurations, each a set of streams that run together: ENABLED says
whether it is active, VERSION counts its changes and MODIFIED is when it last changed. Its
description, source connection, creator and health thresholds are only in `-o json`. The
service returns the configurations whole, so they are read with one request.

Its reads require the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the importtool.import\_trigger scope.

**Command:** `epx imports configurations [flags]`, `epx imports configurations [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx imports configurations delete-schedule`](#epx-imports-configurations-delete-schedule)\
  Delete the run schedule of a configuration.
- [`epx imports configurations get`](#epx-imports-configurations-get)\
  Show one configuration.
- [`epx imports configurations list`](#epx-imports-configurations-list)\
  List import configurations.
- [`epx imports configurations runs`](#epx-imports-configurations-runs)\
  List the runs of a configuration.
- [`epx imports configurations schedule`](#epx-imports-configurations-schedule)\
  Show the run schedule of a configuration.
- [`epx imports configurations stream-order`](#epx-imports-configurations-stream-order)\
  Show the order the streams of a configuration run in.
- [`epx imports configurations streams`](#epx-imports-configurations-streams)\
  List the streams of a configuration.
- [`epx imports configurations trigger`](#epx-imports-configurations-trigger)\
  Start an import run of a configuration.
- [`epx imports configurations update-schedule`](#epx-imports-configurations-update-schedule)\
  Upsert the run schedule of a configuration.

## epx imports configurations delete-schedule

Delete the run schedule of a configuration.

Deletes the schedule of the configuration CONFIGURATION\_ID, which then runs only when it
is triggered. Emporix answers alike when there is no schedule, also of a configuration that
was deleted.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CONFIGURATION\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the importtool.import\_trigger scope.

**Command:** `epx imports configurations delete-schedule CONFIGURATION_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations get

Show one configuration.

Reads import configurations, each a set of streams that run together: ENABLED says
whether it is active, VERSION counts its changes and MODIFIED is when it last changed. Its
description, source connection, creator and health thresholds are only in `-o json`. The
service returns the configurations whole, so they are read with one request.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations get ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, ENABLED, VERSION, MODIFIED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations list

List import configurations.

Reads import configurations, each a set of streams that run together: ENABLED says
whether it is active, VERSION counts its changes and MODIFIED is when it last changed. Its
description, source connection, creator and health thresholds are only in `-o json`. The
service returns the configurations whole, so they are read with one request.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations list [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, ENABLED, VERSION, MODIFIED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations runs

List the runs of a configuration.

Reads the run history of an import configuration, most recent first.
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows; an unknown one lists nothing.
STATUS is QUEUED, RUNNING, RESUMABLE (interrupted; it resumes by itself), SUCCEEDED,
PARTIAL (records failed or a limit skipped some, or a stream did not run), FAILED, CANCELLED
or ABORTED (refused: nothing was read), MODE FULL or DELTA, followed by dry run for a run that
wrote nothing, STARTED when the run was requested, FINISHED when it ended, READ the records it
read and FAILED those that failed. `-o json` adds its other counters, message and queue position.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations runs CONFIGURATION_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, STATUS, MODE, STARTED, FINISHED, READ, FAILED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations schedule

Show the run schedule of a configuration.

Reads when an import configuration runs by itself: CRON is a cron expression with six fields
(seconds first), evaluated in TIME ZONE, and NEXT RUN when it fires next. A configuration
without a schedule is not found (exit 4).
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows; an unknown one is not found.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations schedule CONFIGURATION_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CRON, TIME ZONE, ENABLED, NEXT RUN.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations stream-order

Show the order the streams of a configuration run in.

Reads the order in which the streams of an import configuration run, which the service
derives from their dependencies (also those their mappings add) and sequence: each STREAM
runs after the streams that AFTER names.
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows; an unknown one lists nothing.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations stream-order CONFIGURATION_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among STREAM, AFTER.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations streams

List the streams of a configuration.

Reads the streams of an import configuration, in their sequence, with the columns of
[`epx imports streams get`](/emporix-cli/commands/imports#epx-imports-streams-get). The service returns them whole, so they are read with one request.
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows; an unknown one lists nothing.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports configurations streams CONFIGURATION_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, SOURCE, TARGET TYPE, MODE, ENABLED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports configurations trigger

Start an import run of a configuration.

Starts an import run of the configuration CONFIGURATION\_ID, a DELTA run unless `--mode` FULL,
and prints the run, QUEUED until a worker of the Import Service takes it. `--file` gives the
body instead, e.g. `{"mode": "FULL", "dryRun": true, "streamIds": ["ID"]}` for a dry run of
some streams, which maps and checks records but writes none. A configuration that has a run
QUEUED, RUNNING or RESUMABLE already is a conflict (exit 5). Emporix answers a run that it
refuses, e.g. of streams whose mappings were never published, as any other: the run ends
ABORTED, which `--wait` reports.
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx imports runs get`](/emporix-cli/commands/imports#epx-imports-runs-get)
does and with the importtool.import\_trigger scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended SUCCEEDED, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

Requires the importtool.import\_trigger scope.

**Command:** `epx imports configurations trigger CONFIGURATION_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--mode <MODE>`\
  The MODE of the run, DELTA when not given: FULL or DELTA.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the run, 1s or more (only with `--wait`). Default: `5s`.
- `--wait`\
  Wait until the run ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx imports configurations trigger 7c9e6679-7425-40de-944b-e07fc1f90ae7 --mode FULL --wait
```

## epx imports configurations update-schedule

Upsert the run schedule of a configuration.

Sets when the configuration CONFIGURATION\_ID runs by itself, creating its schedule where
it has none (an upsert), and prints the schedule. `--cron` is a Spring cron expression with
six fields, seconds first, e.g. `"0 0 2 * * *"` for 02:00 every day (Emporix refuses one with
five: 400 Bad Request), evaluated in `--timezone`, an IANA time zone such as Europe/Berlin.
The schedule runs the configuration unless `--enabled=false`, which keeps it without running
it: epx always sends enabled, as the schedule is replaced whole. `--file` gives the body
instead.
CONFIGURATION\_ID is the ID that [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the importtool.import\_trigger scope.

**Command:** `epx imports configurations update-schedule CONFIGURATION_ID [flags]`

**Flags**

- `--cron <EXPRESSION>`\
  The Spring cron EXPRESSION of the schedule, six fields (required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--enabled`\
  Whether the schedule runs the configuration (`--enabled=false:` not). Default: `true`.
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--timezone <ZONE>`\
  The IANA time ZONE the expression is evaluated in.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx imports configurations update-schedule 7c9e6679-7425-40de-944b-e07fc1f90ae7 --cron "0 0 2 * * *" --timezone Europe/Berlin
```

## epx imports health-thresholds

Show the thresholds that rate the health of import streams.

Reads the thresholds that rate an import stream GREEN, AMBER or RED in the statistics: the
tenant's, and the built-in defaults that each one it leaves empty falls back to. A stream is
AMBER or RED once this share of its records failed (FAILURE RATIO, 0 to 1) or its newest data
is this many hours old (STALE HOURS), and RED after a failed run when FAILED RUN RED is true.
A configuration or stream may set its own (in their `-o json`).

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports health-thresholds [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among LEVEL, FAILURE RATIO AMBER, FAILURE RATIO RED, STALE HOURS AMBER, STALE HOURS RED, FAILED RUN RED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports job-groups

List the job groups that scope the import statistics.

Reads the tenant's job groups, named sets of import configurations that the import
statistics can be scoped to: CONFIGURATIONS counts them (`-o json` lists their IDs). The
service returns them whole, so they are read with one request.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports job-groups [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CONFIGURATIONS, MODIFIED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports license

Show the tenant's import limits.

Reads the tenant's import limits: a run writes at most MAX RECORDS PER TYPE records per
target type (it skips the rest and ends PARTIAL), at most MAX CONCURRENT RUNS runs execute at
once (others wait, QUEUED), and BATCH SIZE records are written per request, by WORKERS
writers in parallel. Emporix support raises a limit on request.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports license [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among MAX RECORDS PER TYPE, MAX CONCURRENT RUNS, BATCH SIZE, WORKERS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports records

List the imported records of a type.

Reads the records imported into a target type: KEY is a record's natural key in the source,
OUTCOME how it was last imported (UPSERTED, DELETED at the source, DELETED\_TARGET in
Emporix, FAILED or DRY\_RUN) and IMPORTED when. `-o json` adds its stored fields, the Emporix
id it was written under, and its configuration, stream and run. TYPE is one that
[`epx imports types`](/emporix-cli/commands/imports#epx-imports-types) shows; an unknown one lists nothing. To search keys or filter by
outcome, one page at a time: [`epx api get /importtool/{tenant}/data/records -p type=TYPE -p search=TEXT -p outcome=FAILED --scope importtool.import_trigger`](/emporix-cli/commands/api#epx-api-get).

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports records TYPE [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, KEY, TYPE, OUTCOME, IMPORTED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs

Read, cancel and retry import runs.

Reads import runs; a run's ID is the one that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows.
Its `-o json` adds the progress of each stream of the run, which [`epx imports runs streams`](/emporix-cli/commands/imports#epx-imports-runs-streams) shows.

Its reads require the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the importtool.import\_trigger scope.

**Command:** `epx imports runs [flags]`, `epx imports runs [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx imports runs cancel`](#epx-imports-runs-cancel)\
  Cancel an import run.
- [`epx imports runs diagnostics`](#epx-imports-runs-diagnostics)\
  List the source rows behind a run's repeated keys and missing parents.
- [`epx imports runs errors`](#epx-imports-runs-errors)\
  List the errors of a run.
- [`epx imports runs get`](#epx-imports-runs-get)\
  Show one run.
- [`epx imports runs retry`](#epx-imports-runs-retry)\
  Run the failed records of an import run again.
- [`epx imports runs streams`](#epx-imports-runs-streams)\
  Show the progress of each stream of a run.

## epx imports runs cancel

Cancel an import run.

Asks the import run ID to stop, and prints what Emporix answers: whether it accepted the
cancellation. A QUEUED run is CANCELLED at once; a running one stops when it can, or with
`--force` at once, which also unblocks its configuration. A RESUMABLE run stops only with
`--force`: without it, a conflict (exit 5). Emporix answers a run that is unknown or ended
already with 409 Conflict and `"accepted"`: false, which epx prints as the answer (exit 5).
ID is the one that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the importtool.import\_trigger scope.

**Command:** `epx imports runs cancel ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Stop the run at once.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs diagnostics

List the source rows behind a run's repeated keys and missing parents.

Reads the source rows behind a run's feed-quality counters, which do not fail a run: KIND is
REPEATED\_KEY for a row whose KEY the run had already imported (OCCURRENCE times before), or
UNRESOLVED\_PARENT for a child line whose parent was missing: LINKING FIELD and LINKING VALUE
are the parent key it points at, WAITING LINES the lines waiting for that parent. The service
keeps a capped sample per stream and kind; `-o json` says how many rows it recorded, its cap
and whether a stream reached it (the run's counters give the totals). Every recorded row is
read, up to 50000; a run that recorded more fails, and its rows are read per stream:
[`epx api get /importtool/{tenant}/runs/RUN_ID/diagnostics -p streamId=ID -p limit=50000 --scope importtool.import_trigger`](/emporix-cli/commands/api#epx-api-get).
RUN\_ID is the ID that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows; an unknown one lists nothing.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports runs diagnostics RUN_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KIND, STREAM, KEY, LINKING FIELD, LINKING VALUE, WAITING LINES, OCCURRENCE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs errors

List the errors of a run.

Reads the errors recorded during an import run: KEY is the natural key of the record that
failed, STAGE the step it failed in (such as MAPPING or UPSERT), CODE the error code and
MESSAGE what went wrong. Its stream and time are only in `-o json`.
RUN\_ID is the ID that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows; an unknown one lists nothing.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports runs errors RUN_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, KEY, STAGE, CODE, MESSAGE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs get

Show one run.

Reads import runs; a run's ID is the one that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows.
Its `-o json` adds the progress of each stream of the run, which [`epx imports runs streams`](/emporix-cli/commands/imports#epx-imports-runs-streams) shows.
STATUS is QUEUED, RUNNING, RESUMABLE (interrupted; it resumes by itself), SUCCEEDED,
PARTIAL (records failed or a limit skipped some, or a stream did not run), FAILED, CANCELLED
or ABORTED (refused: nothing was read), MODE FULL or DELTA, followed by dry run for a run that
wrote nothing, STARTED when the run was requested, FINISHED when it ended, READ the records it
read and FAILED those that failed. `-o json` adds its other counters, message and queue position.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports runs get ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, STATUS, MODE, STARTED, FINISHED, READ, FAILED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs retry

Run the failed records of an import run again.

Starts a run that imports again only the records that failed in the import run ID, and
prints it, QUEUED until a worker takes it; the run ID keeps its counters. A configuration
that has a run QUEUED, RUNNING or RESUMABLE already is a conflict (exit 5).
ID is the one that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

`--wait` then reads the job again every `--poll-interval`, as [`epx imports runs get`](/emporix-cli/commands/imports#epx-imports-runs-get)
does and with the importtool.import\_trigger scope too, saying on standard error how its status
changes, until it ends: then it prints the job and exits 0 if it ended SUCCEEDED, else 1.
When `--wait-timeout` passes first, it prints the job as last read and exits 1; the job goes
on. When a read of the job fails, it prints the job as last read and exits as that read
does, e.g. 4 when the job is not found.

Requires the importtool.import\_trigger scope.

**Command:** `epx imports runs retry ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--poll-interval <duration>`\
  How often `--wait` reads the run, 1s or more (only with `--wait`). Default: `5s`.
- `--wait`\
  Wait until the run ends, then print it.
- `--wait-timeout <duration>`\
  How long `--wait` waits, 1s or more (only with `--wait`). Default: `30m`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports runs streams

Show the progress of each stream of a run.

Reads the progress of each stream of an import run: STATUS is QUEUED, RUNNING, SUCCEEDED,
PARTIAL, FAILED, SKIPPED or ABORTED (not run because of its configuration, such as mappings
never published, as MESSAGE says), then the records the stream read, created, updated,
skipped as unchanged, deleted and failed to import. A queued run lists nothing yet.
RUN\_ID is the ID that [`epx imports configurations runs CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-runs) shows; an unknown one is not found.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports runs streams RUN_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among STREAM, STATUS, READ, CREATED, UPDATED, SKIPPED, DELETED, FAILED, MESSAGE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports statistics

Show the import statistics of the last 30 days.

Reads the import statistics of the last 30 days for the whole tenant: RUNS counts the runs
started, SUCCEEDED, PARTIAL, FAILED and CANCELLED how they ended, ACTIVE those queued or
running now, and RECORDS READ and RECORDS FAILED the records. `-o json` adds the rates, a daily
series, each stream's health, the failing streams, an error breakdown and the streams added
or removed. Another window or scope, e.g. one configuration ([`epx imports job-groups`](/emporix-cli/commands/imports#epx-imports-job-groups) lists
sets of them): [`epx api get /importtool/{tenant}/stats -p configId=ID -p from=2026-09-01T00:00:00Z -p granularity=WEEK --scope importtool.import_trigger`](/emporix-cli/commands/api#epx-api-get).

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports statistics [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among RUNS, SUCCEEDED, PARTIAL, FAILED, CANCELLED, ACTIVE, RECORDS READ, RECORDS FAILED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports streams

Read import streams and the records they imported.

Reads an import stream, which imports one source entity of a configuration: SOURCE is that
entity, TARGET TYPE the Emporix type it writes to (a discriminator, in `-o json`, may route
records to others), MODE STANDALONE, or COMPOSITE\_CHILD or COMPOSITE\_MERGE for a stream that
adds to the records of another. Its delta, delete, write and health settings are only in
`-o json`. A stream's ID is the one that [`epx imports configurations streams CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-streams) shows.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports streams [flags]`, `epx imports streams [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx imports streams get`](#epx-imports-streams-get)\
  Show one stream.
- [`epx imports streams records`](#epx-imports-streams-records)\
  List the records a stream imported.

## epx imports streams get

Show one stream.

Reads an import stream, which imports one source entity of a configuration: SOURCE is that
entity, TARGET TYPE the Emporix type it writes to (a discriminator, in `-o json`, may route
records to others), MODE STANDALONE, or COMPOSITE\_CHILD or COMPOSITE\_MERGE for a stream that
adds to the records of another. Its delta, delete, write and health settings are only in
`-o json`. A stream's ID is the one that [`epx imports configurations streams CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-streams) shows.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports streams get ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, SOURCE, TARGET TYPE, MODE, ENABLED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports streams records

List the records a stream imported.

Reads the records a stream imported, with the columns of [`epx imports records`](/emporix-cli/commands/imports#epx-imports-records).
STREAM\_ID is the ID that [`epx imports configurations streams CONFIGURATION_ID`](/emporix-cli/commands/imports#epx-imports-configurations-streams) shows; an unknown one lists nothing.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports streams records STREAM_ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, KEY, TYPE, OUTCOME, IMPORTED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx imports types

List the types that hold imported records.

Reads the target types that hold imported records ([`epx imports records`](/emporix-cli/commands/imports#epx-imports-records) lists those of
one). The service returns them whole, so they are read with one request.

Requires the importtool.import\_trigger scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx imports types [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting importtool.import\_trigger, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among TYPE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

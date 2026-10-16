# epx changelog

> Read the change history (audit log).

## epx changelog

Reads the tenant's change history, an audit log (personal data: who changed what, with
the old and new values in `-o json`), ordered by when the changes occurred. AT is when, TYPE
create, update or delete, ENTITY and ENTITY ID the changed document (an order, customer,
product, site, …), ACTOR who changed it (a name, system, external or unknown) and FIELDS
the changed fields. Without a lower bound on occurredAt (not only inside OR), Emporix
returns the last 30 days, a window that moves while the pages are read: on an active
tenant, read a closed range in the past, e.g. `--query`
`'occurredAt:(>"2026-09-01T00:00:00.000Z" AND <"2026-10-01T00:00:00.000Z")'`, or use
`--limit`; otherwise a change can come or go between pages, so the read fails or misses
it. One document's history: `--query` `'entity:order entityId:ID'`.

Requires the changelog.changelog\_read scope.

**Command:** `epx changelog [flags]`, `epx changelog [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx changelog list`](#epx-changelog-list)\
  List changes.

## epx changelog list

List changes.

Reads the tenant's change history, an audit log (personal data: who changed what, with
the old and new values in `-o json`), ordered by when the changes occurred. AT is when, TYPE
create, update or delete, ENTITY and ENTITY ID the changed document (an order, customer,
product, site, …), ACTOR who changed it (a name, system, external or unknown) and FIELDS
the changed fields. Without a lower bound on occurredAt (not only inside OR), Emporix
returns the last 30 days, a window that moves while the pages are read: on an active
tenant, read a closed range in the past, e.g. `--query`
`'occurredAt:(>"2026-09-01T00:00:00.000Z" AND <"2026-10-01T00:00:00.000Z")'`, or use
`--limit`; otherwise a change can come or go between pages, so the read fails or misses
it. One document's history: `--query` `'entity:order entityId:ID'`.

Requires the changelog.changelog\_read scope.

**Command:** `epx changelog list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among AT, TYPE, ENTITY, ENTITY ID, NAME, ACTOR, FIELDS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

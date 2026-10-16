# epx segments

> Read and change customer segments and what is assigned to them.

## epx segments

Reads and changes customer segments, which group customers to offer them items such as
products or categories, and what is assigned to them: customers, IAM groups and items.

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the customersegment.segment\_manage scope.

**Command:** `epx segments [flags]`, `epx segments [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx segments all-items`](#epx-segments-all-items)\
  List the items of every segment.
- [`epx segments bulk-create`](#epx-segments-bulk-create)\
  Create segments in bulk.
- [`epx segments bulk-delete`](#epx-segments-bulk-delete)\
  Delete segments in bulk.
- [`epx segments bulk-update`](#epx-segments-bulk-update)\
  Upsert segments in bulk.
- [`epx segments create`](#epx-segments-create)\
  Create a segment.
- [`epx segments customers`](#epx-segments-customers)\
  Read and change the customers of a segment.
- [`epx segments delete`](#epx-segments-delete)\
  Delete a segment.
- [`epx segments get`](#epx-segments-get)\
  Show one segment.
- [`epx segments groups`](#epx-segments-groups)\
  Read and change the IAM groups of a segment.
- [`epx segments items`](#epx-segments-items)\
  Read and change the items of a segment.
- [`epx segments list`](#epx-segments-list)\
  List segments.
- [`epx segments match`](#epx-segments-match)\
  List the products that belong to segments.
- [`epx segments patch`](#epx-segments-patch)\
  Change a segment.
- [`epx segments search`](#epx-segments-search)\
  Search segments.
- [`epx segments update`](#epx-segments-update)\
  Upsert a segment.

## epx segments all-items

List the items of every segment.

Reads the items assigned to every segment, products or categories (TYPE), with their
SEGMENT.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments all-items [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, SEGMENT, TYPE, CODE, NAME.
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

## epx segments bulk-create

Create segments in bulk.

Creates the segments in `--file`, a JSON array of segments, each as create takes it.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments bulk-create [flags]`

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

## epx segments bulk-delete

Delete segments in bulk.

Deletes the segments whose IDs `--file` gives, a JSON array such as `["s1", "s2"]`.
Emporix refuses to delete a segment that has items or customers assigned to it
(400 Bad Request, exit 1) unless `--with-assignments` deletes those assignments with it. It
deletes the IAM group assignments of a segment with it either way: a segment with only
IAM groups assigned is deleted without `--with-assignments`.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Deletes the segments that `--file` names with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type the tenant at its prompt.
`--yes` sends without asking (required without a terminal, or with `--file` -);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments bulk-delete [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--with-assignments`\
  Delete the item and customer assignments of a segment with it (its IAM group assignments go either way).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments bulk-update

Upsert segments in bulk.

Replaces the segments in `--file`, a JSON array of segments, each with its id, or creates
those that do not exist (an upsert).

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments bulk-update [flags]`

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

## epx segments create

Create a segment.

Creates the segment in `--file`, a JSON object such as [`epx segments get ID -o json`](/emporix-cli/commands/segments#epx-segments-get)
prints one, with its name, siteCode, validity and status (INACTIVE unless it says
ACTIVE); Emporix makes its id unless the body gives one. A segment with that id that
exists already is a conflict (exit 5).

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments create [flags]`

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

## epx segments customers

Read and change the customers of a segment.

Reads the customers assigned to a segment, personal data included, and assigns and removes
them, for B2B with their legal entity.

Its reads require the customersegment.segment\_read scope.
Its writes require the customersegment.segment\_manage scope.

**Command:** `epx segments customers [flags]`, `epx segments customers [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx segments customers bulk-delete`](#epx-segments-customers-bulk-delete)\
  Delete customer assignments of a segment in bulk.
- [`epx segments customers bulk-update`](#epx-segments-customers-bulk-update)\
  Upsert customer assignments of a segment in bulk.
- [`epx segments customers delete`](#epx-segments-customers-delete)\
  Remove a customer from a segment.
- [`epx segments customers delete-b2b`](#epx-segments-customers-delete-b2b)\
  Remove a customer of a legal entity from a segment.
- [`epx segments customers get`](#epx-segments-customers-get)\
  Show one customer of a segment.
- [`epx segments customers get-b2b`](#epx-segments-customers-get-b2b)\
  Show one customer of a legal entity in a segment.
- [`epx segments customers list`](#epx-segments-customers-list)\
  List the customers of a segment.
- [`epx segments customers search`](#epx-segments-customers-search)\
  Search the customers of a segment.
- [`epx segments customers update`](#epx-segments-customers-update)\
  Assign a customer to a segment.
- [`epx segments customers update-b2b`](#epx-segments-customers-update-b2b)\
  Assign a customer of a legal entity to a segment.

## epx segments customers bulk-delete

Delete customer assignments of a segment in bulk.

Removes the customers in `--file` from the segment SEGMENT\_ID, a JSON array such as
`[{"customerId": "c1", "legalEntityId": "le1"}]` (without legalEntityId for B2C).

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Deletes the customer assignments of a segment that `--file` names with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type SEGMENT\_ID at its prompt.
`--yes` sends without asking (required without a terminal, or with `--file` -);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers bulk-delete SEGMENT_ID [flags]`

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

## epx segments customers bulk-update

Upsert customer assignments of a segment in bulk.

Assigns the customers in `--file` to the segment SEGMENT\_ID, or replaces their assignments
(an upsert): a JSON array of assignments, each with its customerId, for B2B its
legalEntityId, and optionally its mixins and metadata.
A segment holds customers without a legal entity (B2C) or with one (B2B), never both:
Emporix refuses to assign one of the other kind.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers bulk-update SEGMENT_ID [flags]`

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

## epx segments customers delete

Remove a customer from a segment.

Removes the customer CUSTOMER without a legal entity (B2C) from the segment
SEGMENT\_ID; Emporix answers one that is not
assigned as removed.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, and
CUSTOMER the customer's id (the id of [`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get)), not its
NUMBER.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CUSTOMER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers delete SEGMENT_ID CUSTOMER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers delete-b2b

Remove a customer of a legal entity from a segment.

Removes the customer CUSTOMER of the legal entity LEGAL\_ENTITY (B2B) from the segment
SEGMENT\_ID; Emporix answers one that is not
assigned as removed.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows,
CUSTOMER the customer's id (the id of [`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get)), not its
NUMBER, and LEGAL\_ENTITY the ID that
[`epx legal-entities list`](/emporix-cli/commands/legal-entities#epx-legal-entities-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type LEGAL\_ENTITY at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers delete-b2b SEGMENT_ID CUSTOMER LEGAL_ENTITY [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers get

Show one customer of a segment.

Reads one customer assigned to a segment without a legal entity (B2C), personal data
included; one assigned with a legal entity (B2B) is not found (exit 4): get-b2b reads it.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one is not found.
CUSTOMER is the CUSTOMER that [`epx segments customers list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-customers-list) shows; an unknown one is not found.
CUSTOMER is the customer's id, not the NUMBER that [`epx customers get`](/emporix-cli/commands/customers#epx-customers-get) takes; ASSIGNMENT
is MANUAL or AUTOMATIC.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments customers get SEGMENT_ID CUSTOMER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, NAME, EMAIL, LEGAL ENTITY, ASSIGNMENT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers get-b2b

Show one customer of a legal entity in a segment.

Reads one customer assigned to a segment with a legal entity (B2B), personal data included.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one is not found.
CUSTOMER and LEGAL\_ENTITY are the CUSTOMER and LEGAL ENTITY that [`epx segments customers list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-customers-list) shows; an unknown one is not found.
CUSTOMER is the customer's id, not the NUMBER that [`epx customers get`](/emporix-cli/commands/customers#epx-customers-get) takes; ASSIGNMENT
is MANUAL or AUTOMATIC.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments customers get-b2b SEGMENT_ID CUSTOMER LEGAL_ENTITY [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, NAME, EMAIL, LEGAL ENTITY, ASSIGNMENT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers list

List the customers of a segment.

Reads the customers of a segment (for B2B, with their legal entity), personal data included.
list and search sort by id unless `--sort` says otherwise, an order that Emporix keeps
across pages: unsorted, it pages them in a different order per request.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.
CUSTOMER is the customer's id, not the NUMBER that [`epx customers get`](/emporix-cli/commands/customers#epx-customers-get) takes; ASSIGNMENT
is MANUAL or AUTOMATIC.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments customers list SEGMENT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, NAME, EMAIL, LEGAL ENTITY, ASSIGNMENT.
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

## epx segments customers search

Search the customers of a segment.

Reads the customers of a segment that an Emporix q filter finds (`--query`; without it,
every one), personal data included.
list and search sort by id unless `--sort` says otherwise, an order that Emporix keeps
across pages: unsorted, it pages them in a different order per request.
It reads and shows them as [`epx segments customers list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-customers-list) does.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.
CUSTOMER is the customer's id, not the NUMBER that [`epx customers get`](/emporix-cli/commands/customers#epx-customers-get) takes; ASSIGNMENT
is MANUAL or AUTOMATIC.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments customers search SEGMENT_ID [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, NAME, EMAIL, LEGAL ENTITY, ASSIGNMENT.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers update

Assign a customer to a segment.

Assigns the customer CUSTOMER without a legal entity (B2C) to the segment
SEGMENT\_ID, or replaces its assignment (an upsert).
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, and
CUSTOMER the customer's id (the id of [`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get)), not its
NUMBER.
A segment holds customers without a legal entity (B2C) or with one (B2B), never both:
Emporix refuses to assign one of the other kind.
`--file` may give the assignment's mixins and metadata (its version and the schemas of
its mixins); without it, epx sends an empty object.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers update SEGMENT_ID CUSTOMER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments customers update-b2b

Assign a customer of a legal entity to a segment.

Assigns the customer CUSTOMER of the legal entity LEGAL\_ENTITY (B2B) to the segment
SEGMENT\_ID, or replaces its assignment (an upsert).
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows,
CUSTOMER the customer's id (the id of [`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get)), not its
NUMBER, and LEGAL\_ENTITY the ID that
[`epx legal-entities list`](/emporix-cli/commands/legal-entities#epx-legal-entities-list) shows.
A segment holds customers without a legal entity (B2C) or with one (B2B), never both:
Emporix refuses to assign one of the other kind.
`--file` may give the assignment's mixins and metadata (its version and the schemas of
its mixins); without it, epx sends an empty object.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments customers update-b2b SEGMENT_ID CUSTOMER LEGAL_ENTITY [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments delete

Delete a segment.

Deletes the segment ID. ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows.
Emporix refuses to delete a segment that has items or customers assigned to it
(400 Bad Request, exit 1) unless `--with-assignments` deletes those assignments with it. It
deletes the IAM group assignments of a segment with it either way: a segment with only
IAM groups assigned is deleted without `--with-assignments`.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Deletes data, which cannot be undone.
With `--with-assignments`, it deletes the segment and its item, customer and IAM group assignments with one request.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--with-assignments`\
  Delete the item and customer assignments of a segment with it (its IAM group assignments go either way).
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments get

Show one segment.

Reads customer segments, which group customers to offer them items such as products or
categories. STATUS is ACTIVE or INACTIVE, and FROM and TO the validity window.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, STATUS, SITE, FROM, TO.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments groups

Read and change the IAM groups of a segment.

Reads the IAM groups assigned to a segment, and assigns and removes them.

Its reads require the customersegment.segment\_read scope.
Its writes require the customersegment.segment\_manage scope.

**Command:** `epx segments groups [flags]`, `epx segments groups [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx segments groups delete`](#epx-segments-groups-delete)\
  Remove an IAM group from a segment.
- [`epx segments groups get`](#epx-segments-groups-get)\
  Show one IAM group of a segment.
- [`epx segments groups list`](#epx-segments-groups-list)\
  List the IAM groups of a segment.
- [`epx segments groups search`](#epx-segments-groups-search)\
  Search the IAM groups of a segment.
- [`epx segments groups update`](#epx-segments-groups-update)\
  Assign an IAM group to a segment.

## epx segments groups delete

Remove an IAM group from a segment.

Removes the IAM group GROUP, a group of customers (userType CUSTOMER), from the segment
SEGMENT\_ID; Emporix answers one that is not
assigned as removed.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, and GROUP the ID of an IAM group
that [`epx iam groups list`](/emporix-cli/commands/iam#epx-iam-groups-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type GROUP at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments groups delete SEGMENT_ID GROUP [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments groups get

Show one IAM group of a segment.

Reads one IAM group assigned to a segment.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one is not found.
GROUP is the GROUP that [`epx segments groups list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-groups-list) shows; an unknown one is not found.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments groups get SEGMENT_ID GROUP [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among GROUP, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments groups list

List the IAM groups of a segment.

Reads the IAM groups assigned to a segment.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments groups list SEGMENT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among GROUP, NAME.
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

## epx segments groups search

Search the IAM groups of a segment.

Reads the IAM groups assigned to a segment that an Emporix q filter finds (`--query`;
without it, every one).
It reads and shows them as [`epx segments groups list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-groups-list) does.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments groups search SEGMENT_ID [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among GROUP, NAME.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments groups update

Assign an IAM group to a segment.

Assigns the IAM group GROUP, a group of customers (userType CUSTOMER), to the segment
SEGMENT\_ID, or replaces its assignment (an upsert).
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, and GROUP the ID of an IAM group
that [`epx iam groups list`](/emporix-cli/commands/iam#epx-iam-groups-list) shows.
`--file` may give the assignment's mixins and metadata (its version and the schemas of
its mixins); without it, epx sends an empty object.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments groups update SEGMENT_ID GROUP [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments items

Read and change the items of a segment.

Reads the items assigned to a segment, products or categories, and assigns and removes them.

Its reads require the customersegment.segment\_read scope.
Its writes require the customersegment.segment\_manage scope.

**Command:** `epx segments items [flags]`, `epx segments items [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx segments items bulk-delete`](#epx-segments-items-bulk-delete)\
  Delete item assignments of a segment in bulk.
- [`epx segments items bulk-update`](#epx-segments-items-bulk-update)\
  Upsert item assignments of a segment in bulk.
- [`epx segments items delete`](#epx-segments-items-delete)\
  Remove a product or category from a segment.
- [`epx segments items get`](#epx-segments-items-get)\
  Show one item of a segment.
- [`epx segments items list`](#epx-segments-items-list)\
  List the items of a segment.
- [`epx segments items search`](#epx-segments-items-search)\
  Search the items of a segment.
- [`epx segments items update`](#epx-segments-items-update)\
  Assign a product or category to a segment.

## epx segments items bulk-delete

Delete item assignments of a segment in bulk.

Removes the products or categories (TYPE) whose ids `--file` gives from the segment
SEGMENT\_ID, a JSON array such as `["100200", "100300"]`.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows,
and TYPE is PRODUCT or CATEGORY, in capitals.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Deletes the item assignments of a segment that `--file` names with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type TYPE at its prompt.
`--yes` sends without asking (required without a terminal, or with `--file` -);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments items bulk-delete SEGMENT_ID TYPE [flags]`

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

## epx segments items bulk-update

Upsert item assignments of a segment in bulk.

Assigns the products or categories (TYPE) in `--file` to the segment SEGMENT\_ID, or replaces
their assignments (an upsert): a JSON array of assignments, each with the id of its item
and optionally its mixins and metadata.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows,
and TYPE is PRODUCT or CATEGORY, in capitals.
Emporix refuses an ID of no product or category of TYPE.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments items bulk-update SEGMENT_ID TYPE [flags]`

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

## epx segments items delete

Remove a product or category from a segment.

Removes the product or category ID from the segment
SEGMENT\_ID; Emporix answers one that is not
assigned as removed.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, TYPE is PRODUCT or CATEGORY, in capitals, and
ID the id of the product or category (as [`epx products list`](/emporix-cli/commands/products#epx-products-list) or [`epx categories list`](/emporix-cli/commands/categories#epx-categories-list)
shows it).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments items delete SEGMENT_ID TYPE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments items get

Show one item of a segment.

Reads one item assigned to a segment: TYPE is PRODUCT or CATEGORY, in capitals.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one is not found.
TYPE and ID are the TYPE and ID that [`epx segments items list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-items-list) shows; an unknown one is not found.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments items get SEGMENT_ID TYPE ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, SEGMENT, TYPE, CODE, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments items list

List the items of a segment.

Reads the items assigned to a segment, products or categories (TYPE).
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments items list SEGMENT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, SEGMENT, TYPE, CODE, NAME.
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

## epx segments items search

Search the items of a segment.

Reads the items assigned to a segment, products or categories (TYPE), that an Emporix q
filter finds (`--query`; without it, every one).
It reads and shows them as [`epx segments items list SEGMENT_ID`](/emporix-cli/commands/segments#epx-segments-items-list) does.
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows; an unknown one lists nothing.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments items search SEGMENT_ID [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, SEGMENT, TYPE, CODE, NAME.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments items update

Assign a product or category to a segment.

Assigns the product or category ID to the segment
SEGMENT\_ID, or replaces its assignment (an upsert).
SEGMENT\_ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows, TYPE is PRODUCT or CATEGORY, in capitals, and
ID the id of the product or category (as [`epx products list`](/emporix-cli/commands/products#epx-products-list) or [`epx categories list`](/emporix-cli/commands/categories#epx-categories-list)
shows it).
Emporix refuses an ID of no product or category of TYPE.
`--file` may give the assignment's mixins and metadata (its version and the schemas of
its mixins); without it, epx sends an empty object.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments items update SEGMENT_ID TYPE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments list

List segments.

Reads customer segments, which group customers to offer them items such as products or
categories. STATUS is ACTIVE or INACTIVE, and FROM and TO the validity window.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, STATUS, SITE, FROM, TO.
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

## epx segments match

List the products that belong to segments.

Reads which of the products that `--product` names (the product id, as in [`epx products list`](/emporix-cli/commands/products#epx-products-list)) belong to an active segment that `--segment` names, on the site that `--site` names,
directly or through a category. Emporix answers them whole, in one response: a product
named twice is listed twice, and unknown products are left out.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the customersegment.segment\_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx segments match [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting customersegment.segment\_manage, which can also change data (required).
- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PRODUCT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--product <ID>`\
  A product ID to check (repeatable; required without `--body`).
- `--segment <ID>`\
  A segment ID to check against (repeatable; required without `--body`).
- `--site <code>`\
  The site code to match on, e.g. main (required without `--body`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments patch

Change a segment.

Changes the segment ID with the JSON Patch operations (RFC 6902) of `--file`, a JSON array
such as `[{"op": "replace", "path": "/status", "value": "INACTIVE"}]`; epx refuses
anything else, which Emporix would answer as done without applying it, although it
raises the segment's metadata.version.
ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.
In an operation, a path to one translation, such as /name/de, takes a string in that
language.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments patch ID [flags]`

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

## epx segments search

Search segments.

Reads the customer segments that an Emporix q filter finds (`--query`; without it,
every segment).
It reads and shows them as [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the customersegment.segment\_read scope.

**Command:** `epx segments search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, STATUS, SITE, FROM, TO.
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

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx segments update

Upsert a segment.

Replaces the segment ID with the one in `--file`, a JSON object as
[`epx segments get ID -o json`](/emporix-cli/commands/segments#epx-segments-get) prints it, or creates it with that id where none exists
(an upsert); members it leaves out, such as a description, are removed.
ID is the ID that [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) shows.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Localized text in the body is a map of translations, e.g. `{"en": "Acme", "de": "Acme"}`,
as `-o json` of a read prints it, sent with Content-Language: \*. Emporix refuses a string in
its place.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customersegment.segment\_manage scope.

**Command:** `epx segments update ID [flags]`

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

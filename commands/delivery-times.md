# epx delivery-times

> Read and change delivery times and their time slots; read the delivery windows of a cart.

## epx delivery-times

Reads and changes delivery times, the days on which a site delivers to a zone, and their
time slots; reads the delivery windows open to a cart, checks and counts the orders of a
delivery window and names its delivery cycle.

Its reads require different OAuth scopes; the help of each names them.
Its writes require the shipping.shipping\_manage scope.

**Command:** `epx delivery-times [flags]`, `epx delivery-times [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx delivery-times bulk-create`](#epx-delivery-times-bulk-create)\
  Create delivery times in bulk.
- [`epx delivery-times create`](#epx-delivery-times-create)\
  Create a delivery time.
- [`epx delivery-times delete`](#epx-delivery-times-delete)\
  Delete a delivery time.
- [`epx delivery-times generate-cycle`](#epx-delivery-times-generate-cycle)\
  Name the delivery cycle of a delivery window.
- [`epx delivery-times get`](#epx-delivery-times-get)\
  Show one delivery time.
- [`epx delivery-times increment-counter`](#epx-delivery-times-increment-counter)\
  Count one more order in a delivery window.
- [`epx delivery-times list`](#epx-delivery-times-list)\
  List delivery times.
- [`epx delivery-times patch`](#epx-delivery-times-patch)\
  Change a delivery time.
- [`epx delivery-times slots`](#epx-delivery-times-slots)\
  Read and change the time slots of a delivery time.
- [`epx delivery-times update`](#epx-delivery-times-update)\
  Replace a delivery time.
- [`epx delivery-times validate`](#epx-delivery-times-validate)\
  Check whether a delivery window can take an order.
- [`epx delivery-times windows`](#epx-delivery-times-windows)\
  List the delivery windows open to a cart.

## epx delivery-times bulk-create

Create delivery times in bulk.

Creates the delivery times in `--file`, a JSON array of delivery times, each as create
takes it.

Emporix takes up to 50 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times bulk-create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times create

Create a delivery time.

Creates a delivery time from the JSON in `--file`: its siteCode and zoneId (or isForAllZones,
for a day without delivery on every zone of the site), its day (a weekday, singleDate or
datePeriod), isDeliveryDay, timeZoneId and deliveryDayShift, the slots of a delivery day
and optionally the delivery time's id and name, as [`epx delivery-times get ID -o json`](/emporix-cli/commands/delivery-times#epx-delivery-times-get)
prints one; a slot with a cutOffDayShift needs a delivery time without deliveryDayShift.
Emporix answers with its id, and refuses a delivery time that overlaps another of its
zone, or is not valid otherwise, with 400 Bad Request (exit 1).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times delete

Delete a delivery time.

Deletes the delivery time ID with its time slots.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times generate-cycle

Name the delivery cycle of a delivery window.

Prints the name of the delivery cycle of a delivery window, as Emporix names it in the
delivery windows of a cart and in orders (deliveryCycle in
[`epx delivery-times windows CART_ID -o json`](/emporix-cli/commands/delivery-times#epx-delivery-times-windows)): the delivery date, the hour (UTC) of the
slot's cut-off time and the deliveryCycleName of its cutOffTime, such as 2026-06-20-6morning.
The delivery window is the delivery time `--delivery-time` on `--date`, an ISO 8601
date and time such as 2026-11-02T12:00:00.000Z, in its `--slot`: [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list)
and [`epx delivery-times slots list`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) show their IDs.
Emporix answers an unknown delivery time with 403 Forbidden (exit 3), not as not found.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

No OAuth scope is needed.

**Command:** `epx delivery-times generate-cycle [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CYCLE.
- `--date <DATE>`\
  The delivery DATE and time, ISO 8601 (required without `--body`).
- `--delivery-time <ID>`\
  The delivery time's ID (required without `--body`).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--slot <ID>`\
  The ID of the delivery time's slot (required without `--body`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times get

Show one delivery time.

Reads delivery times, the days on which a site delivers to a zone. ZONE is a zone id
([`epx shipping zones list --site S`](/emporix-cli/commands/shipping#epx-shipping-zones-list)), or `"all zones"` for a day without delivery on every
zone; DAY is a weekday, a date or a date period (from/to); DELIVERS is false for a day
without delivery; DAY SHIFT is how many days the delivery follows an order placed before
the cut-off, when the delivery time sets it (a weekday's slots may set their own); SLOTS
counts the time slots, which [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) lists.

Requires the shipping.shipping\_read scope.

**Command:** `epx delivery-times get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, SITE, ZONE, DAY, DELIVERS, TIME ZONE, DAY SHIFT, SLOTS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times increment-counter

Count one more order in a delivery window.

Counts one more order in a delivery window that has capacity left.
The delivery window is the delivery time `--delivery-time` on `--date`, an ISO 8601
date and time such as 2026-11-02T12:00:00.000Z, in its `--slot`: [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list)
and [`epx delivery-times slots list`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) show their IDs.
The Checkout Service counts each order that it places with a delivery window; only a
checkout of your own needs this. Emporix answers true once it counted the order, and a
window that is full or not offered that day, or a date in the past or in another format,
with 400 Bad Request (exit 1). The body may also come whole from `--file`.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times increment-counter [flags]`

**Flags**

- `--date <DATE>`\
  The delivery DATE and time, ISO 8601 (required without `--file`).
- `--delivery-time <ID>`\
  The delivery time's ID (required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--slot <ID>`\
  The ID of the delivery time's slot (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times list

List delivery times.

Reads delivery times, the days on which a site delivers to a zone. ZONE is a zone id
([`epx shipping zones list --site S`](/emporix-cli/commands/shipping#epx-shipping-zones-list)), or `"all zones"` for a day without delivery on every
zone; DAY is a weekday, a date or a date period (from/to); DELIVERS is false for a day
without delivery; DAY SHIFT is how many days the delivery follows an order placed before
the cut-off, when the delivery time sets it (a weekday's slots may set their own); SLOTS
counts the time slots, which [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) lists.

Requires the shipping.shipping\_read scope.

**Command:** `epx delivery-times list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, SITE, ZONE, DAY, DELIVERS, TIME ZONE, DAY SHIFT, SLOTS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times patch

Change a delivery time.

Changes the delivery time ID with the JSON Patch (RFC 6902) in
`--file`: an array of operations whose op is add, remove or replace, such as
`[{"op": "replace", "path": "/name", "value": "Mondays"}]`.
An unknown ID is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots

Read and change the time slots of a delivery time.

Reads and changes the time slots of a delivery time.

Its reads require the shipping.shipping\_read scope.
Its writes require the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots [flags]`, `epx delivery-times slots [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx delivery-times slots create`](#epx-delivery-times-slots-create)\
  Create a time slot of a delivery time.
- [`epx delivery-times slots delete`](#epx-delivery-times-slots-delete)\
  Delete a time slot of a delivery time.
- [`epx delivery-times slots delete-all`](#epx-delivery-times-slots-delete-all)\
  Delete every time slot of a delivery time.
- [`epx delivery-times slots get`](#epx-delivery-times-slots-get)\
  Show one time slot of a delivery time.
- [`epx delivery-times slots list`](#epx-delivery-times-slots-list)\
  List the time slots of a delivery time.
- [`epx delivery-times slots patch`](#epx-delivery-times-slots-patch)\
  Change a time slot of a delivery time.
- [`epx delivery-times slots update`](#epx-delivery-times-slots-update)\
  Replace a time slot of a delivery time.

## epx delivery-times slots create

Create a time slot of a delivery time.

Creates a time slot of a delivery day from the JSON in `--file`: its shippingMethod (a method
of the delivery time's zone), deliveryTimeRange (timeFrom and timeTo, local times such as
08:00), cutOffTime (its time, an ISO 8601 timestamp whose time of day counts, and for a
weekday its cutOffDayShift, which the delivery time must not combine with a
deliveryDayShift), capacity (the orders it takes) and optionally its id, as
[`epx delivery-times slots get DELIVERY_TIME_ID ID -o json`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-get) prints one. Emporix answers
with its id, and a slot id that exists already with 400 Bad Request (exit 1), not as a
conflict.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots create DELIVERY_TIME_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots delete

Delete a time slot of a delivery time.

Deletes the time slot ID of a delivery time.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.
ID is the ID that [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) shows; an unknown one is answered as one that Emporix deleted (204 No Content).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots delete DELIVERY_TIME_ID ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots delete-all

Delete every time slot of a delivery time.

Deletes every time slot of a delivery time.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.

Deletes every time slot of the delivery time with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type DELIVERY\_TIME\_ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots delete-all DELIVERY_TIME_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots get

Show one time slot of a delivery time.

Reads one time slot of a delivery time.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.
ID is the ID that [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) shows; an unknown one is not found.
METHOD is the shipping method of the slot ([`epx shipping zones methods list`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-list)), FROM and TO
its delivery window in the delivery time's TIME ZONE, CUT-OFF the latest time of day to
order for the slot (sent as an ISO 8601 timestamp whose date Emporix ignores), CUT-OFF DAYS
how many days the delivery follows such an order (weekday delivery times only; otherwise
see DAY SHIFT in [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list)), and CAPACITY how many orders the slot takes.

Requires the shipping.shipping\_read scope.

**Command:** `epx delivery-times slots get DELIVERY_TIME_ID ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, METHOD, FROM, TO, CUT-OFF, CUT-OFF DAYS, CAPACITY.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots list

List the time slots of a delivery time.

Reads the time slots of a delivery time.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.
METHOD is the shipping method of the slot ([`epx shipping zones methods list`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-list)), FROM and TO
its delivery window in the delivery time's TIME ZONE, CUT-OFF the latest time of day to
order for the slot (sent as an ISO 8601 timestamp whose date Emporix ignores), CUT-OFF DAYS
how many days the delivery follows such an order (weekday delivery times only; otherwise
see DAY SHIFT in [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list)), and CAPACITY how many orders the slot takes.

Requires the shipping.shipping\_read scope.

**Command:** `epx delivery-times slots list DELIVERY_TIME_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, METHOD, FROM, TO, CUT-OFF, CUT-OFF DAYS, CAPACITY.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots patch

Change a time slot of a delivery time.

Changes the time slot ID of a delivery time with the JSON Patch (RFC 6902) in
`--file`: an array of operations whose op is add, remove or replace, such as
`[{"op": "replace", "path": "/capacity", "value": 50}]`.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.
ID is the ID that [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots patch DELIVERY_TIME_ID ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times slots update

Replace a time slot of a delivery time.

Replaces the time slot ID of a delivery time with the one in `--file`, as create takes it.
DELIVERY\_TIME\_ID is the ID that [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) shows; an unknown one is not found.
ID is the ID that [`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times slots update DELIVERY_TIME_ID ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times update

Replace a delivery time.

Replaces the delivery time ID with the one in `--file`, as create takes it. With
metadata.version in the body, as [`epx delivery-times get ID -o json`](/emporix-cli/commands/delivery-times#epx-delivery-times-get) prints it, Emporix
refuses the write as a conflict (exit 5) when the delivery time changed since it was read;
without it, Emporix replaces it whatever its version. An unknown ID is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the shipping.shipping\_manage scope.

**Command:** `epx delivery-times update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--validate-overlap <string>`\
  Whether Emporix refuses slots of one shipping method whose time ranges overlap (it does unless given false): true or false.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times validate

Check whether a delivery window can take an order.

Checks whether a delivery window can take an order.
The delivery window is the delivery time `--delivery-time` on `--date`, an ISO 8601
date and time such as 2026-11-02T12:00:00.000Z, in its `--slot`: [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list)
and [`epx delivery-times slots list`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) show their IDs.
VALID is true when it can; Emporix answers a date in the past or in another format, or a
window that is full or not offered that day, with 400 Bad Request (exit 1), and an unknown
delivery time as not found (exit 4).

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

No OAuth scope is needed.

**Command:** `epx delivery-times validate [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among VALID.
- `--date <DATE>`\
  The delivery DATE and time, ISO 8601 (required without `--body`).
- `--delivery-time <ID>`\
  The delivery time's ID (required without `--body`).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--slot <ID>`\
  The ID of the delivery time's slot (required without `--body`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx delivery-times windows

List the delivery windows open to a cart.

Reads the delivery windows in which a cart can be delivered to the country and postal code
set on the cart. DATE is the delivery day, FROM and TO the delivery time range on it,
METHOD the shipping method, ZONE the shipping zone, CUT-OFF the latest time to order for
the window, SLOT the time slot of its delivery time, which
[`epx delivery-times slots list DELIVERY_TIME_ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) lists, and ID the id of that delivery
time ([`epx delivery-times get ID`](/emporix-cli/commands/delivery-times#epx-delivery-times-get)); the delivery cycle is only in `-o json`.
For another postal code than the cart's, run
[`epx api get /shipping/{tenant}/actualDeliveryWindows/CART_ID -p postalCode=CODE`](/emporix-cli/commands/api#epx-api-get).
CART\_ID is the id of a cart; an unknown one is not found (exit 4).

No OAuth scope is needed.

**Command:** `epx delivery-times windows CART_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among DATE, FROM, TO, METHOD, ZONE, CUT-OFF, SLOT, ID.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

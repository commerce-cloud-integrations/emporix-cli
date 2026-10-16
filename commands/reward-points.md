# epx reward-points

> Read and change reward points and the ways to redeem them.

## epx reward-points

Reads customers' reward points, personal data included: ACTIVE is the number of points
a customer can redeem now (expired points are left out), ADDED and REDEEMED the points added
and redeemed so far, expired ones included (a failed redemption is refunded as an
addition), so ACTIVE may be less than ADDED minus REDEEMED. Each addition and redemption is
only in `-o json`.
[`epx reward-points redeem-options`](/emporix-cli/commands/reward-points#epx-reward-points-redeem-options) lists the ways to redeem points. The writes create and
delete a customer's reward points entry (create, delete), add and redeem points (add,
redeem), and create, replace and delete the ways to redeem them (redeem-option-create,
\-update, -delete).

Its reads require different OAuth scopes; the help of each names them.
Its writes require the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points [flags]`, `epx reward-points [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx reward-points add`](#epx-reward-points-add)\
  Add reward points to a customer.
- [`epx reward-points balance`](#epx-reward-points-balance)\
  Show the reward points balance of a customer.
- [`epx reward-points create`](#epx-reward-points-create)\
  Create the reward points entry of a customer.
- [`epx reward-points delete`](#epx-reward-points-delete)\
  Delete the reward points of a customer.
- [`epx reward-points get`](#epx-reward-points-get)\
  Show the reward points of a customer.
- [`epx reward-points list`](#epx-reward-points-list)\
  List the reward points of every customer.
- [`epx reward-points redeem`](#epx-reward-points-redeem)\
  Redeem reward points of a customer.
- [`epx reward-points redeem-option-create`](#epx-reward-points-redeem-option-create)\
  Create a way to redeem reward points.
- [`epx reward-points redeem-option-delete`](#epx-reward-points-redeem-option-delete)\
  Delete a way to redeem reward points.
- [`epx reward-points redeem-option-update`](#epx-reward-points-redeem-option-update)\
  Replace a way to redeem reward points.
- [`epx reward-points redeem-options`](#epx-reward-points-redeem-options)\
  List the ways to redeem reward points.

## epx reward-points add

Add reward points to a customer.

Adds `--points` reward points to a customer. Emporix answers without a body, so epx prints
nothing. Sent twice, it adds them twice.
The reward points settings of the site `--site` apply, the default site of the profile in
effect without it, and site main without either: how long the points stay valid. Emporix
keeps them in the site's couponSettings mixin; it accepts a site without them, also an
unknown site code, and the points then stay valid 36 months.
`--valid-until`, which gives these points their own end of validity, takes a time such as
2027-12-31T23:59:59Z: unlike the redeem options, Emporix reads a time that ends in Z.
A customer without a reward points entry is not found (exit 4).
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points add CUSTOMER [flags]`

**Flags**

- `--description <string>`\
  What the points are for.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--order <ID>`\
  The ID of the order that earned the points.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--points <POINTS>`\
  The number of POINTS to add (required without `--file`).
- `--site <CODE>`\
  The site CODE whose reward points settings apply, sent as siteCode.
- `--valid-until <TIME>`\
  The TIME until which the points can be redeemed, e.g. 2027-12-31T23:59:59Z.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx reward-points add C1001 --site AcmeStore_NL --points 50 --description "Goodwill"
```

## epx reward-points balance

Show the reward points balance of a customer.

Reads the reward points balance of a customer, personal data: POINTS, one number, the
points that the customer can redeem now: ACTIVE of [`epx reward-points get`](/emporix-cli/commands/reward-points#epx-reward-points-get), which also shows
the points added and redeemed so far.
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows; an unknown one is not found (exit 4).

Requires the rewardspoints.rewardpoints\_view scope.

**Command:** `epx reward-points balance CUSTOMER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among POINTS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx reward-points create

Create the reward points entry of a customer.

Creates the reward points entry of a customer, with the points it starts with where
`--points` gives them. Emporix answers without a body, so epx prints nothing. A customer
with an entry is a conflict (exit 5). Emporix does not check the customer: an unknown one
gets an entry too.
[`epx reward-points add`](/emporix-cli/commands/reward-points#epx-reward-points-add) adds points to an entry.
The reward points settings of the site `--site` apply, the default site of the profile in
effect without it, and site main without either: how long the points stay valid. Emporix
keeps them in the site's couponSettings mixin; it accepts a site without them, also an
unknown site code, and the points then stay valid 36 months.
The JSON of `--file` must give customerId, CUSTOMER again (Emporix answers 400 Bad Request to
another and 500 Internal Server Error without one, exit 1, and creates no entry), and the
points it starts with as startPoints, e.g.
`{"customerId": "C1001", "startPoints": {"points": 20}}`.
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points create CUSTOMER [flags]`

**Flags**

- `--description <string>`\
  What the points it starts with are for (only with `--points`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--points <POINTS>`\
  The number of POINTS the entry starts with.
- `--site <CODE>`\
  The site CODE whose reward points settings apply, sent as siteCode.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx reward-points create C1001 --site AcmeStore_NL --points 20 --description "Welcome"
```

## epx reward-points delete

Delete the reward points of a customer.

Deletes the reward points entry of a customer, with the points added and redeemed so far.
Emporix answers 200 OK, also for an unknown customer or one without an entry: neither is an
error.
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CUSTOMER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points delete CUSTOMER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx reward-points get

Show the reward points of a customer.

Reads customers' reward points, personal data included: ACTIVE is the number of points
a customer can redeem now (expired points are left out), ADDED and REDEEMED the points added
and redeemed so far, expired ones included (a failed redemption is refunded as an
addition), so ACTIVE may be less than ADDED minus REDEEMED. Each addition and redemption is
only in `-o json`.
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows; one without reward points is not found.

Requires the rewardspoints.rewardpoints\_view scope.

**Command:** `epx reward-points get CUSTOMER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, ACTIVE, ADDED, REDEEMED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx reward-points list

List the reward points of every customer.

Reads customers' reward points, personal data included: ACTIVE is the number of points
a customer can redeem now (expired points are left out), ADDED and REDEEMED the points added
and redeemed so far, expired ones included (a failed redemption is refunded as an
addition), so ACTIVE may be less than ADDED minus REDEEMED. Each addition and redemption is
only in `-o json`.

Requires the rewardspoints.rewardpoints\_view scope.

**Command:** `epx reward-points list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CUSTOMER, ACTIVE, ADDED, REDEEMED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx reward-points redeem

Redeem reward points of a customer.

Redeems `--points` reward points of a customer without making a coupon of them, the oldest
points that are still valid first. Emporix answers without a body, so epx prints nothing.
More points than are active (ACTIVE of [`epx reward-points get`](/emporix-cli/commands/reward-points#epx-reward-points-get)) are a conflict (exit 5).
Sent twice, it redeems them twice.
A customer without a reward points entry is not found (exit 4).
CUSTOMER is the customer id that [`epx reward-points list`](/emporix-cli/commands/reward-points#epx-reward-points-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points redeem CUSTOMER [flags]`

**Flags**

- `--description <string>`\
  What the points are redeemed for.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--points <POINTS>`\
  The number of POINTS to redeem (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx reward-points redeem C1001 --points 50 --description "Gift wrapping"
```

## epx reward-points redeem-option-create

Create a way to redeem reward points.

Creates a way to redeem reward points from the JSON in `--file`: its type (coupon), name,
the points it costs and, for a coupon, the coupon that it makes of them, as
[`epx coupons create`](/emporix-cli/commands/coupons#epx-coupons-create) takes it. It prints the option that Emporix answers with; sending the
same JSON twice creates two. The coupon's restrictions.validFrom and validUntil take epoch
milliseconds or a time with a numeric offset, e.g. 2027-01-01T00:00:00.000+00:00: Emporix
refuses a time that ends in Z (400 Bad Request).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points redeem-option-create [flags]`

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

## epx reward-points redeem-option-delete

Delete a way to redeem reward points.

Deletes the redeem option ID, as [`epx reward-points redeem-options`](/emporix-cli/commands/reward-points#epx-reward-points-redeem-options) shows it; it sends no body.
Emporix answers an unknown ID as one that it deleted (204 No Content).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points redeem-option-delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx reward-points redeem-option-update

Replace a way to redeem reward points.

Replaces the redeem option ID, as [`epx reward-points redeem-options`](/emporix-cli/commands/reward-points#epx-reward-points-redeem-options) shows it, with the
JSON in `--file`, as redeem-option-create takes it.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the rewardspoints.rewardpoints\_manage scope.

**Command:** `epx reward-points redeem-option-update ID [flags]`

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

## epx reward-points redeem-options

List the ways to redeem reward points.

Reads the ways customers can redeem reward points: their TYPE, NAME and the POINTS they cost.

No OAuth scope is needed.

**Command:** `epx reward-points redeem-options [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, NAME, POINTS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

# epx coupons

> Read, create, change and redeem coupons.

## epx coupons

Reads coupons, customer ids included (issuedTo, validFor in `-o json`). DISCOUNT is a
percentage, an amount or FREE_SHIPPING. STATUS is VALID, INACTIVE, EXPIRED or USED, as
Emporix computes it: `--query` status:VALUE matches no coupon and `--sort` status is ignored.
REDEEMED counts redemptions, out of the maximum when there is one; Emporix counts those for
a legal entity apart, per legal entity (legalEntityRedemptions, only in `-o json`). Other
restrictions (valid from, minimum order value, customers, categories, segments) are only in
`-o json`.
[`epx coupons redemptions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemptions) and `redemption` read the redemptions of a coupon, all or one.
[`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) leaves out deleted coupons; [`epx coupons get CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-get) still shows one, in
the table as a live one (`-o json` has deleted: true). Deleted coupons are listed too, one
page at a time, by
[`epx api get /coupon/{tenant}/coupons -p showDeleted=true --scope coupon.coupon_read`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).
[`epx coupons create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-create), `update`, `patch` and `delete` change coupons, `redeem` redeems one
outside a checkout, `redemption-delete` deletes a redemption, and `referral-create` creates
the referral coupon of a customer.

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx coupons [flags]`, `epx coupons [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx coupons create`](#epx-coupons-create)\
  Create a coupon.
- [`epx coupons delete`](#epx-coupons-delete)\
  Delete a coupon.
- [`epx coupons get`](#epx-coupons-get)\
  Show one coupon.
- [`epx coupons list`](#epx-coupons-list)\
  List coupons.
- [`epx coupons patch`](#epx-coupons-patch)\
  Change members of a coupon.
- [`epx coupons redeem`](#epx-coupons-redeem)\
  Redeem a coupon.
- [`epx coupons redemption`](#epx-coupons-redemption)\
  Show one redemption of a coupon.
- [`epx coupons redemption-delete`](#epx-coupons-redemption-delete)\
  Delete a redemption of a coupon.
- [`epx coupons redemptions`](#epx-coupons-redemptions)\
  List the redemptions of a coupon.
- [`epx coupons referral`](#epx-coupons-referral)\
  Show the referral coupon of a customer.
- [`epx coupons referral-create`](#epx-coupons-referral-create)\
  Create the referral coupon of a customer.
- [`epx coupons update`](#epx-coupons-update)\
  Replace a coupon.
- [`epx coupons validate`](#epx-coupons-validate)\
  Check whether a coupon can be redeemed.

## epx coupons create

Create a coupon.

Creates the coupon in `--file`, a JSON object with at least its name and, for a discount,
discountType (PERCENT with discountPercentage, ABSOLUTE with discountAbsolute, or
FREE_SHIPPING), and prints `{"id": CODE}`: the code of the JSON, or one that Emporix makes
without one. Emporix saves codes in upper case unless the tenant configuration
enableCouponCodeCaseSensitivity is true. A coupon with the code exists already: a conflict
(exit 5), also for a deleted one, whose code stays taken.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the coupon.coupon_manage scope.

**Command:** `epx coupons create [flags]`

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
epx coupons create --file coupon.json --dry-run
```

## epx coupons delete

Delete a coupon.

Deletes the coupon CODE, which [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) then leaves out. Emporix keeps it as
deleted: [`epx coupons get CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-get) still shows it, and its code stays taken. An unknown CODE is
not found (exit 4).
Emporix refuses a deleted coupon with 403 Forbidden (exit 1).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CODE at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the coupon.coupon_manage scope.

**Command:** `epx coupons delete CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons get

Show one coupon.

Reads coupons, customer ids included (issuedTo, validFor in `-o json`). DISCOUNT is a
percentage, an amount or FREE_SHIPPING. STATUS is VALID, INACTIVE, EXPIRED or USED, as
Emporix computes it: `--query` status:VALUE matches no coupon and `--sort` status is ignored.
REDEEMED counts redemptions, out of the maximum when there is one; Emporix counts those for
a legal entity apart, per legal entity (legalEntityRedemptions, only in `-o json`). Other
restrictions (valid from, minimum order value, customers, categories, segments) are only in
`-o json`.
[`epx coupons redemptions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemptions) and `redemption` read the redemptions of a coupon, all or one.
[`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) leaves out deleted coupons; [`epx coupons get CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-get) still shows one, in
the table as a live one (`-o json` has deleted: true). Deleted coupons are listed too, one
page at a time, by
[`epx api get /coupon/{tenant}/coupons -p showDeleted=true --scope coupon.coupon_read`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).

Requires the coupon.coupon_read scope.

**Command:** `epx coupons get CODE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, DISCOUNT, STATUS, REDEEMED, VALID UNTIL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons list

List coupons.

Reads coupons, customer ids included (issuedTo, validFor in `-o json`). DISCOUNT is a
percentage, an amount or FREE_SHIPPING. STATUS is VALID, INACTIVE, EXPIRED or USED, as
Emporix computes it: `--query` status:VALUE matches no coupon and `--sort` status is ignored.
REDEEMED counts redemptions, out of the maximum when there is one; Emporix counts those for
a legal entity apart, per legal entity (legalEntityRedemptions, only in `-o json`). Other
restrictions (valid from, minimum order value, customers, categories, segments) are only in
`-o json`.
[`epx coupons redemptions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemptions) and `redemption` read the redemptions of a coupon, all or one.
[`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) leaves out deleted coupons; [`epx coupons get CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-get) still shows one, in
the table as a live one (`-o json` has deleted: true). Deleted coupons are listed too, one
page at a time, by
[`epx api get /coupon/{tenant}/coupons -p showDeleted=true --scope coupon.coupon_read`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api#epx-api-get).

Requires the coupon.coupon_read scope.

**Command:** `epx coupons list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, DISCOUNT, STATUS, REDEEMED, VALID UNTIL.
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

## epx coupons patch

Change members of a coupon.

Changes the members of the coupon CODE that the JSON in `--file` names, such as status or
restrictions; the others keep their values. An unknown CODE is not found (exit 4).
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.
Without one, Emporix changes the coupon anyway.
Emporix refuses a deleted coupon with 403 Forbidden (exit 1).

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the coupon.coupon_manage scope.

**Command:** `epx coupons patch CODE [flags]`

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

## epx coupons redeem

Redeem a coupon.

Redeems a coupon for an order of `--order-total` that it takes `--discount` off, by
`--customer` (a customer number) where given, outside a checkout, and prints what Emporix
answers with: `{"id": ID}`, the redemption's id. Sent twice, it redeems the coupon twice.
Emporix refuses a coupon that cannot be redeemed and says why:
one redeemed as often as it may be with 400 Bad Request (exit 1); one that the customer may
not redeem, or that needs a customer where `--customer` is not given, with
403 Forbidden (exit 1). An unknown coupon is not found (exit 4).
`--legal-entity` names the legal entity that the customer is assigned to: Emporix
refuses it, with 400 Bad Request (exit 1), for a customer without a legal entity and for an
unknown one.
[`epx coupons validate`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-validate) runs the same checks without redeeming it.
`--customer` needs the scope below: Emporix refuses it to a token with coupon.coupon_manage
alone (403 Forbidden).
COUPON_CODE is the CODE that [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the coupon.coupon_redeem_on_behalf scope.

**Command:** `epx coupons redeem COUPON_CODE [flags]`

**Flags**

- `--customer <CUSTOMER_NUMBER>`\
  The CUSTOMER_NUMBER of the customer who redeems it.
- `--discount <AMOUNT CURRENCY>`\
  What the coupon takes off as AMOUNT CURRENCY (required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--legal-entity <ID>`\
  The ID of the customer's legal entity.
- `--order <CODE>`\
  The CODE of the order.
- `--order-total <AMOUNT CURRENCY>`\
  The order's total as AMOUNT CURRENCY, e.g. `"100 EUR"` (required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx coupons redeem SAVE7 --order-total "100 EUR" --discount "7 EUR" --customer 10001 --dry-run
```

## epx coupons redemption

Show one redemption of a coupon.

Reads one redemption of a coupon (personal data).
COUPON_CODE is the CODE that [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) shows; an unknown one is not found.
ID is the ID that [`epx coupons redemptions COUPON_CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemptions) shows; an unknown one is not found.
ORDER TOTAL is the order's value and DISCOUNT the amount the coupon took off it. The
narrower coupon.coupon_redeem scope reads only the redemptions a caller made itself.

Requires the coupon.coupon_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx coupons redemption COUPON_CODE ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting coupon.coupon_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ORDER, CUSTOMER, LEGAL ENTITY, ORDER TOTAL, DISCOUNT, REDEEMED AT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons redemption-delete

Delete a redemption of a coupon.

Deletes the redemption ID of a coupon, which then counts one redemption less: one that was
redeemed as often as it may be can be redeemed again.
COUPON_CODE is the CODE that [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) shows; an unknown one is not found.
ID is the ID that [`epx coupons redemptions COUPON_CODE`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemptions) shows; an unknown one is not found.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the coupon.coupon_manage scope.

**Command:** `epx coupons redemption-delete COUPON_CODE ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons redemptions

List the redemptions of a coupon.

Reads the redemptions of a coupon (personal data): the order, customer and legal entity
of each, and when it was redeemed ([`epx coupons redemption`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-redemption) shows one).
COUPON_CODE is the CODE that [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) shows; an unknown one is not found.
ORDER TOTAL is the order's value and DISCOUNT the amount the coupon took off it. The
narrower coupon.coupon_redeem scope reads only the redemptions a caller made itself.

Requires the coupon.coupon_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx coupons redemptions COUPON_CODE [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting coupon.coupon_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ORDER, CUSTOMER, LEGAL ENTITY, ORDER TOTAL, DISCOUNT, REDEEMED AT.
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

## epx coupons referral

Show the referral coupon of a customer.

Reads the referral coupon of a customer: the coupon issued to the customer (issuedTo in
`-o json`) to share with the people they refer, also a deleted one, which the table shows as
a live one (`-o json` has deleted: true). Its columns are those of [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list).
CUSTOMER_NUMBER is the NUMBER that [`epx customers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-list) shows; a customer without a
referral coupon, or an unknown one, is not found (exit 4).

No OAuth scope is needed.

**Command:** `epx coupons referral CUSTOMER_NUMBER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, NAME, DISCOUNT, STATUS, REDEEMED, VALID UNTIL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons referral-create

Create the referral coupon of a customer.

Creates the referral coupon of a customer, from the referral settings of the main site,
and prints it; a customer has one referral coupon at most, so where it exists, Emporix
answers with it and creates none. A deleted referral coupon stays the customer's: Emporix
answers with it and makes no new one ([`epx coupons referral`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-referral) shows it as VALID). Emporix
keeps the referral settings in the site's couponSettings mixin; where the main site has
none, it answers 400 Bad Request (exit 1). CUSTOMER_NUMBER is the NUMBER that
[`epx customers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-list) shows; an unknown one is not found (exit 4).
It changes data although Emporix needs no write-capable scope for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

No OAuth scope is needed.

**Command:** `epx coupons referral-create CUSTOMER_NUMBER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx coupons update

Replace a coupon.

Replaces the coupon CODE with the JSON in `--file`, as [`epx coupons get CODE -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-get)
prints it: a member that it leaves out is removed. [`epx coupons patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-patch) changes some
members only. An unknown CODE is not found (exit 4), as Emporix creates no coupon here.
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.
Without one, Emporix changes the coupon anyway.
Emporix refuses a deleted coupon with 403 Forbidden (exit 1).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the coupon.coupon_manage scope.

**Command:** `epx coupons update CODE [flags]`

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

## epx coupons validate

Check whether a coupon can be redeemed.

Checks whether a coupon can be redeemed for an order of `--order-total` that it takes
`--discount` off, by `--customer` (a customer number) where given: Emporix runs every check of
a redemption without redeeming the coupon. REDEEMABLE is true when it can be. Emporix
refuses a coupon that cannot be redeemed with an error that says why: one that the customer
may not redeem with 403 Forbidden (exit 1). An unknown coupon is not found (exit 4).
`--legal-entity` names the legal entity that the customer is assigned to: Emporix
refuses it, with 400 Bad Request (exit 1), for a customer without a legal entity and for an
unknown one.
Emporix answers a redeemable coupon without a body; `-o json` prints true then. The scope,
which epx requests also without `--customer`, lets a credential redeem coupons on a
customer's behalf too.
COUPON_CODE is the CODE that [`epx coupons list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons#epx-coupons-list) shows.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the coupon.coupon_redeem_on_behalf scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx coupons validate COUPON_CODE [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting coupon.coupon_redeem_on_behalf, which can also change data (required).
- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among REDEEMABLE.
- `--customer <CUSTOMER_NUMBER>`\
  The CUSTOMER_NUMBER of the customer who redeems it.
- `--discount <AMOUNT CURRENCY>`\
  What the coupon takes off as AMOUNT CURRENCY (required without `--body`).
- `--legal-entity <ID>`\
  The ID of the customer's legal entity.
- `--no-headers`\
  Print table and csv without the header row.
- `--order <CODE>`\
  The CODE of the order.
- `--order-total <AMOUNT CURRENCY>`\
  The order's total as AMOUNT CURRENCY, e.g. `"100 EUR"` (required without `--body`).
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

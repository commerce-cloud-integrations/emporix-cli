# epx customers

> Read and change customers and their addresses.

## epx customers

Reads and changes customer profiles, personal data included, and their addresses,
imports customers with their passwords and sets until when imported passwords are
migrated.

Its reads require different OAuth scopes; the help of each names them.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx customers [flags]`, `epx customers [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx customers addresses`](#epx-customers-addresses)\
  Read and change the addresses of a customer.
- [`epx customers create`](#epx-customers-create)\
  Create a customer.
- [`epx customers delete`](#epx-customers-delete)\
  Delete a customer.
- [`epx customers get`](#epx-customers-get)\
  Show one customer.
- [`epx customers import`](#epx-customers-import)\
  Import customers with their passwords.
- [`epx customers list`](#epx-customers-list)\
  List customers.
- [`epx customers password-migration`](#epx-customers-password-migration)\
  Read and set until when imported customers' passwords are migrated.
- [`epx customers patch`](#epx-customers-patch)\
  Change a customer.
- [`epx customers search`](#epx-customers-search)\
  Search customers.
- [`epx customers update`](#epx-customers-update)\
  Upsert a customer.

## epx customers addresses

Read and change the addresses of a customer.

Reads and changes the addresses of a customer, personal data included, and their tags.

Its reads require the customer.customer\_read scope.
Its writes require the customer.customer\_manage scope.

**Command:** `epx customers addresses [flags]`, `epx customers addresses [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx customers addresses add-tags`](#epx-customers-addresses-add-tags)\
  Tag an address of a customer.
- [`epx customers addresses create`](#epx-customers-addresses-create)\
  Add an address to a customer.
- [`epx customers addresses delete`](#epx-customers-addresses-delete)\
  Delete an address of a customer.
- [`epx customers addresses get`](#epx-customers-addresses-get)\
  Show one address of a customer.
- [`epx customers addresses list`](#epx-customers-addresses-list)\
  List the addresses of a customer.
- [`epx customers addresses patch`](#epx-customers-addresses-patch)\
  Change an address of a customer.
- [`epx customers addresses remove-tags`](#epx-customers-addresses-remove-tags)\
  Remove tags from an address of a customer.
- [`epx customers addresses update`](#epx-customers-addresses-update)\
  Upsert an address of a customer.

## epx customers addresses add-tags

Tag an address of a customer.

Adds the tags that `--tags` names, such as BILLING or SHIPPING, to the address ID of the
customer CUSTOMER\_NUMBER.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows, and ID the ID that
[`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses add-tags CUSTOMER_NUMBER ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--tags <TAGS>`\
  The TAGS to add, separated by commas, e.g. BILLING,SHIPPING (required).
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers addresses create

Add an address to a customer.

Adds the address in `--file`, a JSON object with at least contactName, to the customer
CUSTOMER\_NUMBER; Emporix makes its id unless the body gives one, and answers with it.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses create CUSTOMER_NUMBER [flags]`

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

## epx customers addresses delete

Delete an address of a customer.

Deletes the address ID of the customer CUSTOMER\_NUMBER.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows, and ID the ID that
[`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses delete CUSTOMER_NUMBER ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers addresses get

Show one address of a customer.

Reads one address of a customer, personal data included.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows; an unknown one is not found.
ID is the ID that [`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows; an unknown one is not found.
DEFAULT marks the default address; TAGS are labels such as BILLING or SHIPPING. Phone
numbers, states and extra address lines are only in `-o json`.

Requires the customer.customer\_read scope.

**Command:** `epx customers addresses get CUSTOMER_NUMBER ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CONTACT, COMPANY, STREET, ZIP, CITY, COUNTRY, DEFAULT, TAGS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers addresses list

List the addresses of a customer.

Reads the addresses of a customer, personal data included.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows; an unknown one is not found.
DEFAULT marks the default address; TAGS are labels such as BILLING or SHIPPING. Phone
numbers, states and extra address lines are only in `-o json`.

Requires the customer.customer\_read scope.

**Command:** `epx customers addresses list CUSTOMER_NUMBER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CONTACT, COMPANY, STREET, ZIP, CITY, COUNTRY, DEFAULT, TAGS.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers addresses patch

Change an address of a customer.

Changes the members of the address ID of the customer CUSTOMER\_NUMBER that `--file` gives,
a JSON object such as `{"isDefault": true}`; Emporix keeps the others.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows, and ID the ID that
[`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses patch CUSTOMER_NUMBER ID [flags]`

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

## epx customers addresses remove-tags

Remove tags from an address of a customer.

Removes the tags that `--tags` names from the address ID of the customer CUSTOMER\_NUMBER.
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows, and ID the ID that
[`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses remove-tags CUSTOMER_NUMBER ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--tags <TAGS>`\
  The TAGS to remove, separated by commas, e.g. BILLING,SHIPPING (required).
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers addresses update

Upsert an address of a customer.

Changes the members of the address ID of the customer CUSTOMER\_NUMBER that `--file` gives,
a JSON object as [`epx customers addresses get CUSTOMER_NUMBER ID -o json`](/emporix-cli/commands/customers#epx-customers-addresses-get) prints it,
keeping those it leaves out, such as its tags, or creates the address with that id where
none exists (an upsert).
CUSTOMER\_NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows, and ID the ID that
[`epx customers addresses list CUSTOMER_NUMBER`](/emporix-cli/commands/customers#epx-customers-addresses-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers addresses update CUSTOMER_NUMBER ID [flags]`

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

## epx customers create

Create a customer.

Creates the customer profile in `--file`, a JSON object with at least contactEmail, such as
[`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get) prints one; Emporix makes the customer number unless
the body gives it as id, and answers with it. An account with that email that exists
already is a conflict (exit 5).
Emporix ignores a customerNumber that a profile gives (no error).
Emporix emails a new customer a link to set a password only with `--password-reset-email`
(without it, epx sends sendPasswordResetNotifications=false, where Emporix's default is true).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--password-reset-email`\
  Email the customer a link to set a password once the account is created.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers delete

Delete a customer.

Deletes the customer profile NUMBER and its accounts. Emporix answers 202 Accepted, and
the profile is not found from then on.
NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type NUMBER at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers delete NUMBER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers get

Show one customer.

Reads customer profiles, personal data included. NUMBER is the customer number that get
takes, EMAIL the contact email (that of the account, with which the customer signs in,
may differ), BUSINESS MODEL is B2B (assigned to a company) or B2C, and ON HOLD marks a
suspended profile. B2B data and the default address are only in `-o json`; Emporix leaves
out a profile's addresses, accounts and mixins unless asked to expand them, which epx
does not: [`epx customers addresses list`](/emporix-cli/commands/customers#epx-customers-addresses-list) lists the addresses.

Requires the customer.customer\_read scope.

**Command:** `epx customers get NUMBER [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among NUMBER, NAME, EMAIL, COMPANY, BUSINESS MODEL, ACTIVE, ON HOLD.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers import

Import customers with their passwords.

Imports the customer profiles in `--file`, a JSON array of profiles as create takes them,
each with an account that holds its email and exactly one of passwordHash, a hash that
Emporix checks as its own, or legacyAuth, the hash of another system and its algorithm,
which Emporix replaces with its own as the customer signs in. Customers with legacyAuth
need a password migration retention: [`epx customers password-migration update`](/emporix-cli/commands/customers#epx-customers-password-migration-update) sets it.
Emporix ignores a customerNumber that a profile gives (no error).

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.import\_manage scope.

**Command:** `epx customers import [flags]`

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

## epx customers list

List customers.

Reads customer profiles, personal data included. NUMBER is the customer number that get
takes, EMAIL the contact email (that of the account, with which the customer signs in,
may differ), BUSINESS MODEL is B2B (assigned to a company) or B2C, and ON HOLD marks a
suspended profile. B2B data and the default address are only in `-o json`; Emporix leaves
out a profile's addresses, accounts and mixins unless asked to expand them, which epx
does not: [`epx customers addresses list`](/emporix-cli/commands/customers#epx-customers-addresses-list) lists the addresses.

Requires the customer.customer\_read scope.

**Command:** `epx customers list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among NUMBER, NAME, EMAIL, COMPANY, BUSINESS MODEL, ACTIVE, ON HOLD.
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

## epx customers password-migration

Read and set until when imported customers' passwords are migrated.

Reads, sets and removes the password migration retention of customer imports, until when
customers imported with a password hash from another system get an Emporix one.

Its reads require the customer.import\_read scope.
Its writes require the customer.import\_manage scope.

**Command:** `epx customers password-migration [flags]`, `epx customers password-migration [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx customers password-migration delete`](#epx-customers-password-migration-delete)\
  Remove the password migration retention.
- [`epx customers password-migration get`](#epx-customers-password-migration-get)\
  Show until when imported customers' passwords are migrated.
- [`epx customers password-migration update`](#epx-customers-password-migration-update)\
  Set until when imported customers' passwords are migrated.

## epx customers password-migration delete

Remove the password migration retention.

Removes the password migration retention. Emporix refuses it (400 Bad Request, exit 1)
while a customer still has a legacy password hash to migrate: move `--ends` of update
earlier instead to end the migration sooner. update sets the whole retention, so give
`--reminder` and `--emails` again where it should keep them.

Deletes the password migration retention of the tenant with one request, which cannot be undone.
On a terminal, epx sends DELETE only once you type the tenant at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the customer.import\_manage scope.

**Command:** `epx customers password-migration delete [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers password-migration get

Show until when imported customers' passwords are migrated.

Reads the password migration retention of customer imports: a customer imported with a
password hash from another system gets an Emporix one on signing in until ENDS; after
that, the customers not yet migrated must reset their password. REMINDER is when Emporix
asks them by email to sign in once, and EMAILS whether customers get these reminder and
reset emails. A tenant without this setting is not found (exit 4).

Requires the customer.import\_read scope.

**Command:** `epx customers password-migration get [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ENDS, REMINDER, EMAILS.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx customers password-migration update

Set until when imported customers' passwords are migrated.

Sets the whole password migration retention, or creates it where the tenant has none
(an upsert), from its flags or the JSON object of `--file`: until when customers imported
with legacyAuth get an Emporix password on signing in (`--ends`), when Emporix asks those
not yet migrated by email to sign in once (`--reminder`; without it, 7 days before `--ends`,
or tomorrow if that has passed, also where the retention it replaces had another date)
and whether customers get these reminder and reset emails (`--emails`: without it, epx
sends emailNotificationsEnabled=false; a `--file` without it gets Emporix's default, true).
Emporix answers with the retention.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.import\_manage scope.

**Command:** `epx customers password-migration update [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--emails`\
  Email customers the reminder and the password reset.
- `--ends <DATE>`\
  The DATE the migration ends, e.g. 2027-01-31 (required without `--file`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--reminder <DATE>`\
  The DATE of the reminder email, before `--ends`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx customers password-migration update --ends 2027-01-31 --reminder 2027-01-24 --emails
epx customers password-migration update --ends 2027-01-31 --yes
```

## epx customers patch

Change a customer.

Changes the members of the customer profile NUMBER that `--file` gives, a JSON object such
as `{"active": false}`; Emporix keeps the others.
NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows.
Emporix ignores a customerNumber that a profile gives (no error).
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.
On an existing profile, a contactEmail changes its contact email only, never the email
with which the customer signs in (that of its account).

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers patch NUMBER [flags]`

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

## epx customers search

Search customers.

Reads the customer profiles, personal data included, that an Emporix q filter finds
(`--query`; without it, every profile).
It reads and shows them as [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the customer.customer\_read scope.

**Command:** `epx customers search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among NUMBER, NAME, EMAIL, COMPANY, BUSINESS MODEL, ACTIVE, ON HOLD.
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

## epx customers update

Upsert a customer.

Changes the members of the customer profile NUMBER that `--file` gives, a JSON object as
[`epx customers get NUMBER -o json`](/emporix-cli/commands/customers#epx-customers-get) prints it, keeping those it leaves out, or creates the
profile with that number where none exists (an upsert), as create does: with the email of
an existing account, that is a conflict (exit 5).
NUMBER is the NUMBER that [`epx customers list`](/emporix-cli/commands/customers#epx-customers-list) shows.
Emporix ignores a customerNumber that a profile gives (no error).
Send its metadata.version as read to change it only if nobody changed it since:
Emporix answers 409 Conflict (exit 5) otherwise.
On an existing profile, a contactEmail changes its contact email only, never the email
with which the customer signs in (that of its account).
Emporix emails a new customer a link to set a password only with `--password-reset-email`
(without it, epx sends sendPasswordResetNotifications=false, where Emporix's default is true).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the customer.customer\_manage scope.

**Command:** `epx customers update NUMBER [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--password-reset-email`\
  Email the customer a link to set a password once the account is created.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

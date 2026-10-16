# epx webhooks

> Read and change webhook configurations and subscriptions; read statistics.

## epx webhooks

Reads the tenant's webhooks: where events are sent, which events it subscribes to, how many
webhooks it sent, and a link to its webhook dashboard; changes where events are sent and
which events it subscribes to.

Its reads require different OAuth scopes; the help of each names them.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the webhook.subscription\_manage scope.

**Command:** `epx webhooks [flags]`, `epx webhooks [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx webhooks configurations`](#epx-webhooks-configurations)\
  Read, create, change and delete webhook configurations.
- [`epx webhooks dashboard-access`](#epx-webhooks-dashboard-access)\
  Show a link that logs in to the webhook dashboard.
- [`epx webhooks statistics`](#epx-webhooks-statistics)\
  Show how many webhooks were sent per month.
- [`epx webhooks subscriptions`](#epx-webhooks-subscriptions)\
  Read and change the webhook events that the tenant subscribes to.

## epx webhooks configurations

Read, create, change and delete webhook configurations.

Reads webhook configurations: PROVIDER is SVIX\_SHARED, SVIX or HTTP, DESTINATION where
events are sent and ENTRIES counts the event entries, each with its own destination,
headers and filter (inactive ones included), which only `-o json` shows. The service returns
the configurations whole, so they are read with one request.

Its reads require the webhook.subscription\_read scope.
Its writes require the webhook.subscription\_manage scope.

**Command:** `epx webhooks configurations [flags]`, `epx webhooks configurations [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx webhooks configurations create`](#epx-webhooks-configurations-create)\
  Create a webhook configuration.
- [`epx webhooks configurations delete`](#epx-webhooks-configurations-delete)\
  Delete a webhook configuration.
- [`epx webhooks configurations get`](#epx-webhooks-configurations-get)\
  Show one webhook configuration.
- [`epx webhooks configurations list`](#epx-webhooks-configurations-list)\
  List webhook configurations.
- [`epx webhooks configurations patch`](#epx-webhooks-configurations-patch)\
  Change members of a webhook configuration.
- [`epx webhooks configurations update`](#epx-webhooks-configurations-update)\
  Replace a webhook configuration.

## epx webhooks configurations create

Create a webhook configuration.

Creates the webhook configuration in `--file`, a JSON object with its code, provider
(SVIX\_SHARED, SVIX or HTTP), active and configuration: none for SVIX\_SHARED, the apiKey of
your Svix account for SVIX, and for HTTP the destinationUrl, secretKey, headers and
eventsConfiguration (entries that send events of a type elsewhere, each without an id:
Emporix assigns one and refuses one given). For HTTP, Emporix checks that the destinationUrl
is reachable (400 Bad Request, exit 1, `"not reachable"` otherwise). Emporix answers with the
code. A tenant has at most one configuration per provider: when it has one, Emporix
answers 409 Conflict (exit 5); change that one with update or patch.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the webhook.subscription\_manage scope.

**Command:** `epx webhooks configurations create [flags]`

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

**Examples**

```sh
epx webhooks configurations create --file http.json --dry-run
```

## epx webhooks configurations delete

Delete a webhook configuration.

Deletes the webhook configuration CODE; Emporix deletes an active one only with `--force`.
CODE is the code that [`epx webhooks configurations list`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-list) shows, such as http.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CODE at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the webhook.subscription\_manage scope.

**Command:** `epx webhooks configurations delete CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete the configuration also while it is active.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx webhooks configurations get

Show one webhook configuration.

Reads webhook configurations: PROVIDER is SVIX\_SHARED, SVIX or HTTP, DESTINATION where
events are sent and ENTRIES counts the event entries, each with its own destination,
headers and filter (inactive ones included), which only `-o json` shows. The service returns
the configurations whole, so they are read with one request.

Requires the webhook.subscription\_read scope.

**Command:** `epx webhooks configurations get CODE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, PROVIDER, ACTIVE, DESTINATION, ENTRIES.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx webhooks configurations list

List webhook configurations.

Reads webhook configurations: PROVIDER is SVIX\_SHARED, SVIX or HTTP, DESTINATION where
events are sent and ENTRIES counts the event entries, each with its own destination,
headers and filter (inactive ones included), which only `-o json` shows. The service returns
the configurations whole, so they are read with one request.

Requires the webhook.subscription\_read scope.

**Command:** `epx webhooks configurations list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, PROVIDER, ACTIVE, DESTINATION, ENTRIES.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx webhooks configurations patch

Change members of a webhook configuration.

Changes the webhook configuration CODE with the JSON array of operations in `--file`, each
with op (UPSERT sets the member at path to value; REMOVE removes an item of a list), path
and value, e.g. `[{"op":"UPSERT","path":"/active","value":false}]`. For HTTP, an UPSERT of
/configuration/http/eventsConfigurationEntry adds an entry (without an id), and one of
/configuration/http/eventsConfigurationEntry/ID/active with false stops sending events to
the destination of the entry ID. CODE is the code that [`epx webhooks configurations list`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-list) shows, such as http.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the webhook.subscription\_manage scope.

**Command:** `epx webhooks configurations patch CODE [flags]`

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

**Examples**

```sh
echo '[{"op":"UPSERT","path":"/active","value":false}]' |
  epx webhooks configurations patch http --file - --yes
```

## epx webhooks configurations update

Replace a webhook configuration.

Replaces the webhook configuration CODE with the one in `--file`, a JSON object as create
takes it. Keep the id of each entry of eventsConfiguration that stays, as
[`epx webhooks configurations get CODE -o json`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-get) prints it; Emporix assigns an id to an entry
without one. A read gives secretKeyExists, not the secret: send secretKey (and apiKey for
SVIX) again, or change single members with patch. For HTTP, the destinationUrl must be reachable,
as at create.
CODE is the code that [`epx webhooks configurations list`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-list) shows, such as http.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the webhook.subscription\_manage scope.

**Command:** `epx webhooks configurations update CODE [flags]`

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

## epx webhooks dashboard-access

Show a link that logs in to the webhook dashboard.

Reads a URL that logs in to the tenant's webhook dashboard, the Svix application portal of
a tenant whose webhooks Emporix sends through its shared Svix account (provider
SVIX\_SHARED). The URL holds its own authentication, so treat it as a secret; `-o json` also
prints the portal's token.

Requires the webhook.subscription\_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx webhooks dashboard-access [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting webhook.subscription\_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among URL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx webhooks statistics

Show how many webhooks were sent per month.

Reads the monthly webhook LIMIT and the webhooks sent per month, for the months Emporix
reports by default. For other months, run
[`epx api get /webhook/{tenant}/statistics -p fromYearMonth=2026-01 -p toYearMonth=2026-06 --scope webhook.subscription_read`](/emporix-cli/commands/api#epx-api-get).

Requires the webhook.subscription\_read scope.

**Command:** `epx webhooks statistics [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among LIMIT, SENT PER MONTH.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx webhooks subscriptions

Read and change the webhook events that the tenant subscribes to.

Reads the webhook events and whether the tenant subscribes to them, and subscribes to
events or unsubscribes from them.

Its reads require the webhook.subscription\_read scope.
Its writes require the webhook.subscription\_manage scope.

**Command:** `epx webhooks subscriptions [flags]`, `epx webhooks subscriptions [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx webhooks subscriptions bulk-patch`](#epx-webhooks-subscriptions-bulk-patch)\
  Subscribe to webhook events and unsubscribe from them.
- [`epx webhooks subscriptions list`](#epx-webhooks-subscriptions-list)\
  List the webhook events and whether the tenant subscribes to them.

## epx webhooks subscriptions bulk-patch

Subscribe to webhook events and unsubscribe from them.

Changes the tenant's webhook event subscriptions with the JSON array in `--file`, an item per
event: its eventType, action (SUBSCRIBE or UNSUBSCRIBE), the metadata.version of its
subscription, as [`epx webhooks subscriptions list -o json`](/emporix-cli/commands/webhooks#epx-webhooks-subscriptions-list) prints it (Emporix answers a
stale one with 409 Conflict), and optionally fieldsToUnsubscribe, the fields of the event
to leave out of its payload, and fieldsToSubscribe, those to send again. Each result
names the eventType of its item, as each failure that epx lists does.

Emporix takes up to 200 items in one request (epx refuses more), carries out each on
its own and answers with the result of each: when one failed, epx prints the response,
lists the items that failed on standard error (one that failed with 500, 502, 503 or 504
may have been carried out all the same) and exits with the code that their failures
share (for example 5 for a 409 Conflict, 4 for a 404), else 1.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the webhook.subscription\_manage scope.

**Command:** `epx webhooks subscriptions bulk-patch [flags]`

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

**Examples**

```sh
epx webhooks subscriptions bulk-patch --file subscriptions.json --dry-run
```

## epx webhooks subscriptions list

List the webhook events and whether the tenant subscribes to them.

Reads every webhook event type and whether the tenant receives it: STATUS is SUBSCRIBED,
UNSUBSCRIBED or NONE. The service returns them whole, so they are read with one request.

Requires the webhook.subscription\_read scope.

**Command:** `epx webhooks subscriptions list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among EVENT, STATUS, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](/emporix-cli/commands#global-flags).

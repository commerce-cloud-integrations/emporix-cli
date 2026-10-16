# epx payment-transactions

> Read payment transactions; authorize, capture, refund and cancel payments.

## epx payment-transactions

Reads payment transactions, personal data included, and runs payments through the
tenant's payment providers: authorize starts a transaction for an order, capture collects
an authorized amount, refund returns a captured one and cancel voids an authorization.
Capture, refund and cancel move money.

Its reads require the payment-gateway.paymenttransactions_read scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx payment-transactions [flags]`, `epx payment-transactions [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx payment-transactions authorize`](#epx-payment-transactions-authorize)\
  Authorize a payment.
- [`epx payment-transactions cancel`](#epx-payment-transactions-cancel)\
  Cancel an authorized payment.
- [`epx payment-transactions capture`](#epx-payment-transactions-capture)\
  Capture an authorized payment.
- [`epx payment-transactions get`](#epx-payment-transactions-get)\
  Show one payment transaction.
- [`epx payment-transactions list`](#epx-payment-transactions-list)\
  List payment transactions.
- [`epx payment-transactions refund`](#epx-payment-transactions-refund)\
  Refund a captured payment.

## epx payment-transactions authorize

Authorize a payment.

Authorizes the payment of an order with a payment mode, as `--file`, a JSON object, gives
them: order.id, paymentModeId, amountToAuthorize and currency, and what the provider
needs, such as creditCardToken. The provider reserves the amount, which capture then
collects. Emporix answers with the paymentTransactionId and, where the customer must
finish the payment with the provider, where to send them (externalPaymentRedirectURL).

Emporix answers whether the operation succeeded (successful). epx prints the response and
exits with 1 unless it says that it did: where it says that it failed (successful false,
with a message), and where it does not say, when it may have been carried out all the same.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the payment-gateway.authorize_manage scope.

**Command:** `epx payment-transactions authorize [flags]`

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
epx payment-transactions authorize --file authorize.json --dry-run
```

## epx payment-transactions cancel

Cancel an authorized payment.

Cancels the authorized payment of transaction ID: the provider releases the amount that
it reserved.

This moves money: the confirmation names the transaction in the URL that it shows. ID is
the ID that [`epx payment-transactions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-list) shows; an unknown one is not found (exit 4).

Emporix answers whether the operation succeeded (successful). epx prints the response and
exits with 1 unless it says that it did: where it says that it failed (successful false,
with a message), and where it does not say, when it may have been carried out all the same.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the payment-gateway.cancel_manage scope.

**Command:** `epx payment-transactions cancel ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx payment-transactions cancel T1 --dry-run
```

## epx payment-transactions capture

Capture an authorized payment.

Captures `--amount` in `--currency` of the authorized payment of transaction ID, at most the
amount authorized: the provider collects it from the customer. Emporix answers with the
provider's captureId, which refund `--capture-id` takes.

This moves money: the confirmation names the transaction in the URL that it shows. ID is
the ID that [`epx payment-transactions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-list) shows; an unknown one is not found (exit 4).

Emporix answers whether the operation succeeded (successful). epx prints the response and
exits with 1 unless it says that it did: where it says that it failed (successful false,
with a message), and where it does not say, when it may have been carried out all the same.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the payment-gateway.capture_manage scope.

**Command:** `epx payment-transactions capture ID [flags]`

**Flags**

- `--amount <AMOUNT>`\
  The AMOUNT to capture, e.g. 3.99 (required without `--file`).
- `--currency <CURRENCY>`\
  The CURRENCY of `--amount`, e.g. EUR (required without `--file`).
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
epx payment-transactions capture T1 --amount 3.99 --currency EUR --dry-run
```

## epx payment-transactions get

Show one payment transaction.

Reads payment transactions, personal data included. PAYMENT MODE is the id of the payment
mode used, and ENTRIES the operations the provider ran, in the order sent: type and amount,
marked failed when they failed. Payment details (such as bank details and pay-by-link URLs),
card tokens and the provider's requests and responses are only in `-o json`. list's `--sort`
takes only createdDate (e.g. createdDate:desc); Emporix rejects other and several fields.

Requires the payment-gateway.paymenttransactions_read scope.

**Command:** `epx payment-transactions get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PROVIDER, PAYMENT MODE, ENTRIES.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx payment-transactions list

List payment transactions.

Reads payment transactions, personal data included. PAYMENT MODE is the id of the payment
mode used, and ENTRIES the operations the provider ran, in the order sent: type and amount,
marked failed when they failed. Payment details (such as bank details and pay-by-link URLs),
card tokens and the provider's requests and responses are only in `-o json`. list's `--sort`
takes only createdDate (e.g. createdDate:desc); Emporix rejects other and several fields.

Requires the payment-gateway.paymenttransactions_read scope.

**Command:** `epx payment-transactions list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PROVIDER, PAYMENT MODE, ENTRIES.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given: createdDate only, e.g. createdDate:desc; another field, or several, is refused (400, exit 1).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx payment-transactions refund

Refund a captured payment.

Refunds `--amount` in `--currency` of the captured payment of transaction ID, at most the
amount captured: the provider returns it to the customer. Where the payment was captured
more than once, `--capture-id` names the capture to refund.

This moves money: the confirmation names the transaction in the URL that it shows. ID is
the ID that [`epx payment-transactions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-list) shows; an unknown one is not found (exit 4).

Emporix answers whether the operation succeeded (successful). epx prints the response and
exits with 1 unless it says that it did: where it says that it failed (successful false,
with a message), and where it does not say, when it may have been carried out all the same.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the payment-gateway.refund_manage scope.

**Command:** `epx payment-transactions refund ID [flags]`

**Flags**

- `--amount <AMOUNT>`\
  The AMOUNT to refund, e.g. 3.99 (required without `--file`).
- `--capture-id <ID>`\
  The ID of the capture to refund, as capture answered it.
- `--currency <CURRENCY>`\
  The CURRENCY of `--amount`, e.g. EUR (required without `--file`).
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
epx payment-transactions refund T1 --amount 3.99 --currency EUR --dry-run
```

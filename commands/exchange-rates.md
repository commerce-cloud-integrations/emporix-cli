# epx exchange-rates

> Read, create, change and delete currency exchange rates.

## epx exchange-rates

Reads currency exchange rates. CODE names a rate, e.g. EUR_CHF.

Its reads require the currency.currency_read scope.
Its writes require the currency.currency_manage scope.

**Command:** `epx exchange-rates [flags]`, `epx exchange-rates [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx exchange-rates create`](#epx-exchange-rates-create)\
  Create an exchange rate.
- [`epx exchange-rates delete`](#epx-exchange-rates-delete)\
  Delete an exchange rate.
- [`epx exchange-rates get`](#epx-exchange-rates-get)\
  Show one exchange rate.
- [`epx exchange-rates list`](#epx-exchange-rates-list)\
  List exchange rates.
- [`epx exchange-rates update`](#epx-exchange-rates-update)\
  Change an exchange rate.

## epx exchange-rates create

Create an exchange rate.

Creates the exchange rate from the currency `--source` into `--target` (both defined in the
tenant, as [`epx currencies list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-list) shows them), or the one in the JSON of `--file`. Emporix
answers with its code, SOURCE_TARGET; a rate that exists already is a conflict (exit 5).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the currency.currency_manage scope.

**Command:** `epx exchange-rates create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--rate <RATE>`\
  The RATE: what one unit of the source currency is worth in the target currency, e.g. 0.95 (sent as a string; required without `--file`).
- `--source <CODE>`\
  The CODE of the currency to convert from (required without `--file`).
- `--target <CODE>`\
  The CODE of the currency to convert into (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx exchange-rates create --source EUR --target CHF --rate 0.95
```

## epx exchange-rates delete

Delete an exchange rate.

Deletes the exchange rate CODE; it sends no body.
Emporix answers an unknown CODE as one that it deleted (204 No Content).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type CODE at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the currency.currency_manage scope.

**Command:** `epx exchange-rates delete CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx exchange-rates get

Show one exchange rate.

Reads currency exchange rates. CODE names a rate, e.g. EUR_CHF.

Requires the currency.currency_read scope.

**Command:** `epx exchange-rates get CODE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, SOURCE, TARGET, RATE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx exchange-rates list

List exchange rates.

Reads currency exchange rates. CODE names a rate, e.g. EUR_CHF.

Requires the currency.currency_read scope.

**Command:** `epx exchange-rates list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CODE, SOURCE, TARGET, RATE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx exchange-rates update

Change an exchange rate.

Changes the rate of the exchange rate CODE (its currencies cannot change) to `--rate`, with
its metadata.version as read (`--version`, as [`epx exchange-rates get CODE -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-get) prints
it), or to the JSON of `--file`. A stale version and an unknown CODE are both not found (exit 4):
Emporix answers 404 to both.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the currency.currency_manage scope.

**Command:** `epx exchange-rates update CODE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--rate <RATE>`\
  The RATE: what one unit of the source currency is worth in the target currency, e.g. 0.95 (sent as a string; required without `--file`).
- `--version <VERSION>`\
  The VERSION of the rate as read (metadata.version; required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx exchange-rates update EUR_CHF --rate 0.96 --version 2
```

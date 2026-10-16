# epx invoice-jobs

> Create and read invoice jobs.

## epx invoice-jobs

Creates and reads the jobs that generate the invoices of orders.

Its reads require the invoice.invoice\_view scope.
Its writes require the invoice.invoice\_manage scope.

**Command:** `epx invoice-jobs [flags]`, `epx invoice-jobs [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx invoice-jobs create`](#epx-invoice-jobs-create)\
  Start an invoice job.
- [`epx invoice-jobs get`](#epx-invoice-jobs-get)\
  Show the status of an invoice job.

## epx invoice-jobs create

Start an invoice job.

Starts a job that generates the invoices of orders: of those that `--order` names
(`--type` MANUAL, which requires one at least), or of those ready for an invoice that the
tenant's invoice configuration selects (`--type` AUTOMATIC, which takes no `--order`).
Without an invoice configuration of the tenant's own, an order is ready for an invoice when
its mixins.generalAttributes.extendedOrderStatus is 70, 75 or 76 ([`epx orders create`](/emporix-cli/commands/orders#epx-orders-create)
stores 10; [`epx orders patch`](/emporix-cli/commands/orders#epx-orders-patch) changes it); the job's entry of an order with 10 fails as not
ready. Invoicing a ready order calls the tenant's Pick-Pack service: where that service is
not reachable, the entry fails at the VALIDATION step.
Emporix answers with the job's id (jobId), which [`epx invoice-jobs get ID`](/emporix-cli/commands/invoice-jobs#epx-invoice-jobs-get) reads.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the invoice.invoice\_manage scope.

**Command:** `epx invoice-jobs create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--order <ID>`\
  The ID of an order to invoice, repeatable (only with `--type` MANUAL, which requires it).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--type <TYPE>`\
  The TYPE of job: MANUAL or AUTOMATIC (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx invoice-jobs create --type MANUAL --order O1001 --order O1002
epx invoice-jobs create --type AUTOMATIC --yes
```

## epx invoice-jobs get

Show the status of an invoice job.

Reads an invoice job: STATUS is IN\_PROGRESS or DONE and TYPE MANUAL (orders chosen when
the job was created) or AUTOMATIC (orders that the tenant's configuration selects). ORDERS
counts the orders the job processed, of which INVOICED got an invoice, FAILED did not and
PENDING wait for one. Each order's entry in invoices (orderId and status; errorStep and
message for a failure, e.g. an order not ready for invoicing; for a success, the reference
documents the invoice number and download link) is only in `-o json`.
ID is the job id that creating the job returned. Emporix answers an unknown job with
500 Internal Server Error (the reference documents 404), so get fails with exit code 1.

Requires the invoice.invoice\_view scope.

**Command:** `epx invoice-jobs get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among STATUS, TYPE, ORDERS, INVOICED, FAILED, PENDING.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

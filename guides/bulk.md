# Bulk writes, chains, and jobs

> Write many items in one request, import files of any size, run cart command chains, and wait for jobs.

Some writes change many items at once, or start work that Emporix finishes later. They
follow the same rules as [every write](/emporix-cli/guides/writes), and they report each item that
failed so you know exactly what to send again.

## Bulk writes

```sh
epx availability bulk-update --file stock.json --dry-run
epx categories assignments bulk-create c1 --file assignments.json --yes
```

`bulk-create`, `bulk-update`, `bulk-patch`, and `bulk-delete` send a JSON array of items
to the service's bulk endpoint in one request. Each service takes a maximum number of
items per request, such as 200 for prices, products, and availability. epx refuses a
larger array with exit code 2 before sending anything.

Emporix carries out each item on its own and answers with the result of each. A few bulk
writes, such as `price-lists prices bulk-delete`, are answered as a whole; their help says
so. epx prints the response, and when items failed, it lists each one on standard error by its position
in your array (from 0) and its ID:

```text
item 3 (p-17): 409 Conflict: ...
```

The items that didn't fail were carried out. The command exits with the code the
failures share: 5 when each was a conflict, 4 when each wasn't found, 3 when each was
refused for a scope, and 1 otherwise.

  An item that failed with 500, 502, 503, or 504 may have been carried out all the same,
  as its line says. Read it before you send it again.

## Import products

```sh
epx products import --file products.json --dry-run
epx products import --file products.json --yes
```

`products import` takes a JSON array of any number of products and sends them in requests
of up to 200, in order. `--dry-run` prints every request, and the confirmation says how
many requests it sends.

epx lists the products that failed by their position in the whole file, and goes on
with the next request. If a request fails as a whole, the import stops and says from
which product on nothing was sent.

## Cart command chains

```sh
epx carts execute CART_ID --file commands.json
```

[`epx carts execute`](/emporix-cli/commands/carts#epx-carts-execute) sends up to 10 cart commands,
which Emporix runs in order on one cart. By default, Emporix stops at the first command
that fails but keeps what the commands before it changed; pass `--on-error resume` to run
the rest anyway.

When a command fails, epx lists it by its position and type, says how many commands
didn't run, and exits with the code the failures share:

```text
command 1 (UpdateCartItem): 404 Not Found: ...
```

## Wait for jobs

```sh
epx indexing reindex-jobs create --entity-type PRODUCT --rag --wait
epx search indexes update vehicle vehicles --file index.json --wait --yes
```

Writes that start a job, such as a reindex, print the job as Emporix answers it. Add
`--wait` to follow the job until it ends:

```text
reindex job j1: IN_PROGRESS (2000 of 5000)
```

epx reads the job every `--poll-interval` (default `5s`) and prints a line to standard
error whenever its status or progress changes, until the job ends or `--wait-timeout`
(default `30m`) passes. It then prints the job as last read and exits:

| Result | Exit code |
| --- | --- |
| The job ended successfully. | 0 |
| The job failed, or `--wait-timeout` passed first. The job keeps running, and the error names the command that reads it. | 1 |
| Reading the job failed. | The code of that read, such as 4 when it isn't found. |

`--wait` also requests the scope that reads the job, so a key without it fails with exit
code 3 before the write is sent. A long wait gets a new token before the old one
expires.

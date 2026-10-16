# Change data

> Create, update, and delete data in your tenant with dry runs, confirmation, and explicit scopes.

Reads never change anything. Writes create, change, or delete data in your tenant, and
every write works the same way: it sends only what you give it, requests explicit
scopes, shows its exact request with `--dry-run`, and sends only once you confirm.

```sh
epx units create --file gram.json --dry-run   # show the request
epx units create --file gram.json             # send it after you confirm
```

## Write commands

Write commands are named after what they send:

| Command | Sends |
| --- | --- |
| `create` | A POST that creates an item. |
| `update ID` | A PUT that replaces an item. Some services also create a missing one. |
| `patch ID` | A PATCH that changes some fields of an item. |
| `delete ID` | A DELETE. |
| `bulk-create`, `bulk-update`, `bulk-patch`, `bulk-delete` | Many items in one request. See [Bulk writes](https://commerce-cloud-integrations.github.io/emporix-cli/guides/bulk). |
| An action's verb, such as `reindex`, `capture`, or `redeem` | The action. |

The [CLI reference](https://commerce-cloud-integrations.github.io/emporix-cli/commands) lists every write with its scopes. For an endpoint without
a command of its own, use [`epx api post|put|patch|delete`](https://commerce-cloud-integrations.github.io/emporix-cli/guides/api#send-a-write).

## Provide a request body

```sh
epx prices update 6245aa0a78a8576e338fa9c4 --file price.json
jq -n '{active: false}' | epx sites patch AcmeStore_NL --file - --yes
```

`--file PATH` sends the JSON in a file byte for byte, and `--file -` reads it from
standard input. epx checks that it's one JSON value before anything else, and fails with
exit code 2 otherwise.

Some writes don't need a file:

- Writes with a small, fixed body build it from flags, such as
  `indexing reindex-jobs create --entity-type PRODUCT` or
  `orders transition ORDER_ID --status STATUS`. `--file` replaces those flags.
- A change of one configuration value is `configurations update KEY --file`, which sends the whole
  setting: `configurations get KEY -o json` shows it, to edit and send back.
- A delete of one item, and other writes that Emporix takes without a body, have no
  `--file`.
- Writes that upload a file, such as `schemas references create --json-schema PATH`,
  send a multipart form.

Each command's help shows which flags make its body.

### Localized text

Writes whose body holds localized text, such as a product's name, send a
`Content-Language` header. By default, `--content-language *`, each text is a map of
translations, as `-o json` prints it:

```json
{ "name": { "en": "Gram", "de": "Gramm" } }
```

With a language code, such as `--content-language en`, each text is a string in that
language. Some services refuse strings, so their writes take maps only and have no
`--content-language`. A few, such as the Shipping Service, refuse `Content-Language: *`,
so their writes send maps without the header. Each command's help says how it sends
text.

## Preview with a dry run

```sh
epx api put '/product/{tenant}/products/100200' --scope product.product_manage \
  --file product.json --dry-run
```

`--dry-run` prints the exact request and exits with code 0. It doesn't request a token
for the write and sends nothing. Every write works this way, with no read first.

```text
PUT https://api.emporix.io/product/acme1/products/100200
Accept: application/json
Accept-Language: *
Content-Type: application/json
User-Agent: emporix-cli/...

{"code":"100200", ...}
```

The output shows the method, the full URL with your tenant and query, every header except
`Authorization`, and the body. For a multipart form, it shows each part's name, file
name, content type, and size, and the content of JSON parts. A dry run still needs your
[API key settings](https://commerce-cloud-integrations.github.io/emporix-cli/configure), because it shows the request for your tenant.

## Confirm a write

On a terminal, epx asks before it sends:

```text
Send PUT https://api.emporix.io/product/acme1/products/100200 to tenant acme1? [y/N]
```

Only `y` or `yes` sends the request. Any other answer sends nothing, and the command
fails with `not confirmed; nothing was sent` and exit code 1.

A delete can't be undone, so `y` isn't enough: you type what it deletes, as the prompt
shows. A delete of many items in one request says so:

```text
Send DELETE https://api.emporix.io/category/acme1/categories/c1/assignments to tenant acme1? It deletes the assignments of every product to the category with one request and cannot be undone: type c1 to confirm:
```

### Skip the confirmation

`--yes` sends without asking. epx requires it, and fails with exit code 2 before sending
anything, when it can't ask:

- Standard input isn't a terminal, as in scripts, pipelines, and CI jobs.
- The body comes from standard input (`--file -`).

## Scopes

A write requests only the write scopes of its operation, which its help names. Some
flags add a scope their query parameter needs, and the help says so. For example,
`categories create --publish` adds `category.category_publish`.

`epx api` writes request only the scopes you name with `--scope`.

If your API key lacks a scope, epx fails with exit code 3 before it sends the write.

## Read the response

epx prints the response body as JSON (or YAML with `-o yaml`) with sorted keys and every
value as sent. An empty response, such as a 204 No Content, prints nothing. Any 2xx
status is a success.

Payment operations, such as `payment-transactions capture`, are the exception: Emporix
answers with `successful`, which says whether the payment provider carried out the
operation. epx exits with code 1 unless it's `true`.

## Handle failures

| Failure | Exit code | What to do |
| --- | --- | --- |
| Conflict: the item changed since you read it, or already exists. | 5 | Read the item again, apply your change to its current state, and send it again. |
| The write may have been carried out: a 500, 502, 503, or 504 answer, or a network failure after the connection was made. | 1 | Read the item to check whether the change was made before you retry. The error says `it may have been carried out all the same`. |
| Authentication or a missing scope. | 3 | Fix the API key, or add the scope the help names. Emporix didn't carry out the write. |
| Not found. | 4 | Check the ID. |

Most services that check `metadata.version` answer a stale version with a conflict
(exit code 5), so a write sent with the version you read never overwrites someone else's
change. The Unit and Catalog services answer it with 404, which epx reports as a conflict
when your body carries `metadata.version`. Some services answer it as not found or as a
bad request instead; the command's help says how.

### No repeats

epx never sends a write again once it may have reached Emporix, whatever its method. It
retries only when Emporix certainly didn't act: after 429 Too Many Requests, honoring
`Retry-After` up to 30 seconds, or when the connection couldn't be made. It sends a
request at most three times.

## Best practices

- Run every new write with `--dry-run` first, and check the URL, tenant, and body.
- Try writes on a test tenant before you run them on a tenant that serves customers.
- Use an API key with only the write scopes you need, and keep write scopes off the keys
  your read-only scripts use.
- Read an item right before you change it, and send its `metadata.version` where the
  service checks it.
- In scripts, pass `--yes` explicitly and check the exit code.

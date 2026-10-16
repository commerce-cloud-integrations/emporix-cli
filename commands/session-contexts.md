# epx session-contexts

> Read and change session contexts.

## epx session-contexts

Reads the context of a storefront session: its customer (once logged in, personal data;
ANONYMOUS before), site, currency, language, LOCATION (its target country) and active cart;
its custom attributes are only in `-o json`. Emporix creates the context with the session's
first cart; before that, the session's ID is not found (exit 4). The context of an
anonymous cart may name the tenant's default site rather than the cart's, and
names no cart once the cart was merged into a customer's.

Its reads require the sessioncontext.context\_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.
Its writes require the sessioncontext.context\_manage scope.

**Command:** `epx session-contexts [flags]`, `epx session-contexts [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx session-contexts attributes`](#epx-session-contexts-attributes)\
  Add and delete custom attributes of session contexts.
- [`epx session-contexts get`](#epx-session-contexts-get)\
  Show one session context.
- [`epx session-contexts update`](#epx-session-contexts-update)\
  Replace a session context.

## epx session-contexts attributes

Add and delete custom attributes of session contexts.

Adds and deletes the custom attributes of a session context, as its context member holds
them ([`epx session-contexts get ID -o json`](/emporix-cli/commands/session-contexts#epx-session-contexts-get)).

Requires the sessioncontext.context\_manage scope.

**Command:** `epx session-contexts attributes [flags]`, `epx session-contexts attributes [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx session-contexts attributes create`](#epx-session-contexts-attributes-create)\
  Add a custom attribute to a session context.
- [`epx session-contexts attributes delete`](#epx-session-contexts-attributes-delete)\
  Delete a custom attribute of a session context.

## epx session-contexts attributes create

Add a custom attribute to a session context.

Adds the custom attribute `--key` with the string `--value` to the context of the storefront
session SESSION\_ID; `--file` gives the body instead, e.g. `{"key": "k", "value": {"a": "b"}}` for a
value that is a JSON object.
SESSION\_ID is the ID of [`epx session-contexts get`](/emporix-cli/commands/session-contexts#epx-session-contexts-get); an unknown one is not found (exit 4).
Emporix does not check the version of the context for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the sessioncontext.context\_manage scope.

**Command:** `epx session-contexts attributes create SESSION_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--key <KEY>`\
  The KEY of the attribute (required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--value <VALUE>`\
  The VALUE of the attribute, a string (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx session-contexts attributes delete

Delete a custom attribute of a session context.

Deletes the custom attribute KEY of the context of the storefront session SESSION\_ID.
SESSION\_ID is the ID of [`epx session-contexts get`](/emporix-cli/commands/session-contexts#epx-session-contexts-get); an unknown one is not found (exit 4).
Emporix does not check the version of the context for it.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type KEY at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the sessioncontext.context\_manage scope.

**Command:** `epx session-contexts attributes delete SESSION_ID KEY [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx session-contexts get

Show one session context.

Reads the context of a storefront session: its customer (once logged in, personal data;
ANONYMOUS before), site, currency, language, LOCATION (its target country) and active cart;
its custom attributes are only in `-o json`. Emporix creates the context with the session's
first cart; before that, the session's ID is not found (exit 4). The context of an
anonymous cart may name the tenant's default site rather than the cart's, and
names no cart once the cart was merged into a customer's.

Requires the sessioncontext.context\_manage scope.
Opt-in: epx requests a write-capable scope for a read only when given `--allow-write-scope`.

**Command:** `epx session-contexts get ID [flags]`

**Flags**

- `--allow-write-scope`\
  Opt in to requesting sessioncontext.context\_manage, which can also change data (required).
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, CUSTOMER, SITE, CURRENCY, LANGUAGE, LOCATION, CART.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](/emporix-cli/commands#global-flags).

## epx session-contexts update

Replace a session context.

Replaces the context of the storefront session ID with the JSON in `--file`, as
[`epx session-contexts get ID -o json`](/emporix-cli/commands/session-contexts#epx-session-contexts-get) prints it, with its metadata.version as read, which
Emporix requires. With `--upsert`, Emporix creates the context where the session has none
(an upsert). Emporix answers an anonymous or unknown session with 400 Bad Request (exit 1)
that requires the header saas-token, with or without `--upsert`, which epx does not send; a
session that a customer logged in to needs that customer's token too.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the sessioncontext.context\_manage scope.

**Command:** `epx session-contexts update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--upsert`\
  Create the context where the session has none.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](/emporix-cli/commands#global-flags).

# epx auth

> Check Emporix credentials.

## epx auth

**Command:** `epx auth [flags]`, `epx auth [command]`

**Flags:** only the [global flags](/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx auth status`](#epx-auth-status)\
  Request a token and show the tenant and scopes it grants.

## epx auth status

Request a token and show the tenant and scopes it grants.

Requests a service token for the tenant plus the given scopes and shows what was granted.
Emporix grants only requested scopes, so pass `--scope` to check specific permissions;
a missing scope fails with exit code 3. The token itself is never printed.

**Command:** `epx auth status [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among TENANT, EXPIRES IN, SCOPES.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--scope <strings>`\
  Scope the credential must grant (repeatable).

Also takes the [global flags](/emporix-cli/commands#global-flags).

**Examples**

```sh
epx auth status --scope product.product_read_unpublished --scope indexing.search_view
```

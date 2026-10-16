# Exit codes and scripting

> Use epx in scripts and CI: what each exit code means, how errors look, and how epx retries.

epx is built for scripts. Output goes to standard output, errors go to standard error,
and the exit code tells you what went wrong, so a script can tell a missing product from
a missing scope.

## Exit codes

| Code | Meaning | Examples |
| --- | --- | --- |
| `0` | Success | The command printed its result. |
| `1` | Error | Emporix answered with an error such as 400 or 500, a network failure or timeout, a [paging check](/emporix-cli/guides/lists#how-paging-works) that failed, a declined confirmation, or a write that may have been carried out. |
| `2` | Invalid usage or configuration | An unknown command or flag, a bad flag value, missing API key settings, a body that isn't JSON, or a write without `--yes` where epx can't ask. Nothing was sent. |
| `3` | Authentication failed or scopes missing | Emporix rejected the API key, the token belongs to another tenant, a scope wasn't granted, or Emporix answered 401 or 403. |
| `4` | Not found | Emporix doesn't know the item, such as `epx products get` with an unknown ID. |
| `5` | Conflict | Emporix answered a write with 409 Conflict: the item changed since you read it, or already exists. A stale version sent to `units update` or `catalogs patch`, and a bulk write whose failed items were all conflicts, exit 5 too. Read the item again before you retry. |

  Some commands get a 403 from Emporix that refuses their input rather than the
  credential, such as `orders transition` to a status the order can't take. These exit
  with 1 and show Emporix's message, unless the message names a scope or permission.
  Their help says so.

## Errors

Errors start with `Error:` and go to standard error. When Emporix answers with JSON, epx
shows its message, its error code, and the details of each field, and nothing else of the
response:

```text
Error: 400 Bad Request: The request violates validation constraints. (ENTITY-400001) (name: must not be blank)
```

Usage errors (exit code 2) add a hint to run `epx --help`.

## Retries

epx retries a request, up to three times in total, only when it can't change data twice:

- Any request after 429 Too Many Requests, honoring `Retry-After` up to 30 seconds.
- Any request whose connection couldn't be made, unless making it timed out.
- Reads, never writes, after 502, 503, or 504.

## Script examples

### Check whether a product exists

```sh
if epx products get 100200 -o json > product.json; then
  echo "found"
else
  case $? in
    4) echo "no such product" ;;
    3) echo "check the API key and its scopes" ;;
    2) echo "check the command or the environment" ;;
    *) echo "failed" ;;
  esac
fi
```

### Handle a conflict

```sh
jq -n '{active: false}' | epx api patch '/site/{tenant}/sites/AcmeStore_NL' \
  --scope site.site_manage --file - --yes
case $? in
  0) echo "done" ;;
  5) echo "the site changed meanwhile: read it again"; exit 5 ;;
  *) echo "failed: read the site before you send it again"; exit 1 ;;
esac
```

### Fail early on a missing scope

```sh
epx auth status --scope order.order_read > /dev/null || exit
epx orders count --query status:CREATED -o json
```

### Run in CI

```sh
export EMPORIX_TENANT=acme1
export EMPORIX_CLIENT_ID="$CI_EMPORIX_CLIENT_ID"
export EMPORIX_CLIENT_SECRET="$CI_EMPORIX_CLIENT_SECRET"
epx products list -o json > products.json
```

In CI, set the `EMPORIX_` variables from your CI system's secrets. Profiles keep their
client secret in an OS keychain, which CI runners usually don't have. See
[Authenticate](/emporix-cli/configure).

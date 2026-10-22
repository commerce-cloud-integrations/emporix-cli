# epx

[![Go](https://img.shields.io/badge/go-1.27.3%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-epx-4F46E5)](https://commerce-cloud-integrations.github.io/emporix-cli/)

epx is a command-line client for the [Emporix Commerce APIs](https://developer.emporix.io/api-references).
It helps you read, script, and change the data of an Emporix tenant right from the terminal.

**Documentation:** [commerce-cloud-integrations.github.io/emporix-cli](https://commerce-cloud-integrations.github.io/emporix-cli/) covers installation, authentication, every command, and the Emporix API coverage.

**With epx, you can:**

- Read products, categories, catalogs, prices, availability, customers, orders, and nearly
  every other Emporix resource as a table, CSV, JSON, YAML, or JSON Lines.
- Filter, sort, and page through any list without writing paging code.
- Change data safely: every write shows its exact request with `--dry-run` and asks before
  it sends.
- Call any Emporix API endpoint with `epx api`, using only the OAuth scopes you name.
- Script reliably with stable output and exit codes that tell a missing item from a missing
  scope.

[![epx demo](site/images/demo.gif)](site/images/demo.webm)

> epx is an unofficial project. It isn't affiliated with or endorsed by Emporix.

## Installation

epx is a single binary for macOS, Linux, and Windows, on x64 (amd64) and arm64.

### Install script

On macOS or Linux:

```sh
curl -fsSL https://github.com/commerce-cloud-integrations/emporix-cli/releases/latest/download/install.sh | sh
```

On Windows (PowerShell):

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://github.com/commerce-cloud-integrations/emporix-cli/releases/latest/download/install.ps1 | iex"
```

The script downloads the latest release for your platform, checks it against the release's
checksums and installs `epx` into `~/.local/bin` (on Windows `%LOCALAPPDATA%\Programs\epx\bin`),
without sudo, adding that directory to your `PATH` if needed. Set `EPX_VERSION` (such as
`v0.1.2`) to pin a release, `EPX_INSTALL_DIR` to choose the directory, or `EPX_NO_MODIFY_PATH=1`
to leave your shell profile alone.

### npm

With Node.js 18 or later:

```sh
npm install --global @commercecloudintegrations/epx-cli
```

npm installs only the binary of your platform. Nothing is compiled and no install script
runs. To run epx once without installing it:

```sh
npx @commercecloudintegrations/epx-cli version
```

### Release archive

Each [release](https://github.com/commerce-cloud-integrations/emporix-cli/releases) holds an
archive per platform and a checksums file. For example, on Linux x64:

```sh
version=v0.1.3
base=https://github.com/commerce-cloud-integrations/emporix-cli/releases/download/$version
curl -fsSLO "$base/epx_${version}_linux_amd64.tar.gz" &&
  curl -fsSLO "$base/epx_${version}_checksums.txt" &&
  sha256sum --check --ignore-missing "epx_${version}_checksums.txt" &&
  tar -xzf "epx_${version}_linux_amd64.tar.gz" &&
  sudo install "epx_${version}_linux_amd64/epx" /usr/local/bin/epx
```

On macOS, use `shasum -a 256 --check --ignore-missing` instead of `sha256sum`.

### Go

With Go 1.27.3 or later:

```sh
go install github.com/commerce-cloud-integrations/emporix-cli/cmd/epx@latest
```

### From source

```sh
git clone https://github.com/commerce-cloud-integrations/emporix-cli.git
cd emporix-cli
go build -o bin/epx ./cmd/epx
```

## Upgrading

| Installed with | Upgrade |
| --- | --- |
| npm | `npm update --global @commercecloudintegrations/epx-cli` |
| A release archive | Replace the binary with the one from the new release. |
| Go | Run `go install` again. |

## Uninstalling

1. If you use profiles, remove each one with `epx profile remove NAME`, which also deletes
   its client secret from the OS keychain.
2. Remove epx: `npm uninstall --global @commercecloudintegrations/epx-cli`, or delete the `epx` binary you installed.
3. Delete the `epx` folder from your configuration directory: `~/.config` on Linux,
   `~/Library/Application Support` on macOS, `%AppData%` on Windows.

## Getting started

epx authenticates with an Emporix API key. Create a key with only the scopes of the
commands you run, then give it to epx through the environment:

```sh
export EMPORIX_TENANT=acme1
export EMPORIX_CLIENT_ID=...
export EMPORIX_CLIENT_SECRET=...
```

Or keep it in a profile, whose client secret goes to the OS keychain:

```sh
epx profile add dev --tenant acme1 --client-id ... --site AcmeStore_NL   # asks for the secret
```

Then make your first reads:

```sh
epx auth status --scope product.product_read_unpublished    # check the key and its scopes
epx sites list
epx products list --query published:false --sort code:desc --limit 20
epx products get 100200 -o json
epx api get '/catalog/{tenant}/catalogs'                     # any endpoint
```

And preview a write before you send it:

```sh
epx units create --file gram.json --dry-run    # print the exact request
epx units create --file gram.json              # send it after you confirm
```

## Usage

```sh
epx [command]

# Run `--help` for detailed information about a command, its flags, and its scopes
epx [command] --help
```

## Commands

Every command's help names the OAuth scopes it requests. The
[CLI reference](https://commerce-cloud-integrations.github.io/emporix-cli/commands) documents
every command and flag.

| Area | Commands |
| --- | --- |
| Products and catalog | [`products`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products), [`categories`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories), [`category-trees`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees), [`catalogs`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/catalogs), [`brands`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/brands), [`labels`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/labels), [`media`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media), [`product-templates`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates), [`units`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/units) |
| Prices and promotions | [`prices`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/prices), [`price-lists`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-lists), [`price-models`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/price-models), [`fees`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/fees), [`taxes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/taxes), [`currencies`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies), [`exchange-rates`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates), [`coupons`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/coupons), [`reward-points`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/reward-points) |
| Stock and shipping | [`availability`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/availability), [`shipping`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shipping), [`delivery-times`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/delivery-times) |
| Customers and B2B | [`customers`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers), [`segments`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/segments), [`legal-entities`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities), [`customer-locations`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations), [`contact-assignments`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments), [`shopping-lists`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists) |
| Orders and checkout | [`carts`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/carts), [`quotes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/quotes), [`orders`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders), [`returns`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns), [`invoice-jobs`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/invoice-jobs), [`payment-modes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes), [`payment-transactions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions) |
| Vendors | [`vendors`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors), [`vendor-locations`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations) |
| Tenant setup | [`sites`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites), [`countries`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/countries), [`regions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/regions), [`configurations`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations), [`iam`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam), [`webhooks`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/webhooks), [`schemas`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/schemas), [`custom-entities`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities), [`sequential-ids`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sequential-ids), [`session-contexts`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/session-contexts), [`imports`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/imports), [`changelog`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/changelog) |
| Search and AI | [`search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/search), [`indexing`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing), [`ai`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai) |
| Any endpoint | [`api get\|post\|put\|patch\|delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/api) |
| epx itself | [`auth`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/auth), [`profile`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/profile), [`completion`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/completion), [`version`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/version) |

## Safety

- **Reads change nothing**, including those that Emporix serves as POST or PUT, such as
  searches and shipping quotes.
- **Writes are explicit.** Every write shows its exact request with `--dry-run`, asks before
  it sends (a delete only once you type what it deletes), needs `--yes` in scripts, and is
  never sent again once it may have reached Emporix.
- **Least privilege.** Every command requests a token with only the scopes it needs, and
  fails with exit code 3 when your API key lacks one, instead of silently returning less
  data. Reads that need a write-capable scope run only with `--allow-write-scope`.
- **Credentials stay private.** epx reads its API key from the environment or a profile,
  never from flags, and never prints its token or client secret.
- **One host, HTTPS only.** Requests go only to the configured API host, over HTTPS (plain HTTP
  only to a loopback address, such as a local test server), and redirects are never followed.

`-o json` and `-o yaml` print responses as Emporix sends them, secrets included. See the
[security model](https://commerce-cloud-integrations.github.io/emporix-cli/security).

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | Error |
| `2` | Invalid usage or configuration; nothing was sent |
| `3` | Authentication failed or scopes missing |
| `4` | Not found |
| `5` | Conflict: the item changed since you read it, or already exists |

## Documentation

The [epx documentation](https://commerce-cloud-integrations.github.io/emporix-cli/) covers
installation, authentication, output formats, lists, writes, scripting, and the
[Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage) of
every command. [COMPATIBILITY.md](COMPATIBILITY.md) lists the Emporix API operations epx
sends, per reference version.

## Development

epx is written in Go with [Cobra](https://github.com/spf13/cobra).

```sh
go build -o bin/epx ./cmd/epx
go test ./...    # API calls go to local fake servers; no credentials needed
```

Besides Cobra, epx depends on [go-keyring](https://github.com/zalando/go-keyring) for the
client secrets of profiles in the OS keychain (no cgo), and on
[x/term](https://pkg.go.dev/golang.org/x/term) and [x/sys](https://pkg.go.dev/golang.org/x/sys)
to read a secret without showing it and to lock the configuration file on Windows.

## License

Licensed under the [Apache License 2.0](LICENSE).

epx is maintained by [Commerce Cloud Integrations](https://www.commercecloudintegrations.com).

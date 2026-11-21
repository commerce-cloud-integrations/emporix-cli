# Changelog

All notable changes to epx are listed here. epx follows
[semantic versioning](https://semver.org): while the major version is 0, a minor version can
change or remove commands, columns or flags, and its notes list each such change under
**Breaking changes**.

## Unreleased

### Changed

- `epx fees get` reads the fee only from the `fee` member of the response, where Emporix and its API
  reference put it, and fails (exit 1) on a response without it. Before, it read such a response as the fee.

### Fixed

- `epx ai agents update` help names `enableMemory`, the member Emporix keeps: it ignores
  `enabledMemory`, the name its API reference gives, and says the JSON that `epx ai agents get ID -o json`
  prints can be sent back as it is. `epx ai chat` help says a reply can be one message, not only a
  list.
- `epx ai agents update` help says Emporix refuses a mixin without a schema URL in
  `metadata.mixins`, one whose schema it cannot download and a value its schema does not allow.
- `epx returns update` and `patch` help say Emporix ignores siteCode and restriction there,
  where its reference documents that it refuses them.

## 0.5.0 (2026-11-19)

### Breaking changes

- `--content-language` and `epx quotes pdf --language` send a language code as you type it,
  where epx sent it in lower case before: `--content-language EN-gb` now sends `EN-gb`. Emporix
  matches codes regardless of letter case in the Product, Brand, Label and Site Settings
  services, but other services need the code as the tenant's language configuration writes it.
  A table still finds a translation in any letter case, and a profile keeps its language as
  typed.

### Fixed

- A language code is checked as typed: only ASCII letters and digits, so a letter that only
  becomes ASCII in lower case, such as the Kelvin sign, is refused.

## 0.4.0 (2026-11-19)

### Breaking changes

- `epx imports statistics`: the ABORTED column comes after CANCELLED, so in the table and in
  `-o csv` ACTIVE, RECORDS READ and RECORDS FAILED move from columns 6–8 to 7–9 (`--columns`
  picks columns by name).
- `epx returns list|get`: the SITE column (siteCode) comes after STATUS, so in the table and in
  `-o csv` RECEIVED, ORDERS, REQUESTOR, TOTAL and CREATED move from columns 3–7 to 4–8
  (`--columns` picks columns by name).

### New

- `epx ai chat --allow-file-generation` requests the Media Service's media.asset_manage, which an
  agent with the built-in file generator needs to create files and images; the reply lists them
  as artifacts.
- `epx imports streams files requeue STREAM_ID ID` queues a failed file again, from the content
  Emporix holds, so the stream's next run reads it without an upload.
- `epx imports statistics` shows ABORTED, the runs the Import Service refused to run for their
  configuration.
- `epx imports configurations streams` and `epx imports streams get` show a FAN_OUT stream, which
  writes further objects such as a product's prices for each record of another stream, with
  `fan-out of STREAM` as its SOURCE.
- `--language` and `--content-language` accept every code that Emporix accepts: two or three
  letters, then parts of 2 to 8 letters or digits, in any letter case, so custom codes such as
  `de-ch-zh` or `en-12`, refused before, now work. Codes accepted before are sent as before.

### Changed

- epx is checked against the Emporix API reference of 2026-11-18.
- The Import Service help names, for example, the statuses of a run's streams (CANCELLED; QUEUED
  and SKIPPED only on older runs), the stages of a run's errors (EXTRACT, MAP, VALIDATE, UPSERT),
  the error code FAN_OUT_REMOVE and the failed record (payload), the outcomes of imported records
  (DELETED_TARGET only on older records) and how `search` matches them, what TARGET TYPE means
  for each writer, the stream order (seq, FAN_OUT), the trigger and streams of a run (RETRY,
  streamIds), the newest imported file of a queue (lastProcessed), which a SNAPSHOT stream reads
  again when no file waits, and when Emporix refuses a run: 400 when there is nothing to import,
  409 also while a data purge runs or for a run that cannot be retried.
- `epx indexing providers` help names BatteryIncluded's apiUrl, enrichments and
  siteAwareMixinsStrategy, and `update` says that a body without them sets them back to their
  defaults, as Emporix replaces the configuration whole.
- The AI help names the built-in file generator (`file_generator` in an agent's nativeTools,
  not a tool), the mixins of agents and MCP servers (an update without them clears them), the
  provenance in generated replies and texts, and the reply of an agent job.
- The returns help says that Emporix copies a return's siteCode and restriction from its orders
  and that update and patch do not take them, how `--query siteCode:DE` filters by site, that a
  credential with only a site-limited read scope fails with exit code 3, and that the reference
  now documents 404 for the delete of an unknown return.
- `epx prices match` says how Emporix ranks the price lists that apply, and
  `epx carts items create` that an item whose itemType is EXTERNAL sends no priceId.

### Fixed

- epx is built with Go 1.27.4, which fixes a vulnerability in the standard library's HTTP
  package (GO-2026-6859).

## 0.3.0 (2026-11-05)

### New

- Install with Homebrew on macOS and Linux (`brew install commerce-cloud-integrations/tap/epx`,
  with man pages and shell completion) and with Scoop on Windows (bucket
  `commerce-cloud-integrations/scoop-bucket`). Every release carries the formula and the
  manifest as `epx.rb` and `epx.json`.
- Man pages: `man epx` and one page per top-level command, such as `man epx-products`, with
  the same text as `--help`. Homebrew, the install script and npm install them; the release
  archives for macOS and Linux hold them in `man/`.
- The [documentation site](https://commerce-cloud-integrations.github.io/emporix-cli/) is
  published with every release, with search. The repository has a contributing guide and a
  security policy, with private vulnerability reporting.

### Fixed

- `epx iam groups list` sorts groups by ID unless you pass `--sort`: Emporix orders groups
  differently on every request, so on a tenant with many groups a later page could repeat a
  group and the list failed.

### Changed

- `epx iam groups delete` says that Emporix never deletes a system group (CUSTOMER, BO_USER,
  BO_MANAGER, BO_ADMIN, SUPPORT_USER), not even with `--force-delete`.
- epx is checked against the Emporix API reference of 2026-11-05.
- Command help wraps its paragraphs and flag descriptions to 95 columns and never splits a
  short literal, such as `-o json`, across lines. The words of every help are unchanged.

## 0.2.0 (2026-11-05)

epx follows the Emporix API reference of 2026-11-04: it reads and writes every operation the
reference documents that applies to a service token.

### Breaking changes

Emporix removed these IAM endpoints on 2026-11-03, so epx removes the commands that read them:

| Removed | Use instead |
| --- | --- |
| `epx iam users resource-access-controls USER_ID RESOURCE_ID` | `epx iam users access-controls USER_ID` |
| `epx iam users resource-permissions USER_ID RESOURCE_ID` | `epx iam users scopes USER_ID` |
| `epx iam roles list\|get`, `epx iam permissions list\|get`, `epx iam resources list\|get` | Emporix names no replacement: access controls are now collections of scopes (`epx iam access-controls list`, `epx iam scopes list`) |

The tables of `iam access-controls list|get`, `iam groups access-controls` and
`iam users access-controls` no longer have the RESOURCE and ROLE columns: Emporix no longer
sends them.

`epx imports runs streams` has three new counter columns (UNRESOLVED, COLLAPSED, EMBEDDED)
before MESSAGE, so in `-o csv` MESSAGE moves from column 9 to column 12. SOURCE of
`imports streams get` and `imports configurations streams` now names the file mode of a
file-based stream (for example `STOCK INCREMENTAL files`, or `SNAPSHOT files` for a stream
without a source entity), and a parent's sweep row shows its stream as `NAME sweep`.

### New

- `epx search documents|indexes|queries` take the types `customer` and `legal-entity`;
  `search documents` requests `customer.customer_read` or
  `customermanagement.legalentity_read` for them.
- `epx imports streams files list|limits|upload|order|delete`: the file queue of a
  file-based import stream.
- `epx products list|get|search` and `epx custom-entities instances list|get|search` take
  `--mixin-site SITE`, which merges a site's mixin overlays into the base mixins.
- `epx category-trees list --type STANDARD|CLASSIFICATION`.
- `epx imports runs streams` shows every counter that adds up to READ; `imports runs errors`
  names the codes CONFIGURATION, EXTERNAL_SYSTEM and REJECTED_BY_SOURCE.
- `epx ai rag fields|filter-fields|reindex` and `epx indexing reindex-jobs create
  --entity-type` name and complete CUSTOMER and LEGAL-ENTITY; `ai rag filter-fields` shows
  PARENT; `epx ai tools` names the `search` tool type.

### Releases

- The release archives are reproducible, and the npm package is published once every
  platform version is served.

## 0.1.4 (2026-10-22)

- Releases are built and published by GitHub Actions when a version tag is pushed; the npm
  package carries provenance.
- The npm package links the documentation site as its homepage.

## 0.1.3 (2026-10-22)

- One-line installers for macOS, Linux and Windows (`install.sh`, `install.ps1`), attached
  to every release.
- The npm package is `@commercecloudintegrations/epx-cli`.

## 0.1.2 (2026-10-22)

- The npm platform binaries are versions of the epx package itself, so npm installs only the
  binary of your platform.

## 0.1.1 (2026-10-22)

- npm platform packages published under the `@commercecloudintegrations` scope.

## 0.1.0 (2026-10-16)

The first release: reads of nearly every Emporix resource as a table, CSV, JSON, YAML or JSON
Lines, writes with `--dry-run` and confirmation, `epx api` for any endpoint, profiles with
client secrets in the OS keychain, and static binaries for macOS, Linux and Windows on x64
and arm64.

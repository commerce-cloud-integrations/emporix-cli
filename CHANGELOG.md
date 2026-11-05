# Changelog

All notable changes to epx are listed here. epx follows
[semantic versioning](https://semver.org): while the major version is 0, a minor version can
change or remove commands, columns or flags, and its notes list each such change under
**Breaking changes**.

## Unreleased

### Fixed

- `epx iam groups list` sorts groups by ID unless you pass `--sort`: Emporix orders groups
  differently on every request, so on a tenant with many groups a later page could repeat a
  group and the list failed.

### Changed

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

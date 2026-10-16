# Vendor Service

> Which Vendor Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/vendor-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/vendor-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2025-11-14 ([`a0af6e5`](https://github.com/emporix/api-references/commit/a0af6e5))
- Version: not stated in the reference file
- Operations: 11 (11 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /vendor/{tenant}/vendors` | Built | [`epx vendors create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-create) |
| `GET /vendor/{tenant}/vendors` | Built | [`epx vendors list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-list) |
| `POST /vendor/{tenant}/vendors/search` | Built | [`epx vendors search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-search) |
| `GET /vendor/{tenant}/vendors/{vendorId}` | Built | [`epx vendors get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-get) |
| `PUT /vendor/{tenant}/vendors/{vendorId}` | Built | [`epx vendors update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-update) |
| `DELETE /vendor/{tenant}/vendors/{vendorId}` | Built | [`epx vendors delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendors#epx-vendors-delete) |
| `POST /vendor/{tenant}/locations` | Built | [`epx vendor-locations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations#epx-vendor-locations-create) |
| `GET /vendor/{tenant}/locations` | Built | [`epx vendor-locations list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations#epx-vendor-locations-list) |
| `GET /vendor/{tenant}/locations/{locationId}` | Built | [`epx vendor-locations get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations#epx-vendor-locations-get) |
| `PUT /vendor/{tenant}/locations/{locationId}` | Built | [`epx vendor-locations update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations#epx-vendor-locations-update) |
| `DELETE /vendor/{tenant}/locations/{locationId}` | Built | [`epx vendor-locations delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/vendor-locations#epx-vendor-locations-delete) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

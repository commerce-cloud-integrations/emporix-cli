# Schema Service

> Which Schema Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`utilities/schema/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/schema/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-12 ([`f0072a6`](https://github.com/emporix/api-references/commit/f0072a6))
- Version: not stated in the reference file
- Operations: 31 (31 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /schema/{tenant}/schemas` | Built | [`epx schemas create`](/emporix-cli/commands/schemas#epx-schemas-create) |
| `GET /schema/{tenant}/schemas` | Built | [`epx schemas list`](/emporix-cli/commands/schemas#epx-schemas-list) |
| `GET /schema/{tenant}/schemas/{id}` | Built | [`epx schemas get`](/emporix-cli/commands/schemas#epx-schemas-get) |
| `DELETE /schema/{tenant}/schemas/{id}` | Built | [`epx schemas delete`](/emporix-cli/commands/schemas#epx-schemas-delete) |
| `PUT /schema/{tenant}/schemas/{id}` | Built | [`epx schemas update`](/emporix-cli/commands/schemas#epx-schemas-update) |
| `PUT /schema/{tenant}/schemas/{id}/types` | Built | [`epx schemas update-types`](/emporix-cli/commands/schemas#epx-schemas-update-types) |
| `POST /schema/{tenant}/schemas/file` | Built | [`epx schemas parse`](/emporix-cli/commands/schemas#epx-schemas-parse) |
| `GET /schema/{tenant}/types` | Built | [`epx schemas types`](/emporix-cli/commands/schemas#epx-schemas-types) |
| `POST /schema/{tenant}/references` | Built | [`epx schemas references create`](/emporix-cli/commands/schemas#epx-schemas-references-create) |
| `GET /schema/{tenant}/references` | Built | [`epx schemas references list`](/emporix-cli/commands/schemas#epx-schemas-references-list) |
| `GET /schema/{tenant}/references/{id}` | Built | [`epx schemas references get`](/emporix-cli/commands/schemas#epx-schemas-references-get) |
| `DELETE /schema/{tenant}/references/{id}` | Built | [`epx schemas references delete`](/emporix-cli/commands/schemas#epx-schemas-references-delete) |
| `PUT /schema/{tenant}/references/{id}` | Built | [`epx schemas references update`](/emporix-cli/commands/schemas#epx-schemas-references-update) |
| `POST /schema/{tenant}/custom-entities` | Built | [`epx custom-entities create`](/emporix-cli/commands/custom-entities#epx-custom-entities-create) |
| `GET /schema/{tenant}/custom-entities` | Built | [`epx custom-entities list`](/emporix-cli/commands/custom-entities#epx-custom-entities-list) |
| `GET /schema/{tenant}/custom-entities/{id}` | Built | [`epx custom-entities get`](/emporix-cli/commands/custom-entities#epx-custom-entities-get) |
| `DELETE /schema/{tenant}/custom-entities/{id}` | Built | [`epx custom-entities delete`](/emporix-cli/commands/custom-entities#epx-custom-entities-delete) |
| `PUT /schema/{tenant}/custom-entities/{id}` | Built | [`epx custom-entities update`](/emporix-cli/commands/custom-entities#epx-custom-entities-update) |
| `POST /schema/{tenant}/custom-entities/{type}/instances` | Built | [`epx custom-entities instances create`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-create) |
| `GET /schema/{tenant}/custom-entities/{type}/instances` | Built | [`epx custom-entities instances list`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-list) |
| `GET /schema/{tenant}/custom-entities/{type}/instances/{id}` | Built | [`epx custom-entities instances get`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-get) |
| `DELETE /schema/{tenant}/custom-entities/{type}/instances/{id}` | Built | [`epx custom-entities instances delete`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-delete) |
| `PUT /schema/{tenant}/custom-entities/{type}/instances/{id}` | Built | [`epx custom-entities instances update`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-update) |
| `PATCH /schema/{tenant}/custom-entities/{type}/instances/{id}` | Built | [`epx custom-entities instances patch`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-patch) |
| `POST /schema/{tenant}/custom-entities/{type}/instances/search` | Built | [`epx custom-entities instances search`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-search) |
| `PATCH /schema/{tenant}/custom-entities/{type}/instances/bulk` | Built | [`epx custom-entities instances bulk-patch`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-bulk-patch) |
| `POST /schema/{tenant}/custom-entities/{type}/instances/bulk` | Built | [`epx custom-entities instances bulk-create`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-bulk-create) |
| `PUT /schema/{tenant}/custom-entities/{type}/instances/bulk` | Built | [`epx custom-entities instances bulk-update`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-bulk-update) |
| `DELETE /schema/{tenant}/custom-entities/{type}/instances/bulk` | Built | [`epx custom-entities instances bulk-delete`](/emporix-cli/commands/custom-entities#epx-custom-entities-instances-bulk-delete) |
| `POST /schema/{tenant}/custom-entities/export` | Built | [`epx custom-entities export`](/emporix-cli/commands/custom-entities#epx-custom-entities-export) |
| `POST /schema/{tenant}/custom-entities/import` | Built | [`epx custom-entities import`](/emporix-cli/commands/custom-entities#epx-custom-entities-import) |

All services: [Emporix API coverage](/emporix-cli/coverage).

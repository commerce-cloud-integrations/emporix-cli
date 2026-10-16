# Sequential ID Service

> Which Sequential ID Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`utilities/sequential-id/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/sequential-id/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee))
- Version: not stated in the reference file
- Operations: 8 (8 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /sequential-id/{tenant}/schemas` | Built | [`epx sequential-ids list`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-list) |
| `POST /sequential-id/{tenant}/schemas` | Built | [`epx sequential-ids create`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-create) |
| `GET /sequential-id/{tenant}/schemas/{schemaId}` | Built | [`epx sequential-ids get`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-get) |
| `DELETE /sequential-id/{tenant}/schemas/{schemaId}` | Built | [`epx sequential-ids delete`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-delete) |
| `POST /sequential-id/{tenant}/schemas/{schemaId}/setActive` | Built | [`epx sequential-ids activate`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-activate) |
| `GET /sequential-id/{tenant}/schemas/types/{schemaType}` | Built | [`epx sequential-ids for-type`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-for-type) |
| `POST /sequential-id/{tenant}/schemas/types/{schemaType}/nextId` | Built | [`epx sequential-ids next-id`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-next-id) |
| `POST /sequential-id/sequenceSchemaBatch/nextIds` | Built | [`epx sequential-ids next-ids`](/emporix-cli/commands/sequential-ids#epx-sequential-ids-next-ids) |

All services: [Emporix API coverage](/emporix-cli/coverage).

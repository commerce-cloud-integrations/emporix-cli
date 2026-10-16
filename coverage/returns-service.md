# Returns Service

> Which Returns Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`orders/returns/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/returns/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-03-20 ([`b658888`](https://github.com/emporix/api-references/commit/b658888))
- Version: not stated in the reference file
- Operations: 6 (6 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /return/{tenant}/returns` | Built | [`epx returns create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-create) |
| `GET /return/{tenant}/returns` | Built | [`epx returns list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-list) |
| `GET /return/{tenant}/returns/{returnId}` | Built | [`epx returns get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-get) |
| `PUT /return/{tenant}/returns/{returnId}` | Built | [`epx returns update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-update) |
| `DELETE /return/{tenant}/returns/{returnId}` | Built | [`epx returns delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-delete) |
| `PATCH /return/{tenant}/returns/{returnId}` | Built | [`epx returns patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/returns#epx-returns-patch) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

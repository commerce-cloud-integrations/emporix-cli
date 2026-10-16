# Label Service

> Which Label Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`products-labels-and-brands/label-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/products-labels-and-brands/label-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 6 (6 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /label/labels` | Built | [`epx labels list`](/emporix-cli/commands/labels#epx-labels-list) |
| `POST /label/labels` | Built | [`epx labels create`](/emporix-cli/commands/labels#epx-labels-create) |
| `GET /label/labels/{labelId}` | Built | [`epx labels get`](/emporix-cli/commands/labels#epx-labels-get) |
| `PUT /label/labels/{labelId}` | Built | [`epx labels update`](/emporix-cli/commands/labels#epx-labels-update) |
| `PATCH /label/labels/{labelId}` | Built | [`epx labels patch`](/emporix-cli/commands/labels#epx-labels-patch) |
| `DELETE /label/labels/{labelId}` | Built | [`epx labels delete`](/emporix-cli/commands/labels#epx-labels-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).

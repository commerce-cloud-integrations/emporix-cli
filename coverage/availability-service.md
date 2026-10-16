# Availability Service

> Which Availability Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`orders/availability/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/availability/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`84cc3ec`](https://github.com/emporix/api-references/commit/84cc3ec))
- Version: not stated in the reference file
- Operations: 14 (9 built, 5 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /availability/{tenant}/locations/{site}` | Not applicable | Deprecated by Emporix, removed on 1 December 2026. |
| `POST /availability/{tenant}/locations/{site}` | Not applicable | Deprecated by Emporix, removed on 1 December 2026. |
| `PUT /availability/{tenant}/locations/{site}` | Not applicable | Deprecated by Emporix, removed on 1 December 2026. |
| `DELETE /availability/{tenant}/locations/{location}` | Not applicable | Deprecated by Emporix, removed on 1 December 2026. |
| `POST /availability/{tenant}/search/locations` | Not applicable | Deprecated by Emporix, removed on 1 December 2026. |
| `GET /availability/{tenant}/availability/site/{site}` | Built | [`epx availability list`](/emporix-cli/commands/availability#epx-availability-list) |
| `POST /availability/{tenant}/availability/search` | Built | [`epx availability search`](/emporix-cli/commands/availability#epx-availability-search) |
| `GET /availability/{tenant}/availability/{productId}/{site}` | Built | [`epx availability get`](/emporix-cli/commands/availability#epx-availability-get) |
| `POST /availability/{tenant}/availability/{productId}/{site}` | Built | [`epx availability create`](/emporix-cli/commands/availability#epx-availability-create) |
| `PUT /availability/{tenant}/availability/{productId}/{site}` | Built | [`epx availability update`](/emporix-cli/commands/availability#epx-availability-update) |
| `DELETE /availability/{tenant}/availability/{productId}/{site}` | Built | [`epx availability delete`](/emporix-cli/commands/availability#epx-availability-delete) |
| `POST /availability/{tenant}/availability/bulk` | Built | [`epx availability bulk-create`](/emporix-cli/commands/availability#epx-availability-bulk-create) |
| `PUT /availability/{tenant}/availability/bulk` | Built | [`epx availability bulk-update`](/emporix-cli/commands/availability#epx-availability-bulk-update) |
| `DELETE /availability/{tenant}/availability/bulk` | Built | [`epx availability bulk-delete`](/emporix-cli/commands/availability#epx-availability-bulk-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).

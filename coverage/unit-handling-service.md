# Unit Handling Service

> Which Unit Handling Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/unit-handling-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/unit-handling-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-17 ([`ce5a5d6`](https://github.com/emporix/api-references/commit/ce5a5d6))
- Version: not stated in the reference file
- Operations: 9 (9 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /unit-handling/{tenant}/units` | Built | [`epx units create`](/emporix-cli/commands/units#epx-units-create) |
| `GET /unit-handling/{tenant}/units` | Built | [`epx units list`](/emporix-cli/commands/units#epx-units-list) |
| `DELETE /unit-handling/{tenant}/units` | Built | [`epx units bulk-delete`](/emporix-cli/commands/units#epx-units-bulk-delete) |
| `GET /unit-handling/{tenant}/units/{unitCode}` | Built | [`epx units get`](/emporix-cli/commands/units#epx-units-get) |
| `DELETE /unit-handling/{tenant}/units/{unitCode}` | Built | [`epx units delete`](/emporix-cli/commands/units#epx-units-delete) |
| `PUT /unit-handling/{tenant}/units/{unitCode}` | Built | [`epx units update`](/emporix-cli/commands/units#epx-units-update) |
| `PUT /unit-handling/{tenant}/units/conversion-factor-commands` | Built | [`epx units conversion-factor`](/emporix-cli/commands/units#epx-units-conversion-factor) |
| `PUT /unit-handling/{tenant}/units/convert-unit-commands` | Built | [`epx units convert`](/emporix-cli/commands/units#epx-units-convert) |
| `GET /unit-handling/{tenant}/types` | Built | [`epx units types`](/emporix-cli/commands/units#epx-units-types) |

All services: [Emporix API coverage](/emporix-cli/coverage).

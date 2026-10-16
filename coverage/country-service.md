# Country Service

> Which Country Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/country-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/country-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-06-30 ([`1e29a43`](https://github.com/emporix/api-references/commit/1e29a43))
- Version: not stated in the reference file
- Operations: 5 (5 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /country/{tenant}/countries` | Built | [`epx countries list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/countries#epx-countries-list) |
| `GET /country/{tenant}/countries/{countryCode}` | Built | [`epx countries get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/countries#epx-countries-get) |
| `PATCH /country/{tenant}/countries/{countryCode}` | Built | [`epx countries patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/countries#epx-countries-patch) |
| `GET /country/{tenant}/regions` | Built | [`epx regions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/regions#epx-regions-list) |
| `GET /country/{tenant}/regions/{regionCode}` | Built | [`epx regions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/regions#epx-regions-get) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

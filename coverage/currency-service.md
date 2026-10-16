# Currency Service

> Which Currency Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/currency-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/currency-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-17 ([`2e76d42`](https://github.com/emporix/api-references/commit/2e76d42))
- Version: not stated in the reference file
- Operations: 10 (10 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /currency/{tenant}/currencies` | Built | [`epx currencies create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-create) |
| `GET /currency/{tenant}/currencies` | Built | [`epx currencies list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-list) |
| `GET /currency/{tenant}/currencies/{currencyCode}` | Built | [`epx currencies get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-get) |
| `PUT /currency/{tenant}/currencies/{currencyCode}` | Built | [`epx currencies update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-update) |
| `DELETE /currency/{tenant}/currencies/{currencyCode}` | Built | [`epx currencies delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/currencies#epx-currencies-delete) |
| `POST /currency/{tenant}/exchanges` | Built | [`epx exchange-rates create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-create) |
| `GET /currency/{tenant}/exchanges` | Built | [`epx exchange-rates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-list) |
| `GET /currency/{tenant}/exchanges/{code}` | Built | [`epx exchange-rates get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-get) |
| `PUT /currency/{tenant}/exchanges/{code}` | Built | [`epx exchange-rates update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-update) |
| `DELETE /currency/{tenant}/exchanges/{code}` | Built | [`epx exchange-rates delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/exchange-rates#epx-exchange-rates-delete) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

# Configuration Service

> Which Configuration Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/configuration-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/configuration-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-01-28 ([`9015ed5`](https://github.com/emporix/api-references/commit/9015ed5))
- Version: not stated in the reference file
- Operations: 12 (12 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /configuration/{tenant}/global-configurations` | Built | [`epx configurations global`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-global) |
| `GET /configuration/{tenant}/configurations` | Built | [`epx configurations list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-list) |
| `POST /configuration/{tenant}/configurations` | Built | [`epx configurations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-create) |
| `GET /configuration/{tenant}/configurations/{propertyKey}` | Built | [`epx configurations get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-get) |
| `PUT /configuration/{tenant}/configurations/{propertyKey}` | Built | [`epx configurations update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-update) |
| `DELETE /configuration/{tenant}/configurations/{propertyKey}` | Built | [`epx configurations delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-delete) |
| `GET /configuration/{tenant}/clients` | Built | [`epx configurations clients`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-clients) |
| `GET /configuration/{tenant}/clients/{client}/configurations` | Built | [`epx configurations for-client`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-for-client) |
| `POST /configuration/{tenant}/clients/{client}/configurations` | Built | [`epx configurations client-create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-create) |
| `GET /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | Built | [`epx configurations client-setting`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-setting) |
| `PUT /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | Built | [`epx configurations client-update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-update) |
| `DELETE /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | Built | [`epx configurations client-delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-delete) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

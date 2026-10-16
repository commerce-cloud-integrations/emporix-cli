# Site Settings Service

> Which Site Settings Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/site-settings-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/site-settings-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-09-28 ([`7c702b5`](https://github.com/emporix/api-references/commit/7c702b5))
- Version: not stated in the reference file
- Operations: 13 (13 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /site/{tenant}/sites` | Built | [`epx sites list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-list) |
| `POST /site/{tenant}/sites` | Built | [`epx sites create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-create) |
| `GET /site/{tenant}/sites/{siteCode}` | Built | [`epx sites get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-get) |
| `PUT /site/{tenant}/sites/{siteCode}` | Built | [`epx sites update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-update) |
| `DELETE /site/{tenant}/sites/{siteCode}` | Built | [`epx sites delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-delete) |
| `PATCH /site/{tenant}/sites/{siteCode}` | Built | [`epx sites patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-patch) |
| `GET /site/{tenant}/siteslist` | Built | [`epx sites codes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-codes) |
| `GET /site/{tenant}/sites/{siteCode}/mixins` | Built | [`epx sites mixins list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-list) |
| `POST /site/{tenant}/sites/{siteCode}/mixins` | Built | [`epx sites mixins create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-create) |
| `GET /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | Built | [`epx sites mixins get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-get) |
| `DELETE /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | Built | [`epx sites mixins delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-delete) |
| `PUT /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | Built | [`epx sites mixins update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-update) |
| `PATCH /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | Built | [`epx sites mixins patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/sites#epx-sites-mixins-patch) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

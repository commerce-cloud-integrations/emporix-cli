# Indexing Service

> Which Indexing Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`configuration/indexing-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/indexing-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee))
- Version: `v1`
- Operations: 11 (11 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /indexing/{tenant}/configurations` | Built | [`epx indexing providers create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-create) |
| `GET /indexing/{tenant}/configurations` | Built | [`epx indexing providers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-list) |
| `GET /indexing/{tenant}/configurations/{provider}` | Built | [`epx indexing providers get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-get) |
| `PUT /indexing/{tenant}/configurations/{provider}` | Built | [`epx indexing providers update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-update) |
| `DELETE /indexing/{tenant}/configurations/{provider}` | Built | [`epx indexing providers delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-delete) |
| `GET /indexing/{tenant}/public/configurations` | Built | [`epx indexing providers public list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-public-list) |
| `GET /indexing/{tenant}/public/configurations/{provider}` | Built | [`epx indexing providers public get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-providers-public-get) |
| `POST /indexing/{tenant}/reindex` | Built | [`epx indexing reindex`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex) |
| `POST /indexing/{tenant}/reindex-jobs` | Built | [`epx indexing reindex-jobs create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-create) |
| `GET /indexing/{tenant}/reindex-jobs` | Built | [`epx indexing reindex-jobs list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-list) |
| `GET /indexing/{tenant}/reindex-jobs/{reindexJobId}` | Built | [`epx indexing reindex-jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-get) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

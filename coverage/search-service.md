# Search Service

> Which Search Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`utilities/search-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/search-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-09 ([`8fca29b`](https://github.com/emporix/api-references/commit/8fca29b))
- Version: not stated in the reference file
- Operations: 15 (15 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /search/{tenant}/search/{type}` | Built | [`epx search documents`](/emporix-cli/commands/search#epx-search-documents) |
| `GET /search/{tenant}/search/queries` | Built | [`epx search queries list`](/emporix-cli/commands/search#epx-search-queries-list) |
| `GET /search/{tenant}/search/{type}/queries` | Built | [`epx search queries for-type`](/emporix-cli/commands/search#epx-search-queries-for-type) |
| `PUT /search/{tenant}/search/{type}/queries/{id}` | Built | [`epx search queries update`](/emporix-cli/commands/search#epx-search-queries-update) |
| `GET /search/{tenant}/search/{type}/queries/{id}` | Built | [`epx search queries get`](/emporix-cli/commands/search#epx-search-queries-get) |
| `DELETE /search/{tenant}/search/{type}/queries/{id}` | Built | [`epx search queries delete`](/emporix-cli/commands/search#epx-search-queries-delete) |
| `GET /search/{tenant}/search/indexes` | Built | [`epx search indexes list`](/emporix-cli/commands/search#epx-search-indexes-list) |
| `POST /search/{tenant}/search/indexes/export` | Built | [`epx search indexes export`](/emporix-cli/commands/search#epx-search-indexes-export) |
| `POST /search/{tenant}/search/indexes/import` | Built | [`epx search indexes import`](/emporix-cli/commands/search#epx-search-indexes-import) |
| `GET /search/{tenant}/search/{type}/indexes` | Built | [`epx search indexes for-type`](/emporix-cli/commands/search#epx-search-indexes-for-type) |
| `PUT /search/{tenant}/search/{type}/indexes/{id}` | Built | [`epx search indexes update`](/emporix-cli/commands/search#epx-search-indexes-update) |
| `GET /search/{tenant}/search/{type}/indexes/{id}` | Built | [`epx search indexes get`](/emporix-cli/commands/search#epx-search-indexes-get) |
| `DELETE /search/{tenant}/search/{type}/indexes/{id}` | Built | [`epx search indexes delete`](/emporix-cli/commands/search#epx-search-indexes-delete) |
| `GET /search/{tenant}/jobs` | Built | [`epx search jobs list`](/emporix-cli/commands/search#epx-search-jobs-list) |
| `GET /search/{tenant}/jobs/{id}` | Built | [`epx search jobs get`](/emporix-cli/commands/search#epx-search-jobs-get) |

All services: [Emporix API coverage](/emporix-cli/coverage).

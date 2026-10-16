# Catalog Management

> Which Catalog Management operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`catalogs-and-categories/catalog/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/catalogs-and-categories/catalog/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-16 ([`6d06d44`](https://github.com/emporix/api-references/commit/6d06d44))
- Version: `1.0.0`
- Operations: 7 (7 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /catalog/{tenant}/catalogs` | Built | [`epx catalogs create`](/emporix-cli/commands/catalogs#epx-catalogs-create) |
| `GET /catalog/{tenant}/catalogs` | Built | [`epx catalogs list`](/emporix-cli/commands/catalogs#epx-catalogs-list) |
| `GET /catalog/{tenant}/catalogs/categories/{categoryId}` | Built | [`epx catalogs for-category`](/emporix-cli/commands/catalogs#epx-catalogs-for-category) |
| `GET /catalog/{tenant}/catalogs/{catalogId}` | Built | [`epx catalogs get`](/emporix-cli/commands/catalogs#epx-catalogs-get) |
| `PUT /catalog/{tenant}/catalogs/{catalogId}` | Built | [`epx catalogs update`](/emporix-cli/commands/catalogs#epx-catalogs-update) |
| `DELETE /catalog/{tenant}/catalogs/{catalogId}` | Built | [`epx catalogs delete`](/emporix-cli/commands/catalogs#epx-catalogs-delete) |
| `PATCH /catalog/{tenant}/catalogs/{catalogId}` | Built | [`epx catalogs patch`](/emporix-cli/commands/catalogs#epx-catalogs-patch) |

All services: [Emporix API coverage](/emporix-cli/coverage).

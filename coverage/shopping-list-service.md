# Shopping List Service

> Which Shopping List Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`checkout/shopping-list/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/shopping-list/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee))
- Version: not stated in the reference file
- Operations: 5 (5 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /shoppinglist/{tenant}/shopping-lists` | Built | [`epx shopping-lists create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-create) |
| `GET /shoppinglist/{tenant}/shopping-lists` | Built | [`epx shopping-lists list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-list) |
| `PUT /shoppinglist/{tenant}/shopping-lists/{customerId}` | Built | [`epx shopping-lists update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-update) |
| `GET /shoppinglist/{tenant}/shopping-lists/{customerId}` | Built | [`epx shopping-lists get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-get) |
| `DELETE /shoppinglist/{tenant}/shopping-lists/{customerId}` | Built | [`epx shopping-lists delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-delete), [`epx shopping-lists delete-all`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/shopping-lists#epx-shopping-lists-delete-all) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

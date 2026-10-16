# Cart Service

> Which Cart Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`checkout/cart/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/cart/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-07 ([`1f21665`](https://github.com/emporix/api-references/commit/1f21665))
- Version: not stated in the reference file
- Operations: 25 (25 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /cart/{tenant}/carts` | Built | [`epx carts create`](/emporix-cli/commands/carts#epx-carts-create) |
| `GET /cart/{tenant}/carts` | Built | [`epx carts find`](/emporix-cli/commands/carts#epx-carts-find) |
| `POST /cart/{tenant}/carts/search` | Built | [`epx carts search`](/emporix-cli/commands/carts#epx-carts-search) |
| `GET /cart/{tenant}/carts/{cartId}` | Built | [`epx carts get`](/emporix-cli/commands/carts#epx-carts-get) |
| `PUT /cart/{tenant}/carts/{cartId}` | Built | [`epx carts update`](/emporix-cli/commands/carts#epx-carts-update) |
| `DELETE /cart/{tenant}/carts/{cartId}` | Built | [`epx carts delete`](/emporix-cli/commands/carts#epx-carts-delete) |
| `POST /cart/{tenant}/carts/{cartId}/execute` | Built | [`epx carts execute`](/emporix-cli/commands/carts#epx-carts-execute) |
| `PUT /cart/{tenant}/carts/{cartId}/refresh` | Built | [`epx carts refresh`](/emporix-cli/commands/carts#epx-carts-refresh) |
| `POST /cart/{tenant}/carts/{cartId}/changeSite` | Built | [`epx carts change-site`](/emporix-cli/commands/carts#epx-carts-change-site) |
| `POST /cart/{tenant}/carts/{cartId}/changeCurrency` | Built | [`epx carts change-currency`](/emporix-cli/commands/carts#epx-carts-change-currency) |
| `POST /cart/{tenant}/carts/{cartId}/merge` | Built | [`epx carts merge`](/emporix-cli/commands/carts#epx-carts-merge) |
| `POST /cart/{tenant}/carts/{cartId}/itemsBatch` | Built | [`epx carts items bulk-create`](/emporix-cli/commands/carts#epx-carts-items-bulk-create) |
| `PUT /cart/{tenant}/carts/{cartId}/itemsBatch` | Built | [`epx carts items bulk-update`](/emporix-cli/commands/carts#epx-carts-items-bulk-update) |
| `GET /cart/{tenant}/carts/{cartId}/items` | Built | [`epx carts items list`](/emporix-cli/commands/carts#epx-carts-items-list) |
| `POST /cart/{tenant}/carts/{cartId}/items` | Built | [`epx carts items create`](/emporix-cli/commands/carts#epx-carts-items-create) |
| `DELETE /cart/{tenant}/carts/{cartId}/items` | Built | [`epx carts items delete-all`](/emporix-cli/commands/carts#epx-carts-items-delete-all) |
| `GET /cart/{tenant}/carts/{cartId}/items/{itemId}` | Built | [`epx carts items get`](/emporix-cli/commands/carts#epx-carts-items-get) |
| `PUT /cart/{tenant}/carts/{cartId}/items/{itemId}` | Built | [`epx carts items update`](/emporix-cli/commands/carts#epx-carts-items-update) |
| `DELETE /cart/{tenant}/carts/{cartId}/items/{itemId}` | Built | [`epx carts items delete`](/emporix-cli/commands/carts#epx-carts-items-delete) |
| `POST /cart/{tenant}/carts/{cartId}/discounts` | Built | [`epx carts discounts create`](/emporix-cli/commands/carts#epx-carts-discounts-create) |
| `DELETE /cart/{tenant}/carts/{cartId}/discounts` | Built | [`epx carts discounts delete-all`](/emporix-cli/commands/carts#epx-carts-discounts-delete-all) |
| `GET /cart/{tenant}/carts/{cartId}/discounts` | Built | [`epx carts discounts list`](/emporix-cli/commands/carts#epx-carts-discounts-list) |
| `DELETE /cart/{tenant}/carts/{cartId}/discounts/{discountIndex}` | Built | [`epx carts discounts delete`](/emporix-cli/commands/carts#epx-carts-discounts-delete) |
| `GET /cart/{tenant}/carts/{cartId}/dtRestrictions` | Built | [`epx carts delivery-restrictions`](/emporix-cli/commands/carts#epx-carts-delivery-restrictions) |
| `GET /cart/{tenant}/carts/{cartId}/validate` | Built | [`epx carts validate`](/emporix-cli/commands/carts#epx-carts-validate) |

All services: [Emporix API coverage](/emporix-cli/coverage).

# Order Service

> Which Order Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`orders/order/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/order/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-07 ([`1f21665`](https://github.com/emporix/api-references/commit/1f21665))
- Version: not stated in the reference file
- Operations: 20 (14 built, 6 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /order-v2/{tenant}/salesorders` | Built | [`epx orders list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-list) |
| `POST /order-v2/{tenant}/salesorders` | Built | [`epx orders create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-create) |
| `HEAD /order-v2/{tenant}/salesorders` | Built | [`epx orders count`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-count) |
| `POST /order-v2/{tenant}/salesorders/search` | Built | [`epx orders search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-search) |
| `GET /order-v2/{tenant}/salesorders/{orderId}` | Built | [`epx orders get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-get) |
| `PUT /order-v2/{tenant}/salesorders/{orderId}` | Built | [`epx orders update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-update) |
| `PATCH /order-v2/{tenant}/salesorders/{orderId}` | Built | [`epx orders patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-patch) |
| `DELETE /order-v2/{tenant}/salesorders/{orderId}` | Built | [`epx orders delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-delete) |
| `GET /order-v2/{tenant}/salesorders/{orderId}/transitions` | Built | [`epx orders transitions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-transitions) |
| `POST /order-v2/{tenant}/salesorders/{orderId}/transitions` | Built | [`epx orders transition`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-transition) |
| `GET /order-v2/{tenant}/salesorders/{orderId}/historical-transitions` | Built | [`epx orders history`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-history) |
| `POST /order-v2/{tenant}/salesorders/{orderId}/calculations` | Built | [`epx orders recalculate`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-recalculate) |
| `POST /order-v2/{tenant}/salesorders/{orderId}/entries` | Built | [`epx orders add-entries`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-add-entries) |
| `POST /order-v2/{tenant}/salesorders/{orderId}/split` | Built | [`epx orders split`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/orders#epx-orders-split) |
| `GET /order-v2/{tenant}/orders` | Not applicable | Reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders. |
| `GET /order-v2/{tenant}/orders/{orderId}` | Not applicable | Reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders. |
| `GET /order-v2/{tenant}/orders/{orderId}/transitions` | Not applicable | Reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders. |
| `POST /order-v2/{tenant}/orders/{orderId}/transitions` | Not applicable | Reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders. |
| `GET /order-v2/{tenant}/legal-entity-orders/{legalEntityId}` | Not applicable | Needs a signed-in customer's token (saas-token); `epx orders list --query legalEntityId:ID` lists the orders. |
| `GET /order-v2/{tenant}/legal-entity-orders/{legalEntityId}/{orderId}` | Not applicable | Needs a signed-in customer's token (saas-token); `epx orders list --query legalEntityId:ID` lists the orders. |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

# Fee Service

> Which Fee Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`checkout/fee/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/fee/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-16 ([`b290b28`](https://github.com/emporix/api-references/commit/b290b28))
- Version: `v1`
- Operations: 18 (18 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /fee/{tenant}/fees` | Built | [`epx fees list`](/emporix-cli/commands/fees#epx-fees-list) |
| `POST /fee/{tenant}/fees` | Built | [`epx fees create`](/emporix-cli/commands/fees#epx-fees-create) |
| `GET /fee/{tenant}/fees/{id}` | Built | [`epx fees get`](/emporix-cli/commands/fees#epx-fees-get) |
| `PUT /fee/{tenant}/fees/{id}` | Built | [`epx fees update`](/emporix-cli/commands/fees#epx-fees-update) |
| `DELETE /fee/{tenant}/fees/{id}` | Built | [`epx fees delete`](/emporix-cli/commands/fees#epx-fees-delete) |
| `GET /fee/{tenant}/itemFees` | Built | [`epx fees assignments list`](/emporix-cli/commands/fees#epx-fees-assignments-list) |
| `POST /fee/{tenant}/itemFees` | Built | [`epx fees assignments create`](/emporix-cli/commands/fees#epx-fees-assignments-create) |
| `GET /fee/{tenant}/itemFees/{itemYRN}/fees` | Built | [`epx fees for-item`](/emporix-cli/commands/fees#epx-fees-for-item) |
| `PUT /fee/{tenant}/itemFees/{itemYRN}/fees` | Built | [`epx fees assignments update`](/emporix-cli/commands/fees#epx-fees-assignments-update) |
| `DELETE /fee/{tenant}/itemFees/{itemYRN}/fees` | Built | [`epx fees assignments delete`](/emporix-cli/commands/fees#epx-fees-assignments-delete) |
| `DELETE /fee/{tenant}/itemFees/{itemYRN}/fees/{feeId}` | Built | [`epx fees assignments delete-fee`](/emporix-cli/commands/fees#epx-fees-assignments-delete-fee) |
| `GET /fee/{tenant}/productFees/{productId}/fees` | Built | [`epx fees for-product`](/emporix-cli/commands/fees#epx-fees-for-product) |
| `PUT /fee/{tenant}/productFees/{productId}/fees` | Built | [`epx fees assignments update-product`](/emporix-cli/commands/fees#epx-fees-assignments-update-product) |
| `DELETE /fee/{tenant}/productFees/{productId}/fees` | Built | [`epx fees assignments delete-product`](/emporix-cli/commands/fees#epx-fees-assignments-delete-product) |
| `DELETE /fee/{tenant}/productFees/{productId}/fees/{feeId}` | Built | [`epx fees assignments delete-product-fee`](/emporix-cli/commands/fees#epx-fees-assignments-delete-product-fee) |
| `POST /fee/{tenant}/itemFees/searchByProductId` | Built | [`epx fees assignments search-by-product`](/emporix-cli/commands/fees#epx-fees-assignments-search-by-product) |
| `POST /fee/{tenant}/itemFees/searchByProductIds` | Built | [`epx fees assignments search-by-products`](/emporix-cli/commands/fees#epx-fees-assignments-search-by-products) |
| `POST /fee/{tenant}/itemFees/search` | Built | [`epx fees assignments search`](/emporix-cli/commands/fees#epx-fees-assignments-search) |

All services: [Emporix API coverage](/emporix-cli/coverage).

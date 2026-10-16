# Price Service

> Which Price Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`prices-and-taxes/price-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/prices-and-taxes/price-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 30 (29 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /price/{tenant}/prices` | Built | [`epx prices create`](/emporix-cli/commands/prices#epx-prices-create) |
| `GET /price/{tenant}/prices` | Built | [`epx prices list`](/emporix-cli/commands/prices#epx-prices-list), [`epx prices for-product`](/emporix-cli/commands/prices#epx-prices-for-product) |
| `POST /price/{tenant}/prices/bulk` | Built | [`epx prices bulk-create`](/emporix-cli/commands/prices#epx-prices-bulk-create) |
| `PUT /price/{tenant}/prices/bulk` | Built | [`epx prices bulk-update`](/emporix-cli/commands/prices#epx-prices-bulk-update) |
| `GET /price/{tenant}/prices/{priceId}` | Built | [`epx prices get`](/emporix-cli/commands/prices#epx-prices-get) |
| `PUT /price/{tenant}/prices/{priceId}` | Built | [`epx prices update`](/emporix-cli/commands/prices#epx-prices-update) |
| `DELETE /price/{tenant}/prices/{priceId}` | Built | [`epx prices delete`](/emporix-cli/commands/prices#epx-prices-delete) |
| `POST /price/{tenant}/match-prices` | Built | [`epx prices match`](/emporix-cli/commands/prices#epx-prices-match) |
| `POST /price/{tenant}/match-prices-by-context` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /price/{tenant}/prices/search` | Built | [`epx prices search`](/emporix-cli/commands/prices#epx-prices-search) |
| `POST /price/{tenant}/priceModels` | Built | [`epx price-models create`](/emporix-cli/commands/price-models#epx-price-models-create) |
| `GET /price/{tenant}/priceModels` | Built | [`epx price-models list`](/emporix-cli/commands/price-models#epx-price-models-list) |
| `GET /price/{tenant}/priceModels/{priceModelId}` | Built | [`epx price-models get`](/emporix-cli/commands/price-models#epx-price-models-get) |
| `PUT /price/{tenant}/priceModels/{priceModelId}` | Built | [`epx price-models update`](/emporix-cli/commands/price-models#epx-price-models-update) |
| `DELETE /price/{tenant}/priceModels/{priceModelId}` | Built | [`epx price-models delete`](/emporix-cli/commands/price-models#epx-price-models-delete) |
| `GET /price/{tenant}/price-lists` | Built | [`epx price-lists list`](/emporix-cli/commands/price-lists#epx-price-lists-list) |
| `POST /price/{tenant}/price-lists` | Built | [`epx price-lists create`](/emporix-cli/commands/price-lists#epx-price-lists-create) |
| `POST /price/{tenant}/price-lists/search` | Built | [`epx price-lists search`](/emporix-cli/commands/price-lists#epx-price-lists-search) |
| `GET /price/{tenant}/price-lists/{priceListId}` | Built | [`epx price-lists get`](/emporix-cli/commands/price-lists#epx-price-lists-get) |
| `PUT /price/{tenant}/price-lists/{priceListId}` | Built | [`epx price-lists update`](/emporix-cli/commands/price-lists#epx-price-lists-update) |
| `DELETE /price/{tenant}/price-lists/{priceListId}` | Built | [`epx price-lists delete`](/emporix-cli/commands/price-lists#epx-price-lists-delete) |
| `GET /price/{tenant}/price-lists/{priceListId}/prices` | Built | [`epx price-lists prices list`](/emporix-cli/commands/price-lists#epx-price-lists-prices-list) |
| `POST /price/{tenant}/price-lists/{priceListId}/prices` | Built | [`epx price-lists prices create`](/emporix-cli/commands/price-lists#epx-price-lists-prices-create) |
| `DELETE /price/{tenant}/price-lists/{priceListId}/prices/bulk` | Built | [`epx price-lists prices bulk-delete`](/emporix-cli/commands/price-lists#epx-price-lists-prices-bulk-delete) |
| `POST /price/{tenant}/price-lists/{priceListId}/prices/bulk` | Built | [`epx price-lists prices bulk-create`](/emporix-cli/commands/price-lists#epx-price-lists-prices-bulk-create) |
| `PUT /price/{tenant}/price-lists/{priceListId}/prices/bulk` | Built | [`epx price-lists prices bulk-update`](/emporix-cli/commands/price-lists#epx-price-lists-prices-bulk-update) |
| `POST /price/{tenant}/price-lists/{priceListId}/prices/search` | Built | [`epx price-lists prices search`](/emporix-cli/commands/price-lists#epx-price-lists-prices-search) |
| `GET /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | Built | [`epx price-lists prices get`](/emporix-cli/commands/price-lists#epx-price-lists-prices-get) |
| `PUT /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | Built | [`epx price-lists prices update`](/emporix-cli/commands/price-lists#epx-price-lists-prices-update) |
| `DELETE /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | Built | [`epx price-lists prices delete`](/emporix-cli/commands/price-lists#epx-price-lists-prices-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).

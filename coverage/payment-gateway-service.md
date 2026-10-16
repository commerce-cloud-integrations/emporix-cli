# Payment-Gateway Service

> Which Payment-Gateway Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`checkout/payment-gateway/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/payment-gateway/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-02-10 ([`e61a2e2`](https://github.com/emporix/api-references/commit/e61a2e2))
- Version: not stated in the reference file
- Operations: 15 (13 built, 2 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /payment-gateway/{tenant}/paymentmodes/config` | Built | [`epx payment-modes create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-create) |
| `GET /payment-gateway/{tenant}/paymentmodes/config` | Built | [`epx payment-modes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-list) |
| `GET /payment-gateway/{tenant}/paymentmodes/config/{id}` | Built | [`epx payment-modes get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-get) |
| `PUT /payment-gateway/{tenant}/paymentmodes/config/{id}` | Built | [`epx payment-modes update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-update) |
| `DELETE /payment-gateway/{tenant}/paymentmodes/config/{id}` | Built | [`epx payment-modes delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-delete) |
| `GET /payment-gateway/{tenant}/paymentmodes/frontend` | Built | [`epx payment-modes public list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-public-list) |
| `GET /payment-gateway/{tenant}/paymentmodes/frontend/{id}` | Built | [`epx payment-modes public get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-modes#epx-payment-modes-public-get) |
| `POST /payment-gateway/{tenant}/payment/authorize` | Built | [`epx payment-transactions authorize`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-authorize) |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/capture` | Built | [`epx payment-transactions capture`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-capture) |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/refund` | Built | [`epx payment-transactions refund`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-refund) |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/cancel` | Built | [`epx payment-transactions cancel`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-cancel) |
| `POST /payment-gateway/{tenant}/payment/frontend/initialize` | Not applicable | Sent by a storefront for a customer's checkout; `epx payment-transactions authorize` authorizes the payment of an order with a service token. |
| `POST /payment-gateway/{tenant}/payment/frontend/authorize` | Not applicable | Sent by a storefront for a customer's checkout; `epx payment-transactions authorize` authorizes the payment of an order with a service token. |
| `GET /payment-gateway/{tenant}/transactions` | Built | [`epx payment-transactions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-list) |
| `GET /payment-gateway/{tenant}/transactions/{transactionId}` | Built | [`epx payment-transactions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/payment-transactions#epx-payment-transactions-get) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

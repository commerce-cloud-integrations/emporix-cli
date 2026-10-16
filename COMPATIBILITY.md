# Emporix API compatibility

Which Emporix API operations `epx` sends, checked against the Emporix API reference
([emporix/api-references](https://github.com/emporix/api-references)) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

Emporix versions services in their paths where it versions them at all (such as
`/order-v2/`), and most reference files carry no version number. Each service is
therefore identified by the last change to its reference file up to that commit, and by
`info.version` where the file states one. When a reference file changes after the date
shown, check the commands that use it before relying on them.

Each operation is listed with the epx commands that send it, or why it does not apply
to epx, which acts with a service token (an API key's client credentials), or whether
it is planned. Beyond these commands, `epx api get PATH` reads
any GET operation that answers JSON to a service token, and
`epx api post|put|patch|delete PATH` sends any write with the scopes given, before a
command of its own does.

**Coverage:** 605 of 670 operations built, 65 not applicable, 0 to do; in 43 services (and the
OAuth Service for the token).

## Summary

| Service | Reference file | Last change | Version | Operations | Built | Not applicable |
| --- | --- | --- | --- | --- | --- | --- |
| [AI Rag Indexer](#ai-rag-indexer) | `artificial-intelligence/ai-rag-indexer` | 2026-10-01 (`7fad4ee`) | — | 3 | 3 | 0 |
| [AI Service](#ai-service) | `artificial-intelligence/ai-service` | 2026-10-06 (`9b456b6`) | — | 57 | 56 | 1 |
| [OAuth Service](#oauth-service) | `authentication/oauth-service` | 2026-02-27 (`48eee2b`) | — | 1 | 1 | 0 |
| [Catalog Management](#catalog-management) | `catalogs-and-categories/catalog` | 2026-04-16 (`6d06d44`) | `1.0.0` | 7 | 7 | 0 |
| [Category](#category) | `catalogs-and-categories/category-tree` | 2026-10-01 (`de786a4`) | — | 24 | 23 | 1 |
| [Cart Service](#cart-service) | `checkout/cart` | 2026-10-07 (`1f21665`) | — | 25 | 25 | 0 |
| [Checkout Service](#checkout-service) | `checkout/checkout` | 2026-10-01 (`de786a4`) | — | 1 | 0 | 1 |
| [Fee Service](#fee-service) | `checkout/fee` | 2026-04-16 (`b290b28`) | `v1` | 18 | 18 | 0 |
| [Payment-Gateway Service](#payment-gateway-service) | `checkout/payment-gateway` | 2026-02-10 (`e61a2e2`) | — | 15 | 13 | 2 |
| [Shopping List Service](#shopping-list-service) | `checkout/shopping-list` | 2026-10-01 (`7fad4ee`) | — | 5 | 5 | 0 |
| [Approval Service](#approval-service) | `companies-and-customers/approval-service` | 2026-10-01 (`de786a4`) | — | 7 | 0 | 7 |
| [Customer Management Service](#customer-management-service) | `companies-and-customers/client-management` | 2026-04-27 (`b7c15e6`) | — | 17 | 17 | 0 |
| [Customer Service (customer-managed)](#customer-service-customer-managed) | `companies-and-customers/customer-management` | 2026-10-01 (`7fad4ee`) | — | 25 | 0 | 25 |
| [Customer Segment Service](#customer-segment-service) | `companies-and-customers/customer-segments` | 2026-10-01 (`de786a4`) | — | 36 | 34 | 2 |
| [Customer Service (tenant-managed)](#customer-service-tenant-managed) | `companies-and-customers/customer-service` | 2026-10-01 (`de786a4`) | — | 19 | 19 | 0 |
| [Vendor Service](#vendor-service) | `companies-and-customers/vendor-service` | 2025-11-14 (`a0af6e5`) | — | 11 | 11 | 0 |
| [Configuration Service](#configuration-service) | `configuration/configuration-service` | 2026-01-28 (`9015ed5`) | — | 12 | 12 | 0 |
| [Country Service](#country-service) | `configuration/country-service` | 2026-06-30 (`1e29a43`) | — | 5 | 5 | 0 |
| [Currency Service](#currency-service) | `configuration/currency-service` | 2026-04-17 (`2e76d42`) | — | 10 | 10 | 0 |
| [Indexing Service](#indexing-service) | `configuration/indexing-service` | 2026-10-01 (`7fad4ee`) | `v1` | 11 | 11 | 0 |
| [Site Settings Service](#site-settings-service) | `configuration/site-settings-service` | 2026-09-28 (`7c702b5`) | — | 13 | 13 | 0 |
| [Unit Handling Service](#unit-handling-service) | `configuration/unit-handling-service` | 2026-04-17 (`ce5a5d6`) | — | 9 | 9 | 0 |
| [Shipping Service](#shipping-service) | `delivery-and-shipping/shipping` | 2026-09-11 (`eeddb11`) | — | 45 | 44 | 1 |
| [Media Service](#media-service) | `media/media` | 2026-09-28 (`6149878`) | — | 9 | 8 | 1 |
| [Availability Service](#availability-service) | `orders/availability` | 2026-10-01 (`84cc3ec`) | — | 14 | 9 | 5 |
| [Invoice Service](#invoice-service) | `orders/invoice` | 2025-11-18 (`322697f`) | `1.0.0` | 2 | 2 | 0 |
| [Order Service](#order-service) | `orders/order` | 2026-10-07 (`1f21665`) | — | 20 | 14 | 6 |
| [Returns Service](#returns-service) | `orders/returns` | 2026-03-20 (`b658888`) | — | 6 | 6 | 0 |
| [Price Service](#price-service) | `prices-and-taxes/price-service` | 2026-10-01 (`de786a4`) | — | 30 | 29 | 1 |
| [Tax Service](#tax-service) | `prices-and-taxes/tax-service` | 2025-11-15 (`4ed37c9`) | — | 6 | 6 | 0 |
| [Brand Service](#brand-service) | `products-labels-and-brands/brand-service` | 2026-01-19 (`d7fbba2`) | — | 6 | 6 | 0 |
| [Label Service](#label-service) | `products-labels-and-brands/label-service` | 2026-10-01 (`de786a4`) | — | 6 | 6 | 0 |
| [Product Service](#product-service) | `products-labels-and-brands/product-service` | 2026-10-01 (`de786a4`) | — | 17 | 17 | 0 |
| [Quote Service](#quote-service) | `quotes/quote` | 2026-10-09 (`abf90d6`) | — | 12 | 12 | 0 |
| [Coupon Service](#coupon-service) | `rewards-and-promotions/coupon` | 2026-09-25 (`86c5409`) | — | 13 | 13 | 0 |
| [Reward Points Service](#reward-points-service) | `rewards-and-promotions/reward-points` | 2025-11-17 (`62e8b2b`) | — | 14 | 11 | 3 |
| [IAM Service](#iam-service) | `users-and-permissions/iam` | 2026-07-01 (`d42b1ce`) | — | 42 | 39 | 3 |
| [Session Context Service](#session-context-service) | `users-and-permissions/session-context` | 2026-10-01 (`7fad4ee`) | — | 8 | 4 | 4 |
| [Audit Logs (Changelog) Service](#audit-logs-changelog-service) | `utilities/audit-logs-changelog` | 2026-10-01 (`84cc3ec`) | — | 1 | 1 | 0 |
| [Import Service](#import-service) | `utilities/import-service` | 2026-10-14 (`f1f0640`) | — | 24 | 22 | 2 |
| [Schema Service](#schema-service) | `utilities/schema` | 2026-10-12 (`f0072a6`) | — | 31 | 31 | 0 |
| [Search Service](#search-service) | `utilities/search-service` | 2026-10-09 (`8fca29b`) | — | 15 | 15 | 0 |
| [Sequential ID Service](#sequential-id-service) | `utilities/sequential-id` | 2026-10-01 (`7fad4ee`) | — | 8 | 8 | 0 |
| [Webhook Service](#webhook-service) | `webhooks/webhook-service` | 2026-08-11 (`279d397`) | — | 10 | 10 | 0 |

## Operations by service

### AI Rag Indexer

[`artificial-intelligence/ai-rag-indexer/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/artificial-intelligence/ai-rag-indexer/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /ai-rag-indexer/{tenant}/{type}/reindex` | `epx ai rag reindex` |
| `GET /ai-rag-indexer/{tenant}/{type}/rag-metadata` | `epx ai rag fields` |
| `GET /ai-rag-indexer/{tenant}/{type}/filter-metadata` | `epx ai rag filter-fields` |

### AI Service

[`artificial-intelligence/ai-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/artificial-intelligence/ai-service/api-reference/api.yml), last changed 2026-10-06.

| Operation | epx |
| --- | --- |
| `POST /ai-service/{tenant}/texts` | `epx ai generate-text` |
| `POST /ai-service/{tenant}/completions` | `epx ai generate-completion` |
| `GET /ai-service/{tenant}/jobs` | `epx ai jobs list` |
| `POST /ai-service/{tenant}/jobs/search` | `epx ai jobs search` |
| `GET /ai-service/{tenant}/jobs/{jobId}` | `epx ai jobs get` |
| `DELETE /ai-service/{tenant}/jobs/{jobId}` | `epx ai jobs delete` |
| `POST /ai-service/{tenant}/agentic/chat` | `epx ai chat` |
| `POST /ai-service/{tenant}/agentic/chat-stream` | not applicable: answers with Server-Sent Events, a stream held open while the agent replies; `epx ai chat` prints the same reply whole, and `epx ai chat-async` runs it as a job |
| `POST /ai-service/{tenant}/agentic/chat-async` | `epx ai chat-async` |
| `POST /ai-service/{tenant}/agentic/{agentId}/attachments` | `epx ai agents attach` |
| `GET /ai-service/{tenant}/agentic/templates` | `epx ai templates list` |
| `POST /ai-service/{tenant}/agentic/templates/search` | `epx ai templates search` |
| `POST /ai-service/{tenant}/agentic/templates/{templateId}/agents` | `epx ai agents create` |
| `GET /ai-service/{tenant}/agentic/agents` | `epx ai agents list` |
| `POST /ai-service/{tenant}/agentic/agents/search` | `epx ai agents search` |
| `GET /ai-service/{tenant}/agentic/agents/{agentId}` | `epx ai agents get` |
| `PUT /ai-service/{tenant}/agentic/agents/{agentId}` | `epx ai agents update` |
| `PATCH /ai-service/{tenant}/agentic/agents/{agentId}` | `epx ai agents patch` |
| `DELETE /ai-service/{tenant}/agentic/agents/{agentId}` | `epx ai agents delete` |
| `GET /ai-service/{tenant}/agentic/conversations` | `epx ai conversations list` |
| `POST /ai-service/{tenant}/agentic/conversations/search` | `epx ai conversations search` |
| `GET /ai-service/{tenant}/agentic/tools` | `epx ai tools list` |
| `POST /ai-service/{tenant}/agentic/tools/search` | `epx ai tools search` |
| `GET /ai-service/{tenant}/agentic/tools/{toolId}` | `epx ai tools get` |
| `PUT /ai-service/{tenant}/agentic/tools/{toolId}` | `epx ai tools update` |
| `PATCH /ai-service/{tenant}/agentic/tools/{toolId}` | `epx ai tools patch` |
| `DELETE /ai-service/{tenant}/agentic/tools/{toolId}` | `epx ai tools delete` |
| `GET /ai-service/{tenant}/agentic/tokens` | `epx ai tokens list` |
| `POST /ai-service/{tenant}/agentic/tokens/search` | `epx ai tokens search` |
| `GET /ai-service/{tenant}/agentic/tokens/{tokenId}` | `epx ai tokens get` |
| `PUT /ai-service/{tenant}/agentic/tokens/{tokenId}` | `epx ai tokens update` |
| `PATCH /ai-service/{tenant}/agentic/tokens/{tokenId}` | `epx ai tokens patch` |
| `DELETE /ai-service/{tenant}/agentic/tokens/{tokenId}` | `epx ai tokens delete` |
| `GET /ai-service/{tenant}/agentic/oauths` | `epx ai oauths list` |
| `POST /ai-service/{tenant}/agentic/oauths/search` | `epx ai oauths search` |
| `GET /ai-service/{tenant}/agentic/oauths/{oauthId}` | `epx ai oauths get` |
| `PUT /ai-service/{tenant}/agentic/oauths/{oauthId}` | `epx ai oauths update` |
| `PATCH /ai-service/{tenant}/agentic/oauths/{oauthId}` | `epx ai oauths patch` |
| `DELETE /ai-service/{tenant}/agentic/oauths/{oauthId}` | `epx ai oauths delete` |
| `GET /ai-service/{tenant}/agentic/models` | `epx ai models` |
| `GET /ai-service/{tenant}/agentic/mcp-servers` | `epx ai mcp-servers list` |
| `POST /ai-service/{tenant}/agentic/mcp-servers/search` | `epx ai mcp-servers search` |
| `GET /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | `epx ai mcp-servers get` |
| `PUT /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | `epx ai mcp-servers update` |
| `PATCH /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | `epx ai mcp-servers patch` |
| `DELETE /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | `epx ai mcp-servers delete` |
| `GET /ai-service/{tenant}/agentic/commerce-events` | `epx ai commerce-events` |
| `POST /ai-service/{tenant}/agentic/agents/export` | `epx ai agents export` |
| `POST /ai-service/{tenant}/agentic/agents/import` | `epx ai agents import` |
| `GET /ai-service/{tenant}/agentic/logs/requests` | `epx ai logs requests list` |
| `GET /ai-service/{tenant}/agentic/logs/requests/{requestId}` | `epx ai logs requests get` |
| `POST /ai-service/{tenant}/agentic/logs/requests/search` | `epx ai logs requests search` |
| `GET /ai-service/{tenant}/agentic/logs/sessions` | `epx ai logs sessions list` |
| `GET /ai-service/{tenant}/agentic/logs/sessions/{sessionId}` | `epx ai logs sessions get` |
| `POST /ai-service/{tenant}/agentic/logs/sessions/search` | `epx ai logs sessions search` |
| `GET /ai-service/{tenant}/agentic/analytics` | `epx ai analytics tenant`, `epx ai analytics agent` |
| `GET /ai-service/{tenant}/agentic/analytics/executions` | `epx ai analytics executions` |

### OAuth Service

[`authentication/oauth-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/authentication/oauth-service/api-reference/api.yml), last changed 2026-02-27.

| Operation | epx |
| --- | --- |
| `POST /oauth/token` | every command (client credentials) |

### Catalog Management

[`catalogs-and-categories/catalog/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/catalogs-and-categories/catalog/api-reference/api.yml), last changed 2026-04-16.

| Operation | epx |
| --- | --- |
| `POST /catalog/{tenant}/catalogs` | `epx catalogs create` |
| `GET /catalog/{tenant}/catalogs` | `epx catalogs list` |
| `GET /catalog/{tenant}/catalogs/categories/{categoryId}` | `epx catalogs for-category` |
| `GET /catalog/{tenant}/catalogs/{catalogId}` | `epx catalogs get` |
| `PUT /catalog/{tenant}/catalogs/{catalogId}` | `epx catalogs update` |
| `DELETE /catalog/{tenant}/catalogs/{catalogId}` | `epx catalogs delete` |
| `PATCH /catalog/{tenant}/catalogs/{catalogId}` | `epx catalogs patch` |

### Category

[`catalogs-and-categories/category-tree/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/catalogs-and-categories/category-tree/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /category/{tenant}/assignments/references/{referenceId}` | `epx categories for-product` |
| `DELETE /category/{tenant}/assignments/references/{referenceId}` | `epx categories assignments delete-product-everywhere` |
| `PUT /category/{tenant}/categories/{categoryId}/assignments/references/bulk` | `epx categories assignments bulk-update` |
| `GET /category/{tenant}/categories` | `epx categories list` |
| `POST /category/{tenant}/categories` | `epx categories create` |
| `POST /category/{tenant}/categories/search` | `epx categories search` |
| `GET /category/{tenant}/categories/categoryTree` | not applicable: deprecated by Emporix and no longer served (Emporix answers it as an unknown category, 404); `epx category-trees` reads the category trees |
| `GET /category/{tenant}/categories/{categoryId}` | `epx categories get` |
| `PUT /category/{tenant}/categories/{categoryId}` | `epx categories update` |
| `DELETE /category/{tenant}/categories/{categoryId}` | `epx categories delete` |
| `PATCH /category/{tenant}/categories/{categoryId}` | `epx categories patch` |
| `POST /category/{tenant}/categories/{categoryId}/assignments/bulk` | `epx categories assignments bulk-create` |
| `GET /category/{tenant}/categories/{categoryId}/assignments` | `epx categories assignments list` |
| `POST /category/{tenant}/categories/{categoryId}/assignments` | `epx categories assignments create` |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments` | `epx categories assignments delete-all` |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments/{assignmentId}` | `epx categories assignments delete-assignment` |
| `PUT /category/{tenant}/categories/{categoryId}/assignments/references/{referenceId}` | `epx categories assignments update` |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments/references/{referenceId}` | `epx categories assignments delete` |
| `GET /category/{tenant}/categories/{categoryId}/parents` | `epx categories parents` |
| `GET /category/{tenant}/categories/{categoryId}/subcategories` | `epx categories subcategories` |
| `GET /category/{tenant}/category-trees` | `epx category-trees list` |
| `POST /category/{tenant}/category-trees/search` | `epx category-trees search` |
| `GET /category/{tenant}/category-trees/{categoryId}` | `epx category-trees get` |
| `POST /category/{tenant}/category-trees/{rootCategoryId}/rebuild` | `epx category-trees rebuild` |

### Cart Service

[`checkout/cart/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/cart/api-reference/api.yml), last changed 2026-10-07.

| Operation | epx |
| --- | --- |
| `POST /cart/{tenant}/carts` | `epx carts create` |
| `GET /cart/{tenant}/carts` | `epx carts find` |
| `POST /cart/{tenant}/carts/search` | `epx carts search` |
| `GET /cart/{tenant}/carts/{cartId}` | `epx carts get` |
| `PUT /cart/{tenant}/carts/{cartId}` | `epx carts update` |
| `DELETE /cart/{tenant}/carts/{cartId}` | `epx carts delete` |
| `POST /cart/{tenant}/carts/{cartId}/execute` | `epx carts execute` |
| `PUT /cart/{tenant}/carts/{cartId}/refresh` | `epx carts refresh` |
| `POST /cart/{tenant}/carts/{cartId}/changeSite` | `epx carts change-site` |
| `POST /cart/{tenant}/carts/{cartId}/changeCurrency` | `epx carts change-currency` |
| `POST /cart/{tenant}/carts/{cartId}/merge` | `epx carts merge` |
| `POST /cart/{tenant}/carts/{cartId}/itemsBatch` | `epx carts items bulk-create` |
| `PUT /cart/{tenant}/carts/{cartId}/itemsBatch` | `epx carts items bulk-update` |
| `GET /cart/{tenant}/carts/{cartId}/items` | `epx carts items list` |
| `POST /cart/{tenant}/carts/{cartId}/items` | `epx carts items create` |
| `DELETE /cart/{tenant}/carts/{cartId}/items` | `epx carts items delete-all` |
| `GET /cart/{tenant}/carts/{cartId}/items/{itemId}` | `epx carts items get` |
| `PUT /cart/{tenant}/carts/{cartId}/items/{itemId}` | `epx carts items update` |
| `DELETE /cart/{tenant}/carts/{cartId}/items/{itemId}` | `epx carts items delete` |
| `POST /cart/{tenant}/carts/{cartId}/discounts` | `epx carts discounts create` |
| `DELETE /cart/{tenant}/carts/{cartId}/discounts` | `epx carts discounts delete-all` |
| `GET /cart/{tenant}/carts/{cartId}/discounts` | `epx carts discounts list` |
| `DELETE /cart/{tenant}/carts/{cartId}/discounts/{discountIndex}` | `epx carts discounts delete` |
| `GET /cart/{tenant}/carts/{cartId}/dtRestrictions` | `epx carts delivery-restrictions` |
| `GET /cart/{tenant}/carts/{cartId}/validate` | `epx carts validate` |

### Checkout Service

[`checkout/checkout/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/checkout/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /checkout/{tenant}/checkouts/order` | not applicable: needs a signed-in customer's token, which a service token is not |

### Fee Service

[`checkout/fee/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/fee/api-reference/api.yml), last changed 2026-04-16.

| Operation | epx |
| --- | --- |
| `GET /fee/{tenant}/fees` | `epx fees list` |
| `POST /fee/{tenant}/fees` | `epx fees create` |
| `GET /fee/{tenant}/fees/{id}` | `epx fees get` |
| `PUT /fee/{tenant}/fees/{id}` | `epx fees update` |
| `DELETE /fee/{tenant}/fees/{id}` | `epx fees delete` |
| `GET /fee/{tenant}/itemFees` | `epx fees assignments list` |
| `POST /fee/{tenant}/itemFees` | `epx fees assignments create` |
| `GET /fee/{tenant}/itemFees/{itemYRN}/fees` | `epx fees for-item` |
| `PUT /fee/{tenant}/itemFees/{itemYRN}/fees` | `epx fees assignments update` |
| `DELETE /fee/{tenant}/itemFees/{itemYRN}/fees` | `epx fees assignments delete` |
| `DELETE /fee/{tenant}/itemFees/{itemYRN}/fees/{feeId}` | `epx fees assignments delete-fee` |
| `GET /fee/{tenant}/productFees/{productId}/fees` | `epx fees for-product` |
| `PUT /fee/{tenant}/productFees/{productId}/fees` | `epx fees assignments update-product` |
| `DELETE /fee/{tenant}/productFees/{productId}/fees` | `epx fees assignments delete-product` |
| `DELETE /fee/{tenant}/productFees/{productId}/fees/{feeId}` | `epx fees assignments delete-product-fee` |
| `POST /fee/{tenant}/itemFees/searchByProductId` | `epx fees assignments search-by-product` |
| `POST /fee/{tenant}/itemFees/searchByProductIds` | `epx fees assignments search-by-products` |
| `POST /fee/{tenant}/itemFees/search` | `epx fees assignments search` |

### Payment-Gateway Service

[`checkout/payment-gateway/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/payment-gateway/api-reference/api.yml), last changed 2026-02-10.

| Operation | epx |
| --- | --- |
| `POST /payment-gateway/{tenant}/paymentmodes/config` | `epx payment-modes create` |
| `GET /payment-gateway/{tenant}/paymentmodes/config` | `epx payment-modes list` |
| `GET /payment-gateway/{tenant}/paymentmodes/config/{id}` | `epx payment-modes get` |
| `PUT /payment-gateway/{tenant}/paymentmodes/config/{id}` | `epx payment-modes update` |
| `DELETE /payment-gateway/{tenant}/paymentmodes/config/{id}` | `epx payment-modes delete` |
| `GET /payment-gateway/{tenant}/paymentmodes/frontend` | `epx payment-modes public list` |
| `GET /payment-gateway/{tenant}/paymentmodes/frontend/{id}` | `epx payment-modes public get` |
| `POST /payment-gateway/{tenant}/payment/authorize` | `epx payment-transactions authorize` |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/capture` | `epx payment-transactions capture` |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/refund` | `epx payment-transactions refund` |
| `POST /payment-gateway/{tenant}/payment/{transactionId}/cancel` | `epx payment-transactions cancel` |
| `POST /payment-gateway/{tenant}/payment/frontend/initialize` | not applicable: sent by a storefront for a customer's checkout; `epx payment-transactions authorize` authorizes the payment of an order with a service token |
| `POST /payment-gateway/{tenant}/payment/frontend/authorize` | not applicable: sent by a storefront for a customer's checkout; `epx payment-transactions authorize` authorizes the payment of an order with a service token |
| `GET /payment-gateway/{tenant}/transactions` | `epx payment-transactions list` |
| `GET /payment-gateway/{tenant}/transactions/{transactionId}` | `epx payment-transactions get` |

### Shopping List Service

[`checkout/shopping-list/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/shopping-list/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /shoppinglist/{tenant}/shopping-lists` | `epx shopping-lists create` |
| `GET /shoppinglist/{tenant}/shopping-lists` | `epx shopping-lists list` |
| `PUT /shoppinglist/{tenant}/shopping-lists/{customerId}` | `epx shopping-lists update` |
| `GET /shoppinglist/{tenant}/shopping-lists/{customerId}` | `epx shopping-lists get` |
| `DELETE /shoppinglist/{tenant}/shopping-lists/{customerId}` | `epx shopping-lists delete`, `epx shopping-lists delete-all` |

### Approval Service

[`companies-and-customers/approval-service/approval-api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/approval-service/approval-api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /approval/{tenant}/approvals` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /approval/{tenant}/approvals` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /approval/{tenant}/approvals/{approvalId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `PATCH /approval/{tenant}/approvals/{approvalId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `DELETE /approval/{tenant}/approvals/{approvalId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /approval/{tenant}/approval/permitted` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /approval/{tenant}/search/users` | not applicable: needs a signed-in customer's token, which a service token is not |

### Customer Management Service

[`companies-and-customers/client-management/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/client-management/api-reference/api.yml), last changed 2026-04-27.

| Operation | epx |
| --- | --- |
| `POST /customer-management/{tenant}/legal-entities` | `epx legal-entities create` |
| `GET /customer-management/{tenant}/legal-entities` | `epx legal-entities list` |
| `POST /customer-management/{tenant}/legal-entities/search` | `epx legal-entities search` |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}` | `epx legal-entities get` |
| `PUT /customer-management/{tenant}/legal-entities/{legalEntityId}` | `epx legal-entities update` |
| `DELETE /customer-management/{tenant}/legal-entities/{legalEntityId}` | `epx legal-entities delete` |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}/parent-hierarchy` | `epx legal-entities hierarchy` |
| `POST /customer-management/{tenant}/contact-assignments` | `epx contact-assignments create` |
| `GET /customer-management/{tenant}/contact-assignments` | `epx contact-assignments list` |
| `GET /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | `epx contact-assignments get` |
| `PUT /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | `epx contact-assignments update` |
| `DELETE /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | `epx contact-assignments delete` |
| `POST /customer-management/{tenant}/locations` | `epx customer-locations create` |
| `GET /customer-management/{tenant}/locations` | `epx customer-locations list` |
| `GET /customer-management/{tenant}/locations/{locationId}` | `epx customer-locations get` |
| `PUT /customer-management/{tenant}/locations/{locationId}` | `epx customer-locations update` |
| `DELETE /customer-management/{tenant}/locations/{locationId}` | `epx customer-locations delete` |

### Customer Service (customer-managed)

[`companies-and-customers/customer-management/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-management/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /customerlogin/auth/anonymous/login` | not applicable: issues an anonymous storefront token; epx authenticates with a service token |
| `GET /customerlogin/auth/anonymous/refresh` | not applicable: issues an anonymous storefront token; epx authenticates with a service token |
| `POST /customer/{tenant}/signup` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `POST /customer/{tenant}/login` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `GET /customer/{tenant}/logout` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /customer/{tenant}/refreshauthtoken` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `GET /customer/{tenant}/validateauthtoken` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /customer/{tenant}/exchangeauthtoken` | not applicable: exchanges an external identity provider's token for a customer's token (storefront login) |
| `POST /customer/{tenant}/socialLogin` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `PATCH /customer/{tenant}/me` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /customer/{tenant}/me` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /customer/{tenant}/me/addresses` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /customer/{tenant}/me/addresses` | not applicable: needs a signed-in customer's token, which a service token is not |
| `PATCH /customer/{tenant}/me/addresses/{addressId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /customer/{tenant}/me/addresses/{addressId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `DELETE /customer/{tenant}/me/addresses/{addressId}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /customer/{tenant}/me/addresses/{addressId}/tags` | not applicable: needs a signed-in customer's token, which a service token is not |
| `DELETE /customer/{tenant}/me/addresses/{addressId}/tags` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /customer/{tenant}/me/accounts/internal/email/change` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /customer/{tenant}/me/accounts/internal/email/change/confirm` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `POST /customer/{tenant}/password/reset` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `POST /customer/{tenant}/password/reset/update` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `POST /customer/{tenant}/password/change` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /customer/{tenant}/signup/optin/{token}` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |
| `POST /customer/{tenant}/signup/optin/refresh_token` | not applicable: needs an anonymous storefront token (customer signup and login flows), which a service token is not |

### Customer Segment Service

[`companies-and-customers/customer-segments/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-segments/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /customer-segment/{tenant}/segments` | `epx segments create` |
| `GET /customer-segment/{tenant}/segments` | `epx segments list` |
| `POST /customer-segment/{tenant}/segments/search` | `epx segments search` |
| `GET /customer-segment/{tenant}/segments/me` | not applicable: documented for a signed-in customer's token (customersegment.segment_read_own); a service token gets an empty answer |
| `GET /customer-segment/{tenant}/segments/{segmentId}` | `epx segments get` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}` | `epx segments update` |
| `PATCH /customer-segment/{tenant}/segments/{segmentId}` | `epx segments patch` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}` | `epx segments delete` |
| `POST /customer-segment/{tenant}/segments/bulk` | `epx segments bulk-create` |
| `PUT /customer-segment/{tenant}/segments/bulk` | `epx segments bulk-update` |
| `DELETE /customer-segment/{tenant}/segments/bulk` | `epx segments bulk-delete` |
| `POST /customer-segment/{tenant}/segments/match` | `epx segments match` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers` | `epx segments customers list` |
| `POST /customer-segment/{tenant}/segments/{segmentId}/customers/search` | `epx segments customers search` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | `epx segments customers get` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | `epx segments customers update` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | `epx segments customers delete` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | `epx segments customers get-b2b` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | `epx segments customers update-b2b` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | `epx segments customers delete-b2b` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/bulk` | `epx segments customers bulk-update` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/bulk` | `epx segments customers bulk-delete` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/groups` | `epx segments groups list` |
| `POST /customer-segment/{tenant}/segments/{segmentId}/groups/search` | `epx segments groups search` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | `epx segments groups get` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | `epx segments groups update` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | `epx segments groups delete` |
| `GET /customer-segment/{tenant}/segments/items/category-trees` | not applicable: documented for a signed-in customer's token (customersegment.segment_read_own); a service token gets an empty answer |
| `GET /customer-segment/{tenant}/segments/items` | `epx segments all-items` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/items` | `epx segments items list` |
| `POST /customer-segment/{tenant}/segments/{segmentId}/items/search` | `epx segments items search` |
| `GET /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | `epx segments items get` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | `epx segments items update` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | `epx segments items delete` |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/items/{type}/bulk` | `epx segments items bulk-update` |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/items/{type}/bulk` | `epx segments items bulk-delete` |

### Customer Service (tenant-managed)

[`companies-and-customers/customer-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-service/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /customer/{tenant}/customers` | `epx customers list` |
| `POST /customer/{tenant}/customers` | `epx customers create` |
| `POST /customer/{tenant}/customers/search` | `epx customers search` |
| `PUT /customer/{tenant}/customers/{customerNumber}` | `epx customers update` |
| `PATCH /customer/{tenant}/customers/{customerNumber}` | `epx customers patch` |
| `GET /customer/{tenant}/customers/{customerNumber}` | `epx customers get` |
| `DELETE /customer/{tenant}/customers/{customerNumber}` | `epx customers delete` |
| `GET /customer/{tenant}/customers/{customerNumber}/addresses` | `epx customers addresses list` |
| `POST /customer/{tenant}/customers/{customerNumber}/addresses` | `epx customers addresses create` |
| `PUT /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | `epx customers addresses update` |
| `PATCH /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | `epx customers addresses patch` |
| `GET /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | `epx customers addresses get` |
| `DELETE /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | `epx customers addresses delete` |
| `POST /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}/tags` | `epx customers addresses add-tags` |
| `DELETE /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}/tags` | `epx customers addresses remove-tags` |
| `GET /customer/{tenant}/config/password-migration-retention` | `epx customers password-migration get` |
| `POST /customer/{tenant}/config/password-migration-retention` | `epx customers password-migration update` |
| `DELETE /customer/{tenant}/config/password-migration-retention` | `epx customers password-migration delete` |
| `POST /customer/{tenant}/customers/import` | `epx customers import` |

### Vendor Service

[`companies-and-customers/vendor-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/vendor-service/api-reference/api.yml), last changed 2025-11-14.

| Operation | epx |
| --- | --- |
| `POST /vendor/{tenant}/vendors` | `epx vendors create` |
| `GET /vendor/{tenant}/vendors` | `epx vendors list` |
| `POST /vendor/{tenant}/vendors/search` | `epx vendors search` |
| `GET /vendor/{tenant}/vendors/{vendorId}` | `epx vendors get` |
| `PUT /vendor/{tenant}/vendors/{vendorId}` | `epx vendors update` |
| `DELETE /vendor/{tenant}/vendors/{vendorId}` | `epx vendors delete` |
| `POST /vendor/{tenant}/locations` | `epx vendor-locations create` |
| `GET /vendor/{tenant}/locations` | `epx vendor-locations list` |
| `GET /vendor/{tenant}/locations/{locationId}` | `epx vendor-locations get` |
| `PUT /vendor/{tenant}/locations/{locationId}` | `epx vendor-locations update` |
| `DELETE /vendor/{tenant}/locations/{locationId}` | `epx vendor-locations delete` |

### Configuration Service

[`configuration/configuration-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/configuration-service/api-reference/api.yml), last changed 2026-01-28.

| Operation | epx |
| --- | --- |
| `GET /configuration/{tenant}/global-configurations` | `epx configurations global` |
| `GET /configuration/{tenant}/configurations` | `epx configurations list` |
| `POST /configuration/{tenant}/configurations` | `epx configurations create` |
| `GET /configuration/{tenant}/configurations/{propertyKey}` | `epx configurations get` |
| `PUT /configuration/{tenant}/configurations/{propertyKey}` | `epx configurations update` |
| `DELETE /configuration/{tenant}/configurations/{propertyKey}` | `epx configurations delete` |
| `GET /configuration/{tenant}/clients` | `epx configurations clients` |
| `GET /configuration/{tenant}/clients/{client}/configurations` | `epx configurations for-client` |
| `POST /configuration/{tenant}/clients/{client}/configurations` | `epx configurations client-create` |
| `GET /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | `epx configurations client-setting` |
| `PUT /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | `epx configurations client-update` |
| `DELETE /configuration/{tenant}/clients/{client}/configurations/{propertyKey}` | `epx configurations client-delete` |

### Country Service

[`configuration/country-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/country-service/api-reference/api.yml), last changed 2026-06-30.

| Operation | epx |
| --- | --- |
| `GET /country/{tenant}/countries` | `epx countries list` |
| `GET /country/{tenant}/countries/{countryCode}` | `epx countries get` |
| `PATCH /country/{tenant}/countries/{countryCode}` | `epx countries patch` |
| `GET /country/{tenant}/regions` | `epx regions list` |
| `GET /country/{tenant}/regions/{regionCode}` | `epx regions get` |

### Currency Service

[`configuration/currency-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/currency-service/api-reference/api.yml), last changed 2026-04-17.

| Operation | epx |
| --- | --- |
| `POST /currency/{tenant}/currencies` | `epx currencies create` |
| `GET /currency/{tenant}/currencies` | `epx currencies list` |
| `GET /currency/{tenant}/currencies/{currencyCode}` | `epx currencies get` |
| `PUT /currency/{tenant}/currencies/{currencyCode}` | `epx currencies update` |
| `DELETE /currency/{tenant}/currencies/{currencyCode}` | `epx currencies delete` |
| `POST /currency/{tenant}/exchanges` | `epx exchange-rates create` |
| `GET /currency/{tenant}/exchanges` | `epx exchange-rates list` |
| `GET /currency/{tenant}/exchanges/{code}` | `epx exchange-rates get` |
| `PUT /currency/{tenant}/exchanges/{code}` | `epx exchange-rates update` |
| `DELETE /currency/{tenant}/exchanges/{code}` | `epx exchange-rates delete` |

### Indexing Service

[`configuration/indexing-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/indexing-service/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /indexing/{tenant}/configurations` | `epx indexing providers create` |
| `GET /indexing/{tenant}/configurations` | `epx indexing providers list` |
| `GET /indexing/{tenant}/configurations/{provider}` | `epx indexing providers get` |
| `PUT /indexing/{tenant}/configurations/{provider}` | `epx indexing providers update` |
| `DELETE /indexing/{tenant}/configurations/{provider}` | `epx indexing providers delete` |
| `GET /indexing/{tenant}/public/configurations` | `epx indexing providers public list` |
| `GET /indexing/{tenant}/public/configurations/{provider}` | `epx indexing providers public get` |
| `POST /indexing/{tenant}/reindex` | `epx indexing reindex` |
| `POST /indexing/{tenant}/reindex-jobs` | `epx indexing reindex-jobs create` |
| `GET /indexing/{tenant}/reindex-jobs` | `epx indexing reindex-jobs list` |
| `GET /indexing/{tenant}/reindex-jobs/{reindexJobId}` | `epx indexing reindex-jobs get` |

### Site Settings Service

[`configuration/site-settings-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/site-settings-service/api-reference/api.yml), last changed 2026-09-28.

| Operation | epx |
| --- | --- |
| `GET /site/{tenant}/sites` | `epx sites list` |
| `POST /site/{tenant}/sites` | `epx sites create` |
| `GET /site/{tenant}/sites/{siteCode}` | `epx sites get` |
| `PUT /site/{tenant}/sites/{siteCode}` | `epx sites update` |
| `DELETE /site/{tenant}/sites/{siteCode}` | `epx sites delete` |
| `PATCH /site/{tenant}/sites/{siteCode}` | `epx sites patch` |
| `GET /site/{tenant}/siteslist` | `epx sites codes` |
| `GET /site/{tenant}/sites/{siteCode}/mixins` | `epx sites mixins list` |
| `POST /site/{tenant}/sites/{siteCode}/mixins` | `epx sites mixins create` |
| `GET /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | `epx sites mixins get` |
| `DELETE /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | `epx sites mixins delete` |
| `PUT /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | `epx sites mixins update` |
| `PATCH /site/{tenant}/sites/{siteCode}/mixins/{mixinName}` | `epx sites mixins patch` |

### Unit Handling Service

[`configuration/unit-handling-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/configuration/unit-handling-service/api-reference/api.yml), last changed 2026-04-17.

| Operation | epx |
| --- | --- |
| `POST /unit-handling/{tenant}/units` | `epx units create` |
| `GET /unit-handling/{tenant}/units` | `epx units list` |
| `DELETE /unit-handling/{tenant}/units` | `epx units bulk-delete` |
| `GET /unit-handling/{tenant}/units/{unitCode}` | `epx units get` |
| `DELETE /unit-handling/{tenant}/units/{unitCode}` | `epx units delete` |
| `PUT /unit-handling/{tenant}/units/{unitCode}` | `epx units update` |
| `PUT /unit-handling/{tenant}/units/conversion-factor-commands` | `epx units conversion-factor` |
| `PUT /unit-handling/{tenant}/units/convert-unit-commands` | `epx units convert` |
| `GET /unit-handling/{tenant}/types` | `epx units types` |

### Shipping Service

[`delivery-and-shipping/shipping/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/delivery-and-shipping/shipping/api-reference/api.yml), last changed 2026-09-11.

| Operation | epx |
| --- | --- |
| `GET /shipping/{tenant}/areaDeliveryTimes/{deliveryAreaId}/{cartId}` | not applicable: deprecated by Emporix in 2023; `epx delivery-times windows` reads a cart's delivery windows |
| `GET /shipping/{tenant}/actualDeliveryWindows/{cartId}` | `epx delivery-times windows` |
| `POST /shipping/{tenant}/actualDeliveryWindows/incrementCounter` | `epx delivery-times increment-counter` |
| `POST /shipping/{tenant}/findSite` | `epx shipping find-sites` |
| `POST /shipping/{tenant}/{site}/zones` | `epx shipping zones create` |
| `GET /shipping/{tenant}/{site}/zones` | `epx shipping zones list` |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}` | `epx shipping zones get` |
| `PUT /shipping/{tenant}/{site}/zones/{zoneId}` | `epx shipping zones update` |
| `PATCH /shipping/{tenant}/{site}/zones/{zoneId}` | `epx shipping zones patch` |
| `DELETE /shipping/{tenant}/{site}/zones/{zoneId}` | `epx shipping zones delete` |
| `POST /shipping/{tenant}/deliveryWindowValidation` | `epx delivery-times validate` |
| `POST /shipping/{tenant}/{site}/zones/{zoneId}/methods` | `epx shipping zones methods create` |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}/methods` | `epx shipping zones methods list` |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | `epx shipping zones methods get` |
| `PUT /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | `epx shipping zones methods update` |
| `PATCH /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | `epx shipping zones methods patch` |
| `DELETE /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | `epx shipping zones methods delete` |
| `POST /shipping/{tenant}/{site}/quote` | `epx shipping quote` |
| `POST /shipping/{tenant}/{site}/quote/minimum` | `epx shipping quote-minimum` |
| `POST /shipping/{tenant}/{site}/quote/slot` | `epx shipping quote-slot` |
| `POST /shipping/{tenant}/{site}/groups` | `epx shipping groups create` |
| `GET /shipping/{tenant}/{site}/groups` | `epx shipping groups list` |
| `GET /shipping/{tenant}/{site}/groups/{groupId}` | `epx shipping groups get` |
| `PUT /shipping/{tenant}/{site}/groups/{groupId}` | `epx shipping groups update` |
| `DELETE /shipping/{tenant}/{site}/groups/{groupId}` | `epx shipping groups delete` |
| `POST /shipping/{tenant}/{site}/cgrelations` | `epx shipping customers create` |
| `GET /shipping/{tenant}/{site}/cgrelations` | `epx shipping customers list` |
| `GET /shipping/{tenant}/{site}/cgrelations/{customerId}` | `epx shipping customers get` |
| `PUT /shipping/{tenant}/{site}/cgrelations/{customerId}` | `epx shipping customers update` |
| `DELETE /shipping/{tenant}/{site}/cgrelations/{customerId}` | `epx shipping customers delete` |
| `GET /shipping/{tenant}/delivery-times` | `epx delivery-times list` |
| `POST /shipping/{tenant}/delivery-times` | `epx delivery-times create` |
| `POST /shipping/{tenant}/delivery-times/bulk` | `epx delivery-times bulk-create` |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}` | `epx delivery-times get` |
| `PUT /shipping/{tenant}/delivery-times/{deliveryTimeId}` | `epx delivery-times update` |
| `PATCH /shipping/{tenant}/delivery-times/{deliveryTimeId}` | `epx delivery-times patch` |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}` | `epx delivery-times delete` |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | `epx delivery-times slots list` |
| `POST /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | `epx delivery-times slots create` |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | `epx delivery-times slots delete-all` |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | `epx delivery-times slots get` |
| `PUT /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | `epx delivery-times slots update` |
| `PATCH /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | `epx delivery-times slots patch` |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | `epx delivery-times slots delete` |
| `POST /shipping/{tenant}/delivery-cycles/generate` | `epx delivery-times generate-cycle` |

### Media Service

[`media/media/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/media/media/api-reference/api.yml), last changed 2026-09-28.

| Operation | epx |
| --- | --- |
| `POST /media/{tenant}/assets` | `epx media create` |
| `GET /media/{tenant}/assets` | `epx media list` |
| `GET /media/{tenant}/assets/{assetId}` | `epx media get` |
| `PUT /media/{tenant}/assets/{assetId}` | `epx media update` |
| `PATCH /media/{tenant}/assets/{assetId}` | `epx media patch` |
| `DELETE /media/{tenant}/assets/{assetId}` | `epx media delete` |
| `GET /media/{tenant}/assets/{assetId}/download` | not applicable: answers with the file itself; `epx media download-url` gives a URL to it |
| `POST /media/{tenant}/assets/upload-session` | `epx media start-upload` |
| `GET /media/{tenant}/assets/{assetId}/download-url` | `epx media download-url` |

### Availability Service

[`orders/availability/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/availability/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /availability/{tenant}/locations/{site}` | not applicable: deprecated by Emporix, removed on 1 December 2026 |
| `POST /availability/{tenant}/locations/{site}` | not applicable: deprecated by Emporix, removed on 1 December 2026 |
| `PUT /availability/{tenant}/locations/{site}` | not applicable: deprecated by Emporix, removed on 1 December 2026 |
| `DELETE /availability/{tenant}/locations/{location}` | not applicable: deprecated by Emporix, removed on 1 December 2026 |
| `POST /availability/{tenant}/search/locations` | not applicable: deprecated by Emporix, removed on 1 December 2026 |
| `GET /availability/{tenant}/availability/site/{site}` | `epx availability list` |
| `POST /availability/{tenant}/availability/search` | `epx availability search` |
| `GET /availability/{tenant}/availability/{productId}/{site}` | `epx availability get` |
| `POST /availability/{tenant}/availability/{productId}/{site}` | `epx availability create` |
| `PUT /availability/{tenant}/availability/{productId}/{site}` | `epx availability update` |
| `DELETE /availability/{tenant}/availability/{productId}/{site}` | `epx availability delete` |
| `POST /availability/{tenant}/availability/bulk` | `epx availability bulk-create` |
| `PUT /availability/{tenant}/availability/bulk` | `epx availability bulk-update` |
| `DELETE /availability/{tenant}/availability/bulk` | `epx availability bulk-delete` |

### Invoice Service

[`orders/invoice/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/invoice/api-reference/api.yaml), last changed 2025-11-18.

| Operation | epx |
| --- | --- |
| `POST /invoice/{tenant}/jobs/invoices` | `epx invoice-jobs create` |
| `GET /invoice/{tenant}/jobs/invoices/{jobId}` | `epx invoice-jobs get` |

### Order Service

[`orders/order/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/order/api-reference/api.yml), last changed 2026-10-07.

| Operation | epx |
| --- | --- |
| `GET /order-v2/{tenant}/salesorders` | `epx orders list` |
| `POST /order-v2/{tenant}/salesorders` | `epx orders create` |
| `HEAD /order-v2/{tenant}/salesorders` | `epx orders count` |
| `POST /order-v2/{tenant}/salesorders/search` | `epx orders search` |
| `GET /order-v2/{tenant}/salesorders/{orderId}` | `epx orders get` |
| `PUT /order-v2/{tenant}/salesorders/{orderId}` | `epx orders update` |
| `PATCH /order-v2/{tenant}/salesorders/{orderId}` | `epx orders patch` |
| `DELETE /order-v2/{tenant}/salesorders/{orderId}` | `epx orders delete` |
| `GET /order-v2/{tenant}/salesorders/{orderId}/transitions` | `epx orders transitions` |
| `POST /order-v2/{tenant}/salesorders/{orderId}/transitions` | `epx orders transition` |
| `GET /order-v2/{tenant}/salesorders/{orderId}/historical-transitions` | `epx orders history` |
| `POST /order-v2/{tenant}/salesorders/{orderId}/calculations` | `epx orders recalculate` |
| `POST /order-v2/{tenant}/salesorders/{orderId}/entries` | `epx orders add-entries` |
| `POST /order-v2/{tenant}/salesorders/{orderId}/split` | `epx orders split` |
| `GET /order-v2/{tenant}/orders` | not applicable: reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders |
| `GET /order-v2/{tenant}/orders/{orderId}` | not applicable: reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders |
| `GET /order-v2/{tenant}/orders/{orderId}/transitions` | not applicable: reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders |
| `POST /order-v2/{tenant}/orders/{orderId}/transitions` | not applicable: reads or changes a signed-in customer's own orders (saas-token); `epx orders` reads the tenant's orders |
| `GET /order-v2/{tenant}/legal-entity-orders/{legalEntityId}` | not applicable: needs a signed-in customer's token (saas-token); `epx orders list --query legalEntityId:ID` lists the orders |
| `GET /order-v2/{tenant}/legal-entity-orders/{legalEntityId}/{orderId}` | not applicable: needs a signed-in customer's token (saas-token); `epx orders list --query legalEntityId:ID` lists the orders |

### Returns Service

[`orders/returns/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/returns/api-reference/api.yml), last changed 2026-03-20.

| Operation | epx |
| --- | --- |
| `POST /return/{tenant}/returns` | `epx returns create` |
| `GET /return/{tenant}/returns` | `epx returns list` |
| `GET /return/{tenant}/returns/{returnId}` | `epx returns get` |
| `PUT /return/{tenant}/returns/{returnId}` | `epx returns update` |
| `DELETE /return/{tenant}/returns/{returnId}` | `epx returns delete` |
| `PATCH /return/{tenant}/returns/{returnId}` | `epx returns patch` |

### Price Service

[`prices-and-taxes/price-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/prices-and-taxes/price-service/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /price/{tenant}/prices` | `epx prices create` |
| `GET /price/{tenant}/prices` | `epx prices list`, `epx prices for-product` |
| `POST /price/{tenant}/prices/bulk` | `epx prices bulk-create` |
| `PUT /price/{tenant}/prices/bulk` | `epx prices bulk-update` |
| `GET /price/{tenant}/prices/{priceId}` | `epx prices get` |
| `PUT /price/{tenant}/prices/{priceId}` | `epx prices update` |
| `DELETE /price/{tenant}/prices/{priceId}` | `epx prices delete` |
| `POST /price/{tenant}/match-prices` | `epx prices match` |
| `POST /price/{tenant}/match-prices-by-context` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /price/{tenant}/prices/search` | `epx prices search` |
| `POST /price/{tenant}/priceModels` | `epx price-models create` |
| `GET /price/{tenant}/priceModels` | `epx price-models list` |
| `GET /price/{tenant}/priceModels/{priceModelId}` | `epx price-models get` |
| `PUT /price/{tenant}/priceModels/{priceModelId}` | `epx price-models update` |
| `DELETE /price/{tenant}/priceModels/{priceModelId}` | `epx price-models delete` |
| `GET /price/{tenant}/price-lists` | `epx price-lists list` |
| `POST /price/{tenant}/price-lists` | `epx price-lists create` |
| `POST /price/{tenant}/price-lists/search` | `epx price-lists search` |
| `GET /price/{tenant}/price-lists/{priceListId}` | `epx price-lists get` |
| `PUT /price/{tenant}/price-lists/{priceListId}` | `epx price-lists update` |
| `DELETE /price/{tenant}/price-lists/{priceListId}` | `epx price-lists delete` |
| `GET /price/{tenant}/price-lists/{priceListId}/prices` | `epx price-lists prices list` |
| `POST /price/{tenant}/price-lists/{priceListId}/prices` | `epx price-lists prices create` |
| `DELETE /price/{tenant}/price-lists/{priceListId}/prices/bulk` | `epx price-lists prices bulk-delete` |
| `POST /price/{tenant}/price-lists/{priceListId}/prices/bulk` | `epx price-lists prices bulk-create` |
| `PUT /price/{tenant}/price-lists/{priceListId}/prices/bulk` | `epx price-lists prices bulk-update` |
| `POST /price/{tenant}/price-lists/{priceListId}/prices/search` | `epx price-lists prices search` |
| `GET /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | `epx price-lists prices get` |
| `PUT /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | `epx price-lists prices update` |
| `DELETE /price/{tenant}/price-lists/{priceListId}/prices/{priceId}` | `epx price-lists prices delete` |

### Tax Service

[`prices-and-taxes/tax-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/prices-and-taxes/tax-service/api-reference/api.yml), last changed 2025-11-15.

| Operation | epx |
| --- | --- |
| `POST /tax/{tenant}/taxes` | `epx taxes create` |
| `GET /tax/{tenant}/taxes` | `epx taxes list` |
| `GET /tax/{tenant}/taxes/{locationCode}` | `epx taxes get` |
| `PUT /tax/{tenant}/taxes/{locationCode}` | `epx taxes update` |
| `DELETE /tax/{tenant}/taxes/{locationCode}` | `epx taxes delete` |
| `PUT /tax/{tenant}/taxes/calculation-commands` | `epx taxes calculate` |

### Brand Service

[`products-labels-and-brands/brand-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/products-labels-and-brands/brand-service/api-reference/api.yml), last changed 2026-01-19.

| Operation | epx |
| --- | --- |
| `GET /brand/brands` | `epx brands list` |
| `POST /brand/brands` | `epx brands create` |
| `GET /brand/brands/{brandId}` | `epx brands get` |
| `PUT /brand/brands/{brandId}` | `epx brands update` |
| `PATCH /brand/brands/{brandId}` | `epx brands patch` |
| `DELETE /brand/brands/{brandId}` | `epx brands delete` |

### Label Service

[`products-labels-and-brands/label-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/products-labels-and-brands/label-service/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /label/labels` | `epx labels list` |
| `POST /label/labels` | `epx labels create` |
| `GET /label/labels/{labelId}` | `epx labels get` |
| `PUT /label/labels/{labelId}` | `epx labels update` |
| `PATCH /label/labels/{labelId}` | `epx labels patch` |
| `DELETE /label/labels/{labelId}` | `epx labels delete` |

### Product Service

[`products-labels-and-brands/product-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/products-labels-and-brands/product-service/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `POST /product/{tenant}/products/search` | `epx products search` |
| `GET /product/{tenant}/products` | `epx products list` |
| `POST /product/{tenant}/products` | `epx products create` |
| `POST /product/{tenant}/products/bulk` | `epx products bulk-create` |
| `PUT /product/{tenant}/products/bulk` | `epx products bulk-update`, `epx products import` |
| `POST /product/{tenant}/products/recalculate` | `epx products recalculate` |
| `GET /product/{tenant}/products/recalculate/jobs` | `epx products recalculation-jobs list` |
| `GET /product/{tenant}/products/recalculate/jobs/{jobId}` | `epx products recalculation-jobs get` |
| `GET /product/{tenant}/products/{productId}` | `epx products get` |
| `PUT /product/{tenant}/products/{productId}` | `epx products update` |
| `PATCH /product/{tenant}/products/{productId}` | `epx products patch` |
| `DELETE /product/{tenant}/products/{productId}` | `epx products delete` |
| `GET /product/{tenant}/product-templates` | `epx product-templates list` |
| `POST /product/{tenant}/product-templates` | `epx product-templates create` |
| `GET /product/{tenant}/product-templates/{product-template-id}` | `epx product-templates get` |
| `PUT /product/{tenant}/product-templates/{product-template-id}` | `epx product-templates update` |
| `DELETE /product/{tenant}/product-templates/{product-template-id}` | `epx product-templates delete` |

### Quote Service

[`quotes/quote/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/quotes/quote/api-reference/api.yaml), last changed 2026-10-09.

| Operation | epx |
| --- | --- |
| `GET /quote/{tenant}/quotes` | `epx quotes list` |
| `POST /quote/{tenant}/quotes` | `epx quotes create` |
| `POST /quote/{tenant}/quotes/{quoteId}/pdf` | `epx quotes pdf` |
| `GET /quote/{tenant}/quotes/{quoteId}/history` | `epx quotes history` |
| `GET /quote/{tenant}/quotes/{quoteId}` | `epx quotes get` |
| `PATCH /quote/{tenant}/quotes/{quoteId}` | `epx quotes patch` |
| `DELETE /quote/{tenant}/quotes/{quoteId}` | `epx quotes delete` |
| `POST /quote/{tenant}/quote-reasons` | `epx quotes reasons create` |
| `GET /quote/{tenant}/quote-reasons` | `epx quotes reasons list` |
| `GET /quote/{tenant}/quote-reasons/{quoteReasonId}` | `epx quotes reasons get` |
| `PUT /quote/{tenant}/quote-reasons/{quoteReasonId}` | `epx quotes reasons update` |
| `DELETE /quote/{tenant}/quote-reasons/{quoteReasonId}` | `epx quotes reasons delete` |

### Coupon Service

[`rewards-and-promotions/coupon/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/rewards-and-promotions/coupon/api-reference/api.yml), last changed 2026-09-25.

| Operation | epx |
| --- | --- |
| `GET /coupon/{tenant}/coupons` | `epx coupons list` |
| `POST /coupon/{tenant}/coupons` | `epx coupons create` |
| `GET /coupon/{tenant}/coupons/{code}` | `epx coupons get` |
| `PUT /coupon/{tenant}/coupons/{code}` | `epx coupons update` |
| `PATCH /coupon/{tenant}/coupons/{code}` | `epx coupons patch` |
| `DELETE /coupon/{tenant}/coupons/{code}` | `epx coupons delete` |
| `POST /coupon/{tenant}/coupons/{code}/validation` | `epx coupons validate` |
| `GET /coupon/{tenant}/coupons/{code}/redemptions` | `epx coupons redemptions` |
| `POST /coupon/{tenant}/coupons/{code}/redemptions` | `epx coupons redeem` |
| `GET /coupon/{tenant}/coupons/{code}/redemptions/{id}` | `epx coupons redemption` |
| `DELETE /coupon/{tenant}/coupons/{code}/redemptions/{id}` | `epx coupons redemption-delete` |
| `GET /coupon/{tenant}/referral-coupons/{customerNumber}` | `epx coupons referral` |
| `POST /coupon/{tenant}/referral-coupons/{customerNumber}` | `epx coupons referral-create` |

### Reward Points Service

[`rewards-and-promotions/reward-points/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/rewards-and-promotions/reward-points/api-reference/api.yml), last changed 2025-11-17.

| Operation | epx |
| --- | --- |
| `GET /reward-points/summaryBatch` | `epx reward-points list` |
| `GET /reward-points/customer/{customerId}` | `epx reward-points balance` |
| `POST /reward-points/customer/{customerId}` | `epx reward-points create` |
| `DELETE /reward-points/customer/{customerId}` | `epx reward-points delete` |
| `GET /reward-points/customer/{customerId}/summary` | `epx reward-points get` |
| `POST /reward-points/customer/{customerId}/addPoints` | `epx reward-points add` |
| `POST /reward-points/customer/{customerId}/redeemPoints` | `epx reward-points redeem` |
| `GET /reward-points/public/customer` | not applicable: needs a signed-in customer's token, which a service token is not |
| `GET /reward-points/public/customer/summary` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /reward-points/{tenant}/redeemOptions` | `epx reward-points redeem-option-create` |
| `GET /reward-points/{tenant}/redeemOptions` | `epx reward-points redeem-options` |
| `DELETE /reward-points/{tenant}/redeemOptions/{redeemOptionId}` | `epx reward-points redeem-option-delete` |
| `PUT /reward-points/{tenant}/redeemOptions/{redeemOptionId}` | `epx reward-points redeem-option-update` |
| `POST /reward-points/public/customer/redeem` | not applicable: needs a signed-in customer's token, which a service token is not |

### IAM Service

[`users-and-permissions/iam/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/users-and-permissions/iam/api-reference/api.yml), last changed 2026-07-01.

| Operation | epx |
| --- | --- |
| `GET /iam/{tenant}/access-controls` | `epx iam access-controls list` |
| `GET /iam/{tenant}/access-controls/{accessControlId}` | `epx iam access-controls get` |
| `PUT /iam/{tenant}/access-controls/{accessControlId}` | `epx iam access-controls update` |
| `DELETE /iam/{tenant}/access-controls/{accessControlId}` | `epx iam access-controls delete` |
| `GET /iam/{tenant}/scopes` | `epx iam scopes list` |
| `GET /iam/{tenant}/scopes/{scopeId}` | `epx iam scopes get` |
| `PUT /iam/{tenant}/scopes/{scopeId}` | `epx iam scopes update` |
| `DELETE /iam/{tenant}/scopes/{scopeId}` | `epx iam scopes delete` |
| `GET /iam/{tenant}/groups` | `epx iam groups list` |
| `POST /iam/{tenant}/groups` | `epx iam groups create` |
| `GET /iam/{tenant}/groups/{groupId}` | `epx iam groups get` |
| `DELETE /iam/{tenant}/groups/{groupId}` | `epx iam groups delete` |
| `PUT /iam/{tenant}/groups/{groupId}` | `epx iam groups update` |
| `GET /iam/{tenant}/groups/{groupId}/access-controls` | `epx iam groups access-controls` |
| `GET /iam/{tenant}/groups/{groupId}/users` | `epx iam groups users list` |
| `POST /iam/{tenant}/groups/{groupId}/users` | `epx iam groups users create` |
| `DELETE /iam/{tenant}/groups/{groupId}/users` | `epx iam groups users delete-all` |
| `PUT /iam/{tenant}/groups/{groupId}/users/{userType}/{userId}` | `epx iam groups users update` |
| `DELETE /iam/{tenant}/groups/{groupId}/users/{userId}` | `epx iam groups users delete` |
| `GET /iam/{tenant}/permissions` | `epx iam permissions list` |
| `GET /iam/{tenant}/permissions/{permissionId}` | `epx iam permissions get` |
| `GET /iam/{tenant}/resources` | `epx iam resources list` |
| `GET /iam/{tenant}/resources/{resourceId}` | `epx iam resources get` |
| `GET /iam/{tenant}/templates` | `epx iam templates list` |
| `GET /iam/{tenant}/roles` | `epx iam roles list` |
| `GET /iam/{tenant}/roles/{roleId}` | `epx iam roles get` |
| `GET /iam/{tenant}/users/{userId}/access-controls` | `epx iam users access-controls` |
| `GET /iam/{tenant}/users/{userId}/access-controls/{resourceId}` | `epx iam users resource-access-controls` |
| `GET /iam/{tenant}/users/{userId}/groups` | `epx iam users groups list` |
| `DELETE /iam/{tenant}/users/{userId}/groups` | `epx iam users groups delete-all` |
| `GET /iam/{tenant}/users/{userId}/groups/{groupId}` | `epx iam users groups get` |
| `GET /iam/{tenant}/users/{userId}/permissions/{resourceId}` | `epx iam users resource-permissions` |
| `GET /iam/{tenant}/users/{userId}/scopes` | `epx iam users scopes` |
| `GET /iam/{tenant}/users/me/access-controls` | not applicable: describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID |
| `GET /iam/{tenant}/users/me/scopes` | not applicable: describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID |
| `GET /iam/{tenant}/users` | `epx iam users list` |
| `POST /iam/{tenant}/users` | `epx iam users create` |
| `GET /iam/{tenant}/users/me` | not applicable: describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID |
| `GET /iam/{tenant}/users/{userId}` | `epx iam users get` |
| `PUT /iam/{tenant}/users/{userId}` | `epx iam users update` |
| `DELETE /iam/{tenant}/users/{userId}` | `epx iam users delete` |
| `GET /iam/{tenant}/users/vendors/{vendorId}` | `epx iam users for-vendor` |

### Session Context Service

[`users-and-permissions/session-context/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/users-and-permissions/session-context/api-reference/api.yaml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /session-context/{tenant}/context/{sessionId}` | `epx session-contexts get` |
| `PUT /session-context/{tenant}/context/{sessionId}` | `epx session-contexts update` |
| `DELETE /session-context/{tenant}/context/{sessionId}/attributes/{attributeName}` | `epx session-contexts attributes delete` |
| `POST /session-context/{tenant}/context/{sessionId}/attributes` | `epx session-contexts attributes create` |
| `GET /session-context/{tenant}/me/context` | not applicable: needs a signed-in customer's token, which a service token is not |
| `PATCH /session-context/{tenant}/me/context` | not applicable: needs a signed-in customer's token, which a service token is not |
| `DELETE /session-context/{tenant}/me/context/attributes/{attributeName}` | not applicable: needs a signed-in customer's token, which a service token is not |
| `POST /session-context/{tenant}/me/context/attributes` | not applicable: needs a signed-in customer's token, which a service token is not |

### Audit Logs (Changelog) Service

[`utilities/audit-logs-changelog/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/audit-logs-changelog/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /changelog/{tenant}/changelogs` | `epx changelog list` |

### Import Service

[`utilities/import-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/import-service/api-reference/api.yml), last changed 2026-10-14.

| Operation | epx |
| --- | --- |
| `GET /importtool/{tenant}/configs` | `epx imports configurations list` |
| `GET /importtool/{tenant}/configs/{id}` | `epx imports configurations get` |
| `GET /importtool/{tenant}/configs/{configId}/streams` | `epx imports configurations streams` |
| `GET /importtool/{tenant}/streams/{id}` | `epx imports streams get` |
| `GET /importtool/{tenant}/configs/{configId}/schedule` | `epx imports configurations schedule` |
| `PUT /importtool/{tenant}/configs/{configId}/schedule` | `epx imports configurations update-schedule` |
| `DELETE /importtool/{tenant}/configs/{configId}/schedule` | `epx imports configurations delete-schedule` |
| `GET /importtool/{tenant}/configs/{configId}/stream-order` | `epx imports configurations stream-order` |
| `POST /importtool/{tenant}/configs/{configId}/runs` | `epx imports configurations trigger` |
| `GET /importtool/{tenant}/configs/{configId}/runs` | `epx imports configurations runs` |
| `GET /importtool/{tenant}/runs/{runId}` | `epx imports runs get`, `epx imports runs streams` |
| `GET /importtool/{tenant}/runs/{runId}/events` | not applicable: answers with Server-Sent Events, a stream held open until the run finishes; `epx imports runs get` and `epx imports runs streams` read the run's progress per stream as JSON |
| `POST /importtool/{tenant}/runs/{runId}/cancel` | `epx imports runs cancel` |
| `GET /importtool/{tenant}/runs/{runId}/errors` | `epx imports runs errors` |
| `GET /importtool/{tenant}/runs/{runId}/diagnostics` | `epx imports runs diagnostics` |
| `GET /importtool/{tenant}/runs/{runId}/diagnostics/csv` | not applicable: answers with a CSV file; `epx imports runs diagnostics` reads the same rows as JSON |
| `GET /importtool/{tenant}/data/types` | `epx imports types` |
| `GET /importtool/{tenant}/data/records` | `epx imports records` |
| `GET /importtool/{tenant}/data/streams/{streamId}/records` | `epx imports streams records` |
| `GET /importtool/{tenant}/stats` | `epx imports statistics` |
| `GET /importtool/{tenant}/dashboard/job-groups` | `epx imports job-groups` |
| `GET /importtool/{tenant}/settings/health-thresholds` | `epx imports health-thresholds` |
| `GET /importtool/{tenant}/license` | `epx imports license` |
| `POST /importtool/{tenant}/runs/{runId}/retry` | `epx imports runs retry` |

### Schema Service

[`utilities/schema/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/schema/api-reference/api.yml), last changed 2026-10-12.

| Operation | epx |
| --- | --- |
| `POST /schema/{tenant}/schemas` | `epx schemas create` |
| `GET /schema/{tenant}/schemas` | `epx schemas list` |
| `GET /schema/{tenant}/schemas/{id}` | `epx schemas get` |
| `DELETE /schema/{tenant}/schemas/{id}` | `epx schemas delete` |
| `PUT /schema/{tenant}/schemas/{id}` | `epx schemas update` |
| `PUT /schema/{tenant}/schemas/{id}/types` | `epx schemas update-types` |
| `POST /schema/{tenant}/schemas/file` | `epx schemas parse` |
| `GET /schema/{tenant}/types` | `epx schemas types` |
| `POST /schema/{tenant}/references` | `epx schemas references create` |
| `GET /schema/{tenant}/references` | `epx schemas references list` |
| `GET /schema/{tenant}/references/{id}` | `epx schemas references get` |
| `DELETE /schema/{tenant}/references/{id}` | `epx schemas references delete` |
| `PUT /schema/{tenant}/references/{id}` | `epx schemas references update` |
| `POST /schema/{tenant}/custom-entities` | `epx custom-entities create` |
| `GET /schema/{tenant}/custom-entities` | `epx custom-entities list` |
| `GET /schema/{tenant}/custom-entities/{id}` | `epx custom-entities get` |
| `DELETE /schema/{tenant}/custom-entities/{id}` | `epx custom-entities delete` |
| `PUT /schema/{tenant}/custom-entities/{id}` | `epx custom-entities update` |
| `POST /schema/{tenant}/custom-entities/{type}/instances` | `epx custom-entities instances create` |
| `GET /schema/{tenant}/custom-entities/{type}/instances` | `epx custom-entities instances list` |
| `GET /schema/{tenant}/custom-entities/{type}/instances/{id}` | `epx custom-entities instances get` |
| `DELETE /schema/{tenant}/custom-entities/{type}/instances/{id}` | `epx custom-entities instances delete` |
| `PUT /schema/{tenant}/custom-entities/{type}/instances/{id}` | `epx custom-entities instances update` |
| `PATCH /schema/{tenant}/custom-entities/{type}/instances/{id}` | `epx custom-entities instances patch` |
| `POST /schema/{tenant}/custom-entities/{type}/instances/search` | `epx custom-entities instances search` |
| `PATCH /schema/{tenant}/custom-entities/{type}/instances/bulk` | `epx custom-entities instances bulk-patch` |
| `POST /schema/{tenant}/custom-entities/{type}/instances/bulk` | `epx custom-entities instances bulk-create` |
| `PUT /schema/{tenant}/custom-entities/{type}/instances/bulk` | `epx custom-entities instances bulk-update` |
| `DELETE /schema/{tenant}/custom-entities/{type}/instances/bulk` | `epx custom-entities instances bulk-delete` |
| `POST /schema/{tenant}/custom-entities/export` | `epx custom-entities export` |
| `POST /schema/{tenant}/custom-entities/import` | `epx custom-entities import` |

### Search Service

[`utilities/search-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/search-service/api-reference/api.yml), last changed 2026-10-09.

| Operation | epx |
| --- | --- |
| `POST /search/{tenant}/search/{type}` | `epx search documents` |
| `GET /search/{tenant}/search/queries` | `epx search queries list` |
| `GET /search/{tenant}/search/{type}/queries` | `epx search queries for-type` |
| `PUT /search/{tenant}/search/{type}/queries/{id}` | `epx search queries update` |
| `GET /search/{tenant}/search/{type}/queries/{id}` | `epx search queries get` |
| `DELETE /search/{tenant}/search/{type}/queries/{id}` | `epx search queries delete` |
| `GET /search/{tenant}/search/indexes` | `epx search indexes list` |
| `POST /search/{tenant}/search/indexes/export` | `epx search indexes export` |
| `POST /search/{tenant}/search/indexes/import` | `epx search indexes import` |
| `GET /search/{tenant}/search/{type}/indexes` | `epx search indexes for-type` |
| `PUT /search/{tenant}/search/{type}/indexes/{id}` | `epx search indexes update` |
| `GET /search/{tenant}/search/{type}/indexes/{id}` | `epx search indexes get` |
| `DELETE /search/{tenant}/search/{type}/indexes/{id}` | `epx search indexes delete` |
| `GET /search/{tenant}/jobs` | `epx search jobs list` |
| `GET /search/{tenant}/jobs/{id}` | `epx search jobs get` |

### Sequential ID Service

[`utilities/sequential-id/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/sequential-id/api-reference/api.yml), last changed 2026-10-01.

| Operation | epx |
| --- | --- |
| `GET /sequential-id/{tenant}/schemas` | `epx sequential-ids list` |
| `POST /sequential-id/{tenant}/schemas` | `epx sequential-ids create` |
| `GET /sequential-id/{tenant}/schemas/{schemaId}` | `epx sequential-ids get` |
| `DELETE /sequential-id/{tenant}/schemas/{schemaId}` | `epx sequential-ids delete` |
| `POST /sequential-id/{tenant}/schemas/{schemaId}/setActive` | `epx sequential-ids activate` |
| `GET /sequential-id/{tenant}/schemas/types/{schemaType}` | `epx sequential-ids for-type` |
| `POST /sequential-id/{tenant}/schemas/types/{schemaType}/nextId` | `epx sequential-ids next-id` |
| `POST /sequential-id/sequenceSchemaBatch/nextIds` | `epx sequential-ids next-ids` |

### Webhook Service

[`webhooks/webhook-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/webhooks/webhook-service/api-reference/api.yml), last changed 2026-08-11.

| Operation | epx |
| --- | --- |
| `POST /webhook/{tenant}/config` | `epx webhooks configurations create` |
| `GET /webhook/{tenant}/config` | `epx webhooks configurations list` |
| `PUT /webhook/{tenant}/config/{code}` | `epx webhooks configurations update` |
| `PATCH /webhook/{tenant}/config/{code}` | `epx webhooks configurations patch` |
| `GET /webhook/{tenant}/config/{code}` | `epx webhooks configurations get` |
| `DELETE /webhook/{tenant}/config/{code}` | `epx webhooks configurations delete` |
| `PATCH /webhook/{tenant}/event-subscriptions` | `epx webhooks subscriptions bulk-patch` |
| `GET /webhook/{tenant}/event-subscriptions` | `epx webhooks subscriptions list` |
| `GET /webhook/{tenant}/dashboard-access` | `epx webhooks dashboard-access` |
| `GET /webhook/{tenant}/statistics` | `epx webhooks statistics` |

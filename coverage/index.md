# Emporix API coverage

> Which operations of each Emporix API service epx sends, which do not apply, and which are planned.

Which operations of the Emporix API reference `epx` sends, checked against
[emporix/api-references](https://github.com/emporix/api-references) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

**Coverage:** 605 built, 65 not applicable and 0 not yet built (0 planned, 0 to do), of 670
operations in 43 services
(and the OAuth Service for the token).

Some operations do not apply to `epx`, which acts with a service token (an API key's client
credentials): they need a customer's or a storefront token, are deprecated, or answer with
something other than JSON. Beyond the commands listed, [`epx api get PATH`](/emporix-cli/commands/api#epx-api-get)
reads any other GET operation that answers JSON to a service token, and
[`epx api post PATH`](/emporix-cli/commands/api#epx-api-post), [`epx api put PATH`](/emporix-cli/commands/api#epx-api-put), [`epx api patch PATH`](/emporix-cli/commands/api#epx-api-patch),
or [`epx api delete PATH`](/emporix-cli/commands/api#epx-api-delete) sends any write with the scopes given, before a
command of its own does.

Each service is identified by the last change to its reference file up to that commit, and
by its version where the file states one. When a reference file changes after the date
shown, check the commands that use it before relying on them.

These pages show the same data as [COMPATIBILITY.md](https://github.com/commerce-cloud-integrations/emporix-cli/blob/main/COMPATIBILITY.md).

| Service | Last change | Version | Operations | Built | Not applicable | Planned | To do |
| --- | --- | --- | --- | --- | --- | --- | --- |
| [AI Rag Indexer](/emporix-cli/coverage/ai-rag-indexer) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | — | 3 | 3 | 0 | 0 | 0 |
| [AI Service](/emporix-cli/coverage/ai-service) | 2026-10-06 ([`9b456b6`](https://github.com/emporix/api-references/commit/9b456b6)) | — | 57 | 56 | 1 | 0 | 0 |
| [Approval Service](/emporix-cli/coverage/approval-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 7 | 0 | 7 | 0 | 0 |
| [Audit Logs (Changelog) Service](/emporix-cli/coverage/audit-logs-changelog-service) | 2026-10-01 ([`84cc3ec`](https://github.com/emporix/api-references/commit/84cc3ec)) | — | 1 | 1 | 0 | 0 | 0 |
| [Availability Service](/emporix-cli/coverage/availability-service) | 2026-10-01 ([`84cc3ec`](https://github.com/emporix/api-references/commit/84cc3ec)) | — | 14 | 9 | 5 | 0 | 0 |
| [Brand Service](/emporix-cli/coverage/brand-service) | 2026-01-19 ([`d7fbba2`](https://github.com/emporix/api-references/commit/d7fbba2)) | — | 6 | 6 | 0 | 0 | 0 |
| [Cart Service](/emporix-cli/coverage/cart-service) | 2026-10-07 ([`1f21665`](https://github.com/emporix/api-references/commit/1f21665)) | — | 25 | 25 | 0 | 0 | 0 |
| [Catalog Management](/emporix-cli/coverage/catalog-management) | 2026-04-16 ([`6d06d44`](https://github.com/emporix/api-references/commit/6d06d44)) | `1.0.0` | 7 | 7 | 0 | 0 | 0 |
| [Category](/emporix-cli/coverage/category) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 24 | 23 | 1 | 0 | 0 |
| [Checkout Service](/emporix-cli/coverage/checkout-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 1 | 0 | 1 | 0 | 0 |
| [Configuration Service](/emporix-cli/coverage/configuration-service) | 2026-01-28 ([`9015ed5`](https://github.com/emporix/api-references/commit/9015ed5)) | — | 12 | 12 | 0 | 0 | 0 |
| [Country Service](/emporix-cli/coverage/country-service) | 2026-06-30 ([`1e29a43`](https://github.com/emporix/api-references/commit/1e29a43)) | — | 5 | 5 | 0 | 0 | 0 |
| [Coupon Service](/emporix-cli/coverage/coupon-service) | 2026-09-25 ([`86c5409`](https://github.com/emporix/api-references/commit/86c5409)) | — | 13 | 13 | 0 | 0 | 0 |
| [Currency Service](/emporix-cli/coverage/currency-service) | 2026-04-17 ([`2e76d42`](https://github.com/emporix/api-references/commit/2e76d42)) | — | 10 | 10 | 0 | 0 | 0 |
| [Customer Management Service](/emporix-cli/coverage/customer-management-service) | 2026-04-27 ([`b7c15e6`](https://github.com/emporix/api-references/commit/b7c15e6)) | — | 17 | 17 | 0 | 0 | 0 |
| [Customer Segment Service](/emporix-cli/coverage/customer-segment-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 36 | 34 | 2 | 0 | 0 |
| [Customer Service (customer-managed)](/emporix-cli/coverage/customer-service-customer-managed) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | — | 25 | 0 | 25 | 0 | 0 |
| [Customer Service (tenant-managed)](/emporix-cli/coverage/customer-service-tenant-managed) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 19 | 19 | 0 | 0 | 0 |
| [Fee Service](/emporix-cli/coverage/fee-service) | 2026-04-16 ([`b290b28`](https://github.com/emporix/api-references/commit/b290b28)) | `v1` | 18 | 18 | 0 | 0 | 0 |
| [IAM Service](/emporix-cli/coverage/iam-service) | 2026-07-01 ([`d42b1ce`](https://github.com/emporix/api-references/commit/d42b1ce)) | — | 42 | 39 | 3 | 0 | 0 |
| [Import Service](/emporix-cli/coverage/import-service) | 2026-10-14 ([`f1f0640`](https://github.com/emporix/api-references/commit/f1f0640)) | — | 24 | 22 | 2 | 0 | 0 |
| [Indexing Service](/emporix-cli/coverage/indexing-service) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | `v1` | 11 | 11 | 0 | 0 | 0 |
| [Invoice Service](/emporix-cli/coverage/invoice-service) | 2025-11-18 ([`322697f`](https://github.com/emporix/api-references/commit/322697f)) | `1.0.0` | 2 | 2 | 0 | 0 | 0 |
| [Label Service](/emporix-cli/coverage/label-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 6 | 6 | 0 | 0 | 0 |
| [Media Service](/emporix-cli/coverage/media-service) | 2026-09-28 ([`6149878`](https://github.com/emporix/api-references/commit/6149878)) | — | 9 | 8 | 1 | 0 | 0 |
| [OAuth Service](/emporix-cli/coverage/oauth-service) | 2026-02-27 ([`48eee2b`](https://github.com/emporix/api-references/commit/48eee2b)) | — | 1 | 1 | 0 | 0 | 0 |
| [Order Service](/emporix-cli/coverage/order-service) | 2026-10-07 ([`1f21665`](https://github.com/emporix/api-references/commit/1f21665)) | — | 20 | 14 | 6 | 0 | 0 |
| [Payment-Gateway Service](/emporix-cli/coverage/payment-gateway-service) | 2026-02-10 ([`e61a2e2`](https://github.com/emporix/api-references/commit/e61a2e2)) | — | 15 | 13 | 2 | 0 | 0 |
| [Price Service](/emporix-cli/coverage/price-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 30 | 29 | 1 | 0 | 0 |
| [Product Service](/emporix-cli/coverage/product-service) | 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4)) | — | 17 | 17 | 0 | 0 | 0 |
| [Quote Service](/emporix-cli/coverage/quote-service) | 2026-10-09 ([`abf90d6`](https://github.com/emporix/api-references/commit/abf90d6)) | — | 12 | 12 | 0 | 0 | 0 |
| [Returns Service](/emporix-cli/coverage/returns-service) | 2026-03-20 ([`b658888`](https://github.com/emporix/api-references/commit/b658888)) | — | 6 | 6 | 0 | 0 | 0 |
| [Reward Points Service](/emporix-cli/coverage/reward-points-service) | 2025-11-17 ([`62e8b2b`](https://github.com/emporix/api-references/commit/62e8b2b)) | — | 14 | 11 | 3 | 0 | 0 |
| [Schema Service](/emporix-cli/coverage/schema-service) | 2026-10-12 ([`f0072a6`](https://github.com/emporix/api-references/commit/f0072a6)) | — | 31 | 31 | 0 | 0 | 0 |
| [Search Service](/emporix-cli/coverage/search-service) | 2026-10-09 ([`8fca29b`](https://github.com/emporix/api-references/commit/8fca29b)) | — | 15 | 15 | 0 | 0 | 0 |
| [Sequential ID Service](/emporix-cli/coverage/sequential-id-service) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | — | 8 | 8 | 0 | 0 | 0 |
| [Session Context Service](/emporix-cli/coverage/session-context-service) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | — | 8 | 4 | 4 | 0 | 0 |
| [Shipping Service](/emporix-cli/coverage/shipping-service) | 2026-09-11 ([`eeddb11`](https://github.com/emporix/api-references/commit/eeddb11)) | — | 45 | 44 | 1 | 0 | 0 |
| [Shopping List Service](/emporix-cli/coverage/shopping-list-service) | 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee)) | — | 5 | 5 | 0 | 0 | 0 |
| [Site Settings Service](/emporix-cli/coverage/site-settings-service) | 2026-09-28 ([`7c702b5`](https://github.com/emporix/api-references/commit/7c702b5)) | — | 13 | 13 | 0 | 0 | 0 |
| [Tax Service](/emporix-cli/coverage/tax-service) | 2025-11-15 ([`4ed37c9`](https://github.com/emporix/api-references/commit/4ed37c9)) | — | 6 | 6 | 0 | 0 | 0 |
| [Unit Handling Service](/emporix-cli/coverage/unit-handling-service) | 2026-04-17 ([`ce5a5d6`](https://github.com/emporix/api-references/commit/ce5a5d6)) | — | 9 | 9 | 0 | 0 | 0 |
| [Vendor Service](/emporix-cli/coverage/vendor-service) | 2025-11-14 ([`a0af6e5`](https://github.com/emporix/api-references/commit/a0af6e5)) | — | 11 | 11 | 0 | 0 | 0 |
| [Webhook Service](/emporix-cli/coverage/webhook-service) | 2026-08-11 ([`279d397`](https://github.com/emporix/api-references/commit/279d397)) | — | 10 | 10 | 0 | 0 | 0 |

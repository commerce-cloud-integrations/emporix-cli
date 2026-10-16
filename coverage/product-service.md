# Product Service

> Which Product Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`products-labels-and-brands/product-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/products-labels-and-brands/product-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 17 (17 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /product/{tenant}/products/search` | Built | [`epx products search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-search) |
| `GET /product/{tenant}/products` | Built | [`epx products list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-list) |
| `POST /product/{tenant}/products` | Built | [`epx products create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-create) |
| `POST /product/{tenant}/products/bulk` | Built | [`epx products bulk-create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-bulk-create) |
| `PUT /product/{tenant}/products/bulk` | Built | [`epx products bulk-update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-bulk-update), [`epx products import`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-import) |
| `POST /product/{tenant}/products/recalculate` | Built | [`epx products recalculate`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-recalculate) |
| `GET /product/{tenant}/products/recalculate/jobs` | Built | [`epx products recalculation-jobs list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-recalculation-jobs-list) |
| `GET /product/{tenant}/products/recalculate/jobs/{jobId}` | Built | [`epx products recalculation-jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-recalculation-jobs-get) |
| `GET /product/{tenant}/products/{productId}` | Built | [`epx products get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-get) |
| `PUT /product/{tenant}/products/{productId}` | Built | [`epx products update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-update) |
| `PATCH /product/{tenant}/products/{productId}` | Built | [`epx products patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-patch) |
| `DELETE /product/{tenant}/products/{productId}` | Built | [`epx products delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/products#epx-products-delete) |
| `GET /product/{tenant}/product-templates` | Built | [`epx product-templates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates#epx-product-templates-list) |
| `POST /product/{tenant}/product-templates` | Built | [`epx product-templates create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates#epx-product-templates-create) |
| `GET /product/{tenant}/product-templates/{product-template-id}` | Built | [`epx product-templates get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates#epx-product-templates-get) |
| `PUT /product/{tenant}/product-templates/{product-template-id}` | Built | [`epx product-templates update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates#epx-product-templates-update) |
| `DELETE /product/{tenant}/product-templates/{product-template-id}` | Built | [`epx product-templates delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/product-templates#epx-product-templates-delete) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

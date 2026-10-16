# Category

> Which Category operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`catalogs-and-categories/category-tree/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/catalogs-and-categories/category-tree/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 24 (23 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /category/{tenant}/assignments/references/{referenceId}` | Built | [`epx categories for-product`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-for-product) |
| `DELETE /category/{tenant}/assignments/references/{referenceId}` | Built | [`epx categories assignments delete-product-everywhere`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-delete-product-everywhere) |
| `PUT /category/{tenant}/categories/{categoryId}/assignments/references/bulk` | Built | [`epx categories assignments bulk-update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-bulk-update) |
| `GET /category/{tenant}/categories` | Built | [`epx categories list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-list) |
| `POST /category/{tenant}/categories` | Built | [`epx categories create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-create) |
| `POST /category/{tenant}/categories/search` | Built | [`epx categories search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-search) |
| `GET /category/{tenant}/categories/categoryTree` | Not applicable | Deprecated by Emporix and no longer served (Emporix answers it as an unknown category, 404); `epx category-trees` reads the category trees. |
| `GET /category/{tenant}/categories/{categoryId}` | Built | [`epx categories get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-get) |
| `PUT /category/{tenant}/categories/{categoryId}` | Built | [`epx categories update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-update) |
| `DELETE /category/{tenant}/categories/{categoryId}` | Built | [`epx categories delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-delete) |
| `PATCH /category/{tenant}/categories/{categoryId}` | Built | [`epx categories patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-patch) |
| `POST /category/{tenant}/categories/{categoryId}/assignments/bulk` | Built | [`epx categories assignments bulk-create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-bulk-create) |
| `GET /category/{tenant}/categories/{categoryId}/assignments` | Built | [`epx categories assignments list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-list) |
| `POST /category/{tenant}/categories/{categoryId}/assignments` | Built | [`epx categories assignments create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-create) |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments` | Built | [`epx categories assignments delete-all`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-delete-all) |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments/{assignmentId}` | Built | [`epx categories assignments delete-assignment`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-delete-assignment) |
| `PUT /category/{tenant}/categories/{categoryId}/assignments/references/{referenceId}` | Built | [`epx categories assignments update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-update) |
| `DELETE /category/{tenant}/categories/{categoryId}/assignments/references/{referenceId}` | Built | [`epx categories assignments delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-assignments-delete) |
| `GET /category/{tenant}/categories/{categoryId}/parents` | Built | [`epx categories parents`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-parents) |
| `GET /category/{tenant}/categories/{categoryId}/subcategories` | Built | [`epx categories subcategories`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/categories#epx-categories-subcategories) |
| `GET /category/{tenant}/category-trees` | Built | [`epx category-trees list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees#epx-category-trees-list) |
| `POST /category/{tenant}/category-trees/search` | Built | [`epx category-trees search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees#epx-category-trees-search) |
| `GET /category/{tenant}/category-trees/{categoryId}` | Built | [`epx category-trees get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees#epx-category-trees-get) |
| `POST /category/{tenant}/category-trees/{rootCategoryId}/rebuild` | Built | [`epx category-trees rebuild`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/category-trees#epx-category-trees-rebuild) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

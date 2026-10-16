# Customer Service (tenant-managed)

> Which Customer Service (tenant-managed) operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/customer-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 19 (19 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /customer/{tenant}/customers` | Built | [`epx customers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-list) |
| `POST /customer/{tenant}/customers` | Built | [`epx customers create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-create) |
| `POST /customer/{tenant}/customers/search` | Built | [`epx customers search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-search) |
| `PUT /customer/{tenant}/customers/{customerNumber}` | Built | [`epx customers update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-update) |
| `PATCH /customer/{tenant}/customers/{customerNumber}` | Built | [`epx customers patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-patch) |
| `GET /customer/{tenant}/customers/{customerNumber}` | Built | [`epx customers get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-get) |
| `DELETE /customer/{tenant}/customers/{customerNumber}` | Built | [`epx customers delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-delete) |
| `GET /customer/{tenant}/customers/{customerNumber}/addresses` | Built | [`epx customers addresses list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-list) |
| `POST /customer/{tenant}/customers/{customerNumber}/addresses` | Built | [`epx customers addresses create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-create) |
| `PUT /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | Built | [`epx customers addresses update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-update) |
| `PATCH /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | Built | [`epx customers addresses patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-patch) |
| `GET /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | Built | [`epx customers addresses get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-get) |
| `DELETE /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}` | Built | [`epx customers addresses delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-delete) |
| `POST /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}/tags` | Built | [`epx customers addresses add-tags`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-add-tags) |
| `DELETE /customer/{tenant}/customers/{customerNumber}/addresses/{addressId}/tags` | Built | [`epx customers addresses remove-tags`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-addresses-remove-tags) |
| `GET /customer/{tenant}/config/password-migration-retention` | Built | [`epx customers password-migration get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-password-migration-get) |
| `POST /customer/{tenant}/config/password-migration-retention` | Built | [`epx customers password-migration update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-password-migration-update) |
| `DELETE /customer/{tenant}/config/password-migration-retention` | Built | [`epx customers password-migration delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-password-migration-delete) |
| `POST /customer/{tenant}/customers/import` | Built | [`epx customers import`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customers#epx-customers-import) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

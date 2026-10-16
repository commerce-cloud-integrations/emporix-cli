# Customer Management Service

> Which Customer Management Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/client-management/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/client-management/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-27 ([`b7c15e6`](https://github.com/emporix/api-references/commit/b7c15e6))
- Version: not stated in the reference file
- Operations: 17 (17 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /customer-management/{tenant}/legal-entities` | Built | [`epx legal-entities create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-create) |
| `GET /customer-management/{tenant}/legal-entities` | Built | [`epx legal-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-list) |
| `POST /customer-management/{tenant}/legal-entities/search` | Built | [`epx legal-entities search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-search) |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-get) |
| `PUT /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-update) |
| `DELETE /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-delete) |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}/parent-hierarchy` | Built | [`epx legal-entities hierarchy`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/legal-entities#epx-legal-entities-hierarchy) |
| `POST /customer-management/{tenant}/contact-assignments` | Built | [`epx contact-assignments create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments#epx-contact-assignments-create) |
| `GET /customer-management/{tenant}/contact-assignments` | Built | [`epx contact-assignments list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments#epx-contact-assignments-list) |
| `GET /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments#epx-contact-assignments-get) |
| `PUT /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments#epx-contact-assignments-update) |
| `DELETE /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/contact-assignments#epx-contact-assignments-delete) |
| `POST /customer-management/{tenant}/locations` | Built | [`epx customer-locations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations#epx-customer-locations-create) |
| `GET /customer-management/{tenant}/locations` | Built | [`epx customer-locations list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations#epx-customer-locations-list) |
| `GET /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations#epx-customer-locations-get) |
| `PUT /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations#epx-customer-locations-update) |
| `DELETE /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/customer-locations#epx-customer-locations-delete) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

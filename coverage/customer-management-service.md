# Customer Management Service

> Which Customer Management Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/client-management/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/client-management/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-04-27 ([`b7c15e6`](https://github.com/emporix/api-references/commit/b7c15e6))
- Version: not stated in the reference file
- Operations: 17 (17 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /customer-management/{tenant}/legal-entities` | Built | [`epx legal-entities create`](/emporix-cli/commands/legal-entities#epx-legal-entities-create) |
| `GET /customer-management/{tenant}/legal-entities` | Built | [`epx legal-entities list`](/emporix-cli/commands/legal-entities#epx-legal-entities-list) |
| `POST /customer-management/{tenant}/legal-entities/search` | Built | [`epx legal-entities search`](/emporix-cli/commands/legal-entities#epx-legal-entities-search) |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities get`](/emporix-cli/commands/legal-entities#epx-legal-entities-get) |
| `PUT /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities update`](/emporix-cli/commands/legal-entities#epx-legal-entities-update) |
| `DELETE /customer-management/{tenant}/legal-entities/{legalEntityId}` | Built | [`epx legal-entities delete`](/emporix-cli/commands/legal-entities#epx-legal-entities-delete) |
| `GET /customer-management/{tenant}/legal-entities/{legalEntityId}/parent-hierarchy` | Built | [`epx legal-entities hierarchy`](/emporix-cli/commands/legal-entities#epx-legal-entities-hierarchy) |
| `POST /customer-management/{tenant}/contact-assignments` | Built | [`epx contact-assignments create`](/emporix-cli/commands/contact-assignments#epx-contact-assignments-create) |
| `GET /customer-management/{tenant}/contact-assignments` | Built | [`epx contact-assignments list`](/emporix-cli/commands/contact-assignments#epx-contact-assignments-list) |
| `GET /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments get`](/emporix-cli/commands/contact-assignments#epx-contact-assignments-get) |
| `PUT /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments update`](/emporix-cli/commands/contact-assignments#epx-contact-assignments-update) |
| `DELETE /customer-management/{tenant}/contact-assignments/{contactAssignmentId}` | Built | [`epx contact-assignments delete`](/emporix-cli/commands/contact-assignments#epx-contact-assignments-delete) |
| `POST /customer-management/{tenant}/locations` | Built | [`epx customer-locations create`](/emporix-cli/commands/customer-locations#epx-customer-locations-create) |
| `GET /customer-management/{tenant}/locations` | Built | [`epx customer-locations list`](/emporix-cli/commands/customer-locations#epx-customer-locations-list) |
| `GET /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations get`](/emporix-cli/commands/customer-locations#epx-customer-locations-get) |
| `PUT /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations update`](/emporix-cli/commands/customer-locations#epx-customer-locations-update) |
| `DELETE /customer-management/{tenant}/locations/{locationId}` | Built | [`epx customer-locations delete`](/emporix-cli/commands/customer-locations#epx-customer-locations-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).

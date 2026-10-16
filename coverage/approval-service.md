# Approval Service

> Which Approval Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/approval-service/approval-api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/approval-service/approval-api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 7 (0 built, 7 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /approval/{tenant}/approvals` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /approval/{tenant}/approvals` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /approval/{tenant}/approvals/{approvalId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `PATCH /approval/{tenant}/approvals/{approvalId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `DELETE /approval/{tenant}/approvals/{approvalId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /approval/{tenant}/approval/permitted` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /approval/{tenant}/search/users` | Not applicable | Needs a signed-in customer's token, which a service token is not. |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

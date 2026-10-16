# Session Context Service

> Which Session Context Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`users-and-permissions/session-context/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/users-and-permissions/session-context/api-reference/api.yaml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee))
- Version: not stated in the reference file
- Operations: 8 (4 built, 4 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /session-context/{tenant}/context/{sessionId}` | Built | [`epx session-contexts get`](/emporix-cli/commands/session-contexts#epx-session-contexts-get) |
| `PUT /session-context/{tenant}/context/{sessionId}` | Built | [`epx session-contexts update`](/emporix-cli/commands/session-contexts#epx-session-contexts-update) |
| `DELETE /session-context/{tenant}/context/{sessionId}/attributes/{attributeName}` | Built | [`epx session-contexts attributes delete`](/emporix-cli/commands/session-contexts#epx-session-contexts-attributes-delete) |
| `POST /session-context/{tenant}/context/{sessionId}/attributes` | Built | [`epx session-contexts attributes create`](/emporix-cli/commands/session-contexts#epx-session-contexts-attributes-create) |
| `GET /session-context/{tenant}/me/context` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `PATCH /session-context/{tenant}/me/context` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `DELETE /session-context/{tenant}/me/context/attributes/{attributeName}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /session-context/{tenant}/me/context/attributes` | Not applicable | Needs a signed-in customer's token, which a service token is not. |

All services: [Emporix API coverage](/emporix-cli/coverage).

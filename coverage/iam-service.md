# IAM Service

> Which IAM Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`users-and-permissions/iam/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/users-and-permissions/iam/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-07-01 ([`d42b1ce`](https://github.com/emporix/api-references/commit/d42b1ce))
- Version: not stated in the reference file
- Operations: 42 (39 built, 3 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /iam/{tenant}/access-controls` | Built | [`epx iam access-controls list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-access-controls-list) |
| `GET /iam/{tenant}/access-controls/{accessControlId}` | Built | [`epx iam access-controls get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-access-controls-get) |
| `PUT /iam/{tenant}/access-controls/{accessControlId}` | Built | [`epx iam access-controls update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-access-controls-update) |
| `DELETE /iam/{tenant}/access-controls/{accessControlId}` | Built | [`epx iam access-controls delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-access-controls-delete) |
| `GET /iam/{tenant}/scopes` | Built | [`epx iam scopes list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-scopes-list) |
| `GET /iam/{tenant}/scopes/{scopeId}` | Built | [`epx iam scopes get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-scopes-get) |
| `PUT /iam/{tenant}/scopes/{scopeId}` | Built | [`epx iam scopes update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-scopes-update) |
| `DELETE /iam/{tenant}/scopes/{scopeId}` | Built | [`epx iam scopes delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-scopes-delete) |
| `GET /iam/{tenant}/groups` | Built | [`epx iam groups list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-list) |
| `POST /iam/{tenant}/groups` | Built | [`epx iam groups create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-create) |
| `GET /iam/{tenant}/groups/{groupId}` | Built | [`epx iam groups get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-get) |
| `DELETE /iam/{tenant}/groups/{groupId}` | Built | [`epx iam groups delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-delete) |
| `PUT /iam/{tenant}/groups/{groupId}` | Built | [`epx iam groups update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-update) |
| `GET /iam/{tenant}/groups/{groupId}/access-controls` | Built | [`epx iam groups access-controls`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-access-controls) |
| `GET /iam/{tenant}/groups/{groupId}/users` | Built | [`epx iam groups users list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-users-list) |
| `POST /iam/{tenant}/groups/{groupId}/users` | Built | [`epx iam groups users create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-users-create) |
| `DELETE /iam/{tenant}/groups/{groupId}/users` | Built | [`epx iam groups users delete-all`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-users-delete-all) |
| `PUT /iam/{tenant}/groups/{groupId}/users/{userType}/{userId}` | Built | [`epx iam groups users update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-users-update) |
| `DELETE /iam/{tenant}/groups/{groupId}/users/{userId}` | Built | [`epx iam groups users delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-groups-users-delete) |
| `GET /iam/{tenant}/permissions` | Built | [`epx iam permissions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-permissions-list) |
| `GET /iam/{tenant}/permissions/{permissionId}` | Built | [`epx iam permissions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-permissions-get) |
| `GET /iam/{tenant}/resources` | Built | [`epx iam resources list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-resources-list) |
| `GET /iam/{tenant}/resources/{resourceId}` | Built | [`epx iam resources get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-resources-get) |
| `GET /iam/{tenant}/templates` | Built | [`epx iam templates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-templates-list) |
| `GET /iam/{tenant}/roles` | Built | [`epx iam roles list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-roles-list) |
| `GET /iam/{tenant}/roles/{roleId}` | Built | [`epx iam roles get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-roles-get) |
| `GET /iam/{tenant}/users/{userId}/access-controls` | Built | [`epx iam users access-controls`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-access-controls) |
| `GET /iam/{tenant}/users/{userId}/access-controls/{resourceId}` | Built | [`epx iam users resource-access-controls`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-resource-access-controls) |
| `GET /iam/{tenant}/users/{userId}/groups` | Built | [`epx iam users groups list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-groups-list) |
| `DELETE /iam/{tenant}/users/{userId}/groups` | Built | [`epx iam users groups delete-all`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-groups-delete-all) |
| `GET /iam/{tenant}/users/{userId}/groups/{groupId}` | Built | [`epx iam users groups get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-groups-get) |
| `GET /iam/{tenant}/users/{userId}/permissions/{resourceId}` | Built | [`epx iam users resource-permissions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-resource-permissions) |
| `GET /iam/{tenant}/users/{userId}/scopes` | Built | [`epx iam users scopes`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-scopes) |
| `GET /iam/{tenant}/users/me/access-controls` | Not applicable | Describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID. |
| `GET /iam/{tenant}/users/me/scopes` | Not applicable | Describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID. |
| `GET /iam/{tenant}/users` | Built | [`epx iam users list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-list) |
| `POST /iam/{tenant}/users` | Built | [`epx iam users create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-create) |
| `GET /iam/{tenant}/users/me` | Not applicable | Describes the signed-in user, which a service token is not (Emporix answers 400, constraint validation failed); `epx iam users get`, `scopes` and `access-controls` read any user by ID. |
| `GET /iam/{tenant}/users/{userId}` | Built | [`epx iam users get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-get) |
| `PUT /iam/{tenant}/users/{userId}` | Built | [`epx iam users update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-update) |
| `DELETE /iam/{tenant}/users/{userId}` | Built | [`epx iam users delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-delete) |
| `GET /iam/{tenant}/users/vendors/{vendorId}` | Built | [`epx iam users for-vendor`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/iam#epx-iam-users-for-vendor) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

# Webhook Service

> Which Webhook Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`webhooks/webhook-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/webhooks/webhook-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-08-11 ([`279d397`](https://github.com/emporix/api-references/commit/279d397))
- Version: not stated in the reference file
- Operations: 10 (10 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /webhook/{tenant}/config` | Built | [`epx webhooks configurations create`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-create) |
| `GET /webhook/{tenant}/config` | Built | [`epx webhooks configurations list`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-list) |
| `PUT /webhook/{tenant}/config/{code}` | Built | [`epx webhooks configurations update`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-update) |
| `PATCH /webhook/{tenant}/config/{code}` | Built | [`epx webhooks configurations patch`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-patch) |
| `GET /webhook/{tenant}/config/{code}` | Built | [`epx webhooks configurations get`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-get) |
| `DELETE /webhook/{tenant}/config/{code}` | Built | [`epx webhooks configurations delete`](/emporix-cli/commands/webhooks#epx-webhooks-configurations-delete) |
| `PATCH /webhook/{tenant}/event-subscriptions` | Built | [`epx webhooks subscriptions bulk-patch`](/emporix-cli/commands/webhooks#epx-webhooks-subscriptions-bulk-patch) |
| `GET /webhook/{tenant}/event-subscriptions` | Built | [`epx webhooks subscriptions list`](/emporix-cli/commands/webhooks#epx-webhooks-subscriptions-list) |
| `GET /webhook/{tenant}/dashboard-access` | Built | [`epx webhooks dashboard-access`](/emporix-cli/commands/webhooks#epx-webhooks-dashboard-access) |
| `GET /webhook/{tenant}/statistics` | Built | [`epx webhooks statistics`](/emporix-cli/commands/webhooks#epx-webhooks-statistics) |

All services: [Emporix API coverage](/emporix-cli/coverage).

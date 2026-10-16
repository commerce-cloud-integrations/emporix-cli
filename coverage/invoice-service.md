# Invoice Service

> Which Invoice Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`orders/invoice/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/orders/invoice/api-reference/api.yaml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2025-11-18 ([`322697f`](https://github.com/emporix/api-references/commit/322697f))
- Version: `1.0.0`
- Operations: 2 (2 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /invoice/{tenant}/jobs/invoices` | Built | [`epx invoice-jobs create`](/emporix-cli/commands/invoice-jobs#epx-invoice-jobs-create) |
| `GET /invoice/{tenant}/jobs/invoices/{jobId}` | Built | [`epx invoice-jobs get`](/emporix-cli/commands/invoice-jobs#epx-invoice-jobs-get) |

All services: [Emporix API coverage](/emporix-cli/coverage).

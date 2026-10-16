# Checkout Service

> Which Checkout Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`checkout/checkout/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/checkout/checkout/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 1 (0 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /checkout/{tenant}/checkouts/order` | Not applicable | Needs a signed-in customer's token, which a service token is not. |

All services: [Emporix API coverage](/emporix-cli/coverage).

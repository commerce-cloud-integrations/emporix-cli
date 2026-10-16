# Quote Service

> Which Quote Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`quotes/quote/api-reference/api.yaml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/quotes/quote/api-reference/api.yaml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-09 ([`abf90d6`](https://github.com/emporix/api-references/commit/abf90d6))
- Version: not stated in the reference file
- Operations: 12 (12 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /quote/{tenant}/quotes` | Built | [`epx quotes list`](/emporix-cli/commands/quotes#epx-quotes-list) |
| `POST /quote/{tenant}/quotes` | Built | [`epx quotes create`](/emporix-cli/commands/quotes#epx-quotes-create) |
| `POST /quote/{tenant}/quotes/{quoteId}/pdf` | Built | [`epx quotes pdf`](/emporix-cli/commands/quotes#epx-quotes-pdf) |
| `GET /quote/{tenant}/quotes/{quoteId}/history` | Built | [`epx quotes history`](/emporix-cli/commands/quotes#epx-quotes-history) |
| `GET /quote/{tenant}/quotes/{quoteId}` | Built | [`epx quotes get`](/emporix-cli/commands/quotes#epx-quotes-get) |
| `PATCH /quote/{tenant}/quotes/{quoteId}` | Built | [`epx quotes patch`](/emporix-cli/commands/quotes#epx-quotes-patch) |
| `DELETE /quote/{tenant}/quotes/{quoteId}` | Built | [`epx quotes delete`](/emporix-cli/commands/quotes#epx-quotes-delete) |
| `POST /quote/{tenant}/quote-reasons` | Built | [`epx quotes reasons create`](/emporix-cli/commands/quotes#epx-quotes-reasons-create) |
| `GET /quote/{tenant}/quote-reasons` | Built | [`epx quotes reasons list`](/emporix-cli/commands/quotes#epx-quotes-reasons-list) |
| `GET /quote/{tenant}/quote-reasons/{quoteReasonId}` | Built | [`epx quotes reasons get`](/emporix-cli/commands/quotes#epx-quotes-reasons-get) |
| `PUT /quote/{tenant}/quote-reasons/{quoteReasonId}` | Built | [`epx quotes reasons update`](/emporix-cli/commands/quotes#epx-quotes-reasons-update) |
| `DELETE /quote/{tenant}/quote-reasons/{quoteReasonId}` | Built | [`epx quotes reasons delete`](/emporix-cli/commands/quotes#epx-quotes-reasons-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).

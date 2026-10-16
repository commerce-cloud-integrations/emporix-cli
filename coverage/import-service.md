# Import Service

> Which Import Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`utilities/import-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/utilities/import-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-14 ([`f1f0640`](https://github.com/emporix/api-references/commit/f1f0640))
- Version: not stated in the reference file
- Operations: 24 (22 built, 2 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /importtool/{tenant}/configs` | Built | [`epx imports configurations list`](/emporix-cli/commands/imports#epx-imports-configurations-list) |
| `GET /importtool/{tenant}/configs/{id}` | Built | [`epx imports configurations get`](/emporix-cli/commands/imports#epx-imports-configurations-get) |
| `GET /importtool/{tenant}/configs/{configId}/streams` | Built | [`epx imports configurations streams`](/emporix-cli/commands/imports#epx-imports-configurations-streams) |
| `GET /importtool/{tenant}/streams/{id}` | Built | [`epx imports streams get`](/emporix-cli/commands/imports#epx-imports-streams-get) |
| `GET /importtool/{tenant}/configs/{configId}/schedule` | Built | [`epx imports configurations schedule`](/emporix-cli/commands/imports#epx-imports-configurations-schedule) |
| `PUT /importtool/{tenant}/configs/{configId}/schedule` | Built | [`epx imports configurations update-schedule`](/emporix-cli/commands/imports#epx-imports-configurations-update-schedule) |
| `DELETE /importtool/{tenant}/configs/{configId}/schedule` | Built | [`epx imports configurations delete-schedule`](/emporix-cli/commands/imports#epx-imports-configurations-delete-schedule) |
| `GET /importtool/{tenant}/configs/{configId}/stream-order` | Built | [`epx imports configurations stream-order`](/emporix-cli/commands/imports#epx-imports-configurations-stream-order) |
| `POST /importtool/{tenant}/configs/{configId}/runs` | Built | [`epx imports configurations trigger`](/emporix-cli/commands/imports#epx-imports-configurations-trigger) |
| `GET /importtool/{tenant}/configs/{configId}/runs` | Built | [`epx imports configurations runs`](/emporix-cli/commands/imports#epx-imports-configurations-runs) |
| `GET /importtool/{tenant}/runs/{runId}` | Built | [`epx imports runs get`](/emporix-cli/commands/imports#epx-imports-runs-get), [`epx imports runs streams`](/emporix-cli/commands/imports#epx-imports-runs-streams) |
| `GET /importtool/{tenant}/runs/{runId}/events` | Not applicable | Answers with Server-Sent Events, a stream held open until the run finishes; `epx imports runs get` and `epx imports runs streams` read the run's progress per stream as JSON. |
| `POST /importtool/{tenant}/runs/{runId}/cancel` | Built | [`epx imports runs cancel`](/emporix-cli/commands/imports#epx-imports-runs-cancel) |
| `GET /importtool/{tenant}/runs/{runId}/errors` | Built | [`epx imports runs errors`](/emporix-cli/commands/imports#epx-imports-runs-errors) |
| `GET /importtool/{tenant}/runs/{runId}/diagnostics` | Built | [`epx imports runs diagnostics`](/emporix-cli/commands/imports#epx-imports-runs-diagnostics) |
| `GET /importtool/{tenant}/runs/{runId}/diagnostics/csv` | Not applicable | Answers with a CSV file; `epx imports runs diagnostics` reads the same rows as JSON. |
| `GET /importtool/{tenant}/data/types` | Built | [`epx imports types`](/emporix-cli/commands/imports#epx-imports-types) |
| `GET /importtool/{tenant}/data/records` | Built | [`epx imports records`](/emporix-cli/commands/imports#epx-imports-records) |
| `GET /importtool/{tenant}/data/streams/{streamId}/records` | Built | [`epx imports streams records`](/emporix-cli/commands/imports#epx-imports-streams-records) |
| `GET /importtool/{tenant}/stats` | Built | [`epx imports statistics`](/emporix-cli/commands/imports#epx-imports-statistics) |
| `GET /importtool/{tenant}/dashboard/job-groups` | Built | [`epx imports job-groups`](/emporix-cli/commands/imports#epx-imports-job-groups) |
| `GET /importtool/{tenant}/settings/health-thresholds` | Built | [`epx imports health-thresholds`](/emporix-cli/commands/imports#epx-imports-health-thresholds) |
| `GET /importtool/{tenant}/license` | Built | [`epx imports license`](/emporix-cli/commands/imports#epx-imports-license) |
| `POST /importtool/{tenant}/runs/{runId}/retry` | Built | [`epx imports runs retry`](/emporix-cli/commands/imports#epx-imports-runs-retry) |

All services: [Emporix API coverage](/emporix-cli/coverage).

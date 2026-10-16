# AI Service

> Which AI Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`artificial-intelligence/ai-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/artificial-intelligence/ai-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-06 ([`9b456b6`](https://github.com/emporix/api-references/commit/9b456b6))
- Version: not stated in the reference file
- Operations: 57 (56 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /ai-service/{tenant}/texts` | Built | [`epx ai generate-text`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-generate-text) |
| `POST /ai-service/{tenant}/completions` | Built | [`epx ai generate-completion`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-generate-completion) |
| `GET /ai-service/{tenant}/jobs` | Built | [`epx ai jobs list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-list) |
| `POST /ai-service/{tenant}/jobs/search` | Built | [`epx ai jobs search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-search) |
| `GET /ai-service/{tenant}/jobs/{jobId}` | Built | [`epx ai jobs get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-get) |
| `DELETE /ai-service/{tenant}/jobs/{jobId}` | Built | [`epx ai jobs delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-delete) |
| `POST /ai-service/{tenant}/agentic/chat` | Built | [`epx ai chat`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-chat) |
| `POST /ai-service/{tenant}/agentic/chat-stream` | Not applicable | Answers with Server-Sent Events, a stream held open while the agent replies; `epx ai chat` prints the same reply whole, and `epx ai chat-async` runs it as a job. |
| `POST /ai-service/{tenant}/agentic/chat-async` | Built | [`epx ai chat-async`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-chat-async) |
| `POST /ai-service/{tenant}/agentic/{agentId}/attachments` | Built | [`epx ai agents attach`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-attach) |
| `GET /ai-service/{tenant}/agentic/templates` | Built | [`epx ai templates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-list) |
| `POST /ai-service/{tenant}/agentic/templates/search` | Built | [`epx ai templates search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-search) |
| `POST /ai-service/{tenant}/agentic/templates/{templateId}/agents` | Built | [`epx ai agents create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-create) |
| `GET /ai-service/{tenant}/agentic/agents` | Built | [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) |
| `POST /ai-service/{tenant}/agentic/agents/search` | Built | [`epx ai agents search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-search) |
| `GET /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-get) |
| `PUT /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-update) |
| `PATCH /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-patch) |
| `DELETE /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-delete) |
| `GET /ai-service/{tenant}/agentic/conversations` | Built | [`epx ai conversations list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-conversations-list) |
| `POST /ai-service/{tenant}/agentic/conversations/search` | Built | [`epx ai conversations search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-conversations-search) |
| `GET /ai-service/{tenant}/agentic/tools` | Built | [`epx ai tools list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-list) |
| `POST /ai-service/{tenant}/agentic/tools/search` | Built | [`epx ai tools search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-search) |
| `GET /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-get) |
| `PUT /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-update) |
| `PATCH /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-patch) |
| `DELETE /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-delete) |
| `GET /ai-service/{tenant}/agentic/tokens` | Built | [`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list) |
| `POST /ai-service/{tenant}/agentic/tokens/search` | Built | [`epx ai tokens search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-search) |
| `GET /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-get) |
| `PUT /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-update) |
| `PATCH /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-patch) |
| `DELETE /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-delete) |
| `GET /ai-service/{tenant}/agentic/oauths` | Built | [`epx ai oauths list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-list) |
| `POST /ai-service/{tenant}/agentic/oauths/search` | Built | [`epx ai oauths search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-search) |
| `GET /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-get) |
| `PUT /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-update) |
| `PATCH /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-patch) |
| `DELETE /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-delete) |
| `GET /ai-service/{tenant}/agentic/models` | Built | [`epx ai models`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-models) |
| `GET /ai-service/{tenant}/agentic/mcp-servers` | Built | [`epx ai mcp-servers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-list) |
| `POST /ai-service/{tenant}/agentic/mcp-servers/search` | Built | [`epx ai mcp-servers search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-search) |
| `GET /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-get) |
| `PUT /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-update) |
| `PATCH /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-patch) |
| `DELETE /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers delete`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-delete) |
| `GET /ai-service/{tenant}/agentic/commerce-events` | Built | [`epx ai commerce-events`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-commerce-events) |
| `POST /ai-service/{tenant}/agentic/agents/export` | Built | [`epx ai agents export`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-export) |
| `POST /ai-service/{tenant}/agentic/agents/import` | Built | [`epx ai agents import`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-import) |
| `GET /ai-service/{tenant}/agentic/logs/requests` | Built | [`epx ai logs requests list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-requests-list) |
| `GET /ai-service/{tenant}/agentic/logs/requests/{requestId}` | Built | [`epx ai logs requests get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-requests-get) |
| `POST /ai-service/{tenant}/agentic/logs/requests/search` | Built | [`epx ai logs requests search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-requests-search) |
| `GET /ai-service/{tenant}/agentic/logs/sessions` | Built | [`epx ai logs sessions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-list) |
| `GET /ai-service/{tenant}/agentic/logs/sessions/{sessionId}` | Built | [`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get) |
| `POST /ai-service/{tenant}/agentic/logs/sessions/search` | Built | [`epx ai logs sessions search`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-search) |
| `GET /ai-service/{tenant}/agentic/analytics` | Built | [`epx ai analytics tenant`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-analytics-tenant), [`epx ai analytics agent`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-analytics-agent) |
| `GET /ai-service/{tenant}/agentic/analytics/executions` | Built | [`epx ai analytics executions`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-analytics-executions) |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).

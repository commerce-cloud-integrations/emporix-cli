# AI Service

> Which AI Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`artificial-intelligence/ai-service/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/artificial-intelligence/ai-service/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-06 ([`9b456b6`](https://github.com/emporix/api-references/commit/9b456b6))
- Version: not stated in the reference file
- Operations: 57 (56 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /ai-service/{tenant}/texts` | Built | [`epx ai generate-text`](/emporix-cli/commands/ai#epx-ai-generate-text) |
| `POST /ai-service/{tenant}/completions` | Built | [`epx ai generate-completion`](/emporix-cli/commands/ai#epx-ai-generate-completion) |
| `GET /ai-service/{tenant}/jobs` | Built | [`epx ai jobs list`](/emporix-cli/commands/ai#epx-ai-jobs-list) |
| `POST /ai-service/{tenant}/jobs/search` | Built | [`epx ai jobs search`](/emporix-cli/commands/ai#epx-ai-jobs-search) |
| `GET /ai-service/{tenant}/jobs/{jobId}` | Built | [`epx ai jobs get`](/emporix-cli/commands/ai#epx-ai-jobs-get) |
| `DELETE /ai-service/{tenant}/jobs/{jobId}` | Built | [`epx ai jobs delete`](/emporix-cli/commands/ai#epx-ai-jobs-delete) |
| `POST /ai-service/{tenant}/agentic/chat` | Built | [`epx ai chat`](/emporix-cli/commands/ai#epx-ai-chat) |
| `POST /ai-service/{tenant}/agentic/chat-stream` | Not applicable | Answers with Server-Sent Events, a stream held open while the agent replies; `epx ai chat` prints the same reply whole, and `epx ai chat-async` runs it as a job. |
| `POST /ai-service/{tenant}/agentic/chat-async` | Built | [`epx ai chat-async`](/emporix-cli/commands/ai#epx-ai-chat-async) |
| `POST /ai-service/{tenant}/agentic/{agentId}/attachments` | Built | [`epx ai agents attach`](/emporix-cli/commands/ai#epx-ai-agents-attach) |
| `GET /ai-service/{tenant}/agentic/templates` | Built | [`epx ai templates list`](/emporix-cli/commands/ai#epx-ai-templates-list) |
| `POST /ai-service/{tenant}/agentic/templates/search` | Built | [`epx ai templates search`](/emporix-cli/commands/ai#epx-ai-templates-search) |
| `POST /ai-service/{tenant}/agentic/templates/{templateId}/agents` | Built | [`epx ai agents create`](/emporix-cli/commands/ai#epx-ai-agents-create) |
| `GET /ai-service/{tenant}/agentic/agents` | Built | [`epx ai agents list`](/emporix-cli/commands/ai#epx-ai-agents-list) |
| `POST /ai-service/{tenant}/agentic/agents/search` | Built | [`epx ai agents search`](/emporix-cli/commands/ai#epx-ai-agents-search) |
| `GET /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents get`](/emporix-cli/commands/ai#epx-ai-agents-get) |
| `PUT /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents update`](/emporix-cli/commands/ai#epx-ai-agents-update) |
| `PATCH /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents patch`](/emporix-cli/commands/ai#epx-ai-agents-patch) |
| `DELETE /ai-service/{tenant}/agentic/agents/{agentId}` | Built | [`epx ai agents delete`](/emporix-cli/commands/ai#epx-ai-agents-delete) |
| `GET /ai-service/{tenant}/agentic/conversations` | Built | [`epx ai conversations list`](/emporix-cli/commands/ai#epx-ai-conversations-list) |
| `POST /ai-service/{tenant}/agentic/conversations/search` | Built | [`epx ai conversations search`](/emporix-cli/commands/ai#epx-ai-conversations-search) |
| `GET /ai-service/{tenant}/agentic/tools` | Built | [`epx ai tools list`](/emporix-cli/commands/ai#epx-ai-tools-list) |
| `POST /ai-service/{tenant}/agentic/tools/search` | Built | [`epx ai tools search`](/emporix-cli/commands/ai#epx-ai-tools-search) |
| `GET /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools get`](/emporix-cli/commands/ai#epx-ai-tools-get) |
| `PUT /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools update`](/emporix-cli/commands/ai#epx-ai-tools-update) |
| `PATCH /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools patch`](/emporix-cli/commands/ai#epx-ai-tools-patch) |
| `DELETE /ai-service/{tenant}/agentic/tools/{toolId}` | Built | [`epx ai tools delete`](/emporix-cli/commands/ai#epx-ai-tools-delete) |
| `GET /ai-service/{tenant}/agentic/tokens` | Built | [`epx ai tokens list`](/emporix-cli/commands/ai#epx-ai-tokens-list) |
| `POST /ai-service/{tenant}/agentic/tokens/search` | Built | [`epx ai tokens search`](/emporix-cli/commands/ai#epx-ai-tokens-search) |
| `GET /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens get`](/emporix-cli/commands/ai#epx-ai-tokens-get) |
| `PUT /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens update`](/emporix-cli/commands/ai#epx-ai-tokens-update) |
| `PATCH /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens patch`](/emporix-cli/commands/ai#epx-ai-tokens-patch) |
| `DELETE /ai-service/{tenant}/agentic/tokens/{tokenId}` | Built | [`epx ai tokens delete`](/emporix-cli/commands/ai#epx-ai-tokens-delete) |
| `GET /ai-service/{tenant}/agentic/oauths` | Built | [`epx ai oauths list`](/emporix-cli/commands/ai#epx-ai-oauths-list) |
| `POST /ai-service/{tenant}/agentic/oauths/search` | Built | [`epx ai oauths search`](/emporix-cli/commands/ai#epx-ai-oauths-search) |
| `GET /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths get`](/emporix-cli/commands/ai#epx-ai-oauths-get) |
| `PUT /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths update`](/emporix-cli/commands/ai#epx-ai-oauths-update) |
| `PATCH /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths patch`](/emporix-cli/commands/ai#epx-ai-oauths-patch) |
| `DELETE /ai-service/{tenant}/agentic/oauths/{oauthId}` | Built | [`epx ai oauths delete`](/emporix-cli/commands/ai#epx-ai-oauths-delete) |
| `GET /ai-service/{tenant}/agentic/models` | Built | [`epx ai models`](/emporix-cli/commands/ai#epx-ai-models) |
| `GET /ai-service/{tenant}/agentic/mcp-servers` | Built | [`epx ai mcp-servers list`](/emporix-cli/commands/ai#epx-ai-mcp-servers-list) |
| `POST /ai-service/{tenant}/agentic/mcp-servers/search` | Built | [`epx ai mcp-servers search`](/emporix-cli/commands/ai#epx-ai-mcp-servers-search) |
| `GET /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers get`](/emporix-cli/commands/ai#epx-ai-mcp-servers-get) |
| `PUT /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers update`](/emporix-cli/commands/ai#epx-ai-mcp-servers-update) |
| `PATCH /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers patch`](/emporix-cli/commands/ai#epx-ai-mcp-servers-patch) |
| `DELETE /ai-service/{tenant}/agentic/mcp-servers/{mcpServerId}` | Built | [`epx ai mcp-servers delete`](/emporix-cli/commands/ai#epx-ai-mcp-servers-delete) |
| `GET /ai-service/{tenant}/agentic/commerce-events` | Built | [`epx ai commerce-events`](/emporix-cli/commands/ai#epx-ai-commerce-events) |
| `POST /ai-service/{tenant}/agentic/agents/export` | Built | [`epx ai agents export`](/emporix-cli/commands/ai#epx-ai-agents-export) |
| `POST /ai-service/{tenant}/agentic/agents/import` | Built | [`epx ai agents import`](/emporix-cli/commands/ai#epx-ai-agents-import) |
| `GET /ai-service/{tenant}/agentic/logs/requests` | Built | [`epx ai logs requests list`](/emporix-cli/commands/ai#epx-ai-logs-requests-list) |
| `GET /ai-service/{tenant}/agentic/logs/requests/{requestId}` | Built | [`epx ai logs requests get`](/emporix-cli/commands/ai#epx-ai-logs-requests-get) |
| `POST /ai-service/{tenant}/agentic/logs/requests/search` | Built | [`epx ai logs requests search`](/emporix-cli/commands/ai#epx-ai-logs-requests-search) |
| `GET /ai-service/{tenant}/agentic/logs/sessions` | Built | [`epx ai logs sessions list`](/emporix-cli/commands/ai#epx-ai-logs-sessions-list) |
| `GET /ai-service/{tenant}/agentic/logs/sessions/{sessionId}` | Built | [`epx ai logs sessions get`](/emporix-cli/commands/ai#epx-ai-logs-sessions-get) |
| `POST /ai-service/{tenant}/agentic/logs/sessions/search` | Built | [`epx ai logs sessions search`](/emporix-cli/commands/ai#epx-ai-logs-sessions-search) |
| `GET /ai-service/{tenant}/agentic/analytics` | Built | [`epx ai analytics tenant`](/emporix-cli/commands/ai#epx-ai-analytics-tenant), [`epx ai analytics agent`](/emporix-cli/commands/ai#epx-ai-analytics-agent) |
| `GET /ai-service/{tenant}/agentic/analytics/executions` | Built | [`epx ai analytics executions`](/emporix-cli/commands/ai#epx-ai-analytics-executions) |

All services: [Emporix API coverage](/emporix-cli/coverage).

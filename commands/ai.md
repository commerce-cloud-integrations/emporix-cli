# epx ai

> Read and change AI agents and their setup, chat with them, generate texts.

## epx ai

Reads the AI Service's agents and what they are built from: agent templates, tools, tokens,
OAuth configurations and MCP servers, the LLM models of the supported providers and the
commerce events that can trigger an agent; what agents did: their jobs, conversations,
request and session logs and analytics; and which fields the AI RAG Indexer embeds.
Upserts, changes and deletes agents and what they are built from, exports and imports
agents, chats with them and generates texts, which consumes AI credits. [`epx ai rag reindex`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-rag-reindex) embeds the entities of a type again.

Its reads require the ai.agent_read scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx ai [flags]`, `epx ai [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai agents`](#epx-ai-agents)\
  Read, create, upsert, change, delete, export and import AI agents.
- [`epx ai analytics`](#epx-ai-analytics)\
  Read the analytics of agents.
- [`epx ai chat`](#epx-ai-chat)\
  Send a message to an agent and print its reply (consumes AI credits).
- [`epx ai chat-async`](#epx-ai-chat-async)\
  Send a message to an agent, which replies in a job (consumes AI credits).
- [`epx ai commerce-events`](#epx-ai-commerce-events)\
  List the commerce events that can trigger an agent.
- [`epx ai conversations`](#epx-ai-conversations)\
  Read agent conversations.
- [`epx ai generate-completion`](#epx-ai-generate-completion)\
  Generate the next message of a chat history (consumes AI credits).
- [`epx ai generate-text`](#epx-ai-generate-text)\
  Generate a text from a prompt (consumes AI credits).
- [`epx ai jobs`](#epx-ai-jobs)\
  Read and delete agent jobs.
- [`epx ai logs`](#epx-ai-logs)\
  Read the logs of agent requests and sessions.
- [`epx ai mcp-servers`](#epx-ai-mcp-servers)\
  Read, upsert, change and delete MCP servers.
- [`epx ai models`](#epx-ai-models)\
  List the LLM models that agents can use.
- [`epx ai oauths`](#epx-ai-oauths)\
  Read, upsert, change and delete OAuth configurations.
- [`epx ai rag`](#epx-ai-rag)\
  Read which fields the AI RAG Indexer embeds and filters by; reindex.
- [`epx ai templates`](#epx-ai-templates)\
  Read agent templates.
- [`epx ai tokens`](#epx-ai-tokens)\
  Read, upsert, change and delete AI tokens.
- [`epx ai tools`](#epx-ai-tools)\
  Read, upsert, change and delete agent tools.

## epx ai agents

Read, create, upsert, change, delete, export and import AI agents.

Reads AI agents: TYPE is generic, complaint or anti_fraud, PROVIDER and MODEL the LLM an
agent uses and TRIGGERS what starts it (endpoint, slack, teams or commerce_events). Their
prompts, tools and MCP servers are only in `-o json`.

Its reads require the ai.agent_read scope.
Its writes require different OAuth scopes; the help of each names them.

**Command:** `epx ai agents [flags]`, `epx ai agents [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai agents attach`](#epx-ai-agents-attach)\
  Attach a file to a chat with an agent.
- [`epx ai agents create`](#epx-ai-agents-create)\
  Create an agent from a template.
- [`epx ai agents delete`](#epx-ai-agents-delete)\
  Delete an agent.
- [`epx ai agents export`](#epx-ai-agents-export)\
  Export agents, which creates an export job.
- [`epx ai agents get`](#epx-ai-agents-get)\
  Show one agent.
- [`epx ai agents import`](#epx-ai-agents-import)\
  Import agents.
- [`epx ai agents list`](#epx-ai-agents-list)\
  List AI agents.
- [`epx ai agents patch`](#epx-ai-agents-patch)\
  Change members of an agent.
- [`epx ai agents search`](#epx-ai-agents-search)\
  Search AI agents.
- [`epx ai agents update`](#epx-ai-agents-update)\
  Upsert an agent.

## epx ai agents attach

Attach a file to a chat with an agent.

Uploads the file that `--upload` names (jpg, jpeg, png, pdf, csv, text or txt), or assigns the
media asset that `--media` names ([`epx media list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-list)), to the agent AGENT_ID, the ID that
[`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows, in the session that `--session` names, or a new one. Emporix
answers with the attachment's id and the sessionId, which [`epx ai chat`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-chat) with that agent
takes as `--attachment` and `--session`.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the ai.agentexecution_manage scope.

**Command:** `epx ai agents attach AGENT_ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--media <ID>`\
  The ID of a media asset to attach.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--session <ID>`\
  The ID of the session to go on with, the sessionId of an earlier chat or attachment; without it, a new one.
- `--upload <PATH>`\
  The file in PATH to upload (up to 30 MB including the form), sent with its name.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai agents attach complaint-agent --upload order.pdf --yes
```

## epx ai agents create

Create an agent from a template.

Creates an agent from the agent template TEMPLATE_ID, the ID that [`epx ai templates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-list)
shows, which must be enabled, with the JSON in `--file`: its name and description, its
userPrompt and, optionally, its id. Emporix answers with the id of the agent.

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "gram", "de": "Gramm"}`, as `-o json` of a read prints it, or with `--content-language`
LANGUAGE a string in that language, e.g. `"gram"` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai agents create TEMPLATE_ID [flags]`

**Flags**

- `--content-language <LANGUAGE>`\
  The LANGUAGE of localized text in the body, or \* for maps of translations. Default: `*`.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai agents create complaint-template --file agent.json --dry-run
```

## epx ai agents delete

Delete an agent.

Deletes the agent ID. Emporix refuses to delete one that others use, unless
`--force` unassigns it from them.
ID is the ID that [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai agents delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete it even where other agents collaborate with it.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai agents export

Export agents, which creates an export job.

Exports the agents that `--agent` names (up to 100) with the tools and MCP servers they use,
and prints the export: data, the agents base64-encoded, and its checksum, which
[`epx ai agents import`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-import) takes on another tenant. Emporix keeps each export as a job of type
EXPORT, the jobId of the response, which [`epx ai jobs get JOB_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-get) reads.
It changes data although Emporix needs no write-capable scope for it.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_read scope.

**Command:** `epx ai agents export [flags]`

**Flags**

- `--agent <ID>`\
  The ID of an agent to export (repeatable; required without `--file`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai agents export --agent complaint-agent --yes > agents.json
```

## epx ai agents get

Show one agent.

Reads AI agents: TYPE is generic, complaint or anti_fraud, PROVIDER and MODEL the LLM an
agent uses and TRIGGERS what starts it (endpoint, slack, teams or commerce_events). Their
prompts, tools and MCP servers are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai agents get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, PROVIDER, MODEL, TRIGGERS, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai agents import

Import agents.

Imports the agents of an export, with the tools and MCP servers they use, from the JSON in
`--file`: the data and checksum that [`epx ai agents export`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-export) prints, which Emporix checks.
Imported agents start disabled; Emporix answers with what it imported and the jobId of the
IMPORT job that it keeps ([`epx ai jobs get JOB_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-get)).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai agents import [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai agents import --file agents.json --dry-run
```

## epx ai agents list

List AI agents.

Reads AI agents: TYPE is generic, complaint or anti_fraud, PROVIDER and MODEL the LLM an
agent uses and TRIGGERS what starts it (endpoint, slack, teams or commerce_events). Their
prompts, tools and MCP servers are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai agents list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, PROVIDER, MODEL, TRIGGERS, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai agents patch

Change members of an agent.

Changes the agent ID with the JSON array of operations in `--file`, each with op
(ADD, REMOVE or REPLACE), path and value, e.g.
`[{"op": "REPLACE", "path": "/enabled", "value": false}]`; its other members keep their
values.
ID is the ID that [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai agents patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai agents search

Search AI agents.

Reads the AI agents that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai agents search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, PROVIDER, MODEL, TRIGGERS, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai agents update

Upsert an agent.

Replaces the agent ID with the JSON in `--file`, or creates it with that ID where
none exists (an upsert): Emporix answers with the id of one it created and nothing for one it
replaced. With metadata.version, as [`epx ai agents get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-get) prints it,
Emporix replaces only that version: another is a conflict (exit 5).
The JSON holds its name and description, userPrompt, llmConfig, triggers and mcpServers,
and its nativeTools, agentCollaborations, enabledMemory and maxRecursionLimit.
ID is the ID that [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows.

Localized text in the body, such as a name, is a map of translations, e.g.
`{"en": "gram", "de": "Gramm"}`, as `-o json` of a read prints it, or with `--content-language`
LANGUAGE a string in that language, e.g. `"gram"` with `--content-language` en.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai agents update ID [flags]`

**Flags**

- `--content-language <LANGUAGE>`\
  The LANGUAGE of localized text in the body, or \* for maps of translations. Default: `*`.
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai analytics

Read the analytics of agents.

Reads what Emporix counts of the requests and sessions of agents, as its dashboards show them.

Requires the ai.agent_read scope.

**Command:** `epx ai analytics [flags]`, `epx ai analytics [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai analytics agent`](#epx-ai-analytics-agent)\
  Show the request and session totals of one agent.
- [`epx ai analytics executions`](#epx-ai-analytics-executions)\
  List how often agents ran per calendar quarter, month or week.
- [`epx ai analytics tenant`](#epx-ai-analytics-tenant)\
  Show the request and session totals of every agent.

## epx ai analytics agent

Show the request and session totals of one agent.

Reads the totals of the requests and sessions of one agent: AGENT_ID is the ID that
[`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows; an unknown one counts nothing.
LOGS is requests or sessions, TOTAL how many there were in the last four weeks, WARNING and
ERROR how many of them had that severity and ERROR RATE their share with errors, as Emporix
computes it. The requests per session and the weekly trends are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai analytics agent AGENT_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among LOGS, TOTAL, WARNING, ERROR, ERROR RATE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai analytics executions

List how often agents ran per calendar quarter, month or week.

Reads how often agents ran (their sessions) per calendar quarter, Emporix's default
granularity, or per month or week (`--granularity`): each PERIOD that Emporix reports runs
from FROM (inclusive) to TO (exclusive), in UTC. AGENT_IDS are the IDs that
[`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows, separated by commas without spaces (at most 100); an unknown
one ran 0 times.

Requires the ai.agent_read scope.

**Command:** `epx ai analytics executions AGENT_IDS [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among AGENT, PERIOD, FROM, TO, EXECUTIONS.
- `--granularity <period>`\
  The calendar period of each count: QUARTER, MONTH or WEEK.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai analytics tenant

Show the request and session totals of every agent.

Reads the totals of the requests and sessions of every agent of the tenant.
LOGS is requests or sessions, TOTAL how many there were in the last four weeks, WARNING and
ERROR how many of them had that severity and ERROR RATE their share with errors, as Emporix
computes it. The requests per session and the weekly trends are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai analytics tenant [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among LOGS, TOTAL, WARNING, ERROR, ERROR RATE.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai chat

Send a message to an agent and print its reply (consumes AI credits).

Sends the message that `--message` gives to the agent `--agent`, and prints its reply: the messages
of the agent and of those it collaborates with, each with the sessionId that `--session` takes to
go on.
A reply may take longer than the `--timeout` of a request (30s by default): then give a longer
one, e.g. `--timeout` 5m, or use chat-async.
It consumes AI credits: the LLM tokens of the agent's model, and of the agents it
collaborates with.
With `--session`, the chat goes on in that session, which the agent remembers where its
enabledMemory is true; an attachment must be one of the session, attached to the same agent
([`epx ai agents attach`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-attach)). `--file` gives the body instead, with the caption and purpose of
each attachment.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agentexecution_manage scope.

**Command:** `epx ai chat [flags]`

**Flags**

- `--agent <ID>`\
  The ID of the agent, as [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows it (required without `--file`).
- `--attachment <ID>`\
  The ID of an attachment of the session to send with the message, as [`epx ai agents attach`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-attach) prints it (repeatable) (only with `--session`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--message <TEXT>`\
  The TEXT to send to the agent (required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--session <ID>`\
  The ID of the session to go on with, the sessionId of an earlier chat or attachment; without it, a new one.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai chat --agent complaint-agent --message "Find order EON1243" --yes
```

## epx ai chat-async

Send a message to an agent, which replies in a job (consumes AI credits).

Sends the message that `--message` gives to the agent `--agent`, which replies in a job of type
AGENT_CHAT: Emporix answers with its jobId, which [`epx ai jobs get JOB_ID`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-get) reads.
It consumes AI credits: the LLM tokens of the agent's model, and of the agents it
collaborates with.
With `--session`, the chat goes on in that session, which the agent remembers where its
enabledMemory is true; an attachment must be one of the session, attached to the same agent
([`epx ai agents attach`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-attach)). `--file` gives the body instead, with the caption and purpose of
each attachment.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agentexecution_manage scope.

**Command:** `epx ai chat-async [flags]`

**Flags**

- `--agent <ID>`\
  The ID of the agent, as [`epx ai agents list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-list) shows it (required without `--file`).
- `--attachment <ID>`\
  The ID of an attachment of the session to send with the message, as [`epx ai agents attach`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-agents-attach) prints it (repeatable) (only with `--session`).
- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--message <TEXT>`\
  The TEXT to send to the agent (required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--session <ID>`\
  The ID of the session to go on with, the sessionId of an earlier chat or attachment; without it, a new one.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai chat-async --agent complaint-agent --message "Find order EON1243" --yes
```

## epx ai commerce-events

List the commerce events that can trigger an agent.

Reads the commerce events that can start an agent with a commerce_events trigger, sorted.

Requires the ai.agent_read scope.

**Command:** `epx ai commerce-events [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among EVENT.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai conversations

Read agent conversations.

Reads the conversations of agents in Microsoft Teams channels and group chats: ID is the
conversation in Teams, NAME its channel or chat, SESSION the agent's session ([`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get)) and LAST MESSAGE when it was last active. The channel and the context
reference are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai conversations [flags]`, `epx ai conversations [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai conversations list`](#epx-ai-conversations-list)\
  List agent conversations.
- [`epx ai conversations search`](#epx-ai-conversations-search)\
  Search agent conversations.

## epx ai conversations list

List agent conversations.

Reads the conversations of agents in Microsoft Teams channels and group chats: ID is the
conversation in Teams, NAME its channel or chat, SESSION the agent's session ([`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get)) and LAST MESSAGE when it was last active. The channel and the context
reference are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai conversations list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, AGENT, SESSION, LAST MESSAGE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai conversations search

Search agent conversations.

Reads the agent conversations that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai conversations list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-conversations-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai conversations search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, AGENT, SESSION, LAST MESSAGE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai generate-completion

Generate the next message of a chat history (consumes AI credits).

Generates the next message of the chat history that `--message` gives, one message per
`--message` in their order, and prints it as result.
It consumes AI credits: the LLM tokens of the tenant's own OpenAI account where the tenant
configuration openAiApiToken holds its token, or else Emporix's, of which a tenant may use
3 million input and 1 million output tokens.
A reply may take longer than the `--timeout` of a request (30s by default): then give a longer
one, e.g. `--timeout` 5m.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.completion_manage scope.

**Command:** `epx ai generate-completion [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--message <ROLE:TEXT>`\
  A message of the chat history as ROLE:TEXT, ROLE SYSTEM, USER or ASSISTANT (repeatable; required without `--file`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai generate-completion --message "SYSTEM:You write product texts" \
  --message "USER:Describe product 100200"
```

## epx ai generate-text

Generate a text from a prompt (consumes AI credits).

Generates a text from the prompt that `--prompt` gives, such as a product description, and
prints it as result.
It consumes AI credits: the LLM tokens of the tenant's own OpenAI account where the tenant
configuration openAiApiToken holds its token, or else Emporix's, of which a tenant may use
3 million input and 1 million output tokens.
A reply may take longer than the `--timeout` of a request (30s by default): then give a longer
one, e.g. `--timeout` 5m.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.text_manage scope.

**Command:** `epx ai generate-text [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte instead of the one that epx makes (- reads standard input; sending it then needs `--yes`).
- `--max-tokens <TOKENS>`\
  The most LLM TOKENS to use (Emporix: 4097 without it).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--prompt <TEXT>`\
  The TEXT that asks for the text (required without `--file`).
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai generate-text --prompt "Describe product 100200 in German" --max-tokens 500
```

## epx ai jobs

Read and delete agent jobs.

Reads the jobs of the AI Service: TYPE is AGENT_CHAT (an agent's run), IMPORT or EXPORT (of
agents), STATUS in_progress, success, failure or skipped, AGENT and SESSION the agent that
ran and its session ([`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get)) and EVENT the commerce event that started
it. The agent's message and an import's or export's result are only in `-o json`.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai jobs [flags]`, `epx ai jobs [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai jobs delete`](#epx-ai-jobs-delete)\
  Delete an agent job.
- [`epx ai jobs get`](#epx-ai-jobs-get)\
  Show one agent job.
- [`epx ai jobs list`](#epx-ai-jobs-list)\
  List agent jobs.
- [`epx ai jobs search`](#epx-ai-jobs-search)\
  Search agent jobs.

## epx ai jobs delete

Delete an agent job.

Deletes the agent job ID.
ID is the ID that [`epx ai jobs list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-list) shows; an unknown one is not found.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai jobs delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai jobs get

Show one agent job.

Reads the jobs of the AI Service: TYPE is AGENT_CHAT (an agent's run), IMPORT or EXPORT (of
agents), STATUS in_progress, success, failure or skipped, AGENT and SESSION the agent that
ran and its session ([`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get)) and EVENT the commerce event that started
it. The agent's message and an import's or export's result are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai jobs get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, AGENT, SESSION, EVENT, CREATED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai jobs list

List agent jobs.

Reads the jobs of the AI Service: TYPE is AGENT_CHAT (an agent's run), IMPORT or EXPORT (of
agents), STATUS in_progress, success, failure or skipped, AGENT and SESSION the agent that
ran and its session ([`epx ai logs sessions get`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-get)) and EVENT the commerce event that started
it. The agent's message and an import's or export's result are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai jobs list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, AGENT, SESSION, EVENT, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai jobs search

Search agent jobs.

Reads the agent jobs that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai jobs list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-jobs-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai jobs search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, STATUS, AGENT, SESSION, EVENT, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs

Read the logs of agent requests and sessions.

Reads the logs that agents write: one per request, and one per session, which holds the
messages of its requests.

Requires the ai.agent_read scope.

**Command:** `epx ai logs [flags]`, `epx ai logs [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai logs requests`](#epx-ai-logs-requests)\
  Read agent request logs.
- [`epx ai logs sessions`](#epx-ai-logs-sessions)\
  Read agent session logs.

## epx ai logs requests

Read agent request logs.

Reads the logs of the requests to agents, one per request: ID is the log that get takes,
REQUEST the request it logs and SESSION the session it belongs to.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs requests [flags]`, `epx ai logs requests [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai logs requests get`](#epx-ai-logs-requests-get)\
  Show one agent request log.
- [`epx ai logs requests list`](#epx-ai-logs-requests-list)\
  List agent request logs.
- [`epx ai logs requests search`](#epx-ai-logs-requests-search)\
  Search agent request logs.

## epx ai logs requests get

Show one agent request log.

Reads the logs of the requests to agents, one per request: ID is the log that get takes,
REQUEST the request it logs and SESSION the session it belongs to.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs requests get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, REQUEST, SESSION, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs requests list

List agent request logs.

Reads the logs of the requests to agents, one per request: ID is the log that get takes,
REQUEST the request it logs and SESSION the session it belongs to.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs requests list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, REQUEST, SESSION, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs requests search

Search agent request logs.

Reads the agent request logs that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai logs requests list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-requests-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai logs requests search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, REQUEST, SESSION, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs sessions

Read agent session logs.

Reads the logs of the sessions of agents, one per session, with the messages of each of its
requests.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs sessions [flags]`, `epx ai logs sessions [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai logs sessions get`](#epx-ai-logs-sessions-get)\
  Show one agent session log.
- [`epx ai logs sessions list`](#epx-ai-logs-sessions-list)\
  List agent session logs.
- [`epx ai logs sessions search`](#epx-ai-logs-sessions-search)\
  Search agent session logs.

## epx ai logs sessions get

Show one agent session log.

Reads the logs of the sessions of agents, one per session, with the messages of each of its
requests.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs sessions get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs sessions list

List agent session logs.

Reads the logs of the sessions of agents, one per session, with the messages of each of its
requests.
AGENT is the agent that was called, SEVERITY INFO, WARNING or ERROR, MESSAGES how many
messages were logged (empty when a log comes without them) and SECONDS how long it took; the
messages, the other agents that took part and the LLM tokens used are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai logs sessions list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai logs sessions search

Search agent session logs.

Reads the agent session logs that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai logs sessions list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-logs-sessions-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai logs sessions search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, AGENT, SEVERITY, MESSAGES, SECONDS, CREATED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers

Read, upsert, change and delete MCP servers.

Reads the tenant's MCP servers, whose tools agents can use: TYPE is custom (a server at URL)
or dynamic (a preview at Emporix: tools that the server's configuration defines, only in
`-o json`), TRANSPORT streamable_http or sse.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai mcp-servers [flags]`, `epx ai mcp-servers [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai mcp-servers delete`](#epx-ai-mcp-servers-delete)\
  Delete an MCP server.
- [`epx ai mcp-servers get`](#epx-ai-mcp-servers-get)\
  Show one MCP server.
- [`epx ai mcp-servers list`](#epx-ai-mcp-servers-list)\
  List MCP servers.
- [`epx ai mcp-servers patch`](#epx-ai-mcp-servers-patch)\
  Change members of an MCP server.
- [`epx ai mcp-servers search`](#epx-ai-mcp-servers-search)\
  Search MCP servers.
- [`epx ai mcp-servers update`](#epx-ai-mcp-servers-update)\
  Upsert an MCP server.

## epx ai mcp-servers delete

Delete an MCP server.

Deletes the MCP server ID. Emporix refuses to delete one that others use, unless
`--force` unassigns it from them.
ID is the ID that [`epx ai mcp-servers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai mcp-servers delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete it even where agents use it.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers get

Show one MCP server.

Reads the tenant's MCP servers, whose tools agents can use: TYPE is custom (a server at URL)
or dynamic (a preview at Emporix: tools that the server's configuration defines, only in
`-o json`), TRANSPORT streamable_http or sse.

Requires the ai.agent_read scope.

**Command:** `epx ai mcp-servers get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, TRANSPORT, NAME, URL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers list

List MCP servers.

Reads the tenant's MCP servers, whose tools agents can use: TYPE is custom (a server at URL)
or dynamic (a preview at Emporix: tools that the server's configuration defines, only in
`-o json`), TRANSPORT streamable_http or sse.

Requires the ai.agent_read scope.

**Command:** `epx ai mcp-servers list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, TRANSPORT, NAME, URL.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers patch

Change members of an MCP server.

Changes the MCP server ID with the JSON array of operations in `--file`, each with op
(ADD, REMOVE or REPLACE), path and value, e.g.
`[{"op": "REPLACE", "path": "/enabled", "value": false}]`; its other members keep their
values.
ID is the ID that [`epx ai mcp-servers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai mcp-servers patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers search

Search MCP servers.

Reads the MCP servers that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai mcp-servers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai mcp-servers search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, TRANSPORT, NAME, URL.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai mcp-servers update

Upsert an MCP server.

Replaces the MCP server ID with the JSON in `--file`, or creates it with that ID where
none exists (an upsert): Emporix answers with the id of one it created and nothing for one it
replaced. With metadata.version, as [`epx ai mcp-servers get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-get) prints it,
Emporix replaces only that version: another is a conflict (exit 5).
The JSON holds its type (custom: config with the server's url; dynamic: its tools), name
and transport, and whether it is enabled.
ID is the ID that [`epx ai mcp-servers list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-mcp-servers-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai mcp-servers update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai models

List the LLM models that agents can use.

Reads the LLM models of every provider that agents can use (openai, anthropic and google):
ID is what an agent's LLM configuration names and THINKING whether the model supports
thinking; descriptions are only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai models [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, PROVIDER, THINKING, NAME.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths

Read, upsert, change and delete OAuth configurations.

Reads OAuth configurations (a preview at Emporix): the OAuth 2.0 client credentials with which
agents get a token for a self-hosted LLM from TOKEN URL. SECRET TOKEN is the ID of the token
([`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list)) that holds the client's secret; the scope is only in `-o json`, which
prints the response as Emporix sends it.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai oauths [flags]`, `epx ai oauths [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai oauths delete`](#epx-ai-oauths-delete)\
  Delete an OAuth configuration.
- [`epx ai oauths get`](#epx-ai-oauths-get)\
  Show one OAuth configuration.
- [`epx ai oauths list`](#epx-ai-oauths-list)\
  List OAuth configurations.
- [`epx ai oauths patch`](#epx-ai-oauths-patch)\
  Change members of an OAuth configuration.
- [`epx ai oauths search`](#epx-ai-oauths-search)\
  Search OAuth configurations.
- [`epx ai oauths update`](#epx-ai-oauths-update)\
  Upsert an OAuth configuration.

## epx ai oauths delete

Delete an OAuth configuration.

Deletes the OAuth configuration ID. Emporix refuses to delete one that others use, unless
`--force` unassigns it from them.
ID is the ID that [`epx ai oauths list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai oauths delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete it even where agents use it.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths get

Show one OAuth configuration.

Reads OAuth configurations (a preview at Emporix): the OAuth 2.0 client credentials with which
agents get a token for a self-hosted LLM from TOKEN URL. SECRET TOKEN is the ID of the token
([`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list)) that holds the client's secret; the scope is only in `-o json`, which
prints the response as Emporix sends it.

Requires the ai.agent_read scope.

**Command:** `epx ai oauths get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ENABLED, CLIENT ID, SECRET TOKEN, TOKEN URL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths list

List OAuth configurations.

Reads OAuth configurations (a preview at Emporix): the OAuth 2.0 client credentials with which
agents get a token for a self-hosted LLM from TOKEN URL. SECRET TOKEN is the ID of the token
([`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list)) that holds the client's secret; the scope is only in `-o json`, which
prints the response as Emporix sends it.

Requires the ai.agent_read scope.

**Command:** `epx ai oauths list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ENABLED, CLIENT ID, SECRET TOKEN, TOKEN URL.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths patch

Change members of an OAuth configuration.

Changes the OAuth configuration ID with the JSON array of operations in `--file`, each with op
(ADD, REMOVE or REPLACE), path and value, e.g.
`[{"op": "REPLACE", "path": "/enabled", "value": false}]`; its other members keep their
values.
ID is the ID that [`epx ai oauths list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai oauths patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths search

Search OAuth configurations.

Reads the OAuth configurations that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai oauths list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai oauths search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, ENABLED, CLIENT ID, SECRET TOKEN, TOKEN URL.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai oauths update

Upsert an OAuth configuration.

Replaces the OAuth configuration ID with the JSON in `--file`, or creates it with that ID where
none exists (an upsert): Emporix answers with the id of one it created and nothing for one it
replaced. With metadata.version, as [`epx ai oauths get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-get) prints it,
Emporix replaces only that version: another is a conflict (exit 5).
The JSON holds its url (the token URL), clientId and grantType (client_credentials, the
only one), and its scope, clientSecretToken (the token that holds the secret) and whether it
is enabled.
ID is the ID that [`epx ai oauths list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-oauths-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai oauths update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai rag

Read which fields the AI RAG Indexer embeds and filters by; reindex.

Reads the AI RAG Indexer, which embeds the entities of a type in a vector database for the
retrieval-augmented generation (RAG) tools of agents: the fields it can embed and those a
vector search can filter by; and embeds the entities of a type again.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai rag [flags]`, `epx ai rag [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai rag fields`](#epx-ai-rag-fields)\
  List the fields of an entity type that can be embedded for RAG search.
- [`epx ai rag filter-fields`](#epx-ai-rag-filter-fields)\
  List the fields of an entity type that a vector search can filter by.
- [`epx ai rag reindex`](#epx-ai-rag-reindex)\
  Embed every entity of a type again for RAG search (deprecated by Emporix).

## epx ai rag fields

List the fields of an entity type that can be embedded for RAG search.

Reads the fields of an entity type, by path, that the AI RAG Indexer can embed.
TYPE is PRODUCT, ORDER or a custom entity type ([`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list)). The fields come whole, so they are read with one
request.

Requires the ai.agent_read scope.

**Command:** `epx ai rag fields TYPE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among FIELD.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai rag filter-fields

List the fields of an entity type that a vector search can filter by.

Reads the fields of an entity type, by path, that a vector search can filter by before RAG
retrieval, with their TYPE (string, integer, float, boolean, datetime, date, time,
dictionary, list or object).
TYPE is PRODUCT, ORDER or a custom entity type ([`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list)). The fields come whole, so they are read with one
request.

Requires the ai.agent_read scope.

**Command:** `epx ai rag filter-fields TYPE [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among FIELD, TYPE.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai rag reindex

Embed every entity of a type again for RAG search (deprecated by Emporix).

Embeds every entity of a type again in the vector database, replacing its records there;
it sends no body. Emporix has deprecated this endpoint: [`epx indexing reindex-jobs create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/indexing#epx-indexing-reindex-jobs-create)
with `--rag` starts a reindex job that embeds products too, which its `--wait` follows.
TYPE is PRODUCT, ORDER or a custom entity type ([`epx custom-entities list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/custom-entities#epx-custom-entities-list)). Emporix answers an unknown TYPE with 204 No Content (exit 0),
so check the TYPE first.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai rag reindex TYPE [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx ai rag reindex PRODUCT --dry-run
```

## epx ai templates

Read agent templates.

Reads agent templates, from which agents are made: TYPE is generic, complaint or anti_fraud,
PROVIDER and MODEL the LLM a template uses. Emporix reads no single template:
[`epx ai templates list --query id:ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-list) shows one.

Requires the ai.agent_read scope.

**Command:** `epx ai templates [flags]`, `epx ai templates [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai templates list`](#epx-ai-templates-list)\
  List agent templates.
- [`epx ai templates search`](#epx-ai-templates-search)\
  Search agent templates.

## epx ai templates list

List agent templates.

Reads agent templates, from which agents are made: TYPE is generic, complaint or anti_fraud,
PROVIDER and MODEL the LLM a template uses. Emporix reads no single template:
[`epx ai templates list --query id:ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-list) shows one.

Requires the ai.agent_read scope.

**Command:** `epx ai templates list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, PROVIDER, MODEL, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai templates search

Search agent templates.

Reads the agent templates that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai templates list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-templates-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai templates search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, PROVIDER, MODEL, NAME.
- `--language <language>`\
  Language of localized table text; a missing translation shows another, labeled (`-o json` prints all). Default: `en`.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens

Read, upsert, change and delete AI tokens.

Reads the tokens that hold the API keys and secrets of LLMs, tools, OAuth configurations and
MCP servers, by name: the table shows no value, and the API reference documents none in a
response; `-o json` prints the response as Emporix sends it.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai tokens [flags]`, `epx ai tokens [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai tokens delete`](#epx-ai-tokens-delete)\
  Delete an AI token.
- [`epx ai tokens get`](#epx-ai-tokens-get)\
  Show one token.
- [`epx ai tokens list`](#epx-ai-tokens-list)\
  List AI tokens.
- [`epx ai tokens patch`](#epx-ai-tokens-patch)\
  Change members of an AI token.
- [`epx ai tokens search`](#epx-ai-tokens-search)\
  Search AI tokens.
- [`epx ai tokens update`](#epx-ai-tokens-update)\
  Upsert an AI token.

## epx ai tokens delete

Delete an AI token.

Deletes the token ID. Emporix refuses to delete one that others use, unless
`--force` unassigns it from them.
ID is the ID that [`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tokens delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete it even where agents or MCP servers use it.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens get

Show one token.

Reads the tokens that hold the API keys and secrets of LLMs, tools, OAuth configurations and
MCP servers, by name: the table shows no value, and the API reference documents none in a
response; `-o json` prints the response as Emporix sends it.

Requires the ai.agent_read scope.

**Command:** `epx ai tokens get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CREATED, MODIFIED.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens list

List AI tokens.

Reads the tokens that hold the API keys and secrets of LLMs, tools, OAuth configurations and
MCP servers, by name: the table shows no value, and the API reference documents none in a
response; `-o json` prints the response as Emporix sends it.

Requires the ai.agent_read scope.

**Command:** `epx ai tokens list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CREATED, MODIFIED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens patch

Change members of an AI token.

Changes the token ID with the JSON array of operations in `--file`, each with op
(ADD, REMOVE or REPLACE), path and value, e.g.
`[{"op": "REPLACE", "path": "/enabled", "value": false}]`; its other members keep their
values.
ID is the ID that [`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tokens patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens search

Search AI tokens.

Reads the AI tokens that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai tokens search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, NAME, CREATED, MODIFIED.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tokens update

Upsert an AI token.

Replaces the token ID with the JSON in `--file`, or creates it with that ID where
none exists (an upsert): Emporix answers with the id of one it created and nothing for one it
replaced. With metadata.version, as [`epx ai tokens get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-get) prints it,
Emporix replaces only that version: another is a conflict (exit 5).
The JSON holds its name and value, the API key or secret, which cannot change once the
token is created. A dry run prints the body, value included.
ID is the ID that [`epx ai tokens list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tokens-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tokens update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools

Read, upsert, change and delete agent tools.

Reads the native tools that agents can use: TYPE is slack, teams, rag_custom (retrieval-
augmented generation over the tenant's own vector database) or rag_emporix (over Emporix
data); their configuration is only in `-o json`.

Its reads require the ai.agent_read scope.
Its writes require the ai.agent_manage scope.

**Command:** `epx ai tools [flags]`, `epx ai tools [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx ai tools delete`](#epx-ai-tools-delete)\
  Delete an agent tool.
- [`epx ai tools get`](#epx-ai-tools-get)\
  Show one tool.
- [`epx ai tools list`](#epx-ai-tools-list)\
  List agent tools.
- [`epx ai tools patch`](#epx-ai-tools-patch)\
  Change members of an agent tool.
- [`epx ai tools search`](#epx-ai-tools-search)\
  Search agent tools.
- [`epx ai tools update`](#epx-ai-tools-update)\
  Upsert an agent tool.

## epx ai tools delete

Delete an agent tool.

Deletes the tool ID. Emporix refuses to delete one that others use, unless
`--force` unassigns it from them.
ID is the ID that [`epx ai tools list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-list) shows.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tools delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--force`\
  Delete it even where agents use it.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools get

Show one tool.

Reads the native tools that agents can use: TYPE is slack, teams, rag_custom (retrieval-
augmented generation over the tenant's own vector database) or rag_emporix (over Emporix
data); their configuration is only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai tools get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, NAME.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools list

List agent tools.

Reads the native tools that agents can use: TYPE is slack, teams, rag_custom (retrieval-
augmented generation over the tenant's own vector database) or rag_emporix (over Emporix
data); their configuration is only in `-o json`.

Requires the ai.agent_read scope.

**Command:** `epx ai tools list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, NAME.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, passed as given, e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools patch

Change members of an agent tool.

Changes the tool ID with the JSON array of operations in `--file`, each with op
(ADD, REMOVE or REPLACE), path and value, e.g.
`[{"op": "REPLACE", "path": "/enabled", "value": false}]`; its other members keep their
values.
ID is the ID that [`epx ai tools list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-list) shows; an unknown one is not found.

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tools patch ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools search

Search agent tools.

Reads the agent tools that an Emporix q filter finds (`--query`; without it,
every one).
It reads and shows them as [`epx ai tools list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-list) does.

Changes nothing: Emporix serves this read as POST with a JSON body, which the
flags make or `--body` gives whole.

Requires the ai.agent_read scope.

**Command:** `epx ai tools search [flags]`

**Flags**

- `--body <FILE>`\
  Send the JSON request body in FILE as given (- reads standard input) instead of the one the other flags make.
- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ENABLED, NAME.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--query <filter>`\
  Emporix q filter, sent as given in the request body, not in the URL (empty: no filter), e.g. code:(A1,A2); space-separated conditions must all match (fields mostly as in `-o json`).
- `--sort <order>`\
  Emporix sort order, passed as given, e.g. code or name.en:desc,code; fields mostly as in `-o json` (Emporix may ignore an unknown one).
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx ai tools update

Upsert an agent tool.

Replaces the tool ID with the JSON in `--file`, or creates it with that ID where
none exists (an upsert): Emporix answers with the id of one it created and nothing for one it
replaced. With metadata.version, as [`epx ai tools get ID -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-get) prints it,
Emporix replaces only that version: another is a conflict (exit 5).
The JSON holds its type (slack, teams, rag_custom or rag_emporix), name and config, and
whether it is enabled.
ID is the ID that [`epx ai tools list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/ai#epx-ai-tools-list) shows.

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the ai.agent_manage scope.

**Command:** `epx ai tools update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON request body in PATH, sent byte for byte (- reads standard input; sending it then needs `--yes`; required).
- `--force`\
  Disable it even where enabled agents use it, which disables them too.
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

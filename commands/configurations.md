# epx configurations

> Read and change tenant configurations.

## epx configurations

Reads tenant configurations (settings such as project_lang).
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).
A setting that was never set is not found (exit 4): Emporix then applies its default.
[`epx configurations global`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-global), `clients`, `for-client` and `client-setting` read global
settings and those of clients. [`epx configurations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-create), `update` and `delete`
change settings, and `client-create`, `client-update` and `client-delete` those
of a client.

Its reads require different OAuth scopes; the help of each names them.
Its writes require the configuration.configuration_manage scope.

**Command:** `epx configurations [flags]`, `epx configurations [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx configurations client-create`](#epx-configurations-client-create)\
  Create configurations of a client.
- [`epx configurations client-delete`](#epx-configurations-client-delete)\
  Delete a configuration of a client.
- [`epx configurations client-setting`](#epx-configurations-client-setting)\
  Show one configuration of a client.
- [`epx configurations client-update`](#epx-configurations-client-update)\
  Replace a configuration of a client.
- [`epx configurations clients`](#epx-configurations-clients)\
  List the clients that have configurations.
- [`epx configurations create`](#epx-configurations-create)\
  Create tenant configurations.
- [`epx configurations delete`](#epx-configurations-delete)\
  Delete a tenant configuration.
- [`epx configurations for-client`](#epx-configurations-for-client)\
  List the configurations of a client.
- [`epx configurations get`](#epx-configurations-get)\
  Show one configuration.
- [`epx configurations global`](#epx-configurations-global)\
  List global configurations.
- [`epx configurations list`](#epx-configurations-list)\
  List configurations.
- [`epx configurations update`](#epx-configurations-update)\
  Replace a tenant configuration.

## epx configurations client-create

Create configurations of a client.

Creates configurations of the client CLIENT from the JSON array of settings that `--file` gives,
each with its key and value (and optionally secured, description, readOnly, restricted and
schemaUrl), and prints the settings created. Emporix answers 409 Conflict (exit 5) when one of
the keys exists already (the other keys of the request may be created then): change it with [`epx configurations client-update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-update).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations client-create CLIENT [flags]`

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

## epx configurations client-delete

Delete a configuration of a client.

Deletes the configuration KEY of the client CLIENT, as
[`epx configurations client-setting CLIENT KEY`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-setting) shows it; Emporix then applies its
default, if it has one. A key that is not set is not found (exit 4).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type KEY at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations client-delete CLIENT KEY [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations client-setting

Show one configuration of a client.

Reads one setting of a client.
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).
CLIENT is a name that [`epx configurations clients`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-clients) shows and KEY a key that
[`epx configurations for-client CLIENT`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-for-client) shows; an unknown one is not found (exit 4).

Requires the configuration.configuration_view scope.

**Command:** `epx configurations client-setting CLIENT KEY [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KEY, VALUE, VERSION, SECURED, READ ONLY.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations client-update

Replace a configuration of a client.

Replaces the configuration KEY of the client CLIENT with the setting that `--file` gives, a JSON
object with the key KEY (Emporix refuses another) and its value, and prints the setting. Give
every member to keep, such as secured, description and version, as
[`epx configurations client-setting CLIENT KEY -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-setting) shows them. Emporix refuses a key that was never set (404 Not Found,
exit 4): create it with [`epx configurations client-create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-client-create).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations client-update CLIENT KEY [flags]`

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

## epx configurations clients

List the clients that have configurations.

Reads the clients (such as a storefront or an extension) that keep their own settings.

Requires the configuration.configuration_view scope.

**Command:** `epx configurations clients [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among CLIENT.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations create

Create tenant configurations.

Creates tenant configurations from the JSON array of settings that `--file` gives,
each with its key and value (and optionally secured, description, readOnly, restricted and
schemaUrl), and prints the settings created. Emporix answers 409 Conflict (exit 5) when one of
the keys exists already (the other keys of the request may be created then): change it with [`epx configurations update`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-update).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations create [flags]`

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

## epx configurations delete

Delete a tenant configuration.

Deletes the tenant configuration KEY, as
[`epx configurations get KEY`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-get) shows it; Emporix then applies its
default, if it has one. A key that is not set is not found (exit 4).

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type KEY at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations delete KEY [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations for-client

List the configurations of a client.

Reads the settings of a client.
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).
CLIENT is a name that [`epx configurations clients`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-clients) shows; an unknown one lists nothing.

Requires the configuration.configuration_view scope.

**Command:** `epx configurations for-client CLIENT [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KEY, VALUE, VERSION, SECURED, READ ONLY.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations get

Show one configuration.

Reads tenant configurations (settings such as project_lang).
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).
A setting that was never set is not found (exit 4): Emporix then applies its default.
[`epx configurations global`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-global), `clients`, `for-client` and `client-setting` read global
settings and those of clients. [`epx configurations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-create), `update` and `delete`
change settings, and `client-create`, `client-update` and `client-delete` those
of a client.

Requires the configuration.configuration_view scope.

**Command:** `epx configurations get KEY [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KEY, VALUE, VERSION, SECURED, READ ONLY.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations global

List global configurations.

Reads global configurations, which hold for every tenant.
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).

No OAuth scope is needed.

**Command:** `epx configurations global [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KEY, VALUE, VERSION, SECURED, READ ONLY.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations list

List configurations.

Reads tenant configurations (settings such as project_lang).
Tables show values as sent, secured ones included, and cut long ones (`-o json` prints
them whole).
A setting that was never set is not found (exit 4): Emporix then applies its default.
[`epx configurations global`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-global), `clients`, `for-client` and `client-setting` read global
settings and those of clients. [`epx configurations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-create), `update` and `delete`
change settings, and `client-create`, `client-update` and `client-delete` those
of a client.

Requires the configuration.configuration_view scope.

**Command:** `epx configurations list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among KEY, VALUE, VERSION, SECURED, READ ONLY.
- `--limit <n>`\
  Print at most n items; paged reads stop early where they safely can (0: every item).
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.
- `--stream`\
  Print items as JSON Lines as soon as they are checked, page by page on paged lists, reading up to 10000 pages where a list without it stops after 1000: a check failing on a later page exits 1 after the pages before it (excludes `-o`, `--columns` and `--no-headers`).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx configurations update

Replace a tenant configuration.

Replaces the tenant configuration KEY with the setting that `--file` gives, a JSON
object with the key KEY (Emporix refuses another) and its value, and prints the setting. Give
every member to keep, such as secured, description and version, as
[`epx configurations get KEY -o json`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-get) shows them. Emporix refuses a key that was never set (404 Not Found,
exit 4): create it with [`epx configurations create`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/configurations#epx-configurations-create).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the configuration.configuration_manage scope.

**Command:** `epx configurations update KEY [flags]`

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

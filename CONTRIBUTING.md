# Contributing to epx

Thanks for helping improve epx. Bug reports, documentation fixes and new commands are all
welcome.

- **Questions and bugs:** open an [issue](https://github.com/commerce-cloud-integrations/emporix-cli/issues)
  with the epx version (`epx version`), the command you ran and what happened. Replace
  tenant names, IDs and other data from your tenant with made-up values, and never paste a
  client secret or token.
- **Vulnerabilities:** report them privately, as [SECURITY.md](SECURITY.md) describes.
- **New commands:** open an issue first that names the Emporix API operation, so the command's
  name, flags and OAuth scopes can be agreed before you write it.

## Build and test

You need Go 1.27.3 or later.

```sh
go build -o bin/epx ./cmd/epx
go vet ./...
go test ./...
```

Tests never reach a real Emporix tenant: API calls go to local `httptest` servers, so no
credentials are needed. Format your code with `gofmt` (or `goimports`) before you commit.

### Generated files

Some tracked files are generated from the command tree, and `go test` fails while they are
out of date. When you change a command or its help, regenerate them:

```sh
go test ./internal/cli -run TestCommandReference -update
python3 -I site/sync.py
```

The first command rewrites the command reference of the documentation site
(`site/commands`); the second updates the site's navigation.
To preview the site, run `npx mint dev` in `site/`.

## How epx is organized

- `cmd/epx` wires signals, the environment and the exit code; nothing else.
- `internal/cli` holds the Cobra commands: they parse flags, validate, call the client and
  format the result. Reads of an Emporix collection are declared as a `resource`, writes as
  a `write`; follow the existing declarations rather than building commands by hand.
- `internal/emporix` owns OAuth, the HTTP transport, retries, paging and error mapping. It
  is the only package that looks at HTTP statuses.
- `internal/output` renders tables, CSV, JSON and YAML, and sanitizes API text before it
  reaches the terminal.

Prefer the standard library. A new dependency needs a reason that you state in the pull
request. Package-level mutable state and `init` functions aren't used.

## Rules for commands

- **Least privilege.** Each command requests only the OAuth scopes it needs, and its help
  ends with the sentence that names them.
- **Reads change nothing.** A search served as POST or PUT is a read too.
- **Writes are explicit.** Every write supports `--dry-run`, asks for confirmation on a
  terminal and needs `--yes` otherwise, and is never sent again once it may have reached
  Emporix. Writes are tested against fakes.
- **Identifiers stay strings**, so article numbers keep their leading zeros.
- **Neutral examples.** Tests, help and docs use made-up data: tenant `acme1`, site
  `AcmeStore_NL`, catalog `AcmeStore_catalog`, product `100200`.
- **Exit codes** come from the error classes of `internal/emporix` (see the
  [exit codes guide](https://commerce-cloud-integrations.github.io/emporix-cli/guides/exit-codes));
  don't map HTTP statuses in a command.

## Pull requests

- Keep each pull request to one change, with tests for its behavior, failure paths included.
- Run `go vet ./...` and `go test ./...` before you push, and regenerate the generated files
  when commands changed.
- Write commit messages that describe the change to epx, in the imperative mood
  ("Add `--type` to category-trees list").
- Add a line under **Unreleased** in [CHANGELOG.md](CHANGELOG.md) when users will notice the
  change. Mark changes that break scripts, such as a removed command, flag or column, under
  **Breaking changes**.

By contributing, you agree that your contribution is licensed under the
[Apache License 2.0](LICENSE).

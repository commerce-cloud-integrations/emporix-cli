# Security model

> How epx handles credentials, OAuth scopes, network requests, and the data it prints.

epx works with live commerce data, so it's designed to do only what you ask, with only the
access you grant. This page describes the guarantees it makes.

## Reads change nothing

Every read command only reads, including those that Emporix serves as POST or PUT, such
as searches, shipping quotes, and coupon validation.

## Writes are explicit

Every [write](https://commerce-cloud-integrations.github.io/emporix-cli/guides/writes) sends only what you give it, shows its exact request with
`--dry-run`, and asks before it sends. A delete asks you to type what it deletes. In
scripts, writes require `--yes`. epx never sends a write again once it may have reached
Emporix.

## Least privilege

Every command requests a token with only the OAuth scopes it needs, and its help names
them. Emporix silently drops a scope your API key lacks and then returns less data, so
epx checks every token it receives:

- When a scope is missing, the command fails with exit code 3.
- When the token belongs to another tenant, the command fails with exit code 3.
- epx never falls back to broader scopes.

Reads that Emporix allows only with a write-capable scope request it only when you pass
`--allow-write-scope`. See [Authenticate](https://commerce-cloud-integrations.github.io/emporix-cli/configure#write-capable-scopes-for-reads).

## Credentials

- epx reads its API key from environment variables or from a
  [profile](https://commerce-cloud-integrations.github.io/emporix-cli/configure#store-keys-in-profiles), never from flags.
- A profile's client secret is stored in the OS keychain, never in a file.
- epx never prints its token or client secret.
- When epx rejects a setting, such as an API URL without HTTPS or an invalid tenant, the
  error names the rule, not the value, in case you put a credential in the wrong variable.

## Network

- Requests go only to the configured API host, over HTTPS. Plain HTTP is allowed only for
  a loopback address, for testing.
- epx never follows redirects.

## Output

epx prints responses as Emporix sends them, without redacting anything. JSON, YAML,
`epx api`, and writes print secrets that a response holds, such as API keys, payment
provider credentials, and secured settings. Tables and CSV show only their columns, but
those print values as sent too, such as the value of a secured setting.

Text from Emporix can hold characters that a terminal acts on, such as control
characters, escape sequences, bidirectional overrides, and zero-width characters. epx
replaces them with spaces in tables, CSV, and error messages, and escapes them in JSON
and YAML, where the document still decodes to the same data.

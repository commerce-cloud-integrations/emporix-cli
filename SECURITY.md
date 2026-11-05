# Security policy

epx works with live commerce data and API keys, so security reports get priority.

## Supported versions

Security fixes go into the latest release. Upgrade to it before you report an issue, and
check whether the issue still occurs there.

## Report a vulnerability

Report vulnerabilities privately through GitHub:
[Report a vulnerability](https://github.com/commerce-cloud-integrations/emporix-cli/security/advisories/new)
on the repository's **Security** tab. Don't open a public issue, pull request or discussion
for a vulnerability.

Include what you can of:

- the epx version (`epx version`) and your operating system;
- the command, with its flags, and what it did;
- what you expected instead, and why it's a security issue;
- the steps that reproduce it.

Never include your client secret, an access token or data from your tenant. If the issue
needs a response from Emporix to reproduce, describe the response's shape with made-up
values.

The maintainers aim to reply within 5 working days. Once the issue is confirmed, a fix is
released as soon as it's ready, and the advisory is published with the release that fixes
it, crediting you unless you prefer otherwise.

## In scope

Issues in epx itself, for example:

- epx printing its own access token or client secret, or storing a secret outside the OS
  keychain;
- a request sent to a host other than the configured API host, over plain HTTP to a host
  that isn't a loopback address, or after a redirect;
- a token requested with broader OAuth scopes than the command needs;
- a write sent without `--yes` or a confirmation, sent again after it may have reached
  Emporix, or sent by `--dry-run`;
- text from a response that reaches the terminal as control characters or escape sequences
  in a table, CSV or an error message;
- the installers or release archives installing anything other than the published,
  checksummed binary.

The [security model](https://commerce-cloud-integrations.github.io/emporix-cli/security)
describes these guarantees.

## Out of scope

- Vulnerabilities in the Emporix APIs or the Emporix platform: report them to Emporix.
- Secrets in the output of `-o json`, `-o yaml`, `epx api` or a write: epx prints responses
  as Emporix sends them, by design.
- What a command does with the access your API key grants, when it does what its help says.

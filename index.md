# Overview

> Read and manage your Emporix tenant from the command line.

epx is a command-line client for the [Emporix Commerce APIs](https://developer.emporix.io/api-references).
It helps you inspect, script, and change the data of an Emporix tenant right from the terminal.

With epx, you can:

- Read products, categories, catalogs, prices, availability, customers, orders, and nearly
  every other Emporix resource as a table, CSV, JSON, YAML, or JSON Lines.
- [Filter, sort, and page through lists](/emporix-cli/guides/lists) without writing paging code.
- [Change data safely](/emporix-cli/guides/writes): every write shows its exact request with `--dry-run` and
  asks before it sends.
- [Call any Emporix endpoint](/emporix-cli/guides/api) with `epx api`, using only the OAuth scopes you name.
- [Script reliably](/emporix-cli/guides/exit-codes) with stable output and exit codes that tell a missing item
  from a missing scope.

  <video autoPlay muted loop playsInline src="/images/demo.webm" aria-label="epx reading sites and products, previewing a write, and checking its API key" />

  epx is an unofficial project. It isn't affiliated with or endorsed by Emporix.

## Get started

    Install with npm, from a release archive, or with Go.
    Give epx an API key with only the scopes you need.
    Make your first reads in five minutes.

## Guides

    Choose tables, CSV, JSON, YAML, or JSON Lines, and pick columns and languages.
    Narrow lists with `--limit`, `--query`, and `--sort`.
    Send a request to any Emporix API path with `epx api`.
    Send writes with dry runs, confirmation, and explicit scopes.
    Write many items at once and wait for long-running jobs.
    Use epx in scripts and CI, and handle every failure.

## Reference

    Every command and flag, with the OAuth scopes it requests.
    How epx handles credentials, scopes, and the data it prints.
    Which Emporix API operations epx sends, per service.

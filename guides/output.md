# Format output

> Print what Emporix answers as a table, CSV, JSON, YAML, or JSON Lines, and choose columns and languages.

Every read prints a table by default. Use `-o` to choose another format:

| Format | Flag | Prints |
| --- | --- | --- |
| Table | `-o table` | A few selected columns per item. The default. |
| CSV | `-o csv` | The table's rows as CSV. |
| JSON | `-o json` | The whole response, as Emporix sends it. |
| YAML | `-o yaml` | The whole response, as YAML 1.2. |
| JSON Lines | `--stream` | One item per line, printed while the list is read. |

Results go to standard output and errors to standard error, so you can pipe and
redirect output safely.

## Tables

```sh
epx sites list
epx products list --limit 10
```

A table shows a few columns per item, with the identifier that `get` takes first. Tables
are for reading; use `-o json` in scripts.

To keep tables readable, epx shortens some long values, such as a setting's value or an
error message, to 60 characters ending in `…`. Tables and CSV also drop the trailing zeros of amounts beyond
the cents, so `35.7000000` shows as `35.70`. JSON and YAML print every value as sent.

### Choose columns

```sh
epx products list --columns code,name
epx sites list --columns code --no-headers
```

`--columns` picks the columns and their order by their headers, in any case. Quote a
header that contains a space: `--columns 'id,root product'`. Each read's help lists its
columns. `--no-headers` leaves out the header row.

`--columns` and `--no-headers` apply to tables and CSV only.

## CSV

```sh
epx products list -o csv > products.csv
```

`-o csv` prints the table's rows, quoted as RFC 4180 specifies, with the header row
first unless you pass `--no-headers`. Long values aren't shortened.

  CSV cells aren't made safe for spreadsheets. A cell that starts with `=`, `+`, `-`, or
  `@`, such as a name a shopper typed, can run as a formula when a spreadsheet opens the
  file. Import such files as text.

## JSON

```sh
epx products get 100200 -o json
epx sites list -o json | jq -r '.[].code'
```

`-o json` prints the response as Emporix sends it, indented, with object keys sorted so
that output is stable and easy to compare. Values keep their exact text, and nothing is
redacted. A list prints the items of every page as one JSON array.

`epx api` and every write print their response as JSON, too.

## YAML

```sh
epx products get 100200 -o yaml
```

`-o yaml` prints the same data as `-o json`, as YAML 1.2 in block style. Every string is
double-quoted, so values such as `"no"`, `"1.0"`, and `"null"` stay strings, and numbers
keep their exact text.

  A YAML 1.1 parser reads a number with an exponent, such as `1e5`, as a string unless it
  has a decimal point and a signed exponent (`1.5e+5`).

## JSON Lines

```sh
epx products list --stream > products.jsonl
epx orders list --query status:CREATED --stream | jq -r .id
```

`--stream` prints each item on its own line as compact JSON as soon as its page is read
and checked. Long lists start printing at once and aren't held in memory.

- Most list commands take `--stream`; `epx <command> --help` shows whether one does.
- It works with `--limit`, `--query`, and `--sort`, but not with `-o`, `--columns`, or
  `--no-headers`.
- It reads up to 10,000 pages, where other lists stop at 1,000.
- The [paging checks](https://commerce-cloud-integrations.github.io/emporix-cli/guides/lists#how-paging-works) still run. If a later page fails a
  check, the command exits with code 1 after printing the pages before it. Without
  `--stream`, a list prints nothing unless every check passes.

## Translations

Emporix stores many texts, such as names and descriptions, per language. epx requests
every translation, so `-o json` shows them all. Commands whose tables show text take
`--language` to choose the translation shown. It defaults to `en`, or to your profile's
default language.

```sh
epx products list --language nl
epx catalogs get AcmeStore_catalog --language de
```

When an item has no text in that language, the table shows another translation labeled
with its language, such as `Transistors (fr)`.

## Files

```sh
epx quotes pdf Q1001 --out Q1001.pdf --language de
epx quotes pdf Q1001 --out - | lpr
```

[`epx quotes pdf`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/quotes#epx-quotes-pdf) saves the quote's PDF with `--out`
instead of printing it. `--out -` writes the file to standard
output, but only when standard output isn't a terminal. epx writes the file readable by
you only, and replaces an existing file only once the new one is complete.

## Secrets in output

  JSON, YAML, `epx api`, and writes print every value of a response, secrets included:
  API keys of search providers, payment provider credentials, secured settings, card
  tokens, and webhook destinations. Handle such output, and any file you save it to, as a
  secret.

Characters that a terminal could act on, such as escape sequences, are replaced with
spaces in tables, CSV, and error messages, and escaped in JSON and YAML. See the
[security model](https://commerce-cloud-integrations.github.io/emporix-cli/security#output).

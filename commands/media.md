# epx media

> Read, upload, link and change media assets.

## epx media

Reads media assets: files Emporix stores (BLOB) or links to files elsewhere (LINK). ACCESS
is PUBLIC or PRIVATE; STATUS is PENDING until an upload completes, then READY or empty;
BYTES is the file size as sent, and REFERENCES the items an asset belongs to (type and
id). URLs are only in `-o json`; [`epx media download-url`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-download-url) gives one that downloads the file.

Its reads require the media.asset_read scope.
Its writes require the media.asset_manage scope.

**Command:** `epx media [flags]`, `epx media [command]`

**Flags:** only the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Subcommands**

- [`epx media create`](#epx-media-create)\
  Create a media asset: a link, or a file to upload.
- [`epx media delete`](#epx-media-delete)\
  Delete a media asset.
- [`epx media download-url`](#epx-media-download-url)\
  Show a URL that downloads an asset.
- [`epx media get`](#epx-media-get)\
  Show one asset.
- [`epx media list`](#epx-media-list)\
  List media assets.
- [`epx media patch`](#epx-media-patch)\
  Change members of a media asset.
- [`epx media start-upload`](#epx-media-start-upload)\
  Start an upload of a file straight to storage.
- [`epx media update`](#epx-media-update)\
  Replace a media asset, or the file of one.

## epx media create

Create a media asset: a link, or a file to upload.

Creates a media asset from the JSON in `--file` and prints its id. A LINK asset points to a
file elsewhere, e.g. `{"type": "LINK", "access": "PUBLIC", "url": "https://example.com/a.pdf"}`.
A BLOB asset is a file that Emporix stores: `--upload` PATH sends the file in a
multipart/form-data form, with the JSON of `--file` as its body field, e.g.
`{"type": "BLOB", "access": "PRIVATE", "details": {"filename": "a", "mimeType": "application/pdf"}}`.
Emporix takes a file of up to 30 MB including the form; a large file may need a longer
`--timeout`, e.g. `--timeout` 5m, as it bounds the whole request.
Without `--upload`, the JSON of a BLOB asset is refused before anything is sent, as Emporix
takes one only with its file.
refIds assigns the asset to items, e.g. `[{"type": "PRODUCT", "id": "100200"}]`; an asset of a
PRODUCT or CATEGORY must be PUBLIC. An id that exists already is a conflict (exit 5).

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the media.asset_manage scope.

**Command:** `epx media create [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON in PATH, sent byte for byte: the request body, or the body field of the form of a file to upload (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--upload <PATH>`\
  Upload the file in PATH (up to 30 MB including the form) as the file of a BLOB asset, in a multipart/form-data form with the JSON of `--file`; a large file may need a longer `--timeout`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

**Examples**

```sh
epx media create --file link.json --dry-run
epx media create --upload manual.pdf --file manual.json
```

## epx media delete

Delete a media asset.

Deletes the media asset ID, and the file of a BLOB asset. An unknown ID is no error: Emporix
answers as for a deleted asset.

Deletes data, which cannot be undone.
On a terminal, epx sends DELETE only once you type ID at its prompt.
`--yes` sends without asking (required without a terminal);
`--dry-run` prints the request instead of sending it.

Requires the media.asset_manage scope.

**Command:** `epx media delete ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--yes`\
  Send without asking (required without a terminal).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx media download-url

Show a URL that downloads an asset.

Reads a URL that downloads an asset. For a private stored file it is a signed URL that
works without a token until EXPIRES (15 minutes by default), so treat it as a secret; for
a public file or a link, PROVIDER CLOUDINARY or LINK, it is the asset's permanent URL.
Reading the URL of a pending upload whose file is stored makes Emporix complete it; one
whose file is not stored yet fails.
ASSET_ID is the ID that [`epx media list`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-list) shows; an unknown one is not found.

Requires the media.asset_read scope.

**Command:** `epx media download-url ASSET_ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among PROVIDER, EXPIRES, URL.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx media get

Show one asset.

Reads media assets: files Emporix stores (BLOB) or links to files elsewhere (LINK). ACCESS
is PUBLIC or PRIVATE; STATUS is PENDING until an upload completes, then READY or empty;
BYTES is the file size as sent, and REFERENCES the items an asset belongs to (type and
id). URLs are only in `-o json`; [`epx media download-url`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-download-url) gives one that downloads the file.

Requires the media.asset_read scope.

**Command:** `epx media get ID [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ACCESS, STATUS, FILE, MIME TYPE, BYTES, REFERENCES.
- `--no-headers`\
  Print table and csv without the header row.
- `-o, --output <format>`\
  Output format: table, json, yaml or csv (csv: the table's rows; json and yaml: the whole response). Default: `table`.

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

## epx media list

List media assets.

Reads media assets: files Emporix stores (BLOB) or links to files elsewhere (LINK). ACCESS
is PUBLIC or PRIVATE; STATUS is PENDING until an upload completes, then READY or empty;
BYTES is the file size as sent, and REFERENCES the items an asset belongs to (type and
id). URLs are only in `-o json`; [`epx media download-url`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-download-url) gives one that downloads the file.

Requires the media.asset_read scope.

**Command:** `epx media list [flags]`

**Flags**

- `--columns <names>`\
  Print only these columns of table and csv, in this order: comma-separated names (any case) among ID, TYPE, ACCESS, STATUS, FILE, MIME TYPE, BYTES, REFERENCES.
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

## epx media patch

Change members of a media asset.

Changes the media asset ID with the JSON Patch (RFC 6902) in `--file`, an array of operations
such as `[{"op": "add", "path": "/refIds/-", "value": {"type": "PRODUCT", "id": "100200"}}]`,
also of a BLOB asset, whose file it keeps; its type and access cannot change.
An unknown ID is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PATCH. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the media.asset_manage scope.

**Command:** `epx media patch ID [flags]`

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

## epx media start-upload

Start an upload of a file straight to storage.

Starts an upload session: creates a PENDING BLOB asset from the JSON in `--file`, as create
takes it (`"uploadType"`: `"form"` asks for a form instead of a PUT for a PRIVATE asset), and
prints where to send its file: the method, URL and headers or form fields of a request to
the storage (GCS for a PRIVATE asset, CLOUDINARY for a PUBLIC one), valid for 15 minutes
or an hour, so treat it as a secret. epx does not send the file there; once the storage
has it, the asset is READY. Emporix enables direct uploads per tenant; elsewhere it answers
403 Forbidden (exit 1) and creates no asset.

Changes data: on a terminal, epx asks before it sends POST. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the media.asset_manage scope.

**Command:** `epx media start-upload [flags]`

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

## epx media update

Replace a media asset, or the file of one.

Replaces the media asset ID with the JSON in `--file`; its type and access cannot change.
The JSON of a LINK asset gives its url; a BLOB asset is replaced with its file: `--upload` PATH
sends the new file in a multipart/form-data form, with the JSON as its body field
(up to 30 MB including the form; a large file may need a longer
`--timeout`, e.g. `--timeout` 5m). [`epx media patch`](https://commerce-cloud-integrations.github.io/emporix-cli/commands/media#epx-media-patch) changes a BLOB asset without sending its
file again.
Without `--upload`, the JSON of a BLOB asset is refused before anything is sent, as Emporix
takes one only with its file.
An unknown ID is not found (exit 4).

Changes data: on a terminal, epx asks before it sends PUT. `--yes` sends without
asking (required without a terminal, or with `--file` -); `--dry-run` prints the request
instead of sending it.

Requires the media.asset_manage scope.

**Command:** `epx media update ID [flags]`

**Flags**

- `--dry-run`\
  Print the request instead of sending it (excludes `-o`).
- `--file <PATH>`\
  JSON in PATH, sent byte for byte: the request body, or the body field of the form of a file to upload (- reads standard input; sending it then needs `--yes`; required).
- `-o, --output <format>`\
  Format of the response: json or yaml. Default: `json`.
- `--upload <PATH>`\
  Upload the file in PATH (up to 30 MB including the form) as the file of a BLOB asset, in a multipart/form-data form with the JSON of `--file`; a large file may need a longer `--timeout`.
- `--yes`\
  Send without asking (required without a terminal, or with `--file` -).

Also takes the [global flags](https://commerce-cloud-integrations.github.io/emporix-cli/commands#global-flags).

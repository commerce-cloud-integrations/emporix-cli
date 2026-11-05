"""Turns a `mint export` of site/ into a static site for GitHub Pages under a base path.

    cd EXPORT_DIR && python3 -I .github/scripts/site/pages.py BASE      e.g. BASE=/emporix-cli

It works on the export in the working directory.

- The export's own server and launchers (serve.js, Start Docs.*) and the site's tooling
  (sync.py) are removed.
- A directory's index page, which the export writes to DIR/index/index.html, moves to
  DIR/index.html, so that /commands/ and /coverage/ load; DIR/index/ redirects there.
- Every root-absolute path to one of the export's own top-level entries gains BASE: in
  the HTML of the pages, paths to pages (/commands/sites) and assets (/_next/...,
  /favicon.svg) and links to the root itself; in their payloads (the React Server
  Components data, whose text records are re-measured) and in scripts, styles and data
  files, paths to assets only.
- base.js (beside this file), loaded first by every page, prefixes BASE to the root links
  that Mintlify's router builds at runtime, loads their pages in full and stops the
  requests a static host cannot answer (router prefetches, the export's live reload).
- search.js and search.css (beside this file) are copied in and loaded by every page:
  they open Pagefind's search on the site's search button and Cmd/Ctrl+K. Run pagefind
  over the export afterwards (.github/scripts/site/build.sh does).

Standard library only.
"""

import json
import os
import re
import shutil
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
INDEX_HTML = "index.html"
REMOVED = ("serve.js", "Start Docs.command", "Start Docs.bat", "sync.py")
TEXT_SUFFIXES = (".html", ".js", ".css", ".json", ".txt", ".xml", ".webmanifest")
# Entries that exist only once the build has run, and Mintlify's API paths.
EXTRA_ENTRIES = ("pagefind", "base.js", "search.js", "search.css", "_mintlify")
# The files that agents.py writes beside the pages. Only links in the HTML name them: a page's
# payload (React reads its text back) keeps them as they are, or the page fails to hydrate.
PAGE_ENTRIES = ("llms.txt", "llms-full.txt", "sitemap.xml", "robots.txt", "index.md")
OWN_FILES = ("base.js", "search.js", "search.css")
# Loaded before any other script, so that it sees every link and request (base.js).
BASE_TAG = '<script src="{base}/base.js" data-base="{base}"></script>'
SEARCH_TAGS = '<link rel="stylesheet" href="{base}/search.css"><script defer src="{base}/search.js" data-base="{base}"></script>'


def is_page_directory(entry: Path) -> bool:
    """Whether entry is a directory of pages: one with an index.html at any depth (the
    export writes DIR/index/index.html for a directory's own page)."""
    return entry.is_dir() and entry.name != "_next" and next(entry.rglob(INDEX_HTML), None) is not None


def top_level_paths(export: Path, pages: bool = True) -> list:
    """The export's top-level entries (and EXTRA_ENTRIES), longest first; without pages,
    only those that are not page directories (assets such as _next and favicon.svg)."""
    names = {entry.name for entry in export.iterdir() if pages or not is_page_directory(entry)}
    names = (names | set(EXTRA_ENTRIES) | (set(PAGE_ENTRIES) if pages else set())) - {INDEX_HTML}
    return sorted(names, key=len, reverse=True)


def rewriter(base: str, names: list, root_links: bool = True):
    """Returns a function that prefixes base to root-absolute paths to names, and with
    root_links to links to the root itself."""
    alternatives = "|".join(re.escape(name) for name in names)
    # A path that starts a string, an attribute or a url(): after a quote (also an escaped
    # one, \"), a backtick, "(" or "=".
    # A page's Markdown twin (/install.md) is the page's path too.
    path = re.compile(r'(?<=["\'`(=])/(?=(?:%s)(?:\.md)?(?:[/"\'?#)\\]|$))' % alternatives)
    root = re.compile(r'(?<=href=")/(?=")|(?<=href\\":\\")/(?=\\")|(?<=href":")/(?=")')

    def rewrite(text: str) -> str:
        text = path.sub(base + "/", text)
        return root.sub(base + "/", text) if root_links else text

    return rewrite


# A chunk of a page's React Server Components payload: self.__next_f.push([1,"<JSON string>"]).
PAYLOAD_PUSH = re.compile(r'self\.__next_f\.push\((\[1,".*?"\])\)</script>', re.S)
# A text record of the payload, "<id>:T<length in UTF-8 bytes, hex>,<text>".
TEXT_RECORD = re.compile(rb"([0-9a-f]+):T([0-9a-f]+),")
ROW_ID = re.compile(rb"[0-9a-f]*:")  # hints such as ":HL[...]" have no id


def rewrite_payload(stream: str, rewrite) -> str:
    """Rewrites the records of a page payload: a text record's length counts its bytes, so
    it is recomputed; any other record is one line."""
    data, out, pos = stream.encode("utf-8"), [], 0
    while pos < len(data):
        text = TEXT_RECORD.match(data, pos)
        if text:
            start = text.end()
            end = start + int(text.group(2), 16)
            body = rewrite(data[start:end].decode("utf-8")).encode("utf-8")
            out.append(text.group(1) + b":T" + format(len(body), "x").encode() + b"," + body)
            pos = end
            continue
        if not ROW_ID.match(data, pos):
            raise ValueError("payload record without an id at byte %d" % pos)
        newline = data.find(b"\n", pos)
        end = len(data) if newline < 0 else newline + 1
        out.append(rewrite(data[pos:end].decode("utf-8")).encode("utf-8"))
        pos = end
    return b"".join(out).decode("utf-8")


def rewrite_page(html: str, rewrite, rewrite_assets) -> str:
    """Rewrites a page: its payload record by record with rewrite_assets, gathered into its
    first chunk (the others left empty), and the rest of the page with rewrite. The payload
    keeps its page links: Mintlify compares them with the router's path, which lacks BASE
    (base.js prefixes them in the document)."""
    pushes = list(PAYLOAD_PUSH.finditer(html))
    if not pushes:
        return rewrite(html)
    stream = rewrite_payload("".join(json.loads(push.group(1))[1] for push in pushes), rewrite_assets)
    parts, last = [], 0
    for i, push in enumerate(pushes):
        parts.append(rewrite(html[last:push.start(1)]))
        parts.append(json.dumps([1, stream if i == 0 else ""], ensure_ascii=False).replace("</", "<\\/"))
        last = push.end(1)
    parts.append(rewrite(html[last:]))
    return "".join(parts)


REDIRECT = ('<!doctype html><meta charset="utf-8"><title>Redirecting</title>'
            '<meta http-equiv="refresh" content="0; url={url}"><link rel="canonical" href="{url}">'
            '<a href="{url}">{url}</a>')


def move_index_pages(export: Path, base: str) -> None:
    """Moves DIR/index/index.html to DIR/index.html (where that is free) and leaves a
    redirect to DIR/ in its place, so that search finds each page once."""
    for index in sorted(export.rglob("index/" + INDEX_HTML)):
        directory = index.parent.parent
        if (directory / INDEX_HTML).exists() and directory != export:
            continue
        if directory != export:
            shutil.move(str(index), str(directory / INDEX_HTML))
        relative = directory.relative_to(export).as_posix()
        url = base + "/" + ("" if relative == "." else relative + "/")
        index.write_text(REDIRECT.format(url=url), encoding="utf-8")


def rewrite_files(export: Path, base: str) -> None:
    """Rewrites the text files of export for base (see the module's description)."""
    rewrite = rewriter(base, top_level_paths(export))
    rewrite_assets = rewriter(base, top_level_paths(export, pages=False), root_links=False)
    tags = SEARCH_TAGS.format(base=base)
    first = BASE_TAG.format(base=base)
    for dirpath, _, files in os.walk(export):
        for name in files:
            if not name.endswith(TEXT_SUFFIXES) or name in OWN_FILES:
                continue
            file = Path(dirpath) / name
            text = file.read_text(encoding="utf-8")
            if name.endswith(".html"):
                new = rewrite_page(text, rewrite, rewrite_assets).replace("</head>", tags + "</head>", 1)
                new = re.sub(r"<head[^>]*>", lambda head: head.group(0) + first, new, count=1)
            else:
                new = rewrite_assets(text)  # code may name pages ("/index") in other senses
            if new != text:
                file.write_text(new, encoding="utf-8")


def main(args: list) -> int:
    if len(args) != 1 or not re.fullmatch(r"(/[A-Za-z0-9._-]+)*", args[0]):
        print("usage: pages.py BASE   (in the export; BASE: /name or empty)", file=sys.stderr)
        return 2
    export, base = Path(os.path.realpath(os.getcwd())), args[0]
    if not (export / "_next").is_dir() or not (export / INDEX_HTML).is_file():
        print(f"pages.py: {export} is not an unpacked mint export", file=sys.stderr)
        return 1
    for name in REMOVED:
        (export / name).unlink(missing_ok=True)
    for name in OWN_FILES:
        shutil.copyfile(HERE / name, export / name)
    rewrite_files(export, base)
    move_index_pages(export, base)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

"""Brings the generated parts of the epx documentation site up to date with the repository.

- docs.json: the pages of the "CLI reference" and "API coverage" tabs (index first,
  then the others by name), and a footer link to the Emporix API reference commit that
  COMPATIBILITY.md was checked against.
- A check that coverage/index.mdx shows the reference commit and counts of COMPATIBILITY.md.

The site is the directory of this file; the repository is its parent. Run it after the
command reference or the coverage pages are regenerated:

    python3 -I site/sync.py           # write docs.json
    python3 -I site/sync.py --check   # fail when it is stale, writing nothing

Standard library only; Python 3.9 compatible.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

SITE = Path(__file__).resolve().parent
ROOT = SITE.parent
GITHUB = "https://github.com/commerce-cloud-integrations/emporix-cli"
REFERENCE_COMMITS = "https://github.com/emporix/api-references/commit/"

# The tabs of docs.json whose pages are generated, and the directory that holds them.
GENERATED_TABS = {"CLI reference": "commands", "API coverage": "coverage"}
FOOTER_HEADER = "Project"

# The first API reference commit link of COMPATIBILITY.md or coverage/index.mdx, then its date:
# "[`faad614`](https://github.com/emporix/api-references/commit/<sha>) (2026-10-14)".
REFERENCE_COMMIT = re.compile(re.escape(REFERENCE_COMMITS) + r"([0-9a-f]{40})\)")
REFERENCE_DATE = re.compile(r"\s*\((\d{4}-\d{2}-\d{2})\)")
# The coverage counts (built, not applicable, all operations) of COMPATIBILITY.md and of
# coverage/index.mdx, which must agree.
DOCUMENT_COUNTS = re.compile(r"\*\*Coverage:\*\* (\d+) of (\d+) operations built, (\d+) not applicable")
SITE_COUNTS = re.compile(r"\*\*Coverage:\*\* (\d+) built, (\d+) not applicable and \d+ not yet built \([^)\n]*\), of (\d+)")

class SyncError(Exception):
    """The repository does not hold what the site needs; nothing was written."""


def read(file: Path) -> str:
    try:
        return file.read_text(encoding="utf-8")
    except OSError as err:
        raise SyncError(f"cannot read {file}: {err}") from err


# --- The Emporix API reference ------------------------------------------------


def api_reference(compatibility: Path) -> dict:
    """Returns the API reference commit and date that COMPATIBILITY.md names first, and its
    counts of operations built, not applicable and in all."""
    text = read(compatibility)
    commit = REFERENCE_COMMIT.search(text)
    date = REFERENCE_DATE.match(text, commit.end()) if commit else None
    if not date:
        raise SyncError(
            f"{compatibility} names no API reference commit link followed by its date, "
            f"e.g. ({REFERENCE_COMMITS}<40-hex sha>) (YYYY-MM-DD)"
        )
    counts = DOCUMENT_COUNTS.search(text)
    if not counts:
        raise SyncError(f"{compatibility} has no coverage line (**Coverage:** N of M operations built, ...)")
    built, total, not_applicable = counts.groups()
    return {"commit": commit.group(1), "date": date.group(1), "counts": (built, not_applicable, total)}


def check_coverage(index: Path, reference: dict) -> None:
    """Fails when the coverage index does not show what COMPATIBILITY.md shows: its API
    reference commit and its counts of operations built, not applicable and in all."""
    text = read(index)
    found = REFERENCE_COMMIT.search(text)
    counts = SITE_COUNTS.search(text)
    site = (found.group(1) if found else "no", counts.groups() if counts else None)
    document = (reference["commit"], reference["counts"])
    if site != document:
        raise SyncError(
            f"{index} shows {describe(*site)}, but COMPATIBILITY.md shows {describe(*document)}; "
            "regenerate the coverage pages with COMPATIBILITY.md"
        )


def describe(commit: str, counts) -> str:
    """Describes a coverage summary for an error message."""
    if not counts:
        return f"API reference commit {commit} and no coverage line"
    built, not_applicable, total = counts
    return f"API reference commit {commit}, {built} built, {not_applicable} not applicable of {total} operations"


# --- docs.json -----------------------------------------------------------------


def site_config(config: dict, reference: dict) -> dict:
    """Returns config with the pages of the generated tabs and the footer's reference link."""
    tabs = config["navigation"]["tabs"]
    for tab in tabs:
        directory = GENERATED_TABS.get(tab["tab"])
        if directory:
            if len(tab.get("groups", [])) != 1:
                raise SyncError(f'docs.json: the tab "{tab["tab"]}" must hold one group, for its generated pages')
            tab["groups"][0]["pages"] = directory_pages(directory)
    missing = set(GENERATED_TABS) - {tab["tab"] for tab in tabs}
    if missing:
        raise SyncError(f"docs.json has no tab {', '.join(sorted(missing))} for its generated pages")
    config["footer"]["links"] = [{
        "header": FOOTER_HEADER,
        "items": [
            {"label": "Source on GitHub", "href": GITHUB},
            {
                "label": f"Emporix API reference {reference['commit'][:7]} ({reference['date']})",
                "href": REFERENCE_COMMITS + reference["commit"],
            },
        ],
    }]
    return config


def directory_pages(directory: str) -> list:
    """Returns the pages of a site directory: index first, then the others by name."""
    pages = sorted((SITE / directory).glob("*.mdx"), key=lambda page: (page.name != "index.mdx", page.name))
    if not pages:
        raise SyncError(f"{SITE / directory} holds no pages")
    return [f"{directory}/{page.stem}" for page in pages]


# --- Main ------------------------------------------------------------------------


def outputs() -> dict:
    """Returns {file: content} of the generated files, after checking the coverage pages."""
    reference = api_reference(ROOT / "COMPATIBILITY.md")
    check_coverage(SITE / "coverage" / "index.mdx", reference)
    config_file = SITE / "docs.json"
    try:
        config = json.loads(read(config_file))
    except json.JSONDecodeError as err:
        raise SyncError(f"{config_file} is not JSON: {err}") from err
    return {
        config_file: json.dumps(site_config(config, reference), indent=2, ensure_ascii=False) + "\n",
    }


def main(args: list) -> int:
    if args not in ([], ["--check"]):
        print("usage: python3 -I site/sync.py [--check]", file=sys.stderr)
        return 2
    try:
        files = outputs()
    except SyncError as err:
        print(f"site/sync.py: {err}", file=sys.stderr)
        return 1
    stale = [file for file, content in files.items() if not file.exists() or file.read_text(encoding="utf-8") != content]
    if args == ["--check"]:
        for file in stale:
            print(f"site/sync.py: {file.relative_to(ROOT)} is stale; run python3 -I site/sync.py", file=sys.stderr)
        return 1 if stale else 0
    for file in stale:
        with open(file, "w", encoding="utf-8", newline="\n") as out:
            out.write(files[file])
        print(f"wrote {file.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

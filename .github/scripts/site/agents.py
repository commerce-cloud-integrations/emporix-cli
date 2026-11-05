"""Adds what LLM tools and search crawlers read to a static export of site/, for GitHub Pages.

    python3 -I .github/scripts/site/agents.py EXPORT_DIR SITE_DIR BASE HOST
    e.g. BASE=/emporix-cli HOST=https://commerce-cloud-integrations.github.io

Run after pages.py. For every page of the navigation in SITE_DIR/docs.json it writes:

- a Markdown copy of the page, at the page's path with .md (install.md, commands/products.md),
  which the HTML's text/markdown alternate links name; its links are made absolute;
- a canonical link and a JSON-LD description in the page's <head>.

And, at the root of EXPORT_DIR: llms.txt (the pages, by section, with their Markdown
URLs), llms-full.txt (every page's Markdown in one file), sitemap.xml and robots.txt.

Standard library only.
"""

import html
import json
import re
import sys
from pathlib import Path

FENCE = re.compile(r"^\s*```")
IMPORT = re.compile(r"^import\s.+\sfrom\s.+$")
# A line that is only a JSX tag, such as <Info> or </Card>; its content stays.
JSX_TAG = re.compile(r"^\s*</?[A-Z][\w.]*(\s[^<>]*)?/?>\s*$")
# An opening tag with string attributes, such as <Step title="Set your API key"> or <Card href="/x">.
OPEN_TAG = re.compile(r'^\s*<([A-Z][\w.]*)((?:\s+[\w-]+="[^"]*")*)\s*/?>\s*$')
ATTRIBUTE = re.compile(r'([\w-]+)="([^"]*)"')
MD_LINK = re.compile(r"\]\((/[^)\s]*)\)")
FRONT = re.compile(r"^---\n(.*?)\n---\n", re.S)


def pages_of(node, group="", out=None):
    """The pages of a navigation node, in order, as (group, page): strings are pages, and a
    dict's "group" names the pages beneath it."""
    out = [] if out is None else out
    if isinstance(node, str):
        out.append((group, node))
    elif isinstance(node, list):
        for item in node:
            pages_of(item, group, out)
    elif isinstance(node, dict):
        group = node.get("group", group)
        for key in ("tabs", "groups", "pages"):
            if key in node:
                pages_of(node[key], group, out)
    return out


def frontmatter(text: str) -> tuple:
    """The fields of a page's frontmatter (simple "key: value" lines) and its body."""
    match = FRONT.match(text)
    if not match:
        return {}, text
    fields = {}
    for line in match.group(1).splitlines():
        key, sep, value = line.partition(":")
        if sep:
            fields[key.strip()] = value.strip().strip('"')
    return fields, text[match.end():]


def heading_of(attributes: dict) -> str:
    """The Markdown heading that stands for an opening tag with a title (a step, a card, a tab):
    a card with an href becomes a link. A tag without a title leaves no line."""
    title = attributes.get("title")
    if not title:
        return ""
    href = attributes.get("href")
    if href:
        return "### [%s](%s)" % (title, href)  # its href gains BASE with the other links
    return "### " + title


def markdown(body: str, origin: str) -> str:
    """The body of an MDX page as Markdown: imports and JSX-only lines go, a titled tag becomes
    a heading (its title is content), MDX's escaped braces and underscores are plain, and the
    links to pages of the site become absolute, under origin (outside code blocks)."""
    out, fenced = [], False
    for line in body.splitlines():
        if FENCE.match(line):
            fenced = not fenced
        elif not fenced:
            opening = OPEN_TAG.match(line)
            if IMPORT.match(line) or JSX_TAG.match(line) and not opening:
                continue
            if opening:
                line = heading_of(dict(ATTRIBUTE.findall(opening.group(2))))
                if not line:
                    continue
            line = line.replace("\\{", "{").replace("\\}", "}").replace("\\_", "_")
            line = MD_LINK.sub(lambda link: "](" + origin + link.group(1) + ")", line)
        out.append(line)
    return "\n".join(out).strip() + "\n"


def page_path(page: str) -> str:
    """The URL path of a page under the site: index is the root, and a directory's own page
    (commands/index) is the directory (commands)."""
    if page == "index":
        return ""
    return page[: -len("/index")] if page.endswith("/index") else page


def json_ld(data: dict) -> str:
    text = json.dumps(data, ensure_ascii=False).replace("</", "<\\/")
    return '<script type="application/ld+json">' + text + "</script>"


def add_head(page_html: str, canonical: str, title: str, description: str, site: dict) -> str:
    """Adds the canonical link and JSON-LD to a page's head, once."""
    if 'rel="canonical"' in page_html:
        return page_html
    kind = "WebSite" if canonical == site["url"] else "TechArticle"
    data = {"@context": "https://schema.org", "@type": kind, "headline": title, "name": title,
            "url": canonical, "isPartOf": {"@type": "WebSite", "name": site["name"], "url": site["url"]}}
    if description:
        data["description"] = description
    link = html.escape(canonical, quote=True)
    tags = ('<link rel="canonical" href="%s"><meta property="og:url" content="%s">' % (link, link)) + json_ld(data)
    return page_html.replace("</head>", tags + "</head>", 1)


def main(args: list) -> int:
    if len(args) != 4:
        print("usage: agents.py EXPORT_DIR SITE_DIR BASE HOST", file=sys.stderr)
        return 2
    export, source, base, host = Path(args[0]), Path(args[1]), args[2].rstrip("/"), args[3].rstrip("/")
    if not re.fullmatch(r"(/[A-Za-z0-9._-]+)*", base) or not host.startswith("https://"):
        print("agents.py: BASE is /name or empty, and HOST an https URL", file=sys.stderr)
        return 2
    docs = json.loads((source / "docs.json").read_text(encoding="utf-8"))
    site = {"name": docs.get("name", ""), "url": host + base + "/"}
    entries = []  # (group, page, title, description, canonical, markdown URL, markdown text)
    seen = set()
    for group, page in pages_of(docs["navigation"]):
        if page in seen:
            continue
        seen.add(page)
        mdx = source / (page + ".mdx")
        target = export / (page_path(page) + "/index.html" if page != "index" else "index.html")
        if not mdx.is_file() or not target.is_file():
            print(f"agents.py: page {page} has no source or no exported HTML", file=sys.stderr)
            return 1
        fields, body = frontmatter(mdx.read_text(encoding="utf-8"))
        title = fields.get("title", page)
        description = fields.get("description", "")
        rel = page_path(page)
        canonical = host + base + "/" + (rel + "/" if rel else "")
        md_path = export / (page + ".md")  # the name Mintlify's alternate links use (commands/index.md)
        md_url = host + base + "/" + page + ".md"
        text = "# " + title + "\n\n" + ("> " + description + "\n\n" if description else "") + markdown(body, host + base)
        md_path.parent.mkdir(parents=True, exist_ok=True)
        md_path.write_text(text, encoding="utf-8")
        page_html = target.read_text(encoding="utf-8")
        target.write_text(add_head(page_html, canonical, title, description, site), encoding="utf-8")
        entries.append((group, page, title, description, canonical, md_url, text))
    write_indexes(export, docs, site, host, base, entries)
    return 0


def write_indexes(export: Path, docs: dict, site: dict, host: str, base: str, entries: list) -> None:
    """Writes llms.txt, llms-full.txt, sitemap.xml and robots.txt at the root of export."""
    lines = ["# " + site["name"], "", "> " + docs.get("description", "")]
    current = None
    for group, _, title, description, _, md_url, _ in entries:
        if group != current:
            current = group
            lines += ["", "## " + (group or "Pages"), ""]
        lines.append("- [%s](%s)%s" % (title, md_url, ": " + description if description else ""))
    (export / "llms.txt").write_text("\n".join(lines) + "\n", encoding="utf-8")
    full = ["# %s — full documentation" % site["name"], ""]
    for _, _, _, _, canonical, _, text in entries:
        full += ["Source: " + canonical, "", text, "---", ""]
    (export / "llms-full.txt").write_text("\n".join(full), encoding="utf-8")
    urls = "".join("  <url><loc>%s</loc></url>\n" % html.escape(entry[4]) for entry in entries)
    sitemap = ('<?xml version="1.0" encoding="UTF-8"?>\n'
               '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n' + urls + "</urlset>\n")
    (export / "sitemap.xml").write_text(sitemap, encoding="utf-8")
    (export / "robots.txt").write_text(
        "User-agent: *\nAllow: /\n\nSitemap: %s%s/sitemap.xml\n" % (host, base), encoding="utf-8")


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

#!/bin/bash
# Builds the docs site (site/) as a static site for GitHub Pages: `mint export`, the
# base-path rewrite and search glue of pages.py, the Markdown copies, llms.txt and sitemap
# of agents.py, then a Pagefind index of the pages. .github/workflows/pages.yml runs it for
# every pushed vMAJOR.MINOR.PATCH tag.
# Usage: .github/scripts/site/build.sh [OUTDIR] [BASE]   (default: .local/site-pages, /emporix-cli)
# MINT and PAGEFIND name the two programs (default: as `npm install mint pagefind` puts
# them into .local/mint and .local/pagefind).
# SITE_HOST (default https://commerce-cloud-integrations.github.io) is the origin the sitemap and
# canonical links name.
# Preview: python3 -m http.server 8100 -d "$(dirname OUTDIR)" and open
# http://localhost:8100/emporix-cli/ (the parent directory must be named as BASE; the
# script makes OUTDIR/../pages-root/emporix-cli point at OUTDIR for that).
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../.." && pwd)
out=${1:-$repo/.local/site-pages}
case $out in /*) ;; *) out=$PWD/$out ;; esac
base=${2-/emporix-cli}
mint=${MINT:-$repo/.local/mint/node_modules/.bin/mint}
pagefind=${PAGEFIND:-$repo/.local/pagefind/node_modules/.bin/pagefind}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

python3 -I "$repo/site/sync.py" --check
(cd "$repo/site" && "$mint" export --telemetry false --output "$work/export.zip" >"$work/export.log" 2>&1) || {
	cat "$work/export.log" >&2
	exit 1
}
mkdir "$work/site"
unzip -q "$work/export.zip" -d "$work/site"
(cd "$work/site" && python3 -I "$here/pages.py" "$base" &&
	python3 -I "$here/agents.py" "$base" "${SITE_HOST:-https://commerce-cloud-integrations.github.io}")
"$pagefind" --site "$work/site" --root-selector "#content-area" \
	--exclude-selectors "#page-context-menu, #page-context-menu-button, button" >"$work/pagefind.log"
grep -E "Indexed [0-9]+ pages" "$work/pagefind.log"
rm -rf "$out"
mv "$work/site" "$out"
if [[ -n "$base" ]]; then
	link=$(dirname "$out")/pages-root/${base#/}
	mkdir -p "$(dirname "$link")"
	ln -sfn "$out" "$link"
fi
echo "built $out (base ${base:-/})"

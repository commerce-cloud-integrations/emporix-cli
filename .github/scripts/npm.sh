#!/usr/bin/env bash
#   .github/scripts/npm.sh DIST            pack and check: DIST is dist/<version> as
#                                          release.sh wrote it
#   .github/scripts/npm.sh DIST --publish  the same, then `npm publish` every version that
#                                          the registry does not hold yet: in GitHub Actions
#                                          through trusted publishing (OIDC, with provenance),
#                                          elsewhere after `npm login`. Re-run it after a failure.
set -euo pipefail

fail() { echo "npm: $*" >&2; exit 1; }
usage() { echo "usage: $0 DIST [--publish]" >&2; exit 2; }

[[ $# -ge 1 && $# -le 2 ]] || usage
dist=${1%/}
mode=${2:-pack}
case $mode in pack | --publish) ;; *) usage ;; esac
cd "$(git rev-parse --show-toplevel)"
[[ -d "$dist" ]] || fail "$dist does not exist; run scripts/release.sh first"
dist=$(cd "$dist" && pwd)

tag=$(basename "$dist")
if [[ $tag =~ ^v((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*))$ ]]; then
	version=${BASH_REMATCH[1]}
	commit=$(git rev-parse --verify --quiet "refs/tags/$tag^{commit}") || fail "no local tag $tag"
elif [[ $tag =~ ^snapshot-([0-9a-f]+)$ ]]; then
	[[ "$mode" = pack ]] || fail "a snapshot is never published"
	# The g keeps the identifier alphanumeric: an all-digit hash with a leading zero is not semver.
	version=0.0.0-snapshot.g${BASH_REMATCH[1]}
	commit=$(git rev-parse --verify --quiet "${BASH_REMATCH[1]}^{commit}") || fail "unknown commit in $tag"
else
	fail "$dist is not dist/vMAJOR.MINOR.PATCH or dist/snapshot-<commit>"
fi

stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT

sums=epx_${tag}_checksums.txt
[[ -f "$dist/$sums" ]] || fail "$dist has no $sums"
(cd "$dist" && shasum -a 256 --check --quiet "$sums") || fail "the archives do not match $sums"
archives=()
while read -r _ file; do
	case $file in *.tar.gz | *.zip) archives+=("$file") ;; *) fail "unknown archive type: $file" ;; esac
done <"$dist/$sums"
[[ ${#archives[@]} -gt 0 ]] || fail "$sums lists no archive"

base() {
	local file=$1
	if [[ $file == *.zip ]]; then echo "${file%.zip}"; else echo "${file%.tar.gz}"; fi
}

npm_platform() {
	local file=$1 rest
	rest=$(base "$file")
	rest=${rest#"epx_${tag}_"}
	local goos=${rest%_*} goarch=${rest#*_}
	echo "${goos/windows/win32}-${goarch/amd64/x64}"
}

exe_of() {
	local platform=$1
	if [[ $platform == win32-* ]]; then echo epx.exe; else echo epx; fi
}

# npm refused the unscoped epx-cli as too similar to other names, and epx-cli-win32-* as spam.
pkg=@commercecloudintegrations/epx-cli
tgz_base=commercecloudintegrations-epx-cli
alias_of() {
	local platform=$1
	echo "epx-cli-$platform"
}
tgz_of() {
	local platform=$1
	echo "$out/$tgz_base-$version-$platform.tgz"
}

out=$dist/npm
main_tgz=$out/$tgz_base-$version.tgz
packed=$stage/npm

sources() {
	local into=$1
	if [[ $tag == v* ]]; then
		git archive "$commit" npm/epx-cli LICENSE README.md | tar -x -C "$into"
	else
		mkdir -p "$into/npm"
		cp -R npm/epx-cli "$into/npm/"
		cp LICENSE README.md "$into/"
	fi
}

pack_platform() {
	local file=$1 src=$2 platform name exe dir
	platform=$(npm_platform "$file")
	name=$pkg@$version-$platform
	exe=$(exe_of "$platform")
	dir=$stage/$(alias_of "$platform")
	mkdir -p "$dir/bin"
	case $file in
	*.zip) unzip -q "$dist/$file" -d "$stage" ;;
	*) tar -C "$stage" -xzf "$dist/$file" ;;
	esac
	mv "$stage/$(base "$file")/$exe" "$dir/bin/$exe"
	chmod 755 "$dir/bin/$exe"
	cp "$src/LICENSE" "$dir/"
	jq --arg pkg "$pkg" --arg version "$version-$platform" --arg os "${platform%-*}" \
		--arg cpu "${platform#*-}" --arg exe "bin/$exe" '{homepage, repository, license} + {
			name: $pkg, version: $version,
			description: "The epx binary of \($pkg) for \($os) \($cpu)",
			os: [$os], cpu: [$cpu], files: [$exe], preferUnplugged: true
		}' "$src/npm/epx-cli/package.json" >"$dir/package.json"
	# shellcheck disable=SC2016 # the backticks are Markdown, not a command substitution
	printf '# %s\n\nThe `epx` binary for %s. Install [%s](https://www.npmjs.com/package/%s) instead of this version.\n' \
		"$name" "$platform" "$pkg" "$pkg" >"$dir/README.md"
	(cd "$packed" && npm pack --silent "$dir" >/dev/null) || fail "cannot pack $name"
}

deps() {
	jq -n --arg pkg "$pkg" --arg version "$version" --args \
		'$ARGS.positional | map({("epx-cli-" + .): "npm:\($pkg)@\($version)-\(.)"}) | add' "$@"
}

pack() {
	local src=$stage/src main=$stage/epx-cli file platforms=()
	mkdir -p "$src" "$packed"
	sources "$src"
	node --test --test-reporter=dot "$src"/npm/epx-cli/test/*.test.js || fail "the launcher tests fail"
	for file in "${archives[@]}"; do
		pack_platform "$file" "$src"
		platforms+=("$(npm_platform "$file")")
	done
	cp -R "$src/npm/epx-cli" "$main"
	cp "$src/LICENSE" "$src/README.md" "$main/"
	for file in "${archives[@]}"; do
		[[ $file == *.tar.gz ]] && cp -R "$stage/$(base "$file")/man" "$main/man" && break
	done
	[[ -d "$main/man" ]] || fail "no archive holds the man pages"
	jq --arg version "$version" --argjson deps "$(deps "${platforms[@]}")" --args \
		'del(.private) | .version = $version | .optionalDependencies = $deps
		 | .files += ["man"] | .man = $ARGS.positional' \
		$(cd "$main" && ls man/*.1) <"$src/npm/epx-cli/package.json" >"$main/package.json"
	(cd "$packed" && npm pack --silent "$main" >/dev/null) || fail "cannot pack $pkg"
}

manifest() {
	local tgz=$1
	tar -xOzf "$tgz" package/package.json
}

check() {
	local file platform tgz exe platforms=() actual
	for file in "${archives[@]}"; do
		platform=$(npm_platform "$file")
		platforms+=("$platform")
		tgz=$(tgz_of "$platform")
		exe=$(exe_of "$platform")
		[[ -f "$tgz" ]] || fail "missing $tgz"
		manifest "$tgz" | jq -e --arg pkg "$pkg" --arg version "$version-$platform" \
			--arg os "${platform%-*}" --arg cpu "${platform#*-}" \
			'.name == $pkg and .version == $version and .os == [$os] and .cpu == [$cpu]
			 and (has("scripts") | not) and (has("bin") | not)' >/dev/null ||
			fail "$tgz has a wrong package.json"
		tar -tvzf "$tgz" "package/bin/$exe" | grep -q '^-rwxr-xr-x' || fail "$tgz has no executable bin/$exe"
	done
	[[ -f "$main_tgz" ]] || fail "missing $main_tgz"
	actual=$(cd "$out" && find . -name '*.tgz' ! -name "$(basename "$main_tgz")" | sed 's|^\./||' | sort)
	[[ "$actual" = "$(for platform in "${platforms[@]}"; do basename "$(tgz_of "$platform")"; done | sort)" ]] ||
		fail "$out holds other tarballs than the platforms of $sums"
	manifest "$main_tgz" | jq -e --arg pkg "$pkg" --arg version "$version" --argjson deps "$(deps "${platforms[@]}")" '
		.name == $pkg and .version == $version and (has("private") | not)
		and (has("scripts") | not) and .bin == {"epx": "bin/epx.js"}
		and .optionalDependencies == $deps
		and (.man | length > 1) and (.man | index("man/epx.1"))' >/dev/null || fail "$main_tgz has a wrong package.json"
	for page in $(manifest "$main_tgz" | jq -r '.man[]'); do
		tar -tzf "$main_tgz" "package/$page" >/dev/null 2>&1 || fail "$main_tgz lacks its man page $page"
	done
}

smoke() {
	local proj=$stage/proj host overrides='{}' file platform reported info
	mkdir -p "$proj"
	for file in "${archives[@]}"; do
		platform=$(npm_platform "$file")
		overrides=$(jq --arg name "$(alias_of "$platform")" --arg spec "file:$(tgz_of "$platform")" \
			'. + {($name): $spec}' <<<"$overrides")
	done
	jq -n --arg pkg "$pkg" --arg main "file:$main_tgz" --argjson overrides "$overrides" \
		'{private: true, dependencies: {($pkg): $main}, overrides: $overrides}' >"$proj/package.json"
	(cd "$proj" && npm install --offline --ignore-scripts --silent --no-audit --no-fund) ||
		fail "cannot install the tarballs"
	reported=$("$proj/node_modules/.bin/epx" version) || fail "the installed epx does not run"
	host=$(node -p 'process.platform + "-" + process.arch')
	info=$(go version -m "$(cd "$proj" && node -p "require.resolve('$(alias_of "$host")/bin/$(exe_of "$host")')")") ||
		fail "cannot read the build info of the installed binary"
	grep -q "vcs.revision=$commit" <<<"$info" || fail "the installed epx was not built from $commit"
	if [[ $tag == v* ]] && [[ $reported != "epx $tag "* ]]; then
		fail "the installed epx reports '$reported', not $tag"
	fi
	echo "npm: $out checked ($reported)"
}

# npm never replaces a version.
publish() {
	local at=$1 tgz=$2 dist_tag=$3 published local_sum
	published=$(npm view "$pkg@$at" dist.integrity 2>/dev/null) || published=
	if [[ -n "$published" ]]; then
		local_sum=sha512-$(openssl dgst -sha512 -binary "$tgz" | base64 | tr -d '\n')
		[[ "$published" = "$local_sum" ]] || fail "$pkg@$at on the registry is not $tgz"
		echo "npm: $pkg@$at is already published"
		return
	fi
	npm publish --access public --tag "$dist_tag" "$tgz" || fail "cannot publish $pkg@$at"
}

pack
if [[ -e "$out" ]]; then
	diff -r "$packed" "$out" >/dev/null || fail "$out differs from a fresh pack; delete it to pack again"
	echo "npm: $out matches a fresh pack"
else
	mv "$packed" "$out"
fi
check
smoke
ls -1 "$out"

bounded() {
	local seconds=$1
	shift
	if command -v timeout >/dev/null; then
		timeout "$seconds" "$@"
	else
		"$@"
	fi
}

# npm processes a publish for a few minutes before installs can fetch it.
available() {
	local at=$1 deadline=$2 left
	while left=$((deadline - SECONDS)) && [[ "$left" -gt 0 ]]; do
		if bounded "$left" npm view "$pkg@$at" version --fetch-timeout="$((left * 1000))" --fetch-retries=0 \
			>/dev/null 2>&1; then
			return 0
		fi
		left=$((deadline - SECONDS))
		if [[ "$left" -gt 0 ]]; then
			sleep "$((left < 30 ? left : 30))"
		fi
	done
	fail "$pkg@$at is still not available when the wait for the platform versions ends"
}

if [[ "$mode" = --publish ]]; then
	remote=$(git ls-remote origin "refs/tags/$tag" "refs/tags/$tag^{}" | awk '{ c = $1 } END { print c }')
	[[ "$remote" = "$commit" ]] || fail "origin's tag $tag does not name $commit; push the tag first"
	# GitHub Actions publishes through trusted publishing (OIDC): there is no login to check.
	if [[ -z "${ACTIONS_ID_TOKEN_REQUEST_URL:-}" ]]; then
		npm whoami >/dev/null || fail "not logged in to npm; run npm login"
	fi
	for file in "${archives[@]}"; do
		platform=$(npm_platform "$file")
		publish "$version-$platform" "$(tgz_of "$platform")" "$platform"
	done
	deadline=$((SECONDS + 15 * 60))
	for file in "${archives[@]}"; do
		available "$version-$(npm_platform "$file")" "$deadline"
	done
	publish "$version" "$main_tgz" latest
fi

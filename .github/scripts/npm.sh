#!/usr/bin/env bash
# Builds the npm package of epx from the archives of .github/scripts/release.sh;
# .github/workflows/release.yml runs it for every pushed vMAJOR.MINOR.PATCH tag.
#
#   .github/scripts/npm.sh DIST            pack and check: DIST is dist/<version> as
#                                          release.sh wrote it
#   .github/scripts/npm.sh DIST --publish  the same, then `npm publish` every version that
#                                          the registry does not hold yet: in GitHub Actions
#                                          through trusted publishing (OIDC, with provenance),
#                                          elsewhere after `npm login`. Re-run it after a failure.
#
# The binaries ship inside npm, as @openai/codex does it, in the one package
# @commercecloudintegrations/epx-cli (the organization's scope: npm refused the unscoped
# epx-cli as too similar to other names, and epx-cli-win32-* as spam): one version per
# platform (<version>-<os>-<cpu>, with npm's `os` and `cpu` fields, so npm installs only
# the matching one, published under the dist-tag <os>-<cpu>) and the main version
# <version>, which lists them as optional dependencies under the aliases epx-cli-<os>-<cpu>
# (npm:<package>@<version>-<os>-<cpu>) and runs the right binary from
# npm/epx-cli/bin/epx.js. No install script runs and nothing is
# downloaded from GitHub.
#
# The platforms are those of the release archives, verified against its checksums file.
# For a release, every other published file (launcher, manifest, README, LICENSE) comes
# from the tag, never from the working tree; a snapshot takes them from the working tree.
# Before anything is published, the tarballs are checked (names, versions, os/cpu, the
# optional dependencies of the main version, an executable binary) and installed into a
# temporary project as npm installs them from the registry: the main version from its tarball, its optional
# dependencies resolved to theirs (offline), so npm itself picks the host's package; the
# installed epx must then run and carry the release's commit. A package version already on
# the registry is skipped only when it is byte for byte the local tarball.
# Output: DIST/npm/*.tgz; a later run packs again and must produce the same bytes.
set -euo pipefail

fail() { echo "npm: $*" >&2; exit 1; }
usage() { echo "usage: $0 DIST [--publish]" >&2; exit 2; }

[ $# -ge 1 ] && [ $# -le 2 ] || usage
dist=${1%/}
mode=${2:-pack}
case $mode in pack | --publish) ;; *) usage ;; esac
cd "$(git rev-parse --show-toplevel)"
[ -d "$dist" ] || fail "$dist does not exist; run scripts/release.sh first"
dist=$(cd "$dist" && pwd)

tag=$(basename "$dist")
if [[ $tag =~ ^v((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*))$ ]]; then
	version=${BASH_REMATCH[1]}
	commit=$(git rev-parse --verify --quiet "refs/tags/$tag^{commit}") || fail "no local tag $tag"
elif [[ $tag =~ ^snapshot-([0-9a-f]+)$ ]]; then
	[ "$mode" = pack ] || fail "a snapshot is never published"
	# The g keeps the identifier alphanumeric: an all-digit hash with a leading zero is not semver.
	version=0.0.0-snapshot.g${BASH_REMATCH[1]}
	commit=$(git rev-parse --verify --quiet "${BASH_REMATCH[1]}^{commit}") || fail "unknown commit in $tag"
else
	fail "$dist is not dist/vMAJOR.MINOR.PATCH or dist/snapshot-<commit>"
fi

stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT

# The platforms of the release, from its checksums file: epx_<tag>_<goos>_<goarch>.<ext>.
sums=epx_${tag}_checksums.txt
[ -f "$dist/$sums" ] || fail "$dist has no $sums"
(cd "$dist" && shasum -a 256 --check --quiet "$sums") || fail "the archives do not match $sums"
archives=()
while read -r _ file; do
	case $file in *.tar.gz | *.zip) archives+=("$file") ;; *) fail "unknown archive type: $file" ;; esac
done <"$dist/$sums"
[ ${#archives[@]} -gt 0 ] || fail "$sums lists no archive"

# base strips the extension of an archive name (.tar.gz or .zip, checked above).
base() { if [[ $1 == *.zip ]]; then echo "${1%.zip}"; else echo "${1%.tar.gz}"; fi; }

# npm_platform maps an archive name to npm's <os>-<cpu> (Go's windows/amd64 is win32/x64).
npm_platform() {
	local rest
	rest=$(base "$1")
	rest=${rest#"epx_${tag}_"}
	local goos=${rest%_*} goarch=${rest#*_}
	echo "${goos/windows/win32}-${goarch/amd64/x64}"
}

# exe_of prints the binary name of an npm platform.
exe_of() { if [[ $1 == win32-* ]]; then echo epx.exe; else echo epx; fi; }

# pkg is the npm package (as npm/epx-cli/package.json names it; check verifies it); npm
# pack names its tarballs <scope>-<name>-<version>.tgz. alias_of prints the name under
# which the package installs the version of an npm platform; tgz_of the tarball of that
# version.
pkg=@commercecloudintegrations/epx-cli
tgz_base=commercecloudintegrations-epx-cli
alias_of() { echo "epx-cli-$1"; }
tgz_of() { echo "$out/$tgz_base-$version-$1.tgz"; }

out=$dist/npm
main_tgz=$out/$tgz_base-$version.tgz
packed=$stage/npm

# sources copies the tracked inputs of the packages into $1: from the tag for a release.
sources() {
	if [[ $tag == v* ]]; then
		git archive "$commit" npm/epx-cli LICENSE README.md | tar -x -C "$1"
	else
		mkdir -p "$1/npm"
		cp -R npm/epx-cli "$1/npm/"
		cp LICENSE README.md "$1/"
	fi
}

# pack_platform packs the platform package of one archive.
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

# deps prints the optional dependencies of the main version for the given npm platforms.
deps() {
	jq -n --arg pkg "$pkg" --arg version "$version" --args \
		'$ARGS.positional | map({("epx-cli-" + .): "npm:\($pkg)@\($version)-\(.)"}) | add' "$@"
}

# pack writes every tarball into $packed: the platform versions of the archives, then
# the main version, whose optional dependencies are exactly the platform versions packed.
pack() {
	local src=$stage/src main=$stage/epx-cli file platforms=()
	mkdir -p "$src" "$packed"
	sources "$src"
	# The launcher's own tests, on the launcher that is packed (the tag's for a release).
	node --test --test-reporter=dot "$src"/npm/epx-cli/test/*.test.js || fail "the launcher tests fail"
	for file in "${archives[@]}"; do
		pack_platform "$file" "$src"
		platforms+=("$(npm_platform "$file")")
	done
	cp -R "$src/npm/epx-cli" "$main"
	cp "$src/LICENSE" "$src/README.md" "$main/"
	jq --arg version "$version" --argjson deps "$(deps "${platforms[@]}")" \
		'del(.private) | .version = $version | .optionalDependencies = $deps' \
		<"$src/npm/epx-cli/package.json" >"$main/package.json"
	(cd "$packed" && npm pack --silent "$main" >/dev/null) || fail "cannot pack $pkg"
}

# manifest prints the package.json of a tarball.
manifest() { tar -xOzf "$1" package/package.json; }

# check verifies every tarball: the platform versions of the archives, nothing else,
# and the main version depending on exactly those.
check() {
	local file platform tgz exe platforms=() actual
	for file in "${archives[@]}"; do
		platform=$(npm_platform "$file")
		platforms+=("$platform")
		tgz=$(tgz_of "$platform")
		exe=$(exe_of "$platform")
		[ -f "$tgz" ] || fail "missing $tgz"
		manifest "$tgz" | jq -e --arg pkg "$pkg" --arg version "$version-$platform" \
			--arg os "${platform%-*}" --arg cpu "${platform#*-}" \
			'.name == $pkg and .version == $version and .os == [$os] and .cpu == [$cpu]
			 and (has("scripts") | not) and (has("bin") | not)' >/dev/null ||
			fail "$tgz has a wrong package.json"
		tar -tvzf "$tgz" "package/bin/$exe" | grep -q '^-rwxr-xr-x' || fail "$tgz has no executable bin/$exe"
	done
	[ -f "$main_tgz" ] || fail "missing $main_tgz"
	actual=$(cd "$out" && find . -name '*.tgz' ! -name "$(basename "$main_tgz")" | sed 's|^\./||' | sort)
	[ "$actual" = "$(for platform in "${platforms[@]}"; do basename "$(tgz_of "$platform")"; done | sort)" ] ||
		fail "$out holds other tarballs than the platforms of $sums"
	manifest "$main_tgz" | jq -e --arg pkg "$pkg" --arg version "$version" --argjson deps "$(deps "${platforms[@]}")" '
		.name == $pkg and .version == $version and (has("private") | not)
		and (has("scripts") | not) and .bin == {"epx": "bin/epx.js"}
		and .optionalDependencies == $deps' >/dev/null || fail "$main_tgz has a wrong package.json"
}

# smoke installs the package as npm does from the registry, its optional dependencies
# (aliases) resolved to the tarballs (offline), and runs the epx that npm picked for
# this host.
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
	(cd "$proj" && npm install --offline --silent --no-audit --no-fund) || fail "cannot install the tarballs"
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

# publish publishes the version $1 of the package from the tarball $2 under the dist-tag $3,
# or skips it when the registry already holds this version with the same content (a run
# that stopped halfway); a different content under the same version fails, since npm
# never replaces a version.
publish() {
	local published local_sum
	published=$(npm view "$pkg@$1" dist.integrity 2>/dev/null) || published=
	if [ -n "$published" ]; then
		local_sum=sha512-$(openssl dgst -sha512 -binary "$2" | base64 | tr -d '\n')
		[ "$published" = "$local_sum" ] || fail "$pkg@$1 on the registry is not $2"
		echo "npm: $pkg@$1 is already published"
		return
	fi
	npm publish --access public --tag "$3" "$2" || fail "cannot publish $pkg@$1"
}

# Every run packs from the verified archives (and the tag); packing is deterministic, so a
# later run (a resumed publish) must produce exactly the tarballs already in $out.
pack
if [ -e "$out" ]; then
	diff -r "$packed" "$out" >/dev/null || fail "$out differs from a fresh pack; delete it to pack again"
	echo "npm: $out matches a fresh pack"
else
	mv "$packed" "$out"
fi
check
smoke
ls -1 "$out"

# available waits until the registry serves the version $1 of the package: npm processes a
# publish for a few minutes before installs can fetch it.
available() {
	local tries=0
	until npm view "$pkg@$1" version >/dev/null 2>&1; do
		tries=$((tries + 1))
		[ "$tries" -le 60 ] || fail "$pkg@$1 is still not available after 30 minutes"
		sleep 30
	done
}

if [ "$mode" = --publish ]; then
	remote=$(git ls-remote origin "refs/tags/$tag" "refs/tags/$tag^{}" | awk '{ c = $1 } END { print c }')
	[ "$remote" = "$commit" ] || fail "origin's tag $tag does not name $commit; push the tag first"
	# GitHub Actions publishes through trusted publishing (OIDC): there is no login to check.
	if [ -z "${ACTIONS_ID_TOKEN_REQUEST_URL:-}" ]; then
		npm whoami >/dev/null || fail "not logged in to npm; run npm login"
	fi
	# The platform versions first: the main version must never depend on a version that does not
	# exist yet. Each has its own dist-tag, so that latest stays on the launcher.
	for file in "${archives[@]}"; do
		platform=$(npm_platform "$file")
		publish "$version-$platform" "$(tgz_of "$platform")" "$platform"
	done
	# Until every platform version is served, an install of the main version would leave
	# out the binary of that platform.
	for file in "${archives[@]}"; do
		available "$version-$(npm_platform "$file")"
	done
	publish "$version" "$main_tgz" latest
fi

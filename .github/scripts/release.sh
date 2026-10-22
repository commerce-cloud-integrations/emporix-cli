#!/usr/bin/env bash
# Builds epx release archives. Run inside a checkout; .github/workflows/release.yml runs
# it for every pushed vMAJOR.MINOR.PATCH tag.
#
#   .github/scripts/release.sh             tagged release: HEAD must carry a canonical vMAJOR.MINOR.PATCH
#                                  tag and the tree must be clean (no modified or untracked files)
#   .github/scripts/release.sh --snapshot  any commit, for testing
#   .github/scripts/release.sh --publish   tagged release, then `gh release create` with the
#                                  archives and the installers (install/install.sh,
#                                  install/install.ps1, served as
#                                  releases/latest/download/install.{sh,ps1}), once the remote
#                                  tag is verified to name the built commit
#
# Binaries are static (CGO_ENABLED=0), built with -trimpath and VCS stamping, so
# `epx version` reports the tag without ldflags. Every binary is checked with
# `go version -m` (module version, revision, unmodified tree) before packaging.
# Output: dist/<version>/epx_<version>_<os>_<arch>.{tar.gz,zip} and a checksums file.
set -euo pipefail

fail() { echo "release: $*" >&2; exit 1; }

mode=${1:-release}
case $mode in release | --snapshot | --publish) ;; *) echo "usage: $0 [--snapshot|--publish]" >&2; exit 2 ;; esac
cd "$(git rev-parse --show-toplevel)"
commit=$(git rev-parse HEAD)

if [ "$mode" = --snapshot ]; then
	version=snapshot-$(git rev-parse --short HEAD)
else
	status=$(git status --porcelain) || fail "git status failed"
	[ -z "$status" ] || fail "the tree has modified or untracked files"
	version=$(git describe --exact-match --tags HEAD 2>/dev/null) || fail "HEAD has no tag"
	[[ $version =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] ||
		fail "tag $version is not a canonical vMAJOR.MINOR.PATCH"
fi

out=dist/$version
[ ! -e "$out" ] || fail "$out already exists"
mkdir -p "$out"
stage=$(mktemp -d)
trap 'rm -r "$stage"' EXIT

# check_binary verifies the build info a binary carries.
check_binary() {
	local info
	info=$(go version -m "$1") || fail "cannot read build info of $1"
	grep -q "vcs.revision=$commit" <<<"$info" || fail "$1 was not built from $commit"
	if [ "$mode" != --snapshot ]; then
		grep -q "vcs.modified=false" <<<"$info" || fail "$1 was built from a modified tree"
		awk '$1 == "mod" { print $3 }' <<<"$info" | grep -qx "$version" || fail "$1 does not report $version"
	fi
}

# macOS tar (bsdtar) stores extended attributes such as com.apple.provenance, whose
# headers GNU tar warns about when extracting; GNU tar stores none by default.
tar_flags=()
if tar --version | grep -q bsdtar; then
	tar_flags=(--no-xattrs --no-mac-metadata)
fi

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
	os=${target%/*}
	arch=${target#*/}
	name=epx_${version}_${os}_${arch}
	exe=epx
	[ "$os" = windows ] && exe=epx.exe
	mkdir "$stage/$name"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -buildvcs=true -trimpath -ldflags='-s -w' \
		-o "$stage/$name/$exe" ./cmd/epx
	check_binary "$stage/$name/$exe"
	cp LICENSE README.md "$stage/$name/"
	if [ "$os" = windows ]; then
		(cd "$stage" && zip -qrX "$OLDPWD/$out/$name.zip" "$name")
	else
		COPYFILE_DISABLE=1 tar "${tar_flags[@]}" -C "$stage" -czf "$out/$name.tar.gz" "$name"
	fi
done
(cd "$out" && shasum -a 256 epx_* >"epx_${version}_checksums.txt")

host=$stage/epx_${version}_$(go env GOHOSTOS)_$(go env GOHOSTARCH)/epx
reported=$("$host" version) || fail "the host binary does not run"
echo "release: built $out ($reported)"
ls -1 "$out"

if [ "$mode" = --publish ]; then
	remote=$(git ls-remote origin "refs/tags/$version" "refs/tags/$version^{}" | awk '{ c = $1 } END { print c }')
	[ "$remote" = "$commit" ] || fail "origin's tag $version does not name $commit; push the tag first"
	gh release create "$version" "$out"/* install/install.sh install/install.ps1 \
		--verify-tag --title "$version" --generate-notes
fi

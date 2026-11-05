#!/usr/bin/env bash
# Writes the Homebrew formula (epx.rb) and the Scoop manifest (epx.json) of an epx release
# from the checksums file of its archives. .github/scripts/release.sh runs it for every
# tagged release, so both are assets of the GitHub release, served as
# releases/latest/download/epx.{rb,json}: the Homebrew tap
# (commerce-cloud-integrations/homebrew-tap, brew install commerce-cloud-integrations/tap/epx)
# and the Scoop bucket (commerce-cloud-integrations/scoop-bucket) copy them from there with
# workflows of their own, so this repository holds no credential for them.
#
#   .github/scripts/packages.sh DIST   write DIST/epx.rb and DIST/epx.json; DIST is
#                                      dist/<tag>, holding at least its checksums file
set -euo pipefail

fail() { echo "packages: $*" >&2; exit 1; }
usage() { echo "usage: $0 DIST" >&2; exit 2; }

[[ $# -eq 1 ]] || usage
dist=${1%/}
[[ -d "$dist" ]] || fail "$dist does not exist"

tag=$(basename "$dist")
[[ $tag =~ ^v((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*))$ ]] ||
	fail "$dist is not dist/vMAJOR.MINOR.PATCH"
version=${BASH_REMATCH[1]}

owner=commerce-cloud-integrations
download=https://github.com/$owner/emporix-cli/releases/download/$tag
homepage=https://$owner.github.io/emporix-cli/
summary="Command-line client for the Emporix Commerce APIs"

sums=$dist/epx_${tag}_checksums.txt
[[ -f "$sums" ]] || fail "$dist has no $(basename "$sums")"

# sha prints the SHA-256 of the archive of a platform (os_arch), as the checksums file lists it.
sha() {
	local file=epx_${tag}_$1 sum
	sum=$(awk -v f="$file" '$2 == f { print $1 }' "$sums")
	[[ $(grep -c "  $file\$" "$sums") -eq 1 ]] || fail "$(basename "$sums") lists $file other than once"
	[[ $sum =~ ^[0-9a-f]{64}$ ]] || fail "$(basename "$sums") lists no SHA-256 for $file"
	echo "$sum"
}

# Every checksum is read and checked here, before anything is written: a failure inside
# the command substitutions below would not stop the script.
darwin_arm64=$(sha darwin_arm64.tar.gz)
darwin_amd64=$(sha darwin_amd64.tar.gz)
linux_arm64=$(sha linux_arm64.tar.gz)
linux_amd64=$(sha linux_amd64.tar.gz)
windows_amd64=$(sha windows_amd64.zip)
windows_arm64=$(sha windows_arm64.zip)

# platform prints the url and sha256 lines of the formula for the tarball of os_arch with
# the given checksum.
platform() {
	local target=$1 sum=$2
	printf '      url "%s/epx_%s_%s.tar.gz"\n      sha256 "%s"\n' "$download" "$tag" "$target" "$sum"
}

formula() {
	cat <<EOF
# Written by the release of epx $version (https://github.com/$owner/emporix-cli); do not edit.
class Epx < Formula
  desc "$summary"
  homepage "$homepage"
  license "Apache-2.0"

  on_macos do
    on_arm do
$(platform darwin_arm64 "$darwin_arm64")
    end
    on_intel do
$(platform darwin_amd64 "$darwin_amd64")
    end
  end

  on_linux do
    on_arm do
$(platform linux_arm64 "$linux_arm64")
    end
    on_intel do
$(platform linux_amd64 "$linux_amd64")
    end
  end

  def install
    bin.install "epx"
    man1.install Dir["man/*.1"]
    generate_completions_from_executable(bin/"epx", "completion")
  end

  test do
    assert_match "epx v#{version} ", shell_output("#{bin}/epx version")
  end
end
EOF
}

manifest() {
	jq -n --arg version "$version" --arg summary "$summary" --arg homepage "$homepage" \
		--arg download "$download" --arg tag "$tag" \
		--arg amd64 "$windows_amd64" --arg arm64 "$windows_arm64" '
		def arch($goarch; $hash): {
			url: "\($download)/epx_\($tag)_windows_\($goarch).zip",
			hash: $hash,
			extract_dir: "epx_\($tag)_windows_\($goarch)"
		};
		{
			version: $version,
			description: $summary,
			homepage: $homepage,
			license: "Apache-2.0",
			architecture: {"64bit": arch("amd64"; $amd64), arm64: arch("arm64"; $arm64)},
			bin: "epx.exe"
		}'
}

formula >"$dist/epx.rb"
manifest >"$dist/epx.json"
echo "packages: wrote $dist/epx.rb and $dist/epx.json (epx $version)"

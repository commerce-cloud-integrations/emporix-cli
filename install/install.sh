#!/bin/sh
# Installs epx, a command-line client for the Emporix Commerce APIs, on macOS or Linux:
#
#   curl -fsSL https://github.com/commerce-cloud-integrations/emporix-cli/releases/latest/download/install.sh | sh
#
# It downloads the release archive of this machine from GitHub, checks it against the
# release's checksums file and puts the epx binary into ~/.local/bin, without sudo.
#
# Settings (environment variables; when piping into sh, put them after the pipe:
# `| EPX_VERSION=v0.1.2 sh`):
#   EPX_VERSION         release tag to install, such as v0.1.2 (default: the latest release)
#   EPX_INSTALL_DIR     directory for the binary (default: ~/.local/bin)
#   EPX_NO_MODIFY_PATH  set to 1 to leave the shell profile alone when that directory is
#                       not on PATH
#
# Everything runs inside main, so a download cut off midway runs nothing.

set -eu

repo=https://github.com/commerce-cloud-integrations/emporix-cli
# curl's --proto and --proto-redir: https only, for the first request and every redirect.
https_only='=https'

say() { printf 'epx installer: %s\n' "$*"; }
fail() {
	printf 'epx installer: %s\n' "$*" >&2
	exit 1
}

# fetch writes the body of an https URL into a file: fetch URL FILE.
fetch() {
	fetch_url=$1
	fetch_file=$2
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL --proto "$https_only" --proto-redir "$https_only" --tlsv1.2 -o "$fetch_file" "$fetch_url"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$fetch_file" "$fetch_url"
	else
		fail "neither curl nor wget is installed"
	fi
}

# latest_tag prints the tag of the latest release: with curl from the redirect of
# /releases/latest (no API rate limit), with wget (BusyBox too) from the GitHub API, read
# into $tmp.
latest_tag() {
	if command -v curl >/dev/null 2>&1; then
		url=$(curl -fsSLI --proto "$https_only" --proto-redir "$https_only" --tlsv1.2 -o /dev/null -w '%{url_effective}' \
			"$repo/releases/latest") || url=
		tag=${url##*/tag/}
	elif fetch "https://api.github.com/repos/${repo#https://github.com/}/releases/latest" "$tmp/latest.json"; then
		tag=$(sed -n 's/^ *"tag_name": *"\([^"]*\)".*/\1/p' "$tmp/latest.json")
	else
		tag=
	fi
	case $tag in v[0-9]*) printf '%s\n' "$tag" ;; *) fail "cannot find the latest release at $repo/releases" ;; esac
}

# target prints <os>_<arch> of this machine as the release archives name it.
target() {
	case $(uname -s) in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) fail "$(uname -s) is not supported by this script; on Windows use install.ps1, or download an archive from $repo/releases" ;;
	esac
	case $(uname -m) in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) fail "no epx binary exists for the CPU $(uname -m)" ;;
	esac
	# A shell running under Rosetta reports x86_64 on an Apple silicon Mac.
	if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
		[ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" = 1 ]; then
		arch=arm64
	fi
	printf '%s_%s\n' "$os" "$arch"
}

# sha256 prints the SHA-256 of a file: sha256 FILE.
sha256() {
	sum_file=$1
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$sum_file" | awk '{ print $1 }'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$sum_file" | awk '{ print $1 }'
	elif command -v openssl >/dev/null 2>&1; then
		openssl dgst -sha256 "$sum_file" | awk '{ print $NF }'
	else
		fail "no SHA-256 tool found (sha256sum, shasum or openssl)"
	fi
}

# profile prints the startup file of the user's shell that should set PATH.
profile() {
	case ${SHELL:-} in
	*/zsh) if [ "$(uname -s)" = Darwin ]; then echo "$HOME/.zprofile"; else echo "${ZDOTDIR:-$HOME}/.zshrc"; fi ;;
	*/bash) if [ "$(uname -s)" = Darwin ]; then echo "$HOME/.bash_profile"; else echo "$HOME/.bashrc"; fi ;;
	*) echo "$HOME/.profile" ;;
	esac
}

# add_to_path makes sure that new shells find epx in a directory: add_to_path DIR.
add_to_path() {
	bin_dir=$1
	case ":$PATH:" in *":$bin_dir:"*) return ;; esac
	line="export PATH=\"$bin_dir:\$PATH\""
	if [ "${EPX_NO_MODIFY_PATH:-}" = 1 ]; then
		say "$bin_dir is not on your PATH; add it with: $line"
		return
	fi
	file=$(profile)
	if ! grep -qF '# >>> epx installer >>>' "$file" 2>/dev/null; then
		printf '\n# >>> epx installer >>>\n%s\n# <<< epx installer <<<\n' "$line" >>"$file"
		say "added $bin_dir to PATH in $file"
	fi
	say "to use epx in this terminal, run: $line"
}

main() {
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT
	trap 'exit 1' HUP INT TERM

	version=${EPX_VERSION:-}
	if [ -z "$version" ]; then
		version=$(latest_tag)
	fi
	case $version in v*) ;; *) version=v$version ;; esac
	printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' ||
		fail "EPX_VERSION is not a release tag such as v0.1.2"
	dir=${EPX_INSTALL_DIR:-$HOME/.local/bin}
	name=epx_${version}_$(target)
	base=$repo/releases/download/$version

	say "downloading epx $version ($name)"
	fetch "$base/$name.tar.gz" "$tmp/$name.tar.gz" || fail "cannot download $base/$name.tar.gz"
	fetch "$base/epx_${version}_checksums.txt" "$tmp/checksums.txt" || fail "cannot download the checksums of $version"
	want=$(awk -v f="$name.tar.gz" '$2 == f { print $1 }' "$tmp/checksums.txt")
	[ -n "$want" ] || fail "the checksums of $version do not list $name.tar.gz"
	[ "$(sha256 "$tmp/$name.tar.gz")" = "$want" ] || fail "$name.tar.gz does not match its checksum"

	tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
	mkdir -p "$dir"
	# Copied next to the target, then renamed: an epx that is running keeps working.
	cp "$tmp/$name/epx" "$dir/.epx.$$"
	chmod 755 "$dir/.epx.$$"
	mv -f "$dir/.epx.$$" "$dir/epx"
	reported=$("$dir/epx" version) || fail "the installed $dir/epx does not run"
	case $reported in "epx $version "*) ;; *) fail "$dir/epx reports '$reported', not $version" ;; esac
	say "installed $reported to $dir/epx"

	add_to_path "$dir"
	found=$(command -v epx 2>/dev/null || true)
	if [ -n "$found" ] && [ "$found" != "$dir/epx" ]; then
		say "note: another epx comes first on your PATH: $found"
	fi
}

main "$@"

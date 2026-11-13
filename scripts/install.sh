#!/bin/sh
# Install the latest blerg release binary into /usr/local/bin, or ~/.local/bin if that is not writable.
set -eu

repo="iandesj/Blerg"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
linux | darwin) ;;
*)
	echo "install.sh supports linux and darwin; download a Windows zip from GitHub releases" >&2
	exit 1
	;;
esac

case "$arch" in
x86_64 | amd64) arch="amd64" ;;
aarch64 | arm64) arch="arm64" ;;
*)
	echo "unsupported architecture: $arch" >&2
	exit 1
	;;
esac

archive="blerg_${os}_${arch}.tar.gz"
url="https://github.com/${repo}/releases/latest/download/${archive}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$url" | tar -xz -C "$tmp"
[ -x "$tmp/blerg" ] || chmod +x "$tmp/blerg"

dest="/usr/local/bin"
if [ ! -w "$dest" ]; then
	dest="${HOME}/.local/bin"
	mkdir -p "$dest"
fi

if command -v install >/dev/null 2>&1; then
	install "$tmp/blerg" "${dest}/blerg"
else
	cp "$tmp/blerg" "${dest}/blerg"
	chmod +x "${dest}/blerg"
fi

echo "installed ${dest}/blerg"
echo "ensure ${dest} is on PATH"

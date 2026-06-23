#!/usr/bin/env bash
#
# setup-libnode-linux.sh
#
# Build and install the shared "libnode" that the gojr cgo bridge
# (cmd/gojr/gojr.go) links against on Linux.
#
# Why this exists:
#   gojr embeds the Node.js / V8 C++ embedder API, so it needs Node headers
#   *and* a matching shared library (libnode.so.<ABI>). Node does not publish
#   official shared-library builds, and distro packages (e.g. Ubuntu's
#   libnode-dev) lag several major versions behind. So we build the exact
#   pinned version from source as a shared library and install it into PREFIX
#   (default /usr/local), which the gojr cgo directives reference:
#     #cgo linux CXXFLAGS: -std=c++20 -I/usr/local/include/node -DNODE_SHARED_MODE
#     #cgo linux LDFLAGS:  -L/usr/local/lib -lnode -Wl,-rpath,/usr/local/lib
#
#   NODE_VERSION must match the headers gojr compiles against. Node's
#   `make install` also drops node/npm into PREFIX/bin as a side effect; that
#   is harmless (your PATH/nvm node still wins) and there is no narrower
#   "install just the library" target upstream.
#
# No sudo: PREFIX is expected to be user-writable. If it is not, fix the
# directory's ownership/permissions instead of running this under sudo.
#
# Usage:
#   scripts/setup-libnode-linux.sh
#   NODE_VERSION=23.1.0 PREFIX=/usr/local FORCE=1 scripts/setup-libnode-linux.sh
#
set -euo pipefail

NODE_VERSION="${NODE_VERSION:-23.1.0}"
PREFIX="${PREFIX:-/usr/local}"
FORCE="${FORCE:-0}"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "${script_dir}/.." && pwd)"
build_dir="${project_root}/_build_node_js_temp"

src_name="node-v${NODE_VERSION}"
tarball="${src_name}.tar.xz"
base_url="https://nodejs.org/dist/v${NODE_VERSION}"
jobs="$(nproc 2>/dev/null || echo 4)"

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

# Version (maj.min.patch) of the headers currently installed under PREFIX, if any.
installed_version() {
	local h="${PREFIX}/include/node/node_version.h"
	[ -f "${h}" ] || return 1
	local maj min pat
	maj="$(awk '/#define NODE_MAJOR_VERSION/{print $3}' "${h}")"
	min="$(awk '/#define NODE_MINOR_VERSION/{print $3}' "${h}")"
	pat="$(awk '/#define NODE_PATCH_VERSION/{print $3}' "${h}")"
	printf '%s.%s.%s' "${maj}" "${min}" "${pat}"
}

# Skip the (expensive) build if the right version is already in place.
if [ "${FORCE}" != "1" ] \
	&& [ -e "${PREFIX}/lib/libnode.so" ] \
	&& [ "$(installed_version 2>/dev/null || true)" = "${NODE_VERSION}" ]; then
	log "libnode ${NODE_VERSION} already installed in ${PREFIX} (set FORCE=1 to rebuild)"
	exit 0
fi

# Prerequisites.
for tool in curl tar xz make g++ cc python3; do
	command -v "${tool}" >/dev/null 2>&1 || die "missing required tool: ${tool}"
done

mkdir -p "${PREFIX}"
[ -w "${PREFIX}" ] || die "${PREFIX} is not writable; fix its permissions (do not use sudo)"

mkdir -p "${build_dir}"
cd "${build_dir}"

# Download the source tarball (idempotent).
if [ ! -f "${tarball}" ]; then
	log "downloading ${base_url}/${tarball}"
	curl -fL --retry 3 -o "${tarball}.partial" "${base_url}/${tarball}"
	mv "${tarball}.partial" "${tarball}"
fi

# Verify the checksum when sha256sum is available (best effort).
if command -v sha256sum >/dev/null 2>&1; then
	log "verifying sha256"
	curl -fsSL -o SHASUMS256.txt "${base_url}/SHASUMS256.txt"
	grep " ${tarball}\$" SHASUMS256.txt | sha256sum -c - \
		|| die "checksum verification failed for ${tarball}"
fi

# Extract (idempotent).
if [ ! -d "${src_name}" ]; then
	log "extracting ${tarball}"
	tar xf "${tarball}"
fi

cd "${src_name}"

log "configure --shared --prefix=${PREFIX}"
./configure --shared --prefix="${PREFIX}"

log "make -j${jobs} (this takes a while)"
make -j"${jobs}"

log "make install (no sudo)"
make install

# Node installs only the versioned libnode.so.<ABI>; create the unversioned
# dev symlink that `-lnode` needs at link time.
cd "${PREFIX}/lib"
soname="$(ls -1 libnode.so.* 2>/dev/null | head -n1)"
[ -n "${soname}" ] || die "libnode.so.* not found in ${PREFIX}/lib after install"
ln -sf "${soname}" libnode.so
log "linked ${PREFIX}/lib/libnode.so -> ${soname}"

log "done. Now build gojr with:  make -C ${project_root} install"
log "the source tree in ${build_dir} can be removed to reclaim disk."

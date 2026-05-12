#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ABC_DIR="$SCRIPT_DIR/abc"

: "${EMSDK:=$HOME/go/src/github.com/emscripten-core/emsdk}"
: "${EM_CACHE:=/tmp/ivy-emscripten-cache}"
: "${ABC_WASM_JOBS:=}"
: "${ABC_WASM_LOG:=/private/tmp/abc-wasm-build-$(date +%Y%m%d-%H%M%S).log}"

if [[ -f "$EMSDK/emsdk_env.sh" ]]; then
  EMSDK_QUIET=1 source "$EMSDK/emsdk_env.sh"
elif ! command -v emcc >/dev/null 2>&1; then
  echo "error: emcc was not found, and EMSDK/emsdk_env.sh does not exist at $EMSDK/emsdk_env.sh" >&2
  exit 1
fi

if [[ -z "$ABC_WASM_JOBS" ]]; then
  ABC_WASM_JOBS="$(getconf _NPROCESSORS_ONLN 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)"
fi

mkdir -p "$EM_CACHE"

: "${EMCC:=$(command -v emcc)}"
: "${EMXX:=$(command -v em++)}"
: "${EMAR:=$(command -v emar)}"

cd "$ABC_DIR"

echo "Build log: $ABC_WASM_LOG"
set +e
EM_CACHE="$EM_CACHE" \
EMCC="$EMCC" \
EMXX="$EMXX" \
EMAR="$EMAR" \
emmake make -f Makefile.emscripten -j"$ABC_WASM_JOBS" "$@" 2>&1 | tee "$ABC_WASM_LOG"
make_status=${PIPESTATUS[0]}
set -e

if [[ "$make_status" -ne 0 ]]; then
  echo "Build log: $ABC_WASM_LOG"
  exit "$make_status"
fi

if [[ -f "$ABC_DIR/build/wasm/abc.js" && -f "$ABC_DIR/build/wasm/abc.wasm" ]]; then
  echo "Built $ABC_DIR/build/wasm/abc.js and $ABC_DIR/build/wasm/abc.wasm"
else
  echo "Finished make -f Makefile.emscripten $*"
fi
echo "Build log: $ABC_WASM_LOG"

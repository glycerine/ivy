#!/usr/bin/env bash
set -euo pipefail

# Reproduce the Z3 wasm crash seen in the vendored Z3 unit-test runner.
#
# The failing test is Z3's own src/test model_retrieval module compiled by
# Emscripten. It does not use Go Ivy wasm or the goivy smtZ3Imports.js bridge.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
z3_dir="${script_dir}/z3"
build_dir="${z3_dir}/wasm.build"
runner="${build_dir}/test-z3"
wasm="${build_dir}/test-z3.wasm"
log="${Z3CRASHER_LOG:-${script_dir}/z3crasher.log}"

venv_bin="${GOIVY_VENV_BIN:-/Users/jaten/ivy/pyivy/goivy-venv/bin}"
if [[ -d "${venv_bin}" ]]; then
  export PATH="${venv_bin}:${PATH}"
fi

emsdk_env="${EMSDK_ENV:-/Users/jaten/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh}"
em_cache="${EM_CACHE:-/private/tmp/ivy-emscripten-cache}"

if ! command -v node >/dev/null 2>&1; then
  echo "node not found on PATH" >&2
  exit 1
fi

if [[ ! -f "${runner}" || ! -f "${wasm}" || "${Z3CRASHER_REBUILD:-0}" == "1" ]]; then
  if [[ ! -f "${emsdk_env}" ]]; then
    echo "emsdk environment not found: ${emsdk_env}" >&2
    exit 1
  fi

  echo "[z3crasher] building ${runner} and ${wasm}"
  (
    cd "${build_dir}"
    # The generated old Z3 Makefile compiles the test objects correctly but may
    # fail the first link without wasm EH flags. Keep going so the explicit
    # relink below can finish the command-style JS/wasm test runner.
    set +e
    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" make test-z3 -j${Z3CRASHER_JOBS:-8}"
    first_status=$?
    set -e
    if [[ ${first_status} -ne 0 ]]; then
      echo "[z3crasher] initial make exited ${first_status}; relinking with wasm EH flags"
    fi
    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" make test-z3 -j1 LINK_FLAGS=\"-fwasm-exceptions -sALLOW_MEMORY_GROWTH=1 -sALLOW_TABLE_GROWTH=1 -sEXIT_RUNTIME=1 -sSTACK_SIZE=20MB -sINITIAL_MEMORY=256MB\" LINK_EXTRA_FLAGS=\"\""
  )
fi

echo "[z3crasher] runner: ${runner}"
echo "[z3crasher] wasm  : ${wasm}"
echo "[z3crasher] log   : ${log}"
echo "[z3crasher] command: node ${runner} model_retrieval"

set +e
node "${runner}" model_retrieval 2>&1 | tee "${log}"
node_status=${PIPESTATUS[0]}
set -e

if grep -q "RuntimeError: memory access out of bounds" "${log}"; then
  echo "[z3crasher] reproduced: RuntimeError: memory access out of bounds"
  exit 0
fi

echo "[z3crasher] did not reproduce; node exit status was ${node_status}" >&2
exit "${node_status:-1}"

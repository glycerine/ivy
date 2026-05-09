#!/usr/bin/env bash
set -euo pipefail

# Reproduce the Z3 wasm crash seen in the vendored Z3 unit-test runner.
#
# The failing test is Z3's own src/test model_retrieval module compiled by
# Emscripten. It does not use Go Ivy wasm or the goivy smtZ3Imports.js bridge.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
z3_dir="${script_dir}/z3"
build_dir="${Z3CRASHER_BUILD_DIR:-${z3_dir}/wasm.build}"
runner="${build_dir}/test-z3-model-retrieval"
wasm="${build_dir}/test-z3-model-retrieval.wasm"
log="${Z3CRASHER_LOG:-${script_dir}/z3crasher.log}"
focused_main="${build_dir}/test/model_retrieval_main.cpp"
focused_main_obj="${build_dir}/test/model_retrieval_main.o"
throw_wrap="${build_dir}/test/throw_wrap.cpp"
throw_wrap_obj="${build_dir}/test/throw_wrap.o"
link_stamp="${runner}.link-flags"

venv_bin="${GOIVY_VENV_BIN:-/Users/jaten/ivy/pyivy/goivy-venv/bin}"
if [[ -d "${venv_bin}" ]]; then
  export PATH="${venv_bin}:${PATH}"
fi

emsdk_env="${EMSDK_ENV:-/Users/jaten/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh}"
em_cache="${EM_CACHE:-/private/tmp/ivy-emscripten-cache}"
z3_stack_size="${Z3CRASHER_STACK_SIZE:-83886080}"
debug_flags="${Z3CRASHER_DEBUG_FLAGS:--g2 --profiling-funcs}"
default_link_flags="-std=c++17 -fwasm-exceptions ${debug_flags} -Wl,--wrap=__cxa_throw -sALLOW_MEMORY_GROWTH=1 -sALLOW_TABLE_GROWTH=1 -sEXIT_RUNTIME=1 -sSTACK_SIZE=${z3_stack_size} -sTOTAL_STACK=${z3_stack_size} -sINITIAL_MEMORY=256MB"
link_flags="${Z3CRASHER_LINK_FLAGS:-${default_link_flags}}"

if ! command -v node >/dev/null 2>&1; then
  echo "node not found on PATH" >&2
  exit 1
fi

needs_rebuild=0
if [[ ! -f "${runner}" || ! -f "${wasm}" || ! -f "${link_stamp}" || "${Z3CRASHER_REBUILD:-0}" == "1" ]]; then
  needs_rebuild=1
elif [[ "$(cat "${link_stamp}")" != "${link_flags}" ]]; then
  needs_rebuild=1
fi

if [[ "${needs_rebuild}" == "1" ]]; then
  if [[ ! -f "${emsdk_env}" ]]; then
    echo "emsdk environment not found: ${emsdk_env}" >&2
    exit 1
  fi

  echo "[z3crasher] building focused model_retrieval runner"
  (
    cd "${build_dir}"
    mkdir -p test
    cat > "${focused_main}" <<'CPP'
#include <exception>
#include <iostream>

void tst_model_retrieval();

int main() {
    try {
        tst_model_retrieval();
        return 0;
    }
    catch (std::exception const& ex) {
        std::cerr << "unhandled std::exception: " << ex.what() << "\n";
        return 2;
    }
    catch (...) {
        std::cerr << "unhandled non-std exception\n";
        return 3;
    }
}
CPP
    cat > "${throw_wrap}" <<'CPP'
#include <stdio.h>

extern "C" __attribute__((noreturn)) void __real___cxa_throw(void* thrown_exception, void* tinfo, void (*dest)(void*));

extern "C" __attribute__((noreturn)) void __wrap___cxa_throw(void* thrown_exception, void* tinfo, void (*dest)(void*)) {
    fprintf(stderr, "__cxa_throw called\n");
    fflush(stderr);
    __real___cxa_throw(thrown_exception, tinfo, dest);
    __builtin_unreachable();
}
CPP

    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" make test/model_retrieval.o libz3.a -j${Z3CRASHER_JOBS:-8}"
    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" em++ -std=c++17 -fwasm-exceptions -D_NO_OMP_ -D_MP_INTERNAL -I../src -I../src/test -I../src/util -c \"${focused_main}\" -o \"${focused_main_obj}\""
    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" em++ -std=c++17 -fwasm-exceptions -c \"${throw_wrap}\" -o \"${throw_wrap_obj}\""
    EMSDK_QUIET=1 bash -lc \
      "source \"${emsdk_env}\" && EM_CACHE=\"${em_cache}\" em++ ${link_flags} -o \"${runner}\" \"${focused_main_obj}\" \"${throw_wrap_obj}\" test/model_retrieval.o libz3.a"
    printf '%s\n' "${link_flags}" > "${link_stamp}"
  )
fi

echo "[z3crasher] runner: ${runner}"
echo "[z3crasher] wasm  : ${wasm}"
echo "[z3crasher] log   : ${log}"
echo "[z3crasher] command: node ${runner}"

set +e
node "${runner}" 2>&1 | tee "${log}"
node_status=${PIPESTATUS[0]}
set -e

if grep -q "RuntimeError: memory access out of bounds" "${log}"; then
  echo "[z3crasher] reproduced: RuntimeError: memory access out of bounds"
  exit 0
fi

echo "[z3crasher] did not reproduce; node exit status was ${node_status}" >&2
exit "${node_status:-1}"

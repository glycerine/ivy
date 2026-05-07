#!/bin/bash
set -e

EXPORTED_FUNCS=$(cat src/api/js/exported-functions.json)

emcc \
  -std=c++17 \
  -D_NO_OMP_ \
  -D_MP_INTERNAL \
  -Isrc/api \
  -Lbuild \
  build/libz3.a \
  -sEXPORTED_FUNCTIONS="$EXPORTED_FUNCS" \
  -sEXPORTED_RUNTIME_METHODS='["ccall","cwrap","UTF8ToString","stringToUTF8"]' \
  -sALLOW_MEMORY_GROWTH=1 \
  -sMODULARIZE=1 \
  -sEXPORT_NAME="initZ3" \
  -o build/z3-api.js

echo "Built build/z3-api.js and build/z3-api.wasm"

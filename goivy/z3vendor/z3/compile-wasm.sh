#!/bin/bash

# must use older stable emcc and not homebrew default 4.0.7. we had better luck with 3.1.50
emcc --version | grep 3.1.50

# mk_make.py infers _AMD64_ from the native host. Emscripten targets wasm32, so
# wasm builds must explicitly undefine it without changing native builds.
wasm_cflags="-Wno-deprecated-declarations -fwasm-exceptions -U_AMD64_"

## can try this to address "FP math:        UNKNOWN"
## FPMATH_ENABLED=False AR=emar CXX=em++ CC=emcc CXXFLAGS="${wasm_cflags}" CFLAGS="${wasm_cflags}" python3 scripts/mk_make.py --staticlib  --nofp

rm -rf wasm.build

AR=emar CXX=em++ CC=emcc CXXFLAGS="${wasm_cflags}" CFLAGS="${wasm_cflags}" python3 scripts/mk_make.py --staticlib --build=wasm.build

cd wasm.build
#emmake make -j8 libz3.a
EMSDK_QUIET=1 source "$HOME/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh" && EM_CACHE="/tmp/ivy-emscripten-cache" emmake make -j8 libz3.a

cd .. 
#mv build wasm.build
#ln -s wasm.build build

EMSDK_QUIET=1 source "$HOME/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh" && EM_CACHE="/tmp/ivy-emscripten-cache" EMCC="$HOME/go/src/github.com/emscripten-core/emsdk/upstream/emscripten/emcc" ./build-wasm-api.sh

cp -p wasm.build/z3-api.wasm wasm.build/z3-api.pre-exnref.wasm

/usr/local/bin/wasm-opt --translate-to-exnref \
		--enable-exception-handling \
		--enable-reference-types \
		--enable-bulk-memory \
		--enable-bulk-memory-opt \
		--enable-multivalue \
		--enable-mutable-globals \
		--enable-sign-ext \
		--enable-nontrapping-float-to-int \
		-o wasm.build/z3-api.exnref.wasm wasm.build/z3-api.pre-exnref.wasm

cp -p wasm.build/z3-api.exnref.wasm wasm.build/z3-471-api.wasm

# save them above in case we rebuild native and wipe them out by accident.
for i in z3-api.js z3-api.wasm z3.wasm libz3.a libz3.wasm libz3.dylib; do 
  cp -p wasm.build/$i ../wasm_lib/; 
done

# make the names match what build-wasm-api.sh copies into ../../webui/static/
mv ../wasm_lib/z3-api.js  ../wasm_lib/z3-471-api.js

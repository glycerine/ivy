#!/bin/bash

# must use older stable emcc and not homebrew default 4.0.7. we had better luck with 3.1.50
emcc --version | grep 3.1.50

## can try this to address "FP math:        UNKNOWN"
## FPMATH_ENABLED=False AR=emar CXX=em++ CC=emcc CXXFLAGS="-Wno-deprecated-declarations" CFLAGS="-Wno-deprecated-declarations" python3 scripts/mk_make.py --staticlib  --nofp

AR=emar CXX=em++ CC=emcc CXXFLAGS="-Wno-deprecated-declarations" CFLAGS="-Wno-deprecated-declarations" python3 scripts/mk_make.py --staticlib

cd build
emmake make -j8

cd .. 
mv build wasm.build
ln -s wasm.build build

# save them above in case we rebuild native and wipe them out by accident.
for i in z3-api.js z3-api.wasm z3.wasm libz3.a libz3.wasm libz3.dylib; do cp -p wasm.build/$i ../wasm_lib/; done


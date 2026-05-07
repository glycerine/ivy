#!/bin/bash

# must use older stable emcc and not homebrew default 4.0.7. we had better luck with 3.1.50
emcc --version | grep 3.1.50

CXX=em++ CC=emcc CXXFLAGS="-Wno-deprecated-declarations" CFLAGS="-Wno-deprecated-declarations" python3 scripts/mk_make.py --staticlib

cd build
emmake make -j8

#cp -p 


#!/bin/bash

# 1. Wipe the CMake build directory
rm -rf build

# 2. Run the legacy Python generator, explicitly requesting a static library.
# (We pass the "silence the sprintf blabber" flags via standard environment variables here)
CXXFLAGS="-Wno-deprecated-declarations" CFLAGS="-Wno-deprecated-declarations" python3 scripts/mk_make.py --staticlib

# 3. The script creates a new 'build' directory. Go into it.
cd build

# 4. Compile using standard Make (not CMake)
make -j8

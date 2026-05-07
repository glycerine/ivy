wasm build of 4.7.1 + ken mcm patches into wasm

https://claude.ai/chat/b6273985-b135-4ac1-b8da-2c3467074c65

Perfect! The modern Z3 has a TypeScript parser that extracts all Z3 API functions from the headers. Here's the plan:

Copy the parsing infrastructure to your v4.7.1 build
Run it against v4.7.1's headers
Generate the WASM build with the exported functions

Let's do this step by step:
bashcd ~/go/src/github.com/cpitclaudel/z3.wasm/build/z3

# Create the js API directory structure
mkdir -p src/api/js/scripts

# Copy the parsing and build scripts from modern Z3
cp ~/go/src/github.com/Z3Prover/z3/src/api/js/scripts/parse-api.ts src/api/js/scripts/
cp ~/go/src/github.com/Z3Prover/z3/src/api/js/scripts/build-wasm.ts src/api/js/scripts/
cp ~/go/src/github.com/Z3Prover/z3/src/api/js/scripts/async-fns.ts src/api/js/scripts/ 2>/dev/null || echo "async-fns.ts might not exist"
cp ~/go/src/github.com/Z3Prover/z3/src/api/js/scripts/make-cc-wrapper.ts src/api/js/scripts/ 2>/dev/null || echo "make-cc-wrapper.ts might not exist"

# Copy package.json to install dependencies
cp ~/go/src/github.com/Z3Prover/z3/src/api/js/package.json src/api/js/

# Install dependencies and build
cd src/api/js
npm install
npm run build:ts  # This runs the parser
This will parse v4.7.1's headers and generate the function list. Then you can adapt the build command to use your already-built libz3.a.
Try those commands and let me know what errors you hit.


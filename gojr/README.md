# Go-junior

TypeScript implementation area for the Go-junior spreadsheet language.

This package starts CLI-first: lexer/parser, diagnostics, type/value metadata,
and Node-focused tests live here before browser worker integration.

## Current CLI Slice

```sh
npm run build
node dist/src/cli.js eval --sheet-json '{"A1":40}' 'sheet.A1 + 2'
node dist/src/cli.js run example.gojr
node dist/src/cli.js parse example.gojr
```

## Native Go REPL

The command under `cmd/gojr` embeds the local Homebrew Node/V8 shared library
through CGO and calls the TypeScript Go-junior runtime in-process.

```sh
npm run build
CXX=/usr/local/opt/llvm/bin/clang++ go run ./cmd/gojr
```

From the repository root, use:

```sh
npm run build --prefix gojr
CXX=/usr/local/opt/llvm/bin/clang++ go run ./gojr/cmd/gojr
```

The current bridge links `/usr/local/lib/libnode.141.dylib`, matching the
installed Node 25 ABI. Apple clang 15 cannot parse the current V8 headers, so
Homebrew LLVM is required for now. Inside the REPL, normal input is appended to
the current source buffer and re-evaluated; use `.clear` for a fresh buffer and
`.sheet {"A1":40}` to seed spreadsheet cells.

Implemented in this first runtime slice:

- Chevrotain parse validation plus a typed AST conversion pass.
- Script evaluation for literals, arithmetic, comparisons, locals, assignment,
  returns, `if`, `switch`, simple `for`, `for range`, `break`, `continue`,
  `fallthrough`, `defer`, `panic`, and `panicOn`.
- Explicit spreadsheet refs such as `sheet.A1`, `sheet.$A$1`, cross-sheet
  namespaces such as `Budget.B2`, and ranges such as `sheet.A1:B10`.
- Import-bound host packages, with an early `fmt` package supporting
  `Printf`, `Sprintf`, and `Println`.
- CLI JSON input keeps quoted values as strings, parses integer number tokens
  as exact integers, and parses decimal/exponent number tokens as float64.

Pointers, structs, methods, interfaces, closures, multi-assignment, package
state, type checking, and browser/cache integration are still planned stages,
not part of this initial runtime slice.

# Go-junior

TypeScript implementation area for the Go-junior spreadsheet language,
which is a superset of standard Go.

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
CXX=/usr/local/opt/llvm/bin/clang++ go run ./cmd/gojr build -importpath example.com/demo ./path/to/pkg
```

From the repository root, use:

```sh
npm run build --prefix gojr
CXX=/usr/local/opt/llvm/bin/clang++ go run ./gojr/cmd/gojr
CXX=/usr/local/opt/llvm/bin/clang++ go run ./gojr/cmd/gojr build -importpath example.com/demo ./gojr/test/pkg
```

The current bridge links `/usr/local/lib/libnode.141.dylib`, matching the
installed Node 25 ABI. Apple clang 15 cannot parse the current V8 headers, so
Homebrew LLVM is required for now. Inside the REPL, normal input is evaluated
eagerly against a persistent session. If parsing reaches EOF while more syntax
is needed, the prompt changes to `....>` and keeps the pending multi-line input;
use `.clear` to discard pending input and `.sheet {"A1":40}` to seed
spreadsheet cells.

`gojr build` compiles a package through the embedded JavaScript build API and
writes a Go-like `.a` package archive. The first archive member is `__.PKGDEF`
and contains Go-junior indexed export data for fast importer cache hits; the
generated JavaScript payload lives in a later archive member. By default
artifacts go under `~/go/pkg/js_gojr/<import/path>.a`, mirroring Go's
`~/go/pkg/<goos>_<goarch>/` layout. Use `-pkgdir DIR` to supply a package-cache
parent where `js_gojr` is appended, or `-artifact-root DIR` to supply the exact
artifact root.

Implemented in this first runtime slice:

- Script evaluation for literals, arithmetic, comparisons, locals, assignment,
  returns, `if`, `switch`, simple `for`, `for range`, `break`, `continue`,
  labeled `break` and `continue`, `fallthrough`, labels, `goto`, `defer`,
  `panic`, and `panicOn`.
- Pointers, structs, methods, interfaces, closures, multi-assignment,
  type checking, map/slice builtins, complex numbers, init functions, and
  package/file test execution.
- Explicit spreadsheet refs such as `sheet.A1`, `sheet.$A$1`, cross-sheet
  namespaces such as `Budget.B2`, and ranges such as `sheet.A1:B10`.
- Import-bound host packages, with an early `fmt` package supporting
  `Printf`, `Sprintf`, and `Println`.
- A `testing` package sufficient for the current `.test` and `gojr test`
  workflows.
- Package artifact build envelopes through `gojr build`.
- CLI JSON input keeps quoted values as strings, parses integer number tokens
  as exact integers, and parses decimal/exponent number tokens as float64.

Full Go lowering to executable package JS, import graph builds, browser OPFS
integration, goroutines, channels, `select`, and `recover` remain planned
stages.

## Control Flow Spec

Go-junior supports Go-style labels and branch statements:

```go
Start:
goto Done

Outer:
for i := 0; i < 10; i++ {
    switch i {
    case 5:
        continue Outer
    case 8:
        break Outer
    }
}

Done:
return 0
```

- A label is an identifier followed by `:` and labels the following statement.
- `goto Label` jumps to a label in the active statement scope chain.
- `break` exits the innermost `switch` or `for`.
- `break Label` exits the matching labeled `switch` or `for`.
- `continue` continues the innermost containing `for`, including from inside a
  nested `switch`.
- `continue Label` continues the matching labeled `for`.

The current interpreter treats the AST statement graph as the executable IR.
Future lower IR/codegen stages should preserve the same completion records:
`normal`, `return`, `break(label?)`, `continue(label?)`, `fallthrough`, and
`goto(label)`.

## transliteration from Go into Typescript guidance

The ivy/gojr (Go-junior) directory implements a fully faithful
mechanical transliteration of the /usr/local/go1.27rc1/src/go
Go standard library from Go into Typescript.

This is not a rewrite. This is a transliteration. Preserve 
the structure, the ordering, the decomposition, and the naming. 
Your creative judgment is not needed and not wanted here.

Before translating, list every symbol, function, method, production rule 
in the source. Then translate each one in order. After translating, 
confirm that every item in your inventory appears in the output.

Your job is not to understand or explain the source Go code. Your 
job is to produce a complete mechanical translation into Typescript with 
no omissions. Treat this like a human translator translating 
a legal contract -- every clause must appear in the output, 
even redundant or awkward ones.

If the user says "port" this should be taken as the transliteration
contract-like mechanical converstion described above. There no
room for deviation from the original source logic or naming
in this "port". Exact behavior and symbol level naming must
be preserved.

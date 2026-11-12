# Amap formula parser review

This package is a native Go implementation of the small Boolean equation
language used by Amap gate formulas. The parent `map/amap` package now wires it
through `Amap_ParseFormulaToHop` and uses that bridge when compiling each gate's
truth table. The parent package's unrelated generated drafts still block its
normal package build.

## Behavior reviewed

- Operator precedence is explicit: prefix/postfix negation, AND, XOR, then OR.
- `*` and `&` both mean AND; `+` and `|` both mean OR.
- Adjacent atoms mean AND, matching the source parser's implied conjunction.
- Parentheses and constants `0`/`1` are handled, and malformed/trailing input
  reports a byte offset instead of being accepted as a partial expression.
- `Eval` rejects missing variable assignments. `ToHop` rejects missing Hop
  variables and constructs operations through the existing structurally hashed
  Hop API.
- Traversal and evaluation are read-only, so an `Expr` can be evaluated
  concurrently when callers do not mutate the provided assignment map.

## Verification

The normal companion-package commands pass:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/amap/formula
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/amap/formula
```

The parser bridge and library-equation converter also pass their documented
explicit-file-set tests. Those verify the bridge and exact two-input truth
tables, but they do not prove `go test ./map/amap` passes. The normal parent
package command currently stops on malformed generated code in `amapLib.go`.

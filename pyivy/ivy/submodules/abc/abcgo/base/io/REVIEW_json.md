# JSON review

Scope: `ioJson.go`, shared `ioText.go`, `json_native_test.go`, and `test_vectors/config.json`.

The separate readability pass replaced token-pointer arithmetic with an ordered JSON tree and the standard library decoder. Native nodes retain scalar types, exact number text, duplicate keys, and member order. The original interning/vector API remains available; its scalar-to-string conversion and quoted scalar writer behavior are explicit compatibility semantics.

The separate bug review checked quoted delimiters, escapes, numeric precision, trailing roots, malformed values, vector bounds/arity/cycles, and writer failures. Tests exercise native and vector round trips, layer extraction, a number above float64's exact integer range, duplicate keys, quoted braces/brackets, invalid syntax, cyclic values/vectors, and short writes. Invalid native scalar types are rejected before creating vectors. Recursive extraction validates its input rather than looping over cyclic user-built trees.

Selected-file gates passed:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 base/io/ioText.go base/io/ioJson.go base/io/json_native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet base/io/ioText.go base/io/ioJson.go base/io/json_native_test.go
```

These do not claim full-package completion: ordinary `base/io` testing remains separately blocked by the `base/abc` dependency and other visible I/O drafts.

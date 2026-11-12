# AMAP library review

Scope: `amapLib.go`, `amapLiberty.go`, and scoped library tests.

The separate readability pass replaced parser arena pointers with source spans and explicit child/sibling indices, kept stable source-facing entry points, and used owned Go gate/pin/truth storage. Liberty conversion is independent of SCL and retains the original area-only, unit-delay genlib abstraction; it does not pretend to translate Liberty timing tables.

The separate bug review covered quoted delimiters/comment markers, malformed nesting, sequential/dont_use filtering, stable minimum-area truth selection, full truth keys, writer errors, and round-trip precision. A quoted parenthesized function initially failed parsing; the regression now passes. `dont_use:false` stays usable. A sticky writer now captures errors from the legacy DSD printer, including short writes.

Liberty-only selected race and vet passed:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race map/amap/amap.go map/amap/amapInt.go map/amap/amapCompat.go map/amap/assert.go map/amap/amapLiberty.go map/amap/amapLibraries_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet map/amap/amap.go map/amap/amapInt.go map/amap/amapCompat.go map/amap/assert.go map/amap/amapLiberty.go map/amap/amapLibraries_test.go
```

`amapLib_native_test.go` additionally covers genlib/equation/selection/write/read integration and retained truth/pin values. Its selected integration race gate passed with the actual native read, formula bridge, parser, rule, permutation, and uniqueness sources. Whole-package completion remains a separate gate while other concurrently owned files are changing.

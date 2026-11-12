# Priority-cut mapper port

The package is an active port; the package-wide build is currently blocked by
source-derived C syntax in decoder and mapper files. Do not treat a successful
isolated file compile as package completion.

## Current integration work

`if.go` now has a Go-native first-pass data model: vector, memory and timing
types are aliases to their owning packages; variable-sized cut leaves and
permutations are slices; cut scratch data has explicit fields; callbacks are
typed Go function values; and the manager has typed truth-table and pair-cache
storage. `If_Regular`, `If_Not`, and related helpers replace C pointer-tag
casts with explicit complement proxies. This is an interim compatibility model
whose direct-field behavior still needs whole-mapper validation.

`ifCache.go`, `ifCheck.go`, and `ifCore.go` have had their first syntax and
representation conversions. The library command adapters in `ifCom.go` now
use Go slices and writers and a local flag-only option parser. Algorithm files
remain visible and are being ported in non-overlapping groups.

The current syntax frontier reported by `go test ./map/if` is `ifDec07.go`.
Other parse-invalid files remain, including `ifDec08.go`, `ifDec10.go`,
`ifDec16.go`, `ifDec75.go`, `ifDelay.go`, `ifDsd.go`, `ifLibBox.go`,
`ifLibLut.go`, `ifMan.go`, `ifMap.go`, `ifReduce.go`, `ifSat.go`,
`ifSelect.go`, `ifSeq.go`, `ifTest.go`, `ifTime.go`, `ifTune.go`, and
`ifUtil.go`. Fixing the parser frontier alone will not establish semantic
correctness; manager lifecycle, truth handling, cut enumeration, mapping,
sequential paths, and library I/O need behavior tests.

## Verification

The full package race tests and vet are **not green**. Remaining generated
files contain invalid Go syntax, so the package has not reached full type
checking. The header/cut/truth/count component does pass race tests and vet
when built from its explicit file set, including regressions for complement
identity, cut-copy ownership, and support-mask boundaries:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/if/if.go ./map/if/if_.go ./map/if/ifNativeTypes.go ./map/if/ifCut.go ./map/if/ifTruth.go ./map/if/ifCount.go ./map/if/assert.go ./map/if/ifHeader_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/if/if.go ./map/if/if_.go ./map/if/ifNativeTypes.go ./map/if/ifCut.go ./map/if/ifTruth.go ./map/if/ifCount.go ./map/if/assert.go ./map/if/ifHeader_test.go
```

Recheck the full package with `GOCACHE=/tmp/abcgo-cache go -C abcgo test
./map/if` and `gofmt -d abcgo/map/if/*.go` as owners advance it. Do not run
tests under `experiments/`.

Use `test_vectors/` for fixtures; do not create `testdata/` directories.

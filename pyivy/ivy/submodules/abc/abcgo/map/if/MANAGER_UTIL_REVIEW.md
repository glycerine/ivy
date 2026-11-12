# Manager and utility review

## Test boundary

Reviewed `ifMan.go`, `ifUtil.go`, `ifCache.go`, and `ifVecCompat.go` after the
native ports. The explicit-source race suite and vet gate included these files
plus the cut, truth, and count family:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./map/if/if.go ./map/if/if_.go ./map/if/ifNativeTypes.go \
  ./map/if/ifMan.go ./map/if/ifCache.go ./map/if/ifVecCompat.go \
  ./map/if/ifUtil.go ./map/if/ifCut.go ./map/if/ifTruth.go \
  ./map/if/ifCount.go ./map/if/assert.go \
  ./map/if/ifMan_native_test.go ./map/if/ifUtil_native_test.go \
  ./map/if/ifCut_native_test.go ./map/if/ifTruth_native_test.go \
  ./map/if/ifCount_native_test.go ./map/if/ifHeader_test.go \
  ./map/if/ifCut_c_test.go -count=1
```

Result: the isolated selected-file gate passes (`ok command-line-arguments`,
race enabled). `go vet` on the same explicit source list also passed. This is
not a `map/if` package test.

## Review findings

- Fixed mapping reference accounting so each visited CI/constant receives the
  same reference increment as the source implementation.
- Replaced reverse-order scratch pointer casts with level buckets while
  retaining the original descending-level and reverse-insertion ordering.
- Cross-cut traversal resets visit counts on return, including panic paths.
- Cut traversal clears temporary marks after collecting leaves and fanins.
- Manager stop is idempotent; restart resets graph-owned state and preserves
  reusable caches.

## Remaining limitation

This validates only the explicit source subset. It excludes other package
files, including the still-invalid `ifDec*.go` decomposition drafts and the
in-progress `ifMap.go` and `ifReduce.go`. Therefore plain `go test ./map/if`,
package-wide race tests, and package-wide vet are not green.

# Root ADD numeric family review

Completed `cuddAddAbs.go`, `cuddAddFind.go`, and `cuddAddInv.go` as forwarders to
the sole core implementations. Public sum/product/OR abstraction, extrema,
ADD-valued integer bits, and scalar reciprocal are all implemented. Live
internal recursive entrypoints are retained as forwards; no graph algorithms
are duplicated. Original C and copyright/license headers remain visible.

Readability pass: kept public source names and int32 bit parameters, removed
C-private pointer/cache mechanics, and used explicit short forwarding functions.

Bug pass: root `cuddAddNumeric_test.go` checks public abstraction/extrema/bit and
reciprocal forwarding; detailed boundary, ownership, shuffled-order and
original-C differential tests live beside the authoritative core. Separate
algorithm review: `core/ADD_NUMERIC_REVIEW.md`.

The completed files compile/race-test together with `cuddBridge.go`, `cudd.go`,
and `bddCore.go`. The full root package remains blocked by other visible drafts;
this review does not claim it builds.

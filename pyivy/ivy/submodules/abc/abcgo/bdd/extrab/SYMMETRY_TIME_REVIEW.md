# Symmetry and deadline operators review

Original `extraBddSymm.c` and `extraBddTime.c` remain unchanged and visible.

## Readability and compactness review

Symmetry keeps the original cofactor-pair recursion, skipped-variable handling,
singleton relation and tuple dynamic programming. Typed memo keys distinguish
functions, variable cubes and tuple cardinalities. A typed status replaces the
source's overloaded BDD one/false and ADD-zero sentinels in its structural path
prefilter. That fast scalar prefilter remains distinct from the exact scalar
cofactor test; it is not relabeled as an exact swap test. Both bulk APIs retain
the source matrix conventions (fast upper triangle, naive symmetric matrix).

Timed AND calls guarded authoritative core ITE. Fused timed AND/abstraction and
ordinary core AND/abstraction share one cofactor recurrence, with no intermediate
AND graph. Timed and ordinary BDD transfer also share one recurrence. ADD transfer
retains its separate numerical path. Cancellation state is passed explicitly,
never stored globally or in the manager.

## Bug review

Symmetry validates positive cubes, function ownership and complete support before
recursion. Pair decoding rejects nonpairs/odd literals/out-of-domain variables.
Empty-set membership is preserved in singleton/getVars families, as required by
the source recurrence; stripping it would silently alter pair results. Zero-sized
info allocation is safe, unlike the source's indexing of its first matrix row.
Tests preserve referenced roots across reorder and exercise skipped support.

The C timeout parameter is an absolute process CPU-clock integer; its zero value
means unlimited. Go preserves zero-unlimited using zero `time.Time`, and uses
absolute standard-library deadlines (monotonic component when available). Unlike
C terminal/cache shortcuts, an already expired Go deadline consistently returns
nil even for trivial results. This deliberate boundary difference is documented
at the API. It avoids a misleading integer clock emulation. Cancellation checks
run inside recursion and canonical ITE rebuilds. Only complete nodes/cache entries
are installed; partial successful work stays canonical and retry is supported.
No asynchronous goroutine touches a manager. Deadlines are cooperative, not hard
preemption of allocation, input validation or support traversal.

The old timed transfer claimed ADD support but maps every constant to Boolean
one and complement. The new timed transfer explicitly accepts BDDs only rather
than silently corrupting numeric ADD leaves; ordinary numerical Transfer remains
available. Mapping validation precedes projection growth; projection growth itself
checks cancellation between variables. Negative quantified cubes are rejected.

## Evidence

All 256 three-variable functions match naive pair computation; all 65,536 pairs
match direct cofactor singleton relations. All exact symmetries pass the source
structural prefilter. Tuple cardinalities, constants, invalid literal families and
reordered managers have regressions. The permanent original-C symmetry checksum
`72c5fd61011d557d` includes bulk metadata/matrices, scalar prefilter and all function
pairs. The timeout-disabled original-C operator checksum `f6ee7ba969cdfbe5`
includes all conjunction/abstraction pairs and remapped transfers.

Deadline tests cover zero/future/expired time values, trivial/cached operands and
invalid domains. Core deterministic cancellation contexts interrupt inside AND,
fused abstraction and transfer recursion, then retry on the same managers. A
separate exhaustive 256×256×8 cube test compares fused abstraction against an
independent conjunction followed by existential abstraction. See
`FULL_PACKAGE_REVIEW.md` for normal package race/vet gates.

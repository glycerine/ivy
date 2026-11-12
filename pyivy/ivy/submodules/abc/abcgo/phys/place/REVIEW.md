# Placement foundation review

Scope completed here: `place_base.c.go`, `place_base.h.go`,
`place_gordian.h.go`, and `place_inc.go`. These port the original database,
bounding-box/HPWL calculations, sparse-ID insertion/deletion, partition-member
removal, typed sorting predicates and incremental wirelength estimate. The
inactive `#if 0` experimental fastPlace body is not part of the C API.

Distinct readability/idiomatic pass: typed slices replace pointer arrays and
realloc; growth and numeric comparisons share short helpers. Original internal
names remain for integration with the remaining placement algorithms. Public
constructors and typed wrappers expose geometry/database functions without raw
pointer casts. Counts and IDs remain int32 and coordinates float32.

Distinct bug pass fixed:

- deletion of the final sparse object underflowing the table index;
- empty nets producing negative bounding extents;
- the vertical estimate using the rectangle's x coordinate instead of y;
- stale deletion removing a newer object registered under the same ID;
- overflow in intermediate hash multiplication and allocation-size arithmetic.

Focused regressions cover sparse gaps/deletion to zero, independent x/y distance,
empty nets, partition area/member updates, nil ordering, terminal-list ownership,
and stale object identity. Focused copied-source race tests pass in
`/tmp/abcgo-place-base-check`; the only added fixture is the exact global
`g_place_rootPartition` declaration from place_partition.c.

The whole package is **not yet buildable**: the original generated Bookshelf test
harness, QP solver, partitioning and other placement files still contain C draft
syntax. No workspace source was hidden or excluded to conceal these failures.

## Recursive partitioning and native hypergraph refinement

The next completed group is `place_partition.go` and `libhmetis.go`: root
initialization, corner sorting, recursive equal-area refinement, child-bound
resizing, reallocation, incremental insertion, and weighted hypergraph refinement.
The original hMetis declarations had generated empty Go bodies; those are now
real implementations. The disabled experimental FM and scanline entry points
also perform real gain refinement and balanced geometric min-cut selection.

The standard-library hypergraph engine intentionally differs from proprietary
hMetis: it uses greedy balanced construction, positive-gain single moves and
pair swaps, deterministic multistart ordering, and exact branch-and-bound when
the assignment/edge-work budgets permit. It does not implement hMetis coarsening,
its full tuning-option set, or identical random-seed answers. Both compatibility
entry points use this direct k-way solver. `options[6]` locks nonnegative seed
labels and `options[7]` sets the shuffle seed. Imbalance means the allowed
percentage above average part weight; one oversized vertex is allowed to fit.
Errors preserve the caller's assignments and cut result. Placement refinement
preserves far-from-cut/fixed assignments, using scaled integer cell areas rather
than truncating sub-unit areas to zero.

Separate compactness/readability pass: shared member-accounting and partition
validation helpers replace repeated pointer-array manipulation. Cut-gain trials
inspect incident edges, and work limits have named constants. The original
placement fields and count conventions remain compatible with other source files.

Separate bug pass fixed missing incremental area updates, completed leaves not
reopening after growth, empty-root/sparse incremental scanning, division by zero
for zero-area partition bounds, fractional-distance truncation at cut selection,
intermediate hyperedge offsets escaping the pin array, and arithmetic overflow
in percentage-cap rounding. A failed optimizer leaves the equal-area assignment
available, with an explicit diagnostic.

Focused tests compare random small weighted partition problems against exhaustive
search, exercise known clustered cuts (including a larger heuristic case), fixed
seeds, three-way partitions, weighted edges, infeasible constraints, malformed
inputs, empty graphs, actual placement cut improvement, recursive membership/area
and geometric coverage, zero-area termination, and incremental sparse additions
and reopening. The focused copied-source suite passes with `-race` in
`/tmp/abcgo-place-base-check`; its sole fixture is now the exact
`g_place_numPartitions` declaration from place_gordian.c. The QP solver, Bookshelf
harness and other remaining drafts still prevent whole-package compilation.

## Quadratic placement solver

Completed `place_qpsolver.c.go` and `place_qpsolver.h.go`: float32
preconditioned conjugate gradients, quadratic line minimization, exact elimination
of one dependent cell per disjoint weighted COG group, closed-loop quadratic
constraints, box multipliers, and original qps_init/qps_solve/qps_clean entry
points. The typed `QPProblem` API copies initial data and exposes edge construction,
COG/loop/bounds configuration, Solve, coordinates, objective and failure status.
Production code uses only math/fmt and Go slices; no C runtime or compiler.

Separate compactness/readability pass: typed edges/groups replace private C
pointer-offset buffers. Shared list validation, coordinate expansion, loop energy,
gradient contributions and constraint updates keep the numerical operations
visible. Original CG index/update order and the default preconditioner are retained;
debug logging and the disabled HOIST experiment are not production paths.

Separate bug pass/regressions corrected:

- zero-gradient/all-fixed solves leaving the private objective unset;
- unconstrained solves falsely setting failure because the unused dual step was zero;
- bounds checking stale input positions rather than the current iterates;
- temporary objective estimates overwriting current cost/loop residuals;
- multiple-loop parsing skipping a list item after its sentinel;
- box-only iterations lacking a hard termination limit;
- nonfinite/singular line steps producing NaNs rather than reporting failure;
- overlapping movable COG groups, malformed indices/lists and invalid numeric data;
- translation-nullspace box problems starting outside the feasible region.

The original dual step heuristic can underestimate the required objective bound
or shrink its step before feasible loop/box constraints converge. The port first
runs that update scheme, then recovers failed constrained solves by bracketed
coordinate maximization of the **same quadratic dual**, using the same CG inner
solver. This is an intentional numerical correction, not a replacement objective.
An initial feasible box/COG point handles unanchored translation nullspaces.
Infeasible fixed constraints return an error/failure code with bounded work.

Numerical fixtures cover weighted anchored chains, isolated/all-fixed cells,
weighted COGs including dependent-only groups, active boxes, one/multiple active
loops, infeasible fixed loops, empty problems, repeated solves, finite-difference
gradients, and state preservation during objective estimation. An optional
test-only C harness compares 40 randomized anchored chain/COG cases with the
original place_qpsolver.c; it was actually run with cc and passed (coordinate and
objective tolerances account for float evaluation). This test skips if the
original source/compiler is unavailable; ordinary numerical tests always run.
For copied-source testing, `ABC_QP_REFERENCE_SOURCE` identifies the original C file.

Focused database/partition/QP tests pass with -race in
`/tmp/abcgo-place-base-check`. Whole-package compilation still awaits the remaining
Bookshelf harness, placement driver, pad/bin and QP-generation drafts.

## Completed package integration: driver, geometry and Bookshelf

The remaining source files are now production Go, including the three-file
Bookshelf harness (reusable `RunBookshelf` plus `cmd/bookshelf`). No source drafts
are excluded. `place_test.go` contains workflow tests rather than C syntax.

A separate readability/compactness pass retained the original partition/QP flow,
shared the two density-axis implementations, replaced raw name hashing with Go
maps, and isolated scanner, count/number parsing and buffered output helpers.
The existing conjugate-gradient/constraint solver was preserved.

A separate bug review found and corrected:

- the last leaf COG constraint being discarded by the original count decrement;
- sparse holes and newly added cell coordinates in rebuilt quadratic problems;
- top-boundary cleanup using the horizontal origin/extent in its denominator;
- fixed-cell core extents confusing maximum coordinates with width/height;
- division by zero and nonfinite results for tied density-bin coordinates;
- density limits applying only to bin edges rather than each moved cell;
- the vertical density pass using full height instead of half height at bounds;
- Bookshelf output ignoring its requested filename prefix;
- count mismatches, duplicate names, unknown pins and nonfinite input values;
- partial database/coordinate changes after a failed read, and repeated-terminal
  deletion overflowing the C harness's cell-count-sized temporary buffer;
- ignored file creation, write, flush and close failures.

Preplacement supports netlists without pad cells, uses the largest non-pad height
as row height and the largest pad dimensions for a heterogeneous pad ring. Fixed
cells retain their positions and full physical extents. Analytic global placement
still uses the source's horizontal density pass; vertical spreading remains a
separate callable internal operation. Density correction is not obstacle-aware.
`Legalize` supplies deterministic nearest-row packing around fixed cells/pads;
the original C `legalize` function was empty. It commits positions only when all
movable cells fit, and reports failure otherwise. This greedy legalizer does not
promise to find every feasible arrangement.

Bookshelf retains ABC's center-coordinate convention; orientations and pin
offsets are not modeled. Net import assigns unit weights, as the original
harness did. Package database state is process-global and callers must serialize
mutations; a passing race suite does not make concurrent database access safe.

Geometry regressions cover fixed extents, pads, tied bins and movement limits,
obstacle packing/nonoverlap and transactional failure, all leaf COG constraints,
nonfinite input, sparse clique solves and incremental coordinate initialization.
I/O regressions cover real files, names/fixed flags/connectivity/HPWL round trips,
malformed records, write failures and the complete global/incremental harness.

Verified on the actual full package:

- `go -C abcgo test -race ./phys/place/...` passes, including the command wrapper.
- `go -C abcgo test -count=1 -run TestQPOriginalCComparison -v ./phys/place`
  passes all 40 original-C numerical comparisons (not skipped).

These full-package checks supersede the earlier copied-source-only status above.

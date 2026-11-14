# Multivalued/symbolic simulation review

Implemented the actual saigSimMv.c algorithm in Graph-level aigMvSim.go, shared by
AIG sequential sweeping and the thin SAIG wrapper. Every symbolic frame gets fresh
PI literals. AND reductions and structural hashing preserve exact symbolic
functions until symbolic cutover; later nontrivial operations become undefined.
Known constant and equal/complement cases are still simplified in later frames.
Undefined values remain undefined through complemented edges.

Simulation stores all-zero initial latch state, transfers latches simultaneously,
hashes complete immutable state snapshots and stops on a repeated state. At the
first saturation threshold it retires registers that ever became undefined. At
twice that threshold it retires other nonconstant/nonoscillating registers.
The original oscillator candidate search and all-state const/equivalence tests
are retained. Output mappings merge only registers whose histories never became
undefined and match in every observed state. The shared Scl/ConstReduce path uses
this real default engine; optional injected engines avoid package import cycles.

## Readability/compactness pass

Use Go maps for symbolic AND keys and state lookup; typed state slices replace
fixed allocators and linked hash chains. Symbol identities need a monotonic
literal counter and hash keys, not a duplicate node/level/capacity array: the C
level limit was disabled and its unused accounting fields have no active effect.
Removed unused register-value buffers and allocation-count mirrors. Explicit
Graph traversal supports buffers, XORs, outputs and rewritten DAG ordering.

## Separate bug pass

Validate symbolic/saturation ranges and bound saturation before doubling can
overflow int32. State snapshots own their values and map keys, preventing mutable
state storage from changing old hash entries. Undefined histories are excluded
from equivalence classes. Tests exercise symbolic PI freshness, non-symbolic
cutover, constants, equivalent binary latches, unknown exclusion, oscillator
retirement and repeated-state indices.

A reference test extracts and compiles the actual original C create/hash/find/AND
routines when a C compiler is available. It compares Go against original C on
constant/undefined/complement cases plus 300 chained operations that exercise
hash collisions and original-C storage reallocations. This test passes here;
production Go code has no C compiler/runtime dependency.

Focused actual-core MV/partition tests pass with -race. The thin SAIG forwarding
API and parity tests pass with -race in a temporary integration module containing
the actual converted AIG core and actual dependency packages. Full SAIG-package
build remains pending its other generated files.

## Sequential source audit

The preceding integration report is historical. The current audit compares
the complete simulator with checked-in saigSimMv.c. A red regression found
that allocating every PI symbol before gates changed literal identities when
a PI was created after a gate. Frame processing now preserves source object
order on topological graphs. A cached dependency order handles Go's rewritten
DAGs, buffers and native XORs; a forward-buffer test verifies that extension
without changing graph scratch fields. The graph must remain fixed while a
manager is active. Reusing the frame order and next-state buffer removes the
per-frame traversal arrays and recursive walk.

A compiled full-source fixture compares frame values, symbolic object counts,
state-repeat indices, undefined/constant/binary/oscillator classes, skip masks,
the derived mapping and the complete driver. Another fixture compares period
search and state hashing over 4,367 histories, including unsigned hash overflow
and wraparound of the sixteen primes. The existing original-C AND/hash/create
comparison remains active.

A separate six-register delay-chain fixture reaches aggressive saturation
before convergence. Go matches the original C's complete printed state trace
and final mapping. No-PI undefined values in that trace prove retirement was
exercised rather than ordinary symbolic cutover.

Postprocessing deliberately derives undefined classification from histories:
C's separate counters are never populated. A focused test checks diagnostic
group counts, exclusion of undefined histories and lasso markers. This is an
explicit correction to the diagnostic, not a substitute for the simulation.
See ../../SEQUENTIAL_AUDIT.md for the current validation and package scope.

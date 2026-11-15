# CTAS circuit SAT (giaCTas.c and giaCTas2.c)

The `tas_*.go` files port giaCTas.c's manager, trail/relative reasons, suspended
frontiers, four decision policies, circuit propagation, watched learned clauses,
reason expansion/resolution, backjumping, resource limits, miter workflows and
simulation-pattern packing. Clause storage retains source's integer handles and
three-word headers; Go wrappers reacquire the backing storage after growth.
Source's activity bump/decay operations remain disabled. giaCTas.go preserves
attribution and points to these implementation files instead of retaining a
570-line empty generated comment skeleton.

Native comparisons cover 84,864 single/paired/array queries across four policies,
resource limits and random/hard five-input graphs. Verdicts, per-query/cumulative
search counters and ordered CI witnesses agree with actual C, including all three
outcomes, up to 21 backtracks and frontier ten. Independently enumerate every
completion of each cube and check graph/trail/reason/phase/watch/clause/activity
cleanup after every query. Another 3,600 native public-miter properties compare
statuses, witnesses, logical CI/CO permutations and source phase clearing.

Direct native watch snapshots exercise watch movement, unit propagation,
cancellation, conflict learning, forced queue/store growth, complete reason and
watch state, and both clause reports. Pattern tests compare 512 native/scalar
cross-word merges at bits 0/1/31/32/33/63/64/95 and 64 packing/saturation cases.
An independent scalar bitmap verifies every value/presence bit. Another 80 native
packed-miter workflows cover four budgets, logical ports and 17/1000/1001/2000/
2001-output sampling boundaries; exact simulation words and old-root proof marks
match. Independent truth tables verify proofs and sampled false properties.

Additional exhaustive Go checks cover 94,208 raw-gate single/paired/contradictory/
three-assumption queries and 115,664 five-input root pairs across all policies.
Every verdict and nonempty CI-cube completion is independently checked. These
include raw constants, buffers and repeated/opposite fanins that usual C hashing
simplifies away. Regression-backed fixes preserve constant values, omit fixed
axioms from reasons and recheck a shared second fanin after assigning the first.

Other red regressions restore the original statistics, reject unsupported typed
or resized graphs, validate every assumption before modifying model/trail state,
and cancel remaining graph assignments during active Stop. Destruction also
tolerates entries already manually unassigned by low-level callers. Foreign
clause ownership is checked before watch or reason-queue mutation. Nested saved
frontiers across growth, partial cancellation, strict limits and repeated/stopped
lifecycle cases have dedicated tests. Low-level graph preparation remains the C
caller's responsibility; the public miter workflows perform it themselves.

Compactness restores source reusable mark scratch and phase-based resolution
deduplication instead of per-conflict slices/maps, and uses a fixed two-root array.
Hard single UNSAT queries drop from 3,720B/127 allocations to 448B/28 allocations;
paired queries from 2,056B/76 allocations to 256B/16 allocations. Remaining small
clause-handle wrappers are retained without adding a larger cache. Final timings
are approximately 14.7/8.0 microseconds, versus 27.3/16.5 previously; these are
measurements rather than performance guarantees.

giaCTas2.c is an incomplete allocator experiment. Go exposes its zero-initialized,
sequentially numbered variable records and stable pointers. Ten one-bit flags use
bool rather than uint32, reducing Go records from 76 to 48 bytes. Tests check all
new-record fields, growth and preservation of already modified records. This does
not introduce another SAT algorithm.

The final focused Tas suite passes in 12.512s. Full race/vet integration is
recorded in SEQUENTIAL_AUDIT.md. This review covers the CTAS family; GIA's other
families and the sequential 131-package objective remain open.

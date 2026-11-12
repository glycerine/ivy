# Cooperative context and bounded ISOP review

## Readability and compactness pass

ITE has one authoritative implementation with an optional recursive check; the
ordinary method is its nil-check wrapper. BDD transfer has one shared recurrence
for ordinary/context APIs; numerical ADD transfer remains its existing separate
path. Fused conjunction/abstraction is shared by ordinary/context entrypoints,
using canonical children and a call-local three-operand memo.

Bounded ISOP uses the original Minato recursion with cube-count metadata, not a
second cover algorithm. It never constructs ZDD nodes. Allocation limits use the
existing typed BDD allocator mechanism and preserve a tighter enclosing limit.

## Bug review

Cancelled ITE never caches nil or creates a node with a nil child. Cancellation
checks precede terminal/cache shortcuts, making an expired context predictable.
Transfer validates source ownership/mapping bounds before growth, checks growth
between projections, propagates complement phases, and rejects numeric ADDs at
the explicitly BDD-only context API. Partial complete work remains canonical;
retry does not require manager reset. Contexts do not make managers thread safe.
Fused abstraction validates positive cubes, skips absent quantified variables,
and ORs quantified children through the same guarded ITE. A one child terminates
quantified recursion without evaluating the other child.

ISOP cube counts guard subtraction before sum overflow. Failed node allocations
roll back newly created BDD unique nodes/IDs and flush their ITE cache entries;
existing diagrams and prior allocation ceilings survive. Typed limit recovery
rethrows unrelated panics instead of concealing bugs. Counts include constant
one under strict zero-cube limits. No process-global chosen-phase count cache is
used by extrab's repeated vector calls.

## Regressions

Deterministic contexts cancel within each recursion and then retry on the same
managers. All three-variable conjunction pairs and all eight quantification
subsets compare to separate AND followed by existential abstraction. Existing
transfer/reorder tests remain active. Bounded ISOP counts compare with real ZDD
cover counts for all 256 functions; tests check zero-ZDD allocation, strict limits,
rollback and prior-limit restoration. Full core race/vet gates are recorded in
`../../extrab/FULL_PACKAGE_REVIEW.md`.

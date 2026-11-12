# CUDD core review record

Two distinct review passes were performed (2026-11-12).

## Compactness, idiomatic Go, and readability pass

- The core subpackage is the sole authoritative graph implementation. Completed
  root CUDD modules re-export its types/functions rather than duplicate algorithms.
- Explicit canonical complemented-edge handles replace tagged pointers. A unique
  table and Shannon ITE engine centralize Boolean construction and normalization.
  Node families distinguish BDD and ZDD graphs; numeric zero differs from the
  complemented BDD-one edge.
- Go maps own canonical nodes and computed entries. Manager lifetime, advisory
  reference counts and external synchronization are documented. This is not a
  reproduction of C raw-memory pools or their tuning policies.
- Standard math/big and math/bits implement exact APA arithmetic. Formula parsing
  remains backend-independent; the concrete adapter imports misc/parse, so no
  package cycle or second parser is introduced.
- Existing original modules (public declarations, initialization, BDD ITE,
  abstraction and APA) retain their copyright/license and module organization.
  All executable Go files are gofmt-formatted.

## Bug review pass

Findings were corrected with regressions before integration:

1. Invalid permutation destinations could partially grow the manager or exceed
   the constant-index boundary. Validate every used destination before mutation;
   tests check negative/out-of-range targets and unchanged variable inventory.
2. A BDD-only ownership test made RecursiveDeref ignore numeric ADD constants.
   Lifecycle ownership validation now permits dereferencing those constants;
   a reference-count regression verifies regularization and decrement.
3. IEEE negative zero could receive a distinct constant handle. Normalize signed
   zero before unique-table lookup; a regression checks canonical pointer identity.
4. Node statistics omitted numeric constants and ZDD nodes. Include both unique
   families and the constant table; a statistics regression checks the total.
5. Abstraction initially accepted arbitrary support-bearing expressions as cubes.
   Enforce positive-cube form; tests reject negative/disjunctive cubes.
6. BDD implication/dependency queries must not allocate nodes. Use memoized
   disjointness recursion and assert unchanged unique-table size in tests.
7. Allocation-limit failure could expose partial unique-table/cache state.
   Roll back newly allocated nodes and computed entries; tests force failure
   after partial allocation and then verify successful subsequent operations.

Verification: `GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/cudd/core -race -count=1`.
Coverage includes 1,500 deterministic randomized Boolean operations, exhaustive
four-variable truth tables and pointer canonicity, complement laws, node ordering,
permutation/composition, cofactors, quantification, dependency, exact APA counts
beyond floating precision, arithmetic aliasing/carry/borrow, writer errors, ZDD
universe cardinality, manager ownership/lifetime, and real parser/backend parity
including nested ranks and invalid rank shifts.

## Remaining scope

Root bdd/cudd is still incomplete: roughly 70,000 draft lines covering ADD/ZDD
operations, reordering, exports, advanced decompositions and remaining CUDD APIs
have not yet been ported. They remain in the ordinary source tree and still
prevent root-package compilation; no draft files are hidden or build-tagged.
Dependent completed packages can import this tested core while that work continues.
Passing core tests is evidence for this implementation, not completion of CUDD
or the full ABC port.

## ADD extension: separate follow-up review passes

### Compactness, idiomatic Go, and readability review

- ADD construction, ITE, apply, permutation and queries live only in core; root
  original modules contain forwarding aliases. Numeric nodes use a separate
  unique table, so ADD's real zero is never confused with BDD's complemented one.
- BDD and ADD support now share the same level-sorted traversal rather than
  carrying a second copy in ADD permutation. ITE routines retain specialized
  terminal/complement rules, keeping the distinction visible and reviewable.
- Apply uses typed original-signature callbacks and operation-local memo tables;
  no reflection-based function identity or unsafe conversions are needed.
- Numeric operations use math, and IEEE NaN/infinity values remain numeric
  constants, matching the original double-valued operation domain. Formatting
  and exact arithmetic continue to use standard-library implementations.

### Bug review and fixed findings

- REO integration exposed that generic Support rejected ADD roots. Removed that
  BDD-only gate and deleted the duplicate ADD support traversal. Tests verify
  ADD support indices, canonical BDD support cubes and inverse permutations.
- Generic minterm counting had the same BDD-only entry check. It now counts
  non-background ADD terminals as well; tests compare floating and exact APA
  counts and check numeric zero and node statistics.
- Initial NaN rejection was stricter than the C double-valued operators. Preserve
  IEEE NaNs, including log(-1), instead of returning an unsupported-operation
  result. Signed zero remains canonical; infinities are distinct constants.
- Reviewed callback cache separation: function-pointer code addresses are not
  safe identities for Go closures. Operation-local caches prevent captured
  callbacks from sharing results. A two-offset callback regression verifies this.
- Apply now stops with failure when a callback cannot handle constant operands,
  rather than recursing at an invalid constant level. Tests cover this protocol
  boundary and foreign-family/foreign-manager results.
- ADD tables/computed entries must be cleared on manager shutdown, and statistics
  must include ADD nodes. Lifecycle and accounting regressions verify both.

Follow-up verification:
`GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/cudd/core ./bdd/reo -race -count=1`.
Coverage includes exhaustive numeric leaves after permutation/inverse mapping,
ADD/BDD family isolation, mixed-root REO integration, pointwise arithmetic,
threshold/overwrite/comparison/agreement operators, 0/1 operators, monadic
logarithm/negation, nonallocating constancy/order queries and callback isolation.

Remaining scope is still substantial: advanced ADD abstractions/matrix operations,
ZDD operations, dynamic reordering and many original root CUDD APIs remain drafts.
No dynamic CUDD ReduceHeap or SYMM_SIFT implementation is claimed. The root CUDD
package remains incomplete and its original draft files remain visible.

## ZDD set-operation group: separate review passes

### Compactness, idiomatic Go, and readability review

- The zero-suppressed unique table remains authoritative. One shared ordered
  Shannon traversal implements union/intersection/difference with explicit
  operation symbols and absent-variable cofactors; separate algorithms are not
  copied into root wrappers.
- ZDD ITE uses the exact finite-family identity `(f ∩ g) ∪ (h \ f)`, so its
  meaning remains clear without reproducing C memory/cache machinery.
- Positive/negative cofactors share one traversal. Variable toggling remains a
  small distinct recursion because it exchanges inclusion and exclusion branches.
- Exact counting uses memoized math/big integers, with bounded original-style
  integer/double entry points. Counts are calculated over the graph rather than
  by enumerating member sets. The returned exact integer is caller-owned.
- Original root set/count modules forward only; copyright/licenses are retained
  on the authoritative core modules. All executable files are gofmt-formatted.

### Bug review and regression checks

- ZDD zero is a real terminal, and equal then/else children must not trigger BDD
  reduction. Reviewed all constructors for the ZDD-only rule `then == zero → else`.
  Tests check every live ZDD node's ordering, family and reduction invariants.
- Inclusion queries must not reuse a nonconstant sentinel as an actual result.
  The nonallocating query has its own memo table; set operations use their own
  canonical-edge caches. Exhaustive pair tests verify the inclusion result and
  unchanged unique-table size.
- Universe rebuilding originally assumed identity order. It now uses InvPermZ;
  fresh-manager reversed-order tests verify universe membership and set results.
  Public universe rebuilding rejects invalid permutation inventories before
  changing the universe. This does not implement dynamic graph reordering.
- Variable growth must remain separate from BDD inventory, preserve canonical
  existing families, and validate indices before mutation. Growth/isolation,
  negative/constant-boundary index, and oversized initializer tests cover this.
- Invalid complemented/BDD/foreign-manager edges must not enter ZDD recursion.
  Domain checks and lifecycle regressions verify rejection and cleanup.
- Integer counts must not wrap into plausible results. Checked conversion
  reports the existing failure sentinel on overflow; exact and double APIs
  correctly report a universe containing 2^100 sets. Mutable return-value tests
  verify callers cannot corrupt subsequent results.

Verification:
`GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/cudd/core ./bdd/reo -race -count=1`.
All 256 families over three variables and all 65,536 family pairs are checked for
canonical union/intersection/difference and nonallocating inclusion. All families
are also checked for both cofactors, change/involution and ITE. Tests cover
projection functions, zero suppression, universe/counting, independent variable
inventories and a fresh nonidentity variable order.

The eight original ZDD set-operation APIs and two original counting APIs are now
implemented. ZDD product/division, covers/ISOP, conversions, dynamic reordering,
and remaining original root CUDD APIs are still incomplete. No existing root
draft files are hidden and no dynamic-reordering compatibility is claimed.

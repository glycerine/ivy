# PSAT interface implementation and reviews

## Original-source evidence

This checkout has only m114p.h, m114p_types.h and an empty `SRC +=` manifest under
src/sat/psat. There are no MiniSat-1.14p solver/proof implementation bodies to
port or compile as an original-C oracle. The behavior contract is supplied by
those declarations and src/proof/int/intM114p.c's interpolation consumer.
The typedef compile error is fixed by replacing the external integer handle
with a Go-owned solver pointer. All fifteen declared function names are present.

## Real Go behavior

The package implements an independent proof-producing Davis-Putnam resolution
elimination algorithm; it delegates to no SAT backend. Variables are eliminated
by minimum positive/negative occurrence product. Non-tautological resolvents
retain actual earlier clause IDs and pivots; reverse elimination reconstructs
SAT models. There is no variable-count cutoff. The finite work limit counts
attempted resolutions, with zero/unset meaning unlimited. Result codes match the
consumer: SAT=1, UNSAT=0, undecided=-1. Work statistics are resolution counts, not
claims of the absent MiniSat engine's exact conflict/search trace.

Roots preserve input order and root IDs. A chain has k pivots and k+1 earlier
clause IDs; normal refutations end with the empty clause. This matches the
interpolation consumer's traversal contract. Assumptions are explicit additional
unit roots after BaseRoots. Every produced UNSAT certificate is replay-checked
against the actual base CNF and assumptions before ProofReady becomes true.
Native Proof.Roots/Chains expose copies, and ValidateAgainst anchors input root
identity instead of merely validating a proof of arbitrary supplied roots.

The Go proof-save format is versioned readable JSON with a validating reader;
the absent library's unspecified binary format is not claimed compatible.
C pointer ranges become a Go slice plus an end index. Go callers can instead use
AddClause/Solve and range over Proof.Roots/Chains directly. Header iteration
macros used zero length as end-of-iteration, so they cannot distinguish an empty
input root; native proof data and replay validation handle that case explicitly.
No proof or SAT result is fabricated to conceal missing implementation source.

## Separate reviews

1. **Compactness/idiomatic/readability:** separated native solver API,
   compatibility wrappers, proof checking/serialization and elimination search.
   Go slices own roots, models and chains; there is no opaque global handle
   registry, C allocator, unsafe cast, generated draft or backend alias.
   Clauses are canonicalized once on insertion, with simple linear tautology
   checks thereafter. Root/chain IDs are derived from slice positions.
2. **Independent bug review:** inspected elimination completeness, reverse model
   reconstruction, resolution polarity, proof parent ordering/arity, ownership,
   assumption-root identity, budgets, repeated solves, stale certificates and
   deletion. A red regression exposed ReadProof accepting a second JSON object
   after the certificate; it now requires EOF. Added input-anchored certificate
   validation to reject a valid proof of a different CNF or assumptions.
   Tests cover failed-request atomicity and independent iterator/proof snapshots.

## Verification

- Full-package tests, race tests and vet pass.
- 600 unrestricted random formulas with three assumption trials are checked
  against exhaustive enumeration. Every SAT model and every UNSAT proof is
  independently checked; pigeonhole UNSAT exercises a larger refutation.
- The legacy root/chain iteration is replayed exactly as the interpolation
  consumer does, verifying root counts, earlier IDs and the final empty clause.
- Proof JSON/file round trips, malformed/tampered proofs, empty roots, proof
  mode, budgets/resumption, ownership and destruction are tested.
- A header-contract audit checks all fifteen API names; a source-header fixture
  keeps module-only tests independent of the original repository layout.
- A follow-up 3,000-query incremental oracle checks SAT models and validates
  each UNSAT certificate against the current roots and assumptions. The package
  passes `go test -race -count=1 ./sat/psat` and `go vet ./sat/psat` with
  `GOCACHE=/tmp/abcgo-cache`; finite cases do not prove all proof workloads.

There is no package build blocker. Original MiniSat algorithm/search-statistic
and binary-proof-format parity cannot be proved without the missing library;
this implementation supplies the observable logical/proof contract and records
that boundary explicitly rather than labeling a header-only source an original
implementation or pretending to have run an original-C differential test.

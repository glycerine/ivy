# Preprocessing review

Scope: SimpSolver.go and SimpSolver2.go. Original bounded resolution elimination,
backward subsumption, self-subsuming resolution, optional asymmetric branching,
redundancy checks, frozen variables and reverse model extension are implemented.

## Compactness / idiomatic review

SimpSolver embeds the core solver. Occurrence lists are slices of stable clause
pointers; no relocation offsets or separate length/capacity metadata remain.
Literal occurrence counts are a real elimination-cost cache, updated on addition,
strengthening and removal. Touched flags track pending subsumption work; the
original redundant n_touched counter is removed. Eliminated-clause reconstruction
uses structured records with the eliminated literal first rather than a packed
reverse-length integer array. Records are replayed in the original reverse order.
An explicit FIFO and pointer-set deduplication replace clause-mark mutation when
gathering touched candidates. Candidate snapshots permit safe list modification.

## Separate bug review

Random differential testing exposed an interaction in optional redundant-clause
checking: a new resolvent is implied by its parents, which are about to be
removed. Skipping the resolvent therefore destroys equivalence. The native source
reproduces the bug with (x OR a OR b) AND (!x OR a OR c), frozen a/b/c and assumptions
!a,!b,!c: native reports SAT, although exhaustive truth evaluation says UNSAT.
The checked-in test_vectors/native_rcheck_bug.cpp reproduces it. Derived replacement
clauses now bypass this check; ordinary input clauses still use it. A minimal
regression exercises exactly this four-variable case.

Substitution records now reconstruct the substituted variable's model value;
the source omitted an extension record. A regression requires x=!a and checks
x's returned value. Reset restores all preprocessing state and supports reuse.

Validation includes 1,500 random formulas (up to nine variables), exhaustive SAT
comparison, assumptions and returned-model checking; asymmetric branching and
redundancy checks are enabled on mixed trials. A separate unchanged native
SimpSolver2.cpp/Glucose2.cpp oracle performs 1,500 reproducible default-mode
preprocessing queries with status FNV64 hash818442487605194937, matched in Go.
The harness is test_vectors/native_simp_oracle.cpp; compile as CORE_REVIEW.md plus
SimpSolver2.cpp. All converted core/circuit/preprocessing explicit-file race tests
pass. Utilities and ABC integration drafts remain visible and pending.

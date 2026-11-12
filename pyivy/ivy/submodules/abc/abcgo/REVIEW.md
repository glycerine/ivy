# Package review gate

Every package, including packages already committed, requires these distinct
passes before its next integration commit:

1. Compactness and Go design: eliminate redundant state (especially stored slice
   lengths/capacities), use standard-library facilities, simplify ownership,
   remove unnecessary pointer indirection and C scaffolding, and make names,
   control flow and APIs natural and readable. Preserve necessary algorithmic
   metadata; document why retained caches or counters are needed.
2. Bug review: inspect algorithm invariants, boundaries, aliasing, overflow,
   allocation/growth, error paths, cleanup, mutation, concurrency and API consumers.
   Fix every finding and add behavior-based regression coverage.
3. Verification after fixes: run full package tests and relevant consumer tests;
   run race/differential checks where applicable. A passing test is evidence,
   not a substitute for either review pass.

Each worker must report the files reviewed, findings (or no findings), fixes,
regression tests, and commands/results for both passes separately. Main-thread
integration requires that evidence, and commits only one green subpackage at a
time with a single-sentence message under 60 characters on the current branch.

All twenty original workers have been asked to revisit completed contributions
and apply this gate to ongoing work. Review reports are pending; earlier commits
do not establish that these new review requirements have been satisfied.

## Findings in progress

- GIA: removed a duplicated object-store slice header; added concurrent handle
  pairing and malformed-object/AIGER encoding regressions. Full package port
  remains incomplete, so component tests do not establish package readiness.
- EPD/BBL: numerical underflow and padded-record serialization fixes are under
  review; require worker sign-off and regression/race results before committing.
- BDC/MVC: derive redundant word/count metadata and avoid copying hash chains
  on insertion; coordinate exported API changes with decomposition consumers.
- Integration harness: separate reviews documented in integration/REVIEW.md;
  added full SAT-model/circuit consistency checks, committed after race tests.

## Ownership cleanup idiom

Recognized ABC_FREE expansions that test an lvalue, free its allocation and set
that same lvalue to zero should become a direct nil assignment in Go. Example:
`pGia.Refs = nil`. Keep pointer/slice field names consistent with their actual
ported declarations. Nil assignment is safe for an already nil value. Go GC
reclaims storage only when no aliases retain it; this is not a substitute for
Close on files/OS resources, and unrelated free expressions need ownership review.
The GIA owner is applying this rule across its package with review/regressions.

# ivy2cpp TODO3: whole-hog Python parity completion report

Created: 2026-05-23
Completed: 2026-05-23

This file records the completed TODO3 pass after AUDIT2 items 031-046
were implemented and the Python/Go oracle harness was added.

The oracle harness is live:

```sh
env XTRACE_OFF=1 ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle -v -count=1
```

Current result: all 14 oracle fixtures are marked `PASS` in
`test_vec/oracle/STATUS.md`, and `EXPECTED_FAIL` is no longer accepted by
the oracle status policy.

## DONE TODO3-001 - Make the Go runtime scaffold token-equivalent to Python

Implemented:

- Runtime header/implementation preambles now follow Python token order.
- Go-only `#pragma once` and other local scaffold tokens were removed.
- `ivy_threads.hpp`, `ivy_value.hpp`, `ivy_repl.hpp`, and `ivy_go_z3.hpp`
  now land in the Python-equivalent slots.
- Class skeleton order now matches Python, including `___ivy_choose`
  before assertion hooks and `_generating` across generated targets.
- The Go oracle harness now passes absolute fixture paths to match
  Python's source-location strings in emitted assertions.

## DONE TODO3-002 - Promote oracle fixtures one by one after runtime parity

Implemented:

- All 14 fixtures in `test_vec/oracle` were promoted to `PASS`.
- Codegen was aligned for assignments, quantified updates, bitvectors,
  enums, ranges, destructors, variants, hash thunks, native blocks,
  callback thunks, progress/rely output, and multi-isolate extraction.
- The progress fixture was adjusted to read the state it writes, avoiding
  Ivy isolate erasure that caused both Python and Go generated C++ to
  fail the slow compile gate.

## DONE TODO3-003 - Turn the oracle status policy into a hard gate

Implemented:

- `EXPECTED_FAIL` now fails the suite instead of being accepted.
- `testdata/oracle` references were removed/avoided; the fixture path is
  `test_vec/oracle`.
- Fast and slow gates were run:

```sh
env XTRACE_OFF=1 go test ./ivy2cpp -count=1
env XTRACE_OFF=1 ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle -v -count=1
env XTRACE_OFF=1 ORACLE_TEST=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestOracleCompileGo|TestOracleCompilePython|TestOracleSemanticEquivalence' -count=1 -v
```

Results:

- Fast regression: pass.
- Token oracle: pass=14.
- Go compile oracle: pass=14.
- Python compile oracle: pass=14.
- Semantic equivalence harness: pass with 14 skipped subtests because no
  `.in` transcripts exist yet.

## Remaining Follow-Up

- Add `.in` transcripts for exported-action REPL fixtures so semantic
  equivalence executes real Go/Python binary comparisons instead of
  reporting `no transcript=14`.

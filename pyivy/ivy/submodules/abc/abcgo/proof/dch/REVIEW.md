# DCH integration gate

## Compactness, idiom and readability review

Reviewed the snapshot, simulation, class, CNF, SAT, counterexample, sweep and
choice modules as separate responsibilities before the bug audit.

- Replace scratch `pData` casts, pointer arrays and tagged addresses with typed
  graph edges, local maps, slices and errors. The public adapter depends only on
  AIG and ported BSAT; DAR/CEC remain upstream and introduce no cycle.
- Keep the source's eight 32-bit simulation rounds and random stream, including
  the discarded reset draw and all-zero assignment bit. Use an independent RNG
  instance instead of resetting process-global random state.
- Normalize phase consistently in signatures, CNF, SAT miters and rebuilt
  choices. Explicit proof-result values distinguish a refutation from a timeout.
- Build clauses lazily by actual cones. Retain source soft recycling thresholds;
  do not replace SAT with bounded truth enumeration or silently accept timeouts.
- Centralize metadata copying with the existing AIG clone contracts. Clone timing
  once after final DFS, rather than once before and again during duplication.
- Avoid copying Graph.Nodes just to determine array sizes. Simulation words use
  one contiguous allocation; singleton classes avoid redundant signature hashing.
- TFO counterexamples use a standard-library priority queue and source-order
  ranks. Graphs whose object IDs are not topological are explicitly rejected.

## Independent bug review and fixes

1. Found missing trailing-register/constraint metadata at the snapshot boundary.
   Added interface validation and independent copies of terminal arrays, clock
   domains and Name/Spec. Exhaustive sequential-interface tests treat RO values
   as free inputs and retain the last RI/RO positions.
2. Found redundant timing cloning across choice construction and final DFS.
   Moved metadata cloning to the final graph; a clone-count regression requires
   exactly one clone and verifies mutable arrays do not alias input snapshots.
3. Structural graph checking alone does not detect cycles formed through choice
   links. Added combined logic/choice DFS, alternate ownership/ref checks and
   optional exact structural-support validation. Actual TFI-loop and duplicate
   support cases exercise the filters.
4. Verify only proved representatives are published; simulation-only candidates
   remain private and classes split on real counterexamples. A 30-input rare
   minterm challenges an apparent constant and is preserved, including all-ones.
5. An eight-input parity comparison genuinely exhausts a one-conflict SAT budget;
   it remains unknown and no unproved choice is accepted. Multiple independent
   constant proofs force actual solver recycling, rather than only initialization.
6. Polarity and both full/TFO refinement settings preserve outputs. Incremental
   TFO values match full simulation over 100 random assignment changes. A SAT
   counterexample that fails to refine now returns an error rather than looping.
7. Explicit unknown SAT statuses cannot default to success. CNF contradictions
   and invalid representatives return errors. Simulation dimension overflow is
   rejected before allocation; no arbitrary size cap is introduced.
8. Exhaustively verify every published representative and every choice edge,
   including complemented phases, in 25 deterministic random five-input snapshot
   families. Check outputs, chain acyclicity and graph invariants independently.
9. Cover empty/constant/direct-input interfaces, constant/CI aliases, rejection of native XOR and
   non-topological IDs, parameter errors and unchanged input copies.

## Independent source oracle

`test_vectors/dch_oracle.c` links the original ABC implementation, including its
real SAT solver and allocators. The fixture contains associative snapshots and
redundant-support alternatives with the filter disabled/enabled. Exact total and
result AND counts, choice counts and every output assignment match Go. No C
compiler/runtime is involved in the ordinary Go tests.

The SAT backend uses the ported BSAT solver from the original DCH path.
Source-style supergate and MUX clauses, incremental proved implications, and
the conflict budget are retained. Upstream
synthesis/GIA/new-engine selection flags remain upstream controls as in old DCH.

Validated with DCH vet and full-package race tests for DCH, AIG and BSAT.
No stubs, excluded source files, parser-only success claims or Git operations.

# Satoko cost adapter review

The production pipeline is pure Go: IOA model normalization, source MUX/XOR
coarsening with limit three, MF eight-cut CNF mapping, source clause ordering,
and native Satoko. `EvaluateModel` matches CmdAuto's evaluator signature. The
named `MFBackend` and injectable Backend interface retain explicit provenance.

## Readability review

- The solver runner separates encoding from solve lifecycle and reports native
  status, per-query conflict cost and elapsed time in a small typed result.
- Clauses are inserted in source order, then options configured, then simplified
  and solved, matching Gia_ManSatokoInit/Create/CallOne.
- An early contradiction returns UNSAT with zero cost, as the original driver.
- No CEC/Tseitin encoding substitutes for the required source MF mapped CNF.

## Bug review and verification

The independent C driver initially exposed cost discrepancies even when SAT
status agreed. The Satoko owner corrected actual source behavior: implicit
variable polarity, reverse cancellation reinsertion, binary watcher priority,
blocker caching, activity/minimization order, root simplification heap building,
literal counters and persistent root assignments during clause insertion. No
invented cost offsets were used.

`SATCOST_C_DIFFERENTIAL=1 go test -race ./aig/gia/satcost -run TestOriginalCConflictCost -count=1`
passes all 100 seeded formulas against the original C solver's conflict counter.
The MF package independently matches original ABC's complete variable/clause
ordering on 34 networks and original C cover ordering on wide truth tables.
Native tests check early contradictions, actual search costs, malformed CNF and
explicit missing-backend errors. Full MF/Satcost races and vet pass.

This evidence covers the audited cases. Arbitrary restart/reduction settings
and incremental search-trace identity remain unproven; the native backend stays
injectable and named rather than asserting universal original-search parity.

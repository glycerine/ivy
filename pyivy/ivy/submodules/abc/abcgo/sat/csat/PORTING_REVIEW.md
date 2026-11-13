# CSAT native port review

## Scope

The C API managed an ABC network and later built an AIG target for ABC's CEC/SAT
engines. The Go package now stores a validated, named Boolean network and emits
its gate equivalences directly to the repository's MiniSAT implementation.
Default mode uses the simplification solver; `ABC_UseOnlyCoreSatSolver` selects
the core solver. `ABC_AddTarget` constrains all named signals conjunctively.

## Bug and readability pass

- Gate creation validates all names, fanin counts, and gate arities before it
  mutates the graph. This avoids the partial nodes the C implementation could
  leave behind after discovering a missing fanin.
- Gate IDs and `GateType` remain distinct typed values. All owned slices are
  copied at API boundaries, including target inputs and returned result data.
- Unary XOR/XNOR, constants, complemented gates, and PO aliases have explicit
  CNF equivalences. Truth-table tests cover AND/NAND/OR/NOR/XOR/XNOR for every
  input pair and requested output value, plus unary/constant and mode parity.
- Source comparison (`src/sat/csat/csat_apis.c`, `ABC_AddGate`) confirms these
  are the only implemented gate constructors: `CSAT_CONST`, `CSAT_BPI`,
  `CSAT_BPPI`, `CSAT_BAND`, `CSAT_BNAND`, `CSAT_BOR`, `CSAT_BNOR`, `CSAT_BXOR`,
  `CSAT_BXNOR`, `CSAT_BINV`, `CSAT_BBUF`, `CSAT_BPPO`, and `CSAT_BPO`. Go tests
  cover each boolean family and the source-allowed input arities.
- The C enum also declares `CSAT_BMUX`, `CSAT_BDFF`, `CSAT_BSDFF`, `CSAT_BTRIH`,
  `CSAT_BTRIL`, `CSAT_BBUS`, `CSAT_BCNF`, and `CSAT_BDC`, but none has a
  construction case in `ABC_AddGate`. The C default branch leaves `pObj` nil;
  `CSAT_BDC` also receives no SOP function. Tests verify these declared values
  are rejected with an error instead of assigning invented semantics.
- The C manager's `ABC_SolveInit` strashed the whole network while the target
  network creation was commented out. The Go port instead asserts each target
  explicitly, so target values actually affect SAT results.
- Resource setters now apply solver budgets and elapsed-time interruption.
  Results report timeout when the solver returns unknown under those budgets.
- The BENCH writer emits a constant as `name = 1` and propagates file errors
  into `LastError`; tests exercise its output using a temporary directory.
- Integrity checks walk backward from POs and detect unused internal gates.
  Release is idempotent, and result getters return caller-owned copies.

## Follow-up review and tests

The implication limit now sets a propagation budget independently of the
conflict limit; the stricter remaining cumulative limit also applies. Result
allocation provides indexable name and value slices, and BENCH export retains
named output aliases. Regressions cover these cases. A seeded oracle checks 100
networks, all input assignments, both target polarities, repeated targets, and
both solver modes against direct Boolean evaluation.

`GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/csat` and
`GOCACHE=/tmp/abcgo-cache go vet ./sat/csat` pass. These tests do not establish
equivalence to the legacy Ivy heuristics or cover every resource interaction.

## Known limits

- Mode 0 uses MiniSAT preprocessing rather than ABC's Ivy proof engine. Both
  modes implement the same target satisfiability semantics, but Ivy's staged
  rewriting/proof heuristics and the legacy CEC counters are not reproduced.
- The legacy API has a single current target/result and ignores the target ID;
  the Go getter keeps that behavior and returns an owned copy.

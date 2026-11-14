# Exact solver core source review

Reviewed against `src/base/abci/abcExact.c`, from `Ses_ManComputeTopDec` through `Ses_ManFindMinimumSize` (the assigned solver core range).

## Source correspondence

- Manager setup preserves ABC's four-word truth storage, constant-zero phase normalization, arrival-time normalization, top-decomposition flags, and cleanup restoration.
- Variable accessors use the C offsets and triangular selection-variable indexing. The tests pin representative simulation, output, gate, selection, and depth IDs.
- `Ses_ManDeriveTruth` decodes the compact gate encoding in order and evaluates the three normalized truth-table rows used by each two-input gate. It applies output complementation only when requested, as the C routine does.
- Clause generation keeps the C source ordering and its constraints: selected truth rows, output selection, one operand pair per gate, AIG and binary gate clauses, decomposition assumptions, gate-use clauses, repeated-operand exclusions, co-lexicographic ordering, and symmetric-variable ordering.
- Depth clauses and consistency checks retain the source arrival-time propagation, monotone depth encoding, decomposition checks, and gate-count bounds.
- CEGAR starts with the source's random row sample, solves, derives the candidate truth table, adds the first counterexample clause, and repeats. Sampling uses the existing glibc-compatible generator seeded with `0xCAFE` for each manager, with the same modulo operation as `rand() % nRows`.
- Minimum-size search retains source bottom-up/top-down status handling, resource-limit fallback, and the direct multi-output search path.
- Network extraction preserves the compact solution fields and builds both SOP and GIA representations from the same gate and output literals.

## Focused checks

`abcExact_solver_core_test.go` covers variable index layouts, XOR truth derivation and complemented outputs, caller truth/arrival restoration, single-output truth assumptions, minimum XOR network search, and the seeded glibc sampling sequence.

`gofmt` and `git diff --check` pass for the solver file and tests. Package test execution is currently blocked before type checking by parser errors in the separately owned `abcFpga.go`; rerun `go test ./base/abci` after that tranche is repaired.

## Ownership and representation notes

Go's garbage collector replaces C's explicit frees for temporary solution buffers and local clause slices. SAT state remains owned by `Ses_Man_t` and is reset between gate-count attempts. The per-manager random state reproduces the C seed/draw sequence for a normal single active synthesis manager; C itself exposes a process-global RNG, so interleaving multiple managers differs from that global side effect.

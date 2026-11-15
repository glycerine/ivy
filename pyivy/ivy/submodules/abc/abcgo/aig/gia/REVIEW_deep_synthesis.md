# Deep synthesis driver review

The two entry points in `giaDeep.go` follow `src/aig/gia/giaDeep.c`:
randomized synthesis scripts, strict best-graph replacement, no-improvement
stopping, and independent restarts from an owned initial graph. This review
covers driver control and ownership, not correctness of every synthesis engine
called by those scripts.

## Source agreement

`deep_synthesis_native_review_test.go` compiles the actual C driver with scripted
frame/command boundaries and real graph allocation/duplication. Across 96
configurations it compares 1,968 commands, complete command strings, LUT-size
cycles, random state after completion, batch mode, graph counts, best selection,
and restart behavior. Cases include ties, regressions, intermittent improvements,
seeds, both LUT ranges, both initial batch states and multiple outer trials.
Independent truth-table checks confirm returned graphs compute the intended
conjunction. Native tests require the original ABC objects and C compiler; they
skip when that environment is unavailable.

`deep_synthesis_integration_test.go` additionally runs the actual synthesis
engines through the ordinary dispatcher and standard `abc.rc` aliases, both
directly and through registered `&deepsyn`. Redundant conjunctions reduce to one
AND, all four input assignments retain their values, and the direct result
survives replacement of the frame graph. This small end-to-end example complements
the scripted control matrix; it is not a large-circuit quality benchmark.

## Bugs and repairs

Red regressions preceded three repairs:

- A command panic left temporary batch mode enabled. A small named helper now
  restores it with `defer`, while preserving an initially enabled batch mode.
- Timeout accounting began after the initial graph copy. Timing now starts at
  driver entry. A private injected clock makes copy-cost and deadline tests
  deterministic without mutating a global clock.
- Zero outer trials created a global frame unnecessarily. They now return one
  owned duplicate directly, also avoiding an unused second duplicate.

Tests also cover failures in first/later commands and later outer trials,
missing/wrong-type/typed-nil or removed frame graphs, quality goals, equal-cost
metadata retention, no-improvement boundaries, verbose reporting, precise
simulated deadlines and the full 100,000-iteration limit.

## Explicit source differences

The original C never updates `nAndsMin`, so its advertised quality goal cannot
stop synthesis and its final message may report -1. Retain the existing Go
repair that checks the actual best graph and stops at the CLI's documented goal.
Also retain existing Go cleanup/restoration on command failure and outer-loop
handling of a failed trial, rather than reproducing C leaks or null dereferences.

The portable Go budget uses elapsed wall time and an inclusive deadline; C uses
CPU clock ticks and a strict comparison. Both check budgets after a complete
command, so neither provides cancellation inside a synthesis command. Native
normal-control comparisons disable timed stopping; separate fake-clock tests
verify Go's stated deadline convention. Panics propagate after batch restoration.

Integration results are recorded in `../../SEQUENTIAL_AUDIT.md`.

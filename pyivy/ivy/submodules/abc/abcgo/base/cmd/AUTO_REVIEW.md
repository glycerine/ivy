# Autotuner

## Readability and compactness pass

The parser uses typed parameter rows instead of float bit patterns packed into
integer vectors. Settings contain both immutable Satoko options and their
printable command line. Cartesian expansion and bounded evaluation are separate
functions. Worker results are summed under a mutex; errors propagate to the
caller. The best result retains the first setting on equal costs.

## Separate fidelity and bug pass

Preserved the mandatory first row, one-of choices on subsequent rows, optional
single lowercase switches, source float32 input rounding and three-decimal
argument formatting. All source Satoko option letters map to their corresponding
Go fields, including percentage conversions for learnt and garbage ratios.
Benchmark files are real IOA AIGER models. Failed individual reads print a
diagnostic; no readable benchmarks is an error rather than a meaningless win.

Tests cover Cartesian/toggle combinations, all common option fields, malformed
configuration, concurrent costs, backend errors, missing backend and overflow,
and an entire config/file-list/best-setting run over an actual AIGER file.
The separate bug review fixed cost overflow even after an earlier worker error
and selection when the minimum cost equals MaxInt64.

## Exact SAT-cost integration boundary

`AutoEvaluator func(*ioa.Model, satoko.Options) (int64,error)` defaults to
`aig/gia/satcost.EvaluateModel`. The actual MF8 pipeline preserves source cut
mapping, coarsening and clause order, with native Satoko solving. GIA/Satoko
owners report 34 complete CNF source-oracle comparisons and 100 status/conflict
cost comparisons; broader restart/reduction/incremental traces remain outside
that audit coverage (see satcost/REVIEW.md). No claim of universal trace identity
is made. Explicit alternatives remain injectable through `SetAutoEvaluator` or
frame resource `cmd:auto-evaluator`; a nil evaluator still errors.

A command-level regression writes a real AIGER AND cone, invokes `autotuner`
through the production dispatcher with two choices and concurrent workers, and
checks completed evaluation/selection using the actual default backend. Both
separate review passes confirmed no duplicate solver/CNF implementation and
unchanged injection semantics.

Full `go test -race ./base/cmd` and `go vet ./base/cmd` pass. The optional Windows directory maintenance API group is ported separately in files.go, retaining Windows-only registration.

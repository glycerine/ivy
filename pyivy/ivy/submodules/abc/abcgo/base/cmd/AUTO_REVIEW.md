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

`AutoEvaluator func(*ioa.Model, satoko.Options) (int64,error)` is installed with
`SetAutoEvaluator` or frame resource `cmd:auto-evaluator`. The source objective
is conflict count from MF mapped CNF with 8-input cuts and one OR-of-outputs
clause. A gate-CNF substitute changes that objective and is deliberately not
provided. GIA owner is implementing `aig/gia/satcost.EvaluateModel` with this
exact signature. Until installed, invocation returns an explicit missing-service
error. The real parser, search, concurrency and selection algorithms are tested;
the actual MF evaluator is a remaining integration dependency, not claimed done.

Full `go test -race ./base/cmd` and `go vet ./base/cmd` pass. Windows-only directory
maintenance commands from cmd.c are outside this Linux command set and remain
to be ported if the Windows API group is required.

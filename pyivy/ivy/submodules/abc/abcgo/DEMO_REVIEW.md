# Library demo port review

Source: `src/demo.c`. `DemoMain` is the command entry path; `cmd/abc-demo`
is its thin executable wrapper. `RunDemo` permits an embedding program to use
an initialized frame with its real network commands registered.

## Compactness and readability pass

The source's repeated command/error checks share one local helper. Command order,
resyn/resyn2 scripts, optional statistics/verification, fixed result.blif output,
and diagnostic/timing text remain explicit. Options are only the three original
source switches. No synthesis algorithm or missing command is imitated.

## Separate bug review

Checked argument validation before initialization, stop-on-first-command-error,
all original command sequences, and cleanup. Deliberate source corrections:
framework cleanup also occurs on early failures; dynamically sized commands
avoid the fixed C buffer; filenames are quoted for ABC's own parser so spaces,
quotes, semicolons, comment markers and backslashes remain one argument. Timing
uses Go's monotonic wall clock instead of C process CPU time.

Regression tests use the actual command dispatcher and registered recording
handlers, exercising default/resyn2 order, all 12 default failure positions,
quoted filenames, diagnostics, lifecycle cleanup, and invalid arguments. These
are workflow tests, not substitutes for the network algorithms: those modules
must supply read/balance/rewrite/refactor/write_blif/cec implementations.

Validation now passes for the complete root package and both CLI wrappers:

    go test -race . ./cmd/abc-demo ./cmd/abc-starter ./base/test
    go vet . ./cmd/abc-demo ./cmd/abc-starter ./base/test

## Typed autotuner backend boundary

The demo invokes the command interpreter directly; it does not replace GIA or
SAT algorithms. The command package's separate autotuner accepts
`cmd.AutoEvaluator func(*ioa.Model, satoko.Options) (int64, error)` through
`cmd.SetAutoEvaluator` or a frame resource named `cmd:auto-evaluator`. This is the
explicit boundary for the original `Gia_ManSatokoCallOne` cost computation. Its
implementation must preserve MF mapped-CNF construction and the SAT conflict
count objective. An absent evaluator produces an explicit error; no synthetic
cost, alternate mapper, or successful no-op is supplied by this port.

## Empty test extension hooks

`base/test` faithfully preserves the two empty functions in `src/base/test/test.c`.
Readability review keeps these hooks empty and documents nil acceptance. Separate
bug review confirmed no command registration, frame access, or state to port.
Compile-signature and nil/non-nil invocation tests pass under `go test -race
./base/test`; `go vet ./base/test` also passes.

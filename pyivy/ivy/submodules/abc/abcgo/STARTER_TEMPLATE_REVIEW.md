# Starter and template port review

Sources: `src/starter.c` and `src/template.c`. All original executable behavior
is represented; neither source contains an ABC frame initialization routine.
The starter launches external shell commands; the template allocates a manager,
borrows two pointers, frees the manager, and returns 1 without touching the AIG.

## Compactness and readability pass

`StarterMain` handles arguments/files/diagnostics; `RunStarter` owns bounded
parallel scheduling. `Abc_RunThread` uses the actual platform shell, preserving
system() redirection/pipeline/quoting behavior. One synchronized writer protects
concurrent child output and diagnostics, including caller-provided buffers.
Channels and a wait group replace global thread counters, mutex polling and
pthread bookkeeping. A thin `cmd/abc-starter` wrapper supplies CLI arguments.

The template keeps its original names and parameter layout. Its unused borrowed
AIG handle is opaque (`any`) so callers can supply their existing representation
without an artificial AIG dependency. Allocation retains pointer identity;
free delegates storage reclamation to Go without modifying the borrowed data.
`Xyz_ManPerform` returning 1 is the actual source operation, not an algorithm stub.

## Separate bug review

Verified original line filtering (leading newline/CR/tab/space/#), trailing CRLF
removal, command order at launch, the 1000-command cap, shell failure diagnostics,
and successful launcher status despite individual command failures. CPU requests
reserve one controller slot as in C, with these deliberate source corrections:
a one-CPU request gets one worker instead of spinning forever; bounded channels
sleep instead of busy-waiting; all workers are joined even on read failures;
long input lines remain whole instead of executing truncated 999-byte fragments.
Integer parsing rejects malformed values rather than atoi's silent prefixes.
Timing uses monotonic wall time, meaningful for parallel child processes.

Regression tests cover the single-CPU path, bounded multiworker execution,
concurrent output to the same buffer under the race detector, skipped lines,
long/final unterminated lines, command limits, individual failures, read errors,
usage errors, and an actual shell subprocess workflow. Template tests verify
borrowed identity, unchanged input objects, nil handling, and the original result.
Root/demo/CLI race tests and vet pass; no source files are hidden or excluded.

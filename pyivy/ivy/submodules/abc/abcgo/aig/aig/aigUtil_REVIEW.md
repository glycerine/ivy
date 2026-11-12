# AIG utility port review

## Pass 1 — representation/readability

All 47 source helper names are implemented. Graph ownership and explicit Edge
values replace object-address/tag-bit manipulation. Collections and snapshots
use native typed slices/maps. Mux/XOR recognition returns typed results rather
than output-pointer parameters. Printers use local builders and io.Writer;
BLIF/Verilog dumping uses local net names without overwriting graph node data.
Cio IDs are computed snapshots; cleanup clears the snapshot rather than unrelated
traversal links. Timestamp, file paths and file I/O use the Go standard library.
An explicit random-stream type supports reproducible independent consumers; the
legacy shared-stream entry points serialize access with a mutex. Counterexample
simulation is graph-owned typed frame data with explicit start/stop lifetime.

## Pass 2 — correctness

Separate semantic tests exposed complemented-root mux recognition and reference
reset bugs, which were fixed before integration. Tests now cover all three
control-variable orderings and all eight input phase combinations, both root
polarities, and all input truth assignments. XOR recognition preserves actual
edge polarity. Ref resets include graph-owned output roots.

Additional review fixes and regressions cover traversal-ID rollover without
clearing unrelated typed data, cycle-safe supergate collection, XOR expression
parity with repeated support, non-topological phase calculation, output-phase
updates after constraint inversion, useful zero-value RNG state, and writer
short-write propagation. Deterministic RNG values are checked against literal
ABC reference values. BLIF output is independently interpreted across all input
assignments and compared with Graph.Evaluate, including mux, native XOR and
constant outputs. Sequential counterexample tests cover register transfer,
validation failures, lookup and cleanup; failed starts leave no partial session.

The converted-file AIG suite, including all utility tests, passes with the race
detector. Full-package compilation remains dependent on other workers' remaining
AIG source files; no source files were hidden or disabled for validation.

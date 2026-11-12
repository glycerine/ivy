# Upsizing review

## Compactness/readability pass

The full sclUpsize.c algorithm family is ported in sclUpsize.go: TFO collection,
critical-output/node windows, local recomputation/evaluation windows, cell trials,
gain-ranked nonadjacent updates, iteration cooldown, buffer/inverter bypasses,
best-mapping commits, topology undo, dangling-node cleanup, diagnostics and the
internal/public drivers. The upsize command parses its original numeric/toggle
options and calls the real manager implementation.

The shared SC_Man owns assignments, timing, loads, update pairs and iteration
cooldown. Critical-path marks and temporary gain queues use native maps/slices;
no second network model is introduced. Mutable operations are optional typed
interfaces implemented by the real host graph. Pin indices make duplicate-edge
multiplicity explicit. Requesting bypasses on immutable topology returns an error.
Phased graphs toggle edge polarity; unphased graphs retain the source's paired-
inverter requirement. Barriers are traversed transparently and excluded as cell
replacement candidates, preserving source wrapper behavior without duplicating
an entire network just to remove temporary barrier objects.

## Independent source/bug pass

Compared all original functions and driver stages against sclUpsize.c. Trials
restore assignment, cone times/slews and fanin loads. Candidate search retains
notch limits and the four-nonimproving-candidate stop rule. Updates retain gain
order, ratio limits, neighbor conflict exclusion and five-iteration cooldown.
Best mapping snapshots commit accumulated update pairs; topology edits since the
last improvement are undone in reverse order. Accepted dangling bypass buffers
are then removed and full timing/MIO assignments are recomputed.

Found a source defect in the no-departure single-stage path: TFO collection starts
only from internal fanins, so a gate driven by a CI never updates its output.
Observed a red regression where a3ps→1ps cell improvement was discarded; fixed by
including each path pivot as a TFO root. Zero critical-window input is rejected
rather than entering the original window-doubling infinite loop. Diagnostics and
adapter failures propagate errors. Runtime reporting uses Go duration values.

## Validation

Durable explicit-file manifest: `test_vectors/timing-files.txt` now includes
sclUpsize.go and upsize_test.go. Tests cover departure and non-departure timing,
single-stage propagation, trial-state restoration, missing mutable adapter,
invalid window, real mutable bypass/undo, rejected patch restoring loads,
inverter bypass polarity/undo, dangling cleanup and the actual command mapping.
An independent original ABC executable reads the same generated technology
library and mapped two-cell chain; its upsize command selects the same two larger
gates as Go. This complements the original-C interpolation, full SCL-byte-stream,
and downsizing oracles.

Full-package race tests and vet now pass after the separately owned buffer
source files landed. The buffer owner is adding its own regression suite and
review; concrete ABC topology/mapping adapters remain owned by base/abc.

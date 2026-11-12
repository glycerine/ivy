# Buffer topology and load-based sizing

This port owns `sclBuffer.go` and `sclBufSize.go`, their tests and the buffering vector fixture. Command registration/handlers are separately owned in `scl.c.go` and `commands_buffer.go`; timing/library foundations remain shared.

## Implemented algorithms

The topology family implements unary-chain removal, phase-bit extraction/materialization, reverse-level degree-limited buffering, optional root isolation, and the older critical/noncritical fanout heuristic. The latter retains critical-edge separation, gate duplication and balanced buffer/inverter branching, with a fanout-priority heap and deterministic ID tie breaks. Scaled delays use actual MIO pin delays when present; otherwise they use the existing SC timing estimator. A buffer changed into an inverter receives its actual new mapped gate before further timing work.

The load-based family processes gates in reverse topological order, computes individual fanout-pin capacitances and estimated departure times, sorts by departure/capacitance, groups low-departure fanouts under the target gain and degree, and chooses unary and logic cell sizes from the actual equivalent-cell rings. Buffer mode squares the requested gain as in C. Output load and PI-driver defaults use the existing timing foundation; optional wire loads use the shared interpolation/extrapolation implementation. Barriers transmit load/departure information without cell lookup or sizing.

Public wrappers clone the concrete host graph before editing, clean dangling logic and optionally request its topological clone. The input graph remains unchanged on either success or failure. MIO-backed mappings are synchronized through the existing conversion/host commit mechanism. Hosts advertise node creation, cloning and phase access with typed optional interfaces; absent capabilities return an error. A concrete ABC-to-SCL adapter remains separate integration work, rather than being replaced by a different graph type.

## Independent compactness/readability pass

C object pointers and attached global manager pointers are replaced by explicit manager plus stable node/pin IDs. One `FanoutPin` identifies each consuming edge, including repeated fanins with different pin capacitances. Maps replace spare-capacity/offset arrays and manual vector lifetimes; topology/phase routines keep local traversal state instead of altering host marks/levels. Shared library functions handle cells, truth tables, timing estimates, wire loads and MIO assignment.

The older priority queue uses the standard heap with versioned updates and deterministic ordering. The implementation recomputes its scaled timing after a structural change instead of reproducing the C incremental-cone cache; cone collection APIs remain real traversals. This trades additional timing work for a single authoritative timing state, particularly with sparse or reordered host IDs. It preserves decision formulas but can change scheduling in ties and resource/performance behavior. No product C execution, temporary reconstructed C graph, hidden source or no-op engine is used.

## Separate source/bug pass

Reviewed every branch against the two C sources, including inverter parity, phase propagation through unary chains/barriers, fanout grouping, complete pin loads, cell sizing and mapping synchronization. Regressions cover the findings:

- Duplicate fanins are patched by pin index. Their capacitances and timing arcs stay distinct, repairing C's per-consumer-node scratch ambiguity.
- Equal fanouts compare equal (with deterministic ID/pin ties), fixing C's comparator that returned -1 for identical objects.
- A forced minimum group of two fanouts recomputes the complete load. C could select two but retain only the first edge's accumulated load.
- `BufferPIs` actually controls PI buffering in both engines; the C load-based loop did not read that flag, and the older initial PI queue was commented out.
- Inverter insertion requires readable phase metadata. It returns an error instead of producing a knowingly unverifiable graph. Compensation propagates through transparent unary nodes and barriers to the consuming pin.
- Gate replacement updates SC and MIO views together. Recomputing timing cannot accidentally use a stale MIO buffer delay after replacement by an inverter.
- Result nodes are created before rerouting users, avoiding transient unconnected gates. Group degree is at least two; growth bounds repair the source's capacity assertions/low-gain failure paths with errors.
- Missing timing arcs, malformed tables, nonfinite/negative estimates, malformed graph endpoints and unsafe shared clone mappings return errors instead of panicking or corrupting the source.
- Full timing is recomputed before verbose reports inspect newly inserted edges.

Host cloning must own its topology, mapping slices and phase data. Immutable library/cell objects remain shared. Optional `SclCloneTopological` preserves the concrete graph representation rather than replacing the command frame with an unrelated generic graph.

## Validation

Normal complete-package gates pass:

- `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/scl`
- `GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/scl`

Tests use `test_vectors/buffering_cases.json` plus real mutable mapped-network fixtures. They exhaustively compare Boolean outputs before/after degree trees, inverter trees, unary removal, phase round trips, critical duplication, load/gain sizing, barrier paths and PI controls. A registered command test runs `buffer` followed by `unbuffer`, checks real node insertion/removal and verifies ownership and behavior. Error tests check the original graph remains intact. The durable timing-file manifest includes the new tests.

`TestLoadGainBufferingAgainstOriginalABC` writes a complete Liberty library and mapped BLIF, runs the actual repository ABC `buffer -N 3 -G 300 -S 10`, and requires C and Go to insert the same number of mapped stages. ABC's CEC engine proves their results equivalent. The test skips only when the original executable is unavailable; it executed and passed in this workspace.

The initial reference fixture mixed mapped `.gate` records with `.names` output aliases. The original binary crashed while reading that fixture, before invoking buffering. The final fixture uses fully mapped, distinct output gates and stable port names; the normal package test passes without an environmental exception or ignored subprocess failure. All reference artifacts use the test's temporary directory. No tests run on experiments.

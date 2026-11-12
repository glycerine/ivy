# Source GIA coarsening port

Scope: `Gia_ManDupMuxes` in giaMuxes.c, its recognizers in giaUtil.c, the AND/XOR-real/MUX-real hashing rules in giaHash.c, and mark-used/ordered cleanup in giaScl.c/giaDup.c. Implemented in `coarsen.go` against the neutral MF Network interface, without importing the unported parent GIA package.

`Coarsen(network, limit)` returns owned storage implementing Network, Sibling and MuxesEnabled (always true, including a result containing no MUX). Input is a plain AIG with canonical ordered CI/CO interfaces. Zero limit still recognizes XORs. Already-real XOR/MUX inputs are rejected, matching this pipeline's plain-AIG precondition. Outputs remain driver literals while explicit CO objects stay in the object stream; direct-output-only interfaces are materialized as CO objects.

## Separate design and readability pass

Source object order, fanout-reference calculation, recognizer precedence, and cleanup ordering remain visible. One literal-copy map connects source IDs to hashed IDs; choices use explicit regular IDs. A small typed structural-hash key replaces hash-table plumbing while keeping each original simplification and real-gate canonicalization rule. Cleanup marks through COs, every buffer fanin and sibling alternatives, then copies in original object order. CI and buffer retention matches source cleanup.

## Separate bug pass

Checked MUX control/then/else conventions, recognizer branch precedence, complemented XOR phases, real-MUX data/control canonicalization, original reference-count threshold, sibling barriers and remapping after structural merging, dead logic cleanup, retained dangling buffers, copied interface ownership and topological ordering.

Two malformed-interface cases were first reproduced by red tests: reordered CI lists changed variable interpretation, and explicit CO objects inconsistent with output lists selected a different function. Coarsening now rejects both before construction. Input output slices are copied before normalization, so a provider returning borrowed storage remains unchanged.

A deliberate semantic safeguard concerns a simple-mode AND with the same variable in opposite phases: C Gia_ObjIsBuf tests equal regular IDs and can misclassify x AND !x. The neutral port preserves its false Boolean meaning through hash-AND simplification; equal exact literals retain source buffer treatment. Production strashed AIG inputs exclude this degenerate case. No replacement CNF encoding is introduced.

## Validation

- Exhaustive MUX and XOR input phases, limit zero through three, truth over every input assignment.
- Exhaustive real-MUX hash identities over all constant/complemented three-variable operands; control-dual canonicalization and symmetric XOR hashes.
- Shared-child reference thresholds, sibling absorption barriers, buffer-root retention, source slice ownership and malformed interface regressions.
- Twelve seeded random DAGs compared against the original ABC executable using `&r -s input.aig; &st -m -L 3; &ps -m`: resulting AND/XOR/MUX counts agree, with independent exhaustive Boolean equivalence on the Go result. Original ABC fixture is skipped when the executable is unavailable.

The production path is pure Go. Full MF race tests and vet are the package gate; full-pipeline cut/CNF differential validation is coordinated with the MF mapping owner.

# Netlist hierarchy flattening and blackbox roundtrip

## Readability and compactness pass

Ported the complete abcHie.c API group. Both hierarchical-prefix and numeric
suffix naming modes share a concrete recursive copier, with explicit instance
port bindings and model-cycle checks. Logic and latch objects retain ordered
edges; blackboxes retain their actual model definitions in a separately owned
Design. New nets keep useful source names instead of dropping internal names.
Blackbox extraction and reinsertion operate on named boundary nets and real
objects, with design ownership transferred only after successful construction.

## Separate bug and fidelity pass

Tests cover nested and multiply instantiated modules, both naming modes, wire
aliases, nested property outputs, recursive-model rejection, MV value domains,
BDD/Hop manager transfers, and blackbox extraction/reinsertion. Truth tables are
checked after flatten and blackbox roundtrip. A genuine original-ABC fixture
checks exact ordered net names and SOPs for a reused two-instance module.

Regressions/corrections: aliasing formal PI/PO nets now inserts a representation-
appropriate identity bridge instead of leaving the actual output undriven;
property outputs are not duplicated when already exposed. Reinsertion finds a
PO anywhere in the net fanout list, not only at position zero. Parent MV value
names survive child bindings, and incompatible explicit domains fail. Typed-nil
mapped buffer handles are rejected. BDD and Hop nodes duplicated into another
manager transfer their actual function rather than leaving foreign handles.
The Hop owner also fixed explicit-XOR transfer semantics with independent tests.

Source Copy slots remain conversion scratch. Failed construction can update
scratch slots, as in the original, but does not transfer the source Design.
Terminal names and complete parent value names are retained for readability.
Selected full previously ported ABC core plus hierarchy, functionality and
adapter tests pass race tests and vet. Full ABC next reaches abcHieCec.go; no
whole-package completion is claimed.

The subsequent HIE_CEC cross-check exercises 256 combinations of eight boundary
phase bits against both native direct/cached graph expansion. General flattening
now inserts typed inverse bridges to preserve these phases, including formal PI
net and PO terminal phases. The original binary also reproduces duplicate CO
names when a nested property is already a top-level output; the Go copier avoids
that duplicate export (the original reports failed network checks in this case).

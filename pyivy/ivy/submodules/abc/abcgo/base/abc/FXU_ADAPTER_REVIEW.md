# Concrete FXU graph boundary

## Readability pass

Network implements the FXUNetwork interface directly, without importing FXU or
maintaining a second graph. Export rows use stable object IDs and leave non-SOP
objects empty. A replacement validates all definitions, references and cycles
before allocating or changing objects. Appended virtual rows are mapped to real
new object IDs, including forward references and empty appended slots.

## Separate bug pass

Tests preserve output object identity and exhaustive function values through a
forward-referenced extracted DAG. Additional tests cover complemented SOP edges,
cycle/missing-reference rejection without mutation, and virtual-ID remapping.
The review found an empty false-function SOP could collide with the empty row
sentinel; export now emits an explicit all-don't-care zero cover for that case.
Phase absorption also handles XOR/NXOR output tags rather than merely flipping
irrelevant pattern characters. Existing non-SOP objects remain unchanged.

Selected ABC component race tests and vet pass. The FXU algorithm has independent
coverage in opt/fxu; this adapter review covers actual ABC mutation/semantics.

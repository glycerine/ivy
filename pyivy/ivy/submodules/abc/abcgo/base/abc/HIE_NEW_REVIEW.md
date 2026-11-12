# AU hierarchy port checkpoint

All active abcHieNew.c algorithms are native in abcHieNew.go: independent AU
networks/objects, packed integer literal access, traversal/support, lifecycle,
model ordering/counts, CBLIF parsing, AIG derivation, ternary simulation, and the
ABC-netlist-to-AU bridge. Memory page/header address arithmetic is replaced by
owned slices and explicit stable IDs; those allocation-layout-only inline
helpers have no Go equivalent. Vec interfaces become typed slices. Graph-returning
APIs return the actual importable aig.Graph, whose constant-one edge is zero.

## Readability and compactness pass

Object ownership and ID lookups are explicit. AU fanout ports and terminal
indices have separate fields instead of overloading storage after a variable
length fanin array. Model resolution occurs after the complete CBLIF parse.
Expansion and simulation use per-instance environments. ABC caller Data is
preserved. Existing shared local-SOP/AIG primitives perform Boolean lowering.
Statistics account for native Go allocation sizes rather than C allocator pages.

## Separate bug and fidelity pass

- Forward CBLIF models resolve before flattening; model reordering remaps both
  boxes and fanout-port function IDs. Repeated models have isolated values.
- Source AU box argument binding drops complemented literal bits. The native
  implementation preserves them; the independent complemented-actual truth
  test records this intentional correction. SOP literal input phases likewise
  remain explicit.
- Source ternary output processing reads the stale last visited object instead
  of the current output terminal. Independent constant-zero/constant-one output
  regression verifies each output reads its own fanin.
- Indirect model recursion and combinational cycles fail explicitly rather than
  recurse indefinitely. Support traversal counts each PI once.
- Invalid ranges, undefined signal/model references and port-count mismatches
  fail parsing; panic-based internal parse checks become returned errors.
- Owned slices replace accidental C alignment/page dependence. Complemented AU
  object access uses cached proxy handles with regularization rather than unsafe
  pointer tags. unsafe.Sizeof is used only for memory accounting.
- Source all-object traversal is retained for unused box output cones and
  simulation port accounting, while each instance is expanded once.

Validation: complete truth tables for forward-reference CBLIF with AND/XOR and
complemented boundaries; range/repetition parse tests; all 27 ternary input
patterns compared with independently enumerated Boolean refinements; nested
SOP ABC hierarchy derivation compared against the separate native hierarchy
converter; indirect recursion test; real original-C AIGER oracle across eight
assignments and both outputs. Oracle provenance and the corrected source
polarity defect are documented in test_vectors/AU_ORACLE.md.

Selected-source race and vet gates cover this checkpoint. The ordinary package
gate still fails in subsequent originals; it is not reported as package green.

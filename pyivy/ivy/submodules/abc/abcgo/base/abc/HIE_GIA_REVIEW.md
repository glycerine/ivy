# Buffered hierarchy AIG port

`abcHieGia.go` ports all original abcHieGia.c routines using the importable
`aig.Graph` representation. Edge-valued APIs use `aig.Edge`, with constant one
at edge zero. Vec arguments become typed slices. FlattenGia2 visits all box
ports and supports repeated model instances. FlattenGia performs demand-driven
traversal and deliberately requires unique models, as its C source does.

## Readability and compactness review

The common hierarchy expander has optional entry/exit hooks so buffered and
unbuffered traversal share box ordering and polarity handling. Boundary state
is explicit `HierarchyBufferInfo`: physical buffers, formal ports, canonical
logical signals, and forward-traversal groups. No object address arithmetic or
pointer-tag conversions remain. Reinsertion uses a per-module cone cache and
stages every replacement before committing original module identities.

## Separate bug and fidelity review

- Ordinary AIG cleanup discards unused buffers, while original GIA cleanup
  preserves their identities. The dedicated cleanup roots every physical
  buffer and keeps constant/shared placeholder buffers after sweeping.
- Constant and opposite-phase aliases retain logical signals separately from
  physical placeholders. Repeated sweeping preserves all port correspondence.
- Boundary polarity is composed at PI/net, BI/box, PO/BO and BO/net boundaries.
  In particular an output buffer records the formal PO value before the parent
  BO edge polarity; otherwise reinsertion loses that polarity.
- Original InsertOne_rec adds copied fanins to the source object. This port
  attaches them to the destination and verifies mapped source fanins remain
  unchanged.
- Source-wide Copy scratch cannot identify a cone copy across modules. Native
  insertion caches locally and rejects unbound global primary inputs.
- Constant global signals are never rebound to a local input: specialization
  must preserve the constant rather than turn it into a variable.
- Buffer patching rejects forward drivers and cycles before graph mutation.

Regression tests cover nested/reused models, demand-driven unique-model
precondition, all 256 combinations of eight boundary polarities (four inputs
assignments each), flatten/sweep truth, constant/opposite-phase aliases, unused
boundary preservation, mapped reinsertion, design pointer identity, and mapped
source immutability. Source hierarchy semantics are also covered by the real
ABC flattening oracle documented in HIERARCHY_ORACLE.md.

The complete base/abc package still contains later untranslated originals;
selected-source race/vet checks establish this group's scope, not package-wide
completion.

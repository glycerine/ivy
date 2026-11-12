# FXU separate bug review

Reviewed parsing, row/column ownership, pair canonicalization, complement polarity,
heap mutation, cover reconstruction, graph cycles and extraction limits separately
from the compactness pass.

Corrections and regression evidence:

* Zero-valued heap items have index zero without membership. Insert/update/delete
  now check item identity as well as index bounds; pop resets the removed index.
  Random updates, ordering and sole-item removal/reinsertion are tested.
* Original and extracted-node co-singleton columns have opposite positive parity.
  The writer follows the original boundary convention, with exact original-C
  golden coverage and exhaustive output comparisons.
* Original zero extraction budget enters a do/while once; Go respects zero and
  tests that no covers change. This is a deliberate limit-contract correction.
* Deleted rows are distinct from constant-one rows; output phase is retained.
  Random covers include all-dash constant cubes and complemented outputs.
* Malformed/mixed-phase SOPs, duplicate/subsumed rows, duplicate fanins, out-of-range
  edges and combinational cycles return errors before mutation. Tests cover the
  SCC-free precondition, phases and cycles. Resource limits retain source bounds.

Validation: 200 seeded random five-input three-output networks, exhaustive all32
assignments before/after for each, all extraction mode combinations represented,
nonincreasing literal cost and node-budget checks. Original unmodified C source
produces exact single/double/complement fixture covers and fanin vectors. Golden
regeneration harness and command are in test_vectors. This is semantic and selected
trace coverage, not exhaustive testing of all possible multistep tie traces.

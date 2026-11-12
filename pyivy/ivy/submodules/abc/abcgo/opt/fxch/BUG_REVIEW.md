# FXCH separate bug review

Reviewed ownership, bounds, constant rows, paired-removal polarity, subcube
collisions, output intersections, MUX normalization, dependency cycles and resource
limit semantics separately from the compactness review.

Corrections and tests:

* A red regression showed cyclic owner dependencies accepted by the initial Go
  port. DAG validation now rejects cycles before any caller-owned slice changes.
  Literal order, duplicate/conflicting literals, owners and bounds are validated.
* Original hash removal identifies an entry only by cube ID even when multiple
  same-cube subcubes collide. The Go operation matches both removed literal
  indices as well, tested using forced same-ID collisions.
* Output-set copies avoid mutable aliasing across split memberships. Random tests
  exercise exact and partial-overlap groups; a 70-output test crosses the original
  packed mask boundaries and tests unlimited extraction (limit zero).
* Zero-literal constant rows remain active and are distinguished from deleted rows.
  Mixed constants and unary rows have explicit exhaustive truth tests.
* Manager-local MWC state includes the source's reset-call advancement and new-node
  key generation, avoiding races and replacement deterministic-key approximations.
* The source pointer-based fCompl expression in the second MUX-removal pass is
  represented explicitly as the intended data-input phase flip once complement
  mode has been selected. No pointer truthiness leaks into Go.

Evidence: unchanged original C single, double and MUX fixtures match exact resulting
owner/literal rows. 300 seeded random five-input, three-output networks compare all
32 assignments before and after up to20 extractions. Multioutput, constants,
validation and subcube deletion tests run as ordinary package tests. Native harness
and generation instructions live in test_vectors. Exhaustive semantic coverage of
these inputs does not guarantee identical arbitrary tied multistep optimization
traces; this implementation deliberately rebuilds candidate state.

# MF mapping progress and review

The production leaf is pure Go. `GenerateMapping` ports the source cut merge,
CNF clause-count cost, epsilon ordering, area-flow reference blending, second
mapping round and bounded-depth exact-area round. `NetworkFromModel` preserves
source GIA interface and AND ordering after IOA ID normalization. The required
source-default pipeline additionally runs `Gia_ManDupMuxes(...,3)` before mapping;
that coarsening port is integrated by `GenerateCNF`.

## Readability pass

- Native cuts and per-object slices replace packed cut headers and paged arenas.
- Truth literals retain the original table-ID/complement convention.
- References, flow values and delays live in a compact native object record;
  exported CNF inputs contain only the data the derivation needs.
- Network adapters return independent interface slices and preserve the caller's
  IOA model. Ownership and mutation contracts are documented.

## Bug review and regressions

- Ordered leaf merges check actual cardinality after signature collisions.
- Source equal-score insertion order is preserved; strict supersets are removed.
- Independent truth assignments check expansion, support reduction and all
  real XOR/MUX root values over LUT widths 3 through 8.
- LUT widths at most six use the source six-variable CNF helper; width seven
  repeats its two words to satisfy the eight-variable helper's four-word slice
  contract. This fixed an early narrow-slice panic found by CNF review.
- Malformed topology, outputs, duplicate/unlisted inputs return errors.
- Original source's exact-area depth limit and flow float32 arithmetic remain.

The full source-default pipeline now passes 34 original-ABC clause-by-clause
and variable-order differential cases, including larger/reconvergent circuits.
Uncoarsened plain-AIG flow discounts use the source recognizer precedence;
coarsened graphs disable them through MuxesEnabled. Source choices are seeded
via optional Sibling(id) with independently evaluated zero phases. Private
transient mapper arrays are released after exporting the final mapping.

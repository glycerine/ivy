# Ordered LUT4 decomposition dependency review

Scope: the active If_Dec5/6/7Perform algorithms consumed by GIA's specialized4-4
reconstruction, within the ongoing giaIf.c review. This does not complete IF or
GIA, or the original131-package audit (2/131 closed).

Inspection found a generic cofactor/combination substitute. It found valid
configurations but changed search order, variable permutations and don't-care
encoding. Earlier source comparisons checked existence only, so they did not
prove configuration/graph fidelity. Do not count equal Boolean truth as proof
of the requested source algorithm.

## Regression before restoration

Eleven hand-derived configurations (signed/complemented conjunctions) and one
later-disjoint case fail before repair. Expected constants are derived from
source first-free-variable/cofactor conventions and independently composed
back to their truth tables. For six-input AND: old7054800032108000 versus
source0710808054328000. For seven-input AND: old7654800032108000 versus
source7210800065438000. Five-input AND also differs in constant-group encoding.
F=AND(a,b,c,d) XOR AND(e,f) proves the source continues after finding a shared
candidate and ultimately prefers a later disjoint partition. Log:
/tmp/abcgo-dec07-configuration-order-red.log, exit1.

## Restoration

- Five inputs: ordered free-variable moves, two-shared-variable cofactor test,
  original F/C/D truth composition, constant groups coded1, support-based pin
  packing with trailing sentinel6, and configuration verification.
- Six inputs: exact word masks/swaps, evolving permutations, cofactor counting,
  first shared candidate retained while searching all later disjoint candidates,
  source bound/root truth and pin-field encoding.
- Seven inputs: ordered free triples, cross-word adjacent swaps, byte cofactor
  counting and first successful disjoint configuration.

No generic fallback, alternate solver or C toolchain was introduced. Remove the
old generic small-cascade implementation. Configuration utilities remain shared
with other IF source families; they are not automatically marked reviewed.
The source six-input entry requires full support and seven-input cofactor helper
asserts a second cofactor. Existing Go reduced-support handling is retained by
allowing a single-cofactor direct result outside those C preconditions; normal
full-support source traversal/derivation is unchanged.

## Passing checkpoint and work remaining

Exact regressions plus existing generated/recorded cascades pass0.006s. New192
specialized GIA reconstruction cases exercise five/six/seven inputs, four function
families, signed/permuted/complemented functions and signed supplied leaves.
They evaluate full truth and inspect packing/mapping topology independently of
the configuration decoder; pass0.019s. Full IF Go-only1.707s, decomposition/LUT
race5.437s and IF/GIA vet pass. Logs:
/tmp/abcgo-dec07-ordered-initial.log,
/tmp/abcgo-dec07-special-reconstruction.log,
/tmp/abcgo-dec07-ordered-{if-go-only,race,vet}.log.

Before closure: separately review the port, explicitly check ordered permutation
helpers against scalar remapping (including the cross-word swap), broaden source
priority/configuration evidence, then freeze and run affected/full-module gates.
The wider LUT/cofactor/config/cell inventory and mapping-family gate stay open.

# CAS port review

Source: `src/bdd/cas/casCore.c`, `casDec.c`, and `cas.h` (Alan Mishchenko).
The active output encoding, support remapping, transfer, DD-to-BLIF mux emission,
level-cut LUT decomposition, and non-strict column encoding paths are implemented.
The original commented-out `Experiment2`/legacy encoding alternatives are not
executable APIs. Original C files remain visible and unchanged.

## Readability and compactness pass

- Replaced global fixed-size scratch arrays and pointer arithmetic with local
  slices/maps. Kept the LUT's source field names and stage progression to make
  comparison with `casDec.c` direct.
- Kept graph primitives authoritative in `bdd/cudd/core` and encoding/search
  helpers authoritative in `bdd/extrab/cascade`; CAS does not duplicate either.
  The remaining draft `extrab` package can forward to that helper package when
  its surrounding APIs are ported.
- Removed address-derived BLIF node names in favor of stable traversal IDs.
  Phase bits appear directly in mux truth cubes, eliminating inverter bookkeeping.
- The profile helper originally rebuilt cut paths at every level in this port.
  Review replaced it with the original earliest-incoming-edge interval method,
  making width calculation linear in graph size plus variable count.
- Replaced the command-shell `cec` dependency with direct canonical BDD
  substitution in `VerifyCascade`. Temporary encoding variables make substitution
  simultaneous, preventing newly inserted original-input functions from being
  substituted again accidentally.

## Bug review

- Original final-stage code assumes both Boolean terminal columns occur. Constant
  functions violate this assumption. The port derives the final bit from all true
  terminal paths; exhaustive truth tests include both constants and zero-width
  intermediate stages.
- Rejects foreign-manager nodes, non-Boolean ADD leaves, missing names, invalid
  LUT sizes, and infeasible column multiplicities. Tests cover these boundaries.
- Stable support sorting resolves source `qsort` ties deterministically; it matches
  the tested original implementation's order.
- DD emission handles complemented roots, both child phases, shared nodes,
  terminal false, and terminal true. An independent BLIF interpreter checks the
  emitted networks against native truth tables.
- Reviewed the new core variable insertion: a new index changes no existing
  variable's relative order, so insertion updates level maps directly and does
  not collect unrelated unreferenced diagrams or count as heap reordering.
- Original non-strict encoding retains a process-global minterm pointer cache
  across destroyed managers. The C oracle explicitly clears this stale cache
  between managers; the Go search has per-call maps and no address-lifetime hazard.

## Evidence

- All 256 three-variable Boolean functions: decomposition, symbolic equivalence,
  and independent BLIF truth evaluation.
- Six-variable multistage parity under identity and shuffled orders; 128 seeded
  six-variable random truth tables (including infeasible cases).
- Output-support sorting, unused output codes, support remapping/new managers,
  constants, foreign ownership, and insufficient LUT capacity.
- `test_vectors/encoding_oracle.c` calls the original `GetSingleOutputFunction`:
  4,096 three-output cases × 32 assignments, checksum `3702fd9a65c8d325`.
  The permanent Go fixture requires no C compiler or native library.
- The helper package additionally checks 3,856 non-strict column partitions
  against its independent original-C fixture, checksum `6b850130b3287765`.

Validation commands (from repository root):

```
GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/cas ./bdd/extrab/cascade ./bdd/cudd/core -race -count=1
GOCACHE=/tmp/abcgo-go-cache go -C abcgo vet ./bdd/cas ./bdd/extrab/cascade ./bdd/cudd/core
```

## Limits

This completes active CAS algorithms; it does not complete the large draft root
`bdd/cudd` or `bdd/extrab` packages. CAS imports their authoritative working
subpackages. The bounded encoding heuristic can reject LUT sizes that a broader
search might realize, as in C. Direct code assignments and internal BLIF names
may differ from C's pointer-hash traversal; logical behavior and optimization
counts are tested. Cascades retain their manager's level interpretation and must
not be used after independently changing that manager's order. Filesystem-writing
APIs keep their original void/status shapes; low-level writers accept `io.Writer`.

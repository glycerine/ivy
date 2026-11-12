# SPFD reviews

## Separate compactness, idiomatic Go and readability pass

Reviewed the completed `bdcSpfd.go` independently from algorithm verification.
The common gate operation, recursive marking/unmarking, dictionary loading and
candidate-cover selection share helpers rather than five duplicated phase
branches. Public parameterized `DecomposeSPFD`, `GenerateSPFDDictionary`,
`Read/WriteSPFDDictionary`, and `DecomposeSPFDWithDivisors` return explicit results
or errors. Original benchmark entry points remain wrappers, with no background
large allocations. The original two deliberately inactive C exploratory drivers
remain inactive; their actual C code returns or has a false `fTry` condition.

Storage uses Go slices and canonical truth maps. Retain source fanin indices,
input/output complement flags, XOR flag and mark bits in `Bdc_Ent_t`: these are
algorithm metadata, not redundant copies. `Bdc_Nod_t` retains per-level fanin
coordinates and truth/score because the beam needs all of them. No packed bitfield
layout or hand-managed C capacity remains. Levels/index references deliberately
remain rather than turning the search into a different synthesis algorithm.

XOR expression rendering uses `^`; the original diagnostic printed `+`, which
would misleadingly display an XOR gate as an OR expression. Equal-score beam
candidates are reversed before stable sorting to preserve source reverse-cost
selection order.

## Separate bug-review pass

Walked gate generation, source weight/scoring formulas, complement normalization,
recursive overlap marks, hash-chain slots, search bounds, dictionary binary I/O,
cover progress and error paths. Fixed findings before integration:

- Dictionary bounds equal to elementary-variable count now return immediately,
  preventing a one-entry limit overrun. Regression checks that exact bound.
- Constants now have cover cost zero; the source benchmark subtracts one from
  zero steps and reports negative gate cost. A regression covers both constants.
- Subset benchmark propagates its first output write error, rather than losing
  an error if a later write happens to succeed. A failing-writer test checks it.
- Binary reads validate both record widths and matching truth/weight counts.
  Corrupt-file regression covers rejection; round-trip test covers real data.
- Missing separating divisors return an error instead of reading an uninitialized
  best index, and cover iterations require progress within the 64-minterm bound.
- Canonical dictionary output phase, uniqueness and elementary weights are
  checked. Graph marking tests assert shared-cone overlap and complete cleanup.

Validation: `go -C abcgo test ./bool/bdc` passes against the actual package,
including core tests, SPFD counts/cofactors, small exact searches, dictionary
creation/cover/I/O, overlap/hash metadata and benchmark error handling.

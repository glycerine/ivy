# Independent review

## Compactness and readability

Reviewed `canon.go` independently after the implementation author's review. The leaf package avoids GIA/DAU dependency cycles and retains the source algorithm's configuration order and strict tie comparisons. Bit operations are separated into small helpers; scratch storage is allocated once per canonicalization. No simplification was warranted: replacing the bounded refinement with an exhaustive canonicalizer would change source metadata and complexity.

## Separate bug review

Checked variable flips and swaps across the 32/64-bit boundary, reverse-word ordering, phase-bit permutation, output phase location, cofactor counts, selection-sort ties, and the five-pass stopping rule against the source. Existing exact C-oracle tests cover all three-variable tables and structured/random cases through 16 variables, including tied counts. Transformation reconstruction tests independently verify the returned phase/permutation. No unresolved correctness findings. `go test ./opt/dau/canon -race -count=1` passes.

# CAS bridge primitive review

Scope: `cuddBridge.go` and its tests. The other core algorithms retain their
separate reviews; root CUDD is still incomplete.

## Readability and compactness

BDD-to-ADD and ADD-pattern conversions use one memoized structural walk each,
rebuilding through authoritative ITE operations. No second unique table or graph
representation exists. Variable-at-level insertion directly updates level maps:
existing indices keep their relative order, so rebuilding the heap is unnecessary.
The original CUDD copyright/license header is retained.

## Bug review

Tested constants, complemented roots/children, signed numeric leaves, wrong graph
families, insertion at successive levels, invalid levels, and preserved BDD/ADD
truth values. Insertion clears derived symmetry groups but does not collect client
nodes, increment reorder counts, or change ZDD order. Tests run with the full core
race suite; vet covers the core together with CAS and cascade helpers.

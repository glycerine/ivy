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

## Completed numeric bridge family

`Cudd_addBddThreshold`, `Cudd_addBddStrictThreshold`, `Cudd_addBddInterval`, and
`Cudd_addBddIthBit` now complete the original `cuddBridge.c` public conversion
family. Root `cuddBridge.go` forwards these, the prior Boolean/pattern converters,
and transfer to the sole core implementation. Unreferenced C-private recursion
helpers are replaced by the memoized core walk; live internal entrypoints forward.

Readability pass: shared `addToBoolean` traversal now also implements pattern
conversion, avoiding separate graph-walk implementations for each predicate.
Threshold, interval, and bit semantics remain short leaf predicates.

Bug pass: preserved the original *negated* comparisons instead of rewriting them
as >= or >, since those changes alter NaN behavior. Fractional integer leaves
truncate toward zero. Negative values use 32-bit C-int bit patterns. Bit indices
outside 0..31, nonfinite integer conversions, and out-of-range truncated leaves
return nil rather than relying on the original undefined C operations. Ownership
and graph-family validation occur before traversal. Tests cover numeric boundaries,
wrong families/managers, invalid integer conversions, NaN, infinities, and shuffled
manager order.

Original-C oracle `test_vectors/bridge_oracle.c` checks 56 conversions over eight
numeric assignments: checksum `2c12e60c54a5bb25`. No native compiler is needed by
Go tests. A root-forwarder test compiles independently with the authoritative
aliases; this is explicitly not a claim that the entire root CUDD package builds.

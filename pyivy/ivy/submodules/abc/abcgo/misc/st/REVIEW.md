# Completed reviews

## Compactness / idiomatic Go / readability pass

Reviewed table representation and all st/stmm entry points. The two C allocators
share one Go-managed chained table, with thin compatibility wrappers. Stable
entry objects retain pointer-to-value-slot semantics across rehashing. Public
Table methods and comparator/hash helpers support cross-package callers without
exporting bucket internals. Retained bins, entry links, count, density, growth and
reorder settings are required algorithm state; removed manual pool ownership.

## Separate bug-review pass

Reviewed collision chains, replacement versus duplicate insertion, resizing,
reorder-on-hit, mutable slots, delete links, shallow copies, iteration/delete/stop
and numeric key/value conversions. Found mixed-width numeric records/keys could
panic on int assertions in integer compatibility functions; changed them to
identity-width conversion and added regression coverage with int32/int64/int
keys and negative values. Tests also verify stable slots and independent copied
entry chains. `go test -race ./misc/st` passes.

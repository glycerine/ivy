# Satoko utility review

This package ports the bounded queue, indexed activity heap, sorting helpers and
packed activity arithmetic using Go-owned storage and the standard library.

1. **Readability/compactness pass:** the bounded queue stores only its data slice,
   replacement cursor and running sum; size and capacity come from `len`/`cap`.
   The heap uses a slice plus inverse positions and refers to the shared activity
   vector. Original operation-name wrappers remain alongside public Go methods.
2. **Independent bug pass:** checked queue wrap/clear and uint64 sums, heap
   insertion/removal/update/build and inverse positions, packed mantissa/exponent
   operations and saturation, and sorting bounds. Tests cover rolling averages,
   overflow-resistant sums, weight increases/decreases, heap rebuilding, and
   activity addition/multiplication against floating-point reference values.

Validation: `go test -race ./sat/satoko/utils/...` and
`go vet ./sat/satoko/utils/...` pass from the abcgo module. No generated draft
files, stubs, cgo dependencies or known unresolved package build issues remain.
Vector storage and its separate audit are recorded in `vec/PORTING.md`.

# Satoko vector review

All five scalar vector implementations (bytes, signed/unsigned integers, floats
and packed activities) use a single `data []T` field. Its length and capacity
are authoritative; original operation-name functions and public Go methods are
both retained.

1. **Readability/compactness pass:** removed redundant size/capacity fields and
   updated allocation, growth, resizing, push/pop, copying, sorting and clearing
   to slice operations. Grouped imports and documented active versus reserved
   storage. Duplicate retains a larger destination capacity and owns its data.
2. **Independent bug pass:** checked bounds, empty pop, insertion/drop suffix
   movement, clear/free lifecycle, growth, sorting, aliasing and capacity
   retention. Regressions also directly reslice storage to prove Len/Cap cannot
   become stale. The C insertion shift omitted the last element; the Go port
   shifts the complete suffix. Typed initialization fills values rather than
   reproducing C's byte-pattern memset behavior.

Validation: `go test -race ./sat/satoko/utils/vec` and
`go vet ./sat/satoko/utils/vec` pass from the abcgo module. Full Satoko subtree
and LSAT integration tests also pass. No known unresolved vector package issue
remains; no source files are excluded to obtain these results.

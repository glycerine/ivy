# Hop port status

The package is not yet buildable. The generated drafts remain present and require
further conversion; no files were excluded from the build to conceal errors.

Completed in this pass:

- Node-page allocation and disposal use typed Go slices, eliminating raw byte
  alignment, malloc/free and pointer reinterpretation in hopMem.go.
- Free nodes are linked through pNext. Fetch clears the node, assigns the next ID,
  and increments the manager counter as a separate statement. Recycle links the
  node back into the manager free list.
- Structural-hash bucket storage is []*Hop_Obj_t.
- Hop_ManCheck uses ordinary Go loops and validates PI/PO fanins, internal fanin
  ordering, structural-hash membership and object counts.

Validation: hopMem.go and hopMem_test.go pass a standalone test using a minimal
manager/object type fixture in /tmp/abcgo-hop-memory-check. This verifies the
actual page allocator, free-list cardinality/absence of cycles, second-page
allocation and disposal. It does not establish that the whole package builds.

Remaining work includes tagged/complemented edge representation (currently C
pointer-bit casts), the pData/iData and pNext/PioNum unions, pointer/int vector
integration, manager allocation, structural hashing, graph algorithms and truth
operations, and the bundled cudd2 adapter. The whole-package build currently
fails on retained C syntax beginning in cudd2.c.go/cudd2.h.go; hop.go also still
contains C constructs. Source references are src/aig/hop/*.c and hop.h.

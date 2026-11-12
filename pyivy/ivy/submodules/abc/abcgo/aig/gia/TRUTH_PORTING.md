# GIA truth computation and NPN reduction

`giaTruth.go` ports every function implemented in `src/aig/gia/giaTruth.c`, including the original recursive entry points. No `giaTtopt.c`, `giaTtopt.go`, or corresponding header API exists in this checkout; no substitute interface was invented.

Truth tables are little-endian `[]uint64`, normalized by repeating unused variables below six inputs. The `Gla_ObjTruth*` accessors return slices instead of C pointers. Manager-owned full/cut results remain borrowed until the next computation or workspace release. Explicit cut support order and duplicate-leaf first occurrence are preserved. Scratch-based APIs retain their traversal-ID, object Value, and scratch-vector effects. Workspace internal numbering and postorder node lists remain observable through the original accessors.

A shared memoized evaluator handles multiword tables; a separate single-word evaluator avoids allocating whole tables or repeatedly scanning a manager during scratch recursion. Both handle ordinary ANDs, encoded XORs, buffers, complemented edges, constants, COs, and explicit MUX controls. Invalid cuts reaching an unlisted CI panic rather than silently inventing a truth variable. Six-input APIs tolerate unrelated additional CIs while rejecting cones depending on a seventh indexed CI.

Shannon synthesis preserves the original highest-dependent-variable recursion and raw-versus-hashed construction choice. Constant, duplicate and opposite AND operands are simplified before calling the manager allocator. NPN reduction retains first representatives, minimum-base compression, constant/literal classes, oversized-cone singleton classes, primary inputs, and the transitive closure of reachable registers and their next-state functions.

The exact `Abc_TtCanonicize` implementation lives in the acyclic `opt/dau/canon` leaf package. Importing `opt/dau` directly would introduce a cycle through its GIA bridge. This helper preserves the active C algorithm's count ordering, selection-sort ties, phase/permutation metadata, refinement traversal, and five-pass bound. It does not claim exhaustive mathematical NPN classification beyond the original algorithm.

## Validation

Run `./test_ported_truth.sh -race` from any directory. This focused script explicitly reports its scope; unrelated GIA/AIG drafts remain in their original locations and continue to prevent the full package build.

Tests include all 256 three-variable functions with both raw and hashed synthesis; full, cut, compact and object-indexed scratch tables; randomized six-variable synthesis; nine-variable multiword tables; variable mapping and LUT metadata; cofactors; deterministic workspace numbering; NPN classes and sequential register closure. An optional Linux C compiler oracle includes the actual `giaTruth.c` and compares random complemented AND/XOR graphs and the distinct original multiword computation through twelve inputs.

`go test -race ./opt/dau/canon` compares every three-variable function plus deterministic random and structured tied-count tables through sixteen variables against the actual `dauCanon.c`, including exact truth words, phase masks and permutations. An independent transformation reconstruction test runs without a C compiler, covering zero through ten variables.

The complete `go test ./aig/gia` remains blocked by unrelated draft code (currently syntax errors in `aig/aig/aigRepar.go`). No source is hidden, excluded, replaced with a stub, or moved to make the focused checks pass. Independent compactness/readability and bug-review findings are tracked separately when completed.

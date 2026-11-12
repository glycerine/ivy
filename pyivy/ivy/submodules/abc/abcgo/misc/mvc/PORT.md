# MVC Go port

The active implementation in `src/misc/mvc` is ported using Go-owned slices of
32-bit words and linked cube lists. C names remain exported (`Mvc_*`) to keep
consumer ports recognizable. Fields are exported with conventional Go names:
`Next`, `Data`, `Unused`, `Ones`, `NBits`, `Cubes`,
`Mem`, `Array`, `Lits`, and `Mask`. Predicates return bool. Division functions
return `(quotient, remainder)` instead of writing pointer-to-pointer outputs.
The memory manager is a lifetime grouping marker; Go's garbage collector owns
all allocations. Explicit Free methods release references where useful.

The 1/2/N-word C macro specializations use the common slice-based Go operations.
Iterator macros become ordinary loops over `cover.Cubes.Head` and `cube.Next`.
Internal linked-list merge sorting uses the standard library's stable sorting.
C routines inside literal `#if 0` blocks were not part of the active package and
are not exposed as invented APIs. The header's historical declaration-only
routines without C definitions likewise have no placeholder implementation.

Tests exercise bit/word boundaries, padding, duplication, 200 randomized algebraic
division reconstruction cases, 100 exhaustive containment truth comparisons,
kernels and common cubes, remapping, transposition, cofactors, variable flipping,
and textual cube printing. The bool/dec consumer additionally verifies recursive
factoring exhaustively over all 256 three-variable truth functions.

Review cleanup: cube word count/last index derives from len(Data), and cover word
count/padding derives from NBits. Unused records the final word's logical width
(which a word slice alone cannot encode); Ones is an explicit sort-key cache
refreshed by CoverSetCubeSizes, retaining the original read/set size semantics.

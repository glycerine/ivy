# RPO Go port review

Reviewed against `src/base/abci/abcRpo.c` and the `bool/rpo` Go port.

- Kept RPO truth tables in the source's contiguous 64-bit `word` layout, with per-function slices over one owned backing allocation. Binary I/O is explicitly little-endian, matching ABC's host representation on supported builds.
- Text load still infers variable/function counts through `Abc_TruthGetParams`, reads most-significant hex digits first through the existing shared truth-table parser, stops at the first blank line, and updates the loaded function count. Text output emits one truth table and newline per row; binary output writes every table word without headers.
- The C call passes `word *` to an `unsigned *` factorizer. On the target ABI, those are the low and high uint32 halves in sequence. `rpoTruth32` performs that layout conversion, retaining only the one 32-bit word RPO expects for up to five variables.
- RPO store helpers use a distinct prefix because the package already exposes shared `Abc_TtStore*` and `Abc_TruthStore*` APIs for the NPN/decomposition store. This avoids Go's lack of C translation-unit static linkage collisions while keeping the command entry points unchanged.
- Owned tables release by dropping Go slice references; borrowed buffers remain caller-owned. Invalid dimensions and unreadable files return nil rather than dereferencing failed allocations.

No RPO grouping or factorization logic was introduced here; all decomposition remains delegated to `bool/rpo.Rpo_Factorize`.

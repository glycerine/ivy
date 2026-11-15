# Exclusive-sum-of-products computation

Review all 14 routines in `src/aig/gia/giaEsop.c`, implemented by `esop.go`.
Preserve the source cube interning, ordered literal products, distance-one XOR
minimization, constant sentinel and cover-to-graph construction. The original
algorithm can produce exponentially large covers; this review does not replace
it with another representation or solver.

## Repairs and failing evidence

`esop_review_test.go` first fails on real XOR and MUX gates, in both rebuilt
output graphs and extracted covers. The loop was reading typed gate fanins as
an ordinary AND. The same red run finds lost register dimensions on ordinary
sequential inputs and reordered ports
(`/tmp/abcgo-esop-typed-interface-red.log`).

At the original AND-only computation boundary, lower real XOR/MUX gates using
`Gia_ManDupNoMuxes`, then run the unchanged ESOP algorithm. Restore the caller's
Value scratch after temporary lowering. Extraction retains those values;
rebuilding publishes the same CI copy literals as ordinary conversion. Existing
buffer behavior is separately tested in both polarities.

Set register count before cleanup and copy unchanged interface metadata through
the existing helper. Retain property/constraint roles, reset states, identity,
comments and independently owned input/output names. The original ABCI `&esop`
command replaces its frame graph without prohibiting sequential input, so
silently turning RIs into ordinary outputs changes behavior. This repair
intentionally preserves Go interface semantics beyond the C routine's limited
name/spec copy.

The original-C verbose comparison then fails on the changed timing-field width
(`/tmp/abcgo-esop-reports-red.log`). Restore `%9.2f` source formatting. Only its
numeric clock value is normalized in the comparison. Cube printing already
matches C and requires no repair.

## Source reconciliation and evidence

`test_vectors/esop_review.c` includes the actual original source. The existing
source-oracle helper links original ABC objects; native tests require Linux/GCC
and those objects, and skip when unavailable. There are no algorithm stubs.

| Source routines | Coverage |
| --- | --- |
| `Eso_ManCube`, `Eso_ManAlloc`, `Eso_ManStop` | Initial singleton IDs, canonical interning through exhaustive cube operations, 0..5 variable managers, scratch ownership and returned data after manager disposal. |
| `Eso_ManCoverPrint` | Exact native empty, constant and signed-singleton printing. |
| `Eso_ManCoverDerive` | All 364 native extracted-cover records from 64 generated graphs, with exact literal order and negative output tags; independent interpretation of every output at all assignments. |
| `Eso_ManCoverConvert` | Rebuilt graphs checked exhaustively against original scalar truth; zero-variable, constants, inputs, no outputs, typed mixed cones, reordered ports and sequential interface metadata. |
| `Eso_ManFindDistOneLitEqual`, `Eso_ManFindDistOneLitNotEqual` | All applicable sorted equal-length and adjacent-length pairs among every four-variable cube, compared to native results. |
| `Eso_ManMinimizeAdd`, `Eso_ManMinimizeCopy` | Exact native output ordering for all 6,561 ordered four-variable cube pairs. Independent XOR truth over 66,430 ordered pairs across 0..5 variables and 24,576 deterministic longer sequence prefixes. Covers cancellation, complement collapse and recursive adjacent-bin rewrites. |
| `Eso_ManComputeAnd` | Native product IDs and independent conjunction truth, including contradictory literals, equal/shared terms, sorted merges and constant-product handling at the cover boundary. |
| `Eso_ManComputeOne`, `Eso_ManTransformOne` | Native full-network covers plus independent product/complement truth; zero covers, constant sentinel, singletons, multiple cubes and both phases. |
| `Eso_ManCompute` | Native ordinary end-to-end cover matrix and verbose statistics; typed XOR/MUX/buffer phase matrix in graph and extraction modes; 32 mixed typed/interface/port-order cases and empty/CI-only cases. |

Independent scalar evaluation does not use ESOP scratch or graph hashing. The
native comparison checks exact cover/minimizer ordering rather than accepting
only equivalent Boolean functions. Property and name tests also mutate result
storage to prove it does not alias the source. Typed extraction checks every
original Value field remains unchanged; rebuilding checks published CI mappings
and graph validity after disposing of the source.

## Compactness and retained contracts

The Go interner uses binary string keys and owned literal vectors in place of
C's packed hash arena. IDs and first-encounter order retain the source contract;
the first `2*nVars` IDs are signed singleton literals. The `Cube1` sentinel is
kept distinct from all interned IDs. No extra solver, cache or persistent fields
were added.

Review the temporary merge buffer before rewriting it. A warmed six-variable
cube-product benchmark measures 71.75 ns/op, 24 B/op and one allocation per
product (`/tmp/abcgo-esop-product-before.log`). Retain the current readable merge
rather than introducing aliasing conditions or unsafe string storage for an
unproven benefit. The temporary output is also safe when source cube vectors
remain referenced while a new cube is interned.

Retain the source conservative adjacent-upper-bin condition and all recursive
minimization ordering; exhaustive truth/native tests found no bug there. Cube
variables refer to logical CI positions. `Lit+2` in conversion is intentional:
new CIs are appended densely in logical order before rebuilding products. It is
not an unconverted source object-ID lookup.

Typed normalization is confined to the source's ordinary-AND boundary. The
underlying ESOP product/minimization algorithm remains unchanged. Node-indexed
mapping/timing payloads are not copied as if their old IDs survived. Extracted
covers own their literal arrays and output tags after the temporary manager is
stopped. Graph conversion owns its unchanged interface metadata.

## Validation

Final focused ESOP tests pass in 5.617s
(`/tmp/abcgo-esop-final-focused.log`). Frozen uncached module race tests pass all 130 packages with tests, including
GIA in 482.089s; three command packages have no tests
(`/tmp/abcgo-esop-final-module-race.log`). GIA vet passes with empty output
(`/tmp/abcgo-esop-final-vet.log`). Both exact handles terminated successfully
before results were recorded in `../../SEQUENTIAL_AUDIT.md`. All 14 routines in
this source family are reconciled.
This family review does not complete the remaining GIA package inventory.

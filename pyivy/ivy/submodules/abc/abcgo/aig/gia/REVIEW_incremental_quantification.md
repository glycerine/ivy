# Incremental support-guided existential quantification

Review all 25 routines in `src/aig/gia/giaExist.c`. Go implementations are in
`quant_support.go`, `quant_cone.go` and the support hook in
`manager_auxiliary.go`; typed append hooks also touch `special_gates.go`.
Preserve both source quantification algorithms, support bitmaps, two-copy caches,
cofactor lists, cone collection and reconstruction. Do not replace them with a
new solver or exhaustive-assignment implementation. Enumeration is used only as
an independent test oracle.

## Correctness repairs and red evidence

| Finding | Red evidence and repair |
| --- | --- |
| An intermediate constant is recursively read as an AND when variables remain. | Eliminating every MUX input panics in both output phases in the incremental variant (`/tmp/abcgo-exist-constant-row-red.log`). Add the constant-zero recursive cofactor base case. Both phases are satisfiable and must leave true. Original C also aborts on these cases. |
| An unaffected retained CI in a cofactor list is incorrectly treated as the currently eliminated variable. | Generated seed18 fails independent truth; fresh and repeated cases also fail in unmodified C. Reduce to `exists a,b: c AND (a OR b) == c`, which returned true (`/tmp/abcgo-exist-retained-ci-red.log`). Preserve `[lit,lit]` for a CI unless it carries the current support bit. |
| Appending to a returned support-row slice overwrites its neighbor. | One-word row append changes the next CI word from 2 to 12345 (`/tmp/abcgo-exist-constant-row-red.log`). Bound row capacity while retaining writable-view semantics. |
| Spare vector capacity can be read as an initialized support row. | An appended CI without its support row is silently read as zero (`/tmp/abcgo-exist-row-length-red.log`). Restrict the vector view to stored length before slicing a row, consistent with C's checked access. Use Go-sized offset arithmetic. |
| The direct cone copier reads XOR/MUX data as an AND and omits MUX control. | Both typed output phases/all keep masks receive independent truth checks; cases fail first (`/tmp/abcgo-exist-typed-red.log`). Copy all logical inputs and lower through `giaHashMappedGate` at the ordinary source-algorithm boundary. |
| Typed append operations leave missing support rows; MUX support omits control. | All XOR/MUX/buffer construction cases fail with zero bits or a row-bound panic at 65 inputs (`/tmp/abcgo-exist-typed-support-red.log`). Add existing support hooks after metadata initialization, include control dependencies, and lower collected typed nodes through the same Boolean helper. |
| Constant row zero holds the selection mask and is mistaken for an internal cone node. | Raw zero/one fanin cases panic (`/tmp/abcgo-exist-raw-constants-red.log`). Treat constant zero as a fixed side boundary, reconstructed as its original constant. |
| Selection-marker bits leak into newly appended gates with constant fanins; affected-cone counting includes the constant as a gate. | Separate red regressions show row `[3]` instead of only the real input and a count of two instead of one (`/tmp/abcgo-exist-constant-marker-red.log`). Constants contribute no structural support and stop affected-node recursion. |

The first two findings occur on ordinary canonically hashed graphs and are
reproduced in original C. The repairs preserve the original algorithms while
correcting those boundary decisions. Raw simple-graph and typed cases extend
safe Go handling beyond the source's usual ordinary canonical-AIG use.

## Source reconciliation and coverage

`test_vectors/exist_review.c` includes actual `giaExist.c`, without algorithm
stubs. The existing native helper requires Linux/GCC and original ABC objects;
those comparisons skip when unavailable.

| Source routines | Evidence |
| --- | --- |
| `Gia_ManQuantInfoId`, `Gia_ManQuantInfo` | Exact native row words at every widening boundary; writable bounded rows, neighboring-row append isolation and rejection of uninitialized rows. |
| `Gia_ObjCopyGetTwoArray`, `Gia_ObjCopySetTwoArray` | Used through recursive cached two-cofactor queries, repeated variable eliminations, intermediate constants and retained-CI regressions; checked by complete independent output truth. |
| `Gia_ManQuantSetSuppStart`, `Gia_ManQuantSetSuppZero`, `Gia_ManQuantSetSuppCi`, `Gia_ManQuantSetSuppAnd` | Native full/subset registration and widening at 0/1/2/63/64/65/127/128/129/191/192/193 inputs. Independent mixed typed support unions include every stored word and unused high bit. |
| `Gia_ManQuantUpdateCiSupp`, `Gia_ManQuantCheckSupp` | Native lazy registration, support-bit queries in both source algorithms, typed control at bit64 and independent tracked-CI sets. |
| `Gia_ManQuantVerify_rec`, `Gia_ManQuantVerify` | Independent sets on a 193-input/256-gate typed DAG; no mismatch output, including real MUX control dependencies and constant fanins. |
| `Gia_ManQuantCheckOverlap`, `Gia_ManQuantMarkUsedCis` | Native selected-variable words and every object overlap; independent marker expectations using logical CI IDs and callback data. |
| `Gia_ManQuantCountUsed_rec`, `Gia_ManQuantCountUsed` | Exact native per-variable affected counts, independent reachable affected-node sets on typed DAGs and a constant-boundary regression. |
| `Gia_ManQuantDupConeSupp_rec`, `Gia_ManQuantDupConeSupp` | Direct logical CI maps, selected-variable counts, both output phases, source Value preservation, cleared copy slots and owned results; ordinary/generated/typed end-to-end cases. |
| `Gia_ManQuantExist_rec` | Cached recursive cofactors, current-variable support gating, constant and retained-CI base cases, deep/shared cones and exhaustive small-graph truth. |
| `Gia_ManQuantExist2` | Exact native MUX cases, all generated masks/phases, typed cone lowering without source support state, repeated queries and constant/raw fanins. |
| `Gia_ManQuantCollect_rec`, `Gia_ManQuantCollect` | Side/quantified/internal partitions used through exact native result literals and counts; typed logical inputs, fixed constants, deep sharing, source order and allocation benchmark. |
| `Gia_ManQuantExist2Dup`, `Gia_ManQuantExistInt`, `Gia_ManQuantExist` | Native ordinary results; collected typed lowering with active support, interleaved input object IDs, constants/simple gates and complete independent existential evaluation. |

Native support/marker/affected-count comparisons match 47,187 exact records.
Thirty directed MUX cases match original C result literals, object counts and all
truth bits. Two documented all-quantified incremental MUX cases crash in C and
are excluded only from native equality; permanent Go regression tests validate
their correct result independently. No broad normalization accepts differing
algorithms. The separate retained-CI source defect is reduced and checked by
independent truth rather than making Go reproduce its wrong answer.

Independent enumeration validates 4,608 repeated generated queries on 64 graphs,
both output phases, all sixteen retained-input masks and both algorithms. The
same graph grows across mask changes to detect stale scratch state. Another 192
queries cover interleaved CIs and internal objects, so physical object IDs cannot
stand in for logical CI indices. Typed XOR/MUX/buffer tests exercise both APIs,
all masks and phases, including control support at bit64. Raw zero/one/repeated/
opposite fanins cover both algorithms and phases. Callback context is retained.

A deeply shared 21,001-internal-gate graph tests recursion, memoization, logical
truth and complete copy-slot cleanup in both algorithms. Direct cone copies
preserve source Value scratch and remain evaluable after source disposal.
Independent full/subset support sets on a 193-input mixed DAG check support
unions, marker bits, structural verification and affected-node counts.

## Compactness

The shared slice-returning logical-fanin helper initially introduced 10,002
allocations and 80,327 bytes per collection of a 5,001-node cone. Replace those
per-node slices with a local fixed three-input tuple, preserving stored fanin
order (data then control for MUXes). Native result/copy order remains unchanged.
The 100-iteration benchmark changes from 281,377 ns/op to 239,527 ns/op, zero
allocations per call and 204 amortized bytes including stack overhead; see
`/tmp/abcgo-exist-collect-{before,after}.log`. No new persistent state is added.

Retain the flat word matrix, packed support bits and two-literal cache. Cached
`nSuppWords` is meaningful during row growth; `iSuppPi` is a count during
construction and the selected support-bit index during elimination. Their
roles differ from slice length and from physical fanin lists. The registered-CI
vector maps dense support positions to logical CI indices, including tracked
subsets, and cannot be replaced by object IDs. No speculative field or algorithm
rewrite is warranted after these checks.

## Preconditions and representation boundaries

The incremental API requires support initialized on an empty manager and updated
as CIs/gates are built, with a live hash table for rebuilding. Cone-based Exist2
can start from a graph without support state and constructs its own tracked
ordinary cone. Callbacks receive logical CI indices and caller data. Direct
cone-copy and verification routines retain their source internal-root contracts;
public quantifiers handle constant and CI roots before those paths.

Support is structural and can conservatively include inputs of redundant raw
gates. Row zero is selected-variable scratch, never actual support for a
constant. Internal affected-count helpers require the current support index,
not the construction-time count. Source copy arrays are borrowed scratch and
visited slots are cleared on normal return. Real gates are lowered at the
source AND-only reconstruction boundary; typed support/verification treats all
logical inputs, including control.

Keep source ordering, cofactor-list deduplication and both reconstruction
strategies. Empty/constant results have mathematically correct literal values.
There is no newly invented solver, approximation, variable cap or replacement
quantifier. Source exponential worst-case behavior is not hidden by test or
implementation shortcuts.

## Validation

Focused quantification tests, including original quantifier tests, pass in
5.908s (`/tmp/abcgo-exist-final-focused.log`). Frozen module race and GIA vet
checks are pending. Final evidence will be recorded in
`../../SEQUENTIAL_AUDIT.md`; the overall GIA package review remains in progress.

# Sequential signals, initialized frames and enable removal

Review all ten routines in `src/aig/gia/giaEnable.c`, implemented primarily by
`enable_detect.go` and `enable_unroll.go`. Keep the original structural detection,
zero-initialized frame expansion, candidate transfer and two removal algorithms.

## Correctness repairs and red evidence

| Finding | Repair and evidence |
| --- | --- |
| Three copy paths treated real XOR/MUX nodes as ANDs; raw copying panicked on buffers/simple opposite fanins. | Red exhaustive truth tests in `enable_typed_review_test.go`; hashed copies lower real gates using existing Boolean hash operations, while raw copying uses the existing typed append helper and representation setup. |
| Typed roots/branches could accidentally match bit-field structural MUX patterns and corrupt RI logic. | Red independent Boolean cases in `enable_boundaries_review_test.go`; require actual Boolean AND structure and preserve unrecognized typed transitions. |
| Raw removal discarded nonphysical logical CI/CO order. | Separate red input/output/both permutation cases; use established CI/CO remapping helpers. |
| Supergate recursion flattened typed primitives and expanded shared buffer cones exponentially; detection falsely classified a typed MUX. | Red atomic-leaf, false-enable and five-second subprocess-timeout tests; keep typed primitives atomic and visit each ordinary AND once with an ordered stack. |
| Signal printing scanned the whole supplied buffer instead of the manager's object count and could panic. | Oversized-buffer red regression; bound both scans as in C. |
| Failure reports changed driver/fanin word order. | Native unrecognized-driver reports expose the mismatch; restore source wording. |
| The delegated cofactor helper omitted verbose banner/statistics. | A native verbose pipeline fails before the five-line reporting repair in `cof_duplicate.go`; normalize only representation-dependent memory size. |

Logs and staged outcomes are recorded chronologically in `SEQUENTIAL_AUDIT.md`.
Every discovered repair above has failing evidence before its implementation.

## Source reconciliation and coverage

`test_vectors/enable_review.c` includes the actual original implementation,
including private helpers. Native tests require Linux, GCC and original ABC
objects; they skip when those facilities are unavailable. The ordinary matrix
compares 4,256 trace lines, plus nine high-frequency and seven verbose-pipeline
lines. It covers 16 AND/expanded-MUX phase configurations with two registers.

| Source routines | Evidence |
| --- | --- |
| `Gia_CollectSuper_rec`, `Gia_CollectSuper` | 512 direct positive/complemented/public/prefixed calls; native leaf order and prefix behavior. The public routine remains intentionally one-level. |
| `Gia_ManPrintSignals`, `Gia_ManDetectSeqSignals`, `Gia_ManDetectSeqSignalsWithFanout` | Complete reports, set/reset/enable ordering and threshold queries; thirteen shared enables exercise frequency >10 and real reference counts. Quiet/reporting ownership tests verify borrowed versus cleared reference tables. |
| `Gia_ManTransferFrames` | 48 native cleanup/transfer stages and direct duplicate/absent/zero/unreachable/Value-one boundary cases. |
| `Gia_ManUnrollInit`, `Gia_ManUnrollAndCofactor` | 48 native frame-copy and composed cofactor pipelines, complete copy tables, graph/interface counts and 3,072 native output truth rows. |
| `Gia_ManRemoveEnables`, `Gia_ManRemoveEnables2` | Native recognized removals and raw unrecognized-driver fallback; independent typed/phase, logical-order, signed first-seen class and mixed-failure cases. |

Independent simulation checks 1,872 input sequences through initialized typed
XOR/MUX feedback, one through four frames, ordinary/cofactoring paths and distinct
simultaneously updated registers. It checks source Value preservation, exact
interfaces, copy-table lifecycle and owned results after source disposal.
Additional tests preserve unrecognized input/state/constant/AND/nonholding-MUX
transitions in both phases, while still removing ordinary structural enables
whose data contains real XORs. Signed control classes and class-zero fallback
counts remain deterministic in first-encounter order.

## Compactness

The iterative recursive-entry collector uses local visited/leaf sets, retains
existing output prefixes and first-discovery order, and does not alter manager
marks. The 60-buffer regression exceeds five seconds before repair and completes
normally afterward; native small-cone comparisons verify the preserved ordering.
A separate test prefixes an internal node ID to prevent incorrectly treating
existing output membership as a traversal boundary.

Control classes use one map from canonical signed handles to encounter IDs,
removing the explicit quadratic search and shortening the code. Map iteration
never determines labels. Quiet detection still performs the analysis but avoids
reference-count construction used only by verbose printing. The shared typed
hash-copy helper handles the distinct frame/Value readers without duplicating
the gate logic. No persistent scratch/cache fields were added.

## Source boundaries and explicit Go behavior

Detection/removal recognizes ordinary structural AND/MUX patterns. Real typed
roots/branches are atomic and remain unrecognized when they do not form that
pattern; their logic is preserved. Real gates inside the data of an ordinary
structural enable remain supported. This enforces the source algorithm's Boolean
structure assumptions on Go's broader graph representation rather than adding a
replacement classifier. Hashed frame/removal copies lower real gates into
ordinary AIGs through existing XOR/MUX operations.

Retain earlier Go repairs that preserve unrecognized RIs where C's second
removal variant leaves an invalid Value, and that preserve nonholding structures
where C's first variant asserts. New tests cover both. Missing frame-copy entries
are skipped safely in Go; C assumes those entries exist. Transfer retains source
Value-one behavior and uses object IDs without literal phase, as its source API
does. Direct-recursive constant handling is a safe Go extension. Public collection
still rejects complemented roots, and detection requires registers.

Initialized unrolling requires positive frame counts and starts every RO at
zero. It returns a combinational frame graph with the source's frame-major PO
layout and Name/Spec policy. `UnrollInit` publishes the copy table for callers;
the cofactoring driver erases its temporary copy table. The delegated cofactoring
algorithm retains its own limits and behavior. Only its missing reporting was
reviewed/repaired here; the rest of `giaCof.c` remains a separate family audit.

Final focused and module race/vet results are recorded in
`../../SEQUENTIAL_AUDIT.md`.

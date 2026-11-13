# ACEC source and test inventory

This inventory records exercised paths, not a claim that every C branch has been
compared. Test files named below are in this directory; the checked-in reference
implementations are in `src/proof/acec`.

| C source family | Covered behavior |
| --- | --- |
| `acec.c`, `acecCore.c` | Parameter defaults, miter/witness behavior, adder-box matching, and small end-to-end equivalence cases: `acec_options_review_test.go`, `acec_source_witness_test.go`, `acecCoreMatch_test.go`, `acec_integration_test.go`, `abc_differential_test.go`. |
| `acec2Mult.c`, `acecMult.c` | Cut storage and truth, multiplier scans, partial products, and small exhaustive multiplier checks and a 3×3 end-to-end C/Go comparison: `acec2Mult*test.go`, `acecMult*test.go`, `acec_multiplier_test.go`, `acec_multiplier3_source_test.go`. |
| `acecBo.c` | Booth tree cases: `acecBo_test.go`. |
| `acecCl.c`, `acecCo.c`, `acecCover.c` | CLA detection and rewriting, core ordering/root collection, and cover exploration: `acecCl*test.go`, `acecCo*test.go`, `acecCover*test.go`. |
| `acecFadds.c`, `acecRe.c`, `acecTree.c` | Full/half-adder detection, polarity, adder chains, timing boxes, cut and tree reconstruction: `acecFadds*test.go`, `acecRe*test.go`, `acecTree*test.go`. |
| `acecNorm.c`, `acecOrder.c`, `acecPa.c`, `acecPool.c` | Normalization, ordering, phase propagation, pool membership and roots: `acecNorm_fallback_test.go`, `acecOrder_test.go`, `acecPa*test.go`, `acecPool*test.go`. |
| `acecPo.c`, `acecPolyn.c`, `acecSt.c`, `acecStruct.c` | Signature parsing, polynomial expansions, NPN/stats and structural XOR traversal: `acecPo*test.go`, `acecSignatureSource_test.go`, `acecPolyn_test.go`, `acecSt_test.go`, `acecStruct_test.go`. |
| `acecUtil.c`, `acecXor.c` | Utility operations, XOR roots and boxes: `acecUtil_test.go`, `acecXor*test.go`. |

Remaining work includes differential tests against C for complex Booth and
multiplier layouts, larger carry trees and choice-rich GIA networks, timing-box
configurations, constrained/resource-limited runs, and full source diagnostic
and unusual parameter branches. Native MUX/buffer handling and documented C
soundness defects remain subject to the limitations in `PORTING.md`. Existing
small exhaustive tests and the source-derived cases do not establish equivalence
on all circuits.

# Bugs found and fixed in the original C ABC

Each entry has a test that failed against the unmodified C before the fix,
and states how the Go port was verified.

## 1. `Extra_Truth4VarNPN` puts 221 functions in the constant class

- **File:** `src/misc/extra/extraUtilMisc.c`, `Extra_Truth4VarNPN`.
- **Found:** 2026-11-26, while tracing why Go and C `rewrite` diverged on i10.
- **Bug:** the loop visits truth tables `1 .. 2^15-1`. When it meets a new class
  representative `r` it sets `uMap[r]`. The complement `~r` (always `>= 2^15`)
  gets its canon and phase (`| 16`), but `uMap[~r]` is only written in the
  "already assigned" branch, which never runs for a representative. So for all
  221 non-constant classes, `uMap[~r]` stays 0, the constant class.
- **Effect:** `rewrite` (`Rwr_*` through the DEC manager), `rwt` and
  `abcMeasure` look up class 0, which has no subgraphs, for any 4-cut computing
  the complement of a representative. A 4-input OR (`0xfffe`, the complement of
  representative `0x0001`) is one example. Those cuts were never rewritten.
- **Fix:** when a representative is created, also set `uMap[~uTruth & 0xFFFF]`.
  Canon, phase and permutation tables are unchanged.
- **Test:** `abcgo/misc/extra/npn/npn4_c_oracle_test.go`
  (`TestTruth4VarNPNMatchesC`) compiles the C function from `src/` through
  `abcgo/misc/extra/npn/test_vectors/npn4_tables_oracle.c`. It checks that every
  function's class equals its canon's class, and that all four C tables equal
  the Go tables. Before the fix it reported 442 mismatches (221 functions, two
  checks each).
- **Go port:** already correct. `misc/extra/npn.Extra_Truth4VarNPN` assigns
  classes from canons after the loop. The test above confirms that it now agrees
  with C on all 65,536 entries.

## 2. `&put` crashes on a GIA whose choices are stored as siblings

- **File:** `src/base/abci/abc.c`, `Abc_CommandAbc9Put`.
- **Found:** 2026-11-26, by differential runs of `&get; &st; &synch2; &put`.
- **Bug:** `&synch2` (and other `Gia_ManFromAigChoices` users) store choices in
  `pSibls`, leaving dangling choice nodes and no `pReprs`/`pNexts`. `&put` sees
  the dangling nodes and calls `Gia_ManToAig( p, 1 )`, which asserts
  `p->pNexts && p->pReprs`. Release builds compile the assert out and then
  dereference NULL, so native `abc` segfaults.
- **Fix:** take the choice conversion only when `pReprs` and `pNexts` exist;
  otherwise convert without choices, as the no-dangling branch already does.
- **Test:** `abcgo/base/abci/native_binary_test.go`
  (`TestNativeSynch2PutDoesNotCrash`) runs the script through the native binary
  named by `ABC_BINARY`. It segfaulted before the fix; native `abc` now reports
  1,861 ANDs at level 28 on i10.
- **Go port:** it panicked in the same place (`gia: choice export requires
  representatives and next links`) and has the same fix
  (`TestSynch2PutDoesNotPanic`). The Go `&synch2` result itself still differs
  from native ABC and is being investigated separately.

## 3. `Sfm_TruthToCnf` fixes the first fanin, not the output, of a constant node

- **File:** `src/opt/sfm/sfmCnf.c`, `Sfm_TruthToCnf`.
- **Found:** 2026-11-26, while porting the SFM engine of `&mfs` line by line;
  native `abc` segfaulted on `&if -K 4; &mfs -a -l` for i10 and sort16.
- **Bug:** for a constant function the CNF is one unit clause on variable 0
  (`(char)(Truth == 0)`, and `1`/`0` for more than six inputs). Variable `nVars`
  is the node output; variable 0 is the output only when the node has no fanins
  (note the commented-out `assert( nVars == 0 )`). `Sfm_NodeResubOne` (`-l`,
  `fUseDcs`) can turn a node into a constant while it keeps its fanins, so its
  CNF then forced the first fanin to a constant and left the output free. The
  next window over that node is inconsistent: `Sfm_ComputeInterpolantInt` finds
  a minterm in both the onset and the offset and returns `l_True`, which
  `Sfm_NodeResubOne` asserts cannot happen. Release builds skip the assert, use
  `SFM_SAT_SAT` as a truth table and write counter-example bits past
  `vDivCexes`, and crash.
- **Fix:** emit the unit clause on the output, `Abc_Var2Lit(nVars, Truth == 0)`
  (and `Abc_Var2Lit(nVars, 1)`/`Abc_Var2Lit(nVars, 0)` for more than six
  inputs). For `nVars == 0`, the only case the other callers produce, the CNF
  is unchanged.
- **Test:** `abcgo/opt/sfm/cnf_oracle_test.go` (`TestTruthToCnfMatchesC`)
  compiles `Sfm_TruthToCnf` from `src/` through
  `test_vectors/sfm/cnf_oracle.c`, compares its CNFs with the Go port on random
  and constant functions of 0 to 8 inputs, and checks that each CNF holds
  exactly when the output equals the function; before the fix it reported the
  constant functions with fanins. `abcgo/base/abci/mfs_sfm_dcs_test.go`
  (`TestNativeMfsDcsConstantNodeDoesNotCrash`) runs
  `read i10.aig; &get; &if -K 4; &mfs -a -l; &ps` through the native binary
  named by `ABC_BINARY`; it segfaulted before the fix and now reports
  `lut = 798 edge = 2681 lev = 17`.
- **Go port:** the line-by-line port of the SFM engine (`opt/sfm`) had the same
  CNF (`TestTruthToCnfConstantWithFanins`) and panicked on the assert; it has
  the same fix, and `TestMfsDcsConstantNodeMatchesNative` checks that it gives
  the native result.

## 4. `supercl` violates `Abc_NodeLutMap`'s assertion on every node

- **File:** `src/base/abci/abcLut.c`, `Abc_NodeLutMap`.
- **Found:** 2026-11-26, by running every command with default arguments on
  mul8 and s27; the Go port (which keeps its assertions) panicked in `supercl`.
- **Bug:** `Abc_NodeLutMap` asserts `pObj->Level == 0` before recomputing the
  node's LUT level, but its driver `Abc_NtkSuperChoiceLut` clears only the CI
  levels (its `pObj->Level = 0;` for nodes is commented out), so every node
  still has its level from strashing. Release builds compile the assertion out;
  any assertion-enabled build aborts on the first node.
- **Fix:** remove the stale assertion. The function overwrites the level on the
  next line, so release behavior is unchanged.
- **Test:** `abcgo/base/abci/supercl_c_assert_test.go`
  (`TestNativeSuperclAssertionsHold`) compiles `abcLut.c` from `src/` with
  assertions enabled, links it ahead of the native library named by
  `ABC_LIBABC`, and runs `read mul8.aig; strash; supercl`. It aborted on the
  assertion before the fix and now reports 483 nodes, as the release build does.
- **Go port:** it had the same assertion and the same fix. In addition, its cut
  recursion now walks the live network as `abcCut.c` does (`supercl` creates
  nodes after the cut manager starts). `strash; supercl` is in the native-golden
  sweep and gives the native result on all five benchmarks.

## 5. `&print_truth` reads past its truth tables on sequential AIGs

- **File:** `src/base/abci/abc.c`, `Abc_CommandAbc9PrintTruth`.
- **Found:** 2026-11-26, by running every command with default arguments on
  s27; the Go port panicked where native `abc` printed hexadecimal noise.
- **Bug:** the command refuses AIGs with more than six *primary inputs*, but
  `Gia_LutComputeTruth6Simple` indexes `s_Truths6` by CI number and
  `Extra_PrintHex` prints `Gia_ManCiNum` variables from one 64-bit word. With
  registers there are more CIs than PIs (s27: 4 PIs, 7 CIs), so both read out
  of bounds.
- **Fix:** check the number of CIs instead of PIs (same message).
- **Test:** `abcgo/base/abci/print_truth_c_bounds_test.go`
  (`TestNativePrintTruthStaysInBounds`) compiles `giaTruth.c` and `abc.c` from
  `src/` with AddressSanitizer, links them ahead of the native library named by
  `ABC_LIBABC`, and runs `&print_truth` on s27. Before the fix ASan reported a
  global-buffer-overflow in `Gia_LutComputeTruth6Simple_rec`; now the command
  refuses the AIG.
- **Go port:** it had the same PI check and panicked on the missing CI; it has
  the same fix (`TestPrintTruthRefusesMoreThanSixCis`).

## 6. `&muxpos` adds outputs for nodes that already drive outputs

- **File:** `src/aig/gia/giaUtil.c`, `Gia_ManDupWithMuxPos`.
- **Found:** 2026-11-26, comparing `&muxpos` output counts (native 196 POs on
  mul8, the port 190).
- **Bug:** the function marks the CO drivers of the duplicate `pNew` (`fMark0`)
  so that they get no extra output, but then reads the candidate objects from
  the original `p` (`Gia_ManObj( p, ... )`), whose marks are never set. Every
  MUX-related node that already drives a CO received a second output.
- **Fix:** read the candidates from `pNew` (the IDs are `pNew`'s).
- **Test:** `abcgo/aig/gia/muxpos_c_oracle_test.go`
  (`TestNativeMuxPosSkipsCoDrivers`) links `test_vectors/muxpos_oracle.c` and
  `giaUtil.c` from `src/` against the native library named by `ABC_LIBABC` and
  runs it on mul8. Before the fix it reported 180 added outputs, 6 of them
  duplicating an existing output's driver; now 174 and none.
- **Go port:** it already skipped the CO drivers (174 outputs,
  `TestMuxPosSkipsCoDrivers`). It also asserted a combinational GIA, which
  release ABC does not enforce; that panic is removed so `&muxpos` runs on
  sequential AIGs as native ABC does.

## 7. `&mfs` destroys sequential designs

- **File:** `src/base/abci/abc.c`, `Abc_CommandAbc9Mfs` (with
  `src/aig/gia/giaMfs.c`, `Gia_ManPerformMfs`).
- **Found:** 2026-11-26, while porting `&flow3`, which runs `&mfs`; the Go port
  (which keeps `Gia_ManPerformMfs`'s assertions) panicked on s27.
- **Bug:** `Gia_ManPerformMfs` asserts a combinational AIG, but the command does
  not check. Release builds go on and return a network without registers and
  with almost no logic: after `&if -K 4`, s27 (3 flops, 12 ANDs) becomes 1 AND,
  s713 (19 flops, 184 ANDs) becomes 0, and s1494 (6 flops, 808 ANDs) becomes 9.
  `&flow3` and other scripts that call `&mfs` silently lost the design.
- **Fix:** the command refuses sequential AIGs, like its other preconditions
  (`The current AIG is sequential; "&mfs" works only for combinational
  networks.`), and leaves the AIG unchanged.
- **Test:** `abcgo/base/abci/mfs_sequential_native_test.go`
  (`TestNativeMfsRefusesSequentialAig`) runs the native binary named by
  `ABC_BINARY` on s27; before the fix the flops were gone, now the command
  refuses and `&ps` still shows `ff = 3`.
- **Go port:** it panicked in `Gia_ManPerformMfs`; the command now refuses the
  same way (`TestMfsRefusesSequentialAig`).

## 8. `&resim` without a counter-example crashes

- **File:** `src/base/abci/abc.c`, `Abc_CommandAbc9Resim`.
- **Found:** 2026-11-26, by running every command with default arguments;
  native `abc` dumped core on `&resim` for every design.
- **Bug:** `Cec_ManSeqResimulateCounter` reports a missing counter-example
  ("Counter-example is not available.") and returns, but the command then
  reads `pAbc->pCex->iFrame` with `pAbc->pCex == NULL`.
- **Fix:** update `pAbc->nFrames` only when there is a counter-example.
- **Test:** `abcgo/base/abci/resim_nocex_test.go`
  (`TestNativeResimWithoutCexDoesNotCrash`) runs
  `read s27.blif; strash; &get; &resim; &ps` through the native binary named by
  `ABC_BINARY`; it segfaulted before the fix and now prints the message and the
  unchanged AIG.
- **Go port:** it panicked on the same dereference, and its
  `Cec_ManSeqResimulateCounter` returned silently on each failed precondition;
  it now has the fix and ABC's messages (`TestResimWithoutCexDoesNotPanic`).

## 9. `saucy3` crashes on sequential AIGs

- **File:** `src/base/abci/abc.c`, `Abc_CommandSaucy` (with
  `src/base/abci/abcSaucy.c`).
- **Found:** 2026-11-26, by running every command with default arguments;
  native `abc` dumped core in `saucyGateWay` on s27.
- **Bug:** the symmetry search sizes its data (colors, dependency arrays) by the
  numbers of PIs and POs, but the functional supports it computes range over
  all CIs and COs, so register outputs and inputs index past the arrays.
- **Fix:** the command refuses AIGs with latches (`This command works only for
  combinational AIGs.`), next to its existing "only for AIGs" check.
- **Test:** `abcgo/base/abci/saucy_sequential_test.go`
  (`TestNativeSaucyRefusesSequentialAig`) runs `saucy3` on s27 through the
  native binary named by `ABC_BINARY`; it segfaulted before the fix.
- **Go port:** it panicked with an index out of range; it refuses the same way
  (`TestSaucyRefusesSequentialAig`).

## 10. `%test` crashes when no word-level design is loaded

- **File:** `src/base/wlc/wlcCom.c`, `Abc_CommandTest`.
- **Found:** 2026-11-26, by running every command with default arguments.
- **Bug:** the command's "no current design" check is commented out, and
  `Wlc_NtkExploreMem` dereferences the missing design.
- **Fix:** restore the check.
- **Test:** `TestNativeWlcTestWithoutDesign` (same file) segfaulted before the
  fix; `TestWlcTestWithoutDesign` checks the port, which panicked the same way
  and has the same fix.

## 11. `&anorm` crashes on AIGs without adder trees

- **File:** `src/proof/acec/acecNorm.c`, `Acec_Normalize`.
- **Found:** 2026-11-26, by running every command with default arguments;
  native `abc` dumped core in `Acec_InsertBox` on s27.
- **Bug:** `Acec_DeriveBox` returns NULL when it finds no adder tree, and
  `Acec_Normalize` passed it straight to `Acec_InsertBox`.
- **Fix:** report it and return a copy of the AIG, as `Acec_ManDecla` already
  does for the same case (`Cannot find arithmetic boxes.`).
- **Test:** `abcgo/base/abci/anorm_nobox_test.go`
  (`TestNativeAnormWithoutAdders`) runs `&anorm` on s27 through the native
  binary named by `ABC_BINARY`; it segfaulted before the fix.
- **Go port:** it panicked on the nil box; same fix (`TestAnormWithoutAdders`).

## 12. `&decla` crashes when the adder box does not reach every output

- **File:** `src/proof/acec/acecCl.c`, `Acec_RewriteTop` and `Acec_ManDecla`.
- **Found:** 2026-11-26, by running every command with default arguments;
  native `abc` dumped core in `Acec_RewriteReplace` on mul8. With asserts
  enabled it stops at `assert( Vec_IntSize(vRes) >= Gia_ManCoNum(p) )`.
- **Bug:** the root ranks of the adder box found in mul8 cover fewer bits than
  the AIG has outputs, so the rewritten-output vector is short and
  `Acec_RewriteReplace` reads past its end (the same happens if the first
  output is not a root: `iStart` runs off the end). On i10 the box also has
  empty root ranks, and the search for the first output read entry 0 of one
  (an assert-enabled build stops in `Vec_IntEntry`).
- **Fix:** the search skips empty ranks; `Acec_RewriteTop` returns NULL when
  it meets an empty rank or cannot cover every output; `Acec_ManDecla` prints
  `Cannot match the outputs of the arithmetic box.` and returns the AIG
  cleaned of the adders already added to it.
- **Test:** `abcgo/base/abci/decla_outputs_test.go`
  (`TestNativeDeclaWhenBoxMissesOutputs` on mul8,
  `TestNativeDeclaWithEmptyRootRanks` on i10) segfaulted before the fix.
- **Go port:** it panicked ("insufficient rewritten outputs"); same fix
  (`TestDeclaWhenBoxMissesOutputs`), and its time line now uses ABC's format.

## 13. `&deepsyn` crashes when one of its commands fails

- **File:** `src/aig/gia/giaDeep.c`, `Gia_ManDeepSyn` and `Gia_ManDeepSynOne`.
- **Found:** 2026-11-26, by running every command with default arguments.
  Without an `abc.rc`, the `compress2rs` alias that `&deepsyn` runs is
  unknown.
- **Bug:** `Gia_ManDeepSynOne` returns NULL when a command fails, and
  `Gia_ManDeepSyn` passed it to `Gia_ManAndNum`. The failure path also left
  the frame in batch mode and leaked the best AIG.
- **Fix:** `Gia_ManDeepSyn` stops and returns the best AIG so far;
  `Gia_ManDeepSynOne` restores the batch mode and frees its copy.
- **Test:** `abcgo/base/abci/deepsyn_failure_test.go`
  (`TestNativeDeepSynSurvivesFailedCommand`) segfaulted before the fix.
- **Go port:** it returned nil, so the command failed and left the partial
  result of the failed script as the current AIG; it now keeps the best AIG
  as ABC does (`TestDeepSynSurvivesFailedCommand`).

## 14. Functional-support simulation writes past an input's row

- **File:** `src/opt/sim/simSupp.c`, `Sim_UtilAssignFromFifo`.
- **Found:** 2026-11-26: the port, which checks the bound, stopped with
  "source FIFO flip would exceed simulation row" running `saucy3` on i10
  (257 inputs); native ABC silently corrupts memory instead.
- **Bug:** each SAT pattern fills a series of `nSuppWords` words of the
  2048-bit (64-word) simulation rows and flips bit `i` of input `i` in it.
  When `nSuppWords` does not divide 64 the last series is shorter, and
  flipping bit `i >= 32 * (iWordLim - iWord)` writes into the rows of the
  following objects (the rows share one allocation, so no tool reports it).
  The corrupted simulation values feed the functional-support computation
  (`saucy3`, `print_supp -s`, `bm`).
- **Fix:** flip only the bits that fit in the series.
- **Test:** `abcgo/opt/sim/fifo_flip_test.go`
  (`TestNativeFifoFlipStaysInRow`) links `opt/sim/test_vectors/fifo_oracle.c`,
  which includes `simSupp.c` from `../src`, against `ABC_LIBABC`; on i10 it
  counted 225 corrupted rows before the fix and 0 after.
- **Go port:** it panicked; it now skips the same flips
  (`TestFifoFlipStaysInRow`).

## 15. `If_ObjPerformMappingAnd` can end a node with no cuts

- **File:** `src/map/if/ifMap.c`, `If_ObjPerformMappingAnd`.
- **Found:** 2026-11-26, when Go's `renode` panicked on i10 after
  `strash; if -K 6; lutpack; strash` (the vector
  `abcgo/base/abci/test_vectors/i10_lutpack_strash.aig` is native's AIG at that
  point).
- **Bug:** in the area rounds the previous best cut is saved as the first cut
  of the set. `If_CutFilter` drops it when a new cut with fewer leaves
  dominates it, and that new cut (like every other candidate of node 709 under
  renode's pin delays) is then rejected because its delay exceeds the
  required time. The set ends empty and `assert( pCutSet->nCuts > 0 )` fails.
  Release builds compile the assertion out, copy whatever stale cut sits in
  slot 0 (here the last rejected cut) into the node's best cut, and print
  "Warning! Node ... has delay ... exceeding the required times" downstream.
- **Fix:** when the set is empty after the first iteration, put the previous
  best cut back as the only cut. Native renode on the vector now gives
  2,350 nodes (2,372 with the stale cut).
- **Test:** `abcgo/base/abci/renode_c_assert_test.go`
  (`TestNativeRenodeIfAssertionsHold`) compiles `ifMap.c` with assertions
  enabled and links it ahead of `ABC_LIBABC`. Before the fix it aborted on the
  assertion.
- **Go port:** it panicked on the same assertion. It has the same fix
  (`TestRenodeKeepsBestCutWhenAllCutsViolateRequired`), and the full script
  above now prints native's statistics.

## 16. `Str_CountBits` counts two short for words with three or more bits

- **File:** `src/aig/gia/giaStr.c`, `Str_CountBits`.
- **Found:** 2026-11-27, while bisecting why `&flow` on i10 mapped to 559
  LUTs natively and 561 in Go. The first differing step was `&blut -a -K 6`;
  the smallest reproducer is one 23-AND output cone of mul8, where native
  `&blut` gave depth 7 and Go depth 6.
- **Bug:** the function (documented "Count 1s.") returns early for 0, 1 and 2
  bits by clearing the lowest set bit twice. For three or more bits it then
  computes the popcount of the cleared word, without adding back the two bits
  it removed. Three bits count as 1, four as 2, and so on, so the count is
  not even monotonic. `Str_NtkBalanceMulti` uses it to choose which supergate
  inputs to pair by shared-fanin affinity, so `&blut` (and `&flow`, which runs
  it) pairs inputs by a wrong score.
- **Fix:** return `2 + popcount` of the remaining bits. With the fix native
  `&blut` matches the port on all four inputs tried (the mul8 cone, mul8, i10,
  and the AIG at the diverging `&flow` step), and native `&flow` on i10 gives
  the port's 561 LUTs, 2219 edges, level 7.
- **Test:** `abcgo/aig/gia/str_countbits_c_oracle_test.go`
  (`TestStrCountBitsMatchesC`) extracts `Str_CountBits` from `giaStr.c`,
  compiles it, and compares it with a popcount and with the port on boundary
  words. Before the fix it reported 7 -> 1 and 0xf -> 2, among others.
- **Go port:** already correct (`math/bits.OnesCount64`). End-to-end checks:
  `TestBalanceLutLikeNative` and `TestFlowLikeNative` in `abcgo/base/abci`.
  The same review found that the port's `Str_ManVectorAffinity` stopped the
  whole fanin collection at the limit, where C ends only the AND-fanin loop;
  that is fixed under `TestStrAffinityLimitStopsOnlyFaninLoop`.

## 17. `Io_ReadAiger` drops the names of bad-state and constraint outputs

- **File:** `src/base/io/ioReadAiger.c`, `Io_ReadAiger` (symbol table).
- **Found:** 2026-11-27, comparing `write_blif` after reading
  `abcgo/base/io/test_vectors/aiger_binary_extended.aig` (header
  `aig 3 1 1 1 1 1 1`, symbols `o0 out`, `b0 bad`, `c0 constraint`): native
  named the POs `out n3 n4`, the port `out bad constraint`.
- **Bug:** two faults in reading the symbol table.
  1. AIGER numbers `b`, `c`, `j` and `f` symbols within their own sections,
     but the reader used the number as an index into the whole PO list.
     `b0` therefore named PO 0, which `o0` had already named, and the bad
     output kept a default name.
  2. The symbol loop stopped at the first line starting with `c`, which is
     the comment section but also every constraint symbol (`c0 ...`). So no
     constraint name was ever read.
- **Fix:** offset `b`/`c`/`j`/`f` indices by the sizes of the sections
  before them (outputs, bad, constraints, justice), and end the symbol table
  only at a `c` that is not followed by a digit.
- **Test:** `abcgo/base/abci/aiger_constraint_native_test.go`
  (`TestNativeAigerSectionSymbolsLikePort`, needs `ABC_BINARY`). It reads the
  vector with native and with the port and requires
  `.outputs out bad constraint`. Before the fix native printed
  `.outputs out n3 n4`; after the first fix alone, `out bad n4`.
- **Go port:** already numbered the sections correctly (`ioa` keeps the
  symbol kinds). The same review found that the port did not name latch
  boxes `<name>L` and latch inputs `<name>_in` as the reader does; that is
  fixed under `TestAigerLatchNamesLikeNative`, and the vector's whole BLIF
  now equals native's.

## 18. `write_blif -S` with a three-LUT structure crashes in `If_CluCheck3`

- **File:** `src/map/if/ifDec16.c`, `If_CluCheck3`.
- **Found:** 2026-11-27, porting `Io_WriteBlifSpecial` and comparing
  `read base/abci/test_vectors/native_mul8_if6.blif; strash; if -S 444 -K 9;
  write_blif -S 444 out.blif`: native died with a segmentation fault and an
  empty file.
- **Bug:** `Io_NtkWriteNodeIntStruct` calls `If_CluCheckExt3( NULL, ... )`,
  which passes the NULL manager to `If_CluCheck3`. That function guards `p`
  for its hash table, but when the first two-LUT split leaves too many
  variables for the root it reads `p->pPars->nStructType`, so any node that
  needs the second leaf LUT crashes the writer.
- **Fix:** without a manager there is no structure restriction: test
  `p == NULL || p->pPars->nStructType == 0` for the "allowed" case.
- **Test:** `abcgo/base/abci/write_blif_special_native_test.go`
  (`TestWriteBlifSpecialLikeNative`, needs `ABC_BINARY`). It failed with
  "signal: segmentation fault"; now native's file and messages (nine nodes
  "not decomposable") equal the port's.
- **Go port:** `If_CluCheck3` already treated a nil manager as structure
  type 0.

## 19. `bmc3 -a` crashes when one model fails several outputs

- **File:** `src/sat/bmc/bmcBmc3.c`, `Saig_ManBmcScalable` (solve-all branch).
- **Found:** 2026-11-27, comparing `read base/abci/test_vectors/seq/s713.blif;
  strash; bmc3 -F 4 -a; print_status` with the port: native died with a
  segmentation fault.
- **Bug:** without `-x` (store CEXes) or the bridge, a failed output's entry
  is the sentinel `(Abc_Cex_t *)1`. After a SAT call the engine checks
  which other outputs the same model fails and records each one with
  `pCexDup = Abc_CexDup(pCexNew, ...); pCexDup->iPo = k;`. `Abc_CexDup`
  returns the sentinel unchanged, so the store writes through address 1 as
  soon as a second output fails under one model.
- **Fix:** set `iPo` only when `pCexDup` is not the sentinel.
- **Test:** `abcgo/base/abci/bmc3_solveall_native_test.go`
  (`TestNativeBmc3SolveAllDoesNotCrash`, needs `ABC_BINARY`). It failed with
  "signal: segmentation fault"; native now reports 19 of 23 outputs SAT
  after 4 frames.
- **Go port:** did not crash, but solve-all kept no per-output record unless
  `-x` was given, so it reported "None of the 23 outputs is found to be SAT";
  it also printed no per-output lines and reported the failing frame as
  `iFrame`. Fixed under `TestBmc3SolveAllLikeNative` and
  `TestBmc3FramesLikeNative`; `print_status` now recognizes the sentinel in
  both of its Go forms.

## 20. `&bmc` corrupts the heap writing Mf's CNF for an unminimized buffer cut

- **File:** `src/aig/gia/giaMf.c`, `Mf_ManDeriveCnf`.
- **Found:** 2026-11-27 by a differential sweep: `read
  base/abci/test_vectors/seq/s1494.blif; strash; &get; &bmc -F 4` aborted
  native with "malloc(): invalid size (unsorted)". An ASan build reported a
  heap-buffer-overflow writing `pCnf->pClauses[0]` at giaMf.c:446; with
  assertions on, `assert( Mf_CubeLit(pCubes[c], k) )` failed.
- **Bug:** each function's CNF is derived once and cached by truth-table id
  (`Mf_ManDeriveCnfs`); functions 0 and 1 (constant, buffer) are derived for
  0 and 1 variables. The writer then puts the output literal at position
  `Mf_CutSize(pCut)`. A cut that is not support-minimized (here a 4-leaf cut
  whose function is the buffer of its first leaf) therefore reads the output
  literal from an empty field: the clause loses its output, the cube's fields
  are misread, and the literals written exceed the counted `nLits`. The
  constant case had already been special-cased ("bug fix" comment); the
  buffer, and any function seen first with a different cut size, was not.
- **Fix:** take the output position from the cached CNF itself (the top
  field of every cube) and place the node's variable there; input literals
  are read up to that position.
- **Test:** `abcgo/base/abci/bmc_and_native_test.go`
  (`TestNativeAbc9BmcDoesNotCorruptHeap`, needs `ABC_BINARY`): failed with
  "signal: aborted"; now native reports output 0 asserted in frame 0, as the
  other engines do, and the ASan build is clean.
- **Go port:** unaffected: `mf.DeriveCNF` decodes each cover's output at the
  width the cover was derived for and sizes clauses from the cover.

## 21. `&reachm` and `&reachn` print the wrong frame (and can crash)

- **Files:** `src/bdd/llb/llb1Reach.c` (`Llb_ManReachability`) and
  `src/bdd/llb/llb3Nonlin.c` (`Llb_NonlinReachability`).
- **Found:** 2026-11-27: `read base/abci/test_vectors/native_s27_drw.blif;
  strash; &get; &reachm` printed "was asserted in frame -1622481680" (the
  value changed from run to run); `&reachn` printed `miter "(null)"` and a
  garbage frame. `&reachy` reports frame 0.
- **Bug:** argument lists that do not match their formats. llb1Reach.c
  passes the miter name twice, so the frame's `%d` reads the second pointer;
  llb3Nonlin.c passes no name, so `%s` reads `nIters` as a `char *`: frame 0
  printed "(null)", and any later frame dereferences a small integer.
- **Fix:** pass exactly the miter name and `nIters`.
- **Test:** `abcgo/base/abci/reach_native_test.go`
  (`TestNativeReachReportsFailingFrame`, needs `ABC_BINARY`): requires
  `Output 0 of miter "s27.bench" was asserted in frame 0.` from &reachm,
  &reachn and &reachy.
- **Go port:** the &reach* commands are not registered yet (task #39).

## 22. `Abc_Print`/`printf` reports with missing, extra or misplaced arguments

- **Files:** `src/base/abci/abc.c` (`powerdown` usage),
  `src/proof/cec/cecCore.c` (`Cec_ManSimulation`),
  `src/base/abci/abcNpnSave.c` (`Npn_ManRead`), `src/aig/gia/giaGiarf.c`
  (`Gia_ComputeEquivalences`), `src/sat/proof/pr.c` (`Pr_ManProofTraceOne`
  resolvent message), `src/proof/pdr/pdrIncr.c` (`IPdr_ManSolveInt`).
- **Found:** 2026-11-27, after bug 21, by compiling every C file with
  `-include` of an `Abc_Print` prototype carrying
  `__attribute__((format(printf,2,3)))` (ABC declares none, so gcc never
  checked these calls) and `-Wformat`.
- **Bugs:**
  - `powerdown -h`: "5% means" in a format string; `% m` is a directive, so
    native printed "(e.g. 5No such file or directoryeans hot wires ...".
  - `&equiv -s`: "Performing rounds of random simulation of %d frames with
    %d words." was passed rounds, frames and words, so it reported 20 frames
    and 100 words for 20 rounds of 100 frames with 31 words.
  - `npnload`: "Skipping token %s ..." had no argument (it printed the
    right token only because the pointer was still in the argument
    register).
  - `Gia_ComputeEquivalences`: "Iters = %3d" had no argument.
  - pr.c: the argument was inside the string literal (`"...\n, pFinal->Id"`).
  - `IPdr_ManSolveInt`: "Reached conflict limit (%d) in frame." dropped the
    frame it was passed.
- **Fix:** match each format to its arguments ("5%%", "Performing %d rounds
  ...", pass `pToken`, `nIters`, `pFinal->Id`, "in frame %d").
- **Tests:** `abcgo/base/abci/printf_format_native_test.go`:
  `TestNativeAbcPrintFormatsMatchArguments` (the gcc check over these files,
  red before), `TestNativePowerdownUsageText`, `TestEquivSeqSimulationReport`
  (native and port).
- **Go port:** copied the `&equiv -s` report (fixed); the others were
  already right. Not fixed here: pr.c keeps writing the resolvent after
  reporting that it ran out of space (no test reaches it), and several
  `%d` are given a `long`/`size_t`, which prints correctly for small values
  on LP64.

## 23. A formula ending in an operator crashes the EQN and genlib readers

- **Files:** `src/misc/parse/parseEqn.c` and `src/misc/parse/parseCore.c`
  (`Parse_ParserPerformTopOp`), `src/map/mio/mioParse.c`
  (`Mio_ParseFormulaOper`), `src/map/mio/mioFunc.c` (`Mio_GateParseFormula`).
- **Found:** 2026-11-27 by a malformed-input sweep of the readers:
  `read` of an EQN file with `y = a * ;` and `read_genlib` of a gate with
  `O=a*;` dumped core.
- **Bug:** at the closing parenthesis the parsers apply the pending binary
  operator to the top two entries of the function stack, but a formula
  ending in an operator leaves only one there. parseEqn/parseCore pop a NULL
  (after "Trying to extract data from the empty stack!") and pass it to
  `Hop_And`/`Cudd_bddAnd`; mioParse pops past the stack (ASan:
  heap-buffer-overflow in `Vec_PtrPop`). And `Mio_GateParseFormula` used the
  NULL expression a failed parse returns.
- **Fix:** `Parse_ParserPerformTopOp` and `Mio_ParseFormulaOper` return NULL
  when an operand is missing, which their callers already report as
  "Unknown operation" before failing the read; `Mio_GateParseFormula`
  returns 1 for a NULL expression, as its earlier BDD version did, so
  `read_genlib` reports "Mio_LibraryRead: Had problems parsing formulas.".
- **Test:** `abcgo/base/abci/parse_operand_native_test.go`
  (`TestNativeParserMissingOperandFails`, needs `ABC_BINARY`): both reads
  segfaulted; now they fail with the messages above.
- **Go port:** its formula parsers reject the missing operand ("expected an
  operand"); they word their errors differently.

## 24. `&fftest -N` segfaults with the formula and delay fault models

- **File:** `src/sat/bmc/bmcFault.c` (`Gia_ManFaultTest`).
- **Found:** 2026-11-27 while porting `&fftest` line by line:
  `&r native_mul4_gen.aig; &fftest -A 0 -S ((a&b)^p) -N 2` and
  `read native_s27_drw.blif; strash; &get; &fftest -A 1 -N 2` dumped core.
- **Bug:** every `-N` iterations `Gia_ManFaultTest` calls
  `Gia_ManFaultAnalyze` to find fault parameters that can be fixed to 0 and
  removes them from the fault map `vMap`. Only the stuck-at, complement and
  functionally-observable models (`-A 2..4`) have a fault map; for `-A 0`
  (formula) and `-A 1` (delay) `vMap` is empty, so the
  `assert( Vec_IntSize(vPars) == Vec_IntSize(vMap) )` fails, and in a
  release build the loop reads `vLits`/`vMap` (copies of the empty map)
  past their end.
- **Fix:** run the parameter analysis (and the miter recomputation that
  follows it) only when `Vec_IntSize(vMap) > 0`, so `-N` has no effect
  for the models without a fault map.
- **Test:** `abcgo/base/abci/fftest_native_test.go`
  (`TestNativeFftestIterCheckWithoutFaultMap`, needs `ABC_BINARY`): native
  segfaulted; now its output equals the run without `-N`, and the port's.
- **Go port:** the same condition in `Gia_ManFaultTest`
  (`sat/bmc/bmcFault.go`); the port had panicked on the assertion.

## 25. `&mfsd -K 5` and `-K 6` smash the stack

- **File:** `src/opt/sbd/sbdLut.c` (`Sbd_ProblemAddClauses`).
- **Found:** 2026-11-27 while porting `&mfsd` line by line:
  `&r i10_native_demo.aig; &mfsd -K 5` (and `-K 6`) aborted with
  "*** stack smashing detected ***", although the help allows
  `2 <= K <= 6`. ASan: stack-buffer-overflow at sbdLut.c:74.
- **Bug:** the clause for one minterm of a LUT structure holds the
  `nVarIns` input literals, the parameter and the output, but the buffer was
  `pLits[SBD_SIZE_MAX+2]` = 6 entries. `Sbd_ManExploreCut` builds LUTs with
  `nVarIns = nLutSize`, so 5- and 6-input LUTs wrote one and two literals
  past it.
- **Fix:** size the buffer `SBD_DIV_MAX+2` (LUTs have at most 6 inputs,
  selectors use 3 literals).
- **Tests:** `abcgo/base/abci/mfsd_native_test.go`:
  `TestMfsdLutSizesLikeNative` (needs `ABC_BINARY`; native aborted, now its
  &ps, `-v` report and &w output equal the port's for `-K 5`, `-K 6`,
  `-K 5 -c`, `&if -K 6; &mfsd -K 6 -p`) and `TestNativeMfsdStaysInBounds`
  (needs `ABC_LIBABC`; compiles sbdWin.c, sbdLut.c and sbdCore.c with
  AddressSanitizer and runs `&mfsd` with every LUT size).
- **Go port:** `Sbd_ProblemAddClauses` (`opt/sbd/sbdLut.go`) uses the same
  `SBD_DIV_MAX+2` buffer; a smaller one panics on the index.

## 26. `&mfsd` reads past `s_Truths6` in `Sbd_ManSolve`

- **File:** `src/opt/sbd/sbdWin.c` (`Sbd_ManSolve`).
- **Found:** 2026-11-27 with an ASan build: plain `&mfsd` on
  i10_native_demo.aig, and `&if -K 6; &mfsd`, report a
  global-buffer-overflow reading `s_Truths6` at sbdWin.c:222, called from
  `Sbd_ManExploreCut` (sbdCore.c:1308).
- **Bug:** `Sbd_ManSolve` ORs a cube into the truth table for each conflict,
  using `s_Truths6[iVar]` for every divisor in the final conflict.
  `Sbd_ManExploreCut` calls it with up to `nCutSize-1` = 9 divisors while it
  tries to drop cut leaves, so `iVar` reaches 8. The truth table is only
  meaningful for at most six divisors; for more, the caller only tests the
  result against `SBD_SAT_SAT`/`SBD_SAT_UNDEC`.
- **Fix:** add a divisor's literal to the cube only when `iVar < 6`. The
  clause added to the solver, and so the search, are unchanged; outputs of
  the release binary do not change.
- **Tests:** `TestNativeMfsdStaysInBounds` (ASan report before the fix) and
  `abcgo/opt/sbd/sbdWin_test.go` (`TestSbdManSolveMoreThanSixDivisors`: the
  port without the guard panics with "index out of range [7]").
- **Go port:** `Sbd_ManSolve` (`opt/sbd/sbdWin.go`) has the same guard.

## 27. `&mfsd -K 2` segfaults: 3-LUT structures with too many cut leaves

- **File:** `src/opt/sbd/sbdCore.c` (`Sbd_ManExploreCut`).
- **Found:** 2026-11-27 while comparing the port with native ABC:
  `&r i10_c_rw.aig; &mfsd -K 2` (and i10_native_demo.aig) segfaulted in
  `sat_solver_addclause` called from `Sbd_ProblemSolve`; `-K 3` ran but
  violated the same function's assertions.
- **Bug:** the cut size defaults to `(4-1)*3+1 = 10` and does not follow
  `-K`. The LUT(LUT,LUT) and LUT(LUT(LUT)) structures assume at most
  `3*nLutSize-2` cut leaves (10 for `K = 4`); with more, moving the surplus
  bottom leaves to the top LUT leaves it more than `nLutSize-2` inputs, so
  `nNodesDiff` becomes negative and `Sbd_ProblemSolve` gets fewer than
  three (even zero or negatively many) structures. It then reads
  `pVarsQbf[nVars]`, which was never set, and adds clauses on a garbage
  variable. (ASan does not see the uninitialized read; the release binary
  crashed.)
- **Fix:** `Sbd_ManExploreCut` returns 0 before the three-LUT structures
  when the cut has more than `3*nLutSize-2` leaves. For `K >= 4` the cut
  size is at most 10, so nothing changes there; `-K 2` and `-K 3` now run
  (on these designs their results equal the unfixed `-K 3` and the ASan
  build's `-K 2`).
- **Tests:** `TestMfsdLutSizesLikeNative` (`-K 2` on i10_c_rw segfaulted;
  `-K 3`) and `TestNativeMfsdStaysInBounds`.
- **Go port:** the same check in `Sbd_ManExploreCut`
  (`opt/sbd/sbdCore.go`); the port had failed the C assertions there.

## 28. `&mfsd` silently breaks AND/XOR/MUX graphs

- **File:** `src/base/abci/abc.c` (`Abc_CommandAbc9Mfsd`).
- **Found:** 2026-11-27 while checking `&mfsd` on the graphs `&st -m`
  makes: `&r native_mul8_canon.aig; &st -m; &mfsd -m -M 1; &st` (which
  resynthesizes nothing) fails `&cec` against the input ("Output 3
  trivially differs"); so do i10_c_rw and i10_native_demo.
- **Bug:** `Sbd_NtkPerform` handles AND and XOR nodes only. A MUX node is
  treated as the AND of its first two fanins: cut merging only asserts
  `!Gia_ObjIsMux` (compiled out in release builds), windows and SAT clauses
  ignore the control fanin, and `Sbd_ManDerive` rebuilds it with
  `Gia_ManHashAnd`, so the result loses every MUX.
- **Fix:** the command refuses a GIA with MUXes
  ("Abc_CommandAbc9Mfsd(): This command does not work with MUXes."), as it
  already does for barrier buffers. (`&st` turns the graph back into an
  AIG.)
- **Test:** `abcgo/base/abci/mfsd_native_test.go` (`TestMfsdRejectsMuxes`;
  native and port print the error; the port without the check fails the
  `!Gia_ObjIsMux` assertion). The non-equivalence was confirmed with
  a native build of abc.c without the check.
- **Go port:** the same check in `Abc_CommandAbc9Mfsd` (`base/abci/abc.go`).

## 29. `&if` maps XOR and MUX nodes as ANDs

- **File:** `src/base/abci/abc.c` (`Abc_CommandAbc9If`).
- **Found:** 2026-11-27: `&r native_mul4_gen.aig; &st -m; &if -K 4; &put;
  cec -n native_mul4_gen.aig` reports "Networks are NOT EQUIVALENT" (also
  with `-K 6`, with `&st -L 1 -m`, and on native_mul8_canon.aig), while
  `&st -m; &put` alone is equivalent.
- **Bug:** `Gia_ManToIf` (`src/aig/gia/giaIf.c`) creates exactly one mapper
  object per GIA object (`assert( i == If_ObjId(pIfObj) )`) and builds every
  internal node with `If_ManCreateAnd` from its first two fanins. An XOR
  node of a `&st -m` graph becomes the AND of its fanins and a MUX loses its
  control input, so the mapping (and the AIG derived from it) computes a
  different function. The command checks for barrier buffers and existing
  mappings but not for XORs or MUXes.
- **Fix:** the command refuses a GIA with XORs or MUXes ("This command does
  not work with XORs or MUXes. Run "&st"."), as `&mfsd` now does for MUXes
  (bug 28).
- **Test:** `abcgo/base/abci/if_muxes_native_test.go`
  (`TestIfRejectsMuxGraphsLikeNative`; red before: native `cec` reported
  NOT EQUIVALENT for all six flows; now native and port print the error).
- **Go port:** its `Gia_ManToIf` builds XORs and MUXes with
  `If_ManCreateXor`/`If_ManCreateMux` and so mapped correctly, but it now
  refuses such graphs in `Abc_CommandAbc9If` as C does.

## 30. `&pack -N 5` then `&edge -p` or `&ps` smashes the stack

- **Files:** `src/aig/gia/giaEdge.c` (`Gia_ManConvertPackingToEdges`),
  `src/aig/gia/giaIf.c` (`Gia_ManPrintPackingStats`).
- **Found:** 2026-11-27, checking the fixed arrays of giaEdge.c with an ASan
  build while porting `&edge`.
- **Bug:** both functions assert that a packing block has one to three LUTs
  (`assert( Entry > 0 && Entry < 4 )`), but `&pack -N num` accepts any
  block size of at least 2 and `Gia_ManLutPacking` fills blocks up to it.
  Release builds compile the assert out: `Gia_ManConvertPackingToEdges`
  copies every block member into `int nNodes[4]` and `&ps` increments
  `nNumStr[Entry]` of `int nNumStr[5]`. With `-N 5` or more both write past
  the stack arrays (ASan: stack-buffer-overflow at giaEdge.c:135 and
  giaIf.c:593; release `abc` aborts with "stack smashing detected" on
  `&if -K 4; &pack -N 8; &edge -p`).
- **Fix:** store only the first four members (`if ( k < 4 )`) and count only
  block sizes below 5 (`if ( Entry < 5 )`). Blocks of two or three LUTs, the
  only ones that produce edges or are printed, give the same results as before;
  larger blocks get no edges, as blocks of four already did.
- **Test:** `abcgo/base/abci/edge_native_test.go`: `TestNativeEdgeStaysInBounds`
  (compiles giaEdge.c and giaSatEdge.c with ASan against `ABC_LIBABC`) and
  `TestEdgeCrashesFixedLikeNative` (`&pack -N 5/8`, `&if -K 6; &pack -N 6`,
  on three designs, through `ABC_BINARY`); both failed before the fix.
- **Go port:** `Gia_ManConvertPackingToEdges`, `Gia_ManPrintPackingStats`,
  `Gia_ManTransferPacking` and the AIGER "k" writer panicked or refused blocks
  of more than three LUTs; they now follow the fixed C. `Gia_ManLutPacking`
  (giaPack.c) was ported line by line too (its stable sort of fanin delays
  differed from C's selection sort, so `&pack` built other blocks).

## 31. `&edge -D` below the feasible delay segfaults

- **File:** `src/aig/gia/giaSatEdge.c` (`Seg_ManComputeDelay`).
- **Found:** 2026-11-27, comparing `&edge` options with the port.
- **Bug:** the SAT search starts at the delay limit `-D`; when that limit is
  below the smallest delay achievable with edges, the first call is UNSAT and
  no assignment is saved. `Gia_ManEdgeFromArray( p->pGia, vEdges2 )` is then
  called with `vEdges2 == NULL` and dereferences it, so
  `&if -K 4; &edge -D 5` segfaults on i10.
- **Fix:** call `Gia_ManEdgeFromArray` only when an assignment was found; the
  network keeps the edges it had (`-v` already reports "Proved UNSAT for
  delay N").
- **Test:** `abcgo/base/abci/edge_native_test.go`
  (`TestEdgeCrashesFixedLikeNative`, `TestEdgeOptionsLikeNative` with `-D 9`,
  `TestNativeEdgeStaysInBounds`); native `abc` segfaulted before the fix.
- **Go port:** `Seg_ManComputeDelay` (now a line-by-line port of giaSatEdge.c)
  has the same guard.

## 32. `&acec -m` reads past `vXorRoots` when the adder trees form no chain

- **File:** `src/proof/acec/acecXor.c` (`Acec_OrderTreeRoots`,
  `Acec_ProduceBox`).
- **Found:** 2026-11-27: `&r mul4; &miter cec_mul4_syn2.aig` gives a
  dual-output miter of two independent multipliers; `&acec -m` on it
  segfaults (in `Acec_RankTrees`, then, with only the read fixed, in
  `Acec_FindBox`).
- **Bug:** `Acec_OrderTreeRoots` looks for the XOR tree that no other tree
  feeds. When there is none it relies on `assert( i < Vec_IntSize(vMove) )`,
  compiled out in release builds, and reads `vXorRoots[i]` one past the end;
  `Acec_RankTrees` then writes through that garbage root. The rest of
  `Acec_ProduceBox` also assumes a chain.
- **Fix:** `Acec_OrderTreeRoots` returns NULL when the trees form no chain,
  and `Acec_ProduceBox` then returns no box, so `Acec_Solve` prints "Cannot
  find arithmetic boxes in both LHS and RHS. Trying regular CEC." and
  verifies the miter normally (here: "Networks are equivalent.").
- **Test:** `abcgo/base/abci/acec_order_roots_native_test.go`
  (`TestAcecOrderRootsWithoutChainLikeNative`, vector
  `test_vectors/acec_mul4_dual_miter.aig`; red before: native segfault, port
  panic).
- **Go port:** it panicked ("acec: root ordering requires an XOR chain") at
  the same point; it now returns nil and no box as C does. Four Go tests that
  expected the panic now expect no box / the CEC fallback.

## 33. `&gla` and `&gla -n` on an output driven by a primary input

- **Files:** `src/proof/abs/absGla.c` (`Gia_ManPerformGla`),
  `src/proof/abs/absGlaOld.c` (`Gia_ManPerformGlaOld`).
- **Found:** 2026-11-27, porting absGla.c: s713 has an output that is a
  primary input, and `&gla` runs per output.
- **Bug:** the initial abstraction holds the output's driver, but only flops
  and AND nodes get abstraction IDs. `Gia_ManPerformGla` then reads the
  output literal with `Ga2_ObjFindLit`, i.e. `vIds[-1]` and the frame map
  at that index (release builds; the `assert` is compiled out), and prints
  whatever follows ("GLA finished 10 frames and produced a 6-stable
  abstraction." for an output that fails in frame 0).
  `Gia_ManPerformGlaOld` reads the PI's empty `vFrames` in
  `Gla_ManGetOutLit` and segfaults.
- **Fix:** both procedures treat an output driven by a PI as they treat a
  constant-1 output: a frame-0 counter-example (the PI set to 1 unless the
  output is complemented) and "Sequential miter is trivially SAT.".
- **Test:** `abcgo/base/abci/gla_native_test.go`
  (`TestGlaPrimaryInputOutputLikeNative`; `TestGlaOldLikeNative` on s713);
  red before: native printed "GLA finished ..." or segfaulted, the port
  panicked ("abs: invalid abstraction ID").
- **Go port:** the same checks in `Gia_ManPerformGla` and
  `Gia_ManPerformGlaOld` (the port's earlier PI shortcut printed nothing).

## 34. `&gla -n` segfaults in `Rnm_ManFilterSelected`

- **File:** `src/proof/abs/absRef.c` (`Rnm_ManRefine`).
- **Found:** 2026-11-27, comparing `&gla -n` with the port: native abc
  segfaulted on s27, s713 and s1494 whenever a refinement was not empty.
- **Bug:** `Rnm_ManRefine` filters the selected pseudo-PIs with
  `Rnm_ManFilterSelected` (or `...New` under `-r`), which reads the cut
  leaves through `Ga2_ObjLeaves`, i.e. the GIA's `vMapping` built by
  `Ga2_ManMarkup` for the new engine. The old engine's working GIA
  (`Gla_ManStart`, `Gia_ManDupMapped`) has no `vMapping`, so
  `Vec_IntEntry(NULL, ...)` crashes.
- **Fix:** filter only when the GIA has a mapping; the old engine uses the
  justified pseudo-PIs as they are.
- **Test:** `abcgo/base/abci/gla_native_test.go` (`TestGlaOldLikeNative`);
  red before: native segfault.
- **Go port:** it had built a Ga2-style mapping on the old engine's GIA
  just for the filter; that is removed and `Rnm_ManRefine` has the same
  guard.

## 35. `&gla -a` divides by zero when every pseudo-PI is a PI

- **Files:** `src/proof/abs/absGla.c` (`Ga2_ManRefine`),
  `src/proof/abs/absGlaOld.c` (`Gia_ManPerformGlaOld`).
- **Found:** 2026-11-27, `read_truth 8; strash; &get; &gla -a -F 1` (and
  `&gla -n -a`) on a combinational AND: native SIGFPE.
- **Bug:** with `-a` the refinement is "all pseudo-PIs that are not PIs".
  When the abstraction is complete that set is empty, but it is still used
  as a refinement: the proof logger is restarted with width 0
  (`Prf_BitWordNum(0) == 0`) and `Prf_ManSize` divides by `nWords`. With
  `-p` (old engine, full proofs) the same SAT query repeats until `-T`.
- **Fix:** when the layer is empty the counter-example is checked as
  without `-a` (`Rnm_ManRefine`, which reports it as a true
  counter-example).
- **Test:** `abcgo/base/abci/gla_native_test.go`
  (`TestGlaAddLayerAllPisLikeNative`), `abcgo/proof/abs/absGlaOldLayer_test.go`
  (`TestOldGlaEmptyLayerChecksCounterexample`, which previously asserted the
  wait-for-timeout behavior); red before: native SIGFPE.
- **Go port:** same change in `Ga2_ManRefine` and `Gia_ManPerformGlaOld`.

## 36. `write_cex` after `bmc3 -a` (without `-x`) segfaults

- **File:** `src/base/io/io.c` (`IoCommandWriteCex`).
- **Found:** 2026-11-27, comparing `bmc3 -a; write_cex` with the port on
  s27, s713 and s1494: native segfaulted.
- **Bug:** when solving all outputs without storing CEXes, `bmc3` (and
  other solve-all engines) record a failed output as the sentinel
  `(Abc_Cex_t *)1` in the CEX vector. `write_cex` skips only `NULL` entries
  and passes the sentinel to `Abc_NtkDumpOneCex`, which dereferences it.
- **Fix:** skip the sentinel like a missing CEX; the file then holds only
  `# DONE`.
- **Test:** `abcgo/base/abci/bmc3_native_test.go`
  (`TestBmc3WriteCexLikeNative`); red before: native segfault.
- **Go port:** it printed the "CEX for output" header and then failed with
  "counterexample dimensions do not match network"; it now skips the
  sentinel too.

## 37. `bmc3 -a` reads a missing SAT model when an output is constant 1

- **File:** `src/sat/bmc/bmcBmc3.c` (`Saig_ManBmcScalable`,
  `Saig_ManGenerateCex`).
- **Found:** 2026-11-27, `read s713.blif; strash; bmc3 -a -g -F 8`: native
  segfault in `glucose_solver_read_cex_varvalue` (ASan: NULL read).
- **Bug:** when ternary simulation proves an output constant 1 in a frame,
  `Saig_ManCallSolver` returns `l_True` without calling the solver
  (`Lit == 1`). The SAT branch then derives the CEX from, and tests every
  other output against, the solver's model, which belongs to no solve:
  glucose has an empty model (NULL dereference), `sat_solver` returns
  uninitialized (`ABC_REALLOC`ed) or stale values from an earlier frame, so
  other outputs were declared failed by a meaningless assignment.
- **Fix:** with `Lit == 1` the CEX keeps all PI bits 0 (any assignment
  fails a constant-1 output) and the other outputs are not tested against a
  model; each is solved in its own turn.
- **Test:** `abcgo/base/abci/bmc3_native_test.go`
  (`TestBmc3SolveAllConstantOutputLikeNative`); red before: native segfault.
  On s713 and s1494 the default engine's `bmc3 -a` report is unchanged
  (the real solves find the same outputs); `TestBmc3LikeNative` and
  `TestBmc3SolveAllLikeNative` cover it.
- **Go port:** the same `fModel` guard in `sat/bmc/bmcBmc3.go`.

## 38. `bmc3 -W file` segfaults unless `-a -H` is given

- **File:** `src/sat/bmc/bmcBmc3.c` (`Saig_ManBmcScalable`).
- **Found:** 2026-11-27, `bmc3 -F 3 -W log` on the s1494 miter: native
  segfault.
- **Bug:** the per-output log line reads `p->pTime4Outs[i]` for every
  solved output, but that array is allocated only for `-H` with `-a`
  (`nTimeOutOne` is reset otherwise), so it is `NULL`.
- **Fix:** print 0 for the remaining-time column without `pTime4Outs`.
- **Test:** `abcgo/base/abci/bmc3_native_test.go`
  (`TestBmc3LogFileLikeNative`); red before: native segfault.
- **Go port:** same guard.

## 39. `&bmc` drops every frame before the last one that needs no earlier state

- **File:** `src/sat/bmc/bmcBmcAnd.c` (`Gia_ManBmcUnroll`).
- **Found:** 2026-11-27, porting `&bmc`: on s713, `&bmc` reported output 8
  failing in frame 1, `&bmc -F 1` in frame 0.
- **Bug:** the backward pass over the ternary-simulated frames collects each
  frame's x-valued outputs plus the register leaves wanted by the later
  frames, and stopped (`break`) at the first frame whose cones need no
  register values from the frame before. The forward pass then turned every
  output of all earlier frames into constant 0, though those frames have
  x-valued outputs of their own. Failures in those frames were never
  checked, while `pPars->iFrame` (print_status "Frames") counted them as
  explored: `bmc_and_early.blif` (fails only in frame 0) gave
  `Status = -1  Frames = 49`; bmc3 and `&bmc -F 1` find the frame-0 CEX.
- **Fix:** do not stop the backward pass when the leaves run out (and drop
  the `assert( f >= 0 )` that depended on it).
- **Test:** `abcgo/base/abci/gia_bmc_native_test.go`
  (`TestNativeAbc9BmcFindsEveryFailure`, `TestAbc9BmcBugCasesLikeNative`;
  test vector `base/abci/test_vectors/seq/bmc_and_early.blif`); red before.
  Every CEX checked with `testcex` ("The cex is correct").
- **Go port:** `sat/bmc/bmcBmcAnd.go` (`Gia_ManBmcUnroll`), the same change.

## 40. `&bmc` never reports an output that ternary simulation proves 1

- **File:** `src/sat/bmc/bmcBmcAnd.c` (`Gia_ManBmcUnroll`,
  `Gia_ManBmcPerformInt`).
- **Found:** 2026-11-27, with #39: bmc3 reports s713 output 0 failing in
  frame 0 (for any inputs), `&bmc` never did.
- **Bug:** the unrolling keeps only x-valued outputs and sets the others to
  constant 0, including outputs whose ternary value is 1. The engine's own
  handling of a constant-1 output was to print "Output %d is trivially SAT."
  and go on, without a CEX or a SAT status. `bmc_and_const1.blif` (output 0
  is 1 in frame 0) gave `Status = -1` with `&bmc -F 1`.
- **Fix:** an output with ternary value 1 is unrolled as constant 1; a
  constant-1 output is SAT (`RetValue = 0`) and, as there is no SAT call
  and no model, gets the all-zero-input CEX (as bmc3 does, #37).
- **Test:** as for #39 (`bmc_and_const1.blif`, s713); red before.
- **Go port:** same change.

## 41. `&bmc` loops forever when no output is x-valued, or with `-A 0`

- **File:** `src/sat/bmc/bmcBmcAnd.c` (`Bmc_MnaTernary`).
- **Found:** 2026-11-27, `&bmc -A 0` on s27: native ran out of memory.
- **Bug:** without `-F`, ternary simulation stops `nFramesAdd` frames after
  the first frame with an x-valued output, testing `f == iFirst +
  nFramesAdd` at the start of frame `f > iFirst`: never true for
  `-A 0`. When no output ever becomes x-valued (`bmc_and_zero.blif`: an
  output that is 0 in every frame) there is no stop at all. Each frame
  allocates a state, so abc runs until it is killed or out of memory.
- **Fix:** test `f >= iFirst + nFramesAdd`; while no output is x-valued,
  stop when the new ternary state repeats an earlier one (then no output
  ever will be), for which the states are now zero-allocated (`ABC_CALLOC`:
  `memcmp` reads the unused bits of the last word).
- **Test:** as for #39 (`bmc_and_zero.blif`, `s27.blif` with `-A 0`); red
  before: native killed after the test's 10-second limit.
- **Go port:** same change (`bmcMnaStateSeen`).

## 42. `int -b` (without `-t`) reports safe properties as failing

- **File:** `src/proof/int/intCore.c` (`Inter_ManPerformInterpolation`).
- **Found:** 2026-11-27, porting the int command: `int -b` on the
  s27 dc2-miter (and on `int_backward_safe.blif`, a register that stays 0)
  printed "Error: The problem should be SAT but it is UNSAT." and "Output -1
  of miter ... was asserted in frame 1" (`Status = 0`); on the s713
  dc2-miter it ran bmc3 forever looking for the nonexistent CEX.
- **Bug:** backward interpolation takes its initial interpolant from
  `Inter_ManStartOneOutput( pAig, 1 )`, whose CIs include the extra control
  input of the transition loop, but without `-t` the transition relation is
  `Inter_ManStartDuplicated` (one CI fewer). `Inter_ManDeriveSatSolver`
  then connects the interpolant's CIs to the wrong register variables (its
  `assert( Aig_ManCiNum(pAig) == Aig_ManCiNum(pInter) )` fails, but the
  release build has `-DNDEBUG`), so the first SAT run is SAT and taken as a
  real counterexample. (The usage says "-b ... (works with -t)".)
- **Fix:** like `-K` > 1, `-b` enables the transition loop
  (`if ( pPars->fUseBackward ) pPars->fTransLoop = 1;`), so `int -b` runs
  as `int -b -t`.
- **Test:** `abcgo/base/abci/int_native_test.go`
  (`TestNativeIntBackwardProves`; test vector
  `base/abci/test_vectors/seq/int_backward_safe.blif`); red before.
- **Go port:** `proof/int/intCore.go`, the same change.

## 43. `dsd` crashes on every design (pointers truncated to `int`)

- **Files:** `src/bdd/dsd/dsdInt.h`, `dsd.h`, `dsdApi.c` (the `Mark` field
  and `Dsd_NodeReadMark`/`Dsd_NodeSetMark`), `src/base/abci/abcDsd.c`.
- **Found:** 2026-11-27: `read s27; strash; dsd` (and mul4, mul8, i10;
  with `-r`, `-g`) dumped core in `Abc_ObjAddFanin` (ASan: SEGV, called
  from `Abc_NtkDsdConstructNode`).
- **Bug:** `Abc_NtkDsdConstruct`/`Abc_NtkDsdConstructNode` store the new
  network's `Abc_Obj_t *` in each DSD node's `int Mark`
  (`Dsd_NodeSetMark( pNodeDsd, (int)(ABC_PTRINT_T)pNodeNew )`) and read it
  back as a pointer. On 64-bit hosts the upper 32 bits are lost, so the
  fanin pointers are garbage.
- **Fix:** `Mark` and its accessors are `ABC_PTRINT_T` (pointer-sized); the
  `(int)` casts in `abcDsd.c` are dropped. The other user (`giaClp.c`)
  stores literals, unaffected.
- **Test:** `abcgo/base/abci/dsd_native_test.go`
  (`TestNativeDsdDoesNotCrash`); red before (segmentation fault).
- **Go port:** keeps the objects themselves (no truncation). Its `dsd`
  output differed from C's (to be ported; `dsd_like_native_test.go`).

## 44. `match -m` crashes when the miter cannot be demitered

- **Files:** `src/base/abci/abcDar.c` (`Abc_NtkDarMatch`).
- **Found:** 2026-11-27: `read s713.blif; strash; dc2; miter s713.blif;
  match -m` dumped core (segmentation fault) after printing the miter's
  statistics (the single-output miter of `miter` has no output pairs to
  demiter).
- **Bug:** `Abc_NtkDarMatch` declares `Aig_Man_t * pManRes` uninitialized
  and passes `&pManRes` to `Saig_StrSimPerformMatching`, which returns
  `NULL` without setting it when `Saig_ManDemiterSimpleDiff` fails
  ("Demitering has failed."); `Abc_NtkFromAigPhase( pManRes )` then reads
  a garbage pointer.
- **Fix:** `pManRes = NULL`; when it stays `NULL`, the managers are freed
  and `NULL` is returned, so the command reports "Matching has failed."
- **Test:** `abcgo/base/abci/match_native_test.go`
  (`TestMatchLikeNative`, the script `miter s713.blif; match -m`); red
  before (native abc failed: segmentation fault).
- **Go port:** `base/abci/abcDar.go` (`Abc_NtkDarMatch`), the same change.

## 45. `print_box` crashes when no box library was read

- **File:** `src/map/if/ifCom.c` (`If_CommandPrintBox`).
- **Found:** 2026-11-27, native crash sweep (every command with default
  arguments on s27): `print_box` without a preceding `read_box` dumped core
  in `If_LibBoxPrint` (ASan: SEGV reading `p->vBoxes` of a NULL library).
- **Bug:** the command passes `Abc_FrameReadLibBox()` to `If_LibBoxPrint`
  without checking it; there is no box library until `read_box`.
- **Fix:** print "Box library is not available." and return 1 when the
  library is NULL.
- **Test:** `abcgo/base/abci/print_box_native_test.go`
  (`TestPrintBoxLikeNative`, red before: segmentation fault; also compares
  the printed library after `read_box test_vectors/box/two.box`).
- **Go port:** printed nothing without a library and its own format with
  one; `If_CommandPrintBox` now reports the missing library and prints
  `If_LibBoxPrint`'s format (`ifLibBoxPrintNative`).
- **Note:** `If_LibBoxPrint` writes each delay row's newline with `printf`
  (to stdout) rather than to `pFile`, so `If_LibBoxWrite` would write rows
  without line breaks; no command calls `If_LibBoxWrite`, so it is left as
  is.

## 46. `&cfraig` overflows the heap on sequential AIGs

- **File:** `src/aig/gia/gia.h` (`Gia_ObjSimWords`).
- **Found:** 2026-11-27, native crash sweep: `read s27; strash; &get;
  &cfraig` aborted (s713 too); ASan: heap-buffer-overflow WRITE in
  `Ssc_GiaTransferPiPattern` (sscSim.c), from `Ssc_PerformSweepingInt`.
- **Bug:** `Ssc_GiaResetPiPattern` sizes `vSimsPi` as `nWords *
  Gia_ManCiNum(p)` and every ssc user indexes it by CI (`Gia_ManForEachCi`,
  `Vec_WrdDoubleSimInfo(..., Gia_ManCiNum)`), but `Gia_ObjSimWords` divided
  the size by `Gia_ManPiNum`. With latches (CIs > PIs) the per-CI stride was
  too large (s27: 35 words / 4 PIs = 8 instead of 5), so `Gia_ObjSimPi(p, i)`
  for later CIs pointed past the array. Combinational AIGs were unaffected.
- **Fix:** `Gia_ObjSimWords` divides by `Gia_ManCiNum` (its only users are
  in `proof/ssc`).
- **Test:** `abcgo/base/abci/cfraig_native_test.go` (`TestCFraigLikeNative`;
  s27, s713, i10: output and written AIG equal native); red before (native
  abort).
- **Go port:** `proof/ssc` keeps a per-CI stride; its results equal C's.

## 47. `kcs` crashes on a network without a `csLiveTarget_` output

- **File:** `src/proof/live/kliveness.c` (`Abc_CommandCS_kLiveness`).
- **Found:** 2026-11-27, native crash sweep: `read s27; strash; kcs` (and
  `-c`, `-m`, `-C`, `-g`; also on a logic network) dumped core in
  `createConstrained0LiveCone` (kLiveConstraints.c, ASan SEGV reading the
  NULL consequent).
- **Bug:** `collectCSSignals` takes the liveness target from an output named
  `csLiveTarget_*` and only `assert`s that one exists; the release build
  continues with a NULL consequent and dereferences it.
- **Fix:** the command checks for a `csLiveTarget_` output after building the
  AIG and otherwise prints "The network has no liveness target output
  (named "csLiveTarget_...")." and returns 1.
- **Test:** `abcgo/base/abci/kcs_native_test.go`
  (`TestKcsWithoutTargetLikeNative`; red before: segmentation fault).
- **Go port:** printed its own error ("live: csLiveTarget_ output missing")
  and skipped the check for `-m`; `Abc_CommandCS_kLiveness` (base/abci) now
  prints C's note for non-strashed input and C's message, for every flag.

## 48. `&acec`, `&verify` and `&test` crash when no AIG is loaded

- **File:** `src/base/abci/abc.c` (`Abc_CommandAbc9Acec`,
  `Abc_CommandAbc9Verify`, `Abc_CommandAbc9Test`).
- **Found:** 2026-11-27, native crash sweep: `read s27; strash; &acec`
  (also `&verify`, `&test`) without `&get` dumped core reading
  `pAbc->pGia->pSpec` / `Gia_ManLevelNum(NULL)`.
- **Bug:** `&acec` checks `pAbc->pGia` only in its `-m` branch, `&verify`
  not at all, and `&test` has the check commented out.
- **Fix:** each prints "Abc_CommandAbc9...(): There is no AIG." and returns 1.
- **Test:** `abcgo/base/abci/no_aig_native_test.go`
  (`TestNoAigCommandsLikeNative`); red before (segmentation fault).
- **Go port:** panicked the same way; the three commands now check, as C.

## 49. `&verify` crashes when the spec cannot be read

- **Files:** `src/aig/gia/giaAiger.c` (`Gia_AigerRead`), `src/aig/gia/giaTim.c`
  (`Gia_ManVerifyWithBoxes`).
- **Found:** 2026-11-27, after #48: `read s27.blif; strash; &get; &verify`
  (the default spec is `pGia->pSpec`, the BLIF file) and
  `&verify /nonexistent.aig` dumped core.
- **Bug:** `Gia_AigerRead` calls `fread`/`fclose` on the result of `fopen`
  without checking it (a missing file crashes in `fclose(NULL)`, after
  `Gia_FileSize` has already reported the file unavailable), and
  `Gia_ManVerifyWithBoxes` uses the spec without checking for NULL (a non-AIGER
  file: "The parameter line is in a wrong format.", then a NULL dereference).
- **Fix:** `Gia_AigerRead` returns NULL when the file cannot be opened;
  `Gia_ManVerifyWithBoxes` prints `Cannot read spec AIG "<file>".` and
  returns when the spec is NULL.
- **Test:** `abcgo/base/abci/no_aig_native_test.go`
  (`TestVerifyUnreadableSpecLikeNative`); red before (segmentation faults).
- **Go port:** reported its own error texts ("Boxed verification: ...",
  "Gia_AigerRead(): gia: invalid AIGER header"); the reader's header error is
  now `gia.ErrAigerParamLine` and callers print C's messages
  (`gia.GiaAigerReadMessage`), and `&verify` prints `giaTim.c`'s messages.

## 50. `bm` and `bm2` crash on sequential networks

- **File:** `src/base/abci/abc.c` (`Abc_CommandBm`, `Abc_CommandBm2`).
- **Found:** 2026-11-27, native crash sweep: `read s27; strash; bm` (heap
  overflow in `getDependencies`, abcBm.c) and `bm2` (crash in
  `saucyGateWay`/`Abc_NtkDup`).
- **Bug:** Boolean matching is combinational, but neither command rejects
  latches. On s27 the functional supports from `Sim_ComputeFunSupp` include
  the latch outputs (CIs), which `getDependencies` uses as indices into the
  per-PI `iDep` array.
- **Fix:** both commands print "Boolean matching works only for combinational
  networks (use "comb")." and return 1 when either network has latches.
- **Test:** `abcgo/base/abci/bm_native_test.go`
  (`TestBmSequentialLikeNative`; s27 and mul4 x bm, bm -P, bm2); red before
  (native segmentation fault).
- **Go port:** panicked the same way; both commands now check, as C.

## 51. `cascade` overflows two buffers (stack value and file names)

- **Files:** `src/bdd/extrab/extraBddCas.c` (`Extra_ProfileWidth`),
  `src/bdd/cas/casCore.c` (`Abc_CascadeExperiment`, `Abc_CascadeExperiment2`
  file names).
- **Found:** 2026-11-27, native crash sweep: `read <absolute path>/s27.blif;
  strash; cascade` aborted ("buffer overflow detected"); under ASan also a
  stack-buffer-overflow in `st__gen` from `Extra_ProfileWidth`.
- **Bug 1:** `st__foreach_item( tNodeTopRef, gen, ..., (char**)&LevelStart )`
  stores the pointer-sized record into `int LevelStart`, overwriting the
  stack on 64-bit hosts (silently in the release build here).
- **Bug 2:** the encoded and LUT file names are built with `strcpy`/`strcat`
  of the design's generic name into `char FileNameIni[100]`/`FileNameFin[100]`;
  a long path overflows them (fortified release build: abort).
- **Fix:** read the record into a `char *` and convert it to `int`; build the
  file names with `snprintf( ..., sizeof(...), "%s_ENC.blif", ... )`.
- **Test:** `abcgo/base/abci/cascade_native_test.go` (`TestCascadeLikeNative`;
  s27, mul4 and s27 by absolute path, `cascade` and `cascade -v`: output and
  BLIF equal native); red before (native abort on the long path).
- **Go port:** no overflow (Go strings and maps); `cascade -v` was missing
  C's "(%) Memory" suffix and the decomposition/encoding time lines, now
  printed (`bdd/cas/casDec.go`).

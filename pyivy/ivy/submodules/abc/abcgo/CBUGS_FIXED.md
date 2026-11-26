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

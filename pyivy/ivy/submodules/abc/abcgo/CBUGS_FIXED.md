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

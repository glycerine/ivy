# opt/sbd port

`&mfsd` (`Sbd_NtkPerform`) is a line-by-line port of `src/opt/sbd` that runs
on the user's `Gia_Man_t`, as C does: it appends the resynthesized logic to
that manager, keeps C's object IDs, mirrors, cut lists and LUT levels, and
derives the result with `Sbd_ManDerive`. Its `&ps`, `-v` report (times
masked) and `&w` output equal native ABC's (`base/abci/mfsd_native_test.go`).

| C file      | Go file      | Notes |
|-------------|--------------|-------|
| sbd.h       | sbd.h.go     | `Sbd_Par_t` with C field names (exported). |
| sbdInt.h    | sbdInt.go    | |
| sbdCore.c   | sbdCore.go   | Manager, windows, explore, implement, derive. |
| sbdCut.c    | sbdCut.go    | Incremental cuts; cut lists in C's flat `Vec_Wec` format. |
| sbdCut2.c   | sbdCut2.go   | Cut server. |
| sbdWin.c    | sbdWin.go    | Window SAT solver, `Sbd_ManSolve`. |
| sbdLut.c    | sbdLut.go    | QBF synthesis of LUT structures. |
| sbdPath.c   | sbdPath.go   | Critical path of a mapping (`-p`). |
| sbdSim.c    | sbdSim.go    | Experimental; simulation words live in an `sbdSim` value instead of `Gia_Man_t` fields. |
| sbdCnf.c    | sbdCnf.go    | |
| sbdSat.c, sbd.c | sbdSat.go, sbd.c.go | Earlier ports (sbdSat.go deviates from the C experiment, see git history); not used by `&mfsd` and not re-checked here. |

Boundary helpers: `gia_bridge.go` (GIA accessors) and `sat_bridge.go`
(`sat_solver_*` on `bsat.Solver` through its C-API bridges, so a failed
`sat_solver_addclause` does not poison the solver; `sat_solver_random_polarity`
keeps C's indexing, which randomizes only variables 0..63).

Details that matter for equality with C:

- `Dsm_ManTruthToGia` is called with C's "pretend to have MUXes" bracket
  (`sbdTruthToGia`): a manager with XORs but no MUX array gets a temporary
  one so XORs are built as XOR nodes.
- `Sbd_StoStoreResult` appends to the object's cut list and
  `Sbd_StoInitResult` reuses the previous node's cut storage, as in C.
- `-C` is parsed but, as in C, never reaches the SAT calls (`nBTLimit = 0`).
- C assertions that release builds compile out and that fail for legal
  non-default options (`Sbd_CutCompare2`'s `nLeaves > 4` with `-K 2/3 -c`,
  `Sbd_ManImplement2`'s delay check with `-K 5/6 -c` and `-K 3 -d`) are
  behind `sbdDebugAsserts`. The quadratic cut-set check is behind
  `sbdCheckCuts`; tests enable it.

C bugs fixed in both (CBUGS_FIXED.md): #25 `Sbd_ProblemAddClauses` literal
buffer too small for 5/6-input LUTs; #26 `Sbd_ManSolve` read past
`s_Truths6` with more than six divisors; #27 `Sbd_ManExploreCut` built
three-LUT structures for cuts larger than `3*nLutSize-2` (`-K 2/3`). #28 (in `Abc_CommandAbc9Mfsd`): MUX graphs are refused, since
`Sbd_NtkPerform` turns MUX nodes into ANDs.

## Verification

```
go test ./opt/sbd
ABC_BINARY=$PWD/.scratch/abc-release/abc ABC_LIBABC=$PWD/.scratch/libabc.a go test ./base/abci -run Mfsd
```

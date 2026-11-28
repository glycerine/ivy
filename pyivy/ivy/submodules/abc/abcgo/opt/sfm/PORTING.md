# SFM port status

The C engine of `mfs2`/`&mfs` (sfmCore.c, sfmNtk.c, sfmWin.c, sfmSat.c,
sfmCnf.c) is ported in this package around `Sfm_Ntk_t`.

The `mfs3` engine is a line-by-line port of the C sources:

- `sfmLib.go` here: sfmLib.c (the one- and two-gate implementation library,
  `Sfm_LibPrepare`, matching and implementation).
- `base/abc/sfmDec.go`: sfmDec.c (`Abc_NtkPerformMfs3`, `Sfm_DecStart`,
  window extraction, SAT-based decomposition, insertion and statistics).
- `base/abc/sfmTim.go`: sfmTim.c (the timing manager of the delay mode).
- `base/abc/sfmMit.go`: sfmMit.c (place holders; `Sfm_MitStart` returns nil).
- `base/abc/sfmArea.go`: sfmArea.c (`Abc_NtkChangePerform`, run by
  `mfs3 -a` and `phase_map`).

The decomposition engine and timing manager live in `base/abc` because, as
in C, they work on the mapped `Abc_Ntk_t` directly (they need its object
IDs, levels, traversal IDs and fanout counts).

Two asserts that release builds of ABC skip are not checked either: the
`DelayMin` check in `Abc_NtkDelayOpt` (it fails with `-R` > 1) and the
superset checks of `Abc_TtExpand` in `Sfm_DecCombineDec` (they fail with
`-m`, whose AND/OR supports are not sorted).

`base/abci/mfs3_native_test.go` compares `mfs3` and `phase_map` with native
ABC (statistics, verbose output and the written BLIF).

`sfm_mapped_ntk.go` (`NewNetwork`, `Evaluate`) and the `Gate`/`Pin` types are
remnants of an earlier Go-native mapped optimizer; nothing outside tests uses
them now.

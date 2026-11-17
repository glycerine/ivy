# Bug fixes applied to this vendored z3 4.7.1

This tree is z3 4.7.1 (`dot.git.z3_v4.7.1_425291ee2`). The fixes below were
found while porting z3 to Go (the g3 project, github.com/glycerine/g3). That
work compares g3 against a z3 4.7.1 fork line by line. The same fixes are in
the g3 fork's `src/` tree:

* g3 commit c3c0016bd: relevancy and pareto backports
* g3 commit d91cef3ad: qsat/qe and theory_pb fixes
* g3 commit 0c41ffbae: bvsmul_noovfl/noudfl folding backport
* g3 commits deaf40e62, 048427fed, 74ab0e1a2, b75598ba8, ebcebc895, 62c5b6b34,
  aec02f8de: fixes 6-12 below, found by the g3 audit (g3 `pkg/AUDIT.md`);
  their red tests are in g3 commits a26556e2b, 1d959a9d5, 7ff877661 and
  5c8cf87b9
* g3 commits 6d012f4f8, fab4cf5cc, 420af90fe, 8c311df5b, 5d2088e13,
  3f918f05c, fd7eb9739, 2c5aaab9b, 0bcc96ef7, 60e06fba0, 242626470,
  243a6ce6f, 93d29c7e9, ced98bfe3, 2c41861cf: fixes 13-27 below, also found
  by the g3 audit; their red tests are in g3 commits 0c3b80131, 1a5c59778,
  76e1e9145, 831076c04, 48ba2c7c9, 71efc17ad, 20250af9c, 35dc5b753,
  da3e2adb5, 82e35924e, 574034129, b06bb5e1b, ce22629b9, dd423cba7 and
  42133d3c7, and the `FORK FIX #N` site comments are in g3 commit 365b01e5a
* g3 commits 223907cb8, 1473b32f1, f56da6929, 5126cea09, 3991679a6,
  a54282cae, 8259a67e3, 9936ca711, a7b4f8aa5, babfca1d3, 7f68cde3c,
  3f18d0a9d, 95bc91a7c, 013cb6dca (with 60ae897e7, its `update_api.py` half)
  and 4efcbe742: fixes 28-42 below, found by the g3 audit; their red tests are
  in g3 commits cd7fe0901, 93addf6c7, adeae7578, 587c74a58, 4207b714f,
  e0791ea54, b55867e91, 24c640b35, 69e363adc, 8b1260507, ac01db9c1,
  e1eec76ed, 25230dcf7, 6e1bb5c0a and 9f4457c59, and the `FORK FIX #N` site
  comments are in g3 commit 2f26a23d9
* g3 commits 018d8d387 and 1b4ae811e: fixes 43-44 below (red tests in g3
  commits 3f6706dec and c2b07b691); d3d15a7b4 repairs the `FORK FIX #40`
  comment, which had been placed inside a macro
* g3 commit eff65eb03: fix 45 below (red test in g3 commit 85f87a878, `FORK FIX
  #45` comment in 503d4fa8a)
* g3 commits 64f41b16d and c9d71d841: fixes 46-47 below (red tests in g3
  commits d21dc52c5 and 7f1d18f5d, `FORK FIX` comments in 70817364f)
* g3 commits 6d5697bfe, 088bffab7, aeeab9dba and 66c0fa4e4: fixes 48-51 below
  (red tests in g3 commits 800698807, 59e18dbb9, f337ddd39 and 0b5de3027,
  `FORK FIX` comments in 6bf3dd797)
* g3 commits 0e5916f46 and 060eea5c7: fixes 52-53 below (red tests in g3
  commits dd9706fb4 and 437084826, `FORK FIX` comments in 0edfff976)
* g3 commits 748e0e081 and 62affb150: fixes 54-55 below (red tests in g3
  commits 390424642 and 170a2fd3e, `FORK FIX` comments in 9778bc6d0)
* g3 commits ca3497492 and 93a286ac9: fixes 56-57 below (red tests in g3
  commits 574cf4f54 and e75aff46d, `FORK FIX` comments in ca7dae915)
* g3 commits 5ff7e54dd and e807b6c55: fixes 58-59 below (red tests in g3
  commits e23727c84 and fd1cf03ba, `FORK FIX` comments in the fix commits)
* g3 commits da4297199 and baec5580d: fixes 60-61 below (red tests in g3
  commits 7aca00862 and 42a8fc491, `FORK FIX` comments in 8aed59184 and
  baec5580d)
* g3 commit 8a785a252: fix 62 below (red test in g3 commit f5825ce1d,
  `FORK FIX` comment in the fix commit)
* g3 commit a771cbed6: fix 63 below (red test in g3 commit 542594f7e,
  `FORK FIX` comments in the fix commit)
* g3 commits 843a255af, 78f442aa6, 4372996c3, 7a5001629, 61cfc469a,
  a6e72d800, 392fc24a8, e0b4784b5, e0b592b0d, 3ab8fbf8b, c048c9a39,
  86591a87d, a16326820 and e743ee36f: fixes 64-77 below (red tests in g3
  commits f93d4ebeb, bef8e3063, 8e955ab03, 4deebbea4, 45a6078e6, 3e3334ab0,
  fb447d6a1, 61d3347be, 95c2a728e, 627df6fc1, 61cd5da3c, dcd04f514 and
  a138c4b5c; `FORK FIX` comments in the fix commits)
* g3 commits 2b6e876f6, fa6d9a3f6, e661b489c, 97ec8930e, 1c15f296d,
  8ed299553, 07e5e0e66, 9568f2110, c787bcd7c, 2636ac7d3, b1800ceec and
  69c959c58: fixes 79-90 below (red tests in g3 commits e7249ff89,
  59b15ad99, c59d036bc, bc8c8f2dd, 2990cc1df, be34a9c03, a103c88fb,
  d158e06c6, 5176da850, d9872fb66, b9203b17e and 6ffb6ec78; `FORK FIX`
  comments in the fix commits).
* g3 commits bb9a1f251, ddf0de888, aba28b511, f22e24f67 and 8035a892b:
  fixes 78 and 91-94 below (red tests in g3 commits 7348fa79a, c1f490483,
  ec4e0af1a, 601c9a3ca and 799016783; `FORK FIX` comments in the fix
  commits)
* g3 commit af3832cd0: fix 95 below (red test in g3 commit f1c5fe437,
  `FORK FIX` comment in 469bcbdd6)
* g3 commit 047e3d468: fix 96 below (red test in g3 commit df6c3cf06,
  `FORK FIX` comments in the fix commit)
* g3 commits 6044bbcd4 and 96159413f: fixes 97-98 below (red tests in g3
  commits 966187a42 and fa595827a, `FORK FIX` comments in the fix commits)
* g3 commits 1473e5d07 and e6a24e040: fixes 99-100 below (red tests in g3
  commits 5dc0e4c22 and 06764a57b, `FORK FIX` comments in the fix commits)
* g3 commit ff6f079e3: fix 101 below (red test in g3 commit e01903763,
  `FORK FIX` comments in the fix commit)
* g3 commits ca711590e and 124110e7a: fixes 102-103 below (red tests in g3
  commits 541401db9 and 98a5c9b8d, `FORK FIX` comments in the fix commits)
* g3 commits 543dfc268 and 607582d05: fixes 104-105 below (red tests in g3
  commits 97cb47b53 and b669b839f, `FORK FIX` comments in the fix commits)
* g3 commits 95a41b7f6, cdd4a72f5 and 0ba697b1f: fixes 106-108 below (red
  tests in g3 commits d36c8ce02, 10b5f2ebe and 61ceb4fae, `FORK FIX`
  comments in the fix commits)
* g3 commit 54601015d: fix 109 below (red test in g3 commit 930056541,
  `FORK FIX` comment in the fix commit)
* g3 commits d04ee6e17 and 1986d9747: fixes 110-111 below (red tests in g3
  commits 5abfed8ff and 27502ffed, `FORK FIX` comments in the fix commits)
* g3 commit 0b9b30eba: fix 112 below (red test in g3 commit 8ea2ebfe8,
  `FORK FIX` comment in the fix commit)
* g3 commit ca1f814c4: fix 113 below (red test in g3 commit 8d6be5d29,
  `FORK FIX` comment in the fix commit)
* g3 commit 34e04b88d: fix 114 below (red test in g3 commit d658343f7,
  `FORK FIX` comment in the fix commit)
* g3 commit e1402e49a: fixes 115-119 below (red tests in g3 commit 6d35e5439,
  `FORK FIX` comments in the fix commit)
* g3 commit 68656a40b: fix 120 below (red test and `FORK FIX` comments in
  the same commit)

They are copied here unchanged.

| # | Symptom in 4.7.1 | Fix | Files |
|---|---|---|---|
| 1 | Wrong models after `smt.relevancy` is raised | backport of upstream 5dfe4a4b4 | `smt/smt_context.{h,cpp}`, `smt/smt_internalizer.cpp`, `smt/smt_relevancy.cpp` (comment), `smt/params/smt_params.cpp` (comment) |
| 2 | Pareto optimization never terminates | backport of upstream e18702330 (#1699) | `opt/opt_context.cpp` |
| 3 | `qe_rec` segfaults on formulas with free constants and can return unsound results | backport of upstream 1ce0d7512 / 8ad939a10 (#2974, #2990) | `qe/qsat.cpp`, `qe/qe.cpp` |
| 4 | `free(): invalid pointer` abort in theory_pb with proofs on | local fix (upstream rewrote this code in 4.8.5) | `smt/theory_pb.cpp` |
| 5 | `bvsmul_noovfl`/`bvsmul_noudfl` fold constant operands to the wrong value | backport of upstream (z3 4.17.0) `mk_bvsmul_no_overflow` | `ast/rewriter/bv_rewriter.{h,cpp}` |
| 6 | `dom-simplify` turns a satisfiable goal into `false` | backport of upstream 0ef8ebe89f (#4895) | `tactic/core/dom_simplify_tactic.cpp` |
| 7 | `sat` on an unsat problem after `pop`: term-ite cache survives pop | backport of upstream 4c0db00a7b | `smt/elim_term_ite.h` |
| 8 | `sat` on an unsat problem after `smt.macro_finder` is turned off | local fix (still present upstream) | `smt/asserted_formulas.cpp` |
| 9 | `pb-preprocess` answers `sat` on unsat goals, segfaults | backport of upstream da8530e2db, 939860148f, 42f206171d, 302c0d178c | `tactic/core/pb_preprocess_tactic.cpp` |
| 10 | `unsat` on a satisfiable datatype problem with `smt.relevancy=1` | backport of upstream 7fc4653e47 (#4623) | `smt/theory_datatype.cpp` |
| 11 | `dom-simplify` never terminates on some goals | backport of upstream 0ef8ebe89f (#4895) round cap | `tactic/core/dom_simplify_tactic.cpp` |
| 12 | `update-field` over a constructor leaves the field unchanged (wrong `sat`) | backport of upstream f989e4eb38 (#2276) | `ast/datatype_decl_plugin.{h,cpp}`, `ast/rewriter/datatype_rewriter.cpp` |
| 13 | PDR answers `sat` on an unsat problem with nullary predicates, then segfaults | local fix (PDR was removed upstream) | `muz/pdr/pdr_context.{h,cpp}` |
| 14 | Duality fails with "Wrong number of arguments" and answers `unknown` | local fix (Duality was removed upstream) | `duality/duality_rpfp.cpp` |
| 15 | `lazy_sparse` tables segfault on a nullary union | local fix (still present upstream) | `muz/rel/dl_lazy_table.cpp` |
| 16 | QBMC aborts on a VERIFY failure | partial backport of upstream 784b613b5 (#11540) | `muz/bmc/dl_bmc_engine.cpp` |
| 17 | DDNF segfaults on `extract` equalities | local fix (still present upstream) | `muz/ddnf/ddnf.cpp` |
| 18 | `min_cut` never terminates (spacer `minimize_unsat_core`) | backport of upstream 8150cd6e8 (#11401) | `util/min_cut.cpp` |
| 19 | Quantifier-instantiation transform segfaults (no proof mode, swapped premises) | local fix (still present upstream) | `muz/transforms/dl_mk_quantifier_instantiation.cpp` |
| 20 | `pb2bv` with a numeral on the left: `sat` on an unsat problem | local fix (still present upstream) | `tactic/arith/pb2bv_tactic.cpp` |
| 21 | `eq2bv` keeps its rewriter cache across goals: `sat` on an unsat problem | local fix (still present upstream) | `tactic/arith/eq2bv_tactic.cpp` |
| 22 | `purify-arith` decides unsafe trig arguments too late: `sat` with an invalid model | local fix (still present upstream) | `tactic/arith/purify_arith_tactic.cpp` |
| 23 | UFBV demodulator leaves a variable unbound: wrong `sat` | backport of upstream b76ed6a47 (#6476) | `tactic/ufbv/ufbv_rewriter.cpp` |
| 24 | `recover-01` reads past the end of the goal (segfault) | backport of upstream c230d89a3 (#5294) | `tactic/arith/recover_01_tactic.cpp` |
| 25 | `distribute-forall` segfaults with proofs on | backport of upstream d530d1314 (#3276) | `tactic/core/distribute_forall_tactic.cpp` |
| 26 | `pb2bv` builds a malformed circuit for fractional 0/1 sums | backport of upstream 10768bd00 (#3727), f7a7b9e1f (#4108) | `ast/rewriter/pb2bv_rewriter.cpp` |
| 27 | `lia2card` never terminates on an equation with no terms | local fix (still present upstream with `compile_equality=false`) | `tactic/arith/lia2card_tactic.cpp` |
| 28 | nlqsat never terminates: a negative sign is stored as unsigned in infinity projection | backport of upstream 363d7aad2 | `nlsat/nlsat_explain.cpp` |
| 29 | nlsat evaluates an atom's sign wrongly when the factor-sign product overflows 32 bits | backport of upstream f79dccccf, 6071797ba (#2238) | `nlsat/nlsat_evaluator.cpp` |
| 30 | nlsat `solve_eq` uses a stale power table across polynomials | backport of upstream 19e2f8c9d (#1694) | `nlsat/nlsat_explain.cpp` |
| 31 | nlsat builds an invalid explain lemma (`s == -1` instead of `s < 0`) | backport of upstream 0f20175bd (#2556) | `nlsat/nlsat_explain.cpp` |
| 32 | nlsat `inline_vars`: segfault on a vanished factor, invalid model on a constant atom | backport of upstream 29f3f6a7a (#3144), 99784a92e (#3225) | `nlsat/nlsat_solver.cpp` |
| 33 | nlsat `inline_vars` then reorder gives an invalid model | backport of upstream 983a55232 (#3167) | `nlsat/nlsat_solver.cpp` |
| 34 | `qe-sat` never terminates: `project_real` picks the bound family from raw counts | local fix (qe-sat was removed upstream in 8717c7d7f) | `qe/qe_arith_plugin.cpp` |
| 35 | `qe` with `mod` by 0 or a negative constant returns an unsound `false` | local fix (still present upstream) | `qe/qe_arith_plugin.cpp` |
| 36 | `qe` datatype substitution drops the recognizer (unsound) | local fix (still present upstream) | `qe/qe_datatype_plugin.cpp` |
| 37 | qe-light Fourier-Motzkin uses a stale Boolean-variable map (wrong result) | local fix (still present upstream) | `qe/qe_lite.cpp` |
| 38 | maxres `add_upper_bound_block` stops at a non-optimal cost (`pb_util::mk_lt` aliasing) | local fix (upstream rewrote `mk_lt`) | `ast/pb_decl_plugin.cpp` |
| 39 | OPB `max:` objectives are always rejected | local fix (still present upstream) | `opt/opt_parse.cpp` |
| 40 | probe `+` and `*` use the first argument twice | local fix (still present upstream) | `cmd_context/tactic_cmds.cpp` |
| 41 | z3 cannot replay its own `Z3_mk_pble`/`pbge`/`pbeq` logs | backport of upstream fed977b49 (#1782) | `scripts/update_api.py`, `api/z3_logger.h`, `api/z3_replayer.cpp` |
| 42 | `(get-user-tactics)` segfaults | backport of upstream 202ed79a2 (#5445) | `util/sexpr.cpp` |
| 43 | Duality with `xform.quantify_arrays`: later queries answer wrongly (lost or stale rules) | local fix (Duality was removed upstream) | `muz/duality/duality_dl_interface.cpp` |
| 44 | Duality segfaults on the query after a failed one | local fix (Duality was removed upstream) | `muz/duality/duality_dl_interface.cpp` |
| 45 | `model_implicant` evaluates a nested array disequality to false (UNREACHABLE, segfault) | local fix (still present upstream) | `model/model_implicant.cpp` |
| 46 | Duality with `use_underapprox` aborts with "error in ExpansionChoices" on a sat problem | local fix (Duality was removed upstream) | `duality/duality_rpfp.cpp` |
| 47 | Duality with `use_underapprox` fails in iZ3 ("unknown assertion") on a sat problem | local fix (Duality was removed upstream) | `duality/duality_solver.cpp` |
| 48 | PDR proof trace proves the wrong fact when nullary predicates share the state `true` | local fix (PDR was removed upstream) | `muz/pdr/pdr_context.cpp` |
| 49 | PDR proof trace concludes from an unresolved duplicate leaf | local fix (PDR was removed upstream) | `muz/pdr/pdr_context.cpp` |
| 50 | PDR Boogie certificate without `generate_proof_trace` segfaults | local fix (PDR was removed upstream) | `muz/pdr/pdr_context.cpp` |
| 51 | PDR proof trace segfaults after `remove_duplicate_tails` merged a tail | local fix (PDR was removed upstream) | `muz/pdr/pdr_context.cpp` |
| 52 | Model evaluation after a `max_steps` exception returns a stale result (sort mismatch, segfault) | backport of upstream 18bb90f93 (#4426), b013df9a9 (#4431) | `ast/rewriter/rewriter_def.h` |
| 53 | `qe` on finite-domain sorts: `exists x. x != c` eliminates to `false` (unsound) | local fix (still present upstream) | `qe/qe_dl_plugin.cpp` |
| 54 | Sieve-relation negation drops inner/inner equalities and empties the target | local fix (still present upstream) | `muz/rel/dl_sieve_relation.cpp` |
| 55 | `datalog.default_table_checked` with a sparse table segfaults on a nullary empty table | local fix (still present upstream) | `muz/rel/dl_check_table.cpp` |
| 56 | `datalog.default_table=bitvector` segfaults on a nullary table | backport of upstream 8dac9b7b94 | `muz/rel/dl_table.cpp` |
| 57 | `smt.case_split=6` reads an uninitialized decision phase (always negative in the reference build) | as upstream (`lbool phase = l_undef`) | `smt/smt_context.cpp` |
| 58 | `pb.enable_simplex`: a pb constraint over a nested pb atom answers `unsat` on satisfiable input, or segfaults | local fix (`pb.enable_simplex` was removed upstream) | `smt/theory_pb.cpp` |
| 59 | `pb.enable_simplex` under push/pop: a reused bool var's slack column gets a 0-1 bound, giving `unsat` on satisfiable input | local fix (`pb.enable_simplex` was removed upstream) | `smt/theory_pb.cpp` |
| 60 | `smt_solver::updt_params` resets solver-level smt settings (e.g. `mbqi=false` turns back on after `Z3_get_implied_equalities`; segfault in the model checker) | as upstream (merged `solver::get_params()`) | `smt/smt_solver.cpp` |
| 61 | A pb equality whose arguments are all assigned at internalization is dropped: `sat` on unsat input (default options) | local fix (upstream redesigned pb equalities in bd96eaff47) | `smt/theory_pb.cpp` |
| 62 | `theory_pb::assert_ge` (test helper) segfaults: its local theory is never initialized (`./test-z3 theory_pb`) | as upstream becf423c77 | `smt/theory_pb.cpp` |
| 63 | Duality certificates give eliminated rule variables the model default (invalid certificate) | local fix (Duality was removed upstream); port of g3's witness search | `muz/duality/duality_dl_interface.cpp` |
| 64 | `qe` finite-domain plugin: wrong result when the domain size equals the number of equality atoms | local fix (still present upstream) | `qe/qe_dl_plugin.cpp` |
| 65 | `((_ to_fp eb sb) <Real>)`: out-of-bounds read of `domain[1]` | as upstream (one-Real branch removed) | `ast/fpa_decl_plugin.cpp` |
| 66 | `z3 -dl <directory>` prints NOT IMPLEMENTED YET! and hangs on POSIX | local fix (still present upstream) | `muz/base/dl_util.cpp` |
| 67 | WPA not-in-sort warning prints with MSVC's `%I64u` (undefined behavior with glibc) | local fix (still present upstream) | `muz/fp/datalog_parser.cpp` |
| 68 | `read_uint64` rejects values near `ULLONG_MAX` | local fix (still present upstream) | `muz/base/dl_util.cpp` |
| 69 | WPA map line with an empty name reads the name's last character | local fix (still present upstream) | `muz/fp/datalog_parser.cpp` |
| 70 | `goal::update` pushes the unflattened formula, so a goal can fail to become inconsistent | local fix (still present upstream) | `tactic/goal.cpp` |
| 71 | Model evaluation of mutually recursive `define-fun-rec` predicates never terminates | local fix (upstream replaced it with recfun) | `model/model_evaluator.cpp`, `ast/rewriter/rewriter_def.h`, `ast/ast.h`, `cmd_context/cmd_context.cpp` |
| 72 | Nested parametric datatypes: use-after-free in `get_sort_size` (segfault) | backport of upstream f1c3e1aa7, e08abb321 | `ast/datatype_decl_plugin.{h,cpp}` |
| 73 | `and_then`: a decided-sat subgoal does not decide the goal (`unknown`) | as upstream (`result.reset()`) | `tactic/tactical.cpp` |
| 74 | `repeat`: same as 73, and a null-model segfault with models on | local fix (still present upstream) | `tactic/tactical.cpp` |
| 75 | opt `from_fmls`: an assert-soft group after maximize/minimize is checked against the wrong term | backport of upstream b33f5f879 | `opt/opt_context.cpp` |
| 76 | qe array `solve_store`: wrong result when nested store indices coincide | local fix (still present upstream) | `qe/qe_array_plugin.cpp` |
| 77 | Symbolic automaton product ignores the epsilon closure of the initial states | backport of upstream 606754c09, a921b4ff4 | `math/automata/symbolic_automata_def.h` |
| 78 | datalog `string_to_uint64` accepts a blank string and reads an uninitialized value | local fix (still present upstream) | `muz/base/dl_util.cpp` |
| 79 | `elim-small-bv` enumerates only `bv_sz` values (unsound `sat`), crashes on non-BV quantifiers | backport of upstream b6ee0b151, 8e033c1e7, aa901c4e8 | `tactic/bv/elim_small_bv_tactic.cpp` |
| 80 | `sine-filter` segfaults on a symbol-free last formula or an empty goal | backport of upstream fcd1f2b3c, plus a local empty-goal guard | `tactic/sine_filter.cpp` |
| 81 | `symmetry-reduce` fails `VERIFY` / segfaults on quantified goals | backport of upstream 330b3cc8d | `tactic/tactic.{h,cpp}`, `tactic/core/symmetry_reduce_tactic.cpp` |
| 82 | `bvarray2uf` hits `NOT_IMPLEMENTED_YET` on quantifiers | backport of upstream 4bad2dd92 | `tactic/bv/bvarray2uf_rewriter.cpp` |
| 83 | `get-value` under `model_evaluator.max_steps` segfaults | backport of upstream 335d672bf | `parsers/smt2/smt2parser.cpp` |
| 84 | A pop forgets parametric datatype instances created outside the popped scope | backport of upstream 19f655c69, 120467159, 2d358a9a5, 984767509 | `cmd_context/pdecl.{h,cpp}`, `cmd_context/cmd_context.cpp` |
| 85 | Datatype constructors unknown after pop when reached through a `define-sort` alias | local fix (still present upstream) | `cmd_context/pdecl.{h,cpp}` |
| 86 | WPA loader `fread`s a null `FILE` for an unreadable `.map`/`.rel` file | local fix (still present upstream) | `muz/fp/datalog_parser.cpp` |
| 87 | QBMC hyper-resolution positions recorded as (0,1) | backport of upstream 784b613b5 | `muz/bmc/dl_bmc_engine.cpp` |
| 88 | Nonlinear BMC proof reconstruction Skolemizes head variables (segfault) | local fix (still present upstream as an error) | `muz/bmc/dl_bmc_engine.cpp` |
| 89 | Boogie certificate prints leftover debug lines on stdout | as upstream 73a24ca0a | `muz/base/dl_boogie_proof.cpp` |
| 90 | `psort_nw::m_t` uninitialized: sortmax's network depends on garbage (wrong optima) | as upstream (`m_t(GE)`) | `util/sorting_network.h` |
| 91 | Duality certificates print and prove each edge with the previous rule's clause after dead-edge removal | local fix (Duality removed upstream) | `muz/duality/duality_dl_interface.cpp` |
| 92 | Duality `fixedpoint.print_answer` segfaults (`get_answer` returns null) | local fix (Duality removed upstream) | `muz/duality/duality_dl_interface.cpp` |
| 93 | Duality `get_proof` builds hyper-resolve steps the proof checker rejects | local fix (Duality removed upstream); port of g3's proof shape | `muz/duality/duality_dl_interface.cpp` |
| 94 | Linear BMC hyper-resolution positions recorded as (0,1) | local fix (still present upstream) | `muz/bmc/dl_bmc_engine.cpp` |
| 95 | qe-sat: a resource-limited projection check fails a `VERIFY` and segfaults | local fix (qe-sat was removed upstream) | `qe/qe_sat_tactic.cpp` |
| 96 | Duality livelocks re-proposing a candidate the feasibility check already rejected | local fix (Duality removed upstream) | `duality/duality_solver.cpp` |
| 97 | Simple-joins result rule order depends on heap addresses (pointer-keyed map) | local fix (still present upstream); id/source order as g3 | `muz/rel/dl_mk_simple_joins.cpp` |
| 98 | theory_str iterates pointer-keyed maps and sets in address order | local fix (theory_str removed upstream); AST-id order as g3 | `smt/theory_str.{h,cpp}` |
| 99 | Duality answers an unsound `sat` when an edge solver with `mbqi.id=itp` ignores a quantifier | local fix (Duality removed upstream): validate the counterexample | `muz/duality/duality_dl_interface.cpp` |
| 100 | iZ3 refuses rewrite and and-elim steps on formulas propagate_values mixed across frames (Duality "Unsupported proof rule") | local fix (iZ3 removed upstream) | `interp/iz3translate.cpp` |
| 101 | iZ3 still refuses a mixed rewrite whose congruence side is already mixed (second propagate_values round) | local fix (iZ3 removed upstream) | `interp/iz3translate.cpp` |
| 102 | Datalog product-relation kinds accepted for any signature, then segfault reading a missing spec | local fix | `muz/rel/dl_product_relation.cpp`, `muz/rel/dl_relation_manager.h` |
| 103 | Sieve plugin refuses its own kinds, products lose their sieves, aligned union calls a null functor | local fix | `muz/rel/dl_sieve_relation.cpp`, `muz/rel/dl_product_relation.cpp` |
| 104 | z3str3 reads `str.in.re` terms popped out of the context (segfault in `get_assignment`, or a final-check loop) | narrower than upstream 077f2248c / 4fc64ab57 / c1a0ce086 (theory_str later removed) | `smt/theory_str.cpp` |
| 105 | z3str3 legacy regex unroll dereferences a missing `unroll_tries_map` entry | local fix (legacy unroll solver later removed upstream) | `smt/theory_str.cpp` |
| 106 | check_relation operations throw `std::bad_cast` when an argument is another plugin's relation | local fix | `muz/rel/check_relation.cpp` |
| 107 | table_relation_plugin asks its table plugin for a signature it cannot handle (sparse_table segfault) | local fix | `muz/rel/dl_table_relation.cpp` |
| 108 | `table_base::complement` stops on `VERIFY(sig.first_functional() == 1)` for several non-functional columns | local fix | `muz/rel/dl_base.cpp` |
| 109 | `inline_linear` inlines away predicates that hold relation facts when `xform.inline_eager=false` | local fix | `muz/transforms/dl_mk_rule_inliner.cpp` |
| 110 | `finite_product_relation` `filter_interpreted_fn` calls a null inner filter functor (SIGSEGV) | local fix | `muz/rel/dl_finite_product_relation.cpp` |
| 111 | `mk_explanations` with `default_relation doc` calls a null union functor in `transform_facts` (SIGSEGV) | local fix | `muz/rel/dl_mk_explanations.cpp` |
| 112 | `rule_transformer` leaks the rule set it is building when a plugin throws (e.g. `xform.karr`) | local fix | `muz/base/dl_rule_transformer.cpp` |
| 113 | `tbv_manager::set(tbv&, uint64, hi, lo)` shifts `1ULL` by 64 or more for columns wider than 64 bits (UB; wrong udoc facts) | local fix | `muz/rel/tbv.cpp` |
| 114 | `mpz_manager::gcd` reads the cell of an inline operand when the other equals INT_MAX stored as a cell (segfault) | local fix | `util/mpz.cpp` |
| 115 | `sparse_table` `negated_join_fn` removes rows in increasing offset order, so `remove_offset` drops the wrong rows | local fix | `muz/rel/dl_sparse_table.cpp` |
| 116 | `sparse_table::contains_fact` compares the wrong cells when the table has functional columns | local fix | `muz/rel/dl_sparse_table.cpp` |
| 117 | `sparse_table` `negated_join_fn` uses src2's join column numbers as src2's own columns | local fix | `muz/rel/dl_sparse_table.cpp` |
| 118 | `relation_manager::mk_filter_by_intersection_fn(table)` asks for a negation filter, so intersection removes the shared rows | local fix | `muz/rel/dl_relation_manager.cpp` |
| 119 | `sparse_table` `negated_join_fn` never looks up an all-zero first key | local fix | `muz/rel/dl_sparse_table.cpp` |
| 120 | `xform.bit_blast` drops the facts in the blasted predicates' relations, so a satisfiable query answers unsat | local fix | `muz/transforms/dl_mk_bit_blast.cpp`, `muz/rel/rel_context.cpp`, `muz/rel/rel_context.h`, `muz/base/dl_context.h` |

Every changed site has a comment that starts with "backport of upstream",
"BACKPORT upstream", or "FORK FIX".

## Contents

* `fork_bugfixes/fork_bugfixes.patch`: all source changes as one unified
  diff against the original 4.7.1 sources, applied with `patch -p1` from this
  directory. It includes the test registration in `src/test/main.cpp` and
  `src/test/CMakeLists.txt`. It does not include the new file
  `src/test/fork_bugfixes.cpp`. Keep the patch for re-vendoring.
* `src/test/fork_bugfixes.cpp`: the C++ unit test, `TST(fork_bugfixes)`.
* `fork_bugfixes/test_vectors/*.smt2` with `.expected` outputs, plus
  `fork_bugfixes/run_tests.sh`, which checks a z3 binary against them.

## Building and testing

`compile.sh` regenerates the Makefile with `scripts/mk_make.py`, so the new
test file is picked up automatically. To run the tests:

```
./compile.sh                                  # produces native.build/
cd native.build && make test-z3 && ./test-z3 fork_bugfixes
cd .. && fork_bugfixes/run_tests.sh           # uses native.build/z3 by default
```

The binaries under `native.build/` and `../native_lib/` were built before
these fixes. `run_tests.sh` fails against them until `compile.sh` rebuilds
them.

Watch for a nested build directory. `compile.sh` ends with
`mv build native.build`. If `native.build` already exists, `mv` puts the new
build inside it as `native.build/build`, and `native.build/` keeps the old
build and its old generated Makefile. Linking `test-z3` there then fails
with "undefined reference to `tst_fork_bugfixes()`", because that Makefile
predates `src/test/fork_bugfixes.cpp`. Remove the old `native.build` before
the `mv` (or run `rm -rf native.build` before `compile.sh`). The files
`compile.sh` copies into `../native_lib` are correct either way.

Expected `test-z3` output:

```
fork_bugfixes: relevancy_update_scopes ok
fork_bugfixes: pareto_model_completion ok
fork_bugfixes: qe_rec_free_boolean ok
fork_bugfixes: qe_rec_real_disjunction ok
fork_bugfixes: pb_strengthening_proof ok
fork_bugfixes: bv_smul_noovfl_fold ok
PASS
```

### Verification

* **Patched build.** I built the patched sources with the existing
  `native.build` configuration. Both `test-z3 fork_bugfixes` and
  `run_tests.sh` pass.
* **Controls.** I reverted each fix in turn, rebuilt, and confirmed that its
  unit test fails:

  | Fix reverted | Result |
  |---|---|
  | pareto | hangs (killed by the 60 s timeout) |
  | qsat | segfaults |
  | theory_pb | `free(): invalid pointer` |
  | relevancy | `relevancy_update_scopes FAILED` |

* **Unpatched binary.** All five `run_tests.sh` vectors for fixes 1-4 fail
  on an unpatched 4.7.1 binary.
* **Fix 5.** I built a scratch copy of this tree from the existing
  `native.build`. `test-z3 fork_bugfixes` and `run_tests.sh` both pass
  (6 of 6). The previous `native.build/z3` fails `bv_smul_noovfl_fold`:
  it prints `true` for `#b1 * #b1` and `false` for `#xa * #x3`.

---

## 1. Relevancy raised after setup gives models that violate assertions

**Upstream:** 5dfe4a4b4. The smt context has its own relevancy level, which
can be lowered but never raised after setup.

**Symptom.** The commands below make z3 4.7.1 answer `sat` with `p false`,
although `p` is asserted:

1. `(set-option :smt.relevancy 0)`;
2. assert `p`;
3. `check-sat`;
4. `(set-option :smt.relevancy 2)`;
5. `check-sat` and `get-value`.

The vector is `test_vectors/relevancy_update_scopes.smt2`.

**Cause.**
* `smt_solver::updt_params` copies the new `smt.relevancy` into the shared
  `smt_params`.
* `context::relevancy()` read `m_fparams.m_relevancy_lvl` directly.
* Assertions internalized while relevancy was 0 were never marked relevant
  in `relevancy_propagator_imp::mark_as_relevant`, and nothing re-marks them
  later.
* Model generation skips irrelevant atoms, so the model can violate those
  assertions.

**Fix.**
* `context` gets an `unsigned m_relevancy_lvl`, initialized to 2.
* `setup_context` sets it from `m_fparams` the first time. After that it
  only lowers it: `std::min(m_fparams.m_relevancy_lvl, m_relevancy_lvl)`.
* `relevancy()`, the new `relevancy_lvl()`, the atom-propagation checks in
  `smt_context.cpp`, and the lemma check in
  `context::use_binary_clause_opt` (`smt_internalizer.cpp`) use the context
  level.
* The case-split-queue warning keeps reading the parameter, because the
  queue is created in the constructor.
* `smt_relevancy.cpp` and `smt_params.cpp` only get comments.

**Behaviour change.** Scripts that raise `smt.relevancy` mid-session keep
the lower level. In the g3 vectors, stdout changed only for the bug's own
reproduction.

## 2. Pareto optimization loops forever

**Upstream:** e18702330, fix for issue #1699, adapted to the 4.7.1 API.

**Symptom.** With `:opt.priority pareto` and two `maximize` objectives over
`x, y` in `[0,1]` with `x + y = 1`, the first `check-sat` never returns. The
vector is `test_vectors/pareto_model_completion.smt2`. It hangs with or
without its `opt.optsmt_engine basic` option. The optimizer's preprocessing
turns these bounded integers into 0-1 variables.

**Cause.**
* `opt::context::mk_cmp` evaluates each objective in the current model
  without model completion.
* When the model leaves a preprocessed 0-1 variable undefined, the
  objective stays symbolic and `is_numeral` fails. The dominance bound is
  then trivially true.
* `gia_pareto` finds the same model again, forever.

**Fix.** In `opt_context.cpp`, `mk_cmp` calls
`mdl->eval(term, val, true)` (model completion) in both the
`O_MAXIMIZE`/`O_MINIMIZE` and `O_MAXSMT` cases.

## 3. `qe_rec` crash and unsound result (qsat `elim`)

**Upstream:** 1ce0d7512 and 8ad939a10, fixes for issues #2974 and #2990.

**Symptoms.**
* **Crash.** This script segfaults in `pred_abs::abstract_atoms`
  (`m_elevel.find`). The vector is `test_vectors/qe_rec_free_boolean.smt2`.

  ```
  (declare-const p Bool)
  (assert (exists ((x Bool)) (or x p)))
  (assert (not p))
  (check-sat-using (then qe_rec smt))
  ```

  The free constant `p` was never given a quantifier level.
* **Unsound result.** `(apply qe_rec)` on
  `exists x. a < x and 2x < b and (x > 5 or x + b < 7)` returns a goal
  that is a conjunction containing `false`. The formula is satisfiable for
  suitable `a`, `b`. The vector is `test_vectors/qe_rec_real_disjunction.smt2`.

**Fix.** In `qsat.cpp`, `qsat::elim` (used by `elim_rec`) takes the same
path as qe mode:
1. It returns a null `expr_ref` if the body still has quantifiers.
2. It calls `reset()`, rebuilds `exists vars. fml`, then calls `push_not`
   and `hoist`. `hoist` assigns levels to every free constant.
3. It throws `tactic_exception("formula is not hoistable")` if the result
   is not ground.
4. On `l_false` it returns `mk_and(m_answer)`, existentially closed over
   only those projected free variables that still `occur` in the answer.
5. For any other result it returns null.

Related changes:
* In `elim_rec`, a null result keeps the original subformula.
* The `qsat_qe_rec` special case `num_scopes = 2` is removed.
* `#include "ast/occurs.h"` is added.
* In `qe.cpp`, an assertion is relaxed from `l_true == check()` to
  `l_false != check()`, as upstream did.

**Behaviour change.** On the disjunction vector, `qe_rec` now leaves the
existential unchanged instead of returning an unsound `false`.

## 4. theory_pb: region-allocated justification freed with `dealloc`

**Upstream:** none. Upstream removed this code when it rewrote theory_pb,
first in 4.8.5. This is a local one-site fix.

**Symptom.** With proofs on, `smt.pb.conflict_frequency 0` and
`smt.pb.enable_compilation false`, two conflicting PB constraints make z3
print the answer, core and proof, then abort with `free(): invalid pointer`.
The vector is `test_vectors/pb_strengthening_proof.smt2`.

**Cause.**
* `theory_pb::resolve_conflict(ineq&)`, in its `l_false` case, adds a
  `CLS_AUX_LEMMA` clause justified by `justify(m_ineq_literals)`.
* That justification is allocated in the context region.
* A lemma owns its justification and frees it on the heap
  (`clause::deallocate`), which aborts.

**Fix.** When proofs are enabled, allocate the justification on the heap
with `alloc(theory_lemma_justification, get_id(), ctx, n, lits)`, as the
other lemma sites in the file already do. When proofs are off, pass
`nullptr`.

---

## 5. `bvsmul_noovfl` / `bvsmul_noudfl` constant folding

**Upstream:** the current `bv_rewriter::mk_bvsmul_no_overflow` (checked
against z3 4.17.0). It takes an `is_overflow` flag and replaces
`mk_bvsmul_no_underflow`. The new code is upstream's, unchanged.

**Symptom.** `(simplify (bvsmul_noovfl #xa #x3))` gives `false`. Signed,
that is -6 * 3 = -18. That is an underflow at width 4, not an overflow, so
z3 4.17.0 gives `true`. At width 1, `(bvsmul_noovfl #b1 #b1)` and
`(bvsmul_noudfl #b1 #b1)` both give `true`, although `#b1` is -1 and
(-1) * (-1) = 1 overflows. The vector is
`test_vectors/bv_smul_noovfl_fold.smt2`, with 12 cases.

**Cause.**
* `mk_bvsmul_no_overflow` multiplies the numerals' unsigned values and
  compares the product with 2^(w-1). This ignores the signs.
* Both functions fold "an operand is 1" to `true`. At width 1, bit pattern
  1 is -1, so that shortcut is wrong there.

**Fix.**
* Replace the overflow function with upstream's version, and route
  `OP_BSMUL_NO_UDFL` to it with `is_overflow == false`.
  * Each operand becomes a magnitude plus a sign flag.
  * Overflow holds when the signs are equal and the product is at least
    2^(w-1). Underflow holds when the signs differ and the product is more
    than 2^(w-1).
* The "operand is 1" shortcut is skipped when `bv_sz == 1`.
* Delete `mk_bvsmul_no_underflow` and its TRACE.

---

## 6. dom-simplify re-simplifies the dominator-tree root

**Upstream:** 0ef8ebe89f (2020-12-14, "fix #4895"), which added the same
guard to `simplify_rec` (now `src/ast/simplifiers/dominator_simplifier.cpp`).

**Symptom.** `(assert (= true (ite p1 p0 p1)))` is satisfied by p0 = p1 =
true, but `(check-sat-using (then dom-simplify smt))` answers `unsat` and
`(apply dom-simplify)` prints the goal `(= true false)`.

**Cause.** `dom_simplify_tactic::simplify_rec` recurses into every child of
`e` in the dominator tree. The root is its own child in `tree(e)`, so the
root is simplified again under the assumption made from its own value.

**Fix.** `src/tactic/core/dom_simplify_tactic.cpp:290`: recurse only
`if (child != e)` (in `simplify_rec`, `src/tactic/core/dom_simplify_tactic.cpp:263`).

**Test.** `TST(z3_4_7_1_dom_simplify_root)`, `src/test/z3_4_7_1_bugs.cpp:251`
(registered at `src/test/main.cpp:212`).

---

## 7. elim_term_ite cache survives pop

**Upstream:** 4c0db00a7b (2019-08-30, "fix push/pop bug for ite-elimination").

**Symptom.** In `(set-logic QF_LRA)`, after `(push)`, an assertion
containing `(ite c a b)`, `(check-sat)` and `(pop)`, a second scope that
asserts `a = 1`, `b = 1`, `x = 0` and `x + 2*(ite c a b) < 1` is reported
`sat`; the correct answer is `unsat`.

**Cause.** `asserted_formulas` keeps one `elim_term_ite` rewriter for its
whole life. `elim_term_ite_rw` declared its own `reset()` that only reset
the config, shadowing `rewriter_tpl::reset()`, so the rewriter cache was
never cleared. After `pop` the term-ite was replaced by a name whose
definition had been popped.

**Fix.** `src/smt/elim_term_ite.h:48`: the shadowing `reset()` is removed
(a comment marks the place), so `rewriter_tpl::reset()` clears both the new
definitions and the cache after each pass.

**Test.** `TST(z3_4_7_1_elim_term_ite_pop)`, `src/test/z3_4_7_1_bugs.cpp:265`
(registered at `src/test/main.cpp:213`).

---

## 8. Stored macros not expanded after `smt.macro_finder` is turned off

**Upstream:** none. The bug is still present in upstream z3 (checked at
1a1d88dde7, 2026-11-13, v5.4.2), which answers `sat sat` on the
reproduction below. A draft upstream bug report is kept with the g3
project.

**Symptom.**
```
(set-option :smt.macro_finder true)
(declare-fun f (Int) Int)
(push)
(assert (forall ((x Int)) (= (f x) (+ x 1))))
(check-sat)
(set-option :smt.macro_finder false)
(assert (= (f 1) 5))
(check-sat)
```
4.7.1 answers `sat sat`; the second answer must be `unsat`.

**Cause.** The quantifier that became a macro was removed from the asserted
formulas. `asserted_formulas::reduce` applies stored macros only by running
find-macros, which no longer runs once the option is off, so the new
assertion is never checked against the macro. `expand_macros()` was declared
in the header but never defined.

**Fix.** `src/smt/asserted_formulas.cpp:233-241`: when macros exist, run
find-macros if it applies, otherwise call `expand_macros()`; the new
definition of `asserted_formulas::expand_macros()` is at
`src/smt/asserted_formulas.cpp:344`. With the finder on, the path is
unchanged.

**Test.** `TST(z3_4_7_1_macro_finder_off)`, `src/test/z3_4_7_1_bugs.cpp:298`
(registered at `src/test/main.cpp:214`).

---

## 9. pb-preprocess: wrong goal indices, unsound resolve, crash in replace

**Upstream:** da8530e2db (2021-08-02, #5447), 939860148f (2021-08-04, #5452,
re-applied in 391db898d3), 42f206171d (2021-12-27, #5741) and 302c0d178c
(2022-02-26, #5867).

**Symptom.** `(then pb-preprocess smt)` answers `sat` on unsatisfiable goals,
for example `p1`, `(> x 0)`, `p0`, `(not p1)`, or `(or (not p0) (not p1))`,
`p1`, `p0`; on some goals z3 segfaults. Over 2000 random scripts 4.7.1 gave
18 wrong answers and 5 segfaults; with this fix, none.

**Cause.**
* `classify_vars(i, ...)` recorded the position in `m_ge` where later code
  reads a goal index.
* The subsumption pass appended to `args2`/`coeffs2` across candidates.
* `resolve` never set `min_index` for the first complement, and deleted a
  unit clause even when no complement was found.
* `replace` kept updating the goal after it had become inconsistent.

**Fix.** All in `src/tactic/core/pb_preprocess_tactic.cpp`:
* `:183` `classify_vars(m_ge[i], ...)`;
* `:252-253` reset `args2`/`coeffs2` before each subsumption candidate;
* `:542-555` in `resolve`: set `min_index` when `i == 0` and return unless a
  complement was found;
* `:643` stop the `replace` loop once `g->inconsistent()`.

**Tests.** `TST(z3_4_7_1_pb_preprocess_index)`,
`src/test/z3_4_7_1_bugs.cpp:333`, and `TST(z3_4_7_1_pb_preprocess_resolve)`,
`src/test/z3_4_7_1_bugs.cpp:349` (registered at `src/test/main.cpp:215-216`).

---

## 10. Datatype recognizer assigned false before it is relevant

**Upstream:** 7fc4653e47 (2020-08-11, "fix #4623").

**Symptom.** With `smt.relevancy=1`, a problem where a recognizer such as
`(is-cons l)` is assigned false before any recognizer of that class is
relevant is reported `unsat`; the correct answer is `sat`.

**Cause.** `theory_datatype::propagate_recognizer` looked the recognizer up
in the class's recognizer table, which was still empty, and treated the
missing entry as a conflict.

**Fix.** `src/smt/theory_datatype.cpp:750-754` (in `propagate_recognizer`,
`src/smt/theory_datatype.cpp:740`): if the class has no recognizers yet,
register the triggering one with `add_recognizer(w, recognizer)` first.

**Test.** `TST(z3_4_7_1_dt_recognizer)`, `src/test/z3_4_7_1_bugs.cpp:388`
(registered at `src/test/main.cpp:217`).

**Note on fixes 6-12.** Their tests live in `src/test/z3_4_7_1_bugs.cpp`
(listed in `src/test/CMakeLists.txt:119`), not in
`src/test/fork_bugfixes.cpp`, and their source changes are not in
`fork_bugfixes/fork_bugfixes.patch`. The changed sites carry explanatory
comments but not the "FORK FIX" prefix.

---

## 11. dom-simplify loops forever

**Upstream:** 0ef8ebe89f (2020-12-14, "fix #4895"), the same commit as fix 6;
this is its round cap. Upstream's later rewrite of dom-simplify also
terminates and simplifies the goal further, to `false`.

**Symptom.** On some goals the forward and backward passes of
`simplify_goal` keep changing the goal and never reach a fixed point: z3
prints the first `unsat` and then hangs on `(then dom-simplify smt)` or
`(apply dom-simplify)`. With the fix the tactic stops after 10 rounds, in
about 0.05 s; its goal is equivalent to `false`.

**Cause.** `dom_simplify_tactic::simplify_goal` loops `while (change)` with
no bound.

**Fix.** `src/tactic/core/dom_simplify_tactic.cpp:376`:
`while (change && n < 10)`, with the counter `n` declared at `:372` and
incremented at `:378` (in `simplify_goal`, `src/tactic/core/dom_simplify_tactic.cpp:368`).

**Test.** `TST(z3_4_7_1_dom_simplify_oscillation)`,
`src/test/z3_4_7_1_bugs.cpp:438` (registered at `src/test/main.cpp:218`). A
20-second SIGALRM watchdog turns the hang into a test failure (POSIX only).

---

## 12. `update-field` over a constructor leaves the field unchanged

**Upstream:** f989e4eb38 (2019-05-14, "fix #2276", "update-field does not
update").

**Symptom.** Rewriting `((_ update-field f) (mk a b) v)` returns the
constructor term unchanged, so a problem that asserts the updated field is
reported `sat` with the old field value (for example `sat (mkp i red)`
where the correct answer is `unsat (mkp j red)`).

**Cause.** `datatype_rewriter::mk_app_core` called
`get_accessor_constructor(f)` on the update-field declaration `f` itself
instead of on the accessor it updates, so the constructor never matched and
the term was returned unchanged.

**Fix.**
* `src/ast/datatype_decl_plugin.h:377` and
  `src/ast/datatype_decl_plugin.cpp:834-837`: new
  `datatype::util::get_update_accessor(update)`, which returns the accessor
  stored in the update-field declaration's parameter.
* `src/ast/rewriter/datatype_rewriter.cpp:73-74`: `mk_app_core` takes the
  accessor with `get_update_accessor(f)` and compares the constructor of
  that accessor (and matches `acc` against the constructor's accessors at
  `:84`).

**Test.** `TST(z3_4_7_1_update_field_constructor)`,
`src/test/z3_4_7_1_bugs.cpp:465` (registered at `src/test/main.cpp:219`).

---

## 13. PDR: nullary predicates share one search-cache entry

**Upstream:** none. PDR was removed upstream in 6adaed718 without a fix.

**Symptom.** With two or more predicates that take no arguments, PDR prints
`sat` on an unsat problem and then segfaults.

**Cause.** `model_search` keys its node cache, and `update_models`, by the
state formula only. Every nullary predicate has the state `true`, so the
obligation for one predicate is closed as a duplicate of another's.

**Fix.** The cache is keyed by level, then predicate, then state, and
`update_models` uses an `obj_pair_map`; `context::reset` frees the search
nodes before the predicate transformers (otherwise the fixed version
crashes in the destructor). `FORK FIX #13` comments at
`src/muz/pdr/pdr_context.h:264` and `src/muz/pdr/pdr_context.cpp:52`,
`:1004`, `:1155`, `:1175` and `:1487`.

**Test.** `TST(z3_4_7_1_pdr_nullary_cache)`, `src/test/z3_4_7_1_bugs.cpp:511`
(registered at `src/test/main.cpp:228`).

---

## 14. Duality: a loop predicate used twice in one body breaks unhoisting

**Upstream:** none. Duality was removed upstream in 4f5775c53 without a fix.

**Symptom.** When the loop predicate appears twice in one rule body, Duality
fails with `Wrong number of arguments (4) passed to function inv!2` and
answers `unknown`. The correct answer is `sat`.

**Cause.** `UnhoistLoop` rewrites an incoming edge once per child position,
so an edge with two positions is rewritten twice.

**Fix.** `src/duality/duality_rpfp.cpp:3946` (`FORK FIX #14`): skip an edge
that has already been rewritten.

**Test.** `TST(z3_4_7_1_duality_unhoist_repeated_child)`,
`src/test/z3_4_7_1_bugs.cpp:541` (registered at `src/test/main.cpp:229`).

---

## 15. `lazy_sparse` tables: union into an empty nullary register segfaults

**Upstream:** still present (z3 5.4.2 segfaults with the `fp.` options).

**Symptom.** With `datalog.default_table=lazy_sparse`, a union into an empty
relation with no columns segfaults in
`sparse_table::column_layout::make_byte_aligned_end`.

**Cause.** `lazy_table_plugin::mk_empty` always asks the wrapped sparse
plugin, which cannot build a table with no columns.

**Fix.** `src/muz/rel/dl_lazy_table.cpp:36` (`FORK FIX #15`): when the
wrapped plugin cannot handle the signature, fall back to
`get_manager().mk_empty_table(s)`.

**Test.** `TST(z3_4_7_1_lazy_sparse_nullary_union)`,
`src/test/z3_4_7_1_bugs.cpp:564` (registered at `src/test/main.cpp:230`).

---

## 16. QBMC evaluates rule variables at the rule index

**Upstream:** 784b613b5 (#11540). Only the evaluation-level part is ported
here; see "Known 4.7.1 bugs" for the hyper-resolution positions it also
changes.

**Symptom.** Quantified BMC aborts on a `VERIFY` failure while it builds the
answer.

**Cause.** `get_model` evaluates the rule variables at the rule's index in the
rule set instead of at the trace level, and does not stop at a fact.

**Fix.** `src/muz/bmc/dl_bmc_engine.cpp:290` and `:348` (`FORK FIX #16`):
evaluate at `level`, stop at any fact, `VERIFY(level > 0)`.

**Test.** `TST(z3_4_7_1_qbmc_eval_level)`, `src/test/z3_4_7_1_bugs.cpp:593`
(registered at `src/test/main.cpp:231`). It checks that the output starts
with `sat` and that the proof contains `(asserted (p 0))`; the exact proof
depends on the model.

---

## 17. DDNF: `extract` equalities use the constant's width

**Upstream:** still present (z3 5.4.2 segfaults).

**Symptom.** A rule with an equality over `((_ extract i j) x)` segfaults the
DDNF engine.

**Cause.** `compile_eq` looks up the DDNF by the width of the constant
instead of the width of the variable.

**Fix.** `src/muz/ddnf/ddnf.cpp:855` (`FORK FIX #17`): use
`bv.get_bv_size(v)`.

**Test.** `TST(z3_4_7_1_ddnf_extract_width)`, `src/test/z3_4_7_1_bugs.cpp:629`
(registered at `src/test/main.cpp:232`).

---

## 18. `min_cut` never terminates on graphs with a shared node

**Upstream:** 8150cd6e8 (#11401).

**Symptom.** Spacer with `fixedpoint.spacer.minimize_unsat_core=true` hangs in
`min_cut::augment_path` when two augmenting paths share a node.

**Cause.** `augment_path` adds the reverse arc to the sink instead of to the
predecessor.

**Fix.** `src/util/min_cut.cpp:166` (`FORK FIX #18`): add the reverse arc
to `edge(l, max)`.

**Test.** `TST(z3_4_7_1_min_cut_shared_node)`, `src/test/z3_4_7_1_bugs.cpp:682`
(registered at `src/test/main.cpp:233`), with a 20-second watchdog.

---

## 19. Quantifier instantiation builds its weakening proof wrongly

**Upstream:** still present.

**Symptom.** With `fixedpoint.xform.instantiate_quantifiers=true`, a query
segfaults in `resolve_rule` / `mk_hyper_resolve`. With proofs on, it also
segfaults earlier.

**Cause.** `instantiate_rule` builds the weakening proof without turning on
proof mode, so in a normal query the rule gets no proof. It also passes the
modus-ponens premises in the wrong order, which is undefined behaviour.

**Fix.** `src/muz/transforms/dl_mk_quantifier_instantiation.cpp:31` (include)
and `:242` (`FORK FIX #19`): build the proof under `scoped_proof` and call
`mk_modus_ponens(p1, mk_def_axiom(...))`.

**Test.** `TST(z3_4_7_1_qi_weakening_proof)`, `src/test/z3_4_7_1_bugs.cpp:715`
(registered at `src/test/main.cpp:234`), with proof mode both off and on.

---

## 20. `pb2bv` with a numeral on the left

**Upstream:** still present in 5.4.2: `(= 3 x)` is still encoded as `x`.

**Symptom.** For a 0/1 variable `x`, `(= 3 x)` is encoded as `x`, so
`(then pb2bv smt)` answers `sat` on an unsat problem. `(<= c x)` and
`(>= c x)` skip their range checks.

**Cause.** `pb2bv_tactic::imp::convert` has a separate numeral-on-the-left
branch without the checks of the numeral-on-the-right branch.

**Fix.** `src/tactic/arith/pb2bv_tactic.cpp:636` (`FORK FIX #20`): swap the
operands, flip `<=`/`>=`, and use the checked branch.

**Test.** `TST(z3_4_7_1_pb2bv_numeral_lhs)`, `src/test/z3_4_7_1_bugs.cpp:786`
(registered at `src/test/main.cpp:220`).

---

## 21. `eq2bv` keeps its rewriter cache across goals

**Upstream:** still present in 5.4.2.

**Symptom.** After `split-clause`, a bit-vector created for the first goal
leaks into the second, so `(then split-clause eq2bv smt)` answers `sat` on an
unsat problem.

**Cause.** The tactic's rewriter `m_rw` is never reset between goals.

**Fix.** `src/tactic/arith/eq2bv_tactic.cpp:158` (`FORK FIX #21`):
`m_rw.reset()` per goal.

**Test.** `TST(z3_4_7_1_eq2bv_stale_cache)`, `src/test/z3_4_7_1_bugs.cpp:808`
(registered at `src/test/main.cpp:221`).

---

## 22. `purify-arith` finds the unsafe trig arguments too late

**Upstream:** still present in 5.4.2.

**Symptom.** `(then purify-arith qfnra)` answers `sat` with an invalid model
(for example `x = acos(...)` that violates the assertions).

**Cause.** The scan for unsafe `sin`/`cos` arguments runs lazily, after the
formulas that mention them have already been replaced by purified ones.

**Fix.** `src/tactic/arith/purify_arith_tactic.cpp:746` (`FORK FIX #22`):
call `find_unsafe()` at the start of `operator()`.

**Test.** `TST(z3_4_7_1_purify_trig_unsafe)`, `src/test/z3_4_7_1_bugs.cpp:829`
(registered at `src/test/main.cpp:222`).

---

## 23. UFBV demodulator leaves a variable unbound

**Upstream:** b76ed6a47 ("proper fix to #6476", 2022-12-04).

**Symptom.** The UFBV demodulator rewrites with a substitution that leaves a
variable unbound, giving a wrong `sat`.

**Cause.** The matcher skips identical subterms without binding the variables
they contain.

**Fix.** `src/tactic/ufbv/ufbv_rewriter.cpp:781` (`FORK FIX #23`): bind the
unbound variables of identical subterms.

**Test.** `TST(z3_4_7_1_ufbv_demod_unbound_var)`,
`src/test/z3_4_7_1_bugs.cpp:855` (registered at `src/test/main.cpp:223`).
The bug depends on AST ids, so the test runs the script through a shell-like
`cmd_context` rather than `Z3_eval_smtlib2_string`.

---

## 24. `recover-01` reads past the end of the goal

**Upstream:** c230d89a3 ("fix #5294").

**Symptom.** Segfault when a rewrite turns the goal into `false` before the
last formula.

**Cause.** The rewrite loop keeps the goal size it read before the loop.

**Fix.** `src/tactic/arith/recover_01_tactic.cpp:369` (`FORK FIX #24`):
re-read the size each round.

**Test.** `TST(z3_4_7_1_recover01_inconsistent)`,
`src/test/z3_4_7_1_bugs.cpp:888` (registered at `src/test/main.cpp:224`).

---

## 25. `distribute-forall` crashes with proofs on

**Upstream:** d530d1314 ("fix #3276").

**Symptom.** `(apply distribute-forall)` segfaults with `:produce-proofs true`.

**Cause.** The rewrite is applied without building its proof.

**Fix.** `src/tactic/core/distribute_forall_tactic.cpp:57` and `:81`
(`FORK FIX #25`): prove the split with `mk_push_quant`.

**Test.** `TST(z3_4_7_1_distribute_forall_proof)`,
`src/test/z3_4_7_1_bugs.cpp:927` (registered at `src/test/main.cpp:225`).

---

## 26. `pb2bv` on fractional 0/1 sums

**Upstream:** 10768bd00 ("fix #3727") and f7a7b9e1f ("fix #4108").

**Symptom.** A sum such as `(+ (ite a 0.5 0.0) ...)` is turned into a
malformed, always-true circuit, and z3 answers `sat`.

**Cause.** The rewriter does not check that the bound and the coefficients are
integral.

**Fix.** `src/ast/rewriter/pb2bv_rewriter.cpp:316` and `:348`
(`FORK FIX #26`): leave the sum unchanged unless the bound and every
coefficient are integral.

**Test.** `TST(z3_4_7_1_pb2bv_fractional)`, `src/test/z3_4_7_1_bugs.cpp:946`
(registered at `src/test/main.cpp:226`).

---

## 27. `lia2card` never terminates on an equation with no terms

**Upstream:** still present in 5.4.2 with `:compile_equality false` (the
upstream default is now `true`, which hides it).

**Symptom.** `(assert (= 2 3))` with `(apply (using-params lia2card
:compile_equality false))` never terminates.

**Cause.** `lia2card_tactic::mk_eq` builds `(and false true)`. `goal::update`
flattens it to `false` but appends the unflattened formula, so the goal never
becomes inconsistent and the rewrite loop keeps appending.

**Fix.** `src/tactic/arith/lia2card_tactic.cpp:273` (`FORK FIX #27`): fold a
sum with no terms in `mk_eq` when `compile_equality` is off. The underlying
`goal::update` defect is listed below.

**Test.** `TST(z3_4_7_1_lia2card_empty_eq)`, `src/test/z3_4_7_1_bugs.cpp:974`
(registered at `src/test/main.cpp:227`), with a 20-second watchdog.

---

## 28. nlsat: negative sign stored as unsigned in infinity projection

**Upstream:** 363d7aad2.

**Symptom.** On a quantified nonlinear real problem with `nlqsat`, z3 never terminates where the answer is `unsat`.

**Cause.** `project_plus_infinity` and `project_minus_infinity` keep the sign of a coefficient in an `unsigned`, so a negative sign is treated as positive.

**Fix.** `src/nlsat/nlsat_explain.cpp:1673` and `:1689` (`FORK FIX #28`): the sign is an `int`.

**Test.** `TST(z3_4_7_1_nlsat_signed_infinity)`, `src/test/z3_4_7_1_bugs.cpp:1040`
(registered at `src/test/main.cpp:235`).

---

## 29. nlsat: atom sign wrong when the factor-sign product overflows

**Upstream:** f79dccccf and 6071797ba (#2238).

**Symptom.** With large factor values (for example three factors of 2^22) the nlsat search diverges: the atom gets the wrong truth value.

**Cause.** `eval_ineq` multiplies the raw (unnormalized) signs of the factors in 32-bit `int`, which overflows.

**Fix.** `src/nlsat/nlsat_evaluator.cpp:387` (`FORK FIX #29`): normalize each sign to -1, 0 or 1 before multiplying.

**Test.** `TST(z3_4_7_1_nlsat_eval_sign_overflow)`, `src/test/z3_4_7_1_bugs.cpp:1081`
(registered at `src/test/main.cpp:236`).

---

## 30. nlsat: `solve_eq` keeps a stale power table

**Upstream:** 19e2f8c9d (#1694).

**Symptom.** When a later polynomial has a higher degree than an earlier one, `solve_eq` computes a wrong value (seen at unit level; no SMT-LIB input diverged).

**Cause.** The vector of powers of the leading coefficient is never reset between polynomials.

**Fix.** `src/nlsat/nlsat_explain.cpp:1729` and `:1747` (`FORK FIX #30`): each polynomial indexes its own power tables (upstream's As/Bs).

**Test.** `TST(z3_4_7_1_nlsat_solve_eq_powers)`, `src/test/z3_4_7_1_bugs.cpp:1118`
(registered at `src/test/main.cpp:237`).

---

## 31. nlsat: `s == -1` instead of `s < 0` in normalize/simplify

**Upstream:** 0f20175bd (#2556).

**Symptom.** The explain step can produce an invalid lemma (shown at unit level); on the audit input only the trace changes.

**Cause.** `normalize` and `simplify` test the sign with `s == -1`, but the sign can be any negative number.

**Fix.** `src/nlsat/nlsat_explain.cpp:432` and `:1138` (`FORK FIX #31`): test `s < 0`.

**Test.** `TST(z3_4_7_1_nlsat_simplify_core_sign)`, `src/test/z3_4_7_1_bugs.cpp:1158`
(registered at `src/test/main.cpp:238`).

---

## 32. nlsat `inline_vars`: vanished factors and constant atoms

**Upstream:** 29f3f6a7a (#3144) and 99784a92e (#3225).

**Symptom.** With `nlsat.inline_vars=true`, z3 segfaults when a factor vanishes after substitution, and gives an invalid model when an atom becomes constant.

**Cause.** The substitution does not give a zero product its value, does not fold constant atoms, and leaves an empty clause instead of `false`.

**Fix.** `src/nlsat/nlsat_solver.cpp:2373`, `:2385`, `:2409` and `:2451` (`FORK FIX #32`).

**Test.** `TST(z3_4_7_1_nlsat_inline_vars)`, `src/test/z3_4_7_1_bugs.cpp:1220`
(registered at `src/test/main.cpp:239`).

---

## 33. nlsat: reorder after `inline_vars` gives an invalid model

**Upstream:** 983a55232 (#3167).

**Symptom.** With inlined variables and variable reordering, the model assigns values to the wrong variables.

**Cause.** `fix_patch` runs after `reorder` and indexes the reordered assignment with pre-reorder variables.

**Fix.** `src/nlsat/nlsat_solver.cpp:2019` (`FORK FIX #33`): `can_reorder()` returns false while patches exist.

**Test.** `TST(z3_4_7_1_nlsat_inline_reorder)`, `src/test/z3_4_7_1_bugs.cpp:1243`
(registered at `src/test/main.cpp:240`).

---

## 34. `qe-sat`: `project_real` chooses the bound family from raw counts

**Upstream:** none; `qe-sat`, the only caller, was removed upstream in 8717c7d7f, so the code is unreachable there.

**Symptom.** `(check-sat-using (using-params qe-sat :projection_mode false))` never terminates on inputs with strict bounds.

**Cause.** `project_real` compares raw lower and upper bound counts, while `assign` uses `get_bound_sizes`, which doubles non-strict bounds for reals, so the two disagree on the branch.

**Fix.** `src/qe/qe_arith_plugin.cpp:1931` (`FORK FIX #34`): choose the bound family with `get_bound_sizes`.

**Test.** `TST(z3_4_7_1_qe_project_real_direction)`, `src/test/z3_4_7_1_bugs.cpp:1275`
(registered at `src/test/main.cpp:241`).

---

## 35. `qe`: `mod` by 0 or by a negative constant

**Upstream:** still present (negative cases).

**Symptom.** `(apply qe)` returns `false` for `(exists ((x Int)) (and (= (mod (+ x a) 0) 0) (< x b)))` and for negative moduli, which is unsound (x=3, a=0, b=10 satisfies the negative-modulus example).

**Cause.** `is_divides` and the nested-mod scan accept a zero or negative divisor, so `div_z` works with a zero or negative lcm.

**Fix.** `src/qe/qe_arith_plugin.cpp:45` and `:1329` (`FORK FIX #35`): a divisibility atom needs a positive modulus; others are left alone, so the quantifier is kept.

**Test.** `TST(z3_4_7_1_qe_mod_divisor)`, `src/test/z3_4_7_1_bugs.cpp:1298`
(registered at `src/test/main.cpp:242`).

---

## 36. `qe`: datatype substitution drops the recognizer

**Upstream:** still present.

**Symptom.** `(apply qe)` turns `(exists ((x Lst)) (and ((_ is nil) x) (or (= x y) (= x z))))` into `true`; with `is-cons y` and `is-cons z` added, `(then qe smt)` says `sat` on an unsat problem.

**Cause.** `datatype_plugin::subst_rec` substitutes `x := t` without keeping the recognizer constraint `r(t)`.

**Fix.** `src/qe/qe_datatype_plugin.cpp:725` (`FORK FIX #36`): conjoin `r(t)`.

**Test.** `TST(z3_4_7_1_qe_dt_eq_recognizer)`, `src/test/z3_4_7_1_bugs.cpp:1318`
(registered at `src/test/main.cpp:243`).

---

## 37. qe-light: Fourier-Motzkin keeps a stale Boolean-variable map

**Upstream:** still present (`qe_lite_tactic.cpp` `fm::init`).

**Symptom.** With two quantified formulas in one goal, qe-light maps a Boolean literal to a stale variable: it prints `(or b c)` where `(or b a)` is correct.

**Cause.** `fm::init` resets `m_expr2var` twice and `m_expr2bvar` never.

**Fix.** `src/qe/qe_lite.cpp:1544` (`FORK FIX #37`): reset `m_expr2bvar`.

**Test.** `TST(z3_4_7_1_qe_lite_fm_bvar)`, `src/test/z3_4_7_1_bugs.cpp:1337`
(registered at `src/test/main.cpp:244`).

---

## 38. maxres upper-bound block: `pb_util::mk_lt` aliases its coefficients

**Upstream:** fixed upstream by rewriting `mk_lt` as `not(mk_ge)` (commit not identified).

**Symptom.** With `opt.maxres.add_upper_bound_block=true`, maxres stops at the initial model's cost (19 where the optimum is 3).

**Cause.** `mk_lt` normalizes into `m_coeffs` and passes that vector to `mk_ge`, whose `normalize` starts with `m_coeffs.reset()`, so the blocker becomes a `pbge` with all-zero coefficients (`false`).

**Fix.** `src/ast/pb_decl_plugin.cpp:177` (`FORK FIX #38`): pass copies of the coefficients.

**Test.** `TST(z3_4_7_1_pb_mk_lt_coeffs)`, `src/test/z3_4_7_1_bugs.cpp:1359`
(registered at `src/test/main.cpp:245`).

---

## 39. OPB parser rejects `max:` objectives

**Upstream:** still present.

**Symptom.** An OPB file with `max: +2 x1 +3 x2 ;` fails with `unexpected char: a` and exit status 3.

**Cause.** `opb::parse` calls `parse_token("min:")` first, and `parse_token` consumes the matching `m`.

**Fix.** `src/opt/opt_parse.cpp:292` (`FORK FIX #39`): consume the shared `m` once, then test `in:` / `ax:`. The meaning kept for `max:` (minimize the negated sum) is unchanged and itself questionable.

**Test.** `TST(z3_4_7_1_opb_max_objective)`, `src/test/z3_4_7_1_bugs.cpp:1379`
(registered at `src/test/main.cpp:246`).

---

## 40. Probe `+` and `*` use the first argument twice

**Upstream:** still present.

**Symptom.** `(+ 1 2)` evaluates to 4 and `(* 2 3)` to 12 in probe expressions.

**Cause.** `MK_NARY_PROBE` starts the fold at child 1 after initializing the result with child 1.

**Fix.** `src/cmd_context/tactic_cmds.cpp:763` (`FORK FIX #40`): fold from child 2.

**Test.** `TST(z3_4_7_1_probe_nary_fold)`, `src/test/z3_4_7_1_bugs.cpp:1397`
(registered at `src/test/main.cpp:247`).

---

## 41. z3 cannot replay its own pseudo-Boolean logs

**Upstream:** fed977b49 (#1782).

**Symptom.** Replaying a log that calls `Z3_mk_pble`, `Z3_mk_pbge` or `Z3_mk_pbeq` always fails.

**Cause.** The generated logger writes `int` array arguments as unsigned values with `Au`, but the replayer reads them back as a signed array.

**Fix.** `scripts/update_api.py:958` (`FORK FIX #41`, generates `api_log_macros.cpp`), `src/api/z3_logger.h:45` and `src/api/z3_replayer.cpp:500` and `:509`: log and replay `int` arrays as `int` arrays.

**Test.** `TST(z3_4_7_1_pb_log_replay)`, `src/test/z3_4_7_1_bugs.cpp:1413`
(registered at `src/test/main.cpp:248`).

---

## 42. `(get-user-tactics)` segfaults

**Upstream:** 202ed79a2 (#5445).

**Symptom.** Any `(get-user-tactics)` after `declare-tactic` segfaults.

**Cause.** `sexpr::display` falls through after displaying an atom and reads it as a composite.

**Fix.** `src/util/sexpr.cpp:188` (`FORK FIX #42`): return after `display_atom`.

**Test.** `TST(z3_4_7_1_sexpr_display_atom)`, `src/test/z3_4_7_1_bugs.cpp:1454`
(registered at `src/test/main.cpp:249`).

---

## 43. Duality with `xform.quantify_arrays` does not restore the rules

**Upstream:** none. Duality was removed upstream.

**Symptom.** With `fixedpoint.xform.quantify_arrays=true`, queries after the
first one are answered against the transformed rules of an earlier query.
With rules `p(1.0)`, `q1 <- p(1.0)`, `q2 <- p(2.0)` and the queries `q1, q2,
q1`, z3 prints `sat unsat unsat`, and with `xform.scale=true` it prints `sat
sat sat`. The correct answers are `sat unsat sat`.

**Cause.** The quantify_arrays branch of `dl_interface::query` runs
`apply_default_transformation(m_ctx)` on the context's own rules and never
restores them, unlike PDR and Spacer, which save `old_rules` and call
`replace_rules`. The first query's cone-of-influence filter removes the other
queries' rules, and its `query!0` output predicate stays in the context.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:169-170` and `:195-196` (`FORK FIX #43`): save
`old_rules` before `mk_query`, and call `m_ctx.reopen();
m_ctx.replace_rules(old_rules);` after the transformed formulas are read.

**Test.** `TST(z3_4_7_1_duality_quantify_arrays_requery)`, `src/test/z3_4_7_1_bugs.cpp:1470`
(registered at `src/test/main.cpp:250`).

---

## 44. Duality segfaults on the query after a failed one

**Upstream:** none. Duality was removed upstream.

**Symptom.** After a query that fails (for example "theory not supported by
interpolation or bad proof"), the next query segfaults.

**Cause.** `dl_interface::query` replaces `_d` with a fresh `duality_data`
before the query can fail, so after the exception `old_rs` is null, and the
next query calls `old_rs->GetCounterexample()`.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:144` (`FORK FIX #44`): skip
the counterexample swap when `old_rs` is null, as the existing guard on
`LearnFrom` already does.

**Test.** `TST(z3_4_7_1_duality_failed_query_requery)`, `src/test/z3_4_7_1_bugs.cpp:1497`
(registered at `src/test/main.cpp:251`).

---

## 45. `model_implicant`: a nested array equality stays marked true

**Upstream:** still present in z3 5.4.2 (`src/model/model_implicant.cpp:600-604`).

**Symptom.** For two arrays of arrays that agree at one index and differ at
another, `(not (= a b))` evaluates to false; `minimize_literals` then hits
`UNREACHABLE` ("formula false in model") and the process crashes. Reached
from PDR/Spacer model implicants.

**Cause.** In `eval_array_eq`, the nested-array branch marks the inner
equality true at one index and `continue`s without clearing the mark, so the
later differing index never makes the outer equality false.

**Fix.** `src/model/model_implicant.cpp:605` (`FORK FIX #45`): `set_unknown(e);`
before the `continue`.

**Test.** `TST(z3_4_7_1_implicant_nested_array_eq)`, `src/test/z3_4_7_1_bugs.cpp:1522`
(registered at `src/test/main.cpp:252`).

---

## 46. Duality: underapproximation flags fixed in `FixCurrentState`

**Upstream:** none. Duality was removed upstream.

**Symptom.** With `fixedpoint.duality.use_underapprox=true`, a satisfiable
problem aborts with an uncaught "error in ExpansionChoices".

**Cause.** `FixCurrentState` passes an empty don't-care set, so the
`@under` flag literals are frozen to their current values, and the expansion
choices later contradict them.

**Fix.** `src/duality/duality_rpfp.cpp:2521` (`FORK FIX #46`): add the
`underapprox_flag_rev` keys to `dont_cares`. Its red test passes only together
with fix 47.

**Test.** `TST(z3_4_7_1_duality_underapprox_flag_literals)`, `src/test/z3_4_7_1_bugs.cpp:1566`
(registered at `src/test/main.cpp:253`).

---

## 47. Duality: `Build` interpolates a stale assumption-based proof

**Upstream:** none. Duality was removed upstream.

**Symptom.** With `use_underapprox`, iZ3 fails with "unknown assertion" on a
satisfiable problem.

**Cause.** `Build` interpolates the proof of the assumption-based
`CheckWithUnderapprox`; the plain `tree->Check(top)` that would produce an
interpolatable proof is commented out.

**Fix.** `src/duality/duality_solver.cpp:1958` (`FORK FIX #47`): re-enable
`tree->Check(top);` before interpolating.

**Test.** `TST(z3_4_7_1_duality_underapprox_stale_proof)`, `src/test/z3_4_7_1_bugs.cpp:1587`
(registered at `src/test/main.cpp:254`).

---

## 48. PDR proof trace: cache keyed by state only

**Upstream:** none. PDR was removed upstream (6adaed718).

**Symptom.** With `fixedpoint.generate_proof_trace=true`, when the `query!0` root and a body predicate are nullary (state `true`), the certificate proves the wrong fact (for example `(forall x (P x))` instead of `query!0`).

**Cause.** `model_search::get_proof_trace` caches proofs and rules by state only, as fix 13 did for the search cache.

**Fix.** `src/muz/pdr/pdr_context.cpp:1308` and the other `FORK FIX #48` sites: key both caches by `(pt().head(), state())` with `obj_pair_map`.

**Test.** `TST(z3_4_7_1_pdr_proof_nullary_root)`, `src/test/z3_4_7_1_bugs.cpp:1606`
(registered at `src/test/main.cpp:255`).

---

## 49. PDR proof trace: unresolved duplicate leaf

**Upstream:** none. PDR was removed upstream (6adaed718).

**Symptom.** A repeated obligation that the search closed as a duplicate without children is reached first, so its unresolved rule is cached and the certificate concludes `(=> (a true) query!0)`.

**Cause.** `get_proof_trace` resolves nodes in visit order and caches the first one seen.

**Fix.** `src/muz/pdr/pdr_context.cpp:1320` and `:1345` (`FORK FIX #49`): collect the first expanded node per (predicate, state) and resolve it in place of a childless duplicate.

**Test.** `TST(z3_4_7_1_pdr_proof_duplicate_leaf)`, `src/test/z3_4_7_1_bugs.cpp:1639`
(registered at `src/test/main.cpp:256`).

---

## 50. PDR Boogie certificate without a proof trace

**Upstream:** none. PDR was removed upstream (6adaed718).

**Symptom.** `fixedpoint.print_boogie_certificate=true` without `generate_proof_trace` segfaults.

**Cause.** The rules then have no proof; only `SASSERT(p1)` guarded the dereference.

**Fix.** `src/muz/pdr/pdr_context.cpp:1374` (`FORK FIX #50`): use `m.mk_asserted(rule formula)` when a rule has no proof.

**Test.** `TST(z3_4_7_1_pdr_boogie_without_proof_trace)`, `src/test/z3_4_7_1_bugs.cpp:1690`
(registered at `src/test/main.cpp:257`).

---

## 51. PDR proof trace after merged tails

**Upstream:** none. PDR was removed upstream (6adaed718).

**Symptom.** When `remove_duplicate_tails` merged a child's tail away, `get_proof_trace` segfaults in `rule_unifier::unify_rules` (reads `get_tail(0)` of a rule with no tails left).

**Cause.** It always unifies tail 0 of the partly resolved rule with the next child.

**Fix.** `src/muz/pdr/pdr_context.cpp:1406`, `:1441` and `:1451` (`FORK FIX #51`): unify only while an uninterpreted tail is left; if unification fails, print the IF_VERBOSE message and skip that child; keep the resolved children in a separate premises vector.

**Test.** `TST(z3_4_7_1_pdr_proof_merged_tails)`, `src/test/z3_4_7_1_bugs.cpp:1718`
(registered at `src/test/main.cpp:258`).

---

## 52. Rewriter keeps stale stacks after a `max_steps` exception

**Upstream:** 18bb90f93 (2020-05-21, "fix #4426") and b013df9a9 (2020-06-04,
"fix #4431").

**Symptom.** After a rewrite stops with "max. steps exceeded", the next
rewrite on the same rewriter (for example the next model evaluation) returns a
stale result: "Sort mismatch at argument #2 ... supplied sort is Bool", or a
segfault for small `model_evaluator.max_steps`. Seen in model finding with
`model.partial`/MBQI and `max_steps` (g3 vector
`z3_bug_proto_eval_max_steps.smt2`).

**Cause.** `rewriter_tpl::operator()` starts with the frame, result, scope
and cache stacks left over from the interrupted rewrite.

**Fix.** `src/ast/rewriter/rewriter_def.h:743` (`FORK FIX #52`): reset the
frames, results, result proofs, scopes and cache when they are not empty, as
upstream does.

**Test.** `TST(z3_4_7_1_rewriter_stale_stacks)`, `src/test/z3_4_7_1_bugs.cpp:1758`
(registered at `src/test/main.cpp:259`).

---

## 53. `qe` finite-domain plugin: negative atoms in the "other value" branch

**Upstream:** still present (`src/qe/qe_dl_plugin.cpp:167`).

**Symptom.** Over a finite-domain sort of size 4, `(exists ((x S)) (not (= x
c)))` eliminates to `false`; the correct result is `true`.

**Cause.** `dl_plugin::subst_large_domain` substitutes `true` for the
negatively occurring atom `(= x c)` in the "other value" branch, contradicting
`assign_large_domain` (which asserts `not(neq_atom)`) and the datatype
plugin's `subst_large_domain`.

**Fix.** `src/qe/qe_dl_plugin.cpp:165` (`FORK FIX #53`): substitute `false`.

**Test.** `TST(z3_4_7_1_qe_dl_other_value_neq)`, `src/test/z3_4_7_1_bugs.cpp:1799`
(registered at `src/test/main.cpp:260`).

---

## 54. Sieve-relation negation: inverted inner-column test

**Upstream:** still present (`src/muz/rel/dl_sieve_relation.cpp:636` in 5.4.2).

**Symptom.** `sieve_relation_plugin::mk_filter_by_negation_fn` mishandles
the equality columns. An equality between two inner columns is dropped, so the
inner negation runs on zero columns and removes every target row whenever the
negated relation is nonempty. An equality on a full (sieved) column of the
negated relation, which always holds, yields the identity filter instead of
subtracting. Sieve relations come from `mk_explanations` with
`datalog.explanations_on_relation_level=true` (API `query(expr)` path).

**Cause.** `r_col_inner = r_sieved && !sr->is_inner_col(r_cols[i])` (and the
same for `neg_col_inner`) is true exactly when the column is *not* inner, and
`get_inner_col(i)` uses the loop index instead of `r_cols[i]` / `neg_cols[i]`.

**Fix.** `src/muz/rel/dl_sieve_relation.cpp:637` (`FORK FIX #54`):
`r_col_inner = !r_sieved || sr->is_inner_col(r_cols[i])`, the same for the
negated side, and `get_inner_col(r_cols[i])` / `get_inner_col(neg_cols[i])`
(a non-sieve side uses the column itself).

**Test.** `TST(z3_4_7_1_sieve_negation_inner_cols)`, `src/test/z3_4_7_1_bugs.cpp:1835`
(registered at `src/test/main.cpp:262`).

---

## 55. `check_table` with a sparse member: nullary `mk_empty` segfaults

**Upstream:** still present (`src/muz/rel/dl_check_table.cpp:52` in 5.4.2).

**Symptom.** With `fixedpoint.datalog.default_table_checked=true` and sparse as
either the default table or the checker, a rule that projects to a nullary
table before a negation segfaults in
`sparse_table::column_layout::make_byte_aligned_end`; the correct answer is
`sat`.

**Cause.** `instr_union` creates its empty target with
`r_src.get_plugin().mk_empty(r_src)`. `check_table_plugin::mk_empty` passes the
0-column signature to both wrapped plugins without checking
`can_handle_signature`, and `sparse_table_plugin` needs at least one column
(the same pattern as fix 15).

**Fix.** `src/muz/rel/dl_check_table.cpp:52` (`FORK FIX #55`): a wrapped
plugin that cannot handle the signature is replaced by
`get_manager().mk_empty_table(s)`, as `lazy_table_plugin::mk_empty` does.

**Test.** `TST(z3_4_7_1_check_table_nullary_empty)`, `src/test/z3_4_7_1_bugs.cpp:1908`
(registered at `src/test/main.cpp:263`).

---

## 56. `bitvector_table`: nullary tables reserve no bits

**Upstream:** fixed in 8dac9b7b94 ("fix #3814").

**Symptom.** With `fixedpoint.datalog.default_table=bitvector`, any nullary
predicate (for example `(declare-rel g ()) (rule g) (query g)`) segfaults in
`bitvector_table::add_fact`; the correct answer is `sat`.

**Cause.** The `bitvector_table` constructor calls `m_bv.reserve(1 << shift)`
inside the loop over the columns, so a table with no columns reserves nothing.

**Fix.** `src/muz/rel/dl_table.cpp:257` (`FORK FIX #56`): reserve after the
loop, as upstream does; a nullary table holds one bit.

**Test.** `TST(z3_4_7_1_bitvector_table_nullary)`, `src/test/z3_4_7_1_bugs.cpp:1950`
(registered at `src/test/main.cpp:264`).

---

## 57. `smt.case_split=6`: uninitialized decision phase

**Upstream:** fixed (`src/smt/smt_context.cpp:2006` declares
`lbool phase = l_undef;`).

**Symptom.** With `smt.case_split=6` (theory-aware branching) the decision
phase is undefined; the reference build decides negative every time, whatever
`smt.phase_selection` says, so models and statistics differ from g3 and are
not guaranteed across compilers.

**Cause.** `context::decide` declares `lbool phase;` without a value and
`theory_aware_branching_queue::next_case_split` returns without writing it.

**Fix.** `src/smt/smt_context.cpp:1809` (`FORK FIX #57`): initialize the
phase to `l_undef`, so the configured phase selection decides, as in upstream
and in g3 (AUDIT.md smtcore_02-F9).

**Test.** `TST(z3_4_7_1_case_split6_phase)`, `src/test/z3_4_7_1_bugs.cpp:1985`
(registered at `src/test/main.cpp:265`). The unfixed code reads undefined
behavior, so the test is red in this build but not guaranteed red everywhere.

---

## 58. PB simplex: rows over nested pb atoms

**Upstream:** `pb.enable_simplex` no longer exists (1a1d88dde7 rejects it).

**Symptom.** With `pb.enable_simplex=true`, a pb constraint that has another
pb atom as an argument, for example
`((_ pbge 3 2 1 1) ((_ pbge 2 1 1 1) (not b) (not d) a) b d)`, answers `unsat`
although a=b=true, d=false satisfies it; other inputs segfault. On 400 random
nested scripts the release build was wrong in 131.

**Cause.** In `theory_pb::internalize_atom` the simplex column of a pb atom
`abv` is both the atom's slack column and a 0-1 column of any outer
constraint that mentions it, so the outer row puts [0,1] bounds on the inner
slack (the code carries a "TBD" comment about this case).

**Fix.** `src/smt/theory_pb.cpp:543` (`FORK FIX #58`): a constraint with an
argument that is itself a pb atom with a row gets no simplex row;
`assign_eh` (`:829`) and `pop_scope_eh` (`:1347`) skip atoms without a row
(replacing a `VERIFY`).

**Test.** `TST(z3_4_7_1_pb_simplex_nested)`, `src/test/z3_4_7_1_bugs.cpp:2032`
(registered at `src/test/main.cpp:266`).

---

## 59. PB simplex: 0-1 bound on a reused slack column after pop

**Upstream:** `pb.enable_simplex` no longer exists (1a1d88dde7 rejects it).

**Symptom.** With `pb.enable_simplex=true` and push/pop, satisfiable problems
answer `unsat` after a pop (minimal repro: sat, then unsat; correct: sat,
sat).

**Cause.** `theory_pb::assign_eh` bounds a bool var's 0-1 column whenever
`m_lit_watch[sign] != nullptr`. `pop_scope_eh`/`clear_watch` empty but keep
the watch lists of popped constraints, so when a freed bool var is reused as
a pb atom, whose column is that atom's slack, assigning it puts a 0-1 bound
on the weighted sum.

**Fix.** `src/smt/theory_pb.cpp:797` (`FORK FIX #59`): apply the literal bound
only to 0-1 columns (`m_enable_simplex && m_vars.contains(v)`).

**Test.** `TST(z3_4_7_1_pb_simplex_pop)`, `src/test/z3_4_7_1_bugs.cpp:2060`
(registered at `src/test/main.cpp:267`).

---

## 60. `smt_solver::updt_params` resets solver-level settings

**Upstream:** fixed (`smt_solver::updt_params` uses `solver::get_params()`;
z3 4.15.2 keeps `mbqi=false` in the scenarios below).

**Symptom.** `Z3_get_implied_equalities` calls `updt_params` with just
`produce_models=true`. On a solver created with `smt.mbqi=false` that
re-enables mbqi: a later quantified problem answers `sat` instead of
`unknown`, and if a quantifier was already internalized under `mbqi=false`
the next check segfaults in `model_finder::get_flat_quantifier` (from
`model_checker::check`).

**Cause.** `smt_solver::updt_params(p)` passes `p` alone to
`smt_params::updt_params`, `context::updt_params` and `smt_params_helper`, so
every smt setting not in `p` is re-read from the global defaults.

**Fix.** `src/smt/smt_solver.cpp:79` (`FORK FIX #60`): pass the merged
`solver::get_params()`, as upstream does (AUDIT.md smtcore_03-F10 follow-up).

**Test.** `TST(z3_4_7_1_solver_updt_params)`, `src/test/z3_4_7_1_bugs.cpp:2099`
(registered at `src/test/main.cpp:268`).

---

## 61. `theory_pb`: a pb equality left with no arguments is dropped

**Upstream:** fixed by redesign in bd96eaff47 ("axiomatize pb-eq"); upstream
answers `unsat`. The fork fix is minimal and does not backport the redesign.

**Symptom.** With default options,
`(assert (or a b)) (assert (not b)) (assert ((_ pbeq 1 1 1) a c)) (assert (or c b))`
answers `sat`; `model_validate=true` reports the model invalid. The problem
is unsat.

**Cause.** `theory_pb::internalize_atom` drops the arguments that are already
assigned and subtracts the true ones' coefficients from the bound. When every
argument of an equality is assigned, the constraint keeps no arguments and a
negative bound; `pb_rewriter_util::normalize` only rejects a bound above the
coefficient sum, so it returns `l_undef`, and a constraint with no arguments
is never watched or propagated.

**Fix.** `src/smt/theory_pb.cpp:475` (`FORK FIX #61`): after `normalize`, an
undecided equality with no arguments is false, matching
`pb_rewriter::mk_app_core` (`sz == 0 && k != 0`).

**Test.** `TST(z3_4_7_1_pb_eq_no_args)`, `src/test/z3_4_7_1_bugs.cpp:2135`
(registered at `src/test/main.cpp:269`).

---

## 62. `theory_pb::assert_ge`: uninitialized local theory

**Upstream:** fixed in becf423c77 (theories are built from the context;
`assert_ge` became `theory_pb th(ctx);`).

**Symptom.** `./test-z3 theory_pb` segfaults right after its fuzz phase, on
the first `assert_ge` case ("k:0 4"). Only the unit test reaches this helper;
solver results are unaffected.

**Cause.** `theory_pb::assert_ge` builds a local
`theory_pb th(ctx.get_manager(), p)` and never calls `th.init(&ctx)`, so
`theory::m_manager` stays null; the sorting network's first clause reaches
`theory_pb::justify`, whose `proofs_enabled()` dereferences it.

**Fix.** `src/smt/theory_pb.cpp:1244` (`FORK FIX #62`): call `th.init(&ctx)`
after constructing `th`.

**Test.** `TST(z3_4_7_1_pb_assert_ge_uninit)`, `src/test/z3_4_7_1_bugs.cpp:2163`
(registered at `src/test/main.cpp:270`).

---

## 63. Duality certificates: witnesses for eliminated rule variables

**Upstream:** Duality was removed upstream; no fix to backport.

**Symptom.** With `fixedpoint.engine=duality` and `:print-certificate true`,
a counterexample step prints each rule variable as `tree->Eval(&edge, skolem)`.
A variable that qe-lite eliminated from the edge formula is unconstrained in
the counterexample model and prints the model default, so the certificate is
invalid (for example it certifies `q0(5)` by `x1 = x + 3` with `(= x 0)`).

**Cause.** `print_proof` and `extract_proof` evaluate the original rule's
variables in the edge model, which no longer mentions eliminated variables.

**Fix.** `src/muz/duality/duality_dl_interface.cpp` (`FORK FIX #63`): keep
the raw background axioms (:255) and, per step, search a witness for the
original rule's variables that is consistent with the step's head and body
facts (`witness_clause_shape`, `witness_node_fact`, `find_step_witness`,
:405-603), as g3 does (pkg/duality/counterexample_proof.go). `print_proof`
(:617) and `extract_proof` (:769) use it, falling back to the old value when
no witness is found (g3 AUDIT.md duality_01-F4).

**Test.** `TST(z3_4_7_1_duality_certificate_eliminated_var)`,
`src/test/z3_4_7_1_bugs.cpp:1658` (registered at `src/test/main.cpp:256`).

---

## 64. qe finite-domain plugin: domain size equal to the number of atoms

**Upstream:** still present (`src/qe/qe_dl_plugin.cpp:122` in 5.4.2).

**Symptom.** Over a finite-domain sort of size 2, qe eliminates `(exists ((x S)) (and (not (= x y)) (not (= x z))))` to `true` (with fix 53) where the correct result is `(= y z)`, and `(exists ((x S)) (and (not (= x 0)) (not (= x 1))))` to `true` instead of `false`.

**Cause.** `is_small_domain` uses `domain_size < num_eqs + num_neqs`; at equality the large-domain "other value" branch assumes a value outside the atoms exists, and none does.

**Fix.** `src/qe/qe_dl_plugin.cpp:120` (`FORK FIX #64`): enumerate when `domain_size <= num_eqs + num_neqs` (g3 AUDIT.md qe_01-F4).

**Test.** `TST(z3_4_7_1_qe_dl_domain_boundary)`, `src/test/z3_4_7_1_bugs.cpp:1995`
(registered at `src/test/main.cpp:263`).

---

## 65. `to_fp` of one Real: out-of-bounds read

**Upstream:** the one-Real branch was removed; upstream rejects the input with "Unexpected argument combination".

**Symptom.** `((_ to_fp 3 3) r)` with `r` a Real reads `domain[1]` of a one-element domain (undefined behavior); the reference build reports "sort mismatch, expected one argument of Real sort", and accepting the input would crash `fpa2bv` (no rounding mode).

**Cause.** `fpa_decl_plugin::mk_to_fp` has a "1 Real -> 1 FP" branch that checks `domain[1]`.

**Fix.** `src/ast/fpa_decl_plugin.cpp:514` (`FORK FIX #65`): drop the branch and "(Real)" from the supported-combinations message, as upstream (g3 AUDIT.md ast_05-F2).

**Test.** `TST(z3_4_7_1_to_fp_one_real)`, `src/test/z3_4_7_1_bugs.cpp:2379`
(registered at `src/test/main.cpp:281`).

---

## 66. `z3 -dl <directory>` on POSIX

**Upstream:** still present (`NOT_IMPLEMENTED_YET()` in 5.4.2).

**Symptom.** `z3 -dl wpa/` prints `NOT IMPLEMENTED YET!` and hangs (`UNREACHABLE` in release).

**Cause.** `get_file_names` is implemented only for `_WINDOWS`.

**Fix.** `src/muz/base/dl_util.cpp:601` (`FORK FIX #66`): opendir/readdir with sorted names, the same extension filter and recursion as the Windows code (g3 AUDIT.md cmd_context_02-F11).

**Test.** `TST(z3_4_7_1_get_file_names_posix)`, `src/test/z3_4_7_1_bugs.cpp:2427`
(registered at `src/test/main.cpp:282`).

---

## 67. WPA not-in-sort warning format

**Upstream:** still present.

**Symptom.** The warning prints `4294967301` as `5` padded to 64 columns.

**Cause.** `%I64u` is MSVC-only; glibc reads it as a 64-wide `%u` of a 32-bit argument (undefined behavior).

**Fix.** `src/muz/fp/datalog_parser.cpp:1291` (`FORK FIX #67`): `PRIu64` (also the unreachable "unknown symbol number" message).

**Test.** `TST(z3_4_7_1_wpa_warning_number)`, `src/test/z3_4_7_1_bugs.cpp:2519`
(registered at `src/test/main.cpp:283`).

---

## 68. `read_uint64` near `ULLONG_MAX`

**Upstream:** still present.

**Symptom.** `read_uint64` rejects `ULLONG_MAX` and every value from 18446744073709551610 up.

**Cause.** The overflow-guarded branch never advances past the digit.

**Fix.** `src/muz/base/dl_util.cpp:697` (`FORK FIX #68`): consume the digit (`s++`).

**Test.** `TST(z3_4_7_1_read_uint64_max)`, `src/test/z3_4_7_1_bugs.cpp:2542`
(registered at `src/test/main.cpp:284`).

---

## 69. WPA map line with an empty name

**Upstream:** still present.

**Symptom.** A map line such as `"1 "` gives an empty name, and the parser reads `rest_of_line[size()-1]` (undefined behavior; a release build happens to read a harmless byte, so the test is not red).

**Cause.** `parse_map_line` tests the name's last character without checking for an empty name.

**Fix.** `src/muz/fp/datalog_parser.cpp:1497` (`FORK FIX #69`): test emptiness first; the name is the empty symbol.

**Test.** `TST(z3_4_7_1_wpa_empty_map_name)`, `src/test/z3_4_7_1_bugs.cpp:2569`
(registered at `src/test/main.cpp:285`).

---

## 70. `goal::update` pushes the unflattened formula

**Upstream:** still present (5.4.2).

**Symptom.** `(declare-const a Bool)(assert (not (=> false a)))(apply nnf)` keeps `(not (=> false a))`, `(not a)` and `(and false (not a))`; the goal should be `false`. The same path made lia2card loop with `:compile_equality false` (fix 27 avoided that trigger).

**Cause.** Without proofs, `goal::update` flattens the new formula but calls `push_back(f, nullptr, d)` instead of `push_back(fr, ...)`.

**Fix.** `src/tactic/goal.cpp:292` (`FORK FIX #70`): push `fr`.

**Test.** `TST(z3_4_7_1_goal_update_false)`, `src/test/z3_4_7_1_bugs.cpp:2591`
(registered at `src/test/main.cpp:278`).

---

## 71. Model evaluation of mutually recursive `define-fun-rec`

**Upstream:** replaced by the recfun theory; no direct fix.

**Symptom.** `get-value` of `(even 0)` for mutually recursive Boolean `define-fun-rec`s never terminates.

**Cause.** The model evaluator expands both branches of the recursive `ite`/`and`/`or` unboundedly.

**Fix.** `FORK FIX #71`: `ast_manager` notes that `define-fun-rec` was used (`src/ast/ast.h:1488`, set in `cmd_context::insert_rec_fun`), and the model evaluator (`src/model/model_evaluator.cpp:515`) then lets `rewriter_tpl::constant_fold` stop an `and` at a false argument and an `or` at a true one (`src/ast/rewriter/rewriter_def.h`), as g3 does.

**Test.** `TST(z3_4_7_1_rec_fun_model_eval)`, `src/test/z3_4_7_1_bugs.cpp:2629`
(registered at `src/test/main.cpp:279`).

---

## 72. Nested parametric datatypes: `get_sort_size` use-after-free

**Upstream:** fixed in f1c3e1aa7 ("fix #2077") and e08abb321 ("fix #2504").

**Symptom.** Declaring `(Box T)` and `(Wrap U)` with a `(Box U)` field and asserting three distinct `(Wrap Bool)` values segfaults; the answer is `unsat`.

**Cause.** `sparam::subst` returns a size held in `S`, which `get_sort_size` then `dec_ref`s and frees before `mk_sort` uses it; the substitution was also keyed by the wrong parameters.

**Fix.** `src/ast/datatype_decl_plugin.cpp:587` and `:678` (`FORK FIX #72`) with the `m_refs` member in `datatype_decl_plugin.h`: key by `d.params()` and keep the sizes alive until `set_sort_size`, as upstream.

**Test.** `TST(z3_4_7_1_datatype_param_size)`, `src/test/z3_4_7_1_bugs.cpp:2657`
(registered at `src/test/main.cpp:280`).

---

## 73. `and_then`: a decided-sat subgoal does not decide the goal

**Upstream:** fixed (`result.reset()` in `and_then_tactical::operator()`).

**Symptom.** `(check-sat-using (then split-clause simplify))` answers `unknown` although a later subgoal is decided sat with a valid model; `apply` prints the stale undecided subgoals too.

**Cause.** The sat branch is appended to `result`, which still holds the undecided subgoals of earlier iterations.

**Fix.** `src/tactic/tactical.cpp:161` (`FORK FIX #73`): reset `result` before pushing the sat goal (g3 AUDIT.md tactic_top_01-F1).

**Test.** `TST(z3_4_7_1_and_then_sat_branch)`, `src/test/z3_4_7_1_bugs.cpp:2679`
(registered at `src/test/main.cpp:276`).

---

## 74. `repeat`: a decided-sat subgoal does not decide the goal

**Upstream:** still present in 5.4.2 (the null-model crash is gone with redesigned model converters).

**Symptom.** As fix 73 for `repeat`; with models on (the shell default), `(check-sat-using (repeat (or-else split-clause solve-eqs)))` over `(or (> x 0) (= y 3))` also segfaults in `extension_model_converter` because the branch's converter runs on a null model.

**Cause.** `repeat_tactical::operator()` keeps earlier undecided goals and applies the model converters to a null model.

**Fix.** `src/tactic/tactical.cpp:1004` (`FORK FIX #74`): reset `result`, and build the model as `and_then` does (`md = alloc(model, m); apply(mc2, md, 0); apply(mc1, md, i);`).

**Test.** `TST(z3_4_7_1_repeat_sat_branch)`, `src/test/z3_4_7_1_bugs.cpp:2703`
(registered at `src/test/main.cpp:277`).

---

## 75. opt `from_fmls`: stale objective term

**Upstream:** fixed in b33f5f879.

**Symptom.** A plain `assert-soft` group after a `maximize`/`minimize` is verified against the objective's term, maxres rejects every improving model, and the stale initial cost is reported (`(q (- 1))` instead of `(q (- 3))`).

**Cause.** `orig_term` is kept across loop iterations and `is_maxsat` does not reset it.

**Fix.** `src/opt/opt_context.cpp:816` and `:965` (`FORK FIX #75`): declare `orig_term` inside the loop and reset it in `is_maxsat`, as upstream.

**Test.** `TST(z3_4_7_1_opt_stale_orig_term)`, `src/test/z3_4_7_1_bugs.cpp:2730`
(registered at `src/test/main.cpp:273`).

---

## 76. qe array `solve_store` with coinciding indices

**Upstream:** still present (5.4.2).

**Symptom.** `(exists ((a (Array Int Int))) (= b (store (store a i v1) j v2)))` with `i = j`, `v1 = 0`, `v2 = 1` is unsat under `(then qe smt)` although plain check-sat says sat.

**Cause.** `array_plugin::solve_store` adds `select(store_T, i_k) = v_k` with the outer writes stored into `t`, which forces `v_k` to equal an outer value when the indices coincide.

**Fix.** `src/qe/qe_array_plugin.cpp:239` (`FORK FIX #76`): each write gives `(or (= i_k i_o) ... (= (select t i_k) v_k))` over the outer writes, as g3 does.

**Test.** `TST(z3_4_7_1_qe_nested_store)`, `src/test/z3_4_7_1_bugs.cpp:2768`
(registered at `src/test/main.cpp:274`).

---

## 77. Symbolic automaton product: epsilon closure of the initial states

**Upstream:** fixed in 606754c09 ("fix #2262") and a921b4ff4; the code was later removed (z3-4.15.3).

**Symptom.** `(str.in.re x (re.inter (re.union (str.to.re "") (str.to.re "a")) (re.union (str.to.re "") (str.to.re "b"))))` is unsat although `x = ""` satisfies it.

**Cause.** `mk_product` tests the initial states' finality without their epsilon closure.

**Fix.** `src/math/automata/symbolic_automata_def.h:382` and `:474` (`FORK FIX #77`): use the epsilon closure of the initial states, as upstream.

**Test.** `TST(z3_4_7_1_automaton_product_epsilon)`, `src/test/z3_4_7_1_bugs.cpp:2807`
(registered at `src/test/main.cpp:275`).

---

## 78. datalog `string_to_uint64` and blank strings

**Upstream:** still present (1a1d88dde7).

**Symptom.** An `int` domain fact such as `P("")` parses with an indeterminate numeral (a different value each run).

**Cause.** `string_to_uint64` treats `sscanf`'s EOF result for a blank string as success.

**Fix.** `src/muz/base/dl_util.cpp:670` (`FORK FIX #78`): require `converted == 1`; the parser then reports `Invalid integer: ""` (g3 AUDIT.md cmd_context_01-F9).

**Test.** `TST(z3_4_7_1_dl_string_to_uint64_blank)`, `src/test/z3_4_7_1_bugs.cpp:3290` (registered at `src/test/main.cpp:300`).

---

## 79. `elim-small-bv`: incomplete enumeration

**Upstream:** fixed in b6ee0b151, 8e033c1e7 and aa901c4e8.

**Symptom.** `(forall ((y (_ BitVec 2))) (not (= y #b11)))` becomes `true`, so an unsat goal answers sat; a goal quantified over Int segfaults; under `max_steps` a partial expansion is kept.

**Cause.** `rw_cfg::reduce_quantifier` calls `get_bv_size` before checking the sort, enumerates `j < bv_sz` values instead of `2^bv_sz`, and keeps a partial expansion.

**Fix.** `src/tactic/bv/elim_small_bv_tactic.cpp:143` (`FORK FIX #79`): as upstream (g3 had copied the unsoundness and is fixed too).

**Test.** `TST(z3_4_7_1_elim_small_bv)`, `src/test/z3_4_7_1_bugs.cpp:2794` (registered at `src/test/main.cpp:290`).

---

## 80. `sine-filter` crashes

**Upstream:** fcd1f2b3c fixes the missing entry; sine-filter was removed upstream (490dc66ec).

**Symptom.** A last formula with no uninterpreted root such as `(> x 0)` segfaults; an empty goal reads `g->form(UINT_MAX)`.

**Cause.** `filter_expressions` dereferences a missing `obj_map` entry and indexes the last formula of an empty goal.

**Fix.** `src/tactic/sine_filter.cpp:224` (`FORK FIX #80`).

**Test.** `TST(z3_4_7_1_sine_filter_component)`, `src/test/z3_4_7_1_bugs.cpp:2822` (registered at `src/test/main.cpp:291`).

---

## 81. `symmetry-reduce` on quantified goals

**Upstream:** fixed in 330b3cc8d.

**Symptom.** A quantified goal fails `VERIFY(colors1.find(a, v2))` in `merge_colors` (release: segfault).

**Cause.** The tactic does not reject quantified goals.

**Fix.** `src/tactic/tactic.h:162`, `src/tactic/tactic.cpp:255`, `src/tactic/core/symmetry_reduce_tactic.cpp:644` (`FORK FIX #81`): `fail_if_has_quantifiers`, as upstream.

**Test.** `TST(z3_4_7_1_symmetry_reduce_quantified)`, `src/test/z3_4_7_1_bugs.cpp:2845` (registered at `src/test/main.cpp:292`).

---

## 82. `bvarray2uf` on quantifiers

**Upstream:** fixed in 4bad2dd92.

**Symptom.** Any quantifier, even `(forall ((x Int)) (p x))`, reaches `NOT_IMPLEMENTED_YET()`.

**Cause.** `reduce_var`, `reduce_quantifier` and two `reduce_app` cases are unimplemented.

**Fix.** `src/tactic/bv/bvarray2uf_rewriter.cpp:217`, `:368`, `:401`, `:408` (`FORK FIX #82`): raise `default_exception("not handled by bvarray2uf")`, as upstream.

**Test.** `TST(z3_4_7_1_bvarray2uf_quantifier)`, `src/test/z3_4_7_1_bugs.cpp:2866` (registered at `src/test/main.cpp:293`).

---

## 83. `get-value` under `model_evaluator.max_steps`

**Upstream:** fixed in 335d672bf.

**Symptom.** `get-value` segfaults in `shared_occs` when evaluation exceeds `max_steps` (fix 52 does not cover this path).

**Cause.** `parse_get_value` calls `model::eval`, which swallows the exception and leaves a null value that `display` dereferences.

**Fix.** `src/parsers/smt2/smt2parser.cpp:34`, `:2597` (`FORK FIX #83`): evaluate with a `model_evaluator` so the exception becomes a command error.

**Test.** `TST(z3_4_7_1_get_value_max_steps)`, `src/test/z3_4_7_1_bugs.cpp:2885` (registered at `src/test/main.cpp:294`).

---

## 84. Parametric datatype instances after pop

**Upstream:** fixed in 19f655c69, 120467159, 2d358a9a5 and 984767509.

**Symptom.** After a pop, the next `(List Real)` created outside the popped scope fails with "invalid declaration, constant 'nil' ... already declared" and the declaration that used it is lost.

**Cause.** A pop resets the whole instance cache of the parametric datatype.

**Fix.** `src/cmd_context/pdecl.h`, `pdecl.cpp` and `cmd_context.cpp:1333`, `:1452`, `:2040` (`FORK FIX #84`): the notified-sort set with a push/pop trail, as upstream.

**Test.** `TST(z3_4_7_1_datatype_inst_after_pop)`, `src/test/z3_4_7_1_bugs.cpp:2917` (registered at `src/test/main.cpp:295`).

---

## 85. `define-sort` alias after pop

**Upstream:** still present (5.4.2).

**Symptom.** `(define-sort L () (List Int)) (push 1) (declare-const k L) (pop 1) (declare-const k L) (assert (= k (insert 1 nil)))` reports `unknown constant nil`.

**Cause.** The alias's psort instance cache survives the pop, which removed the constructors, and a cache hit never declares them again.

**Fix.** `src/cmd_context/pdecl.cpp:217`, `:320`, `:734` and `pdecl.h` (`FORK FIX #85`): `needs_renotify` re-instantiates the definition on such cache hits.

**Test.** `TST(z3_4_7_1_define_sort_alias_after_pop)`, `src/test/z3_4_7_1_bugs.cpp:2953` (registered at `src/test/main.cpp:296`).

---

## 86. WPA loader: unreadable map or relation file

**Upstream:** still present (5.4.2).

**Symptom.** `chmod 0 dir/V.map; z3 -dl dir` segfaults.

**Cause.** `parse_rel_file` and `parse_map_file` `fread` from a null `FILE`.

**Fix.** `src/muz/fp/datalog_parser.cpp:1383`, `:1526` (`FORK FIX #86`): report "ERROR: could not open file '%s'." as `parse_file` does.

**Test.** `TST(z3_4_7_1_wpa_unreadable_file)`, `src/test/z3_4_7_1_bugs.cpp:2979` (registered at `src/test/main.cpp:289`).

---

## 87. QBMC hyper-resolution positions

**Upstream:** fixed in 784b613b5.

**Symptom.** QBMC proofs record `(_ hyper-res 0 1)`; the proof checker rejects them.

**Cause.** `qlinear::get_model` resolves tail 0 of the derivation (premise 0) against the selected rule's head (premise 1), so the positions are (1,0).

**Fix.** `src/muz/bmc/dl_bmc_engine.cpp:323` (`FORK FIX #87`): (1,0), as upstream.

**Test.** `TST(z3_4_7_1_qbmc_hyper_res_positions)`, `src/test/z3_4_7_1_bugs.cpp:3010` (registered at `src/test/main.cpp:286`).

---

## 88. Nonlinear BMC proof reconstruction

**Upstream:** still present (5.4.2 reports "could not expand BMC rule").

**Symptom.** Nonlinear BMC queries such as `r(x) :- p(x), p(x+1)` segfault in proof reconstruction; without a proof trace `mk_hyper_resolve` crashes; the root proof concludes an internal level predicate.

**Cause.** `nonlinear::get_proof` Skolemizes every rule variable while `compile` binds head variables to the head arguments, so no tail rule is selected and a null rule is dereferenced.

**Fix.** `src/muz/bmc/dl_bmc_engine.cpp:521`, `:532`, `:570` (`FORK FIX #88`): bind head variables to the goal atom's arguments, assert a rule without a proof, and conclude the predicate at the model values.

**Test.** `TST(z3_4_7_1_bmc_nonlinear_head_vars)`, `src/test/z3_4_7_1_bugs.cpp:3040` (registered at `src/test/main.cpp:287`).

---

## 89. Boogie certificate debug output

**Upstream:** removed upstream in 73a24ca0a.

**Symptom.** `fixedpoint.print_boogie_certificate` prints `set proof`, `proof set`, and per step `get_subst` with the whole proof, on stdout.

**Cause.** Leftover debug output in `boogie_proof::set_proof` and `get_subst`.

**Fix.** `src/muz/base/dl_boogie_proof.cpp:124`, `:205` (`FORK FIX #89`): drop the debug lines, as upstream.

**Test.** `TST(z3_4_7_1_boogie_proof_debug_output)`, `src/test/z3_4_7_1_bugs.cpp:3068` (registered at `src/test/main.cpp:288`).

---

## 90. Sorting network comparison mode uninitialized

**Upstream:** fixed (`psort_nw(psort_expr& c): ctx(c), m_t(GE) {}`).

**Symptom.** sortmax reports optima below the true cost (for example 4 where the optimum is 5, or cost 0 with every soft false), and its clauses change between builds.

**Cause.** `psort_nw`'s constructor leaves `m_t` uninitialized, and sortmax calls `sorting()` without `ge`/`le`/`eq`.

**Fix.** `src/util/sorting_network.h:188` (`FORK FIX #90`): initialize `m_t` to `GE`, as upstream (g3 builds the GE network too).

**Test.** `TST(z3_4_7_1_sorting_uninit_cmp)`, `src/test/z3_4_7_1_bugs.cpp:3128` (registered at `src/test/main.cpp:285`).

---

## 91. Duality edge-to-clause map after dead-edge removal

**Upstream:** Duality was removed upstream.

**Symptom.** Every edge after a removed dead edge prints, and proves with, the previous rule's clause (extra substitutions such as `(= x1 0)` in certificates).

**Cause.** `dl_interface::query` maps `edges[i]` to `i`, but `FromClauses` -> `RemoveDeadNodes` has already deleted edges.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:312` (`FORK FIX #91`): map each edge by its `number - 1`.

**Test.** `TST(z3_4_7_1_duality_certificate_dead_edge)`, `src/test/z3_4_7_1_bugs.cpp:1771` (registered at `src/test/main.cpp:256`).

---

## 92. Duality `get_answer`

**Upstream:** Duality was removed upstream.

**Symptom.** `fixedpoint.print_answer` with the Duality engine segfaults.

**Cause.** `dl_interface::get_answer` is `SASSERT(false)` and returns null, which the caller dereferences.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:713` (`FORK FIX #92`): throw "duality does not support get_answer", as g3 reports.

**Test.** `TST(z3_4_7_1_duality_print_answer)`, `src/test/z3_4_7_1_bugs.cpp:1820` (registered at `src/test/main.cpp:257`).

---

## 93. Duality `get_proof` hyper-resolve steps

**Upstream:** Duality was removed upstream.

**Symptom.** `(get-proof)` after a Duality query prints steps the proof checker rejects: substitutions in declaration order instead of variable-index order, positions of the child instead of the resolved body literal, and `@Fail!0` instead of the clause head as the query step's conclusion.

**Cause.** `extract_proof` builds the steps from the edge order rather than from the original clause.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:764`, `:793` and following (`FORK FIX #93`): port g3's proof shape (variable order, one premise per resolved literal, the original head, background axioms resolving matching body literals).

**Test.** `TST(z3_4_7_1_duality_proof_hyper_resolve)`, `src/test/z3_4_7_1_bugs.cpp:1841` (registered at `src/test/main.cpp:258`).

---

## 94. Linear BMC hyper-resolution positions

**Upstream:** still present (5.4.2); 784b613b5 fixed only qlinear (fix 87).

**Symptom.** Linear BMC proofs record `(_ hyper-res 0 1)`; the proof checker cannot establish complementarity.

**Cause.** `linear::get_model` resolves tail 0 of the derivation (premise 0) against the selected rule's head (premise 1) but pushes `(0, 1)`.

**Fix.** `src/muz/bmc/dl_bmc_engine.cpp:1287` (`FORK FIX #94`): `(1, 0)`, as fix 87 does for qlinear.

**Test.** `TST(z3_4_7_1_bmc_linear_hyper_res_positions)`, `src/test/z3_4_7_1_bugs.cpp:3314` (registered at `src/test/main.cpp:301`).

---

## 95. qe-sat projection under a resource limit

**Upstream:** qe-sat was removed upstream.

**Symptom.** Under `rlimit`, `(check-sat-using qe-sat)` prints "Failed to verify" and then segfaults (projection with no model). g3 answers `unknown` with reason "max. resource limit exceeded".

**Cause.** `qsat::project` does `VERIFY(l_true == m_solvers[idx+1]->check())`; under a resource limit the check is `l_undef`.

**Fix.** `src/qe/qe_sat_tactic.cpp:640` (`FORK FIX #95`): throw `tactic_exception(get_cancel_msg())` when the check is undecided, as g3 does (pkg/qe/sat_project_level.go).

**Test.** `TST(z3_4_7_1_qe_sat_projection_canceled)`, `src/test/z3_4_7_1_bugs.cpp:3345` (registered at `src/test/main.cpp:302`).

---

## 96. Duality: a rejected candidate proposed forever

**Upstream:** Duality was removed upstream.

**Symptom.** For a rule `err <- forall z. p(z)` with `p = {1}` (unsat), z3 loops forever on "reject 2 err" (48k report lines in 10 s).

**Cause.** The induction check's edge solver uses `mbqi.id=itp`, so it ignores the non-interpolant quantifier and says sat; `CandidateFeasible` refutes the candidate, but no state changes, so the same candidate comes back.

**Fix.** `src/duality/duality_solver.cpp` (`FORK FIX #96`): count state changes (`ExpandNode` :702, `Update` :1593, `Extend` :1745; member at :296) and, in `SolveMainInt` (:1106), throw the existing `Solver::Incompleteness` when an already rejected candidate returns with no state change since. `dl_interface::query` reports it as "query failed: incompleteness" and `unknown`, as g3 does.

**Test.** `TST(z3_4_7_1_duality_rejected_candidate_loop)`, `src/test/z3_4_7_1_bugs.cpp:1880` (registered at `src/test/main.cpp:259`).

---

## 97. Simple joins: rule order from heap addresses

**Upstream:** still present (5.4.2 walks `m_rules_content`).

**Symptom.** The rules `join_planner::run` produces come out in an order that depends on heap addresses, so compiled register numbers and load order change between builds and runs (e.g. `datalog_query_sequence_three -v:2`: "load path into 0 / load edge into 1" swapped).

**Cause.** `m_rules_content` is keyed by rule pointers with `ptr_hash<rule>`, and `run` iterates it.

**Fix.** `src/muz/rel/dl_mk_simple_joins.cpp:708` (`FORK FIX #97`): walk `source.get_rule(ri)` in source order and look each rule up in `m_rules_content` (owner decision 2026-11-15: heap-address order is replaced by id order in both the fork and g3; g3 already walked source order).

**Test.** `TST(z3_4_7_1_simple_joins_rule_order)`, `src/test/z3_4_7_1_bugs.cpp:3397` (registered at `src/test/main.cpp:304`).

---

## 98. theory_str: pointer-ordered containers

**Upstream:** theory_str was removed upstream.

**Symptom.** z3str3 models, the "concats N" verbose lines and recursion depth depend on heap addresses (e.g. a model `c = "abcabcabcabcabcabcabc"` vs g3's).

**Cause.** Every `std::map<expr*,...>` and `std::set<expr*>` in theory_str is iterated in pointer order, including `varEqConcatMap[node].begin()`.

**Fix.** `src/smt/theory_str.h:44-66` (`FORK FIX #98`) adds id comparators (`str_expr_id_lt`, `str_expr_vector_id_lt`, `str_expr_pair_id_lt`) and aliases used by every such container in `theory_str.{h,cpp}`; comment at `theory_str.cpp:5750`. g3 orders the same containers by AST id.

**Test.** `TST(z3_4_7_1_str_container_id_order)`, `src/test/z3_4_7_1_bugs.cpp:3444` (registered at `src/test/main.cpp:305`).

---

## 99. Duality: unchecked quantified `sat`

**Upstream:** Duality was removed upstream.

**Symptom.** Duality answers `sat` with an invalid counterexample when a rule or background axiom has a quantifier that the edge solvers (configured with `mbqi.id=itp`, MBQI only for interpolant quantifiers) never checked; on a generated corpus with brute-force answers z3 gave 336 wrong `sat` answers.

**Cause.** Duality trusts any edge-solver `sat`; with `mbqi.id=itp` the non-interpolant quantifiers are not instantiated.

**Fix.** `src/muz/duality/duality_dl_interface.cpp:140` and the call in `query` at `:536` (`FORK FIX #99`): before answering `sat`, `validate_counterexample` instantiates each counterexample step's original clause with fresh constants, replaces positively occurring predicate atoms by the facts derived by earlier steps, and checks the whole with a full `smt_solver` (proofs off); anything but `l_true` throws `Solver::Incompleteness` ("query failed: incompleteness", `unknown`). Quantifier-free counterexamples are accepted unchanged. g3 does the same (pkg/duality/counterexample_validation.go).

**Test.** `TST(z3_4_7_1_duality_itp_unchecked_sat)`, `src/test/z3_4_7_1_bugs.cpp:1921` (registered at `src/test/main.cpp:260`).

---

## 100. iZ3: rewrites and eliminations on mixed formulas

**Upstream:** iZ3 was removed upstream.

**Symptom.** Duality queries fail with "Unsupported proof rule" (or later on skolemization) when a proof contains a `rewrite` (singleton `and`, flattened `and`, `(=> true a)`) or an and-elim / not-or-elim on a formula that mixes frames.

**Cause.** `asserted_formulas::propagate_values` substitutes constants for constants, directed by decl ids, so an equality from one frame is substituted into another frame's assertion. A rewrite of such a mixed formula belongs to no frame, and `translate_main` refuses it; elims of mixed conjunctions were localized into quantified interpolants.

**Fix.** `src/interp/iz3translate.cpp:1250-1416` and hooks at `:2006`, `:2019` (`FORK FIX #100`): `localize_rewrite` and `localize_elim` map the substituted constants back, take the step locally and return to the original terms by congruence or modus ponens; if the renamed step is still not local, the old behavior applies. g3 does the same (pkg/interp/translation_localize.go). Interpolants pass iz3check on the 81-script Duality corpus.

**Test.** `TST(z3_4_7_1_duality_mixed_rewrite)`, `src/test/z3_4_7_1_bugs.cpp:1956` (registered at `src/test/main.cpp:261`).

---

## 101. iZ3: twice-mixed rewrites

**Upstream:** iZ3 was removed upstream.

**Symptom.** After fix 100, a Duality query can still fail with "theory not supported by interpolation or bad proof" when a rewrite's congruence left side was itself mixed by an earlier propagate_values round, so renaming with the congruence's own constant equalities leaves the step mixed (e.g. the 128 corpus script; spacer answers sat).

**Cause.** `localize_rewrite` (fix 100) only uses the constant-to-constant equalities proved inside the step's own congruence.

**Fix.** `src/interp/iz3translate.cpp:1263` (`global_theta`), `collect_global_equivs` at `:1380`, `proof_busy` at `:1400`, the retry in `localize_rewrite` at `:1377`ff, and the root setup in `translate()` at `:2336`/`:2382` (`FORK FIX #101`): when the rewrite stays mixed, retry with the proof's other hypothesis-free constant equalities (depth first, first proof of each constant wins), skipping any equality whose proof is still being translated (`proof_busy`, a null `translate_main` memo entry) to avoid recursion cycles. g3 does the same (pkg/interp/translation_localize.go, `clauseBusy`/`termBusy`). Interpolants pass iz3check.

**Test.** `TST(z3_4_7_1_duality_second_round_rewrite)`, `src/test/z3_4_7_1_bugs.cpp:2025` (registered at `src/test/main.cpp:262`).

---

## 102. Datalog: product kinds for another signature

**Symptom.** Segfault in `product_relation_plugin::mk_empty(s, kind)` from `get_relation` or `mk_unary_singleton` registers of a product-kind predicate, e.g. `(declare-rel r (Bool Bool) interval_relation bound_relation)` with a fact; 12 crash probes.

**Cause.** `product_relation_plugin::can_handle_signature(s, kind)` returned `true` for every signature, so a kind made for one signature was used for another whose spec is missing.

**Fix.** `src/muz/rel/dl_product_relation.cpp:93` and `rel_spec_store::contains_kind` at `src/muz/rel/dl_relation_manager.h:695` (`FORK FIX #102`): a product kind is accepted only for the signature it was made for. g3 already behaved this way.

**Test.** `TST(z3_4_7_1_product_kind_spec)`, `src/test/z3_4_7_1_bugs.cpp:3635` (registered at `src/test/main.cpp:309`).

---

## 103. Datalog: sieve kinds in product relations

**Symptom.** Segfault, or "Failed to verify: can_do_inner_union(i,i)" followed by a crash, in `product_relation_plugin::aligned_union_fn` on a negated mixed Int/Bool predicate with an unbound Int column, under any default relation; 77 crash probes.

**Cause.** The sieve plugin refused its own kinds in `can_handle_signature(s, kind)`, so products lost their sieves (the empty union target and `convert_spec` built tables or full relations instead), and the `VERIFY` does not stop the null union call.

**Fix.** `src/muz/rel/dl_sieve_relation.cpp:204` (`can_handle_signature` with kind), `:266` (`mk_full` with kind), declarations in `dl_sieve_relation.h`, and `src/muz/rel/dl_product_relation.cpp:652`, which reports a missing diagonal union as an error before doing any work (`FORK FIX #103`). g3 does the same (pkg/muz/rel/dl_sieve_relation.go).

**Test.** `TST(z3_4_7_1_sieve_kinds)`, `src/test/z3_4_7_1_bugs.cpp:3657` (registered at `src/test/main.cpp:310`).

---

## 104. z3str3: regex membership terms after pop

**Upstream:** 077f2248c (fix #3756) pins every `set_up_axioms` term in `m_trail`; 4fc64ab57 re-internalizes a regex term that is no longer internalized; c1a0ce086 (#3818) resets theory_str state in `init_search_eh`; theory_str was removed in fcd3a70c9. The fix here is narrower and changes no AST order in g3's 1441-vector corpus.

**Symptom.** Under z3str3 with the default `str.regex_automata=true`, a `str.in.re` term asserted inside push/pop is read by the next check-sat after the pop: z3 segfaults in `smt::context::get_assignment` called from `theory_str::final_check_eh`, or (when only the reads are guarded) final check loops forever because `gen_len_val_options_for_free_var` still refuses value testing for the variable.

**Cause.** `instantiate_axiom_RegexIn` records the term by raw pointer in `regex_in_bool_map`, `regex_in_var_reg_str_map`, `regex_terms`, `regex_terms_by_string` and `axiomatized_terms`; neither `pop_scope_eh` nor the trail removes the entries, and in automata mode nothing keeps the term alive.

**Fix.** `src/smt/theory_str.cpp` (`FORK FIX #104`): `instantiate_axiom_RegexIn` pins the term in `m_trail` (`:1865`), and four readers skip a term that is not `b_internalized`: the final-check `regex_terms` loop (`:9938`), the `regex_terms_by_string` intersection loop after the length readers (`:10498`), the path-constraint completeness loop (`:10728`), and the `gen_len_val_options_for_free_var` regex gate (`:12202`). g3 does the same (pkg/smtcore theory_str regex stages and sequential options).

**Test.** `TST(z3_4_7_1_str_regex_pop)`, `src/test/z3_4_7_1_bugs.cpp:3695` (registered at `src/test/main.cpp:311`).

---

## 105. z3str3: legacy regex unroll tries map

**Upstream:** the legacy unroll solver and later theory_str were removed upstream.

**Symptom.** With `smt.str.regex_automata=false`, any regex membership that needs an unroll, e.g. `x in (re.+ (str.to.re "a"))` in a single check-sat, segfaults in `gen_unroll_conditional_options`.

**Cause.** `unroll_tries_map[var][unrolls]` uses `obj_map::operator[]`, which is a find that never inserts; no entry is ever added, so the release build dereferences null.

**Fix.** `src/smt/theory_str.cpp:11518` (`FORK FIX #105`): `gen_unroll_conditional_options` inserts an empty map for the variable when it is missing. g3 already had insert-on-miss semantics.

**Test.** `TST(z3_4_7_1_str_regex_unroll_tries)`, `src/test/z3_4_7_1_bugs.cpp:3723` (registered at `src/test/main.cpp:312`).

---

## 106. Datalog: check_relation operations on other plugins' relations

**Symptom.** "terminate called after throwing an instance of 'std::bad_cast'" from `check_relation_plugin::mk_join_fn` under `instr_join::perform`, when `datalog.check_relation` is set and a predicate the checked plugin cannot handle is joined with one it can (e.g. the p_E_chk_bound_relation / p_E_chk_interval_relation probes, or check_relation(doc) over a Bool table through the C API).

**Cause.** The binary `mk_*_fn` of `check_relation_plugin` assumed every argument was a `check_relation` (`dynamic_cast` to a reference).

**Fix.** `src/muz/rel/check_relation.cpp:157`-`:164` (`is_checked`) and its uses at `:220`, `:251`, `:524`, `:532`, `:717` (`FORK FIX #106`): the functor factories return `nullptr` unless every argument is a check_relation, so the relation manager falls back to the default functors. g3 already behaved this way.

**Test.** `TST(z3_4_7_1_check_relation_mixed_join)`, `src/test/z3_4_7_1_bugs.cpp:3740` (registered at `src/test/main.cpp:313`).

---

## 107. Datalog: table relations of signatures their table rejects

**Symptom.** Segfault in `sparse_table::column_layout::make_byte_aligned_end` from `table_relation_plugin::mk_empty` <- `check_relation_plugin::mk_full` <- `sieve_relation_plugin::full` <- `product_relation_plugin` join (`datalog.check_relation tr_sparse`, a Bool predicate and a negated (Int, Bool) predicate).

**Cause.** `table_relation_plugin` did not check `can_handle_signature` before asking its table plugin for a table.

**Fix.** `src/muz/rel/dl_table_relation.cpp:53` (`mk_empty`) and `:71` (`mk_full_relation`) (`FORK FIX #107`): use a table plugin that can handle the signature. g3 does the same (pkg/muz/rel table relation plugin).

**Test.** `TST(z3_4_7_1_table_relation_unhandled_signature)`, `src/test/z3_4_7_1_bugs.cpp:3777` (registered at `src/test/main.cpp:314`).

---

## 108. Datalog: table complement of several columns

**Symptom.** "Failed to verify: sig.first_functional() == 1" followed by a segfault, from `table_base::complement` <- `plugin_object::mk_full` <- `product_relation_plugin` `join_fn::get_full_tableish_relation` (doc favourite, a two-column udoc predicate joined with a Bool table into an unseeded head).

**Cause.** The complement loop only handles one non-functional column.

**Fix.** `src/muz/rel/dl_base.cpp:386` (`FORK FIX #108`): enumerate all non-functional columns. g3's `TableRelationPlugin.MkFull` follows z3's `mk_full` (an empty helper table, then its complement).

**Test.** `TST(z3_4_7_1_table_complement_columns)`, `src/test/z3_4_7_1_bugs.cpp:3801` (registered at `src/test/main.cpp:315`).

---

## 109. Datalog: linear inlining loses relation facts

**Symptom.** Wrong answer. Facts added with `Z3_fixedpoint_add_fact` to a predicate disappear when that predicate's rule is inlined into its readers: with `xform.inline_eager=false` and `xform.inline_linear_branch=true`, a query answers {false} instead of {true, false}.

**Cause.** `mk_rule_inliner` fills `m_preds_with_facts` (`count_pred_occurrences` -> `rel_context::collect_non_empty_predicates`) only from `plan_inlining`, which runs only with `xform.inline_eager`. `inline_linear` then treats every predicate as fact-free.

**Fix.** `src/muz/transforms/dl_mk_rule_inliner.cpp:860` (`FORK FIX #109`): `mk_rule_inliner::operator()` collects the non-empty predicates before `inline_linear` when eager inlining is off. g3 was already correct; its `TestAPIFixedpointInlinerLinearFacts` holds the fixed answer.

**Test.** `TST(z3_4_7_1_inliner_linear_facts)`, `src/test/z3_4_7_1_bugs.cpp:3852` (registered at `src/test/main.cpp:316`).

---

## 110. Datalog: explanations with an interpreted tail over table columns

**Symptom.** Segfault with `datalog.generate_explanations` on a rule with an interpreted tail over table and explanation columns, e.g. `node(y) :- path(x,y), x != y`.

**Cause.** `finite_product_relation`'s `filter_interpreted_fn` calls the inner relation filter functor without checking that the plugin returned one.

**Fix.** `src/muz/rel/dl_finite_product_relation.cpp:1368` and `:1427` (`FORK FIX #110`): throw "trying to perform unsupported filter_interpreted operation on a relation of kind fact_explanation" instead. g3 reports the same error.

**Test.** `TST(z3_4_7_1_fpr_filter_interpreted)`, `src/test/z3_4_7_1_bugs.cpp:3966` (registered at `src/test/main.cpp:317`).

---

## 111. Datalog: explanations with `default_relation doc`

**Symptom.** Segfault with `datalog.generate_explanations` and `datalog.default_relation=doc`.

**Cause.** `mk_explanations::transform_facts` calls the join and union functors of the finite-product/product relation without checking for null.

**Fix.** `src/muz/rel/dl_mk_explanations.cpp:856` and `:871` (`FORK FIX #111`): throw when the join or union functor is missing. g3 reports the same error.

**Test.** `TST(z3_4_7_1_explanations_doc_union)`, `src/test/z3_4_7_1_bugs.cpp:3978` (registered at `src/test/main.cpp:318`).

---

## 112. Datalog: rule transformer leaks rules when a plugin throws

**Symptom.** With `xform.karr=true`, `mk_karr_invariants` always throws (its inner datalog context's `check_rules` rejects the loop counter's Int column). The rule set being built, and the ASTs of its rules, stay allocated until `~ast_manager`: a leak, and AST deletion order that differs from a correct implementation.

**Cause.** `rule_transformer::operator()` calls each plugin with no cleanup of `new_rules` on an exception.

**Fix.** `src/muz/base/dl_rule_transformer.cpp:94` (`FORK FIX #112`): catch, deallocate `new_rules`, rethrow. g3 already frees them; its AST event digests for `fixedpoint_spacer_karr_infinite_sorts.smt2` were recorded from the fixed fork.

**Test.** `TST(z3_4_7_1_transformer_throw_leak)`, `src/test/z3_4_7_1_bugs.cpp:4029` (registered at `src/test/main.cpp:319`): the AST count must return to its value before the throw.

---

## 113. Datalog: udoc facts in bit-vector columns wider than 64 bits

**Symptom.** Wrong answer. Under `datalog.default_relation=doc`, `Z3_fixedpoint_add_fact` of 5 to `p((_ BitVec 160))` answers `#x0000000500000000000000050000000000000005`: the value repeats in every 64-bit lane.

**Cause.** `tbv_manager::set(tbv&, uint64, hi, lo)` tests bit `i` with `1ULL << i` for every bit of the column; for `i >= 64` the shift is undefined (x86 masks it to `i % 64`).

**Fix.** `src/muz/rel/tbv.cpp:140` (`FORK FIX #113`): bits at or above 64 are zero. g3 routes such predicates to udoc relations as `mk_empty_relation` does and sets the bits from the numeral of any width.

**Test.** `TST(z3_4_7_1_tbv_wide_uint64)`, `src/test/z3_4_7_1_bugs.cpp:4049` (registered at `src/test/main.cpp:320`).

---

## 114. Numerals: gcd of an inline value and INT_MAX stored in a cell

**Symptom.** Segfault (null `m_ptr` dereference) in `mpz_manager::gcd`, e.g. from `mpq_manager::set(q, 0x7fffffff, 0x7fffffff)` normalizing the fraction.

**Cause.** `set(mpz&, uint64)` keeps values `>= INT_MAX` in a cell, so INT_MAX can be a cell while an equal value is inline. The Lehmer gcd entry swaps its operands only by magnitude; with a small `a1` and a big `b1` of equal value it does not swap and then reads `size(a1)` through the inline operand's null cell pointer. Upstream z3 has the same code.

**Fix.** `src/util/mpz.cpp:735`-`:739` (`FORK FIX #114`): the entry swap also swaps when `a1` is inline and `b1` is a cell. g3's `mpz_gcd.go` has the same fix (`TestIntegerGCDIntMaxCell`).

**Test.** `TST(z3_4_7_1_gcd_int_max_cell)`, `src/test/z3_4_7_1_bugs.cpp:4081` (registered at `src/test/main.cpp:321`).

---

## 115. Datalog: sparse negated join removes the wrong rows

**Symptom.** `filter_by_negated_join` (e.g. lazy_table filter by negation of a lazy join) leaves rows and indexer disagreeing: t = {1, 2, 3, 4} minus {1, 4} iterates {2, 4}.

**Cause.** `remove_offset` moves the last row into the removed slot, but the offsets to remove were collected and removed in increasing order.

**Fix.** `src/muz/rel/dl_sparse_table.cpp:1327` (`FORK FIX #115`): remove the largest offsets first, as `negation_filter_fn` does.

**Test.** `TST(z3_4_7_1_sparse_negated_join_order)`, `src/test/z3_4_7_1_bugs.cpp:4110` (registered at `src/test/main.cpp:322`).

---

## 116. Datalog: sparse contains_fact with functional columns

**Symptom.** An index table holding (0) reports that it contains (1); (1, 2, 3) with two functional columns makes (1, 0, 3) contained.

**Cause.** The cells are compared from `functional_columns()` on, but the functional columns are the last ones.

**Fix.** `src/muz/rel/dl_sparse_table.cpp:539` (`FORK FIX #116`): compare the columns before `sz - func_col_cnt`.

**Test.** `TST(z3_4_7_1_sparse_contains_fact_functional)`, `src/test/z3_4_7_1_bugs.cpp:4167` (registered at `src/test/main.cpp:323`).

---

## 117. Datalog: sparse negated join column numbers of the second source

**Symptom.** t = {(1,1), (1,5)} minus the join {(1,1,5)} on both columns removes (1,5) instead of (1,1).

**Cause.** A column of src2 keeps its number in the joined signature when used as src2's own column.

**Fix.** `src/muz/rel/dl_sparse_table.cpp:1311` (`FORK FIX #117`): subtract src1's column count.

**Test.** `TST(z3_4_7_1_sparse_negated_join_second_columns)`, `src/test/z3_4_7_1_bugs.cpp:4210` (registered at `src/test/main.cpp:324`).

---

## 118. Datalog: table intersection filter is a negation filter

**Symptom.** product_relation's aligned union intersects table relations by removing the rows the other has (the m_E_* probes run `filter_by_negation` where the intersection belongs). Outputs happen to agree; rlimit counts and table recycling differ.

**Cause.** `relation_manager::mk_filter_by_intersection_fn(table)` asks the table plugins for `mk_filter_by_negation_fn`.

**Fix.** `src/muz/rel/dl_relation_manager.cpp:1487` (`FORK FIX #118`): ask for `mk_filter_by_intersection_fn`; no table plugin has one, so the default join and union are used, as g3 does. The m_E_* probes' rlimit count becomes 93 (was 73).

**Test.** `TST(z3_4_7_1_table_intersection_filter)`, `src/test/z3_4_7_1_bugs.cpp:4262` (registered at `src/test/main.cpp:325`).

---

## 119. Datalog: sparse negated join misses an all-zero first key

**Symptom.** t = {0} minus {0} joined with {0} keeps its row.

**Cause.** The key buffers start as zeros and are looked up only when `update_key` reports a change, so a first key of all zeros is never looked up.

**Fix.** `src/muz/rel/dl_sparse_table.cpp:1354` (`FORK FIX #119`): look the first keys up.

**Test.** `TST(z3_4_7_1_sparse_negated_join_zero_key)`, `src/test/z3_4_7_1_bugs.cpp:4308` (registered at `src/test/main.cpp:326`).

---

## 120. Datalog: bit-blasting drops the facts of the blasted predicates

**Symptom.** With `fixedpoint.xform.bit_blast=true`, facts added with `Z3_fixedpoint_add_fact` (or left in a relation by an earlier query) are ignored: with the fact p(1) and the rule q(x) :- p(x), the query exists x. q(x) answers unsat (sat without bit_blast). Upstream master's `dl_mk_bit_blast.cpp` still rewrites the rules only.

**Cause.** `mk_bit_blast` replaces every predicate whose arguments are all bit-vectors by a fresh predicate over their bits and rewrites the rules, but the original predicate's relation in `rel_context`, which holds the facts, is never consulted by the new rules.

**Fix.** `src/muz/transforms/dl_mk_bit_blast.cpp:311` (`FORK FIX #120`): after building the model converter, call `rel_context_base::add_bit_blasted_facts(old, new)` for each blasted predicate. `src/muz/rel/rel_context.cpp:575` copies the rows of a table relation into the new predicate, each bit-vector column expanded into its bits, low bit first. The virtual is declared in `src/muz/base/dl_context.h:133` (empty default for engines without relations) and overridden in `src/muz/rel/rel_context.h:104`. g3 already copied the facts.

**Test.** `TST(z3_4_7_1_bit_blast_add_fact)`, `src/test/z3_4_7_1_bugs.cpp:4349` (registered at `src/test/main.cpp:326`).

---

## Known 4.7.1 bugs that are not fixed here

The g3 fork documents these only in comments; no code changed for them.
They are listed so their behaviour isn't mistaken for a regression.

None at present: every bug that was listed here is now fixed above.

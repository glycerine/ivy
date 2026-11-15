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

**Test.** `TST(z3_4_7_1_dom_simplify_root)`, `src/test/z3_4_7_1_bugs.cpp:201`
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

**Test.** `TST(z3_4_7_1_elim_term_ite_pop)`, `src/test/z3_4_7_1_bugs.cpp:215`
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

**Test.** `TST(z3_4_7_1_macro_finder_off)`, `src/test/z3_4_7_1_bugs.cpp:248`
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
`src/test/z3_4_7_1_bugs.cpp:283`, and `TST(z3_4_7_1_pb_preprocess_resolve)`,
`src/test/z3_4_7_1_bugs.cpp:299` (registered at `src/test/main.cpp:215-216`).

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

**Test.** `TST(z3_4_7_1_dt_recognizer)`, `src/test/z3_4_7_1_bugs.cpp:338`
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
`src/test/z3_4_7_1_bugs.cpp:388` (registered at `src/test/main.cpp:218`). A
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
`src/test/z3_4_7_1_bugs.cpp:415` (registered at `src/test/main.cpp:219`).

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

**Test.** `TST(z3_4_7_1_pdr_nullary_cache)`, `src/test/z3_4_7_1_bugs.cpp:461`
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
`src/test/z3_4_7_1_bugs.cpp:491` (registered at `src/test/main.cpp:229`).

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
`src/test/z3_4_7_1_bugs.cpp:514` (registered at `src/test/main.cpp:230`).

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

**Test.** `TST(z3_4_7_1_qbmc_eval_level)`, `src/test/z3_4_7_1_bugs.cpp:543`
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

**Test.** `TST(z3_4_7_1_ddnf_extract_width)`, `src/test/z3_4_7_1_bugs.cpp:579`
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

**Test.** `TST(z3_4_7_1_min_cut_shared_node)`, `src/test/z3_4_7_1_bugs.cpp:632`
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

**Test.** `TST(z3_4_7_1_qi_weakening_proof)`, `src/test/z3_4_7_1_bugs.cpp:665`
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

**Test.** `TST(z3_4_7_1_pb2bv_numeral_lhs)`, `src/test/z3_4_7_1_bugs.cpp:736`
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

**Test.** `TST(z3_4_7_1_eq2bv_stale_cache)`, `src/test/z3_4_7_1_bugs.cpp:758`
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

**Test.** `TST(z3_4_7_1_purify_trig_unsafe)`, `src/test/z3_4_7_1_bugs.cpp:779`
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
`src/test/z3_4_7_1_bugs.cpp:805` (registered at `src/test/main.cpp:223`).
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
`src/test/z3_4_7_1_bugs.cpp:838` (registered at `src/test/main.cpp:224`).

---

## 25. `distribute-forall` crashes with proofs on

**Upstream:** d530d1314 ("fix #3276").

**Symptom.** `(apply distribute-forall)` segfaults with `:produce-proofs true`.

**Cause.** The rewrite is applied without building its proof.

**Fix.** `src/tactic/core/distribute_forall_tactic.cpp:57` and `:81`
(`FORK FIX #25`): prove the split with `mk_push_quant`.

**Test.** `TST(z3_4_7_1_distribute_forall_proof)`,
`src/test/z3_4_7_1_bugs.cpp:877` (registered at `src/test/main.cpp:225`).

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

**Test.** `TST(z3_4_7_1_pb2bv_fractional)`, `src/test/z3_4_7_1_bugs.cpp:896`
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

**Test.** `TST(z3_4_7_1_lia2card_empty_eq)`, `src/test/z3_4_7_1_bugs.cpp:924`
(registered at `src/test/main.cpp:227`), with a 20-second watchdog.

---

## 28. nlsat: negative sign stored as unsigned in infinity projection

**Upstream:** 363d7aad2.

**Symptom.** On a quantified nonlinear real problem with `nlqsat`, z3 never terminates where the answer is `unsat`.

**Cause.** `project_plus_infinity` and `project_minus_infinity` keep the sign of a coefficient in an `unsigned`, so a negative sign is treated as positive.

**Fix.** `src/nlsat/nlsat_explain.cpp:1673` and `:1689` (`FORK FIX #28`): the sign is an `int`.

**Test.** `TST(z3_4_7_1_nlsat_signed_infinity)`, `src/test/z3_4_7_1_bugs.cpp:990`
(registered at `src/test/main.cpp:235`).

---

## 29. nlsat: atom sign wrong when the factor-sign product overflows

**Upstream:** f79dccccf and 6071797ba (#2238).

**Symptom.** With large factor values (for example three factors of 2^22) the nlsat search diverges: the atom gets the wrong truth value.

**Cause.** `eval_ineq` multiplies the raw (unnormalized) signs of the factors in 32-bit `int`, which overflows.

**Fix.** `src/nlsat/nlsat_evaluator.cpp:387` (`FORK FIX #29`): normalize each sign to -1, 0 or 1 before multiplying.

**Test.** `TST(z3_4_7_1_nlsat_eval_sign_overflow)`, `src/test/z3_4_7_1_bugs.cpp:1031`
(registered at `src/test/main.cpp:236`).

---

## 30. nlsat: `solve_eq` keeps a stale power table

**Upstream:** 19e2f8c9d (#1694).

**Symptom.** When a later polynomial has a higher degree than an earlier one, `solve_eq` computes a wrong value (seen at unit level; no SMT-LIB input diverged).

**Cause.** The vector of powers of the leading coefficient is never reset between polynomials.

**Fix.** `src/nlsat/nlsat_explain.cpp:1729` and `:1747` (`FORK FIX #30`): each polynomial indexes its own power tables (upstream's As/Bs).

**Test.** `TST(z3_4_7_1_nlsat_solve_eq_powers)`, `src/test/z3_4_7_1_bugs.cpp:1068`
(registered at `src/test/main.cpp:237`).

---

## 31. nlsat: `s == -1` instead of `s < 0` in normalize/simplify

**Upstream:** 0f20175bd (#2556).

**Symptom.** The explain step can produce an invalid lemma (shown at unit level); on the audit input only the trace changes.

**Cause.** `normalize` and `simplify` test the sign with `s == -1`, but the sign can be any negative number.

**Fix.** `src/nlsat/nlsat_explain.cpp:432` and `:1138` (`FORK FIX #31`): test `s < 0`.

**Test.** `TST(z3_4_7_1_nlsat_simplify_core_sign)`, `src/test/z3_4_7_1_bugs.cpp:1108`
(registered at `src/test/main.cpp:238`).

---

## 32. nlsat `inline_vars`: vanished factors and constant atoms

**Upstream:** 29f3f6a7a (#3144) and 99784a92e (#3225).

**Symptom.** With `nlsat.inline_vars=true`, z3 segfaults when a factor vanishes after substitution, and gives an invalid model when an atom becomes constant.

**Cause.** The substitution does not give a zero product its value, does not fold constant atoms, and leaves an empty clause instead of `false`.

**Fix.** `src/nlsat/nlsat_solver.cpp:2373`, `:2385`, `:2409` and `:2451` (`FORK FIX #32`).

**Test.** `TST(z3_4_7_1_nlsat_inline_vars)`, `src/test/z3_4_7_1_bugs.cpp:1170`
(registered at `src/test/main.cpp:239`).

---

## 33. nlsat: reorder after `inline_vars` gives an invalid model

**Upstream:** 983a55232 (#3167).

**Symptom.** With inlined variables and variable reordering, the model assigns values to the wrong variables.

**Cause.** `fix_patch` runs after `reorder` and indexes the reordered assignment with pre-reorder variables.

**Fix.** `src/nlsat/nlsat_solver.cpp:2019` (`FORK FIX #33`): `can_reorder()` returns false while patches exist.

**Test.** `TST(z3_4_7_1_nlsat_inline_reorder)`, `src/test/z3_4_7_1_bugs.cpp:1193`
(registered at `src/test/main.cpp:240`).

---

## 34. `qe-sat`: `project_real` chooses the bound family from raw counts

**Upstream:** none; `qe-sat`, the only caller, was removed upstream in 8717c7d7f, so the code is unreachable there.

**Symptom.** `(check-sat-using (using-params qe-sat :projection_mode false))` never terminates on inputs with strict bounds.

**Cause.** `project_real` compares raw lower and upper bound counts, while `assign` uses `get_bound_sizes`, which doubles non-strict bounds for reals, so the two disagree on the branch.

**Fix.** `src/qe/qe_arith_plugin.cpp:1931` (`FORK FIX #34`): choose the bound family with `get_bound_sizes`.

**Test.** `TST(z3_4_7_1_qe_project_real_direction)`, `src/test/z3_4_7_1_bugs.cpp:1225`
(registered at `src/test/main.cpp:241`).

---

## 35. `qe`: `mod` by 0 or by a negative constant

**Upstream:** still present (negative cases).

**Symptom.** `(apply qe)` returns `false` for `(exists ((x Int)) (and (= (mod (+ x a) 0) 0) (< x b)))` and for negative moduli, which is unsound (x=3, a=0, b=10 satisfies the negative-modulus example).

**Cause.** `is_divides` and the nested-mod scan accept a zero or negative divisor, so `div_z` works with a zero or negative lcm.

**Fix.** `src/qe/qe_arith_plugin.cpp:45` and `:1329` (`FORK FIX #35`): a divisibility atom needs a positive modulus; others are left alone, so the quantifier is kept.

**Test.** `TST(z3_4_7_1_qe_mod_divisor)`, `src/test/z3_4_7_1_bugs.cpp:1248`
(registered at `src/test/main.cpp:242`).

---

## 36. `qe`: datatype substitution drops the recognizer

**Upstream:** still present.

**Symptom.** `(apply qe)` turns `(exists ((x Lst)) (and ((_ is nil) x) (or (= x y) (= x z))))` into `true`; with `is-cons y` and `is-cons z` added, `(then qe smt)` says `sat` on an unsat problem.

**Cause.** `datatype_plugin::subst_rec` substitutes `x := t` without keeping the recognizer constraint `r(t)`.

**Fix.** `src/qe/qe_datatype_plugin.cpp:725` (`FORK FIX #36`): conjoin `r(t)`.

**Test.** `TST(z3_4_7_1_qe_dt_eq_recognizer)`, `src/test/z3_4_7_1_bugs.cpp:1268`
(registered at `src/test/main.cpp:243`).

---

## 37. qe-light: Fourier-Motzkin keeps a stale Boolean-variable map

**Upstream:** still present (`qe_lite_tactic.cpp` `fm::init`).

**Symptom.** With two quantified formulas in one goal, qe-light maps a Boolean literal to a stale variable: it prints `(or b c)` where `(or b a)` is correct.

**Cause.** `fm::init` resets `m_expr2var` twice and `m_expr2bvar` never.

**Fix.** `src/qe/qe_lite.cpp:1544` (`FORK FIX #37`): reset `m_expr2bvar`.

**Test.** `TST(z3_4_7_1_qe_lite_fm_bvar)`, `src/test/z3_4_7_1_bugs.cpp:1287`
(registered at `src/test/main.cpp:244`).

---

## 38. maxres upper-bound block: `pb_util::mk_lt` aliases its coefficients

**Upstream:** fixed upstream by rewriting `mk_lt` as `not(mk_ge)` (commit not identified).

**Symptom.** With `opt.maxres.add_upper_bound_block=true`, maxres stops at the initial model's cost (19 where the optimum is 3).

**Cause.** `mk_lt` normalizes into `m_coeffs` and passes that vector to `mk_ge`, whose `normalize` starts with `m_coeffs.reset()`, so the blocker becomes a `pbge` with all-zero coefficients (`false`).

**Fix.** `src/ast/pb_decl_plugin.cpp:177` (`FORK FIX #38`): pass copies of the coefficients.

**Test.** `TST(z3_4_7_1_pb_mk_lt_coeffs)`, `src/test/z3_4_7_1_bugs.cpp:1309`
(registered at `src/test/main.cpp:245`).

---

## 39. OPB parser rejects `max:` objectives

**Upstream:** still present.

**Symptom.** An OPB file with `max: +2 x1 +3 x2 ;` fails with `unexpected char: a` and exit status 3.

**Cause.** `opb::parse` calls `parse_token("min:")` first, and `parse_token` consumes the matching `m`.

**Fix.** `src/opt/opt_parse.cpp:292` (`FORK FIX #39`): consume the shared `m` once, then test `in:` / `ax:`. The meaning kept for `max:` (minimize the negated sum) is unchanged and itself questionable.

**Test.** `TST(z3_4_7_1_opb_max_objective)`, `src/test/z3_4_7_1_bugs.cpp:1329`
(registered at `src/test/main.cpp:246`).

---

## 40. Probe `+` and `*` use the first argument twice

**Upstream:** still present.

**Symptom.** `(+ 1 2)` evaluates to 4 and `(* 2 3)` to 12 in probe expressions.

**Cause.** `MK_NARY_PROBE` starts the fold at child 1 after initializing the result with child 1.

**Fix.** `src/cmd_context/tactic_cmds.cpp:763` (`FORK FIX #40`): fold from child 2.

**Test.** `TST(z3_4_7_1_probe_nary_fold)`, `src/test/z3_4_7_1_bugs.cpp:1347`
(registered at `src/test/main.cpp:247`).

---

## 41. z3 cannot replay its own pseudo-Boolean logs

**Upstream:** fed977b49 (#1782).

**Symptom.** Replaying a log that calls `Z3_mk_pble`, `Z3_mk_pbge` or `Z3_mk_pbeq` always fails.

**Cause.** The generated logger writes `int` array arguments as unsigned values with `Au`, but the replayer reads them back as a signed array.

**Fix.** `scripts/update_api.py:958` (`FORK FIX #41`, generates `api_log_macros.cpp`), `src/api/z3_logger.h:45` and `src/api/z3_replayer.cpp:500` and `:509`: log and replay `int` arrays as `int` arrays.

**Test.** `TST(z3_4_7_1_pb_log_replay)`, `src/test/z3_4_7_1_bugs.cpp:1363`
(registered at `src/test/main.cpp:248`).

---

## 42. `(get-user-tactics)` segfaults

**Upstream:** 202ed79a2 (#5445).

**Symptom.** Any `(get-user-tactics)` after `declare-tactic` segfaults.

**Cause.** `sexpr::display` falls through after displaying an atom and reads it as a composite.

**Fix.** `src/util/sexpr.cpp:188` (`FORK FIX #42`): return after `display_atom`.

**Test.** `TST(z3_4_7_1_sexpr_display_atom)`, `src/test/z3_4_7_1_bugs.cpp:1404`
(registered at `src/test/main.cpp:249`).

---

## Known 4.7.1 bugs that are not fixed here

The g3 fork documents these only in comments; no code changed for them.
They are listed so their behaviour isn't mistaken for a regression.

* **`opt/opt_context.cpp` `from_fmls`.** `orig_term` is kept across loop
  iterations, so a plain `assert-soft` group that follows a
  `maximize`/`minimize` is verified against the wrong term, and maxres
  reports the stale initial cost. Upstream fixed this; z3 4.17.0 prints the
  optimum.
* **`ast/rewriter/rewriter_def.h`.** Model evaluation of mutually recursive
  `define-fun-rec` predicates does not terminate (for example, `get-value`
  of `(even 0)`). Upstream replaced this with its recfun theory.
* **`qe/qe_array_plugin.cpp` `solve_store`.** It rewrites
  `store(store(A,j,u),i,v) = t` to `store(t,i,v)[j] = u`, which is wrong
  when `i = j`. This is still present in upstream 4.17.0.
* **`math/automata/symbolic_automata_def.h` `mk_product`.** It ignores the
  epsilon closure of final states.
* **`tactic/goal.cpp` `goal::update`.** It flattens the new formula but
  appends the unflattened one (`f` instead of `fr`), so a goal can fail to
  become inconsistent. Fix 27 avoids the one known trigger; changing
  `goal::update` itself would affect every tactic. Still present upstream.
* **`muz/bmc/dl_bmc_engine.cpp` QBMC hyper-resolution positions.** Upstream
  784b613b5 also changes them from (0,1) to (1,0). Fix 16 does not port that
  part, because it would change every printed QBMC proof.
* **Nonlinear BMC with a quantified tail** segfaults (seen in g3 work input
  `bmc4.smt2`, second query); not analysed.
* **Other crashes and assertion failures that were not analysed:**
  * the `sine-filter` and `elim-small-bv` tactics;
  * a `symmetry-reduce` assertion;
  * `bvarray2uf` NOT IMPLEMENTED;
  * `get-value` under `model_evaluator.max_steps`;
  * `define-sort` after `pop`;
  * qlinear `unify_rules`.

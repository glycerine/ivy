# Bug fixes applied to this vendored z3 4.7.1

This tree is z3 4.7.1 (`dot.git.z3_v4.7.1_425291ee2`). The fixes below were
found while porting z3 to Go (the g3 project, github.com/glycerine/g3). That
work compares g3 against a z3 4.7.1 fork line by line. The same fixes are in
the g3 fork's `src/` tree:

* g3 commit c3c0016bd: relevancy and pareto backports
* g3 commit d91cef3ad: qsat/qe and theory_pb fixes
* g3 commit 0c41ffbae: bvsmul_noovfl/noudfl folding backport

They are copied here unchanged.

| # | Symptom in 4.7.1 | Fix | Files |
|---|---|---|---|
| 1 | Wrong models after `smt.relevancy` is raised | backport of upstream 5dfe4a4b4 | `smt/smt_context.{h,cpp}`, `smt/smt_internalizer.cpp`, `smt/smt_relevancy.cpp` (comment), `smt/params/smt_params.cpp` (comment) |
| 2 | Pareto optimization never terminates | backport of upstream e18702330 (#1699) | `opt/opt_context.cpp` |
| 3 | `qe_rec` segfaults on formulas with free constants and can return unsound results | backport of upstream 1ce0d7512 / 8ad939a10 (#2974, #2990) | `qe/qsat.cpp`, `qe/qe.cpp` |
| 4 | `free(): invalid pointer` abort in theory_pb with proofs on | local fix (upstream rewrote this code in 4.8.5) | `smt/theory_pb.cpp` |
| 5 | `bvsmul_noovfl`/`bvsmul_noudfl` fold constant operands to the wrong value | backport of upstream (z3 4.17.0) `mk_bvsmul_no_overflow` | `ast/rewriter/bv_rewriter.{h,cpp}` |

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
* **Other crashes and assertion failures that were not analysed:**
  * the `sine-filter` and `elim-small-bv` tactics;
  * a `symmetry-reduce` assertion;
  * `bvarray2uf` NOT IMPLEMENTED;
  * `get-value` under `model_evaluator.max_steps`;
  * `define-sort` after `pop`;
  * qlinear `unify_rules`.

/*++
Module Name:

    fork_bugfixes.cpp

Abstract:

    Regression tests for the bug fixes applied to this z3 4.7.1 tree
    (see ../../FORK_BUGFIXES.md). Each test runs the SMT-LIB2 reproduction
    script through the C API and checks the output. Before the fixes:

      relevancy  - models violating the assertions (p false although p is asserted)
      pareto     - infinite loop in gia_pareto
      qe_rec     - segmentation fault (free Boolean constant)
      qe_rec_lra - unsound result (exists ... false)
      pb_proof   - free(): invalid pointer at context destruction

    Run with:  test-z3 fork_bugfixes

    The same scripts and expected outputs are in
    ../../fork_bugfixes/test_vectors and can be checked against the z3
    binary with ../../fork_bugfixes/run_tests.sh.
--*/

#include "api/z3.h"
#include "util/debug.h"
#include <iostream>
#include <string>

namespace {

    std::string run_script(char const * script) {
        Z3_config cfg = Z3_mk_config();
        Z3_context ctx = Z3_mk_context(cfg);
        Z3_del_config(cfg);
        std::string out = Z3_eval_smtlib2_string(ctx, script);
        Z3_del_context(ctx);
        // set-option on module parameters (smt.relevancy, opt.priority, ...)
        // changes the global parameters; do not leak them into the next test.
        Z3_global_param_reset_all();
        return out;
    }

    void check(char const * name, char const * script, char const * expected) {
        std::string out = run_script(script);
        if (out != expected) {
            std::cerr << "fork_bugfixes: " << name << " FAILED\n--- expected:\n"
                      << expected << "--- got:\n" << out << "\n";
        }
        ENSURE(out == expected);
        std::cout << "fork_bugfixes: " << name << " ok\n";
    }

    bool contains(std::string const & s, char const * sub) {
        return s.find(sub) != std::string::npos;
    }

    // Upstream 5dfe4a4b4: raising smt.relevancy after assertions were
    // internalized with relevancy 0 must not drop them from the model.
    void test_relevancy_raised_after_setup() {
        check("relevancy_update_scopes",
              "(set-option :smt.relevancy 0)\n"
              "(set-option :produce-models true)\n"
              "(declare-const p Bool)\n"
              "(declare-const a Bool)\n"
              "(declare-const b Bool)\n"
              "(assert p)\n"
              "(assert (or a b))\n"
              "(check-sat)\n"
              "(get-value (p (or a b)))\n"
              "(push 1)\n"
              "(set-option :smt.relevancy 2)\n"
              "(check-sat)\n"
              "(get-value (p (or a b)))\n"
              "(pop 1)\n"
              "(check-sat)\n"
              "(get-value (p (or a b)))\n"
              "(push 1)\n"
              "(assert (not p))\n"
              "(check-sat)\n"
              "(pop 1)\n"
              "(check-sat)\n"
              "(get-value (p (or a b)))\n",
              "sat\n((p true)\n ((or a b) true))\n"
              "sat\n((p true)\n ((or a b) true))\n"
              "sat\n((p true)\n ((or a b) true))\n"
              "unsat\n"
              "sat\n((p true)\n ((or a b) true))\n");
    }

    // Upstream e18702330 (#1699): pareto objectives are evaluated with model
    // completion. Without it this loops forever.
    // (Z3_eval_smtlib2_string has no maximize/get-objectives commands, so
    // this one uses the optimize API; the SMT-LIB2 form is
    // fork_bugfixes/test_vectors/pareto_model_completion.smt2.)
    void test_pareto_model_completion() {
        Z3_config cfg = Z3_mk_config();
        Z3_context ctx = Z3_mk_context(cfg);
        Z3_del_config(cfg);
        Z3_optimize opt = Z3_mk_optimize(ctx);
        Z3_optimize_inc_ref(ctx, opt);
        Z3_params p = Z3_mk_params(ctx);
        Z3_params_inc_ref(ctx, p);
        Z3_params_set_symbol(ctx, p, Z3_mk_string_symbol(ctx, "priority"),
                             Z3_mk_string_symbol(ctx, "pareto"));
        Z3_params_set_symbol(ctx, p, Z3_mk_string_symbol(ctx, "optsmt_engine"),
                             Z3_mk_string_symbol(ctx, "basic"));
        Z3_optimize_set_params(ctx, opt, p);
        Z3_params_dec_ref(ctx, p);
        Z3_optimize_from_string(ctx, opt,
              "(declare-const x Int)\n"
              "(declare-const y Int)\n"
              "(assert (>= x 0)) (assert (<= x 1)) (assert (>= y 0)) (assert (<= y 1))\n"
              "(assert (= (+ x y) 1))\n"
              "(maximize x)\n"
              "(maximize y)\n");
        // Each check returns the next point of the pareto front; the
        // front is (1,0) and (0,1), then the search is exhausted.
        std::string front;
        unsigned checks = 0;
        for (; checks < 10; ++checks) {
            Z3_lbool r = Z3_optimize_check(ctx, opt);
            if (r != Z3_L_TRUE) {
                ENSURE(r == Z3_L_FALSE);
                break;
            }
            front += "(";
            front += Z3_ast_to_string(ctx, Z3_optimize_get_upper(ctx, opt, 0));
            front += ",";
            front += Z3_ast_to_string(ctx, Z3_optimize_get_upper(ctx, opt, 1));
            front += ")";
        }
        Z3_optimize_dec_ref(ctx, opt);
        Z3_del_context(ctx);
        Z3_global_param_reset_all();
        if (front != "(1,0)(0,1)" || checks != 2) {
            std::cerr << "fork_bugfixes: pareto_model_completion FAILED: front "
                      << front << " after " << checks << " checks\n";
        }
        ENSURE(front == "(1,0)(0,1)");
        ENSURE(checks == 2);
        std::cout << "fork_bugfixes: pareto_model_completion ok\n";
    }

    // Upstream 1ce0d7512 / 8ad939a10 (#2974, #2990): qe_rec on a formula
    // with free constants. Before: segfault in m_elevel.find.
    void test_qe_rec_free_boolean() {
        check("qe_rec_free_boolean",
              "(set-logic ALL)\n"
              "(declare-const p Bool)\n"
              "(assert (exists ((x Bool)) (or x p)))\n"
              "(assert (not p))\n"
              "(check-sat-using (then qe_rec smt))\n",
              "sat\n");
    }

    // Same backport: qe_rec must not turn a satisfiable existential with
    // free variables into false. The fixed elim leaves it unchanged.
    void test_qe_rec_real_disjunction() {
        std::string out = run_script(
            "(declare-const a Real)(declare-const b Real)"
            "(assert (exists ((x Real)) (and (< a x) (< (* 2 x) b) "
            "(or (> x 5.0) (< (+ x b) 7.0)))))"
            "(apply qe_rec)\n");
        if (contains(out, "false")) {
            std::cerr << "fork_bugfixes: qe_rec_real_disjunction FAILED (unsound false):\n" << out << "\n";
        }
        ENSURE(!contains(out, "false"));
        ENSURE(contains(out, "(exists ((x Real))"));
        std::cout << "fork_bugfixes: qe_rec_real_disjunction ok\n";
    }

    // theory_pb resolve_conflict justified a learned lemma with a
    // region-allocated justification that was later dealloc'ed.
    // Before: free(): invalid pointer when the context is destroyed.
    // (The API's SMT-LIB2 evaluator cannot turn on proofs/cores, so this
    // uses the solver API; the SMT-LIB2 form is
    // fork_bugfixes/test_vectors/pb_strengthening_proof.smt2.)
    void test_pb_strengthening_proof() {
        Z3_global_param_set("auto_config", "false");
        Z3_global_param_set("smt.core.validate", "true");
        Z3_global_param_set("smt.pb.conflict_frequency", "0");
        Z3_global_param_set("smt.pb.enable_compilation", "false");
        Z3_config cfg = Z3_mk_config();
        Z3_set_param_value(cfg, "proof", "true");
        Z3_context ctx = Z3_mk_context(cfg);
        Z3_del_config(cfg);
        char const * decls =
            "(declare-const x Bool)\n"
            "(declare-const y Bool)\n"
            "(declare-const u Bool)\n"
            "(declare-const z Bool)\n"
            "(declare-const w Bool)\n";
        std::string a = std::string(decls) + "(assert ((_ pbge 6 4 3 2) x y u))\n";
        std::string bb = std::string(decls) + "(assert ((_ pbge 4 3 2 1) (not x) z w))\n";
        Z3_ast fml_A = Z3_parse_smtlib2_string(ctx, a.c_str(), 0, nullptr, nullptr, 0, nullptr, nullptr);
        Z3_inc_ref(ctx, fml_A);
        Z3_ast fml_B = Z3_parse_smtlib2_string(ctx, bb.c_str(), 0, nullptr, nullptr, 0, nullptr, nullptr);
        Z3_inc_ref(ctx, fml_B);
        Z3_solver s = Z3_mk_solver(ctx);
        Z3_solver_inc_ref(ctx, s);
        Z3_sort b = Z3_mk_bool_sort(ctx);
        Z3_ast cut_A = Z3_mk_const(ctx, Z3_mk_string_symbol(ctx, "cut_A"), b);
        Z3_ast cut_B = Z3_mk_const(ctx, Z3_mk_string_symbol(ctx, "cut_B"), b);
        Z3_solver_assert_and_track(ctx, s, fml_A, cut_A);
        Z3_solver_assert_and_track(ctx, s, fml_B, cut_B);
        Z3_lbool r = Z3_solver_check(ctx, s);
        ENSURE(r == Z3_L_FALSE);
        Z3_ast_vector core = Z3_solver_get_unsat_core(ctx, s);
        Z3_ast_vector_inc_ref(ctx, core);
        unsigned core_size = Z3_ast_vector_size(ctx, core);
        Z3_ast_vector_dec_ref(ctx, core);
        std::string proof = Z3_ast_to_string(ctx, Z3_solver_get_proof(ctx, s));
        Z3_solver_dec_ref(ctx, s);
        Z3_dec_ref(ctx, fml_A);
        Z3_dec_ref(ctx, fml_B);
        // Deleting the context is where the unfixed code crashed.
        Z3_del_context(ctx);
        Z3_global_param_reset_all();
        if (core_size != 2 || !contains(proof, "(th-lemma")) {
            std::cerr << "fork_bugfixes: pb_strengthening_proof FAILED: core size "
                      << core_size << "\n" << proof << "\n";
        }
        ENSURE(core_size == 2);
        ENSURE(contains(proof, "(th-lemma"));
        std::cout << "fork_bugfixes: pb_strengthening_proof ok\n";
    }
}

void tst_fork_bugfixes() {
    test_relevancy_raised_after_setup();
    test_pareto_model_completion();
    test_qe_rec_free_boolean();
    test_qe_rec_real_disjunction();
    test_pb_strengthening_proof();
}

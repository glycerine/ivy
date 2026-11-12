package satoko

import "fmt"

import "os"

//===--- solver_api.h -------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===
//#include <stdio.h>
//#include <assert.h>
//#include <string.h>
//#include <math.h>

//#include "act_var.h"
//#include "solver.h"
//#include "utils/misc.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_IMPL_START

//===------------------------------------------------------------------------===
// Satoko internal functions
//===------------------------------------------------------------------------===
func solver_rebuild_order(s *solver_t) {
    var var uint32;
    var vars *vec_uint_t  = vec_uint_alloc(vec_char_size(s.assigns));

    for var = 0; var < vec_char_size(s.assigns); var++ {
        if var_value(s, var) == SATOKO_VAR_UNASSING {
            vec_uint_push_back(vars, var); }
    }
    heap_build(s.var_order, vars);
    vec_uint_free(vars);
}

static inline int clause_is_satisfied(solver_t *s, struct clause *clause)
{
    var i uint32;
    var lits *uint32  = &(clause.data[0].lit);
    for i = 0; i < clause.size; i++ {
        if lit_value(s, lits[i]) == SATOKO_LIT_TRUE {
            return SATOKO_OK; }
    }
    return SATOKO_ERR;
}

func solver_clean_stats(s *solver_t) {
    var n_conflicts_all long  = s.stats.n_conflicts_all;
    var n_propagations_all long  = s.stats.n_propagations_all;
    memset(&(s.stats), 0, sizeof(struct satoko_stats));
    s.stats.n_conflicts_all = n_conflicts_all;
    s.stats.n_propagations_all = n_propagations_all;
}

func print_opts(s *solver_t) {
    fmt.Printf("+-[ BLACK MAGIC - PARAMETERS ]-+\n");
    fmt.Printf("|                              |\n");
    fmt.Printf("|--> Restarts heuristic        |\n");
    fmt.Printf("|    * LBD Queue   = %6d    |\n", s.opts.sz_lbd_bqueue);
    fmt.Printf("|    * Trail Queue = %6d    |\n", s.opts.sz_trail_bqueue);
    fmt.Printf("|    * f_rst       = %6.2f    |\n", s.opts.f_rst);
    fmt.Printf("|    * b_rst       = %6.2f    |\n", s.opts.b_rst);
    fmt.Printf("|                              |\n");
    fmt.Printf("|--> Clause DB reduction:      |\n");
    fmt.Printf("|    * First       = %6d    |\n", s.opts.n_conf_fst_reduce);
    fmt.Printf("|    * Inc         = %6d    |\n", s.opts.inc_reduce);
    fmt.Printf("|    * Special Inc = %6d    |\n", s.opts.inc_special_reduce);
    fmt.Printf("|    * Protected (LBD) < %2d    |\n", s.opts.lbd_freeze_clause);
    fmt.Printf("|                              |\n");
    fmt.Printf("|--> Binary resolution:        |\n");
    fmt.Printf("|    * Clause size < %3d       |\n", s.opts.clause_max_sz_bin_resol);
    fmt.Printf("|    * Clause lbd  < %3d       |\n", s.opts.clause_min_lbd_bin_resol);
    fmt.Printf("+------------------------------+\n\n");
}

func print_stats(s *solver_t) {
    fmt.Printf("starts        : %10d\n", s.stats.n_starts);
    fmt.Printf("conflicts     : %10d\n", s.stats.n_conflicts);
    fmt.Printf("decisions     : %10d\n", s.stats.n_decisions);
    fmt.Printf("propagations  : %10d\n", s.stats.n_propagations);
}

//===------------------------------------------------------------------------===
// Satoko external functions
//===------------------------------------------------------------------------===
func satoko_create() *solver_t {
    // inlining of macro satoko_calloc
    var s *solver_t  = &solver_t{};

    satoko_default_opts(&s.opts);
    s.status = SATOKO_OK;
    /* User data */
    s.assumptions = vec_uint_alloc(0);
    s.final_conflict = vec_uint_alloc(0);
    /* Clauses Database */
    s.all_clauses = cdb_alloc(0);
    s.originals = vec_uint_alloc(0);
    s.learnts = vec_uint_alloc(0);
    s.watches = vec_wl_alloc(0);
    /* Activity heuristic */
    s.var_act_inc = VAR_ACT_INIT_INC;
    s.clause_act_inc = CLAUSE_ACT_INIT_INC;
    /* Variable Information */
    // inlining of macro vec_act_alloc
    s.activity = vec_sdbl_alloc(0);
    s.var_order = heap_alloc(s.activity);
    s.levels = vec_uint_alloc(0);
    s.reasons = vec_uint_alloc(0);
    s.assigns = vec_char_alloc(0);
    s.polarity = vec_char_alloc(0);
    /* Assignments */
    s.trail = vec_uint_alloc(0);
    s.trail_lim = vec_uint_alloc(0);
    /* Temporary data used by Search method */
    s.bq_trail = b_queue_alloc(s.opts.sz_trail_bqueue);
    s.bq_lbd = b_queue_alloc(s.opts.sz_lbd_bqueue);
    s.n_confl_bfr_reduce = s.opts.n_conf_fst_reduce;
    s.RC1 = 1;
    s.RC2 = s.opts.n_conf_fst_reduce;
    /* Temporary data used by Analyze */
    s.temp_lits = vec_uint_alloc(0);
    s.seen = vec_char_alloc(0);
    s.tagged = vec_uint_alloc(0);
    s.stack = vec_uint_alloc(0);
    s.last_dlevel = vec_uint_alloc(0);
    /* Misc temporary */
    s.stamps = vec_uint_alloc(0);
    return s;
}

func satoko_destroy(s *solver_t) {
    vec_uint_free(s.assumptions);
    vec_uint_free(s.final_conflict);
    cdb_free(s.all_clauses);
    vec_uint_free(s.originals);
    vec_uint_free(s.learnts);
    vec_wl_free(s.watches);
    // inlining of macro vec_act_free
    vec_sdbl_free(s.activity);
    heap_free(s.var_order);
    vec_uint_free(s.levels);
    vec_uint_free(s.reasons);
    vec_char_free(s.assigns);
    vec_char_free(s.polarity);
    vec_uint_free(s.trail);
    vec_uint_free(s.trail_lim);
    b_queue_free(s.bq_lbd);
    b_queue_free(s.bq_trail);
    vec_uint_free(s.temp_lits);
    vec_char_free(s.seen);
    vec_uint_free(s.tagged);
    vec_uint_free(s.stack);
    vec_uint_free(s.last_dlevel);
    vec_uint_free(s.stamps);
    if s.marks {
        vec_char_free(s.marks); }
    // inlining of macro satoko_free
    for { free(s); s = nil; 
if !(false) { break }
} 
}

func satoko_default_opts(opts *satoko_opts_t) {
    memset(opts, 0, sizeof(satoko_opts_t));
    opts.verbose = 0;
    opts.no_simplify = 0;
    /* Limits */
    opts.conf_limit = 0;
    opts.prop_limit  = 0;
    /* Constants used for restart heuristic */
    opts.f_rst = 0.8;
    opts.b_rst = 1.4;
    opts.fst_block_rst   = 10000;
    opts.sz_lbd_bqueue   = 50;
    opts.sz_trail_bqueue = 5000;
    /* Constants used for clause database reduction heuristic */
    opts.n_conf_fst_reduce = 2000;
    opts.inc_reduce = 300;
    opts.inc_special_reduce = 1000;
    opts.lbd_freeze_clause = 30;
    opts.learnt_ratio = 0.5;
    /* VSIDS heuristic */
    opts.var_act_limit = VAR_ACT_LIMIT;
    opts.var_act_rescale = VAR_ACT_RESCALE;
    opts.var_decay = 0.95;
    opts.clause_decay = (clause_act_t) 0.995;
    /* Binary resolution */
    opts.clause_max_sz_bin_resol = 30;
    opts.clause_min_lbd_bin_resol = 6;

    opts.garbage_max_ratio = float32( 0.3);
}

/**
 * TODO: sanity check on configuration options
 */
func satoko_configure(s *satoko_t, user_opts *satoko_opts_t) {
    assert((user_opts != nil), "user_opts");
    memcpy(&s.opts, user_opts, sizeof(satoko_opts_t));
}

func satoko_simplify(s *solver_t) int32 {
    var i, j uint32;
    var cref uint32;

    assert(solver_dlevel(s) == 0, "solver_dlevel(s) == 0");
    if solver_propagate(s) != UNDEF {
        return SATOKO_ERR; }
    if s.n_assigns_simplify == vec_uint_size(s.trail) || s.n_props_simplify > 0 {
        return SATOKO_OK; }

    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(s.originals)) && (((cref) = vec_uint_at(s.originals, i)), 1); i++ {
        struct clause *clause = clause_fetch(s, cref);

    if clause_is_satisfied(s, clause) {
            clause.f_mark = 1;
            s.stats.n_original_lits -= clause.size;
            clause_unwatch(s, cref);
        } else {
            vec_uint_assign(s.originals, j++, cref); }
    }
    vec_uint_shrink(s.originals, j);
    solver_rebuild_order(s);
    s.n_assigns_simplify = vec_uint_size(s.trail);
    s.n_props_simplify = s.stats.n_original_lits + s.stats.n_learnt_lits;
    return SATOKO_OK;
}

func satoko_setnvars(s *solver_t, nvars int32) {
    var i int32;
    for i = satoko_varnum(s); i < nvars; i++ {
        satoko_add_variable(s, 0);
    }
}

func satoko_add_variable(s *solver_t, sign byte) int32 {
    // inlining of macro vec_act_size
    var var uint32  = vec_sdbl_size(s.activity);
    vec_wl_push(s.watches);
    vec_wl_push(s.watches);
    // inlining of macro vec_act_push_back
    vec_sdbl_push_back(s.activity, 0);
    vec_uint_push_back(s.levels, 0);
    vec_char_push_back(s.assigns, SATOKO_VAR_UNASSING);
    vec_char_push_back(s.polarity, sign);
    vec_uint_push_back(s.reasons, UNDEF);
    vec_uint_push_back(s.stamps, 0);
    vec_char_push_back(s.seen, 0);
    heap_insert(s.var_order, var);
    if s.marks {
        vec_char_push_back(s.marks, 0); }
    return var;
}

func satoko_add_clause(s *solver_t, lits *int32, size int32) int32 {
    var i, j uint32;
    var prev_lit uint32;
    var max_var uint32;
    var cref uint32;

    qsort((void *) lits, (size_t)size, sizeof(unsigned), stk_uint_compare);
    max_var = lit2var(lits[size - 1]);
    // inlining of macro vec_act_size
    for max_var >= vec_sdbl_size(s.activity) {
        satoko_add_variable(s, SATOKO_LIT_FALSE);
    }

    vec_uint_clear(s.temp_lits);
    j = 0;
    prev_lit = UNDEF;
    for i = 0; i < uint32(size); i++ {
        if uint32(lits[i]) == lit_compl(prev_lit) || lit_value(s, lits[i]) == SATOKO_LIT_TRUE {
            return SATOKO_OK;
        } else if uint32(lits[i]) != prev_lit && var_value(s, lit2var(lits[i])) == SATOKO_VAR_UNASSING {
            prev_lit = lits[i];
            vec_uint_push_back(s.temp_lits, lits[i]);
        }
    }

    if vec_uint_size(s.temp_lits) == 0 {
        s.status = SATOKO_ERR;
        return SATOKO_ERR;
    }
 if vec_uint_size(s.temp_lits) == 1 {
        solver_enqueue(s, vec_uint_at(s.temp_lits, 0), UNDEF);
        return (s.status = (solver_propagate(s) == UNDEF));
    }
    if false {
        for  i = 0; i < vec_uint_size(s.temp_lits); i++  {
            var lit int32  = vec_uint_at(s.temp_lits, i);
            fmt.Printf("%s%d ", lit&1 ? "!":"", lit>>1);
        }
        fmt.Printf("\n");
    }
    cref = solver_clause_create(s, s.temp_lits, 0);
    clause_watch(s, cref);
    return SATOKO_OK;
}

func satoko_assump_push(s *solver_t, lit int32) {
    assert(lit2var(lit) < uint32(satoko_varnum(s)), "lit2var(lit) < uint32(satoko_varnum(s))");
    // printf("[Satoko] Push assumption: %d\n", lit);
    vec_uint_push_back(s.assumptions, lit);
    vec_char_assign(s.polarity, lit2var(lit), lit_polarity(lit));
}

func satoko_assump_pop(s *solver_t) {
    assert(vec_uint_size(s.assumptions) > 0, "vec_uint_size(s.assumptions) > 0");
    // printf("[Satoko] Pop assumption: %d\n", vec_uint_pop_back(s->assumptions));
    vec_uint_pop_back(s.assumptions);
    solver_cancel_until(s, vec_uint_size(s.assumptions));
}

func satoko_solve(s *solver_t) int32 {
    var status int32  = SATOKO_UNDEC;

    assert(s, "s");
    solver_clean_stats(s);
    //if (s->opts.verbose)
    //    print_opts(s);
    if s.status == SATOKO_ERR {
        fmt.Printf("Satoko in inconsistent state\n");
        return SATOKO_UNDEC;
    }

    if !s.opts.no_simplify {
        if satoko_simplify(s) != SATOKO_OK {
            return SATOKO_UNDEC; } }

    for status == SATOKO_UNDEC {
        status = solver_search(s);
        if solver_check_limits(s) == 0 || solver_stop(s) {
            break; }
        if s.nRuntimeLimit && Abc_Clock() > s.nRuntimeLimit {
            break; }
        if s.pFuncStop && s.pFuncStop(s.RunId) {
            break; }
    }
    if s.opts.verbose {
        print_stats(s); }
    
    solver_cancel_until(s, vec_uint_size(s.assumptions));
    return status;
}

func satoko_solve_assumptions(s *solver_t, plits *int32, nlits int32) int32 {
    var i, status int32;
    // printf("\n[Satoko] Solve with assumptions.. (%d)\n", vec_uint_size(s->assumptions));
    // printf("[Satoko]   + Variables: %d\n", satoko_varnum(s));
    // printf("[Satoko]   + Clauses: %d\n", satoko_clausenum(s));
    // printf("[Satoko]   + Trail size: %d\n", vec_uint_size(s->trail));
    // printf("[Satoko]   + Queue head: %d\n", s->i_qhead);
    // solver_debug_check_trail(s);
    for  i = 0; i < nlits; i++  {
        satoko_assump_push( s, plits[i] );
    }
    status = satoko_solve( s );
    for  i = 0; i < nlits; i++  {
        satoko_assump_pop( s );
    }
    return status;
}

func satoko_solve_assumptions_limit(s *satoko_t, plits *int32, nlits int32, nconflim int32) int32 {
    var temp int32 = s.opts.conf_limit;
    var status int32;
    if nconflim != 0 {
        s.opts.conf_limit = s.stats.n_conflicts + nconflim;
    } else {
        s.opts.conf_limit = 0;
    }
    status = satoko_solve_assumptions(s, plits, nlits);
    s.opts.conf_limit = temp;
    return status;
}
func satoko_minimize_assumptions(s *satoko_t, plits *int32, nlits int32, nconflim int32) int32 {
    var i, nlitsL, nlitsR, nresL, nresR, status int32;
    if nlits == 1 {
        // since the problem is UNSAT, we try to solve it without assuming the last literal
        // if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
        status = satoko_solve_assumptions_limit( s, nil, 0, nconflim );
        return int32((status != SATOKO_UNSAT)); // return 1 if the problem is not UNSAT
    }
    assert(nlits >= 2, "nlits >= 2");
    nlitsL = nlits / 2;
    nlitsR = nlits - nlitsL;
    // assume the left lits
    for  i = 0; i < nlitsL; i++  {
        satoko_assump_push(s, plits[i]);
    }
    // solve with these assumptions
    status = satoko_solve_assumptions_limit( s, nil, 0, nconflim );
    if status == SATOKO_UNSAT { // these are enough
    
        for  i = 0; i < nlitsL; i++  {
            satoko_assump_pop(s);
        }
        return satoko_minimize_assumptions( s, plits, nlitsL, nconflim );
    }
    // these are not enoguh
    // solve for the right lits
    if nlitsR == 1 {
        nresL = 1;
    } else {
        nresL = satoko_minimize_assumptions( s, plits + nlitsL, nlitsR, nconflim );
    }
    for  i = 0; i < nlitsL; i++  {
        satoko_assump_pop(s);
    }
    // swap literals
    vec_uint_clear(s.temp_lits);
    for  i = 0; i < nlitsL; i++  {
        vec_uint_push_back(s.temp_lits, plits[i]);
    }
    for  i = 0; i < nresL; i++  {
        plits[i] = plits[nlitsL+i];
    }
    for  i = 0; i < nlitsL; i++  {
        plits[nresL+i] = vec_uint_at(s.temp_lits, i);
    }
    // assume the right lits
    for  i = 0; i < nresL; i++  {
        satoko_assump_push(s, plits[i]);
    }
    // solve with these assumptions
    status = satoko_solve_assumptions_limit( s, nil, 0, nconflim );
    if status == SATOKO_UNSAT { // these are enough
    
        for  i = 0; i < nresL; i++  {
            satoko_assump_pop(s);
        }
        return nresL;
    }
    // solve for the left lits
    if nlitsL == 1 {
        nresR = 1;
    } else {
        nresR = satoko_minimize_assumptions( s, plits + nresL, nlitsL, nconflim );
    }
    for  i = 0; i < nresL; i++  {
        satoko_assump_pop(s);
    }
    return nresL + nresR;
}

func satoko_final_conflict(s *solver_t, out **int32) int32 {
    *out = (int *)vec_uint_data(s.final_conflict);
    return vec_uint_size(s.final_conflict);
}

func satoko_stats(s *satoko_t) *satoko_stats_t {
    return &s.stats;
}

func satoko_options(s *satoko_t) *satoko_opts_t {
    return &s.opts;
}

func satoko_bookmark(s *satoko_t) {
    // printf("[Satoko] Bookmark.\n");
    assert(s.status == SATOKO_OK, "s.status == SATOKO_OK");
    assert(solver_dlevel(s) == 0, "solver_dlevel(s) == 0");
    s.book_cl_orig = vec_uint_size(s.originals);
    s.book_cl_lrnt = vec_uint_size(s.learnts);
    s.book_vars = vec_char_size(s.assigns);
    s.book_trail = vec_uint_size(s.trail);
    // s->book_qhead = s->i_qhead;
    s.opts.no_simplify = 1;
}

func satoko_unbookmark(s *satoko_t) {
    // printf("[Satoko] Unbookmark.\n");
    assert(s.status == SATOKO_OK, "s.status == SATOKO_OK");
    s.book_cl_orig = 0;
    s.book_cl_lrnt = 0;
    s.book_cdb = 0;
    s.book_vars = 0;
    s.book_trail = 0;
    // s->book_qhead = 0;
    s.opts.no_simplify = 0;
}

func satoko_reset(s *satoko_t) {
    // printf("[Satoko] Reset.\n");
    vec_uint_clear(s.assumptions);
    vec_uint_clear(s.final_conflict);
    cdb_clear(s.all_clauses);
    vec_uint_clear(s.originals);
    vec_uint_clear(s.learnts);
    vec_wl_clean(s.watches);
    // inlining of macro vec_act_clear
    vec_sdbl_clear(s.activity);
    heap_clear(s.var_order);
    vec_uint_clear(s.levels);
    vec_uint_clear(s.reasons);
    vec_char_clear(s.assigns);
    vec_char_clear(s.polarity);
    vec_uint_clear(s.trail);
    vec_uint_clear(s.trail_lim);
    b_queue_clean(s.bq_lbd);
    b_queue_clean(s.bq_trail);
    vec_uint_clear(s.temp_lits);
    vec_char_clear(s.seen);
    vec_uint_clear(s.tagged);
    vec_uint_clear(s.stack);
    vec_uint_clear(s.last_dlevel);
    vec_uint_clear(s.stamps);
    s.status = SATOKO_OK;
    s.var_act_inc = VAR_ACT_INIT_INC;
    s.clause_act_inc = CLAUSE_ACT_INIT_INC;
    s.n_confl_bfr_reduce = s.opts.n_conf_fst_reduce;
    s.RC1 = 1;
    s.RC2 = s.opts.n_conf_fst_reduce;
    s.book_cl_orig = 0;
    s.book_cl_lrnt = 0;
    s.book_cdb = 0;
    s.book_vars = 0;
    s.book_trail = 0;
    s.i_qhead = 0;
}

func satoko_rollback(s *satoko_t) {
    var i, cref uint32;
    var n_originals uint32  = vec_uint_size(s.originals) - s.book_cl_orig;
    var n_learnts uint32  = vec_uint_size(s.learnts) - s.book_cl_lrnt;
    struct clause **cl_to_remove;

    // printf("[Satoko] rollback.\n");
    assert(s.status == SATOKO_OK, "s.status == SATOKO_OK");
    assert(solver_dlevel(s) == 0, "solver_dlevel(s) == 0");
    if !s.book_vars {
        satoko_reset(s);
        return;
    }
    // inlining of macro satoko_alloc
    cl_to_remove = ((struct clause * *) malloc((n_originals + n_learnts) * sizeof(struct clause *)));
    /* Mark clauses */
    // inlining of macro vec_uint_foreach_start
    for i = s.book_cl_orig; (i < vec_uint_size(s.originals)) && (((cref) = vec_uint_at(s.originals, i)), 1); i++ {
        cl_to_remove[i] = clause_fetch(s, cref);
    }
    // inlining of macro vec_uint_foreach_start
    for i = s.book_cl_lrnt; (i < vec_uint_size(s.learnts)) && (((cref) = vec_uint_at(s.learnts, i)), 1); i++ {
        cl_to_remove[n_originals + i] = clause_fetch(s, cref);
    }
    for i = 0; i < n_originals + n_learnts; i++ {
        clause_unwatch(s, cdb_cref(s.all_clauses, (unsigned *)cl_to_remove[i]));
        cl_to_remove[i].f_mark = 1;
    }
    // inlining of macro satoko_free
    for { free(cl_to_remove); cl_to_remove = nil; 
if !(false) { break }
} 
    vec_uint_shrink(s.originals, s.book_cl_orig);
    vec_uint_shrink(s.learnts, s.book_cl_lrnt);
    /* Shrink variable related vectors */
    for i = s.book_vars; i < 2 * vec_char_size(s.assigns); i++ {
        vec_wl_at(s.watches, i).size = 0;
        vec_wl_at(s.watches, i).n_bin = 0;
    }
    // s->i_qhead = s->book_qhead;
    s.watches.size = s.book_vars;
    // inlining of macro vec_act_shrink
    vec_sdbl_shrink(s.activity, s.book_vars);
    vec_uint_shrink(s.levels, s.book_vars);
    vec_uint_shrink(s.reasons, s.book_vars);
    vec_uint_shrink(s.stamps, s.book_vars);
    vec_char_shrink(s.assigns, s.book_vars);
    vec_char_shrink(s.seen, s.book_vars);
    vec_char_shrink(s.polarity, s.book_vars);
    solver_rebuild_order(s);
    /* Rewind solver and cancel level 0 assignments to the trail */
    solver_cancel_until(s, 0);
    vec_uint_shrink(s.trail, s.book_trail);
    if s.book_cdb {
        s.all_clauses.size = s.book_cdb; }
    s.book_cl_orig = 0;
    s.book_cl_lrnt = 0;
    s.book_vars = 0;
    s.book_trail = 0;
    // s->book_qhead = 0;
}

func satoko_mark_cone(s *satoko_t, pvars *int32, n_vars int32) {
    var i int32;
    if !solver_has_marks(s) {
        s.marks = vec_char_init(satoko_varnum(s), 0); }
    for i = 0; i < n_vars; i++ {
        var_set_mark(s, pvars[i]);
        vec_sdbl_assign(s.activity, pvars[i], 0);
        if !heap_in_heap(s.var_order, pvars[i]) {
            heap_insert(s.var_order, pvars[i]); }
    }
}

func satoko_unmark_cone(s *satoko_t, pvars *int32, n_vars int32) {
    var i int32;
    assert(solver_has_marks(s), "solver_has_marks(s)");
    for i = 0; i < n_vars; i++ {
        var_clean_mark(s, pvars[i]);
    }
}

func satoko_write_dimacs(s *satoko_t, fname string, wrt_lrnt int32, zero_var int32) {
    var file *os.File;
    var i uint32;
    // inlining of macro vec_act_size
    var n_vars uint32  = vec_sdbl_size(s.activity);
    var n_orig uint32  = vec_uint_size(s.originals) + vec_uint_size(s.trail);
    var n_lrnts uint32  = vec_uint_size(s.learnts);
    var array *uint32;

    assert(wrt_lrnt == 0 || wrt_lrnt == 1, "wrt_lrnt == 0 || wrt_lrnt == 1");
    assert(zero_var == 0 || zero_var == 1, "zero_var == 0 || zero_var == 1");
    if fname != nil {
        file = os.OpenFile(fname, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    } else {
        file = os.Stdout; }
    
    if file == nil {
        fmt.Printf("Error: Cannot open output file.\n");
        return;
    }
    fmt.Fprintf(file, "p cnf %d %d\n", n_vars, wrt_lrnt ? n_orig + n_lrnts : n_orig);
    for i = 0; i < vec_char_size(s.assigns); i++ {
        if var_value(s, i) != SATOKO_VAR_UNASSING {
            if zero_var != 0 {
                fmt.Fprintf(file, "%d\n", var_value(s, i) == SATOKO_LIT_FALSE ? -int32((i)) : i);
            } else {
                fmt.Fprintf(file, "%d 0\n", var_value(s, i) == SATOKO_LIT_FALSE ? -int32((i + 1)) : i + 1); }
        }
    }
    array = vec_uint_data(s.originals);
    for i = 0; i < vec_uint_size(s.originals); i++ {
        clause_dump(file, clause_fetch(s, array[i]), !zero_var);
    }
    
    if wrt_lrnt != 0 {
        array = vec_uint_data(s.learnts);
        for i = 0; i < n_lrnts; i++ {
            clause_dump(file, clause_fetch(s, array[i]), !zero_var);
        }
    }
    _ = (file).Close();

}

func satoko_varnum(s *satoko_t) int32 {
    return vec_char_size(s.assigns);
}

func satoko_clausenum(s *satoko_t) int32 {
    return vec_uint_size(s.originals);
}

func satoko_learntnum(s *satoko_t) int32 {
    return vec_uint_size(s.learnts);
}

func satoko_conflictnum(s *satoko_t) int32 {
    return satoko_stats(s).n_conflicts_all;
}

func satoko_set_stop(s *satoko_t, pstop *int32) {
    s.pstop = pstop;
}

func satoko_set_stop_func(s *satoko_t, fnct func(int32) int32) {
    s.pFuncStop = fnct;
}

func satoko_set_runid(s *satoko_t, id int32) {
    s.RunId = id;
}

func satoko_read_cex_varvalue(s *satoko_t, ivar int32) int32 {
    return satoko_var_polarity(s, ivar) == SATOKO_LIT_TRUE;
}

func satoko_set_runtime_limit(s *satoko_t, Limit abctime) abctime {
    var nRuntimeLimit abctime  = s.nRuntimeLimit;
    s.nRuntimeLimit = Limit;
    return nRuntimeLimit;
}

func satoko_var_polarity(s *satoko_t, var uint32) byte {
    return vec_char_at(s.polarity, var);
}

//ABC_NAMESPACE_IMPL_END

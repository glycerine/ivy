package satoko

import (
	"fmt"
	"os"
)

//===--- solver.c -----------------------------------------------------------===
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

//#include "act_clause.h"
//#include "act_var.h"
//#include "solver.h"
//#include "utils/heap.h"
//#include "utils/mem.h"
//#include "utils/sort.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_IMPL_START

//===------------------------------------------------------------------------===
// Lit funtions
//===------------------------------------------------------------------------===
/**
 *  A literal is said to be redundant in a given clause if and only if all
 *  variables in its reason are either present in that clause or (recursevely)
 *  redundant.
 */
func lit_is_removable(s *solver_t, lit uint32, min_level uint32) int32 {
    var top uint32  = vec_uint_size(s.tagged);

    assert(lit_reason(s, lit) != UNDEF, "lit_reason(s, lit) != UNDEF");
    vec_uint_clear(s.stack);
    vec_uint_push_back(s.stack, lit2var(lit));
    for vec_uint_size(s.stack) {
        var i uint32;
        var var uint32  = vec_uint_pop_back(s.stack);
        struct clause *c = clause_fetch(s, var_reason(s, var));
        var lits *uint32  = &(c.data[0].lit);

        assert(var_reason(s, var) != UNDEF, "var_reason(s, var) != UNDEF");
        if c.size == 2 && lit_value(s, lits[0]) == SATOKO_LIT_FALSE {
            assert(lit_value(s, lits[1]) == SATOKO_LIT_TRUE, "lit_value(s, lits[1]) == SATOKO_LIT_TRUE");
            // inlining of macro stk_swap
            { var t uint32  = lits[0]; lits[0] = lits[1]; lits[1] = t; };
        }

        /* Check scan the literals of the reason clause.
         * The first literal is skiped because is the literal itself. */
        for i = 1; i < c.size; i++ {
            var = lit2var(lits[i]);

            /* Check if the variable has already been seen or if it
             * was assinged a value at the decision level 0. In a
             * positive case, there is no need to look any further */
            if vec_char_at(s.seen, var) || var_dlevel(s, var) == 0 {
                continue; }

            /* If the variable has a reason clause and if it was
             * assingned at a 'possible' level, then we need to
             * check if it is recursively redundant, otherwise the
             * literal being checked is not redundant */
            if var_reason(s, var) != UNDEF && ((1 << (var_dlevel(s, var) & 31)) & min_level) {
                vec_uint_push_back(s.stack, var);
                vec_uint_push_back(s.tagged, var);
                vec_char_assign(s.seen, var, 1);
            } else {
                // inlining of macro vec_uint_foreach_start
                for i = top; (i < vec_uint_size(s.tagged)) && (((var) = vec_uint_at(s.tagged, i)), 1); i++ {
                    vec_char_assign(s.seen, var, 0);
                }
                vec_uint_shrink(s.tagged, top);
                return 0;
            }
        }
    }
    return 1;
}

//===------------------------------------------------------------------------===
// Clause functions
//===------------------------------------------------------------------------===
/* Calculate clause LBD (Literal Block Distance):
 * - It's the number of variables in the final conflict clause that come from
 * different decision levels
 */
func clause_clac_lbd(s *solver_t, lits *uint32, size uint32) uint32 {
    var i uint32;
    var lbd uint32;

    s.cur_stamp++;
    for i = 0; i < size; i++ {
        var level uint32  = lit_dlevel(s, lits[i]);
        if vec_uint_at(s.stamps, level) != s.cur_stamp {
            vec_uint_assign(s.stamps, level, s.cur_stamp);
            lbd++;
        }
    }
    return lbd;
}

func clause_bin_resolution(s *solver_t, clause_lits *vec_uint_t) {
    var lits *uint32  = vec_uint_data(clause_lits);
    var counter, sz, i uint32;
    var lit uint32;
    var neg_lit uint32  = lit_compl(lits[0]);
    struct watcher *w;

    s.cur_stamp++;
    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(clause_lits)) && (((lit) = vec_uint_at(clause_lits, i)), 1); i++ {
        vec_uint_assign(s.stamps, lit2var(lit), s.cur_stamp);
    }

    counter = 0;
    // inlining of macro watch_list_foreach_bin
    for w = watch_list_array(vec_wl_at(s.watches, neg_lit));          w < watch_list_array(vec_wl_at(s.watches, neg_lit)) + vec_wl_at(s.watches, neg_lit).n_bin;      w++ {
        var imp_lit uint32  = w.blocker;
        if vec_uint_at(s.stamps, lit2var(imp_lit)) == s.cur_stamp &&
            lit_value(s, imp_lit) == SATOKO_LIT_TRUE {
            counter++;
            vec_uint_assign(s.stamps, lit2var(imp_lit), (s.cur_stamp - 1));
        }
    }
    if counter > 0 {
        sz = vec_uint_size(clause_lits) - 1;
        for i = 1; i < vec_uint_size(clause_lits) - counter; i++ {
            if vec_uint_at(s.stamps, lit2var(lits[i])) != s.cur_stamp {
                // inlining of macro stk_swap
                { var t uint32  = lits[i]; lits[i] = lits[sz]; lits[sz] = t; };
                i--;
                sz--;
            }
        }
        vec_uint_shrink(clause_lits, vec_uint_size(clause_lits) - counter);
    }
}

func clause_minimize(s *solver_t, clause_lits *vec_uint_t) {
    var i, j uint32;
    var lits *uint32  = vec_uint_data(clause_lits);
    var min_level uint32;
    var clause_size uint32;

    for i = 1; i < vec_uint_size(clause_lits); i++ {
        var level uint32  = lit_dlevel(s, lits[i]);
        min_level |= 1 << (level & 31);
    }

    /* Remove reduntant literals */
    // inlining of macro vec_uint_foreach
    for j = 0; (j < vec_uint_size(clause_lits)) && (((i) = vec_uint_at(clause_lits, j)), 1); j++ {
        vec_uint_push_back(s.tagged, lit2var(i));
    }
    for i = j = 1; i < vec_uint_size(clause_lits); i++ {
        if lit_reason(s, lits[i]) == UNDEF || !(lit_is_removable(s, lits[i], min_level) != 0) {
            lits[j++] = lits[i]; }
    }
    vec_uint_shrink(clause_lits, j);

    /* Binary Resolution */
    clause_size = vec_uint_size(clause_lits);
    if clause_size <= s.opts.clause_max_sz_bin_resol &&
        clause_clac_lbd(s, lits, clause_size) <= s.opts.clause_min_lbd_bin_resol {
        clause_bin_resolution(s, clause_lits); }
}

static inline void clause_realloc(struct cdb *dest, struct cdb *src, unsigned *cref)
{
    var new_cref uint32;
    struct clause *new_clause;
    struct clause *old_clause = cdb_handler(src, *cref);

    if old_clause.f_reallocd {
        *cref = uint32( old_clause.size);
        return;
    }
    new_cref = cdb_append(dest, 3 + old_clause.f_learnt + old_clause.size);
    new_clause = cdb_handler(dest, new_cref);
    memcpy(new_clause, old_clause, (size_t)((3 + old_clause.f_learnt + old_clause.size) * 4));
    old_clause.f_reallocd = 1;
    old_clause.size = uint32( new_cref);
    *cref = new_cref;
}

//===------------------------------------------------------------------------===
// Solver internal functions
//===------------------------------------------------------------------------===
func solver_decide(s *solver_t) uint32 {
    var next_var uint32  = UNDEF;

    for next_var == UNDEF || var_value(s, next_var) != SATOKO_VAR_UNASSING {
        if heap_size(s.var_order) == 0 {
            next_var = UNDEF;
            return UNDEF;
        }
        next_var = heap_remove_min(s.var_order);
        if solver_has_marks(s) && !var_mark(s, next_var) {
            next_var = UNDEF; }
    }
    return var2lit(next_var, satoko_var_polarity(s, next_var));
}

func solver_new_decision(s *solver_t, lit uint32) {
    if solver_has_marks(s) && !var_mark(s, lit2var(lit)) {
        return; }
    assert(var_value(s, lit2var(lit)) == SATOKO_VAR_UNASSING, "var_value(s, lit2var(lit)) == SATOKO_VAR_UNASSING");
    vec_uint_push_back(s.trail_lim, vec_uint_size(s.trail));
    solver_enqueue(s, lit, UNDEF);
}

/* Calculate Backtrack Level from the learnt clause */
func solver_calc_bt_level(s *solver_t, learnt *vec_uint_t) uint32 {
    var i, tmp uint32;
        var i_max uint32  = 1;
    var lits *uint32  = vec_uint_data(learnt);
        var max uint32  = lit_dlevel(s, lits[1]);

    if vec_uint_size(learnt) == 1 {
        return 0; }
    for i = 2; i < vec_uint_size(learnt); i++ {
        if lit_dlevel(s, lits[i]) > max {
            max   = lit_dlevel(s, lits[i]);
            i_max = i;
        }
    }
        tmp         = lits[1];
        lits[1]     = lits[i_max];
        lits[i_max] = tmp;
        return lit_dlevel(s, lits[1]);
}

/**
 *  Most books and papers explain conflict analysis and the calculation of the
 *  1UIP (first Unique Implication Point) using an implication graph. This
 *  function, however, do not explicity constructs the graph! It inspects the
 *  trail in reverse and figure out which literals should be added to the
 *  to-be-learnt clause using the reasons of each assignment.
 *
 *  cur_lit: current literal being analyzed.
 *  n_paths: number of unprocessed paths from conlfict node to the current
 *           literal being analyzed (cur_lit).
 *
 *  This functions performs a backward BFS (breadth-first search) for 1UIP node.
 *  The trail works as the BFS queue. The counter of unprocessed but seen
 *  variables (n_paths) allows us to identify when 'cur_lit' is the closest
 *  cause of conflict.
 *
 *  When 'n_paths' reaches zero it means there are no unprocessed reverse paths
 *  back from the conflict node to 'cur_lit' - meaning it is the 1UIP decision
 *  variable.
 *
 */
func solver_analyze(s *solver_t, cref uint32, learnt *vec_uint_t, bt_level *uint32, lbd *uint32) {
    var i uint32;
    var trail *uint32  = vec_uint_data(s.trail);
    var idx uint32  = vec_uint_size(s.trail) - 1;
    var n_paths uint32;
    var p uint32  = UNDEF;
    var var uint32;

    vec_uint_push_back(learnt, UNDEF);
    do {
        struct clause *clause;
        var lits *uint32;
        var j uint32;

        assert(cref != UNDEF, "cref != UNDEF");
        clause = clause_fetch(s, cref);
        lits = &(clause.data[0].lit);

        if p != UNDEF && clause.size == 2 && lit_value(s, lits[0]) == SATOKO_LIT_FALSE {
            assert(lit_value(s, lits[1]) == SATOKO_LIT_TRUE, "lit_value(s, lits[1]) == SATOKO_LIT_TRUE");
            // inlining of macro stk_swap
            { var t uint32  = lits[0]; lits[0] = lits[1]; lits[1] = t; };
        }

        if clause.f_learnt {
            clause_act_bump(s, clause); }

        if clause.f_learnt && clause.lbd > 2 {
            var n_levels uint32  = clause_clac_lbd(s, lits, clause.size);
            if n_levels + 1 < clause.lbd {
                if clause.lbd <= s.opts.lbd_freeze_clause {
                    clause.f_deletable = 0; }
                clause.lbd = n_levels;
            }
        }

        for j = (p == UNDEF ? 0 : 1); j < clause.size; j++ {
            var = lit2var(lits[j]);
            if vec_char_at(s.seen, var) || var_dlevel(s, var) == 0 {
                continue; }
            vec_char_assign(s.seen, var, 1);
            var_act_bump(s, var);
            if var_dlevel(s, var) == solver_dlevel(s) {
                n_paths++;
                if var_reason(s, var) != UNDEF && clause_fetch(s, var_reason(s, var)).f_learnt {
                    vec_uint_push_back(s.last_dlevel, var); }
            } else {
                vec_uint_push_back(learnt, lits[j]); }
        }

        for !vec_char_at(s.seen, lit2var(trail[idx--])) {; }

        p = trail[idx + 1];
        cref = lit_reason(s, p);
        vec_char_assign(s.seen, lit2var(p), 0);
        n_paths--;
    } while (n_paths > 0);

    vec_uint_data(learnt)[0] = lit_compl(p);
    clause_minimize(s, learnt);
    *bt_level = solver_calc_bt_level(s, learnt);
    *lbd = clause_clac_lbd(s, vec_uint_data(learnt), vec_uint_size(learnt));

    if vec_uint_size(s.last_dlevel) > 0 {
        // inlining of macro vec_uint_foreach
        for i = 0; (i < vec_uint_size(s.last_dlevel)) && (((var) = vec_uint_at(s.last_dlevel, i)), 1); i++ {
            if clause_fetch(s, var_reason(s, var)).lbd < *lbd {
                var_act_bump(s, var); }
        }
        vec_uint_clear(s.last_dlevel);
    }
    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(s.tagged)) && (((var) = vec_uint_at(s.tagged, i)), 1); i++ {
        vec_char_assign(s.seen, var, 0);
    }
    vec_uint_clear(s.tagged);
}

func solver_rst(s *solver_t) int32 {
    return b_queue_is_valid(s.bq_lbd) &&
           ((int64(b_queue_avg(s.bq_lbd)) * s.opts.f_rst) > (s.sum_lbd / s.stats.n_conflicts));
}

func solver_block_rst(s *solver_t) int32 {
    return s.stats.n_conflicts > int32(s.opts.fst_block_rst) &&
           b_queue_is_valid(s.bq_lbd) &&
           (int64(vec_uint_size(s.trail)) > (s.opts.b_rst * int64(b_queue_avg(s.bq_trail))));
}

func solver_handle_conflict(s *solver_t, confl_cref uint32) {
    var bt_level uint32;
    var lbd uint32;
    var cref uint32;

    vec_uint_clear(s.temp_lits);
    solver_analyze(s, confl_cref, s.temp_lits, &bt_level, &lbd);
    s.sum_lbd += lbd;
    b_queue_push(s.bq_lbd, lbd);
    solver_cancel_until(s, bt_level);
    cref = UNDEF;
    if vec_uint_size(s.temp_lits) > 1 {
        cref = solver_clause_create(s, s.temp_lits, 1);
        clause_watch(s, cref);
    }
    solver_enqueue(s, vec_uint_at(s.temp_lits, 0), cref);
    var_act_decay(s);
    clause_act_decay(s);
}

func solver_analyze_final(s *solver_t, lit uint32) {
    var i uint32;

    // printf("[Satoko] Analize final..\n");
    // printf("[Satoko] Conflicting lit: %d\n", lit);
    vec_uint_clear(s.final_conflict);
    vec_uint_push_back(s.final_conflict, lit);
    if solver_dlevel(s) == 0 {
        return; }
    vec_char_assign(s.seen, lit2var(lit), 1);
    for i = vec_uint_size(s.trail); i -. vec_uint_at(s.trail_lim, 0); {
        var var uint32  = lit2var(vec_uint_at(s.trail, i));

        if vec_char_at(s.seen, var) {
            var reason uint32  = var_reason(s, var);
            if reason == UNDEF {
                assert(var_dlevel(s, var) > 0, "var_dlevel(s, var) > 0");
                vec_uint_push_back(s.final_conflict, lit_compl(vec_uint_at(s.trail, i)));
            } else {
                var j uint32;
                struct clause *clause = clause_fetch(s, reason);
                for j = (clause.size == 2 ? 0 : 1); j < clause.size; j++ {
                    if lit_dlevel(s, clause.data[j].lit) > 0 {
                        vec_char_assign(s.seen, lit2var(clause.data[j].lit), 1); }
                }
            }
            vec_char_assign(s.seen, var, 0);
        }
    }
    vec_char_assign(s.seen, lit2var(lit), 0);
    // solver_debug_check_unsat(s);
}

func solver_garbage_collect(s *solver_t) {
    var i uint32;
    var array *uint32;
    struct cdb *new_cdb = cdb_alloc(cdb_capacity(s.all_clauses) - cdb_wasted(s.all_clauses));

    if s.book_cdb {
        s.book_cdb = 0; }

    for i = 0; i < 2 * vec_char_size(s.assigns); i++ {
        struct watcher *w;
        // inlining of macro watch_list_foreach
        for w = watch_list_array(vec_wl_at(s.watches, i));          w < watch_list_array(vec_wl_at(s.watches, i)) + watch_list_size(vec_wl_at(s.watches, i));      w++ {
            clause_realloc(new_cdb, s.all_clauses, &(w.cref));
        }
    }

    /* Update CREFS */
    for i = 0; i < vec_uint_size(s.trail); i++ {
        if lit_reason(s, vec_uint_at(s.trail, i)) != UNDEF {
            clause_realloc(new_cdb, s.all_clauses, &(vec_uint_data(s.reasons)[lit2var(vec_uint_at(s.trail, i))])); }
    }

    array = vec_uint_data(s.learnts);
    for i = 0; i < vec_uint_size(s.learnts); i++ {
        clause_realloc(new_cdb, s.all_clauses, &(array[i]));
    }

    array = vec_uint_data(s.originals);
    for i = 0; i < vec_uint_size(s.originals); i++ {
        clause_realloc(new_cdb, s.all_clauses, &(array[i]));
    }

    cdb_free(s.all_clauses);
    s.all_clauses = new_cdb;
}

func solver_reduce_cdb(s *solver_t) {
    var i, limit uint32;
    var n_learnts uint32  = vec_uint_size(s.learnts);
    var cref uint32;
    struct clause *clause;
    struct clause **learnts_cls;

    // inlining of macro satoko_alloc
    learnts_cls = ((struct clause * *) malloc((n_learnts) * sizeof(struct clause *)));
    // inlining of macro vec_uint_foreach_start
    for i = s.book_cl_lrnt; (i < vec_uint_size(s.learnts)) && (((cref) = vec_uint_at(s.learnts, i)), 1); i++ {
        learnts_cls[i] = clause_fetch(s, cref);
    }

    limit = uint32((n_learnts * s.opts.learnt_ratio));

    satoko_sort((void **)learnts_cls, n_learnts,
            (int (*)(void *, void *)) clause_compare);

    if learnts_cls[n_learnts / 2].lbd <= 3 {
        s.RC2 += s.opts.inc_special_reduce; }
    if learnts_cls[n_learnts - 1].lbd <= 6 {
        s.RC2 += s.opts.inc_special_reduce; }

    vec_uint_clear(s.learnts);
    for i = 0; i < n_learnts; i++ {
        clause = learnts_cls[i];
        cref = cdb_cref(s.all_clauses, (unsigned *)clause);
        assert(clause.f_mark == 0, "clause.f_mark == 0");
        if clause.f_deletable && clause.lbd > 2 && clause.size > 2 && lit_reason(s, clause.data[0].lit) != cref && (i < limit) {
            clause.f_mark = 1;
            s.stats.n_learnt_lits -= clause.size;
            clause_unwatch(s, cref);
            cdb_remove(s.all_clauses, clause);
        } else {
            if !clause.f_deletable {
                limit++; }
            clause.f_deletable = 1;
            vec_uint_push_back(s.learnts, cref);
        }
    }
    // inlining of macro satoko_free
    for { free(learnts_cls); learnts_cls = nil; 
if !(false) { break }
} 

    if s.opts.verbose {
        fmt.Printf("reduceDB: Keeping %7d out of %7d clauses (%5.2f %%) \n", vec_uint_size(s.learnts), n_learnts, 100.0 * vec_uint_size(s.learnts) / n_learnts);
        fflush(os.Stdout);
    }
    if cdb_wasted(s.all_clauses) > cdb_size(s.all_clauses) * s.opts.garbage_max_ratio {
        solver_garbage_collect(s); }
}

//===------------------------------------------------------------------------===
// Solver external functions
//===------------------------------------------------------------------------===
func solver_clause_create(s *solver_t, lits *vec_uint_t, f_learnt uint32) uint32 {
    struct clause *clause;
    var cref uint32;
    var n_words uint32;

    assert(vec_uint_size(lits) > 1, "vec_uint_size(lits) > 1");
    assert(f_learnt == 0 || f_learnt == 1, "f_learnt == 0 || f_learnt == 1");

    n_words = 3 + f_learnt + vec_uint_size(lits);
    cref = cdb_append(s.all_clauses, n_words);
    clause = clause_fetch(s, cref);
    clause.f_learnt = f_learnt;
    clause.f_mark = 0;
    clause.f_reallocd = 0;
    clause.f_deletable = f_learnt;
    clause.size = vec_uint_size(lits);
    memcpy(&(clause.data[0].lit), vec_uint_data(lits), sizeof(unsigned) * vec_uint_size(lits));

    if f_learnt != 0 {
        vec_uint_push_back(s.learnts, cref);
        clause.lbd = clause_clac_lbd(s, vec_uint_data(lits), vec_uint_size(lits));
        clause.data[clause.size].act = 0;
        s.stats.n_learnt_lits += vec_uint_size(lits);
        clause_act_bump(s, clause);
    } else {
        vec_uint_push_back(s.originals, cref);
        s.stats.n_original_lits += vec_uint_size(lits);
    }
    return cref;
}

func solver_cancel_until(s *solver_t, level uint32) {
    var i uint32;

    if solver_dlevel(s) <= level {
        return; }
    for i = vec_uint_size(s.trail); i -. vec_uint_at(s.trail_lim, level); {
        var var uint32  = lit2var(vec_uint_at(s.trail, i));

        vec_char_assign(s.assigns, var, SATOKO_VAR_UNASSING);
        vec_uint_assign(s.reasons, var, UNDEF);
        if !heap_in_heap(s.var_order, var) {
            heap_insert(s.var_order, var); }
    }
    s.i_qhead = vec_uint_at(s.trail_lim, level);
    vec_uint_shrink(s.trail, vec_uint_at(s.trail_lim, level));
    vec_uint_shrink(s.trail_lim, level);
}

func solver_propagate(s *solver_t) uint32 {
    var conf_cref uint32  = UNDEF;
    var lits *uint32;
    var neg_lit uint32;
    var n_propagations uint32;

    for s.i_qhead < vec_uint_size(s.trail) {
        var p uint32  = vec_uint_at(s.trail, s.i_qhead++);
        struct watch_list *ws;
        struct watcher *begin;
        struct watcher *end;
        struct watcher *i, *j;

        n_propagations++;
        // inlining of macro watch_list_foreach_bin
        for i = watch_list_array(vec_wl_at(s.watches, p));          i < watch_list_array(vec_wl_at(s.watches, p)) + vec_wl_at(s.watches, p).n_bin;      i++ {
            if solver_has_marks(s) && !var_mark(s, lit2var(i.blocker)) {
                continue; }
            if var_value(s, lit2var(i.blocker)) == SATOKO_VAR_UNASSING {
                solver_enqueue(s, i.blocker, i.cref);
            } else if lit_value(s, i.blocker) == SATOKO_LIT_FALSE {
                return i.cref; }
        }

        ws = vec_wl_at(s.watches, p);
        begin = watch_list_array(ws);
        end = begin + watch_list_size(ws);
        for i = j = begin + ws.n_bin; i < end; {
            struct clause *clause;
            struct watcher w;

            if solver_has_marks(s) && !var_mark(s, lit2var(i.blocker)) {
                *j++ = *i++;
                continue;
            }
            if lit_value(s, i.blocker) == SATOKO_LIT_TRUE {
                *j++ = *i++;
                continue;
            }

            clause = clause_fetch(s, i.cref);
            lits = &(clause.data[0].lit);

            // Make sure the false literal is data[1]:
            neg_lit = lit_compl(p);
            if lits[0] == neg_lit {
                // inlining of macro stk_swap
                 var t uint32  = lits[0]; lits[0] = lits[1]; lits[1] = t; };
            assert(lits[1] == neg_lit, "lits[1] == neg_lit");

            w.cref = i.cref;
            w.blocker = lits[0];

            /* If 0th watch is true, then clause is already satisfied. */
            if lits[0] != i.blocker && lit_value(s, lits[0]) == SATOKO_LIT_TRUE {
                *j++ = w;
            } else {
                /* Look for new watch */
                var k uint32;
                for k = 2; k < clause.size; k++ {
                    if lit_value(s, lits[k]) != SATOKO_LIT_FALSE {
                        lits[1] = lits[k];
                        lits[k] = neg_lit;
                        watch_list_push(vec_wl_at(s.watches, lit_compl(lits[1])), w, 0);
                        goto next;
                    }
                }

                *j++ = w;

                /* Clause becomes unit under this assignment */
                if lit_value(s, lits[0]) == SATOKO_LIT_FALSE {
                    conf_cref = i.cref;
                    s.i_qhead = vec_uint_size(s.trail);
                    i++;
                    // Copy the remaining watches:
                    for i < end {
                        *j++ = *i++;
                    }
                } else {
                    solver_enqueue(s, lits[0], i.cref); }
            }
        next:
            i++;
        }

        s.stats.n_inspects += j - watch_list_array(ws);
        watch_list_shrink(ws, j - watch_list_array(ws));
    }
    s.stats.n_propagations += n_propagations;
    s.stats.n_propagations_all += n_propagations;
    s.n_props_simplify -= n_propagations;
    return conf_cref;
}

func solver_search(s *solver_t) byte {
    s.stats.n_starts++;
    for {
        var confl_cref uint32  = solver_propagate(s);
        if confl_cref != UNDEF {
            s.stats.n_conflicts++;
            s.stats.n_conflicts_all++;
            if solver_dlevel(s) == 0 {
                return SATOKO_UNSAT; }
            /* Restart heuristic */
            b_queue_push(s.bq_trail, vec_uint_size(s.trail));
            if solver_block_rst(s) != 0 {
                b_queue_clean(s.bq_lbd); }
            solver_handle_conflict(s, confl_cref);
        } else {
            // solver_debug_check_clauses(s);
            /* No conflict */
            var next_lit uint32;

            if (solver_rst(s) != 0) || solver_check_limits(s) == 0 || solver_stop(s) || 
                (s.nRuntimeLimit && (s.stats.n_conflicts & 63) == 0 && Abc_Clock() > s.nRuntimeLimit) {
                b_queue_clean(s.bq_lbd);
                solver_cancel_until(s, 0);
                return SATOKO_UNDEC;
            }
            if !s.opts.no_simplify && solver_dlevel(s) == 0 {
                satoko_simplify(s); }

            /* Reduce the set of learnt clauses */
            if s.opts.learnt_ratio && vec_uint_size(s.learnts) > 100 &&
                s.stats.n_conflicts >= s.n_confl_bfr_reduce {
                s.RC1 = (s.stats.n_conflicts / s.RC2) + 1;
                solver_reduce_cdb(s);
                s.RC2 += s.opts.inc_reduce;
                s.n_confl_bfr_reduce = s.RC1 * s.RC2;
            }

            /* Make decisions based on user assumptions */
            next_lit = UNDEF;
            for solver_dlevel(s) < vec_uint_size(s.assumptions) {
                var lit uint32  = vec_uint_at(s.assumptions, solver_dlevel(s));
                if lit_value(s, lit) == SATOKO_LIT_TRUE {
                    vec_uint_push_back(s.trail_lim, vec_uint_size(s.trail));
                } else if lit_value(s, lit) == SATOKO_LIT_FALSE {
                    solver_analyze_final(s, lit_compl(lit));
                    return SATOKO_UNSAT;
                } else {
                    next_lit = lit;
                    break;
                }

            }
            if next_lit == UNDEF {
                s.stats.n_decisions++;
                next_lit = solver_decide(s);
                if next_lit == UNDEF {
                    // solver_debug_check(s, SATOKO_SAT);
                    return SATOKO_SAT;
                }
            }
            solver_new_decision(s, next_lit);
        }
    }
}

//===------------------------------------------------------------------------===
// Debug procedures
//===------------------------------------------------------------------------===
func solver_debug_check_trail(s *solver_t) {
    var i uint32;
    var array *uint32;
    var trail_dup *vec_uint_t  = vec_uint_alloc(0);
    fmt.Fprintf(os.Stdout, "[Satoko] Checking for trail(%d) inconsistencies...\n", vec_uint_size(s.trail));
    vec_uint_duplicate(trail_dup, s.trail);
    vec_uint_sort(trail_dup, 1);
    array = vec_uint_data(trail_dup);
    for i = 1; i < vec_uint_size(trail_dup); i++ {
        if array[i - 1] == lit_compl(array[i]) {
            fmt.Fprintf(os.Stdout, "[Satoko] Inconsistent trail: %d %d\n", array[i - 1], array[i]);
            assert(false, "0");
            return;
        }
    }
    for i = 0; i < vec_uint_size(trail_dup); i++ {
        if var_value(s, lit2var(array[i])) != lit_polarity(array[i]) {
            fmt.Fprintf(os.Stdout, "[Satoko] Inconsistent trail assignment: %d, %d\n", vec_char_at(s.assigns, lit2var(array[i])), array[i]);
            assert(false, "0");
            return;
        }
    }
    fmt.Fprintf(os.Stdout, "[Satoko] Trail OK.\n");
    vec_uint_print(trail_dup);
    vec_uint_free(trail_dup);

}

func solver_debug_check_clauses(s *solver_t) {
    var cref, i uint32;

    fmt.Fprintf(os.Stdout, "[Satoko] Checking clauses (%d)...\n", vec_uint_size(s.originals));
    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(s.originals)) && (((cref) = vec_uint_at(s.originals, i)), 1); i++ {
        var j uint32;
        struct clause *clause = clause_fetch(s, cref);
        for j = 0; j < clause.size; j++ {
            if vec_uint_find(s.trail, lit_compl(clause.data[j].lit)) {
                continue;
            }
            break;
        }
        if j == clause.size {
            vec_uint_print(s.trail);
            fmt.Fprintf(os.Stdout, "[Satoko] FOUND UNSAT CLAUSE]: (%d) ", i);
            clause_print(clause);
            assert(false, "0");
        }
    }
    fmt.Fprintf(os.Stdout, "[Satoko] All SAT - OK\n");
}

func solver_debug_check(s *solver_t, result int32) {
    var cref, i uint32;
    solver_debug_check_trail(s);
    fmt.Fprintf(os.Stdout, "[Satoko] Checking clauses (%d)... \n", vec_uint_size(s.originals));
    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(s.originals)) && (((cref) = vec_uint_at(s.originals, i)), 1); i++ {
        var j uint32;
        struct clause *clause = clause_fetch(s, cref);
        for j = 0; j < clause.size; j++ {
            if vec_uint_find(s.trail, clause.data[j].lit) {
                break;
            }
        }
        if result == SATOKO_SAT && j == clause.size {
            fmt.Fprintf(os.Stdout, "[Satoko] FOUND UNSAT CLAUSE: (%d) ", i);
            clause_print(clause);
            assert(false, "0");
        }
    }
    fmt.Fprintf(os.Stdout, "[Satoko] All SAT - OK\n");
}

//ABC_NAMESPACE_IMPL_END

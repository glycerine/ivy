package satoko

//===--- solver.h -----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "clause.h"
//#include "cdb.h"
//#include "satoko.h"
//#include "types.h"
//#include "watch_list.h"
//#include "utils/b_queue.h"
//#include "utils/heap.h"
//#include "utils/mem.h"
//#include "utils/misc.h"
//#include "utils/vec/vec_char.h"
//#include "utils/vec/vec_sdbl.h"
//#include "utils/vec/vec_uint.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START



const UNDEF = 0xFFFFFFFF

   
type solver_t struct {
    status int32;
    /* User data */
    assumptions *vec_uint_t;
    final_conflict *vec_uint_t;

    /* Clauses Database */
    struct cdb *all_clauses;
    learnts *vec_uint_t;
    originals *vec_uint_t;
    watches *vec_wl_t;

    /* Activity heuristic */
    var_act_inc act_t; /* Amount to bump next variable with. */
    clause_act_inc clause_act_t; /* Amount to bump next clause with. */

    /* Variable Information */
    activity *vec_act_t; /* A heuristic measurement of the activity of a variable. */
    var_order *heap_t;
    levels *vec_uint_t;  /* Decision level of the current assignment */
    reasons *vec_uint_t; /* Reason (clause) of the current assignment */
    assigns *vec_char_t;
    polarity *vec_char_t;

    /* Assignments */
    trail *vec_uint_t;
    trail_lim *vec_uint_t; /* Separator indices for different decision levels in 'trail'. */
    i_qhead uint32; /* Head of propagation queue (as index into the trail). */
    n_assigns_simplify uint32; /* Number of top-level assignments since
                    last execution of 'simplify()'. */
    n_props_simplify long;  /* Remaining number of propagations that
                   must be made before next execution of
                   'simplify()'. */

    /* Temporary data used by Analyze */
    temp_lits *vec_uint_t;
    seen *vec_char_t;
    tagged *vec_uint_t; /* Stack */
    stack *vec_uint_t;
    last_dlevel *vec_uint_t;

    /* Temporary data used by Search method */
    bq_trail *b_queue_t;
    bq_lbd *b_queue_t;
    RC1 long;
    RC2 long;
    n_confl_bfr_reduce long;
    sum_lbd float32;

    /* Misc temporary */
    cur_stamp uint32; /* Used for marking literals and levels of interest */
    stamps *vec_uint_t; /* Multipurpose stamp used to calculate LBD and
                 * clauses minimization with binary resolution */

    /* Bookmark */
    book_cl_orig uint32; /* Bookmark for orignal problem clauses vector */
    book_cl_lrnt uint32; /* Bookmark for learnt clauses vector */
    book_cdb uint32;     /* Bookmark clause database size */
    book_vars uint32;    /* Bookmark number of variables */
    book_trail uint32;   /* Bookmark trail size */
    // unsigned book_qhead;   /* Bookmark propagation queue head */

    /* Temporary data used for solving cones */
    marks *vec_char_t;
    
    /* Callbacks to stop the solver */
    nRuntimeLimit abctime;
    pstop    *int32;
    RunId     int32;           
    int   (*pFuncStop)(int);  

    struct satoko_stats stats;
    struct satoko_opts opts;
}

//===------------------------------------------------------------------------===
// extern unsigned solver_clause_create(solver_t *, vec_uint_t *, unsigned);

// extern char solver_search(solver_t *);

// extern void solver_cancel_until(solver_t *, unsigned);

// extern unsigned solver_propagate(solver_t *);


/* Debuging */
// extern void solver_debug_check(solver_t *, int);

// extern void solver_debug_check_trail(solver_t *);

// extern void solver_debug_check_clauses(solver_t *);


//===------------------------------------------------------------------------===
// Inline var/lit functions
//===------------------------------------------------------------------------===
func var2lit(var uint32, polarity byte) uint32 {
    return var + var + (uint32( polarity) != 0);
}

func lit2var(lit uint32) uint32 {
    return lit >> 1;
}
//===------------------------------------------------------------------------===
// Inline var functions
//===------------------------------------------------------------------------===
func var_value(s *solver_t, var uint32) byte {
    return vec_char_at(s.assigns, var);
}

func var_dlevel(s *solver_t, var uint32) uint32 {
    return vec_uint_at(s.levels, var);
}

func var_reason(s *solver_t, var uint32) uint32 {
    return vec_uint_at(s.reasons, var);
}
func var_mark(s *solver_t, var uint32) int32 {
    return int32(vec_char_at(s.marks, var));
}
func var_set_mark(s *solver_t, var uint32) {
    vec_char_assign(s.marks, var, 1);
}
func var_clean_mark(s *solver_t, var uint32) {
    vec_char_assign(s.marks, var, 0);
}
//===------------------------------------------------------------------------===
// Inline lit functions
//===------------------------------------------------------------------------===
func lit_compl(lit uint32) uint32 {
    return lit ^ 1;
}

func lit_polarity(lit uint32) byte {
    return byte((lit & 1));
}

func lit_value(s *solver_t, lit uint32) byte {
    return lit_polarity(lit) ^ vec_char_at(s.assigns, lit2var(lit));
}

func lit_dlevel(s *solver_t, lit uint32) uint32 {
    return vec_uint_at(s.levels, lit2var(lit));
}

func lit_reason(s *solver_t, lit uint32) uint32 {
    return vec_uint_at(s.reasons, lit2var(lit));
}
//===------------------------------------------------------------------------===
// Inline solver minor functions
//===------------------------------------------------------------------------===
func solver_check_limits(s *solver_t) uint32 {
    return (s.opts.conf_limit == 0 || s.opts.conf_limit >= s.stats.n_conflicts) &&
           (s.opts.prop_limit == 0 || s.opts.prop_limit >= s.stats.n_propagations);
}

/** Returns current decision level */
func solver_dlevel(s *solver_t) uint32 {
    return vec_uint_size(s.trail_lim);
}

func solver_enqueue(s *solver_t, lit uint32, reason uint32) int32 {
    var var uint32  = lit2var(lit);

    vec_char_assign(s.assigns, var, lit_polarity(lit));
    vec_char_assign(s.polarity, var, lit_polarity(lit));
    vec_uint_assign(s.levels, var, solver_dlevel(s));
    vec_uint_assign(s.reasons, var, reason);
    vec_uint_push_back(s.trail, lit);
    return SATOKO_OK;
}

func solver_has_marks(s *satoko_t) int32 {
    return int32((s.marks != nil));
}

func solver_stop(s *satoko_t) int32 {
    return s.pstop && *s.pstop;
}

//===------------------------------------------------------------------------===
// Inline clause functions
//===------------------------------------------------------------------------===
func clause_fetch(s *solver_t, cref uint32) *struct clause {
    return cdb_handler(s.all_clauses, cref);
}

func clause_watch(s *solver_t, cref uint32) {
    struct clause *clause = cdb_handler(s.all_clauses, cref);
    struct watcher w1;
    struct watcher w2;

    w1.cref = cref;
    w2.cref = cref;
    w1.blocker = clause.data[1].lit;
    w2.blocker = clause.data[0].lit;
    watch_list_push(vec_wl_at(s.watches, lit_compl(clause.data[0].lit)), w1, (clause.size == 2));
    watch_list_push(vec_wl_at(s.watches, lit_compl(clause.data[1].lit)), w2, (clause.size == 2));
}

func clause_unwatch(s *solver_t, cref uint32) {
    struct clause *clause = cdb_handler(s.all_clauses, cref);
    watch_list_remove(vec_wl_at(s.watches, lit_compl(clause.data[0].lit)), cref, (clause.size == 2));
    watch_list_remove(vec_wl_at(s.watches, lit_compl(clause.data[1].lit)), cref, (clause.size == 2));
}

//ABC_NAMESPACE_HEADER_END


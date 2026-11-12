package espresso

import (
	"fmt"
	"os"
)

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
//#include "mincov_int.h"

//ABC_NAMESPACE_IMPL_START


/*
 *  mincov.c
 */

#define USE_GIMPEL
#define USE_INDEP_SET

// static int select_column();

// static void select_essential();

// static int verify_cover();


// Original macro:
// #define fail(why) {\
//     (void) fprintf(stderr, "Fatal error: file %s, line %d\n%s\n",\
//     __FILE__, __LINE__, why);\
//     (void) fflush(stdout);\
//     abort();\
// }


sm_row *
sm_minimum_cover(A, weight, heuristic, debug_level)
var A *sm_matrix;
var weight *int32;
var heuristic int32;        /* set to 1 for a heuristic covering */
var debug_level int32;    /* how deep in the recursion to provide info */
{
    stats_t stats;
    solution_t *best, *select;
    sm_row *prow, *sol;
    sm_col *pcol;
    sm_matrix *dup_A;
    int nelem, bound;
    double sparsity;

    /* Avoid sillyness */
    if A.nrows <= 0 {
    return sm_row_alloc();        /* easy to cover */
    }

    /* Initialize debugging structure */
    stats.start_time = util_cpu_time();
    stats.debug = debug_level > 0;
    stats.max_print_depth = debug_level;
    stats.max_depth = -1;
    stats.nodes = 0;
    stats.component = stats.comp_count = 0;
    stats.gimpel = stats.gimpel_count = 0;
    stats.no_branching = heuristic != 0;
    stats.lower_bound = -1;

    /* Check the matrix sparsity */
    nelem = 0;
    // inlining of macro sm_foreach_row
    for prow = A.first_row; prow != 0; prow = prow.next_row {
    nelem += prow.length;
    }
    sparsity = float64( nelem) / float64( (A.nrows * A.ncols));

    /* Determine an upper bound on the solution */
    bound = 1;
    // inlining of macro sm_foreach_col
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
    bound += WEIGHT(weight, pcol.col_num);
    }

    /* Perform the covering */
    select = solution_alloc();
    dup_A = sm_dup(A);
    best = sm_mincov(dup_A, select, weight, 0, bound, 0, &stats);
    sm_free(dup_A);
    solution_free(select);

    if stats.debug {
    if stats.no_branching {
        (void) fmt.Printf("**** heuristic covering ...\n");
        (void) fmt.Printf("lower bound = %d\n", stats.lower_bound);
    }
    (void) fmt.Printf("matrix     = %d by %d with %d elements (%4.3f%%)\n", A.nrows, A.ncols, nelem, sparsity * 100.0);
    (void) fmt.Printf("cover size = %d elements\n", best.row.length);
    (void) fmt.Printf("cover cost = %d\n", best.cost);
    (void) fmt.Printf("time       = %s\n", util_print_time(util_cpu_time() - stats.start_time));
    (void) fmt.Printf("components = %d\n", stats.comp_count);
    (void) fmt.Printf("gimpel     = %d\n", stats.gimpel_count);
    (void) fmt.Printf("nodes      = %d\n", stats.nodes);
    (void) fmt.Printf("max_depth  = %d\n", stats.max_depth);
    }

    sol = sm_row_dup(best.row);
    if ! verify_cover(A, sol) {
    // inlining of macro fail
    {    (void) fmt.Fprintf(os.Stderr, "Fatal error: file %s, line %d\n%s\n", __FILE__, __LINE__, "mincov: internal error -- cover verification failed\n");    (void) fflush(os.Stdout);    abort();};
    }
    solution_free(best);
    return sol;
}

/*
 *  Find the best cover for 'A' (given that 'select' already selected);
 *
 *    - abort search if a solution cannot be found which beats 'bound'
 *
 *    - if any solution meets 'lower_bound', then it is the optimum solution
 *      and can be returned without further work.
 */

solution_t * 
sm_mincov(A, select, weight, lb, bound, depth, stats)
var A *sm_matrix;
var select *solution_t;
var weight *int32;
var lb int32;
var bound int32;
var depth int32;
var stats *stats_t;
{
    sm_matrix *A1, *A2, *L, *R;
    sm_element *p;
    solution_t *select1, *select2, *best, *best1, *best2, *indep;
    int pick, lb_new, debug;

    /* Start out with some debugging information */
    stats.nodes++;
    if (depth > stats.max_depth) stats.max_depth = depth;
    debug = stats.debug && (depth <= stats.max_print_depth);

    /* Apply row dominance, column dominance, and select essentials */
    select_essential(A, select, weight, bound);
    if select.cost >= bound {
    return NIL(solution_t);
    }

    /* See if gimpel's reduction technique applies ... */
                 
    if weight == NIL(int) {    /* hack until we fix it */
    if gimpel_reduce(A, select, weight, lb, bound, depth, stats, &best) {
        return best;
    }
    }
      

                    
    /* Determine bound from here to final solution using independent-set */
    indep = sm_maximal_independent_set(A, weight);

    /* make sure the lower bound is monotonically increasing */
    lb_new = MAX(select.cost + indep.cost, lb);
    pick = select_column(A, weight, indep);
    solution_free(indep);
     
                                           
                                                     
      

    if depth == 0 {
    stats.lower_bound = lb_new + stats.gimpel;
    }

    if debug {
        (void) fmt.Printf("ABSMIN[%2d]%s", depth, stats.component ? "*" : " ");
        (void) fmt.Printf(" %3dx%3d sel=%3d bnd=%3d lb=%3d %12s ", A.nrows, A.ncols, select.cost + stats.gimpel, bound + stats.gimpel, lb_new + stats.gimpel, util_print_time(util_cpu_time()-stats.start_time));
    }

    /* Check for bounding based on no better solution possible */
    if lb_new >= bound {
    if (debug) (void) fmt.Printf("bounded\n");
    best = NIL(solution_t);


    /* Check for new best solution */
    } else if A.nrows == 0 {
    best = solution_dup(select);
    if (debug) (void) fmt.Printf("BEST\n");
    if stats.debug && stats.component == 0 {
            (void) fmt.Printf("new 'best' solution %d at level %d (time is %s)\n", best.cost + stats.gimpel, depth, util_print_time(util_cpu_time() - stats.start_time));
        }


    /* Check for a partition of the problem */
    } else if sm_block_partition(A, &L, &R) {
    /* Make L the smaller problem */
    if L.ncols > R.ncols {
        A1 = L;
        L = R;
        R = A1;
    }
    if (debug) (void) fmt.Printf("comp %d %d\n", L.nrows, R.nrows);
    stats.comp_count++;

    /* Solve problem for L */
    select1 = solution_alloc();
    stats.component++;
    best1 = sm_mincov(L, select1, weight, 0, 
                    bound-select.cost, depth+1, stats);
    stats.component--;
    solution_free(select1);
    sm_free(L);

    /* Add best solution to the selected set */
    if best1 == NIL(solution_t) {
        best = NIL(solution_t);
    } else {
        for p = best1.row.first_col; p != 0; p = p.next_col {
        solution_add(select, weight, p.col_num);
        }
        solution_free(best1);

        /* recur for the remaining block */
        best = sm_mincov(R, select, weight, lb_new, bound, depth+1, stats);
    }
    sm_free(R);

    /* We've tried as hard as possible, but now we must split and recur */
    } else {
    if (debug) (void) fmt.Printf("pick=%d\n", pick);

        /* Assume we choose this column to be in the covering set */
    A1 = sm_dup(A);
    select1 = solution_dup(select);
    solution_accept(select1, A1, weight, pick);
        best1 = sm_mincov(A1, select1, weight, lb_new, bound, depth+1, stats);
    solution_free(select1);
    sm_free(A1);

    /* Update the upper bound if we found a better solution */
    if best1 != NIL(solution_t) && bound > best1.cost {
        bound = best1.cost;
    }

    /* See if this is a heuristic covering (no branching) */
    if stats.no_branching {
        return best1;
    }

    /* Check for reaching lower bound -- if so, don't actually branch */
    if best1 != NIL(solution_t) && best1.cost == lb_new {
        return best1;
    }

        /* Now assume we cannot have that column */
    A2 = sm_dup(A);
    select2 = solution_dup(select);
    solution_reject(select2, A2, weight, pick);
        best2 = sm_mincov(A2, select2, weight, lb_new, bound, depth+1, stats);
    solution_free(select2);
    sm_free(A2);

    best = solution_choose_best(best1, best2);
    }

    return best;
}

static int 
select_column(A, weight, indep)
var A *sm_matrix;
var weight *int32;
var indep *solution_t;
{
    register sm_col *pcol;
    register sm_row *prow, *indep_cols;
    register sm_element *p, *p1;
    double w, best;
    int best_col;

    indep_cols = sm_row_alloc();
    if indep != NIL(solution_t) {
    /* Find which columns are in the independent sets */
    for p = indep.row.first_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_row
        prow = (((p.col_num) >= 0 && (p.col_num) < (A).rows_size) ?     (A).rows[p.col_num] : (sm_row *) 0);
        for p1 = prow.first_col; p1 != 0; p1 = p1.next_col {
        (void) sm_row_insert(indep_cols, p1.col_num);
        }
    }
    } else {
    /* select out of all columns */
    // inlining of macro sm_foreach_col
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
        (void) sm_row_insert(indep_cols, pcol.col_num);
    }
    }

    /* Find the best column */
    best_col = -1;
    best = -1;

    /* Consider only columns which are in some independent row */
    // inlining of macro sm_foreach_row_element
    for p1 = indep_cols.first_col; p1 != 0; p1 = p1.next_col {
    // inlining of macro sm_get_col
    pcol = (((p1.col_num) >= 0 && (p1.col_num) < (A).cols_size) ?     (A).cols[p1.col_num] : (sm_col *) 0);

    /* Compute the total 'value' of all things covered by the column */
    w = 0.0;
    for p = pcol.first_row; p != 0; p = p.next_row {
        // inlining of macro sm_get_row
        prow = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
        w += 1.0 / (float64( prow.length) - 1.0);
    }

    /* divide this by the relative cost of choosing this column */
    w = w / float64( WEIGHT(weight, pcol.col_num));

    /* maximize this ratio */
    if w  > best {
        best_col = pcol.col_num;
        best = w;
    }
    }

    sm_row_free(indep_cols);
    return best_col;
}

static void 
select_essential(A, select, weight, bound)
var A *sm_matrix;
var select *solution_t;
var weight *int32;
var bound int32;            /* must beat this solution */
{
    register sm_element *p;
    register sm_row *prow, *essen;
    int delcols, delrows, essen_count;

    for {
    /*  Check for dominated columns  */
    delcols = sm_col_dominance(A, weight);

    /*  Find the rows with only 1 element (the essentials) */
    essen = sm_row_alloc();
    // inlining of macro sm_foreach_row
    for prow = A.first_row; prow != 0; prow = prow.next_row {
        if prow.length == 1 {
        (void) sm_row_insert(essen, prow.first_col.col_num);
        }
    }

    /* Select all of the elements */
    // inlining of macro sm_foreach_row_element
    for p = essen.first_col; p != 0; p = p.next_col {
        solution_accept(select, A, weight, p.col_num);
        /* Make sure solution still looks good */
        if select.cost >= bound {
        sm_row_free(essen);
        return;
        }
    }
    essen_count = essen.length;
    sm_row_free(essen);

    /*  Check for dominated rows  */
    delrows = sm_row_dominance(A);

    
if !(delcols > 0 || delrows > 0 || essen_count > 0) { break }
} 
}

static int 
verify_cover(A, cover)
var A *sm_matrix;
var cover *sm_row;
{
    sm_row *prow;

    // inlining of macro sm_foreach_row
    for prow = A.first_row; prow != 0; prow = prow.next_row {
    if ! sm_row_intersects(prow, cover) {
        return 0;
    }
    }
    return 1;
}
//ABC_NAMESPACE_IMPL_END


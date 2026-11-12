package espresso

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



func solution_alloc() *solution_t {
    var sol *solution_t;

    sol = ALLOC(solution_t, 1);
    sol.cost = 0;
    sol.row = sm_row_alloc();
    return sol;
}


void
solution_free(sol)
var sol *solution_t;
{
    sm_row_free(sol.row);
    FREE(sol);
}


solution_t *
solution_dup(sol)
var sol *solution_t;
{
    solution_t *new_sol;

    new_sol = ALLOC(solution_t, 1);
    new_sol.cost = sol.cost;
    new_sol.row = sm_row_dup(sol.row);
    return new_sol;
}


void 
solution_add(sol, weight, col)
var sol *solution_t;
var weight *int32;
var col int32;
{
    (void) sm_row_insert(sol.row, col);
    sol.cost += WEIGHT(weight, col);
}


void 
solution_accept(sol, A, weight, col)
var sol *solution_t;
var A *sm_matrix;
var weight *int32;
var col int32;
{
    register sm_element *p, *pnext;
    sm_col *pcol;

    solution_add(sol, weight, col);

    /* delete rows covered by this column */
    // inlining of macro sm_get_col
    pcol = (((col) >= 0 && (col) < (A).cols_size) ?     (A).cols[col] : (sm_col *) 0);
    for p = pcol.first_row; p != 0; p = pnext {
    pnext = p.next_row;        /* grab it before it disappears */
    sm_delrow(A, p.row_num);
    }
}


/* ARGSUSED */
void 
solution_reject(sol, A, weight, col)
var sol *solution_t;
var A *sm_matrix;
var weight *int32;
var col int32;
{
    sm_delcol(A, col);
}


solution_t *
solution_choose_best(best1, best2)
var best1, best2 *solution_t;
{
    if best1 != NIL(solution_t) {
    if best2 != NIL(solution_t) {
        if best1.cost <= best2.cost {
        solution_free(best2);
        return best1;
        } else {
        solution_free(best1);
        return best2;
        }
    } else {
        return best1;
    }
    } else {
    if best2 != NIL(solution_t) {
        return best2;
    } else {
        return NIL(solution_t);
    }
    }
}
//ABC_NAMESPACE_IMPL_END


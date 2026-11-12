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


// static sm_matrix *build_intersection_matrix();



     
  
                                                              
   
           
                          
             
              
 
                                  
                                

                                                        
                                     
                                                       
                                           
                                             
                 
         
     
     
             
 
      

solution_t * 
sm_maximal_independent_set(A, weight)
var A *sm_matrix;
var weight *int32;
{
    register sm_row *best_row, *prow;
    register sm_element *p;
    int least_weight;
    sm_row *save;
    sm_matrix *B;
    solution_t *indep;

    indep = solution_alloc();
    B = build_intersection_matrix(A);

    while (B.nrows > 0) {
    /*  Find the row which is disjoint from a maximum number of rows */
    best_row = B.first_row;
    for prow = B.first_row.next_row; prow != 0; prow = prow.next_row {
        if prow.length < best_row.length {
        best_row = prow;
        }
    }

    /* Find which element in this row has least weight */
    if weight == NIL(int) {
        least_weight = 1;
    } else {
        // inlining of macro sm_get_row
        prow = (((best_row.row_num) >= 0 && (best_row.row_num) < (A).rows_size) ?     (A).rows[best_row.row_num] : (sm_row *) 0);
        least_weight = weight[prow.first_col.col_num];
        for p = prow.first_col.next_col; p != 0; p = p.next_col {
        if weight[p.col_num] < least_weight {
            least_weight = weight[p.col_num];
        }
        }
    }
    indep.cost += least_weight;
    (void) sm_row_insert(indep.row, best_row.row_num);

    /*  Discard the rows which intersect this row */
    save = sm_row_dup(best_row);
    for p = save.first_col; p != 0; p = p.next_col {
        sm_delrow(B, p.col_num);
        sm_delcol(B, p.col_num);
    }
    sm_row_free(save);
    }

    sm_free(B);

/*
    if (! verify_indep_set(A, indep->row)) {
    fail("sm_maximal_independent_set: row set is not independent");
    }
*/
    return indep;
}

static sm_matrix *
build_intersection_matrix(A)
var A *sm_matrix;
{
    register sm_row *prow, *prow1;
    register sm_element *p, *p1;
    register sm_col *pcol;
    sm_matrix *B;

    /* Build row-intersection matrix */
    B = sm_alloc();
    for prow = A.first_row; prow != 0; prow = prow.next_row {

    /* Clear flags on all rows we can reach from row 'prow' */
    for p = prow.first_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_col
        pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        for p1 = pcol.first_row; p1 != 0; p1 = p1.next_row {
        // inlining of macro sm_get_row
        prow1 = (((p1.row_num) >= 0 && (p1.row_num) < (A).rows_size) ?     (A).rows[p1.row_num] : (sm_row *) 0);
        prow1.flag = 0;
        }
    }

    /* Now record which rows can be reached */
    for p = prow.first_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_col
        pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        for p1 = pcol.first_row; p1 != 0; p1 = p1.next_row {
        // inlining of macro sm_get_row
        prow1 = (((p1.row_num) >= 0 && (p1.row_num) < (A).rows_size) ?     (A).rows[p1.row_num] : (sm_row *) 0);
        if ! prow1.flag {
            prow1.flag = 1;
            (void) sm_insert(B, prow.row_num, prow1.row_num);
        }
        }
    }
    }

    return B;
}
//ABC_NAMESPACE_IMPL_END


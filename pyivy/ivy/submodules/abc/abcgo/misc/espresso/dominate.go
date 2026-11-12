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



int 
sm_row_dominance(A)
var A *sm_matrix; 
{
    register sm_row *prow, *prow1;
    register sm_col *pcol, *least_col;
    register sm_element *p, *pnext;
    int rowcnt;

    rowcnt = A.nrows;

    /* Check each row against all other rows */
    for prow = A.first_row; prow != 0; prow = prow.next_row {

    /* Among all columns with a 1 in this row, choose smallest */
    // inlining of macro sm_get_col
    least_col = (((prow.first_col.col_num) >= 0 && (prow.first_col.col_num) < (A).cols_size) ?     (A).cols[prow.first_col.col_num] : (sm_col *) 0);
    for p = prow.first_col.next_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_col
        pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        if pcol.length < least_col.length {
        least_col = pcol;
        }
    }

    /* Only check for containment against rows in this column */
    for p = least_col.first_row; p != 0; p = pnext {
        pnext = p.next_row;

        // inlining of macro sm_get_row
        prow1 = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
        if (prow1.length > prow.length) ||
                  (prow1.length == prow.length && 
                  prow1.row_num > prow.row_num) {
        if sm_row_contains(prow, prow1) {
            sm_delrow(A, prow1.row_num);
        }
        }
    }
    }

    return rowcnt - A.nrows;
}

int 
sm_col_dominance(A, weight)
var A *sm_matrix;
var weight *int32;
{
    register sm_row *prow;
    register sm_col *pcol, *pcol1;
    register sm_element *p;
    sm_row *least_row;
    sm_col *next_col;
    int colcnt;

    colcnt = A.ncols;

    /* Check each column against all other columns */
    for pcol = A.first_col; pcol != 0; pcol = next_col {
    next_col = pcol.next_col;

    /* Check all rows to find the one with fewest elements */
    // inlining of macro sm_get_row
    least_row = (((pcol.first_row.row_num) >= 0 && (pcol.first_row.row_num) < (A).rows_size) ?     (A).rows[pcol.first_row.row_num] : (sm_row *) 0);
    for p = pcol.first_row.next_row; p != 0; p = p.next_row {
        // inlining of macro sm_get_row
        prow = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
        if prow.length < least_row.length {
        least_row = prow;
        }
    }

    /* Only check for containment against columns in this row */
    for p = least_row.first_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_col
        pcol1 = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        if (weight != 0 && weight[pcol1.col_num] > weight[pcol.col_num])
        continue;
        if (pcol1.length > pcol.length) ||
           (pcol1.length == pcol.length && 
                   pcol1.col_num > pcol.col_num) {
        if sm_col_contains(pcol, pcol1) {
            sm_delcol(A, pcol.col_num);
            break;
        }
        }
    }
    }

    return colcnt - A.ncols;
}
//ABC_NAMESPACE_IMPL_END


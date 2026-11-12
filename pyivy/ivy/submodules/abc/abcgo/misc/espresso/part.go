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


// static int visit_col();


static void
copy_row(A, prow)
register sm_matrix *A;
register sm_row *prow;
{
    register sm_element *p;

    for p = prow.first_col; p != 0; p = p.next_col {
    (void) sm_insert(A, p.row_num, p.col_num);
    }
}


static int
visit_row(A, prow, rows_visited, cols_visited)
var A *sm_matrix;
var prow *sm_row;
var rows_visited *int32;
var cols_visited *int32;
{
    sm_element *p;
    sm_col *pcol;

    if ! prow.flag {
    prow.flag = 1;
    (*rows_visited)++;
    if *rows_visited == A.nrows {
        return 1;
    }
    for p = prow.first_col; p != 0; p = p.next_col {
        // inlining of macro sm_get_col
        pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        if ! pcol.flag {
        if visit_col(A, pcol, rows_visited, cols_visited) {
            return 1;
        }
        }
    }
    }
    return 0;
}


static int
visit_col(A, pcol, rows_visited, cols_visited)
var A *sm_matrix;
var pcol *sm_col;
var rows_visited *int32;
var cols_visited *int32;
{
    sm_element *p;
    sm_row *prow;

    if ! pcol.flag {
    pcol.flag = 1;
    (*cols_visited)++;
    if *cols_visited == A.ncols {
        return 1;
    }
    for p = pcol.first_row; p != 0; p = p.next_row {
        // inlining of macro sm_get_row
        prow = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
        if ! prow.flag {
        if visit_row(A, prow, rows_visited, cols_visited) {
            return 1;
        }
        }
    }
    }
    return 0;
}

int
sm_block_partition(A, L, R)
var A *sm_matrix;
var L, R **sm_matrix;
{
    int cols_visited, rows_visited;
    register sm_row *prow;
    register sm_col *pcol;

    /* Avoid the trivial case */
    if A.nrows == 0 {
    return 0;
    }

    /* Reset the visited flags for each row and column */
    for prow = A.first_row; prow != 0; prow = prow.next_row {
    prow.flag = 0;
    }
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
    pcol.flag = 0;
    }

    cols_visited = rows_visited = 0;
    if visit_row(A, A.first_row, &rows_visited, &cols_visited) {
    /* we found all of the rows */
    return 0;
    } else {
    *L = sm_alloc();
    *R = sm_alloc();
    for prow = A.first_row; prow != 0; prow = prow.next_row {
        if prow.flag {
        copy_row(*L, prow);
        } else {
        copy_row(*R, prow);
        }
    }
    return 1;
    }
}
//ABC_NAMESPACE_IMPL_END


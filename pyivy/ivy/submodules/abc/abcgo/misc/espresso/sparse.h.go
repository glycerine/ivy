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




//ABC_NAMESPACE_HEADER_START


/*
 *  sparse.h -- sparse matrix package header file
 */

type sm_element = sm_element_struct;
type sm_row = sm_row_struct;
type sm_col = sm_col_struct;
type sm_matrix = sm_matrix_struct;


/*
 *  sparse matrix element
 */
type sm_element_struct struct {
    row_num int32;        /* row number of this element */
    col_num int32;        /* column number of this element */
    next_row *sm_element;    /* next row in this column */
    prev_row *sm_element;    /* previous row in this column */
    next_col *sm_element;    /* next column in this row */
    prev_col *sm_element;    /* previous column in this row */
    user_word string;        /* user-defined word */
}


/*
 *  row header
 */
type sm_row_struct struct {
    row_num int32;        /* the row number */
    length int32;            /* number of elements in this row */
    flag int32;            /* user-defined word */
    first_col *sm_element;    /* first element in this row */
    last_col *sm_element;    /* last element in this row */
    next_row *sm_row;        /* next row (in sm_matrix linked list) */
    prev_row *sm_row;        /* previous row (in sm_matrix linked list) */
    user_word string;        /* user-defined word */
}


/*
 *  column header
 */
type sm_col_struct struct {
    col_num int32;        /* the column number */
    length int32;            /* number of elements in this column */
    flag int32;            /* user-defined word */
    first_row *sm_element;    /* first element in this column */
    last_row *sm_element;    /* last element in this column */
    next_col *sm_col;        /* next column (in sm_matrix linked list) */
    prev_col *sm_col;        /* prev column (in sm_matrix linked list) */
    user_word string;        /* user-defined word */
}


/*
 *  A sparse matrix
 */
type sm_matrix_struct struct {
    rows **sm_row;        /* pointer to row headers (by row #) */
    rows_size int32;        /* alloc'ed size of above array */
    cols **sm_col;        /* pointer to column headers (by col #) */
    cols_size int32;        /* alloc'ed size of above array */
    first_row *sm_row;        /* first row (linked list of all rows) */
    last_row *sm_row;        /* last row (linked list of all rows) */
    nrows int32;            /* number of rows */
    first_col *sm_col;        /* first column (linked list of columns) */
    last_col *sm_col;        /* last column (linked list of columns) */
    ncols int32;            /* number of columns */
    user_word string;        /* user-defined word */
}


// Original macro:
// #define sm_get_col(A, colnum)    \
//     (((colnum) >= 0 && (colnum) < (A)->cols_size) ? \
//     (A)->cols[colnum] : (sm_col *) 0)


// Original macro:
// #define sm_get_row(A, rownum)    \
//     (((rownum) >= 0 && (rownum) < (A)->rows_size) ? \
//     (A)->rows[rownum] : (sm_row *) 0)


// Original macro:
// #define sm_foreach_row(A, prow)    \
//     for(prow = A->first_row; prow != 0; prow = prow->next_row)


// Original macro:
// #define sm_foreach_col(A, pcol)    \
//     for(pcol = A->first_col; pcol != 0; pcol = pcol->next_col)


// Original macro:
// #define sm_foreach_row_element(prow, p)    \
//     for(p = prow->first_col; p != 0; p = p->next_col)


// Original macro:
// #define sm_foreach_col_element(pcol, p)    \
//     for(p = pcol->first_row; p != 0; p = p->next_row)


// Original macro:
// #define sm_put(x, val) \
//     (x->user_word = (char *) val)


// Original macro:
// #define sm_get(type, x) \
//     ((type) (x->user_word))


extern sm_matrix *sm_alloc(), *sm_alloc_size(), *sm_dup();
extern void sm_free(), sm_delrow(), sm_delcol(), sm_resize();
extern void sm_write(), sm_print(), sm_dump(), sm_cleanup();
extern void sm_copy_row(), sm_copy_col();
extern void sm_remove(), sm_remove_element();
extern sm_element *sm_insert(), *sm_find();
// extern sm_row *sm_longest_row();

// extern sm_col *sm_longest_col();

extern int sm_read(), sm_read_compressed();

extern sm_row *sm_row_alloc(), *sm_row_dup(), *sm_row_and();
extern void sm_row_free(), sm_row_remove(), sm_row_print();
extern sm_element *sm_row_insert(), *sm_row_find();
extern int sm_row_contains(), sm_row_intersects();
extern int sm_row_compare(), sm_row_hash();

extern sm_col *sm_col_alloc(), *sm_col_dup(), *sm_col_and();
extern void sm_col_free(), sm_col_remove(), sm_col_print();
extern sm_element *sm_col_insert(), *sm_col_find();
extern int sm_col_contains(), sm_col_intersects();
extern int sm_col_compare(), sm_col_hash();

extern int sm_row_dominance(), sm_col_dominance(), sm_block_partition();



//ABC_NAMESPACE_HEADER_END



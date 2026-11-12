package espresso

import "fmt"

import "os"

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
//#include "port.h"
//#include "sparse_int.h"

//ABC_NAMESPACE_IMPL_START


/*
 *  free-lists are only used if 'FAST_AND_LOOSE' is set; this is because
 *  we lose the debugging capability of libmm_t which trashes objects when
 *  they are free'd.  However, FAST_AND_LOOSE is much faster if matrices
 *  are created and freed frequently.
 */

#ifdef FAST_AND_LOOSE
sm_element *sm_element_freelist;
sm_row *sm_row_freelist;
sm_col *sm_col_freelist;
#endif


func sm_alloc() *sm_matrix {
    register sm_matrix *A;

    A = ALLOC(sm_matrix, 1);
    A.rows = NIL(sm_row *);
    A.cols = NIL(sm_col *);
    A.nrows = A.ncols = 0;
    A.rows_size = A.cols_size = 0;
    A.first_row = A.last_row = NIL(sm_row);
    A.first_col = A.last_col = NIL(sm_col);
    A.user_word = NIL(char);        /* for our user ... */
    return A;
}


sm_matrix *
sm_alloc_size(row, col)
var row, col int32;
{
    register sm_matrix *A;

    A = sm_alloc();
    sm_resize(A, row, col);
    return A;
}


void
sm_free(A)
var A *sm_matrix;
{
#ifdef FAST_AND_LOOSE
    register sm_row *prow;

    if (A->first_row != 0) {
    for(prow = A->first_row; prow != 0; prow = prow->next_row) {
        /* add the elements to the free list of elements */
        prow->last_col->next_col = sm_element_freelist;
        sm_element_freelist = prow->first_col;
    }

    /* Add the linked list of rows to the row-free-list */
    A->last_row->next_row = sm_row_freelist;
    sm_row_freelist = A->first_row;

    /* Add the linked list of cols to the col-free-list */
    A->last_col->next_col = sm_col_freelist;
    sm_col_freelist = A->first_col;
    }
#else
    register sm_row *prow, *pnext_row;
    register sm_col *pcol, *pnext_col;

    for(prow = A->first_row; prow != 0; prow = pnext_row) {
    pnext_row = prow->next_row;
    sm_row_free(prow);
    }
    for(pcol = A->first_col; pcol != 0; pcol = pnext_col) {
    pnext_col = pcol->next_col;
    pcol->first_row = pcol->last_row = NIL(sm_element);
    sm_col_free(pcol);
    }
#endif

    /* Free the arrays to map row/col numbers into pointers */
    FREE(A.rows);
    FREE(A.cols);
    FREE(A);
}


sm_matrix *
sm_dup(A)
var A *sm_matrix; 
{
    register sm_row *prow;
    register sm_element *p;
    register sm_matrix *B;

    B = sm_alloc();
    if A.last_row != 0 {
    sm_resize(B, A.last_row.row_num, A.last_col.col_num);
    for prow = A.first_row; prow != 0; prow = prow.next_row {
        for p = prow.first_col; p != 0; p = p.next_col {
        (void) sm_insert(B, p.row_num, p.col_num);
        }
    }
    }
    return B;
}


void 
sm_resize(A, row, col)
register sm_matrix *A;
var row, col int32;
{
    register int i, new_size;

    if row >= A.rows_size {
    new_size = MAX(A.rows_size*2, row+1);
    A.rows = REALLOC(sm_row *, A.rows, new_size);
    for i = A.rows_size; i < new_size; i++ {
        A.rows[i] = NIL(sm_row);
    }
    A.rows_size = new_size;
    }

    if col >= A.cols_size {
    new_size = MAX(A.cols_size*2, col+1);
    A.cols = REALLOC(sm_col *, A.cols, new_size);
    for i = A.cols_size; i < new_size; i++ {
        A.cols[i] = NIL(sm_col);
    }
    A.cols_size = new_size;
    }
}


/*  
 *  insert -- insert a value into the matrix
 */
sm_element *
sm_insert(A, row, col)
register sm_matrix *A;
register int row, col;
{
    register sm_row *prow;
    register sm_col *pcol;
    register sm_element *element;
    sm_element *save_element;

    if row >= A.rows_size || col >= A.cols_size {
    sm_resize(A, row, col);
    }

    prow = A.rows[row];
    if prow == NIL(sm_row) {
    prow = A.rows[row] = sm_row_alloc();
    prow.row_num = row;
    sorted_insert(sm_row, A.first_row, A.last_row, A.nrows, 
            next_row, prev_row, row_num, row, prow);
    }

    pcol = A.cols[col];
    if pcol == NIL(sm_col) {
    pcol = A.cols[col] = sm_col_alloc();
    pcol.col_num = col;
    sorted_insert(sm_col, A.first_col, A.last_col, A.ncols, 
            next_col, prev_col, col_num, col, pcol);
    }

    /* get a new item, save its address */
    sm_element_alloc(element);
    save_element = element;

    /* insert it into the row list */
    sorted_insert(sm_element, prow.first_col, prow.last_col, 
        prow.length, next_col, prev_col, col_num, col, element);

    /* if it was used, also insert it into the column list */
    if element == save_element {
    sorted_insert(sm_element, pcol.first_row, pcol.last_row, 
        pcol.length, next_row, prev_row, row_num, row, element);
    } else {
    /* otherwise, it was already in matrix -- free element we allocated */
    sm_element_free(save_element);
    }
    return element;
}


sm_element *
sm_find(A, rownum, colnum)
var A *sm_matrix;
var rownum, colnum int32;
{
    sm_row *prow;
    sm_col *pcol;

    // inlining of macro sm_get_row
    prow = (((rownum) >= 0 && (rownum) < (A).rows_size) ?     (A).rows[rownum] : (sm_row *) 0);
    if prow == NIL(sm_row) {
    return NIL(sm_element);
    } else {
    // inlining of macro sm_get_col
    pcol = (((colnum) >= 0 && (colnum) < (A).cols_size) ?     (A).cols[colnum] : (sm_col *) 0);
    if pcol == NIL(sm_col) {
        return NIL(sm_element);
    }
    if prow.length < pcol.length {
        return sm_row_find(prow, colnum);
    } else {
        return sm_col_find(pcol, rownum);
    }
    }
}


void
sm_remove(A, rownum, colnum)
var A *sm_matrix;
var rownum, colnum int32;
{
    sm_remove_element(A, sm_find(A, rownum, colnum));
}



void
sm_remove_element(A, p)
register sm_matrix *A; 
register sm_element *p;
{
    register sm_row *prow;
    register sm_col *pcol;

    if (p == 0) return;

    /* Unlink the element from its row */
    // inlining of macro sm_get_row
    prow = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
    // inlining of macro dll_unlink
    {     if p.prev_col == 0 {     prow.first_col = p.next_col;     } else {     p.prev_col.next_col = p.next_col;     }
     if p.next_col == 0 {     prow.last_col = p.prev_col;     } else {     p.next_col.prev_col = p.prev_col;     }
     prow.length--; };

    /* if no more elements in the row, discard the row header */
    if prow.first_col == NIL(sm_element) {
    sm_delrow(A, p.row_num);
    }

    /* Unlink the element from its column */
    // inlining of macro sm_get_col
    pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
    // inlining of macro dll_unlink
    {     if p.prev_row == 0 {     pcol.first_row = p.next_row;     } else {     p.prev_row.next_row = p.next_row;     }
     if p.next_row == 0 {     pcol.last_row = p.prev_row;     } else {     p.next_row.prev_row = p.prev_row;     }
     pcol.length--; };

    /* if no more elements in the column, discard the column header */
    if pcol.first_row == NIL(sm_element) {
    sm_delcol(A, p.col_num);
    }

    sm_element_free(p);
}

void 
sm_delrow(A, i)
var A *sm_matrix;
var i int32;
{
    register sm_element *p, *pnext;
    sm_col *pcol;
    sm_row *prow;

    // inlining of macro sm_get_row
    prow = (((i) >= 0 && (i) < (A).rows_size) ?     (A).rows[i] : (sm_row *) 0);
    if prow != NIL(sm_row) {
    /* walk across the row */
    for p = prow.first_col; p != 0; p = pnext {
        pnext = p.next_col;

        /* unlink the item from the column (and delete it) */
        // inlining of macro sm_get_col
        pcol = (((p.col_num) >= 0 && (p.col_num) < (A).cols_size) ?     (A).cols[p.col_num] : (sm_col *) 0);
        sm_col_remove_element(pcol, p);

        /* discard the column if it is now empty */
        if pcol.first_row == NIL(sm_element) {
        sm_delcol(A, pcol.col_num);
        }
    }

    /* discard the row -- we already threw away the elements */ 
    A.rows[i] = NIL(sm_row);
    // inlining of macro dll_unlink
    {     if prow.prev_row == 0 {     A.first_row = prow.next_row;     } else {     prow.prev_row.next_row = prow.next_row;     }
     if prow.next_row == 0 {     A.last_row = prow.prev_row;     } else {     prow.next_row.prev_row = prow.prev_row;     }
     A.nrows--; };
    prow.first_col = prow.last_col = NIL(sm_element);
    sm_row_free(prow);
    }
}

void 
sm_delcol(A, i)
var A *sm_matrix;
var i int32;
{
    register sm_element *p, *pnext;
    sm_row *prow;
    sm_col *pcol;

    // inlining of macro sm_get_col
    pcol = (((i) >= 0 && (i) < (A).cols_size) ?     (A).cols[i] : (sm_col *) 0);
    if pcol != NIL(sm_col) {
    /* walk down the column */
    for p = pcol.first_row; p != 0; p = pnext {
        pnext = p.next_row;

        /* unlink the element from the row (and delete it) */
        // inlining of macro sm_get_row
        prow = (((p.row_num) >= 0 && (p.row_num) < (A).rows_size) ?     (A).rows[p.row_num] : (sm_row *) 0);
        sm_row_remove_element(prow, p);

        /* discard the row if it is now empty */
        if prow.first_col == NIL(sm_element) {
        sm_delrow(A, prow.row_num);
        }
    }

    /* discard the column -- we already threw away the elements */ 
    A.cols[i] = NIL(sm_col);
    // inlining of macro dll_unlink
    {     if pcol.prev_col == 0 {     A.first_col = pcol.next_col;     } else {     pcol.prev_col.next_col = pcol.next_col;     }
     if pcol.next_col == 0 {     A.last_col = pcol.prev_col;     } else {     pcol.next_col.prev_col = pcol.prev_col;     }
     A.ncols--; };
    pcol.first_row = pcol.last_row = NIL(sm_element);
    sm_col_free(pcol);
    }
}

void
sm_copy_row(dest, dest_row, prow)
register sm_matrix *dest;
var dest_row int32;
var prow *sm_row;
{
    register sm_element *p;

    for p = prow.first_col; p != 0; p = p.next_col {
    (void) sm_insert(dest, dest_row, p.col_num);
    }
}


void
sm_copy_col(dest, dest_col, pcol)
register sm_matrix *dest;
var dest_col int32;
var pcol *sm_col;
{
    register sm_element *p;

    for p = pcol.first_row; p != 0; p = p.next_row {
    (void) sm_insert(dest, dest_col, p.row_num);
    }
}


sm_row *
sm_longest_row(A)
var A *sm_matrix;
{
    register sm_row *large_row, *prow;
    register int max_length;

    max_length = 0;
    large_row = NIL(sm_row);
    for prow = A.first_row; prow != 0; prow = prow.next_row {
    if prow.length > max_length {
        max_length = prow.length;
        large_row = prow;
    }
    }
    return large_row;
}


sm_col *
sm_longest_col(A)
var A *sm_matrix;
{
    register sm_col *large_col, *pcol;
    register int max_length;

    max_length = 0;
    large_col = NIL(sm_col);
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
    if pcol.length > max_length {
        max_length = pcol.length;
        large_col = pcol;
    }
    }
    return large_col;
}

int
sm_num_elements(A)
var A *sm_matrix;
{
    register sm_row *prow;
    register int num;

    num = 0;
    // inlining of macro sm_foreach_row
    for prow = A.first_row; prow != 0; prow = prow.next_row {
    num += prow.length;
    }
    return num;
}

int 
sm_read(fp, A)
var fp *os.File;
var A **sm_matrix;
{
    int i, j, err;

    *A = sm_alloc();
    while (! feof(fp)) {
    err = fscanf(fp, "%d %d", &i, &j);
    if err == -1 {
        return 1;
    } else if err != 2 {
        return 0;
    }
    (void) sm_insert(*A, i, j);
    }
    return 1;
}


int 
sm_read_compressed(fp, A)
var fp *os.File;
var A **sm_matrix;
{
    int i, j, k, nrows, ncols;
    unsigned long x;

    *A = sm_alloc();
    if fscanf(fp, "%d %d", &nrows, &ncols) != 2 {
    return 0;
    }
    sm_resize(*A, nrows, ncols);

    for i = 0; i < nrows; i++ {
    if fscanf(fp, "%lx", &x) != 1 {
        return 0;
    }
    for j = 0; j < ncols; j += 32 {
        if fscanf(fp, "%lx", &x) != 1 {
        return 0;
        }
        for k = j; x != 0; x >>= 1, k++ {
        if x & 1 {
            (void) sm_insert(*A, i, k);
        }
        }
    }
    }
    return 1;
}


void 
sm_write(fp, A)
var fp *os.File;
var A *sm_matrix;
{
    register sm_row *prow;
    register sm_element *p;

    for prow = A.first_row; prow != 0; prow = prow.next_row {
    for p = prow.first_col; p != 0; p = p.next_col {
        (void) fmt.Fprintf(fp, "%d %d\n", p.row_num, p.col_num);
    }
    }
}

void 
sm_print(fp, A)
var fp *os.File;
var A *sm_matrix;
{
    register sm_row *prow;
    register sm_col *pcol;
    int c;

    if A.last_col.col_num >= 100 {
    (void) fmt.Fprintf(fp, "    ");
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
        (void) fmt.Fprintf(fp, "%d", (pcol.col_num / 100)%10);
    }
    putc('\n', fp);
    }

    if A.last_col.col_num >= 10 {
    (void) fmt.Fprintf(fp, "    ");
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
        (void) fmt.Fprintf(fp, "%d", (pcol.col_num / 10)%10);
    }
    putc('\n', fp);
    }

    (void) fmt.Fprintf(fp, "    ");
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
    (void) fmt.Fprintf(fp, "%d", pcol.col_num % 10);
    }
    putc('\n', fp);

    (void) fmt.Fprintf(fp, "    ");
    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
    (void) fmt.Fprintf(fp, "-");
    }
    putc('\n', fp);

    for prow = A.first_row; prow != 0; prow = prow.next_row {
    (void) fmt.Fprintf(fp, "%3d:", prow.row_num);

    for pcol = A.first_col; pcol != 0; pcol = pcol.next_col {
        c = sm_row_find(prow, pcol.col_num) ? '1' : '.';
        putc(c, fp);
    }
    putc('\n', fp);
    }
}


void 
sm_dump(A, s, max)
var A *sm_matrix;
var s string;
var max int32;
{
    FILE *fp = os.Stdout;

    (void) fmt.Fprintf(fp, "%s %d rows by %d cols\n", s, A.nrows, A.ncols);
    if A.nrows < max {
    sm_print(fp, A);
    }
}

func sm_cleanup() {
#ifdef FAST_AND_LOOSE
    register sm_element *p, *pnext;
    register sm_row *prow, *pnextrow;
    register sm_col *pcol, *pnextcol;

    for(p = sm_element_freelist; p != 0; p = pnext) {
    pnext = p->next_col;
    FREE(p);
    }
    sm_element_freelist = 0;

    for(prow = sm_row_freelist; prow != 0; prow = pnextrow) {
    pnextrow = prow->next_row;
    FREE(prow);
    }
    sm_row_freelist = 0;

    for(pcol = sm_col_freelist; pcol != 0; pcol = pnextcol) {
    pnextcol = pcol->next_col;
    FREE(pcol);
    }
    sm_col_freelist = 0;
#endif
}
//ABC_NAMESPACE_IMPL_END


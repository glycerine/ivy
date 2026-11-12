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
//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START



pset
do_sm_minimum_cover(A)
var A pset_family;
{
    sm_matrix *M;
    sm_row *sparse_cover;
    sm_element *pe;
    pset cover;
    register int i, base, rownum;
    register unsigned val;
    register pset last, p;

    M = sm_alloc();
    rownum = 0;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    // inlining of macro foreach_set_element
    // inlining of macro LOOP
    for(i = (p[0] & 0x03ff); i > 0; )                    for(val = p[i], base = --i << LOGBPI; val != 0; base++, val >>= 1)          if val & 1 {
        (void) sm_insert(M, rownum, base);
    }
    rownum++;
    }

    sparse_cover = sm_minimum_cover(M, NIL(int), 1, 0);
    sm_free(M);

    // inlining of macro set_new
    cover = set_clear(ALLOC(unsigned int, SET_SIZE(A.sf_size)), A.sf_size);
    sm_foreach_row_element(sparse_cover, pe) {
    // inlining of macro set_insert
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (cover[(((pe.col_num) >> LOGBPI) + 1)] |= 1 << ((pe.col_num) & (BPI-1)));
    }
    sm_row_free(sparse_cover);

    return cover;
}
//ABC_NAMESPACE_IMPL_END


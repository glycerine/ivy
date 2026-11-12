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
/*
 *  unate.c -- routines for dealing with unate functions
 */

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


// static pset_family abs_covered();

// static pset_family abs_covered_many();

// static int abs_select_restricted();


pcover map_cover_to_unate(T)
pcube *T;
{
    register unsigned int word_test, word_set, bit_test, bit_set;
    register pcube p, pA;
    pset_family A;
    pcube *T1;
    int ncol, i;

    // inlining of macro CUBELISTSIZE
    A = sf_new((((pcube *) T[1] - T) - 3), cdata.vars_unate);
    // inlining of macro CUBELISTSIZE
    A.count = (((pcube *) T[1] - T) - 3);
    // inlining of macro foreachi_set
    for p, i = A.data, 0;i<A.count;p+=A.wsize,i++ {
    (void) set_clear(p, A.sf_size);
    }
    ncol = 0;

    for i = 0; i < cube.size; i++ {
    if cdata.part_zeros[i] > 0 {
        assert(ncol <= cdata.vars_unate, "ncol <= cdata.vars_unate");

        /* Copy a column from T to A */
        // inlining of macro WHICH_WORD
        word_test = (((i) >> LOGBPI) + 1);
        // inlining of macro WHICH_BIT
        bit_test = 1 << ((i) & (BPI-1));
        // inlining of macro WHICH_WORD
        word_set = (((ncol) >> LOGBPI) + 1);
        // inlining of macro WHICH_BIT
        bit_set = 1 << ((ncol) & (BPI-1));

        pA = A.data;
        for T1 = T+2; (p = *T1++) != 0;  {
        if (p[word_test] & bit_test) == 0 {
            pA[word_set] |= bit_set;
        }
        pA += A.wsize;
        }

        ncol++;
    }
    }

    return A;
}

pcover map_unate_to_cover(A)
var A pset_family;
{
    register int i, ncol, lp;
    register pcube p, pB;
    int var, nunate, *unate;
    pcube last;
    pset_family B;

    B = sf_new(A.count, cube.size);
    B.count = A.count;

    /* Find the unate variables */
    unate = ALLOC(int, cube.num_vars);
    nunate = 0;
    for var = 0; var < cube.num_vars; var++ {
    if cdata.is_unate[var] {
        unate[nunate++] = var;
    }
    }

    /* Loop for each set of A */
    pB = B.data;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {

    /* Initialize this set of B */
    INLINEset_fill(pB, cube.size);

    /* Now loop for the unate variables; if the part is in A,
     * then this variable of B should be a single 1 in the unate
     * part.
     */
    for ncol = 0; ncol < nunate; ncol++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((ncol) >> LOGBPI) + 1)] & (1 << ((ncol) & (BPI-1))) {
        lp = cube.last_part[unate[ncol]];
        for i = cube.first_part[unate[ncol]]; i <= lp; i++ {
            if cdata.part_zeros[i] == 0 {
            // inlining of macro set_remove
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (pB[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
            }
        }
        }
    }
    pB += B.wsize;
    }

    FREE(unate);
    return B;
}

/*
 *  unate_compl
 */

pset_family unate_compl(A)
var A pset_family;
{
    register pset p, last;

    /* Make sure A is single-cube containment minimal */
/*    A = sf_rev_contain(A);*/

    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    PUTSIZE(p, set_ord(p));
    }

    /* Recursively find the complement */
    A = unate_complement(A);

    /* Now, we can guarantee a minimal result by containing the result */
    A = sf_rev_contain(A);
    return A;
}


/*
 *  Assume SIZE(p) records the size of each set
 */
pset_family unate_complement(A)
var A pset_family;            /* disposes of A */
{
    pset_family Abar;
    register pset p, p1, restrict;
    register int i;
    int max_i, min_set_ord, j;

    /* Check for no sets in the matrix -- complement is the universe */
    if A.count == 0 {
    sf_free(A);
    Abar = sf_new(1, A.sf_size);
    // inlining of macro GETSET
    (void) set_clear(((Abar).data + (Abar).wsize * (Abar.count++)), A.sf_size);

    /* Check for a single set in the maxtrix -- compute de Morgan complement */
    } else if A.count == 1 {
    p = A.data;
    Abar = sf_new(A.sf_size, A.sf_size);
    for i = 0; i < A.sf_size; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        // inlining of macro GETSET
        p1 = set_clear(((Abar).data + (Abar).wsize * (Abar.count++)), A.sf_size);
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p1[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
        }
    }
    sf_free(A);

    } else {

    /* Select splitting variable as the variable which belongs to a set
     * of the smallest size, and which has greatest column count
     */
    // inlining of macro set_new
    restrict = set_clear(ALLOC(unsigned int, SET_SIZE(A.sf_size)), A.sf_size);
    min_set_ord = A.sf_size + 1;
    // inlining of macro foreachi_set
    for p, i = A.data, 0;i<A.count;p+=A.wsize,i++ {
        if SIZE(p) < min_set_ord {
        set_copy(restrict, p);
        min_set_ord = SIZE(p);
        } else if SIZE(p) == min_set_ord {
        set_or(restrict, restrict, p);
        }
    }

    /* Check for no data (shouldn't happen ?) */
    if min_set_ord == 0 {
        A.count = 0;
        Abar = A;

    /* Check for "essential" columns */
    } else if min_set_ord == 1 {
        Abar = unate_complement(abs_covered_many(A, restrict));
        sf_free(A);
        // inlining of macro foreachi_set
        for p, i = Abar.data, 0;i<Abar.count;p+=Abar.wsize,i++ {
        set_or(p, p, restrict);
        }

    /* else, recur as usual */
    } else {
        max_i = abs_select_restricted(A, restrict);

        /* Select those rows of A which are not covered by max_i,
         * recursively find all minimal covers of these rows, and
         * then add back in max_i
         */
        Abar = unate_complement(abs_covered(A, max_i));
        // inlining of macro foreachi_set
        for p, i = Abar.data, 0;i<Abar.count;p+=Abar.wsize,i++ {
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p[(((max_i) >> LOGBPI) + 1)] |= 1 << ((max_i) & (BPI-1)));
        }

        /* Now recur on A with all zero's on column max_i */
        // inlining of macro foreachi_set
        for p, i = A.data, 0;i<A.count;p+=A.wsize,i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((max_i) >> LOGBPI) + 1)] & (1 << ((max_i) & (BPI-1))) {
            // inlining of macro set_remove
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (p[(((max_i) >> LOGBPI) + 1)] &= ~ (1 << ((max_i) & (BPI-1))));
            j = SIZE(p) - 1;
            PUTSIZE(p, j);
        }
        }

        Abar = sf_append(Abar, unate_complement(A));
    }
    // inlining of macro set_free
    FREE(restrict);
    }

    return Abar;
}

pset_family exact_minimum_cover(T)
IN pset_family T;
{
    register pset p, last, p1;
    register int i, n;
    int lev, lvl;
    pset nlast;
    pset_family temp;
    long start = ptime();
    struct {
    pset_family sf;
    int level;
    }
 stack[32];                /* 32 suffices for 2 ** 32 cubes ! */

    if (T.count <= 0)
    return sf_new(1, T.sf_size);
    for(n = T.count, lev = 0; n != 0; n >>= 1, lev++)   ;

    /* A simple heuristic ordering */
    T = lex_sort(sf_save(T));

    /* Push a full set on the stack to get things started */
    n = 1;
    stack[0].sf = sf_new(1, T.sf_size);
    stack[0].level = lev;
    // inlining of macro GETSET
    set_fill(((stack[0].sf).data + (stack[0].sf).wsize * (stack[0].sf.count++)), T.sf_size);

    // inlining of macro GETSET
    nlast = ((T).data + (T).wsize * (T.count - 1));
    // inlining of macro foreach_set
    for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {

    /* "unstack" the set into a family */
    temp = sf_new(set_ord(p), T.sf_size);
    for(i = 0; i < T.sf_size; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        // inlining of macro GETSET
        p1 = set_fill(((temp).data + (temp).wsize * (temp.count++)), T.sf_size);
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p1[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
        }
    stack[n].sf = temp;
    stack[n++].level = lev;

    /* Pop the stack and perform (leveled) intersections */
    while (n > 1 && (stack[n-1].level==stack[n-2].level || p == nlast)) {
        temp = unate_intersect(stack[n-1].sf, stack[n-2].sf, FALSE);
        lvl = MIN(stack[n-1].level, stack[n-2].level) - 1;
        if debug & MINCOV && lvl < 10 {
        fmt.Printf("# EXACT_MINCOV[%d]: %4d = %4d x %4d, time = %s\n", lvl, temp.count, stack[n-1].sf.count, stack[n-2].sf.count, print_time(ptime() - start));
        (void) fflush(os.Stdout);
        }
        sf_free(stack[n-2].sf);
        sf_free(stack[n-1].sf);
        stack[n-2].sf = temp;
        stack[n-2].level = lvl;
        n--;
    }
    }

    temp = stack[0].sf;
    // inlining of macro set_new
    p1 = set_fill(set_clear(ALLOC(unsigned int, SET_SIZE(T.sf_size)), T.sf_size), T.sf_size);
    // inlining of macro foreach_set
    for(p=temp.data,last=p+temp.count*temp.wsize;p<last;p+=temp.wsize)
    INLINEset_diff(p, p1, p);
    // inlining of macro set_free
    FREE(p1);
    if debug & MINCOV1 {
    fmt.Printf("MINCOV: family of all minimal coverings is\n");
    sf_print(temp);
    }
    sf_free(T);         /* this is the copy of T we made ... */
    return temp;
}

/*
 *  unate_intersect -- intersect two unate covers
 *
 *  If largest_only is TRUE, then only the largest cube(s) are returned
 */

const MAGIC = 500               /* save 500 cubes before containment */

pset_family unate_intersect(A, B, largest_only)
var A, B pset_family;
var largest_only bool;
{
    register pset pi, pj, lasti, lastj, pt;
    pset_family T, Tsave;
    bool save;
    int maxord, ord;

    /* How large should each temporary result cover be ? */
    T = sf_new(MAGIC, A.sf_size);
    Tsave = NULL;
    maxord = 0;
    pt = T.data;

    /* Form pairwise intersection of each set of A with each cube of B */
    // inlining of macro foreach_set
    for pi=A.data,lasti=pi+A.count*A.wsize;pi<lasti;pi+=A.wsize {

    // inlining of macro foreach_set
    for pj=B.data,lastj=pj+B.count*B.wsize;pj<lastj;pj+=B.wsize {

        save = set_andp(pt, pi, pj);

        /* Check if we want the largest only */
        if save && largest_only {
        if (ord = set_ord(pt)) > maxord {
            /* discard Tsave and T */
            if Tsave != NULL {
            sf_free(Tsave);
            Tsave = NULL;
            }
            pt = T.data;
            T.count = 0;
            /* Re-create pt (which was just thrown away) */
            (void) set_and(pt, pi, pj);
            maxord = ord;
        } else if ord < maxord {
            save = FALSE;
        }
        }

        if save {
        if ++T.count >= T.capacity {
            T = sf_contain(T);
            Tsave = (Tsave == NULL) ? T : sf_union(Tsave, T);
            T = sf_new(MAGIC, A.sf_size);
            pt = T.data;
        } else {
            pt += T.wsize;
        }
        }
    }
    }


    /* Contain the final result and merge it into Tsave */
    T = sf_contain(T);
    Tsave = (Tsave == NULL) ? T : sf_union(Tsave, T);

    return Tsave;
}

/*
 *  abs_covered -- after selecting a new column for the selected set,
 *  create a new matrix which is only those rows which are still uncovered
 */
static pset_family
abs_covered(A, pick)
var A pset_family;
register int pick;
{
    register pset last, p, pdest;
    register pset_family Aprime;

    Aprime = sf_new(A.count, A.sf_size);
    pdest = Aprime.data;
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if ! (p[(((pick) >> LOGBPI) + 1)] & (1 << ((pick) & (BPI-1)))) {
        INLINEset_copy(pdest, p);
        Aprime.count++;
        pdest += Aprime.wsize;
    }
    return Aprime;
}


/*
 *  abs_covered_many -- after selecting many columns for ther selected set,
 *  create a new matrix which is only those rows which are still uncovered
 */
static pset_family
abs_covered_many(A, pick_set)
var A pset_family;
register pset pick_set;
{
    register pset last, p, pdest;
    register pset_family Aprime;

    Aprime = sf_new(A.count, A.sf_size);
    pdest = Aprime.data;
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    if setp_disjoint(p, pick_set) {
        INLINEset_copy(pdest, p);
        Aprime.count++;
        pdest += Aprime.wsize;
    }
    return Aprime;
}


/*
 *  abs_select_restricted -- select the column of maximum column count which
 *  also belongs to the set "restrict"; weight each column of a set as
 *  1 / (set_ord(p) - 1).
 */
static int
abs_select_restricted(A, restrict)
var A pset_family;
var restrict pset;
{
    register int i, best_var, best_count, *count;

    /* Sum the elements in these columns */
    count = sf_count_restricted(A, restrict);

    /* Find which variable has maximum weight */
    best_var = -1;
    best_count = 0;
    for i = 0; i < A.sf_size; i++ {
    if count[i] > best_count {
        best_var = i;
        best_count = count[i];
    }
    }
    FREE(count);

    if (best_var == -1)
    fatal("abs_select_restricted: should not have best_var == -1");

    return best_var;
}
//ABC_NAMESPACE_IMPL_END


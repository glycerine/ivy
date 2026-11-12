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


// static bool primes_consensus_special_cases();

static pcover primes_consensus_merge();
static pcover and_with_cofactor(); 


/* primes_consensus -- generate primes using consensus */
pcover primes_consensus(T)
pcube *T;            /* T will be disposed of */
{
    register pcube cl, cr;
    register int best;
    pcover Tnew, Tl, Tr;

    if primes_consensus_special_cases(T, &Tnew) == MAYBE {
    // inlining of macro new_cube
    // inlining of macro set_new
    cl = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cr = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    best = binate_split_select(T, cl, cr, COMPL);

    Tl = primes_consensus(scofactor(T, cl, best));
    Tr = primes_consensus(scofactor(T, cr, best));
    Tnew = primes_consensus_merge(Tl, Tr, cl, cr);

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cl);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cr);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    }

    return Tnew;
}

static bool 
primes_consensus_special_cases(T, Tnew)
pcube *T;            /* will be disposed if answer is determined */
pcover *Tnew;            /* returned only if answer determined */
{
    register pcube *T1, p, ceil, cof=T[0];
    pcube last;
    pcover A;

    /* Check for no cubes in the cover */
    if T[2] == NULL {
    // inlining of macro new_cover
    *Tnew = sf_new(0, cube.size);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for only a single cube in the cover */
    if T[3] == NULL {
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), set_or(cof, cof, T[2]));
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for a row of all 1's (implies function is a tautology) */
    for T1 = T+2; (p = *T1++) != NULL;  {
    if full_row(p, cof) {
        // inlining of macro new_cover
        *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
        // inlining of macro free_cubelist
        FREE(T[0]); FREE(T);;
        return TRUE;
    }
    }

    /* Check for a column of all 0's which can be factored out */
    // inlining of macro set_save
    ceil = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cof))), cof);
    for T1 = T+2; (p = *T1++) != NULL;  {
    INLINEset_or(ceil, ceil, p);
    }
    if ! setp_equal(ceil, cube.fullset) {
    // inlining of macro new_cube
    // inlining of macro set_new
    p = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    (void) set_diff(p, cube.fullset, ceil);
    (void) set_or(cof, cof, p);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(p);

    A = primes_consensus(T);
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
        INLINEset_and(p, p, ceil);
    }
    *Tnew = A;
    // inlining of macro set_free
    FREE(ceil);
    return TRUE;
    }
    // inlining of macro set_free
    FREE(ceil);

    /* Collect column counts, determine unate variables, etc. */
    massive_count(T);

    /* If single active variable not factored out above, then tautology ! */
    if cdata.vars_active == 1 {
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Check for unate cover */
    } else if cdata.vars_unate == cdata.vars_active {
    A = cubeunlist(T);
    *Tnew = sf_contain(A);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Not much we can do about it */
    } else {
    return MAYBE;
    }
}

static pcover 
primes_consensus_merge(Tl, Tr, cl, cr)
pcover Tl, Tr;
pcube cl, cr;
{
    register pcube pl, pr, lastl, lastr, pt;
    pcover T, Tsave;

    Tl = and_with_cofactor(Tl, cl);
    Tr = and_with_cofactor(Tr, cr);

    T = sf_new(500, Tl.sf_size);
    pt = T.data;
    Tsave = sf_contain(sf_join(Tl, Tr));

    // inlining of macro foreach_set
    for pl=Tl.data,lastl=pl+Tl.count*Tl.wsize;pl<lastl;pl+=Tl.wsize {
    // inlining of macro foreach_set
    for pr=Tr.data,lastr=pr+Tr.count*Tr.wsize;pr<lastr;pr+=Tr.wsize {
        if cdist01(pl, pr) == 1 {
        consensus(pt, pl, pr);
        if ++T.count >= T.capacity {
            Tsave = sf_union(Tsave, sf_contain(T));
            T = sf_new(500, Tl.sf_size);
            pt = T.data;
        } else {
            pt += T.wsize;
        }
        }
    }
    }
    // inlining of macro free_cover
    sf_free(Tl);
    // inlining of macro free_cover
    sf_free(Tr);

    Tsave = sf_union(Tsave, sf_contain(T));
    return Tsave;
}


static pcover
and_with_cofactor(A, cof)
var A pset_family;
register pset cof;
{
    register pset last, p;

    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    INLINEset_and(p, p, cof);
    if cdist(p, cube.fullset) > 0 {
        // inlining of macro RESET
        (p[0] &= ~ (ACTIVE));
    } else {
        // inlining of macro SET
        (p[0] |= (ACTIVE));
    }
    }
    return sf_inactive(A);
}
//ABC_NAMESPACE_IMPL_END


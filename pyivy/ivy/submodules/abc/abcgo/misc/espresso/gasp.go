package espresso

import "fmt"

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
    module: gasp.c

    The "last_gasp" heuristic computes the reduction of each cube in
    the cover (without replacement) and then performs an expansion of
    these cubes.  The cubes which expand to cover some other cube are
    added to the original cover and irredundant finds a minimal subset.

    If one of the reduced cubes expands to cover some other reduced
    cube, then the new prime thus generated is a candidate for reducing
    the size of the cover.

    super_gasp is a variation on this strategy which extracts a minimal
    subset from the set of all prime implicants which cover all
    maximally reduced cubes.
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START



/*
 *  reduce_gasp -- compute the maximal reduction of each cube of F
 *
 *  If a cube does not reduce, it remains prime; otherwise, it is marked
 *  as nonprime.   If the cube is redundant (should NEVER happen here) we
 *  just crap out ...
 *
 *  A cover with all of the cubes of F is returned.  Those that did
 *  reduce are marked "NONPRIME"; those that reduced are marked "PRIME".
 *  The cubes are in the same order as in F.
 */
static pcover reduce_gasp(F, D)
pcover F, D;
{
    pcube p, last, cunder, *FD;
    pcover G;

    // inlining of macro new_cover
    G = sf_new(F.count, cube.size);
    FD = cube2list(F, D);

    /* Reduce cubes of F without replacement */
    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    cunder = reduce_cube(FD, p);
    if setp_empty(cunder) {
        fatal("empty reduction in reduce_gasp, shouldn't happen");
    } else if setp_equal(cunder, p) {
        // inlining of macro SET
        (cunder[0] |= (PRIME));            /* just to make sure */
        G = sf_addset(G, p);        /* it did not reduce ... */
    } else {
        // inlining of macro RESET
        (cunder[0] &= ~ (PRIME));        /* it reduced ... */
        G = sf_addset(G, cunder);
    }
    if debug & GASP {
        fmt.Printf("REDUCE_GASP: %s reduced to %s\n", pc1(p), pc2(cunder));
    }
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cunder);
    }

    // inlining of macro free_cubelist
    FREE(FD[0]); FREE(FD);;
    return G;
}

/*
 *  expand_gasp -- expand each nonprime cube of F into a prime implicant
 *
 *  The gasp strategy differs in that only those cubes which expand to
 *  cover some other cube are saved; also, all cubes are expanded
 *  regardless of whether they become covered or not.
 */

pcover expand_gasp(F, D, R, Foriginal)
INOUT pcover F;
IN pcover D;
IN pcover R;
IN pcover Foriginal;
{
    int c1index;
    pcover G;

    /* Try to expand each nonprime and noncovered cube */
    // inlining of macro new_cover
    G = sf_new(10, cube.size);
    for c1index = 0; c1index < F.count; c1index++ {
    expand1_gasp(F, D, R, Foriginal, c1index, &G);
    }
    G = sf_dupl(G);
    G = expand(G, R, /*nonsparse*/ FALSE);    /* Make them prime ! */
    return G;
}



/*
 *  expand1 -- Expand a single cube against the OFF-set, using the gasp strategy
 */
void expand1_gasp(F, D, R, Foriginal, c1index, G)
pcover F;        /* reduced cubes of ON-set */
pcover D;        /* DC-set */
pcover R;        /* OFF-set */
pcover Foriginal;    /* ON-set before reduction (same order as F) */
var c1index int32;        /* which index of F (or Freduced) to be checked */
pcover *G;
{
    register int c2index;
    register pcube p, last, c2under;
    pcube RAISE, FREESET, temp, *FD, c2essential;
    pcover F1;

    if debug & EXPAND1 {
    // inlining of macro GETSET
    fmt.Printf("\nEXPAND1_GASP:    \t%s\n", pc1(((F).data + (F).wsize * (c1index))));
    }

    // inlining of macro new_cube
    // inlining of macro set_new
    RAISE = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    FREESET = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    temp = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);

    /* Initialize the OFF-set */
    R.active_count = R.count;
    // inlining of macro foreach_set
    for p=R.data,last=p+R.count*R.wsize;p<last;p+=R.wsize {
    // inlining of macro SET
    (p[0] |= (ACTIVE));
    }
    /* Initialize the reduced ON-set, all nonprime cubes become active */
    F.active_count = F.count;
    // inlining of macro foreachi_set
    for c2under, c2index = F.data, 0;c2index<F.count;c2under+=F.wsize,c2index++ {
    // inlining of macro TESTP
    if c1index == c2index || (c2under[0] & (PRIME)) {
        F.active_count--;
        // inlining of macro RESET
        (c2under[0] &= ~ (ACTIVE));
    } else {
        // inlining of macro SET
        (c2under[0] |= (ACTIVE));
    }
    }

    /* Initialize the raising and unassigned sets */
    // inlining of macro GETSET
    (void) set_copy(RAISE, ((F).data + (F).wsize * (c1index)));
    (void) set_diff(FREESET, cube.fullset, RAISE);

    /* Determine parts which must be lowered */
    essen_parts(R, F, RAISE, FREESET);

    /* Determine parts which can always be raised */
    essen_raising(R, RAISE, FREESET);

    /* See which, if any, of the reduced cubes we can cover */
    // inlining of macro foreachi_set
    for c2under, c2index = F.data, 0;c2index<F.count;c2under+=F.wsize,c2index++ {
    // inlining of macro TESTP
    if c2under[0] & (ACTIVE) {
        /* See if this cube can be covered by an expansion */
        if setp_implies(c2under, RAISE) ||
          feasibly_covered(R, c2under, RAISE, temp) {
        
        /* See if c1under can expanded to cover c2 reduced against
         * (F - c1) u c1under; if so, c2 can definitely be removed !
         */

        /* Copy F and replace c1 with c1under */
        F1 = sf_save(Foriginal);
        // inlining of macro GETSET
        (void) set_copy(((F1).data + (F1).wsize * (c1index)), ((F).data + (F).wsize * (c1index)));

        /* Reduce c2 against ((F - c1) u c1under) */
        FD = cube2list(F1, D);
        // inlining of macro GETSET
        c2essential = reduce_cube(FD, ((F1).data + (F1).wsize * (c2index)));
        // inlining of macro free_cubelist
        FREE(FD[0]); FREE(FD);;
        sf_free(F1);

        /* See if c2essential is covered by an expansion of c1under */
        if feasibly_covered(R, c2essential, RAISE, temp) {
            (void) set_or(temp, RAISE, c2essential);
            // inlining of macro RESET
            (temp[0] &= ~ (PRIME));        /* cube not prime */
            *G = sf_addset(*G, temp);
        }
        // inlining of macro set_free
        FREE(c2essential);
        }
    }
    }

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(RAISE);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(FREESET);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(temp);
}

/* irred_gasp -- Add new primes to F and find an irredundant subset */
pcover irred_gasp(F, D, G)
pcover F, D, G;                 /* G is disposed of */
{
    if (G.count != 0)
    F = irredundant(sf_append(F, G), D);
    else
    // inlining of macro free_cover
    sf_free(G);
    return F;
}


/* last_gasp */
pcover last_gasp(F, D, R, cost)
pcover F, D, R;
var cost *cost_t;
{
    pcover G, G1;

    // inlining of macro EXECUTE
    {long t=ptime();G = reduce_gasp(F, D);totals(t,GREDUCE_TIME,G,&(*cost));};
    // inlining of macro EXECUTE
    {long t=ptime();G1 = expand_gasp(G, D, R, F);totals(t,GEXPAND_TIME,G1,&(*cost));};
    // inlining of macro free_cover
    sf_free(G);
    // inlining of macro EXECUTE
    {long t=ptime();F = irred_gasp(F, D, G1);totals(t,GIRRED_TIME,F,&(*cost));};
    return F;
}


/* super_gasp */
pcover super_gasp(F, D, R, cost)
pcover F, D, R;
var cost *cost_t;
{
    pcover G, G1;

    // inlining of macro EXECUTE
    {long t=ptime();G = reduce_gasp(F, D);totals(t,GREDUCE_TIME,G,&(*cost));};
    // inlining of macro EXECUTE
    {long t=ptime();G1 = all_primes(G, R);totals(t,GEXPAND_TIME,G1,&(*cost));};
    // inlining of macro free_cover
    sf_free(G);
    // inlining of macro EXEC
    {long t=ptime();G = sf_dupl(sf_append(F, G1));if(trace)print_trace(G,"NEWPRIMES",ptime()-t);};
    // inlining of macro EXECUTE
    {long t=ptime();F = irredundant(G, D);totals(t,IRRED_TIME,F,&(*cost));};
    return F;
}
//ABC_NAMESPACE_IMPL_END


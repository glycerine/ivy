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
    module: expand.c
    purpose: Perform the Espresso-II Expansion Step

    The idea is to take each nonprime cube of the on-set and expand it
    into a prime implicant such that we can cover as many other cubes
    of the on-set.  If no cube of the on-set can be covered, then we
    expand each cube into a large prime implicant by transforming the
    problem into a minimum covering problem which is solved by the
    heuristics of minimum_cover.

    These routines revolve around having a representation of the
    OFF-set.  (In contrast to the Espresso-II manuscript, we do NOT
    require an "unwrapped" version of the OFF-set).

    Some conventions on variable names:

    SUPER_CUBE is the supercube of all cubes which can be covered
    by an expansion of the cube being expanded

    OVEREXPANDED_CUBE is the cube which would result from expanding
    all parts which can expand individually of the cube being expanded

    RAISE is the current expansion of the current cube

    FREESET is the set of parts which haven't been raised or lowered yet.

    INIT_LOWER is a set of parts to be removed from the free parts before
    starting the expansion
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


/*
    expand -- expand each nonprime cube of F into a prime implicant

    If nonsparse is true, only the non-sparse variables will be expanded;
    this is done by forcing all of the sparse variables out of the free set.
*/

pcover expand(F, R, nonsparse)
INOUT pcover F;
IN pcover R;
IN bool nonsparse;              /* expand non-sparse variables only */
{
    register pcube last, p;
    pcube RAISE, FREESET, INIT_LOWER, SUPER_CUBE, OVEREXPANDED_CUBE;
    int var, num_covered;
    bool change;

    /* Order the cubes according to "chewing-away from the edges" of mini */
    if (use_random_order)
    F = random_order(F);
    else
    F = mini_sort(F, ascend);

    /* Allocate memory for variables needed by expand1() */
    // inlining of macro new_cube
    // inlining of macro set_new
    RAISE = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    FREESET = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    INIT_LOWER = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    SUPER_CUBE = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    OVEREXPANDED_CUBE = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);

    /* Setup the initial lowering set (differs only for nonsparse) */
    if (nonsparse)
    for(var = 0; var < cube.num_vars; var++)
        if (cube.sparse[var])
        (void) set_or(INIT_LOWER, INIT_LOWER, cube.var_mask[var]);

    /* Mark all cubes as not covered, and maybe essential */
    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    // inlining of macro RESET
    (p[0] &= ~ (COVERED));
    // inlining of macro RESET
    (p[0] &= ~ (NONESSEN));
    }

    /* Try to expand each nonprime and noncovered cube */
    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    /* do not expand if PRIME or if covered by previous expansion */
    // inlining of macro TESTP
    if ! (p[0] & (PRIME)) && ! (p[0] & (COVERED)) {

        /* expand the cube p, result is RAISE */
        expand1(R, F, RAISE, FREESET, OVEREXPANDED_CUBE, SUPER_CUBE,
        INIT_LOWER, &num_covered, p);
        if (debug & EXPAND)
        fmt.Printf("EXPAND: %s (covered %d)\n", pc1(p), num_covered);
        (void) set_copy(p, RAISE);
        // inlining of macro SET
        (p[0] |= (PRIME));
        // inlining of macro RESET
        (p[0] &= ~ (COVERED));        /* not really necessary */

        /* See if we generated an inessential prime */
        if num_covered == 0 && ! setp_equal(p, OVEREXPANDED_CUBE) {
        // inlining of macro SET
        (p[0] |= (NONESSEN));
        }
    }
    }

    /* Delete any cubes of F which became covered during the expansion */
    F.active_count = 0;
    change = FALSE;
    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    // inlining of macro TESTP
    if p[0] & (COVERED) {
        // inlining of macro RESET
        (p[0] &= ~ (ACTIVE));
        change = TRUE;
    } else {
        // inlining of macro SET
        (p[0] |= (ACTIVE));
        F.active_count++;
    }
    }
    if (change)
    F = sf_inactive(F);

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(RAISE);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(FREESET);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(INIT_LOWER);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(SUPER_CUBE);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(OVEREXPANDED_CUBE);
    return F;
}

/*
    expand1 -- Expand a single cube against the OFF-set
*/
void expand1(BB, CC, RAISE, FREESET, OVEREXPANDED_CUBE, SUPER_CUBE,
        INIT_LOWER, num_covered, c)
pcover BB;            /* Blocking matrix (OFF-set) */
pcover CC;            /* Covering matrix (ON-set) */
pcube RAISE;            /* The current parts which have been raised */
pcube FREESET;            /* The current parts which are free */
pcube OVEREXPANDED_CUBE;    /* Overexpanded cube of c */
pcube SUPER_CUBE;        /* Supercube of all cubes of CC we cover */
pcube INIT_LOWER;        /* Parts to initially remove from FREESET */
var num_covered *int32;        /* Number of cubes of CC which are covered */
pcube c;            /* The cube to be expanded */
{
    int bestindex;

    if (debug & EXPAND1)
    fmt.Printf("\nEXPAND1:    \t%s\n", pc1(c));

    /* initialize BB and CC */
    // inlining of macro SET
    (c[0] |= (PRIME));        /* don't try to cover ourself */
    setup_BB_CC(BB, CC);

    /* initialize count of # cubes covered, and the supercube of them */
    *num_covered = 0;
    (void) set_copy(SUPER_CUBE, c);

    /* Initialize the lowering, raising and unassigned sets */
    (void) set_copy(RAISE, c);
    (void) set_diff(FREESET, cube.fullset, RAISE);

    /* If some parts are forced into lowering set, remove them */
    if ! setp_empty(INIT_LOWER) {
    (void) set_diff(FREESET, FREESET, INIT_LOWER);
    elim_lowering(BB, CC, RAISE, FREESET);
    }

    /* Determine what can be raised, and return the over-expanded cube */
    essen_parts(BB, CC, RAISE, FREESET);
    (void) set_or(OVEREXPANDED_CUBE, RAISE, FREESET);

    /* While there are still cubes which can be covered, cover them ! */
    if CC.active_count > 0 {
    select_feasible(BB, CC, RAISE, FREESET, SUPER_CUBE, num_covered);
    }

    /* While there are still cubes covered by the overexpanded cube ... */
    while (CC.active_count > 0) {
    bestindex = most_frequent(CC, FREESET);
    // inlining of macro set_insert
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (RAISE[(((bestindex) >> LOGBPI) + 1)] |= 1 << ((bestindex) & (BPI-1)));
    // inlining of macro set_remove
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (FREESET[(((bestindex) >> LOGBPI) + 1)] &= ~ (1 << ((bestindex) & (BPI-1))));
    essen_parts(BB, CC, RAISE, FREESET);
    }

    /* Finally, when all else fails, choose the largest possible prime */
    /* We will loop only if we decide unravelling OFF-set is too expensive */
    while (BB.active_count > 0) {
    mincov(BB, RAISE, FREESET);
    }

    /* Raise any remaining free coordinates */
    (void) set_or(RAISE, RAISE, FREESET);
}

/*
    essen_parts -- determine which parts are forced into the lowering
    set to insure that the cube be orthognal to the OFF-set.

    If any cube of the OFF-set is distance 1 from the raising cube,
    then we must lower all parts of the conflicting variable.  (If the
    cube is distance 0, we detect this error here.)

    If there are essentially lowered parts, we can remove from consideration
    any cubes of the OFF-set which are more than distance 1 from the
    overexpanded cube of RAISE.
*/

void essen_parts(BB, CC, RAISE, FREESET)
pcover BB, CC;
pcube RAISE, FREESET;
{
    register pcube p, r = RAISE;
    pcube lastp, xlower = cube.temp[0];
    int dist;

    (void) set_copy(xlower, cube.emptyset);

    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=BB.data,lastp=p+BB.count*BB.wsize;p<lastp;p+=BB.wsize) if p[0] & (ACTIVE) {
#ifdef NO_INLINE
    if ((dist = cdist01(p, r)) > 1) goto exit_if;
#else
 {register int w,last;register unsigned int x;dist=0;if((last=cube.inword)!=-1)
{x=p[last]&r[last];if((x=~(x|x>>1)&cube.inmask))if((dist=count_ones(x))>1)goto
exit_if;for(w=1;w<last;w++){x=p[w]&r[w];if((x=~(x|x>>1)&DISJOINT))if(dist==1||(
dist+=count_ones(x))>1)goto exit_if;}}}{register int w,var,last;register pcube
mask;for(var=cube.num_binary_vars;var<cube.num_vars;var++){mask=cube.var_mask[
var];last=cube.last_word[var];for(w=cube.first_word[var];w<=last;w++)if(p[w]&r[
w]&mask[w])goto nextvar;if(++dist>1)goto exit_if;nextvar:;}}
#endif
    if dist == 0 {
        fatal("ON-set and OFF-set are not orthogonal");
    } else {
        (void) force_lower(xlower, p, r);
        BB.active_count--;
        // inlining of macro RESET
        (p[0] &= ~ (ACTIVE));
    }
exit_if: ;
    }

    if ! setp_empty(xlower) {
    (void) set_diff(FREESET, FREESET, xlower);/* remove from free set */
    elim_lowering(BB, CC, RAISE, FREESET);
    }

    if (debug & EXPAND1)
    fmt.Printf("ESSEN_PARTS:\tRAISE=%s FREESET=%s\n", pc1(RAISE), pc2(FREESET));
}

/*
    essen_raising -- determine which parts may always be added to
    the raising set without restricting further expansions

    General rule: if some part is not blocked by any cube of BB, then
    this part can always be raised.
*/

void essen_raising(BB, RAISE, FREESET)
register pcover BB;
pcube RAISE, FREESET;
{
    register pcube last, p, xraise = cube.temp[0];

    /* Form union of all cubes of BB, and then take complement wrt FREESET */
    (void) set_copy(xraise, cube.emptyset);
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=BB.data,last=p+BB.count*BB.wsize;p<last;p+=BB.wsize) if ((p[0] & (ACTIVE)))
    INLINEset_or(xraise, xraise, p);
    (void) set_diff(xraise, FREESET, xraise);

    (void) set_or(RAISE, RAISE, xraise);         /* add to raising set */
    (void) set_diff(FREESET, FREESET, xraise);       /* remove from free set */

    if (debug & EXPAND1)
    fmt.Printf("ESSEN_RAISING:\tRAISE=%s FREESET=%s\n", pc1(RAISE), pc2(FREESET));
}

/*
    elim_lowering -- after removing parts from FREESET, we can reduce the
    size of both BB and CC.

    We mark as inactive any cube of BB which does not intersect the
    overexpanded cube (i.e., RAISE + FREESET).  Likewise, we remove
    from CC any cube which is not covered by the overexpanded cube.
*/

void elim_lowering(BB, CC, RAISE, FREESET)
pcover BB, CC;
pcube RAISE, FREESET;
{
    register pcube p, r = set_or(cube.temp[0], RAISE, FREESET);
    pcube last;

    /*
     *  Remove sets of BB which are orthogonal to future expansions
     */
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=BB.data,last=p+BB.count*BB.wsize;p<last;p+=BB.wsize) if p[0] & (ACTIVE) {
#ifdef NO_INLINE
    if (! cdist0(p, r))
#else
 {register int w,lastw;register unsigned int x;if((lastw=cube.inword)!=-1){x=p[
lastw]&r[lastw];if(~(x|x>>1)&cube.inmask)goto false;for(w=1;w<lastw;w++){x=p[w]
&r[w];if(~(x|x>>1)&DISJOINT)goto false;}}}{register int w,var,lastw;register
pcube mask;for(var=cube.num_binary_vars;var<cube.num_vars;var++){mask=cube.
var_mask[var];lastw=cube.last_word[var];for(w=cube.first_word[var];w<=lastw;w++)
if(p[w]&r[w]&mask[w])goto nextvar;goto false;nextvar:;}}continue;false:
#endif
        // inlining of macro RESET
        BB.active_count--, (p[0] &= ~ (ACTIVE));
    }


    /*
     *  Remove sets of CC which cannot be covered by future expansions
     */
    if CC != (pcover) NULL {
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=CC.data,last=p+CC.count*CC.wsize;p<last;p+=CC.wsize) if p[0] & (ACTIVE) {
#ifdef NO_INLINE
        if (! setp_implies(p, r))
#else
        INLINEsetp_implies(p, r, /* when false => */ goto false1);
        /* when true => go to end of loop */ continue;
        false1:
#endif
        // inlining of macro RESET
        CC.active_count--, (p[0] &= ~ (ACTIVE));
    }
    }
}

/*
    most_frequent -- When all else fails, select a reasonable part to raise
    The active cubes of CC are the cubes which are covered by the
    overexpanded cube of the original cube (however, we know that none
    of them can actually be covered by a feasible expansion of the
    original cube).  We resort to the MINI strategy of selecting to
    raise the part which will cover the same part in the most cubes of CC.
*/
int most_frequent(CC, FREESET)
pcover CC;
pcube FREESET;
{
    register int i, best_part, best_count, *count;
    register pset p, last;

    /* Count occurences of each variable */
    count = ALLOC(int, cube.size);
    for(i = 0; i < cube.size; i++)
    count[i] = 0;
    if (CC != (pcover) NULL)
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=CC.data,last=p+CC.count*CC.wsize;p<last;p+=CC.wsize) if ((p[0] & (ACTIVE)))
        set_adjcnt(p, count, 1);

    /* Now find which free part occurs most often */
    best_count = best_part = -1;
    for(i = 0; i < cube.size; i++)
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if (FREESET[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) && count[i] > best_count {
        best_part = i;
        best_count = count[i];
    }
    FREE(count);

    if (debug & EXPAND1)
    fmt.Printf("MOST_FREQUENT:\tbest=%d FREESET=%s\n", best_part, pc2(FREESET));
    return best_part;
}

/*
    setup_BB_CC -- set up the blocking and covering set families;

    Note that the blocking family is merely the set of cubes of R, and
    that CC is the set of cubes of F which might possibly be covered
    (i.e., nonprime cubes, and cubes not already covered)
*/

void setup_BB_CC(BB, CC)
register pcover BB, CC;
{
    register pcube p, last;

    /* Create the block and cover set families */
    BB.active_count = BB.count;
    // inlining of macro foreach_set
    for(p=BB.data,last=p+BB.count*BB.wsize;p<last;p+=BB.wsize)
    // inlining of macro SET
    (p[0] |= (ACTIVE));

    if CC != (pcover) NULL {
    CC.active_count = CC.count;
    // inlining of macro foreach_set
    for(p=CC.data,last=p+CC.count*CC.wsize;p<last;p+=CC.wsize)
        // inlining of macro TESTP
        if ((p[0] & (COVERED)) || (p[0] & (PRIME)))
        // inlining of macro RESET
        CC.active_count--, (p[0] &= ~ (ACTIVE));
        else
        // inlining of macro SET
        (p[0] |= (ACTIVE));
    }
}

/*
    select_feasible -- Determine if there are cubes which can be covered,
    and if so, raise those parts necessary to cover as many as possible.

    We really don't check to maximize the number that can be covered;
    instead, we check, for each fcc, how many other fcc remain fcc
    after expanding to cover the fcc.  (Essentially one-level lookahead).
*/

void select_feasible(BB, CC, RAISE, FREESET, SUPER_CUBE, num_covered)
pcover BB, CC;
pcube RAISE, FREESET, SUPER_CUBE;
var num_covered *int32;
{
    register pcube p, last;
    register pcube bestfeas = NULL; // Suppress "might be used uninitialized"
    register pcube *feas;
    register int i, j;
    pcube *feas_new_lower;
    int bestcount, bestsize, count, size, numfeas, lastfeas;
    pcover new_lower;

    /*  Start out with all cubes covered by the over-expanded cube as
     *  the "possibly" feasibly-covered cubes (pfcc)
     */
    feas = ALLOC(pcube, CC.active_count);
    numfeas = 0;
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=CC.data,last=p+CC.count*CC.wsize;p<last;p+=CC.wsize) if ((p[0] & (ACTIVE)))
    feas[numfeas++] = p;

    /* Setup extra cubes to record parts forced low after a covering */
    feas_new_lower = ALLOC(pcube, CC.active_count);
    // inlining of macro new_cover
    new_lower = sf_new(numfeas, cube.size);
    for(i = 0; i < numfeas; i++)
    // inlining of macro GETSET
    feas_new_lower[i] = ((new_lower).data + (new_lower).wsize * (i));


loop:
    /* Find the essentially raised parts -- this might cover some cubes
       for us, without having to find out if they are fcc or not
    */
    essen_raising(BB, RAISE, FREESET);

    /* Now check all "possibly" feasibly covered cubes to check feasibility */
    lastfeas = numfeas;
    numfeas = 0;
    for i = 0; i < lastfeas; i++ {
    p = feas[i];

    /* Check active because essen_parts might have removed it */
    // inlining of macro TESTP
    if p[0] & (ACTIVE) {

        /*  See if the cube is already covered by RAISE --
         *  this can happen because of essen_raising() or because of
         *  the previous "loop"
         */
        if setp_implies(p, RAISE) {
        (*num_covered) += 1;
        (void) set_or(SUPER_CUBE, SUPER_CUBE, p);
        CC.active_count--;
        // inlining of macro RESET
        (p[0] &= ~ (ACTIVE));
        // inlining of macro SET
        (p[0] |= (COVERED));
        /* otherwise, test if it is feasibly covered */
        } else if feasibly_covered(BB,p,RAISE,feas_new_lower[numfeas]) {
        feas[numfeas] = p;            /* save the fcc */
        numfeas++;
        }
    }
    }
    if (debug & EXPAND1)
    fmt.Printf("SELECT_FEASIBLE: started with %d pfcc, ended with %d fcc\n", lastfeas, numfeas);

    /* Exit here if there are no feasibly covered cubes */
    if numfeas == 0 {
    FREE(feas);
    FREE(feas_new_lower);
    // inlining of macro free_cover
    sf_free(new_lower);
    return;
    }

    /* Now find which is the best feasibly covered cube */
    bestcount = 0;
    bestsize = 9999;
    for i = 0; i < numfeas; i++ {
    size = set_dist(feas[i], FREESET);    /* # of newly raised parts */
    count = 0;    /* # of other cubes which remain fcc after raising */

#define NEW
          
    for(j = 0; j < numfeas; j++)
        if (setp_disjoint(feas_new_lower[i], feas[j]))
        count++;
     
                                
                                           
                
      
    if count > bestcount {
        bestcount = count;
        bestfeas = feas[i];
        bestsize = size;
    } else if count == bestcount && size < bestsize {
        bestfeas = feas[i];
        bestsize = size;
    }
    }

    /* Add the necessary parts to the raising set */
    (void) set_or(RAISE, RAISE, bestfeas);
    (void) set_diff(FREESET, FREESET, RAISE);
    if (debug & EXPAND1)
    fmt.Printf("FEASIBLE:  \tRAISE=%s FREESET=%s\n", pc1(RAISE), pc2(FREESET));
    essen_parts(BB, CC, RAISE, FREESET);
    goto loop;
/* NOTREACHED */
}

/*
    feasibly_covered -- determine if the cube c is feasibly covered
    (i.e., if it is possible to raise all of the necessary variables
    while still insuring orthogonality with R).  Also, if c is feasibly
    covered, then compute the new set of parts which are forced into
    the lowering set.
*/

bool feasibly_covered(BB, c, RAISE, new_lower)
pcover BB;
pcube c, RAISE, new_lower;
{
    register pcube p, r = set_or(cube.temp[0], RAISE, c);
    int dist;
    pcube lastp;

    set_copy(new_lower, cube.emptyset);
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=BB.data,lastp=p+BB.count*BB.wsize;p<lastp;p+=BB.wsize) if p[0] & (ACTIVE) {
#ifdef NO_INLINE
    if ((dist = cdist01(p, r)) > 1) goto exit_if;
#else
 {register int w,last;register unsigned int x;dist=0;if((last=cube.inword)!=-1)
{x=p[last]&r[last];if((x=~(x|x>>1)&cube.inmask))if((dist=count_ones(x))>1)goto
exit_if;for(w=1;w<last;w++){x=p[w]&r[w];if((x=~(x|x>>1)&DISJOINT))if(dist==1||(
dist+=count_ones(x))>1)goto exit_if;}}}{register int w,var,last;register pcube
mask;for(var=cube.num_binary_vars;var<cube.num_vars;var++){mask=cube.var_mask[
var];last=cube.last_word[var];for(w=cube.first_word[var];w<=last;w++)if(p[w]&r[
w]&mask[w])goto nextvar;if(++dist>1)goto exit_if;nextvar:;}}
#endif
    if (dist == 0)
        return FALSE;
    else
        (void) force_lower(new_lower, p, r);
    exit_if: ;
    }
    return TRUE;
}

/*
    mincov -- transform the problem of expanding a cube to a maximally-
    large prime implicant into the problem of selecting a minimum
    cardinality cover over a family of sets.

    When we get to this point, we must unravel the remaining off-set.
    This may be painful.
*/

void mincov(BB, RAISE, FREESET)
pcover BB;
pcube RAISE, FREESET;
{
    int expansion, nset, var, dist;
    pset_family B;
    register pcube xraise=cube.temp[0], xlower, p, last, plower;

#ifdef RANDOM_MINCOV
#if defined(_POSIX_SOURCE) || defined(__SVR4)
    dist = rand() % set_ord(FREESET);
#else
    dist = random() % set_ord(FREESET);
#endif
    for(var = 0; var < cube.size && dist >= 0; var++) {
    if (is_in_set(FREESET, var)) {
        dist--;
    }
    }

    set_insert(RAISE, var);
    set_remove(FREESET, var);
    (void) essen_parts(BB, /*CC*/ (pcover) NULL, RAISE, FREESET);
#else

    /* Create B which are those cubes which we must avoid intersecting */
    B = new_cover(BB->active_count);
    foreach_active_set(BB, last, p) {
    plower = set_copy(GETSET(B, B->count++), cube.emptyset);
    (void) force_lower(plower, p, RAISE);
    }

    /* Determine how many sets it will blow up into after the unravel */
    nset = 0;
    foreach_set(B, last, p) {
    expansion = 1;
    for(var = cube.num_binary_vars; var < cube.num_vars; var++) {
        if ((dist=set_dist(p, cube.var_mask[var])) > 1) {
        expansion *= dist;
        if (expansion > 500) goto heuristic_mincov;
        }
    }
    nset += expansion;
    if (nset > 500) goto heuristic_mincov;
    }

    B = unravel(B, cube.num_binary_vars);
    xlower = do_sm_minimum_cover(B);

    /* Add any remaining free parts to the raising set */
    (void) set_or(RAISE, RAISE, set_diff(xraise, FREESET, xlower));
    (void) set_copy(FREESET, cube.emptyset);    /* free set is empty */
    BB->active_count = 0;            /* BB satisfied */
    if (debug & EXPAND1) {
    printf("MINCOV:    \tRAISE=%s FREESET=%s\n", pc1(RAISE), pc2(FREESET));
    }
    sf_free(B);
    set_free(xlower);
    return;

heuristic_mincov:
    sf_free(B);
    /* most_frequent will pick first free part */
    set_insert(RAISE, most_frequent(/*CC*/ (pcover) NULL, FREESET));
    (void) set_diff(FREESET, FREESET, RAISE);
    essen_parts(BB, /*CC*/ (pcover) NULL, RAISE, FREESET);
    return;
#endif
}

/*
    find_all_primes -- find all of the primes which cover the
    currently reduced BB
*/
pcover find_all_primes(BB, RAISE, FREESET)
pcover BB;
register pcube RAISE, FREESET;
{
    register pset last, p, plower;
    pset_family B, B1;

    if BB.active_count == 0 {
    // inlining of macro new_cover
    B1 = sf_new(1, cube.size);
    // inlining of macro GETSET
    p = ((B1).data + (B1).wsize * (B1.count++));
    (void) set_copy(p, RAISE);
    // inlining of macro SET
    (p[0] |= (PRIME));
    } else {
    // inlining of macro new_cover
    B = sf_new(BB.active_count, cube.size);
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=BB.data,last=p+BB.count*BB.wsize;p<last;p+=BB.wsize) if p[0] & (ACTIVE) {
        // inlining of macro GETSET
        plower = set_copy(((B).data + (B).wsize * (B.count++)), cube.emptyset);
        (void) force_lower(plower, p, RAISE);
    }
    B = sf_rev_contain(unravel(B, cube.num_binary_vars));
    B1 = exact_minimum_cover(B);
    // inlining of macro foreach_set
    for p=B1.data,last=p+B1.count*B1.wsize;p<last;p+=B1.wsize {
        INLINEset_diff(p, FREESET, p);
        INLINEset_or(p, p, RAISE);
        // inlining of macro SET
        (p[0] |= (PRIME));
    }
    // inlining of macro free_cover
    sf_free(B);
    }
    return B1;
}

/*
    all_primes -- foreach cube in F, generate all of the primes
    which cover the cube.
*/

pcover all_primes(F, R)
pcover F, R;
{
    register pcube last, p, RAISE, FREESET;
    pcover Fall_primes, B1;

    // inlining of macro new_cube
    // inlining of macro set_new
    FREESET = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    RAISE = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cover
    Fall_primes = sf_new(F.count, cube.size);

    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    // inlining of macro TESTP
    if p[0] & (PRIME) {
        Fall_primes = sf_addset(Fall_primes, p);
    } else {
        /* Setup for call to essential parts */
        (void) set_copy(RAISE, p);
        (void) set_diff(FREESET, cube.fullset, RAISE);
        setup_BB_CC(R, /* CC */ (pcover) NULL);
        essen_parts(R, /* CC */ (pcover) NULL, RAISE, FREESET);

        /* Find all of the primes, and add them to the prime set */
        B1 = find_all_primes(R, RAISE, FREESET);
        Fall_primes = sf_append(Fall_primes, B1);
    }
    }

    // inlining of macro set_free
    FREE(RAISE);
    // inlining of macro set_free
    FREE(FREESET);
    return Fall_primes;
}
//ABC_NAMESPACE_IMPL_END

